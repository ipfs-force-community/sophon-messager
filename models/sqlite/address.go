package sqlite

import (
	"context"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/big"
	shared "github.com/filecoin-project/venus/venus-shared/types"
	"gorm.io/gorm"

	types "github.com/filecoin-project/venus/venus-shared/types/messager"
	"github.com/ipfs-force-community/sophon-messager/models/mtypes"
	"github.com/ipfs-force-community/sophon-messager/models/repo"
)

type sqliteAddress struct {
	ID        shared.UUID        `gorm:"column:id;type:varchar(256);primary_key"`
	Addr      string             `gorm:"column:addr;type:varchar(256);uniqueIndex;NOT NULL"`
	Nonce     uint64             `gorm:"column:nonce;type:unsigned bigint;index;NOT NULL"`
	Weight    int64              `gorm:"column:weight;type:bigint;index;NOT NULL"`
	State     types.AddressState `gorm:"column:state;type:int;index;default:1"`
	SelMsgNum uint64             `gorm:"column:sel_msg_num;type:unsigned bigint;NOT NULL"`

	FeeSpec

	IsDeleted int       `gorm:"column:is_deleted;index;default:-1;NOT NULL"` // 是否删除 1:是  -1:否
	CreatedAt time.Time `gorm:"column:created_at;index;NOT NULL"`            // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;index;NOT NULL"`            // 更新时间
}

func (s sqliteAddress) TableName() string {
	return "addresses"
}

func fromAddress(addr *types.Address) *sqliteAddress {
	return &sqliteAddress{
		ID:        addr.ID,
		Addr:      addr.Addr.String(),
		Nonce:     addr.Nonce,
		Weight:    addr.Weight,
		State:     addr.State,
		SelMsgNum: addr.SelMsgNum,
		FeeSpec: FeeSpec{
			GasOverEstimation: addr.GasOverEstimation,
			GasOverPremium:    addr.GasOverPremium,
			MaxFee:            mtypes.SafeFromGo(addr.MaxFee.Int),
			GasFeeCap:         mtypes.SafeFromGo(addr.GasFeeCap.Int),
			BaseFee:           mtypes.SafeFromGo(addr.BaseFee.Int),
		},
		IsDeleted: addr.IsDeleted,
		CreatedAt: addr.CreatedAt,
		UpdatedAt: addr.UpdatedAt,
	}
}

func (s sqliteAddress) Address() (*types.Address, error) {
	addr, err := address.NewFromString(s.Addr)
	if err != nil {
		return nil, err
	}

	return &types.Address{
		ID:        s.ID,
		Addr:      addr,
		Nonce:     s.Nonce,
		Weight:    s.Weight,
		State:     s.State,
		SelMsgNum: s.SelMsgNum,
		FeeSpec: types.FeeSpec{
			GasOverEstimation: s.GasOverEstimation,
			GasOverPremium:    s.GasOverPremium,
			MaxFee:            big.Int(mtypes.SafeFromGo(s.MaxFee.Int)),
			GasFeeCap:         big.Int(mtypes.SafeFromGo(s.GasFeeCap.Int)),
			BaseFee:           big.Int(mtypes.SafeFromGo(s.BaseFee.Int)),
		},
		IsDeleted: s.IsDeleted,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}, nil
}

type sqliteAddressRepo struct {
	*gorm.DB
}

func newSqliteAddressRepo(db *gorm.DB) *sqliteAddressRepo {
	return &sqliteAddressRepo{DB: db}
}

func (s sqliteAddressRepo) SaveAddress(ctx context.Context, addr *types.Address) error {
	return s.DB.WithContext(ctx).Save(fromAddress(addr)).Error
}

func (s sqliteAddressRepo) GetAddress(ctx context.Context, addr address.Address) (*types.Address, error) {
	var a sqliteAddress
	if err := s.DB.WithContext(ctx).Take(&a, "addr = ? and is_deleted = -1", addr.String()).Error; err != nil {
		return nil, err
	}

	return a.Address()
}

func (s sqliteAddressRepo) GetAddressByID(ctx context.Context, id shared.UUID) (*types.Address, error) {
	var a sqliteAddress
	if err := s.DB.WithContext(ctx).Where("id = ? and is_deleted = -1", id).First(&a).Error; err != nil {
		return nil, err
	}

	return a.Address()
}

