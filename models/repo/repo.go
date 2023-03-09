package repo

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/filecoin-project/go-state-types/crypto"
	"github.com/filecoin-project/go-state-types/exitcode"
	"github.com/filecoin-project/venus/venus-shared/types"
	"gorm.io/gorm"
)

const (
	NotDeleted = -1
	Deleted    = 1
)

type Repo interface {
	GetDb() *gorm.DB
	Transaction(func(txRepo TxRepo) error) error
	DbClose() error
	AutoMigrate() error

	TxRepo
}

type TxRepo interface {
	ActorCfgRepo() ActorCfgRepo
	MessageRepo() MessageRepo
	AddressRepo() AddressRepo
	SharedParamsRepo() SharedParamsRepo
	NodeRepo() NodeRepo
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

type SqlMsgReceipt struct {
	ExitCode exitcode.ExitCode `gorm:"column:exit_code;default:-1"`
	Return   []byte            `gorm:"column:return_value;type:blob;"`
	GasUsed  int64             `gorm:"column:gas_used;type:bigint;NOT NULL"`
}

func (s *SqlMsgReceipt) MsgReceipt() *types.MessageReceipt {
	if s == nil {
		return nil
	}

	return &types.MessageReceipt{
		ExitCode: s.ExitCode,
		Return:   s.Return,
		GasUsed:  s.GasUsed,
	}
}

func FromMsgReceipt(receipt *types.MessageReceipt) *SqlMsgReceipt {
	var s SqlMsgReceipt
	if receipt == nil {
		return nil
	}

	s.GasUsed = receipt.GasUsed
	s.Return = receipt.Return
	s.ExitCode = receipt.ExitCode
	return &s
}

var ErrRecordNotFound = gorm.ErrRecordNotFound
