package controller

import (
	"context"
	"github.com/filecoin-project/go-address"

	"github.com/ipfs-force-community/venus-messager/service"
	"github.com/ipfs-force-community/venus-messager/types"
)

type WalletController struct {
	BaseController
	WalletService *service.WalletService
}

func (walletController WalletController) SaveWallet(ctx context.Context, wallet *types.Wallet) (string, error) {
	return walletController.WalletService.SaveWallet(ctx, wallet)
}

func (walletController WalletController) GetWallet(ctx context.Context, uuid types.UUID) (*types.Wallet, error) {
	return walletController.WalletService.GetWallet(ctx, uuid)
}

func (walletController WalletController) ListWallet(ctx context.Context) ([]*types.Wallet, error) {
	return walletController.WalletService.ListWallet(ctx)
}

func (walletController WalletController) ListWalletAddress(ctx context.Context, name string) ([]address.Address, error) {
	return walletController.WalletService.ListWalletAddress(ctx, name)
}