func (s sqliteAddressRepo) GetOneRecord(ctx context.Context, addr string) (*types.Address, error) {
	var a sqliteAddress
	if err := s.DB.WithContext(ctx).Take(&a, "addr = ?", addr).Error; err != nil {
		return nil, err
	}

	return a.Address()
}

func (s sqliteAddressRepo) HasAddress(ctx context.Context, addr address.Address) (bool, error) {
	var count int64
	if err := s.DB.WithContext(ctx).Model((*sqliteAddress)(nil)).
		Where("addr = ? and is_deleted = -1", addr.String()).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s sqliteAddressRepo) ListAddress(ctx context.Context) ([]*types.Address, error) {
	var list []*sqliteAddress
	if err := s.DB.WithContext(ctx).Find(&list, "is_deleted = ?", -1).Error; err != nil {
		return nil, err
	}

	result := make([]*types.Address, len(list))
	for index, r := range list {
		addr, err := r.Address()
		if err != nil {
			return nil, err
		}
		result[index] = addr
	}

	return result, nil
}

func (s sqliteAddressRepo) ListActiveAddress(ctx context.Context) ([]*types.Address, error) {
	var list []*sqliteAddress
	if err := s.DB.WithContext(ctx).Find(&list, "is_deleted = ? and state = ?", -1, types.AddressStateAlive).Error; err != nil {
		return nil, err
	}

	result := make([]*types.Address, len(list))
	for index, r := range list {
		addr, err := r.Address()
		if err != nil {
			return nil, err
		}
		result[index] = addr
	}

	return result, nil
}

func (s sqliteAddressRepo) UpdateNonce(addr address.Address, nonce uint64) (int64, error) {
	query := s.DB.Model((*sqliteAddress)(nil)).Where("addr = ? and is_deleted = -1", addr.String()).
		UpdateColumns(map[string]interface{}{"nonce": nonce, "updated_at": time.Now()})
	return query.RowsAffected, query.Error
}

func (s sqliteAddressRepo) UpdateState(ctx context.Context, addr address.Address, state types.AddressState) error {
	return s.DB.WithContext(ctx).Model((*sqliteAddress)(nil)).Where("addr = ? and is_deleted = -1", addr.String()).
		UpdateColumns(map[string]interface{}{"state": state, "updated_at": time.Now()}).Error
}

func (s sqliteAddressRepo) UpdateSelectMsgNum(ctx context.Context, addr address.Address, num uint64) error {
	return s.DB.WithContext(ctx).Model((*sqliteAddress)(nil)).Where("addr = ? and is_deleted = -1", addr.String()).
		UpdateColumns(map[string]interface{}{"sel_msg_num": num, "updated_at": time.Now()}).Error
}

func (s sqliteAddressRepo) UpdateFeeParams(ctx context.Context, addr address.Address, gasOverEstimation, gasOverPremium float64, maxFee, gasFeeCap, baseFee big.Int) error {
	updateColumns := make(map[string]interface{}, 6)
	if !maxFee.Nil() {
		updateColumns["max_fee"] = mtypes.NewFromGo(maxFee.Int)
	}
	if !gasFeeCap.Nil() {
		updateColumns["gas_fee_cap"] = mtypes.NewFromGo(gasFeeCap.Int)
	}
	if !baseFee.Nil() {
		updateColumns["base_fee"] = mtypes.NewFromGo(baseFee.Int)
	}
	if gasOverEstimation != 0 {
		updateColumns["gas_over_estimation"] = gasOverEstimation
	}
	if gasOverPremium != 0 {
		updateColumns["gas_over_premium"] = gasOverPremium
	}
	if len(updateColumns) == 0 {
		return nil
	}

	updateColumns["updated_at"] = time.Now()

	return s.DB.WithContext(ctx).Model((*sqliteAddress)(nil)).Where("addr = ? and is_deleted = -1", addr.String()).UpdateColumns(updateColumns).Error
}

func (s sqliteAddressRepo) DelAddress(ctx context.Context, addr string) error {
	return s.DB.WithContext(ctx).Where("addr = ?", addr).Delete(&sqliteAddress{}).Error
}

var _ repo.AddressRepo = &sqliteAddressRepo{}
