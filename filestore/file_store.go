package filestore

import (
	"os"
	"path/filepath"

	"github.com/filecoin-project/venus-messager/config"
	"github.com/filecoin-project/venus-messager/utils"
)

const (
	ConfigFile = "config.toml"
	TipsetFile = "tipset.json"
	SqliteFile = "message.db"
)

type FSRepo interface {
	Path() string
	Config() *config.Config
	ReplaceConfig(cfg *config.Config) error
	TipsetFile() string
	SqliteFile() string
}

type fsRepo struct {
	path string
	cfg  *config.Config
}

func NewFSRepo(repoPath string) (FSRepo, error) {
	r := &fsRepo{path: repoPath}
	cfg := config.DefaultConfig()
	err := utils.ReadConfig(filepath.Join(repoPath, ConfigFile), cfg)
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

	if err := utils.WriteConfig(filepath.Join(repoPath, ConfigFile), cfg); err != nil {
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

func (r *fsRepo) TipsetFile() string {
	return filepath.Join(r.path, TipsetFile)
}

func (r *fsRepo) SqliteFile() string {
	return filepath.Join(r.path, SqliteFile)
}

func (r *fsRepo) ReplaceConfig(cfg *config.Config) error {
	if err := utils.WriteConfig(filepath.Join(r.path, ConfigFile), cfg); err != nil {
		return err
	}
	r.cfg = cfg

	return nil
}
