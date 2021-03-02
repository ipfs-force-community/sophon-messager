package sqlite

import (
	"testing"
	"time"

	"github.com/ipfs-force-community/venus-messager/config"
	"github.com/ipfs-force-community/venus-messager/types"
	"github.com/stretchr/testify/assert"
)

func TestMessage(t *testing.T) {
	repo, err := OpenSqlite(&config.SqliteConfig{Path: "sqlite.db"})
	assert.NoError(t, err)
	//defer func() {
	//	//	assert.NoError(t, repo.DbClose())
	//	//}()
	err = repo.AutoMigrate()
	assert.NoError(t, err)

	messageRepo := repo.MessageRepo()

	msg := &types.Message{
		Id:         "22222222222",
		Version:    0,
		To:         "11",
		From:       "22",
		Nonce:      0,
		Value:      nil,
		GasLimit:   0,
		GasFeeCap:  nil,
		GasPremium: nil,
		Method:     0,
		Params:     nil,
		SignData:   nil,
		IsDeleted:  -1,
		CreatedAt:  time.Time{},
		UpdatedAt:  time.Time{},
	}

	id, err := messageRepo.SaveMessage(msg)
	assert.NoError(t, err)

	result, err := messageRepo.GetMessage(id)
	assert.NoError(t, err)
	t.Logf("%+v", result)

	results, err := messageRepo.ListMessage()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(results))
}
