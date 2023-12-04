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

	"github.com/filecoin-project/go-address"
	v1Mock "github.com/filecoin-project/venus/venus-shared/api/chain/v1/mock"
	"go.uber.org/fx/fxtest"

	"github.com/golang/mock/gomock"

	"github.com/ipfs/go-cid"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"

	"github.com/stretchr/testify/assert"

	"github.com/ipfs-force-community/sophon-messager/config"
	"github.com/ipfs-force-community/sophon-messager/filestore"
	"github.com/ipfs-force-community/sophon-messager/gateway"
	"github.com/ipfs-force-community/sophon-messager/models"
	"github.com/ipfs-force-community/sophon-messager/models/repo"
	"github.com/ipfs-force-community/sophon-messager/publisher"
	"github.com/ipfs-force-community/sophon-messager/testhelper"

	"github.com/filecoin-project/venus/pkg/constants"
	"github.com/filecoin-project/venus/venus-shared/testutil"
	shared "github.com/filecoin-project/venus/venus-shared/types"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
)

func TestVerifyNetworkName(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	msh := newMessageServiceHelper(ctx, t)
	ms := msh.MessageService
	assert.NoError(t, msh.MessageService.tsCache.Save(msh.fsRepo.TipsetFile()))

	tipsetCache := &TipsetCache{
		NetworkName: string(shared.NetworkNameMain),
	}
	assert.NoError(t, tipsetCache.Save(msh.fsRepo.TipsetFile()))

	networkName, err := msh.fullNode.StateNetworkName(ctx)
	assert.NoError(t, err)
	ms.tsCache.NetworkName = string(shared.NetworkNameButterfly)
	err = ms.verifyNetworkName()
	expectErrStr := fmt.Sprintf("network name not match, expect %s, actual %s, please remove `%s`",
		networkName, ms.tsCache.NetworkName, msh.fsRepo.TipsetFile())
	assert.Equal(t, expectErrStr, err.Error())
}

