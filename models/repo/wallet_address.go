package repo

import (
	"github.com/filecoin-project/go-address"
	"github.com/ipfs-force-community/venus-messager/types"
)

type WalletAddressRepo interface {
	SaveWalletAddress(wa *types.WalletAddress) error
	GetWalletAddress(walletName string, addr address.Address) (*types.WalletAddress, error)
	GetOneRecord(walletName string, addr address.Address) (*types.WalletAddress, error)
	HasWalletAddress(walletName string, addr address.Address) (bool, error)
	ListWalletAddress() ([]*types.WalletAddress, error)
	UpdateAddressState(walletName string, addr address.Address, state types.State) error
	UpdateSelectMsgNum(walletName string, addr address.Address, selMsgNum uint64) error
	DelWalletAddress(walletName string, addr address.Address) error
}
