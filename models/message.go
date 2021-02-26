package models

import (
	"time"
)

type Message struct {
	Id      string `gorm:"column:id;primary_key;"json:"id"` // 主键
	Version uint64 `gorm:"column:version;"json:"version"`

	To    string `gorm:"column:to;type:varchar(128);NOT NULL"json:"to"`
	From  string `gorm:"column:from;type:varchar(128);NOT NULL"json:"from"`
	Nonce uint64 `gorm:"column:nonce;"json:"nonce"`

	Value uint64 `gorm:"column:value;"json:"value"`

	GasLimit   int64  `gorm:"column:gaslimit;"json:"gasLimit"`
	GasFeeCap  uint64 `gorm:"column:gasfeecap;"json:"gasFeeCap"`
	GasPremium uint64 `gorm:"column:gaspremium;"json:"gasPremium"`

	Method   int    `gorm:"column:method;"json:"method"`
	Params   []byte `gorm:"column:params;type:varchar(128);"json:"params"`
	SignData []byte `gorm:"column:signdata;type:varchar(128);"json:"params"`

	IsDeleted int       `gorm:"column:is_deleted;default:-1;NOT NULL"`                // 是否删除 1:是  -1:否
	CreatedAt time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP;NOT NULL"` // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP;NOT NULL"` // 更新时间
}

func (m *Message) TableName() string {
	return "messages"
}

type MessageRepo interface {
	SaveMessage(msg *Message) (string, error)
}

var _ MessageRepo = (*messageRepo)(nil)

type messageRepo struct {
	Repo
}

func NewMessageRepo(db Repo) MessageRepo {
	return &messageRepo{
		db,
	}
}

func (m messageRepo) SaveMessage(msg *Message) (string, error) {
	err := m.GetDb().Save(msg).Error
	return msg.Id, err
}
