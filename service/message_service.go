package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/venus/pkg/constants"
	v1 "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	venusTypes "github.com/filecoin-project/venus/venus-shared/types"
	"github.com/ipfs/go-cid"
	"gorm.io/gorm"

	"github.com/filecoin-project/venus-messager/config"
	"github.com/filecoin-project/venus-messager/filestore"
	"github.com/filecoin-project/venus-messager/gateway"
	"github.com/filecoin-project/venus-messager/log"
	"github.com/filecoin-project/venus-messager/models/repo"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
)

var errAlreadyInMpool = fmt.Errorf("already in mpool: validation failure")
var errMinimumNonce = errors.New("minimum expected nonce")

const (
	MaxHeadChangeProcess = 5

	LookBackLimit = 900
)

type MessageService struct {
	repo           repo.Repo
	log            *log.Logger
	fsRepo         filestore.FSRepo
	nodeClient     v1.FullNode
	messageState   *MessageState
	addressService *AddressService
	walletClient   gateway.IWalletClient

	triggerPush chan *venusTypes.TipSet
	headChans   chan *headChan

	tsCache *TipsetCache

	messageSelector *MessageSelector

	sps         *SharedParamsService
	nodeService *NodeService

	preCancel context.CancelFunc

	cleanUnFillMsgFunc chan func() (int, error)
	cleanUnFillMsgRes  chan cleanUnFillMsgResult
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

func NewMessageService(repo repo.Repo,
	nc v1.FullNode,
	logger *log.Logger,
	fsRepo filestore.FSRepo,
	messageState *MessageState,
	addressService *AddressService,
	sps *SharedParamsService,
	nodeService *NodeService,
	walletClient *gateway.IWalletCli) (*MessageService, error) {
	selector := NewMessageSelector(repo, logger, &fsRepo.Config().MessageService, nc, addressService, sps, walletClient)
	ms := &MessageService{
		repo:            repo,
		log:             logger,
		nodeClient:      nc,
		fsRepo:          fsRepo,
		messageSelector: selector,
		headChans:       make(chan *headChan, MaxHeadChangeProcess),

		messageState:       messageState,
		addressService:     addressService,
		walletClient:       walletClient,
		tsCache:            newTipsetCache(),
		triggerPush:        make(chan *venusTypes.TipSet, 20),
		sps:                sps,
		nodeService:        nodeService,
		cleanUnFillMsgFunc: make(chan func() (int, error)),
		cleanUnFillMsgRes:  make(chan cleanUnFillMsgResult),
	}
	ms.refreshMessageState(context.TODO())
	if err := ms.tsCache.Load(ms.fsRepo.TipsetFile()); err != nil {
		ms.log.Infof("load tipset file failed: %v", err)
	}

	return ms, ms.verifyNetworkName()
}

func (ms *MessageService) verifyNetworkName() error {
	networkName, err := ms.nodeClient.StateNetworkName(context.Background())
	if err != nil {
		return err
	}
	if len(ms.tsCache.NetworkName) != 0 {
		if ms.tsCache.NetworkName != string(networkName) {
			return fmt.Errorf("network name not match, need %s, had %s, please remove `%s`",
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

	//replace address
	if msg.From.Protocol() == address.ID {
		fromA, err := ms.nodeClient.StateAccountKey(ctx, msg.From, venusTypes.EmptyTSK)
		if err != nil {
			return fmt.Errorf("getting key address: %w", err)
		}
		ms.log.Warnf("Push from ID address (%s), adjusting to %s", msg.From, fromA)
		msg.From = fromA
	}

	has, err := ms.walletClient.WalletHas(ctx, msg.WalletName, msg.From)
	if err != nil {
		return err
	}
	if !has {
		return fmt.Errorf("wallet(%s) address %s not exists", msg.WalletName, msg.From)
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
				State:     types.AddressStateAlive,
				IsDeleted: repo.NotDeleted,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}); err != nil {
				return fmt.Errorf("save address %s failed %v", msg.From.String(), err)
			}
			ms.log.Infof("add new address %s", msg.From.String())
		}
		return err
	}); err != nil {
		return err
	}
	if addrInfo != nil && addrInfo.State == types.AddressStateForbbiden {
		ms.log.Errorf("address(%s) is forbidden", msg.From.String())
		return fmt.Errorf("address(%s) is forbidden", msg.From.String())
	}

	msg.Nonce = 0
	err = ms.repo.MessageRepo().CreateMessage(msg)
	if err == nil {
		ms.messageState.SetMessage(msg.ID, msg)
	}

	return err
}

