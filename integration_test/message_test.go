package integration

import (
	"context"
	"errors"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/venus-auth/auth"
	"github.com/filecoin-project/venus-auth/core"
	"github.com/filecoin-project/venus-auth/jwtclient"
	"github.com/gbrlsnchs/jwt/v3"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/xerrors"
	"gorm.io/gorm"

	"github.com/filecoin-project/venus-messager/config"
	"github.com/filecoin-project/venus-messager/service"
	"github.com/filecoin-project/venus-messager/testhelper"

	"github.com/filecoin-project/venus/pkg/constants"
	"github.com/filecoin-project/venus/venus-shared/api/messager"
	"github.com/filecoin-project/venus/venus-shared/testutil"
	shared "github.com/filecoin-project/venus/venus-shared/types"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
)

func TestHasMessageByUid(t *testing.T) {
	p := prepare(t)
	defer p.closer()

	t.Run("test has message by uid", func(t *testing.T) {
		testHasMessageByUid(p.ctx, t, p.apiAdmin, p.apiSign, p.addrs)
	})
	assert.NoError(t, p.ms.stop(p.ctx))
}

func TestMarkBadMessage(t *testing.T) {
	p := prepare(t)
	defer p.closer()

	t.Run("test mark bad message", func(t *testing.T) {
		testMarkBadMessage(p.ctx, t, p.apiAdmin, p.apiSign, p.addrs, p.blockDelay)
	})
	assert.NoError(t, p.ms.stop(p.ctx))
}

func TestUpdateFilledMessageByID(t *testing.T) {
	p := prepare(t)
	defer p.closer()

	t.Run("test update filled message by id", func(t *testing.T) {
		testUpdateFilledMessageByID(p.ctx, t, p.apiAdmin, p.apiSign, p.addrs, p.blockDelay)
	})
	assert.NoError(t, p.ms.stop(p.ctx))
}

func TestListMessage(t *testing.T) {
	p := prepare(t)
	defer p.closer()

	t.Run("test list message", func(t *testing.T) {
		testListMessage(p.ctx, t, p.apiAdmin, p.apiSign, p.addrs)
	})
	assert.NoError(t, p.ms.stop(p.ctx))
}

func TestMessageAPI(t *testing.T) {
	p := prepare(t)
	defer p.closer()

	t.Run("test push message", func(t *testing.T) {
		testPushMessage(p.ctx, t, p.apiAdmin, p.apiSign, p.addrs)
	})
	t.Run("test push message with id", func(t *testing.T) {
		testPushMessageWithID(p.ctx, t, p.apiAdmin, p.apiSign, p.addrs)
	})
	t.Run("test get message by uid", func(t *testing.T) {
		testGetMessageByUid(p.ctx, t, p.apiAdmin, p.apiSign, p.addrs)
	})
	t.Run("test wait message", func(t *testing.T) {
		testWaitMessage(p.ctx, t, p.apiAdmin, p.apiSign, p.addrs)
	})
	t.Run("test get message by signed cid", func(t *testing.T) {
		testGetMessageBySignedCID(p.ctx, t, p.apiAdmin, p.apiSign, p.addrs)
	})
	t.Run("test get message By unsigned cid", func(t *testing.T) {
		testGetMessageByUnsignedCID(p.ctx, t, p.apiAdmin, p.apiSign, p.addrs)
	})
	t.Run("test get message by from and nonce", func(t *testing.T) {
		testGetMessageByFromAndNonce(p.ctx, t, p.apiAdmin, p.apiSign, p.addrs)
	})
	t.Run("test list message by from state", func(t *testing.T) {
		testListMessageByFromState(p.ctx, t, p.apiAdmin, p.apiSign, p.addrs)
	})
	t.Run("test list message by address", func(t *testing.T) {
		testListMessageByAddress(p.ctx, t, p.apiAdmin, p.apiSign)
	})
	t.Run("test list failed message", func(t *testing.T) {
		testListFailedMessage(p.ctx, t, p.apiAdmin, p.apiSign, p.addrs, p.blockDelay)
	})
	t.Run("test list blocked message", func(t *testing.T) {
		testListBlockedMessage(p.ctx, t, p.apiAdmin, p.apiSign, p.addrs, p.blockDelay)
	})
	t.Run("test update message state by id", func(t *testing.T) {
		testUpdateMessageStateByID(p.ctx, t, p.apiAdmin, p.apiSign, p.addrs, p.blockDelay)
	})
	t.Run("test update all filled message", func(t *testing.T) {
		testUpdateAllFilledMessage(p.ctx, t, p.apiAdmin, p.apiSign, p.addrs, p.blockDelay)
	})
	t.Run("test replace message", func(t *testing.T) {
		testReplaceMessage(p.ctx, t, p.apiAdmin, p.apiSign, p.addrs, p.blockDelay)
	})
	t.Run("test recover failed msg", func(t *testing.T) {
		testRecoverFailedMsg(p.ctx, t, p.apiAdmin, p.apiSign, p.addrs, p.blockDelay)
	})

	assert.NoError(t, p.ms.stop(p.ctx))
}

