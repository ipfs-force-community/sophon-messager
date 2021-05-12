package models

import (
	"testing"
	"time"

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
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		fc2 := &types.FeeConfig{
			ID:                types.NewUUID(),
			WalletID:          types.NewUUID(),
			MethodType:        1,
			GasOverEstimation: 1.2,
			MaxFee:            big.NewInt(11),
			MaxFeeCap:         big.NewInt(0),
			IsDeleted:         -1,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
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
