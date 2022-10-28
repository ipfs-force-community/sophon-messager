// stm: #unit
package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"testing"
	"time"

	v1Mock "github.com/filecoin-project/venus/venus-shared/api/chain/v1/mock"

	"github.com/golang/mock/gomock"

	"github.com/ipfs/go-cid"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"

	"github.com/stretchr/testify/assert"
	"go.uber.org/fx/fxtest"

	"github.com/filecoin-project/venus-messager/config"
	"github.com/filecoin-project/venus-messager/filestore"
	"github.com/filecoin-project/venus-messager/testhelper"

	"github.com/filecoin-project/venus/pkg/constants"
	"github.com/filecoin-project/venus/venus-shared/testutil"
	shared "github.com/filecoin-project/venus/venus-shared/types"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
)

func TestVerifyNetworkName(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.DefaultConfig()
	cfg.MessageService.SkipPushMessage = true
	cfg.MessageService.WaitingChainHeadStableDuration = time.Second * 2
	blockDelay := cfg.MessageService.WaitingChainHeadStableDuration * 2

	fsRepo := filestore.NewMockFileStore(t.TempDir())
	tipsetCache := &TipsetCache{
		NetworkName: string(shared.NetworkNameMain),
	}
	assert.NoError(t, tipsetCache.Save(fsRepo.TipsetFile()))

	msh, err := newMessageServiceHelper(ctx, cfg, blockDelay, fsRepo, testhelper.NewMockAuthClient())
	assert.NoError(t, err)

	networkName, err := msh.fullNode.StateNetworkName(ctx)
	assert.NoError(t, err)

	msh.ms.tsCache.NetworkName = string(shared.NetworkNameButterfly)
	err = msh.ms.verifyNetworkName()

	expectErrStr := fmt.Sprintf("network name not match, expect %s, actual %s, please remove `%s`",
		networkName, msh.ms.tsCache.NetworkName, fsRepo.TipsetFile())
	assert.Equal(t, expectErrStr, err.Error())
}

