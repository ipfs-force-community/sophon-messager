package mtypes

import (
	"database/sql/driver"
	"fmt"
	"math/big"
)

type Int struct {
	*big.Int
}

func NewInt(i int64) Int {
	return Int{big.NewInt(0).SetInt64(i)}
}

func NewFromGo(i *big.Int) Int {
	return Int{big.NewInt(0).Set(i)}
}

// Value implement sql.Scanner
func (bi Int) Value() (driver.Value, error) {
	if bi.Int != nil {
		return (bi).String(), nil
	}
	return "0", nil
}

// Scan assigns a value from a database driver.
// An error should be returned if the value cannot be stored
// without loss of information.
//
// Reference types such as []byte are only valid until the next call to Scan
// and should not be retained. Their underlying memory is owned by the driver.
// If retention is necessary, copy their values before the next call to Scan.
func (bi *Int) Scan(value interface{}) error {
	bi.Int = new(big.Int)
	if value == nil {
		return nil
	}
	switch t := value.(type) {
	case int64:
		bi.SetInt64(t)
	case []byte:
		bi.SetString(string(value.([]byte)), 10)
	case string:
		bi.SetString(t, 10)
	default:
		return fmt.Errorf("could not scan type %T into BigInt ", t)
	}
	return nil
}
