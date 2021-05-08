package repo

import "github.com/filecoin-project/venus-messager/types"

type NodeRepo interface {
	CreateNode(node *types.Node) error
	SaveNode(node *types.Node) error
	GetNode(name string) (*types.Node, error)
	HasNode(name string) (bool, error)
	ListNode() ([]*types.Node, error)
	DelNode(name string) error
}
