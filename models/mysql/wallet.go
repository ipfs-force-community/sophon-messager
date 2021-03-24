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
	Id types.UUID `gorm:"column:id;primary_key;" json:"id"` // 主键

	Name  string `gorm:"column:name;type:varchar(256);NOT NULL"`
	Url   string `gorm:"column:url;type:varchar(256);NOT NULL"`
	Token string `gorm:"column:token;type:varchar(256);NOT NULL"`

	IsDeleted int       `gorm:"column:is_deleted;default:-1;NOT NULL"`                // 是否删除 1:是  -1:否
	CreatedAt time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP;NOT NULL"` // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP;NOT NULL"` // 更新时间
}

func FromWallet(msg types.Wallet) *mysqlWallet {
	return automapper.MustMapper(&msg, TMysqlWallet).(*mysqlWallet)
}

func (sqliteWallet mysqlWallet) Wallet() *types.Wallet {
	return automapper.MustMapper(&sqliteWallet, TWallet).(*types.Wallet)
}

func (sqliteWallet mysqlWallet) TableName() string {
	return "wallets"
}

var _ repo.WalletRepo = (*mysqlWalletRepo)(nil)

type mysqlWalletRepo struct {
	*gorm.DB
}

func (s mysqlWalletRepo) HasWallet(name string) (bool, error) {
	panic("implement me")
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
	if err := s.DB.Where(&mysqlWallet{Id: uuid, IsDeleted: -1}).First(&wallet).Error; err != nil {
		return nil, err
	}
	return wallet.Wallet(), nil
}

func (s mysqlWalletRepo) GetWalletByName(name string) (*types.Wallet, error) {
	panic("implement me")
}

func (s mysqlWalletRepo) ListWallet() ([]*types.Wallet, error) {
	var internalMsg []*mysqlWallet
	if err := s.DB.Find(&internalMsg, "is_deleted = ?", -1).Error; err != nil {
		return nil, err
	}

	result, err := automapper.Mapper(internalMsg, reflect.TypeOf([]*types.Message{}))
	if err != nil {
		return nil, err
	}
	return result.([]*types.Wallet), nil
}

func (s mysqlWalletRepo) DelWallet(uuid types.UUID) error {
	var wallet mysqlWallet
	if err := s.DB.Where(&mysqlWallet{Id: uuid, IsDeleted: -1}).First(&wallet).Error; err != nil {
		return err
	}
	wallet.IsDeleted = 1

	return s.DB.Save(&wallet).Error
}
