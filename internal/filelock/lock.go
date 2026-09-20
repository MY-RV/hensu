package filelock

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// LockSuffix is appended to the config path to form the coordination file.
// A sidecar is required: Set uses atomic rename, so flock on the data file
// would bind an inode that disappears on replace.
const LockSuffix = ".hensu-lock"

// With serializes fn under an exclusive lock for path.
//
// The lock is always taken on "<path>.hensu-lock" (stable inode). The sidecar
// is created with mode 0600 and intentionally kept — removing it races waiters.
// ctx cancels lock acquisition; a lock obtained after cancel is released.
func With(ctx context.Context, path string, fn func() error) error {
	if fn == nil {
		return fmt.Errorf("filelock: nil func")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	lockPath := path + LockSuffix
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return err
	}
	return hold(ctx, f, fn)
}

func hold(ctx context.Context, f *os.File, fn func() error) error {
	acquired := make(chan error, 1)
	go func() {
		acquired <- exclusive(f)
	}()

	select {
	case <-ctx.Done():
		go func() {
			if err := <-acquired; err == nil {
				_ = unlock(f)
			}
			_ = f.Close()
		}()
		return ctx.Err()
	case err := <-acquired:
		if err != nil {
			_ = f.Close()
			return err
		}
	}

	defer func() {
		_ = unlock(f)
		_ = f.Close()
	}()

	done := make(chan error, 1)
	go func() {
		done <- fn()
	}()

	select {
	case <-ctx.Done():
		err := <-done
		if err != nil {
			return err
		}
		return ctx.Err()
	case err := <-done:
		return err
	}
}
