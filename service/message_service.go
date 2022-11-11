package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/ipfs/go-cid"
	"gorm.io/gorm"

	"go.opencensus.io/stats"
	"go.opencensus.io/tag"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/venus-auth/jwtclient"

	"github.com/filecoin-project/venus/pkg/constants"
	v1 "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	gatewayAPI "github.com/filecoin-project/venus/venus-shared/api/gateway/v2"
	venusTypes "github.com/filecoin-project/venus/venus-shared/types"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"

	"github.com/filecoin-project/venus-messager/filestore"
	"github.com/filecoin-project/venus-messager/metrics"
	"github.com/filecoin-project/venus-messager/models/repo"
	"github.com/filecoin-project/venus-messager/publisher"
)

const (
	MaxHeadChangeProcess = 5
	LookBackLimit        = 900
)

type MessageService struct {
	repo           repo.Repo
	fsRepo         filestore.FSRepo
	nodeClient     v1.FullNode
	addressService *AddressService
	walletClient   gatewayAPI.IWalletClient

	publisher publisher.IMsgPublisher

	triggerPush chan *venusTypes.TipSet
	headChans   chan *headChan

	tsCache *TipsetCache

	messageSelector *MessageSelector

	sps *SharedParamsService

	preCancel context.CancelFunc

	cleanUnFillMsgFunc chan func() (int, error)
	cleanUnFillMsgRes  chan cleanUnFillMsgResult

	blockDelay time.Duration
}

type headChan struct {
	apply, revert []*venusTypes.TipSet
	isReconnect   bool
	done          chan error
}

type cleanUnFillMsgResult struct {
	count int
	err   error
}

func NewMessageService(ctx context.Context,
	repo repo.Repo,
	nc v1.FullNode,
	fsRepo filestore.FSRepo,
	addressService *AddressService,
	sps *SharedParamsService,
	walletClient gatewayAPI.IWalletClient, msgPublisher publisher.IMsgPublisher) (*MessageService, error) {
	selector := NewMessageSelector(repo, &fsRepo.Config().MessageService, nc, addressService, sps, walletClient)
	ms := &MessageService{
		repo:               repo,
		nodeClient:         nc,
		fsRepo:             fsRepo,
		messageSelector:    selector,
		headChans:          make(chan *headChan, MaxHeadChangeProcess),
		publisher:          msgPublisher,
		addressService:     addressService,
		walletClient:       walletClient,
		tsCache:            newTipsetCache(),
		triggerPush:        make(chan *venusTypes.TipSet, 20),
		sps:                sps,
		cleanUnFillMsgFunc: make(chan func() (int, error)),
		cleanUnFillMsgRes:  make(chan cleanUnFillMsgResult),
	}
	ms.refreshMessageState(ctx)
	if err := ms.tsCache.Load(ms.fsRepo.TipsetFile()); err != nil {
		log.Infof("load tipset file failed: %v", err)
	}

	// 本身缺少 global context
	if fsRepo.Config().Metrics.Enabled {
		go ms.recordMetricsProc(ctx)
	}

	networkParams, err := ms.nodeClient.StateGetNetworkParams(ctx)
	if err != nil {
		return nil, fmt.Errorf("get network params failed %v", err)
	}
	ms.blockDelay = time.Duration(networkParams.BlockDelaySecs) * time.Second

	return ms, ms.verifyNetworkName()
}

func (ms *MessageService) verifyNetworkName() error {
	networkName, err := ms.nodeClient.StateNetworkName(context.Background())
	if err != nil {
		return err
	}
	if len(ms.tsCache.NetworkName) != 0 {
		if ms.tsCache.NetworkName != string(networkName) {
			return fmt.Errorf("network name not match, expect %s, actual %s, please remove `%s`",
				networkName, ms.tsCache.NetworkName, ms.fsRepo.TipsetFile())
		}
		return nil
	}
	ms.tsCache.NetworkName = string(networkName)

	return ms.tsCache.Save(ms.fsRepo.TipsetFile())
}

