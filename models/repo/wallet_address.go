package repo

import (
	"github.com/filecoin-project/venus-messager/types"
)

type WalletAddressRepo interface {
	SaveWalletAddress(wa *types.WalletAddress) error
	GetWalletAddress(walletID, addrID types.UUID) (*types.WalletAddress, error)
	GetOneRecord(walletID, addrID types.UUID) (*types.WalletAddress, error)
	GetWalletAddressByWalletID(walletID types.UUID) ([]*types.WalletAddress, error)
	HasWalletAddress(walletID, addrID types.UUID) (bool, error)
	ListWalletAddress() ([]*types.WalletAddress, error)
	UpdateAddressState(walletID, addrID types.UUID, state types.State) error
	UpdateSelectMsgNum(walletID, addrID types.UUID, selMsgNum uint64) error
	DelWalletAddress(walletID, addrID types.UUID) error
}
