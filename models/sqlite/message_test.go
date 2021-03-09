package sqlite

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/crypto"
	venustypes "github.com/filecoin-project/venus/pkg/types"
	"github.com/stretchr/testify/assert"

	"github.com/ipfs-force-community/venus-messager/config"
	"github.com/ipfs-force-community/venus-messager/models/repo"
	"github.com/ipfs-force-community/venus-messager/types"
	"github.com/ipfs-force-community/venus-messager/utils"
)

var db repo.Repo

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	shutdown()
	os.Exit(code)
}

func shutdown() {
	if db == nil {
		return
	}
	if err := db.DbClose(); err != nil {
		fmt.Printf("shutdown postgre client failed. %v", err)
	}
}

func setup() {
	var err error
	if db, err = OpenSqlite(&config.SqliteConfig{Path: "sqlite.db"}); err != nil {
		panic(err)
	}
	if err = db.AutoMigrate(); err != nil {
		panic(err)
	}
}

func TestSaveMessages(t *testing.T) {
	msg := utils.NewTestMsg()

	_, err := db.MessageRepo().SaveMessage(msg)
	assert.NoError(t, err)
}

func TestUpdateMessageReceipt(t *testing.T) {
	msg := utils.NewTestMsg()
	msg.Signature = &crypto.Signature{Type: crypto.SigTypeBLS, Data: []byte{1, 2, 3}}
	signedCid := msg.SignedCid()

	_, err := db.MessageRepo().SaveMessage(msg)
	assert.NoError(t, err)

	rec := &venustypes.MessageReceipt{
		ExitCode:    0,
		ReturnValue: []byte{'g', 'd'},
		GasUsed:     34,
	}
	height := abi.ChainEpoch(10)
	state := types.OnChain
	_, err = db.MessageRepo().UpdateMessageReceipt(signedCid.String(), rec, height, state)
	assert.NoError(t, err)

	msg2, err := db.MessageRepo().GetMessageByCid(signedCid.String())
	assert.NoError(t, err)
	assert.Equal(t, uint64(height), msg2.Height)
	assert.Equal(t, rec, msg2.Receipt)
	assert.Equal(t, state, msg2.State)
}

func TestMessage(t *testing.T) {
	msgDb := db.MessageRepo()
	msg := utils.NewTestMsg()
	beforeSave := utils.ObjectToString(msg)

	uuid, err := msgDb.SaveMessage(msg)
	assert.NoError(t, err)

	result, err := msgDb.GetMessage(uuid)
	assert.NoError(t, err)

	afterSave := utils.ObjectToString(result)
	assert.Equal(t, beforeSave, afterSave)

	allMsg, err := msgDb.ListMessage()
	assert.NoError(t, err)
	assert.LessOrEqual(t, 1, len(allMsg))

	unchainedMsgs, err := msgDb.ListUnchainedMsgs()
	assert.NoError(t, err)
	assert.LessOrEqual(t, 1, len(unchainedMsgs))

	startTime := time.Now().Add(-time.Second * 3600)
	msgs, err := msgDb.GetMessageByTime(startTime)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(msgs))
}

func TestUpdateMessageStateByCid(t *testing.T) {
	msg := utils.NewTestMsg()
	msg.Signature = &crypto.Signature{Type: crypto.SigTypeBLS, Data: []byte{1, 2, 3}}
	msg.State = types.Signed
	signedCid := msg.SignedCid()

	_, err := db.MessageRepo().SaveMessage(msg)
	assert.NoError(t, err)

	assert.NoError(t, db.MessageRepo().UpdateMessageStateByCid(signedCid.String(), types.OnChain))

	msg2, err := db.MessageRepo().GetMessage(msg.ID)
	assert.NoError(t, err)
	assert.Equal(t, types.OnChain, msg2.State)
}
