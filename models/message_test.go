package models

import (
	"math/rand"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/filecoin-project/go-address"

	"github.com/filecoin-project/go-state-types/abi"
	venustypes "github.com/filecoin-project/venus/pkg/types"
	"github.com/stretchr/testify/assert"

	"github.com/ipfs-force-community/venus-messager/models/repo"
	"github.com/ipfs-force-community/venus-messager/types"
	"github.com/ipfs-force-community/venus-messager/utils"
)

func TestSaveAndGetMessage(t *testing.T) {
	sqliteRepo, mysqlRepo := setupRepo(t)

	messageRepoTest := func(t *testing.T, messageRepo repo.MessageRepo) {
		msg := NewMessage()

		err := messageRepo.CreateMessage(msg)
		assert.NoError(t, err)

		result, err := messageRepo.GetMessageByUid(msg.ID)
		assert.NoError(t, err)

		beforeSave := ObjectToString(msg)
		afterSave := ObjectToString(result)
		assert.Equal(t, beforeSave, afterSave)

		allMsg, err := messageRepo.ListMessage()
		assert.NoError(t, err)
		assert.LessOrEqual(t, 1, len(allMsg))

		unchainedMsgs, err := messageRepo.ListUnchainedMsgs()
		assert.NoError(t, err)
		assert.LessOrEqual(t, 1, len(unchainedMsgs))

		signedMsg := NewSignedMessages(1)[0]
		err = messageRepo.CreateMessage(signedMsg)
		assert.NoError(t, err)
		msg2, err := messageRepo.GetMessageBySignedCid(*signedMsg.SignedCid)
		assert.NoError(t, err)
		assert.Equal(t, signedMsg.SignedCid, msg2.SignedCid)
	}

	t.Run("SaveAndGetMessage", func(t *testing.T) {
		t.Run("sqlite", func(t *testing.T) {
			messageRepoTest(t, sqliteRepo.MessageRepo())
		})
		t.Run("mysql", func(t *testing.T) {
			t.SkipNow()
			messageRepoTest(t, mysqlRepo.MessageRepo())
		})
	})
}

func TestUpdateMessageInfoByCid(t *testing.T) {

	sqliteRepo, mysqlRepo := setupRepo(t)

	messageRepoTest := func(t *testing.T, messageRepo repo.MessageRepo) {
		msg := NewSignedMessages(1)[0]
		unsignedCid := msg.UnsignedCid

		err := messageRepo.CreateMessage(msg)
		assert.NoError(t, err)

		rec := &venustypes.MessageReceipt{
			ExitCode:    0,
			ReturnValue: []byte{'g', 'd'},
			GasUsed:     34,
		}
		tsKeyStr := "{ bafy2bzacec7ymsvmwjgew5whbhs4c3gg5k76pu6y7tun67lqw6unt6xo2nn62 bafy2bzacediq3wdlglhbc6ezlmnks46hdl2kyc3alghiov3c6jpt5qcf76s32 bafy2bzacebjjsg2vqadraxippg46rkysbyucgl27qzu6p6bgepcn7ybgjmqxs }"
		tsKey, err := utils.StringToTipsetKey(tsKeyStr)
		assert.NoError(t, err)

		height := abi.ChainEpoch(10)
		state := types.OnChainMsg
		err = messageRepo.UpdateMessageInfoByCid(unsignedCid.String(), rec, height, state, tsKey)
		assert.NoError(t, err)

		msg2, err := messageRepo.GetMessageByCid(*unsignedCid)
		assert.NoError(t, err)
		assert.Equal(t, int64(height), msg2.Height)
		assert.Equal(t, rec, msg2.Receipt)
		assert.Equal(t, state, msg2.State)
		assert.Equal(t, tsKeyStr, msg2.TipSetKey.String())
	}
	t.Run("UpdateMessageInfoByCid", func(t *testing.T) {
		t.Run("sqlite", func(t *testing.T) {
			messageRepoTest(t, sqliteRepo.MessageRepo())
		})
		t.Run("mysql", func(t *testing.T) {
			t.SkipNow()
			messageRepoTest(t, mysqlRepo.MessageRepo())
		})
	})
}

