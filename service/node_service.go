package service

import (
	"context"

	"github.com/filecoin-project/venus-messager/config"
	"github.com/filecoin-project/venus-messager/log"
	"github.com/filecoin-project/venus-messager/models/repo"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
)

type NodeService struct {
	repo repo.Repo
	log  *log.Logger
}

func NewNodeService(repo repo.Repo, logger *log.Logger) *NodeService {
	return &NodeService{repo: repo, log: logger}
}

func (ns *NodeService) SaveNode(ctx context.Context, node *types.Node) error {
	// try connect node
	_, close, err := NewNodeClient(context.TODO(), &config.NodeConfig{Token: node.Token, Url: node.URL})
	if err != nil {
		return err
	}
	close()
	if err := ns.repo.NodeRepo().SaveNode(node); err != nil {
		return err
	}
	ns.log.Infof("add node %s", node.Name)

	return nil
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

func (ns *NodeService) DeleteNode(ctx context.Context, name string) error {
	if err := ns.repo.NodeRepo().DelNode(name); err != nil {
		return err
	}
	ns.log.Infof("delete node %s", name)

	return nil
}