func testPushMessage(ctx context.Context, t *testing.T, api, apiSign messager.IMessager, addrs []address.Address) {
	msgs := genMessageWithAddress(addrs)
	sendSpecs := testhelper.MockSendSpecs()

	for _, msg := range msgs {
		meta := sendSpecs[rand.Intn(len(sendSpecs))]
		id, err := api.PushMessage(ctx, &msg.Message, meta)
		assert.NoError(t, err)

		_, err = apiSign.PushMessage(ctx, &msg.Message, meta)
		assert.Equal(t, jwtclient.ErrorPermissionDeny.Error(), err.Error())

		res, err := api.GetMessageByUid(ctx, id)
		assert.NoError(t, err)
		checkUnsignedMsg(t, &msg.Message, &res.Message)
		checkSendSpec(t, meta, res.Meta)
	}
}

func testPushMessageWithID(ctx context.Context, t *testing.T, api, apiSign messager.IMessager, addrs []address.Address) {
	msgs := genMessageWithAddress(addrs)
	sendSpecs := testhelper.MockSendSpecs()

	for _, msg := range msgs {
		meta := sendSpecs[rand.Intn(len(sendSpecs))]
		id, err := api.PushMessageWithId(ctx, msg.ID, &msg.Message, meta)
		assert.NoError(t, err)
		assert.Equal(t, msg.ID, id)

		_, err = apiSign.PushMessageWithId(ctx, msg.ID, &msg.Message, meta)
		assert.Equal(t, jwtclient.ErrorPermissionDeny.Error(), err.Error())

		res, err := api.GetMessageByUid(ctx, msg.ID)
		assert.NoError(t, err)
		checkUnsignedMsg(t, &msg.Message, &res.Message)
		checkSendSpec(t, meta, res.Meta)
	}
}

func testHasMessageByUid(ctx context.Context, t *testing.T, api, apiSign messager.IMessager, addrs []address.Address) {
	msgs := genMessageWithAddress(addrs)
	for _, msg := range msgs {
		id, err := api.PushMessageWithId(ctx, msg.ID, &msg.Message, nil)
		assert.NoError(t, err)
		assert.Equal(t, msg.ID, id)

		has, err := api.HasMessageByUid(ctx, msg.ID)
		assert.NoError(t, err)
		assert.True(t, has)
	}

	has, err := api.HasMessageByUid(ctx, shared.NewUUID().String())
	assert.NoError(t, err)
	assert.False(t, has)
}

func testGetMessageByUid(ctx context.Context, t *testing.T, api, apiSign messager.IMessager, addrs []address.Address) {
	msgs := genMessageWithAddress(addrs)
	sendSpecs := testhelper.MockSendSpecs()

	for _, msg := range msgs {
		meta := sendSpecs[rand.Intn(len(sendSpecs))]
		id, err := api.PushMessage(ctx, &msg.Message, meta)
		assert.NoError(t, err)

		res, err := api.GetMessageByUid(ctx, id)
		assert.NoError(t, err)
		checkSendSpec(t, meta, res.Meta)
		checkUnsignedMsg(t, &msg.Message, &res.Message)

		_, err = apiSign.GetMessageByUid(ctx, id)
		assert.Equal(t, jwtclient.ErrorPermissionDeny.Error(), err.Error())
	}

	res, err := api.GetMessageByUid(ctx, shared.NewUUID().String())
	assert.Contains(t, err.Error(), gorm.ErrRecordNotFound.Error())
	assert.Nil(t, res)
}

func testWaitMessage(ctx context.Context, t *testing.T, api, apiSign messager.IMessager, addrs []address.Address) {
	msgs := genMessageWithAddress(addrs)
	sendSpecs := testhelper.MockSendSpecs()

	for _, msg := range msgs {
		meta := sendSpecs[rand.Intn(len(sendSpecs))]
		id, err := api.PushMessageWithId(ctx, msg.ID, &msg.Message, meta)
		assert.NoError(t, err)
		assert.Equal(t, msg.ID, id)
	}

	for _, msg := range msgs {
		_, err := apiSign.WaitMessage(ctx, msg.ID, constants.MessageConfidence)
		assert.Equal(t, jwtclient.ErrorPermissionDeny.Error(), err.Error())
	}

	for _, msg := range msgs {
		waitMessage(ctx, t, api, msg)
	}
}

