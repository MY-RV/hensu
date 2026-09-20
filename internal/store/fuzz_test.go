package store_test

import (
	"context"
	"testing"
	"unicode/utf8"

	"github.com/my-rv/hensu/internal/store"
)

func FuzzCanonicalKey(f *testing.F) {
	f.Add("INTERNAL:FOO")
	f.Add("a::::b")
	f.Add("")
	f.Add("already__flat")
	f.Fuzz(func(t *testing.T, in string) {
		out := store.CanonicalKey(in)
		if out != store.CanonicalKey(out) {
			t.Fatalf("not idempotent: %q -> %q", in, out)
		}
		if out != "" && out != store.CanonicalKey(out) {
			t.Fatal("unstable")
		}
	})
}

func FuzzParseDocument(f *testing.F) {
	f.Add("")
	f.Add("# FORMAT: DOTENV\nA=1\n")
	f.Add("# FORMAT: YAML\nA: b\n")
	f.Add("# FORMAT: JSON\n{\"A\":\"1\"}\n")
	f.Add("# FORMAT: NOPE\n")
	f.Fuzz(func(t *testing.T, raw string) {
		doc, err := store.ParseDocument(raw)
		if err != nil {
			return
		}
		if !doc.Format.Valid() {
			t.Fatalf("invalid format %q", doc.Format)
		}
		_ = store.SerializeDocument(doc, doc.Body)
	})
}

func FuzzRevealPeekUTF8(f *testing.F) {
	f.Add("short")
	f.Add("áéíóúáéíóúáéíóúáéíóú")
	f.Fuzz(func(t *testing.T, v string) {
		if !utf8.ValidString(v) {
			return
		}
		got := store.Lookup(store.Map{"K": v}, []string{"K"}, store.RevealPeek)["K"].Value
		if !utf8.ValidString(got) {
			t.Fatalf("invalid utf8 tip %q from %q", got, v)
		}
	})
}

func TestStore_canceledContext(t *testing.T) {
	fs := store.NewMemFileSystem()
	s := store.NewStore("/x.env", store.WithFileSystem(fs))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := s.Set(ctx, map[string]string{"A": "1"}); err == nil {
		t.Fatal("expected canceled set")
	}
	if _, _, err := s.Load(ctx); err == nil {
		t.Fatal("expected canceled load")
	}
}
