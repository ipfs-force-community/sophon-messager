package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/venus/pkg/constants"
	"github.com/filecoin-project/venus/venus-shared/testutil"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
	"github.com/stretchr/testify/assert"
	"go.uber.org/fx/fxtest"

	"github.com/filecoin-project/venus-messager/config"
	"github.com/filecoin-project/venus-messager/filestore"
	"github.com/filecoin-project/venus-messager/gateway"
	"github.com/filecoin-project/venus-messager/log"
	"github.com/filecoin-project/venus-messager/models"
	"github.com/filecoin-project/venus-messager/pubsub"
	"github.com/filecoin-project/venus-messager/testhelper"
	"github.com/filecoin-project/venus-messager/utils"
)

const defaultLocalToken = "defaultLocalToken"

func TestMergeMsgSpec(t *testing.T) {
	defParams := &Params{
		SharedSpec: DefSharedParams,
	}
	defParams.GasOverPremium = 1.0
	defParams.GasFeeCap = big.NewInt(10000)

	sendSpec := &types.SendSpec{
		GasOverEstimation: 1.4,
		MaxFee:            big.NewInt(40000),
		GasOverPremium:    4.0,
	}
	emptySendSpec := &types.SendSpec{}

	addrInfo := &types.Address{
		GasOverEstimation: 1.5,
		MaxFee:            big.NewInt(50000),
		GasFeeCap:         big.NewInt(50000),
		GasOverPremium:    5.0,
	}
	emptyAddrInfo := &types.Address{}

	msg := testhelper.NewMessage()
	msg2 := testhelper.NewMessage()
	msg2.GasFeeCap = testhelper.DefGasFeeCap

	tests := []struct {
		params   *Params
		sendSpec *types.SendSpec
		addrInfo *types.Address
		msg      *types.Message

		expect *GasSpec
	}{
		{
			defParams,
			sendSpec,
			addrInfo,
			msg,
			&GasSpec{GasOverEstimation: sendSpec.GasOverEstimation, MaxFee: sendSpec.MaxFee, GasOverPremium: sendSpec.GasOverPremium, GasFeeCap: addrInfo.GasFeeCap},
		},
		{
			defParams,
			emptySendSpec,
			addrInfo,
			msg,
			&GasSpec{GasOverEstimation: addrInfo.GasOverEstimation, MaxFee: addrInfo.MaxFee, GasOverPremium: addrInfo.GasOverPremium, GasFeeCap: addrInfo.GasFeeCap},
		},
		{
			defParams,
			emptySendSpec,
			emptyAddrInfo,
			msg,
			&GasSpec{GasOverEstimation: defParams.GasOverEstimation, MaxFee: defParams.MaxFee, GasOverPremium: defParams.GasOverPremium, GasFeeCap: defParams.GasFeeCap},
		},
		{
			defParams,
			emptySendSpec,
			addrInfo,
			msg2,
			&GasSpec{GasOverEstimation: addrInfo.GasOverEstimation, MaxFee: addrInfo.MaxFee, GasOverPremium: addrInfo.GasOverPremium},
		},
		{
			defParams,
			emptySendSpec,
			emptyAddrInfo,
			msg2,
			&GasSpec{GasOverEstimation: defParams.GasOverEstimation, MaxFee: defParams.MaxFee, GasOverPremium: defParams.GasOverPremium},
		},
	}

	for _, test := range tests {
		gasSpec := mergeMsgSpec(test.params, test.sendSpec, test.addrInfo, test.msg)
		assert.Equal(t, test.expect, gasSpec)
	}
}

