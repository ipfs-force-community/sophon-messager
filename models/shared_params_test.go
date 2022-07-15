package models

import (
	"context"
	"testing"

	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/venus-messager/models/repo"
	"github.com/filecoin-project/venus/venus-shared/types/messager"
	"github.com/stretchr/testify/assert"
)

func TestSharedParams(t *testing.T) {
	sqliteRepo, mysqlRepo := setupRepo(t)

	sharedParamsTest := func(t *testing.T, r repo.SharedParamsRepo) {
		ctx := context.Background()
		params := &messager.SharedSpec{
			ID:                1,
			GasOverEstimation: 1.5,
			MaxFee:            big.NewInt(10),
			GasFeeCap:         big.NewInt(100),
			GasOverPremium:    1.6,
			SelMsgNum:         30,
		}
		_, err := r.SetSharedParams(ctx, params)
		assert.Nil(t, err)

		res, err := r.GetSharedParams(ctx)
		assert.Nil(t, err)
		assert.Equal(t, params, res)
	}

	t.Run("TestSharedParams", func(t *testing.T) {
		t.Run("sqlite", func(t *testing.T) {
			sharedParamsTest(t, sqliteRepo.SharedParamsRepo())
		})
		t.Run("mysql", func(t *testing.T) {
			t.SkipNow()
			sharedParamsTest(t, mysqlRepo.SharedParamsRepo())
		})
	})
}
