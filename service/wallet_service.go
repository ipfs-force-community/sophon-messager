package service

import (
	"context"
	"sync"

	"github.com/filecoin-project/go-address"
	"golang.org/x/xerrors"

	"github.com/sirupsen/logrus"

	"github.com/ipfs-force-community/venus-messager/models/repo"
	"github.com/ipfs-force-community/venus-messager/types"
)

type WalletService struct {
	repo repo.Repo
	log  *logrus.Logger

	walletClients map[types.UUID]IWalletClient

	l sync.Mutex
}

func NewWalletService(repo repo.Repo, logger *logrus.Logger) (*WalletService, error) {
	ws := &WalletService{
		repo:          repo,
		log:           logger,
		walletClients: make(map[types.UUID]IWalletClient),
	}

	walletList, err := ws.ListWallet(context.TODO())
	if err != nil {
		return nil, err
	}

	for _, w := range walletList {
		cli, _, err := newWalletClient(context.Background(), w.Url, w.Token)
		if err != nil {
			return nil, err
		}

		if _, ok := ws.walletClients[w.ID]; !ok {
			ws.walletClients[w.ID] = &cli
		}
	}

	return ws, err
}

func (walletService *WalletService) SaveWallet(ctx context.Context, wallet *types.Wallet) (types.UUID, error) {
	uuid, err := walletService.repo.WalletRepo().SaveWallet(wallet)
	if err != nil {
		return types.UUID{}, err
	}
	if err := walletService.updateWalletClient(ctx, wallet); err != nil {
		return types.UUID{}, err
	}

	return uuid, nil
}

func (walletService *WalletService) GetWalletByID(ctx context.Context, uuid types.UUID) (*types.Wallet, error) {
	return walletService.repo.WalletRepo().GetWalletByID(uuid)
}

func (walletService *WalletService) GetWalletByName(ctx context.Context, name string) (*types.Wallet, error) {
	return walletService.repo.WalletRepo().GetWalletByName(name)
}

func (walletService *WalletService) ListWallet(ctx context.Context) ([]*types.Wallet, error) {
	return walletService.repo.WalletRepo().ListWallet()
}

func (walletService *WalletService) ListRemoteWalletAddress(ctx context.Context, uuid types.UUID) ([]address.Address, error) {
	cli, ok := walletService.walletClients[uuid]
	if !ok {
		return nil, xerrors.Errorf("wallet %v not exit", uuid)
	}

	return cli.WalletList(ctx)
}

// nolint
func (walletService *WalletService) updateWalletClient(ctx context.Context, wallet *types.Wallet) error {
	walletService.l.Lock()
	defer walletService.l.Unlock()
	if _, ok := walletService.walletClients[wallet.ID]; !ok {
		cli, _, err := newWalletClient(ctx, wallet.Url, wallet.Token)
		if err != nil {
			return err
		}
		walletService.walletClients[wallet.ID] = &cli
	}

	return nil
}
