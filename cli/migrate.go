package cli

import (
	"github.com/urfave/cli/v2"

	"github.com/filecoin-project/venus-messager/models"
)

var MirtateCmd = &cli.Command{
	Name:  "migrate",
	Usage: "auto migrate database",
	Action: func(cctx *cli.Context) error {
		fsRepo, err := getRepo(cctx)
		if err != nil {
			return err
		}
		cfg := fsRepo.Config()
		if cfg.DB.Type == "mysql" {
			cfg.DB.MySql.Debug = true
		} else {
			cfg.DB.Sqlite.Debug = true
		}
		if err := fsRepo.ReplaceConfig(cfg); err != nil {
			return err
		}

		r, err := models.SetDataBase(fsRepo)
		if err != nil {
			return err
		}

		return r.AutoMigrate()
	},
}
