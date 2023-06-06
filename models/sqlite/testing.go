package sqlite

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ipfs-force-community/sophon-messager/filestore"
	"github.com/ipfs-force-community/sophon-messager/models/repo"
)

func setupRepo(t *testing.T) repo.Repo {
	fs := filestore.NewMockFileStore(t.TempDir())
	// cfg := fs.Config()
	// cfg.DB.Sqlite.Debug = true
	// assert.NoError(t, fs.ReplaceConfig(cfg))
	sqliteRepo, err := OpenSqlite(fs)
	assert.NoError(t, err)
	assert.NoError(t, sqliteRepo.AutoMigrate())

	return sqliteRepo
}
