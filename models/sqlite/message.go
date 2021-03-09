package sqlite

import (
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/crypto"
	"github.com/filecoin-project/go-state-types/exitcode"
	types2 "github.com/filecoin-project/venus/pkg/types"
	venustypes "github.com/filecoin-project/venus/pkg/types"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/ipfs-force-community/venus-messager/models/repo"
	"github.com/ipfs-force-community/venus-messager/types"
)

type sqliteMessage struct {
	gorm.Model
	ID      string `gorm:"column:id;uniqueIndex"json:"id"`
	Version uint64 `gorm:"column:version;"json:"version"`

	To    string `gorm:"column:to;type:varchar(256);NOT NULL"json:"to"`
	From  string `gorm:"column:from;type:varchar(256);NOT NULL;index:idx_from_nonce"json:"from"`
	Nonce uint64 `gorm:"column:nonce;index:idx_from_nonce"json:"nonce"`

	Value decimal.NullDecimal `gorm:"column:value;type:varchar(256);"json:"value"`

	GasLimit   int64               `gorm:"column:gaslimit;"json:"gasLimit"`
	GasFeeCap  decimal.NullDecimal `gorm:"column:gasfeecap;type:varchar(256);"json:"gasFeeCap"`
	GasPremium decimal.NullDecimal `gorm:"column:gaspremium;type:varchar(256);"json:"gasPremium"`

	Method int `gorm:"column:method;" json:"method"`

	Params []byte `gorm:"column:params;type:text;"json:"params"`

	Signature *repo.SqlSignature `gorm:"column:signed_data"json:"params"`

	Cid       string `gorm:"column:cid;"`
	SignedCid string `gorm:"column:signed_cid;index:idx_epoch_txid"`

	Height   uint64              `gorm:"column:height;index:idx_epoch_txid"`
	ExitCode exitcode.ExitCode   `gorm:"column:exit_code;default:-1"`
	Receipt  *repo.SqlMsgReceipt `grom:"column:receipt;embedded"`

	Meta *types.MsgMeta `gorm:"column:meta;blob"`

	State types.MessageState `gorm:"column:state;"json:"version"`
}

var toDeicimal = func(b big.Int) decimal.NullDecimal {
	return decimal.NullDecimal{decimal.NewFromBigInt(b.Int, 0), b.Int != nil}
}

var fromDecimal = func(decimal decimal.NullDecimal) big.Int {
	i := big.NewInt(0)
	if decimal.Valid {
		i.Int.SetString(decimal.Decimal.String(), 10)
	}
	return i
}

func FromMessage(srcMsg *types.Message) *sqliteMessage {
	destMsg := &sqliteMessage{
		ID:         srcMsg.ID,
		Version:    srcMsg.Version,
		To:         srcMsg.To.String(),
		From:       srcMsg.From.String(),
		Nonce:      srcMsg.Nonce,
		Value:      toDeicimal(srcMsg.Value),
		GasLimit:   srcMsg.GasLimit,
		GasFeeCap:  toDeicimal(srcMsg.GasFeeCap),
		GasPremium: toDeicimal(srcMsg.GasPremium),
		Method:     int(srcMsg.Method),
		Params:     srcMsg.Params,
		Signature:  (*repo.SqlSignature)(srcMsg.Signature),
		Cid:        srcMsg.UnsingedCid().String(),
		SignedCid:  srcMsg.SignedCid().String(),
		ExitCode:   repo.ExitCodeToExec,
		Height:     srcMsg.Height,
		Receipt:    (&repo.SqlMsgReceipt{}).FromMsgReceipt(srcMsg.Receipt),
		Meta:       srcMsg.Meta,
		State:      srcMsg.State,
	}

	if srcMsg.Receipt != nil {
		destMsg.ExitCode = srcMsg.Receipt.ExitCode
	}

	return destMsg
}

