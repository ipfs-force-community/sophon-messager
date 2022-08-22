package testhelper

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/asaskevich/EventBus"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-bitfield"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/builtin/v8/miner"
	"github.com/filecoin-project/go-state-types/crypto"
	"github.com/filecoin-project/go-state-types/dline"
	"github.com/filecoin-project/go-state-types/network"
	lminer "github.com/filecoin-project/venus/venus-shared/actors/builtin/miner"
	v1 "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	"github.com/filecoin-project/venus/venus-shared/types"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p-core/metrics"
	network2 "github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
)

const (
	maxMsgInBlock  = 200
	newTipsetTopic = "new_tipset"
)

var ErrGasLimitNegative = errors.New("gas limit is negative")

var (
	DefGasUsed    = int64(10000)
	DefGasPremium = abi.NewTokenAmount(1000)
	DefGasFeeCap  = abi.NewTokenAmount(10000)
	defBalance    = abi.NewTokenAmount(1000)

	// MinPackedPremium If the gas premium is lower than this value, the message will not be packaged
	MinPackedPremium = abi.NewTokenAmount(500)
)

type MockFullNode struct {
	miner address.Address

	actors map[address.Address]*types.Actor

	ts     map[types.TipSetKey]*types.TipSet
	currTS *types.TipSet

	blockDelay  time.Duration
	blockInfos  map[cid.Cid]*blockInfo
	chainMsgs   map[cid.Cid]*types.SignedMessage
	msgReceipts map[cid.Cid]*types.MessageReceipt

	pendingMsgs []*types.SignedMessage

	eventBus EventBus.Bus

	l sync.Mutex
}

type blockInfo struct {
	bh   *types.BlockHeader
	msgs []cid.Cid
}

func NewMockFullNode(blockDelay time.Duration) (*MockFullNode, error) {
	miner, err := address.NewIDAddress(10001)
	if err != nil {
		return nil, err
	}
	node := &MockFullNode{
		blockDelay:  blockDelay,
		miner:       miner,
		actors:      make(map[address.Address]*types.Actor),
		ts:          make(map[types.TipSetKey]*types.TipSet),
		blockInfos:  make(map[cid.Cid]*blockInfo),
		chainMsgs:   make(map[cid.Cid]*types.SignedMessage),
		msgReceipts: make(map[cid.Cid]*types.MessageReceipt),
		eventBus:    EventBus.New(),
	}
	bh, err := genBlockHead(miner, 0, []cid.Cid{})
	if err != nil {
		return nil, err
	}
	ts, err := types.NewTipSet([]*types.BlockHeader{bh})
	if err != nil {
		return nil, err
	}
	node.ts[ts.Key()] = ts
	node.setHead(ts)
	node.blockInfos[bh.Cid()] = &blockInfo{bh: bh}
	node.eventBus.Publish(newTipsetTopic, ts)

	go node.tipsetProvider()

	return node, nil
}

func (f *MockFullNode) AddActors(addrs []address.Address) error {
	var err error
	for _, addr := range addrs {
		if addr.Protocol() == address.ID {
			addr, err = ResolveIDAddr(addr)
			if err != nil {
				return err
			}
		}
		_, ok := f.actors[addr]
		if !ok {
			f.actors[addr] = &types.Actor{Nonce: 0, Balance: defBalance}
		}
	}
	return nil
}

func (f *MockFullNode) tipsetProvider() {
	ticker := time.NewTicker(f.blockDelay)
	defer ticker.Stop()

	for range ticker.C {
		bh, err := f.blockProvider()
		if err != nil {
			panic(err)
		}
		ts, err := types.NewTipSet([]*types.BlockHeader{bh})
		if err != nil {
			panic(err)
		}
		f.l.Lock()
		f.ts[ts.Key()] = ts
		f.l.Unlock()
		f.setHead(ts)
		f.eventBus.Publish(newTipsetTopic, ts)
	}
}

func (f *MockFullNode) setHead(ts *types.TipSet) {
	f.l.Lock()
	defer f.l.Unlock()
	f.currTS = ts
}

func (f *MockFullNode) blockProvider() (*types.BlockHeader, error) {
	head, err := f.ChainHead(context.Background())
	if err != nil {
		return nil, err
	}
	bh, err := genBlockHead(f.miner, head.Height()+1, head.Cids())
	if err != nil {
		return nil, err
	}
	f.l.Lock()
	pendingMsgLen := len(f.pendingMsgs)
	cids := make([]cid.Cid, 0, pendingMsgLen)
	cidMap := make(map[cid.Cid]struct{}, pendingMsgLen)
	for i, msg := range f.pendingMsgs {
		actor, ok := f.actors[msg.Message.From]
		if !ok {
			continue
		}
		if msg.Message.GasPremium.LessThan(MinPackedPremium) {
			continue
		}

		actor.Nonce++
		c := msg.Cid()
		f.chainMsgs[c] = msg
		f.msgReceipts[c] = &types.MessageReceipt{ExitCode: -1, GasUsed: DefGasUsed}
		cids = append(cids, c)
		cidMap[c] = struct{}{}
		if i >= maxMsgInBlock {
			break
		}
	}

	tmpMsg := make([]*types.SignedMessage, 0, pendingMsgLen)
	for _, msg := range f.pendingMsgs {
		if _, ok := cidMap[msg.Cid()]; !ok {
			tmpMsg = append(tmpMsg, msg)
		}
	}
	f.pendingMsgs = tmpMsg

	f.blockInfos[bh.Cid()] = &blockInfo{
		bh:   bh,
		msgs: cids,
	}
	f.l.Unlock()

	return bh, nil
}

