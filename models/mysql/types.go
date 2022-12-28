package mysql

import "github.com/filecoin-project/venus-messager/models/mtypes"

// SelectSpec just use in this package, do not use in others
type SelectSpec struct {
	SelMsgNum         uint64     `gorm:"column:sel_msg_num;type:bigint unsigned;NOT NULL"`
	GasOverEstimation float64    `gorm:"column:gas_over_estimation;type:decimal(10,2);NOT NULL"`
	MaxFee            mtypes.Int `gorm:"column:max_fee;type:varchar(256);default:0"`
	GasFeeCap         mtypes.Int `gorm:"column:gas_fee_cap;type:varchar(256);default:0"`
	GasOverPremium    float64    `gorm:"column:gas_over_premium;type:decimal(10,2);NOT NULL"`
	BaseFee           mtypes.Int `gorm:"column:base_fee;type:varchar(256);default:0"`
}
