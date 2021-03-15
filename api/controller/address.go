package controller

import (
	"context"

	"github.com/ipfs-force-community/venus-messager/service"
	"github.com/ipfs-force-community/venus-messager/types"
)

type Address struct {
	BaseController
	AddressService *service.AddressService
}

func (a Address) SaveAddress(ctx context.Context, address *types.Address) (types.UUID, error) {
	return a.AddressService.SaveAddress(ctx, address)
}

func (a Address) GetAddress(ctx context.Context, addr string) (*types.Address, error) {
	return a.AddressService.GetAddress(ctx, addr)
}

func (a Address) ListAddress(ctx context.Context) ([]*types.Address, error) {
	return a.AddressService.ListAddress(ctx)
}
