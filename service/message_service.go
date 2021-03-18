package service

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/venus/pkg/messagepool"
	venusTypes "github.com/filecoin-project/venus/pkg/types"
	"github.com/ipfs-force-community/venus-messager/utils"
	"github.com/ipfs-force-community/venus-wallet/core"
	"github.com/ipfs/go-cid"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"

	"github.com/ipfs-force-community/venus-messager/config"
	"github.com/ipfs-force-community/venus-messager/models/repo"
	"github.com/ipfs-force-community/venus-messager/types"
)

const (
	MaxHeadChangeProcess = 5

	LookBackLimit = 900

	maxStoreTipsetCount = 900
)

type MessageService struct {
	repo           repo.Repo
	log            *logrus.Logger
	cfg            *config.MessageServiceConfig
	nodeClient     *NodeClient
	messageState   *MessageState
	addressService *AddressService

	triggerPush chan *venusTypes.TipSet
	headChans   chan *headChan
	failedHeads []failedHead

	readFileOnce sync.Once
	tsCache      *TipsetCache
}

type headChan struct {
	apply, revert []*venusTypes.TipSet
}

type failedHead struct {
	headChan
	time.Time
}

type TipsetCache struct {
	Cache      map[int64]*tipsetFormat
	CurrHeight int64

	l sync.Mutex
}

func NewMessageService(repo repo.Repo,
	nc *NodeClient,
	logger *logrus.Logger,
	cfg *config.MessageServiceConfig,
	messageState *MessageState,
	addressService *AddressService) (*MessageService, error) {
	ms := &MessageService{
		repo:       repo,
		log:        logger,
		nodeClient: nc,
		cfg:        cfg,
		headChans:  make(chan *headChan, MaxHeadChangeProcess),

		messageState:   messageState,
		addressService: addressService,
		tsCache: &TipsetCache{
			Cache:      make(map[int64]*tipsetFormat, maxStoreTipsetCount),
			CurrHeight: 0,
		},
		triggerPush: make(chan *venusTypes.TipSet, 20),
		failedHeads: make([]failedHead, 0),
	}
	ms.refreshMessageState(context.TODO())

	return ms, nil
}

func (ms *MessageService) PushMessage(ctx context.Context, msg *types.Message) (types.UUID, error) {
	//replace address
	if msg.From.Protocol() == address.ID {
		fromA, err := ms.nodeClient.ResolveToKeyAddr(ctx, msg.From, nil)
		if err != nil {
			return types.UUID{}, xerrors.Errorf("getting key address: %w", err)
		}
		ms.log.Warnf("Push from ID address (%s), adjusting to %s", msg.From, fromA)
		msg.From = fromA
	}

	has, err := ms.repo.AddressRepo().HasAddress(ctx, msg.From)
	if err != nil {
		return types.UUID{}, err
	}
	if !has {
		return types.UUID{}, xerrors.Errorf("address %s not in wallet", msg.From)
	}
	msg.State = types.UnFillMsg
	msg.Nonce = 0
	id, err := ms.repo.MessageRepo().SaveMessage(msg)
	if err == nil {
		ms.messageState.SetMessage(msg.ID, msg)
	}

	return id, err
}

func (ms *MessageService) GetMessage(ctx context.Context, uuid types.UUID) (*types.Message, error) {
	ts, err := ms.nodeClient.ChainHead(ctx)
	if err != nil {
		return nil, err
	}
	msg, err := ms.repo.MessageRepo().GetMessage(uuid)
	if err != nil {
		return nil, err
	}
	if msg.State == types.OnChainMsg {
		msg.Confidence = int64(ts.Height()) - msg.Height
	}
	return msg, nil
}

func (ms *MessageService) GetMessageState(ctx context.Context, uuid types.UUID) (types.MessageState, error) {
	return ms.repo.MessageRepo().GetMessageState(uuid)
}

func (ms *MessageService) GetMessageByCid(ctx context.Context, unsignedCid cid.Cid) (*types.Message, error) {
	return ms.repo.MessageRepo().GetMessageByCid(unsignedCid.String())
}

func (ms *MessageService) GetMessageBySignedCid(ctx context.Context, signedCid cid.Cid) (*types.Message, error) {
	return ms.repo.MessageRepo().GetMessageBySignedCid(signedCid.String())
}

func (ms *MessageService) GetMessageByUnsignedCid(ctx context.Context, unsignedCid cid.Cid) (*types.Message, error) {
	return ms.repo.MessageRepo().GetMessageByCid(unsignedCid.String())
}