func TestReplaceMessage(t *testing.T) {
	// stm: @MESSENGER_SERVICE_REPLACE_MESSAGE_001, @MESSENGER_SERVICE_REPLACE_MESSAGE_002
	// stm: @MESSENGER_SERVICE_REPLACE_MESSAGE_003, @MESSENGER_SERVICE_REPLACE_MESSAGE_004
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.DefaultConfig()
	cfg.MessageService.SkipPushMessage = true
	cfg.MessageService.WaitingChainHeadStableDuration = time.Second * 2
	blockDelay := cfg.MessageService.WaitingChainHeadStableDuration * 2
	fsRepo := filestore.NewMockFileStore(t.TempDir())
	authClient := testhelper.NewMockAuthClient()
	msh, err := newMessageServiceHelper(ctx, cfg, blockDelay, fsRepo, authClient)
	assert.NoError(t, err)
	ms := msh.ms

	addrCount := 10
	addrs := testhelper.ResolveAddrs(t, testhelper.RandAddresses(t, addrCount))
	authClient.AddMockUserAndSigner(defaultLocalToken, addrs)
	assert.NoError(t, msh.walletProxy.AddAddress(defaultLocalToken, addrs))
	assert.NoError(t, msh.fullNode.AddActors(addrs))

	lc := fxtest.NewLifecycle(t)
	_ = StartNodeEvents(lc, msh.fullNode, ms)
	assert.NoError(t, lc.Start(ctx))
	defer lc.RequireStop()

	blockedMsgs := make(map[string]*types.Message, 0)
	msgs := genMessages(addrs, len(addrs)*10)
	for i, msg := range msgs {
		if i%2 == 0 {
			msg.GasPremium = big.Sub(testhelper.MinPackedPremium, big.NewInt(100))
			blockedMsgs[msg.ID] = msg
		}
	}
	assert.NoError(t, pushMessage(ctx, ms, msgs))

	ts, err := msh.fullNode.ChainHead(ctx)
	assert.NoError(t, err)
	selectResult, err := ms.messageSelector.SelectMessage(ctx, ts)
	assert.NoError(t, err)
	assert.Len(t, selectResult.SelectMsg, len(msgs))
	assert.Len(t, selectResult.ErrMsg, 0)
	assert.Len(t, selectResult.ModifyAddress, len(addrs))
	assert.Len(t, selectResult.ExpireMsg, 0)
	assert.Len(t, selectResult.ToPushMsg, 0)
	testhelper.IsSortedByNonce(t, selectResult.SelectMsg)
	assert.NoError(t, saveAndPushMsgs(ctx, ms, selectResult))

	notBlockedMsgs := make([]*types.Message, 0)
	for _, msg := range selectResult.SelectMsg {
		if _, ok := blockedMsgs[msg.ID]; !ok {
			notBlockedMsgs = append(notBlockedMsgs, msg)
		}
	}
	checkMsgs(ctx, t, ms, msgs, notBlockedMsgs)

	replacedMsgs := make([]*types.Message, 0, len(blockedMsgs))
	for _, msg := range blockedMsgs {
		params := &types.ReplacMessageParams{
			ID:   msg.ID,
			Auto: true,
		}
		if msg.Meta != nil {
			params.MaxFee = msg.Meta.MaxFee
			params.GasOverPremium = msg.Meta.GasOverPremium
		}
		_, err := ms.ReplaceMessage(ctx, params)
		assert.NoError(t, err)

		res, err := ms.GetMessageByUid(ctx, msg.ID)
		assert.NoError(t, err)
		replacedMsgs = append(replacedMsgs, res)
	}

	ctx, calcel := context.WithTimeout(ctx, time.Minute*3)
	defer calcel()
	for _, msg := range replacedMsgs {
		res, err := waitMsgWithTimeout(ctx, ms, msg.ID)
		assert.NoError(t, err)
		assert.Equal(t, msg.GasLimit, res.GasLimit)
		assert.Equal(t, msg.GasFeeCap, res.GasFeeCap)
		assert.Equal(t, msg.GasPremium, res.GasPremium)
	}

	failedMessageReplace := func(*types.ReplacMessageParams) {
		_, err = ms.ReplaceMessage(ctx, nil)
		assert.Error(t, err)
	}

	// param is nil, expect an error
	failedMessageReplace(nil)
	// message can't find, expect an error
	failedMessageReplace(&types.ReplacMessageParams{ID: shared.NewUUID().String(), Auto: true})
	// message is already on chain, expect an error
	failedMessageReplace(&types.ReplacMessageParams{ID: replacedMsgs[0].ID, Auto: true})
}

