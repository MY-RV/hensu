package store_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/my-rv/hensu/internal/store"
)

func TestRevealPeek_utf8Runes(t *testing.T) {
	// 20 runes, multi-byte; must not split code points.
	long := "áéíóúáéíóúáéíóúáéíóú" // 20 runes
	got := store.Lookup(store.Map{"K": long}, []string{"K"}, store.RevealPeek)
	want := "áéíó***éíóú"
	if got["K"].Value != want {
		t.Fatalf("got %q want %q", got["K"].Value, want)
	}
}

func TestRevealPeek_shortMultibyteFullyHidden(t *testing.T) {
	short := "日本語テスト" // 6 runes
	got := store.Lookup(store.Map{"K": short}, []string{"K"}, store.RevealPeek)
	if got["K"].Value != "***" {
		t.Fatalf("got %q", got["K"].Value)
	}
}

func TestNormalizeMap_colonSugar(t *testing.T) {
	m := store.NormalizeMap(map[string]string{"internal:sessions:foo": "1"})
	if m["INTERNAL__SESSIONS__FOO"] != "1" {
		t.Fatalf("%#v", m)
	}
}

func TestCanonicalKey_edge(t *testing.T) {
	if store.CanonicalKey("  a::::b  ") != "A__B" {
		t.Fatal(store.CanonicalKey("  a::::b  "))
	}
	if store.CanonicalKey("") != "" || store.CanonicalKey("   ") != "" {
		t.Fatal("blank")
	}
}

func TestEntry_jsonNullWhenUndefined(t *testing.T) {
	b, err := json.Marshal(store.Entry{Defined: false})
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != `{"value":null,"defined":false}` {
		t.Fatalf("%s", b)
	}
	b, err = json.Marshal(store.Entry{Value: "x", Defined: true})
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != `{"value":"x","defined":true}` {
		t.Fatalf("%s", b)
	}
}

func TestRead_typedErrors(t *testing.T) {
	s := store.NewStore("/no/such/"+t.Name(), store.WithFileSystem(store.NewMemFileSystem()))
	_, err := s.Read(context.Background(), "")
	if !errors.Is(err, store.ErrEmptyKey) {
		t.Fatalf("empty: %v", err)
	}
	_, err = s.Read(context.Background(), "MISSING")
	var nd *store.ErrNotDefined
	if !errors.As(err, &nd) || nd.Key != "MISSING" {
		t.Fatalf("missing: %v", err)
	}
	_, err = s.Get(context.Background(), nil, store.RevealPeek)
	if !errors.Is(err, store.ErrExpectedKeys) {
		t.Fatalf("expected keys: %v", err)
	}
}

func TestStore_pathImmutable(t *testing.T) {
	fs := store.NewMemFileSystem()
	s := store.NewStore("/cfg/.env", store.WithFileSystem(fs))
	if s.Path() != "/cfg/.env" {
		t.Fatal(s.Path())
	}
	if err := s.Set(context.Background(), map[string]string{"A": "1"}); err != nil {
		t.Fatal(err)
	}
	// Path() is the only accessor; field is unexported.
	v, err := s.Read(context.Background(), "A")
	if err != nil || v != "1" {
		t.Fatal(v, err)
	}
}

func TestSet_concurrentWriters(t *testing.T) {
	fs := store.NewMemFileSystem()
	path := "/concurrent.env"
	s := store.NewStore(path, store.WithFileSystem(fs))
	if err := s.Set(context.Background(), map[string]string{"N": "0"}); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	errCh := make(chan error, 32)
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := "K" + string(rune('A'+i%26))
			if err := s.Set(context.Background(), map[string]string{key: "v"}); err != nil {
				errCh <- err
			}
		}(i)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Fatal(err)
	}
	keys, err := s.Keys(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) < 2 {
		t.Fatalf("keys=%v", keys)
	}
}

func TestSet_emptyKeyInUpdates(t *testing.T) {
	s := store.NewStore("/x", store.WithFileSystem(store.NewMemFileSystem()))
	err := s.Set(context.Background(), map[string]string{"  ": "v"})
	if !errors.Is(err, store.ErrEmptyKey) {
		t.Fatalf("%v", err)
	}
}

func TestSet_osFileMode0600(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	s := store.NewStore(path)
	if err := s.Set(context.Background(), map[string]string{"SECRET": "x"}); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode %o", info.Mode().Perm())
	}
}

func TestLoad_invalidFormat(t *testing.T) {
	fs := store.NewMemFileSystem()
	path := "/bad.env"
	_ = fs.WriteAtomic(context.Background(), path, []byte("# FORMAT: XML\nA=1\n"), 0o600)
	s := store.NewStore(path, store.WithFileSystem(fs))
	_, _, err := s.Load(context.Background())
	var pe *store.ParseError
	if !errors.As(err, &pe) {
		t.Fatalf("want ParseError, got %T %v", err, err)
	}
}
