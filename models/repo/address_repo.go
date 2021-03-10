package repo

import (
	"context"

	"github.com/ipfs-force-community/venus-messager/types"
)

type AddressRepo interface {
	SaveAddress(ctx context.Context, address *types.Address) (string, error)
	GetAddress(ctx context.Context, addr string) (*types.Address, error)
	ListAddress(ctx context.Context) ([]*types.Address, error)
	DelAddress(ctx context.Context, addr string) error
}