func TestUpdateMessageStateByCid(t *testing.T) {
	sqliteRepo, mysqlRepo := setupRepo(t)

	messageRepoTest := func(t *testing.T, messageRepo repo.MessageRepo) {
		msg := NewSignedMessages(1)[0]
		msg.State = types.FillMsg
		cid := msg.UnsignedMessage.Cid()
		msg.UnsignedCid = &cid

		err := messageRepo.CreateMessage(msg)
		assert.NoError(t, err)

		err = messageRepo.UpdateMessageStateByCid(cid.String(), types.OnChainMsg)
		assert.NoError(t, err)

		msg2, err := messageRepo.GetMessageByUid(msg.ID)
		assert.NoError(t, err)
		assert.Equal(t, types.OnChainMsg, msg2.State)
	}
	t.Run("UpdateMessageStateByCid", func(t *testing.T) {
		t.Run("sqlite", func(t *testing.T) {
			messageRepoTest(t, sqliteRepo.MessageRepo())
		})
		t.Run("mysql", func(t *testing.T) {
			t.SkipNow()
			messageRepoTest(t, mysqlRepo.MessageRepo())
		})
	})
}

func TestExpireMessage(t *testing.T) {
	sqliteRepo, mysqlRepo := setupRepo(t)

	messageRepoTest := func(t *testing.T, messageRepo repo.MessageRepo) {
		msg := NewSignedMessages(1)[0]

		err := messageRepo.CreateMessage(msg)
		assert.NoError(t, err)

		err = messageRepo.ExpireMessage([]*types.Message{msg})
		assert.NoError(t, err)

		msg2, err := messageRepo.GetMessageByUid(msg.ID)
		assert.NoError(t, err)
		assert.Equal(t, types.FailedMsg, msg2.State)
	}
	t.Run("ExpireMessage", func(t *testing.T) {
		t.Run("sqlite", func(t *testing.T) {
			messageRepoTest(t, sqliteRepo.MessageRepo())
		})
		t.Run("mysql", func(t *testing.T) {
			t.SkipNow()
			messageRepoTest(t, mysqlRepo.MessageRepo())
		})
	})
}

func TestGetMessageState(t *testing.T) {
	sqliteRepo, mysqlRepo := setupRepo(t)

	messageRepoTest := func(t *testing.T, messageRepo repo.MessageRepo) {
		msg := NewMessage()
		err := messageRepo.CreateMessage(msg)
		assert.NoError(t, err)
		state, err := messageRepo.GetMessageState(msg.ID)
		assert.NoError(t, err)
		assert.Equal(t, state, types.UnKnown)

		for _, state := range []types.MessageState{types.UnFillMsg, types.FillMsg, types.OnChainMsg, types.FailedMsg} {
			msg.State = state
			err = messageRepo.SaveMessage(msg)
			assert.NoError(t, err)
			state, err = messageRepo.GetMessageState(msg.ID)
			assert.NoError(t, err)
			assert.Equal(t, state, state)
		}
	}
	t.Run("GetMessageState", func(t *testing.T) {
		t.Run("sqlite", func(t *testing.T) {
			messageRepoTest(t, sqliteRepo.MessageRepo())
		})
		t.Run("mysql", func(t *testing.T) {
			t.SkipNow()
			messageRepoTest(t, mysqlRepo.MessageRepo())
		})
	})
}

func TestGetSignedMessageByTime(t *testing.T) {
	sqliteRepo, mysqlRepo := setupRepo(t)

	messageRepoTest := func(t *testing.T, messageRepo repo.MessageRepo) {
		msg := NewMessage()
		err := messageRepo.CreateMessage(msg)
		assert.NoError(t, err)

		signedMsgs := NewSignedMessages(10)
		for _, msg := range signedMsgs {
			err := messageRepo.CreateMessage(msg)
			assert.NoError(t, err)
		}
		startTime := time.Now().Add(-time.Second * 3600)
		msgs, err := messageRepo.GetSignedMessageByTime(startTime)
		assert.NoError(t, err)
		assert.LessOrEqual(t, 10, len(msgs))
	}
	t.Run("GetSignedMessageByTime", func(t *testing.T) {
		t.Run("sqlite", func(t *testing.T) {
			messageRepoTest(t, sqliteRepo.MessageRepo())
		})
		t.Run("mysql", func(t *testing.T) {
			t.SkipNow()
			messageRepoTest(t, mysqlRepo.MessageRepo())
		})
	})
}

