// stm: #unit
package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/venus/venus-shared/testutil"
	shared "github.com/filecoin-project/venus/venus-shared/types"
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
)

const defaultLocalToken = "defaultLocalToken"

func TestMergeMsgSpec(t *testing.T) {
	defSharedPramsCopy := *DefSharedParams
	defSharedParams := &defSharedPramsCopy
	defSharedParams.GasOverPremium = 1.0
	defSharedParams.GasFeeCap = big.NewInt(10000)
	defSharedParams.BaseFee = big.NewInt(10000)

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
		BaseFee:           big.NewInt(50001),
	}
	emptyAddrInfo := &types.Address{}

	msg := testhelper.NewMessage()
	msg2 := testhelper.NewMessage()
	msg2.GasFeeCap = testhelper.DefGasFeeCap

	tests := []struct {
		globalSpec *types.SharedSpec
		sendSpec   *types.SendSpec
		addrInfo   *types.Address
		msg        *types.Message

		expect *GasSpec
	}{
		{
			globalSpec: DefSharedParams,
			sendSpec:   emptySendSpec,
			addrInfo:   emptyAddrInfo,
			msg:        msg,
			expect:     &GasSpec{GasOverEstimation: DefSharedParams.GasOverEstimation, MaxFee: DefSharedParams.MaxFee, GasOverPremium: 0, GasFeeCap: DefSharedParams.GasFeeCap, BaseFee: DefSharedParams.BaseFee},
		},
		{
			defSharedParams,
			sendSpec,
			addrInfo,
			msg,
			&GasSpec{GasOverEstimation: sendSpec.GasOverEstimation, MaxFee: sendSpec.MaxFee, GasOverPremium: sendSpec.GasOverPremium, GasFeeCap: addrInfo.GasFeeCap, BaseFee: addrInfo.BaseFee},
		},
		{
			defSharedParams,
			emptySendSpec,
			addrInfo,
			msg,
			&GasSpec{GasOverEstimation: addrInfo.GasOverEstimation, MaxFee: addrInfo.MaxFee, GasOverPremium: addrInfo.GasOverPremium, GasFeeCap: addrInfo.GasFeeCap, BaseFee: addrInfo.BaseFee},
		},
		{
			defSharedParams,
			emptySendSpec,
			emptyAddrInfo,
			msg,
			&GasSpec{GasOverEstimation: defSharedParams.GasOverEstimation, MaxFee: defSharedParams.MaxFee, GasOverPremium: defSharedParams.GasOverPremium, GasFeeCap: defSharedParams.GasFeeCap, BaseFee: defSharedParams.BaseFee},
		},
		{
			defSharedParams,
			emptySendSpec,
			addrInfo,
			msg2,
			&GasSpec{GasOverEstimation: addrInfo.GasOverEstimation, MaxFee: addrInfo.MaxFee, GasOverPremium: addrInfo.GasOverPremium, BaseFee: addrInfo.BaseFee},
		},
		{
			defSharedParams,
			emptySendSpec,
			emptyAddrInfo,
			msg2,
			&GasSpec{GasOverEstimation: defSharedParams.GasOverEstimation, MaxFee: defSharedParams.MaxFee, GasOverPremium: defSharedParams.GasOverPremium, BaseFee: defSharedParams.BaseFee},
		},
	}

	for _, test := range tests {
		gasSpec := mergeMsgSpec(test.globalSpec, test.sendSpec, test.addrInfo, test.msg)
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

	sharedParams, err := sps.GetSharedParams(ctx)
	assert.NoError(t, err)

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
		addr2: sharedParams.SelMsgNum,
	}

	addrNum := msgSelector.addrSelectMsgNum(addrList, sharedParams.SelMsgNum)

	for _, addrInfo := range addrList {
		num, ok := addrNum[addrInfo.Addr]
		assert.True(t, ok)
		expectNum, ok := expect[addrInfo.Addr]
		assert.True(t, ok)
		assert.Equal(t, expectNum, num)
	}
}