func (ms *MessageService) GetMessageByFromAndNonce(ctx context.Context, from string, nonce uint64) (*types.Message, error) {
	return ms.repo.MessageRepo().GetMessageByFromAndNonce(from, nonce)
}

func (ms *MessageService) ListMessage(ctx context.Context) ([]*types.Message, error) {
	ts, err := ms.nodeClient.ChainHead(ctx)
	if err != nil {
		return nil, err
	}
	msgs, err := ms.repo.MessageRepo().ListMessage()
	if err != nil {
		return nil, err
	}

	for _, msg := range msgs {
		if msg.State == types.OnChainMsg {
			msg.Confidence = int64(ts.Height()) - msg.Height
		}
	}
	return msgs, nil
}

func (ms *MessageService) ListFilledMessageByAddress(ctx context.Context, addr address.Address) ([]*types.Message, error) {
	return ms.repo.MessageRepo().ListFilledMessageByAddress(addr)
}

func (ms *MessageService) UpdateMessageStateByCid(ctx context.Context, cid string, state types.MessageState) (string, error) {
	return ms.repo.MessageRepo().UpdateMessageStateByCid(cid, state)
}

func (ms *MessageService) UpdateMessageStateByID(ctx context.Context, id types.UUID, state types.MessageState) (types.UUID, error) {
	return ms.repo.MessageRepo().UpdateMessageStateByID(id, state)
}

func (ms *MessageService) UpdateMessageInfoByCid(unsignedCid string, receipt *venusTypes.MessageReceipt, height abi.ChainEpoch, state types.MessageState, tsKey string) (string, error) {
	return ms.repo.MessageRepo().UpdateMessageInfoByCid(unsignedCid, receipt, height, state, tsKey)
}

func (ms *MessageService) ProcessNewHead(ctx context.Context, apply, revert []*venusTypes.TipSet) error {
	ms.log.Infof("receive new head from chain")
	if !ms.cfg.IsProcessHead {
		ms.log.Infof("skip process new head")
		return nil
	}
	ms.headChans <- &headChan{
		apply:  apply,
		revert: revert,
	}
	ms.log.Infof("%d head wait to process", len(ms.headChans))
	return nil
}

func (ms *MessageService) ReconnectCheck(ctx context.Context, head *venusTypes.TipSet) error {
	ms.log.Infof("reconnect to node")

	ms.readFileOnce.Do(func() {
		tsCache, err := readTipsetFile(ms.cfg.TipsetFilePath)
		if err != nil {
			ms.log.Errorf("read tipset file failed %v", err)
		}
		ms.tsCache = tsCache
	})

	if len(ms.tsCache.Cache) == 0 {
		return nil
	}

	tsList := ms.tsCache.ListTs()
	sort.Sort(tsList)

	// long time not use
	if int64(head.Height())-tsList[0].Height >= LookBackLimit {
		count, err := ms.UpdateAllFilledMessage(ctx)
		if err != nil {
			return err
		}
		ms.log.Infof("gap height %v, update filled message count %v", int64(head.Height())-tsList[0].Height, count)
		return nil
	}

	if tsList[0].Height == int64(head.Height()) && tsList[0].Key == head.String() {
		ms.log.Infof("The head does not change and returns directly.")
		return nil
	}

	gapTipset, revertTipset, err := ms.lookAncestors(ctx, tsList, head)
	if err != nil {
		return err
	}

	ms.headChans <- &headChan{
		apply:  gapTipset,
		revert: revertTipset,
	}

	return nil
}

func (ms *MessageService) lookAncestors(ctx context.Context, localTipset tipsetList, head *venusTypes.TipSet) ([]*venusTypes.TipSet, []*venusTypes.TipSet, error) {
	var err error

	ts := &venusTypes.TipSet{}
	*ts = *head

	idx := 0
	localTsLen := len(localTipset)

	gapTipset := make([]*venusTypes.TipSet, 0)
	loopCount := 0
	for {
		if loopCount > LookBackLimit {
			break
		}
		if idx >= localTsLen {
			break
		}
		localTs := localTipset[idx]

		if ts.Height() == 0 {
			break
		}
		if localTs.Height > int64(ts.Height()) {
			idx++
		} else if localTs.Height == int64(ts.Height()) {
			if localTs.Key == ts.String() {
				break
			}
			idx++
		} else {
			gapTipset = append(gapTipset, ts)
			ts, err = ms.nodeClient.ChainGetTipSet(ctx, ts.Parents())
			if err != nil {
				return nil, nil, xerrors.Errorf("get tipset failed %v", err)
			}
		}
		loopCount++
	}

	var revertTsf []*tipsetFormat
	if idx >= localTsLen {
		idx = localTsLen
	}
	revertTsf = localTipset[:idx]

	revertTs, err := ms.convertTipsetFormatToTipset(revertTsf)

	return gapTipset, revertTs, err
}

