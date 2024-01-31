package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/filecoin-project/go-state-types/network"

	lru "github.com/hashicorp/golang-lru"

	"gorm.io/gorm"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"go.opencensus.io/tag"
	"go.uber.org/zap"
	"modernc.org/mathutil"

	"github.com/filecoin-project/venus/pkg/constants"
	"github.com/filecoin-project/venus/pkg/crypto"
	"github.com/filecoin-project/venus/venus-shared/actors/builtin"
	v1 "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	gatewayAPI "github.com/filecoin-project/venus/venus-shared/api/gateway/v2"
	venusTypes "github.com/filecoin-project/venus/venus-shared/types"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
	logging "github.com/ipfs/go-log/v2"

	"github.com/ipfs-force-community/sophon-messager/config"
	"github.com/ipfs-force-community/sophon-messager/metrics"
	"github.com/ipfs-force-community/sophon-messager/models/repo"
	"github.com/ipfs-force-community/sophon-messager/publisher"
	"github.com/ipfs-force-community/sophon-messager/utils"
)

const (
	gasEstimate = "gas estimate: "
)

var msgSelectLog = logging.Logger("msg-select")

type MsgSelectMgr struct {
	ctx            context.Context
	repo           repo.Repo
	cfg            *config.MessageServiceConfig
	fullNode       v1.FullNode
	addressService *AddressService
	sps            *SharedParamsService
	walletClient   gatewayAPI.IWalletClient

	works       map[address.Address]*work
	msgReceiver publisher.MessageReceiver
	lk          sync.Mutex
}

func newMsgSelectMgr(ctx context.Context,
	repo repo.Repo,
	cfg *config.MessageServiceConfig,
	fullNode v1.FullNode,
	addressService *AddressService,
	sps *SharedParamsService,
	walletClient gatewayAPI.IWalletClient,
	msgReceiver publisher.MessageReceiver,
) (*MsgSelectMgr, error) {
	ms := &MsgSelectMgr{
		ctx:            ctx,
		repo:           repo,
		cfg:            cfg,
		fullNode:       fullNode,
		addressService: addressService,
		sps:            sps,
		walletClient:   walletClient,

		msgReceiver: msgReceiver,
		works:       make(map[address.Address]*work),
	}

	addrInfos, err := ms.addressService.ListActiveAddress(ctx)
	if err != nil {
		return nil, err
	}

	return ms, ms.tryUpdateWorks(addressMap(addrInfos))
}

// SelectMessage not concurrency safe
func (msgSelectMgr *MsgSelectMgr) SelectMessage(ctx context.Context, ts *venusTypes.TipSet) error {
	sharedParams, err := msgSelectMgr.sps.GetSharedParams(ctx)
	if err != nil {
		return err
	}

	activeAddrs, err := msgSelectMgr.addressService.ListActiveAddress(ctx)
	if err != nil {
		return err
	}
	addrSelMsgNum := addrSelectMsgNum(activeAddrs, sharedParams.SelMsgNum)
	addrInfos := addressMap(activeAddrs)

	msgSelectMgr.lk.Lock()
	defer msgSelectMgr.lk.Unlock()

	if err := msgSelectMgr.tryUpdateWorks(addrInfos); err != nil {
		msgSelectLog.Warnf("failed to update work %v", err)
	}

	appliedNonce, err := msgSelectMgr.getNonceInTipset(ctx, ts)
	if err != nil {
		return err
	}

	for _, w := range msgSelectMgr.works {
		go w.startSelectMessage(appliedNonce, addrInfos[w.addr], ts, addrSelMsgNum[w.addr], sharedParams)
	}

	return nil
}

