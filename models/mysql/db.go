package mysql

import (
	"time"

	"golang.org/x/xerrors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/filecoin-project/venus-messager/config"
	"github.com/filecoin-project/venus-messager/models/repo"
)

type MysqlRepo struct {
	*gorm.DB
}

func (d MysqlRepo) WalletAddressRepo() repo.WalletAddressRepo {
	return newMysqlWalletAddressRepo(d.DB)
}

func (d MysqlRepo) MessageRepo() repo.MessageRepo {
	return newMysqlMessageRepo(d.DB)
}

func (d MysqlRepo) WalletRepo() repo.WalletRepo {
	return newMysqlWalletRepo(d.DB)
}

func (d MysqlRepo) AddressRepo() repo.AddressRepo {
	return newMysqlAddressRepo(d.DB)
}

func (d MysqlRepo) SharedParamsRepo() repo.SharedParamsRepo {
	return newMysqlSharedParamsRepo(d.DB)
}

func (d MysqlRepo) NodeRepo() repo.NodeRepo {
	return newMysqlNodeRepo(d.DB)
}

func (d MysqlRepo) FeeConfigRepo() repo.FeeConfigRepo {
	return newMysqlFeeConfigRepo(d.DB)
}

func (d MysqlRepo) AutoMigrate() error {
	err := d.GetDb().AutoMigrate(mysqlMessage{})
	if err != nil {
		return err
	}

	if err := d.GetDb().AutoMigrate(mysqlAddress{}); err != nil {
		return err
	}

	if err := d.GetDb().AutoMigrate(mysqlSharedParams{}); err != nil {
		return err
	}

	if err := d.GetDb().AutoMigrate(mysqlNode{}); err != nil {
		return err
	}

	if err := d.GetDb().AutoMigrate(mysqlWalletAddress{}); err != nil {
		return err
	}

	if err := d.GetDb().AutoMigrate(mysqlFeeConfig{}); err != nil {
		return err
	}

	return d.GetDb().AutoMigrate(mysqlWallet{})
}

func (d MysqlRepo) GetDb() *gorm.DB {
	return d.DB
}

func (d MysqlRepo) DbClose() error {
	// return d.DbClose()
	// todo:
	return nil
}

func (d MysqlRepo) Transaction(cb func(txRepo repo.TxRepo) error) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		txRepo := &TxMysqlRepo{tx}
		return cb(txRepo)
	})
}

var _ repo.TxRepo = (*TxMysqlRepo)(nil)

type TxMysqlRepo struct {
	*gorm.DB
}

func (t *TxMysqlRepo) WalletAddressRepo() repo.WalletAddressRepo {
	return newMysqlWalletAddressRepo(t.DB)
}

func (t *TxMysqlRepo) WalletRepo() repo.WalletRepo {
	return newMysqlWalletRepo(t.DB)
}

func (t *TxMysqlRepo) MessageRepo() repo.MessageRepo {
	return newMysqlMessageRepo(t.DB)
}

func (t *TxMysqlRepo) AddressRepo() repo.AddressRepo {
	return newMysqlAddressRepo(t.DB)
}

func OpenMysql(cfg *config.MySqlConfig) (repo.Repo, error) {
	db, err := gorm.Open(mysql.Open(cfg.ConnectionString), &gorm.Config{
		//Logger: logger.Default.LogMode(logger.Info), // 日志配置
	})

	if err != nil {
		return nil, xerrors.Errorf("[db connection failed] Database name: %s %w", cfg.ConnectionString, err)
	}

	db.Set("gorm:table_options", "CHARSET=utf8mb4")
	if cfg.Debug {
		db = db.Debug()
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConn)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConn)
	sqlDB.SetConnMaxLifetime(time.Minute * cfg.ConnMaxLifeTime)

	// 使用插件
	//db.Use(&TracePlugin{})
	return &MysqlRepo{
		db,
	}, nil
}
