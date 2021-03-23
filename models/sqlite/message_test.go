package sqlite

import (
	"os"
	"testing"
	"time"

	"github.com/filecoin-project/go-state-types/abi"
	venustypes "github.com/filecoin-project/venus/pkg/types"
	"github.com/stretchr/testify/assert"

	"github.com/ipfs-force-community/venus-messager/config"
	"github.com/ipfs-force-community/venus-messager/models/repo"
	"github.com/ipfs-force-community/venus-messager/types"
	"github.com/ipfs-force-community/venus-messager/utils"
)

func setup(path string) repo.Repo {
	db, err := OpenSqlite(&config.SqliteConfig{Path: path})
	if err != nil {
		panic(err)
	}
	if err = db.AutoMigrate(); err != nil {
		panic(err)
	}

	return db
}

func TestSageAndGetMessage(t *testing.T) {
	name := "TestSageAndGetMessage.db"
	db := setup(name)
	defer func() {
		assert.NoError(t, os.Remove(name))
	}()

	msgDb := db.MessageRepo()
	msg := NewMessage()

	id, err := msgDb.SaveMessage(msg)
	assert.NoError(t, err)

	result, err := msgDb.GetMessageByUid(id)
	assert.NoError(t, err)

	beforeSave := ObjectToString(msg)
	afterSave := ObjectToString(result)
	assert.Equal(t, beforeSave, afterSave)

	allMsg, err := msgDb.ListMessage()
	assert.NoError(t, err)
	assert.LessOrEqual(t, 1, len(allMsg))

	unchainedMsgs, err := msgDb.ListUnchainedMsgs()
	assert.NoError(t, err)
	assert.LessOrEqual(t, 1, len(unchainedMsgs))

	signedMsg := NewSignedMessages(1)[0]
	_, err = msgDb.SaveMessage(signedMsg)
	assert.NoError(t, err)
	msg2, err := msgDb.GetMessageBySignedCid(*signedMsg.SignedCid)
	assert.NoError(t, err)
	assert.Equal(t, signedMsg.SignedCid, msg2.SignedCid)
}

func TestUpdateMessageInfoByCid(t *testing.T) {
	name := "TestUpdateMessageInfoByCid.db"
	db := setup(name)
	defer func() {
		assert.NoError(t, os.Remove(name))
	}()

	msg := NewSignedMessages(1)[0]
	unsignedCid := msg.UnsignedCid

	_, err := db.MessageRepo().SaveMessage(msg)
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
	_, err = db.MessageRepo().UpdateMessageInfoByCid(unsignedCid.String(), rec, height, state, tsKey)
	assert.NoError(t, err)

	msg2, err := db.MessageRepo().GetMessageByCid(*unsignedCid)
	assert.NoError(t, err)
	assert.Equal(t, int64(height), msg2.Height)
	assert.Equal(t, rec, msg2.Receipt)
	assert.Equal(t, state, msg2.State)
	assert.Equal(t, tsKeyStr, msg2.TipSetKey.String())
}

func TestUpdateMessageStateByCid(t *testing.T) {
	name := "TestSageAndGetMessage.db"
	db := setup(name)
	defer func() {
		assert.NoError(t, os.Remove(name))
	}()

	msg := NewSignedMessages(1)[0]
	msg.State = types.FillMsg
	cid := msg.UnsignedMessage.Cid()
	msg.UnsignedCid = &cid

	_, err := db.MessageRepo().SaveMessage(msg)
	assert.NoError(t, err)

	_, err = db.MessageRepo().UpdateMessageStateByCid(cid.String(), types.OnChainMsg)
	assert.NoError(t, err)

	msg2, err := db.MessageRepo().GetMessageByUid(msg.ID)
	assert.NoError(t, err)
	assert.Equal(t, types.OnChainMsg, msg2.State)
}

func Test_sqliteMessageRepo_GetMessageState(t *testing.T) {

}

func TestSqliteMessageRepo_GetSignedMessageByTime(t *testing.T) {
	name := "GetSignedMessageByTim.db"
	db := setup(name)
	defer func() {
		assert.NoError(t, os.Remove(name))
	}()

	msgDb := db.MessageRepo()
	msg := NewMessage()
	_, err := msgDb.SaveMessage(msg)
	assert.NoError(t, err)

	signedMsgs := NewSignedMessages(10)
	for _, msg := range signedMsgs {
		_, err := msgDb.SaveMessage(msg)
		assert.NoError(t, err)
	}
	startTime := time.Now().Add(-time.Second * 3600)
	msgs, err := msgDb.GetSignedMessageByTime(startTime)
	assert.NoError(t, err)
	assert.Equal(t, 10, len(msgs))
}

func TestSqliteMessageRepo_GetSignedMessageByHeight(t *testing.T) {
	name := "GetSignedMessageByHeight.db"
	db := setup(name)
	defer func() {
		assert.NoError(t, os.Remove(name))
	}()

	msgDb := db.MessageRepo()
	msg := NewMessage()
	_, err := msgDb.SaveMessage(msg)
	assert.NoError(t, err)

	signedMsgs := NewSignedMessages(10)
	for i, msg := range signedMsgs {
		msg.Height = int64(i)
		_, err := msgDb.SaveMessage(msg)
		assert.NoError(t, err)
	}
	height := abi.ChainEpoch(5)
	msgs, err := msgDb.GetSignedMessageByHeight(height)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(msgs))
}

func TestSqliteMessageRepo_GetMessageByFromAndNonce(t *testing.T) {
	name := "GetMessageByFromAndNonce.db"
	db := setup(name)
	defer func() {
		assert.NoError(t, os.Remove(name))
	}()

	msgDb := db.MessageRepo()
	msg := NewSignedMessages(1)[0]
	_, err := msgDb.SaveMessage(msg)
	assert.NoError(t, err)

	result, err := msgDb.GetMessageByFromAndNonce(msg.From, msg.Nonce)
	assert.NoError(t, err)

	assert.Equal(t, ObjectToString(msg), ObjectToString(result))
}

func TestSqliteMessageRepo_ListFilledMessageByHeight(t *testing.T) {
	name := "ListFilledMessageByHeight.db"
	db := setup(name)
	defer func() {
		assert.NoError(t, os.Remove(name))
	}()
	msgDb := db.MessageRepo()
	for _, msg := range NewSignedMessages(10) {
		msg.Height = 10
		msg.State = types.FillMsg
		_, err := msgDb.SaveMessage(msg)
		assert.NoError(t, err)
	}

	result, err := msgDb.ListFilledMessageByHeight(10)
	assert.NoError(t, err)
	assert.Len(t, result, 10)
}
