package store

import (
	"context"
	"os"
	"sync"
)

// MemFileSystem is an in-memory FileSystem for tests and hermetic use.
type MemFileSystem struct {
	mu        sync.Mutex
	files     map[string][]byte
	modes     map[string]os.FileMode
	pathLocks map[string]*sync.Mutex
}

// NewMemFileSystem returns an empty memory-backed FileSystem.
func NewMemFileSystem() *MemFileSystem {
	return &MemFileSystem{
		files:     make(map[string][]byte),
		modes:     make(map[string]os.FileMode),
		pathLocks: make(map[string]*sync.Mutex),
	}
}

// ReadFile implements FileSystem. Missing path → (nil, nil).
func (m *MemFileSystem) ReadFile(ctx context.Context, path string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	data, ok := m.files[path]
	if !ok {
		return nil, nil
	}
	out := make([]byte, len(data))
	copy(out, data)
	return out, nil
}

// WriteAtomic implements FileSystem.
func (m *MemFileSystem) WriteAtomic(ctx context.Context, path string, data []byte, mode os.FileMode) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := make([]byte, len(data))
	copy(cp, data)
	m.files[path] = cp
	m.modes[path] = mode
	return nil
}

// WithLock serializes fn per path without holding the data mutex across fn.
func (m *MemFileSystem) WithLock(ctx context.Context, path string, fn func() error) error {
	if fn == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	lk := m.lockFor(path)

	acquired := make(chan struct{})
	go func() {
		lk.Lock()
		close(acquired)
	}()

	select {
	case <-ctx.Done():
		go func() {
			<-acquired
			lk.Unlock()
		}()
		return ctx.Err()
	case <-acquired:
	}

	defer lk.Unlock()

	done := make(chan error, 1)
	go func() { done <- fn() }()
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

func (m *MemFileSystem) lockFor(path string) *sync.Mutex {
	m.mu.Lock()
	defer m.mu.Unlock()
	lk, ok := m.pathLocks[path]
	if !ok {
		lk = &sync.Mutex{}
		m.pathLocks[path] = lk
	}
	return lk
}

// Mode returns the last written mode for path, if any.
func (m *MemFileSystem) Mode(path string) (os.FileMode, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	mode, ok := m.modes[path]
	return mode, ok
}
