package sqlite

import (
	"context"

	"gorm.io/gorm"

	"github.com/hunjixin/automapper"

	"github.com/filecoin-project/go-state-types/abi"

	"github.com/ipfs-force-community/venus-messager/models/repo"
	"github.com/ipfs-force-community/venus-messager/types"
)

type sqliteSharedParams struct {
	ID uint `gorm:"primary_key;column:id;type:INT unsigned AUTO_INCREMENT;NOT NULL" json:"id"`

	ExpireEpoch       abi.ChainEpoch `gorm:"column:expire_epoch;type:INT;NOT NULL"`
	GasOverEstimation float64        `gorm:"column:gas_over_estimation;type:REAL;NOT NULL"`
	MaxFee            int64          `gorm:"column:max_fee;type:UNSIGNED BIG INT;NOT NULL"`
	MaxFeeCap         int64          `gorm:"column:max_fee_cap;type:UNSIGNED BIG INT;NOT NULL"`

	SelMsgNum uint64 `gorm:"column:sel_msg_num;type:UNSIGNED BIG INT;NOT NULL"`

	ScanInterval int `gorm:"column:scan_interval;NOT NULL"`

	MaxEstFailNumOfMsg uint64 `gorm:"column:max_ext_fail_num_of_msg;type:UNSIGNED BIG INT;NOT NULL"`
}

func FromSharedParams(sp types.SharedParams) *sqliteSharedParams {
	return automapper.MustMapper(&sp, TSqliteSharedParams).(*sqliteSharedParams)
}

func (ssp sqliteSharedParams) SharedParams() *types.SharedParams {
	return automapper.MustMapper(&ssp, TSharedParams).(*types.SharedParams)
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

func (s sqliteSharedParamsRepo) GetSharedParams(ctx context.Context) (*types.SharedParams, error) {
	var ssp sqliteSharedParams
	if err := s.DB.Take(&ssp).Error; err != nil {
		return nil, err
	}
	return ssp.SharedParams(), nil
}

func (s sqliteSharedParamsRepo) SetSharedParams(ctx context.Context, params *types.SharedParams) error {
	var ssp sqliteSharedParams
	if err := s.DB.Where("id = ?", 1).Take(&ssp).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			if params.ID == 0 {
				params.ID = 1
			}
			if err := s.DB.Save(FromSharedParams(*params)).Error; err != nil {
				return err
			}
			return nil
		}
		return err
	}

	ssp.ExpireEpoch = params.ExpireEpoch
	ssp.GasOverEstimation = params.GasOverEstimation
	ssp.MaxFeeCap = params.MaxFeeCap
	ssp.MaxFee = params.MaxFee

	ssp.SelMsgNum = params.SelMsgNum

	ssp.ScanInterval = params.ScanInterval

	ssp.MaxEstFailNumOfMsg = params.MaxEstFailNumOfMsg

	if err := s.DB.Save(&ssp).Error; err != nil {
		return err
	}

	return nil
}
