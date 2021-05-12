package repo

import "github.com/filecoin-project/venus-messager/types"

type FeeConfigRepo interface {
	SaveFeeConfig(fc *types.FeeConfig) error
	GetFeeConfig(walletID types.UUID, methodType int64) (*types.FeeConfig, error)
	GetGlobalFeeConfig() (*types.FeeConfig, error)
	GetWalletFeeConfig(walletID types.UUID) (*types.FeeConfig, error)
	HasFeeConfig(walletID types.UUID, methodType int64) (bool, error)
	ListFeeConfig() ([]*types.FeeConfig, error)
	DeleteFeeConfig(walletID types.UUID, methodType int64) error
}
