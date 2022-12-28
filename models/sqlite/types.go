package sqlite

import (
	"database/sql/driver"
	"errors"

	"github.com/filecoin-project/venus-messager/models/mtypes"
)

type SelectSpec struct {
	SelMsgNum uint64     `gorm:"column:sel_msg_num;type:unsigned bigint;NOT NULL"`
	BaseFee   mtypes.Int `gorm:"column:base_fee;type:varchar(256);default:0"` //not include in message

	GasOverEstimation float64    `gorm:"column:gas_over_estimation;type:REAL;NOT NULL"`
	MaxFee            mtypes.Int `gorm:"column:max_fee;type:varchar(256);default:0"`
	GasFeeCap         mtypes.Int `gorm:"column:gas_fee_cap;type:varchar(256);default:0"`
	GasOverPremium    float64    `gorm:"column:gas_over_premium;type:REAL;NOT NULL;default:0"`
}

type sqliteUint64 uint64

func newSqliteUint64(val uint64) sqliteUint64 {
	return sqliteUint64(val)
}

func (c *sqliteUint64) Scan(value interface{}) error {
	switch value := value.(type) {
	case int64:
		*c = sqliteUint64(value)
	case int:
		*c = sqliteUint64(value)
	default:
		return errors.New("address should be a `[]byte` or `string`")
	}

	return nil
}

func (c sqliteUint64) Value() (driver.Value, error) {
	return int64(c), nil
}