func TestReplaceMessage(t *testing.T) {
	// stm: @MESSENGER_SERVICE_REPLACE_MESSAGE_001, @MESSENGER_SERVICE_REPLACE_MESSAGE_002
	// stm: @MESSENGER_SERVICE_REPLACE_MESSAGE_003, @MESSENGER_SERVICE_REPLACE_MESSAGE_004
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	msh := newMessageServiceHelper(ctx, t)
	addrs := msh.genAddresses()
	ms := msh.MessageService
	msh.start()
	defer msh.stop()

	smallerPremium := big.Sub(testhelper.MinPackedPremium, big.NewInt(20))
	blockedMsgs := make(map[string]*types.Message, 0)
	msgs := genMessages(addrs, len(addrs)*10)
	for i, msg := range msgs {
		if i%2 == 0 {
			msg.GasPremium = smallerPremium
			blockedMsgs[msg.ID] = msg
		}
	}
	assert.NoError(t, pushMessage(ctx, ms, msgs))

	ts, err := msh.fullNode.ChainHead(ctx)
	assert.NoError(t, err)

	selectResult := selectMsgWithAddress(ctx, t, msh, addrs, ts)
	assert.Len(t, selectResult.SelectMsg, len(msgs))

	notBlockedMsgs := make([]*types.Message, 0)
	for _, msg := range selectResult.SelectMsg {
		if _, ok := blockedMsgs[msg.ID]; !ok {
			notBlockedMsgs = append(notBlockedMsgs, msg)
		}
	}
	ms.msgSelectMgr.msgReceiver <- selectResult.ToPushMsg
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

		// check gas premium
		defPremium := testhelper.DefGasPremium
		expectPremium := big.Div(big.Mul(smallerPremium, big.NewInt(int64(testhelper.DefReplaceByFeePercent))), big.NewInt(100))
		if msg.Meta != nil && msg.Meta.GasOverPremium != 0 {
			gasOverPremium := big.Mul(big.NewInt(int64(100*msg.Meta.GasOverPremium)), big.NewInt(100))
			expectPremium = big.Mul(expectPremium, gasOverPremium)
			defPremium = big.Mul(defPremium, gasOverPremium)
		}
		assert.Equal(t, big.Max(expectPremium, defPremium), res.GasPremium)
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

	msh := newMessageServiceHelper(ctx, t)

	t.Run("tipset cache is empty", func(t *testing.T) {
		ms := newMessageService(msh)
		ts, err := msh.fullNode.ChainHead(ctx)
		assert.NoError(t, err)
		assert.NoError(t, ms.ReconnectCheck(ctx, ts))
	})

	t.Run("head not change", func(t *testing.T) {
		ms := newMessageService(msh)
		ts, err := msh.fullNode.ChainHead(ctx)
		assert.NoError(t, err)
		ms.tsCache.Add(ts)
		go func() {
			assert.NoError(t, ms.ReconnectCheck(ctx, ts))
		}()
		<-ms.headChans

		next, err := testhelper.GenTipset(ts.Height()+1, 1, ts.Cids())
		assert.NoError(t, err)
		assert.NoError(t, ms.ReconnectCheck(ctx, next))
	})

	t.Run("normal", func(t *testing.T) {
		ms := newMessageService(msh)
		ts, err := msh.fullNode.ChainHead(ctx)
		assert.NoError(t, err)
		ms.tsCache.Add(ts)

		expectTS := ts
		expectHeight := abi.ChainEpoch(5) + ts.Height()
		tsMap := make(map[abi.ChainEpoch]shared.TipSet, 5)

		ticker := time.NewTicker(msh.blockDelay)
		defer ticker.Stop()

		ctx, cancel := context.WithTimeout(ctx, msh.blockDelay*2*time.Duration(expectHeight))
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
		assert.Len(t, headChange.apply, int(expectTS.Height()-ts.Height())-1)
		assert.Len(t, headChange.revert, 0)
		for _, ts := range headChange.apply {
			assert.Equal(t, tsMap[ts.Height()], *ts)
		}
		headChange.done <- nil
	})

	t.Run("test with revert", func(t *testing.T) {
		ms := newMessageService(msh)
		ts, err := msh.fullNode.ChainHead(ctx)
		assert.NoError(t, err)
		ms.tsCache.Add(ts)

		expectTS := ts
		expectHeight := abi.ChainEpoch(10) + ts.Height()
		revertHeight := abi.ChainEpoch(5) + ts.Height()

		ticker := time.NewTicker(msh.blockDelay)
		defer ticker.Stop()

		ctx, cancel := context.WithTimeout(ctx, msh.blockDelay*2*time.Duration(expectHeight))
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
		assert.Len(t, headChange.apply, int(expectTS.Height()-revertedTS[len(revertedTS)-1].Height()))
		revert := headChange.revert
		apply := headChange.apply
		pts, err := msh.nodeClient.ChainGetTipSet(ctx, apply[len(apply)-1].Parents())
		assert.NoError(t, err)
		assert.Equal(t, revert[len(revert)-1].Parents(), pts.Parents())
		assert.Equal(t, revertedTS[len(revertedTS)-1], revert[len(revert)-1])
		headChange.done <- nil
	})
}

