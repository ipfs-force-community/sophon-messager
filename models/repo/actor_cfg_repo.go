package repo

import (
	"context"

	shared "github.com/filecoin-project/venus/venus-shared/types"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
)

type ActorCfgRepo interface {
	SaveActorCfg(ctx context.Context, address *types.ActorCfg) error
	GetActorCfgByMethodType(ctx context.Context, methodType *types.MethodType) (*types.ActorCfg, error)
	GetActorCfgByID(ctx context.Context, id shared.UUID) (*types.ActorCfg, error)
	ListActorCfg(ctx context.Context) ([]*types.ActorCfg, error)
	DelActorCfgByMethodType(ctx context.Context, addr *types.MethodType) error
	DelActorCfgById(ctx context.Context, id shared.UUID) error
	UpdateSelectSpecById(ctx context.Context, id shared.UUID, spec *types.ChangeGasSpecParams) error
}
