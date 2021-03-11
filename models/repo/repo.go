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
	DbClose() error
	AutoMigrate() error

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

const ExitCodeToExec = exitcode.ExitCode(-1)

type SqlMsgReceipt struct {
	ExitCode    exitcode.ExitCode `json:"exitCode"`
	ReturnValue []byte            `json:"return"`
	GasUsed     int64             `json:"gasUsed"`
}

func (s *SqlMsgReceipt) Scan(value interface{}) error {
	sqlBin, isok := value.([]byte)
	if !isok {
		return fmt.Errorf("value must be []byte")
	}
	return json.Unmarshal(sqlBin, s)
}

func (s *SqlMsgReceipt) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
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

func (s *SqlMsgReceipt) FromMsgReceipt(receipt *types.MessageReceipt) *SqlMsgReceipt {
	if receipt == nil {
		return nil
	}

	s.GasUsed = receipt.GasUsed
	s.ReturnValue = receipt.ReturnValue
	s.ExitCode = receipt.ExitCode
	return s
}
