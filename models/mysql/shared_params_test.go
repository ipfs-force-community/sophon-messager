package mysql

import (
	"context"
	"regexp"
	"testing"

	"gorm.io/gorm"

	"github.com/filecoin-project/go-state-types/big"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"

	"github.com/stretchr/testify/assert"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/ipfs-force-community/sophon-messager/models/repo"
)

func TestSharedParams(t *testing.T) {
	r, mock, sqlDB := setup(t)

	t.Run("mysql test get shared params", wrapper(testGetSharedParams, r, mock))
	t.Run("mysql test set shared params", wrapper(testSetSharedParams, r, mock))

	assert.NoError(t, closeDB(mock, sqlDB))
}

func testGetSharedParams(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `shared_params` LIMIT 1")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	res, err := r.SharedParamsRepo().GetSharedParams(ctx)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), res.ID)
}

func testSetSharedParams(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	ctx := context.Background()
	params := &types.SharedSpec{
		ID:        1,
		SelMsgNum: 10,
		FeeSpec: types.FeeSpec{
			GasOverEstimation: 1.25,
			MaxFee:            big.NewInt(100),
			GasFeeCap:         big.NewInt(1000),
			BaseFee:           big.NewInt(1001),
			GasOverPremium:    4.4,
		},
	}

	mysqlParams := fromSharedParams(*params)
	updateSql, updateArgs := genUpdateSQL(mysqlParams, true)
	updateArgs = append(updateArgs, uint(1))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `shared_params` WHERE id = ? LIMIT 1")).
		WithArgs(1).WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(updateSql)).
		WithArgs(updateArgs...).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	id, err := r.SharedParamsRepo().SetSharedParams(ctx, params) //create
	assert.NoError(t, err)
	assert.Equal(t, uint(1), id)

	// update params but ID not 1
	params2 := *params
	params2.ID = 3
	mysqlParams2 := fromSharedParams(params2)
	updateSql2, updateArgs2 := genUpdateSQL(mysqlParams2, true)
	updateArgs2 = append(updateArgs2, uint(1))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `shared_params` WHERE id = ? LIMIT 1")).
		WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"gas_over_estimation"}).AddRow(params.GasOverEstimation))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(updateSql2)).
		WithArgs(updateArgs2...).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	id, err = r.SharedParamsRepo().SetSharedParams(ctx, &params2) //update
	assert.NoError(t, err)
	assert.Equal(t, uint(1), id)
}