func (ms *MessageService) pushMessage(ctx context.Context, msg *types.Message) error {
	if len(msg.ID) == 0 {
		return errors.New("empty uid")
	}

	// replace address
	if msg.From.Protocol() == address.ID {
		fromA, err := ms.nodeClient.StateAccountKey(ctx, msg.From, venusTypes.EmptyTSK)
		if err != nil {
			return fmt.Errorf("getting key address: %w", err)
		}
		log.Warnf("Push from ID address (%s), adjusting to %s", msg.From, fromA)
		msg.From = fromA
	}

	accounts, err := ms.addressService.GetAccountsOfSigner(ctx, msg.From)
	if err != nil {
		return fmt.Errorf("get accounts for %s: %w", msg.From.String(), err)
	}
	has, err := ms.walletClient.WalletHas(ctx, msg.From, accounts)
	if err != nil {
		return err
	}
	if !has {
		return fmt.Errorf("signer address %s not exists", msg.From)
	}
	var addrInfo *types.Address
	if err := ms.repo.Transaction(func(txRepo repo.TxRepo) error {
		addrInfo, err = txRepo.AddressRepo().GetAddress(ctx, msg.From)
		if err == nil {
			return nil
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err = txRepo.AddressRepo().SaveAddress(ctx, &types.Address{
				ID:        venusTypes.NewUUID(),
				Addr:      msg.From,
				Nonce:     0,
				SelMsgNum: 0,
				State:     types.AddressStateAlive,
				IsDeleted: repo.NotDeleted,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}); err != nil {
				return fmt.Errorf("save address %s failed %v", msg.From.String(), err)
			}
			log.Infof("add new address %s", msg.From.String())
		}
		return err
	}); err != nil {
		return err
	}
	if addrInfo != nil && addrInfo.State == types.AddressStateForbbiden {
		log.Errorf("address(%s) is forbidden", msg.From.String())
		return fmt.Errorf("address(%s) is forbidden", msg.From.String())
	}

	msg.Nonce = 0

	return ms.repo.MessageRepo().CreateMessage(msg)
}

func (ms *MessageService) PushMessage(ctx context.Context, msg *venusTypes.Message, meta *types.SendSpec) (string, error) {
	return ms.PushMessageWithId(ctx, venusTypes.NewUUID().String(), msg, meta)
}

func (ms *MessageService) PushMessageWithId(ctx context.Context, id string, msg *venusTypes.Message, meta *types.SendSpec) (string, error) {
	account, _ := jwtclient.CtxGetName(ctx)
	if err := ms.pushMessage(ctx, &types.Message{
		ID:         id,
		Message:    *msg,
		Meta:       meta,
		WalletName: account,
		State:      types.UnFillMsg,
	}); err != nil {
		log.Errorf("push message %s failed %v", id, err)
		return id, err
	}

	return id, nil
}

func (ms *MessageService) WaitMessage(ctx context.Context, id string, confidence uint64) (*types.Message, error) {
	d := time.Second * 30
	if ms.blockDelay > 0 {
		d = ms.blockDelay
	}
	tm := time.NewTicker(d)
	defer tm.Stop()

	doneCh := make(chan struct{}, 1)
	doneCh <- struct{}{}

	for {
		select {
		case <-doneCh:
			msg, err := ms.GetMessageByUid(ctx, id)
			if err != nil {
				return nil, err
			}

			switch msg.State {
			// OffChain
			case types.FillMsg:
				fallthrough
			case types.UnFillMsg:
				fallthrough
			case types.UnKnown:
				continue
			// OnChain
			case types.ReplacedMsg:
				if msg.Confidence > int64(confidence) {
					return msg, nil
				}
				continue
			case types.OnChainMsg:
				if msg.Confidence > int64(confidence) {
					return msg, nil
				}
				continue
			// Error
			case types.FailedMsg:
				return msg, nil
			case types.NoWalletMsg:
				return nil, errors.New("msg failed due to wallet disappear")
			}

		case <-tm.C:
			doneCh <- struct{}{}
		case <-ctx.Done():
			return nil, errors.New("exit by client ")
		}
	}
}

func (ms *MessageService) GetMessageByUid(ctx context.Context, id string) (*types.Message, error) {
	ts, err := ms.nodeClient.ChainHead(ctx)
	if err != nil {
		return nil, err
	}
	msg, err := ms.repo.MessageRepo().GetMessageByUid(id)
	if err != nil {
		return nil, err
	}
	if isChainMsg(msg.State) {
		msg.Confidence = int64(ts.Height()) - msg.Height
	}
	return msg, nil
}

func isChainMsg(msgState types.MessageState) bool {
	return msgState == types.OnChainMsg || msgState == types.ReplacedMsg
}

func (ms *MessageService) HasMessageByUid(ctx context.Context, id string) (bool, error) {
	return ms.repo.MessageRepo().HasMessageByUid(id)
}

