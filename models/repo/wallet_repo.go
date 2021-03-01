package repo

import "github.com/ipfs-force-community/venus-messager/types"

type WalletRepo interface {
	SaveWallet(msg *types.Wallet) (string, error)
	GetWallet(uuid string) (types.Wallet, error)
	ListWallet() ([]types.Wallet, error)
	DelWallet(uuid string) error
}
