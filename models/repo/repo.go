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

func (smr *SqlMsgReceipt) MsgReceipt() *types.MessageReceipt {
	if smr == nil {
		return nil
	}

	return &types.MessageReceipt{
		ExitCode:    smr.ExitCode,
		ReturnValue: smr.ReturnValue,
		GasUsed:     smr.GasUsed,
	}
}

func (smr *SqlMsgReceipt) FromMsgReceipt(receipt *types.MessageReceipt) *SqlMsgReceipt {
	if receipt == nil {
		return nil
	}

	smr.GasUsed = receipt.GasUsed
	smr.ReturnValue = receipt.ReturnValue
	smr.ExitCode = receipt.ExitCode
	return smr
}
