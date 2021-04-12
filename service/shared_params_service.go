package service

import (
	"context"
	"reflect"
	"time"

	"golang.org/x/xerrors"
	"gorm.io/gorm"

	"github.com/sirupsen/logrus"

	"github.com/ipfs-force-community/venus-messager/models/repo"
	"github.com/ipfs-force-community/venus-messager/types"
)

const referParamsInterval = time.Second * 10

var defParams = &types.SharedParams{
	ID:                 0,
	ExpireEpoch:        0,
	GasOverEstimation:  0,
	MaxFee:             100000,
	MaxFeeCap:          20000,
	SelMsgNum:          20,
	ScanInterval:       10,
	MaxEstFailNumOfMsg: 5,
}

type SharedParamsService struct {
	repo repo.Repo
	log  *logrus.Logger

	params *Params
}

type Params struct {
	*types.SharedParams

	ScanIntervalChan chan time.Duration
}

func NewSharedParamsService(repo repo.Repo, logger *logrus.Logger) (*SharedParamsService, error) {
	sps := &SharedParamsService{
		repo: repo,
		log:  logger,
		params: &Params{
			SharedParams:     &types.SharedParams{},
			ScanIntervalChan: make(chan time.Duration, 5),
		},
	}
	ctx := context.TODO()
	params, err := sps.GetSharedParams(ctx)
	if err != nil {
		if !xerrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if _, err = sps.SetSharedParams(ctx, defParams); err != nil {
			return nil, err
		}
		params = defParams
	}

	sps.params.SharedParams = params
	sps.refreshParamsLoop()

	return sps, nil
}

func (sps *SharedParamsService) GetSharedParams(ctx context.Context) (*types.SharedParams, error) {
	return sps.repo.SharedParamsRepo().GetSharedParams(ctx)
}

func (sps *SharedParamsService) SetSharedParams(ctx context.Context, params *types.SharedParams) (struct{}, error) {
	id, err := sps.repo.SharedParamsRepo().SetSharedParams(ctx, params)
	if err != nil {
		return struct{}{}, err
	}
	params.ID = id
	sps.SetParams(params)

	return struct{}{}, err
}

func (sps *SharedParamsService) GetParams() *Params {
	return sps.params
}

func (sps *SharedParamsService) SetParams(sharedParams *types.SharedParams) {
	if sharedParams == nil {
		sps.log.Warnf("params is nil")
		return
	}
	sps.log.Infof("old params %v ", sps.params.SharedParams)
	if sharedParams.GetMsgMeta() != nil {
		sps.params.ExpireEpoch = sharedParams.ExpireEpoch
		sps.params.GasOverEstimation = sharedParams.GasOverEstimation
		sps.params.MaxFee = sharedParams.MaxFee
		sps.params.MaxFeeCap = sharedParams.MaxFeeCap
	}
	if sharedParams.SelMsgNum > 0 {
		sps.params.SelMsgNum = sharedParams.SelMsgNum
	}
	if sharedParams.ScanInterval > 0 {
		if sps.params.ScanInterval != sharedParams.ScanInterval {
			sps.params.ScanInterval = sharedParams.ScanInterval
			sps.params.ScanIntervalChan <- time.Duration(sharedParams.ScanInterval) * time.Second
		}
	}
	sps.params.MaxEstFailNumOfMsg = sharedParams.MaxEstFailNumOfMsg
	sps.log.Infof("new params %v", sharedParams)
}

func (sps *SharedParamsService) RefreshSharedParams(ctx context.Context) (struct{}, error) {
	params, err := sps.GetSharedParams(ctx)
	if err != nil {
		return struct{}{}, err
	}
	sps.SetParams(params)
	return struct{}{}, nil
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
			if !reflect.DeepEqual(sps.params.SharedParams, params) {
				sps.SetParams(params)
			}
		}
	}()
}