func TestAddrSelectMsgNum(t *testing.T) {
	log := log.New()
	ctx := context.Background()
	fsRepo := filestore.NewMockFileStore(t.TempDir())
	repo, err := models.SetDataBase(fsRepo)
	assert.NoError(t, err)
	assert.NoError(t, repo.AutoMigrate())

	sps, err := NewSharedParamsService(ctx, repo, log)
	assert.NoError(t, err)
	msgSelector := &MessageSelector{
		sps: sps,
	}

	addr := testutil.IDAddressProvider()(t)
	addr2 := testutil.IDAddressProvider()(t)
	addrList := []*types.Address{
		{
			Addr:      addr,
			SelMsgNum: 10,
		},
		{
			Addr:      addr,
			SelMsgNum: 4,
		},
		{
			Addr: addr2,
		},
	}
	expect := map[address.Address]uint64{
		addr:  10,
		addr2: sps.GetParams().SelMsgNum,
	}

	addrNum := msgSelector.addrSelectMsgNum(addrList)

	for _, addrInfo := range addrList {
		num, ok := addrNum[addrInfo.Addr]
		assert.True(t, ok)
		expectNum, ok := expect[addrInfo.Addr]
		assert.True(t, ok)
		assert.Equal(t, expectNum, num)
	}
}

func TestSelectMessage(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.DefaultConfig()
	cfg.MessageService.SkipPushMessage = true
	cfg.MessageService.WaitingChainHeadStableDuration = time.Second * 2
	blockDelay := cfg.MessageService.WaitingChainHeadStableDuration * 2
	fsRepo := filestore.NewMockFileStore(t.TempDir())
	msh, err := newMessageServiceHelper(ctx, cfg, blockDelay, fsRepo)
	assert.NoError(t, err)
	ms := msh.ms

	account := defaultLocalToken
	addrCount := 10
	addrs := testhelper.ResolveAddrs(t, testhelper.RandAddresses(t, addrCount))
	assert.NoError(t, msh.walletProxy.AddAddress(account, addrs))
	assert.NoError(t, msh.fullNode.AddActors(addrs))

	lc := fxtest.NewLifecycle(t)
	_ = StartNodeEvents(lc, msh.fullNode, msh.ms, ms.log)
	assert.NoError(t, lc.Start(ctx))
	defer lc.RequireStop()

	ts, err := msh.fullNode.ChainHead(ctx)
	assert.NoError(t, err)
	res, err := ms.messageSelector.SelectMessage(ctx, ts)
	assert.NoError(t, err)
	assert.Equal(t, &MsgSelectResult{}, res)

	totalMsg := len(addrs) * 10
	msgs := genMessages(addrs, account, totalMsg)
	assert.NoError(t, pushMessage(ctx, ms, msgs))

	ts, err = msh.fullNode.ChainHead(ctx)
	assert.NoError(t, err)
	selectResult, err := ms.messageSelector.SelectMessage(ctx, ts)
	assert.NoError(t, err)
	assert.Len(t, selectResult.SelectMsg, totalMsg)
	assert.Len(t, selectResult.ErrMsg, 0)
	assert.Len(t, selectResult.ModifyAddress, len(addrs))
	assert.Len(t, selectResult.ExpireMsg, 0)
	assert.Len(t, selectResult.ToPushMsg, 0)
	testhelper.IsSortedByNonce(t, selectResult.SelectMsg)
	assert.NoError(t, saveAndPushMsgs(ctx, ms, selectResult))

	checkMsgs(ctx, t, ms, msgs, selectResult.SelectMsg)
}

