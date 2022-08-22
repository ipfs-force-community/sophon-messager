package filestore

import (
	"fmt"
	"path/filepath"

	"github.com/filecoin-project/venus-messager/config"
)

type mockFileStore struct {
	path  string
	cfg   *config.Config
	token []byte
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
	return filepath.Join(mfs.Path(), TipsetFile)
}

func (mfs *mockFileStore) SqliteFile() string {
	// SQLite In-Memory
	return ":memory:"
}

func (mfs *mockFileStore) GetToken() ([]byte, error) {
	if mfs.token != nil {
		return mfs.token, nil
	}
	return nil, fmt.Errorf("token not found")
}

func (mfs *mockFileStore) SaveToken(token []byte) error {
	mfs.token = token
	return nil
}

var _ FSRepo = (*mockFileStore)(nil)