func waitMessage(ctx context.Context, t *testing.T, api messager.IMessager, msg *types.Message) *types.Message {
	res, err := api.WaitMessage(ctx, msg.ID, constants.MessageConfidence)
	assert.NoError(t, err)
	assert.Equal(t, types.OnChainMsg, res.State)

	return res
}

func genMessagesAndWait(ctx context.Context, t *testing.T, api messager.IMessager, addrs []address.Address) []*types.Message {
	msgs := genMessageWithAddress(addrs)
	sendSpecs := testhelper.MockSendSpecs()

	for _, msg := range msgs {
		meta := sendSpecs[rand.Intn(len(sendSpecs))]
		id, err := api.PushMessageWithId(ctx, msg.ID, &msg.Message, meta)
		assert.NoError(t, err)
		assert.Equal(t, msg.ID, id)
	}

	newMsgs := make([]*types.Message, 0, len(msgs))
	for _, msg := range msgs {
		newMsgs = append(newMsgs, waitMessage(ctx, t, api, msg))
	}

	return newMsgs
}

func testGetMessageByUnsignedCID(ctx context.Context, t *testing.T, api, apiSign messager.IMessager, addrs []address.Address) {
	msgs := genMessagesAndWait(ctx, t, api, addrs)
	for _, msg := range msgs {
		res, err := api.GetMessageByUnsignedCid(ctx, *msg.UnsignedCid)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, res.Confidence, msg.Confidence)
		msg.Confidence = res.Confidence
		assert.Equal(t, msg, res)

		_, err = apiSign.GetMessageByUnsignedCid(ctx, *msg.UnsignedCid)
		assert.Equal(t, jwtclient.ErrorPermissionDeny.Error(), err.Error())
	}

	res, err := api.GetMessageByUnsignedCid(ctx, testutil.CidProvider(32)(t))
	assert.Contains(t, err.Error(), gorm.ErrRecordNotFound.Error())
	assert.Nil(t, res)
}

func testGetMessageBySignedCID(ctx context.Context, t *testing.T, api, apiSign messager.IMessager, addrs []address.Address) {
	msgs := genMessagesAndWait(ctx, t, api, addrs)
	for _, msg := range msgs {
		res, err := api.GetMessageBySignedCid(ctx, *msg.SignedCid)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, res.Confidence, msg.Confidence)
		msg.Confidence = res.Confidence
		assert.Equal(t, msg, res)

		_, err = apiSign.GetMessageBySignedCid(ctx, *msg.SignedCid)
		assert.Equal(t, jwtclient.ErrorPermissionDeny.Error(), err.Error())
	}

	res, err := api.GetMessageByUnsignedCid(ctx, testutil.CidProvider(32)(t))
	assert.Contains(t, err.Error(), gorm.ErrRecordNotFound.Error())
	assert.Nil(t, res)
}

func testGetMessageByFromAndNonce(ctx context.Context, t *testing.T, api, apiSign messager.IMessager, addrs []address.Address) {
	msgs := genMessagesAndWait(ctx, t, api, addrs)
	for _, msg := range msgs {
		res, err := api.GetMessageByFromAndNonce(ctx, msg.From, msg.Nonce)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, res.Confidence, msg.Confidence)
		msg.Confidence = res.Confidence
		assert.Equal(t, msg, res)

		_, err = apiSign.GetMessageByFromAndNonce(ctx, msg.From, msg.Nonce)
		assert.Equal(t, jwtclient.ErrorPermissionDeny.Error(), err.Error())
	}

	res, err := api.GetMessageByFromAndNonce(ctx, testutil.AddressProvider()(t), 1)
	assert.Contains(t, err.Error(), gorm.ErrRecordNotFound.Error())
	assert.Nil(t, res)
}

func testListMessage(ctx context.Context, t *testing.T, api, apiSign messager.IMessager, addrs []address.Address) {
	msgs := genMessagesAndWait(ctx, t, api, addrs)
	list, err := api.ListMessage(ctx, &types.MsgQueryParams{})
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(list), len(msgs))

	listNull, err := apiSign.ListMessage(ctx, &types.MsgQueryParams{})
	assert.NoError(t, err)
	assert.Equal(t, 0, len(listNull))

	msgMap := make(map[string]*types.Message, len(list))
	for _, msg := range list {
		msgMap[msg.ID] = msg
	}

	for _, msg := range msgs {
		tmpMsg, ok := msgMap[msg.ID]
		assert.True(t, ok)
		assert.LessOrEqual(t, msg.Confidence, tmpMsg.Confidence)
		tmpMsg.Confidence = msg.Confidence
		assert.Equal(t, tmpMsg, msg)
	}
}

