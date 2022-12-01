package repo

import types "github.com/filecoin-project/venus/venus-shared/types/messager"

type INodeProvider interface {
	ListNode() ([]*types.Node, error)
}

//go:generate  mockgen -destination=../../mocks/mock_node_repo.go -package=mocks github.com/filecoin-project/venus-messager/models/repo NodeRepo

type NodeRepo interface {
	CreateNode(node *types.Node) error
	SaveNode(node *types.Node) error
	GetNode(name string) (*types.Node, error)
	HasNode(name string) (bool, error)
	ListNode() ([]*types.Node, error)
	DelNode(name string) error
}

func NewINodeRepo(repo Repo) NodeRepo {
	return repo.NodeRepo()
}

func NewINodeProvider(s NodeRepo) INodeProvider {
	return s
}
