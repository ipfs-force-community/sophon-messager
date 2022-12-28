package mysql

import (
	"context"
	"math/rand"
	"regexp"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/venus/venus-shared/testutil"
	venustypes "github.com/filecoin-project/venus/venus-shared/types"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
	"github.com/stretchr/testify/assert"

	"github.com/filecoin-project/venus-messager/models/repo"
)

func TestAddress(t *testing.T) {
	r, mock, sqlDB := setup(t)

	t.Run("mysql test save address", wrapper(testSaveAddress, r, mock))
	t.Run("mysql test get address", wrapper(testGetAddress, r, mock))
	t.Run("mysql test get address by id", wrapper(testGetAddressByID, r, mock))
	t.Run("mysql test get one record", wrapper(testGetOneRecord, r, mock))
	t.Run("mysql test has address", wrapper(testHasAddress, r, mock))
	t.Run("mysql test list address", wrapper(testListAddress, r, mock))
	t.Run("mysql test list active address", wrapper(testListActiveAddress, r, mock))
	t.Run("mysql test delete address", wrapper(testDelAddress, r, mock))
	t.Run("mysql test update nonce", wrapper(testUpdateNonce, r, mock))
	t.Run("mysql test update state", wrapper(testUpdateState, r, mock))
	t.Run("mysql test update select message num", wrapper(testUpdateSelectMsgNum, r, mock))
	t.Run("mysql test update fee params", wrapper(testUpdateFeeParams, r, mock))

	assert.NoError(t, closeDB(mock, sqlDB))
}

func testSaveAddress(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	ctx := context.Background()

	addrInfo, err := newAddressInfo(t)
	assert.NoError(t, err)

	mysqlAddr := fromAddress(addrInfo)
	updateSQL, updateArgs := genUpdateSQL(mysqlAddr, false)
	updateArgs = append(updateArgs, mysqlAddr.ID)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(updateSQL)).
		WithArgs(updateArgs...).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `addresses` WHERE `id` = ? ORDER BY `addresses`.`id` LIMIT 1")).
		WithArgs(mysqlAddr.ID).
		WillReturnError(gorm.ErrRecordNotFound)

	insertSql, insertArgs := genInsertSQL(mysqlAddr)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(insertSql)).
		WithArgs(insertArgs...).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	assert.Nil(t, r.AddressRepo().SaveAddress(ctx, addrInfo))
}

func testGetAddress(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	ctx := context.Background()
	addr := testutil.AddressProvider()(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `addresses` WHERE addr = ? and is_deleted = ? LIMIT 1")).
		WithArgs(addr.String(), repo.NotDeleted).
		WillReturnRows(sqlmock.NewRows([]string{"addr"}).AddRow(addr.String()))

	res, err := r.AddressRepo().GetAddress(ctx, addr)
	assert.NoError(t, err)
	assert.Equal(t, addr, res.Addr)
}

func testGetAddressByID(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	ctx := context.Background()
	id := venustypes.NewUUID()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `addresses` WHERE id = ? and is_deleted = ? ORDER BY `addresses`.`id` LIMIT 1")).
		WithArgs(id, repo.NotDeleted).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))

	res, err := r.AddressRepo().GetAddressByID(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, id, res.ID)
}

func testGetOneRecord(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	ctx := context.Background()
	addr := testutil.AddressProvider()(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `addresses` WHERE addr = ? LIMIT 1")).
		WithArgs(addr.String()).
		WillReturnRows(sqlmock.NewRows([]string{"addr"}).AddRow(addr.String()))

	res, err := r.AddressRepo().GetOneRecord(ctx, addr)
	assert.NoError(t, err)
	assert.Equal(t, addr, res.Addr)
}

func testHasAddress(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	ctx := context.Background()
	addr := testutil.AddressProvider()(t)

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT count(*) FROM `addresses` WHERE addr = ? and is_deleted = ?")).
		WithArgs(addr.String(), repo.NotDeleted).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	has, err := r.AddressRepo().HasAddress(ctx, addr)
	assert.NoError(t, err)
	assert.True(t, has)
}

func testListAddress(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT * FROM `addresses` WHERE is_deleted = ?")).
		WithArgs(repo.NotDeleted).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(venustypes.NewUUID()).AddRow(venustypes.NewUUID()))

	res, err := r.AddressRepo().ListAddress(ctx)
	assert.NoError(t, err)
	assert.Len(t, res, 2)
}

