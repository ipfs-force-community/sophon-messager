package mysql

import (
	"reflect"
	"time"

	"github.com/hunjixin/automapper"
	"gorm.io/gorm"

	"github.com/filecoin-project/venus-messager/models/repo"
	"github.com/filecoin-project/venus-messager/types"
)

type mysqlWalletAddress struct {
	ID           types.UUID  `gorm:"column:id;type:varchar(256);primary_key"`
	WalletID     types.UUID  `gorm:"column:wallet_id;type:varchar(256);NOT NULL"`
	AddrID       types.UUID  `gorm:"column:addr_id;type:varchar(256);NOT NULL"`
	AddressState types.State `gorm:"column:addr_state;type:int;index:wallet_addr_state;"`
	SelMsgNum    uint64      `gorm:"column:sel_msg_num;type:bigint unsigned;NOT NULL"`

	IsDeleted int       `gorm:"column:is_deleted;index;default:-1;NOT NULL"` // 是否删除 1:是  -1:否
	CreatedAt time.Time `gorm:"column:created_at;index;NOT NULL"`            // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;index;NOT NULL"`            // 更新时间
}

func FromWalletAddress(walletAddr types.WalletAddress) *mysqlWalletAddress {
	return automapper.MustMapper(&walletAddr, TMysqlWalletAddress).(*mysqlWalletAddress)
}

func (mysqlWalletAddress mysqlWalletAddress) WalletAddress() *types.WalletAddress {
	return automapper.MustMapper(&mysqlWalletAddress, TWalletAddress).(*types.WalletAddress)
}

func (mysqlWalletAddress mysqlWalletAddress) TableName() string {
	return "wallet_addresses"
}

var _ repo.WalletAddressRepo = (*mysqlWalletAddressRepo)(nil)

type mysqlWalletAddressRepo struct {
	*gorm.DB
}

func newMysqlWalletAddressRepo(db *gorm.DB) mysqlWalletAddressRepo {
	return mysqlWalletAddressRepo{DB: db}
}

func (s mysqlWalletAddressRepo) SaveWalletAddress(wa *types.WalletAddress) error {
	mysqlWalletAddress := FromWalletAddress(*wa)
	mysqlWalletAddress.UpdatedAt = time.Now()
	return s.DB.Save(mysqlWalletAddress).Error
}

func (s mysqlWalletAddressRepo) GetWalletAddress(walletID, addrID types.UUID) (*types.WalletAddress, error) {
	var wa mysqlWalletAddress
	if err := s.DB.Where("wallet_id = ? and addr_id = ? and is_deleted = -1", walletID, addrID).
		First(&wa).Error; err != nil {
		return nil, err
	}
	return wa.WalletAddress(), nil
}

func (s mysqlWalletAddressRepo) GetOneRecord(walletID, addrID types.UUID) (*types.WalletAddress, error) {
	var wa mysqlWalletAddress
	if err := s.DB.Where("wallet_id = ? and addr_id = ?", walletID, addrID).First(&wa).Error; err != nil {
		return nil, err
	}
	return wa.WalletAddress(), nil
}

func (s mysqlWalletAddressRepo) GetWalletAddressByWalletID(walletID types.UUID) ([]*types.WalletAddress, error) {
	var internalWalletAddress []*mysqlWalletAddress
	if err := s.DB.Find(&internalWalletAddress, "wallet_id = ? and is_deleted = ?", walletID, -1).Error; err != nil {
		return nil, err
	}

	result, err := automapper.Mapper(internalWalletAddress, reflect.TypeOf([]*types.WalletAddress{}))
	if err != nil {
		return nil, err
	}

	return result.([]*types.WalletAddress), nil
}

func (s mysqlWalletAddressRepo) HasWalletAddress(walletID, addrID types.UUID) (bool, error) {
	var count int64
	if err := s.DB.Model(&mysqlWalletAddress{}).Where("wallet_id = ? and addr_id = ?", walletID, addrID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s mysqlWalletAddressRepo) ListWalletAddress() ([]*types.WalletAddress, error) {
	var internalWalletAddress []*mysqlWalletAddress
	if err := s.DB.Find(&internalWalletAddress, "is_deleted = ?", -1).Error; err != nil {
		return nil, err
	}

	result, err := automapper.Mapper(internalWalletAddress, reflect.TypeOf([]*types.WalletAddress{}))
	if err != nil {
		return nil, err
	}

	return result.([]*types.WalletAddress), nil
}

func (s mysqlWalletAddressRepo) UpdateAddressState(walletID, addrID types.UUID, state types.State) error {
	return s.DB.Model((*mysqlWalletAddress)(nil)).Where("wallet_id = ? and addr_id = ?", walletID, addrID).
		UpdateColumn("addr_state", state).Error
}

func (s mysqlWalletAddressRepo) UpdateSelectMsgNum(walletID, addrID types.UUID, selMsgNum uint64) error {
	return s.DB.Model((*mysqlWalletAddress)(nil)).Where("wallet_id = ? and addr_id = ?", walletID, addrID).
		UpdateColumn("sel_msg_num", selMsgNum).Error
}

func (s mysqlWalletAddressRepo) DelWalletAddress(walletID, addrID types.UUID) error {
	var wa mysqlWalletAddress
	if err := s.DB.Where("wallet_id = ? and addr_id = ? and is_deleted = -1", walletID, addrID).
		First(&wa).Error; err != nil {
		return err
	}
	wa.IsDeleted = repo.Deleted
	wa.AddressState = types.Removed

	return s.DB.Save(&wa).Error
}
