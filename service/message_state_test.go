package service

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/ipfs-force-community/venus-messager/config"
	"github.com/ipfs-force-community/venus-messager/models/sqlite"
	"github.com/ipfs-force-community/venus-messager/types"
)

func TestMessageStateCache(t *testing.T) {
	db, err := sqlite.OpenSqlite(&config.SqliteConfig{Path: "sqlite.db"})
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, os.Remove("sqlite.db"))
	}()
	assert.NoError(t, db.AutoMigrate())

	msgs := sqlite.NewSignedMessages(10)
	for _, msg := range msgs {
		_, err := db.MessageRepo().SaveMessage(msg)
		assert.NoError(t, err)
	}

	msgState, err := NewMessageState(db, logrus.New(), &config.MessageStateConfig{
		BackTime:          60,
		CleanupInterval:   3,
		DefaultExpiration: 2,
	})
	assert.NoError(t, err)

	msgList, err := msgState.repo.MessageRepo().ListMessage()
	assert.NoError(t, err)
	assert.Equal(t, 10, len(msgList))

	assert.NoError(t, msgState.loadRecentMessage())
	assert.Equal(t, 10, len(msgState.idCids.cache))

	state, flag := msgState.GetMessageStateByCid(msgs[0].Cid().String())
	assert.True(t, flag)
	assert.Equal(t, msgs[0].State, state)

	msgState.UpdateMessageStateByCid(msgs[1].Cid().String(), types.OnChainMsg)
	state, flag = msgState.GetMessageStateByCid(msgs[1].Cid().String())
	assert.True(t, flag)
	assert.Equal(t, types.OnChainMsg, state)
}
