package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/venus/pkg/constants"
	shared "github.com/filecoin-project/venus/venus-shared/types"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
	"github.com/stretchr/testify/assert"
	"go.uber.org/fx/fxtest"

	"github.com/filecoin-project/venus-messager/config"
	"github.com/filecoin-project/venus-messager/filestore"
	"github.com/filecoin-project/venus-messager/testhelper"
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

	msh, err := newMessageServiceHelper(ctx, cfg, blockDelay, fsRepo)
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

	blockedMsgs := make(map[string]*types.Message, 0)
	msgs := genMessages(addrs, defaultLocalToken, len(addrs)*10)
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
	for _, msg := range replacedMsgs {
		res, err := ms.WaitMessage(ctx, msg.ID, constants.MessageConfidence)
		assert.NoError(t, err)
		assert.Equal(t, msg.GasLimit, res.GasLimit)
		assert.Equal(t, msg.GasFeeCap, res.GasFeeCap)
		assert.Equal(t, msg.GasPremium, res.GasPremium)
	}
}

func TestReconnectCheck(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.DefaultConfig()
	cfg.MessageService.WaitingChainHeadStableDuration = time.Second * 2
	blockDelay := cfg.MessageService.WaitingChainHeadStableDuration * 2
	fsRepo := filestore.NewMockFileStore(t.TempDir())
	msh, err := newMessageServiceHelper(ctx, cfg, blockDelay, fsRepo)
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
				t.Log(expectTS.Height(), time.Now().String())
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.DefaultConfig()
	cfg.MessageService.WaitingChainHeadStableDuration = time.Second * 2
	blockDelay := cfg.MessageService.WaitingChainHeadStableDuration * 2
	fsRepo := filestore.NewMockFileStore(t.TempDir())
	msh, err := newMessageServiceHelper(ctx, cfg, blockDelay, fsRepo)
	assert.NoError(t, err)

	t.Run("tipset cache is empty", func(t *testing.T) {
		ms := newMessageService(msh, filestore.NewMockFileStore(t.TempDir()))
		ts, err := msh.fullNode.ChainHead(ctx)
		assert.NoError(t, err)
		apply := []*shared.TipSet{ts}
		go func() {
			assert.NoError(t, ms.ProcessNewHead(ctx, apply, nil))
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
			assert.NoError(t, ms.ProcessNewHead(ctx, apply, nil))
		}()
		t.Log(ms.tsCache.Cache[int64(ts.Height())].Height(), ts.Height())
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
		apply := []*shared.TipSet{expectTS}
		go func() {
			assert.NoError(t, ms.ProcessNewHead(ctx, apply, nil))
		}()
		headChange := <-ms.headChans
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
			assert.NoError(t, ms.ProcessNewHead(ctx, []*shared.TipSet{expectTS}, nil))
		}()

		headChange := <-ms.headChans
		assert.Len(t, headChange.apply, int(expectTS.Height()-revertedTS[len(revertedTS)-1].Height())+1)
		revert := headChange.revert
		apply := headChange.apply
		assert.Equal(t, revert[len(revert)-1].Parents(), apply[len(apply)-1].Parents())
		assert.Equal(t, revertedTS[len(revertedTS)-1], revert[len(revert)-1])
		headChange.done <- nil
	})
}

func newMessageService(msh *messageServiceHelper, fsRepo filestore.FSRepo) *MessageService {
	return &MessageService{
		repo:           msh.ms.repo,
		log:            msh.ms.log,
		fsRepo:         fsRepo,
		nodeClient:     msh.fullNode,
		messageState:   msh.ms.messageState,
		addressService: msh.ms.addressService,
		walletClient:   msh.walletProxy,
		Pubsub:         msh.ms.Pubsub,
		triggerPush:    msh.ms.triggerPush,
		headChans:      make(chan *headChan, 10),
		tsCache:        newTipsetCache(),
	}
}
