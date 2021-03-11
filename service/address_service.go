package service

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"

	"github.com/ipfs-force-community/venus-messager/config"
	"github.com/ipfs-force-community/venus-messager/models/repo"
	"github.com/ipfs-force-community/venus-messager/types"
)

type AddressService struct {
	repo repo.Repo
	log  *logrus.Logger

	walletService *WalletService
	nodeClient    *NodeClient
	cfg           *config.AddressConfig

	addrInfo map[string]*AddressInfo
	l        sync.Mutex
}

type AddressInfo struct {
	Nonce        uint64
	WalletClient *WalletClient
}

func NewAddressService(repo repo.Repo, logger *logrus.Logger, walletService *WalletService, nodeClient *NodeClient, cfg *config.AddressConfig) (*AddressService, error) {
	addressService := &AddressService{
		repo:          repo,
		log:           logger,
		walletService: walletService,
		nodeClient:    nodeClient,
		cfg:           cfg,
		addrInfo:      make(map[string]*AddressInfo),
	}

	if err := addressService.listenAddressChange(context.TODO()); err != nil {
		return nil, err
	}

	return addressService, nil
}

func (addressService *AddressService) SaveAddress(ctx context.Context, address *types.Address) (string, error) {
	return addressService.repo.AddressRepo().SaveAddress(ctx, address)
}

func (addressService *AddressService) GetAddress(ctx context.Context, addr string) (*types.Address, error) {
	return addressService.repo.AddressRepo().GetAddress(ctx, addr)
}

func (addressService *AddressService) ListAddress(ctx context.Context) ([]*types.Address, error) {
	return addressService.repo.AddressRepo().ListAddress(ctx)
}
func (addressService *AddressService) getLocalAddressAndNonce() error {
	addrs, err := addressService.ListAddress(context.Background())
	if err != nil {
		return err
	}

	for _, addr := range addrs {
		addressService.SetAddressInfo(addr.Addr, &AddressInfo{
			Nonce:        addr.Nonce,
			WalletClient: nil,
		})
	}

	return nil
}

func (addressService *AddressService) listenAddressChange(ctx context.Context) error {
	if err := addressService.getLocalAddressAndNonce(); err != nil {
		return xerrors.Errorf("get local address and nonce failed: %v", err)
	}
	go func() {
		ticker := time.NewTicker(time.Duration(addressService.cfg.RemoteWalletSweepInterval) * time.Second)
		for {
			select {
			case <-ticker.C:
				for key, cli := range addressService.walletService.walletClients {
					if err := addressService.ProcessWallet(ctx, cli); err != nil {
						addressService.log.Errorf("process wallet failed, name: %s, error: %v", key, err)
					}
				}
			default:
			}
		}
	}()

	return nil
}

func (addressService *AddressService) ProcessWallet(ctx context.Context, cli *WalletClient) error {
	addrs, err := cli.WalletList(ctx)
	if err != nil {
		return xerrors.Errorf("get wallet list failed error: %v", err)
	}
	for _, addr := range addrs {
		if info, ok := addressService.addrInfo[addr.String()]; ok {
			if info.WalletClient == nil {
				info.WalletClient = cli
			}
			continue
		}

		var nonce uint64
		actor, err := addressService.nodeClient.GetActor(context.Background(), addr)
		if err != nil {
			addressService.log.Warnf("get actor failed, addr: %s, err: %v", addr, err)
		} else {
			nonce = actor.Nonce
		}

		ta := &types.Address{
			Addr:      addr.String(),
			Nonce:     nonce,
			UpdatedAt: time.Now(),
		}
		if err := addressService.SetNonceToLocal(addr.String(), actor.Nonce); err != nil {
			addressService.log.Errorf("set nonce failed addr: %v, err: %v", ta, err)
		}
	}

	return nil
}

func (addressService *AddressService) GetNonce(addr string) uint64 {
	addressService.l.Lock()
	defer addressService.l.Unlock()
	if info, ok := addressService.addrInfo[addr]; ok {
		return info.Nonce
	}

	return 0
}

func (addressService *AddressService) SetNonce(addr string, nonce uint64) {
	addressService.l.Lock()
	defer addressService.l.Unlock()
	if info, ok := addressService.addrInfo[addr]; ok {
		info.Nonce = nonce
	}
}

func (addressService *AddressService) SetAddressInfo(addr string, info *AddressInfo) {
	addressService.l.Lock()
	defer addressService.l.Unlock()

	addressService.addrInfo[addr] = info
}

func (addressService *AddressService) GetAddressInfo(addr string) (*AddressInfo, bool) {
	addressService.l.Lock()
	defer addressService.l.Unlock()
	if info, ok := addressService.addrInfo[addr]; ok {
		return info, ok
	}

	return nil, false
}

func (addressService *AddressService) SetNonceToLocal(addr string, nonce uint64) error {
	_, err := addressService.SaveAddress(context.Background(), &types.Address{
		Addr:      addr,
		Nonce:     nonce,
		UpdatedAt: time.Now(),
	})

	return err
}
