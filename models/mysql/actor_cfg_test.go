package mysql

import (
	"context"
	"regexp"
	"testing"

	"github.com/ipfs/go-cid"

	"github.com/ipfs-force-community/sophon-messager/models/mtypes"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/filecoin-project/venus/venus-shared/testutil"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
	"github.com/ipfs-force-community/sophon-messager/models/repo"
	"gorm.io/gorm"

	"github.com/stretchr/testify/assert"
)

func Test_mysqlActorCfgRepo_SaveActorCfg(t *testing.T) {
	r, mock, sqlDB := setup(t)
	t.Run("mysql test save actor config", wrapper(testSaveActorCfg, r, mock))
	t.Run("mysql test get actor config by id", wrapper(testGetActorTypeById, r, mock))
	t.Run("mysql test has actor config", wrapper(testHasActorCfg, r, mock))
	t.Run("mysql test get actor config by method type", wrapper(testGetActorTypeByMethodType, r, mock))
	t.Run("mysql test list actor config by id", wrapper(testListActorType, r, mock))
	t.Run("mysql test delete actor config by method types", wrapper(testDeleteActorCfgByMethodType, r, mock))
	t.Run("mysql test delete actor config by id", wrapper(testDeleteActorCfgById, r, mock))
	t.Run("mysql test update actor config", wrapper(testUpdateSelectSpec, r, mock))
	assert.NoError(t, closeDB(mock, sqlDB))
}

func testSaveActorCfg(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	ctx := context.Background()

	var actorCfg types.ActorCfg
	testutil.Provide(t, &actorCfg)

	mysqlActorCfg := fromActorCfg(&actorCfg)
	updateSQL, updateArgs := genUpdateSQL(mysqlActorCfg, false)
	updateArgs = append(updateArgs, mysqlActorCfg.ID)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(updateSQL)).
		WithArgs(updateArgs...).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `actor_cfg` WHERE `id` = ? ORDER BY `actor_cfg`.`id` LIMIT 1")).
		WithArgs(mysqlActorCfg.ID).
		WillReturnError(gorm.ErrRecordNotFound)

	insertSql, insertArgs := genInsertSQL(mysqlActorCfg)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(insertSql)).
		WithArgs(insertArgs...).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	assert.Nil(t, r.ActorCfgRepo().SaveActorCfg(ctx, &actorCfg))
}

