package mysql

import (
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/crypto"
	venustypes "github.com/filecoin-project/venus/venus-shared/types"
	"github.com/ipfs/go-cid"
	"gorm.io/gorm"

	"github.com/filecoin-project/venus-messager/models/mtypes"
	"github.com/filecoin-project/venus-messager/models/repo"
	"github.com/filecoin-project/venus-messager/utils"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
)

type mysqlMessage struct {
	ID      string `gorm:"column:id;type:varchar(256);primary_key"`
	Version uint64 `gorm:"column:version;type:bigint unsigned"`

	From  string `gorm:"column:from_addr;type:varchar(256);NOT NULL;index:msg_from;index:idx_from_nonce;index:msg_from_state;index:idx_messages_create_at_state_from_addr;"`
	Nonce uint64 `gorm:"column:nonce;type:bigint unsigned;index:msg_nonce;index:idx_from_nonce"`
	To    string `gorm:"column:to;type:varchar(256);NOT NULL"`

	Value mtypes.Int `gorm:"column:value;type:varchar(256);"`

	GasLimit   int64      `gorm:"column:gas_limit;type:bigint"`
	GasFeeCap  mtypes.Int `gorm:"column:gas_fee_cap;type:varchar(256);"`
	GasPremium mtypes.Int `gorm:"column:gas_premium;type:varchar(256);"`

	Method int `gorm:"column:method;type:int"`

	Params []byte `gorm:"column:params;type:blob;"`

	Signature *repo.SqlSignature `gorm:"column:signed_data;type:blob;"`

	UnsignedCid string `gorm:"column:unsigned_cid;type:varchar(256);index:msg_unsigned_cid;"`
	SignedCid   string `gorm:"column:signed_cid;type:varchar(256);index:msg_signed_cid"`

	Height    int64               `gorm:"column:height;type:bigint;index:msg_height"`
	Receipt   *repo.SqlMsgReceipt `gorm:"embedded;embeddedPrefix:receipt_"`
	TipsetKey string              `gorm:"column:tipset_key;type:varchar(2048);"`

	Meta *mtypes.MsgMeta `gorm:"embedded;embeddedPrefix:meta_"`

	WalletName string `gorm:"column:wallet_name;type:varchar(256)"`
	FromUser   string `gorm:"column:from_user;type:varchar(256)"`

	State types.MessageState `gorm:"column:state;type:int;index:msg_state;index:msg_from_state;index:idx_messages_create_at_state_from_addr;"`

	IsDeleted int       `gorm:"column:is_deleted;index;default:-1;NOT NULL"`                                   // 是否删除 1:是  -1:否
	CreatedAt time.Time `gorm:"column:created_at;index;index:idx_messages_create_at_state_from_addr;NOT NULL"` // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;index;NOT NULL"`                                              // 更新时间
}

func (sqlMsg *mysqlMessage) TableName() string {
	return "messages"
}

func (sqlMsg *mysqlMessage) Message() *types.Message {
	var destMsg = &types.Message{
		ID: sqlMsg.ID,
		Message: venustypes.Message{
			Version:    sqlMsg.Version,
			Nonce:      sqlMsg.Nonce,
			Value:      big.NewFromGo(sqlMsg.Value.Int),
			GasLimit:   sqlMsg.GasLimit,
			GasFeeCap:  big.NewFromGo(sqlMsg.GasFeeCap.Int),
			GasPremium: big.NewFromGo(sqlMsg.GasPremium.Int),
			Method:     abi.MethodNum(sqlMsg.Method),
			Params:     sqlMsg.Params,
		},
		Height:     sqlMsg.Height,
		Receipt:    sqlMsg.Receipt.MsgReceipt(),
		Signature:  (*crypto.Signature)(sqlMsg.Signature),
		Meta:       sqlMsg.Meta.Meta(),
		WalletName: sqlMsg.WalletName,
		FromUser:   sqlMsg.FromUser,
		State:      sqlMsg.State,
		UpdatedAt:  sqlMsg.UpdatedAt,
		CreatedAt:  sqlMsg.CreatedAt,
	}
	destMsg.From, _ = address.NewFromString(sqlMsg.From)
	destMsg.To, _ = address.NewFromString(sqlMsg.To)
	if len(sqlMsg.UnsignedCid) > 0 {
		unsignedCid, _ := cid.Decode(sqlMsg.UnsignedCid)
		destMsg.UnsignedCid = &unsignedCid
	}
	if len(sqlMsg.SignedCid) > 0 {
		signedCid, _ := cid.Decode(sqlMsg.SignedCid)
		destMsg.SignedCid = &signedCid
	}
	if len(sqlMsg.TipsetKey) > 0 {
		destMsg.TipSetKey, _ = utils.StringToTipsetKey(sqlMsg.TipsetKey)
	}

	return destMsg
}

