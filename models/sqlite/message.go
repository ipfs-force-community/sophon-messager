package sqlite

import (
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/crypto"
	types2 "github.com/filecoin-project/venus/pkg/types"
	venustypes "github.com/filecoin-project/venus/pkg/types"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/ipfs-force-community/venus-messager/models/repo"
	"github.com/ipfs-force-community/venus-messager/types"
)

type sqliteMessage struct {
	ID      types.UUID `gorm:"column:id;type:varchar(256);primary_key"`
	Version uint64     `gorm:"column:version;unsigned bigint"`

	From  string `gorm:"column:from;type:varchar(256);NOT NULL;index;index:idx_from_nonce"`
	Nonce uint64 `gorm:"column:nonce;type:unsigned bigint;index:idx_from_nonce"`
	To    string `gorm:"column:to;type:varchar(256);NOT NULL"`

	Value types.Int `gorm:"column:value;type:varchar(256);"`

	GasLimit   int64     `gorm:"column:gas_limit;type:bigint"`
	GasFeeCap  types.Int `gorm:"column:gas_fee_cap;type:varchar(256);"`
	GasPremium types.Int `gorm:"column:gas_premium;type:varchar(256);"`

	Method int `gorm:"column:method;type:int"`

	Params []byte `gorm:"column:params;type:blob;"`

	Signature *repo.SqlSignature `gorm:"column:signed_data;type:blob;"`

	UnsignedCid string `gorm:"column:unsigned_cid;type:varchar(256);index:unsigned_cid;"`
	SignedCid   string `gorm:"column:signed_cid;type:varchar(256);index:signed_cid"`

	Height  uint64              `gorm:"column:height;type:unsigned bigint;index:height"`
	Receipt *repo.SqlMsgReceipt `gorm:"embedded;embeddedPrefix:receipt_"`

	Meta *MsgMeta `gorm:"embedded;embeddedPrefix:meta_"`

	State types.MessageState `gorm:"column:state;type:int;"`

	IsDeleted int       `gorm:"column:is_deleted;index:is_deleted;default:-1;NOT NULL"`                // 是否删除 1:是  -1:否
	CreatedAt time.Time `gorm:"column:created_at;index:created_at;default:CURRENT_TIMESTAMP;NOT NULL"` // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;index:update_at;default:CURRENT_TIMESTAMP;NOT NULL"`  // 更新时间
}

func (sqlMsg *sqliteMessage) TableName() string {
	return "messages"
}

func (sqlMsg *sqliteMessage) Message() *types.Message {
	var destMsg = &types.Message{
		ID: sqlMsg.ID,
		UnsignedMessage: types2.UnsignedMessage{
			Version:    sqlMsg.Version,
			Nonce:      sqlMsg.Nonce,
			Value:      big.NewFromGo(sqlMsg.Value.Int),
			GasLimit:   sqlMsg.GasLimit,
			GasFeeCap:  big.NewFromGo(sqlMsg.GasFeeCap.Int),
			GasPremium: big.NewFromGo(sqlMsg.GasPremium.Int),
			Method:     abi.MethodNum(sqlMsg.Method),
			Params:     sqlMsg.Params,
		},
		Height:    sqlMsg.Height,
		Receipt:   sqlMsg.Receipt.MsgReceipt(),
		Signature: (*crypto.Signature)(sqlMsg.Signature),
		Meta:      sqlMsg.Meta.Meta(),
		State:     sqlMsg.State,
	}
	destMsg.From, _ = address.NewFromString(sqlMsg.From)
	destMsg.To, _ = address.NewFromString(sqlMsg.To)
	return destMsg
}

func FromMessage(srcMsg *types.Message) *sqliteMessage {
	destMsg := &sqliteMessage{
		ID:      srcMsg.ID,
		Version: srcMsg.Version,
		To:      srcMsg.To.String(),
		From:    srcMsg.From.String(),
		Nonce:   srcMsg.Nonce,
		//Value:      types.NewFromGo(srcMsg.Value.Int),
		GasLimit: srcMsg.GasLimit,
		//GasFeeCap:  toDeicimal(srcMsg.GasFeeCap),
		//GasPremium: toDeicimal(srcMsg.GasPremium),
		Method:    int(srcMsg.Method),
		Params:    srcMsg.Params,
		Signature: (*repo.SqlSignature)(srcMsg.Signature),
		//Cid:       srcMsg.UnsingedCid().String(),
		//SignedCid: srcMsg.SignedCid().String(),
		//	ExitCode:   repo.ExitCodeToExec,
		Height:  srcMsg.Height,
		Receipt: (&repo.SqlMsgReceipt{}).FromMsgReceipt(srcMsg.Receipt),
		Meta:    FromMeta(srcMsg.Meta),
		State:   srcMsg.State,
	}

	if srcMsg.UnsignedCid != nil {
		destMsg.UnsignedCid = srcMsg.UnsignedCid.String()
	}

	if srcMsg.SignedCid != nil {
		destMsg.SignedCid = srcMsg.SignedCid.String()
	}

	if srcMsg.Value.Int != nil {
		destMsg.Value = types.Int{Int: srcMsg.Value.Int}
	}

	if srcMsg.GasFeeCap.Int != nil {
		destMsg.GasFeeCap = types.Int{Int: srcMsg.GasFeeCap.Int}
	}

	if srcMsg.GasPremium.Int != nil {
		destMsg.GasPremium = types.Int{Int: srcMsg.GasPremium.Int}
	}

	return destMsg
}

