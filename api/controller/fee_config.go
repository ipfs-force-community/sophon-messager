package controller

import (
	"context"

	"github.com/filecoin-project/venus-messager/service"
	"github.com/filecoin-project/venus-messager/types"
)

type FeeConfig struct {
	BaseController
	FeeConfigService *service.FeeConfigService
}

func (fc FeeConfig) SaveFeeConfig(ctx context.Context, feeConfig *types.FeeConfig) (types.UUID, error) {
	return fc.FeeConfigService.SaveFeeConfig(ctx, feeConfig)
}

func (fc FeeConfig) GetFeeConfig(ctx context.Context, walletID types.UUID, methodType int64) (*types.FeeConfig, error) {
	return fc.FeeConfigService.GetFeeConfig(ctx, walletID, methodType)
}

func (fc FeeConfig) GetWalletFeeConfig(ctx context.Context, walletID types.UUID) (*types.FeeConfig, error) {
	return fc.FeeConfigService.GetWalletFeeConfig(ctx, walletID)
}

func (fc FeeConfig) GetGlobalFeeConfig(ctx context.Context) (*types.FeeConfig, error) {
	return fc.FeeConfigService.GetGlobalFeeConfig(ctx)
}

func (fc FeeConfig) ListFeeConfig(ctx context.Context) ([]*types.FeeConfig, error) {
	return fc.FeeConfigService.ListFeeConfig(ctx)
}

func (fc FeeConfig) HasFeeConfig(ctx context.Context, walletID types.UUID, methodType int64) (bool, error) {
	return fc.FeeConfigService.HasFeeConfig(ctx, walletID, methodType)
}

func (fc FeeConfig) DeleteFeeConfig(ctx context.Context, walletID types.UUID, methodType int64) (types.UUID, error) {
	return fc.FeeConfigService.DeleteFeeConfig(ctx, walletID, methodType)
}
