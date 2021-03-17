package repo

import (
	"context"

	"github.com/filecoin-project/go-address"

	"github.com/ipfs-force-community/venus-messager/types"
)

type AddressRepo interface {
	HasAddress(ctx context.Context, addr address.Address) (bool, error)
	SaveAddress(ctx context.Context, address *types.Address) (types.UUID, error)
	UpdateNonce(ctx context.Context, uuid types.UUID, nonce uint64) (types.UUID, error)
	GetAddress(ctx context.Context, addr string) (*types.Address, error)
	ListAddress(ctx context.Context) ([]*types.Address, error)
	DelAddress(ctx context.Context, addr string) error
}