func (msgSelectMgr *MsgSelectMgr) getNonceInTipset(ctx context.Context, ts *venusTypes.TipSet) (*utils.NonceMap, error) {
	applied := utils.NewNonceMap()
	selectMsg := func(m *venusTypes.Message) error {
		// The first match for a sender is guaranteed to have correct nonce -- the block isn't valid otherwise
		if _, ok := applied.Get(m.From); !ok {
			applied.Add(m.From, m.Nonce)
		}

		val, _ := applied.Get(m.From)
		if val != m.Nonce {
			return nil
		}
		val++
		applied.Add(m.From, val)
		return nil
	}

	msgs, err := msgSelectMgr.fullNode.ChainGetMessagesInTipset(ctx, ts.Key())
	if err != nil {
		return nil, fmt.Errorf("failed to get message in tipset %v", err)
	}
	for _, msg := range msgs {
		err := selectMsg(msg.Message)
		if err != nil {
			return nil, fmt.Errorf("failed to decide whether to select message for block: %w", err)
		}
	}

	return applied, nil
}

func (msgSelectMgr *MsgSelectMgr) tryUpdateWorks(addrInfos map[address.Address]*types.Address) error {
	ws := make(map[address.Address]*work, len(addrInfos))
	for _, addrInfo := range addrInfos {
		w, ok := msgSelectMgr.works[addrInfo.Addr]
		if !ok {
			msgSelectLog.Infof("add a work %v", addrInfo.Addr)
			ws[addrInfo.Addr] = newWork(msgSelectMgr.ctx, addrInfo.Addr, msgSelectMgr.cfg, msgSelectMgr.fullNode, msgSelectMgr.repo, msgSelectMgr.addressService, msgSelectMgr.walletClient, msgSelectMgr.msgReceiver)
		} else {
			ws[addrInfo.Addr] = w
			delete(msgSelectMgr.works, addrInfo.Addr)
		}
	}
	for addr, w := range msgSelectMgr.works {
		if _, ok := ws[addr]; !ok {
			select {
			case w.controlChan <- struct{}{}:
				w.close()
				delete(msgSelectMgr.works, addr)
				msgSelectLog.Infof("remove a work %v", addr)
			default:
				ws[addr] = w
			}
		}
	}
	msgSelectMgr.works = ws

	return nil
}

func addressMap(addrList []*types.Address) map[address.Address]*types.Address {
	addrs := make(map[address.Address]*types.Address, len(addrList))
	for i, addr := range addrList {
		addrs[addr.Addr] = addrList[i]
	}

	return addrs
}

func addrSelectMsgNum(addrList []*types.Address, defSelMsgNum uint64) map[address.Address]uint64 {
	selMsgNum := make(map[address.Address]uint64)
	for _, addr := range addrList {
		if num, ok := selMsgNum[addr.Addr]; ok && addr.SelMsgNum > 0 && num < addr.SelMsgNum {
			selMsgNum[addr.Addr] = addr.SelMsgNum
		} else if !ok {
			if addr.SelMsgNum == 0 {
				selMsgNum[addr.Addr] = defSelMsgNum
			} else {
				selMsgNum[addr.Addr] = addr.SelMsgNum
			}
		}
	}

	return selMsgNum
}

func recordMetric(ctx context.Context, addr address.Address, selectResult *MsgSelectResult) {
	ctx, _ = tag.New(ctx, tag.Upsert(metrics.WalletAddress, addr.String()))

	metrics.SelectedMsgNumOfLastRound.Set(ctx, int64(len(selectResult.SelectMsg)))
	metrics.ToPushMsgNumOfLastRound.Set(ctx, int64(len(selectResult.ToPushMsg)))
	metrics.ErrMsgNumOfLastRound.Set(ctx, int64(len(selectResult.ErrMsg)))
}

var errSingMessage = errors.New("sign message failed")

type MsgSelectResult struct {
	Address   *types.Address
	SelectMsg []*types.Message
	ToPushMsg []*venusTypes.SignedMessage
	ErrMsg    []msgErrInfo
}

type msgErrInfo struct {
	id  string
	err string
}

type work struct {
	ctx    context.Context
	cancel context.CancelFunc

	addr           address.Address
	cfg            *config.MessageServiceConfig
	fullNode       v1.FullNode
	repo           repo.Repo
	addressService *AddressService
	walletClient   gatewayAPI.IWalletClient
	msgReceiver    publisher.MessageReceiver

	start       time.Time
	controlChan chan struct{}

	actorCache *lru.ARCCache

	log *zap.SugaredLogger
}

