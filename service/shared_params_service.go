package service

import (
	"context"
	"errors"
	"time"

	"github.com/filecoin-project/go-state-types/big"
	venusTypes "github.com/filecoin-project/venus/venus-shared/types"
	"gorm.io/gorm"

	"github.com/filecoin-project/venus-messager/log"
	"github.com/filecoin-project/venus-messager/models/repo"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
)

const referParamsInterval = time.Second * 10

var DefaultMaxFee = venusTypes.MustParseFIL("0.007")

var DefSharedParams = &types.SharedSpec{
	ID:                1,
	GasOverEstimation: 1.25,
	MaxFee:            big.Int{Int: DefaultMaxFee.Int},
	GasFeeCap:         big.NewInt(0),
	GasOverPremium:    0,
	SelMsgNum:         20,
}

type SharedParamsService struct {
	repo repo.Repo
	log  *log.Logger

	params *Params
}

type Params struct {
	*types.SharedSpec

	ScanIntervalChan chan time.Duration
}

func NewSharedParamsService(ctx context.Context, repo repo.Repo, logger *log.Logger) (*SharedParamsService, error) {
	sps := &SharedParamsService{
		repo: repo,
		log:  logger,
		params: &Params{
			SharedSpec:       &types.SharedSpec{},
			ScanIntervalChan: make(chan time.Duration, 5),
		},
	}
	params, err := sps.GetSharedParams(ctx)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		// avoid data race
		sharedParamsCopy := *DefSharedParams
		if err = sps.SetSharedParams(ctx, &sharedParamsCopy); err != nil {
			return nil, err
		}
		params = &sharedParamsCopy
	}

	sps.params.SharedSpec = params

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
	sps.SetParams(params)

	return nil
}

func (sps *SharedParamsService) GetParams() *Params {
	return sps.params
}

func (sps *SharedParamsService) SetParams(sharedParams *types.SharedSpec) {
	if sharedParams == nil {
		sps.log.Warnf("params is nil")
		return
	}
	sps.log.Infof("old params %v ", sps.params.SharedSpec)

	sps.params.SharedSpec = sharedParams

	sps.log.Infof("new params %v", sharedParams)
}

func (sps *SharedParamsService) RefreshSharedParams(ctx context.Context) error {
	params, err := sps.GetSharedParams(ctx)
	if err != nil {
		return err
	}
	sps.SetParams(params)
	return nil
}