func TestGetSignedMessageByHeight(t *testing.T) {
	sqliteRepo, mysqlRepo := setupRepo(t)

	messageRepoTest := func(t *testing.T, messageRepo repo.MessageRepo) {
		msg := NewMessage()
		err := messageRepo.CreateMessage(msg)
		assert.NoError(t, err)

		signedMsgs := NewSignedMessages(10)
		for i, msg := range signedMsgs {
			msg.Height = int64(i)
			err := messageRepo.CreateMessage(msg)
			assert.NoError(t, err)
		}
		height := abi.ChainEpoch(5)
		msgs, err := messageRepo.GetSignedMessageByHeight(height)
		assert.NoError(t, err)
		assert.LessOrEqual(t, 5, len(msgs))
	}
	t.Run("GetSignedMessageByHeight", func(t *testing.T) {
		t.Run("sqlite", func(t *testing.T) {
			messageRepoTest(t, sqliteRepo.MessageRepo())
		})
		t.Run("mysql", func(t *testing.T) {
			t.SkipNow()
			messageRepoTest(t, mysqlRepo.MessageRepo())
		})
	})
}

func TestGetMessageByFromAndNonce(t *testing.T) {
	sqliteRepo, mysqlRepo := setupRepo(t)

	messageRepoTest := func(t *testing.T, messageRepo repo.MessageRepo) {
		msg := NewSignedMessages(1)[0]
		err := messageRepo.CreateMessage(msg)
		assert.NoError(t, err)

		result, err := messageRepo.GetMessageByFromAndNonce(msg.From, msg.Nonce)
		assert.NoError(t, err)

		assert.Equal(t, ObjectToString(msg), ObjectToString(result))
	}
	t.Run("GetMessageByFromAndNonce", func(t *testing.T) {
		t.Run("sqlite", func(t *testing.T) {
			messageRepoTest(t, sqliteRepo.MessageRepo())
		})
		t.Run("mysql", func(t *testing.T) {
			t.SkipNow()
			messageRepoTest(t, mysqlRepo.MessageRepo())
		})
	})
}

func TestListFilledMessageByHeight(t *testing.T) {
	sqliteRepo, mysqlRepo := setupRepo(t)

	messageRepoTest := func(t *testing.T, messageRepo repo.MessageRepo) {
		randHeight := rand.Uint64() / 2
		for _, msg := range NewSignedMessages(10) {
			msg.Height = int64(randHeight)
			msg.State = types.FillMsg
			err := messageRepo.CreateMessage(msg)
			assert.NoError(t, err)
		}

		result, err := messageRepo.ListFilledMessageByHeight(abi.ChainEpoch(randHeight))
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 10)
	}
	t.Run("ListFilledMessageByHeight", func(t *testing.T) {
		t.Run("sqlite", func(t *testing.T) {
			messageRepoTest(t, sqliteRepo.MessageRepo())
		})
		t.Run("mysql", func(t *testing.T) {
			t.SkipNow()
			messageRepoTest(t, mysqlRepo.MessageRepo())
		})
	})
}

func TestListFilledMessageByAddress(t *testing.T) {
	sqliteRepo, mysqlRepo := setupRepo(t)

	messageRepoTest := func(t *testing.T, messageRepo repo.MessageRepo) {
		uid, err := uuid.NewUUID()
		assert.NoError(t, err)
		addr, err := address.NewActorAddress(uid[:])
		assert.NoError(t, err)

		msgs, err := messageRepo.ListFilledMessageByAddress(addr)
		assert.NoError(t, err)
		assert.Len(t, msgs, 0)

		count := 10
		signedMsgs := NewSignedMessages(count)
		for i, msg := range signedMsgs {
			if i%2 == 0 {
				msg.State = types.FillMsg
			}
			msg.From = addr
			err := messageRepo.CreateMessage(msg)
			assert.NoError(t, err)
		}

		msgs, err = messageRepo.ListFilledMessageByAddress(signedMsgs[0].From)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(msgs), count/2)
	}
	t.Run("ListFilledMessageByAddress", func(t *testing.T) {
		t.Run("sqlite", func(t *testing.T) {
			messageRepoTest(t, sqliteRepo.MessageRepo())
		})
		t.Run("mysql", func(t *testing.T) {
			t.SkipNow()
			messageRepoTest(t, mysqlRepo.MessageRepo())
		})
	})
}

