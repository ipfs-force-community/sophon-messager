package sqlite

import (
	"time"

	"github.com/hunjixin/automapper"

	"gorm.io/gorm"

	"github.com/filecoin-project/venus-messager/models/repo"

	"github.com/filecoin-project/venus-messager/types"
)

type sqliteFeeConfig struct {
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

func (sqlFeeConfig *sqliteFeeConfig) TableName() string {
	return "fee_config"
}

type sqliteFeeConfigRepo struct {
	*gorm.DB
}

func newSqliteFeeConfigRepo(db *gorm.DB) *sqliteFeeConfigRepo {
	return &sqliteFeeConfigRepo{db}
}

func fromFeeConfig(fc *types.FeeConfig) *sqliteFeeConfig {
	return automapper.MustMapper(fc, TSqliteFeeConfig).(*sqliteFeeConfig)
}

func feeConfig(sfc sqliteFeeConfig) *types.FeeConfig {
	return automapper.MustMapper(&sfc, TFeeConfig).(*types.FeeConfig)
}

func (sfc *sqliteFeeConfigRepo) SaveFeeConfig(fc *types.FeeConfig) error {
	return sfc.Save(fromFeeConfig(fc)).Error
}

func (sfc *sqliteFeeConfigRepo) GetFeeConfig(walletID types.UUID, methodType int64) (*types.FeeConfig, error) {
	var fc sqliteFeeConfig
	if err := sfc.Take(&fc, "wallet_id = ? and method_type = ? and is_deleted = -1", walletID, methodType).Error; err != nil {
		return nil, err
	}

	return feeConfig(fc), nil
}

func (sfc *sqliteFeeConfigRepo) GetGlobalFeeConfig() (*types.FeeConfig, error) {
	var fc sqliteFeeConfig
	if err := sfc.Take(&fc, "id = ? and wallet_id = ? and method_type = ? and is_deleted = -1", types.EmptyUUID, types.EmptyUUID, -1).Error; err != nil {
		return nil, err
	}

	return feeConfig(fc), nil
}

func (sfc *sqliteFeeConfigRepo) GetWalletFeeConfig(walletID types.UUID) (*types.FeeConfig, error) {
	var fc sqliteFeeConfig
	if err := sfc.Take(&fc, "wallet_id = ? and method_type = ? and is_deleted = -1", walletID, -1).Error; err != nil {
		return nil, err
	}

	return feeConfig(fc), nil
}

func (sfc *sqliteFeeConfigRepo) HasFeeConfig(walletID types.UUID, methodType int64) (bool, error) {
	var count int64
	if err := sfc.Model((*sqliteFeeConfig)(nil)).Where("wallet_id = ? and method_type = ? and is_deleted = -1", walletID, methodType).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (sfc *sqliteFeeConfigRepo) ListFeeConfig() ([]*types.FeeConfig, error) {
	var sfcList []sqliteFeeConfig
	if err := sfc.Find(&sfcList, "is_deleted = -1").Error; err != nil {
		return nil, err
	}

	fcList := make([]*types.FeeConfig, 0, len(sfcList))
	for _, fc := range sfcList {
		fcList = append(fcList, feeConfig(fc))
	}

	return fcList, nil
}

func (sfc *sqliteFeeConfigRepo) DeleteFeeConfig(walletID types.UUID, methodType int64) error {
	return sfc.Model((*sqliteFeeConfig)(nil)).Where("wallet_id = ? and method_type = ? and is_deleted = -1", walletID, methodType).
		UpdateColumns(map[string]interface{}{"is_deleted": repo.Deleted}).Error
}

var _ repo.FeeConfigRepo = (*sqliteFeeConfigRepo)(nil)
