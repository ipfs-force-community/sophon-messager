package filestore

import (
	"path/filepath"

	"github.com/filecoin-project/venus-messager/config"
)

type mockFileStore struct {
	path string
	cfg  *config.Config
}

func NewMockFileStore(cfg *config.Config) FSRepo {
	mfs := &mockFileStore{path: "./", cfg: config.DefaultConfig()}
	if cfg != nil {
		mfs.cfg = cfg
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
	return filepath.Join(mfs.path, TipsetFile)
}

func (mfs *mockFileStore) SqliteFile() string {
	return filepath.Join(mfs.path, SqliteFile)
}

var _ FSRepo = (*mockFileStore)(nil)
