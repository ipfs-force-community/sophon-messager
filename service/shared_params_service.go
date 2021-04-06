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
	params, err := sps.GetSharedParams(context.TODO())
	if err != nil && !xerrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if params != nil {
		sps.params.SharedParams = params
	}
	sps.refreshParamsLoop()

	return sps, nil
}

func (sps *SharedParamsService) GetSharedParams(ctx context.Context) (*types.SharedParams, error) {
	sp, err := sps.repo.SharedParamsRepo().GetSharedParams(ctx)
	if err != nil {
		return nil, err
	}
	return sp, nil
}

// TODO: check set params?
func (sps *SharedParamsService) SetSharedParams(ctx context.Context, params *types.SharedParams) (struct{}, error) {
	err := sps.repo.SharedParamsRepo().SetSharedParams(ctx, params)
	if err != nil {
		return struct{}{}, err
	}
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