func newWork(ctx context.Context,
	addr address.Address,
	cfg *config.MessageServiceConfig,
	fullNode v1.FullNode,
	repo repo.Repo,
	addressService *AddressService,
	walletClient gatewayAPI.IWalletClient,
	msgReceiver publisher.MessageReceiver,
) *work {
	ctx, cancel := context.WithCancel(ctx)
	cache, _ := lru.NewARC(100)
	return &work{
		ctx:            ctx,
		cancel:         cancel,
		addr:           addr,
		cfg:            cfg,
		addressService: addressService,
		fullNode:       fullNode,
		repo:           repo,
		walletClient:   walletClient,
		msgReceiver:    msgReceiver,
		controlChan:    make(chan struct{}, 1),
		actorCache:     cache,
		log:            msgSelectLog.With("address", addr),
	}
}

func (w *work) startSelectMessage(
	appliedNonce *utils.NonceMap,
	addrInfo *types.Address,
	ts *venusTypes.TipSet,
	maxAllowPendingMessage uint64,
	sharedParams *types.SharedSpec,
) {
	// first check w.ctx, avoid w.controlChan closed
	select {
	case <-w.ctx.Done():
		w.log.Infof("context done: %s, skip select message", w.ctx.Err())
		return
	default:
	}

	select {
	case w.controlChan <- struct{}{}:
	default:
		w.log.Infof("already selecting message, had took %v", time.Since(w.start))
		return
	}

	w.start = time.Now()
	ctx, cancel := context.WithTimeout(w.ctx, w.cfg.SignMessageTimeout+w.cfg.EstimateMessageTimeout)
	defer w.finish()
	defer cancel()

	selectResult, err := w.selectMessage(ctx, appliedNonce, addrInfo, ts, maxAllowPendingMessage, sharedParams)
	if err != nil {
		w.log.Errorf("select message failed: %v", err)
		return
	}
	w.log.Infof("select message result | SelectMsg: %d | ToPushMsg: %d | ErrMsg: %d | took: %v", len(selectResult.SelectMsg),
		len(selectResult.ToPushMsg), len(selectResult.ErrMsg), time.Since(w.start))

	recordMetric(ctx, w.addr, selectResult)

	if err := w.saveSelectedMessages(ctx, selectResult); err != nil {
		w.log.Errorf("failed to save selected messages to db %v", err)
		return
	}

	for _, msg := range selectResult.SelectMsg {
		selectResult.ToPushMsg = append(selectResult.ToPushMsg, &venusTypes.SignedMessage{
			Message:   msg.Message,
			Signature: *msg.Signature,
		})
	}

	if len(selectResult.ToPushMsg) > 0 {
		// send messages to push
		select {
		case w.msgReceiver <- selectResult.ToPushMsg:
		default:
			w.log.Errorf("message receiver channel is full, skip %d messages", len(selectResult.ToPushMsg))
		}
	}
}

