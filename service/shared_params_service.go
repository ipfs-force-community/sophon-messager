package service

import (
	"context"
	"errors"
	"reflect"
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

var defParams = &types.SharedSpec{
	ID:                0,
	GasOverEstimation: 1.25,
	MaxFee:            big.Int{Int: DefaultMaxFee.Int},
	MaxFeeCap:         big.NewInt(0),
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

func NewSharedParamsService(repo repo.Repo, logger *log.Logger) (*SharedParamsService, error) {
	sps := &SharedParamsService{
		repo: repo,
		log:  logger,
		params: &Params{
			SharedSpec:       &types.SharedSpec{},
			ScanIntervalChan: make(chan time.Duration, 5),
		},
	}
	ctx := context.TODO()
	params, err := sps.GetSharedParams(ctx)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if err = sps.SetSharedParams(ctx, defParams); err != nil {
			return nil, err
		}
		params = defParams
	}

	sps.params.SharedSpec = params
	sps.refreshParamsLoop()

	return sps, nil
}

func (sps *SharedParamsService) GetSharedParams(ctx context.Context) (*types.SharedSpec, error) {
	return sps.repo.SharedParamsRepo().GetSharedParams(ctx)
}

func (sps *SharedParamsService) SetSharedParams(ctx context.Context, params *types.SharedSpec) error {
	id, err := sps.repo.SharedParamsRepo().SetSharedParams(ctx, params)
	if err != nil {
		return err
	}
	params.ID = id
	sps.SetParams(params)

	return err
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
	if sharedParams.GetSendSpec() != nil {
		sps.params.GasOverEstimation = sharedParams.GasOverEstimation
		sps.params.MaxFee = sharedParams.MaxFee
		sps.params.MaxFeeCap = sharedParams.MaxFeeCap
		sps.params.GasOverPremium = sharedParams.GasOverPremium
	}
	if sharedParams.SelMsgNum > 0 {
		sps.params.SelMsgNum = sharedParams.SelMsgNum
	}
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

func (sps *SharedParamsService) refreshParamsLoop() {
	go func() {
		ticker := time.NewTicker(referParamsInterval)
		defer ticker.Stop()

		for range ticker.C {
			params, err := sps.GetSharedParams(context.TODO())
			if err != nil {
				sps.log.Warnf("get shared params %v", err)
				continue
			}
			if !reflect.DeepEqual(sps.params.SharedSpec, params) {
				sps.SetParams(params)
			}
		}
	}()
}
