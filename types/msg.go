package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/crypto"
	"github.com/filecoin-project/venus/pkg/types"
	"github.com/ipfs/go-cid"
	"time"
)

// Deprecated: use 'Message'
type DeprecatedMessage struct {
	Id      string `json:"id"` // 主键
	Version uint64 `json:"version"`

	To    string `json:"to"`
	From  string `json:"from"`
	Nonce uint64 `json:"nonce"`

	Value *Int `json:"value"`

	GasLimit   int64 `json:"gasLimit"`
	GasFeeCap  *Int  `json:"gasFeeCap"`
	GasPremium *Int  `json:"gasPremium"`

	Method   int    `json:"method"`
	Params   []byte `json:"params"`
	SignData []byte `json:"signData"`

	Epoch uint64 `json:"epoch"` // had message been mined on any block, yes:1, no:0, todo: save exact epoch

	IsDeleted int       `json:"isDeleted"` // 是否删除 1:是  -1:否
	CreatedAt time.Time `json:"createAt"`  // 创建时间
	UpdatedAt time.Time `json:"updateAt"`  // 更新时间
}

func (m *DeprecatedMessage) TableName() string {
	return "messages"
}

type Message struct {
	Uid string `json:"uuid"` // 主键

	types.UnsignedMessage

	*crypto.Signature
	Meta *MsgMeta
}

func (m *Message) Cid() cid.Cid {
	if m.Signature != nil {
		return m.SignedCid()
	}
	return m.Cid()
}

func (m *Message) UnsingedCid() cid.Cid {
	return m.UnsignedMessage.Cid()
}

func (m *Message) SignedCid() cid.Cid {
	if m.Signature != nil {
		return cid.Undef
	}
	return (&types.SignedMessage{m.UnsignedMessage, *m.Signature}).Cid()
}

func (m *Message) TableName() string {
	return "messages"
}

type MsgMeta struct {
	ExpireEpoch       abi.ChainEpoch `json:"expireEpoch"`
	GasOverEstimation float64        `json:"gasOverEstimation"`
	MaxFee            big.Int        `json:"maxFee,omitempty"`
	MaxFeeCap         big.Int        `json:"maxFeeCap"`
}

func (me *MsgMeta) Scan(value interface{}) error {
	sqlBin, isok := value.([]byte)
	if !isok {
		return fmt.Errorf("value must be []byte")
	}
	return json.Unmarshal(sqlBin, me)
}

func (me *MsgMeta) Value() (driver.Value, error) {
	return json.Marshal(me)
}