//// full api ////

func (f *MockFullNode) StateAccountKey(ctx context.Context, addr address.Address, tsk types.TipSetKey) (address.Address, error) {
	if addr.Protocol() != address.ID {
		return addr, nil
	}
	return ResolveIDAddr(addr)
}

func (f *MockFullNode) StateNetworkName(ctx context.Context) (types.NetworkName, error) {
	return types.NetworkNameMain, nil
}

func (f *MockFullNode) ChainGetParentMessages(ctx context.Context, bcid cid.Cid) ([]types.MessageCID, error) {
	f.l.Lock()
	defer f.l.Unlock()
	blkInfo, ok := f.blockInfos[bcid]
	if !ok {
		return nil, fmt.Errorf("not found block %v", bcid)
	}
	msgCid := make([]types.MessageCID, 0, len(blkInfo.msgs))
	for _, c := range blkInfo.msgs {
		msgCid = append(msgCid, types.MessageCID{
			Cid:     c,
			Message: f.chainMsgs[c].VMMessage(),
		})
	}

	return msgCid, nil
}

func (f *MockFullNode) ChainGetParentReceipts(ctx context.Context, bcid cid.Cid) ([]*types.MessageReceipt, error) {
	f.l.Lock()
	defer f.l.Unlock()
	blkInfo, ok := f.blockInfos[bcid]
	if !ok {
		return nil, fmt.Errorf("not found block %v", bcid)
	}
	receipts := make([]*types.MessageReceipt, 0, len(blkInfo.msgs))
	for _, c := range blkInfo.msgs {
		receipts = append(receipts, f.msgReceipts[c])
	}

	return receipts, nil
}

func (f *MockFullNode) ChainGetTipSet(ctx context.Context, key types.TipSetKey) (*types.TipSet, error) {
	f.l.Lock()
	defer f.l.Unlock()
	ts, ok := f.ts[key]
	if !ok {
		return nil, fmt.Errorf("not found %s", key)
	}
	return ts, nil
}

func (f *MockFullNode) ChainGetMessagesInTipset(ctx context.Context, key types.TipSetKey) ([]types.MessageCID, error) {
	f.l.Lock()
	defer f.l.Unlock()
	_, ok := f.ts[key]
	if !ok {
		return nil, fmt.Errorf("not found tipset %v", key)
	}
	msgs := make([]types.MessageCID, 0)
	for _, c := range key.Cids() {
		blkInfo, ok := f.blockInfos[c]
		if !ok {
			continue
		}
		for _, c := range blkInfo.msgs {
			msgs = append(msgs, types.MessageCID{
				Cid:     c,
				Message: f.chainMsgs[c].VMMessage(),
			})
		}
	}
	return msgs, nil
}

func (f *MockFullNode) ChainHead(ctx context.Context) (*types.TipSet, error) {
	f.l.Lock()
	defer f.l.Unlock()
	return f.currTS, nil
}

func (f *MockFullNode) StateGetActor(ctx context.Context, addr address.Address, tsk types.TipSetKey) (*types.Actor, error) {
	f.l.Lock()
	defer f.l.Unlock()
	actor, ok := f.actors[addr]
	if !ok {
		return nil, fmt.Errorf("not found actor %v", addr)
	}
	return actor, nil
}

func (f *MockFullNode) GasBatchEstimateMessageGas(ctx context.Context, estimateMessages []*types.EstimateMessage, fromNonce uint64, tsk types.TipSetKey) ([]*types.EstimateResult, error) {
	var err error
	res := make([]*types.EstimateResult, 0, len(estimateMessages))
	for _, msg := range estimateMessages {
		msg.Msg, err = f.GasEstimateMessageGas(ctx, msg.Msg, msg.Spec, tsk)
		if err != nil {
			res = append(res, &types.EstimateResult{
				Msg: msg.Msg,
				Err: err.Error(),
			})
			continue
		}
		msg.Msg.Nonce = fromNonce
		res = append(res, &types.EstimateResult{
			Msg: msg.Msg,
			Err: "",
		})
		fromNonce++
	}

	return res, nil
}

func (f *MockFullNode) GasEstimateMessageGas(ctx context.Context, msg *types.Message, spec *types.MessageSendSpec, tsk types.TipSetKey) (*types.Message, error) {
	err := estimateGasLimit(msg, spec)
	if err != nil {
		return nil, err
	}

	if msg.GasPremium.NilOrZero() {
		msg.GasPremium = DefGasPremium
		if spec != nil && spec.GasOverPremium > 0 {
			msg.GasPremium = big.Div(big.Mul(msg.GasPremium, big.NewInt(int64(spec.GasOverPremium*10000))), big.NewInt(10000))
		}
	}
	if msg.GasFeeCap.NilOrZero() {
		msg.GasFeeCap = big.Add(DefGasFeeCap, msg.GasPremium)
	}

	if spec != nil && !spec.MaxFee.NilOrZero() {
		gl := types.NewInt(uint64(msg.GasLimit))
		totalFee := types.BigMul(msg.GasFeeCap, gl)

		if !totalFee.LessThanEqual(spec.MaxFee) {
			msg.GasFeeCap = big.Div(spec.MaxFee, gl)
			msg.GasPremium = big.Min(msg.GasFeeCap, msg.GasPremium)
		}
	}

	return msg, nil
}

