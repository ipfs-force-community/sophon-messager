package sqlite

import (
	"math/rand"
	"testing"

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
	nodeRepo := setupRepo(t).NodeRepo()

	node := randNode()
	node2 := randNode()
	node3 := randNode()
	randName := venustypes.NewUUID().String()

	t.Run("create node", func(t *testing.T) {
		assert.NoError(t, nodeRepo.CreateNode(node))
		assert.NoError(t, nodeRepo.CreateNode(node2))
		assert.NoError(t, nodeRepo.CreateNode(node3))
	})

	t.Run("get node", func(t *testing.T) {
		res, err := nodeRepo.GetNode(node2.Name)
		assert.NoError(t, err)
		assert.Equal(t, node2, res)

		_, err = nodeRepo.GetNode(randName)
		assert.Error(t, err)
	})

	t.Run("save node", func(t *testing.T) {
		tmp := *node
		tmp.URL = "url"
		tmp.Token = "token"
		tmp.Type = types.FullNode

		assert.NoError(t, nodeRepo.SaveNode(&tmp))
		res, err := nodeRepo.GetNode(tmp.Name)
		assert.NoError(t, err)
		assert.Equal(t, &tmp, res)
	})

	t.Run("list node", func(t *testing.T) {
		list, err := nodeRepo.ListNode()
		assert.NoError(t, err)
		assert.Equal(t, len(list), 3)
	})

	t.Run("has node", func(t *testing.T) {
		has, err := nodeRepo.HasNode(node.Name)
		assert.NoError(t, err)
		assert.True(t, has)

		has, err = nodeRepo.HasNode(randName)
		assert.NoError(t, err)
		assert.False(t, has)
	})

	t.Run("delete node", func(t *testing.T) {
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