func (ms *MessageService) GetMessageByCid(ctx context.Context, id cid.Cid) (*types.Message, error) {
	ts, err := ms.nodeClient.ChainHead(ctx)
	if err != nil {
		return nil, err
	}
	msg, err := ms.repo.MessageRepo().GetMessageByCid(id)
	if err != nil {
		return nil, err
	}
	if isChainMsg(msg.State) {
		msg.Confidence = int64(ts.Height()) - msg.Height
	}
	return msg, nil
}

func (ms *MessageService) GetMessageState(ctx context.Context, id string) (types.MessageState, error) {
	return ms.repo.MessageRepo().GetMessageState(id)
}

func (ms *MessageService) GetMessageBySignedCid(ctx context.Context, signedCid cid.Cid) (*types.Message, error) {
	ts, err := ms.nodeClient.ChainHead(ctx)
	if err != nil {
		return nil, err
	}
	msg, err := ms.repo.MessageRepo().GetMessageBySignedCid(signedCid)
	if err != nil {
		return nil, err
	}
	if isChainMsg(msg.State) {
		msg.Confidence = int64(ts.Height()) - msg.Height
	}
	return msg, nil
}

func (ms *MessageService) GetMessageByUnsignedCid(ctx context.Context, unsignedCid cid.Cid) (*types.Message, error) {
	ts, err := ms.nodeClient.ChainHead(ctx)
	if err != nil {
		return nil, err
	}
	msg, err := ms.repo.MessageRepo().GetMessageByCid(unsignedCid)
	if err != nil {
		return nil, err
	}
	if isChainMsg(msg.State) {
		msg.Confidence = int64(ts.Height()) - msg.Height
	}
	return msg, nil
}

func (ms *MessageService) GetMessageByFromAndNonce(ctx context.Context, from address.Address, nonce uint64) (*types.Message, error) {
	ts, err := ms.nodeClient.ChainHead(ctx)
	if err != nil {
		return nil, err
	}
	msg, err := ms.repo.MessageRepo().GetMessageByFromAndNonce(from, nonce)
	if err != nil {
		return nil, err
	}
	if isChainMsg(msg.State) {
		msg.Confidence = int64(ts.Height()) - msg.Height
	}
	return msg, nil
}

func (ms *MessageService) ListMessageByFromState(ctx context.Context, from address.Address, state types.MessageState, isAsc bool, pageIndex, pageSize int) ([]*types.Message, error) {
	ts, err := ms.nodeClient.ChainHead(ctx)
	if err != nil {
		return nil, err
	}
	msgs, err := ms.repo.MessageRepo().ListMessageByFromState(from, state, isAsc, pageIndex, pageSize)
	if err != nil {
		return nil, err
	}

	for _, msg := range msgs {
		if isChainMsg(msg.State) {
			msg.Confidence = int64(ts.Height()) - msg.Height
		}
	}
	return msgs, nil
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
		if isChainMsg(msg.State) {
			msg.Confidence = int64(ts.Height()) - msg.Height
		}
	}
	return msgs, nil
}

func (ms *MessageService) ListMessageByAddress(ctx context.Context, addr address.Address) ([]*types.Message, error) {
	ts, err := ms.nodeClient.ChainHead(ctx)
	if err != nil {
		return nil, err
	}
	msgs, err := ms.repo.MessageRepo().ListMessageByAddress(addr)
	if err != nil {
		return nil, err
	}

	for _, msg := range msgs {
		if isChainMsg(msg.State) {
			msg.Confidence = int64(ts.Height()) - msg.Height
		}
	}
	return msgs, nil
}

func (ms *MessageService) ListFailedMessage(ctx context.Context) ([]*types.Message, error) {
	return ms.repo.MessageRepo().ListFailedMessage()
}

func (ms *MessageService) ListFilledMessageByAddress(ctx context.Context, addr address.Address) ([]*types.Message, error) {
	return ms.repo.MessageRepo().ListFilledMessageByAddress(addr)
}

func (ms *MessageService) ListBlockedMessage(ctx context.Context, addr address.Address, d time.Duration) ([]*types.Message, error) {
	var msgs []*types.Message
	if addr != address.Undef {
		return ms.repo.MessageRepo().ListBlockedMessage(addr, d)
	}

	addrList, err := ms.addressService.ListActiveAddress(ctx)
	if err != nil {
		return nil, err
	}
	for _, a := range addrList {
		msgsT, err := ms.repo.MessageRepo().ListBlockedMessage(a.Addr, d)
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, msgsT...)
	}

	return msgs, nil
}

func (ms *MessageService) UpdateMessageStateByCid(ctx context.Context, cid string, state types.MessageState) (string, error) {
	return cid, ms.repo.MessageRepo().UpdateMessageStateByCid(cid, state)
}