func testListMessageByFromState(ctx context.Context, t *testing.T, api, apiSign messager.IMessager, addrs []address.Address) {
	// insert message
	genMessagesAndWait(ctx, t, api, addrs)
	genMessagesAndWait(ctx, t, api, addrs)

	state := types.OnChainMsg
	isAsc := true
	pageIndex := 1
	pageSize := 20

	checkCreatedAt := func(msgs []*types.Message, isAsc bool) {
		msgLen := len(msgs)
		for i := 0; i < msgLen-1; i++ {
			if isAsc {
				assert.True(t, msgs[i].CreatedAt.Before(msgs[i+1].CreatedAt))
			} else {
				assert.True(t, msgs[i].CreatedAt.After(msgs[i+1].CreatedAt))
			}
		}
	}

	tmpMsgs := make([]*types.Message, pageSize*2)
	msgs, err := api.ListMessageByFromState(ctx, address.Undef, state, isAsc, pageIndex, pageSize)
	assert.NoError(t, err)
	assert.Len(t, msgs, pageSize)
	checkCreatedAt(msgs, isAsc)
	copy(tmpMsgs, msgs)

	pageIndex = 2
	msgs, err = api.ListMessageByFromState(ctx, address.Undef, state, isAsc, pageIndex, pageSize)
	assert.NoError(t, err)
	assert.Len(t, msgs, pageSize)
	checkCreatedAt(msgs, isAsc)
	copy(tmpMsgs[20:], msgs)
	assert.Equal(t, tmpMsgs[20:], msgs)

	pageSize = 40
	pageIndex = 1
	msgs, err = api.ListMessageByFromState(ctx, address.Undef, state, isAsc, pageIndex, pageSize)
	assert.NoError(t, err)
	assert.Len(t, msgs, pageSize)
	checkCreatedAt(msgs, isAsc)
	for i, msg := range msgs {
		tmpMsg := tmpMsgs[i]
		assert.LessOrEqual(t, tmpMsg.Confidence, msg.Confidence)
		tmpMsg.Confidence = msg.Confidence
		assert.Equal(t, tmpMsg, msg)
	}

	isAsc = false
	msgs, err = api.ListMessageByFromState(ctx, address.Undef, state, isAsc, pageIndex, pageSize)
	assert.NoError(t, err)
	assert.Len(t, msgs, pageSize)
	checkCreatedAt(msgs, isAsc)

	allMsgs, err := api.ListMessage(ctx, &types.MsgQueryParams{})
	assert.NoError(t, err)
	msgIDs := make(map[address.Address][]string, len(allMsgs))
	for _, msg := range allMsgs {
		msgIDs[msg.From] = append(msgIDs[msg.From], msg.ID)
	}
	for addr, ids := range msgIDs {
		idsLen := len(ids)
		msgs, err = api.ListMessageByFromState(ctx, addr, state, isAsc, pageIndex, idsLen)
		assert.NoError(t, err)
		assert.Len(t, msgs, idsLen)
		checkCreatedAt(msgs, isAsc)
		for i, msg := range msgs {
			assert.Equal(t, ids[idsLen-1-i], msg.ID)
		}
	}
}

func testListMessageByAddress(ctx context.Context, t *testing.T, api, apiSign messager.IMessager) {
	allMsgs, err := api.ListMessage(ctx, &types.MsgQueryParams{})
	assert.NoError(t, err)
	msgIDs := make(map[address.Address][]string)
	for _, msg := range allMsgs {
		msgIDs[msg.From] = append(msgIDs[msg.From], msg.ID)
	}
	for addr, ids := range msgIDs {
		idsLen := len(ids)
		msgs, err := api.ListMessageByAddress(ctx, addr)
		assert.NoError(t, err)
		assert.Len(t, msgs, idsLen)
		for _, msg := range msgs {
			assert.Contains(t, ids, msg.ID)
		}
	}
}

