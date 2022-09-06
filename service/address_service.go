package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/big"

	"github.com/filecoin-project/venus-messager/log"
	"github.com/filecoin-project/venus-messager/models/repo"

	"github.com/filecoin-project/venus/venus-shared/api/gateway/v1"
	venusTypes "github.com/filecoin-project/venus/venus-shared/types"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
)

var errAddressNotExists = errors.New("address not exists")

type AddressService struct {
	repo         repo.Repo
	log          *log.Logger
	walletClient gateway.IWalletClient
}

func NewAddressService(repo repo.Repo, logger *log.Logger, walletClient gateway.IWalletClient) *AddressService {
	addressService := &AddressService{
		repo: repo,
		log:  logger,

		walletClient: walletClient,
	}

	return addressService
}

func (addressService *AddressService) SaveAddress(ctx context.Context, address *types.Address) (venusTypes.UUID, error) {
	err := addressService.repo.Transaction(func(txRepo repo.TxRepo) error {
		has, err := txRepo.AddressRepo().HasAddress(ctx, address.Addr)
		if err != nil {
			return err
		}
		if has {
			return fmt.Errorf("address already exists")
		}
		return txRepo.AddressRepo().SaveAddress(ctx, address)
	})

	return address.ID, err
}

func (addressService *AddressService) UpdateNonce(ctx context.Context, addr address.Address, nonce uint64) error {
	return addressService.repo.AddressRepo().UpdateNonce(ctx, addr, nonce)
}

func (addressService *AddressService) GetAddress(ctx context.Context, addr address.Address) (*types.Address, error) {
	return addressService.repo.AddressRepo().GetAddress(ctx, addr)
}

func (addressService *AddressService) WalletHas(ctx context.Context, account string, addr address.Address) (bool, error) {
	return addressService.walletClient.WalletHas(ctx, account, addr)
}

func (addressService *AddressService) HasAddress(ctx context.Context, addr address.Address) (bool, error) {
	return addressService.repo.AddressRepo().HasAddress(ctx, addr)
}

func (addressService *AddressService) ListAddress(ctx context.Context) ([]*types.Address, error) {
	return addressService.repo.AddressRepo().ListAddress(ctx)
}

func (addressService *AddressService) ListActiveAddress(ctx context.Context) ([]*types.Address, error) {
	return addressService.repo.AddressRepo().ListActiveAddress(ctx)
}

func (addressService *AddressService) DeleteAddress(ctx context.Context, addr address.Address) error {
	return addressService.repo.AddressRepo().DelAddress(ctx, addr)
}

func (addressService *AddressService) ForbiddenAddress(ctx context.Context, addr address.Address) error {
	if err := addressService.repo.AddressRepo().UpdateState(ctx, addr, types.AddressStateForbbiden); err != nil {
		return err
	}
	addressService.log.Infof("forbidden address %v success", addr.String())

	return nil
}

func (addressService *AddressService) ActiveAddress(ctx context.Context, addr address.Address) error {
	if err := addressService.repo.AddressRepo().UpdateState(ctx, addr, types.AddressStateAlive); err != nil {
		return err
	}
	addressService.log.Infof("active address %v success", addr.String())

	return nil
}

func (addressService *AddressService) SetSelectMsgNum(ctx context.Context, addr address.Address, num uint64) error {
	if err := addressService.repo.AddressRepo().UpdateSelectMsgNum(ctx, addr, num); err != nil {
		return err
	}
	addressService.log.Infof("set select msg num: %s %d", addr.String(), num)

	return nil
}

func (addressService *AddressService) SetFeeParams(ctx context.Context, params *types.AddressSpec) error {
	has, err := addressService.repo.AddressRepo().HasAddress(ctx, params.Address)
	if err != nil {
		return err
	}
	if !has {
		return errAddressNotExists
	}
	var maxFee, gasFeeCap, baseFee big.Int

	if len(params.MaxFeeStr) != 0 {
		maxFee, err = venusTypes.BigFromString(params.MaxFeeStr)
		if err != nil {
			return fmt.Errorf("parsing maxfee failed %v", err)
		}
	}
	if len(params.GasFeeCapStr) != 0 {
		gasFeeCap, err = venusTypes.BigFromString(params.GasFeeCapStr)
		if err != nil {
			return fmt.Errorf("parsing gas-feecap failed %v", err)
		}
	}
	if len(params.BaseFeeStr) != 0 {
		baseFee, err = venusTypes.BigFromString(params.BaseFeeStr)
		if err != nil {
			return fmt.Errorf("parsing basefee failed %v", err)
		}
	}

	return addressService.repo.AddressRepo().UpdateFeeParams(ctx, params.Address, params.GasOverEstimation, params.GasOverPremium, maxFee, gasFeeCap, baseFee)
}

func (addressService *AddressService) ActiveAddresses(ctx context.Context) map[address.Address]struct{} {
	addrs := make(map[address.Address]struct{})
	addrList, err := addressService.ListActiveAddress(ctx)
	if err != nil {
		addressService.log.Errorf("list address %v", err)
		return addrs
	}

	for _, addr := range addrList {
		addrs[addr.Addr] = struct{}{}
	}

	return addrs
}