func testListActiveAddress(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT * FROM `addresses` WHERE is_deleted = ? and state = ?")).
		WithArgs(repo.NotDeleted, types.AddressStateAlive).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(venustypes.NewUUID()).AddRow(venustypes.NewUUID()))

	res, err := r.AddressRepo().ListActiveAddress(ctx)
	assert.NoError(t, err)
	assert.Len(t, res, 2)
}

func testDelAddress(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	ctx := context.Background()
	addr := testutil.AddressProvider()(t)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(
		"UPDATE `addresses` SET `is_deleted`=?,`state`=?,`updated_at`=? WHERE addr = ? and is_deleted = ?")).
		WithArgs(repo.Deleted, types.AddressStateRemoved, anyTime{}, addr.String(), repo.NotDeleted).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := r.AddressRepo().DelAddress(ctx, addr)
	assert.NoError(t, err)
}

func testUpdateNonce(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	ctx := context.Background()
	addr := testutil.AddressProvider()(t)
	nonce := uint64(10)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(
		"UPDATE `addresses` SET `nonce`=?,`updated_at`=? WHERE addr = ? and is_deleted = ?")).
		WithArgs(nonce, anyTime{}, addr.String(), repo.NotDeleted).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := r.AddressRepo().UpdateNonce(ctx, addr, nonce)
	assert.NoError(t, err)
}

func testUpdateState(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	ctx := context.Background()
	addr := testutil.AddressProvider()(t)
	states := []types.AddressState{
		types.AddressState(0),
		types.AddressStateAlive,
		types.AddressStateRemoving,
		types.AddressStateRemoved,
		types.AddressStateForbbiden,
	}

	for _, state := range states {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(
			"UPDATE `addresses` SET `state`=?,`updated_at`=? WHERE addr = ? and is_deleted = ?")).
			WithArgs(state, anyTime{}, addr.String(), repo.NotDeleted).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := r.AddressRepo().UpdateState(ctx, addr, state)
		assert.NoError(t, err)
	}
}

func testUpdateSelectMsgNum(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	ctx := context.Background()
	addr := testutil.AddressProvider()(t)
	selectMsgNum := uint64(10)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(
		"UPDATE `addresses` SET `sel_msg_num`=?,`updated_at`=? WHERE addr = ? and is_deleted = ?")).
		WithArgs(selectMsgNum, anyTime{}, addr.String(), repo.NotDeleted).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := r.AddressRepo().UpdateSelectMsgNum(ctx, addr, selectMsgNum)
	assert.NoError(t, err)
}

