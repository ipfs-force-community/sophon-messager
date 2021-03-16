package repo

import "github.com/ipfs-force-community/venus-messager/types"

type WalletRepo interface {
	SaveWallet(wallet *types.Wallet) (types.UUID, error)
	GetWalletByID(uuid types.UUID) (*types.Wallet, error)
	GetWalletByName(name string) (*types.Wallet, error)
	ListWallet() ([]*types.Wallet, error)
	DelWallet(uuid types.UUID) error
}
