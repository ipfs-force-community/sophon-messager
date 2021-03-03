package sqlite

import (
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/crypto"
	types2 "github.com/filecoin-project/venus/pkg/types"
	"github.com/ipfs-force-community/venus-messager/models/repo"
	"github.com/ipfs-force-community/venus-messager/types"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type sqliteMessage struct {
	gorm.Model
	Uid     string `gorm:"column:uuid;uniqueIndex"json:"uuid"`
	Version uint64 `gorm:"column:version;"json:"version"`

	To    string `gorm:"column:to;type:varchar(256);NOT NULL"json:"to"`
	From  string `gorm:"column:from;type:varchar(256);NOT NULL;index:idx_from_nonce"json:"from"`
	Nonce uint64 `gorm:"column:nonce;index:idx_from_nonce"json:"nonce"`

	Value decimal.NullDecimal `gorm:"column:value;type:varchar(256);"json:"value"`

	GasLimit   int64               `gorm:"column:gaslimit;"json:"gasLimit"`
	GasFeeCap  decimal.NullDecimal `gorm:"column:gasfeecap;type:varchar(256);"json:"gasFeeCap"`
	GasPremium decimal.NullDecimal `gorm:"column:gaspremium;type:varchar(256);"json:"gasPremium"`

	Method int `gorm:"column:method;" json:"method"`

	Params    []byte             `gorm:"column:params;type:text;"json:"params"`
	Epoch     uint64             `gorm:"index:idx_epoch_txid"`
	Signature *repo.SqlSignature `gorm:"column:signdata" json:"params"`
	Cid       string             `gorm:"uniqueIndex"`
	SignedCid string             `gorm:"uniqueIndex;index:idx_epoch_txid"`

	Meta *types.MsgMeta `gorm:"blob"`
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
		Uid:        srcMsg.Uid,
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
		Meta:       srcMsg.Meta}

	return destMsg
}

func (sqlMsg *sqliteMessage) Message() *types.Message {
	var destMsg = &types.Message{
		Uid: sqlMsg.Uid,
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
		Signature: (*crypto.Signature)(sqlMsg.Signature),
		Meta:      sqlMsg.Meta,
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

	// err := m.GetDb().Debug().Clauses(clause.OnConflict{
	// 	Columns:   []clause.Column{{Name: "from"}, {Name: "nonce"}},
	// 	UpdateAll: true,
	// 	DoUpdates: clause.AssignmentColumns([]string{"private_key"}),
	// }).Save(sqlMsg).Error

	err := m.GetDb().Debug().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "uuid"}},
		DoNothing: true}).Save(sqlMsg).Error

	return msg.Uid, err
}

func (m *sqliteMessageRepo) GetMessage(uuid string) (*types.Message, error) {
	var msg sqliteMessage
	if err := m.GetDb().Where("uuid = ?", uuid).First(&msg).Error; err != nil {
		return nil, err
	}
	return msg.Message(), nil
}

func (m *sqliteMessageRepo) ListMessage() ([]*types.Message, error) {
	var sqlMsgs []*sqliteMessage
	if err := m.GetDb().Find(&sqlMsgs).Error; err != nil {
		return nil, err
	}

	var result = make([]*types.Message, len(sqlMsgs))

	for idx, msg := range sqlMsgs {
		result[idx] = msg.Message()
	}
	return result, nil
}

func (m *sqliteMessageRepo) ListUnchainedMsgs() ([]*types.Message, error) {
	var sqlMsgs []*sqliteMessage
	var err error
	if err = m.Repo.GetDb().Debug().Model((*sqliteMessage)(nil)).
		Where("epoch=0 and signdata is null").
		Find(&sqlMsgs).Error; err != nil {
		return nil, err
	}

	var result = make([]*types.Message, len(sqlMsgs))

	for idx, msg := range sqlMsgs {
		result[idx] = msg.Message()
	}
	return result, nil
}