func testListFailedMessage(ctx context.Context, t *testing.T, api, apiSign messager.IMessager, addrs []address.Address, blockDelay time.Duration) {
	msgs := genMessageWithAddress(addrs)
	for _, msg := range msgs {
		msg.Message.GasLimit = -1
		id, err := api.PushMessageWithId(ctx, msg.ID, &msg.Message, nil)
		assert.NoError(t, err)
		assert.Equal(t, msg.ID, id)

		res, err := api.GetMessageByUid(ctx, id)
		assert.NoError(t, err)
		checkUnsignedMsg(t, &msg.Message, &res.Message)
	}

	time.Sleep(blockDelay * 2)

	list, err := api.ListFailedMessage(ctx)
	assert.NoError(t, err)
	assert.Equal(t, len(msgs), len(list))
	for _, msg := range list {
		assert.Equal(t, types.UnFillMsg, msg.State)
		assert.True(t, strings.Contains(msg.ErrorMsg, testhelper.ErrGasLimitNegative.Error()))
	}

	// mark bad message
	for _, msg := range msgs {
		err := api.MarkBadMessage(ctx, msg.ID)
		assert.NoError(t, err)

		res, err := api.GetMessageByUid(ctx, msg.ID)
		assert.NoError(t, err)
		assert.Equal(t, types.FailedMsg, res.State)
	}

	listNull, err := apiSign.ListFailedMessage(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(listNull))
}

func testListBlockedMessage(ctx context.Context, t *testing.T, api, apiSign messager.IMessager, addrs []address.Address, blockDelay time.Duration) {
	msgs := genMessageWithAddress(addrs)
	addrMsg := make(map[address.Address][]*types.Message, len(addrs))
	for _, msg := range msgs {
		msg.GasPremium = big.Sub(testhelper.MinPackedPremium, big.NewInt(100))
		id, err := api.PushMessage(ctx, &msg.Message, nil)
		assert.NoError(t, err)

		res, err := api.GetMessageByUid(ctx, id)
		assert.NoError(t, err)
		checkUnsignedMsg(t, &msg.Message, &res.Message)

		addrMsg[res.From] = append(addrMsg[res.From], res)
	}

	time.Sleep(blockDelay * 2)

	for addr, msgs := range addrMsg {
		list, err := api.ListBlockedMessage(ctx, addr, blockDelay)
		assert.NoError(t, err)
		assert.Equal(t, len(msgs), len(list))

		_, err = apiSign.ListBlockedMessage(ctx, addr, blockDelay)
		assert.Equal(t, jwtclient.ErrorPermissionDeny.Error(), err.Error())

		for i, msg := range list {
			idx := len(msgs) - 1 - i
			assert.Equal(t, types.FillMsg, msg.State)
			assert.Equal(t, msgs[idx].GasPremium, msg.GasPremium)
			if i < len(list)-1 {
				assert.True(t, list[i].CreatedAt.Before(list[i+1].CreatedAt))
			}
		}
	}

	// replace message
	for _, msgs := range addrMsg {
		for _, msg := range msgs {
			params := &types.ReplacMessageParams{
				ID:             msg.ID,
				Auto:           false,
				GasLimit:       testhelper.DefGasUsed,
				GasPremium:     testhelper.DefGasPremium,
				GasFeecap:      testhelper.DefGasFeeCap,
				GasOverPremium: 0,
			}
			c, err := api.ReplaceMessage(ctx, params)
			assert.NoError(t, err)
			assert.True(t, c.Defined())
		}
	}

	for _, msgs := range addrMsg {
		for _, msg := range msgs {
			waitMessage(ctx, t, api, msg)
		}
	}
}

func testUpdateMessageStateByID(ctx context.Context, t *testing.T, api, apiSign messager.IMessager, addrs []address.Address, blockDelay time.Duration) {
	msgs := genMessageWithAddress(addrs)
	for _, msg := range msgs {
		msg.Message.GasLimit = -1
		id, err := api.PushMessageWithId(ctx, msg.ID, &msg.Message, nil)
		assert.NoError(t, err)
		assert.Equal(t, msg.ID, id)

		res, err := api.GetMessageByUid(ctx, id)
		assert.NoError(t, err)
		checkUnsignedMsg(t, &msg.Message, &res.Message)
	}

	time.Sleep(blockDelay * 2)

	list, err := api.ListFailedMessage(ctx)
	assert.NoError(t, err)
	assert.Equal(t, len(msgs), len(list))
	for _, msg := range list {
		assert.Equal(t, types.UnFillMsg, msg.State)
		assert.True(t, strings.Contains(msg.ErrorMsg, testhelper.ErrGasLimitNegative.Error()))
	}

	for _, msg := range msgs {
		assert.NoError(t, api.UpdateMessageStateByID(ctx, msg.ID, types.FailedMsg))
		err := apiSign.UpdateMessageStateByID(ctx, msg.ID, types.FailedMsg)
		assert.Equal(t, jwtclient.ErrorPermissionDeny.Error(), err.Error())

		res, err := api.GetMessageByUid(ctx, msg.ID)
		assert.NoError(t, err)
		assert.Equal(t, types.FailedMsg, res.State)
	}
}

