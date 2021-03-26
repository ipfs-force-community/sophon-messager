package service

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/ipfs-force-community/venus-messager/models/repo"
	"github.com/ipfs-force-community/venus-messager/types"
)

type SharedParamsService struct {
	repo repo.Repo
	log  *logrus.Logger
}

func NewSharedParamsService(repo repo.Repo, logger *logrus.Logger) (*SharedParamsService, error) {
	sps := &SharedParamsService{
		repo: repo,
		log:  logger,
	}

	return sps, nil
}

func (sps *SharedParamsService) GetSharedParams(ctx context.Context) (*types.SharedParams, error) {
	sp, err := sps.repo.SharedParamsRepo().GetSharedParams(ctx)
	if err != nil {
		return nil, err
	}
	return sp, nil
}

func (sps *SharedParamsService) SetSharedParams(ctx context.Context, params *types.SharedParams) (*types.SharedParams, error) {
	return sps.repo.SharedParamsRepo().SetSharedParams(ctx, params)
}
