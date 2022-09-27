package filestore

import (
	"crypto/rand"
	"io"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/filecoin-project/venus-messager/config"
	"github.com/stretchr/testify/assert"
)

func TestNewFSRepo(t *testing.T) {
	path := t.TempDir()
	defCfg := config.DefaultConfig()
	assert.Nil(t, config.WriteConfig(filepath.Join(path, ConfigFile), defCfg))

	fsRepo, err := NewFSRepo(path)
	assert.Nil(t, err)

	assert.Equal(t, config.DefaultConfig(), fsRepo.Config())
	assert.Equal(t, path, fsRepo.Path())

	cfg := config.DefaultConfig()
	cfg.MessageService.TipsetFilePath = ""
	cfg.DB.Type = "mysql"
	cfg.JWT.Local.Secret = "secret"
	cfg.JWT.Local.Token = "token"
	assert.Nil(t, fsRepo.ReplaceConfig(cfg))
	assert.Equal(t, cfg, fsRepo.Config())

	t.Run("use default value when timeout is zero", func(t *testing.T) {
		cfgCopy := *config.DefaultConfig()
		cfgCopy.MessageService.DefaultTimeout = 0
		cfgCopy.MessageService.SignMessageTimeout = 0
		cfgCopy.MessageService.EstimateMessageTimeout = 0

		repoPath := t.TempDir()
		assert.Nil(t, config.WriteConfig(filepath.Join(repoPath, ConfigFile), &cfgCopy))
		fsRepo, err := NewFSRepo(repoPath)
		assert.Nil(t, err)
		fsRepo.Config().MessageService.DefaultTimeout = config.DefaultTimeout
		fsRepo.Config().MessageService.SignMessageTimeout = config.SignMessageTimeout
		fsRepo.Config().MessageService.EstimateMessageTimeout = config.EstimateMessageTimeout
	})
}

func TestInitFSRepo(t *testing.T) {
	defCfg := config.DefaultConfig()
	path := t.TempDir()
	defCfg.DB.Sqlite.File = filepath.Join(path, SqliteFile)
	assert.Nil(t, randFile(defCfg.DB.Sqlite.File))
	assert.Nil(t, randFile(filepath.Join(path, "message.db-shm")))
	assert.Nil(t, randFile(filepath.Join(path, "message.db-wal")))
	defCfg.MessageService.TipsetFilePath = filepath.Join(path, TipsetFile)
	assert.Nil(t, randFile(defCfg.MessageService.TipsetFilePath))

	fsPath := t.TempDir()
	fsRepo, err := InitFSRepo(fsPath, defCfg)
	assert.Nil(t, err)
	cfg := fsRepo.Config()
	assert.Equal(t, "", cfg.DB.Sqlite.File)
	compareFile(t, filepath.Join(path, SqliteFile), fsRepo.SqliteFile())
	compareFile(t, filepath.Join(path, "message.db-shm"), filepath.Join(fsPath, "message.db-shm"))
	compareFile(t, filepath.Join(path, "message.db-wal"), filepath.Join(fsPath, "message.db-wal"))
	assert.Equal(t, "", cfg.MessageService.TipsetFilePath)
	compareFile(t, filepath.Join(path, TipsetFile), fsRepo.TipsetFile())

	defCfg2 := config.DefaultConfig()
	path2 := t.TempDir()
	defCfg.DB.Sqlite.File = filepath.Join(path2, "message.db")
	defCfg.MessageService.TipsetFilePath = filepath.Join(path2, "tipset.json")

	fsPath2 := t.TempDir()
	fsRepo2, err := InitFSRepo(fsPath2, defCfg2)
	assert.Nil(t, err)
	cfg2 := fsRepo2.Config()
	assert.Equal(t, "", cfg2.DB.Sqlite.File)
	assert.Equal(t, "", cfg2.MessageService.TipsetFilePath)
}

func randFile(path string) error {
	data, err := ioutil.ReadAll(io.LimitReader(rand.Reader, 32))
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, 0644)
}

func compareFile(t *testing.T, f1, f2 string) {
	data, err := ioutil.ReadFile(f1)
	assert.Nil(t, err)
	data2, err := ioutil.ReadFile(f2)
	assert.Nil(t, err)
	assert.Equal(t, data, data2)
}