func testUpdateAllFilledMessage(ctx context.Context, t *testing.T, api, apiSign messager.IMessager, addrs []address.Address, blockDelay time.Duration) {
	msgs := genMessageWithAddress(addrs)
	for _, msg := range msgs {
		id, err := api.PushMessageWithId(ctx, msg.ID, &msg.Message, nil)
		assert.NoError(t, err)
		assert.Equal(t, msg.ID, id)

		res, err := api.GetMessageByUid(ctx, id)
		assert.NoError(t, err)
		checkUnsignedMsg(t, &msg.Message, &res.Message)
	}
	ctx, cancel := context.WithTimeout(ctx, blockDelay*2)
	defer cancel()
	ticker := time.NewTicker(blockDelay / 10)
	defer ticker.Stop()

	updateTotal := 0
	for {
		select {
		case <-ticker.C:
			if updateTotal == len(msgs) {
				return
			}
			count, err := api.UpdateAllFilledMessage(ctx)
			assert.NoError(t, err)
			updateTotal += count
		case <-ctx.Done():
			assert.NoError(t, ctx.Err())
			return
		}
	}
}

func testUpdateFilledMessageByID(ctx context.Context, t *testing.T, api, apiSign messager.IMessager, addrs []address.Address, blockDelay time.Duration) {
	msgs := genMessageWithAddress(addrs)
	for _, msg := range msgs {
		id, err := api.PushMessageWithId(ctx, msg.ID, &msg.Message, nil)
		assert.NoError(t, err)
		assert.Equal(t, msg.ID, id)

		res, err := api.GetMessageByUid(ctx, id)
		assert.NoError(t, err)
		checkUnsignedMsg(t, &msg.Message, &res.Message)
	}
	ctx, cancel := context.WithTimeout(ctx, blockDelay*4)
	defer cancel()
	wg := sync.WaitGroup{}

	update := func(msg *types.Message) {
		ticker := time.NewTicker(blockDelay / 10)
		defer ticker.Stop()
		defer wg.Done()

		for {
			select {
			case <-ticker.C:
				_, err := api.UpdateFilledMessageByID(ctx, msg.ID)
				if err == nil || strings.Contains(err.Error(), "has been final") {
					res, err := api.GetMessageByUid(ctx, msg.ID)
					assert.NoError(t, err)
					if res.SignedCid != nil {
						assert.Equal(t, types.OnChainMsg, res.State)
						assert.True(t, res.Height > 0)
						assert.False(t, res.TipSetKey.IsEmpty())
						return
					}
				} else {
					assert.True(t, strings.Contains(err.Error(), "not found "))
				}
			case <-ctx.Done():
				assert.NoError(t, ctx.Err())
				return
			}
		}
	}

	for i := range msgs {
		_, err := apiSign.UpdateFilledMessageByID(ctx, msgs[i].ID)
		assert.Equal(t, jwtclient.ErrorPermissionDeny.Error(), err.Error())

		wg.Add(1)
		go update(msgs[i])
	}
	wg.Wait()
}

func testReplaceMessage(ctx context.Context, t *testing.T, api, apiSign messager.IMessager, addrs []address.Address, blockDelay time.Duration) {
	msgs := genMessageWithAddress(addrs)
	for _, msg := range msgs {
		msg.GasPremium = big.Sub(testhelper.MinPackedPremium, big.NewInt(100))
		id, err := api.PushMessageWithId(ctx, msg.ID, &msg.Message, nil)
		assert.NoError(t, err)
		assert.Equal(t, msg.ID, id)

		res, err := api.GetMessageByUid(ctx, id)
		assert.NoError(t, err)
		checkUnsignedMsg(t, &msg.Message, &res.Message)
	}

	time.Sleep(blockDelay * 2)

	params := testhelper.MockReplaceMessageParams()
	paramsLen := len(params)

	// replace message
	for i, msg := range msgs {
		param := params[i%paramsLen]
		param.ID = msg.ID
		c, err := api.ReplaceMessage(ctx, param)
		assert.NoError(t, err)
		assert.True(t, c.Defined())

		_, err = apiSign.ReplaceMessage(ctx, param)
		assert.Equal(t, jwtclient.ErrorPermissionDeny.Error(), err.Error())
	}

	check := func(idx int, msg *types.Message) {
		assert.Equal(t, types.OnChainMsg, msg.State)
		gasLimit := testhelper.DefGasUsed
		gasFeeCap := testhelper.DefGasFeeCap
		gasPremium := testhelper.DefGasPremium
		gasOverEstimation := service.DefSharedParams.GasOverEstimation
		if msg.Meta != nil && msg.Meta.GasOverEstimation != 0 {
			gasOverEstimation = msg.Meta.GasOverEstimation
		}
		gasLimit = int64(float64(gasLimit) * gasOverEstimation)
		switch idx {
		case 0:
			gasFeeCap = big.Add(testhelper.DefGasFeeCap, testhelper.DefGasPremium)
		case 1:
			gasPremium = big.Mul(big.NewInt(int64(params[1].GasOverPremium*10000/10000)), testhelper.DefGasPremium)
			gasFeeCap = big.Add(testhelper.DefGasFeeCap, gasPremium)
		case 2:
			gasFeeCap = big.Div(params[2].MaxFee, big.NewInt(gasLimit))
			gasPremium = big.Min(gasFeeCap, gasPremium)
		case 3:
			gasLimit = params[3].GasLimit
			gasFeeCap = params[3].GasFeecap
			gasPremium = params[3].GasPremium
		default:
			t.Errorf("idx %d > %d", idx, paramsLen)
		}
		assert.Equal(t, gasLimit, msg.GasLimit)
		assert.Equal(t, gasFeeCap, msg.GasFeeCap)
		assert.Equal(t, gasPremium, msg.GasPremium)
	}

	for i, msg := range msgs {
		res, err := api.WaitMessage(ctx, msg.ID, constants.MessageConfidence)
		assert.NoError(t, err)
		check(i%paramsLen, res)
	}
}