type MsgMeta struct {
	ExpireEpoch       abi.ChainEpoch `gorm:"column:expire_epoch;type:bigint;"`
	GasOverEstimation float64        `gorm:"column:gas_over_estimation;type:decimal;"`
	MaxFee            types.Int      `gorm:"column:max_fee;type:varchar(256);"`
	MaxFeeCap         types.Int      `gorm:"column:max_fee_cap;type:varchar(256);"`
}

func (meta *MsgMeta) Meta() *types.MsgMeta {
	return &types.MsgMeta{
		ExpireEpoch:       meta.ExpireEpoch,
		GasOverEstimation: meta.GasOverEstimation,
		MaxFee:            big.NewFromGo(meta.MaxFee.Int),
		MaxFeeCap:         big.NewFromGo(meta.MaxFeeCap.Int),
	}
}

func FromMeta(srcMeta *types.MsgMeta) *MsgMeta {
	meta := &MsgMeta{
		ExpireEpoch:       srcMeta.ExpireEpoch,
		GasOverEstimation: srcMeta.GasOverEstimation,
	}

	if srcMeta.MaxFee.Int != nil {
		meta.MaxFee = types.Int{Int: srcMeta.MaxFee.Int}
	}

	if srcMeta.MaxFeeCap.Int != nil {
		meta.MaxFeeCap = types.Int{Int: srcMeta.MaxFeeCap.Int}
	}
	return meta
}

var _ repo.MessageRepo = (*sqliteMessageRepo)(nil)

type sqliteMessageRepo struct {
	*gorm.DB
}

func (m *sqliteMessageRepo) GetMessageState(uuid types.UUID) (types.MessageState, error) {
	type Result struct {
		State int
	}

	var result Result
	err := m.DB.Table("messages").
		Select("state").
		Where("name = ?", "Antonio").
		Scan(&result).Error
	if err != nil {
		return types.UnKnown, err
	}

	return types.MessageState(result.State), nil
}

func (m *sqliteMessageRepo) ExpireMessage(msgs []*types.Message) error {
	for _, msg := range msgs {
		err := m.DB.Table("messages").UpdateColumn("state", types.ExpireMsg).Where("id=?", msg.ID.String()).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *sqliteMessageRepo) ListUnChainMessageByAddress(addr address.Address) ([]*types.Message, error) {
	var sqlMsgs []*sqliteMessage
	err := m.DB.Find(&sqlMsgs, "from=? && state=?", addr.String(), types.UnFillMsg).Error
	if err != nil {
		return nil, err
	}
	result := make([]*types.Message, len(sqlMsgs))
	for index, sqlMsg := range sqlMsgs {
		result[index] = sqlMsg.Message()
	}
	return result, nil
}

//todo better batch update
func (m *sqliteMessageRepo) BatchSaveMessage(msgs []*types.Message) error {
	for _, msg := range msgs {
		_, err := m.SaveMessage(msg)
		if err != nil {
			return err
		}
	}
	return nil
}

func newSqliteMessageRepo(db *gorm.DB) *sqliteMessageRepo {
	return &sqliteMessageRepo{DB: db}
}

func (m *sqliteMessageRepo) SaveMessage(msg *types.Message) (types.UUID, error) {
	sqlMsg := FromMessage(msg)
	//todo check
	err := m.DB.Debug().Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: []clause.Assignment{
			{
				Column: clause.Column{
					Table: "messages",
					Name:  "updated_at",
				},
				Value: time.Now(),
			},
		}}).Save(sqlMsg).Error

	return msg.ID, err
}

func (m *sqliteMessageRepo) GetMessage(uuid types.UUID) (*types.Message, error) {
	var msg sqliteMessage
	if err := m.DB.Where("id = ?", uuid.String()).First(&msg).Error; err != nil {
		return nil, err
	}
	return msg.Message(), nil
}

func (m *sqliteMessageRepo) GetMessageByCid(cid string) (*types.Message, error) {
	var msg sqliteMessage
	if err := m.DB.Where("signed_cid = ?", cid).First(&msg).Error; err != nil {
		return nil, err
	}
	return msg.Message(), nil
}

func (m *sqliteMessageRepo) GetMessageByTime(start time.Time) ([]*types.Message, error) {
	var sqlMsgs []*sqliteMessage
	if err := m.DB.Where("created_at >= ?", start).Find(&sqlMsgs).Error; err != nil {
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
	if err := m.DB.Find(&sqlMsgs).Error; err != nil {
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
	if err := m.DB.Debug().Model((*sqliteMessage)(nil)).
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
		Height:  uint64(height),
		Receipt: (&repo.SqlMsgReceipt{}).FromMsgReceipt(receipt),
		State:   state,
	}
	return cid, m.DB.Debug().Model(&sqliteMessage{}).
		Where("signed_cid = ?", cid).
		Select("height", "exit_code", "receipt", "state").
		UpdateColumns(sqlMsg).Error
}

func (m *sqliteMessageRepo) UpdateMessageStateByCid(cid string, state types.MessageState) error {
	return m.DB.Debug().Model(&sqliteMessage{}).
		Where("signed_cid = ?", cid).UpdateColumn("state", state).Error
}
