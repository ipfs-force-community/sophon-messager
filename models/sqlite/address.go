package sqlite

import (
	"context"
	"reflect"
	"time"

	"github.com/filecoin-project/go-address"

	"gorm.io/gorm"

	"github.com/hunjixin/automapper"

	"github.com/ipfs-force-community/venus-messager/models/repo"
	"github.com/ipfs-force-community/venus-messager/types"
)

type sqliteAddress struct {
	ID       types.UUID         `gorm:"column:id;type:varchar(256);primary_key"`
	Addr     string             `gorm:"column:addr;type:varchar(256);uniqueIndex;NOT NULL"` // 主键
	Nonce    uint64             `gorm:"column:nonce;type:unsigned bigint;index;NOT NULL"`
	Weight   int64              `gorm:"column:weight;type:bigint;index;NOT NULL"`
	WalletID types.UUID         `gorm:"column:wallet_id;type:varchar(256)"`
	State    types.AddressState `gorm:"column:state;type:int;index:addr_state;"`

	IsDeleted int       `gorm:"column:is_deleted;index;default:-1;NOT NULL"` // 是否删除 1:是  -1:否
	CreatedAt time.Time `gorm:"column:created_at;index;NOT NULL"`            // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;index;NOT NULL"`            // 更新时间
}

func (s sqliteAddress) TableName() string {
	return "addresses"
}

func FromAddress(address *types.Address) *sqliteAddress {
	return automapper.MustMapper(address, TSqliteAddress).(*sqliteAddress)
}

func (s sqliteAddress) Address() *types.Address {
	return automapper.MustMapper(&s, TAddress).(*types.Address)
}

type sqliteAddressRepo struct {
	*gorm.DB
}

func newSqliteAddressRepo(db *gorm.DB) *sqliteAddressRepo {
	return &sqliteAddressRepo{DB: db}
}

func (s sqliteAddressRepo) SaveAddress(ctx context.Context, addr *types.Address) (types.UUID, error) {
	err := s.DB.Save(FromAddress(addr)).Error
	if err != nil {
		return types.UUID{}, err
	}

	return addr.ID, nil
}

func (s sqliteAddressRepo) UpdateAddress(ctx context.Context, addr *types.Address) error {
	return s.DB.Model(&sqliteAddress{}).Where("addr = ?", addr.Addr).
		Updates(map[string]interface{}{"nonce": addr.Nonce, "is_deleted": addr.IsDeleted, "state": addr.State, "wallet_id": addr.WalletID}).Error
}

func (s sqliteAddressRepo) UpdateNonce(ctx context.Context, addr address.Address, nonce uint64) (address.Address, error) {
	return addr, s.DB.Model(&sqliteAddress{}).Where("addr = ?", addr.String()).UpdateColumn("nonce", nonce).Error
}

func (s sqliteAddressRepo) UpdateAddressState(ctx context.Context, addr address.Address, state types.AddressState) (address.Address, error) {
	return addr, s.DB.Model(&sqliteAddress{}).Where("addr = ?", addr.String()).UpdateColumn("state", state).Error
}

func (s sqliteAddressRepo) HasAddress(ctx context.Context, addr address.Address) (bool, error) {
	var count int64
	err := s.DB.Model(&sqliteAddress{}).Where("addr=?", addr.String()).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s sqliteAddressRepo) GetAddress(ctx context.Context, addr address.Address) (*types.Address, error) {
	var a sqliteAddress
	if err := s.DB.Where("addr = ? and is_deleted = -1", addr.String()).First(&a).Error; err != nil {
		return nil, err
	}

	return a.Address(), nil
}

func (s sqliteAddressRepo) DelAddress(ctx context.Context, addr address.Address) error {
	var a sqliteAddress
	if err := s.DB.Where("addr = ? and is_deleted = -1", addr.String()).First(&a).Error; err != nil {
		return err
	}
	a.IsDeleted = 1
	a.State = types.Removed

	return s.DB.Save(&a).Error
}

func (s sqliteAddressRepo) ListAddress(ctx context.Context) ([]*types.Address, error) {
	var list []*sqliteAddress
	if err := s.DB.Find(&list, "is_deleted = ?", -1).Error; err != nil {
		return nil, err
	}

	result, err := automapper.Mapper(list, reflect.TypeOf([]*types.Address{}))
	if err != nil {
		return nil, err
	}

	return result.([]*types.Address), nil
}

var _ repo.AddressRepo = &sqliteAddressRepo{}
