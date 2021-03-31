package service

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/sirupsen/logrus"

	"github.com/ipfs-force-community/venus-messager/config"
	"github.com/ipfs-force-community/venus-messager/models/repo"
	"github.com/ipfs-force-community/venus-messager/types"
)

type NodeService struct {
	repo repo.Repo
	log  *logrus.Logger

	nodeInfos []*NodeInfo
}

type NodeInfo struct {
	url   string
	token string
	name  string
	cli   *NodeClient
}

func NewNodeService(repo repo.Repo, logger *logrus.Logger) (*NodeService, error) {
	ns := &NodeService{repo: repo, log: logger}

	var err error
	ns.nodeInfos, err = ns.loadNodeFromDB(context.TODO())
	if err != nil {
		return nil, err
	}

	return ns, nil
}

func (ns *NodeService) loadNodeFromDB(ctx context.Context) ([]*NodeInfo, error) {
	nodeList, err := ns.repo.NodeRepo().ListNode()
	if err != nil {
		return nil, err
	}
	nodeInfos := make([]*NodeInfo, len(nodeList))
	for i, node := range nodeList {
		nodeInfos[i] = &NodeInfo{
			url:   node.URL,
			token: node.Token,
			name:  node.Name,
		}
		cli, _, err := NewNodeClient(ctx, &config.NodeConfig{Token: node.Token, Url: node.URL})
		if err != nil {
			return nil, err
		}
		nodeInfos[i].cli = cli
	}

	return nodeInfos, err
}

func (ns *NodeService) SaveNode(ctx context.Context, node *types.Node) (struct{}, error) {
	if err := ns.checkNode(node); err != nil {
		return struct{}{}, err
	}
	cli, _, err := NewNodeClient(context.TODO(), &config.NodeConfig{Token: node.Token, Url: node.URL})
	if err != nil {
		return struct{}{}, err
	}
	if err := ns.repo.NodeRepo().SaveNode(node); err != nil {
		return struct{}{}, err
	}
	ns.nodeInfos = append(ns.nodeInfos, &NodeInfo{name: node.Name, url: node.URL, token: node.Token, cli: cli})
	ns.log.Infof("add node %s %s %s", node.Name, node.URL, node.Token)

	return struct{}{}, nil
}

func (ns *NodeService) checkNode(node *types.Node) error {
	urlToken := node.URL + node.Token
	for _, info := range ns.nodeInfos {
		if node.Name == info.name {
			return xerrors.Errorf("the same node name exists")
		}
		if info.url+info.token == urlToken {
			return xerrors.Errorf("the same url and token exists")
		}
	}

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

func (ns *NodeService) DeleteNode(ctx context.Context, name string) (struct{}, error) {
	if err := ns.repo.NodeRepo().DelNode(name); err != nil {
		return struct{}{}, err
	}
	ns.removeNode(name)
	ns.log.Infof("remove node %s", name)

	return struct{}{}, nil
}

func (ns *NodeService) removeNode(name string) {
	newNodeInfos := make([]*NodeInfo, 0, len(ns.nodeInfos)-1)
	for _, node := range ns.nodeInfos {
		if node.name == name {
			continue
		}
		newNodeInfos = append(newNodeInfos, node)
	}
	ns.nodeInfos = newNodeInfos
}