func TestReconnectCheck(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.DefaultConfig()
	cfg.MessageService.WaitingChainHeadStableDuration = time.Second * 2
	blockDelay := cfg.MessageService.WaitingChainHeadStableDuration * 2
	fsRepo := filestore.NewMockFileStore(t.TempDir())
	msh, err := newMessageServiceHelper(ctx, cfg, blockDelay, fsRepo, testhelper.NewMockAuthClient())
	assert.NoError(t, err)

	t.Run("tipset cache is empty", func(t *testing.T) {
		ms := newMessageService(msh, filestore.NewMockFileStore(t.TempDir()))
		ts, err := msh.fullNode.ChainHead(ctx)
		assert.NoError(t, err)
		assert.NoError(t, ms.ReconnectCheck(ctx, ts))
	})

	t.Run("head not change", func(t *testing.T) {
		ms := newMessageService(msh, filestore.NewMockFileStore(t.TempDir()))
		ts, err := msh.fullNode.ChainHead(ctx)
		assert.NoError(t, err)
		ms.tsCache.Add(ts)
		assert.NoError(t, ms.ReconnectCheck(ctx, ts))
	})

	t.Run("normal", func(t *testing.T) {
		ms := newMessageService(msh, filestore.NewMockFileStore(t.TempDir()))
		ts, err := msh.fullNode.ChainHead(ctx)
		assert.NoError(t, err)
		ms.tsCache.Add(ts)

		expectTS := ts
		expectHeight := abi.ChainEpoch(5) + ts.Height()
		tsMap := make(map[abi.ChainEpoch]shared.TipSet, 5)

		ticker := time.NewTicker(blockDelay)
		defer ticker.Stop()

		ctx, cancel := context.WithTimeout(ctx, blockDelay*2*time.Duration(expectHeight))
		defer cancel()

		for expectTS.Height() < expectHeight {
			select {
			case <-ticker.C:
				expectTS, err = msh.fullNode.ChainHead(ctx)
				assert.NoError(t, err)
				tsMap[expectTS.Height()] = *expectTS
			case <-ctx.Done():
				t.Errorf("not found tipset")
			}
		}
		go func() {
			assert.NoError(t, ms.ReconnectCheck(ctx, expectTS))
		}()
		headChange := <-ms.headChans
		assert.True(t, headChange.isReconnect)
		assert.Len(t, headChange.apply, int(expectTS.Height()-ts.Height()))
		assert.Len(t, headChange.revert, 0)
		for _, ts := range headChange.apply {
			assert.Equal(t, tsMap[ts.Height()], *ts)
		}
		headChange.done <- nil
	})

	t.Run("test with revert", func(t *testing.T) {
		ms := newMessageService(msh, filestore.NewMockFileStore(t.TempDir()))
		ts, err := msh.fullNode.ChainHead(ctx)
		assert.NoError(t, err)
		ms.tsCache.Add(ts)

		expectTS := ts
		expectHeight := abi.ChainEpoch(10) + ts.Height()
		revertHeight := abi.ChainEpoch(5) + ts.Height()

		ticker := time.NewTicker(blockDelay)
		defer ticker.Stop()

		ctx, cancel := context.WithTimeout(ctx, blockDelay*2*time.Duration(expectHeight))
		defer cancel()

		revertSignal := &testhelper.RevertSignal{ExpectRevertCount: 3, RevertedTS: make(chan []*shared.TipSet, 1)}

		for expectTS.Height() < expectHeight {
			select {
			case <-ticker.C:
				expectTS, err = msh.fullNode.ChainHead(ctx)
				assert.NoError(t, err)
				if expectTS.Height() < revertHeight {
					ms.tsCache.Add(expectTS)
				} else if expectTS.Height() == revertHeight {
					msh.fullNode.SendRevertSignal(revertSignal)
				}
			case <-ctx.Done():
				t.Errorf("not found tipset")
			}
		}
		revertedTS := <-revertSignal.RevertedTS
		go func() {
			assert.NoError(t, ms.ReconnectCheck(ctx, expectTS))
		}()

		headChange := <-ms.headChans
		assert.True(t, headChange.isReconnect)
		assert.Len(t, headChange.apply, int(expectTS.Height()-revertedTS[len(revertedTS)-1].Height())+1)
		revert := headChange.revert
		apply := headChange.apply
		assert.Equal(t, revert[len(revert)-1].Parents(), apply[len(apply)-1].Parents())
		assert.Equal(t, revertedTS[len(revertedTS)-1], revert[len(revert)-1])
		headChange.done <- nil
	})
}