func estimateGasLimit(msg *types.Message, spec *types.MessageSendSpec) error {
	// Estimation failure when GasLimit is less than 0
	if msg.GasLimit < 0 {
		return fmt.Errorf("failed to estimate gas: %w", ErrGasLimitNegative)
	}
	if msg.GasLimit > 0 {
		return nil
	}
	msg.GasLimit = DefGasUsed
	if spec != nil {
		if spec.GasOverEstimation > 0 {
			msg.GasLimit = int64(float64(msg.GasLimit) * spec.GasOverEstimation)
		}
	}

	return nil
}

func (f *MockFullNode) MpoolBatchPush(ctx context.Context, smsgs []*types.SignedMessage) ([]cid.Cid, error) {
	f.l.Lock()
	defer f.l.Unlock()
	cids := make([]cid.Cid, 0, len(smsgs))
	for _, msg := range smsgs {
		// todo: check nonce
		for i, m := range f.pendingMsgs {
			if m.Message.From == msg.Message.From && m.Message.Nonce == msg.Message.Nonce {
				f.pendingMsgs[i] = msg
				break
			}
		}
		f.pendingMsgs = append(f.pendingMsgs, msg)
		cids = append(cids, msg.Cid())
	}
	return cids, nil
}

func (f *MockFullNode) StateSearchMsg(ctx context.Context, from types.TipSetKey, msgCid cid.Cid, limit abi.ChainEpoch, allowReplaced bool) (*types.MsgLookup, error) {
	f.l.Lock()
	defer f.l.Unlock()
	_, ok := f.chainMsgs[msgCid]
	if !ok {
		return nil, fmt.Errorf("not found %s", msgCid)
	}
	msgLookup := &types.MsgLookup{
		Message: msgCid,
		Receipt: *f.msgReceipts[msgCid],
		TipSet:  types.TipSetKey{},
	}
	for _, blkInfo := range f.blockInfos {
		for _, c := range blkInfo.msgs {
			if c == msgCid {
				msgLookup.Height = blkInfo.bh.Height
				goto loopOver
			}
		}
	}
loopOver:
	for _, ts := range f.ts {
		if ts.Height() == msgLookup.Height {
			msgLookup.TipSet = ts.Key()
		}
	}
	return msgLookup, nil
}

func (f *MockFullNode) ChainNotify(ctx context.Context) (<-chan []*types.HeadChange, error) {
	head, err := f.ChainHead(context.Background())
	if err != nil {
		return nil, err
	}
	out := make(chan []*types.HeadChange, 16)
	out <- []*types.HeadChange{
		{
			Type: types.HCCurrent,
			Val:  head,
		},
	}
	var done bool
	_ = f.eventBus.Subscribe(newTipsetTopic, func(ts *types.TipSet) {
		if !done {
			// to test UpdateAllFilledMessage and testUpdateFilledMessageByID
			time.Sleep(f.blockDelay / 4)
			out <- []*types.HeadChange{
				{
					Type: types.HCApply,
					Val:  ts,
				},
			}
		}
	})
	go func() {
		<-ctx.Done()
		close(out)
		done = true
	}()

	return out, nil
}

func (f *MockFullNode) ChainReadObj(ctx context.Context, cid cid.Cid) ([]byte, error) {
	panic("implement me")
}

func (f *MockFullNode) ChainDeleteObj(ctx context.Context, obj cid.Cid) error {
	panic("implement me")
}

func (f *MockFullNode) ChainHasObj(ctx context.Context, obj cid.Cid) (bool, error) {
	panic("implement me")
}

func (f *MockFullNode) ChainStatObj(ctx context.Context, obj cid.Cid, base cid.Cid) (types.ObjStat, error) {
	panic("implement me")
}

func (f *MockFullNode) ChainPutObj(ctx context.Context, block blocks.Block) error {
	panic("implement me")
}

func (f *MockFullNode) ListActor(ctx context.Context) (map[address.Address]*types.Actor, error) {
	panic("implement me")
}

func (f *MockFullNode) StateMinerSectorAllocated(ctx context.Context, maddr address.Address, s abi.SectorNumber, tsk types.TipSetKey) (bool, error) {
	panic("implement me")
}

func (f *MockFullNode) StateSectorPreCommitInfo(ctx context.Context, maddr address.Address, n abi.SectorNumber, tsk types.TipSetKey) (miner.SectorPreCommitOnChainInfo, error) {
	panic("implement me")
}

func (f *MockFullNode) StateSectorGetInfo(ctx context.Context, maddr address.Address, n abi.SectorNumber, tsk types.TipSetKey) (*miner.SectorOnChainInfo, error) {
	panic("implement me")
}

func (f *MockFullNode) StateSectorPartition(ctx context.Context, maddr address.Address, sectorNumber abi.SectorNumber, tsk types.TipSetKey) (*lminer.SectorLocation, error) {
	panic("implement me")
}