func testGetActorTypeById(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	ctx := context.Background()
	var actorCfg types.ActorCfg
	testutil.Provide(t, &actorCfg)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `actor_cfg` WHERE id = ? LIMIT 1")).
		WithArgs(actorCfg.ID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	_, err := r.ActorCfgRepo().GetActorCfgByID(ctx, actorCfg.ID)
	assert.Equal(t, repo.ErrRecordNotFound, err)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `actor_cfg` WHERE id = ? LIMIT 1")).
		WithArgs(actorCfg.ID).
		WillReturnRows(genSelectResult(fromActorCfg(&actorCfg)))

	actorCfgR, err := r.ActorCfgRepo().GetActorCfgByID(ctx, actorCfg.ID)
	assert.NoError(t, err)
	assert.Equal(t, actorCfg, *actorCfgR)
}

func testHasActorCfg(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	ctx := context.Background()
	var actorCfg types.ActorCfg
	testutil.Provide(t, &actorCfg)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `actor_cfg` WHERE code = ? and method = ?")).
		WithArgs(mtypes.NewDBCid(actorCfg.Code), actorCfg.Method).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	has, err := r.ActorCfgRepo().HasActorCfg(ctx, &types.MethodType{
		Code:   actorCfg.Code,
		Method: actorCfg.Method,
	})
	assert.NoError(t, err)
	assert.True(t, has)
}

func testGetActorTypeByMethodType(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	ctx := context.Background()
	var actorCfg types.ActorCfg
	testutil.Provide(t, &actorCfg)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `actor_cfg` WHERE code = ? and method = ? LIMIT 1")).
		WithArgs(mtypes.NewDBCid(actorCfg.Code), actorCfg.Method).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	_, err := r.ActorCfgRepo().GetActorCfgByMethodType(ctx, &types.MethodType{
		Code:   actorCfg.Code,
		Method: actorCfg.Method,
	})
	assert.Equal(t, repo.ErrRecordNotFound, err)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `actor_cfg` WHERE code = ? and method = ? LIMIT 1")).
		WithArgs(mtypes.NewDBCid(actorCfg.Code), actorCfg.Method).
		WillReturnRows(genSelectResult(fromActorCfg(&actorCfg)))

	actorCfgR, err := r.ActorCfgRepo().GetActorCfgByMethodType(ctx, &types.MethodType{
		Code:   actorCfg.Code,
		Method: actorCfg.Method,
	})
	assert.NoError(t, err)
	assert.Equal(t, actorCfg, *actorCfgR)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `actor_cfg` WHERE code = ? and method = ? LIMIT 1")).
		WithArgs(mtypes.UndefDBCid, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	_, err = r.ActorCfgRepo().GetActorCfgByMethodType(ctx, &types.MethodType{
		Code:   cid.Undef,
		Method: 1,
	})
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func testListActorType(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	ctx := context.Background()
	actorCfgs := make([]*types.ActorCfg, 10)
	testutil.Provide(t, &actorCfgs)

	actorMysqlCfgs := make([]*mysqlActorCfg, len(actorCfgs))
	for index, actorCfg := range actorCfgs {
		actorMysqlCfgs[index] = fromActorCfg(actorCfg)
	}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `actor_cfg`")).
		WithArgs().
		WillReturnRows(genSelectResult(actorMysqlCfgs))

	val, err := r.ActorCfgRepo().ListActorCfg(ctx)
	assert.NoError(t, err)
	assertActorCfgArrValue(t, actorCfgs, val)
}

func assertActorCfgValue(t *testing.T, expectVal, actualVal *types.ActorCfg) {
	assert.Equal(t, expectVal.ID, actualVal.ID)
	assert.Equal(t, expectVal.ActorVersion, actualVal.ActorVersion)
	assert.Equal(t, expectVal.MethodType, actualVal.MethodType)
	assert.Equal(t, expectVal.FeeSpec, actualVal.FeeSpec)
}

func assertActorCfgArrValue(t *testing.T, expectVal, actualVal []*types.ActorCfg) {
	assert.Equal(t, len(expectVal), len(actualVal))

	for index, val := range expectVal {
		assertActorCfgValue(t, val, actualVal[index])
	}
}

func testDeleteActorCfgByMethodType(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	ctx := context.Background()
	t.Run("correct ", func(t *testing.T) {
		var actorCfg types.ActorCfg
		testutil.Provide(t, &actorCfg)

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `actor_cfg` WHERE code = ? and method = ?")).
			WithArgs(mtypes.NewDBCid(actorCfg.Code), actorCfg.Method).
			WillReturnResult(driverResult{0, 1})
		mock.ExpectCommit()

		err := r.ActorCfgRepo().DelActorCfgByMethodType(ctx, &types.MethodType{
			Code:   actorCfg.Code,
			Method: actorCfg.Method,
		})
		assert.NoError(t, err)
	})

	t.Run("code cid is undefined", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `actor_cfg` WHERE code = ? and method = ?")).
			WithArgs(mtypes.UndefDBCid, 0).
			WillReturnResult(driverResult{0, 1})
		mock.ExpectCommit()

		err := r.ActorCfgRepo().DelActorCfgByMethodType(ctx, &types.MethodType{
			Code:   cid.Undef,
			Method: 0,
		})
		assert.NoError(t, err)
	})

}

func testDeleteActorCfgById(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	ctx := context.Background()
	var actorCfg types.ActorCfg
	testutil.Provide(t, &actorCfg)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `actor_cfg` WHERE id = ?")).
		WithArgs(actorCfg.ID).
		WillReturnResult(driverResult{0, 1})
	mock.ExpectCommit()

	err := r.ActorCfgRepo().DelActorCfgById(ctx, actorCfg.ID)
	assert.NoError(t, err)
}

func testUpdateSelectSpec(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	ctx := context.Background()
	var actorCfg types.ActorCfg
	testutil.Provide(t, &actorCfg)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `actor_cfg` SET `base_fee`=?,`gas_fee_cap`=?,`gas_over_estimation`=?,`gas_over_premium`=?,`max_fee`=?,`updated_at`=? WHERE id = ?")).
		WithArgs(actorCfg.BaseFee.String(), actorCfg.GasFeeCap.String(), actorCfg.GasOverEstimation, actorCfg.GasOverPremium, actorCfg.MaxFee.String(), anyTime{}, actorCfg.ID).
		WillReturnResult(driverResult{0, 1})
	mock.ExpectCommit()

	err := r.ActorCfgRepo().UpdateSelectSpecById(ctx, actorCfg.ID,
		&types.ChangeGasSpecParams{
			GasOverEstimation: &actorCfg.GasOverEstimation,
			MaxFee:            actorCfg.MaxFee,
			GasFeeCap:         actorCfg.GasFeeCap,
			GasOverPremium:    &actorCfg.GasOverPremium,
			BaseFee:           actorCfg.BaseFee,
		})
	assert.NoError(t, err)

	//only update select num
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `actor_cfg` SET `gas_over_premium`=?,`updated_at`=? WHERE id = ?")).
		WithArgs(actorCfg.GasOverPremium, anyTime{}, actorCfg.ID).
		WillReturnResult(driverResult{0, 1})
	mock.ExpectCommit()

	err = r.ActorCfgRepo().UpdateSelectSpecById(ctx, actorCfg.ID,
		&types.ChangeGasSpecParams{
			GasOverPremium: &actorCfg.GasOverPremium,
		})
	assert.NoError(t, err)

	//only update max fee
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `actor_cfg` SET `max_fee`=?,`updated_at`=? WHERE id = ?")).
		WithArgs(actorCfg.MaxFee.String(), anyTime{}, actorCfg.ID).
		WillReturnResult(driverResult{0, 1})
	mock.ExpectCommit()

	err = r.ActorCfgRepo().UpdateSelectSpecById(ctx, actorCfg.ID,
		&types.ChangeGasSpecParams{
			MaxFee: actorCfg.MaxFee,
		})
	assert.NoError(t, err)
}
