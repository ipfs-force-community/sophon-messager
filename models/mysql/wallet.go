package mysql

import (
	"github.com/hunjixin/automapper"
	"github.com/ipfs-force-community/venus-messager/models/repo"
	"github.com/ipfs-force-community/venus-messager/types"
	"reflect"
	"time"
)

type mysqlWallet struct {
	Id string `gorm:"column:id;primary_key;"json:"id"` // 主键

	Name  string `gorm:"column:to;type:varchar(256);NOT NULL"`
	Url   string `gorm:"column:to;type:varchar(256);NOT NULL"`
	token string `gorm:"column:to;type:varchar(256);NOT NULL"`

	IsDeleted int       `gorm:"column:is_deleted;default:-1;NOT NULL"`                // 是否删除 1:是  -1:否
	CreatedAt time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP;NOT NULL"` // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP;NOT NULL"` // 更新时间
}

func FromWallet(msg types.Wallet) *mysqlWallet {
	return automapper.MustMapper(&msg, TMysqlWallet).(*mysqlWallet)
}

func (sqliteWallet mysqlWallet) Wallet() types.Wallet {
	return automapper.MustMapper(sqliteWallet, TWallet).(types.Wallet)
}

func (sqliteWallet mysqlWallet) TableName() string {
	return "wallets"
}

var _ repo.WalletRepo = (*mysqlWalletRepo)(nil)

type mysqlWalletRepo struct {
	repo.Repo
}

func newMysqlWalletRepo(repo repo.Repo) mysqlWalletRepo {
	return mysqlWalletRepo{repo}
}

func (s mysqlWalletRepo) SaveWallet(wallet *types.Wallet) (string, error) {
	err := s.GetDb().Save(FromWallet(*wallet)).Error
	return wallet.Id, err
}

func (s mysqlWalletRepo) GetWallet(uuid string) (types.Wallet, error) {
	var wallet mysqlWallet
	if err := s.GetDb().First(&wallet, "id = ?", uuid, "is_deleted = ?", -1).Error; err != nil {
		return types.Wallet{}, err
	}
	return wallet.Wallet(), nil
}

func (s mysqlWalletRepo) ListWallet() ([]types.Wallet, error) {
	var internalMsg []mysqlWallet
	if err := s.GetDb().Find(&internalMsg, "is_deleted = ?", -1).Error; err != nil {
		return nil, err
	}

	result, err := automapper.Mapper(internalMsg, reflect.TypeOf([]types.Message{}))
	if err != nil {
		return nil, err
	}
	return result.([]types.Wallet), nil
}

func (s mysqlWalletRepo) DelWallet(uuid string) error {
	var wallet mysqlWallet
	if err := s.GetDb().First(&wallet, uuid, "is_deleted = ?", -1).Error; err != nil {
		return err
	}
	wallet.IsDeleted = 1

	return s.GetDb().Save(&wallet).Error
}
