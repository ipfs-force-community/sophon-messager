package service

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/fx/fxtest"

	"github.com/filecoin-project/venus-messager/config"
	"github.com/filecoin-project/venus-messager/filestore"
	"github.com/filecoin-project/venus-messager/testhelper"

	"github.com/filecoin-project/venus/pkg/constants"
	shared "github.com/filecoin-project/venus/venus-shared/types"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
)

func TestDoRefreshMessageState(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.DefaultConfig()
	cfg.MessageService.WaitingChainHeadStableDuration = time.Second * 2
	blockDelay := cfg.MessageService.WaitingChainHeadStableDuration * 2
	fsRepo := filestore.NewMockFileStore(t.TempDir())
	authClient := testhelper.NewMockAuthClient()
	msh, err := newMessageServiceHelper(ctx, cfg, blockDelay, fsRepo, authClient)
	assert.NoError(t, err)

	addrCount := 10
	addrs := testhelper.ResolveAddrs(t, testhelper.RandAddresses(t, addrCount))
	authClient.AddMockUserAndSigner(defaultLocalToken, addrs)
	assert.NoError(t, msh.walletProxy.AddAddress(defaultLocalToken, addrs))
	assert.NoError(t, msh.fullNode.AddActors(addrs))

	lc := fxtest.NewLifecycle(t)
	_ = StartNodeEvents(lc, msh.fullNode, msh.ms)
	assert.NoError(t, lc.Start(ctx))
	defer lc.RequireStop()

	t.Run("normal", func(t *testing.T) {
		ctx, calcel := context.WithTimeout(ctx, time.Minute*3)
		defer calcel()
		wg := sync.WaitGroup{}
		for i := 0; i < 10; i++ {
			wg.Add(1)
			msgs := genMessages(addrs, len(addrs)*10)
			assert.NoError(t, pushMessage(ctx, msh.ms, msgs))
			go func(msgs []*types.Message) {
				defer wg.Done()

				for _, msg := range msgs {
					res := waitMsgAndCheck(ctx, t, msg.ID, msh.ms)

					msgLookup, err := msh.fullNode.StateSearchMsg(ctx, shared.EmptyTSK, *res.SignedCid, constants.LookbackNoLimit, true)
					assert.NoError(t, err)
					assert.Equal(t, msgLookup.Height, abi.ChainEpoch(res.Height))
					assert.Equal(t, msgLookup.TipSet, res.TipSetKey)
					assert.Equal(t, msgLookup.Receipt, *res.Receipt)
				}
			}(msgs)
		}
		wg.Wait()
	})

	t.Run("revert", func(t *testing.T) {
		ticker := time.NewTicker(blockDelay)
		defer ticker.Stop()

		loop := 10
		i := 0
		rs := &testhelper.RevertSignal{ExpectRevertCount: 3, RevertedTS: make(chan []*shared.TipSet, 1)}
		for i < loop {
			select {
			case <-ticker.C:
				msgs := genMessages(addrs, len(addrs)*2*(i+1))
				assert.NoError(t, pushMessage(ctx, msh.ms, msgs))
				if i == 6 {
					msh.fullNode.SendRevertSignal(rs)
				}
				i++
			case <-ctx.Done():
				return
			}
		}
		revertedTs := <-rs.RevertedTS
		mayRevertMsg := make(map[cid.Cid]shared.TipSetKey, 0)
		for _, ts := range revertedTs {
			msgs, err := msh.fullNode.ChainGetMessagesInTipset(ctx, ts.Key())
			assert.NoError(t, err)
			for _, msg := range msgs {
				mayRevertMsg[msg.Cid] = ts.Key()
			}
		}

		time.Sleep(blockDelay*2 + time.Second)

		revertedMsgCount := 0
		for signedCID, tsk := range mayRevertMsg {
			res, err := msh.ms.GetMessageBySignedCid(ctx, signedCID)
			assert.NoError(t, err)
			if !res.TipSetKey.Equals(tsk) {
				revertedMsgCount++
				assert.Equal(t, types.OnChainMsg, res.State)
				msgLookup, err := msh.fullNode.StateSearchMsg(ctx, shared.EmptyTSK, signedCID, constants.LookbackNoLimit, true)
				assert.NoError(t, err)
				assert.Equal(t, msgLookup.Height, abi.ChainEpoch(res.Height))
				assert.Equal(t, msgLookup.TipSet, res.TipSetKey)
				assert.Equal(t, msgLookup.Receipt, *res.Receipt)
			}
		}
		assert.Greater(t, revertedMsgCount, 1)
	})

	t.Run("replace message", func(t *testing.T) {
		ctx, cancel := context.WithCancel(ctx)
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

		msgs := genMessages(addrs, len(addrs)*10)
		assert.NoError(t, pushMessage(ctx, ms, msgs))

		ts, err := msh.fullNode.ChainHead(ctx)
		assert.NoError(t, err)
		selectResult, err := ms.messageSelector.SelectMessage(ctx, ts)
		assert.NoError(t, err)
		assert.Len(t, selectResult.SelectMsg, len(addrs)*10)
		assert.Len(t, selectResult.ErrMsg, 0)
		assert.Len(t, selectResult.ModifyAddress, len(addrs))
		assert.Len(t, selectResult.ExpireMsg, 0)
		assert.Len(t, selectResult.ToPushMsg, 0)
		testhelper.IsSortedByNonce(t, selectResult.SelectMsg)

		conflictCount := 20
		type conflictMessage struct {
			srcMsgs      []*types.Message
			replacedMsgs []*types.Message
		}
		cm := &conflictMessage{}
		addrMsgs := testhelper.MsgGroupByAddress(selectResult.SelectMsg)
		idx := 0
		count := 0
		for count < conflictCount {
			for _, msgs := range addrMsgs {
				msg := msgs[idx]
				cm.srcMsgs = append(cm.srcMsgs, msg)
				msgCopy := *msg
				msgCopy.Method = 1
				msgCopy.GasLimit = int64(float64(msgCopy.GasLimit) * 1.5)
				msgCopy.GasFeeCap = big.Mul(msgCopy.GasFeeCap, big.NewInt(2))
				c := msgCopy.Message.Cid()
				msgCopy.UnsignedCid = &c
				signedCID := (&shared.SignedMessage{
					Message:   msgCopy.Message,
					Signature: *msg.Signature,
				}).Cid()
				msgCopy.SignedCid = &signedCID
				cm.replacedMsgs = append(cm.replacedMsgs, &msgCopy)
				count++
				continue
			}
			idx++
		}

		assert.NoError(t, saveMsgsToDB(ctx, ms, selectResult))
		for _, msg := range cm.replacedMsgs {
			selectResult.ToPushMsg = append(selectResult.ToPushMsg, &shared.SignedMessage{
				Message:   msg.Message,
				Signature: *msg.Signature,
			})
		}

		ctx, calcel := context.WithTimeout(ctx, time.Minute*3)
		defer calcel()

		go func() {
			ms.multiPushMessages(ctx, selectResult)
		}()
		for i, msg := range cm.srcMsgs {
			res, err := waitMsgWithTimeout(ctx, msh.ms, msg.ID)
			assert.NoError(t, err)
			assert.Equal(t, types.ReplacedMsg, res.State)
			assert.Equal(t, msg.Method, res.Method)
			assert.Equal(t, msg.GasLimit, res.GasLimit)
			assert.Equal(t, msg.GasFeeCap, res.GasFeeCap)
			assert.Equal(t, msg.UnsignedCid, res.UnsignedCid)
			assert.Equal(t, msg.SignedCid, res.SignedCid)

			msgLookup, err := msh.fullNode.StateSearchMsg(ctx, shared.EmptyTSK, *cm.replacedMsgs[i].SignedCid, constants.LookbackNoLimit, true)
			assert.NoError(t, err)
			assert.Equal(t, msgLookup.Height, abi.ChainEpoch(res.Height))
			assert.Equal(t, msgLookup.TipSet, res.TipSetKey)
			assert.Equal(t, msgLookup.Receipt, *res.Receipt)
		}
	})

	t.Run("tow message nonce is zero", func(t *testing.T) {
		ctx, cancel := context.WithCancel(ctx)
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

		addrCount := 1
		addrs := testhelper.ResolveAddrs(t, testhelper.RandAddresses(t, addrCount))
		authClient.AddMockUserAndSigner(defaultLocalToken, addrs)
		assert.NoError(t, msh.walletProxy.AddAddress(defaultLocalToken, addrs))
		assert.NoError(t, msh.fullNode.AddActors(addrs))

		lc := fxtest.NewLifecycle(t)
		_ = StartNodeEvents(lc, msh.fullNode, ms)
		assert.NoError(t, lc.Start(ctx))
		defer lc.RequireStop()

		// first message will estimate gas failed
		// second message will on chain
		// both messages nonce is 0
		msgs := genMessages(addrs, 2)
		msg := msgs[0]
		msg.GasLimit = -1
		assert.NoError(t, pushMessage(ctx, ms, msgs))

		ts, err := msh.fullNode.ChainHead(ctx)
		assert.NoError(t, err)
		selectResult, err := ms.messageSelector.SelectMessage(ctx, ts)
		assert.NoError(t, err)
		assert.Len(t, selectResult.SelectMsg, 1)
		assert.Equal(t, msgs[1].ID, selectResult.SelectMsg[0].ID)
		assert.Len(t, selectResult.ErrMsg, 1)
		assert.Equal(t, msgs[0].ID, selectResult.ErrMsg[0].id)

		assert.NoError(t, saveMsgsToDB(ctx, ms, selectResult))

		ctx, calcel := context.WithTimeout(ctx, time.Minute*3)
		defer calcel()
		go func() {
			ms.multiPushMessages(ctx, selectResult)
		}()

		fillMsg := selectResult.SelectMsg[0]
		res, err := waitMsgWithTimeout(ctx, msh.ms, fillMsg.ID)
		assert.NoError(t, err)
		assert.Equal(t, fillMsg.ID, res.ID)
		assert.Equal(t, types.OnChainMsg, res.State)
		assert.Equal(t, fillMsg.UnsignedCid, res.UnsignedCid)
		assert.Equal(t, fillMsg.SignedCid, res.SignedCid)
	})
}

