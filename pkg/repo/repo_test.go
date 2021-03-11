package repo

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ipfs-force-community/venus-messager/config"
)

func TestInitRepo(t *testing.T) {
	emptyPath := ""
	defaultRepoDir, err := GetRepoPath(emptyPath)
	assert.NoError(t, err)
	assert.Equal(t, defaultRepoDir, defaultRepoDir)

	repoDir := "./venus-messager-test"
	defer func() {
		assert.NoError(t, os.RemoveAll(repoDir))
	}()
	err = InitRepo(repoDir)
	assert.NoError(t, err)

	defaultCfg := config.DefaultConfig()

	cfgPath := filepath.Join(repoDir, ConfigFilename)
	exist, err := fileExist(cfgPath)
	assert.NoError(t, err)
	assert.True(t, exist)

	cfg, err := config.ReadConfig(cfgPath)
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(repoDir, defaultCfg.MessageService.TipsetFilePath), cfg.MessageService.TipsetFilePath)
	assert.Equal(t, filepath.Join(repoDir, defaultCfg.DB.Sqlite.Path), cfg.DB.Sqlite.Path)

}