func (f *MockFullNode) StateMinerSectorSize(ctx context.Context, maddr address.Address, tsk types.TipSetKey) (abi.SectorSize, error) {
	panic("implement me")
}

func (f *MockFullNode) StateMinerInfo(ctx context.Context, maddr address.Address, tsk types.TipSetKey) (types.MinerInfo, error) {
	panic("implement me")
}

func (f *MockFullNode) StateMinerWorkerAddress(ctx context.Context, maddr address.Address, tsk types.TipSetKey) (address.Address, error) {
	panic("implement me")
}

func (f *MockFullNode) StateMinerRecoveries(ctx context.Context, maddr address.Address, tsk types.TipSetKey) (bitfield.BitField, error) {
	panic("implement me")
}

func (f *MockFullNode) StateMinerFaults(ctx context.Context, maddr address.Address, tsk types.TipSetKey) (bitfield.BitField, error) {
	panic("implement me")
}

func (f *MockFullNode) StateMinerProvingDeadline(ctx context.Context, maddr address.Address, tsk types.TipSetKey) (*dline.Info, error) {
	panic("implement me")
}

func (f *MockFullNode) StateMinerPartitions(ctx context.Context, maddr address.Address, dlIdx uint64, tsk types.TipSetKey) ([]types.Partition, error) {
	panic("implement me")
}

func (f *MockFullNode) StateMinerDeadlines(ctx context.Context, maddr address.Address, tsk types.TipSetKey) ([]types.Deadline, error) {
	panic("implement me")
}

func (f *MockFullNode) StateMinerSectors(ctx context.Context, maddr address.Address, sectorNos *bitfield.BitField, tsk types.TipSetKey) ([]*miner.SectorOnChainInfo, error) {
	panic("implement me")
}

func (f *MockFullNode) StateMarketStorageDeal(ctx context.Context, dealID abi.DealID, tsk types.TipSetKey) (*types.MarketDeal, error) {
	panic("implement me")
}

func (f *MockFullNode) StateMinerPreCommitDepositForPower(ctx context.Context, maddr address.Address, pci miner.SectorPreCommitInfo, tsk types.TipSetKey) (big.Int, error) {
	panic("implement me")
}

func (f *MockFullNode) StateMinerInitialPledgeCollateral(ctx context.Context, maddr address.Address, pci miner.SectorPreCommitInfo, tsk types.TipSetKey) (big.Int, error) {
	panic("implement me")
}

func (f *MockFullNode) StateVMCirculatingSupplyInternal(ctx context.Context, tsk types.TipSetKey) (types.CirculatingSupply, error) {
	panic("implement me")
}

func (f *MockFullNode) StateCirculatingSupply(ctx context.Context, tsk types.TipSetKey) (abi.TokenAmount, error) {
	panic("implement me")
}

func (f *MockFullNode) StateMarketDeals(ctx context.Context, tsk types.TipSetKey) (map[string]*types.MarketDeal, error) {
	panic("implement me")
}

func (f *MockFullNode) StateMinerActiveSectors(ctx context.Context, maddr address.Address, tsk types.TipSetKey) ([]*miner.SectorOnChainInfo, error) {
	panic("implement me")
}

func (f *MockFullNode) StateLookupID(ctx context.Context, addr address.Address, tsk types.TipSetKey) (address.Address, error) {
	panic("implement me")
}

func (f *MockFullNode) StateLookupRobustAddress(ctx context.Context, address address.Address, key types.TipSetKey) (address.Address, error) {
	panic("implement me")
}

func (f *MockFullNode) StateListMiners(ctx context.Context, tsk types.TipSetKey) ([]address.Address, error) {
	panic("implement me")
}

func (f *MockFullNode) StateListActors(ctx context.Context, tsk types.TipSetKey) ([]address.Address, error) {
	panic("implement me")
}

func (f *MockFullNode) StateMinerPower(ctx context.Context, addr address.Address, tsk types.TipSetKey) (*types.MinerPower, error) {
	panic("implement me")
}

func (f *MockFullNode) StateMinerAvailableBalance(ctx context.Context, maddr address.Address, tsk types.TipSetKey) (big.Int, error) {
	panic("implement me")
}

func (f *MockFullNode) StateSectorExpiration(ctx context.Context, maddr address.Address, sectorNumber abi.SectorNumber, tsk types.TipSetKey) (*lminer.SectorExpiration, error) {
	panic("implement me")
}

func (f *MockFullNode) StateMinerSectorCount(ctx context.Context, addr address.Address, tsk types.TipSetKey) (types.MinerSectors, error) {
	panic("implement me")
}

func (f *MockFullNode) StateMarketBalance(ctx context.Context, addr address.Address, tsk types.TipSetKey) (types.MarketBalance, error) {
	panic("implement me")
}

func (f *MockFullNode) StateDealProviderCollateralBounds(ctx context.Context, size abi.PaddedPieceSize, verified bool, tsk types.TipSetKey) (types.DealCollateralBounds, error) {
	panic("implement me")
}

func (f *MockFullNode) StateVerifiedClientStatus(ctx context.Context, addr address.Address, tsk types.TipSetKey) (*abi.StoragePower, error) {
	panic("implement me")
}

func (f *MockFullNode) BlockTime(ctx context.Context) time.Duration {
	panic("implement me")
}

