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
		ID:        1,
		SelMsgNum: 30,
		FeeSpec: messager.FeeSpec{
			GasOverEstimation: 1.5,
			MaxFee:            big.NewInt(10),
			GasFeeCap:         big.NewInt(100),
			BaseFee:           big.NewInt(1000),
			GasOverPremium:    1.6,
		},
	}
	_, err := r.SetSharedParams(ctx, params)
	assert.Nil(t, err)

	res, err := r.GetSharedParams(ctx)
	assert.Nil(t, err)
	assert.Equal(t, params, res)

	// update params but ID not 1
	params2 := &messager.SharedSpec{
		ID:        3,
		SelMsgNum: 10,
		FeeSpec: messager.FeeSpec{
			GasOverEstimation: 3.5,
			MaxFee:            big.NewInt(310),
			GasFeeCap:         big.NewInt(3100),
			BaseFee:           big.NewInt(0),
			GasOverPremium:    3.6,
		},
	}
	_, err = r.SetSharedParams(ctx, params2)
	assert.Nil(t, err)

	res, err = r.GetSharedParams(ctx)
	assert.Nil(t, err)
	params2.ID = 1
	assert.Equal(t, params2, res)
}
