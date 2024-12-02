package mysql

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

type mysqlAddress struct {
	ID        shared.UUID        `gorm:"column:id;type:varchar(256);primary_key"`
	Addr      string             `gorm:"column:addr;type:varchar(256);uniqueIndex;NOT NULL"`
	Nonce     uint64             `gorm:"column:nonce;type:bigint unsigned;index;NOT NULL"`
	Weight    int64              `gorm:"column:weight;type:bigint;index;NOT NULL"`
	State     types.AddressState `gorm:"column:state;type:int;index;default:1"`
	SelMsgNum uint64             `gorm:"column:sel_msg_num;type:bigint unsigned;NOT NULL"`

	FeeSpec

	IsDeleted int       `gorm:"column:is_deleted;index;default:-1;NOT NULL"` // 是否删除 1:是  -1:否
	CreatedAt time.Time `gorm:"column:created_at;index;NOT NULL"`            // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;index;NOT NULL"`            // 更新时间
}

func (s mysqlAddress) TableName() string {
	return "addresses"
}

func fromAddress(addr *types.Address) *mysqlAddress {
	return &mysqlAddress{
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

func (s mysqlAddress) Address() (*types.Address, error) {
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

type mysqlAddressRepo struct {
	*gorm.DB
}

var _ repo.AddressRepo = &mysqlAddressRepo{}

func newMysqlAddressRepo(db *gorm.DB) *mysqlAddressRepo {
	return &mysqlAddressRepo{DB: db}
}

func (s mysqlAddressRepo) SaveAddress(ctx context.Context, a *types.Address) error {
	return s.DB.WithContext(ctx).Save(fromAddress(a)).Error
}

func (s mysqlAddressRepo) GetAddress(ctx context.Context, addr address.Address) (*types.Address, error) {
	var a mysqlAddress
	if err := s.DB.WithContext(ctx).Take(&a, "addr = ? and is_deleted = ?", addr.String(), repo.NotDeleted).Error; err != nil {
		return nil, err
	}

	return a.Address()
}

func (s mysqlAddressRepo) GetAddressByID(ctx context.Context, id shared.UUID) (*types.Address, error) {
	var a mysqlAddress
	if err := s.DB.WithContext(ctx).Where("id = ? and is_deleted = ?", id, repo.NotDeleted).First(&a).Error; err != nil {
		return nil, err
	}

	return a.Address()
}

func (s mysqlAddressRepo) GetOneRecord(ctx context.Context, addr address.Address) (*types.Address, error) {
	var a mysqlAddress
	if err := s.DB.WithContext(ctx).Take(&a, "addr = ?", addr.String()).Error; err != nil {
		return nil, err
	}

	return a.Address()
}

func (s mysqlAddressRepo) HasAddress(ctx context.Context, addr address.Address) (bool, error) {
	var count int64
	if err := s.DB.WithContext(ctx).Model(&mysqlAddress{}).Where("addr = ? and is_deleted = ?", addr.String(), repo.NotDeleted).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s mysqlAddressRepo) ListAddress(ctx context.Context) ([]*types.Address, error) {
	var list []*mysqlAddress
	if err := s.DB.WithContext(ctx).Find(&list, "is_deleted = ?", repo.NotDeleted).Error; err != nil {
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

func (s mysqlAddressRepo) ListActiveAddress(ctx context.Context) ([]*types.Address, error) {
	var list []*mysqlAddress
	if err := s.DB.WithContext(ctx).Find(&list, "is_deleted = ? and state = ?", repo.NotDeleted, types.AddressStateAlive).Error; err != nil {
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

func (s mysqlAddressRepo) DelAddress(ctx context.Context, addr address.Address) error {
	return s.DB.WithContext(ctx).Model((*mysqlAddress)(nil)).Where("addr = ? and is_deleted = ?", addr.String(), repo.NotDeleted).
		UpdateColumns(map[string]interface{}{"is_deleted": repo.Deleted, "state": types.AddressStateRemoved, "updated_at": time.Now()}).Error
}

func (s mysqlAddressRepo) UpdateNonce(addr address.Address, nonce uint64) (int64, error) {
	query := s.DB.Model(&mysqlAddress{}).Where("addr = ? and is_deleted = ?", addr.String(), repo.NotDeleted).
		UpdateColumns(map[string]interface{}{"nonce": nonce, "updated_at": time.Now()})
	return query.RowsAffected, query.Error
}

func (s mysqlAddressRepo) UpdateState(ctx context.Context, addr address.Address, state types.AddressState) error {
	return s.DB.WithContext(ctx).Model(&mysqlAddress{}).Where("addr = ? and is_deleted = ?", addr.String(), repo.NotDeleted).
		UpdateColumns(map[string]interface{}{"state": state, "updated_at": time.Now()}).Error
}

func (s mysqlAddressRepo) UpdateSelectMsgNum(ctx context.Context, addr address.Address, num uint64) error {
	return s.DB.WithContext(ctx).Model((*mysqlAddress)(nil)).Where("addr = ? and is_deleted = ?", addr.String(), repo.NotDeleted).
		UpdateColumns(map[string]interface{}{"sel_msg_num": num, "updated_at": time.Now()}).Error
}

func (s mysqlAddressRepo) UpdateFeeParams(ctx context.Context, addr address.Address, gasOverEstimation, gasOverPremium float64, maxFee, gasFeeCap, baseFee big.Int) error {
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

	return s.DB.WithContext(ctx).Model((*mysqlAddress)(nil)).Where("addr = ? and is_deleted = ?", addr.String(), repo.NotDeleted).UpdateColumns(updateColumns).Error
}
