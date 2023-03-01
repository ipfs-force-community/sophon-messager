// stm: #unit
package service

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/big"

	"github.com/filecoin-project/venus-messager/filestore"
	"github.com/filecoin-project/venus-messager/models"
	"github.com/filecoin-project/venus-messager/models/repo"
	"github.com/filecoin-project/venus-messager/testhelper"

	"github.com/filecoin-project/venus/venus-shared/testutil"
	shared "github.com/filecoin-project/venus/venus-shared/types"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
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
		FeeSpec: types.FeeSpec{
			GasOverEstimation: 1.5,
			MaxFee:            big.NewInt(50000),
			GasFeeCap:         big.NewInt(50000),
			GasOverPremium:    5.0,
			BaseFee:           big.NewInt(50001),
		},
	}

	actorCfg := &types.ActorCfg{
		FeeSpec: types.FeeSpec{
			GasOverEstimation: 2.0,
			MaxFee:            big.NewInt(60000),
			GasFeeCap:         big.NewInt(60000),
			GasOverPremium:    6.0,
			BaseFee:           big.NewInt(60001),
		},
	}
	emptyAddrInfo := &types.Address{}

	msg := testhelper.NewMessage()
	msg2 := testhelper.NewMessage()
	msg2.GasFeeCap = testhelper.DefGasFeeCap

	tests := []struct {
		globalSpec *types.SharedSpec
		sendSpec   *types.SendSpec
		addrInfo   *types.Address
		actorCfg   *types.ActorCfg
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
			nil,
			msg,
			&GasSpec{GasOverEstimation: sendSpec.GasOverEstimation, MaxFee: sendSpec.MaxFee, GasOverPremium: sendSpec.GasOverPremium, GasFeeCap: addrInfo.GasFeeCap, BaseFee: addrInfo.BaseFee},
		},
		{
			defSharedParams,
			emptySendSpec,
			addrInfo,
			nil,
			msg,
			&GasSpec{GasOverEstimation: addrInfo.GasOverEstimation, MaxFee: addrInfo.MaxFee, GasOverPremium: addrInfo.GasOverPremium, GasFeeCap: addrInfo.GasFeeCap, BaseFee: addrInfo.BaseFee},
		},
		{
			defSharedParams,
			emptySendSpec,
			emptyAddrInfo,
			nil,
			msg,
			&GasSpec{GasOverEstimation: defSharedParams.GasOverEstimation, MaxFee: defSharedParams.MaxFee, GasOverPremium: defSharedParams.GasOverPremium, GasFeeCap: defSharedParams.GasFeeCap, BaseFee: defSharedParams.BaseFee},
		},
		{
			defSharedParams,
			emptySendSpec,
			addrInfo,
			nil,
			msg2,
			&GasSpec{GasOverEstimation: addrInfo.GasOverEstimation, MaxFee: addrInfo.MaxFee, GasOverPremium: addrInfo.GasOverPremium, BaseFee: addrInfo.BaseFee},
		},
		{
			defSharedParams,
			emptySendSpec,
			emptyAddrInfo,
			nil,
			msg2,
			&GasSpec{GasOverEstimation: defSharedParams.GasOverEstimation, MaxFee: defSharedParams.MaxFee, GasOverPremium: defSharedParams.GasOverPremium, BaseFee: defSharedParams.BaseFee},
		},
		{
			defSharedParams,
			emptySendSpec,
			emptyAddrInfo,
			actorCfg,
			msg2,
			&GasSpec{GasOverEstimation: actorCfg.GasOverEstimation, MaxFee: actorCfg.MaxFee, GasOverPremium: actorCfg.GasOverPremium, BaseFee: actorCfg.BaseFee},
		},
	}

	for _, test := range tests {
		gasSpec := mergeMsgSpec(test.globalSpec, test.sendSpec, test.addrInfo, test.actorCfg, test.msg)
		assert.Equal(t, test.expect, gasSpec)
	}
}

