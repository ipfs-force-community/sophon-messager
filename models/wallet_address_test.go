package models

import (
	"math/rand"
	"testing"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/ipfs-force-community/venus-messager/models/repo"
	"github.com/ipfs-force-community/venus-messager/types"
	"github.com/stretchr/testify/assert"
)

func TestWalletAddress(t *testing.T) {
	sqliteRepo, mysqlRepo := setupRepo(t)

	walletAddrRepoTest := func(t *testing.T, waRepo repo.WalletAddressRepo) {
		rand.Seed(time.Now().Unix())
		addr, err := address.NewIDAddress(rand.Uint64() / 2)
		assert.NoError(t, err)
		addr2, err := address.NewIDAddress(rand.Uint64() / 2)
		assert.NoError(t, err)
		wa := &types.WalletAddress{
			ID:           types.NewUUID(),
			WalletName:   "venus_wallet",
			Addr:         addr,
			AddressState: types.Alive,
			SelMsgNum:    10,
			IsDeleted:    -1,
		}

		wa2 := &types.WalletAddress{
			ID:           types.NewUUID(),
			WalletName:   "venus_wallet",
			Addr:         addr2,
			AddressState: types.Alive,
			SelMsgNum:    10,
			IsDeleted:    -1,
		}

		err = waRepo.SaveWalletAddress(wa)
		assert.NoError(t, err)
		err = waRepo.SaveWalletAddress(wa2)
		assert.NoError(t, err)
		r, err := waRepo.GetWalletAddress(wa.WalletName, wa.Addr)
		assert.NoError(t, err)
		assert.Equal(t, wa.AddressState, r.AddressState)
		assert.Equal(t, wa.SelMsgNum, r.SelMsgNum)
		assert.Equal(t, wa.IsDeleted, r.IsDeleted)
		assert.Equal(t, wa.ID, r.ID)

		newState := types.Removing
		err = waRepo.UpdateAddressState(wa.WalletName, wa.Addr, newState)
		assert.NoError(t, err)
		r2, err := waRepo.GetWalletAddress(wa.WalletName, wa.Addr)
		assert.NoError(t, err)
		assert.Equal(t, newState, r2.AddressState)

		selMsgNum := uint64(50)
		err = waRepo.UpdateSelectMsgNum(wa.WalletName, wa.Addr, selMsgNum)
		assert.NoError(t, err)
		r3, err := waRepo.GetWalletAddress(wa.WalletName, wa.Addr)
		assert.NoError(t, err)
		assert.Equal(t, selMsgNum, r3.SelMsgNum)

		err = waRepo.DelWalletAddress(wa.WalletName, wa.Addr)
		assert.NoError(t, err)
		r, err = waRepo.GetWalletAddress(wa.WalletName, wa.Addr)
		assert.Error(t, err)
		assert.Nil(t, r)

		rs, err := waRepo.ListWalletAddress()
		assert.NoError(t, err)
		assert.LessOrEqual(t, 1, len(rs))
	}

	t.Run("TestAddress", func(t *testing.T) {
		t.Run("sqlite", func(t *testing.T) {
			walletAddrRepoTest(t, sqliteRepo.WalletAddressRepo())
		})
		t.Run("mysql", func(t *testing.T) {
			t.SkipNow()
			walletAddrRepoTest(t, mysqlRepo.WalletAddressRepo())
		})
	})
}
