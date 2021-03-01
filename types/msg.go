package types

import (
	"time"
)

type Message struct {
	Id      string `json:"id"` // 主键
	Version uint64 `"json:"version"`

	To    string `json:"to"`
	From  string `json:"from"`
	Nonce uint64 `json:"nonce"`

	Value *Int `json:"value"`

	GasLimit   int64 `json:"gasLimit"`
	GasFeeCap  *Int  `json:"gasFeeCap"`
	GasPremium *Int  `json:"gasPremium"`

	Method   int    `json:"method"`
	Params   []byte `json:"params"`
	SignData []byte `json:"params"`

	IsDeleted int       `json:"isDeleted"` // 是否删除 1:是  -1:否
	CreatedAt time.Time `json:"createAt"`  // 创建时间
	UpdatedAt time.Time `json:"updateAt"`  // 更新时间
}

func (m *Message) TableName() string {
	return "messages"
}