func (sqlMsg *sqliteMessage) Message() *types.Message {
	var destMsg = &types.Message{
		ID: sqlMsg.ID,
		UnsignedMessage: types2.UnsignedMessage{
			Version:    sqlMsg.Version,
			Nonce:      sqlMsg.Nonce,
			Value:      fromDecimal(sqlMsg.Value),
			GasLimit:   sqlMsg.GasLimit,
			GasFeeCap:  fromDecimal(sqlMsg.GasFeeCap),
			GasPremium: fromDecimal(sqlMsg.GasPremium),
			Method:     abi.MethodNum(sqlMsg.Method),
			Params:     sqlMsg.Params,
		},
		Height:    sqlMsg.Height,
		Receipt:   sqlMsg.Receipt.MsgReceipt(),
		Signature: (*crypto.Signature)(sqlMsg.Signature),
		Meta:      sqlMsg.Meta,
		State:     sqlMsg.State,
	}
	destMsg.From, _ = address.NewFromString(sqlMsg.From)
	destMsg.To, _ = address.NewFromString(sqlMsg.To)
	return destMsg
}

func (m *sqliteMessage) TableName() string {
	return "messages"
}

var _ repo.MessageRepo = (*sqliteMessageRepo)(nil)

type sqliteMessageRepo struct {
	repo.Repo
}

func newSqliteMessageRepo(repo repo.Repo) *sqliteMessageRepo {
	return &sqliteMessageRepo{repo}
}

func (m *sqliteMessageRepo) SaveMessage(msg *types.Message) (string, error) {
	sqlMsg := FromMessage(msg)

	err := m.GetDb().Debug().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoNothing: true}).Save(sqlMsg).Error

	return msg.ID, err
}

func (m *sqliteMessageRepo) GetMessage(uuid string) (*types.Message, error) {
	var msg sqliteMessage
	if err := m.GetDb().Where("id = ?", uuid).First(&msg).Error; err != nil {
		return nil, err
	}
	return msg.Message(), nil
}

func (m *sqliteMessageRepo) GetMessageByCid(cid string) (*types.Message, error) {
	var msg sqliteMessage
	if err := m.GetDb().Where("signed_cid = ?", cid).First(&msg).Error; err != nil {
		return nil, err
	}
	return msg.Message(), nil
}

func (m *sqliteMessageRepo) GetMessageByTime(start time.Time) ([]*types.Message, error) {
	var sqlMsgs []*sqliteMessage
	if err := m.GetDb().Where("created_at >= ?", start).Find(&sqlMsgs).Error; err != nil {
		return nil, err
	}
	result := make([]*types.Message, len(sqlMsgs))
	for idx, msg := range sqlMsgs {
		result[idx] = msg.Message()
	}

	return result, nil
}

func (m *sqliteMessageRepo) ListMessage() ([]*types.Message, error) {
	var sqlMsgs []*sqliteMessage
	if err := m.GetDb().Find(&sqlMsgs).Error; err != nil {
		return nil, err
	}

	result := make([]*types.Message, len(sqlMsgs))
	for idx, msg := range sqlMsgs {
		result[idx] = msg.Message()
	}
	return result, nil
}

func (m *sqliteMessageRepo) ListUnchainedMsgs() ([]*types.Message, error) {
	var sqlMsgs []*sqliteMessage
	if err := m.Repo.GetDb().Debug().Model((*sqliteMessage)(nil)).
		Where("height=0 and signed_data is null").
		Find(&sqlMsgs).Error; err != nil {
		return nil, err
	}

	var result = make([]*types.Message, len(sqlMsgs))

	for idx, msg := range sqlMsgs {
		result[idx] = msg.Message()
	}
	return result, nil
}

func (m *sqliteMessageRepo) UpdateMessageReceipt(cid string, receipt *venustypes.MessageReceipt, height abi.ChainEpoch, state types.MessageState) (string, error) {
	sqlMsg := sqliteMessage{
		Height:   uint64(height),
		ExitCode: receipt.ExitCode,
		Receipt:  (&repo.SqlMsgReceipt{}).FromMsgReceipt(receipt),
		State:    state,
	}
	return cid, m.GetDb().Debug().Model(&sqliteMessage{}).
		Where("signed_cid = ?", cid).
		Select("height", "exit_code", "receipt", "state").
		UpdateColumns(sqlMsg).Error
}

func (m *sqliteMessageRepo) UpdateMessageStateByCid(cid string, state types.MessageState) error {
	return m.GetDb().Debug().Model(&sqliteMessage{}).
		Where("signed_cid = ?", cid).UpdateColumn("state", state).Error
}