func TestAddrSelectMsgNum(t *testing.T) {
	ctx := context.Background()
	fsRepo := filestore.NewMockFileStore(t.TempDir())
	repo, err := models.SetDataBase(fsRepo)
	assert.NoError(t, err)
	assert.NoError(t, repo.AutoMigrate())

	sps, err := NewSharedParamsService(ctx, repo)
	assert.NoError(t, err)

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

	addrNum := addrSelectMsgNum(addrList, sharedParams.SelMsgNum)

	for _, addrInfo := range addrList {
		num, ok := addrNum[addrInfo.Addr]
		assert.True(t, ok)
		expectNum, ok := expect[addrInfo.Addr]
		assert.True(t, ok)
		assert.Equal(t, expectNum, num)
	}
}

func BenchmarkSelect(b *testing.B) {
	var total, chCount, ctxCount int
	for i := 0; i < b.N; i++ {
		total++
		ch := make(chan struct{})
		close(ch)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		select {
		case <-ctx.Done():
			ctxCount++
		case <-ch:
			chCount++
		default:
			b.Error("select default")
		}
	}
	b.Log(chCount, ctxCount, total)
	assert.LessOrEqual(b, chCount, total)
	assert.LessOrEqual(b, ctxCount, total)
}

func TestSelectMessage(t *testing.T) {
	// stm: @MESSENGER_SELECTOR_SELECT_MESSAGE_001, @MESSENGER_SELECTOR_SELECT_MESSAGE_002
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	msh := newMessageServiceHelper(ctx, t, skipPushMessage())
	addrs := msh.genAddresses()
	ms := msh.MessageService
	msh.start()
	defer msh.stop()

	// If an error occurs retrieving nonce in tipset, return that error
	err := ms.msgSelectMgr.SelectMessage(ctx, &shared.TipSet{})
	assert.Error(t, err)

	totalMsg := len(addrs) * 10
	msgs := msh.genAndPushMessages(totalMsg)

	ts, err := msh.fullNode.ChainHead(ctx)
	assert.NoError(t, err)
	err = ms.msgSelectMgr.SelectMessage(ctx, ts)
	assert.NoError(t, err)

	selectedMsgs := make([]*types.Message, 0, totalMsg)
	for _, msg := range msgs {
		res, err := ms.GetMessageByUid(ctx, msg.ID)
		assert.NoError(t, err)
		selectedMsgs = append(selectedMsgs, res)
	}
	assert.Equal(t, totalMsg, len(selectedMsgs))

	checkMsgs(ctx, t, ms, msgs, selectedMsgs)
}

func TestSelectNum(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	msh := newMessageServiceHelper(ctx, t, skipPushMessage())
	addrs := msh.genAddresses()
	ms := msh.MessageService
	msh.start()
	defer msh.stop()

	defSelectedNum := int(DefSharedParams.SelMsgNum)
	totalMsg := len(addrs) * 50
	msgs := msh.genAndPushMessages(totalMsg)

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
	selectResult := selectMsgWithAddress(ctx, t, msh, addrs, ts)
	ms.msgSelectMgr.msgReceiver <- selectResult.ToPushMsg
	assert.Len(t, selectResult.SelectMsg, len(addrs)*defSelectedNum)
	checkSelectNum(selectResult.SelectMsg, map[address.Address]int{}, defSelectedNum)
	sort.SliceIsSorted(selectResult.SelectMsg, func(i, j int) bool {
		return selectResult.SelectMsg[i].CreatedAt.After(selectResult.SelectMsg[j].CreatedAt)
	})
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
	selectResult = selectMsgWithAddress(ctx, t, msh, addrs, ts)
	ms.msgSelectMgr.msgReceiver <- selectResult.ToPushMsg
	assert.Len(t, selectResult.SelectMsg, expectNum)
	checkSelectNum(selectResult.SelectMsg, addrNum, defSelectedNum)
	checkMsgs(ctx, t, ms, msgs, selectResult.SelectMsg)
}