func (w *work) selectMessage(ctx context.Context, appliedNonce *utils.NonceMap, addrInfo *types.Address, ts *venusTypes.TipSet, maxAllowPendingMessage uint64, sharedParams *types.SharedSpec) (*MsgSelectResult, error) {
	// 没有绑定账号肯定无法签名
	accounts, err := w.addressService.GetAccountsOfSigner(ctx, addrInfo.Addr)
	if err != nil {
		return nil, fmt.Errorf("get account failed: %v", err)
	}

	// 判断是否需要推送消息
	nonceInLatestTs, actorNonce, err := w.getNonce(ctx, ts, appliedNonce)
	if err != nil {
		return nil, err
	}
	if nonceInLatestTs > addrInfo.Nonce {
		w.log.Warnf("nonce in db %d is smaller than nonce on chain %d, update to latest", addrInfo.Nonce, nonceInLatestTs)
		addrInfo.Nonce = nonceInLatestTs
		addrInfo.UpdatedAt = time.Now()
		err := w.repo.AddressRepo().UpdateNonce(ctx, addrInfo.Addr, addrInfo.Nonce)
		if err != nil {
			return nil, fmt.Errorf("update nonce failed: %v", err)
		}
	}

	toPushMessage := w.getFilledMessage(nonceInLatestTs)

	// calc the message needed
	nonceGap := addrInfo.Nonce - nonceInLatestTs
	if nonceGap >= maxAllowPendingMessage {
		w.log.Errorf("there are %d message not to be package, nonce gap: %d", len(toPushMessage), nonceGap)
		return &MsgSelectResult{
			ToPushMsg: toPushMessage,
			Address:   addrInfo,
		}, nil
	}
	wantCount := maxAllowPendingMessage - nonceGap
	w.log.Infof("state actor nonce %d, latest nonce in ts %d, assigned nonce %d, nonce gap %d, want %d", actorNonce, nonceInLatestTs, addrInfo.Nonce, nonceGap, wantCount)

	// get unfill message
	selectCount := mathutil.MinUint64(wantCount, 100)
	messages, err := w.repo.MessageRepo().ListUnChainMessageByAddress(addrInfo.Addr, int(selectCount))
	if err != nil {
		return nil, fmt.Errorf("list unfill message error: %v", err)
	}

	if len(messages) == 0 {
		w.log.Infof("have no unfill message")
		return &MsgSelectResult{
			ToPushMsg: toPushMessage,
			Address:   addrInfo,
		}, nil
	}

	var errMsg []msgErrInfo
	count := uint64(0)
	selectMsg := make([]*types.Message, 0, len(messages))

	estimateResult, candidateMessages, err := w.estimateMessage(ctx, ts, messages, sharedParams, addrInfo)
	if err != nil {
		return nil, fmt.Errorf("estimate message failed: %v", err)
	}

	// sign
	for index, msg := range candidateMessages {
		// if error print error message
		if len(estimateResult[index].Err) != 0 {
			errMsg = append(errMsg, msgErrInfo{id: msg.ID, err: gasEstimate + estimateResult[index].Err})
			w.log.Errorf("estimate message %s fail %s", msg.ID, estimateResult[index].Err)
			continue
		}
		estimateMsg := estimateResult[index].Msg
		if count >= wantCount {
			break
		}

		// 检查 gaslimit 是否超出上限
		if estimateMsg.GasLimit > constants.BlockGasLimit {
			err := fmt.Sprintf("%s gas limit %d over limit %d", gasEstimate, estimateMsg.GasLimit, constants.BlockGasLimit)
			errMsg = append(errMsg, msgErrInfo{id: msg.ID, err: err})
			w.log.Errorf(err)
			continue
		}

		// 分配nonce
		msg.Nonce = addrInfo.Nonce
		msg.GasFeeCap = estimateMsg.GasFeeCap
		msg.GasPremium = estimateMsg.GasPremium
		msg.GasLimit = estimateMsg.GasLimit

		unsignedCid := msg.Message.Cid()
		msg.UnsignedCid = &unsignedCid

		// 签名
		sig, err := w.signMessage(ctx, msg, accounts)
		if err != nil {
			if strings.Contains(err.Error(), errSingMessage.Error()) {
				errMsg = append(errMsg, msgErrInfo{id: msg.ID, err: err.Error()})
				w.log.Errorf("msg: %v, error: %v", msg.ID, err)
				break
			}
			w.log.Errorf("msg: %v, error: %v", msg.ID, err)
			continue
		}

		msg.Signature = sig
		msg.State = types.FillMsg

		// signed cid for t1 address
		signedMsg := venusTypes.SignedMessage{
			Message:   msg.Message,
			Signature: *msg.Signature,
		}
		signedCid := signedMsg.Cid()
		msg.SignedCid = &signedCid

		// remove errored info during the previous message selection
		if len(msg.ErrorMsg) > 0 {
			msg.ErrorMsg = ""
		}

		selectMsg = append(selectMsg, msg)
		addrInfo.Nonce++
		count++
	}

	return &MsgSelectResult{
		SelectMsg: selectMsg,
		ToPushMsg: toPushMessage,
		Address:   addrInfo,
		ErrMsg:    errMsg,
	}, nil
}

