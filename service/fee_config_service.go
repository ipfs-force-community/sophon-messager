package service

import (
	"time"

	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/venus-messager/models/repo"
	"github.com/filecoin-project/venus-messager/types"
	venusTypes "github.com/filecoin-project/venus/pkg/types"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
	"gorm.io/gorm"
)

var DefaultMaxFee = venusTypes.MustParseFIL("0.07")

var globalFeeConfig = types.FeeConfig{
	ID:                types.UUID{},
	WalletID:          types.UUID{},
	MethodType:        -1,
	GasOverEstimation: 1.25,
	MaxFee:            big.NewInt(DefaultMaxFee.Int64()),
	MaxFeeCap:         big.NewInt(0),
}

type FeeConfigService struct {
	repo repo.Repo
	log  *logrus.Logger
}

func NewFeeConfigService(repo repo.Repo, logger *logrus.Logger) (*FeeConfigService, error) {
	fcs := &FeeConfigService{
		repo: repo,
		log:  logger,
	}
	_, err := fcs.repo.FeeConfigRepo().GetGlobalFeeConfig()
	if err != nil {
		if !xerrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		gfc := &globalFeeConfig
		gfc.CreatedAt = time.Now()
		if err := fcs.repo.FeeConfigRepo().SaveFeeConfig(gfc); err != nil {
			return nil, xerrors.Errorf("save global fee config failed %v", err)
		}
	}

	return fcs, nil
}

func (fcs *FeeConfigService) SelectFeeConfig(walletName string, methodType uint64) (*types.FeeConfig, error) {
	wallet, err := fcs.repo.WalletRepo().GetWalletByName(walletName)
	if err != nil {
		return nil, xerrors.Errorf("got wallet(%s) failed %v", walletName, err)
	}
	fc, err := fcs.repo.FeeConfigRepo().GetFeeConfig(wallet.ID, int64(methodType))
	if err != nil {
		if !xerrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		fc, err = fcs.repo.FeeConfigRepo().GetWalletFeeConfig(wallet.ID)
		if err != nil && xerrors.Is(err, gorm.ErrRecordNotFound) {
			return fcs.repo.FeeConfigRepo().GetGlobalFeeConfig()
		}

		return fc, err
	}

	return fc, nil
}
