package sqlite

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/filecoin-project/go-address"
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

	addr, err := address.NewFromString("f01000")
	assert.NoError(t, err)
	addr2, err := address.NewFromString("f01001")
	assert.NoError(t, err)

	addrInfo := &types.Address{
		ID:        types.NewUUID(),
		Addr:      addr.String(),
		Nonce:     3,
		IsDeleted: -1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	addrInfo2 := &types.Address{
		ID:        types.NewUUID(),
		Addr:      addr2.String(),
		Nonce:     2,
		IsDeleted: -1,
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	}

	ctx := context.Background()

	list, err := addressRepo.ListAddress(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(list))

	_, err = addressRepo.SaveAddress(ctx, addrInfo)
	assert.NoError(t, err)
	_, err = addressRepo.SaveAddress(ctx, addrInfo2)
	assert.NoError(t, err)

	r, err := addressRepo.GetAddress(ctx, addr)
	assert.NoError(t, err)
	assert.Equal(t, addrInfo.Nonce, r.Nonce)

	newNonce := uint64(5)
	_, err = addressRepo.UpdateNonce(ctx, addr, newNonce)
	assert.NoError(t, err)
	r2, err := addressRepo.GetAddress(ctx, addr)
	assert.NoError(t, err)
	assert.Equal(t, newNonce, r2.Nonce)

	err = addressRepo.DelAddress(ctx, addr)
	assert.NoError(t, err)

	r, err = addressRepo.GetAddress(ctx, addr)
	assert.Error(t, err)
	assert.Nil(t, r)

	rs, err := addressRepo.ListAddress(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(rs))
}

func TestSqliteAddressRepo_UpdateAddressState(t *testing.T) {
	path := "UpdateAddressState.db"
	repo, err := OpenSqlite(&config.SqliteConfig{Path: path})
	assert.NoError(t, err)
	assert.NoError(t, repo.AutoMigrate())
	defer func() {
		assert.NoError(t, os.Remove(path))
	}()

	addressRepo := repo.AddressRepo()
	addr, err := address.NewFromString("f01000")
	assert.NoError(t, err)

	addrInfo := &types.Address{
		ID:        types.NewUUID(),
		Addr:      addr.String(),
		Nonce:     3,
		IsDeleted: -1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err = addressRepo.SaveAddress(context.TODO(), addrInfo)
	assert.NoError(t, err)

	result, err := addressRepo.GetAddress(context.TODO(), addr)
	assert.NoError(t, err)
	assert.Equal(t, types.Alive, result.State)

	_, err = addressRepo.UpdateAddressState(context.TODO(), addr, types.Notfound)
	assert.NoError(t, err)

	addrInfo, err = addressRepo.GetAddress(context.TODO(), addr)
	assert.NoError(t, err)
	assert.Equal(t, types.Notfound, addrInfo.State)
}