func (ms *MessageService) UpdateMessageStateByID(ctx context.Context, id string, state types.MessageState) error {
	return ms.repo.MessageRepo().UpdateMessageStateByID(id, state)
}

func (ms *MessageService) UpdateMessageInfoByCid(unsignedCid string, receipt *venusTypes.MessageReceipt,
	height abi.ChainEpoch, state types.MessageState, tsKey venusTypes.TipSetKey,
) (string, error) {
	return unsignedCid, ms.repo.MessageRepo().UpdateMessageInfoByCid(unsignedCid, receipt, height, state, tsKey)
}

func (ms *MessageService) ProcessNewHead(ctx context.Context, apply []*venusTypes.TipSet) error {
	log.Infof("receive new head from chain")
	if ms.fsRepo.Config().MessageService.SkipProcessHead {
		log.Infof("skip process new head")
		return nil
	}

	if len(apply) == 0 {
		log.Errorf("expect apply blocks, but got none")
		return nil
	}

	tsList := ms.tsCache.List()
	sort.Slice(tsList, func(i, j int) bool {
		return tsList[i].Height() > tsList[j].Height()
	})
	smallestTs := apply[len(apply)-1]

	defer log.Infof("%d head wait to process", len(ms.headChans))

	if len(tsList) == 0 || smallestTs.Parents().Equals(tsList[0].Key()) {
		log.Infof("apply a block height %d %s", apply[0].Height(), apply[0].String())
		done := make(chan error)
		ms.headChans <- &headChan{
			apply:  apply,
			revert: nil,
			done:   done,
		}
		return <-done
	}

	localApply, revertTipset, err := ms.lookAncestors(ctx, tsList, smallestTs)
	if err != nil {
		log.Errorf("look ancestor error from %s and %s, error: %v", smallestTs, tsList[0].Key(), err)
		return nil
	}

	if len(apply) > 1 {
		localApply = append(apply[:len(apply)-1], localApply...)
	}
	done := make(chan error)
	ms.headChans <- &headChan{
		apply:  localApply,
		revert: revertTipset,
		done:   done,
	}
	return <-done
}

func (ms *MessageService) ReconnectCheck(ctx context.Context, head *venusTypes.TipSet) error {
	log.Infof("reconnect to node")

	if len(ms.tsCache.Cache) == 0 {
		count, err := ms.UpdateAllFilledMessage(ctx)
		if err != nil {
			return err
		}
		log.Infof("update filled message count %v", count)
		return nil
	}

	tsList := ms.tsCache.List()
	sort.Slice(tsList, func(i, j int) bool {
		return tsList[i].Height() > tsList[j].Height()
	})

	// long time not use
	if head.Height()-tsList[0].Height() >= LookBackLimit {
		count, err := ms.UpdateAllFilledMessage(ctx)
		if err != nil {
			return err
		}
		log.Infof("gap height %v, update filled message count %v", head.Height()-tsList[0].Height(), count)
		return nil
	}

	if tsList[0].Height() == head.Height() && tsList[0].Equals(head) {
		log.Infof("The head does not change and returns directly.")
		return nil
	}

	gapTipset, revertTipset, err := ms.lookAncestors(ctx, tsList, head)
	if err != nil {
		return err
	}

	done := make(chan error)
	ms.headChans <- &headChan{
		apply:       gapTipset,
		revert:      revertTipset,
		done:        done,
		isReconnect: true,
	}

	return <-done
}

func (ms *MessageService) lookAncestors(ctx context.Context, localTipset []*venusTypes.TipSet, head *venusTypes.TipSet) ([]*venusTypes.TipSet, []*venusTypes.TipSet, error) {
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
		if localTs.Height() > ts.Height() {
			idx++
		} else if localTs.Height() == ts.Height() {
			if localTs.Equals(ts) {
				break
			}
			idx++
		} else {
			gapTipset = append(gapTipset, ts)
			ts, err = ms.nodeClient.ChainGetTipSet(ctx, ts.Parents())
			if err != nil {
				return nil, nil, fmt.Errorf("got tipset(%s) failed %v", ts.Parents(), err)
			}
		}
		loopCount++
	}

	if idx >= localTsLen {
		idx = localTsLen
	}
	revertTs := localTipset[:idx]

	return gapTipset, revertTs, err
}

///   Message push    ////