func TestMessageService_ProcessNewHead(t *testing.T) {
	// stm: @MESSENGER_SERVICE_LIST_MESSAGE_BY_ADDRESS_001
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	msh := newMessageServiceHelper(ctx, t)

	t.Run("apply is empty", func(t *testing.T) {
		ms := newMessageService(msh)
		assert.NoError(t, ms.ProcessNewHead(ctx, nil))
	})

	t.Run("tipset cache is empty", func(t *testing.T) {
		ms := newMessageService(msh)
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
		ms := newMessageService(msh)
		ts, err := msh.fullNode.ChainHead(ctx)
		assert.NoError(t, err)
		ms.tsCache.Add(ts)

		next, err := testhelper.GenTipset(ts.Height()+1, 1, ts.Cids())
		assert.NoError(t, err)
		apply := []*shared.TipSet{next}
		assert.NoError(t, ms.ProcessNewHead(ctx, apply))
	})

	getExpectTS := func(currTS *shared.TipSet, expectHeight abi.ChainEpoch) (*shared.TipSet, map[abi.ChainEpoch]shared.TipSet, error) {
		expectTS := currTS
		tsMap := make(map[abi.ChainEpoch]shared.TipSet, expectHeight)
		ticker := time.NewTicker(msh.blockDelay)
		defer ticker.Stop()

		ctx, cancel := context.WithTimeout(ctx, msh.blockDelay*2*time.Duration(expectHeight))
		defer cancel()

		var err error
		for expectTS.Height() < expectHeight {
			select {
			case <-ticker.C:
				expectTS, err = msh.fullNode.ChainHead(ctx)
				assert.NoError(t, err)
				tsMap[expectTS.Height()] = *expectTS
			case <-ctx.Done():
				return nil, nil, fmt.Errorf("context done: %v", ctx.Err())
			}
		}
		return expectTS, tsMap, nil
	}

	t.Run("head no gap", func(t *testing.T) {
		ms := newMessageService(msh)
		ts, err := msh.fullNode.ChainHead(ctx)
		assert.NoError(t, err)
		ms.tsCache.Add(ts)

		expectHeight := ts.Height() + 1
		expectTS, _, err := getExpectTS(ts, expectHeight)
		assert.NoError(t, err)
		assert.Equal(t, expectHeight, expectTS.Height())

		assert.NoError(t, ms.ProcessNewHead(ctx, []*shared.TipSet{expectTS}))
	})

	t.Run("normal", func(t *testing.T) {
		ms := newMessageService(msh)
		ts, err := msh.fullNode.ChainHead(ctx)
		assert.NoError(t, err)
		ms.tsCache.Add(ts)

		expectHeight := ts.Height() + 5
		expectTS, tsMap, err := getExpectTS(ts, expectHeight)
		assert.NoError(t, err)
		assert.Equal(t, expectHeight, expectTS.Height())

		go func() {
			assert.NoError(t, ms.ProcessNewHead(ctx, []*shared.TipSet{expectTS}))
		}()
		headChange := <-ms.headChans

		var apply []*shared.TipSet
		for i := expectHeight; i > ts.Height()+1; i-- {
			val, ok := tsMap[i]
			assert.True(t, ok)
			apply = append(apply, &val)
		}
		assert.Equal(t, apply, headChange.apply)
		assert.Len(t, headChange.revert, 0)
		headChange.done <- nil
	})

	t.Run("test with revert apply multiple tipset", func(t *testing.T) {
		ms := newMessageService(msh)
		ts, err := msh.fullNode.ChainHead(ctx)
		assert.NoError(t, err)
		ms.tsCache.Add(ts)

		expectTS := ts
		expectHeight := abi.ChainEpoch(10) + ts.Height()
		revertHeight := abi.ChainEpoch(7) + ts.Height()

		ticker := time.NewTicker(msh.blockDelay)
		defer ticker.Stop()

		ctx, cancel := context.WithTimeout(ctx, msh.blockDelay*2*time.Duration(expectHeight))
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
		assert.Len(t, headChange.apply, int(expectTS.Height()-revertedTS[len(revertedTS)-1].Height()))
		revert := headChange.revert
		apply := headChange.apply
		pts, err := msh.nodeClient.ChainGetTipSet(ctx, apply[len(apply)-1].Parents())
		assert.NoError(t, err)
		assert.Equal(t, revert[len(revert)-1].Parents(), pts.Parents())
		assert.Equal(t, revertedTS[len(revertedTS)-1], revert[len(revert)-1])
		headChange.done <- nil
	})

	t.Run("test with revert", func(t *testing.T) {
		testRevert := func(revertFrom, applyFrom int) {
			ms := newMessageService(msh)
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
				tipSets = append(tipSets, ts)
				parent = ts.Cids()
				apply = append(apply, ts)
			}

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

			sort.Slice(revert, func(i, j int) bool {
				return revert[i].Height() > revert[j].Height()
			})

			headChange := <-ms.headChans
			headChange.done <- nil
			sort.Slice(apply, func(i, j int) bool {
				return apply[i].Height() > apply[j].Height()
			})
			assert.EqualValues(t, headChange.apply, apply[:len(apply)-1])
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

	msh := newMessageServiceHelper(ctx, t)
	addr := testutil.BlsAddressProvider()(t)
	msh.addAddresses([]address.Address{addr})
	ms := msh.MessageService
	msh.start()
	defer msh.stop()

	var pushedMsg *types.Message

	t.Run("push message:", func(t *testing.T) {
		// stm: @MESSENGER_SERVICE_PUSH_MESSAGE_001, @MESSENGER_SERVICE_PUSH_MESSAGE_002,
		// stm: @MESSENGER_SERVICE_PUSH_MESSAGE_WITH_ID_001, @MESSENGER_SERVICE_PUSH_MESSAGE_WITH_ID_002
		// stm: @MESSENGER_SERVICE_GET_MESSAGE_BY_UID_001, @MESSENGER_SERVICE_GET_MESSAGE_BY_UID_002
		// stm: @MESSENGER_SERVICE_LIST_MESSAGE_001
		rawMsg := testhelper.NewUnsignedMessage()
		rawMsg.From = addr
		rawMsg.To = addr
		uidStr, err := msh.PushMessage(ctx, &rawMsg, nil)
		assert.NoError(t, err)
		_, err = shared.ParseUUID(uidStr)
		assert.NoError(t, err)

		// pushing message would be failed
		pushFailedMsg := testhelper.NewUnsignedMessage()
		_, err = ms.PushMessage(ctx, &pushFailedMsg, nil)
		assert.Error(t, err)
		// msg with uuid not exists, expect an error
		_, err = ms.GetMessageByUid(ctx, shared.NewUUID().String())
		assert.Error(t, err)

		pushedMsg, err = ms.GetMessageByUid(ctx, uidStr)
		assert.NoError(t, err)
		assert.Equal(t, pushedMsg.ID, uidStr)

		{ // list messages
			msgs, err := ms.ListMessage(ctx, &repo.MsgQueryParams{})
			assert.NoError(t, err)
			assert.Equal(t, len(msgs), 1)
			assert.Equal(t, msgs[0].ID, uidStr)

			msgs, err = ms.ListMessageByAddress(ctx, addr)
			assert.NoError(t, err)
			assert.Equal(t, len(msgs), 1)
			assert.Equal(t, msgs[0].ID, uidStr)
		}
	})

	t.Run("wait message:", func(t *testing.T) {
		// stm: @MESSENGER_SERVICE_WAIT_MESSAGE_001, @MESSENGER_SERVICE_WAIT_MESSAGE_002
		ctx, cancel := context.WithTimeout(ctx, time.Minute*1)
		defer cancel()
		_, err := waitMsgWithTimeout(ctx, ms, shared.NewUUID().String())
		assert.Error(t, err)

		wg := sync.WaitGroup{}

		waitOneMsg := func(msgID string, expectErr bool) {
			wg.Add(1)
			go func() {
				defer wg.Done()
				res, err := waitMsgWithTimeout(ctx, ms, msgID)
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

type messageServiceHelper struct {
	ctx context.Context

	t  *testing.T
	lc *fxtest.Lifecycle

	fullNode    *testhelper.MockFullNode
	walletProxy *gateway.MockWalletProxy
	authClient  *testhelper.AuthClient
	*MessageService

	addrs      []address.Address
	token      string
	blockDelay time.Duration
}

type options struct {
	skipPushMessage bool
}
type opt = func(opts *options)

func skipPushMessage() opt {
	return func(opts *options) {
		opts.skipPushMessage = true
	}
}

func newMessageServiceHelper(ctx context.Context, t *testing.T, opts ...opt) *messageServiceHelper {
	opt := &options{}
	for _, o := range opts {
		o(opt)
	}

	cfg := config.DefaultConfig()
	cfg.MessageService.WaitingChainHeadStableDuration = time.Second * 1
	blockDelay := cfg.MessageService.WaitingChainHeadStableDuration * 2
	if opt.skipPushMessage {
		cfg.MessageService.SkipPushMessage = opt.skipPushMessage
	}

	fsRepo := filestore.NewMockFileStore(t.TempDir())
	assert.NoError(t, fsRepo.ReplaceConfig(cfg))

	fullNode, err := testhelper.NewMockFullNode(ctx, blockDelay)
	assert.NoError(t, err)

	repo, err := models.SetDataBase(fsRepo)
	assert.NoError(t, err)
	assert.NoError(t, repo.AutoMigrate())

	authClient := testhelper.NewMockAuthClient(t)
	walletProxy := gateway.NewMockWalletProxy()
	addressService := NewAddressService(repo, walletProxy, authClient)
	sharedParamsService, err := NewSharedParamsService(ctx, repo)
	assert.NoError(t, err)

	rpcPublisher := publisher.NewRpcPublisher(ctx, fullNode, repo.NodeRepo(), false, repo.MessageRepo())
	networkParams := &shared.NetworkParams{BlockDelaySecs: 30}
	msgPublisher, err := publisher.NewIMsgPublisher(ctx, networkParams, cfg.Publisher, nil, rpcPublisher)
	assert.NoError(t, err)

	msgReceiver, err := publisher.NewMessageReceiver(ctx, msgPublisher)
	assert.NoError(t, err)
	ms, err := NewMessageService(ctx, repo, fullNode, fsRepo, addressService, sharedParamsService,
		walletProxy, msgReceiver)
	assert.NoError(t, err)

	return &messageServiceHelper{
		ctx:            ctx,
		t:              t,
		lc:             fxtest.NewLifecycle(t),
		fullNode:       fullNode,
		walletProxy:    walletProxy,
		authClient:     authClient,
		MessageService: ms,
		token:          defaultLocalToken,
		blockDelay:     blockDelay,
	}
}

func (msh *messageServiceHelper) start() {
	_ = StartNodeEvents(msh.lc, msh.fullNode, msh.MessageService)
	assert.NoError(msh.t, msh.lc.Start(msh.ctx))
}

func (msh *messageServiceHelper) stop() {
	msh.lc.RequireStop()
}

func (msh *messageServiceHelper) genAddresses() []address.Address {
	addrCount := 10
	addrs := testhelper.ResolveAddrs(msh.t, testhelper.RandAddresses(msh.t, addrCount))
	msh.addAddresses(addrs)

	return addrs
}

func (msh *messageServiceHelper) addAddresses(addrs []address.Address) {
	account := msh.token
	msh.addrs = addrs
	msh.authClient.Init(account, addrs)
	assert.NoError(msh.t, msh.walletProxy.AddAddress(account, addrs))
	assert.NoError(msh.t, msh.fullNode.AddActors(addrs))
}

func (msh *messageServiceHelper) genAndPushMessages(count int) []*types.Message {
	msgs := genMessages(msh.addrs, count)
	assert.NoError(msh.t, pushMessage(msh.ctx, msh.MessageService, msgs))

	return msgs
}

func newMessageService(msh *messageServiceHelper) *MessageService {
	return &MessageService{
		repo:           msh.MessageService.repo,
		fsRepo:         filestore.NewMockFileStore(msh.t.TempDir()),
		nodeClient:     msh.fullNode,
		addressService: msh.MessageService.addressService,
		walletClient:   msh.walletProxy,
		triggerPush:    msh.MessageService.triggerPush,
		headChans:      make(chan *headChan, 10),
		tsCache:        newTipsetCache(),
	}
}
