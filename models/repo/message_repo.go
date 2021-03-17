package repo

import (
	"time"

	"github.com/filecoin-project/go-address"
	venustypes "github.com/filecoin-project/venus/pkg/types"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/ipfs-force-community/venus-messager/types"
)

type MessageRepo interface {
	ExpireMessage(msg []*types.Message) error
	BatchSaveMessage(msg []*types.Message) error
	SaveMessage(msg *types.Message) (types.UUID, error)

	GetMessage(uuid types.UUID) (*types.Message, error)
	GetMessageState(uuid types.UUID) (types.MessageState, error)
	GetMessageByCid(unsignedCid string) (*types.Message, error)
	GetSignedMessageByTime(start time.Time) ([]*types.Message, error)
	GetSignedMessageByHeight(height abi.ChainEpoch) ([]*types.Message, error)
	ListMessage() ([]*types.Message, error)
	ListUnChainMessageByAddress(addr address.Address) ([]*types.Message, error)
	ListFilledMessageByAddress(addr address.Address) ([]*types.Message, error)
	ListUnchainedMsgs() ([]*types.Message, error)

	UpdateMessageStateByCid(unsignedCid string, state types.MessageState) error
	UpdateMessageInfoByCid(unsignedCid string, receipt *venustypes.MessageReceipt, height abi.ChainEpoch, state types.MessageState, tsKey string) (string, error)
}
