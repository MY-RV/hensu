package store_test

import (
	"context"
	"testing"

	"github.com/my-rv/hensu/internal/store"
)

func TestParseDocument_headerAndBody(t *testing.T) {
	doc, err := store.ParseDocument("# FORMAT: YAML\nFOO:\n  BAR: 1\n")
	if err != nil {
		t.Fatal(err)
	}
	if doc.Format != store.FormatYAML || !doc.HadHeader {
		t.Fatalf("%#v", doc)
	}
	if !stringsHasPrefix(doc.Body, "FOO:") {
		t.Fatalf("body %q", doc.Body)
	}
}

func TestParseDocument_noHeaderIsDotenv(t *testing.T) {
	doc, err := store.ParseDocument("FOO=bar\n")
	if err != nil {
		t.Fatal(err)
	}
	if doc.Format != store.FormatDotenv || doc.HadHeader {
		t.Fatalf("%#v", doc)
	}
}

func TestSerializeDocument_omitsHeaderForPlainDotenv(t *testing.T) {
	out := store.SerializeDocument(store.Document{Format: store.FormatDotenv}, "FOO=1\n")
	if out != "FOO=1\n" {
		t.Fatalf("%q", out)
	}
}

func TestSerializeDocument_forcesHeaderForJSON(t *testing.T) {
	out := store.SerializeDocument(store.Document{Format: store.FormatJSON}, "{\n  \"A\": \"1\"\n}\n")
	if !stringsHasPrefix(out, "# FORMAT: JSON\n") {
		t.Fatalf("%q", out)
	}
}

func TestMemFileSystem_roundTrip(t *testing.T) {
	fs := store.NewMemFileSystem()
	s := store.NewStore("/virtual/.env", store.WithFileSystem(fs))
	if err := s.Set(context.Background(), map[string]string{"FOO": "bar"}); err != nil {
		t.Fatal(err)
	}
	m, _, err := s.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if m["FOO"] != "bar" {
		t.Fatalf("%#v", m)
	}
	mode, ok := fs.Mode("/virtual/.env")
	if !ok || mode != 0o600 {
		t.Fatalf("mode %v ok=%v", mode, ok)
	}
}

func stringsHasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
