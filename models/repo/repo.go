package repo

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/filecoin-project/go-state-types/crypto"
	"gorm.io/gorm"
)

type Repo interface {
	GetDb() *gorm.DB
	DbClose() error
	AutoMigrate() error

	WalletRepo() WalletRepo
	MessageRepo() MessageRepo
}

type ISqlField interface {
	Scan(value interface{}) error
	Value() (driver.Value, error)
}

type SqlSignature crypto.Signature

func (s *SqlSignature) Scan(value interface{}) error {
	sqlBin, isok := value.([]byte)
	if !isok {
		return fmt.Errorf("value must be []byte")
	}
	return json.Unmarshal(sqlBin, s)
}

func (s *SqlSignature) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

