package service

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/filecoin-project/venus-messager/config"
	"github.com/filecoin-project/venus-messager/filestore"
	"github.com/filecoin-project/venus-messager/models"
	shared "github.com/filecoin-project/venus/venus-shared/types"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
	"github.com/stretchr/testify/assert"
)

func TestNodeService(t *testing.T) {
	ctx := context.Background()
	cfg := config.DefaultConfig()
	fsRepo := filestore.NewMockFileStore(t.TempDir())
	assert.NoError(t, fsRepo.ReplaceConfig(cfg))

	repo, err := models.SetDataBase(fsRepo)
	assert.NoError(t, err)
	assert.NoError(t, repo.AutoMigrate())

	nodeService := NewNodeService(repo.NodeRepo())

	nodeEquals := func(a, b *types.Node) bool {
		return a.ID == b.ID &&
			a.Name == b.Name &&
			a.URL == b.URL &&
			a.Token == b.Token &&
			a.Type == b.Type
	}

	nodeCases := make([]*types.Node, 0, 10)
	nodeMap := make(map[string]*types.Node)
	for i := 0; i < 10; i++ {
		node := &types.Node{
			ID:    shared.NewUUID(),
			Name:  fmt.Sprintf("node-%d", i),
			URL:   fmt.Sprintf("http://%d", i),
			Token: "token",
			Type:  types.NodeType(rand.Intn(2) + 1),
		}
		nodeCases = append(nodeCases, node)
		nodeMap[node.Name] = node
	}

	// save node
	for _, node := range nodeCases {
		assert.NoError(t, nodeService.SaveNode(ctx, node))
	}

	// get node
	for _, node := range nodeCases {
		n, err := nodeService.GetNode(ctx, node.Name)
		assert.NoError(t, err)
		assert.True(t, nodeEquals(n, node))

		// has node
		has, err := nodeService.HasNode(ctx, n.Name)
		assert.NoError(t, err)
		assert.True(t, has)
	}
	_, err = nodeService.GetNode(ctx, shared.NewUUID().String())
	assert.Error(t, err)

	has, err := nodeService.HasNode(ctx, shared.NewUUID().String())
	assert.NoError(t, err)
	assert.False(t, has)

	// list node
	nodes, err := nodeService.ListNode(ctx)
	assert.NoError(t, err)
	assert.Len(t, nodes, len(nodeCases))
	for _, n := range nodes {
		assert.True(t, nodeEquals(n, nodeMap[n.Name]))
	}

	// delete node
	assert.NoError(t, nodeService.DeleteNode(ctx, nodeCases[0].Name))
	has, err = nodeService.HasNode(ctx, nodeCases[0].Name)
	assert.NoError(t, err)
	assert.False(t, has)
}
