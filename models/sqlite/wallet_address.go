package sqlite

import (
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/ipfs-force-community/venus-messager/types"
	"gorm.io/gorm"

	"github.com/ipfs-force-community/venus-messager/models/repo"
)

type sqliteWalletAddress struct {
	ID           types.UUID  `gorm:"column:id;type:varchar(256);primary_key"`
	WalletName   string      `gorm:"column:wallet_name;type:varchar(256)"`
	Addr         string      `gorm:"column:addr;type:varchar(256);NOT NULL"` // 主键
	AddressState types.State `gorm:"column:addr_state;type:int;index:wallet_addr_state;"`
	SelectMsgNum uint64      `gorm:"column:select_msg_num;type:int;NOT NULL"`

	IsDeleted int       `gorm:"column:is_deleted;index;default:-1;NOT NULL"` // 是否删除 1:是  -1:否
	CreatedAt time.Time `gorm:"column:created_at;index;NOT NULL"`            // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;index;NOT NULL"`            // 更新时间
}

func FromWalletAddress(wa types.WalletAddress) *sqliteWalletAddress {
	sqliteWalletAddress := &sqliteWalletAddress{
		ID:           wa.ID,
		WalletName:   wa.WalletName,
		Addr:         wa.Addr.String(),
		AddressState: wa.AddressState,
		SelectMsgNum: wa.SelMsgNum,
		IsDeleted:    wa.IsDeleted,
		CreatedAt:    wa.CreatedAt,
		UpdatedAt:    wa.UpdatedAt,
	}
	return sqliteWalletAddress
}

func (sqliteWalletAddress sqliteWalletAddress) WalletAddress() (*types.WalletAddress, error) {
	wa := &types.WalletAddress{
		ID:           sqliteWalletAddress.ID,
		WalletName:   sqliteWalletAddress.WalletName,
		AddressState: sqliteWalletAddress.AddressState,
		SelMsgNum:    sqliteWalletAddress.SelectMsgNum,
		IsDeleted:    sqliteWalletAddress.IsDeleted,
		CreatedAt:    sqliteWalletAddress.CreatedAt,
		UpdatedAt:    sqliteWalletAddress.UpdatedAt,
	}
	var err error
	wa.Addr, err = address.NewFromString(sqliteWalletAddress.Addr)
	if err != nil {
		return nil, err
	}
	return wa, nil
}

func (sqliteWalletAddress sqliteWalletAddress) TableName() string {
	return "wallet_addresses"
}

var _ repo.WalletAddressRepo = (*sqliteWalletAddressRepo)(nil)

type sqliteWalletAddressRepo struct {
	*gorm.DB
}

func newSqliteWalletAddressRepo(db *gorm.DB) sqliteWalletAddressRepo {
	return sqliteWalletAddressRepo{DB: db}
}

func (s sqliteWalletAddressRepo) SaveWalletAddress(wa *types.WalletAddress) error {
	sqliteWalletAddress := FromWalletAddress(*wa)
	sqliteWalletAddress.UpdatedAt = time.Now()
	return s.DB.Save(sqliteWalletAddress).Error
}

func (s sqliteWalletAddressRepo) GetWalletAddress(walletName string, addr address.Address) (*types.WalletAddress, error) {
	var wa sqliteWalletAddress
	if err := s.DB.Where("wallet_name = ? and addr = ? and is_deleted = -1", walletName, addr.String()).
		First(&wa).Error; err != nil {
		return nil, err
	}
	return wa.WalletAddress()
}

func (s sqliteWalletAddressRepo) GetOneRecord(walletName string, addr address.Address) (*types.WalletAddress, error) {
	var wa sqliteWalletAddress
	if err := s.DB.Where("wallet_name = ? and addr = ?", walletName, addr.String()).First(&wa).Error; err != nil {
		return nil, err
	}
	return wa.WalletAddress()
}

func (s sqliteWalletAddressRepo) HasWalletAddress(walletName string, addr address.Address) (bool, error) {
	var count int64
	if err := s.DB.Model(&sqliteWalletAddress{}).Where("wallet_name = ? and addr = ?", walletName, addr.String()).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s sqliteWalletAddressRepo) ListWalletAddress() ([]*types.WalletAddress, error) {
	var internalWalletAddress []*sqliteWalletAddress
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

func (s sqliteWalletAddressRepo) UpdateAddressState(walletName string, addr address.Address, state types.State) error {
	return s.DB.Model((*sqliteWalletAddress)(nil)).Where("wallet_name = ? and addr = ?", walletName, addr.String()).
		UpdateColumn("addr_state", state).Error
}

func (s sqliteWalletAddressRepo) UpdateSelectMsgNum(walletName string, addr address.Address, selMsgNum uint64) error {
	return s.DB.Model((*sqliteWalletAddress)(nil)).Where("wallet_name = ? and addr = ?", walletName, addr.String()).
		UpdateColumn("select_msg_num", selMsgNum).Error
}

func (s sqliteWalletAddressRepo) DelWalletAddress(walletName string, addr address.Address) error {
	var wa sqliteWalletAddress
	if err := s.DB.Where("wallet_name = ? and addr = ? and is_deleted = -1", walletName, addr.String()).
		First(&wa).Error; err != nil {
		return err
	}
	wa.IsDeleted = 1
	wa.AddressState = types.Removed

	return s.DB.Save(&wa).Error
}
