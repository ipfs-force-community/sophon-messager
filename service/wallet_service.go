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
	delWalletChan chan types.UUID

	l sync.RWMutex
}

func NewWalletService(repo repo.Repo, logger *logrus.Logger) (*WalletService, error) {
	ws := &WalletService{
		repo:          repo,
		log:           logger,
		walletClients: make(map[types.UUID]IWalletClient),
		delWalletChan: make(chan types.UUID, 10),
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
	cli, _, err := newWalletClient(ctx, wallet.Url, wallet.Token)
	if err != nil {
		return types.UUID{}, err
	}
	err = walletService.repo.WalletRepo().SaveWallet(wallet)
	if err != nil {
		return types.UUID{}, err
	}
	if err := walletService.addWalletClient(wallet.ID, &cli); err != nil {
		return types.UUID{}, err
	}

	return wallet.ID, nil
}

func (walletService *WalletService) GetWalletByID(ctx context.Context, uuid types.UUID) (*types.Wallet, error) {
	return walletService.repo.WalletRepo().GetWalletByID(uuid)
}

func (walletService *WalletService) GetWalletByName(ctx context.Context, name string) (*types.Wallet, error) {
	return walletService.repo.WalletRepo().GetWalletByName(name)
}

func (walletService *WalletService) HasWallet(ctx context.Context, name string) (bool, error) {
	return walletService.repo.WalletRepo().HasWallet(name)
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

func (walletService *WalletService) DeleteWallet(ctx context.Context, name string) (string, error) {
	w, err := walletService.GetWalletByName(ctx, name)
	if err != nil {
		return "", err
	}
	if err := walletService.repo.WalletRepo().DelWallet(w.ID); err != nil {
		return "", err
	}

	walletService.removeWalletClient(w.ID)
	walletService.delWalletChan <- w.ID
	walletService.log.Infof("delete wallet %s", name)

	return name, nil
}

/// wallet client ///
func (walletService *WalletService) GetWalletClient(walletId types.UUID) (IWalletClient, bool) {
	walletService.l.RLock()
	defer walletService.l.RUnlock()
	cli, ok := walletService.walletClients[walletId]

	return cli, ok
}

func (walletService *WalletService) ListWalletClient() ([]types.UUID, []IWalletClient) {
	walletService.l.RLock()
	defer walletService.l.RUnlock()
	clis := make([]IWalletClient, 0, len(walletService.walletClients))
	ids := make([]types.UUID, 0, len(walletService.walletClients))
	for id, cli := range walletService.walletClients {
		clis = append(clis, cli)
		ids = append(ids, id)
	}

	return ids, clis
}

func (walletService *WalletService) addWalletClient(walletID types.UUID, cli IWalletClient) error {
	walletService.l.Lock()
	defer walletService.l.Unlock()
	if _, ok := walletService.walletClients[walletID]; !ok {
		walletService.walletClients[walletID] = cli
	}

	return nil
}

func (walletService *WalletService) removeWalletClient(walletId types.UUID) {
	walletService.l.Lock()
	defer walletService.l.Unlock()

	delete(walletService.walletClients, walletId)
}