func (ms *MessageService) convertTipsetFormatToTipset(tf []*tipsetFormat) ([]*venusTypes.TipSet, error) {
	var tsList []*venusTypes.TipSet
	var err error
	for _, t := range tf {
		key, err := utils.StringToTipsetKey(t.Key)
		if err != nil {
			return nil, err
		}
		blocks := make([]*venusTypes.BlockHeader, len(key.Cids()))
		for i, cid := range key.Cids() {
			blocks[i], err = ms.nodeClient.ChainGetBlock(context.TODO(), cid)
			if err != nil {
				return nil, err
			}
		}
		ts, err := venusTypes.NewTipSet(blocks...)
		if err != nil {
			return nil, err
		}
		tsList = append(tsList, ts)
	}

	return tsList, err
}

///   Message push    ////

func (ms *MessageService) pushMessageToPool(ctx context.Context, ts *venusTypes.TipSet) error {
	addrList, err := ms.addressService.ListAddress(ctx)
	if err != nil {
		return err
	}
	//sort by addr weight
	sort.Slice(addrList, func(i, j int) bool {
		return addrList[i].Weight < addrList[j].Weight
	})

	var toPushMessage []*venusTypes.SignedMessage
	for _, addr := range addrList {
		if err = ms.repo.Transaction(func(txRepo repo.TxRepo) error {
			addrInfo, exit := ms.addressService.GetAddressInfo(addr.Addr)
			if !exit {
				return xerrors.Errorf("no wallet cliet of address %s", addr.Addr)
			}

			mAddr, err := address.NewFromString(addr.Addr)
			if err != nil {
				return err
			}

			//判断是否需要推送消息
			actor, err := ms.nodeClient.StateGetActor(ctx, mAddr, ts.Key())
			if err != nil {
				ms.log.Warnf("actor of address %s not found", mAddr)
				return nil
			}

			if actor.Nonce > addr.Nonce {
				ms.log.Warnf("%s nonce in db %d is smaller than nonce on chain %d", addr.Addr, addr.Nonce, actor.Nonce)
				addr.Nonce = actor.Nonce
				if _, err := txRepo.AddressRepo().SaveAddress(ctx, addr); err != nil {
					ms.log.Errorf("save address %v", err)
				}
			}
			//todo push sigined but not onchain message, when to resend message
			filledMessage, err := txRepo.MessageRepo().ListFilledMessageByAddress(mAddr)
			if err != nil {
				ms.log.Errorf("found filled message %v", err)
			}
			for _, msg := range filledMessage {
				toPushMessage = append(toPushMessage, &venusTypes.SignedMessage{
					Message:   msg.UnsignedMessage,
					Signature: *msg.Signature,
				})
			}

			//sign new message
			nonceGap := addr.Nonce - actor.Nonce
			if nonceGap > 20 {
				ms.log.Debugf("%s there are %d message not to be package ", addr.Addr, nonceGap)
				return nil
			}
			selectCount := 20 - nonceGap
			//消息排序
			messages, err := txRepo.MessageRepo().ListUnChainMessageByAddress(mAddr)
			if err != nil {
				return err
			}
			messages, expireMsgs := ms.excludeExpire(ts, messages)
			//todo 如何筛选
			if len(messages) == 0 {
				ms.log.Debugf("%s have no message", addr.Addr)
				return nil
			}
			var selectMsg []*types.Message
			var count = uint64(0)
			for _, msg := range messages {
				//分配nonce
				msg.Nonce = addr.Nonce
				addr.Nonce++
				//todo 估算gas, spec怎么做？
				//通过配置影响 maxfee
				newMsg, err := ms.nodeClient.GasEstimateMessageGas(ctx, msg.VMMessage(), &venusTypes.MessageSendSpec{MaxFee: msg.Meta.MaxFee}, ts.Key())
				if err != nil {
					ms.log.Errorf("GasEstimateMessageGas msg id fail, id: %v, error: %v", msg.ID.String(), err)
					continue
				}
				msg.GasFeeCap = newMsg.GasFeeCap
				msg.GasPremium = newMsg.GasPremium
				msg.GasLimit = newMsg.GasLimit

				signedMsg, err := ToSignedMsg(ctx, addrInfo.WalletClient, msg)
				if err != nil {
					ms.log.Errorf("wallet sign failed %s fail %v", msg.ID.String(), err)
					continue
				}

				signedCid := signedMsg.Cid()
				msg.SignedCid = &signedCid

				if count < selectCount {
					selectMsg = append(selectMsg, msg)
					count++
				}
			}

			//保存消息
			err = txRepo.MessageRepo().ExpireMessage(expireMsgs)
			if err != nil {
				return err
			}

			err = txRepo.MessageRepo().BatchSaveMessage(selectMsg)
			if err != nil {
				return err
			}

			_, err = txRepo.AddressRepo().SaveAddress(ctx, addr)
			if err != nil {
				return err
			}

			for _, msg := range selectMsg {
				toPushMessage = append(toPushMessage, &venusTypes.SignedMessage{
					Message:   msg.UnsignedMessage,
					Signature: *msg.Signature,
				})
				//update cache
				err := ms.messageState.MutatorMessage(msg.ID, func(message *types.Message) error {
					message.SignedCid = msg.SignedCid
					message.UnsignedCid = msg.UnsignedCid
					message.UnsignedMessage = msg.UnsignedMessage
					message.State = msg.State
					message.Signature = msg.Signature
					message.Nonce = msg.Nonce
					return nil
				})
				if err != nil {
					return err
				}
			}

			return nil
		}); err != nil {
			ms.log.Errorf("select message of %s failed %v", addr.Addr, err)
			return err
		}
	}

	//广播推送
	//todo 多点推送
	_, err = ms.nodeClient.MpoolBatchPush(ctx, toPushMessage)
	return err
}

