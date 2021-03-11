package service

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/ipfs-force-community/venus-messager/config"
	"github.com/ipfs-force-community/venus-messager/models/sqlite"
	"github.com/ipfs-force-community/venus-messager/types"
	"github.com/ipfs-force-community/venus-messager/utils"
)

func TestMessageStateCache(t *testing.T) {
	db, err := sqlite.OpenSqlite(&config.SqliteConfig{Path: "sqlite.db"})
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, os.Remove("sqlite.db"))
	}()
	assert.NoError(t, db.AutoMigrate())

	msgs := utils.NewTestSignedMsgs(10)
	for _, msg := range msgs {
		_, err := db.MessageRepo().SaveMessage(msg)
		assert.NoError(t, err)
	}

	ms, err := NewMessageState(db, logrus.New(), &config.MessageStateConfig{
		BackTime:          60,
		CleanupInterval:   3,
		DefaultExpiration: 2,
	})

	assert.NoError(t, err)
	assert.NoError(t, ms.loadRecentMessage())
	assert.Equal(t, 10, len(ms.idCids.cache))

	state, flag := ms.GetMessageState(msgs[0].Cid().String())
	assert.True(t, flag)
	assert.Equal(t, msgs[0].State, state)

	ms.SetMessageState(msgs[1].Cid().String(), types.OnChain)
	state, flag = ms.GetMessageState(msgs[1].Cid().String())
	assert.True(t, flag)
	assert.Equal(t, types.OnChain, state)

}
