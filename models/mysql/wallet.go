package mysql

import (
	"reflect"
	"time"

	"github.com/hunjixin/automapper"
	"gorm.io/gorm"

	"github.com/filecoin-project/venus-messager/models/repo"
	"github.com/filecoin-project/venus-messager/types"
)

type mysqlWallet struct {
	ID types.UUID `gorm:"column:id;type:varchar(256);primary_key;"` // 主键

	Name  string      `gorm:"column:name;type:varchar(256);NOT NULL"`
	Url   string      `gorm:"column:url;type:varchar(256);NOT NULL"`
	Token string      `gorm:"column:token;type:varchar(256);NOT NULL"`
	State types.State `gorm:"column:state;type:int;"`

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

func (s mysqlWalletRepo) SaveWallet(wallet *types.Wallet) error {
	return s.DB.Save(FromWallet(*wallet)).Error
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

func (s mysqlWalletRepo) GetOneRecord(name string) (*types.Wallet, error) {
	var wallet mysqlWallet
	if err := s.DB.Where("name = ?", name).First(&wallet).Error; err != nil {
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

func (s mysqlWalletRepo) UpdateState(name string, state types.State) error {
	return s.DB.Model((*mysqlWallet)(nil)).Where("name = ? and is_deleted = -1", name).
		UpdateColumn("state", state).Error
}

func (s mysqlWalletRepo) DelWallet(name string) error {
	var wallet mysqlWallet
	if err := s.DB.Where("name = ? and is_deleted = -1", name).First(&wallet).Error; err != nil {
		return err
	}
	wallet.IsDeleted = repo.Deleted
	wallet.State = types.Removed

	return s.DB.Save(&wallet).Error
}
