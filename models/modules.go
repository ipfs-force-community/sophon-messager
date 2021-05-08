package models

import (
	"golang.org/x/xerrors"

	"github.com/filecoin-project/venus-messager/config"
	"github.com/filecoin-project/venus-messager/models/mysql"
	"github.com/filecoin-project/venus-messager/models/repo"
	"github.com/filecoin-project/venus-messager/models/sqlite"
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
