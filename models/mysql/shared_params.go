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
	ID uint `gorm:"primary_key;column:id;type:SMALLINT(2) unsigned AUTO_INCREMENT;NOT NULL"`

	SelectSpec
}

func fromSharedParams(sp types.SharedSpec) *mysqlSharedParams {
	return &mysqlSharedParams{
		ID: sp.ID,
		SelectSpec: SelectSpec{
			BaseFee:           mtypes.SafeFromGo(sp.BaseFee.Int),
			SelMsgNum:         sp.SelMsgNum,
			GasOverEstimation: sp.GasOverEstimation,
			MaxFee:            mtypes.SafeFromGo(sp.MaxFee.Int),
			GasFeeCap:         mtypes.SafeFromGo(sp.GasFeeCap.Int),
			GasOverPremium:    sp.GasOverPremium,
		},
	}
}

func (ssp mysqlSharedParams) SharedParams() *types.SharedSpec {
	return &types.SharedSpec{
		ID: ssp.ID,
		SelectSpec: types.SelectSpec{
			GasOverEstimation: ssp.GasOverEstimation,
			MaxFee:            big.Int(mtypes.SafeFromGo(ssp.MaxFee.Int)),
			GasFeeCap:         big.Int(mtypes.SafeFromGo(ssp.GasFeeCap.Int)),
			BaseFee:           big.Int(mtypes.SafeFromGo(ssp.BaseFee.Int)),
			GasOverPremium:    ssp.GasOverPremium,
			SelMsgNum:         ssp.SelMsgNum,
		},
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
	// make sure ID is 1
	params.ID = 1
	if err := s.DB.Where("id = ?", 1).Take(&ssp).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			if err := s.DB.Save(fromSharedParams(*params)).Error; err != nil {
				return 0, err
			}
			return params.ID, nil
		}
		return 0, err
	}

	if err := s.DB.Save(fromSharedParams(*params)).Error; err != nil {
		return 0, err
	}

	return params.ID, nil
}
