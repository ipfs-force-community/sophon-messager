package sqlite

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ipfs-force-community/venus-messager/config"
	"github.com/ipfs-force-community/venus-messager/types"
)

func newWallet() *types.Wallet {
	return &types.Wallet{
		ID:   types.NewUUID(),
		Name: types.NewUUID().String(),
		Url:  "http://127.0.0.1:8080",

		IsDeleted: -1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func TestWallet(t *testing.T) {
	path := "sqlite_wallet.db"
	repo, err := OpenSqlite(&config.SqliteConfig{Path: path})
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, os.Remove(path))
	}()
	assert.NoError(t, repo.AutoMigrate())

	walletRepo := repo.WalletRepo()

	w := newWallet()
	w2 := newWallet()
	w2.IsDeleted = 1

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
	assert.Equal(t, 1, len(rs))

	err = walletRepo.DelWallet(id)
	assert.NoError(t, err)

	_, err = walletRepo.GetWalletByID(id)
	assert.Error(t, err)
}

func TestSqliteWalletRepo_HasWallet(t *testing.T) {
	path := "HasWallet.db"
	repo, err := OpenSqlite(&config.SqliteConfig{Path: path})
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, os.Remove(path))
	}()
	assert.NoError(t, repo.AutoMigrate())

	walletRepo := repo.WalletRepo()

	w := newWallet()
	_, err = walletRepo.SaveWallet(w)
	assert.NoError(t, err)

	has, err := walletRepo.HasWallet(w.Name)
	assert.NoError(t, err)
	assert.True(t, has)

	assert.NoError(t, walletRepo.DelWallet(w.ID))

	has, err = walletRepo.HasWallet(w.Name)
	assert.NoError(t, err)
	assert.True(t, has)
}
