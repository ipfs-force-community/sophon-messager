package repo

import "github.com/ipfs-force-community/venus-messager/types"

type WalletRepo interface {
	SaveWallet(wallet *types.Wallet) error
	GetWalletByID(uuid types.UUID) (*types.Wallet, error)
	GetWalletByName(name string) (*types.Wallet, error)
	HasWallet(name string) (bool, error)
	ListWallet() ([]*types.Wallet, error)
	DelWallet(uuid types.UUID) error
}