func testUpdateFeeParams(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	ctx := context.Background()
	addr := testutil.AddressProvider()(t)
	gasOverEstimation := 1.25
	gasOverPremium := 4.0
	maxFee := big.NewInt(100)
	gasFeeCap := big.NewInt(1000)
	baseFee := big.NewInt(1000)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(
		"UPDATE `addresses` SET `base_fee`=?,`gas_fee_cap`=?,`gas_over_estimation`=?,`gas_over_premium`=?,`max_fee`=?,`updated_at`=? WHERE addr = ? and is_deleted = ?")).
		WithArgs(baseFee.String(), gasFeeCap.String(), gasOverEstimation, gasOverPremium, maxFee.String(), anyTime{}, addr.String(), repo.NotDeleted).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// gasFeeCap is nil
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(
		"UPDATE `addresses` SET `base_fee`=?,`gas_over_estimation`=?,`gas_over_premium`=?,`max_fee`=?,`updated_at`=? WHERE addr = ? and is_deleted = ?")).
		WithArgs(baseFee.String(), gasOverEstimation, gasOverPremium, maxFee.String(), anyTime{}, addr.String(), repo.NotDeleted).
		WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectCommit()

	// gasOverEstimation is 0
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(
		"UPDATE `addresses` SET `base_fee`=?,`gas_fee_cap`=?,`gas_over_premium`=?,`max_fee`=?,`updated_at`=? WHERE addr = ? and is_deleted = ?")).
		WithArgs(baseFee.String(), gasFeeCap.String(), gasOverPremium, maxFee.String(), anyTime{}, addr.String(), repo.NotDeleted).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// gasOverPremium is 0
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(
		"UPDATE `addresses` SET `base_fee`=?,`gas_fee_cap`=?,`gas_over_estimation`=?,`max_fee`=?,`updated_at`=? WHERE addr = ? and is_deleted = ?")).
		WithArgs(baseFee.String(), gasFeeCap.String(), gasOverEstimation, maxFee.String(), anyTime{}, addr.String(), repo.NotDeleted).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// maxFee is nil
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(
		"UPDATE `addresses` SET `base_fee`=?,`gas_fee_cap`=?,`gas_over_estimation`=?,`gas_over_premium`=?,`updated_at`=? WHERE addr = ? and is_deleted = ?")).
		WithArgs(baseFee.String(), gasFeeCap.String(), gasOverEstimation, gasOverPremium, anyTime{}, addr.String(), repo.NotDeleted).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// baseFee is nil
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(
		"UPDATE `addresses` SET `gas_fee_cap`=?,`gas_over_estimation`=?,`gas_over_premium`=?,`max_fee`=?,`updated_at`=? WHERE addr = ? and is_deleted = ?")).
		WithArgs(gasFeeCap.String(), gasOverEstimation, gasOverPremium, maxFee.String(), anyTime{}, addr.String(), repo.NotDeleted).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	assert.NoError(t, r.AddressRepo().UpdateFeeParams(ctx, addr, gasOverEstimation, gasOverPremium, maxFee, gasFeeCap, baseFee))
	assert.NoError(t, r.AddressRepo().UpdateFeeParams(ctx, addr, gasOverEstimation, gasOverPremium, maxFee, big.Int{}, baseFee))
	assert.NoError(t, r.AddressRepo().UpdateFeeParams(ctx, addr, 0, gasOverPremium, maxFee, gasFeeCap, baseFee))
	assert.NoError(t, r.AddressRepo().UpdateFeeParams(ctx, addr, gasOverEstimation, 0, maxFee, gasFeeCap, baseFee))
	assert.NoError(t, r.AddressRepo().UpdateFeeParams(ctx, addr, gasOverEstimation, gasOverPremium, big.Int{}, gasFeeCap, baseFee))
	assert.NoError(t, r.AddressRepo().UpdateFeeParams(ctx, addr, gasOverEstimation, gasOverPremium, maxFee, gasFeeCap, big.Int{}))
}

func newAddressInfo(t *testing.T) (*types.Address, error) {
	randNum := rand.Int63n(1000)
	return &types.Address{
		ID:     venustypes.NewUUID(),
		Addr:   testutil.AddressProvider()(t),
		Nonce:  uint64(randNum),
		Weight: randNum,
		// Any zero value like 0, '', false won’t be saved into the database for those fields defined default value,
		// you might want to use pointer type or Scanner/Valuer to avoid this.
		State: types.AddressState(rand.Intn(4) + 1),
		SelectSpec: types.SelectSpec{
			GasOverEstimation: float64(randNum),
			GasOverPremium:    float64(randNum),
			MaxFee:            big.NewInt(randNum),
			GasFeeCap:         big.NewInt(randNum),
			BaseFee:           big.NewInt(randNum),

			SelMsgNum: uint64(randNum),
		},
		IsDeleted: repo.NotDeleted,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}
