package repo

import types "github.com/filecoin-project/venus/venus-shared/types/messager"

type NodeRepo interface {
	CreateNode(node *types.Node) error
	SaveNode(node *types.Node) error
	GetNode(name string) (*types.Node, error)
	HasNode(name string) (bool, error)
	ListNode() ([]*types.Node, error)
	DelNode(name string) error
}
