package mysql

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/venus/venus-shared/testutil"
	venusTypes "github.com/filecoin-project/venus/venus-shared/types"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
	"github.com/stretchr/testify/assert"

	"github.com/filecoin-project/venus-messager/models/repo"
	"github.com/filecoin-project/venus-messager/testhelper"
)

func TestMessage(t *testing.T) {
	r, mock, sqlDB := setup(t)

	t.Run("mysql test expire message", wrapper(testExpireMessage, r, mock))
	t.Run("mysql test create message", wrapper(testCreateMessage, r, mock))
	t.Run("mysql test save message", wrapper(testSaveMessage, r, mock))
	t.Run("mysql test batch save message", wrapper(testBatchSaveMessage, r, mock))

	t.Run("mysql test get message by from and nonce", wrapper(testGetMessageByFromAndNonce, r, mock))
	t.Run("mysql test get message by from nonce and state", wrapper(testGetMessageByFromNonceAndState, r, mock))
	t.Run("mysql test get message by uid", wrapper(testGetMessageByUid, r, mock))
	t.Run("mysql test has message by uid", wrapper(testHasMessageByUid, r, mock))
	t.Run("mysql test get message state", wrapper(testGetMessageState, r, mock))
	t.Run("mysql test get message by cid", wrapper(testGetMessageByCid, r, mock))
	t.Run("mysql test get message by signed cid", wrapper(testGetMessageBySignedCid, r, mock))
	t.Run("mysql test get signed message by time", wrapper(testGetSignedMessageByTime, r, mock))
	t.Run("mysql test get signed message by height", wrapper(testGetSignedMessageByHeight, r, mock))
	t.Run("mysql test get signed message by height", wrapper(testGetSignedMessageFromFailedMsg, r, mock))

	t.Run("mysql test list message", wrapper(testListMessage, r, mock))
	t.Run("mysql test list message by from state", wrapper(testListMessageByFromState, r, mock))
	t.Run("mysql test list message by address", wrapper(testListMessageByAddress, r, mock))
	t.Run("mysql test list failed message", wrapper(testListFailedMessage, r, mock))
	t.Run("mysql test list blocked message", wrapper(testListBlockedMessage, r, mock))
	t.Run("mysql test list unchain message by address", wrapper(testListUnChainMessageByAddress, r, mock))
	t.Run("mysql test list failed message by address", wrapper(testListFilledMessageByAddress, r, mock))
	t.Run("mysql test list chain message by height", wrapper(testListChainMessageByHeight, r, mock))
	t.Run("mysql test list unfilled message", wrapper(testListUnFilledMessage, r, mock))
	t.Run("mysql test list signed message", wrapper(testListSignedMsgs, r, mock))
	t.Run("mysql test list filled message below nonce", wrapper(testListFilledMessageBelowNonce, r, mock))

	t.Run("mysql test update message info by cid", wrapper(testUpdateMessageInfoByCid, r, mock))
	t.Run("mysql test update message state by cid", wrapper(testUpdateMessageStateByCid, r, mock))
	t.Run("mysql test update message state by id", wrapper(testUpdateMessageStateByID, r, mock))
	t.Run("mysql test mark bad message", wrapper(testMarkBadMessage, r, mock))
	t.Run("mysql test update return value", wrapper(testUpdateReturnValue, r, mock))

	assert.NoError(t, closeDB(mock, sqlDB))
}

func testExpireMessage(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	msgs := testhelper.NewMessages(2)

	for i, msg := range msgs {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("UPDATE `messages` SET `state`=?,`updated_at`=? WHERE id = ?")).
			WithArgs(types.FailedMsg, anyTime{}, msg.ID).WillReturnResult(sqlmock.NewResult(int64(i+1), 1))
		mock.ExpectCommit()
	}

	assert.NoError(t, r.MessageRepo().ExpireMessage(msgs))
}

func testCreateMessage(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	msg := testhelper.NewMessage()

	mysqlMsg := fromMessage(msg)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(genInsertSQL(mysqlMsg))).
		WithArgs(getStructFieldValue(mysqlMsg)...).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	assert.NoError(t, r.MessageRepo().CreateMessage(msg))
}

