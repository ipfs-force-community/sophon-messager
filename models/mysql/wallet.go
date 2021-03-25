package mysql

import (
	"reflect"
	"time"

	"gorm.io/gorm"

	"github.com/hunjixin/automapper"

	"github.com/ipfs-force-community/venus-messager/models/repo"
	"github.com/ipfs-force-community/venus-messager/types"
)

type mysqlWallet struct {
	ID types.UUID `gorm:"column:id;type:varchar(256);primary_key;"` // 主键

	Name  string `gorm:"column:name;uniqueIndex;type:varchar(256);NOT NULL"`
	Url   string `gorm:"column:url;type:varchar(256);NOT NULL"`
	Token string `gorm:"column:token;type:varchar(256);NOT NULL"`

	IsDeleted int       `gorm:"column:is_deleted;index;default:-1;NOT NULL"` // 是否删除 1:是  -1:否
	CreatedAt time.Time `gorm:"column:created_at;index;NOT NULL"`            // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;index;NOT NULL"`            // 更新时间
}

func FromWallet(msg types.Wallet) *mysqlWallet {
	return automapper.MustMapper(&msg, TMysqlWallet).(*mysqlWallet)
}

func (mysqlWallet mysqlWallet) Wallet() *types.Wallet {
	return automapper.MustMapper(&mysqlWallet, TWallet).(*types.Wallet)
}

func (mysqlWallet mysqlWallet) TableName() string {
	return "wallets"
}

var _ repo.WalletRepo = (*mysqlWalletRepo)(nil)

type mysqlWalletRepo struct {
	*gorm.DB
}

func newMysqlWalletRepo(db *gorm.DB) mysqlWalletRepo {
	return mysqlWalletRepo{DB: db}
}

func (s mysqlWalletRepo) SaveWallet(wallet *types.Wallet) (types.UUID, error) {
	err := s.DB.Save(FromWallet(*wallet)).Error
	return wallet.ID, err
}

func (s mysqlWalletRepo) GetWalletByID(uuid types.UUID) (*types.Wallet, error) {
	var wallet mysqlWallet
	if err := s.DB.Where("id = ? and is_deleted = -1", uuid.String()).First(&wallet).Error; err != nil {
		return nil, err
	}
	return wallet.Wallet(), nil
}

func (s mysqlWalletRepo) GetWalletByName(name string) (*types.Wallet, error) {
	var wallet mysqlWallet
	if err := s.DB.Where("name = ? and is_deleted = -1", name).First(&wallet).Error; err != nil {
		return nil, err
	}
	return wallet.Wallet(), nil
}

func (s mysqlWalletRepo) HasWallet(name string) (bool, error) {
	var count int64
	if err := s.DB.Model(&mysqlWallet{}).Where("name = ?", name).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s mysqlWalletRepo) ListWallet() ([]*types.Wallet, error) {
	var internalWallet []*mysqlWallet
	if err := s.DB.Find(&internalWallet, "is_deleted = ?", -1).Error; err != nil {
		return nil, err
	}

	result, err := automapper.Mapper(internalWallet, reflect.TypeOf([]*types.Wallet{}))
	if err != nil {
		return nil, err
	}
	return result.([]*types.Wallet), nil
}

func (s mysqlWalletRepo) DelWallet(uuid types.UUID) error {
	var wallet mysqlWallet
	if err := s.DB.Where("id = ? and is_deleted = -1", uuid).First(&wallet).Error; err != nil {
		return err
	}
	wallet.IsDeleted = 1

	return s.DB.Save(&wallet).Error
}