func (w *work) getNonce(ctx context.Context, ts *venusTypes.TipSet, appliedNonce *utils.NonceMap) (uint64, uint64, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, w.cfg.DefaultTimeout)
	defer cancel()
	actorI, err := handleTimeout(timeoutCtx, w.fullNode.StateGetActor, []interface{}{w.addr, ts.Key()})
	if err != nil {
		return 0, 0, fmt.Errorf("get actor failed: %v", err)
	}
	actor := actorI.(*venusTypes.Actor)
	nonceInLatestTs := actor.Nonce
	// todo actor nonce maybe the latest ts. not need appliedNonce
	if nonceInTs, ok := appliedNonce.Get(w.addr); ok {
		w.log.Infof("nonce in ts %d, nonce in actor %d", nonceInTs, nonceInLatestTs)
		nonceInLatestTs = nonceInTs
	}

	return nonceInLatestTs, actor.Nonce, nil
}

func (w *work) getFilledMessage(nonceInLatestTs uint64) []*venusTypes.SignedMessage {
	filledMessage, err := w.repo.MessageRepo().ListFilledMessageByAddress(w.addr)
	if err != nil {
		w.log.Warnf("list filled message %v", err)
	}
	msgs := make([]*venusTypes.SignedMessage, 0, len(filledMessage))
	for _, msg := range filledMessage {
		if nonceInLatestTs > msg.Nonce {
			continue
		}
		msgs = append(msgs, &venusTypes.SignedMessage{
			Message:   msg.Message,
			Signature: *msg.Signature,
		})
	}

	return msgs
}

func (w *work) estimateMessage(ctx context.Context,
	ts *venusTypes.TipSet,
	msgs []*types.Message,
	sharedParams *types.SharedSpec,
	addrInfo *types.Address,
) ([]*venusTypes.EstimateResult, []*types.Message, error) {
	candidateMessages := make([]*types.Message, 0, len(msgs))
	estimateMessages := make([]*venusTypes.EstimateMessage, 0, len(msgs))

	nv, err := w.fullNode.StateNetworkVersion(ctx, venusTypes.EmptyTSK)
	if err != nil {
		return nil, nil, fmt.Errorf("get network version failed: %v", err)
	}
	for _, msg := range msgs {
		actorCfg, err := w.getActorCfg(ctx, msg, nv)
		if err != nil {
			return nil, nil, fmt.Errorf("get actor config failed: %v", err)
		}
		newMsgMeta := mergeMsgSpec(sharedParams, msg.Meta, addrInfo, actorCfg, msg)

		if msg.GasFeeCap.NilOrZero() && !newMsgMeta.GasFeeCap.NilOrZero() {
			msg.GasFeeCap = newMsgMeta.GasFeeCap
		}

		baseFee := ts.At(0).ParentBaseFee
		if !newMsgMeta.BaseFee.NilOrZero() && baseFee.GreaterThan(newMsgMeta.BaseFee) {
			w.log.Infof("skip msg %v, base fee too height %v(local) < %v(chain), height %v", msg.ID, newMsgMeta.BaseFee, baseFee, ts.Height())
			continue
		}

		candidateMessages = append(candidateMessages, msg)
		estimateMessages = append(estimateMessages, &venusTypes.EstimateMessage{
			Msg: &msg.Message,
			Spec: &venusTypes.MessageSendSpec{
				MaxFee:            newMsgMeta.MaxFee,
				GasOverEstimation: newMsgMeta.GasOverEstimation,
				GasOverPremium:    newMsgMeta.GasOverPremium,
			},
		})

		w.log.Infof("estimate message %s, gas fee cap %s, gas limit %v, gas premium: %s, "+
			"meta maxfee %s, over estimation %f, gas over premium %f, gas fee cap %v", msg.ID, msg.GasFeeCap, msg.GasLimit, msg.GasPremium,
			newMsgMeta.MaxFee, newMsgMeta.GasOverEstimation, newMsgMeta.GasOverPremium, newMsgMeta.GasFeeCap)
	}

	estimateMsgCtx, estimateMsgCancel := context.WithTimeout(ctx, w.cfg.EstimateMessageTimeout)
	defer estimateMsgCancel()

	estimateResult, err := w.fullNode.GasBatchEstimateMessageGas(estimateMsgCtx, estimateMessages, addrInfo.Nonce, ts.Key())

	return estimateResult, candidateMessages, err
}