func (ms *MessageService) PushMessage(ctx context.Context, account string, msg *venusTypes.Message, meta *types.SendSpec) (string, error) {
	newId := venusTypes.NewUUID()
	if err := ms.pushMessage(ctx, &types.Message{
		ID:         newId.String(),
		Message:    *msg,
		Meta:       meta,
		State:      types.UnFillMsg,
		WalletName: account,
		FromUser:   account,
	}); err != nil {
		ms.log.Errorf("push message %s failed %v", newId.String(), err)
		return newId.String(), err
	}

	return newId.String(), nil
}

func (ms *MessageService) PushMessageWithId(ctx context.Context, account string, id string, msg *venusTypes.Message, meta *types.SendSpec) (string, error) {
	if err := ms.pushMessage(ctx, &types.Message{
		ID:         id,
		Message:    *msg,
		Meta:       meta,
		State:      types.UnFillMsg,
		WalletName: account,
		FromUser:   account,
	}); err != nil {
		ms.log.Errorf("push message %s failed %v", id, err)
		return id, err
	}

	return id, nil
}

func (ms *MessageService) WaitMessage(ctx context.Context, id string, confidence uint64) (*types.Message, error) {
	tm := time.NewTicker(time.Second * 30)
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
			//OffChain
			case types.FillMsg:
				fallthrough
			case types.UnFillMsg:
				fallthrough
			case types.UnKnown:
				continue
			//OnChain
			case types.ReplacedMsg:
				fallthrough
			case types.OnChainMsg:
				if msg.Confidence > int64(confidence) {
					return msg, nil
				}
				continue
			//Error
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
	if msg.State == types.OnChainMsg {
		msg.Confidence = int64(ts.Height()) - msg.Height
	}
	return msg, nil
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
	if msg.State == types.OnChainMsg {
		msg.Confidence = int64(ts.Height()) - msg.Height
	}
	return msg, nil
}

func (ms *MessageService) GetMessageState(ctx context.Context, id string) (types.MessageState, error) {
	return ms.repo.MessageRepo().GetMessageState(id)
}

func (ms *MessageService) GetMessageBySignedCid(ctx context.Context, signedCid cid.Cid) (*types.Message, error) {
	return ms.repo.MessageRepo().GetMessageBySignedCid(signedCid)
}

func (ms *MessageService) GetMessageByUnsignedCid(ctx context.Context, unsignedCid cid.Cid) (*types.Message, error) {
	return ms.repo.MessageRepo().GetMessageByCid(unsignedCid)
}

func (ms *MessageService) GetMessageByFromAndNonce(ctx context.Context, from address.Address, nonce uint64) (*types.Message, error) {
	return ms.repo.MessageRepo().GetMessageByFromAndNonce(from, nonce)
}

func (ms *MessageService) ListMessageByFromState(ctx context.Context, from address.Address, state types.MessageState, isAsc bool, pageIndex, pageSize int) ([]*types.Message, error) {
	return ms.repo.MessageRepo().ListMessageByFromState(from, state, isAsc, pageIndex, pageSize)
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
		if msg.State == types.OnChainMsg {
			msg.Confidence = int64(ts.Height()) - msg.Height
		}
	}
	return msgs, nil
}

func (ms *MessageService) ListFailedMessage(ctx context.Context) ([]*types.Message, error) {
	return ms.repo.MessageRepo().ListFailedMessage()
}

func (ms *MessageService) ListFilledMessageByAddress(ctx context.Context, addr address.Address) ([]*types.Message, error) {
	msgs, err := ms.repo.MessageRepo().ListFilledMessageByAddress(addr)
	if len(msgs) > 0 {
		ids := make([]string, 0, len(msgs))
		for _, msg := range msgs {
			ids = append(ids, msg.ID)
		}
		ms.log.Infof("list failed message by address %s %s", addr, strings.Join(ids, ","))
	}

	return msgs, err
}

func (ms *MessageService) ListBlockedMessage(ctx context.Context, addr address.Address, d time.Duration) ([]*types.Message, error) {
	var msgs []*types.Message
	var err error
	if addr != address.Undef {
		msgs, err = ms.repo.MessageRepo().ListBlockedMessage(addr, d)
	} else {
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
	}

	if len(msgs) > 0 {
		ids := make([]string, 0, len(msgs))
		for _, msg := range msgs {
			ids = append(ids, msg.ID)
		}
		ms.log.Infof("list blocked message by address %s %s", addr, strings.Join(ids, ","))
	}

	return msgs, err
}

