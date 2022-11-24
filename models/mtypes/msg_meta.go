package mtypes

import (
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"

	types "github.com/filecoin-project/venus/venus-shared/types/messager"
)

type MsgMeta struct {
	ExpireEpoch       abi.ChainEpoch `gorm:"column:expire_epoch;type:bigint;NOT NULL"`
	GasOverEstimation float64        `gorm:"column:gas_over_estimation;type:decimal(10,2)"`

	// todo set GasOverEstimation not null after https://github.com/go-gorm/sqlite/issues/121
	// GasOverEstimation float64        `gorm:"column:gas_over_estimation;type:decimal(10,2);NOT NULL"`
	MaxFee         Int     `gorm:"column:max_fee;type:varchar(256);default:0"`
	GasOverPremium float64 `gorm:"column:gas_over_premium;type:decimal(10,2);"`
}

func (meta *MsgMeta) Meta() *types.SendSpec {
	if meta == nil {
		return &types.SendSpec{
			MaxFee: big.NewInt(0),
		}
	}
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
			MaxFee: NewInt(0),
		}
	}

	return &MsgMeta{
		ExpireEpoch:       srcMeta.ExpireEpoch,
		GasOverEstimation: srcMeta.GasOverEstimation,
		GasOverPremium:    srcMeta.GasOverPremium,
		MaxFee:            SafeFromGo(srcMeta.MaxFee.Int),
	}
}