func TestSelectMessage(t *testing.T) {
	// stm: @MESSENGER_SELECTOR_SELECT_MESSAGE_001, @MESSENGER_SELECTOR_SELECT_MESSAGE_002
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

	// If an error occurs retrieving nonce in tipset, return that error
	_, err = ms.messageSelector.SelectMessage(ctx, &shared.TipSet{})
	assert.Error(t, err)

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
		addrMsgs := testhelper.MsgGroupByAddress(msgs)
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
	// stm: @MESSENGER_SELECTOR_ESTIMATE_MESSAGE_GAS_001
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
		params := &types.AddressSpec{
			Address:           addr,
			GasOverEstimation: float64(i) * gasOverEstimation,
			GasOverPremium:    float64(i) * gasOverPremium,
			MaxFeeStr:         big.Mul(testhelper.DefMaxFee, big.NewInt(int64(i))).String(),
			GasFeeCapStr:      big.Mul(testhelper.DefGasFeeCap, big.NewInt(int64(i))).String(),
			BaseFeeStr:        big.Mul(testhelper.DefBaseFee, big.NewInt(int64(i))).String(),
		}
		assert.NoError(t, ms.addressService.SetFeeParams(ctx, params))
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

func TestBaseFee(t *testing.T) {
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

	totalMsg := len(addrs) * int(DefSharedParams.SelMsgNum)
	msgs := genMessages(addrs, account, totalMsg)
	assert.NoError(t, pushMessage(ctx, ms, msgs))

	// global basefee too low
	sharedParams, err := ms.sps.GetSharedParams(ctx)
	assert.NoError(t, err)
	sharedParams.BaseFee = big.Div(testhelper.DefBaseFee, big.NewInt(2))
	assert.NoError(t, ms.sps.SetSharedParams(ctx, sharedParams))

	ts, err := msh.fullNode.ChainHead(ctx)
	assert.NoError(t, err)
	selectResult, err := ms.messageSelector.SelectMessage(ctx, ts)
	assert.NoError(t, err)
	assert.Len(t, selectResult.SelectMsg, 0)
	assert.Len(t, selectResult.ErrMsg, 0)
	assert.Len(t, selectResult.ModifyAddress, 0)
	assert.Len(t, selectResult.ExpireMsg, 0)
	assert.Len(t, selectResult.ToPushMsg, 0)

	// set basefee for address
	heightBaseFeeAddrs := make(map[address.Address]struct{}, addrCount/2)
	for i, addr := range addrs {
		if i%2 == 0 {
			addrSpec := types.AddressSpec{
				Address:    addr,
				BaseFeeStr: big.Mul(testhelper.DefBaseFee, big.NewInt(2)).String(),
			}
			heightBaseFeeAddrs[addr] = struct{}{}
			assert.NoError(t, ms.addressService.SetFeeParams(ctx, &addrSpec))
		}
	}

	ts, err = msh.fullNode.ChainHead(ctx)
	assert.NoError(t, err)
	selectResult, err = ms.messageSelector.SelectMessage(ctx, ts)
	assert.NoError(t, err)
	assert.Len(t, selectResult.SelectMsg, totalMsg/2)
	assert.Len(t, selectResult.ErrMsg, 0)
	assert.Len(t, selectResult.ModifyAddress, addrCount/2)
	assert.Len(t, selectResult.ExpireMsg, 0)
	assert.Len(t, selectResult.ToPushMsg, 0)
	for _, addr := range selectResult.ModifyAddress {
		_, ok := heightBaseFeeAddrs[addr.Addr]
		assert.True(t, ok)
	}
	for addr := range testhelper.MsgGroupByAddress(selectResult.SelectMsg) {
		_, ok := heightBaseFeeAddrs[addr]
		assert.True(t, ok)
	}
	assert.NoError(t, saveAndPushMsgs(ctx, ms, selectResult))
	checkMsgs(ctx, t, ms, msgs, selectResult.SelectMsg)

	// increase basefee
	for _, addr := range addrs {
		if _, ok := heightBaseFeeAddrs[addr]; !ok {
			addrSpec := types.AddressSpec{
				Address:    addr,
				BaseFeeStr: big.Mul(testhelper.DefBaseFee, big.NewInt(2)).String(),
			}
			heightBaseFeeAddrs[addr] = struct{}{}
			assert.NoError(t, ms.addressService.SetFeeParams(ctx, &addrSpec))
		}
	}
	ts, err = msh.fullNode.ChainHead(ctx)
	assert.NoError(t, err)
	selectResult, err = ms.messageSelector.SelectMessage(ctx, ts)
	assert.NoError(t, err)
	assert.Len(t, selectResult.SelectMsg, totalMsg/2)
	assert.Len(t, selectResult.ErrMsg, 0)
	assert.Len(t, selectResult.ModifyAddress, addrCount/2)
	assert.Len(t, selectResult.ExpireMsg, 0)
	assert.Len(t, selectResult.ToPushMsg, 0)
	for _, addr := range selectResult.ModifyAddress {
		_, ok := heightBaseFeeAddrs[addr.Addr]
		assert.True(t, ok)
	}
	for addr := range testhelper.MsgGroupByAddress(selectResult.SelectMsg) {
		_, ok := heightBaseFeeAddrs[addr]
		assert.True(t, ok)
	}
	testhelper.IsSortedByNonce(t, selectResult.SelectMsg)
	assert.NoError(t, saveAndPushMsgs(ctx, ms, selectResult))
	checkMsgs(ctx, t, ms, msgs, selectResult.SelectMsg)
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

func TestCapGasFee(t *testing.T) {
	// stm: @MESSENGER_SELECTOR_CAP_MESSAGE_GAS_001
	msg := testhelper.NewMessage().Message
	maxfee := func(msg *shared.Message) big.Int {
		return big.Mul(big.NewInt(msg.GasLimit), msg.GasFeeCap)
	}
	oldFeeCap := big.NewInt(1000)
	oldGasPremium := oldFeeCap
	msg.GasLimit = 10000
	msg.GasFeeCap = oldFeeCap
	msg.GasPremium = oldGasPremium
	oldMaxFee := maxfee(&msg)
	descedMaxFee := big.Div(oldMaxFee, big.NewInt(10))
	CapGasFee(&msg, descedMaxFee)
	newMaxFee := maxfee(&msg)
	assert.Less(t, big.Cmp(msg.GasPremium, oldGasPremium), 0)
	assert.Less(t, big.Cmp(newMaxFee, oldMaxFee), 0)
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

	ms, err := NewMessageService(ctx, repo, fullNode, log, fsRepo, addressService, sharedParamsService,
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
		// avoid been modified
		msgCopy := *msg
		if err := ms.pushMessage(ctx, &msgCopy); err != nil {
			return err
		}
	}
	return nil
}

func saveAndPushMsgs(ctx context.Context, ms *MessageService, selectResult *MsgSelectResult) error {
	if err := saveMsgsToDB(ctx, ms, selectResult); err != nil {
		return err
	}
	go func() {
		ms.multiPushMessages(ctx, selectResult)
	}()
	return nil
}

func saveMsgsToDB(ctx context.Context, ms *MessageService, selectResult *MsgSelectResult) error {
	if err := ms.saveSelectedMessagesToDB(ctx, selectResult); err != nil {
		return err
	}
	for _, msg := range selectResult.SelectMsg {
		selectResult.ToPushMsg = append(selectResult.ToPushMsg, &shared.SignedMessage{
			Message:   msg.Message,
			Signature: *msg.Signature,
		})
	}
	return nil
}

func checkMsgs(ctx context.Context, t *testing.T, ms *MessageService, srcMsgs []*types.Message, selectedMsgs []*types.Message) {
	ctx, calcel := context.WithTimeout(ctx, time.Minute*3)
	defer calcel()

	sharedParams, err := ms.sps.GetSharedParams(ctx)
	assert.NoError(t, err)
	addrInfos := make(map[address.Address]*types.Address)
	idMsgMap := testhelper.SliceToMap(srcMsgs)
	for _, msg := range selectedMsgs {
		res := waitMsgAndCheck(ctx, t, msg.ID, ms)

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
		res, err := ms.WaitMessage(ctx, msgID, 1)
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

func waitMsgAndCheck(ctx context.Context, t *testing.T, msgID string, ms *MessageService) *types.Message {
	res, err := waitMsgWithTimeout(ctx, ms, msgID)
	assert.NoError(t, err)
	assert.Equal(t, msgID, res.ID)
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
	meta := &types.SendSpec{}
	if srcMsgs.Meta != nil {
		meta = srcMsgs.Meta
	}
	gasSpec := mergeMsgSpec(sharedParams, meta, addrInfo, srcMsgs)
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
