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

}

func TestInitFSRepo(t *testing.T) {
	defCfg := config.DefaultConfig()
	path := t.TempDir()

	fsRepo, err := InitFSRepo(path, defCfg)
	assert.NoError(t, err)

	assert.Equal(t, defCfg, fsRepo.Config())
}
