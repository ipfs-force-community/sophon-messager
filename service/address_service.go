package service

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/filecoin-project/go-address"
	"github.com/sirupsen/logrus"

	"github.com/filecoin-project/venus-messager/gateway"
	"github.com/filecoin-project/venus-messager/models/repo"
	"github.com/filecoin-project/venus-messager/types"
)

var errAddressNotExists = xerrors.New("address not exists")

type AddressService struct {
	repo repo.Repo
	log  *logrus.Logger

	sps          *SharedParamsService
	walletClient *gateway.IWalletCli
}

func NewAddressService(repo repo.Repo, logger *logrus.Logger, sps *SharedParamsService, walletClient *gateway.IWalletCli) *AddressService {
	addressService := &AddressService{
		repo: repo,
		log:  logger,

		sps:          sps,
		walletClient: walletClient,
	}

	return addressService
}

func (addressService *AddressService) SaveAddress(ctx context.Context, address *types.Address) (types.UUID, error) {
	err := addressService.repo.Transaction(func(txRepo repo.TxRepo) error {
		has, err := txRepo.AddressRepo().HasAddress(ctx, address.WalletName, address.Addr)
		if err != nil {
			return err
		}
		if has {
			return xerrors.Errorf("address already exists")
		}
		return txRepo.AddressRepo().SaveAddress(ctx, address)
	})

	return address.ID, err
}

func (addressService *AddressService) UpdateNonce(ctx context.Context, addr address.Address, nonce uint64) (address.Address, error) {
	return addr, addressService.repo.AddressRepo().UpdateNonce(ctx, addr, nonce)
}

func (addressService *AddressService) GetAddress(ctx context.Context, walletName string, addr address.Address) (*types.Address, error) {
	addrInfo, err := addressService.repo.AddressRepo().GetAddress(ctx, walletName, addr)
	if err != nil {
		return nil, err
	}
	has, err := addressService.walletClient.WalletHas(ctx, walletName, addr)
	if err != nil {
		return nil, xerrors.Errorf("call 'WalletHas' failed %v", err)
	}
	if has {
		return addrInfo, nil
	}

	return nil, errAddressNotExists
}

//TODO: Make sure the wallet has an address ？
func (addressService *AddressService) HasAddress(ctx context.Context, walletName string, addr address.Address) (bool, error) {
	return addressService.repo.AddressRepo().HasAddress(ctx, walletName, addr)
}

// Deprecated: use HasAddress instead.
func (addressService *AddressService) HasWalletAddress(ctx context.Context, walletName string, addr address.Address) (bool, error) {
	return addressService.repo.AddressRepo().HasAddress(ctx, walletName, addr)
}

func (addressService *AddressService) ListAddress(ctx context.Context) ([]*types.Address, error) {
	return addressService.listAddress(ctx, true)
}

func (addressService *AddressService) listAddress(ctx context.Context, verify bool) ([]*types.Address, error) {
	list, err := addressService.repo.AddressRepo().ListAddress(ctx)
	if err != nil {
		return nil, err
	}
	if !verify {
		return list, err
	}

	addrList := make([]*types.Address, 0, len(list))
	for _, addrInfo := range list {
		has, err := addressService.walletClient.WalletHas(ctx, addrInfo.WalletName, addrInfo.Addr)
		if err != nil {
			return nil, xerrors.Errorf("call 'WalletHas' failed %v", err)
		}
		if has {
			addrList = append(addrList, addrInfo)
		}
	}
	return addrList, nil
}

func (addressService *AddressService) DeleteAddress(ctx context.Context, walletName string, addr address.Address) (address.Address, error) {
	return addr, addressService.repo.AddressRepo().DelAddress(ctx, walletName, addr)
}

func (addressService *AddressService) ForbiddenAddress(ctx context.Context, walletName string, addr address.Address) (address.Address, error) {
	if err := addressService.repo.AddressRepo().UpdateState(ctx, walletName, addr, types.Forbiden); err != nil {
		return address.Undef, err
	}
	addressService.log.Infof("forbidden address %v", addr.String())

	return addr, nil
}

func (addressService *AddressService) ActiveAddress(ctx context.Context, walletName string, addr address.Address) (address.Address, error) {
	if err := addressService.repo.AddressRepo().UpdateState(ctx, walletName, addr, types.Alive); err != nil {
		return address.Undef, err
	}
	addressService.log.Infof("active address %v", addr.String())

	return addr, nil
}

func (addressService *AddressService) SetSelectMsgNum(ctx context.Context, walletName string, addr address.Address, num uint64) (address.Address, error) {
	if err := addressService.repo.AddressRepo().UpdateSelectMsgNum(ctx, walletName, addr, num); err != nil {
		return addr, err
	}
	addressService.log.Infof("set select msg num: %s %s %d", walletName, addr.String(), num)

	return addr, nil
}

func (addressService *AddressService) Addresses() map[address.Address]struct{} {
	addrs := make(map[address.Address]struct{})
	addrList, err := addressService.listAddress(context.Background(), false)
	if err != nil {
		addressService.log.Errorf("list address %v", err)
		return addrs
	}

	for _, addr := range addrList {
		addrs[addr.Addr] = struct{}{}
	}

	return addrs
}
