package controller

import (
	"context"

	"github.com/filecoin-project/go-address"

	"github.com/filecoin-project/venus-messager/service"
	"github.com/filecoin-project/venus-messager/types"
)

type Address struct {
	BaseController
	AddressService *service.AddressService
}

func (a Address) SaveAddress(ctx context.Context, address *types.Address) (types.UUID, error) {
	return a.AddressService.SaveAddress(ctx, address)
}

func (a Address) GetAddress(ctx context.Context, walletName string, addr address.Address) (*types.Address, error) {
	return a.AddressService.GetAddress(ctx, walletName, addr)
}

func (a Address) HasAddress(ctx context.Context, walletName string, addr address.Address) (bool, error) {
	return a.AddressService.HasAddress(ctx, walletName, addr)
}

func (a *Address) ListAddress(ctx context.Context) ([]*types.Address, error) {
	return a.AddressService.ListAddress(ctx)
}

func (a Address) UpdateNonce(ctx context.Context, addr address.Address, nonce uint64) (address.Address, error) {
	return a.AddressService.UpdateNonce(ctx, addr, nonce)
}

func (a Address) DeleteAddress(ctx context.Context, walletName string, addr address.Address) (address.Address, error) {
	return a.AddressService.DeleteAddress(ctx, walletName, addr)
}

func (a Address) ForbiddenAddress(ctx context.Context, walletName string, addr address.Address) (address.Address, error) {
	return a.AddressService.ForbiddenAddress(ctx, walletName, addr)
}

func (a Address) ActiveAddress(ctx context.Context, walletName string, addr address.Address) (address.Address, error) {
	return a.AddressService.ActiveAddress(ctx, walletName, addr)
}

func (a Address) SetSelectMsgNum(ctx context.Context, walletName string, addr address.Address, num uint64) (address.Address, error) {
	return a.AddressService.SetSelectMsgNum(ctx, walletName, addr, num)
}
