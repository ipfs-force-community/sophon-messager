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
	db := setup("message.db")

	msgDb := db.MessageRepo()
	msg := NewMessage()
	beforeSave := ObjectToString(msg)

	uuid, err := msgDb.SaveMessage(msg)
	assert.NoError(t, err)

	result, err := msgDb.GetMessage(uuid)
	assert.NoError(t, err)

	afterSave := ObjectToString(result)
	assert.Equal(t, beforeSave, afterSave)

	allMsg, err := msgDb.ListMessage()
	assert.NoError(t, err)
	assert.LessOrEqual(t, 1, len(allMsg))

	unchainedMsgs, err := msgDb.ListUnchainedMsgs()
	assert.NoError(t, err)
	assert.LessOrEqual(t, 1, len(unchainedMsgs))
}

func TestUpdateMessageReceipt(t *testing.T) {
	db := setup("message.db")

	msg := NewSignedMessages(1)[0]
	unsignedCid := msg.UnsignedCid

	_, err := db.MessageRepo().SaveMessage(msg)
	assert.NoError(t, err)

	rec := &venustypes.MessageReceipt{
		ExitCode:    0,
		ReturnValue: []byte{'g', 'd'},
		GasUsed:     34,
	}
	height := abi.ChainEpoch(10)
	state := types.OnChainMsg
	_, err = db.MessageRepo().UpdateMessageReceipt(unsignedCid.String(), rec, height, state)
	assert.NoError(t, err)

	msg2, err := db.MessageRepo().GetMessageByCid(unsignedCid.String())
	assert.NoError(t, err)
	assert.Equal(t, uint64(height), msg2.Height)
	assert.Equal(t, rec, msg2.Receipt)
	assert.Equal(t, state, msg2.State)
}

func TestUpdateMessageStateByCid(t *testing.T) {
	db := setup("message.db")

	msg := NewSignedMessages(1)[0]
	msg.State = types.FillMsg
	cid := msg.UnsignedMessage.Cid()
	msg.UnsignedCid = &cid

	_, err := db.MessageRepo().SaveMessage(msg)
	assert.NoError(t, err)

	assert.NoError(t, db.MessageRepo().UpdateMessageStateByCid(cid.String(), types.OnChainMsg))

	msg2, err := db.MessageRepo().GetMessage(msg.ID)
	assert.NoError(t, err)
	assert.Equal(t, types.OnChainMsg, msg2.State)
}

func Test_sqliteMessageRepo_GetMessageState(t *testing.T) {

}

func TestSqliteMessageRepo_GetSignedMessageByTime(t *testing.T) {
	db := setup("message2.db")
	defer func() {
		assert.NoError(t, os.Remove("message2.db"))
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
	db := setup("message3.db")
	defer func() {
		assert.NoError(t, os.Remove("message3.db"))
	}()

	msgDb := db.MessageRepo()
	msg := NewMessage()
	_, err := msgDb.SaveMessage(msg)
	assert.NoError(t, err)

	signedMsgs := NewSignedMessages(10)
	for i, msg := range signedMsgs {
		msg.Height = uint64(i)
		_, err := msgDb.SaveMessage(msg)
		assert.NoError(t, err)
	}
	height := abi.ChainEpoch(5)
	msgs, err := msgDb.GetSignedMessageByHeight(height)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(msgs))
}
