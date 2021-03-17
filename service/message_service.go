package service

import (
	"context"
	"github.com/ipfs-force-community/venus-wallet/core"
	"sort"
	"sync"
	"time"

	"github.com/ipfs-force-community/venus-messager/utils"

	"github.com/filecoin-project/go-address"

	venusTypes "github.com/filecoin-project/venus/pkg/types"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"

	"github.com/ipfs-force-community/venus-messager/config"
	"github.com/ipfs-force-community/venus-messager/models/repo"
	"github.com/ipfs-force-community/venus-messager/types"
)

const (
	MaxHeadChangeProcess = 5

	LookBackLimit = 1000

	maxStoreTipsetCount = 1000
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
	Cache      map[uint64]*tipsetFormat
	CurrHeight uint64

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
			Cache:      make(map[uint64]*tipsetFormat, maxStoreTipsetCount),
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

func (ms *MessageService) GetMessageByCid(background context.Context, cid string) (*types.Message, error) {
	return ms.repo.MessageRepo().GetMessageByCid(cid)
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
			msg.Confidence = int64(ts.Height()) - int64(msg.Height)
		}
	}
	return msgs, nil
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

	if tsList[0].Height == uint64(head.Height()) && tsList[0].Key == head.String() {
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
		if localTs.Height > uint64(ts.Height()) {
			idx++
		} else if localTs.Height == uint64(ts.Height()) {
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
				txRepo.AddressRepo().SaveAddress(ctx, addr)
			}
			//todo push sigined but not onchain message, when to resend message
			filledMessage, err := txRepo.MessageRepo().ListFilledMessageByAddress(mAddr)
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
					ms.log.Errorf("GasEstimateMessageGas msg id fail %v", msg.ID.String(), err)
					continue
				}
				msg.GasFeeCap = newMsg.GasFeeCap
				msg.GasPremium = newMsg.GasPremium
				msg.GasLimit = newMsg.GasLimit

				unsignedCid := msg.UnsignedMessage.Cid()
				msg.UnsignedCid = &unsignedCid
				//签名
				data, err := msg.UnsignedMessage.ToStorageBlock()
				if err != nil {
					ms.log.Errorf("calc message unsigned message id %s fail %v", msg.ID.String(), err)
					continue
				}
				sig, err := addrInfo.WalletClient.WalletSign(ctx, mAddr, unsignedCid.Bytes(), core.MsgMeta{
					Type:  core.MTChainMsg,
					Extra: data.RawData(),
				})
				if err != nil {
					//todo client net crash?
					ms.log.Errorf("wallet sign failed %s fail %v", msg.ID.String(), err)
					continue
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