func testMarkBadMessage(ctx context.Context, t *testing.T, api, apiSign messager.IMessager, addrs []address.Address, blockDelay time.Duration) {
	msgs := genMessageWithAddress(addrs)
	for _, msg := range msgs {
		msg.Message.GasLimit = -1
		id, err := api.PushMessageWithId(ctx, msg.ID, &msg.Message, nil)
		assert.NoError(t, err)
		assert.Equal(t, msg.ID, id)

		res, err := api.GetMessageByUid(ctx, id)
		assert.NoError(t, err)
		checkUnsignedMsg(t, &msg.Message, &res.Message)
	}

	time.Sleep(blockDelay * 2)

	for _, msg := range msgs {
		res, err := api.GetMessageByUid(ctx, msg.ID)
		assert.NoError(t, err)
		assert.Equal(t, types.UnFillMsg, res.State)
		assert.True(t, strings.Contains(res.ErrorMsg, testhelper.ErrGasLimitNegative.Error()))

		assert.NoError(t, api.MarkBadMessage(ctx, msg.ID))
		res, err = api.GetMessageByUid(ctx, msg.ID)
		assert.NoError(t, err)
		assert.Equal(t, types.FailedMsg, res.State)

		err = apiSign.MarkBadMessage(ctx, msg.ID)
		assert.Equal(t, jwtclient.ErrorPermissionDeny.Error(), err.Error())

	}

	assert.Error(t, api.MarkBadMessage(ctx, shared.NewUUID().String()))
}

func testRecoverFailedMsg(ctx context.Context, t *testing.T, api, apiSign messager.IMessager, addrs []address.Address, blockDelay time.Duration) {
	msgs := genMessageWithAddress(addrs)
	addrIDs := make(map[address.Address]map[string]struct{})
	for _, msg := range msgs {
		msg.GasPremium = big.Sub(testhelper.MinPackedPremium, big.NewInt(100))
		id, err := api.PushMessageWithId(ctx, msg.ID, &msg.Message, nil)
		assert.NoError(t, err)
		assert.Equal(t, msg.ID, id)

		res, err := api.GetMessageByUid(ctx, id)
		assert.NoError(t, err)
		checkUnsignedMsg(t, &msg.Message, &res.Message)

		ids, ok := addrIDs[res.From]
		if ok {
			ids[res.ID] = struct{}{}
		} else {
			addrIDs[res.From] = map[string]struct{}{res.ID: {}}
		}
	}

	time.Sleep(blockDelay * 2)

	for _, msg := range msgs {
		assert.NoError(t, api.MarkBadMessage(ctx, msg.ID))
	}

	for addr, ids := range addrIDs {
		recoverIDs, err := api.RecoverFailedMsg(ctx, addr)
		assert.NoError(t, err)
		assert.Equal(t, len(ids), len(recoverIDs))
		for _, id := range recoverIDs {
			_, ok := ids[id]
			assert.True(t, ok)
		}

		_, err = apiSign.RecoverFailedMsg(ctx, addr)
		assert.Equal(t, jwtclient.ErrorPermissionDeny.Error(), err.Error())
	}
}

