package sqlite

import (
	"context"
	"testing"

	"github.com/filecoin-project/go-state-types/big"
	"github.com/stretchr/testify/assert"

	"github.com/filecoin-project/venus/venus-shared/types/messager"
)

func TestSharedParams(t *testing.T) {
	r := setupRepo(t).SharedParamsRepo()

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