func (ms *MessageService) UpdateMessageStateByCid(ctx context.Context, cid string, state types.MessageState) (string, error) {
	return cid, ms.repo.MessageRepo().UpdateMessageStateByCid(cid, state)
}

func (ms *MessageService) UpdateMessageStateByID(ctx context.Context, id string, state types.MessageState) error {
	return ms.repo.MessageRepo().UpdateMessageStateByID(id, state)
}

func (ms *MessageService) UpdateMessageInfoByCid(unsignedCid string, receipt *venusTypes.MessageReceipt,
	height abi.ChainEpoch, state types.MessageState, tsKey venusTypes.TipSetKey) (string, error) {
	return unsignedCid, ms.repo.MessageRepo().UpdateMessageInfoByCid(unsignedCid, receipt, height, state, tsKey)
}

func (ms *MessageService) ProcessNewHead(ctx context.Context, apply, revert []*venusTypes.TipSet) error {
	ms.log.Infof("receive new head from chain")
	if ms.fsRepo.Config().MessageService.SkipProcessHead {
		ms.log.Infof("skip process new head")
		return nil
	}

	if len(apply) == 0 {
		ms.log.Errorf("expect apply blocks, but got none")
		return nil
	}

	tsList := ms.tsCache.List()
	sort.Slice(tsList, func(i, j int) bool {
		return tsList[i].Height() > tsList[j].Height()
	})
	smallestTs := apply[len(apply)-1]

	defer ms.log.Infof("%d head wait to process", len(ms.headChans))

	if len(tsList) == 0 || smallestTs.Parents().Equals(tsList[0].Key()) {
		ms.log.Infof("apply a block height %d %s", apply[0].Height(), apply[0].String())
		done := make(chan error)
		ms.headChans <- &headChan{
			apply:  apply,
			revert: nil,
			done:   done,
		}
		return <-done
	}
	apply, revertTipset, err := ms.lookAncestors(ctx, tsList, smallestTs)
	if err != nil {
		ms.log.Errorf("look ancestor error from %s and %s, error: %v", smallestTs, tsList[0].Key(), err)
		return nil
	}

	done := make(chan error)
	ms.headChans <- &headChan{
		apply:  apply,
		revert: revertTipset,
		done:   done,
	}
	return <-done
}

