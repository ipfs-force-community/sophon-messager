package mysql

import (
	"context"

	"github.com/filecoin-project/go-state-types/big"

	"gorm.io/gorm"

	"github.com/filecoin-project/venus-messager/models/mtypes"
	"github.com/filecoin-project/venus-messager/models/repo"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
)

type mysqlSharedParams struct {
	ID uint `gorm:"primary_key;column:id;type:SMALLINT(2) unsigned AUTO_INCREMENT;NOT NULL" json:"id"`

	GasOverEstimation float64    `gorm:"column:gas_over_estimation;type:DOUBLE;NOT NULL"`
	MaxFee            mtypes.Int `gorm:"column:max_fee;type:varchar(256);NOT NULL"`
	MaxFeeCap         mtypes.Int `gorm:"column:max_fee_cap;type:varchar(256);NOT NULL"`
	GasOverPremium    float64    `gorm:"column:gas_over_premium;type:DOUBLE;NOT NULL;default:0;"`
	SelMsgNum         uint64     `gorm:"column:sel_msg_num;type:BIGINT(20) UNSIGNED;NOT NULL"`
}

func fromSharedParams(sp types.SharedSpec) *mysqlSharedParams {
	return &mysqlSharedParams{
		ID:                sp.ID,
		GasOverEstimation: sp.GasOverEstimation,
		MaxFee:            mtypes.Int{Int: sp.MaxFee.Int},
		MaxFeeCap:         mtypes.Int{Int: sp.MaxFeeCap.Int},
		GasOverPremium:    sp.GasOverPremium,
		SelMsgNum:         sp.SelMsgNum,
	}
}

func (ssp mysqlSharedParams) SharedParams() *types.SharedSpec {
	return &types.SharedSpec{
		ID:                ssp.ID,
		GasOverEstimation: ssp.GasOverEstimation,
		MaxFee:            big.Int(mtypes.SafeFromGo(ssp.MaxFee.Int)),
		MaxFeeCap:         big.Int(mtypes.SafeFromGo(ssp.MaxFeeCap.Int)),
		GasOverPremium:    ssp.GasOverPremium,
		SelMsgNum:         ssp.SelMsgNum,
	}
}

func (ssp mysqlSharedParams) TableName() string {
	return "shared_params"
}

var _ repo.SharedParamsRepo = (*mysqlSharedParamsRepo)(nil)

type mysqlSharedParamsRepo struct {
	*gorm.DB
}

func newMysqlSharedParamsRepo(db *gorm.DB) mysqlSharedParamsRepo {
	return mysqlSharedParamsRepo{DB: db}
}

func (s mysqlSharedParamsRepo) GetSharedParams(ctx context.Context) (*types.SharedSpec, error) {
	var ssp mysqlSharedParams
	if err := s.DB.Take(&ssp).Error; err != nil {
		return nil, err
	}
	return ssp.SharedParams(), nil
}

func (s mysqlSharedParamsRepo) SetSharedParams(ctx context.Context, params *types.SharedSpec) (uint, error) {
	var ssp mysqlSharedParams
	if err := s.DB.Where("id = ?", 1).Take(&ssp).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			if params.ID == 0 {
				params.ID = 1
			}
			if err := s.DB.Save(fromSharedParams(*params)).Error; err != nil {
				return 0, err
			}
			return params.ID, nil
		}
		return 0, err
	}

	ssp.GasOverEstimation = params.GasOverEstimation
	ssp.MaxFeeCap = mtypes.Int{Int: params.MaxFeeCap.Int}
	ssp.MaxFee = mtypes.Int{Int: params.MaxFee.Int}
	ssp.GasOverPremium = params.GasOverPremium

	ssp.SelMsgNum = params.SelMsgNum

	if err := s.DB.Save(&ssp).Error; err != nil {
		return 0, err
	}

	return params.ID, nil
}
