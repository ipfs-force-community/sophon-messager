package mysql

import (
	"context"
	"reflect"
	"time"

	"github.com/hunjixin/automapper"

	"github.com/ipfs-force-community/venus-messager/models/repo"
	"github.com/ipfs-force-community/venus-messager/types"
)

type mysqlAddress struct {
	Addr  string `gorm:"column:addr;primary_key;NOT NULL"json:"id"` // 主键
	Nonce uint64 `gorm:"column:nonce;"json:"nonce"`

	IsDeleted int       `gorm:"column:is_deleted;default:-1;NOT NULL"`                // 是否删除 1:是  -1:否
	CreatedAt time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP;NOT NULL"` // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP;NOT NULL"` // 更新时间
}

func (s mysqlAddress) TableName() string {
	return "addresses"
}

func FromAddress(address *types.Address) *mysqlAddress {
	return automapper.MustMapper(address, TMysqlAddress).(*mysqlAddress)
}

func (s mysqlAddress) Address() *types.Address {
	return automapper.MustMapper(&s, TAddress).(*types.Address)
}

type mysqlAddressRepo struct {
	repo.Repo
}

func newMysqlAddressRepo(repo repo.Repo) *mysqlAddressRepo {
	return &mysqlAddressRepo{repo}
}

func (s mysqlAddressRepo) SaveAddress(ctx context.Context, address *types.Address) (string, error) {
	return address.Addr, s.GetDb().Save(FromAddress(address)).Error
}

func (s mysqlAddressRepo) GetAddress(ctx context.Context, addr string) (*types.Address, error) {
	var a mysqlAddress
	if err := s.GetDb().Where(&mysqlAddress{
		Addr:      addr,
		IsDeleted: -1,
	}).First(&a).Error; err != nil {
		return nil, err
	}

	return a.Address(), nil
}

func (s mysqlAddressRepo) DelAddress(ctx context.Context, addr string) error {
	var a mysqlAddress
	if err := s.GetDb().Where(&mysqlAddress{
		Addr:      addr,
		IsDeleted: -1,
	}).First(&a).Error; err != nil {
		return err
	}
	a.IsDeleted = 1

	return s.GetDb().Save(&a).Error
}

func (s mysqlAddressRepo) ListAddress(ctx context.Context) ([]*types.Address, error) {
	var list []*mysqlAddress
	if err := s.GetDb().Find(&list, "is_deleted = ?", -1).Error; err != nil {
		return nil, err
	}

	result, err := automapper.Mapper(list, reflect.TypeOf([]*types.Address{}))
	if err != nil {
		return nil, err
	}

	return result.([]*types.Address), nil
}

var _ repo.AddressRepo = &mysqlAddressRepo{}