func TestUpdateUnFilledMessageState(t *testing.T) {
	sqliteRepo, mysqlRepo := setupRepo(t)

	messageRepoTest := func(t *testing.T, messageRepo repo.MessageRepo) {

		msgs := NewMessages(10)
		for _, msg := range msgs {
			msg.State = types.UnFillMsg
			err := messageRepo.CreateMessage(msg)
			assert.NoError(t, err)
		}

		assert.NoError(t, messageRepo.UpdateUnFilledMessageState(msgs[0].WalletName, msgs[0].From, types.NoWalletMsg))

		msg, err := messageRepo.GetMessageByUid(msgs[0].ID)
		assert.NoError(t, err)
		assert.Equal(t, types.NoWalletMsg, msg.State)
	}
	t.Run("UpdateUnFilledMessageState", func(t *testing.T) {
		t.Run("sqlite", func(t *testing.T) {
			messageRepoTest(t, sqliteRepo.MessageRepo())
		})
		t.Run("mysql", func(t *testing.T) {
			t.SkipNow()
			messageRepoTest(t, mysqlRepo.MessageRepo())
		})
	})
}

func TestMarkBadMessage(t *testing.T) {
	sqliteRepo, mysqlRepo := setupRepo(t)

	messageRepoTest := func(t *testing.T, messageRepo repo.MessageRepo) {

		msgs := NewMessages(1)
		for _, msg := range msgs {
			msg.State = types.UnFillMsg
			err := messageRepo.CreateMessage(msg)
			assert.NoError(t, err)
		}

		_, err := messageRepo.MarkBadMessage(msgs[0].ID)
		assert.NoError(t, err)

		msg, err := messageRepo.GetMessageByUid(msgs[0].ID)
		assert.NoError(t, err)
		assert.Equal(t, types.FailedMsg, msg.State)
	}
	t.Run("UpdateUnFilledMessageState", func(t *testing.T) {
		t.Run("sqlite", func(t *testing.T) {
			messageRepoTest(t, sqliteRepo.MessageRepo())
		})
		t.Run("mysql", func(t *testing.T) {
			t.SkipNow()
			messageRepoTest(t, mysqlRepo.MessageRepo())
		})
	})
}

func TestUpdateReturnValue(t *testing.T) {
	sqliteRepo, mysqlRepo := setupRepo(t)

	messageRepoTest := func(t *testing.T, messageRepo repo.MessageRepo) {

		msgs := NewMessages(2)
		for _, msg := range msgs {
			msg.State = types.UnFillMsg
			err := messageRepo.CreateMessage(msg)
			assert.NoError(t, err)
		}
		failedInfo := "gas estimate failed"
		err := messageRepo.UpdateReturnValue(msgs[0].ID, failedInfo)
		assert.NoError(t, err)
		msg, err := messageRepo.GetMessageByUid(msgs[0].ID)
		assert.NoError(t, err)
		assert.Equal(t, failedInfo, string(msg.Receipt.ReturnValue))

		failedMsgs, err := messageRepo.ListFailedMessage()
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(failedMsgs), 1)
	}
	t.Run("UpdateUnFilledMessageState", func(t *testing.T) {
		t.Run("sqlite", func(t *testing.T) {
			messageRepoTest(t, sqliteRepo.MessageRepo())
		})
		t.Run("mysql", func(t *testing.T) {
			t.SkipNow()
			messageRepoTest(t, mysqlRepo.MessageRepo())
		})
	})
}