func TestSelectNum(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.DefaultConfig()
	cfg.MessageService.SkipPushMessage = true
	cfg.MessageService.WaitingChainHeadStableDuration = time.Second * 2
	blockDelay := cfg.MessageService.WaitingChainHeadStableDuration * 2
	fsRepo := filestore.NewMockFileStore(t.TempDir())
	msh, err := newMessageServiceHelper(ctx, cfg, blockDelay, fsRepo)
	assert.NoError(t, err)
	ms := msh.ms

	account := defaultLocalToken
	addrCount := 10
	addrs := testhelper.ResolveAddrs(t, testhelper.RandAddresses(t, addrCount))
	assert.NoError(t, msh.walletProxy.AddAddress(account, addrs))
	assert.NoError(t, msh.fullNode.AddActors(addrs))

	lc := fxtest.NewLifecycle(t)
	_ = StartNodeEvents(lc, msh.fullNode, ms, ms.log)
	assert.NoError(t, lc.Start(ctx))
	defer lc.RequireStop()

	defSelectedNum := int(DefSharedParams.SelMsgNum)
	totalMsg := len(addrs) * 50
	msgs := genMessages(addrs, account, totalMsg)
	assert.NoError(t, pushMessage(ctx, ms, msgs))

	checkSelectNum := func(msgs []*types.Message, addrNum map[address.Address]int, defNum int) {
		addrMsgs := utils.MsgGroupByAddress(msgs)
		for addr, m := range addrMsgs {
			num, ok := addrNum[addr]
			if ok {
				assert.Len(t, m, num)
			} else {
				assert.Len(t, m, defNum)
			}
		}
	}

	ts, err := msh.fullNode.ChainHead(ctx)
	assert.NoError(t, err)
	selectResult, err := ms.messageSelector.SelectMessage(ctx, ts)
	assert.NoError(t, err)
	assert.Len(t, selectResult.SelectMsg, len(addrs)*defSelectedNum)
	assert.Len(t, selectResult.ErrMsg, 0)
	assert.Len(t, selectResult.ModifyAddress, len(addrs))
	assert.Len(t, selectResult.ExpireMsg, 0)
	assert.Len(t, selectResult.ToPushMsg, 0)
	assert.NoError(t, saveAndPushMsgs(ctx, ms, selectResult))
	testhelper.IsSortedByNonce(t, selectResult.SelectMsg)
	checkSelectNum(selectResult.SelectMsg, map[address.Address]int{}, defSelectedNum)
	checkMsgs(ctx, t, ms, msgs, selectResult.SelectMsg)

	addrNum := make(map[address.Address]int, len(addrs))
	expectNum := 0
	for i, addr := range addrs {
		num := i + 5
		assert.NoError(t, ms.addressService.SetSelectMsgNum(ctx, addr, uint64(num)))
		addrNum[addr] = num
		expectNum += num
	}

	ts, err = msh.fullNode.ChainHead(ctx)
	assert.NoError(t, err)
	selectResult, err = ms.messageSelector.SelectMessage(ctx, ts)
	assert.NoError(t, err)
	assert.Len(t, selectResult.SelectMsg, expectNum)
	assert.Len(t, selectResult.ErrMsg, 0)
	assert.Len(t, selectResult.ModifyAddress, len(addrs))
	assert.Len(t, selectResult.ExpireMsg, 0)
	assert.Len(t, selectResult.ToPushMsg, 0)
	assert.NoError(t, saveAndPushMsgs(ctx, ms, selectResult))
	testhelper.IsSortedByNonce(t, selectResult.SelectMsg)
	checkSelectNum(selectResult.SelectMsg, addrNum, defSelectedNum)
	checkMsgs(ctx, t, ms, msgs, selectResult.SelectMsg)
}

