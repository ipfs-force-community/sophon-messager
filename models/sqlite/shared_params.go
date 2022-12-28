package sqlite

import (
	"context"

	"github.com/filecoin-project/go-state-types/big"

	"gorm.io/gorm"

	"github.com/filecoin-project/venus-messager/models/mtypes"
	"github.com/filecoin-project/venus-messager/models/repo"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
)

type sqliteSharedParams struct {
	ID        uint   `gorm:"primary_key;column:id;type:INT unsigned AUTO_INCREMENT;NOT NULL" json:"id"`
	SelMsgNum uint64 `gorm:"column:sel_msg_num;type:unsigned bigint;NOT NULL"`
	FeeSpec
}

func fromSharedParams(sp types.SharedSpec) *sqliteSharedParams {
	return &sqliteSharedParams{
		ID:        sp.ID,
		SelMsgNum: sp.SelMsgNum,
		FeeSpec: FeeSpec{
			BaseFee:           mtypes.SafeFromGo(sp.BaseFee.Int),
			GasOverEstimation: sp.GasOverEstimation,
			MaxFee:            mtypes.SafeFromGo(sp.MaxFee.Int),
			GasFeeCap:         mtypes.SafeFromGo(sp.GasFeeCap.Int),
			GasOverPremium:    sp.GasOverPremium,
		},
	}
}

func (ssp sqliteSharedParams) SharedParams() *types.SharedSpec {
	return &types.SharedSpec{
		ID:        ssp.ID,
		SelMsgNum: ssp.SelMsgNum,
		FeeSpec: types.FeeSpec{
			GasOverEstimation: ssp.GasOverEstimation,
			MaxFee:            big.Int(mtypes.SafeFromGo(ssp.MaxFee.Int)),
			GasFeeCap:         big.Int(mtypes.SafeFromGo(ssp.GasFeeCap.Int)),
			BaseFee:           big.Int(mtypes.SafeFromGo(ssp.BaseFee.Int)),
			GasOverPremium:    ssp.GasOverPremium,
		},
	}
}

func (ssp sqliteSharedParams) TableName() string {
	return "shared_params"
}

var _ repo.SharedParamsRepo = (*sqliteSharedParamsRepo)(nil)

type sqliteSharedParamsRepo struct {
	*gorm.DB
}

func newSqliteSharedParamsRepo(db *gorm.DB) sqliteSharedParamsRepo {
	return sqliteSharedParamsRepo{DB: db}
}

func (s sqliteSharedParamsRepo) GetSharedParams(ctx context.Context) (*types.SharedSpec, error) {
	var ssp sqliteSharedParams
	if err := s.DB.Take(&ssp).Error; err != nil {
		return nil, err
	}
	return ssp.SharedParams(), nil
}

func (s sqliteSharedParamsRepo) SetSharedParams(ctx context.Context, params *types.SharedSpec) (uint, error) {
	var ssp sqliteSharedParams
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