func (ms *MessageService) excludeExpire(ts *venusTypes.TipSet, msgs []*types.Message) ([]*types.Message, []*types.Message) {
	//todo check whether message is expired
	var result []*types.Message
	var expireMsg []*types.Message
	for _, msg := range msgs {
		if msg.Meta.ExpireEpoch != 0 && msg.Meta.ExpireEpoch <= ts.Height() {
			//expire
			msg.State = types.ExpireMsg
			expireMsg = append(expireMsg, msg)
			continue
		}
		result = append(result, msg)
	}
	return result, expireMsg
}

func (ms *MessageService) StartPushMessage(ctx context.Context) {
	tm := time.NewTicker(time.Second * 30)
	defer tm.Stop()

	for {
		select {
		case <-ctx.Done():
			ms.log.Infof("Stop push message")
		case <-tm.C:
			newHead, err := ms.nodeClient.ChainHead(ctx)
			if err != nil {
				ms.log.Errorf("fail to get chain head %v", err)
			}
			err = ms.pushMessageToPool(ctx, newHead)
			if err != nil {
				ms.log.Errorf("push message error %v", err)
			}
		case newHead := <-ms.triggerPush:
			err := ms.pushMessageToPool(ctx, newHead)
			if err != nil {
				ms.log.Errorf("push message error %v", err)
			}
		}
	}
}

func (ms *MessageService) UpdateAllFilledMessage(ctx context.Context) (int, error) {
	msgs := make([]*types.Message, 0)
	for addrStr := range ms.addressService.ListAddressInfo() {
		addr, err := address.NewFromString(addrStr)
		if err != nil {
			return 0, xerrors.Errorf("invalid address %v", addrStr)
		}
		filledMsgs, err := ms.repo.MessageRepo().ListFilledMessageByAddress(addr)
		if err != nil {
			return 0, err
		}
		msgs = append(msgs, filledMsgs...)
	}

	updateCount := 0
	for _, msg := range msgs {
		if err := ms.updateFilledMessage(ctx, msg); err != nil {
			return 0, err
		}
		updateCount++
	}

	return updateCount, nil
}

func (ms *MessageService) updateFilledMessage(ctx context.Context, msg *types.Message) error {
	cid := msg.SignedCid
	if msg.From.Protocol() == address.BLS {
		cid = msg.UnsignedCid
	}
	if cid != nil {
		msgLookup, err := ms.nodeClient.StateSearchMsg(ctx, *cid)
		if err != nil || msgLookup == nil {
			return xerrors.Errorf("search message from node %s %v", cid.String(), err)
		}
		if _, err := ms.UpdateMessageInfoByCid(msg.UnsignedCid.String(), &msgLookup.Receipt, msgLookup.Height, types.OnChainMsg, msgLookup.TipSet.String()); err != nil {
			return err
		}
		ms.log.Infof("update message by node success %v", msg.ID)
	}

	return nil
}

func (ms *MessageService) UpdateFilledMessageByID(ctx context.Context, uuid types.UUID) (types.UUID, error) {
	msg, err := ms.GetMessage(ctx, uuid)
	if err != nil {
		return uuid, err
	}

	return uuid, ms.updateFilledMessage(ctx, msg)
}