func testBatchSaveMessage(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	msgs := testhelper.NewMessages(10)

	for _, msg := range msgs {
		mysqlMsg := fromMessage(msg)
		args := getStructFieldValue(mysqlMsg)
		id := args[0]
		tmpArgs := args[1:]
		tmpArgs = append(tmpArgs, id)

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(genUpdateSQL(mysqlMsg))).
			WithArgs(tmpArgs...).
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectCommit()

		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `messages` WHERE `id` = ? ORDER BY `messages`.`id` LIMIT 1")).
			WithArgs(mysqlMsg.ID).
			WillReturnError(gorm.ErrRecordNotFound)

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(genInsertSQL(mysqlMsg))).
			WithArgs(args...).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
	}

	assert.NoError(t, r.MessageRepo().BatchSaveMessage(msgs))
}

func testSaveMessage(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	msg := testhelper.NewMessage()

	mysqlMsg := fromMessage(msg)
	args := getStructFieldValue(mysqlMsg)
	id := args[0]
	tmpArgs := args[1:]
	tmpArgs = append(tmpArgs, id)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(genUpdateSQL(mysqlMsg))).
		WithArgs(tmpArgs...).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `messages` WHERE `id` = ? ORDER BY `messages`.`id` LIMIT 1")).
		WithArgs(mysqlMsg.ID).
		WillReturnError(gorm.ErrRecordNotFound)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(genInsertSQL(mysqlMsg))).
		WithArgs(args...).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	assert.NoError(t, r.MessageRepo().SaveMessage(msg))
}

func testGetMessageByFromAndNonce(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	from := testutil.AddressProvider()(t)
	nonce := uint64(10)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `messages` WHERE from_addr = ? and nonce = ? LIMIT 1")).
		WithArgs(from.String(), nonce).WillReturnRows(sqlmock.NewRows([]string{"from_addr", "nonce"}).AddRow(from.String(), nonce))

	res, err := r.MessageRepo().GetMessageByFromAndNonce(from, nonce)
	assert.NoError(t, err)
	assert.Equal(t, from, res.From)
	assert.Equal(t, nonce, res.Nonce)
}

func testGetMessageByFromNonceAndState(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	from := testutil.AddressProvider()(t)
	nonce := uint64(10)
	state := types.OnChainMsg

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `messages` WHERE from_addr = ? and nonce = ? and state = ? LIMIT 1")).
		WithArgs(from.String(), nonce, state).
		WillReturnRows(sqlmock.NewRows([]string{"from_addr", "nonce", "state"}).AddRow(from.String(), nonce, state))

	res, err := r.MessageRepo().GetMessageByFromNonceAndState(from, nonce, state)
	assert.NoError(t, err)
	assert.Equal(t, from, res.From)
	assert.Equal(t, nonce, res.Nonce)
	assert.Equal(t, state, res.State)
}

func testGetMessageByUid(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	uid := venusTypes.NewUUID().String()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `messages` WHERE id = ? LIMIT 1")).
		WithArgs(uid).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uid))

	res, err := r.MessageRepo().GetMessageByUid(uid)
	assert.NoError(t, err)
	assert.Equal(t, uid, res.ID)
}

func testHasMessageByUid(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	uid := venusTypes.NewUUID().String()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `messages` WHERE id = ?")).
		WithArgs(uid).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	has, err := r.MessageRepo().HasMessageByUid(uid)
	assert.NoError(t, err)
	assert.True(t, has)
}

func testGetMessageState(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	uid := venusTypes.NewUUID().String()
	state := types.FailedMsg

	mock.ExpectQuery(regexp.QuoteMeta("SELECT state FROM `messages` WHERE id = ?")).
		WithArgs(uid).WillReturnRows(sqlmock.NewRows([]string{"state"}).AddRow(state))

	state, err := r.MessageRepo().GetMessageState(uid)
	assert.NoError(t, err)
	assert.Equal(t, state, state)
}

