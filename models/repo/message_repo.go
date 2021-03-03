package repo

import (
	"github.com/ipfs-force-community/venus-messager/types"
)

type MessageRepo interface {
	SaveMessage(msg *types.Message, spec *types.SendSpec) (string, error)
	GetMessage(uuid string) (*types.Message, error)
	ListMessage() ([]*types.Message, error)
}