func TestEstimateMessageGas(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.DefaultConfig()
	cfg.MessageService.SkipPushMessage = true
	cfg.MessageService.WaitingChainHeadStableDuration = time.Second * 2
	blockDelay := cfg.MessageService.WaitingChainHeadStableDuration * 2
	fsRepo := filestore.NewMockFileStore(t.TempDir())
	msh, err := newMessageServiceHelper(ctx, cfg, blockDelay, fsRepo)
	assert.NoError(t, err)
	ms := msh.ms

	account := defaultLocalToken
	addrCount := 10
	addrs := testhelper.ResolveAddrs(t, testhelper.RandAddresses(t, addrCount))
	assert.NoError(t, msh.walletProxy.AddAddress(account, addrs))
	assert.NoError(t, msh.fullNode.AddActors(addrs))

	lc := fxtest.NewLifecycle(t)
	_ = StartNodeEvents(lc, msh.fullNode, ms, ms.log)
	assert.NoError(t, lc.Start(ctx))
	defer lc.RequireStop()

	msgs := genMessages(addrs, defaultLocalToken, len(addrs)*10)
	for _, msg := range msgs {
		// will estimate gas failed
		msg.GasLimit = -1
	}
	assert.NoError(t, pushMessage(ctx, ms, msgs))
	msgsMap := testhelper.SliceToMap(msgs)

	ts, err := msh.fullNode.ChainHead(ctx)
	assert.NoError(t, err)
	selectResult, err := ms.messageSelector.SelectMessage(ctx, ts)
	assert.NoError(t, err)
	assert.Len(t, selectResult.SelectMsg, 0)
	assert.Len(t, selectResult.ErrMsg, len(msgs))
	assert.Len(t, selectResult.ModifyAddress, 0)
	assert.Len(t, selectResult.ExpireMsg, 0)
	assert.Len(t, selectResult.ToPushMsg, 0)
	testhelper.IsSortedByNonce(t, selectResult.SelectMsg)
	assert.NoError(t, saveAndPushMsgs(ctx, ms, selectResult))

	list, err := ms.ListFailedMessage(ctx)
	assert.NoError(t, err)
	for _, msg := range list {
		_, ok := msgsMap[msg.ID]
		assert.True(t, ok)
		assert.Contains(t, string(msg.Receipt.Return), testhelper.ErrGasLimitNegative.Error())

		assert.NoError(t, ms.MarkBadMessage(ctx, msg.ID))
		res, err := ms.GetMessageByUid(ctx, msg.ID)
		assert.NoError(t, err)
		assert.Equal(t, types.FailedMsg, res.State)
	}

	gasOverEstimation := 1.25
	gasOverPremium := 1.0
	for i, addr := range addrs {
		gasOverEstimation = float64(i) * gasOverEstimation
		gasOverPremium = float64(i) * gasOverPremium
		maxFeeStr := big.Mul(testhelper.DefMaxFee, big.NewInt(int64(i))).String()
		gasFeeCapStr := big.Mul(testhelper.DefGasFeeCap, big.NewInt(int64(i))).String()
		assert.NoError(t, ms.addressService.SetFeeParams(ctx, addr, gasOverEstimation, gasOverPremium, maxFeeStr, gasFeeCapStr))
	}

	msgs = genMessages(addrs, defaultLocalToken, len(addrs)*10)
	for _, msg := range msgs {
		// use the fee params in the address table
		msg.Meta = nil
	}
	assert.NoError(t, pushMessage(ctx, ms, msgs))

	ts, err = msh.fullNode.ChainHead(ctx)
	assert.NoError(t, err)
	selectResult, err = ms.messageSelector.SelectMessage(ctx, ts)
	assert.NoError(t, err)
	assert.Len(t, selectResult.SelectMsg, len(msgs))
	assert.Len(t, selectResult.ErrMsg, 0)
	assert.Len(t, selectResult.ModifyAddress, len(addrs))
	assert.Len(t, selectResult.ExpireMsg, 0)
	assert.Len(t, selectResult.ToPushMsg, 0)
	testhelper.IsSortedByNonce(t, selectResult.SelectMsg)
	assert.NoError(t, saveAndPushMsgs(ctx, ms, selectResult))
}

func TestSignMessageFailed(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.DefaultConfig()
	cfg.MessageService.SkipPushMessage = true
	cfg.MessageService.WaitingChainHeadStableDuration = time.Second * 2
	blockDelay := cfg.MessageService.WaitingChainHeadStableDuration * 2
	fsRepo := filestore.NewMockFileStore(t.TempDir())
	msh, err := newMessageServiceHelper(ctx, cfg, blockDelay, fsRepo)
	assert.NoError(t, err)
	ms := msh.ms

	account := defaultLocalToken
	addrCount := 10
	addrs := testhelper.ResolveAddrs(t, testhelper.RandAddresses(t, addrCount))
	assert.NoError(t, msh.walletProxy.AddAddress(account, addrs))
	assert.NoError(t, msh.fullNode.AddActors(addrs))

	lc := fxtest.NewLifecycle(t)
	_ = StartNodeEvents(lc, msh.fullNode, ms, ms.log)
	assert.NoError(t, lc.Start(ctx))
	defer lc.RequireStop()

	msgs := genMessages(addrs, defaultLocalToken, len(addrs)*10)
	assert.NoError(t, pushMessage(ctx, ms, msgs))

	removedAddrs := addrs[:len(addrs)/2]
	aliveAddrs := addrs[len(addrs)/2:]
	assert.NoError(t, msh.walletProxy.RemoveAddress(defaultLocalToken, removedAddrs))

	ts, err := msh.fullNode.ChainHead(ctx)
	assert.NoError(t, err)
	selectResult, err := ms.messageSelector.SelectMessage(ctx, ts)
	assert.NoError(t, err)
	assert.Len(t, selectResult.SelectMsg, (len(addrs)-len(removedAddrs))*10)
	assert.Len(t, selectResult.ErrMsg, len(removedAddrs))
	assert.Len(t, selectResult.ModifyAddress, len(aliveAddrs))
	assert.Len(t, selectResult.ExpireMsg, 0)
	assert.Len(t, selectResult.ToPushMsg, 0)
	testhelper.IsSortedByNonce(t, selectResult.SelectMsg)
	assert.NoError(t, saveAndPushMsgs(ctx, ms, selectResult))
	checkMsgs(ctx, t, ms, msgs, selectResult.SelectMsg)

	removedAddrMap := make(map[address.Address]struct{})
	for _, addr := range removedAddrs {
		removedAddrMap[addr] = struct{}{}
	}
	for _, errInfo := range selectResult.ErrMsg {
		res, err := ms.GetMessageByUid(ctx, errInfo.id)
		assert.NoError(t, err)
		assert.Contains(t, string(res.Receipt.Return), signMsg)

		_, ok := removedAddrMap[res.From]
		assert.True(t, ok)
	}
}