func testGetMessageByCid(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	cid := testutil.CidProvider(32)(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `messages` WHERE unsigned_cid = ? LIMIT 1")).
		WithArgs(cid.String()).WillReturnRows(sqlmock.NewRows([]string{"unsigned_cid"}).AddRow(cid.String()))

	res, err := r.MessageRepo().GetMessageByCid(cid)
	assert.NoError(t, err)
	assert.Equal(t, cid, *res.UnsignedCid)
}

func testGetMessageBySignedCid(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	cid := testutil.CidProvider(32)(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `messages` WHERE signed_cid = ? LIMIT 1")).
		WithArgs(cid.String()).WillReturnRows(sqlmock.NewRows([]string{"signed_cid"}).AddRow(cid.String()))

	res, err := r.MessageRepo().GetMessageBySignedCid(cid)
	assert.NoError(t, err)
	assert.Equal(t, cid, *res.SignedCid)
}

func testGetSignedMessageByTime(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	now := time.Now()
	afterTimes := []time.Time{now.Add(1 * time.Second), now.Add(1 * time.Hour)}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `messages` WHERE created_at >= ? and signed_data is not null")).
		WithArgs(now).WillReturnRows(sqlmock.NewRows([]string{"created_at"}).AddRow(afterTimes[0]).AddRow(afterTimes[1]))

	res, err := r.MessageRepo().GetSignedMessageByTime(now)
	assert.NoError(t, err)
	for _, msg := range res {
		assert.True(t, now.Before(msg.CreatedAt))
	}
}

func testGetSignedMessageByHeight(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	height := abi.ChainEpoch(1000)
	bigger := []abi.ChainEpoch{100000, 1000001}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `messages` WHERE height >= ? and signed_data is not null")).
		WithArgs(height).WillReturnRows(sqlmock.NewRows([]string{"height"}).AddRow(bigger[0]).AddRow(bigger[1]))

	res, err := r.MessageRepo().GetSignedMessageByHeight(height)
	assert.NoError(t, err)
	for _, msg := range res {
		assert.Less(t, height, msg.Height)
	}
}

func testGetSignedMessageFromFailedMsg(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	addr := testutil.AddressProvider()(t)
	state := types.FailedMsg

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `messages` WHERE state = ? and from_addr = ? and signed_data is not null")).
		WithArgs(state, addr.String()).
		WillReturnRows(sqlmock.NewRows([]string{"state", "from_addr"}).AddRow(state, addr.String()).AddRow(state, addr.String()))

	res, err := r.MessageRepo().GetSignedMessageFromFailedMsg(addr)
	assert.NoError(t, err)
	for _, msg := range res {
		assert.Equal(t, state, msg.State)
		assert.Equal(t, addr, msg.From)
	}
}

func testListMessage(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	ids := []string{"msg1", "msg2", "msg3"}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `messages`")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ids[0]).AddRow(ids[1]).AddRow(ids[2]))

	res, err := r.MessageRepo().ListMessage()
	assert.NoError(t, err)
	checkMsgWithIDs(t, res, ids)
}

func testListMessageByFromState(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	ids := []string{"msg1", "msg2", "msg3"}
	from := testutil.AddressProvider()(t)
	state := types.OnChainMsg
	isAsc := false
	pageIndex := 1
	pageSize := 3

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `messages` WHERE from_addr = ? AND state = ? ORDER BY created_at DESC LIMIT 3")).
		WithArgs(from.String(), state).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ids[0]).AddRow(ids[1]).AddRow(ids[2]))

	// from is empty
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `messages` WHERE state = ? ORDER BY created_at DESC LIMIT 3")).
		WithArgs(state).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ids[0]).AddRow(ids[1]).AddRow(ids[2]))

	// isAsc = true
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `messages` WHERE from_addr = ? AND state = ? ORDER BY created_at ASC LIMIT 3")).
		WithArgs(from.String(), state).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ids[0]).AddRow(ids[1]).AddRow(ids[2]))

	// pageIndex = 2 pageSize = 2
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `messages` WHERE from_addr = ? AND state = ? ORDER BY created_at DESC LIMIT 2 OFFSET 2")).
		WithArgs(from.String(), state).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ids[2]))

	res, err := r.MessageRepo().ListMessageByFromState(from, state, isAsc, pageIndex, pageSize)
	assert.NoError(t, err)
	checkMsgWithIDs(t, res, ids)

	res, err = r.MessageRepo().ListMessageByFromState(address.Undef, state, isAsc, pageIndex, pageSize)
	assert.NoError(t, err)
	checkMsgWithIDs(t, res, ids)

	res, err = r.MessageRepo().ListMessageByFromState(from, state, true, pageIndex, pageSize)
	assert.NoError(t, err)
	checkMsgWithIDs(t, res, ids)

	res, err = r.MessageRepo().ListMessageByFromState(from, state, isAsc, 2, 2)
	assert.NoError(t, err)
	checkMsgWithIDs(t, res, []string{ids[2]})
}

