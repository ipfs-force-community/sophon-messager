package controller

import (
	"context"
	"github.com/ipfs-force-community/venus-messager/types"
)

type WalletController struct {
	BaseController
}

func (walletController WalletController) SaveWallet(ctx context.Context, wallet *types.Wallet) (string, error) {
	return walletController.Repo.WalletRepo().SaveWallet(wallet)
}

func (walletController WalletController) GetWallet(ctx context.Context, uuid string) (types.Wallet, error) {
	return walletController.Repo.WalletRepo().GetWallet(uuid)
}

func (walletController WalletController) ListWallet(ctx context.Context) ([]types.Wallet, error) {
	return walletController.Repo.WalletRepo().ListWallet()
}
