package filestore

import (
	"path/filepath"
	"testing"

	"github.com/filecoin-project/venus-messager/config"
	"github.com/filecoin-project/venus-messager/utils"
	"github.com/stretchr/testify/assert"
)

func TestNewFSRepo(t *testing.T) {
	path := t.TempDir()
	defCfg := config.DefaultConfig()
	assert.Nil(t, utils.WriteConfig(filepath.Join(path, ConfigFile), defCfg))

	fsRepo, err := NewFSRepo(path)
	assert.NoError(t, err)

	assert.Equal(t, config.DefaultConfig(), fsRepo.Config())
	assert.Equal(t, path, fsRepo.Path())
	assert.Equal(t, filepath.Join(path, TipsetFile), fsRepo.TipsetFile())
	assert.Equal(t, filepath.Join(path, SqliteFile), fsRepo.SqliteFile())

	token := []byte("test-token")
	err = fsRepo.SaveToken(token)
	assert.NoError(t, err)

	token2, err := fsRepo.GetToken()
	assert.NoError(t, err)
	assert.Equal(t, token, token2)

	t.Run("use default value when timeout is zero", func(t *testing.T) {
		cfgCopy := *config.DefaultConfig()
		cfgCopy.MessageService.DefaultTimeout = 0
		cfgCopy.MessageService.SignMessageTimeout = 0
		cfgCopy.MessageService.EstimateMessageTimeout = 0

		repoPath := t.TempDir()
		assert.Nil(t, utils.WriteConfig(filepath.Join(repoPath, ConfigFile), &cfgCopy))
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

	fsRepo, err := InitFSRepo(path, defCfg)
	assert.NoError(t, err)

	assert.Equal(t, defCfg, fsRepo.Config())
}
