package controller

import (
	"context"
	"github.com/ipfs-force-community/venus-messager/service"
	"github.com/ipfs-force-community/venus-messager/types"
)

type WalletController struct {
	BaseController
	walletService service.WalletService
}

func (walletController WalletController) SaveWallet(ctx context.Context, wallet *types.Wallet) (string, error) {
	return walletController.walletService.SaveWallet(ctx, wallet)
}

func (walletController WalletController) GetWallet(ctx context.Context, uuid string) (types.Wallet, error) {
	return walletController.walletService.GetWallet(ctx, uuid)
}

func (walletController WalletController) ListWallet(ctx context.Context) ([]types.Wallet, error) {
	return walletController.walletService.ListWallet(ctx)
}