func TestMessageService_ProcessNewHead(t *testing.T) {
	// stm: @MESSENGER_SERVICE_LIST_MESSAGE_BY_ADDRESS_001
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.DefaultConfig()
	cfg.MessageService.WaitingChainHeadStableDuration = time.Second * 2
	blockDelay := cfg.MessageService.WaitingChainHeadStableDuration * 2
	fsRepo := filestore.NewMockFileStore(t.TempDir())
	msh, err := newMessageServiceHelper(ctx, cfg, blockDelay, fsRepo, testhelper.NewMockAuthClient())
	assert.NoError(t, err)

	t.Run("tipset cache is empty", func(t *testing.T) {
		ms := newMessageService(msh, filestore.NewMockFileStore(t.TempDir()))
		ts, err := msh.fullNode.ChainHead(ctx)
		assert.NoError(t, err)
		apply := []*shared.TipSet{ts}
		go func() {
			assert.NoError(t, ms.ProcessNewHead(ctx, apply))
		}()

		headChange := <-ms.headChans
		assert.Equal(t, apply, headChange.apply)
		assert.Nil(t, headChange.revert)
		headChange.done <- nil
	})

	t.Run("head not change", func(t *testing.T) {
		ms := newMessageService(msh, filestore.NewMockFileStore(t.TempDir()))
		ts, err := msh.fullNode.ChainHead(ctx)
		assert.NoError(t, err)
		ms.tsCache.Add(ts)
		apply := []*shared.TipSet{ts}
		go func() {
			assert.NoError(t, ms.ProcessNewHead(ctx, apply))
		}()
		headChange := <-ms.headChans
		assert.Len(t, headChange.apply, 0)
		assert.Len(t, headChange.revert, 0)
		headChange.done <- nil
	})

	getExpectTS := func(currTS *shared.TipSet, expectHeight abi.ChainEpoch) (*shared.TipSet, map[abi.ChainEpoch]shared.TipSet, error) {
		expectTS := currTS
		tsMap := make(map[abi.ChainEpoch]shared.TipSet, expectHeight)
		ticker := time.NewTicker(blockDelay)
		defer ticker.Stop()

		ctx, cancel := context.WithTimeout(ctx, blockDelay*2*time.Duration(expectHeight))
		defer cancel()

		for expectTS.Height() < expectHeight {
			select {
			case <-ticker.C:
				expectTS, err = msh.fullNode.ChainHead(ctx)
				assert.NoError(t, err)
				tsMap[expectTS.Height()] = *expectTS
			case <-ctx.Done():
				return nil, nil, fmt.Errorf("context done: %v", err)
			}
		}
		return expectTS, tsMap, nil
	}

	t.Run("head no gap", func(t *testing.T) {
		ms := newMessageService(msh, filestore.NewMockFileStore(t.TempDir()))
		ts, err := msh.fullNode.ChainHead(ctx)
		assert.NoError(t, err)
		ms.tsCache.Add(ts)

		expectHeight := ts.Height() + 1
		expectTS, _, err := getExpectTS(ts, expectHeight)
		assert.NoError(t, err)

		apply := []*shared.TipSet{expectTS}
		go func() {
			assert.NoError(t, ms.ProcessNewHead(ctx, apply))
		}()
		headChange := <-ms.headChans
		assert.Equal(t, apply, headChange.apply)
		assert.Nil(t, headChange.revert)
		headChange.done <- nil
	})

	t.Run("normal", func(t *testing.T) {
		ms := newMessageService(msh, filestore.NewMockFileStore(t.TempDir()))
		ts, err := msh.fullNode.ChainHead(ctx)
		assert.NoError(t, err)
		ms.tsCache.Add(ts)

		expectHeight := ts.Height() + 5
		expectTS, tsMap, err := getExpectTS(ts, expectHeight)
		assert.NoError(t, err)
		apply := []*shared.TipSet{expectTS}
		go func() {
			assert.NoError(t, ms.ProcessNewHead(ctx, apply))
		}()
		headChange := <-ms.headChans
		assert.Len(t, headChange.apply, int(expectTS.Height()-ts.Height()))
		assert.Len(t, headChange.revert, 0)
		for _, ts := range headChange.apply {
			assert.Equal(t, tsMap[ts.Height()], *ts)
		}
		headChange.done <- nil
	})

	t.Run("test with revert apply multiple tipset", func(t *testing.T) {
		ms := newMessageService(msh, filestore.NewMockFileStore(t.TempDir()))
		ts, err := msh.fullNode.ChainHead(ctx)
		assert.NoError(t, err)
		ms.tsCache.Add(ts)

		expectTS := ts
		expectHeight := abi.ChainEpoch(10) + ts.Height()
		revertHeight := abi.ChainEpoch(7) + ts.Height()

		ticker := time.NewTicker(blockDelay)
		defer ticker.Stop()

		ctx, cancel := context.WithTimeout(ctx, blockDelay*2*time.Duration(expectHeight))
		defer cancel()

		revertSignal := &testhelper.RevertSignal{ExpectRevertCount: 4, RevertedTS: make(chan []*shared.TipSet, 1)}

		for expectTS.Height() < expectHeight {
			select {
			case <-ticker.C:
				expectTS, err = msh.fullNode.ChainHead(ctx)
				assert.NoError(t, err)
				if expectTS.Height() < revertHeight {
					ms.tsCache.Add(expectTS)
				} else if expectTS.Height() == revertHeight {
					msh.fullNode.SendRevertSignal(revertSignal)
				}
			case <-ctx.Done():
				t.Errorf("not found tipset")
			}
		}
		revertedTS := <-revertSignal.RevertedTS
		go func() {
			assert.NoError(t, ms.ProcessNewHead(ctx, []*shared.TipSet{expectTS}))
		}()

		headChange := <-ms.headChans
		assert.Len(t, headChange.apply, int(expectTS.Height()-revertedTS[len(revertedTS)-1].Height())+1)
		revert := headChange.revert
		apply := headChange.apply
		assert.Equal(t, revert[len(revert)-1].Parents(), apply[len(apply)-1].Parents())
		assert.Equal(t, revertedTS[len(revertedTS)-1], revert[len(revert)-1])
		headChange.done <- nil
	})

	t.Run("test with revert", func(t *testing.T) {
		testRevert := func(revertFrom, applyFrom int) {
			ms := newMessageService(msh, filestore.NewMockFileStore(t.TempDir()))
			var tipSets []*shared.TipSet
			var revert []*shared.TipSet
			var parent []cid.Cid
			for i := 0; i < 6; i++ {
				ts, err := testhelper.GenTipset(abi.ChainEpoch(i), 1, parent)
				assert.NoError(t, err)
				parent = ts.Cids()
				tipSets = append(tipSets, ts)
				ms.tsCache.Add(ts)
				if i > revertFrom {
					revert = append(revert, ts)
				}
			}

			var apply []*shared.TipSet
			parent = tipSets[applyFrom-1].Cids()
			for i := applyFrom; i < 6; i++ {
				ts, err := testhelper.GenTipset(abi.ChainEpoch(i), 2, parent)
				assert.NoError(t, err)
				parent = ts.Cids()
				apply = append(apply, ts)
			}

			sort.Slice(apply, func(i, j int) bool {
				return apply[i].Height() > apply[j].Height()
			})

			full := v1Mock.NewMockFullNode(gomock.NewController(t))
			ms.nodeClient = full
			full.EXPECT().ChainGetTipSet(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(arg0 context.Context, arg1 shared.TipSetKey) (*shared.TipSet, error) {
				for _, ts := range tipSets {
					if ts.Key().Equals(arg1) {
						return ts, nil
					}
				}
				return nil, errors.New("not found tipset")
			})

			go func() {
				assert.NoError(t, ms.ProcessNewHead(context.Background(), apply))
			}()

			headChange := <-ms.headChans
			headChange.done <- nil
			assert.EqualValues(t, headChange.apply, apply)
			sort.Slice(revert, func(i, j int) bool {
				return revert[i].Height() > revert[j].Height()
			})
			assert.EqualValues(t, headChange.revert, revert)
		}
		// 1,2,3,4,5
		// revert 3,4,5
		testRevert(2, 3)

		// apply 5
		// revert 5
		testRevert(4, 5)
	})
}