func (w *work) signMessage(ctx context.Context, msg *types.Message, accounts []string) (*crypto.Signature, error) {
	data, err := msg.Message.ToStorageBlock()
	if err != nil {
		return nil, fmt.Errorf("serialize message failed: %v", err)
	}

	sb, err := msg.Message.SigningBytes(venusTypes.AddressProtocol2SignType(msg.Message.From.Protocol()))
	if err != nil {
		return nil, fmt.Errorf("get signing bytes failed: %v", err)
	}

	signMsgCtx, signMsgCancel := context.WithTimeout(ctx, w.cfg.SignMessageTimeout)
	sigI, err := handleTimeout(signMsgCtx, w.walletClient.WalletSign, []interface{}{w.addr, accounts, sb, venusTypes.MsgMeta{
		Type:  venusTypes.MTChainMsg,
		Extra: data.RawData(),
	}})
	signMsgCancel()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errSingMessage.Error(), err)
	}

	return sigI.(*crypto.Signature), nil
}

func (w *work) saveSelectedMessages(ctx context.Context, selectResult *MsgSelectResult) error {
	startSaveDB := time.Now()
	w.log.Infof("start save messages to database")
	err := w.repo.Transaction(func(txRepo repo.TxRepo) error {
		if len(selectResult.SelectMsg) > 0 {
			if err := txRepo.MessageRepo().BatchSaveMessage(selectResult.SelectMsg); err != nil {
				return err
			}

			addrInfo := selectResult.Address
			if err := txRepo.AddressRepo().UpdateNonce(ctx, addrInfo.Addr, addrInfo.Nonce); err != nil {
				return err
			}
		}

		for _, m := range selectResult.ErrMsg {
			w.log.Infof("update message %s error info with error %s", m.id, m.err)
			if err := txRepo.MessageRepo().UpdateErrMsg(m.id, m.err); err != nil {
				return err
			}
		}
		return nil
	})
	w.log.Infof("end save messages to database, took %v, err %v", time.Since(startSaveDB), err)

	return err
}

func (w *work) getActorCfg(ctx context.Context, msg *types.Message, nv network.Version) (*types.ActorCfg, error) {
	key := msg.To.String() + "-" + strconv.Itoa(int(nv))
	var actor *venusTypes.Actor
	actorI, has := w.actorCache.Get(key)
	if has {
		actor = actorI.(*venusTypes.Actor)
	} else {
		var err error
		actor, err = w.fullNode.StateGetActor(ctx, msg.To, venusTypes.EmptyTSK)
		if err != nil {
			// msg.To may be a new address that needs to be created
			if msg.Message.Method == builtin.MethodSend &&
				strings.Contains(err.Error(), venusTypes.ErrActorNotFound.Error()) {
				return nil, nil
			}
			return nil, err
		}
		w.actorCache.Add(key, actor)
	}

	mt := &types.MethodType{
		Code:   actor.Code,
		Method: msg.Method,
	}

	// https://github.com/filecoin-project/venus/issues/5939
	has, err := w.repo.ActorCfgRepo().HasActorCfg(ctx, mt)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}

	actorCfg, err := w.repo.ActorCfgRepo().GetActorCfgByMethodType(ctx, mt)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return actorCfg, nil
}

func (w *work) finish() {
	<-w.controlChan
}

