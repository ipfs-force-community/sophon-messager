package models

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ipfs-force-community/venus-messager/models/repo"

	"github.com/filecoin-project/go-address"
	"github.com/stretchr/testify/assert"

	"github.com/ipfs-force-community/venus-messager/types"
)

func TestAddress(t *testing.T) {
	sqliteRepo, mysqlRepo := setupRepo(t)

	addressRepoTest := func(t *testing.T, addressRepo repo.AddressRepo) {
		rand.Seed(time.Now().Unix())
		addr, err := address.NewIDAddress(rand.Uint64() / 2)
		assert.NoError(t, err)
		addr2, err := address.NewIDAddress(rand.Uint64() / 2)
		assert.NoError(t, err)

		addrInfo := &types.Address{
			ID:        types.NewUUID(),
			Addr:      addr,
			Nonce:     3,
			IsDeleted: -1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		addrInfo2 := &types.Address{
			ID:        types.NewUUID(),
			Addr:      addr2,
			Nonce:     2,
			IsDeleted: -1,
			CreatedAt: time.Time{},
			UpdatedAt: time.Time{},
		}

		ctx := context.Background()

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
		assert.LessOrEqual(t, 1, len(rs))
	}

	t.Run("TestAddress", func(t *testing.T) {
		t.Run("sqlite", func(t *testing.T) {
			addressRepoTest(t, sqliteRepo.AddressRepo())
		})
		t.Run("mysql", func(t *testing.T) {
			t.SkipNow()
			addressRepoTest(t, mysqlRepo.AddressRepo())
		})
	})
}

func TestSqliteAddressRepo_UpdateAddressState(t *testing.T) {
	sqliteRepo, mysqlRepo := setupRepo(t)

	addressRepoTest := func(t *testing.T, addressRepo repo.AddressRepo) {
		addr, err := address.NewIDAddress(rand.Uint64() / 2)
		assert.NoError(t, err)

		addrInfo := &types.Address{
			ID:        types.NewUUID(),
			Addr:      addr,
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

	t.Run("UpdateAddressState", func(t *testing.T) {
		t.Run("sqlite", func(t *testing.T) {
			addressRepoTest(t, sqliteRepo.AddressRepo())
		})
		t.Run("mysql", func(t *testing.T) {
			t.SkipNow()
			addressRepoTest(t, mysqlRepo.AddressRepo())
		})
	})
}

func TestSqliteAddressRepo_UpdateAddressNonce(t *testing.T) {
	sqliteRepo, mysqlRepo := setupRepo(t)

	addressRepoTest := func(t *testing.T, addressRepo repo.AddressRepo) {
		uid, err := uuid.NewUUID()
		assert.NoError(t, err)
		addr, err := address.NewActorAddress(uid[:])
		assert.NoError(t, err)

		addrInfo := &types.Address{
			ID:        types.NewUUID(),
			Addr:      addr,
			Nonce:     3,
			IsDeleted: -1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		_, err = addressRepo.SaveAddress(context.TODO(), addrInfo)
		assert.NoError(t, err)

		result, err := addressRepo.GetAddress(context.TODO(), addr)
		assert.NoError(t, err)
		assert.EqualValues(t, 3, result.Nonce)

		_, err = addressRepo.UpdateNonce(context.TODO(), addr, 1000)
		assert.NoError(t, err)

		addrInfo, err = addressRepo.GetAddress(context.TODO(), addr)
		assert.NoError(t, err)
		assert.EqualValues(t, 1000, addrInfo.Nonce)
	}

	t.Run("UpdateAddressNonce", func(t *testing.T) {
		t.Run("sqlite", func(t *testing.T) {
			addressRepoTest(t, sqliteRepo.AddressRepo())
		})
		t.Run("mysql", func(t *testing.T) {
			t.SkipNow()
			addressRepoTest(t, mysqlRepo.AddressRepo())
		})
	})
}

func TestSqliteAddressRepo_Delete(t *testing.T) {
	sqliteRepo, mysqlRepo := setupRepo(t)

	addressRepoTest := func(t *testing.T, addressRepo repo.AddressRepo) {
		id, err := uuid.NewUUID()
		assert.NoError(t, err)
		addr, err := address.NewActorAddress(id[:])
		assert.NoError(t, err)
		addrInfo := &types.Address{
			ID:        types.NewUUID(),
			Addr:      addr,
			Nonce:     3,
			IsDeleted: -1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		_, err = addressRepo.SaveAddress(context.TODO(), addrInfo)
		assert.NoError(t, err)

		_, err = addressRepo.GetAddress(context.TODO(), addr)
		assert.NoError(t, err)

		err = addressRepo.DelAddress(context.TODO(), addr)
		assert.NoError(t, err)

		addrInfo, err = addressRepo.GetAddress(context.TODO(), addr)
		assert.Error(t, err)
		assert.Containsf(t, err.Error(), "record not found", "expect not found error")
	}

	t.Run("DeleteAddress", func(t *testing.T) {
		t.Run("sqlite", func(t *testing.T) {
			addressRepoTest(t, sqliteRepo.AddressRepo())
		})
		t.Run("mysql", func(t *testing.T) {
			t.SkipNow()
			addressRepoTest(t, mysqlRepo.AddressRepo())
		})
	})
}

func TestSqliteAddressRepo_HasAddress(t *testing.T) {
	sqliteRepo, mysqlRepo := setupRepo(t)

	addressRepoTest := func(t *testing.T, addressRepo repo.AddressRepo) {
		uid, err := uuid.NewUUID()
		assert.NoError(t, err)
		addr, err := address.NewActorAddress(uid[:])
		assert.NoError(t, err)

		addrInfo := &types.Address{
			ID:        types.NewUUID(),
			Addr:      addr,
			Nonce:     3,
			IsDeleted: -1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		_, err = addressRepo.SaveAddress(context.TODO(), addrInfo)
		assert.NoError(t, err)

		has, err := addressRepo.HasAddress(context.TODO(), addr)
		assert.NoError(t, err)
		assert.Equal(t, has, true)

		err = addressRepo.DelAddress(context.TODO(), addr)
		assert.NoError(t, err)

		_, err = addressRepo.GetAddress(context.TODO(), addr)
		assert.Error(t, err)
		assert.Containsf(t, err.Error(), "record not found", "expect not found error")
	}

	t.Run("HasAddress", func(t *testing.T) {
		t.Run("sqlite", func(t *testing.T) {
			addressRepoTest(t, sqliteRepo.AddressRepo())
		})
		t.Run("mysql", func(t *testing.T) {
			t.SkipNow()
			addressRepoTest(t, mysqlRepo.AddressRepo())
		})
	})
}
