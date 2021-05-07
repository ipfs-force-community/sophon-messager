package repo

import (
	"context"

	"github.com/filecoin-project/venus-messager/types"
)

type SharedParamsRepo interface {
	GetSharedParams(ctx context.Context) (*types.SharedParams, error)
	SetSharedParams(ctx context.Context, params *types.SharedParams) (uint, error)
}
