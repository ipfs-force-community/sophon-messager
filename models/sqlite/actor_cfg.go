package sqlite

import (
	"context"
	"time"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"

	"github.com/filecoin-project/go-state-types/network"

	"github.com/filecoin-project/venus-messager/models/mtypes"

	shared "github.com/filecoin-project/venus/venus-shared/types"
	"gorm.io/gorm"

	"github.com/filecoin-project/venus-messager/models/repo"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
)

type sqliteActorCfg struct {
	ID      shared.UUID  `gorm:"column:id;type:varchar(256);primary_key;"` // 主键
	Nv      uint         `gorm:"column:network_v;type:INTEGER;NOT NULL"`
	CodeCid mtypes.DBCid `gorm:"column:code_cid;type:varchar(256);index:idx_code_cid_method,unique;NOT NULL"`
	Method  sqliteUint64 `gorm:"column:method;type:INTEGER;index:idx_code_cid_method,unique;NOT NULL"`

	SelectSpec

	CreatedAt time.Time `gorm:"column:created_at;index;NOT NULL"` // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;index;NOT NULL"` // 更新时间
}

func fromActorCfg(actorCfg *types.ActorCfg) *sqliteActorCfg {
	return &sqliteActorCfg{
		ID:      actorCfg.ID,
		Nv:      uint(actorCfg.NVersion),
		CodeCid: mtypes.NewDBCid(actorCfg.CodeCid),
		Method:  sqliteUint64(actorCfg.Method),
		SelectSpec: SelectSpec{
			SelMsgNum:         actorCfg.SelMsgNum,
			GasOverEstimation: actorCfg.GasOverEstimation,
			GasOverPremium:    actorCfg.GasOverPremium,
			MaxFee:            mtypes.SafeFromGo(actorCfg.MaxFee.Int),
			GasFeeCap:         mtypes.SafeFromGo(actorCfg.GasFeeCap.Int),
			BaseFee:           mtypes.SafeFromGo(actorCfg.BaseFee.Int),
		},
		CreatedAt: actorCfg.CreatedAt,
		UpdatedAt: actorCfg.UpdatedAt,
	}
}

func (sqliteActorCfg sqliteActorCfg) ActorCfg() *types.ActorCfg {
	return &types.ActorCfg{
		ID:       sqliteActorCfg.ID,
		NVersion: network.Version(sqliteActorCfg.Nv),
		MethodType: types.MethodType{
			CodeCid: sqliteActorCfg.CodeCid.Cid(),
			Method:  abi.MethodNum(sqliteActorCfg.Method),
		},
		SelectSpec: types.SelectSpec{
			SelMsgNum:         sqliteActorCfg.SelMsgNum,
			GasOverEstimation: sqliteActorCfg.GasOverEstimation,
			GasOverPremium:    sqliteActorCfg.GasOverPremium,
			MaxFee:            big.Int(mtypes.SafeFromGo(sqliteActorCfg.MaxFee.Int)),
			GasFeeCap:         big.Int(mtypes.SafeFromGo(sqliteActorCfg.GasFeeCap.Int)),
			BaseFee:           big.Int(mtypes.SafeFromGo(sqliteActorCfg.BaseFee.Int)),
		},
		CreatedAt: sqliteActorCfg.CreatedAt,
		UpdatedAt: sqliteActorCfg.UpdatedAt,
	}
}

func (sqliteActorCfg sqliteActorCfg) TableName() string {
	return "actor_cfg"
}

var _ repo.ActorCfgRepo = (*sqliteActorCfgRepo)(nil)

type sqliteActorCfgRepo struct {
	*gorm.DB
}

func newSqliteActorCfgRepo(db *gorm.DB) *sqliteActorCfgRepo {
	return &sqliteActorCfgRepo{DB: db}
}

func (s *sqliteActorCfgRepo) SaveActorCfg(ctx context.Context, actorCfg *types.ActorCfg) error {
	return s.DB.Save(fromActorCfg(actorCfg)).Error
}

func (s *sqliteActorCfgRepo) GetActorCfgByMethodType(ctx context.Context, methodType *types.MethodType) (*types.ActorCfg, error) {
	var a sqliteActorCfg
	if err := s.DB.Take(&a, "code_cid = ? and method = ?", methodType.CodeCid.String(), sqliteUint64(methodType.Method)).Error; err != nil {
		return nil, err
	}

	return a.ActorCfg(), nil
}

func (s *sqliteActorCfgRepo) GetActorCfgByID(ctx context.Context, id shared.UUID) (*types.ActorCfg, error) {
	var a sqliteActorCfg
	if err := s.DB.Take(&a, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return a.ActorCfg(), nil
}

func (s *sqliteActorCfgRepo) ListActorCfg(ctx context.Context) ([]*types.ActorCfg, error) {
	var list []*sqliteActorCfg
	if err := s.DB.Find(&list).Error; err != nil {
		return nil, err
	}

	result := make([]*types.ActorCfg, len(list))
	for index, r := range list {
		result[index] = r.ActorCfg()
	}

	return result, nil
}

func (s *sqliteActorCfgRepo) DelActorCfgByMethodType(ctx context.Context, methodType *types.MethodType) error {
	return s.DB.Delete(sqliteActorCfg{}, "code_cid = ? and method = ?", methodType.CodeCid.String(), sqliteUint64(methodType.Method)).Error
}

func (s *sqliteActorCfgRepo) DelActorCfgById(ctx context.Context, id shared.UUID) error {
	return s.DB.Delete(sqliteActorCfg{}, "id = ?", id).Error
}

func (s *sqliteActorCfgRepo) UpdateSelectSpecById(ctx context.Context, id shared.UUID, spec *types.ChangeSelectSpecParams) error {
	updateColumns := make(map[string]interface{}, 6)
	if !spec.GasFeeCap.Nil() {
		updateColumns["gas_fee_cap"] = spec.GasFeeCap.String()
	}
	if !spec.BaseFee.Nil() {
		updateColumns["base_fee"] = spec.BaseFee.String()
	}
	if !spec.MaxFee.Nil() {
		updateColumns["max_fee"] = spec.MaxFee.String()
	}

	if spec.SelMsgNum != nil {
		//todo unable to change value tozero, maybe need change type to pointer, but this pointer value not
		updateColumns["sel_msg_num"] = *spec.SelMsgNum
	}
	if spec.GasOverEstimation != nil {
		updateColumns["gas_over_estimation"] = *spec.GasOverEstimation
	}
	if spec.GasOverPremium != nil {
		updateColumns["gas_over_premium"] = *spec.GasOverPremium
	}

	if len(updateColumns) == 0 {
		return nil
	}

	updateColumns["updated_at"] = time.Now()

	return s.DB.Model((*sqliteActorCfg)(nil)).Where("id = ?", id).UpdateColumns(updateColumns).Error
}