func (ms *MessageService) ReplaceMessage(ctx context.Context, uuid types.UUID, auto bool, maxFee string, gasLimit int64, gasPremium string, gasFeecap string) (cid.Cid, error) {
	msg, err := ms.GetMessage(ctx, uuid)
	if err != nil {
		return cid.Undef, xerrors.Errorf("found message %v", err)
	}
	if msg.State == types.OnChainMsg {
		return cid.Undef, xerrors.Errorf("message already on chain")
	}

	if auto {
		minRBF := messagepool.ComputeMinRBF(msg.GasPremium)

		var mss *venusTypes.MessageSendSpec
		if len(maxFee) > 0 {
			maxFee, err := venusTypes.BigFromString(maxFee)
			if err != nil {
				return cid.Undef, fmt.Errorf("parsing max-spend: %w", err)
			}
			mss = &venusTypes.MessageSendSpec{
				MaxFee: maxFee,
			}
		}

		// msg.GasLimit = 0 // TODO: need to fix the way we estimate gas limits to account for the messages already being in the mempool
		msg.GasFeeCap = abi.NewTokenAmount(0)
		msg.GasPremium = abi.NewTokenAmount(0)
		retm, err := ms.nodeClient.GasEstimateMessageGas(ctx, &msg.UnsignedMessage, mss, venusTypes.EmptyTSK)
		if err != nil {
			return cid.Undef, fmt.Errorf("failed to estimate gas values: %w", err)
		}

		msg.GasPremium = big.Max(retm.GasPremium, minRBF)
		msg.GasFeeCap = big.Max(retm.GasFeeCap, msg.GasPremium)

		mff := func() (abi.TokenAmount, error) {
			return abi.TokenAmount(venusTypes.DefaultDefaultMaxFee), nil
		}

		messagepool.CapGasFee(mff, &msg.UnsignedMessage, mss)
	} else {
		if gasLimit > 0 {
			msg.GasLimit = gasLimit
		}
		msg.GasPremium, err = venusTypes.BigFromString(gasPremium)
		if err != nil {
			return cid.Undef, fmt.Errorf("parsing gas-premium: %w", err)
		}
		// TODO: estimate fee cap here
		msg.GasFeeCap, err = venusTypes.BigFromString(gasFeecap)
		if err != nil {
			return cid.Undef, fmt.Errorf("parsing gas-feecap: %w", err)
		}
	}

	addrInfo, exist := ms.addressService.GetAddressInfo(msg.From.String())
	if !exist {
		return cid.Undef, xerrors.Errorf("address not found %s", msg.From.String())
	}

	signedMsg, err := ToSignedMsg(ctx, addrInfo.WalletClient, msg)
	if err != nil {
		return cid.Undef, err
	}

	if _, err := ms.repo.MessageRepo().SaveMessage(msg); err != nil {
		return cid.Undef, err
	}
	err = ms.messageState.MutatorMessage(msg.ID, func(message *types.Message) error {
		message.SignedCid = msg.SignedCid
		message.UnsignedCid = msg.UnsignedCid
		message.UnsignedMessage = msg.UnsignedMessage
		message.State = msg.State
		message.Signature = msg.Signature
		message.Nonce = msg.Nonce
		return nil
	})
	if err != nil {
		return cid.Undef, err
	}

	_, err = ms.nodeClient.MpoolBatchPush(ctx, []*venusTypes.SignedMessage{&signedMsg})

	return signedMsg.Cid(), err
}

func ToSignedMsg(ctx context.Context, walletCli IWalletClient, msg *types.Message) (venusTypes.SignedMessage, error) {
	unsignedCid := msg.UnsignedMessage.Cid()
	msg.UnsignedCid = &unsignedCid
	//签名
	data, err := msg.UnsignedMessage.ToStorageBlock()
	if err != nil {
		return venusTypes.SignedMessage{}, xerrors.Errorf("calc message unsigned message id %s fail %v", msg.ID.String(), err)
	}
	sig, err := walletCli.WalletSign(ctx, msg.From, unsignedCid.Bytes(), core.MsgMeta{
		Type:  core.MTChainMsg,
		Extra: data.RawData(),
	})
	if err != nil {
		return venusTypes.SignedMessage{}, xerrors.Errorf("wallet sign failed %s fail %v", msg.ID.String(), err)
	}

	msg.Signature = sig
	//state
	msg.State = types.FillMsg

	signedMsg := venusTypes.SignedMessage{
		Message:   msg.UnsignedMessage,
		Signature: *msg.Signature,
	}
	signedCid := signedMsg.Cid()
	msg.SignedCid = &signedCid

	return signedMsg, nil
}