func TestEstimateMessageGas(t *testing.T) {
	// stm: @MESSENGER_SELECTOR_ESTIMATE_MESSAGE_GAS_001
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	msh := newMessageServiceHelper(ctx, t, skipPushMessage())
	addrs := msh.genAddresses()
	ms := msh.MessageService
	msh.start()
	defer msh.stop()

	msgs := genMessages(addrs, len(addrs)*10)
	for _, msg := range msgs {
		// will estimate gas failed
		msg.GasLimit = -1
	}
	assert.NoError(t, pushMessage(ctx, ms, msgs))
	msgsMap := testhelper.SliceToMap(msgs)

	ts, err := msh.fullNode.ChainHead(ctx)
	assert.NoError(t, err)
	selectResult := selectMsgWithAddress(ctx, t, msh, addrs, ts)
	assert.Len(t, selectResult.SelectMsg, 0)
	assert.Len(t, selectResult.ErrMsg, len(msgs))
	assert.Len(t, selectResult.ToPushMsg, 0)

	list, err := ms.ListFailedMessage(ctx, &repo.MsgQueryParams{})
	assert.NoError(t, err)
	for _, msg := range list {
		_, ok := msgsMap[msg.ID]
		assert.True(t, ok)
		assert.Contains(t, msg.ErrorMsg, testhelper.ErrGasLimitNegative.Error())

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

	msgs = genMessages(addrs, len(addrs)*10)
	for _, msg := range msgs {
		// use the fee params in the address table
		msg.Meta = nil
	}
	assert.NoError(t, pushMessage(ctx, ms, msgs))

	ts, err = msh.fullNode.ChainHead(ctx)
	assert.NoError(t, err)
	selectResult = selectMsgWithAddress(ctx, t, msh, addrs, ts)
	assert.Len(t, selectResult.SelectMsg, len(msgs))

	for _, addr := range addrs {
		addrInfo, err := ms.addressService.GetAddress(ctx, addr)
		assert.NoError(t, err)
		assert.Equal(t, uint64(10), addrInfo.Nonce)
	}
}

func TestBaseFee(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	msh := newMessageServiceHelper(ctx, t, skipPushMessage())
	addrs := msh.genAddresses()
	ms := msh.MessageService
	msh.start()
	defer msh.stop()

	addrCount := len(addrs)
	totalMsg := addrCount * int(DefSharedParams.SelMsgNum)
	msgs := msh.genAndPushMessages(totalMsg)

	// global basefee too low
	sharedParams, err := ms.sps.GetSharedParams(ctx)
	assert.NoError(t, err)
	sharedParams.BaseFee = big.Div(testhelper.DefBaseFee, big.NewInt(2))
	assert.NoError(t, ms.sps.SetSharedParams(ctx, sharedParams))

	ts, err := msh.fullNode.ChainHead(ctx)
	assert.NoError(t, err)
	selectResult := selectMsgWithAddress(ctx, t, msh, addrs, ts)
	assert.Len(t, selectResult.SelectMsg, 0)
	assert.Len(t, selectResult.ErrMsg, 0)
	assert.Len(t, selectResult.ToPushMsg, 0)

	// increase basefee for address
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

	selectResult = selectMsgWithAddress(ctx, t, msh, addrs, ts)
	addrMsgs := testhelper.MsgGroupByAddress(selectResult.SelectMsg)
	for addr, msgs := range addrMsgs {
		if _, ok := heightBaseFeeAddrs[addr]; ok {
			assert.Len(t, msgs, int(DefSharedParams.SelMsgNum))
		} else {
			assert.Len(t, selectResult.SelectMsg, 0)
		}
	}
	ms.msgSelectMgr.msgReceiver <- selectResult.ToPushMsg
	checkMsgs(ctx, t, ms, msgs, selectResult.SelectMsg)
}

func TestSignMessageFailed(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	msh := newMessageServiceHelper(ctx, t, skipPushMessage())
	addrs := msh.genAddresses()
	ms := msh.MessageService
	msh.start()
	defer msh.stop()

	addrCount := len(addrs)
	msgs := msh.genAndPushMessages(len(addrs) * 10)

	removedAddrs := addrs[:addrCount/2]
	aliveAddrs := addrs[addrCount/2:]
	assert.NoError(t, msh.walletProxy.RemoveAddress(msh.token, removedAddrs))
	aliveAddrMap := make(map[address.Address]struct{}, len(aliveAddrs))
	for _, addr := range aliveAddrs {
		aliveAddrMap[addr] = struct{}{}
	}

	ts, err := msh.fullNode.ChainHead(ctx)
	assert.NoError(t, err)
	selectResult := selectMsgWithAddress(ctx, t, msh, addrs, ts)
	assert.Len(t, selectResult.SelectMsg, len(aliveAddrs)*10)
	assert.Len(t, selectResult.ErrMsg, len(removedAddrs))
	assert.Len(t, selectResult.ToPushMsg, len(aliveAddrs)*10)

	ms.msgSelectMgr.msgReceiver <- selectResult.ToPushMsg
	checkMsgs(ctx, t, ms, msgs, selectResult.SelectMsg)

	removedAddrMap := make(map[address.Address]struct{})
	for _, addr := range removedAddrs {
		removedAddrMap[addr] = struct{}{}
	}
	for _, errInfo := range selectResult.ErrMsg {
		res, err := ms.GetMessageByUid(ctx, errInfo.id)
		assert.NoError(t, err)
		assert.Contains(t, res.ErrorMsg, signMsg)

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
	gasSpec := mergeMsgSpec(sharedParams, meta, addrInfo, nil, srcMsgs)
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

func genMessages(addrs []address.Address, count int) []*types.Message {
	msgs := testhelper.NewMessages(count)
	sendSpecs := testhelper.MockSendSpecs()
	for i, msg := range msgs {
		msg.From = addrs[i%len(addrs)]
		msg.To = addrs[i%len(addrs)]
		msg.Meta = sendSpecs[i%len(sendSpecs)]
	}
	return msgs
}

func selectMsgWithAddress(ctx context.Context,
	t *testing.T,
	msh *messageServiceHelper,
	addrs []address.Address,
	ts *shared.TipSet,
) *MsgSelectResult {
	ms := msh.MessageService
	sharedParams, err := ms.sps.GetSharedParams(ctx)
	assert.NoError(t, err)
	activeAddrs, err := ms.addressService.ListActiveAddress(ctx)
	assert.NoError(t, err)
	addrSelMsgNum := addrSelectMsgNum(activeAddrs, sharedParams.SelMsgNum)
	allSelectRes := &MsgSelectResult{}
	for _, addr := range addrs {
		work := newWork(ctx, addr, ms.msgSelectMgr.cfg, msh.fullNode, ms.repo, ms.addressService, ms.walletClient, ms.msgReceiver)
		appliedNonce, err := ms.msgSelectMgr.getNonceInTipset(ctx, ts)
		assert.NoError(t, err)
		addrInfo, err := ms.addressService.GetAddress(ctx, addr)
		assert.NoError(t, err)
		selectResult, err := work.selectMessage(ctx, appliedNonce, addrInfo, ts, addrSelMsgNum[addr], sharedParams)
		assert.NoError(t, err)
		testhelper.IsSortedByNonce(t, selectResult.SelectMsg)

		allSelectRes.SelectMsg = append(allSelectRes.SelectMsg, selectResult.SelectMsg...)
		for _, msg := range selectResult.SelectMsg {
			allSelectRes.ToPushMsg = append(allSelectRes.ToPushMsg, &shared.SignedMessage{
				Message:   msg.Message,
				Signature: *msg.Signature,
			})
		}
		allSelectRes.ErrMsg = append(allSelectRes.ErrMsg, selectResult.ErrMsg...)

		assert.NoError(t, work.saveSelectedMessages(ctx, selectResult))
	}

	return allSelectRes
}
