package mysql

import (
	"math/rand"
	"regexp"
	"testing"

	"gorm.io/gorm"

	"github.com/DATA-DOG/go-sqlmock"
	venustypes "github.com/filecoin-project/venus/venus-shared/types"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
	"github.com/stretchr/testify/assert"

	"github.com/filecoin-project/venus-messager/models/repo"
)

func TestNode(t *testing.T) {
	r, mock, sqlDB := setup(t)

	t.Run("mysql test create node", wrapper(testCreateNode, r, mock))
	t.Run("mysql test save node", wrapper(testSaveNode, r, mock))
	t.Run("mysql test get node", wrapper(testGetNode, r, mock))
	t.Run("mysql test has node", wrapper(testHasNode, r, mock))
	t.Run("mysql test list node", wrapper(testListNode, r, mock))
	t.Run("mysql test delete node", wrapper(testDelNode, r, mock))

	assert.NoError(t, closeDB(mock, sqlDB))
}

func testCreateNode(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	node := &types.Node{
		ID:    venustypes.NewUUID(),
		Name:  venustypes.NewUUID().String(),
		URL:   venustypes.NewUUID().String(),
		Token: venustypes.NewUUID().String(),
		Type:  types.NodeType(rand.Intn(2)),
	}

	mysqlNode := fromNode(node)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(genInsertSQL(mysqlNode))).
		WithArgs(getStructFieldValue(mysqlNode)...).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	assert.NoError(t, r.NodeRepo().CreateNode(node))
}

func testSaveNode(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	node := &types.Node{
		ID:    venustypes.NewUUID(),
		Name:  venustypes.NewUUID().String(),
		URL:   venustypes.NewUUID().String(),
		Token: venustypes.NewUUID().String(),
		Type:  types.NodeType(rand.Intn(2)),
	}

	mysqlNode := fromNode(node)
	args := getStructFieldValue(mysqlNode)
	id := args[0]
	tmpArgs := args[1:]
	tmpArgs = append(tmpArgs, id)
	updateSQL := genUpdateSQL(mysqlNode)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(updateSQL)).
		WithArgs(tmpArgs...).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `nodes` WHERE `id` = ? ORDER BY `nodes`.`id` LIMIT 1")).
		WithArgs(mysqlNode.ID).
		WillReturnError(gorm.ErrRecordNotFound)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(genInsertSQL(mysqlNode))).
		WithArgs(args...).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	assert.Nil(t, r.NodeRepo().SaveNode(node))
}

func testGetNode(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	nodeName := "node1"

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `nodes` WHERE name = ? and is_deleted = ? LIMIT 1")).
		WithArgs(nodeName, repo.NotDeleted).
		WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow(nodeName))

	res, err := r.NodeRepo().GetNode(nodeName)
	assert.NoError(t, err)
	assert.Equal(t, nodeName, res.Name)
}

func testHasNode(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	nodeName := "node1"

	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `nodes` WHERE name = ? and is_deleted = ?")).
		WithArgs(nodeName, repo.NotDeleted).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	has, err := r.NodeRepo().HasNode(nodeName)
	assert.NoError(t, err)
	assert.True(t, has)
}

func testListNode(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `nodes` WHERE is_deleted = ?")).
		WithArgs(repo.NotDeleted).
		WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("node1").AddRow("node2"))

	list, err := r.NodeRepo().ListNode()
	assert.NoError(t, err)
	assert.Len(t, list, 2)
}

func testDelNode(t *testing.T, r repo.Repo, mock sqlmock.Sqlmock) {
	id := venustypes.NewUUID()
	nodeName := "node1"

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `nodes` WHERE name = ? and is_deleted = ? LIMIT 1")).
		WithArgs(nodeName, repo.NotDeleted).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(id, nodeName))

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `nodes` SET `name`=?,`url`=?,`token`=?,`node_type`=?,`is_deleted`=?,`created_at`=?,`updated_at`=? WHERE `id` = ?")).
		WithArgs(nodeName, "", "", 0, repo.Deleted, anyTime{}, anyTime{}, id).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	assert.NoError(t, r.NodeRepo().DelNode(nodeName))
}