func (ms *MessageService) pushMessageToPool(ctx context.Context, ts *venusTypes.TipSet) error {
	// select message
	startSelectMsg := time.Now()
	selectResult, err := ms.messageSelector.SelectMessage(ctx, ts)
	if err != nil {
		return err
	}
	selectMsgSpent := time.Since(startSelectMsg)
	log.Infof("current loop select result | SelectMsg: %d | ExpireMsg: %d | ToPushMsg: %d | ErrMsg: %d", len(selectResult.SelectMsg), len(selectResult.ExpireMsg), len(selectResult.ToPushMsg), len(selectResult.ErrMsg))
	stats.Record(ctx, metrics.SelectedMsgNumOfLastRound.M(int64(len(selectResult.SelectMsg))))
	stats.Record(ctx, metrics.ToPushMsgNumOfLastRound.M(int64(len(selectResult.ToPushMsg))))
	stats.Record(ctx, metrics.ExpiredMsgNumOfLastRound.M(int64(len(selectResult.ExpireMsg))))
	stats.Record(ctx, metrics.ErrMsgNumOfLastRound.M(int64(len(selectResult.ErrMsg))))

	startSaveDB := time.Now()
	log.Infof("start to save to database")
	if err := ms.saveSelectedMessagesToDB(ctx, selectResult); err != nil {
		return err
	}
	saveDBSpent := time.Since(startSaveDB)
	log.Infof("success to save to database")

	for _, msg := range selectResult.SelectMsg {
		selectResult.ToPushMsg = append(selectResult.ToPushMsg, &venusTypes.SignedMessage{
			Message:   msg.Message,
			Signature: *msg.Signature,
		})
	}

	// broad cast push to node in config, push to multi node in db config, publish to pubsub
	go func() {
		startPush := time.Now()
		ms.multiPushMessages(ctx, selectResult)
		log.Infof("Push message select spent: %v, save db spent: %v, push to node spent: %v",
			selectMsgSpent,
			saveDBSpent,
			time.Since(startPush),
		)
	}()
	return err
}

func (ms *MessageService) saveSelectedMessagesToDB(ctx context.Context, selectResult *MsgSelectResult) error {
	if err := ms.repo.Transaction(func(txRepo repo.TxRepo) error {
		// 保存消息
		err := txRepo.MessageRepo().ExpireMessage(selectResult.ExpireMsg)
		if err != nil {
			return err
		}

		err = txRepo.MessageRepo().BatchSaveMessage(selectResult.SelectMsg)
		if err != nil {
			return err
		}

		for _, addr := range selectResult.ModifyAddress {
			err = txRepo.AddressRepo().UpdateNonce(ctx, addr.Addr, addr.Nonce)
			if err != nil {
				return err
			}
		}

		for _, m := range selectResult.ErrMsg {
			log.Infof("update message %s return value with error %s", m.id, m.err)
			err := txRepo.MessageRepo().UpdateReturnValue(m.id, m.err)
			if err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		log.Errorf("save signed message failed %v", err)
		return err
	}
	return nil
}

func (ms *MessageService) multiPushMessages(ctx context.Context, selectResult *MsgSelectResult) {
	pushMsgByAddr := make(map[address.Address][]*venusTypes.SignedMessage)
	for _, msg := range selectResult.ToPushMsg {
		if val, ok := pushMsgByAddr[msg.Message.From]; ok {
			pushMsgByAddr[msg.Message.From] = append(val, msg)
		} else {
			pushMsgByAddr[msg.Message.From] = []*venusTypes.SignedMessage{msg}
		}
	}

	for addr, msgs := range pushMsgByAddr {
		sort.Slice(msgs, func(i, j int) bool {
			return msgs[i].Message.Nonce < msgs[j].Message.Nonce
		})
		pushMsgByAddr[addr] = msgs
	}

	log.Infof("start to push message %d to mpool", len(selectResult.ToPushMsg))

	for addr, msgs := range pushMsgByAddr {
		err := ms.publisher.PublishMessages(ctx, msgs)
		if err != nil {
			log.Errorf("publish message from address %s failed %v", addr, err)
		}
	}
}

func (ms *MessageService) StartPushMessage(ctx context.Context, skipPushMsg bool) {
	tm := time.NewTicker(time.Second * 30)
	defer tm.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Warnf("stop push message: %v", ctx.Err())
			return
		case <-tm.C:
			//newHead, err := ms.nodeClient.ChainHead(ctx)
			//if err != nil {
			//	log.Errorf("fail to get chain head %v", err)
			//}
			//err = ms.pushMessageToPool(ctx, newHead)
			//if err != nil {
			//	log.Errorf("push message error %v", err)
			//}
		case newHead := <-ms.triggerPush:
			// Clear all unfill messages by address
			ms.tryClearUnFillMsg()

			if skipPushMsg {
				log.Info("skip push message")
				continue
			}
			start := time.Now()
			log.Infof("start to push message %s task wait task %d", newHead.String(), len(ms.triggerPush))
			err := ms.pushMessageToPool(ctx, newHead)
			if err != nil {
				log.Errorf("push message error %v", err)
			}
			log.Infof("end push message spent %d ms", time.Since(start).Milliseconds())
		}
	}
}

