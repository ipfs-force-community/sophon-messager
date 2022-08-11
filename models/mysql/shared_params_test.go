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

	"github.com/filecoin-project/venus-messager/models/repo"
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
		ID:                1,
		GasOverEstimation: 1.25,
		MaxFee:            big.NewInt(100),
		GasFeeCap:         big.NewInt(1000),
		GasOverPremium:    4.4,
		SelMsgNum:         10,
	}

	mysqlParams := fromSharedParams(*params)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `shared_params` WHERE id = ? LIMIT 1")).
		WithArgs(1).WillReturnError(gorm.ErrRecordNotFound)

	args := getStructFieldValue(mysqlParams)
	args[0], args[len(args)-1] = args[len(args)-1], args[0]

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(genUpdateSQL(mysqlParams))).
		WithArgs().
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	id, err := r.SharedParamsRepo().SetSharedParams(ctx, params)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), id)
}
