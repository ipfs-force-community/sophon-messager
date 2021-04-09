package models

import (
	"testing"

	"github.com/ipfs-force-community/venus-messager/models/repo"
	"github.com/ipfs-force-community/venus-messager/types"
	"github.com/stretchr/testify/assert"
)

func TestWalletAddress(t *testing.T) {
	sqliteRepo, mysqlRepo := setupRepo(t)

	walletAddrRepoTest := func(t *testing.T, waRepo repo.WalletAddressRepo) {
		wa := &types.WalletAddress{
			ID:           types.NewUUID(),
			WalletID:     types.NewUUID(),
			AddrID:       types.NewUUID(),
			AddressState: types.Alive,
			SelMsgNum:    10,
			IsDeleted:    -1,
		}

		wa2 := &types.WalletAddress{
			ID:           types.NewUUID(),
			WalletID:     types.NewUUID(),
			AddrID:       types.NewUUID(),
			AddressState: types.Alive,
			SelMsgNum:    10,
			IsDeleted:    -1,
		}

		err := waRepo.SaveWalletAddress(wa)
		assert.NoError(t, err)
		err = waRepo.SaveWalletAddress(wa2)
		assert.NoError(t, err)
		r, err := waRepo.GetWalletAddress(wa.WalletID, wa.AddrID)
		assert.NoError(t, err)
		assert.Equal(t, wa.AddressState, r.AddressState)
		assert.Equal(t, wa.SelMsgNum, r.SelMsgNum)
		assert.Equal(t, wa.IsDeleted, r.IsDeleted)
		assert.Equal(t, wa.ID, r.ID)

		newState := types.Removing
		err = waRepo.UpdateAddressState(wa.WalletID, wa.AddrID, newState)
		assert.NoError(t, err)
		r2, err := waRepo.GetWalletAddress(wa.WalletID, wa.AddrID)
		assert.NoError(t, err)
		assert.Equal(t, newState, r2.AddressState)

		selMsgNum := uint64(50)
		err = waRepo.UpdateSelectMsgNum(wa.WalletID, wa.AddrID, selMsgNum)
		assert.NoError(t, err)
		r3, err := waRepo.GetWalletAddress(wa.WalletID, wa.AddrID)
		assert.NoError(t, err)
		assert.Equal(t, selMsgNum, r3.SelMsgNum)

		err = waRepo.DelWalletAddress(wa.WalletID, wa.AddrID)
		assert.NoError(t, err)
		r, err = waRepo.GetWalletAddress(wa.WalletID, wa.AddrID)
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