func (f *MockFullNode) ChainList(ctx context.Context, tsKey types.TipSetKey, count int) ([]types.TipSetKey, error) {
	panic("implement me")
}

func (f *MockFullNode) ChainSetHead(ctx context.Context, key types.TipSetKey) error {
	panic("implement me")
}

func (f *MockFullNode) ChainGetTipSetByHeight(ctx context.Context, height abi.ChainEpoch, tsk types.TipSetKey) (*types.TipSet, error) {
	panic("implement me")
}

func (f *MockFullNode) ChainGetTipSetAfterHeight(ctx context.Context, height abi.ChainEpoch, tsk types.TipSetKey) (*types.TipSet, error) {
	panic("implement me")
}

func (f *MockFullNode) StateGetRandomnessFromTickets(ctx context.Context, personalization crypto.DomainSeparationTag, randEpoch abi.ChainEpoch, entropy []byte, tsk types.TipSetKey) (abi.Randomness, error) {
	panic("implement me")
}

func (f *MockFullNode) StateGetRandomnessFromBeacon(ctx context.Context, personalization crypto.DomainSeparationTag, randEpoch abi.ChainEpoch, entropy []byte, tsk types.TipSetKey) (abi.Randomness, error) {
	panic("implement me")
}

func (f *MockFullNode) StateGetBeaconEntry(ctx context.Context, epoch abi.ChainEpoch) (*types.BeaconEntry, error) {
	panic("implement me")
}

func (f *MockFullNode) ChainGetBlock(ctx context.Context, id cid.Cid) (*types.BlockHeader, error) {
	panic("implement me")
}

func (f *MockFullNode) ChainGetMessage(ctx context.Context, msgID cid.Cid) (*types.Message, error) {
	panic("implement me")
}

func (f *MockFullNode) ChainGetBlockMessages(ctx context.Context, bid cid.Cid) (*types.BlockMessages, error) {
	panic("implement me")
}

func (f *MockFullNode) ChainGetReceipts(ctx context.Context, id cid.Cid) ([]types.MessageReceipt, error) {
	panic("implement me")
}

func (f *MockFullNode) StateVerifiedRegistryRootKey(ctx context.Context, tsk types.TipSetKey) (address.Address, error) {
	panic("implement me")
}

func (f *MockFullNode) StateVerifierStatus(ctx context.Context, addr address.Address, tsk types.TipSetKey) (*abi.StoragePower, error) {
	panic("implement me")
}

func (f *MockFullNode) GetFullBlock(ctx context.Context, id cid.Cid) (*types.FullBlock, error) {
	panic("implement me")
}

func (f *MockFullNode) GetActor(ctx context.Context, addr address.Address) (*types.Actor, error) {
	panic("implement me")
}

func (f *MockFullNode) GetParentStateRootActor(ctx context.Context, ts *types.TipSet, addr address.Address) (*types.Actor, error) {
	panic("implement me")
}

func (f *MockFullNode) GetEntry(ctx context.Context, height abi.ChainEpoch, round uint64) (*types.BeaconEntry, error) {
	panic("implement me")
}

func (f *MockFullNode) MessageWait(ctx context.Context, msgCid cid.Cid, confidence, lookback abi.ChainEpoch) (*types.ChainMessage, error) {
	panic("implement me")
}

func (f *MockFullNode) ProtocolParameters(ctx context.Context) (*types.ProtocolParams, error) {
	panic("implement me")
}

func (f *MockFullNode) ResolveToKeyAddr(ctx context.Context, addr address.Address, ts *types.TipSet) (address.Address, error) {
	panic("implement me")
}

func (f *MockFullNode) StateWaitMsg(ctx context.Context, cid cid.Cid, confidence uint64, limit abi.ChainEpoch, allowReplaced bool) (*types.MsgLookup, error) {
	panic("implement me")
}

func (f *MockFullNode) StateNetworkVersion(ctx context.Context, tsk types.TipSetKey) (network.Version, error) {
	panic("implement me")
}

func (f *MockFullNode) VerifyEntry(parent, child *types.BeaconEntry, height abi.ChainEpoch) bool {
	panic("implement me")
}

func (f *MockFullNode) ChainExport(ctx context.Context, epoch abi.ChainEpoch, b bool, key types.TipSetKey) (<-chan []byte, error) {
	panic("implement me")
}

func (f *MockFullNode) ChainGetPath(ctx context.Context, from types.TipSetKey, to types.TipSetKey) ([]*types.HeadChange, error) {
	panic("implement me")
}

func (f *MockFullNode) StateGetNetworkParams(ctx context.Context) (*types.NetworkParams, error) {
	panic("implement me")
}

func (f *MockFullNode) StateActorCodeCIDs(ctx context.Context, version network.Version) (map[string]cid.Cid, error) {
	panic("implement me")
}

func (f *MockFullNode) StateMarketParticipants(ctx context.Context, tsk types.TipSetKey) (map[string]types.MarketBalance, error) {
	panic("implement me")
}

func (f *MockFullNode) MinerGetBaseInfo(ctx context.Context, maddr address.Address, round abi.ChainEpoch, tsk types.TipSetKey) (*types.MiningBaseInfo, error) {
	panic("implement me")
}

