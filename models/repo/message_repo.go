package repo

import (
	"time"

	"github.com/ipfs/go-cid"

	"github.com/filecoin-project/go-address"
	venustypes "github.com/filecoin-project/venus/pkg/types"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/venus-messager/types"
)

type MessageRepo interface {
	ExpireMessage(msg []*types.Message) error
	BatchSaveMessage(msg []*types.Message) error
	CreateMessage(msg *types.Message) error
	SaveMessage(msg *types.Message) error

	GetMessageByFromAndNonce(from address.Address, nonce uint64) (*types.Message, error)
	GetMessageByUid(id string) (*types.Message, error)
	HasMessageByUid(id string) (bool, error)
	GetMessageState(id string) (types.MessageState, error)
	GetMessageByCid(unsignedCid cid.Cid) (*types.Message, error)
	GetMessageBySignedCid(signedCid cid.Cid) (*types.Message, error)
	GetSignedMessageByTime(start time.Time) ([]*types.Message, error)
	GetSignedMessageByHeight(height abi.ChainEpoch) ([]*types.Message, error)
	ListMessage() ([]*types.Message, error)
	ListMessageByFromState(from address.Address, state types.MessageState, pageIndex, pageSize int) ([]*types.Message, error)
	ListMessageByAddress(addr address.Address) ([]*types.Message, error)
	ListFailedMessage() ([]*types.Message, error)
	ListBlockedMessage(addr address.Address, d time.Duration) ([]*types.Message, error)
	ListUnChainMessageByAddress(addr address.Address, topN int) ([]*types.Message, error)
	ListFilledMessageByAddress(addr address.Address) ([]*types.Message, error)
	ListFilledMessageByWallet(walletName string, addr address.Address) ([]*types.Message, error)
	ListFilledMessageByHeight(height abi.ChainEpoch) ([]*types.Message, error)
	ListUnchainedMsgs() ([]*types.Message, error)
	ListSignedMsgs() ([]*types.Message, error)
	ListFilledMessageBelowNonce(addr address.Address, nonce uint64) ([]*types.Message, error)

	UpdateMessageInfoByCid(unsignedCid string, receipt *venustypes.MessageReceipt, height abi.ChainEpoch, state types.MessageState, tsKey venustypes.TipSetKey) error
	UpdateMessageStateByCid(unsignedCid string, state types.MessageState) error
	UpdateMessageStateByID(id string, state types.MessageState) error
	UpdateUnFilledMessageState(walletName string, addr address.Address, state types.MessageState) error
	MarkBadMessage(id string) (struct{}, error)
	UpdateReturnValue(id string, returnVal string) error
}
