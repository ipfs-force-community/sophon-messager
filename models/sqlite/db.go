package sqlite

import (
	"github.com/ipfs-force-community/venus-messager/config"
	"github.com/ipfs-force-community/venus-messager/models/repo"
	"golang.org/x/xerrors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type SqlLiteRepo struct {
	*gorm.DB
}

func (d SqlLiteRepo) MessageRepo() repo.MessageRepo {
	return newSqliteMessageRepo(d)
}

func (d SqlLiteRepo) WalletRepo() repo.WalletRepo {
	return newSqliteWalletRepo(d)
}

func (d SqlLiteRepo) AutoMigrate() error {
	err := d.GetDb().AutoMigrate(sqliteMessage{})
	if err != nil {
		return err
	}

	return d.GetDb().AutoMigrate(sqliteWallet{})
}

func (d SqlLiteRepo) GetDb() *gorm.DB {
	return d.DB
}

func (d SqlLiteRepo) DbClose() error {
	return d.DbClose()
}

func OpenSqlite(cfg *config.SqliteConfig) (repo.Repo, error) {
	db, err := gorm.Open(sqlite.Open(cfg.Path), &gorm.Config{
		//Logger: logger.Default.LogMode(logger.Info), // 日志配置
	})
	if err != nil {
		return nil, xerrors.Errorf("fail to connect sqlite: %s %w", cfg.Path, err)
	}
	db.Set("gorm:table_options", "CHARSET=utf8mb4")

	return &SqlLiteRepo{
		db,
	}, nil
}