func (ms *MessageService) tryClearUnFillMsg() {
	select {
	case f := <-ms.cleanUnFillMsgFunc:
		count, err := f()
		ms.cleanUnFillMsgRes <- cleanUnFillMsgResult{
			count: count,
			err:   err,
		}
	default:
	}
}

func (ms *MessageService) UpdateAllFilledMessage(ctx context.Context) (int, error) {
	msgs := make([]*types.Message, 0)

	for addr := range ms.addressService.ActiveAddresses(ctx) {
		filledMsgs, err := ms.repo.MessageRepo().ListFilledMessageByAddress(addr)
		if err != nil {
			log.Errorf("list filled message %v %v", addr, err)
			continue
		}
		msgs = append(msgs, filledMsgs...)
	}

	log.Infof("%d messages need to sync", len(msgs))
	updateCount := 0
	for _, msg := range msgs {
		if err := ms.updateFilledMessage(ctx, msg); err != nil {
			log.Errorf("failed to update filled message: %v", err)
			continue
		}
		updateCount++
	}

	return updateCount, nil
}

func (ms *MessageService) updateFilledMessage(ctx context.Context, msg *types.Message) error {
	cid := msg.SignedCid
	if cid != nil {
		msgLookup, err := ms.nodeClient.StateSearchMsg(ctx, venusTypes.EmptyTSK, *cid, constants.LookbackNoLimit, true)
		if err != nil || msgLookup == nil {
			return fmt.Errorf("search message %s from node %v", cid.String(), err)
		}
		if _, err := ms.UpdateMessageInfoByCid(msg.UnsignedCid.String(), &msgLookup.Receipt, msgLookup.Height, types.OnChainMsg, msgLookup.TipSet); err != nil {
			return err
		}
		log.Infof("update message %v by node success, height: %d", msg.ID, msgLookup.Height)
	}

	return nil
}

func (ms *MessageService) UpdateFilledMessageByID(ctx context.Context, id string) (string, error) {
	msg, err := ms.GetMessageByUid(ctx, id)
	if err != nil {
		return id, err
	}

	return id, ms.updateFilledMessage(ctx, msg)
}

func (ms *MessageService) ReplaceMessage(ctx context.Context, params *types.ReplacMessageParams) (cid.Cid, error) {
	if params == nil {
		return cid.Undef, fmt.Errorf("params is nil")
	}
	msg, err := ms.GetMessageByUid(ctx, params.ID)
	if err != nil {
		return cid.Undef, fmt.Errorf("found message %v", err)
	}
	if msg.State == types.OnChainMsg {
		return cid.Undef, fmt.Errorf("message already on chain")
	}

	if params.Auto {
		minRBF := computeMinRBF(msg.GasPremium)

		mss := &venusTypes.MessageSendSpec{
			MaxFee:         params.MaxFee,
			GasOverPremium: params.GasOverPremium,
		}

		// msg.GasLimit = 0 // TODO: need to fix the way we estimate gas limits to account for the messages already being in the mempool
		msg.GasFeeCap = abi.NewTokenAmount(0)
		msg.GasPremium = abi.NewTokenAmount(0)
		retm, err := ms.nodeClient.GasEstimateMessageGas(ctx, &msg.Message, mss, venusTypes.EmptyTSK)
		if err != nil {
			return cid.Undef, fmt.Errorf("failed to estimate gas values: %w", err)
		}

		msg.GasPremium = big.Max(retm.GasPremium, minRBF)
		msg.GasFeeCap = big.Max(retm.GasFeeCap, msg.GasPremium)

		if mss.MaxFee.NilOrZero() {
			addrInfo, err := ms.addressService.GetAddress(ctx, msg.From)
			if err != nil {
				return cid.Undef, err
			}
			maxFee := addrInfo.MaxFee
			sharedParams, err := ms.sps.GetSharedParams(ctx)
			if err != nil {
				return cid.Undef, err
			}
			if maxFee.NilOrZero() {
				maxFee = sharedParams.MaxFee
			}
			mss.MaxFee = maxFee
		}

		CapGasFee(&msg.Message, mss.MaxFee)
	} else {
		if params.GasLimit > 0 {
			msg.GasLimit = params.GasLimit
		}
		if big.Cmp(params.GasPremium, big.Zero()) <= 0 {
			return cid.Undef, fmt.Errorf("gas premium(%s) must bigger than zero", params.GasPremium)
		}
		if big.Cmp(params.GasFeecap, big.Zero()) <= 0 {
			return cid.Undef, fmt.Errorf("gas feecap(%s) must bigger than zero", params.GasFeecap)
		}
		if big.Cmp(msg.GasFeeCap, msg.GasPremium) < 0 {
			return cid.Undef, fmt.Errorf("gas feecap(%s) must bigger or equal than gas premium (%s)", msg.GasFeeCap, msg.GasPremium)
		}
		msg.GasPremium = params.GasPremium
		// TODO: estimate fee cap here
		msg.GasFeeCap = params.GasFeecap
	}

	accounts, err := ms.addressService.GetAccountsOfSigner(ctx, msg.From)
	if err != nil {
		return cid.Undef, err
	}
	signedMsg, err := ToSignedMsg(ctx, ms.walletClient, msg, accounts)
	if err != nil {
		return cid.Undef, err
	}

	if err := ms.repo.MessageRepo().SaveMessage(msg); err != nil {
		return cid.Undef, err
	}

	_, err = ms.nodeClient.MpoolBatchPush(ctx, []*venusTypes.SignedMessage{&signedMsg})

	return signedMsg.Cid(), err
}

