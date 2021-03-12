package repo

import "github.com/ipfs-force-community/venus-messager/types"

type WalletRepo interface {
	SaveWallet(wallet *types.Wallet) (string, error)
	GetWallet(uuid types.UUID) (*types.Wallet, error)
	ListWallet() ([]*types.Wallet, error)
	DelWallet(uuid types.UUID) error
}
