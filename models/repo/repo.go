package repo

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/filecoin-project/go-state-types/crypto"
	"github.com/filecoin-project/go-state-types/exitcode"
	"github.com/filecoin-project/venus/pkg/types"
	"gorm.io/gorm"
)

type Repo interface {
	GetDb() *gorm.DB
	Transaction(func(txRepo TxRepo) error) error
	DbClose() error
	AutoMigrate() error

	WalletRepo() WalletRepo
	MessageRepo() MessageRepo
	AddressRepo() AddressRepo
	SharedParamsRepo() SharedParamsRepo
	NodeRepo() NodeRepo
}

type TxRepo interface {
	WalletRepo() WalletRepo
	MessageRepo() MessageRepo
	AddressRepo() AddressRepo
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
	ExitCode    exitcode.ExitCode `gorm:"column:exit_code;default:-1"`
	ReturnValue []byte            `gorm:"column:return_value;type:blob;"`
	GasUsed     int64             `gorm:"column:gas_used;type:bigint;"`
}

func (s *SqlMsgReceipt) MsgReceipt() *types.MessageReceipt {
	if s == nil {
		return nil
	}

	return &types.MessageReceipt{
		ExitCode:    s.ExitCode,
		ReturnValue: s.ReturnValue,
		GasUsed:     s.GasUsed,
	}
}

func FromMsgReceipt(receipt *types.MessageReceipt) *SqlMsgReceipt {
	var s SqlMsgReceipt
	if receipt == nil {
		return nil
	}

	s.GasUsed = receipt.GasUsed
	s.ReturnValue = receipt.ReturnValue
	s.ExitCode = receipt.ExitCode
	return &s
}
