package sqlite

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ipfs-force-community/venus-messager/config"
	"github.com/ipfs-force-community/venus-messager/types"
)

func TestAddress(t *testing.T) {
	path := "sqlite_address.db"
	repo, err := OpenSqlite(&config.SqliteConfig{Path: path})
	assert.NoError(t, err)
	assert.NoError(t, repo.AutoMigrate())
	defer func() {
		assert.NoError(t, os.Remove(path))
	}()

	addressRepo := repo.AddressRepo()

	a := &types.Address{
		ID:        types.NewUUID(),
		Addr:      "test1",
		Nonce:     3,
		IsDeleted: -1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	a2 := &types.Address{
		ID:        types.NewUUID(),
		Addr:      "test2",
		Nonce:     2,
		IsDeleted: -1,
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	}

	ctx := context.Background()

	list, err := addressRepo.ListAddress(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(list))

	_, err = addressRepo.SaveAddress(ctx, a)
	assert.NoError(t, err)
	_, err = addressRepo.SaveAddress(ctx, a2)
	assert.NoError(t, err)

	r, err := addressRepo.GetAddress(ctx, a.Addr)
	assert.NoError(t, err)
	assert.Equal(t, a.Nonce, r.Nonce)

	newNonce := uint64(5)
	_, err = addressRepo.UpdateNonce(ctx, a.ID, newNonce)
	assert.NoError(t, err)
	r2, err := addressRepo.GetAddress(ctx, a.Addr)
	assert.NoError(t, err)
	assert.Equal(t, newNonce, r2.Nonce)

	err = addressRepo.DelAddress(ctx, a.Addr)
	assert.NoError(t, err)

	r, err = addressRepo.GetAddress(ctx, a.Addr)
	assert.Error(t, err)
	assert.Nil(t, r)

	rs, err := addressRepo.ListAddress(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(rs))
}
