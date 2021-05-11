package repo

import "github.com/filecoin-project/venus-messager/types"

type FeeConfigRepo interface {
	SaveFeeConfig(fc *types.FeeConfig) error
	GetFeeConfig(walletID types.UUID, methodType uint64) (*types.FeeConfig, error)
	HasFeeConfig(walletID types.UUID, methodType uint64) (bool, error)
	ListFeeConfig() ([]*types.FeeConfig, error)
	DeleteFeeConfig(walletID types.UUID, methodType uint64) error
}
