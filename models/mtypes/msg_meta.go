package mtypes

import (
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"

	types "github.com/filecoin-project/venus/venus-shared/types/messager"
)

type MsgMeta struct {
	ExpireEpoch       abi.ChainEpoch `gorm:"column:expire_epoch;type:bigint;"`
	GasOverEstimation float64        `gorm:"column:gas_over_estimation;type:decimal(10,2);"`
	MaxFee            Int            `gorm:"column:max_fee;type:varchar(256);"`
	GasOverPremium    float64        `gorm:"column:gas_over_premium;type:decimal(10,2);"`
}

func (meta *MsgMeta) Meta() *types.SendSpec {
	return &types.SendSpec{
		ExpireEpoch:       meta.ExpireEpoch,
		GasOverEstimation: meta.GasOverEstimation,
		MaxFee:            big.Int(SafeFromGo(meta.MaxFee.Int)),
		GasOverPremium:    meta.GasOverPremium,
	}
}

func FromMeta(srcMeta *types.SendSpec) *MsgMeta {
	if srcMeta == nil {
		return &MsgMeta{
			ExpireEpoch:       0,
			GasOverEstimation: 0,
			MaxFee:            Int{},
		}
	}
	meta := &MsgMeta{
		ExpireEpoch:       srcMeta.ExpireEpoch,
		GasOverEstimation: srcMeta.GasOverEstimation,
		GasOverPremium:    srcMeta.GasOverPremium,
	}

	if srcMeta.MaxFee.Int != nil {
		meta.MaxFee = Int{Int: srcMeta.MaxFee.Int}
	}
	return meta
}
