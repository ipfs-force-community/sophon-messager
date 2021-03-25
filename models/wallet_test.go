package models

import (
	"github.com/ipfs-force-community/venus-messager/models/repo"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ipfs-force-community/venus-messager/types"
)

func TestWallet(t *testing.T) {
	sqliteRepo, mysqlRepo := setupRepo(t)

	walletRepoTest := func(t *testing.T, walletRepo repo.WalletRepo) {

		w := &types.Wallet{
			ID:   types.NewUUID(),
			Name: types.NewUUID().String(),
			Url:  "http://127.0.0.1:8080",

			IsDeleted: -1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		w2 := &types.Wallet{
			ID:   types.NewUUID(),
			Name: types.NewUUID().String(),
			Url:  "http://127.0.0.1:8080",

			IsDeleted: 1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		id, err := walletRepo.SaveWallet(w)
		assert.NoError(t, err)

		w.ID = types.NewUUID()
		_, err = walletRepo.SaveWallet(w)
		assert.Error(t, err)

		id2, err := walletRepo.SaveWallet(w2)
		assert.NoError(t, err)

		r, err := walletRepo.GetWalletByID(id)
		assert.NoError(t, err)
		assert.Equal(t, w.Name, r.Name)
		assert.Equal(t, w.Url, r.Url)

		r2, err := walletRepo.GetWalletByName(w.Name)
		assert.NoError(t, err)
		assert.Equal(t, w.Name, r2.Name)
		assert.Equal(t, w.Url, r2.Url)

		_, err = walletRepo.GetWalletByID(id2)
		assert.Error(t, err)

		rs, err := walletRepo.ListWallet()
		assert.NoError(t, err)
		assert.LessOrEqual(t, 1, len(rs))

		err = walletRepo.DelWallet(id)
		assert.NoError(t, err)

		_, err = walletRepo.GetWalletByID(id)
		assert.Error(t, err)
	}
	t.Run("TestWallet", func(t *testing.T) {
		t.Run("sqlite", func(t *testing.T) {
			walletRepoTest(t, sqliteRepo.WalletRepo())
		})
		t.Run("mysql", func(t *testing.T) {
			walletRepoTest(t, mysqlRepo.WalletRepo())
		})
	})

}

func TestSqliteWalletRepo_HasWallet(t *testing.T) {
	sqliteRepo, mysqlRepo := setupRepo(t)

	walletRepoTest := func(t *testing.T, walletRepo repo.WalletRepo) {
		w := &types.Wallet{
			ID:   types.NewUUID(),
			Name: types.NewUUID().String(),
			Url:  "http://127.0.0.1:8080",

			IsDeleted: -1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		_, err := walletRepo.SaveWallet(w)
		assert.NoError(t, err)

		has, err := walletRepo.HasWallet(w.Name)
		assert.NoError(t, err)
		assert.True(t, has)

		assert.NoError(t, walletRepo.DelWallet(w.ID))

		has, err = walletRepo.HasWallet(w.Name)
		assert.NoError(t, err)
		assert.True(t, has)
	}

	t.Run("Has_Wallet", func(t *testing.T) {
		t.Run("sqlite", func(t *testing.T) {
			walletRepoTest(t, sqliteRepo.WalletRepo())
		})
		t.Run("mysql", func(t *testing.T) {
			walletRepoTest(t, mysqlRepo.WalletRepo())
		})
	})
}

func TestSqliteWalletRepo_GetWalletByID(t *testing.T) {
	sqliteRepo, mysqlRepo := setupRepo(t)

	walletRepoTest := func(t *testing.T, walletRepo repo.WalletRepo) {
		w := &types.Wallet{
			ID:   types.NewUUID(),
			Name: types.NewUUID().String(),
			Url:  "http://127.0.0.1:8080",

			IsDeleted: -1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		_, err := walletRepo.SaveWallet(w)
		assert.NoError(t, err)

		_, err = walletRepo.GetWalletByID(w.ID)
		assert.NoError(t, err)

		assert.NoError(t, walletRepo.DelWallet(w.ID))

		_, err = walletRepo.GetWalletByID(w.ID)
		assert.Containsf(t, err.Error(), "record not found", "expect not found error")
	}

	t.Run("GetWalletByID", func(t *testing.T) {
		t.Run("sqlite", func(t *testing.T) {
			walletRepoTest(t, sqliteRepo.WalletRepo())
		})
		t.Run("mysql", func(t *testing.T) {
			t.SkipNow()
			walletRepoTest(t, mysqlRepo.WalletRepo())
		})
	})
}

func TestSqliteWalletRepo_GetWalletByName(t *testing.T) {
	sqliteRepo, mysqlRepo := setupRepo(t)

	walletRepoTest := func(t *testing.T, walletRepo repo.WalletRepo) {
		w := &types.Wallet{
			ID:   types.NewUUID(),
			Name: types.NewUUID().String(),
			Url:  "http://127.0.0.1:8080",

			IsDeleted: -1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		_, err := walletRepo.SaveWallet(w)
		assert.NoError(t, err)

		_, err = walletRepo.GetWalletByName(w.Name)
		assert.NoError(t, err)

		assert.NoError(t, walletRepo.DelWallet(w.ID))

		_, err = walletRepo.GetWalletByName(w.Name)
		assert.Containsf(t, err.Error(), "record not found", "expect not found error")
	}

	t.Run("GetWalletByID", func(t *testing.T) {
		t.Run("sqlite", func(t *testing.T) {
			walletRepoTest(t, sqliteRepo.WalletRepo())
		})
		t.Run("mysql", func(t *testing.T) {
			t.SkipNow()
			walletRepoTest(t, mysqlRepo.WalletRepo())
		})
	})
}