type messageServiceHelper struct {
	fullNode    *testhelper.MockFullNode
	walletProxy *gateway.MockWalletProxy

	ms *MessageService
}

func newMessageServiceHelper(ctx context.Context, cfg *config.Config, blockDelay time.Duration, fsRepo filestore.FSRepo) (*messageServiceHelper, error) {
	fullNode, err := testhelper.NewMockFullNode(ctx, blockDelay)
	if err != nil {
		return nil, err
	}
	walletProxy := gateway.NewMockWalletProxy()

	if err := fsRepo.ReplaceConfig(cfg); err != nil {
		return nil, err
	}

	log := log.New()
	repo, err := models.SetDataBase(fsRepo)
	if err != nil {
		return nil, err
	}
	err = repo.AutoMigrate()
	if err != nil {
		return nil, err
	}

	addressService := NewAddressService(repo, log, walletProxy)
	sharedParamsService, err := NewSharedParamsService(ctx, repo, log)
	if err != nil {
		return nil, err
	}

	msgState, err := NewMessageState(repo, log, &cfg.MessageState)
	if err != nil {
		return nil, err
	}

	ms, err := NewMessageService(ctx, repo, fullNode, log, fsRepo, msgState, addressService, sharedParamsService,
		NewNodeService(repo, log), walletProxy, &pubsub.MessagerPubSubStub{})
	if err != nil {
		return nil, err
	}

	return &messageServiceHelper{
		fullNode:    fullNode,
		walletProxy: walletProxy,
		ms:          ms,
	}, nil
}

func pushMessage(ctx context.Context, ms *MessageService, msgs []*types.Message) error {
	for _, msg := range msgs {
		// avoid being modified
		msgCopy := *msg
		if err := ms.pushMessage(ctx, &msgCopy); err != nil {
			return err
		}
	}
	return nil
}

func saveAndPushMsgs(ctx context.Context, ms *MessageService, selectResult *MsgSelectResult) error {
	if err := saveMsgsAndUpdateCache(ctx, ms, selectResult); err != nil {
		return err
	}
	go func() {
		ms.multiPushMessages(ctx, selectResult)
	}()
	return nil
}

func saveMsgsAndUpdateCache(ctx context.Context, ms *MessageService, selectResult *MsgSelectResult) error {
	if err := ms.saveSelectedMessagesToDB(ctx, selectResult); err != nil {
		return err
	}
	return ms.updateCacheForSelectedMessages(selectResult)
}

