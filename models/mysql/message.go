package mysql

import (
	"reflect"
	"time"

	"github.com/filecoin-project/go-address"
	"gorm.io/gorm"

	"github.com/filecoin-project/go-state-types/abi"

	venustypes "github.com/filecoin-project/venus/pkg/types"
	"github.com/hunjixin/automapper"

	"github.com/ipfs-force-community/venus-messager/models/repo"
	"github.com/ipfs-force-community/venus-messager/types"
)

type mysqlMessage struct {
	Id      string `gorm:"column:id;primary_key;" json:"id"` // 主键
	Version uint64 `gorm:"column:version;" json:"version"`

	To    string `gorm:"column:to;type:varchar(256);NOT NULL" json:"to"`
	From  string `gorm:"column:from;type:varchar(256);NOT NULL" json:"from"`
	Nonce uint64 `gorm:"column:nonce;" json:"nonce"`

	Value *types.Int `gorm:"column:value;type:varchar(256);" json:"value"`

	GasLimit   int64      `gorm:"column:gaslimit;" json:"gasLimit"`
	GasFeeCap  *types.Int `gorm:"column:gasfeecap;type:varchar(256);" json:"gasFeeCap"`
	GasPremium *types.Int `gorm:"column:gaspremium;type:varchar(256);" json:"gasPremium"`

	Method int `gorm:"column:method;" json:"method"`

	Params   []byte `gorm:"column:params;type:text;" json:"params"`
	SignData []byte `gorm:"column:signdata;type:varchar(256);" json:"signData"`

	IsDeleted int       `gorm:"column:is_deleted;default:-1;NOT NULL"`                // 是否删除 1:是  -1:否
	CreatedAt time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP;NOT NULL"` // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP;NOT NULL"` // 更新时间
}

func FromMessage(msg types.Message) *mysqlMessage {
	return automapper.MustMapper(&msg, TMysqlMessage).(*mysqlMessage)
}

func (m mysqlMessage) Message() *types.Message {
	return automapper.MustMapper(&m, TMessage).(*types.Message)
}

func (m *mysqlMessage) TableName() string {
	return "messages"
}

var _ repo.MessageRepo = (*mysqlMessageRepo)(nil)

type mysqlMessageRepo struct {
	*gorm.DB
}

func (m mysqlMessageRepo) ListFilledMessageByAddress(addr address.Address) ([]*types.Message, error) {
	panic("implement me")
}

func (m mysqlMessageRepo) GetMessageState(uuid types.UUID) (types.MessageState, error) {
	panic("implement me")
}

func (m mysqlMessageRepo) ExpireMessage(msg []*types.Message) error {
	panic("implement me")
}

func (m mysqlMessageRepo) ListUnChainMessageByAddress(addr address.Address) ([]*types.Message, error) {
	panic("implement me")
}

func (m mysqlMessageRepo) BatchSaveMessage(msg []*types.Message) error {
	panic("implement me")
}

func newMysqlMessageRepo(db *gorm.DB) mysqlMessageRepo {
	return mysqlMessageRepo{DB: db}
}

func (m mysqlMessageRepo) SaveMessage(msg *types.Message) (types.UUID, error) {
	err := m.DB.Save(msg).Error
	return msg.ID, err
}

func (m mysqlMessageRepo) GetMessage(uuid types.UUID) (*types.Message, error) {
	var msg *mysqlMessage
	if err := m.DB.Where(&mysqlMessage{Id: uuid.String(), IsDeleted: -1}).First(&msg).Error; err != nil {
		return nil, err
	}
	return msg.Message(), nil
}

func (m mysqlMessageRepo) GetMessageByCid(cid string) (*types.Message, error) {
	panic("implement me")
}

func (m mysqlMessageRepo) ListUnchainedMsgs() ([]*types.Message, error) {
	panic("implement me")
}

func (m mysqlMessageRepo) GetSignedMessageByTime(start time.Time) ([]*types.Message, error) {
	panic("implement me")
}

func (m mysqlMessageRepo) GetSignedMessageByHeight(height abi.ChainEpoch) ([]*types.Message, error) {
	panic("implement me")
}

func (m mysqlMessageRepo) ListMessage() ([]*types.Message, error) {
	var internalMsg []*mysqlMessage
	if err := m.DB.Find(&internalMsg, "is_deleted = ?", -1).Error; err != nil {
		return nil, err
	}

	result, err := automapper.Mapper(internalMsg, reflect.TypeOf([]*types.Message{}))
	if err != nil {
		return nil, err
	}
	return result.([]*types.Message), nil
}

func (m mysqlMessageRepo) UpdateMessageInfoByCid(unsignedCid string, receipt *venustypes.MessageReceipt, height abi.ChainEpoch, state types.MessageState, tsKey string) (string, error) {
	panic("implement me")
}

func (m mysqlMessageRepo) UpdateMessageStateByCid(cid string, state types.MessageState) error {
	panic("implement me")
}
