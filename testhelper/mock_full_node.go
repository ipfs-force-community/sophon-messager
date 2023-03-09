package testhelper

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/filecoin-project/go-state-types/network"

	"github.com/asaskevich/EventBus"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	v1 "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	"go.uber.org/atomic"

	mockV1 "github.com/filecoin-project/venus/venus-shared/api/chain/v1/mock"
	"github.com/filecoin-project/venus/venus-shared/types"
	"github.com/ipfs/go-cid"
)

const (
	maxMsgInBlock   = 200
	headChangeTopic = "head_change"
)

var ErrGasLimitNegative = errors.New("gas limit is negative")

var (
	DefGasUsed           = int64(10000)
	DefGasPremium        = abi.NewTokenAmount(1000)
	DefGasFeeCap         = abi.NewTokenAmount(10000)
	DefBaseFee           = abi.NewTokenAmount(1000)
	DefGasOverEstimation = 1.25
	DefGasOverPremium    = 4.0
	DefMaxFee            = big.Mul(big.NewInt(DefGasUsed*10), DefGasFeeCap)

	defBalance = abi.NewTokenAmount(1000)

	// MinPackedPremium If the gas premium is lower than this value, the message will not be packaged
	MinPackedPremium = abi.NewTokenAmount(500)
)

type MockFullNode struct {
	ctx context.Context

	miner address.Address

	actors map[address.Address]*types.Actor

	ts        map[types.TipSetKey]*types.TipSet
	heightKey map[abi.ChainEpoch]types.TipSetKey
	currTS    *types.TipSet

	blockDelay  time.Duration
	blockInfos  map[cid.Cid]*blockInfo
	chainMsgs   map[cid.Cid]*types.SignedMessage
	msgReceipts map[cid.Cid]*types.MessageReceipt

	pendingMsgs []*types.SignedMessage

	eventBus EventBus.Bus

	revertSignReceiver chan *RevertSignal

	l sync.Mutex

	mockV1.MockFullNode
}

type blockInfo struct {
	bh   *types.BlockHeader
	msgs []cid.Cid
}

func NewMockFullNode(ctx context.Context, blockDelay time.Duration) (*MockFullNode, error) {
	miner, err := address.NewIDAddress(10001)
	if err != nil {
		return nil, err
	}
	node := &MockFullNode{
		ctx:                ctx,
		blockDelay:         blockDelay,
		miner:              miner,
		actors:             make(map[address.Address]*types.Actor),
		ts:                 make(map[types.TipSetKey]*types.TipSet),
		heightKey:          make(map[abi.ChainEpoch]types.TipSetKey),
		blockInfos:         make(map[cid.Cid]*blockInfo),
		chainMsgs:          make(map[cid.Cid]*types.SignedMessage),
		msgReceipts:        make(map[cid.Cid]*types.MessageReceipt),
		eventBus:           EventBus.New(),
		revertSignReceiver: make(chan *RevertSignal, 5),
	}
	bh, err := GenBlockHead(miner, 0, []cid.Cid{})
	if err != nil {
		return nil, err
	}
	ts, err := types.NewTipSet([]*types.BlockHeader{bh})
	if err != nil {
		return nil, err
	}
	node.setHead(ts)
	node.blockInfos[bh.Cid()] = &blockInfo{bh: bh}
	node.pub(headChangeTopic, []*types.HeadChange{
		{
			Type: types.HCApply,
			Val:  ts,
		},
	})

	go node.tipsetProvider()

	return node, nil
}

func checkErr(err error) {
	if err != nil {
		panic(fmt.Errorf("%s %s", string(debug.Stack()), err))
	}
}

