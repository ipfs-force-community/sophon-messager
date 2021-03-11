package repo

import (
	"os"
	"path/filepath"

	"github.com/ipfs-force-community/venus-messager/config"
)

const (
	ConfigFilename = "messager.toml"
)

func InitRepo(repoDir string) error {
	if !isExist(repoDir) {
		if err := os.MkdirAll(repoDir, 0775); err != nil {
			return err
		}
	}

	cfgPath := filepath.Join(repoDir, ConfigFilename)
	exist, err := fileExist(cfgPath)
	if err != nil {
		return err
	}

	defCfg := config.DefaultConfig()
	if !exist {
		cfg := defCfg
		cfg.MessageService.TipsetFilePath = filepath.Join(repoDir, cfg.MessageService.TipsetFilePath)
		cfg.DB.Sqlite.Path = filepath.Join(repoDir, cfg.DB.Sqlite.Path)

		if err := config.WriteFile(cfgPath, cfg); err != nil {
			return err
		}
	}

	tipsetFilePath := filepath.Join(repoDir, defCfg.MessageService.TipsetFilePath)
	exist, err = fileExist(tipsetFilePath)
	if err != nil {
		return err
	}
	if !exist {
		if _, err := os.Create(tipsetFilePath); err != nil {
			return err
		}
	}

	return nil
}

func fileExist(path string) (bool, error) {
	_, err := os.Stat(path)
	switch {
	case err == nil:
		return true, nil
	case os.IsNotExist(err):
		return false, nil
	default:
		return false, err
	}
}

func isExist(path string) bool {
	f, err := os.Stat(path)
	if err != nil {
		return false
	}
	return f.IsDir()
}

func Exists(p string) bool {
	_, err := os.Stat(filepath.Join(p, ConfigFilename))

	return !os.IsNotExist(err)
}
