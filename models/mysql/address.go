package mysql

import (
	"context"
	"github.com/filecoin-project/go-address"
	"github.com/ipfs-force-community/venus-messager/models/repo"
	"gorm.io/gorm"
	"reflect"
	"time"

	"github.com/hunjixin/automapper"

	"github.com/ipfs-force-community/venus-messager/types"
)

type mysqlAddress struct {
	ID       types.UUID         `gorm:"column:id;type:varchar(256);primary_key"`
	Addr     string             `gorm:"column:addr;type:varchar(256);uniqueIndex;NOT NULL"` // 主键
	Nonce    uint64             `gorm:"column:nonce;type:bigint unsigned;index;NOT NULL"`
	Weight   int64              `gorm:"column:weight;type:bigint;index;NOT NULL"`
	WalletID types.UUID         `gorm:"column:wallet_id;type:varchar(256)"`
	State    types.AddressState `gorm:"column:state;type:int;index:addr_state;"`

	IsDeleted int       `gorm:"column:is_deleted;index;default:-1;NOT NULL"` // 是否删除 1:是  -1:否
	CreatedAt time.Time `gorm:"column:created_at;index;NOT NULL"`            // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;index;NOT NULL"`            // 更新时间
}

func (s mysqlAddress) TableName() string {
	return "addresses"
}

func FromAddress(addr *types.Address) *mysqlAddress {
	return &mysqlAddress{
		ID:        addr.ID,
		Addr:      addr.Addr.String(),
		Nonce:     addr.Nonce,
		Weight:    addr.Weight,
		WalletID:  addr.WalletID,
		State:     addr.State,
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
		WalletID:  s.WalletID,
		State:     s.State,
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

func (s mysqlAddressRepo) SaveAddress(ctx context.Context, addr *types.Address) (types.UUID, error) {
	err := s.DB.Save(FromAddress(addr)).Error
	if err != nil {
		return types.UUID{}, err
	}

	return addr.ID, nil
}

func (s mysqlAddressRepo) UpdateAddress(ctx context.Context, addr *types.Address) error {
	return s.DB.Model(&mysqlAddress{}).Where("addr = ?", addr.Addr).
		Updates(map[string]interface{}{"nonce": addr.Nonce, "is_deleted": addr.IsDeleted, "state": addr.State, "wallet_id": addr.WalletID}).Error
}

func (s mysqlAddressRepo) UpdateNonce(ctx context.Context, addr address.Address, nonce uint64) (address.Address, error) {
	return addr, s.DB.Model(&mysqlAddress{}).Where("addr = ?", addr.String()).UpdateColumn("nonce", nonce).Error
}

func (s mysqlAddressRepo) UpdateAddressState(ctx context.Context, addr address.Address, state types.AddressState) (address.Address, error) {
	return addr, s.DB.Model(&mysqlAddress{}).Where("addr = ?", addr.String()).UpdateColumn("state", state).Error
}

func (s mysqlAddressRepo) HasAddress(ctx context.Context, addr address.Address) (bool, error) {
	var count int64
	err := s.DB.Model(&mysqlAddress{}).Where("addr=?", addr.String()).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s mysqlAddressRepo) GetAddress(ctx context.Context, addr address.Address) (*types.Address, error) {
	var a mysqlAddress
	if err := s.DB.Where("addr = ? and is_deleted = -1", addr.String()).First(&a).Error; err != nil {
		return nil, err
	}

	return a.Address()
}

func (s mysqlAddressRepo) DelAddress(ctx context.Context, addr address.Address) error {
	var a mysqlAddress
	if err := s.DB.Where("addr = ? and is_deleted = -1", addr.String()).First(&a).Error; err != nil {
		return err
	}
	a.IsDeleted = 1
	a.State = types.Removed

	return s.DB.Save(&a).Error
}

func (s mysqlAddressRepo) ListAddress(ctx context.Context) ([]*types.Address, error) {
	var list []*mysqlAddress
	if err := s.DB.Find(&list, "is_deleted = ?", -1).Error; err != nil {
		return nil, err
	}

	result, err := automapper.Mapper(list, reflect.TypeOf([]*types.Address{}))
	if err != nil {
		return nil, err
	}

	return result.([]*types.Address), nil
}
