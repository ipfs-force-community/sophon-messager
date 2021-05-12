package controller

import (
	"github.com/filecoin-project/venus-messager/service"
	"github.com/filecoin-project/venus-messager/types"
)

type FeeConfig struct {
	BaseController
	FeeConfigService *service.FeeConfigService
}

func (fc FeeConfig) SaveFeeConfig(feeConfig *types.FeeConfig) (types.UUID, error) {
	return fc.FeeConfigService.SaveFeeConfig(feeConfig)
}

func (fc FeeConfig) GetFeeConfig(walletID types.UUID, methodType int64) (*types.FeeConfig, error) {
	return fc.FeeConfigService.GetFeeConfig(walletID, methodType)
}

func (fc FeeConfig) GetWalletFeeConfig(walletID types.UUID) (*types.FeeConfig, error) {
	return fc.FeeConfigService.GetWalletFeeConfig(walletID)
}

func (fc FeeConfig) GetGlobalFeeConfig() (*types.FeeConfig, error) {
	return fc.FeeConfigService.GetGlobalFeeConfig()
}

func (fc FeeConfig) ListFeeConfig() ([]*types.FeeConfig, error) {
	return fc.FeeConfigService.ListFeeConfig()
}

func (fc FeeConfig) HasFeeConfig(walletID types.UUID, methodType int64) (bool, error) {
	return fc.FeeConfigService.HasFeeConfig(walletID, methodType)
}

func (fc FeeConfig) DeleteFeeConfig(walletID types.UUID, methodType int64) error {
	return fc.FeeConfigService.DeleteFeeConfig(walletID, methodType)
}