func (f *MockFullNode) MinerCreateBlock(ctx context.Context, bt *types.BlockTemplate) (*types.BlockMsg, error) {
	panic("implement me")
}

func (f *MockFullNode) MpoolDeleteByAdress(ctx context.Context, addr address.Address) error {
	panic("implement me")
}

func (f *MockFullNode) MpoolPublishByAddr(ctx context.Context, address address.Address) error {
	panic("implement me")
}

func (f *MockFullNode) MpoolPublishMessage(ctx context.Context, smsg *types.SignedMessage) error {
	panic("implement me")
}

func (f *MockFullNode) MpoolPush(ctx context.Context, smsg *types.SignedMessage) (cid.Cid, error) {
	panic("implement me")
}

func (f *MockFullNode) MpoolGetConfig(ctx context.Context) (*types.MpoolConfig, error) {
	panic("implement me")
}

func (f *MockFullNode) MpoolSetConfig(ctx context.Context, cfg *types.MpoolConfig) error {
	panic("implement me")
}

func (f *MockFullNode) MpoolSelect(ctx context.Context, key types.TipSetKey, f2 float64) ([]*types.SignedMessage, error) {
	panic("implement me")
}

func (f *MockFullNode) MpoolSelects(ctx context.Context, key types.TipSetKey, float64s []float64) ([][]*types.SignedMessage, error) {
	panic("implement me")
}

func (f *MockFullNode) MpoolPending(ctx context.Context, tsk types.TipSetKey) ([]*types.SignedMessage, error) {
	panic("implement me")
}

func (f *MockFullNode) MpoolClear(ctx context.Context, local bool) error {
	panic("implement me")
}

func (f *MockFullNode) MpoolPushUntrusted(ctx context.Context, smsg *types.SignedMessage) (cid.Cid, error) {
	panic("implement me")
}

func (f *MockFullNode) MpoolPushMessage(ctx context.Context, msg *types.Message, spec *types.MessageSendSpec) (*types.SignedMessage, error) {
	panic("implement me")
}

func (f *MockFullNode) MpoolBatchPushUntrusted(ctx context.Context, smsgs []*types.SignedMessage) ([]cid.Cid, error) {
	panic("implement me")
}

func (f *MockFullNode) MpoolBatchPushMessage(ctx context.Context, msgs []*types.Message, spec *types.MessageSendSpec) ([]*types.SignedMessage, error) {
	panic("implement me")
}

func (f *MockFullNode) MpoolGetNonce(ctx context.Context, addr address.Address) (uint64, error) {
	panic("implement me")
}

func (f *MockFullNode) MpoolSub(ctx context.Context) (<-chan types.MpoolUpdate, error) {
	panic("implement me")
}

func (f *MockFullNode) GasEstimateFeeCap(ctx context.Context, msg *types.Message, maxqueueblks int64, tsk types.TipSetKey) (big.Int, error) {
	panic("implement me")
}

func (f *MockFullNode) GasEstimateGasPremium(ctx context.Context, nblocksincl uint64, sender address.Address, gaslimit int64, tsk types.TipSetKey) (big.Int, error) {
	panic("implement me")
}

func (f *MockFullNode) GasEstimateGasLimit(ctx context.Context, msgIn *types.Message, tsk types.TipSetKey) (int64, error) {
	panic("implement me")
}

func (f *MockFullNode) MpoolCheckMessages(ctx context.Context, protos []*types.MessagePrototype) ([][]types.MessageCheckStatus, error) {
	panic("implement me")
}

func (f *MockFullNode) MpoolCheckPendingMessages(ctx context.Context, addr address.Address) ([][]types.MessageCheckStatus, error) {
	panic("implement me")
}

func (f *MockFullNode) MpoolCheckReplaceMessages(ctx context.Context, msg []*types.Message) ([][]types.MessageCheckStatus, error) {
	panic("implement me")
}

func (f *MockFullNode) MsigCreate(ctx context.Context, req uint64, addrs []address.Address, duration abi.ChainEpoch, val types.BigInt, src address.Address, gp types.BigInt) (*types.MessagePrototype, error) {
	panic("implement me")
}

func (f *MockFullNode) MsigPropose(ctx context.Context, msig address.Address, to address.Address, amt types.BigInt, src address.Address, method uint64, params []byte) (*types.MessagePrototype, error) {
	panic("implement me")
}

func (f *MockFullNode) MsigAddPropose(ctx context.Context, msig address.Address, src address.Address, newAdd address.Address, inc bool) (*types.MessagePrototype, error) {
	panic("implement me")
}

func (f *MockFullNode) MsigAddApprove(ctx context.Context, msig address.Address, src address.Address, txID uint64, proposer address.Address, newAdd address.Address, inc bool) (*types.MessagePrototype, error) {
	panic("implement me")
}

func (f *MockFullNode) MsigAddCancel(ctx context.Context, msig address.Address, src address.Address, txID uint64, newAdd address.Address, inc bool) (*types.MessagePrototype, error) {
	panic("implement me")
}

func (f *MockFullNode) MsigCancelTxnHash(ctx context.Context, address address.Address, u uint64, address2 address.Address, bigInt types.BigInt, address3 address.Address, u2 uint64, bytes []byte) (*types.MessagePrototype, error) {
	panic("implement me")
}

