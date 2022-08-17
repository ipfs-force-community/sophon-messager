package filestore

import (
	"path/filepath"

	"github.com/filecoin-project/venus-messager/config"
)

type mockFileStore struct {
	path string
	cfg  *config.Config
}

func NewMockFileStore(path string) FSRepo {
	mfs := &mockFileStore{path: "./", cfg: config.DefaultConfig()}
	if len(path) != 0 {
		mfs.path = path
	}
	return mfs
}

func (mfs *mockFileStore) Path() string {
	return mfs.path
}

func (mfs *mockFileStore) Config() *config.Config {
	return mfs.cfg
}

func (mfs *mockFileStore) ReplaceConfig(cfg *config.Config) error {
	mfs.cfg = cfg
	return nil
}

func (mfs *mockFileStore) TipsetFile() string {
	// SQLite In-Memory Model
	return ":memory:"
}

func (mfs *mockFileStore) SqliteFile() string {
	return filepath.Join(mfs.path, SqliteFile)
}

var _ FSRepo = (*mockFileStore)(nil)
