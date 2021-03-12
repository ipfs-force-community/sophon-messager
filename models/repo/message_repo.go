package repo

import (
	"github.com/filecoin-project/go-address"
	"time"

	"github.com/filecoin-project/go-state-types/abi"
	venustypes "github.com/filecoin-project/venus/pkg/types"
	"github.com/ipfs-force-community/venus-messager/types"
)

type MessageRepo interface {
	ExpireMessage(msg []*types.Message) error
	BatchSaveMessage(msg []*types.Message) error
	SaveMessage(msg *types.Message) (types.UUID, error)
	GetMessage(uuid types.UUID) (*types.Message, error)
	UpdateMessageReceipt(cid string, receipt *venustypes.MessageReceipt, height abi.ChainEpoch, state types.MessageState) (string, error)
	ListMessage() ([]*types.Message, error)
	ListUnChainMessageByAddress(addr address.Address) ([]*types.Message, error)
	ListUnchainedMsgs() ([]*types.Message, error)
	GetMessageByCid(cid string) (*types.Message, error)
	GetMessageByTime(start time.Time) ([]*types.Message, error)
	UpdateMessageStateByCid(cid string, state types.MessageState) error
}
