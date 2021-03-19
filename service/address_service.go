package service

import (
	"context"
	"sync"
	"time"

	"github.com/filecoin-project/go-address"

	venustypes "github.com/filecoin-project/venus/pkg/types"
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
	UUID         types.UUID
	WalletClient IWalletClient
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

func (addressService *AddressService) SaveAddress(ctx context.Context, address *types.Address) (types.UUID, error) {
	return addressService.repo.AddressRepo().SaveAddress(ctx, address)
}

func (addressService *AddressService) UpdateNonce(ctx context.Context, uuid types.UUID, nonce uint64) (types.UUID, error) {
	return addressService.repo.AddressRepo().UpdateNonce(ctx, uuid, nonce)
}

func (addressService *AddressService) GetAddress(ctx context.Context, addr string) (*types.Address, error) {
	return addressService.repo.AddressRepo().GetAddress(ctx, addr)
}

func (addressService *AddressService) HasAddress(ctx context.Context, addr address.Address) (bool, error) {
	return addressService.repo.AddressRepo().HasAddress(ctx, addr)
}

func (addressService *AddressService) ListAddress(ctx context.Context) ([]*types.Address, error) {
	return addressService.repo.AddressRepo().ListAddress(ctx)
}

func (addressService *AddressService) DeleteAddress(ctx context.Context, addr string) (string, error) {
	err := addressService.repo.AddressRepo().DelAddress(ctx, addr)
	if err != nil {
		return "", err
	}

	// TODO: 确定已签名消息都已上链
	addressService.RemoveAddressInfo(addr)
	addressService.log.Infof("delete address %v", addr)

	return addr, nil
}

func (addressService *AddressService) getLocalAddress() error {
	addrsInfo, err := addressService.ListAddress(context.Background())
	if err != nil {
		return err
	}

	for _, info := range addrsInfo {
		cli, ok := addressService.walletService.walletClients[info.WalletID]
		if !ok {
			addressService.log.Errorf("not found wallet client, uuid: %v", info.WalletID)
			continue
		}

		addressService.SetAddressInfo(info.Addr, &AddressInfo{
			UUID:         info.ID,
			WalletClient: cli,
		})
	}

	return nil
}

func (addressService *AddressService) listenAddressChange(ctx context.Context) error {
	if err := addressService.getLocalAddress(); err != nil {
		return xerrors.Errorf("get local address and nonce failed: %v", err)
	}
	go func() {
		ticker := time.NewTicker(time.Duration(addressService.cfg.RemoteWalletSweepInterval) * time.Second)
		for {
			select {
			case <-ticker.C:
				for walletID, cli := range addressService.walletService.walletClients {
					if err := addressService.ProcessWallet(ctx, walletID, cli); err != nil {
						addressService.log.Errorf("process wallet failed, name: %s, error: %v", walletID, err)
					}
				}
			case <-ctx.Done():
				addressService.log.Warnf("context error: %v", ctx.Err())
				return
			}
		}
	}()

	return nil
}

func (addressService *AddressService) ProcessWallet(ctx context.Context, walletID types.UUID, cli IWalletClient) error {
	addrs, err := cli.WalletList(ctx)
	if err != nil {
		return xerrors.Errorf("get wallet list failed error: %v", err)
	}
	for _, addr := range addrs {
		if _, ok := addressService.addrInfo[addr.String()]; ok {
			continue
		}

		var nonce uint64
		actor, err := addressService.nodeClient.StateGetActor(context.Background(), addr, venustypes.EmptyTSK)
		if err != nil {
			addressService.log.Warnf("get actor failed, addr: %s, err: %v", addr, err)
		} else {
			nonce = actor.Nonce //current nonce should big than nonce on chain
		}

		ta := &types.Address{
			ID:        types.NewUUID(),
			Addr:      addr.String(),
			Nonce:     nonce,
			WalletID:  walletID,
			UpdatedAt: time.Now(),
			IsDeleted: -1,
		}
		_, err = addressService.SaveAddress(context.Background(), ta)
		if err != nil {
			addressService.log.Errorf("save address failed, addr: %v, err: %v", addr.String(), err)
			continue
		}
		addressService.SetAddressInfo(addr.String(), &AddressInfo{
			UUID:         ta.ID,
			WalletClient: cli,
		})
	}

	return nil
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

func (addressService *AddressService) RemoveAddressInfo(addr string) {
	addressService.l.Lock()
	defer addressService.l.Unlock()

	delete(addressService.addrInfo, addr)
}

func (addressService *AddressService) ListAddressInfo() map[string]AddressInfo {
	addressService.l.Lock()
	defer addressService.l.Unlock()
	addrInfos := make(map[string]AddressInfo, len(addressService.addrInfo))
	for addr, info := range addressService.addrInfo {
		addrInfos[addr] = *info
	}

	return addrInfos
}
