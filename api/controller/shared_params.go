package controller

import (
	"context"

	"github.com/filecoin-project/venus-messager/service"
	"github.com/filecoin-project/venus-messager/types"
)

type SharedParamsCtrl struct {
	BaseController
	SharedParamsService *service.SharedParamsService
}

func (spc SharedParamsCtrl) GetSharedParams(ctx context.Context) (*types.SharedParams, error) {
	return spc.SharedParamsService.GetSharedParams(ctx)
}

func (spc SharedParamsCtrl) SetSharedParams(ctx context.Context, params *types.SharedParams) (struct{}, error) {
	return spc.SharedParamsService.SetSharedParams(ctx, params)
}

func (spc SharedParamsCtrl) RefreshSharedParams(ctx context.Context) (struct{}, error) {
	return spc.SharedParamsService.RefreshSharedParams(ctx)
}
