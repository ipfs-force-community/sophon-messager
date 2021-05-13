package mysql

import (
	"time"

	"github.com/hunjixin/automapper"

	"gorm.io/gorm"

	"github.com/filecoin-project/venus-messager/models/repo"

	"github.com/filecoin-project/venus-messager/types"
)

type mysqlFeeConfig struct {
	ID types.UUID `gorm:"column:id;type:varchar(256);primary_key;"`

	WalletID          types.UUID `gorm:"column:wallet_id;type:varchar(256);NOT NULL"`
	MethodType        int64      `gorm:"column:method_type;type:bigint;NOT NULL"`
	GasOverEstimation float64    `gorm:"column:gas_over_estimation;type:decimal(10,2);"`
	MaxFee            types.Int  `gorm:"column:max_fee;type:varchar(256);"`
	MaxFeeCap         types.Int  `gorm:"column:max_fee_cap;type:varchar(256);"`

	IsDeleted int       `gorm:"column:is_deleted;index;default:-1;NOT NULL"`
	CreatedAt time.Time `gorm:"column:created_at;index;NOT NULL"`
	UpdatedAt time.Time `gorm:"column:updated_at;index;NOT NULL"`
}

func (mysqlFeeConfig *mysqlFeeConfig) TableName() string {
	return "fee_config"
}

type mysqlFeeConfigRepo struct {
	*gorm.DB
}

func newMysqlFeeConfigRepo(db *gorm.DB) *mysqlFeeConfigRepo {
	return &mysqlFeeConfigRepo{db}
}

func fromFeeConfig(fc *types.FeeConfig) *mysqlFeeConfig {
	return automapper.MustMapper(fc, TMysqlFeeConfig).(*mysqlFeeConfig)
}

func feeConfig(sfc mysqlFeeConfig) *types.FeeConfig {
	return automapper.MustMapper(&sfc, TFeeConfig).(*types.FeeConfig)
}

func (sfc *mysqlFeeConfigRepo) SaveFeeConfig(fc *types.FeeConfig) error {
	return sfc.Save(fromFeeConfig(fc)).Error
}

func (sfc *mysqlFeeConfigRepo) GetFeeConfig(walletID types.UUID, methodType int64) (*types.FeeConfig, error) {
	var fc mysqlFeeConfig
	if err := sfc.Take(&fc, "wallet_id = ? and method_type = ? and is_deleted = -1", walletID, methodType).Error; err != nil {
		return nil, err
	}

	return feeConfig(fc), nil
}

func (sfc *mysqlFeeConfigRepo) GetGlobalFeeConfig() (*types.FeeConfig, error) {
	var fc mysqlFeeConfig
	if err := sfc.Take(&fc, "id = ? and wallet_id = ? and method_type = ? and is_deleted = -1", types.DefGlobalFeeCfgID, types.UUID{}, -1).Error; err != nil {
		return nil, err
	}

	return feeConfig(fc), nil
}

func (sfc *mysqlFeeConfigRepo) GetWalletFeeConfig(walletID types.UUID) (*types.FeeConfig, error) {
	var fc mysqlFeeConfig
	if err := sfc.Take(&fc, "wallet_id = ? and method_type = ? and is_deleted = -1", walletID, -1).Error; err != nil {
		return nil, err
	}

	return feeConfig(fc), nil
}

func (sfc *mysqlFeeConfigRepo) HasFeeConfig(walletID types.UUID, methodType int64) (bool, error) {
	var count int64
	if err := sfc.Model((*mysqlFeeConfig)(nil)).Where("wallet_id = ? and method_type = ? and is_deleted = -1", walletID, methodType).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (sfc *mysqlFeeConfigRepo) ListFeeConfig() ([]*types.FeeConfig, error) {
	var sfcList []mysqlFeeConfig
	if err := sfc.Find(&sfcList, "is_deleted = -1").Error; err != nil {
		return nil, err
	}

	fcList := make([]*types.FeeConfig, 0, len(sfcList))
	for _, fc := range sfcList {
		fcList = append(fcList, feeConfig(fc))
	}

	return fcList, nil
}

func (sfc *mysqlFeeConfigRepo) DeleteFeeConfig(walletID types.UUID, methodType int64) error {
	return sfc.Model((*mysqlFeeConfig)(nil)).Where("wallet_id = ? and method_type = ? and is_deleted = -1", walletID, methodType).
		Update("is_deleted", repo.Deleted).Error
}

var _ repo.FeeConfigRepo = (*mysqlFeeConfigRepo)(nil)