func testListMessageByAddress(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	ids := []string{"msg1", "msg2"}
	addr := testutil.AddressProvider()(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `messages` WHERE from_addr=?")).
		WithArgs(addr.String()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ids[0]).AddRow(ids[1]))

	res, err := r.MessageRepo().ListMessageByAddress(addr)
	assert.NoError(t, err)
	checkMsgWithIDs(t, res, ids)
}

func testListFailedMessage(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	ids := []string{"msg1", "msg2"}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `messages` WHERE state = ? AND receipt_return_value is not null ORDER BY created_at")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ids[0]).AddRow(ids[1]))

	res, err := r.MessageRepo().ListFailedMessage()
	assert.NoError(t, err)
	checkMsgWithIDs(t, res, ids)
}

func testListBlockedMessage(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	ids := []string{"msg1", "msg2"}
	from := testutil.AddressProvider()(t)
	blocked := time.Second * 3

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `messages` WHERE from_addr = ? AND state = ? AND created_at < ? ORDER BY created_at")).
		WithArgs(from.String(), types.FillMsg, anyTime{}).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ids[0]).AddRow(ids[1]))

	res, err := r.MessageRepo().ListBlockedMessage(from, blocked)
	assert.NoError(t, err)
	checkMsgWithIDs(t, res, ids)
}

func testListUnChainMessageByAddress(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	ids := []string{"msg1", "msg2", "msg3", "msg4"}
	from := testutil.AddressProvider()(t)
	topN := 3

	mock.ExpectQuery(regexp.QuoteMeta(fmt.Sprintf("SELECT * FROM `messages` WHERE from_addr=? AND state=? ORDER BY created_at LIMIT %d", topN))).
		WithArgs(from.String(), types.UnFillMsg).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ids[0]).AddRow(ids[1]).AddRow(ids[2]))

	zero := 0
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `messages` WHERE from_addr=? AND state=? ORDER BY created_at")).
		WithArgs(from.String(), types.UnFillMsg).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ids[0]).AddRow(ids[1]).AddRow(ids[2]).AddRow(ids[3]))

	res, err := r.MessageRepo().ListUnChainMessageByAddress(from, topN)
	assert.NoError(t, err)
	checkMsgWithIDs(t, res, ids[:3])

	res, err = r.MessageRepo().ListUnChainMessageByAddress(from, zero)
	assert.NoError(t, err)
	checkMsgWithIDs(t, res, ids)
}

func testListFilledMessageByAddress(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	ids := []string{"msg1", "msg2"}
	from := testutil.AddressProvider()(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `messages` WHERE from_addr=? AND state=?")).
		WithArgs(from.String(), types.FillMsg).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ids[0]).AddRow(ids[1]))

	res, err := r.MessageRepo().ListFilledMessageByAddress(from)
	assert.NoError(t, err)
	checkMsgWithIDs(t, res, ids)
}

func testListChainMessageByHeight(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	ids := []string{"msg1", "msg2"}
	height := abi.ChainEpoch(100)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `messages` WHERE height=? AND state=?")).
		WithArgs(height, types.OnChainMsg).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ids[0]).AddRow(ids[1]))

	res, err := r.MessageRepo().ListChainMessageByHeight(height)
	assert.NoError(t, err)
	checkMsgWithIDs(t, res, ids)
}

