package sqlite

import (
	"golang.org/x/xerrors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/ipfs-force-community/venus-messager/config"
	"github.com/ipfs-force-community/venus-messager/models/repo"
)

type SqlLiteRepo struct {
	*gorm.DB
}

func (d SqlLiteRepo) MessageRepo() repo.MessageRepo {
	return newSqliteMessageRepo(d.DB)
}

func (d SqlLiteRepo) WalletRepo() repo.WalletRepo {
	return newSqliteWalletRepo(d.DB)
}

func (d SqlLiteRepo) AddressRepo() repo.AddressRepo {
	return newSqliteAddressRepo(d.DB)
}

func (d SqlLiteRepo) AutoMigrate() error {
	err := d.GetDb().AutoMigrate(sqliteMessage{})
	if err != nil {
		return err
	}

	if err := d.GetDb().AutoMigrate(sqliteAddress{}); err != nil {
		return err
	}

	return d.GetDb().AutoMigrate(sqliteWallet{})
}

func (d SqlLiteRepo) GetDb() *gorm.DB {
	return d.DB
}

func (d SqlLiteRepo) Transaction(cb func(txRepo repo.TxRepo) error) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		txRepo := &TxSqlliteRepo{tx}
		return cb(txRepo)
	})
}

var _ repo.TxRepo = (*TxSqlliteRepo)(nil)

type TxSqlliteRepo struct {
	*gorm.DB
}

func (t *TxSqlliteRepo) WalletRepo() repo.WalletRepo {
	return newSqliteWalletRepo(t.DB)
}

func (t *TxSqlliteRepo) MessageRepo() repo.MessageRepo {
	return newSqliteMessageRepo(t.DB)
}

func (t *TxSqlliteRepo) AddressRepo() repo.AddressRepo {
	return newSqliteAddressRepo(t.DB)
}

func (d SqlLiteRepo) DbClose() error {
	// todo: if '*gorm.DB' need to dispose?
	return nil
}

func OpenSqlite(cfg *config.SqliteConfig) (repo.Repo, error) {
	//cache=shared&_journal_mode=wal&sync=normal
	//cache=shared&sync=full
	db, err := gorm.Open(sqlite.Open(cfg.Path+"?cache=shared&_journal_mode=wal&sync=normal"), &gorm.Config{
		// Logger: logger.Default.LogMode(logger.Info), // 日志配置
	})
	if err != nil {
		return nil, xerrors.Errorf("fail to connect sqlite: %s %w", cfg.Path, err)
	}
	db.Set("gorm:table_options", "CHARSET=utf8mb4")

	if cfg.Debug {
		db = db.Debug()
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 设置连接池 用于设置最大打开的连接数，默认值为0表示不限制.设置最大的连接数，可以避免并发太高导致连接mysql出现too many connections的错误。
	sqlDB.SetMaxOpenConns(1)

	// 设置最大连接数 用于设置闲置的连接数.设置闲置的连接数则当开启的一个连接使用完成后可以放在池里等候下一次使用。
	sqlDB.SetMaxIdleConns(1)

	return &SqlLiteRepo{
		db,
	}, nil
}