func TestMessageService_PushMessage(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.DefaultConfig()
	cfg.MessageService.WaitingChainHeadStableDuration = time.Second * 2
	blockDelay := cfg.MessageService.WaitingChainHeadStableDuration * 2
	fsRepo := filestore.NewMockFileStore(t.TempDir())
	authClient := testhelper.NewMockAuthClient()
	msh, err := newMessageServiceHelper(ctx, cfg, blockDelay, fsRepo, authClient)
	assert.NoError(t, err)

	account := defaultLocalToken
	addr := testutil.BlsAddressProvider()(t)
	assert.NoError(t, msh.fullNode.AddActors([]address.Address{addr}))
	authClient.AddMockUserAndSigner(account, []address.Address{addr})
	assert.NoError(t, msh.walletProxy.AddAddress(account, []address.Address{addr}))

	lc := fxtest.NewLifecycle(t)
	_ = StartNodeEvents(lc, msh.fullNode, msh.ms)
	assert.NoError(t, lc.Start(ctx))
	defer lc.RequireStop()

	var pushedMsg *types.Message

	t.Run("push message:", func(t *testing.T) {
		// stm: @MESSENGER_SERVICE_PUSH_MESSAGE_001, @MESSENGER_SERVICE_PUSH_MESSAGE_002,
		// stm: @MESSENGER_SERVICE_PUSH_MESSAGE_WITH_ID_001, @MESSENGER_SERVICE_PUSH_MESSAGE_WITH_ID_002
		// stm: @MESSENGER_SERVICE_GET_MESSAGE_BY_UID_001, @MESSENGER_SERVICE_GET_MESSAGE_BY_UID_002
		// stm: @MESSENGER_SERVICE_LIST_MESSAGE_001
		rawMsg := testhelper.NewUnsignedMessage()
		rawMsg.From = addr
		uidStr, err := msh.ms.PushMessage(ctx, &rawMsg, nil)
		assert.NoError(t, err)
		_, err = shared.ParseUUID(uidStr)
		assert.NoError(t, err)

		// pushing message would be failed
		pushFailedMsg := testhelper.NewUnsignedMessage()
		_, err = msh.ms.PushMessage(ctx, &pushFailedMsg, nil)
		assert.Error(t, err)
		// msg with uuid not exists, expect an error
		_, err = msh.ms.GetMessageByUid(ctx, shared.NewUUID().String())
		assert.Error(t, err)

		pushedMsg, err = msh.ms.GetMessageByUid(ctx, uidStr)
		assert.NoError(t, err)
		assert.Equal(t, pushedMsg.ID, uidStr)

		{ // list messages
			msgs, err := msh.ms.ListMessage(ctx)
			assert.NoError(t, err)
			assert.Equal(t, len(msgs), 1)
			assert.Equal(t, msgs[0].ID, uidStr)

			msgs, err = msh.ms.ListMessageByAddress(ctx, addr)
			assert.NoError(t, err)
			assert.Equal(t, len(msgs), 1)
			assert.Equal(t, msgs[0].ID, uidStr)
		}
	})

	t.Run("wait message:", func(t *testing.T) {
		// stm: @MESSENGER_SERVICE_WAIT_MESSAGE_001, @MESSENGER_SERVICE_WAIT_MESSAGE_002
		ctx, cancel := context.WithTimeout(context.TODO(), time.Minute*3)
		defer cancel()
		_, err := waitMsgWithTimeout(ctx, msh.ms, shared.NewUUID().String())
		assert.Error(t, err)

		wg := sync.WaitGroup{}

		waitOneMsg := func(msgID string, expectErr bool) {
			wg.Add(1)
			go func() {
				defer wg.Done()
				res, err := waitMsgWithTimeout(ctx, msh.ms, msgID)
				if expectErr {
					assert.Error(t, err)
					return
				}
				assert.Equal(t, res.ID, msgID)
				msgLookup, err := msh.fullNode.StateSearchMsg(ctx, shared.EmptyTSK, *res.SignedCid, constants.LookbackNoLimit, true)
				assert.NoError(t, err)
				assert.Equal(t, msgLookup.Height, abi.ChainEpoch(res.Height))
				assert.Equal(t, msgLookup.TipSet, res.TipSetKey)
				assert.Equal(t, msgLookup.Receipt, *res.Receipt)
			}()
		}

		waitOneMsg(pushedMsg.ID, false)
		waitOneMsg(shared.NewUUID().String(), true)
		wg.Wait()
	})
}

func newMessageService(msh *messageServiceHelper, fsRepo filestore.FSRepo) *MessageService {
	return &MessageService{
		repo:           msh.ms.repo,
		fsRepo:         fsRepo,
		nodeClient:     msh.fullNode,
		addressService: msh.ms.addressService,
		walletClient:   msh.walletProxy,
		Pubsub:         msh.ms.Pubsub,
		triggerPush:    msh.ms.triggerPush,
		headChans:      make(chan *headChan, 10),
		tsCache:        newTipsetCache(),
	}
}
