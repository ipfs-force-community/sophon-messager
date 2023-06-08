package models

import (
	"fmt"

	"github.com/ipfs-force-community/sophon-messager/filestore"
	"github.com/ipfs-force-community/sophon-messager/models/mysql"
	"github.com/ipfs-force-community/sophon-messager/models/repo"
	"github.com/ipfs-force-community/sophon-messager/models/sqlite"
	"go.uber.org/fx"
)

func SetDataBase(fsRepo filestore.FSRepo) (repo.Repo, error) {
	switch fsRepo.Config().DB.Type {
	case "sqlite":
		return sqlite.OpenSqlite(fsRepo)
	case "mysql":
		return mysql.OpenMysql(&fsRepo.Config().DB.MySql)
	default:
		return nil, fmt.Errorf("unexpected db type %s (want 'sqlite' or 'mysql')", fsRepo.Config().DB.Type)
	}
}

func AutoMigrate(repo repo.Repo) error {
	return repo.AutoMigrate()
}

func Options() fx.Option {
	return fx.Options(
		fx.Provide(SetDataBase),
		fx.Invoke(AutoMigrate),
		// repo
		fx.Provide(repo.NewINodeRepo),
		fx.Provide(repo.NewINodeProvider),
	)
}
