package store

import (
	"context"
	"os"

	"github.com/my-rv/hensu/internal/atomicfile"
	"github.com/my-rv/hensu/internal/filelock"
)

// FileSystem abstracts store persistence.
// ReadFile must return (nil, nil) when the path does not exist.
type FileSystem interface {
	ReadFile(ctx context.Context, path string) ([]byte, error)
	WriteAtomic(ctx context.Context, path string, data []byte, mode os.FileMode) error
	WithLock(ctx context.Context, path string, fn func() error) error
}

// OSFileSystem is the production FileSystem backed by the local OS.
type OSFileSystem struct{}

// ReadFile implements FileSystem.
func (OSFileSystem) ReadFile(ctx context.Context, path string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return data, nil
}

// WriteAtomic implements FileSystem.
func (OSFileSystem) WriteAtomic(ctx context.Context, path string, data []byte, mode os.FileMode) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return atomicfile.Write(path, data, mode)
}

// WithLock implements FileSystem using an exclusive cross-process lock.
func (OSFileSystem) WithLock(ctx context.Context, path string, fn func() error) error {
	return filelock.With(ctx, path, fn)
}
