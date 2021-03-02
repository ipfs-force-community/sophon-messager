package service

import (
	"context"

	"github.com/ipfs-force-community/venus-messager/models/repo"
	"github.com/ipfs-force-community/venus-messager/types"
	"github.com/sirupsen/logrus"
)

type AddressService struct {
	repo repo.Repo
	log  *logrus.Logger
}

func NewAddressService(repo repo.Repo, logger *logrus.Logger) *AddressService {
	return &AddressService{
		repo: repo,
		log:  logger,
	}
}

func (addressService AddressService) SaveAddress(ctx context.Context, address *types.Address) (string, error) {
	return addressService.repo.AddressRepo().SaveAddress(ctx, address)
}

func (addressService AddressService) GetAddress(ctx context.Context, addr string) (*types.Address, error) {
	return addressService.repo.AddressRepo().GetAddress(ctx, addr)
}

func (addressService AddressService) ListAddress(ctx context.Context) ([]*types.Address, error) {
	return addressService.repo.AddressRepo().ListAddress(ctx)
}
