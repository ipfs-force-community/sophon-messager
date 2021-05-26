package controller

import (
	"context"

	"github.com/filecoin-project/go-address"

	"github.com/filecoin-project/venus-messager/service"
	"github.com/filecoin-project/venus-messager/types"
)

type WalletController struct {
	BaseController
	WalletService *service.WalletService
}

func (walletController WalletController) SaveWallet(ctx context.Context, wallet *types.Wallet) (types.UUID, error) {
	return walletController.WalletService.SaveWallet(ctx, wallet)
}

func (walletController WalletController) GetWalletByName(ctx context.Context, name string) (*types.Wallet, error) {
	return walletController.WalletService.GetWalletByName(ctx, name)
}

func (walletController WalletController) GetWalletByID(ctx context.Context, id types.UUID) (*types.Wallet, error) {
	return walletController.WalletService.GetWalletByID(ctx, id)
}

func (walletController WalletController) HasWallet(ctx context.Context, name string) (bool, error) {
	return walletController.WalletService.HasWallet(ctx, name)
}

func (walletController WalletController) ListWallet(ctx context.Context) ([]*types.Wallet, error) {
	return walletController.WalletService.ListWallet(ctx)
}

func (walletController WalletController) ListRemoteWalletAddress(ctx context.Context, walletName string) ([]address.Address, error) {
	return walletController.WalletService.ListRemoteWalletAddress(ctx, walletName)
}

func (walletController WalletController) DeleteWallet(ctx context.Context, name string) (string, error) {
	return walletController.WalletService.DeleteWallet(ctx, name)
}

func (walletController WalletController) HasWalletAddress(ctx context.Context, walletName string, addr address.Address) (bool, error) {
	return walletController.WalletService.HasWalletAddress(ctx, walletName, addr)
}

func (walletController WalletController) ListWalletAddress(ctx context.Context) ([]*types.WalletAddress, error) {
	return walletController.WalletService.ListWalletAddress(ctx)
}

func (walletController WalletController) GetWalletAddress(ctx context.Context, walletName string, addr address.Address) (*types.WalletAddress, error) {
	return walletController.WalletService.GetWalletAddress(ctx, walletName, addr)
}