func (ms *MessageService) MarkBadMessage(ctx context.Context, id string) error {
	return ms.repo.MessageRepo().MarkBadMessage(id)
}

func (ms *MessageService) RecoverFailedMsg(ctx context.Context, addr address.Address) ([]string, error) {
	recoverIDs := make([]string, 0)
	actor, err := ms.nodeClient.StateGetActor(ctx, addr, venusTypes.EmptyTSK)
	if err != nil {
		return nil, err
	}
	addrInfo, err := ms.repo.AddressRepo().GetAddress(ctx, addr)
	if err != nil {
		return nil, err
	}
	if addrInfo.Nonce < actor.Nonce {
		return recoverIDs, nil
	}
	msgs, err := ms.repo.MessageRepo().GetSignedMessageFromFailedMsg(addr)
	if err != nil {
		return nil, err
	}
	for _, msg := range msgs {
		if msg.Nonce >= actor.Nonce {
			if err = ms.repo.MessageRepo().UpdateMessageStateByID(msg.ID, types.FillMsg); err != nil {
				return nil, err
			}
			recoverIDs = append(recoverIDs, msg.ID)
		}
	}

	return recoverIDs, nil
}

func (ms *MessageService) RepublishMessage(ctx context.Context, id string) error {
	msg, err := ms.GetMessageByUid(ctx, id)
	if err != nil {
		return nil
	}
	if msg.State != types.FillMsg {
		return fmt.Errorf("need FillMsg got %s", msg.State)
	}
	signedMsg := &venusTypes.SignedMessage{
		Message:   msg.Message,
		Signature: *msg.Signature,
	}

	msgs := []*venusTypes.SignedMessage{signedMsg}
	err = ms.publisher.PublishMessages(ctx, msgs)
	return err
}

func ToSignedMsg(ctx context.Context, walletCli gatewayAPI.IWalletClient, msg *types.Message, accounts []string) (venusTypes.SignedMessage, error) {
	unsignedCid := msg.Message.Cid()
	msg.UnsignedCid = &unsignedCid
	// 签名
	data, err := msg.Message.ToStorageBlock()
	if err != nil {
		return venusTypes.SignedMessage{}, fmt.Errorf("calc message unsigned message id %s fail %v", msg.ID, err)
	}
	sig, err := walletCli.WalletSign(ctx, msg.From, accounts, unsignedCid.Bytes(), venusTypes.MsgMeta{
		Type:  venusTypes.MTChainMsg,
		Extra: data.RawData(),
	})
	if err != nil {
		return venusTypes.SignedMessage{}, fmt.Errorf("wallet sign failed %s fail %v", msg.ID, err)
	}

	msg.Signature = sig
	// state
	msg.State = types.FillMsg

	signedMsg := venusTypes.SignedMessage{
		Message:   msg.Message,
		Signature: *msg.Signature,
	}
	signedCid := signedMsg.Cid()
	msg.SignedCid = &signedCid

	return signedMsg, nil
}

