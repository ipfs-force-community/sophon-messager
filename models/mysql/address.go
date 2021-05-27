package mysql

import (
	"context"
	"time"

	"github.com/filecoin-project/go-address"
	"gorm.io/gorm"

	"github.com/filecoin-project/venus-messager/models/repo"
	"github.com/filecoin-project/venus-messager/types"
)

type mysqlAddress struct {
	ID         types.UUID  `gorm:"column:id;type:varchar(256);primary_key"`
	Addr       string      `gorm:"column:addr;type:varchar(256);NOT NULL"` // 主键
	Nonce      uint64      `gorm:"column:nonce;type:bigint unsigned;index;NOT NULL"`
	Weight     int64       `gorm:"column:weight;type:bigint;index;NOT NULL"`
	SelMsgNum  uint64      `gorm:"column:sel_msg_num;type:bigint unsigned;NOT NULL"`
	State      types.State `gorm:"column:state;type:int;index;"`
	WalletName string      `gorm:"column:wallet_name;type:varchar(256);NOT NULL"`

	IsDeleted int       `gorm:"column:is_deleted;index;default:-1;NOT NULL"` // 是否删除 1:是  -1:否
	CreatedAt time.Time `gorm:"column:created_at;index;NOT NULL"`            // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;index;NOT NULL"`            // 更新时间
}

func (s mysqlAddress) TableName() string {
	return "addresses"
}

func FromAddress(addr *types.Address) *mysqlAddress {
	return &mysqlAddress{
		ID:         addr.ID,
		Addr:       addr.Addr.String(),
		Nonce:      addr.Nonce,
		Weight:     addr.Weight,
		SelMsgNum:  addr.SelMsgNum,
		State:      addr.State,
		WalletName: addr.WalletName,
		IsDeleted:  addr.IsDeleted,
		CreatedAt:  addr.CreatedAt,
		UpdatedAt:  addr.UpdatedAt,
	}
}

func (s mysqlAddress) Address() (*types.Address, error) {
	addr, err := address.NewFromString(s.Addr)
	if err != nil {
		return nil, err
	}
	return &types.Address{
		ID:         s.ID,
		Addr:       addr,
		Nonce:      s.Nonce,
		Weight:     s.Weight,
		SelMsgNum:  s.SelMsgNum,
		State:      s.State,
		WalletName: s.WalletName,
		IsDeleted:  s.IsDeleted,
		CreatedAt:  s.CreatedAt,
		UpdatedAt:  s.UpdatedAt,
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
	return s.DB.Save(FromAddress(a)).Error
}

func (s mysqlAddressRepo) GetAddress(ctx context.Context, walletName string, addr address.Address) (*types.Address, error) {
	var a mysqlAddress
	if err := s.DB.Take(&a, "wallet_name = ? and addr = ? and is_deleted = -1", walletName, addr.String()).Error; err != nil {
		return nil, err
	}

	return a.Address()
}

func (s mysqlAddressRepo) GetAddressByID(ctx context.Context, id types.UUID) (*types.Address, error) {
	var a mysqlAddress
	if err := s.DB.Where("id = ? and is_deleted = -1", id).First(&a).Error; err != nil {
		return nil, err
	}

	return a.Address()
}

func (s mysqlAddressRepo) GetOneRecord(ctx context.Context, walletName string, addr address.Address) (*types.Address, error) {
	var a mysqlAddress
	if err := s.DB.Take(&a, "wallet_name = ? and addr = ?", walletName, addr.String()).Error; err != nil {
		return nil, err
	}

	return a.Address()
}

func (s mysqlAddressRepo) HasAddress(ctx context.Context, walletName string, addr address.Address) (bool, error) {
	var count int64
	if err := s.DB.Model(&mysqlAddress{}).Where("wallet_name = ? and addr = ? and is_deleted = -1", walletName, addr.String()).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s mysqlAddressRepo) ListAddress(ctx context.Context) ([]*types.Address, error) {
	var list []*mysqlAddress
	if err := s.DB.Find(&list, "is_deleted = ?", -1).Error; err != nil {
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

func (s mysqlAddressRepo) DelAddress(ctx context.Context, walletName string, addr address.Address) error {
	return s.DB.Model((*mysqlAddress)(nil)).Where("wallet_name = ? and addr = ? and is_deleted = -1", walletName, addr.String()).
		UpdateColumns(map[string]interface{}{"is_deleted": repo.Deleted, "state": types.Removed, "updated_at": time.Now()}).Error
}

func (s mysqlAddressRepo) UpdateNonce(ctx context.Context, addr address.Address, nonce uint64) error {
	return s.DB.Model(&mysqlAddress{}).Where("addr = ? and is_deleted = -1", addr.String()).
		UpdateColumns(map[string]interface{}{"nonce": nonce, "updated_at": time.Now()}).Error
}

func (s mysqlAddressRepo) UpdateState(ctx context.Context, walletName string, addr address.Address, state types.State) error {
	return s.DB.Model(&mysqlAddress{}).Where("wallet_name = ? and addr = ? and is_deleted = -1", walletName, addr.String()).
		UpdateColumns(map[string]interface{}{"state": state, "updated_at": time.Now()}).Error
}

func (s mysqlAddressRepo) UpdateSelectMsgNum(ctx context.Context, walletName string, addr address.Address, num uint64) error {
	return s.DB.Model((*mysqlAddress)(nil)).Where("wallet_name = ? and addr = ? and is_deleted = -1", walletName, addr.String()).
		UpdateColumns(map[string]interface{}{"sel_msg_num": num, "updated_at": time.Now()}).Error
}