func genMessageWithAddress(addrs []address.Address) []*types.Message {
	msgs := testhelper.NewMessages(len(addrs) * 2)
	for _, msg := range msgs {
		msg.From = addrs[rand.Intn(len(addrs))]
		msg.To = addrs[rand.Intn(len(addrs))]
	}

	return msgs
}

//// check ////

func checkUnsignedMsg(t *testing.T, expect, actual *shared.Message) {
	assert.Equal(t, expect.Version, actual.Version)
	assert.Equal(t, expect.To, actual.To)
	assert.Equal(t, expect.Value, actual.Value)
	assert.Equal(t, expect.Method, actual.Method)
	assert.Equal(t, expect.Params, actual.Params)
	assert.Equal(t, testhelper.ResolveAddr(t, expect.From), actual.From)
	// todo: finish estimate gas
	if actual.Nonce > 0 {
	} else {
		assert.Equal(t, expect.GasLimit, actual.GasLimit)
		assert.Equal(t, expect.GasFeeCap, actual.GasFeeCap)
		assert.Equal(t, expect.GasPremium, actual.GasPremium)
	}
}

func checkSendSpec(t *testing.T, expect, actual *types.SendSpec) {
	if expect == nil {
		assert.Equal(t, big.NewInt(0), actual.MaxFee)
		assert.Equal(t, float64(0), actual.GasOverPremium)
		assert.Equal(t, float64(0), actual.GasOverEstimation)
		return
	}
	if expect.MaxFee.NilOrZero() {
		assert.Equal(t, big.NewInt(0), actual.MaxFee)
	}
	assert.Equal(t, expect.GasOverPremium, actual.GasOverPremium)
	assert.Equal(t, expect.GasOverEstimation, actual.GasOverEstimation)
}

type testParams struct {
	ctx        context.Context
	cfg        *config.Config
	authClient *testhelper.AuthClient
	ms         *messagerServer
	addrCount  int
	account    string
	addrs      []address.Address
	apiAdmin   messager.IMessager
	apiSign    messager.IMessager
	blockDelay time.Duration
	closer     func()
}

func prepare(t *testing.T) *testParams {
	ctx := context.Background()
	// ctx = auth.WithPerm(ctx, []string{"admin", "sign", "write", "read"})
	cfg := config.DefaultConfig()
	cfg.API.Address = "/ip4/0.0.0.0/tcp/0"
	cfg.MessageService.WaitingChainHeadStableDuration = 1 * time.Second
	blockDelay := cfg.MessageService.WaitingChainHeadStableDuration * 2
	authClient := testhelper.NewMockAuthClient(t)
	ms, err := mockMessagerServer(ctx, t.TempDir(), cfg, authClient)
	assert.NoError(t, err)

	go ms.start(ctx)
	assert.NoError(t, <-ms.appStartErr)

	// account with local token of admin perm and 10 signers
	addrCount := 10
	account := defaultLocalToken
	addrs := testhelper.RandAddresses(t, addrCount)
	authClient.Init(account, addrs)
	assert.NoError(t, ms.walletCli.AddAddress(account, addrs))
	assert.NoError(t, ms.fullNode.AddActors(addrs))

	apiAdmin, closer, err := newMessagerClient(ctx, ms.port, ms.token)
	assert.NoError(t, err)

	// account with token of sign permission and no signers
	accountSign := "accountSign"
	playLoad := &auth.JWTPayload{
		Name: accountSign,
		Perm: core.PermSign,
	}
	tokenSign, err := genToken(playLoad)
	assert.NoError(t, err)
	authClient.EXPECT().Verify(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, token string) (*auth.VerifyResponse, error) {
		if token == tokenSign {
			return playLoad, nil
		}
		return nil, errors.New("not exist")
	}).AnyTimes()
	apiSign, closer2, err := newMessagerClient(ctx, ms.port, tokenSign)
	assert.NoError(t, err)

	return &testParams{
		ctx:        ctx,
		cfg:        cfg,
		authClient: authClient,
		ms:         ms,
		addrCount:  addrCount,
		account:    account,
		addrs:      addrs,
		apiAdmin:   apiAdmin,
		apiSign:    apiSign,

		blockDelay: blockDelay,
		closer:     func() { closer(); closer2() },
	}
}

func genToken(pl interface{}) (string, error) {
	secret, err := jwtclient.RandSecret()
	if err != nil {
		return "", xerrors.Errorf("rand secret %v", err)
	}
	tk, err := jwt.Sign(pl, jwt.NewHS256(secret))
	if err != nil {
		return core.EmptyString, xerrors.Errorf("gen token failed :%s", err)
	}
	return string(tk), nil
}