func (ms *MessageService) ReconnectCheck(ctx context.Context, head *venusTypes.TipSet) error {
	ms.log.Infof("reconnect to node")

	if len(ms.tsCache.Cache) == 0 {
		count, err := ms.UpdateAllFilledMessage(ctx)
		if err != nil {
			return err
		}
		ms.log.Infof("update filled message count %v", count)
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
		ms.log.Infof("gap height %v, update filled message count %v", head.Height()-tsList[0].Height(), count)
		return nil
	}

	if tsList[0].Height() == head.Height() && tsList[0].Equals(head) {
		ms.log.Infof("The head does not change and returns directly.")
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
	tSelect := time.Now()
	selectResult, err := ms.messageSelector.SelectMessage(ctx, ts)
	if err != nil {
		return err
	}
	ms.log.Infof("current loop select result | SelectMsg: %d | ExpireMsg: %d | ToPushMsg: %d | ErrMsg: %d", len(selectResult.SelectMsg), len(selectResult.ExpireMsg), len(selectResult.ToPushMsg), len(selectResult.ErrMsg))
	tSaveDb := time.Now()
	ms.log.Infof("start to save to database")
	//save to db
	if err = ms.repo.Transaction(func(txRepo repo.TxRepo) error {
		//保存消息
		err = txRepo.MessageRepo().ExpireMessage(selectResult.ExpireMsg)
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
			ms.log.Infof("update message %s return value with error %s", m.id, m.err)
			err := txRepo.MessageRepo().UpdateReturnValue(m.id, m.err)
			if err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		ms.log.Errorf("save signed message failed %v", err)
		return err
	}

	ms.log.Infof("success to save to database")

	tCacheUpdate := time.Now()
	//update cache
	for _, msg := range selectResult.SelectMsg {
		selectResult.ToPushMsg = append(selectResult.ToPushMsg, &venusTypes.SignedMessage{
			Message:   msg.Message,
			Signature: *msg.Signature,
		})
		//update cache
		err := ms.messageState.MutatorMessage(msg.ID, func(message *types.Message) error {
			message.SignedCid = msg.SignedCid
			message.UnsignedCid = msg.UnsignedCid
			message.Message = msg.Message
			message.State = msg.State
			message.Signature = msg.Signature
			message.Nonce = msg.Nonce
			if message.Receipt != nil {
				message.Receipt.Return = nil //cover data for err before
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	for _, m := range selectResult.ErrMsg {
		err := ms.messageState.MutatorMessage(m.id, func(message *types.Message) error {
			if message.Receipt != nil {
				message.Receipt.Return = []byte(m.err)
			} else {
				message.Receipt = &venusTypes.MessageReceipt{Return: []byte(m.err)}
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	ms.log.Infof("success to update memory cache")

	//broad cast  push to node in config ,push to multi node in db config
	go func() {
		tPush := time.Now()
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

		ms.log.Infof("start to push message %d to mpool", len(selectResult.ToPushMsg))
		for addr, msgs := range pushMsgByAddr {
			//use batchpush instead of push one by one, push single may cause messsage send to different nodes when through chain-co
			//issue https://github.com/filecoin-project/venus/issues/4860
			if _, pushErr := ms.nodeClient.MpoolBatchPush(ctx, msgs); pushErr != nil {
				if !strings.Contains(pushErr.Error(), errMinimumNonce.Error()) && !strings.Contains(pushErr.Error(), errAlreadyInMpool.Error()) {
					ms.log.Errorf("push message in address %s to node failed %v", addr, pushErr)
				}
			}
		}

		ms.multiNodeToPush(ctx, pushMsgByAddr)

		ms.log.Infof("Push message select spent:%d , save db spent:%d ,update cache spent:%d, push to node spent: %d",
			time.Since(tSelect).Milliseconds(),
			time.Since(tSaveDb).Milliseconds(),
			time.Since(tCacheUpdate).Milliseconds(),
			time.Since(tPush).Milliseconds(),
		)
	}()
	return err
}

type nodeClient struct {
	name  string
	cli   v1.FullNode
	close jsonrpc.ClientCloser
}

func (ms *MessageService) multiNodeToPush(ctx context.Context, msgsByAddr map[address.Address][]*venusTypes.SignedMessage) {
	if len(msgsByAddr) == 0 {
		return
	}

	nodeList, err := ms.nodeService.ListNode(context.TODO())
	if err != nil {
		ms.log.Errorf("list node %v", err)
		return
	}

	nc := make([]nodeClient, 0, len(nodeList))
	for _, node := range nodeList {
		cli, closer, err := NewNodeClient(context.TODO(), &config.NodeConfig{Token: node.Token, Url: node.URL})
		if err != nil {
			ms.log.Warnf("connect node(%s) %v", node.Name, err)
			continue
		}
		nc = append(nc, nodeClient{name: node.Name, cli: cli, close: closer})
	}

	if len(nc) == 0 {
		ms.log.Infof("no available broadcast node config")
		return
	}

	for _, node := range nc {
		for addr, msgs := range msgsByAddr {
			//use batchpush instead of push one by one, push single may cause messsage send to different nodes when through chain-co
			//issue https://github.com/filecoin-project/venus/issues/4860
			if _, err := node.cli.MpoolBatchPush(ctx, msgs); err != nil {
				//skip error
				if !strings.Contains(err.Error(), errMinimumNonce.Error()) && !strings.Contains(err.Error(), errAlreadyInMpool.Error()) {
					ms.log.Errorf("push message from %s to node %s %v", addr, node.name, err)
				}
			}
		}
		ms.log.Infof("start to broadcast message of address")
		for fromAddr := range msgsByAddr {
			if err := node.cli.MpoolPublishByAddr(ctx, fromAddr); err != nil {
				ms.log.Errorf("publish message of address %s to node %s failed %v", fromAddr, node.name, err)
			}
		}

		node.close()
	}
}

func (ms *MessageService) StartPushMessage(ctx context.Context, skipPushMsg bool) {
	tm := time.NewTicker(time.Second * 30)
	defer tm.Stop()

	for {
		select {
		case <-ctx.Done():
			ms.log.Infof("Stop push message")
			return
		case <-tm.C:
			//newHead, err := ms.nodeClient.ChainHead(ctx)
			//if err != nil {
			//	ms.log.Errorf("fail to get chain head %v", err)
			//}
			//err = ms.pushMessageToPool(ctx, newHead)
			//if err != nil {
			//	ms.log.Errorf("push message error %v", err)
			//}
		case newHead := <-ms.triggerPush:
			// Clear all unfill messages by address
			ms.tryClearUnFillMsg()

			if skipPushMsg {
				ms.log.Info("skip push message")
				continue
			}
			start := time.Now()
			ms.log.Infof("start to push message %s task wait task %d", newHead.String(), len(ms.triggerPush))
			err := ms.pushMessageToPool(ctx, newHead)
			if err != nil {
				ms.log.Errorf("push message error %v", err)
			}
			ms.log.Infof("end push message spent %d ms", time.Since(start).Milliseconds())
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

	for addr := range ms.addressService.ActiveAddresses() {
		filledMsgs, err := ms.repo.MessageRepo().ListFilledMessageByAddress(addr)
		if err != nil {
			ms.log.Errorf("list filled message %v %v", addr, err)
			continue
		}
		msgs = append(msgs, filledMsgs...)
	}

	ms.log.Infof("%d messages need to sync", len(msgs))
	updateCount := 0
	for _, msg := range msgs {
		if err := ms.updateFilledMessage(ctx, msg); err != nil {
			ms.log.Errorf("update filled message: %v", err)
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
		ms.log.Infof("update message %v by node success, height: %d", msg.ID, msgLookup.Height)
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
			if maxFee.NilOrZero() {
				maxFee = ms.sps.GetParams().MaxFee
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

	signedMsg, err := ToSignedMsg(ctx, ms.walletClient, msg)
	if err != nil {
		return cid.Undef, err
	}

	if err := ms.repo.MessageRepo().SaveMessage(msg); err != nil {
		return cid.Undef, err
	}
	err = ms.messageState.MutatorMessage(msg.ID, func(message *types.Message) error {
		message.SignedCid = msg.SignedCid
		message.GasLimit = msg.GasLimit
		message.GasPremium = msg.GasPremium
		message.GasFeeCap = msg.GasFeeCap
		message.UnsignedCid = msg.UnsignedCid
		message.Message = msg.Message
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
	if msg.State == types.OnChainMsg {
		return fmt.Errorf("message already on chain")
	}
	if msg.State != types.FillMsg {
		return fmt.Errorf("need FillMsg got %s", msg.State)
	}
	signedMsg := &venusTypes.SignedMessage{
		Message:   msg.Message,
		Signature: *msg.Signature,
	}
	if _, err := ms.nodeClient.MpoolPush(ctx, signedMsg); err != nil {
		return err
	}
	toPush := make(map[address.Address][]*venusTypes.SignedMessage)
	toPush[signedMsg.Message.From] = []*venusTypes.SignedMessage{signedMsg}
	ms.multiNodeToPush(ctx, toPush)
	return nil
}

func ToSignedMsg(ctx context.Context, walletCli gateway.IWalletClient, msg *types.Message) (venusTypes.SignedMessage, error) {
	unsignedCid := msg.Message.Cid()
	msg.UnsignedCid = &unsignedCid
	//签名
	data, err := msg.Message.ToStorageBlock()
	if err != nil {
		return venusTypes.SignedMessage{}, fmt.Errorf("calc message unsigned message id %s fail %v", msg.ID, err)
	}
	sig, err := walletCli.WalletSign(ctx, msg.WalletName, msg.From, unsignedCid.Bytes(), venusTypes.MsgMeta{
		Type:  venusTypes.MTChainMsg,
		Extra: data.RawData(),
	})
	if err != nil {
		return venusTypes.SignedMessage{}, fmt.Errorf("wallet sign failed %s fail %v", msg.ID, err)
	}

	msg.Signature = sig
	//state
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
		ms.log.Infof("clear unfill messages success, address: %v, count: %d", addr.String(), r.count)
		return r.count, r.err
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}

const (
	ReplaceByFeeRatioDefault = 1.25
	RbfDenom                 = 256
)

var rbfNumBig = big.NewInt(int64((ReplaceByFeeRatioDefault - 1) * RbfDenom))
var rbfDenomBig = big.NewInt(RbfDenom)

func computeMinRBF(curPrem abi.TokenAmount) abi.TokenAmount {
	minPrice := big.Add(curPrem, big.Div(big.Mul(curPrem, rbfNumBig), rbfDenomBig))
	return big.Add(minPrice, big.NewInt(1))
}
