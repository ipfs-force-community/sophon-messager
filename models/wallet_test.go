package models

import (
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/filecoin-project/venus-messager/models/repo"

	"github.com/stretchr/testify/assert"

	"github.com/filecoin-project/venus-messager/types"
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

		err := walletRepo.SaveWallet(w)
		assert.NoError(t, err)

		err = walletRepo.SaveWallet(w2)
		assert.NoError(t, err)

		r, err := walletRepo.GetWalletByID(w.ID)
		assert.NoError(t, err)
		assert.Equal(t, w.Name, r.Name)
		assert.Equal(t, w.Url, r.Url)

		r2, err := walletRepo.GetWalletByName(w.Name)
		assert.NoError(t, err)
		assert.Equal(t, w.Name, r2.Name)
		assert.Equal(t, w.Url, r2.Url)

		_, err = walletRepo.GetWalletByID(w2.ID)
		assert.Error(t, err)

		rs, err := walletRepo.ListWallet()
		assert.NoError(t, err)
		assert.LessOrEqual(t, 1, len(rs))

		err = walletRepo.DelWallet(w.Name)
		assert.NoError(t, err)

		_, err = walletRepo.GetWalletByID(w.ID)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	}
	t.Run("TestWallet", func(t *testing.T) {
		t.Run("sqlite", func(t *testing.T) {
			walletRepoTest(t, sqliteRepo.WalletRepo())
		})
		t.Run("mysql", func(t *testing.T) {
			t.SkipNow()
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
		err := walletRepo.SaveWallet(w)
		assert.NoError(t, err)

		has, err := walletRepo.HasWallet(w.Name)
		assert.NoError(t, err)
		assert.True(t, has)

		assert.NoError(t, walletRepo.DelWallet(w.Name))

		has, err = walletRepo.HasWallet(w.Name)
		assert.NoError(t, err)
		assert.True(t, has)
	}

	t.Run("Has_Wallet", func(t *testing.T) {
		t.Run("sqlite", func(t *testing.T) {
			walletRepoTest(t, sqliteRepo.WalletRepo())
		})
		t.Run("mysql", func(t *testing.T) {
			t.SkipNow()
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
		err := walletRepo.SaveWallet(w)
		assert.NoError(t, err)

		_, err = walletRepo.GetWalletByID(w.ID)
		assert.NoError(t, err)

		assert.NoError(t, walletRepo.DelWallet(w.Name))

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
		err := walletRepo.SaveWallet(w)
		assert.NoError(t, err)

		_, err = walletRepo.GetWalletByName(w.Name)
		assert.NoError(t, err)

		assert.NoError(t, walletRepo.DelWallet(w.Name))

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
