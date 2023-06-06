package filestore

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ipfs-force-community/sophon-messager/config"
	"github.com/ipfs-force-community/sophon-messager/utils"
)

const (
	ConfigFile = "config.toml"
	TipsetFile = "tipset.json"
	SqliteFile = "message.db"
	TokenFile  = "token"
)

type FSRepo interface {
	Path() string
	Config() *config.Config
	ReplaceConfig(cfg *config.Config) error
	TipsetFile() string
	SqliteFile() string
	GetToken() ([]byte, error)
	SaveToken([]byte) error
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
	if cfg.MessageService.DefaultTimeout <= 0 {
		cfg.MessageService.DefaultTimeout = config.DefaultTimeout
	}
	if cfg.MessageService.SignMessageTimeout <= 0 {
		cfg.MessageService.SignMessageTimeout = config.SignMessageTimeout
	}
	if cfg.MessageService.EstimateMessageTimeout <= 0 {
		cfg.MessageService.EstimateMessageTimeout = config.EstimateMessageTimeout
	}
	r.cfg = cfg

	return r, nil
}

func InitFSRepo(repoPath string, cfg *config.Config) (FSRepo, error) {
	if err := os.MkdirAll(repoPath, 0o775); err != nil {
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

func (r *fsRepo) SaveToken(token []byte) error {
	err := os.WriteFile(filepath.Join(r.path, TokenFile), token, 0o644)
	if err != nil {
		return fmt.Errorf("write token to token file failed: %v", err)
	}
	return nil
}

func (r *fsRepo) GetToken() ([]byte, error) {
	return os.ReadFile(filepath.Join(r.path, TokenFile))
}
