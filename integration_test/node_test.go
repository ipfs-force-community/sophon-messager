package integration

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/filecoin-project/venus-messager/testhelper"

	shared "github.com/filecoin-project/venus/venus-shared/types"

	"github.com/filecoin-project/venus/venus-shared/api/messager"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"

	"github.com/filecoin-project/venus-messager/config"
	"github.com/stretchr/testify/assert"
)

func TestNodeAPI(t *testing.T) {
	ctx := context.Background()
	cfg := config.DefaultConfig()
	cfg.API.Address = "/ip4/0.0.0.0/tcp/0"
	cfg.MessageService.SkipPushMessage = true
	cfg.MessageService.WaitingChainHeadStableDuration = 2 * time.Second
	ms, err := mockMessagerServer(ctx, t.TempDir(), cfg)
	assert.NoError(t, err)

	go ms.start(ctx)
	assert.NoError(t, <-ms.appStartErr)

	full, err := testhelper.MockFullNodeServer(t)
	assert.NoError(t, err)

	cli, closer, err := newMessagerClient(ctx, ms.port, ms.token)
	assert.NoError(t, err)
	defer closer()

	nodeNum := 10
	nodeNames := make([]string, nodeNum)
	nodes := make([]*types.Node, nodeNum)
	for i := 0; i < nodeNum; i++ {
		nodes[i] = &types.Node{
			ID:    shared.NewUUID(),
			Name:  "node_" + strconv.Itoa(i),
			URL:   fmt.Sprintf("/ip4/127.0.0.1/tcp/%s", full.Port),
			Token: full.Token,
			Type:  types.NodeType(rand.Intn(3)),
		}
		nodeNames[i] = nodes[i].Name
	}

	t.Run("test save node", func(t *testing.T) {
		for _, node := range nodes {
			assert.NoError(t, cli.SaveNode(ctx, node))
		}
	})
	t.Run("test get node", func(t *testing.T) {
		testGetNode(ctx, t, cli, nodes)
	})
	t.Run("test has node", func(t *testing.T) {
		testHasNode(ctx, t, cli, nodeNames)
	})
	t.Run("test list node", func(t *testing.T) {
		testListNode(ctx, t, cli, nodes)
	})
	t.Run("test delete node", func(t *testing.T) {
		testDeleteNode(ctx, t, cli, nodeNames)
	})

	assert.NoError(t, full.Stop(ctx))
	assert.NoError(t, ms.stop(ctx))
}

func testGetNode(ctx context.Context, t *testing.T, cli messager.IMessager, nodes []*types.Node) {
	for i, node := range nodes {
		res, err := cli.GetNode(ctx, node.Name)
		assert.NoError(t, err)
		assert.Equal(t, node, res)

		if i%2 == 0 {
			_, err = cli.GetNode(ctx, node.Name+"_name")
			assert.Contains(t, err.Error(), "record not found")
		}
	}
}

func testHasNode(ctx context.Context, t *testing.T, cli messager.IMessager, nodeNames []string) {
	for i, name := range nodeNames {
		has, err := cli.HasNode(ctx, name)
		assert.NoError(t, err)
		assert.True(t, has)

		if i%2 == 0 {
			has, err = cli.HasNode(ctx, name+"_name")
			assert.NoError(t, err)
			assert.False(t, has)
		}
	}
}

func testListNode(ctx context.Context, t *testing.T, cli messager.IMessager, nodes []*types.Node) {
	list, err := cli.ListNode(ctx)
	assert.NoError(t, err)
	assert.Len(t, list, len(nodes))

	for _, node := range nodes {
		for _, one := range list {
			if node.Name == one.Name {
				assert.Equal(t, node, one)
			}
		}
	}
}

func testDeleteNode(ctx context.Context, t *testing.T, cli messager.IMessager, nodeNames []string) {
	for _, name := range nodeNames {
		err := cli.DeleteNode(ctx, name)
		assert.NoError(t, err)

		_, err = cli.GetNode(ctx, name)
		assert.Error(t, err)
		has, err := cli.HasNode(ctx, name)
		assert.NoError(t, err)
		assert.False(t, has)
	}
}