func TestUpdateMessageState(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.DefaultConfig()
	cfg.MessageService.WaitingChainHeadStableDuration = time.Second * 2
	blockDelay := cfg.MessageService.WaitingChainHeadStableDuration * 2
	fsRepo := filestore.NewMockFileStore(t.TempDir())
	authClient := testhelper.NewMockAuthClient()
	msh, err := newMessageServiceHelper(ctx, cfg, blockDelay, fsRepo, authClient)
	assert.NoError(t, err)

	addrCount := 10
	addrs := testhelper.ResolveAddrs(t, testhelper.RandAddresses(t, addrCount))
	authClient.AddMockUserAndSigner(defaultLocalToken, addrs)
	assert.NoError(t, msh.walletProxy.AddAddress(defaultLocalToken, addrs))
	assert.NoError(t, msh.fullNode.AddActors(addrs))

	lc := fxtest.NewLifecycle(t)
	_ = StartNodeEvents(lc, msh.fullNode, msh.ms)
	assert.NoError(t, lc.Start(ctx))
	defer lc.RequireStop()

	msgs := genMessages(addrs, len(addrs)*10*5)
	assert.NoError(t, pushMessage(ctx, msh.ms, msgs))

	ts, err := msh.fullNode.ChainHead(ctx)
	assert.NoError(t, err)

	ctx, cancel2 := context.WithTimeout(ctx, time.Minute*3)
	defer cancel2()

	for _, msg := range msgs {
		res, err := waitMsgWithTimeout(ctx, msh.ms, msg.ID)
		assert.NoError(t, err)

		assert.Equal(t, types.OnChainMsg, res.State)
		msgLookup, err := msh.fullNode.StateSearchMsg(ctx, ts.Key(), *res.SignedCid, constants.LookbackNoLimit, true)
		assert.NoError(t, err)
		assert.Equal(t, msgLookup.Height, abi.ChainEpoch(res.Height))
		assert.Equal(t, msgLookup.TipSet, res.TipSetKey)
		assert.Equal(t, msgLookup.Receipt, *res.Receipt)
	}
}