func testListUnFilledMessage(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	ids := []string{"msg1", "msg2"}
	addr := testutil.AddressProvider()(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `messages` WHERE from_addr = ? AND state = ?")).
		WithArgs(addr.String(), types.UnFillMsg).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ids[0]).AddRow(ids[1]))

	res, err := r.MessageRepo().ListUnFilledMessage(addr)
	assert.NoError(t, err)
	checkMsgWithIDs(t, res, ids)
}

func testListSignedMsgs(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	ids := []string{"msg1", "msg2"}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `messages` WHERE height=0 and signed_data is not null")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ids[0]).AddRow(ids[1]))

	res, err := r.MessageRepo().ListSignedMsgs()
	assert.NoError(t, err)
	checkMsgWithIDs(t, res, ids)
}

func testListFilledMessageBelowNonce(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	ids := []string{"msg1", "msg2"}
	addr := testutil.AddressProvider()(t)
	nonce := uint64(100)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `messages` WHERE from_addr=? AND state=? AND nonce<?")).
		WithArgs(addr.String(), types.FillMsg, nonce).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ids[0]).AddRow(ids[1]))

	res, err := r.MessageRepo().ListFilledMessageBelowNonce(addr, nonce)
	assert.NoError(t, err)
	checkMsgWithIDs(t, res, ids)
}

func testUpdateMessageInfoByCid(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	cid := testutil.CidProvider(32)(t)
	receipt := &venusTypes.MessageReceipt{
		ExitCode: -1,
		GasUsed:  100,
		Return:   []byte("return"),
	}
	height := abi.ChainEpoch(1000)
	state := types.OnChainMsg
	key := venusTypes.NewTipSetKey(testutil.CidProvider(32)(t))

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `messages` SET `height`=?,`receipt_exit_code`=?,`receipt_gas_used`=?,"+
		"`receipt_return_value`=?,`state`=?,`tipset_key`=?,`updated_at`=? WHERE unsigned_cid = ?")).
		WithArgs(height, receipt.ExitCode, receipt.GasUsed, receipt.Return, state, key.String(), anyTime{}, cid.String()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	assert.NoError(t, r.MessageRepo().UpdateMessageInfoByCid(cid.String(), receipt, height, state, key))
}

func testUpdateMessageStateByCid(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	cid := testutil.CidProvider(32)(t)
	state := types.OnChainMsg

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `messages` SET `state`=?,`updated_at`=? WHERE unsigned_cid = ?")).
		WithArgs(state, anyTime{}, cid.String()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	assert.NoError(t, r.MessageRepo().UpdateMessageStateByCid(cid.String(), state))
}

func testUpdateMessageStateByID(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	id := venusTypes.NewUUID().String()
	state := types.OnChainMsg

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `messages` SET `state`=?,`updated_at`=? WHERE id = ?")).
		WithArgs(state, anyTime{}, id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	assert.NoError(t, r.MessageRepo().UpdateMessageStateByID(id, state))
}

func testMarkBadMessage(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	id := venusTypes.NewUUID().String()
	state := types.FailedMsg

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `messages` SET `state`=?,`updated_at`=? WHERE id = ?")).
		WithArgs(state, anyTime{}, id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	assert.NoError(t, r.MessageRepo().MarkBadMessage(id))
}

func testUpdateReturnValue(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	id := venusTypes.NewUUID().String()
	returnVal := "val"

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `messages` SET `receipt_return_value`=?,`updated_at`=? WHERE id = ?")).
		WithArgs(returnVal, anyTime{}, id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	assert.NoError(t, r.MessageRepo().UpdateReturnValue(id, returnVal))
}

func checkMsgWithIDs(t *testing.T, msgs []*types.Message, ids []string) {
	assert.Equal(t, len(msgs), len(ids))
	for i, msg := range msgs {
		assert.Equal(t, ids[i], msg.ID)
	}
}
