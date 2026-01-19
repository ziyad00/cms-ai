package assets

import (
	"context"
	"io"
	"os"
	"path/filepath"
)

type Storage interface {
	Open(ctx context.Context, path string) (io.ReadSeekCloser, error)
	EnsureDir(ctx context.Context, dir string) error
	Join(elem ...string) string
}

type LocalStorage struct{}

func (LocalStorage) Open(_ context.Context, path string) (io.ReadSeekCloser, error) {
	return os.Open(path)
}

func (LocalStorage) EnsureDir(_ context.Context, dir string) error {
	return os.MkdirAll(dir, 0o755)
}

func (LocalStorage) Join(elem ...string) string {
	return filepath.Join(elem...)
}
