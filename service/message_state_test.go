package service

import (
	"os"
	"testing"

	venustypes "github.com/filecoin-project/venus/pkg/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/ipfs-force-community/venus-messager/config"
	"github.com/ipfs-force-community/venus-messager/models/sqlite"
	"github.com/ipfs-force-community/venus-messager/types"
	"github.com/ipfs-force-community/venus-messager/utils"
)

func TestMessageCache(t *testing.T) {
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

	ms := NewMessageState(db, logrus.New(), &config.MessageStateConfig{
		BackTime:          60,
		CleanupInterval:   3,
		DefaultExpiration: 2,
	})

	assert.NoError(t, ms.loadRecentMessage())
	assert.Equal(t, 10, len(ms.idCids.cache))

	r, flag := ms.GetMessage(msgs[0].ID)
	assert.True(t, flag)
	assert.Equal(t, utils.ObjectToString(msgs[0]), utils.ObjectToString(r))

	rec := &venustypes.MessageReceipt{
		ReturnValue: []byte{'2', '3'},
		ExitCode:    0,
		GasUsed:     10,
	}
	ms.UpdateMessageStateAndReceipt(msgs[1].SignedCid().String(), types.OnChain, rec)
	r, flag = ms.GetMessage(msgs[1].ID)
	assert.True(t, flag)
	assert.Equal(t, rec, r.Receipt)
	assert.Equal(t, types.OnChain, r.State)

}