func (ms *MessageService) clearUnFillMessage(ctx context.Context, addr address.Address) (int, error) {
	var count int
	if err := ms.repo.Transaction(func(txRepo repo.TxRepo) error {
		unFillMsgs, err := txRepo.MessageRepo().ListUnFilledMessage(addr)
		if err != nil {
			return err
		}
		for _, msg := range unFillMsgs {
			if err := txRepo.MessageRepo().MarkBadMessage(msg.ID); err != nil {
				return fmt.Errorf("mark bad message %s failed %v", msg.ID, err)
			}
			count++
		}
		return nil
	}); err != nil {
		return 0, err
	}

	return count, nil
}

func (ms *MessageService) ClearUnFillMessage(ctx context.Context, addr address.Address) (int, error) {
	ms.cleanUnFillMsgFunc <- func() (int, error) {
		return ms.clearUnFillMessage(ctx, addr)
	}

	select {
	case r, ok := <-ms.cleanUnFillMsgRes:
		if !ok {
			return 0, fmt.Errorf("unexpect error")
		}
		log.Infof("clear unfill messages success, address: %v, count: %d", addr.String(), r.count)
		return r.count, r.err
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}

func (ms *MessageService) recordMetricsProc(ctx context.Context) {
	tm := time.NewTicker(time.Second * 60)
	defer tm.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Warnf("stop record metrics: %v", ctx.Err())
			return
		case <-tm.C:
			addrs, err := ms.addressService.ListActiveAddress(ctx)
			if err != nil {
				log.Errorf("get address list err: %s", err)
			}

			for _, addr := range addrs {
				ctx, _ = tag.New(
					ctx,
					tag.Upsert(metrics.WalletAddress, addr.Addr.String()),
				)
				stats.Record(ctx, metrics.WalletDBNonce.M(int64(addr.Nonce)))

				actor, err := ms.nodeClient.StateGetActor(ctx, addr.Addr, venusTypes.EmptyTSK)
				if err != nil {
					log.Errorf("get actor err: %s", err)
				} else {
					balance, _ := strconv.ParseFloat(venusTypes.FIL(actor.Balance).Unitless(), 64)
					stats.Record(ctx, metrics.WalletBalance.M(balance))
					stats.Record(ctx, metrics.WalletChainNonce.M(int64(actor.Nonce)))
				}

				msgs, err := ms.repo.MessageRepo().ListUnFilledMessage(addr.Addr)
				if err != nil {
					log.Errorf("get unFilled msg err: %s", err)
				} else {
					stats.Record(ctx, metrics.NumOfUnFillMsg.M(int64(len(msgs))))
				}

				msgs, err = ms.repo.MessageRepo().ListFilledMessageByAddress(addr.Addr)
				if err != nil {
					log.Errorf("get filled msg err: %s", err)
				} else {
					stats.Record(ctx, metrics.NumOfFillMsg.M(int64(len(msgs))))
				}

				msgs, err = ms.repo.MessageRepo().ListBlockedMessage(addr.Addr, 3*time.Minute)
				if err != nil {
					log.Errorf("get blocked three minutes msg err: %s", err)
				} else {
					stats.Record(ctx, metrics.NumOfMsgBlockedThreeMinutes.M(int64(len(msgs))))
				}

				msgs, err = ms.repo.MessageRepo().ListBlockedMessage(addr.Addr, 5*time.Minute)
				if err != nil {
					log.Errorf("get blocked five minutes msg err: %s", err)
				} else {
					stats.Record(ctx, metrics.NumOfMsgBlockedFiveMinutes.M(int64(len(msgs))))
				}
			}

			msgs, err := ms.repo.MessageRepo().ListFailedMessage()
			if err != nil {
				log.Errorf("get failed msg err: %s", err)
			} else {
				stats.Record(ctx, metrics.NumOfFailedMsg.M(int64(len(msgs))))
			}
		}
	}
}

const (
	ReplaceByFeeRatioDefault = 1.25
	RbfDenom                 = 256
)

var (
	rbfNumBig   = big.NewInt(int64((ReplaceByFeeRatioDefault - 1) * RbfDenom))
	rbfDenomBig = big.NewInt(RbfDenom)
)

func computeMinRBF(curPrem abi.TokenAmount) abi.TokenAmount {
	minPrice := big.Add(curPrem, big.Div(big.Mul(curPrem, rbfNumBig), rbfDenomBig))
	return big.Add(minPrice, big.NewInt(1))
}
