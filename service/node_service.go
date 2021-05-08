package service

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/filecoin-project/venus-messager/config"
	"github.com/filecoin-project/venus-messager/models/repo"
	"github.com/filecoin-project/venus-messager/types"
)

type NodeService struct {
	repo repo.Repo
	log  *logrus.Logger
}

func NewNodeService(repo repo.Repo, logger *logrus.Logger) *NodeService {
	return &NodeService{repo: repo, log: logger}
}

func (ns *NodeService) SaveNode(ctx context.Context, node *types.Node) (struct{}, error) {
	// try connect node
	_, close, err := NewNodeClient(context.TODO(), &config.NodeConfig{Token: node.Token, Url: node.URL})
	if err != nil {
		return struct{}{}, err
	}
	close()
	if err := ns.repo.NodeRepo().SaveNode(node); err != nil {
		return struct{}{}, err
	}
	ns.log.Infof("add node %s", node.Name)

	return struct{}{}, nil
}

func (ns *NodeService) GetNode(ctx context.Context, name string) (*types.Node, error) {
	return ns.repo.NodeRepo().GetNode(name)
}

func (ns *NodeService) HasNode(ctx context.Context, name string) (bool, error) {
	return ns.repo.NodeRepo().HasNode(name)
}

func (ns *NodeService) ListNode(ctx context.Context) ([]*types.Node, error) {
	return ns.repo.NodeRepo().ListNode()
}

func (ns *NodeService) DeleteNode(ctx context.Context, name string) (struct{}, error) {
	if err := ns.repo.NodeRepo().DelNode(name); err != nil {
		return struct{}{}, err
	}
	ns.log.Infof("delete node %s", name)

	return struct{}{}, nil
}
