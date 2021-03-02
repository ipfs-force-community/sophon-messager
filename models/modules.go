package models

import (
	"golang.org/x/xerrors"

	"github.com/ipfs-force-community/venus-messager/config"
	"github.com/ipfs-force-community/venus-messager/models/mysql"
	"github.com/ipfs-force-community/venus-messager/models/repo"
	"github.com/ipfs-force-community/venus-messager/models/sqlite"
)

func SetDataBase(cfg *config.DbConfig) (repo.Repo, error) {
	switch cfg.Type {
	case "sqlite":
		return sqlite.OpenSqlite(&cfg.Sqlite)
	case "mysql":
		return mysql.OpenMysql(&cfg.MySql)
	default:
		return nil, xerrors.Errorf("unsupport db type,(%s, %s)", "sqlite", "mysql")
	}
}

func AutoMigrate(repo repo.Repo) error {
	return repo.AutoMigrate()
}
