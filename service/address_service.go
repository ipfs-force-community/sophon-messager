package service

import (
	"context"

	"github.com/filecoin-project/go-address"
	"github.com/sirupsen/logrus"

	"github.com/filecoin-project/venus-messager/models/repo"
	"github.com/filecoin-project/venus-messager/types"
)

type AddressService struct {
	repo repo.Repo
	log  *logrus.Logger
}

type AddressInfo struct {
	State        types.State
	SelectMsgNum uint64
	WalletClient IWalletClient
}

func NewAddressService(repo repo.Repo, logger *logrus.Logger) *AddressService {
	addressService := &AddressService{
		repo: repo,
		log:  logger,
	}

	return addressService
}

func (addressService *AddressService) SaveAddress(ctx context.Context, address *types.Address) (types.UUID, error) {
	err := addressService.repo.Transaction(func(txRepo repo.TxRepo) error {
		has, err := txRepo.AddressRepo().HasAddress(ctx, address.Addr)
		if err != nil {
			return err
		}
		if has {
			srcAddr, err := txRepo.AddressRepo().GetOneRecord(ctx, address.Addr)
			if err != nil {
				return err
			}
			address.ID = srcAddr.ID
			address.CreatedAt = srcAddr.CreatedAt
			address.Weight = srcAddr.Weight

			if srcAddr.IsDeleted == repo.NotDeleted {
				return ErrRecordExist
			}
		}
		return txRepo.AddressRepo().SaveAddress(ctx, address)
	})

	return address.ID, err
}

func (addressService *AddressService) UpdateAddress(ctx context.Context, address *types.Address) error {
	return addressService.repo.AddressRepo().UpdateAddress(ctx, address)
}

func (addressService *AddressService) UpdateNonce(ctx context.Context, addr address.Address, nonce uint64) (address.Address, error) {
	return addr, addressService.repo.AddressRepo().UpdateNonce(ctx, addr, nonce)
}

func (addressService *AddressService) GetAddress(ctx context.Context, addr address.Address) (*types.Address, error) {
	return addressService.repo.AddressRepo().GetAddress(ctx, addr)
}

func (addressService *AddressService) HasAddress(ctx context.Context, addr address.Address) (bool, error) {
	return addressService.repo.AddressRepo().HasAddress(ctx, addr)
}

func (addressService *AddressService) ListAddress(ctx context.Context) ([]*types.Address, error) {
	return addressService.repo.AddressRepo().ListAddress(ctx)
}

func (addressService *AddressService) DeleteAddress(ctx context.Context, addr address.Address) (address.Address, error) {
	return addr, addressService.repo.AddressRepo().DelAddress(ctx, addr)
}
