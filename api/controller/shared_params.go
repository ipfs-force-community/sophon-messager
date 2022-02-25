package controller

import (
	"context"

	"github.com/filecoin-project/venus-messager/service"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
)

type SharedParamsCtrl struct {
	BaseController
	SharedParamsService *service.SharedParamsService
}

func (spc SharedParamsCtrl) GetSharedParams(ctx context.Context) (*types.SharedSpec, error) {
	return spc.SharedParamsService.GetSharedParams(ctx)
}

func (spc SharedParamsCtrl) SetSharedParams(ctx context.Context, params *types.SharedSpec) error {
	return spc.SharedParamsService.SetSharedParams(ctx, params)
}

func (spc SharedParamsCtrl) RefreshSharedParams(ctx context.Context) error {
	return spc.SharedParamsService.RefreshSharedParams(ctx)
}
