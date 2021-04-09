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
	ID types.UUID `gorm:"column:id;type:varchar(256);primary_key;"` // 主键

	Name  string      `gorm:"column:name;type:varchar(256);NOT NULL"`
	Url   string      `gorm:"column:url;type:varchar(256);NOT NULL"`
	Token string      `gorm:"column:token;type:varchar(256);NOT NULL"`
	State types.State `gorm:"column:state;type:int;"`

	IsDeleted int       `gorm:"column:is_deleted;index;default:-1;NOT NULL"` // 是否删除 1:是  -1:否
	CreatedAt time.Time `gorm:"column:created_at;index;NOT NULL"`            // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;index;NOT NULL"`            // 更新时间
}

func FromWallet(wallet types.Wallet) *sqliteWallet {
	return automapper.MustMapper(&wallet, TSqliteWallet).(*sqliteWallet)
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

func (s sqliteWalletRepo) SaveWallet(wallet *types.Wallet) error {
	return s.DB.Save(FromWallet(*wallet)).Error
}

func (s sqliteWalletRepo) GetWalletByID(uuid types.UUID) (*types.Wallet, error) {
	var wallet sqliteWallet
	if err := s.DB.Where("id = ? and is_deleted = -1", uuid).First(&wallet).Error; err != nil {
		return nil, err
	}
	return wallet.Wallet(), nil
}

func (s sqliteWalletRepo) GetWalletByName(name string) (*types.Wallet, error) {
	var wallet sqliteWallet
	if err := s.DB.Where("name = ? and is_deleted = -1", name).First(&wallet).Error; err != nil {
		return nil, err
	}
	return wallet.Wallet(), nil
}

func (s sqliteWalletRepo) GetOneRecord(name string) (*types.Wallet, error) {
	var wallet sqliteWallet
	if err := s.DB.Where("name = ?", name).First(&wallet).Error; err != nil {
		return nil, err
	}
	return wallet.Wallet(), nil
}

func (s sqliteWalletRepo) HasWallet(name string) (bool, error) {
	var count int64
	if err := s.DB.Model(&sqliteWallet{}).Where("name = ?", name).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
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

func (s sqliteWalletRepo) UpdateState(name string, state types.State) error {
	return s.DB.Model((*sqliteWallet)(nil)).Where("name = ?", name).UpdateColumn("state", state).Error
}

func (s sqliteWalletRepo) DelWallet(name string) error {
	var wallet sqliteWallet
	if err := s.DB.Where("name = ? and is_deleted = -1", name).First(&wallet).Error; err != nil {
		return err
	}
	wallet.IsDeleted = 1
	wallet.State = types.Removed

	return s.DB.Save(&wallet).Error
}
