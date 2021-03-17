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

func (walletController WalletController) SaveWallet(ctx context.Context, wallet *types.Wallet) (types.UUID, error) {
	return walletController.WalletService.SaveWallet(ctx, wallet)
}

func (walletController WalletController) GetWalletByID(ctx context.Context, uuid types.UUID) (*types.Wallet, error) {
	return walletController.WalletService.GetWalletByID(ctx, uuid)
}

func (walletController WalletController) GetWalletByName(ctx context.Context, name string) (*types.Wallet, error) {
	return walletController.WalletService.GetWalletByName(ctx, name)
}

func (walletController WalletController) ListWallet(ctx context.Context) ([]*types.Wallet, error) {
	return walletController.WalletService.ListWallet(ctx)
}

func (walletController WalletController) ListRemoteWalletAddress(ctx context.Context, uuid types.UUID) ([]address.Address, error) {
	return walletController.WalletService.ListRemoteWalletAddress(ctx, uuid)
}
