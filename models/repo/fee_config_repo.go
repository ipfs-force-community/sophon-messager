package repo

import "github.com/filecoin-project/venus-messager/types"

type FeeConfigRepo interface {
	SaveFeeConfig(fc *types.FeeConfig) error
	GetFeeConfig(walletID types.UUID, method int64) (*types.FeeConfig, error)
	GetGlobalFeeConfig() (*types.FeeConfig, error)
	GetWalletFeeConfig(walletID types.UUID) (*types.FeeConfig, error)
	HasFeeConfig(walletID types.UUID, method int64) (bool, error)
	ListFeeConfig() ([]*types.FeeConfig, error)
	DeleteFeeConfig(walletID types.UUID, method int64) error
}