func fromMessage(srcMsg *types.Message) *mysqlMessage {
	destMsg := &mysqlMessage{
		ID:         srcMsg.ID,
		Version:    srcMsg.Version,
		To:         srcMsg.To.String(),
		From:       srcMsg.From.String(),
		Nonce:      srcMsg.Nonce,
		GasLimit:   srcMsg.GasLimit,
		Method:     int(srcMsg.Method),
		Params:     srcMsg.Params,
		Signature:  (*repo.SqlSignature)(srcMsg.Signature),
		Height:     srcMsg.Height,
		Receipt:    repo.FromMsgReceipt(srcMsg.Receipt),
		Meta:       mtypes.FromMeta(srcMsg.Meta),
		WalletName: srcMsg.WalletName,
		FromUser:   srcMsg.FromUser,
		State:      srcMsg.State,
		IsDeleted:  repo.NotDeleted,
	}

	if srcMsg.UnsignedCid != nil {
		destMsg.UnsignedCid = srcMsg.UnsignedCid.String()
	}

	if srcMsg.SignedCid != nil {
		destMsg.SignedCid = srcMsg.SignedCid.String()
	}

	if srcMsg.Value.Int != nil {
		destMsg.Value = mtypes.Int{Int: srcMsg.Value.Int}
	}

	if srcMsg.GasFeeCap.Int != nil {
		destMsg.GasFeeCap = mtypes.Int{Int: srcMsg.GasFeeCap.Int}
	}

	if srcMsg.GasPremium.Int != nil {
		destMsg.GasPremium = mtypes.Int{Int: srcMsg.GasPremium.Int}
	}

	if !srcMsg.TipSetKey.IsEmpty() {
		destMsg.TipsetKey = srcMsg.TipSetKey.String()
	}

	return destMsg
}

var _ repo.MessageRepo = (*mysqlMessageRepo)(nil)

type mysqlMessageRepo struct {
	*gorm.DB
}

func newMysqlMessageRepo(db *gorm.DB) *mysqlMessageRepo {
	return &mysqlMessageRepo{DB: db}
}

func (m *mysqlMessageRepo) ListMessageByFromState(from address.Address, state types.MessageState, isAsc bool, pageIndex, pageSize int) ([]*types.Message, error) {
	query := m.DB.Table("messages").Offset((pageIndex - 1) * pageSize).Limit(pageSize)

	if from != address.Undef {
		query = query.Where("from_addr=?", from.String())
	}

	if isAsc {
		query = query.Order("created_at ASC")
	} else {
		query = query.Order("created_at DESC")
	}

	query = query.Where("state=?", state)

	var sqlMsgs []*mysqlMessage
	err := query.Find(&sqlMsgs).Error
	if err != nil {
		return nil, err
	}

	result := make([]*types.Message, len(sqlMsgs))
	for index, sqlMsg := range sqlMsgs {
		result[index] = sqlMsg.Message()
	}
	return result, err
}

