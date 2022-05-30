package filestore

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/filecoin-project/venus-messager/config"
)

const (
	ConfigFile = "config.toml"
)

type FSRepo interface {
	Path() string
	Config() *config.Config
	ReplaceConfig(cfg *config.Config) error
}

type fsRepo struct {
	path string
	cfg  *config.Config
}

func NewFSRepo(repoPath string) (FSRepo, error) {
	r := &fsRepo{path: repoPath}
	cfg, err := config.ReadConfig(filepath.Join(repoPath, ConfigFile))
	if err != nil {
		return nil, err
	}
	r.cfg = cfg

	return r, nil
}

func InitFSRepo(repoPath string, cfg *config.Config) (FSRepo, error) {
	if err := os.MkdirAll(repoPath, 0775); err != nil {
		return nil, err
	}

	if cfg.DB.Type == "sqlite" {
		filePath, err := filepath.Abs(cfg.DB.Sqlite.File)
		if err != nil {
			return nil, err
		}
		fileName := filepath.Base(cfg.DB.Sqlite.File)
		fileRootPath := filepath.Dir(filePath)
		_, err = os.Stat(filePath)
		if err == nil {
			if err := copyFile(filePath, filepath.Join(repoPath, fileName)); err != nil {
				return nil, err
			}
			if err := copyFile(filepath.Join(fileRootPath, fileName+"-shm"), filepath.Join(repoPath, fileName+"-shm")); err != nil {
				return nil, err
			}
			if err := copyFile(filepath.Join(fileRootPath, fileName+"-wal"), filepath.Join(repoPath, fileName+"-wal")); err != nil {
				return nil, err
			}
		}
		cfg.DB.Sqlite.File = filepath.Join(repoPath, fileName)
	} else {
		cfg.DB.Sqlite.File = filepath.Join(repoPath, filepath.Base(cfg.DB.Sqlite.File))
	}

	tsFile := filepath.Base(cfg.MessageService.TipsetFilePath)
	_, err := os.Stat(cfg.MessageService.TipsetFilePath)
	if err == nil {
		if err := copyFile(cfg.MessageService.TipsetFilePath, filepath.Join(repoPath, tsFile)); err != nil {
			return nil, err
		}
	}
	cfg.MessageService.TipsetFilePath = filepath.Join(repoPath, tsFile)

	if err := config.WriteConfig(filepath.Join(repoPath, ConfigFile), cfg); err != nil {
		return nil, err
	}

	return &fsRepo{path: repoPath, cfg: cfg}, nil
}

func (r *fsRepo) Path() string {
	return r.path
}

func (r *fsRepo) Config() *config.Config {
	return r.cfg
}

func (r *fsRepo) ReplaceConfig(cfg *config.Config) error {
	if err := config.WriteConfig(filepath.Join(r.path, ConfigFile), cfg); err != nil {
		return err
	}
	r.cfg = cfg

	return nil
}

func copyFile(src, dst string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(dst, data, 0644)
}
