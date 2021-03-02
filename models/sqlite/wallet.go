package sqlite

import (
	"reflect"
	"time"

	"github.com/hunjixin/automapper"
	"github.com/ipfs-force-community/venus-messager/models/repo"
	"github.com/ipfs-force-community/venus-messager/types"
)

type sqliteWallet struct {
	Id string `gorm:"column:id;primary_key;"json:"id"` // 主键

	Name  string `gorm:"column:name;type:varchar(256);NOT NULL"`
	Url   string `gorm:"column:url;type:varchar(256);NOT NULL"`
	Token string `gorm:"column:token;type:varchar(256);NOT NULL"`

	IsDeleted int       `gorm:"column:is_deleted;default:-1;NOT NULL"`                // 是否删除 1:是  -1:否
	CreatedAt time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP;NOT NULL"` // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP;NOT NULL"` // 更新时间
}

func FromWallet(msg types.Wallet) *sqliteWallet {
	return automapper.MustMapper(&msg, TSqliteWallet).(*sqliteWallet)
}

func (sqliteWallet sqliteWallet) Wallet() *types.Wallet {
	return automapper.MustMapper(&sqliteWallet, TWallet).(*types.Wallet)
}

func (sqliteWallet sqliteWallet) TableName() string {
	return "wallets"
}

var _ repo.WalletRepo = (*sqliteWalletRepo)(nil)

type sqliteWalletRepo struct {
	repo.Repo
}

func newSqliteWalletRepo(repo repo.Repo) sqliteWalletRepo {
	return sqliteWalletRepo{repo}
}

func (s sqliteWalletRepo) SaveWallet(wallet *types.Wallet) (string, error) {
	err := s.GetDb().Save(FromWallet(*wallet)).Error
	return wallet.Id, err
}

func (s sqliteWalletRepo) GetWallet(uuid string) (*types.Wallet, error) {
	var wallet sqliteWallet
	if err := s.GetDb().Where(&sqliteWallet{Id: uuid, IsDeleted: -1}).First(&wallet).Error; err != nil {
		return nil, err
	}
	return wallet.Wallet(), nil
}

func (s sqliteWalletRepo) ListWallet() ([]*types.Wallet, error) {
	var internalMsg []*sqliteWallet
	if err := s.GetDb().Find(&internalMsg, "is_deleted = ?", -1).Error; err != nil {
		return nil, err
	}

	result, err := automapper.Mapper(internalMsg, reflect.TypeOf([]*types.Wallet{}))
	if err != nil {
		return nil, err
	}
	return result.([]*types.Wallet), nil
}

func (s sqliteWalletRepo) DelWallet(uuid string) error {
	var wallet sqliteWallet
	if err := s.GetDb().Where(&sqliteWallet{Id: uuid, IsDeleted: -1}).First(&wallet).Error; err != nil {
		return err
	}
	wallet.IsDeleted = 1

	return s.GetDb().Save(&wallet).Error
}
