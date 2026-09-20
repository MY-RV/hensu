package filelock_test

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/my-rv/hensu/internal/filelock"
)

func TestWith_serializes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.env")
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	var current int32
	var max int32
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := filelock.With(context.Background(), path, func() error {
				v := atomic.AddInt32(&current, 1)
				for {
					old := atomic.LoadInt32(&max)
					if v <= old || atomic.CompareAndSwapInt32(&max, old, v) {
						break
					}
				}
				atomic.AddInt32(&current, -1)
				return nil
			})
			if err != nil {
				t.Errorf("%v", err)
			}
		}()
	}
	wg.Wait()
	if max != 1 {
		t.Fatalf("max concurrency under lock = %d want 1", max)
	}
}

func TestWith_sidecarStableAcrossRename(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.env")
	lockPath := path + filelock.LockSuffix

	err := filelock.With(context.Background(), path, func() error {
		return os.WriteFile(path, []byte("A=1\n"), 0o600)
	})
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(lockPath)
	if err != nil {
		t.Fatalf("sidecar must exist: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("sidecar mode %o", info.Mode().Perm())
	}

	// Second set still coordinates via the same sidecar (data file may be replaced).
	err = filelock.With(context.Background(), path, func() error {
		return os.WriteFile(path, []byte("A=2\n"), 0o600)
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(lockPath); err != nil {
		t.Fatal(err)
	}
}

func TestWith_canceled(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.env")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := filelock.With(ctx, path, func() error { return nil })
	if err == nil {
		t.Fatal("expected cancel")
	}
}
