package models

import (
	"testing"

	"gorm.io/gorm"

	"github.com/filecoin-project/go-state-types/big"

	"github.com/filecoin-project/venus-messager/models/repo"
	"github.com/filecoin-project/venus-messager/types"
	"github.com/stretchr/testify/assert"
)

func TestFeeConfig(t *testing.T) {
	sqliteRepo, mysqlRepo := setupRepo(t)

	feeConfigRepoTest := func(t *testing.T, feeConfigRepo repo.FeeConfigRepo) {
		fc := &types.FeeConfig{
			ID:                types.NewUUID(),
			WalletID:          types.NewUUID(),
			MethodType:        0,
			GasOverEstimation: 1.25,
			MaxFee:            big.NewInt(0),
			MaxFeeCap:         big.NewInt(10),
			IsDeleted:         -1,
		}

		fc2 := &types.FeeConfig{
			ID:                types.NewUUID(),
			WalletID:          types.NewUUID(),
			MethodType:        1,
			GasOverEstimation: 1.2,
			MaxFee:            big.NewInt(11),
			MaxFeeCap:         big.NewInt(0),
			IsDeleted:         -1,
		}

		globalFC := &types.FeeConfig{
			ID:                types.DefGlobalFeeCfgID,
			WalletID:          types.UUID{},
			MethodType:        -1,
			GasOverEstimation: 1.2,
			MaxFee:            big.NewInt(110),
			MaxFeeCap:         big.NewInt(10),
			IsDeleted:         -1,
		}

		walletFC := &types.FeeConfig{
			ID:                types.NewUUID(),
			WalletID:          types.NewUUID(),
			MethodType:        -1,
			GasOverEstimation: 1.2,
			MaxFee:            big.NewInt(110),
			MaxFeeCap:         big.NewInt(10),
			IsDeleted:         -1,
		}

		t.Run("SaveFeeConfig", func(t *testing.T) {
			assert.NoError(t, feeConfigRepo.SaveFeeConfig(fc))
			assert.NoError(t, feeConfigRepo.SaveFeeConfig(fc2))
		})

		t.Run("GetFeeConfig", func(t *testing.T) {
			result, err := feeConfigRepo.GetFeeConfig(fc.WalletID, fc.MethodType)
			assert.NoError(t, err)
			assert.Equal(t, fc.ID, result.ID)
			assert.Equal(t, fc.GasOverEstimation, result.GasOverEstimation)
			assert.Equal(t, fc.MaxFeeCap, result.MaxFeeCap)
			assert.Equal(t, fc.MaxFee, result.MaxFee)
			assert.Equal(t, fc.IsDeleted, result.IsDeleted)

			result2, err := feeConfigRepo.GetFeeConfig(types.UUID{}, 0)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), gorm.ErrRecordNotFound.Error())
			assert.Nil(t, result2)

			assert.NoError(t, feeConfigRepo.DeleteFeeConfig(fc.WalletID, fc.MethodType))
			_, err = feeConfigRepo.GetFeeConfig(fc.WalletID, fc.MethodType)
			assert.Error(t, err)

		})

		t.Run("GetGlobalFeeConfig", func(t *testing.T) {
			_, err := feeConfigRepo.GetGlobalFeeConfig()
			assert.Contains(t, err.Error(), gorm.ErrRecordNotFound.Error())

			// save global fee config
			assert.NoError(t, feeConfigRepo.SaveFeeConfig(globalFC))

			fc, err := feeConfigRepo.GetGlobalFeeConfig()
			assert.NoError(t, err)
			assert.Equal(t, fc.GasOverEstimation, fc.GasOverEstimation)
			assert.Equal(t, fc.MaxFeeCap, fc.MaxFeeCap)
			assert.Equal(t, fc.MaxFee, fc.MaxFee)
			assert.Equal(t, fc.IsDeleted, fc.IsDeleted)

			assert.NoError(t, feeConfigRepo.DeleteFeeConfig(fc.WalletID, fc.MethodType))
		})

		t.Run("GetWalletFeeConfig", func(t *testing.T) {
			_, err := feeConfigRepo.GetWalletFeeConfig(walletFC.WalletID)
			assert.Contains(t, err.Error(), gorm.ErrRecordNotFound.Error())

			// save wallet fee config
			assert.NoError(t, feeConfigRepo.SaveFeeConfig(walletFC))

			fc, err := feeConfigRepo.GetWalletFeeConfig(walletFC.WalletID)
			assert.NoError(t, err)
			assert.Equal(t, fc.GasOverEstimation, fc.GasOverEstimation)
			assert.Equal(t, fc.MaxFeeCap, fc.MaxFeeCap)
			assert.Equal(t, fc.MaxFee, fc.MaxFee)
			assert.Equal(t, fc.IsDeleted, fc.IsDeleted)

			assert.NoError(t, feeConfigRepo.DeleteFeeConfig(fc.WalletID, fc.MethodType))
		})

		t.Run("HasFeeConfig", func(t *testing.T) {
			has, err := feeConfigRepo.HasFeeConfig(fc2.WalletID, fc2.MethodType)
			assert.NoError(t, err)
			assert.Equal(t, true, has)

			has, err = feeConfigRepo.HasFeeConfig(types.UUID{}, 0)
			assert.NoError(t, err)
			assert.Equal(t, false, has)

			has, err = feeConfigRepo.HasFeeConfig(fc.WalletID, fc.MethodType)
			assert.NoError(t, err)
			assert.Equal(t, false, has)
		})

		t.Run("ListFeeConfig", func(t *testing.T) {
			fcList, err := feeConfigRepo.ListFeeConfig()
			assert.NoError(t, err)
			assert.GreaterOrEqual(t, len(fcList), 1)

			assert.NoError(t, feeConfigRepo.DeleteFeeConfig(fc2.WalletID, fc2.MethodType))
			fcList2, err := feeConfigRepo.ListFeeConfig()
			assert.NoError(t, err)
			t.Log(fcList2)
			assert.Equal(t, 0, len(fcList2))
		})
	}

	t.Run("FeeConfig", func(t *testing.T) {
		t.Run("sqlite", func(t *testing.T) {
			feeConfigRepoTest(t, sqliteRepo.FeeConfigRepo())
		})
		t.Run("mysql", func(t *testing.T) {
			t.SkipNow()
			feeConfigRepoTest(t, mysqlRepo.FeeConfigRepo())
		})
	})
}
