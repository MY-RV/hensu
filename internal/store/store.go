package store

import (
	"context"
	"path/filepath"
	"sort"
	"strings"
)

// Store is a file-backed flat env map. Path is fixed at construction.
type Store struct {
	path string
	fs   FileSystem
	reg  CodecRegistry
}

// StoreOption configures a Store.
type StoreOption func(*Store)

// WithFileSystem overrides the persistence backend.
func WithFileSystem(fs FileSystem) StoreOption {
	return func(s *Store) {
		if fs != nil {
			s.fs = fs
		}
	}
}

// WithRegistry overrides the codec registry.
func WithRegistry(reg CodecRegistry) StoreOption {
	return func(s *Store) {
		if reg != nil {
			s.reg = reg
		}
	}
}

// NewStore returns a Store for path (relative paths are caller's responsibility).
func NewStore(path string, opts ...StoreOption) *Store {
	s := &Store{
		path: filepath.Clean(path),
		fs:   OSFileSystem{},
		reg:  defaultCodecRegistry(),
	}
	for _, opt := range opts {
		opt(s)
	}
	s.reg = requireRegistry(s.reg)
	return s
}

// Path returns the store file path.
func (s *Store) Path() string { return s.path }

// DefaultPath is ./.env when no --file is given.
func DefaultPath() string {
	return ".env"
}

func ctxOrBackground(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

// Load parses the file into a flat map. Missing file → empty map.
func (s *Store) Load(ctx context.Context) (Map, Format, error) {
	ctx = ctxOrBackground(ctx)
	if err := ctx.Err(); err != nil {
		return nil, "", err
	}
	raw, format, body, err := s.readDocument(ctx)
	if err != nil {
		return nil, "", err
	}
	if len(raw) == 0 {
		return Map{}, FormatDotenv, nil
	}
	c, err := lookupCodec(s.reg, format)
	if err != nil {
		return nil, "", err
	}
	m, err := c.Decode(body)
	if err != nil {
		return nil, "", &ParseError{Msg: "decode " + string(format), Err: err}
	}
	return NormalizeMap(m), format, nil
}

// Get returns reveal entries for keys.
func (s *Store) Get(ctx context.Context, keys []string, mode Reveal) (map[string]Entry, error) {
	if err := requireKeys(keys); err != nil {
		return nil, err
	}
	m, _, err := s.Load(ctx)
	if err != nil {
		return nil, err
	}
	return Lookup(m, keys, mode), nil
}

// Read returns the exact raw value for one key.
func (s *Store) Read(ctx context.Context, key string) (string, error) {
	k := CanonicalKey(key)
	if k == "" {
		return "", ErrEmptyKey
	}
	m, _, err := s.Load(ctx)
	if err != nil {
		return "", err
	}
	v, ok := m[k]
	if !ok {
		return "", &ErrNotDefined{Key: k}
	}
	return v, nil
}

// Keys returns sorted canonical key names.
func (s *Store) Keys(ctx context.Context) ([]string, error) {
	m, _, err := s.Load(ctx)
	if err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys, nil
}

// Set merges updates into the file, preserving format and comments when possible.
// The read-modify-write is serialized via FileSystem.WithLock.
func (s *Store) Set(ctx context.Context, updates map[string]string) error {
	ctx = ctxOrBackground(ctx)
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(updates) == 0 {
		return nil
	}
	norm := make(map[string]string, len(updates))
	for k, v := range updates {
		nk := CanonicalKey(k)
		if nk == "" {
			return ErrEmptyKey
		}
		norm[nk] = v
	}

	return s.fs.WithLock(ctx, s.path, func() error {
		if err := ctx.Err(); err != nil {
			return err
		}
		raw, err := s.fs.ReadFile(ctx, s.path)
		if err != nil {
			return err
		}
		if len(raw) == 0 {
			c, err := lookupCodec(s.reg, FormatDotenv)
			if err != nil {
				return err
			}
			body, err := c.Apply("", norm)
			if err != nil {
				return err
			}
			if body != "" && !strings.HasSuffix(body, "\n") {
				body += "\n"
			}
			return s.fs.WriteAtomic(ctx, s.path, []byte(body), 0o600)
		}

		doc, err := ParseDocument(string(raw))
		if err != nil {
			return err
		}
		c, err := lookupCodec(s.reg, doc.Format)
		if err != nil {
			return err
		}
		outBody, err := c.Apply(doc.Body, norm)
		if err != nil {
			return err
		}
		return s.fs.WriteAtomic(ctx, s.path, []byte(SerializeDocument(doc, outBody)), 0o600)
	})
}

func (s *Store) readDocument(ctx context.Context) (raw []byte, format Format, body string, err error) {
	raw, err = s.fs.ReadFile(ctx, s.path)
	if err != nil {
		return nil, "", "", err
	}
	if len(raw) == 0 {
		return raw, FormatDotenv, "", nil
	}
	doc, err := ParseDocument(string(raw))
	if err != nil {
		return nil, "", "", err
	}
	return raw, doc.Format, doc.Body, nil
}

func requireKeys(keys []string) error {
	if len(keys) == 0 {
		return ErrExpectedKeys
	}
	for _, k := range keys {
		if CanonicalKey(k) == "" {
			return ErrEmptyKey
		}
	}
	return nil
}
