package service

import (
	"context"
	"errors"

	"github.com/filecoin-project/go-state-types/big"
	venusTypes "github.com/filecoin-project/venus/venus-shared/types"
	"gorm.io/gorm"

	"github.com/filecoin-project/venus-messager/models/repo"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
)

var DefaultMaxFee = venusTypes.MustParseFIL("0.07")

var DefSharedParams = &types.SharedSpec{
	ID:                1,
	GasOverEstimation: 1.25,
	MaxFee:            big.Int{Int: DefaultMaxFee.Int},
	GasFeeCap:         big.NewInt(0),
	GasOverPremium:    0,
	SelMsgNum:         20,
	BaseFee:           big.NewInt(0),
}

type SharedParamsService struct {
	repo repo.Repo
}

func NewSharedParamsService(ctx context.Context, repo repo.Repo) (*SharedParamsService, error) {
	sps := &SharedParamsService{
		repo: repo,
	}
	_, err := sps.GetSharedParams(ctx)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if err = sps.SetSharedParams(ctx, DefSharedParams); err != nil {
			return nil, err
		}
	}

	return sps, nil
}

func (sps *SharedParamsService) GetSharedParams(ctx context.Context) (*types.SharedSpec, error) {
	return sps.repo.SharedParamsRepo().GetSharedParams(ctx)
}

func (sps *SharedParamsService) SetSharedParams(ctx context.Context, params *types.SharedSpec) error {
	_, err := sps.repo.SharedParamsRepo().SetSharedParams(ctx, params)
	if err != nil {
		return err
	}
	log.Infof("new shared params %v", params)

	return nil
}
