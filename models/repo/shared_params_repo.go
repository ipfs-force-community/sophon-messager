package repo

import (
	"context"

	types "github.com/filecoin-project/venus/venus-shared/types/messager"
)

type SharedParamsRepo interface {
	GetSharedParams(ctx context.Context) (*types.SharedSpec, error)
	SetSharedParams(ctx context.Context, params *types.SharedSpec) (uint, error)
}