func checkMsgs(ctx context.Context, t *testing.T, ms *MessageService, srcMsgs []*types.Message, selectedMsgs []*types.Message) {
	ctx, calcel := context.WithTimeout(ctx, time.Minute*3)
	defer calcel()

	sharedParams, err := ms.sps.GetSharedParams(ctx)
	assert.NoError(t, err)
	addrInfos := make(map[address.Address]*types.Address)
	idMsgMap := testhelper.SliceToMap(srcMsgs)
	for _, msg := range selectedMsgs {
		res := waitMsgAndCheck(ctx, t, msg, ms)

		addrInfo, ok := addrInfos[msg.From]
		if !ok {
			addrInfo, err = ms.addressService.GetAddress(ctx, msg.From)
			assert.NoError(t, err)
			addrInfos[msg.From] = addrInfo
		}

		checkGasFee(t, idMsgMap[msg.ID].(*types.Message), res, sharedParams, addrInfo)
	}
}

type waitMsgRes struct {
	msg *types.Message
	err error
}

func waitMsgWithTimeout(ctx context.Context, ms *MessageService, msgID string) (*types.Message, error) {
	resChan := make(chan *waitMsgRes)

	go func() {
		res, err := ms.WaitMessage(ctx, msgID, constants.MessageConfidence)
		resChan <- &waitMsgRes{
			msg: res,
			err: err,
		}
		close(resChan)
	}()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("context done: %v", ctx.Err())
	case res := <-resChan:
		return res.msg, res.err
	}
}

func waitMsgAndCheck(ctx context.Context, t *testing.T, msg *types.Message, ms *MessageService) *types.Message {
	res, err := waitMsgWithTimeout(ctx, ms, msg.ID)
	assert.NoError(t, err)
	assert.Equal(t, msg.ID, res.ID)
	assert.Equal(t, types.OnChainMsg, res.State)
	assert.Greater(t, res.Height, int64(0))
	assert.NotEmpty(t, res.TipSetKey.String())
	assert.GreaterOrEqual(t, res.Nonce, uint64(0))
	assert.NotNil(t, res.Signature)
	assert.NotNil(t, res.SignedCid)
	assert.NotNil(t, res.UnsignedCid)
	assert.NotNil(t, res.Receipt)

	return res
}

func checkGasFee(t *testing.T, srcMsgs, currMsgs *types.Message, sharedParams *types.SharedSpec, addrInfo *types.Address) {
	params := &Params{SharedSpec: sharedParams}
	meta := &types.SendSpec{}
	if srcMsgs.Meta != nil {
		meta = srcMsgs.Meta
	}
	gasSpec := mergeMsgSpec(params, meta, addrInfo, srcMsgs)
	gasLimit := testhelper.DefGasUsed
	gasPremium := testhelper.DefGasPremium
	if gasSpec.GasOverEstimation != 0 {
		gasLimit = int64(float64(gasLimit) * gasSpec.GasOverEstimation)
	}
	if gasSpec.GasOverPremium != 0 {
		gasPremium = big.Div(big.Mul(gasPremium, big.NewInt(int64(gasSpec.GasOverPremium*10000))), big.NewInt(10000))
	}
	gasFeeCap := big.Add(testhelper.DefGasFeeCap, gasPremium)
	if !gasSpec.GasFeeCap.NilOrZero() && srcMsgs.GasFeeCap.IsZero() {
		gasFeeCap = gasSpec.GasFeeCap
	}
	maxFee := testhelper.DefMaxFee
	if !gasSpec.MaxFee.NilOrZero() {
		maxFee = gasSpec.MaxFee
	}

	gl := big.NewInt(gasLimit)
	totalFee := big.Mul(gasFeeCap, gl)
	if !totalFee.LessThanEqual(maxFee) {
		gasFeeCap = big.Div(maxFee, gl)
		gasPremium = big.Min(gasFeeCap, gasPremium)
	}
	assert.Equal(t, gasLimit, currMsgs.GasLimit)
	assert.Equal(t, gasFeeCap, currMsgs.GasFeeCap)
	assert.Equal(t, gasPremium, currMsgs.GasPremium)
}

func genMessages(addrs []address.Address, account string, count int) []*types.Message {
	msgs := testhelper.NewMessages(count)
	sendSpecs := testhelper.MockSendSpecs()
	for i, msg := range msgs {
		msg.FromUser = account
		msg.WalletName = account
		msg.From = addrs[i%len(addrs)]
		msg.Meta = sendSpecs[i%len(sendSpecs)]
	}

	return msgs
}
