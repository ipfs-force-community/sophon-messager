package controller

import (
	"context"

	"github.com/filecoin-project/venus-messager/service"
	"github.com/filecoin-project/venus-messager/types"
)

type NodeController struct {
	BaseController
	NodeService *service.NodeService
}

func (nodeController NodeController) SaveNode(ctx context.Context, node *types.Node) error {
	return nodeController.NodeService.SaveNode(ctx, node)
}

func (nodeController NodeController) GetNode(ctx context.Context, name string) (*types.Node, error) {
	return nodeController.NodeService.GetNode(ctx, name)
}

func (nodeController NodeController) HasNode(ctx context.Context, name string) (bool, error) {
	return nodeController.NodeService.HasNode(ctx, name)
}

func (nodeController NodeController) ListNode(ctx context.Context) ([]*types.Node, error) {
	return nodeController.NodeService.ListNode(ctx)
}

func (nodeController NodeController) DeleteNode(ctx context.Context, name string) error {
	return nodeController.NodeService.DeleteNode(ctx, name)
}
