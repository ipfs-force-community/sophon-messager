package mysql

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/hunjixin/automapper"

	"github.com/filecoin-project/go-state-types/abi"

	"github.com/ipfs-force-community/venus-messager/models/repo"
	"github.com/ipfs-force-community/venus-messager/types"
)

type mysqlSharedParams struct {
	ID uint `gorm:"primary_key;column:id;type:SMALLINT(2) unsigned AUTO_INCREMENT;NOT NULL" json:"id"`

	ExpireEpoch       abi.ChainEpoch `gorm:"column:expire_epoch;type:BIGINT(20);NOT NULL"`
	GasOverEstimation float64        `gorm:"column:gas_over_estimation;type:DOUBLE;NOT NULL"`
	MaxFee            int64          `gorm:"column:max_fee;type:BIGINT(20);NOT NULL"`
	MaxFeeCap         int64          `gorm:"column:max_fee_cap;type:BIGINT(20);NOT NULL"`

	SelMsgNum uint64 `gorm:"column:sel_msg_num;type:BIGINT(20) UNSIGNED;NOT NULL"`

	ScanInterval time.Duration `gorm:"column:scan_interval;NOT NULL"`

	MaxEstFailNumOfMsg uint64 `gorm:"column:max_ext_fail_num_of_msg;type:BIGINT(20) UNSIGNED;NOT NULL"`
}

func FromSharedParams(sp types.SharedParams) *mysqlSharedParams {
	return automapper.MustMapper(&sp, TMysqlSharedParams).(*mysqlSharedParams)
}

func (ssp mysqlSharedParams) SharedParams() *types.SharedParams {
	return automapper.MustMapper(&ssp, TSharedParams).(*types.SharedParams)
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

func (s mysqlSharedParamsRepo) GetSharedParams(ctx context.Context) (*types.SharedParams, error) {
	var ssp mysqlSharedParams
	if err := s.DB.Take(&ssp).Error; err != nil {
		return nil, err
	}
	return ssp.SharedParams(), nil
}

func (s mysqlSharedParamsRepo) SetSharedParams(ctx context.Context, params *types.SharedParams) (*types.SharedParams, error) {
	var ssp mysqlSharedParams
	if err := s.DB.Where("id = ?", 1).Take(&ssp).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			if params.ID == 0 {
				params.ID = 1
			}
			if err := s.DB.Save(FromSharedParams(*params)).Error; err != nil {
				return nil, err
			}
			return params, nil
		}
		return nil, err
	}

	ssp.ExpireEpoch = params.ExpireEpoch
	ssp.GasOverEstimation = params.GasOverEstimation
	ssp.MaxFeeCap = params.MaxFeeCap
	ssp.MaxFee = params.MaxFee

	ssp.SelMsgNum = params.SelMsgNum

	ssp.ScanInterval = params.ScanInterval

	ssp.MaxEstFailNumOfMsg = params.MaxEstFailNumOfMsg

	if err := s.DB.Save(&ssp).Error; err != nil {
		return nil, err
	}

	return ssp.SharedParams(), nil
}