func (f *MockFullNode) MsigSwapPropose(ctx context.Context, msig address.Address, src address.Address, oldAdd address.Address, newAdd address.Address) (*types.MessagePrototype, error) {
	panic("implement me")
}

func (f *MockFullNode) MsigSwapApprove(ctx context.Context, msig address.Address, src address.Address, txID uint64, proposer address.Address, oldAdd address.Address, newAdd address.Address) (*types.MessagePrototype, error) {
	panic("implement me")
}

func (f *MockFullNode) MsigSwapCancel(ctx context.Context, msig address.Address, src address.Address, txID uint64, oldAdd address.Address, newAdd address.Address) (*types.MessagePrototype, error) {
	panic("implement me")
}

func (f *MockFullNode) MsigApprove(ctx context.Context, msig address.Address, txID uint64, src address.Address) (*types.MessagePrototype, error) {
	panic("implement me")
}

func (f *MockFullNode) MsigApproveTxnHash(ctx context.Context, msig address.Address, txID uint64, proposer address.Address, to address.Address, amt types.BigInt, src address.Address, method uint64, params []byte) (*types.MessagePrototype, error) {
	panic("implement me")
}

func (f *MockFullNode) MsigCancel(ctx context.Context, msig address.Address, txID uint64, src address.Address) (*types.MessagePrototype, error) {
	panic("implement me")
}

func (f *MockFullNode) MsigRemoveSigner(ctx context.Context, msig address.Address, proposer address.Address, toRemove address.Address, decrease bool) (*types.MessagePrototype, error) {
	panic("implement me")
}

func (f *MockFullNode) MsigGetVested(ctx context.Context, addr address.Address, start types.TipSetKey, end types.TipSetKey) (types.BigInt, error) {
	panic("implement me")
}

func (f *MockFullNode) NetFindProvidersAsync(ctx context.Context, key cid.Cid, count int) <-chan peer.AddrInfo {
	panic("implement me")
}

func (f *MockFullNode) NetGetClosestPeers(ctx context.Context, key string) ([]peer.ID, error) {
	panic("implement me")
}

func (f *MockFullNode) NetConnectedness(ctx context.Context, id peer.ID) (network2.Connectedness, error) {
	panic("implement me")
}

func (f *MockFullNode) NetFindPeer(ctx context.Context, p peer.ID) (peer.AddrInfo, error) {
	panic("implement me")
}

func (f *MockFullNode) NetConnect(ctx context.Context, pi peer.AddrInfo) error {
	panic("implement me")
}

func (f *MockFullNode) NetPeers(ctx context.Context) ([]peer.AddrInfo, error) {
	panic("implement me")
}

func (f *MockFullNode) NetPeerInfo(ctx context.Context, p peer.ID) (*types.ExtendedPeerInfo, error) {
	panic("implement me")
}

func (f *MockFullNode) NetAgentVersion(ctx context.Context, p peer.ID) (string, error) {
	panic("implement me")
}

func (f *MockFullNode) NetPing(ctx context.Context, p peer.ID) (time.Duration, error) {
	panic("implement me")
}

func (f *MockFullNode) NetAddrsListen(ctx context.Context) (peer.AddrInfo, error) {
	panic("implement me")
}

func (f *MockFullNode) NetDisconnect(ctx context.Context, p peer.ID) error {
	panic("implement me")
}

func (f *MockFullNode) NetAutoNatStatus(ctx context.Context) (types.NatInfo, error) {
	panic("implement me")
}

func (f *MockFullNode) ID(ctx context.Context) (peer.ID, error) {
	panic("implement me")
}

func (f *MockFullNode) NetBandwidthStats(ctx context.Context) (metrics.Stats, error) {
	panic("implement me")
}

func (f *MockFullNode) NetBandwidthStatsByPeer(ctx context.Context) (map[string]metrics.Stats, error) {
	panic("implement me")
}

func (f *MockFullNode) NetBandwidthStatsByProtocol(ctx context.Context) (map[protocol.ID]metrics.Stats, error) {
	panic("implement me")
}

func (f *MockFullNode) NetProtectAdd(ctx context.Context, acl []peer.ID) error {
	panic("implement me")
}

func (f *MockFullNode) NetProtectRemove(ctx context.Context, acl []peer.ID) error {
	panic("implement me")
}

func (f *MockFullNode) NetProtectList(ctx context.Context) ([]peer.ID, error) {
	panic("implement me")
}

func (f *MockFullNode) PaychGet(ctx context.Context, from, to address.Address, amt types.BigInt, opts types.PaychGetOpts) (*types.ChannelInfo, error) {
	panic("implement me")
}

func (f *MockFullNode) PaychFund(ctx context.Context, from, to address.Address, amt types.BigInt) (*types.ChannelInfo, error) {
	panic("implement me")
}

func (f *MockFullNode) PaychAvailableFunds(ctx context.Context, ch address.Address) (*types.ChannelAvailableFunds, error) {
	panic("implement me")
}

func (f *MockFullNode) PaychAvailableFundsByFromTo(ctx context.Context, from, to address.Address) (*types.ChannelAvailableFunds, error) {
	panic("implement me")
}

func (f *MockFullNode) PaychGetWaitReady(ctx context.Context, sentinel cid.Cid) (address.Address, error) {
	panic("implement me")
}

