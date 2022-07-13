package repo

import (
	"context"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/big"
	shared "github.com/filecoin-project/venus/venus-shared/types"

	types "github.com/filecoin-project/venus/venus-shared/types/messager"
)

type AddressRepo interface {
	SaveAddress(ctx context.Context, address *types.Address) error
	GetAddress(ctx context.Context, addr address.Address) (*types.Address, error)
	GetAddressByID(ctx context.Context, id shared.UUID) (*types.Address, error)
	GetOneRecord(ctx context.Context, addr address.Address) (*types.Address, error)
	HasAddress(ctx context.Context, addr address.Address) (bool, error)
	ListAddress(ctx context.Context) ([]*types.Address, error)
	ListActiveAddress(ctx context.Context) ([]*types.Address, error)
	DelAddress(ctx context.Context, addr address.Address) error
	UpdateNonce(ctx context.Context, addr address.Address, nonce uint64) error
	UpdateState(ctx context.Context, addr address.Address, state types.AddressState) error
	UpdateSelectMsgNum(ctx context.Context, addr address.Address, num uint64) error
	UpdateFeeParams(ctx context.Context, addr address.Address, gasOverEstimation, gasOverPremium float64, maxFee, maxFeeCap big.Int) error
}
