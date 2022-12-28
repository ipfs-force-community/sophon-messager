package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/filecoin-project/venus-messager/mocks"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNodeService(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockNodeRepo := mocks.NewMockNodeRepo(mockCtrl)
	nodeService := NewNodeService(mockNodeRepo)
	tempStore := make(map[string]*types.Node)

	nodeCases := []*types.Node{}
	for i := 0; i < 2; i++ {
		node := &types.Node{}
		node.Name = fmt.Sprintf("node-%d", i)
		node.URL = fmt.Sprintf("http://%d", i)
		nodeCases = append(nodeCases, node)
		// todo: add node to tempStore by SaveNode
		tempStore[node.Name] = node
	}

	nodeEqual := func(a, b *types.Node) bool {
		return a.Name == b.Name && a.URL == b.URL
	}

	// todo: comlete test case of SaveNode
	// t.Run("SaveNode", func(t *testing.T) {
	// 	mockNodeRepo.EXPECT().SaveNode(gomock.Any()).Do(func(node *types.Node) {
	// 		tempStore[node.Name] = node
	// 	}).Return(nil)
	// 	err := nodeService.SaveNode(context.Background(), &types.Node{})
	// 	assert.NoError(t, err)
	// })

	t.Run("GetNode", func(t *testing.T) {
		mockNodeRepo.EXPECT().GetNode(gomock.Any()).Return(tempStore[nodeCases[0].Name], nil)
		node, err := nodeService.GetNode(context.Background(), nodeCases[0].Name)
		assert.NoError(t, err)
		assert.True(t, nodeEqual(node, nodeCases[0]))

		mockNodeRepo.EXPECT().GetNode(gomock.Any()).Return(nil, fmt.Errorf("error"))
		node, err = nodeService.GetNode(context.Background(), nodeCases[0].Name)
		assert.Error(t, err)
		assert.Nil(t, node)
	})

	t.Run("HasNode", func(t *testing.T) {
		// success
		mockNodeRepo.EXPECT().HasNode(gomock.Any()).Return(true, nil)
		has, err := nodeService.HasNode(context.Background(), nodeCases[0].Name)
		assert.NoError(t, err)
		assert.True(t, has)

		// fail
		mockNodeRepo.EXPECT().HasNode(gomock.Any()).Return(false, fmt.Errorf("error"))
		has, err = nodeService.HasNode(context.Background(), nodeCases[0].Name)
		assert.Error(t, err)
		assert.False(t, has)
	})

	t.Run("ListNode", func(t *testing.T) {
		mockNodeRepo.EXPECT().ListNode().Return(nodeCases, nil)
		nodes, err := nodeService.ListNode(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, len(nodes), len(nodeCases))
		for i := 0; i < len(nodes); i++ {
			assert.True(t, nodeEqual(nodes[i], nodeCases[i]))
		}

		mockNodeRepo.EXPECT().ListNode().Return(nil, fmt.Errorf("error"))
		nodes, err = nodeService.ListNode(context.Background())
		assert.Error(t, err)
		assert.Nil(t, nodes)
	})

	t.Run("DelNode", func(t *testing.T) {
		mockNodeRepo.EXPECT().DelNode(gomock.Any()).Return(nil)
		err := nodeService.DeleteNode(context.Background(), nodeCases[0].Name)
		assert.NoError(t, err)

		mockNodeRepo.EXPECT().DelNode(gomock.Any()).Return(fmt.Errorf("error"))
		err = nodeService.DeleteNode(context.Background(), nodeCases[0].Name)
		assert.Error(t, err)
	})
}
