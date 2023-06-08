package sqlite

import (
	"context"
	"errors"
	"time"

	"github.com/ipfs/go-cid"

	"github.com/filecoin-project/go-state-types/actors"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"

	"github.com/ipfs-force-community/sophon-messager/models/mtypes"

	shared "github.com/filecoin-project/venus/venus-shared/types"
	"gorm.io/gorm"

	types "github.com/filecoin-project/venus/venus-shared/types/messager"
	"github.com/ipfs-force-community/sophon-messager/models/repo"
)

type sqliteActorCfg struct {
	ID           shared.UUID  `gorm:"column:id;type:varchar(256);primary_key;"` // 主键
	ActorVersion int          `gorm:"column:actor_v;type:INTEGER;NOT NULL"`
	Code         mtypes.DBCid `gorm:"column:code;type:varchar(256);index:idx_code_method,unique;NOT NULL"`
	Method       sqliteUint64 `gorm:"column:method;type:INTEGER;index:idx_code_method,unique;NOT NULL"`

	FeeSpec

	CreatedAt time.Time `gorm:"column:created_at;index;NOT NULL"` // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;index;NOT NULL"` // 更新时间
}

func fromActorCfg(actorCfg *types.ActorCfg) *sqliteActorCfg {
	return &sqliteActorCfg{
		ID:           actorCfg.ID,
		ActorVersion: int(actorCfg.ActorVersion),
		Code:         mtypes.NewDBCid(actorCfg.Code),
		Method:       sqliteUint64(actorCfg.Method),
		FeeSpec: FeeSpec{
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
		ID:           sqliteActorCfg.ID,
		ActorVersion: actors.Version(sqliteActorCfg.ActorVersion),
		MethodType: types.MethodType{
			Code:   sqliteActorCfg.Code.Cid(),
			Method: abi.MethodNum(sqliteActorCfg.Method),
		},
		FeeSpec: types.FeeSpec{
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
	if actorCfg.Code == cid.Undef {
		return errors.New("code cid is undefined")
	}
	return s.DB.Save(fromActorCfg(actorCfg)).Error
}

func (s *sqliteActorCfgRepo) HasActorCfg(ctx context.Context, methodType *types.MethodType) (bool, error) {
	var count int64
	if err := s.DB.Table("actor_cfg").Where("code = ? and method = ?", mtypes.NewDBCid(methodType.Code),
		methodType.Method).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (s *sqliteActorCfgRepo) GetActorCfgByMethodType(ctx context.Context, methodType *types.MethodType) (*types.ActorCfg, error) {
	var a sqliteActorCfg
	if err := s.DB.Take(&a, "code = ? and method = ?", mtypes.DBCid(methodType.Code), sqliteUint64(methodType.Method)).Error; err != nil {
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
	return s.DB.Delete(sqliteActorCfg{}, "code = ? and method = ?", mtypes.DBCid(methodType.Code), sqliteUint64(methodType.Method)).Error
}

func (s *sqliteActorCfgRepo) DelActorCfgById(ctx context.Context, id shared.UUID) error {
	return s.DB.Delete(sqliteActorCfg{}, "id = ?", id).Error
}

func (s *sqliteActorCfgRepo) UpdateSelectSpecById(ctx context.Context, id shared.UUID, spec *types.ChangeGasSpecParams) error {
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
