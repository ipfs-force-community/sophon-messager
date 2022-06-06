package models

import (
	"golang.org/x/xerrors"

	"github.com/filecoin-project/venus-messager/filestore"
	"github.com/filecoin-project/venus-messager/models/mysql"
	"github.com/filecoin-project/venus-messager/models/repo"
	"github.com/filecoin-project/venus-messager/models/sqlite"
)

func SetDataBase(fsRepo filestore.FSRepo) (repo.Repo, error) {
	switch fsRepo.Config().DB.Type {
	case "sqlite":
		return sqlite.OpenSqlite(fsRepo)
	case "mysql":
		return mysql.OpenMysql(&fsRepo.Config().DB.MySql)
	default:
		return nil, xerrors.Errorf("unexpected db type %s (want 'sqlite' or 'mysql')", fsRepo.Config().DB.Type)
	}
}

func AutoMigrate(repo repo.Repo) error {
	return repo.AutoMigrate()
}
