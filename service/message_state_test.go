package service

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/filecoin-project/venus-messager/config"
	"github.com/filecoin-project/venus-messager/log"
	"github.com/filecoin-project/venus-messager/models"
	"github.com/filecoin-project/venus-messager/models/sqlite"
	"github.com/filecoin-project/venus-messager/types"
)

func TestMessageStateCache(t *testing.T) {
	db, err := sqlite.OpenSqlite(&config.SqliteConfig{Path: "message_state.db"})
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, os.Remove("message_state.db"))
		assert.NoError(t, os.Remove("message_state.db-shm"))
		assert.NoError(t, os.Remove("message_state.db-wal"))
	}()
	assert.NoError(t, db.AutoMigrate())

	msgs := models.NewSignedMessages(10)
	for _, msg := range msgs {
		err := db.MessageRepo().CreateMessage(msg)
		assert.NoError(t, err)
	}

	msgState, err := NewMessageState(db, log.New(), &config.MessageStateConfig{
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

	err = msgState.UpdateMessageByCid(msgs[1].Cid(), func(message *types.Message) error {
		message.State = types.OnChainMsg
		return nil
	})
	assert.NoError(t, err)
	state, flag = msgState.GetMessageStateByCid(msgs[1].Cid().String())
	assert.True(t, flag)
	assert.Equal(t, types.OnChainMsg, state)
}
