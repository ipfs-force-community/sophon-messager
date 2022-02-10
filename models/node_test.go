package models

import (
	"testing"

	"github.com/filecoin-project/venus-messager/models/repo"
	venustypes "github.com/filecoin-project/venus/venus-shared/types"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
	"github.com/stretchr/testify/assert"
)

func randNode() *types.Node {
	return &types.Node{
		ID:    venustypes.NewUUID(),
		Name:  venustypes.NewUUID().String(),
		URL:   venustypes.NewUUID().String(),
		Token: venustypes.NewUUID().String(),
		Type:  0,
	}
}

func TestNode(t *testing.T) {
	sqliteRepo, mysqlRepo := setupRepo(t)

	nodeRepoTest := func(t *testing.T, nodeRepo repo.NodeRepo) {
		node := randNode()
		node2 := randNode()
		node3 := randNode()

		assert.NoError(t, nodeRepo.SaveNode(node))
		assert.NoError(t, nodeRepo.SaveNode(node2))
		assert.NoError(t, nodeRepo.SaveNode(node3))
		list, err := nodeRepo.ListNode()
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(list), 3)

		has, err := nodeRepo.HasNode(node.Name)
		assert.NoError(t, err)
		assert.True(t, has)

		err = nodeRepo.DelNode(node.Name)
		assert.NoError(t, err)
		list, err = nodeRepo.ListNode()
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(list), 2)

		has, err = nodeRepo.HasNode(node.Name)
		assert.NoError(t, err)
		assert.False(t, has)
	}

	t.Run("sqlit", func(t *testing.T) {
		nodeRepoTest(t, sqliteRepo.NodeRepo())
	})

	t.Run("mysql", func(t *testing.T) {
		t.Skip()
		nodeRepoTest(t, mysqlRepo.NodeRepo())
	})
}

func TestGetNode(t *testing.T) {
	sqliteRepo, mysqlRepo := setupRepo(t)

	nodeRepoTest := func(t *testing.T, nodeRepo repo.NodeRepo) {
		node := randNode()

		assert.NoError(t, nodeRepo.SaveNode(node))
		result, err := nodeRepo.GetNode(node.Name)
		assert.NoError(t, err)
		assert.Equal(t, ObjectToString(node), ObjectToString(result))
	}

	t.Run("sqlit", func(t *testing.T) {
		nodeRepoTest(t, sqliteRepo.NodeRepo())
	})

	t.Run("mysql", func(t *testing.T) {
		t.SkipNow()
		nodeRepoTest(t, mysqlRepo.NodeRepo())
	})
}