func (f *MockFullNode) AddActors(addrs []address.Address) error {
	f.l.Lock()
	defer f.l.Unlock()

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

type RevertSignal struct {
	ExpectRevertCount int
	RevertedTS        chan []*types.TipSet
}

func (f *MockFullNode) SendRevertSignal(rs *RevertSignal) {
	select {
	case f.revertSignReceiver <- rs:
	default:
		fmt.Println("receive too many revert channel")
		close(rs.RevertedTS)
	}
}

func (f *MockFullNode) tipsetProvider() {
	ticker := time.NewTicker(f.blockDelay)
	defer ticker.Stop()

	genTS := func() *types.TipSet {
		bh, err := f.blockProvider()
		if err != nil {
			panic(err)
		}
		ts, err := types.NewTipSet([]*types.BlockHeader{bh})
		if err != nil {
			panic(err)
		}
		return ts
	}

	for {
		var ts *types.TipSet
		var revertTS []*types.TipSet
		select {
		case <-ticker.C:
			select {
			case rs := <-f.revertSignReceiver:
				revertTS = f.revertTS(rs.ExpectRevertCount)
				rs.RevertedTS <- revertTS
			default:
			}
			ts = genTS()
		case <-f.ctx.Done():
			return
		}
		f.setHead(ts)
		var headChanges []*types.HeadChange
		headChanges = append(headChanges, &types.HeadChange{
			Type: types.HCApply,
			Val:  ts,
		})
		for _, t := range revertTS {
			headChanges = append(headChanges, &types.HeadChange{
				Type: types.HCRevert,
				Val:  t,
			})
			headChanges = append(headChanges, &types.HeadChange{
				Type: types.HCApply,
				Val:  f.ts[f.heightKey[t.Height()]],
			})
		}
		f.pub(headChangeTopic, headChanges)
	}
}

func (f *MockFullNode) revertTS(revertCount int) []*types.TipSet {
	f.l.Lock()
	defer f.l.Unlock()

	msgs := make([]cid.Cid, 0)
	currHeight := f.currTS.Height()
	revertTS := make([]*types.TipSet, 0, revertCount)

	for i := currHeight; i > 0 && currHeight-i < abi.ChainEpoch(revertCount); i-- {
		key := f.heightKey[i]
		ts := f.ts[key]
		if ts.Height() == 0 {
			break
		}
		for _, blk := range ts.Cids() {
			blkInfo, ok := f.blockInfos[blk]
			if !ok {
				continue
			}
			for i := len(blkInfo.msgs) - 1; i >= 0; i-- {
				msgs = append(msgs, blkInfo.msgs[i])
			}
		}
		revertTS = append(revertTS, ts)
	}

	revertTSLen := len(revertTS)
	msgsLen := len(msgs)
	cidList := make([][]cid.Cid, revertTSLen)
	if revertTSLen == 0 {
		return revertTS
	}
	if msgsLen > 0 {
		// Redistribute messages
		batchSize := msgsLen / revertTSLen
		if msgsLen%revertTSLen != 0 {
			batchSize++
		}

		j := revertTSLen - 1
		for i := 1; i <= revertTSLen; i++ {
			start := (i - 1) * batchSize
			end := i * batchSize
			if end > msgsLen {
				end = msgsLen
			}
			cidList[j] = msgs[start:end]
			j--
		}
	}

	j := 0
	parentKey := revertTS[revertTSLen-1].Parents()
	for i := int(currHeight) - revertTSLen + 1; i <= int(currHeight); i++ {
		bh, err := GenBlockHead(f.miner, abi.ChainEpoch(i), parentKey.Cids())
		checkErr(err)
		f.blockInfos[bh.Cid()] = &blockInfo{
			bh:   bh,
			msgs: cidList[j],
		}
		ts, err := types.NewTipSet([]*types.BlockHeader{bh})
		checkErr(err)
		f.ts[ts.Key()] = ts
		f.heightKey[ts.Height()] = ts.Key()
		parentKey = ts.Key()
		if i == int(currHeight) {
			f.currTS = ts
		}
		j++
	}

	return revertTS
}

func (f *MockFullNode) setHead(ts *types.TipSet) {
	f.l.Lock()
	defer f.l.Unlock()
	f.currTS = ts
	f.ts[ts.Key()] = ts
	f.heightKey[ts.Height()] = ts.Key()
}

func (f *MockFullNode) pub(topic string, arg interface{}) {
	f.l.Lock()
	defer f.l.Unlock()

	f.eventBus.Publish(topic, arg)
}

func (f *MockFullNode) sub(topic string, fn interface{}) {
	f.l.Lock()
	defer f.l.Unlock()

	_ = f.eventBus.Subscribe(topic, fn)
}

func (f *MockFullNode) blockProvider() (*types.BlockHeader, error) {
	head, err := f.ChainHead(context.Background())
	if err != nil {
		return nil, err
	}
	bh, err := GenBlockHead(f.miner, head.Height()+1, head.Cids())
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

func (f *MockFullNode) StateNetworkVersion(arg0 context.Context, arg1 types.TipSetKey) (network.Version, error) {
	return network.Version17, nil
}
func (f *MockFullNode) StateGetNetworkParams(ctx context.Context) (*types.NetworkParams, error) {
	return &types.NetworkParams{
		NetworkName:    types.NetworkNameMain,
		BlockDelaySecs: uint64(f.blockDelay / time.Second),
	}, nil
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

	if key.IsEmpty() {
		return f.currTS, nil
	}
	ts, ok := f.ts[key]
	if !ok {
		return nil, fmt.Errorf("not found %s", key)
	}
	return ts, nil
}

func (f *MockFullNode) ChainList(ctx context.Context, tsKey types.TipSetKey, count int) ([]types.TipSetKey, error) {
	ts, err := f.ChainGetTipSet(ctx, tsKey)
	if err != nil {
		return nil, err
	}

	keys := make([]types.TipSetKey, 0, count)
	for i := 0; i < count; i++ {
		keys = append(keys, ts.Key())

		if ts.Height() == 0 {
			return keys, nil
		}

		ts, err = f.ChainGetTipSet(ctx, ts.Parents())
		if err != nil {
			return nil, err
		}
	}

	return keys, nil
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

	if addr.Protocol() == address.ID {
		var err error
		addr, err = ResolveIDAddr(addr)
		if err != nil {
			return nil, err
		}
	}
	actor, ok := f.actors[addr]
	if !ok {
		return nil, fmt.Errorf("not found actor %v", addr)
	}
	actorCp := *actor
	return &actorCp, nil
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

	maxFee := DefMaxFee
	if spec != nil && !spec.MaxFee.NilOrZero() {
		maxFee = spec.MaxFee
	}
	gl := types.NewInt(uint64(msg.GasLimit))
	totalFee := types.BigMul(msg.GasFeeCap, gl)

	if !totalFee.LessThanEqual(maxFee) {
		msg.GasFeeCap = big.Div(maxFee, gl)
		msg.GasPremium = big.Min(msg.GasFeeCap, msg.GasPremium)
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

func (f *MockFullNode) MpoolBatchPushUntrusted(ctx context.Context, smsgs []*types.SignedMessage) ([]cid.Cid, error) {
	f.l.Lock()
	defer f.l.Unlock()
	cids := make([]cid.Cid, 0, len(smsgs))
	for _, msg := range smsgs {
		if _, ok := f.chainMsgs[msg.Cid()]; ok {
			continue
		}
		// todo: check nonce
		for i, m := range f.pendingMsgs {
			if m.Message.From == msg.Message.From && m.Message.Nonce == msg.Message.Nonce {
				f.pendingMsgs[i] = msg
				cids = append(cids, m.Cid())
				continue
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

	for h, tsk := range f.heightKey {
		for _, blkCID := range tsk.Cids() {
			for _, c := range f.blockInfos[blkCID].msgs {
				if c == msgCid {
					msgLookup.Height = h
					msgLookup.TipSet = tsk

					return msgLookup, nil
				}
			}
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
	done := atomic.NewBool(false)
	f.sub(headChangeTopic, func(hc []*types.HeadChange) {
		// to test UpdateAllFilledMessage and testUpdateFilledMessageByID
		time.Sleep(f.blockDelay / 4)
		if !done.Load() {
			out <- hc
		}
	})
	go func() {
		<-ctx.Done()
		done.Store(true)
		close(out)
	}()

	return out, nil
}

var _ v1.FullNode = (*MockFullNode)(nil)
