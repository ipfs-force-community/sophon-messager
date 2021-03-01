package utils

import (
	"gorm.io/gorm"
	"math/big"
)

// BigInt use for auto scan by query
type BigInt struct {
	*big.Int
}

// Value implement sql.Scanner
func (b *BigInt) Value() (driver.Value, error) {
	if b != nil {
		return (b).String(), nil
	}
	return nil, nil
}

// Scan assigns a value from a database driver.
//
// The src value will be of one of the following types:
//
//    int64
//    float64
//    bool
//    []byte
//    string
//    time.Time
//    nil - for NULL values
//
// An error should be returned if the value cannot be stored
// without loss of information.
//
// Reference types such as []byte are only valid until the next call to Scan
// and should not be retained. Their underlying memory is owned by the driver.
// If retention is necessary, copy their values before the next call to Scan.
func (b *BigInt) Scan(value interface{}) error {
	b.Int = new(big.Int)
	if value == nil {
		return nil
	}
	switch t := value.(type) {
	case int64:
		b.SetInt64(t)
	case []byte:
		b.SetString(string(value.([]byte)), 10)
	case string:
		b.SetString(t, 10)
	default:
		return fmt.Errorf("Could not scan type %T into BigInt ", t)
	}
	return nil
}

type YourModel struct {
	Id  int
	Num BigInt
}

func (t *YourModel) BeforeCreate(scope *gorm.Scope) (err error) {
	for _, f := range scope.Fields() {
		v := f.Field.Type().String()
		if v == "*big.Int" {
			f.IsNormal = true
			t := f.Field.Interface().(*big.Int)
			f.Field = reflect.ValueOf(gorm.Expr("cast(? AS DECIMAL(65,0))", t.String()))
		}
	}
	return
}