func (m *mysqlMessageRepo) HasMessageByUid(id string) (bool, error) {
	var count int64
	err := m.DB.Table("messages").Where("id=?", id).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (m *mysqlMessageRepo) GetMessageState(id string) (types.MessageState, error) {
	type Result struct {
		State int
	}

	var result Result
	err := m.DB.Table("messages").
		Select("state").
		Where("id = ?", id).
		Scan(&result).Error
	if err != nil {
		return types.UnKnown, err
	}

	return types.MessageState(result.State), nil
}

func (m *mysqlMessageRepo) ExpireMessage(msgs []*types.Message) error {
	for _, msg := range msgs {
		updateColumns := map[string]interface{}{
			"state":      types.FailedMsg,
			"updated_at": time.Now(),
		}
		err := m.DB.Table("messages").Where("id = ?", msg.ID).UpdateColumns(updateColumns).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *mysqlMessageRepo) ListFilledMessageByAddress(addr address.Address) ([]*types.Message, error) {
	var sqlMsgs []*mysqlMessage
	err := m.DB.Find(&sqlMsgs, "from_addr=? AND state=?", addr.String(), types.FillMsg).Error
	if err != nil {
		return nil, err
	}
	result := make([]*types.Message, len(sqlMsgs))
	for index, sqlMsg := range sqlMsgs {
		result[index] = sqlMsg.Message()
	}
	return result, nil
}

func (m *mysqlMessageRepo) ListFilledMessageBelowNonce(addr address.Address, nonce uint64) ([]*types.Message, error) {
	var sqlMsgs []*mysqlMessage
	err := m.DB.Find(&sqlMsgs, "from_addr=? AND state=? AND nonce <", addr.String(), types.FillMsg, nonce).Error
	if err != nil {
		return nil, err
	}
	result := make([]*types.Message, len(sqlMsgs))
	for index, sqlMsg := range sqlMsgs {
		result[index] = sqlMsg.Message()
	}
	return result, nil
}

func (m *mysqlMessageRepo) ListFilledMessageByHeight(height abi.ChainEpoch) ([]*types.Message, error) {
	var sqlMsgs []*mysqlMessage
	err := m.DB.Find(&sqlMsgs, "height=? AND state=?", height, types.FillMsg).Error
	if err != nil {
		return nil, err
	}
	result := make([]*types.Message, len(sqlMsgs))
	for index, sqlMsg := range sqlMsgs {
		result[index] = sqlMsg.Message()
	}
	return result, nil
}

func (m *mysqlMessageRepo) ListUnChainMessageByAddress(addr address.Address, topN int) ([]*types.Message, error) {
	var sqlMsgs []*mysqlMessage
	err := m.DB.Limit(topN).Order("created_at").Find(&sqlMsgs, "from_addr=? AND state=?", addr.String(), types.UnFillMsg).Error
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
func (m *mysqlMessageRepo) BatchSaveMessage(msgs []*types.Message) error {
	for _, msg := range msgs {
		err := m.SaveMessage(msg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *mysqlMessageRepo) CreateMessage(msg *types.Message) error {
	sqlMsg := fromMessage(msg)
	sqlMsg.CreatedAt = time.Now()
	sqlMsg.UpdatedAt = time.Now()
	return m.DB.Create(sqlMsg).Error
}

func (m *mysqlMessageRepo) SaveMessage(msg *types.Message) error {
	sqlMsg := fromMessage(msg)
	sqlMsg.UpdatedAt = time.Now()
	return m.DB.Omit("created_at").Save(sqlMsg).Error
}

func (m *mysqlMessageRepo) GetMessageByUid(id string) (*types.Message, error) {
	var msg mysqlMessage
	if err := m.DB.Where("id = ?", id).Take(&msg).Error; err != nil {
		return nil, err
	}
	return msg.Message(), nil
}

func (m *mysqlMessageRepo) GetMessageByCid(unsignedCid cid.Cid) (*types.Message, error) {
	var msg mysqlMessage
	if err := m.DB.Where("unsigned_cid = ?", unsignedCid.String()).Take(&msg).Error; err != nil {
		return nil, err
	}
	return msg.Message(), nil
}

func (m *mysqlMessageRepo) GetMessageBySignedCid(signedCid cid.Cid) (*types.Message, error) {
	var msg mysqlMessage
	if err := m.DB.Where("signed_cid = ?", signedCid.String()).Take(&msg).Error; err != nil {
		return nil, err
	}
	return msg.Message(), nil
}

func (m *mysqlMessageRepo) GetSignedMessageByTime(start time.Time) ([]*types.Message, error) {
	var sqlMsgs []*mysqlMessage
	if err := m.DB.Where("created_at >= ? and signed_data is not null", start).Find(&sqlMsgs).Error; err != nil {
		return nil, err
	}
	result := make([]*types.Message, len(sqlMsgs))
	for idx, msg := range sqlMsgs {
		result[idx] = msg.Message()
	}

	return result, nil
}

func (m *mysqlMessageRepo) GetSignedMessageByHeight(height abi.ChainEpoch) ([]*types.Message, error) {
	var sqlMsgs []*mysqlMessage
	if err := m.DB.Where("height >= ? and signed_data is not null", uint64(height)).Find(&sqlMsgs).Error; err != nil {
		return nil, err
	}
	result := make([]*types.Message, len(sqlMsgs))
	for idx, msg := range sqlMsgs {
		result[idx] = msg.Message()
	}

	return result, nil
}

func (m *mysqlMessageRepo) GetSignedMessageFromFailedMsg(addr address.Address) ([]*types.Message, error) {
	var sqlMsgs []*mysqlMessage
	if err := m.DB.Where("state = ? and from_addr = ? and signed_data is not null", types.FailedMsg, addr.String()).Find(&sqlMsgs).Error; err != nil {
		return nil, err
	}
	result := make([]*types.Message, len(sqlMsgs))
	for idx, msg := range sqlMsgs {
		result[idx] = msg.Message()
	}

	return result, nil
}

func (m *mysqlMessageRepo) GetMessageByFromAndNonce(from address.Address, nonce uint64) (*types.Message, error) {
	var msg mysqlMessage
	if err := m.DB.Where("from_addr = ? and nonce = ?", from.String(), nonce).Take(&msg).Error; err != nil {
		return nil, err
	}
	return msg.Message(), nil
}

func (m *mysqlMessageRepo) GetMessageByFromNonceAndState(from address.Address, nonce uint64, state types.MessageState) (*types.Message, error) {
	var msg mysqlMessage
	if err := m.DB.Where("from_addr = ? and nonce = ? and state = ?", from.String(), nonce, state).Take(&msg).Error; err != nil {
		return nil, err
	}
	return msg.Message(), nil
}

func (m *mysqlMessageRepo) ListMessage() ([]*types.Message, error) {
	var sqlMsgs []*mysqlMessage
	if err := m.DB.Find(&sqlMsgs).Error; err != nil {
		return nil, err
	}

	result := make([]*types.Message, len(sqlMsgs))
	for idx, msg := range sqlMsgs {
		result[idx] = msg.Message()
	}
	return result, nil
}

func (m *mysqlMessageRepo) ListMessageByAddress(addr address.Address) ([]*types.Message, error) {
	var sqlMsgs []*mysqlMessage
	if err := m.DB.Find(&sqlMsgs, "from_addr=?", addr.String()).Error; err != nil {
		return nil, err
	}

	result := make([]*types.Message, len(sqlMsgs))
	for idx, msg := range sqlMsgs {
		result[idx] = msg.Message()
	}
	return result, nil
}

func (m *mysqlMessageRepo) ListFailedMessage() ([]*types.Message, error) {
	var sqlMsgs []*mysqlMessage
	err := m.DB.Order("created_at").Find(&sqlMsgs, "state = ? AND receipt_return_value is not null", types.UnFillMsg).Error
	if err != nil {
		return nil, err
	}
	result := make([]*types.Message, len(sqlMsgs))
	for index, sqlMsg := range sqlMsgs {
		result[index] = sqlMsg.Message()
	}
	return result, nil
}

func (m *mysqlMessageRepo) ListBlockedMessage(addr address.Address, d time.Duration) ([]*types.Message, error) {
	var sqlMsgs []*mysqlMessage
	t := time.Now().Add(-d)
	err := m.DB.Order("created_at").Find(&sqlMsgs, "from_addr = ? AND state = ? AND created_at < ?", addr.String(), types.FillMsg, t).Error
	if err != nil {
		return nil, err
	}
	result := make([]*types.Message, len(sqlMsgs))
	for index, sqlMsg := range sqlMsgs {
		result[index] = sqlMsg.Message()
	}
	return result, nil
}

func (m *mysqlMessageRepo) ListUnFilledMessage(addr address.Address) ([]*types.Message, error) {
	var sqlMsgs []*mysqlMessage
	if err := m.DB.Model((*mysqlMessage)(nil)).
		Find(&sqlMsgs, "from_addr = ? AND state = ?", addr.String(), types.UnFillMsg).Error; err != nil {
		return nil, err
	}

	var result = make([]*types.Message, len(sqlMsgs))

	for idx, msg := range sqlMsgs {
		result[idx] = msg.Message()
	}
	return result, nil
}

func (m *mysqlMessageRepo) ListSignedMsgs() ([]*types.Message, error) {
	var sqlMsgs []*mysqlMessage
	if err := m.DB.Model((*mysqlMessage)(nil)).
		Where("height=0 and signed_data is not null").
		Find(&sqlMsgs).Error; err != nil {
		return nil, err
	}

	var result = make([]*types.Message, len(sqlMsgs))

	for idx, msg := range sqlMsgs {
		result[idx] = msg.Message()
	}
	return result, nil
}

func (m *mysqlMessageRepo) UpdateMessageInfoByCid(unsignedCid string,
	receipt *venustypes.MessageReceipt,
	height abi.ChainEpoch,
	state types.MessageState,
	tsKey venustypes.TipSetKey) error {
	rcp := repo.FromMsgReceipt(receipt)
	updateClause := map[string]interface{}{
		"height":               uint64(height),
		"receipt_exit_code":    rcp.ExitCode,
		"receipt_return_value": rcp.Return,
		"receipt_gas_used":     rcp.GasUsed,
		"state":                state,
		"tipset_key":           tsKey.String(),
		"updated_at":           time.Now(),
	}
	return m.DB.Model(&mysqlMessage{}).
		Where("unsigned_cid = ?", unsignedCid).
		UpdateColumns(updateClause).Error
}

func (m *mysqlMessageRepo) UpdateMessageStateByCid(cid string, state types.MessageState) error {
	updateColumns := map[string]interface{}{
		"state":      state,
		"updated_at": time.Now(),
	}
	return m.DB.Model(&mysqlMessage{}).
		Where("unsigned_cid = ?", cid).UpdateColumns(updateColumns).Error
}

func (m *mysqlMessageRepo) UpdateMessageStateByID(id string, state types.MessageState) error {
	updateColumns := map[string]interface{}{
		"state":      state,
		"updated_at": time.Now(),
	}
	return m.DB.Debug().Model(&mysqlMessage{}).
		Where("id = ?", id).UpdateColumns(updateColumns).Error
}

func (m *mysqlMessageRepo) MarkBadMessage(id string) error {
	updateColumns := map[string]interface{}{
		"state":      types.FailedMsg,
		"updated_at": time.Now(),
	}
	return m.DB.Debug().Model(&mysqlMessage{}).Where("id = ?", id).UpdateColumns(updateColumns).Error
}

func (m *mysqlMessageRepo) UpdateReturnValue(id string, returnVal string) error {
	updateColumns := map[string]interface{}{
		"receipt_return_value": returnVal,
		"updated_at":           time.Now(),
	}
	return m.DB.Model((*mysqlMessage)(nil)).Where("id = ?", id).UpdateColumns(updateColumns).Error
}