func (w *work) close() {
	w.cancel()
	close(w.controlChan)
}

func CapGasFee(msg *venusTypes.Message, maxFee abi.TokenAmount) {
	if maxFee.NilOrZero() {
		return
	}

	gl := venusTypes.NewInt(uint64(msg.GasLimit))
	totalFee := venusTypes.BigMul(msg.GasFeeCap, gl)

	if totalFee.LessThanEqual(maxFee) {
		msg.GasPremium = big.Min(msg.GasFeeCap, msg.GasPremium) // cap premium at FeeCap
		return
	}

	msg.GasFeeCap = big.Div(maxFee, gl)
	msg.GasPremium = big.Min(msg.GasFeeCap, msg.GasPremium) // cap premium at FeeCap
}

type GasSpec struct {
	GasOverEstimation float64
	MaxFee            big.Int
	GasOverPremium    float64
	GasFeeCap         big.Int
	BaseFee           big.Int
}

func mergeMsgSpec(globalSpec *types.SharedSpec, sendSpec *types.SendSpec, addrInfo *types.Address, actorCfg *types.ActorCfg, msg *types.Message) *GasSpec {
	newMsgMeta := &GasSpec{
		GasOverEstimation: sendSpec.GasOverEstimation,
		GasOverPremium:    sendSpec.GasOverPremium,
		MaxFee:            sendSpec.MaxFee,
	}

	//msg > addr > actor > global
	if sendSpec.GasOverEstimation != 0 {
		newMsgMeta.GasOverEstimation = sendSpec.GasOverEstimation
	} else if addrInfo.GasOverEstimation != 0 {
		newMsgMeta.GasOverEstimation = addrInfo.GasOverEstimation
	} else if actorCfg != nil && actorCfg.GasOverEstimation != 0 {
		newMsgMeta.GasOverEstimation = actorCfg.GasOverEstimation
	} else if globalSpec != nil {
		newMsgMeta.GasOverEstimation = globalSpec.GasOverEstimation
	}

	if !sendSpec.MaxFee.NilOrZero() {
		newMsgMeta.MaxFee = sendSpec.MaxFee
	} else if !addrInfo.MaxFee.NilOrZero() {
		newMsgMeta.MaxFee = addrInfo.MaxFee
	} else if actorCfg != nil && !actorCfg.MaxFee.NilOrZero() {
		newMsgMeta.MaxFee = actorCfg.MaxFee
	} else if globalSpec != nil {
		newMsgMeta.MaxFee = globalSpec.MaxFee
	}

	if sendSpec.GasOverPremium != 0 {
		newMsgMeta.GasOverPremium = sendSpec.GasOverPremium
	} else if addrInfo.GasOverPremium != 0 {
		newMsgMeta.GasOverPremium = addrInfo.GasOverPremium
	} else if actorCfg != nil && actorCfg.GasOverPremium != 0 {
		newMsgMeta.GasOverPremium = actorCfg.GasOverPremium
	} else if globalSpec.GasOverPremium != 0 {
		newMsgMeta.GasOverPremium = globalSpec.GasOverPremium
	}

	if msg.GasFeeCap.NilOrZero() {
		if !addrInfo.GasFeeCap.NilOrZero() {
			newMsgMeta.GasFeeCap = addrInfo.GasFeeCap
		} else if actorCfg != nil && !actorCfg.GasFeeCap.NilOrZero() {
			newMsgMeta.GasFeeCap = actorCfg.GasFeeCap
		} else if globalSpec != nil {
			newMsgMeta.GasFeeCap = globalSpec.GasFeeCap
		}
	}

	if !addrInfo.BaseFee.NilOrZero() {
		newMsgMeta.BaseFee = addrInfo.BaseFee
	} else if actorCfg != nil && !actorCfg.BaseFee.NilOrZero() {
		newMsgMeta.BaseFee = actorCfg.BaseFee
	} else if globalSpec != nil {
		newMsgMeta.BaseFee = globalSpec.BaseFee
	}

	return newMsgMeta
}
