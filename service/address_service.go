package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/big"

	"github.com/filecoin-project/venus-auth/core"
	"github.com/filecoin-project/venus-auth/jwtclient"

	gatewayAPI "github.com/filecoin-project/venus/venus-shared/api/gateway/v2"
	venusTypes "github.com/filecoin-project/venus/venus-shared/types"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"

	"github.com/filecoin-project/venus-messager/models/repo"
)

var errAddressNotExists = errors.New("address not exists")

type IAddressService interface {
	SaveAddress(ctx context.Context, address *types.Address) (venusTypes.UUID, error)
	UpdateNonce(ctx context.Context, addr address.Address, nonce uint64) error
	GetAddress(ctx context.Context, addr address.Address) (*types.Address, error)

	// WalletHas 1. 检查请求token绑定的 account是否在addr绑定的用户列表中,不在返回false; eg. from01的绑定关系: from01-acc01,from01-acc02, 判断: acc[token] IN (acc01,acc02)
	// 2. 调用venus-gateway的 `WalletHas(context.Context, []string, address.Address) (bool, error)` 查找在线的venus-wallet，调用 `.WalletHas(ctx, []string{acc01,acc02}, addr)`
	// venus-gateway: venus-wallet的私钥和所有支持账号都绑定，WalletHas从接口的绑定账号列表中查找是否存在在线的venus-wallet channel.
	WalletHas(ctx context.Context, addr address.Address) (bool, error)
	HasAddress(ctx context.Context, addr address.Address) (bool, error)
	ListAddress(ctx context.Context) ([]*types.Address, error)
	ListActiveAddress(ctx context.Context) ([]*types.Address, error)
	DeleteAddress(ctx context.Context, addr address.Address) error
	ForbiddenAddress(ctx context.Context, addr address.Address) error
	ActiveAddress(ctx context.Context, addr address.Address) error
	SetSelectMsgNum(ctx context.Context, addr address.Address, num uint64) error
	SetFeeParams(ctx context.Context, params *types.AddressSpec) error
	ActiveAddresses(ctx context.Context) map[address.Address]struct{}
	GetAccountsOfSigner(ctx context.Context, addr address.Address) ([]string, error)
}

type AddressService struct {
	repo         repo.Repo
	walletClient gatewayAPI.IWalletClient
	authClient   jwtclient.IAuthClient
}

var _ IAddressService = (*AddressService)(nil)

func NewAddressService(repo repo.Repo, walletClient gatewayAPI.IWalletClient, remoteAuthCli jwtclient.IAuthClient) *AddressService {
	addressService := &AddressService{
		repo: repo,

		walletClient: walletClient,
		authClient:   remoteAuthCli,
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

// WalletHas 1. 检查请求token绑定的 account是否在addr绑定的用户列表中,不在返回false; eg. from01的绑定关系: from01-acc01,from01-acc02, 判断: acc[token] IN (acc01,acc02)
// 2. 调用venus-gateway的 `WalletHas(context.Context, []string, address.Address) (bool, error)` 查找在线的venus-wallet，调用 `.WalletHas(ctx, []string{acc01,acc02}, addr)`
// venus-gateway: venus-wallet的私钥和所有支持账号都绑定，WalletHas从接口的绑定账号列表中查找是否存在在线的venus-wallet channel.
func (addressService *AddressService) WalletHas(ctx context.Context, addr address.Address) (bool, error) {
	accounts, err := addressService.GetAccountsOfSigner(ctx, addr)
	if err != nil {
		return false, err
	}

	bExist := false
	name, _ := core.CtxGetName(ctx)
	for _, account := range accounts {
		if name == account {
			bExist = true
			break
		}
	}
	if !bExist {
		return false, nil
	}

	return addressService.walletClient.WalletHas(ctx, addr, accounts)
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
	log.Infof("forbidden address %v success", addr.String())

	return nil
}

func (addressService *AddressService) ActiveAddress(ctx context.Context, addr address.Address) error {
	if err := addressService.repo.AddressRepo().UpdateState(ctx, addr, types.AddressStateAlive); err != nil {
		return err
	}
	log.Infof("active address %v success", addr.String())

	return nil
}

func (addressService *AddressService) SetSelectMsgNum(ctx context.Context, addr address.Address, num uint64) error {
	if err := addressService.repo.AddressRepo().UpdateSelectMsgNum(ctx, addr, num); err != nil {
		return err
	}
	log.Infof("set select msg num: %s %d", addr.String(), num)

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
		log.Errorf("list address %v", err)
		return addrs
	}

	for _, addr := range addrList {
		addrs[addr.Addr] = struct{}{}
	}

	return addrs
}

func (addressService *AddressService) GetAccountsOfSigner(ctx context.Context, addr address.Address) ([]string, error) {
	users, err := addressService.authClient.GetUserBySigner(ctx, addr)
	if err != nil {
		return nil, err
	}

	accounts := make([]string, 0, len(users))
	for _, user := range users {
		accounts = append(accounts, user.Name)
	}

	return accounts, nil
}
