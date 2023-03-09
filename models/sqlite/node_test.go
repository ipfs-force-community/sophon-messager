package sqlite

import (
	"math/rand"
	"testing"

	"github.com/filecoin-project/venus-messager/testhelper"

	"gorm.io/gorm"

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
		Type:  types.NodeType(rand.Intn(2)),
	}
}

func TestNode(t *testing.T) {
	t.Run("create node", func(t *testing.T) {
		nodeRepo := setupRepo(t).NodeRepo()
		node := randNode()
		node2 := randNode()
		node3 := randNode()

		assert.NoError(t, nodeRepo.CreateNode(node))
		assert.NoError(t, nodeRepo.CreateNode(node2))
		assert.NoError(t, nodeRepo.CreateNode(node3))
	})

	t.Run("get node", func(t *testing.T) {
		nodeRepo := setupRepo(t).NodeRepo()
		node := randNode()
		node2 := randNode()
		node3 := randNode()
		randName := venustypes.NewUUID().String()

		assert.NoError(t, nodeRepo.CreateNode(node))
		assert.NoError(t, nodeRepo.CreateNode(node2))
		assert.NoError(t, nodeRepo.CreateNode(node3))

		res, err := nodeRepo.GetNode(node2.Name)
		assert.NoError(t, err)
		testhelper.Equal(t, node2, res)

		_, err = nodeRepo.GetNode(randName)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
	})

	t.Run("save node", func(t *testing.T) {
		nodeRepo := setupRepo(t).NodeRepo()
		node := randNode()
		node2 := randNode()
		node3 := randNode()

		assert.NoError(t, nodeRepo.CreateNode(node))
		assert.NoError(t, nodeRepo.CreateNode(node2))
		assert.NoError(t, nodeRepo.CreateNode(node3))

		node.URL = "url"
		node.Token = "token"
		node.Type = types.FullNode

		assert.NoError(t, nodeRepo.SaveNode(node))
		res, err := nodeRepo.GetNode(node.Name)
		assert.NoError(t, err)
		testhelper.Equal(t, node, res)
	})

	t.Run("list node", func(t *testing.T) {
		nodeRepo := setupRepo(t).NodeRepo()
		node := randNode()
		node2 := randNode()
		node3 := randNode()

		assert.NoError(t, nodeRepo.CreateNode(node))
		assert.NoError(t, nodeRepo.CreateNode(node2))
		assert.NoError(t, nodeRepo.CreateNode(node3))

		list, err := nodeRepo.ListNode()
		assert.NoError(t, err)
		assert.Equal(t, len(list), 3)

		testhelper.Equal(t, []*types.Node{node, node2, node3}, list)
	})

	t.Run("has node", func(t *testing.T) {
		nodeRepo := setupRepo(t).NodeRepo()
		node := randNode()
		node2 := randNode()
		node3 := randNode()
		randName := venustypes.NewUUID().String()

		assert.NoError(t, nodeRepo.CreateNode(node))
		assert.NoError(t, nodeRepo.CreateNode(node2))
		assert.NoError(t, nodeRepo.CreateNode(node3))

		has, err := nodeRepo.HasNode(node.Name)
		assert.NoError(t, err)
		assert.True(t, has)

		has, err = nodeRepo.HasNode(randName)
		assert.NoError(t, err)
		assert.False(t, has)
	})

	t.Run("delete node", func(t *testing.T) {
		nodeRepo := setupRepo(t).NodeRepo()
		node := randNode()
		node2 := randNode()
		node3 := randNode()
		randName := venustypes.NewUUID().String()

		assert.NoError(t, nodeRepo.CreateNode(node))
		assert.NoError(t, nodeRepo.CreateNode(node2))
		assert.NoError(t, nodeRepo.CreateNode(node3))

		err := nodeRepo.DelNode(node.Name)
		assert.NoError(t, err)

		list, err := nodeRepo.ListNode()
		assert.NoError(t, err)
		assert.Equal(t, len(list), 2)

		has, err := nodeRepo.HasNode(node.Name)
		assert.NoError(t, err)
		assert.False(t, has)

		_, err = nodeRepo.GetNode(node.Name)
		assert.Error(t, err)

		err = nodeRepo.DelNode(randName)
		assert.Error(t, err)
	})
}
