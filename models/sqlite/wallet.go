package sqlite

import (
	"reflect"
	"time"

	"github.com/hunjixin/automapper"
	"gorm.io/gorm"

	"github.com/ipfs-force-community/venus-messager/models/repo"
	"github.com/ipfs-force-community/venus-messager/types"
)

type sqliteWallet struct {
	ID types.UUID `gorm:"column:id;primary_key;"` // 主键

	Name  string `gorm:"column:name;uniqueIndex;type:varchar(256);NOT NULL"`
	Url   string `gorm:"column:url;type:varchar(256);NOT NULL"`
	Token string `gorm:"column:token;type:varchar(256);NOT NULL"`

	IsDeleted int       `gorm:"column:is_deleted;index;default:-1;NOT NULL"` // 是否删除 1:是  -1:否
	CreatedAt time.Time `gorm:"column:created_at;index;NOT NULL"`            // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;index;NOT NULL"`            // 更新时间
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
	*gorm.DB
}

func newSqliteWalletRepo(db *gorm.DB) sqliteWalletRepo {
	return sqliteWalletRepo{DB: db}
}

func (s sqliteWalletRepo) SaveWallet(wallet *types.Wallet) (string, error) {
	err := s.DB.Save(FromWallet(*wallet)).Error
	return wallet.ID.String(), err
}

func (s sqliteWalletRepo) GetWallet(uuid types.UUID) (*types.Wallet, error) {
	var wallet sqliteWallet
	if err := s.DB.Where(&sqliteWallet{ID: uuid, IsDeleted: -1}).First(&wallet).Error; err != nil {
		return nil, err
	}
	return wallet.Wallet(), nil
}

func (s sqliteWalletRepo) ListWallet() ([]*types.Wallet, error) {
	var internalWallet []*sqliteWallet
	if err := s.DB.Find(&internalWallet, "is_deleted = ?", -1).Error; err != nil {
		return nil, err
	}

	result, err := automapper.Mapper(internalWallet, reflect.TypeOf([]*types.Wallet{}))
	if err != nil {
		return nil, err
	}
	return result.([]*types.Wallet), nil
}

func (s sqliteWalletRepo) DelWallet(uuid types.UUID) error {
	var wallet sqliteWallet
	if err := s.DB.Where(&sqliteWallet{ID: uuid, IsDeleted: -1}).First(&wallet).Error; err != nil {
		return err
	}
	wallet.IsDeleted = 1

	return s.DB.Save(&wallet).Error
}
