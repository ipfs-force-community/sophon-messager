package sqlite

import (
	"encoding/json"
	"fmt"
	"github.com/ipfs-force-community/venus-messager/config"
	"github.com/ipfs-force-community/venus-messager/models/repo"
	"github.com/ipfs-force-community/venus-messager/testutils"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func objectToString(i interface{}) string {
	res, _ := json.MarshalIndent(i, "", " ")
	return string(res)
}

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

func TestMessage(t *testing.T) {
	msgDb := db.MessageRepo()
	msg := testutils.NewTestMsg()
	beforeSave := objectToString(msg)

	t.Logf("%s", beforeSave)

	uuid, err := msgDb.SaveMessage(msg)
	assert.NoError(t, err)

	result, err := msgDb.GetMessage(uuid)
	assert.NoError(t, err)
	afterSave := objectToString(result)

	assert.Equal(t, beforeSave, afterSave)

	t.Logf("%s", afterSave)

	allMsg, err := msgDb.ListMessage()
	assert.NoError(t, err)
	assert.LessOrEqual(t, 1, len(allMsg))

	unchainedMsgs, err := msgDb.ListUnchainedMsgs()
	assert.NoError(t, err)
	assert.LessOrEqual(t, 1, len(unchainedMsgs))
}
