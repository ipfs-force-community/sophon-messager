package service

import (
	"context"
	"github.com/ipfs-force-community/venus-messager/models/repo"
	"github.com/ipfs-force-community/venus-messager/types"
	"github.com/sirupsen/logrus"
)

type WalletService struct {
	Repo   repo.Repo
	Logger *logrus.Logger
}

func NewWalletService(repo repo.Repo, logger *logrus.Logger) *WalletService {
	return &WalletService{Repo: repo, Logger: logger}
}

func (walletService WalletService) SaveWallet(ctx context.Context, wallet *types.Wallet) (string, error) {
	return walletService.Repo.WalletRepo().SaveWallet(wallet)
}

func (walletService WalletService) GetWallet(ctx context.Context, uuid string) (types.Wallet, error) {
	return walletService.Repo.WalletRepo().GetWallet(uuid)
}

func (walletService WalletService) ListWallet(ctx context.Context) ([]types.Wallet, error) {
	return walletService.Repo.WalletRepo().ListWallet()
}