func (f *MockFullNode) PaychAllocateLane(ctx context.Context, ch address.Address) (uint64, error) {
	panic("implement me")
}

func (f *MockFullNode) PaychNewPayment(ctx context.Context, from, to address.Address, vouchers []types.VoucherSpec) (*types.PaymentInfo, error) {
	panic("implement me")
}

func (f *MockFullNode) PaychList(ctx context.Context) ([]address.Address, error) {
	panic("implement me")
}

func (f *MockFullNode) PaychStatus(ctx context.Context, pch address.Address) (*types.Status, error) {
	panic("implement me")
}

func (f *MockFullNode) PaychSettle(ctx context.Context, addr address.Address) (cid.Cid, error) {
	panic("implement me")
}

func (f *MockFullNode) PaychCollect(ctx context.Context, addr address.Address) (cid.Cid, error) {
	panic("implement me")
}

func (f *MockFullNode) PaychVoucherCheckValid(ctx context.Context, ch address.Address, sv *types.SignedVoucher) error {
	panic("implement me")
}

func (f *MockFullNode) PaychVoucherCheckSpendable(ctx context.Context, ch address.Address, sv *types.SignedVoucher, secret []byte, proof []byte) (bool, error) {
	panic("implement me")
}

func (f *MockFullNode) PaychVoucherAdd(ctx context.Context, ch address.Address, sv *types.SignedVoucher, proof []byte, minDelta big.Int) (big.Int, error) {
	panic("implement me")
}

func (f *MockFullNode) PaychVoucherCreate(ctx context.Context, pch address.Address, amt big.Int, lane uint64) (*types.VoucherCreateResult, error) {
	panic("implement me")
}

func (f *MockFullNode) PaychVoucherList(ctx context.Context, pch address.Address) ([]*types.SignedVoucher, error) {
	panic("implement me")
}

func (f *MockFullNode) PaychVoucherSubmit(ctx context.Context, ch address.Address, sv *types.SignedVoucher, secret []byte, proof []byte) (cid.Cid, error) {
	panic("implement me")
}

func (f *MockFullNode) ChainSyncHandleNewTipSet(ctx context.Context, ci *types.ChainInfo) error {
	panic("implement me")
}

func (f *MockFullNode) SetConcurrent(ctx context.Context, concurrent int64) error {
	panic("implement me")
}

func (f *MockFullNode) SyncerTracker(ctx context.Context) *types.TargetTracker {
	panic("implement me")
}

func (f *MockFullNode) Concurrent(ctx context.Context) int64 {
	panic("implement me")
}

func (f *MockFullNode) ChainTipSetWeight(ctx context.Context, tsk types.TipSetKey) (big.Int, error) {
	panic("implement me")
}

func (f *MockFullNode) SyncSubmitBlock(ctx context.Context, blk *types.BlockMsg) error {
	panic("implement me")
}

func (f *MockFullNode) StateCall(ctx context.Context, msg *types.Message, tsk types.TipSetKey) (*types.InvocResult, error) {
	panic("implement me")
}

func (f *MockFullNode) SyncState(ctx context.Context) (*types.SyncState, error) {
	panic("implement me")
}

func (f *MockFullNode) WalletSign(ctx context.Context, k address.Address, msg []byte, meta types.MsgMeta) (*crypto.Signature, error) {
	panic("implement me")
}

func (f *MockFullNode) WalletExport(ctx context.Context, addr address.Address, password string) (*types.KeyInfo, error) {
	panic("implement me")
}

func (f *MockFullNode) WalletImport(ctx context.Context, key *types.KeyInfo) (address.Address, error) {
	panic("implement me")
}

func (f *MockFullNode) WalletHas(ctx context.Context, addr address.Address) (bool, error) {
	panic("implement me")
}

func (f *MockFullNode) WalletNewAddress(ctx context.Context, protocol address.Protocol) (address.Address, error) {
	panic("implement me")
}

func (f *MockFullNode) WalletBalance(ctx context.Context, addr address.Address) (abi.TokenAmount, error) {
	panic("implement me")
}

func (f *MockFullNode) WalletDefaultAddress(ctx context.Context) (address.Address, error) {
	panic("implement me")
}

func (f *MockFullNode) WalletAddresses(ctx context.Context) []address.Address {
	panic("implement me")
}

func (f *MockFullNode) WalletSetDefault(ctx context.Context, addr address.Address) error {
	panic("implement me")
}

func (f *MockFullNode) WalletSignMessage(ctx context.Context, k address.Address, msg *types.Message) (*types.SignedMessage, error) {
	panic("implement me")
}

func (f *MockFullNode) LockWallet(ctx context.Context) error {
	panic("implement me")
}

func (f *MockFullNode) UnLockWallet(ctx context.Context, password []byte) error {
	panic("implement me")
}

func (f *MockFullNode) SetPassword(ctx context.Context, password []byte) error {
	panic("implement me")
}

func (f *MockFullNode) HasPassword(ctx context.Context) bool {
	panic("implement me")
}

func (f *MockFullNode) WalletState(ctx context.Context) int {
	panic("implement me")
}

func (f *MockFullNode) Version(ctx context.Context) (types.Version, error) {
	panic("implement me")
}

var _ v1.FullNode = (*MockFullNode)(nil)
