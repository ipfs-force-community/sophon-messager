package mysql

import (
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/ipfs-force-community/venus-messager/types"
	"gorm.io/gorm"

	"github.com/ipfs-force-community/venus-messager/models/repo"
)

type mysqlWalletAddress struct {
	ID           types.UUID  `gorm:"column:id;type:varchar(256);primary_key"`
	WalletName   string      `gorm:"column:wallet_name;type:varchar(256)"`
	Addr         string      `gorm:"column:addr;type:varchar(256);NOT NULL"` // 主键
	AddressState types.State `gorm:"column:addr_state;type:int;index:wallet_addr_state;"`
	SelectMsgNum uint64      `gorm:"column:select_msg_num;type:int;NOT NULL"`

	IsDeleted int       `gorm:"column:is_deleted;index;default:-1;NOT NULL"` // 是否删除 1:是  -1:否
	CreatedAt time.Time `gorm:"column:created_at;index;NOT NULL"`            // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;index;NOT NULL"`            // 更新时间
}

func FromWalletAddress(wa types.WalletAddress) *mysqlWalletAddress {
	mysqlWalletAddress := &mysqlWalletAddress{
		ID:           wa.ID,
		WalletName:   wa.WalletName,
		Addr:         wa.Addr.String(),
		AddressState: wa.AddressState,
		SelectMsgNum: wa.SelMsgNum,
		IsDeleted:    wa.IsDeleted,
		CreatedAt:    wa.CreatedAt,
		UpdatedAt:    wa.UpdatedAt,
	}
	return mysqlWalletAddress
}

func (mysqlWalletAddress mysqlWalletAddress) WalletAddress() (*types.WalletAddress, error) {
	wa := &types.WalletAddress{
		ID:           mysqlWalletAddress.ID,
		WalletName:   mysqlWalletAddress.WalletName,
		AddressState: mysqlWalletAddress.AddressState,
		SelMsgNum:    mysqlWalletAddress.SelectMsgNum,
		IsDeleted:    mysqlWalletAddress.IsDeleted,
		CreatedAt:    mysqlWalletAddress.CreatedAt,
		UpdatedAt:    mysqlWalletAddress.UpdatedAt,
	}
	var err error
	wa.Addr, err = address.NewFromString(mysqlWalletAddress.Addr)
	if err != nil {
		return nil, err
	}
	return wa, nil
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

func (s mysqlWalletAddressRepo) GetWalletAddress(walletName string, addr address.Address) (*types.WalletAddress, error) {
	var wa mysqlWalletAddress
	if err := s.DB.Where("wallet_name = ? and addr = ? and is_deleted = -1", walletName, addr.String()).
		First(&wa).Error; err != nil {
		return nil, err
	}
	return wa.WalletAddress()
}

func (s mysqlWalletAddressRepo) GetOneRecord(walletName string, addr address.Address) (*types.WalletAddress, error) {
	var wa mysqlWalletAddress
	if err := s.DB.Where("wallet_name = ? and addr = ?", walletName, addr.String()).First(&wa).Error; err != nil {
		return nil, err
	}
	return wa.WalletAddress()
}

func (s mysqlWalletAddressRepo) HasWalletAddress(walletName string, addr address.Address) (bool, error) {
	var count int64
	if err := s.DB.Model(&mysqlWalletAddress{}).Where("wallet_name = ? and addr = ?", walletName, addr.String()).
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

	var err error
	result := make([]*types.WalletAddress, len(internalWalletAddress))
	for i, wa := range internalWalletAddress {
		if result[i], err = wa.WalletAddress(); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (s mysqlWalletAddressRepo) UpdateAddressState(walletName string, addr address.Address, state types.State) error {
	return s.DB.Model((*mysqlWalletAddress)(nil)).Where("wallet_name = ? and addr = ?", walletName, addr.String()).
		UpdateColumn("addr_state", state).Error
}

func (s mysqlWalletAddressRepo) UpdateSelectMsgNum(walletName string, addr address.Address, selMsgNum uint64) error {
	return s.DB.Model((*mysqlWalletAddress)(nil)).Where("wallet_name = ? and addr = ?", walletName, addr.String()).
		UpdateColumn("select_msg_num", selMsgNum).Error
}

func (s mysqlWalletAddressRepo) DelWalletAddress(walletName string, addr address.Address) error {
	var wa mysqlWalletAddress
	if err := s.DB.Where("wallet_name = ? and addr = ? and is_deleted = -1", walletName, addr.String()).
		First(&wa).Error; err != nil {
		return err
	}
	wa.IsDeleted = 1
	wa.AddressState = types.Removed

	return s.DB.Save(&wa).Error
}
