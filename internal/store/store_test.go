package store_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/my-rv/hensu/internal/store"
)

func TestCanonicalKey_colon(t *testing.T) {
	got := store.CanonicalKey("internal:sessions:max_minutes_timeout")
	want := "INTERNAL__SESSIONS__MAX_MINUTES_TIMEOUT"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	if store.CanonicalKey("INTERNAL__SESSIONS__MAX_MINUTES_TIMEOUT") != want {
		t.Fatal("underscore form mismatch")
	}
}

func TestLoad_missingIsEmpty(t *testing.T) {
	s := store.NewStore(filepath.Join(t.TempDir(), "missing"))
	m, f, err := s.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(m) != 0 || f != store.FormatDotenv {
		t.Fatalf("got %#v %q", m, f)
	}
}

func TestLoad_dotenvRetrocompat(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte("FOO=bar\n# c\nBAZ=1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	m, f, err := store.NewStore(path).Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if f != store.FormatDotenv || m["FOO"] != "bar" || m["BAZ"] != "1" {
		t.Fatalf("%#v %q", m, f)
	}
}

func TestLoad_shexport(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	raw := "# FORMAT: SHEXPORT\nexport FOO=\"bar\"\n"
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	m, f, err := store.NewStore(path).Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if f != store.FormatShexport || m["FOO"] != "bar" {
		t.Fatalf("%#v %q", m, f)
	}
}

func TestLoad_yamlNested(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.yaml")
	raw := `# FORMAT: YAML
PRESENTATION:
    API:
        PORT: 8080
INTERNAL:
    CHAT:
        ENABLED: true
        DEFAULT_MODEL: deepseek/flash
EXTERNALS:
    SUPABASE:
        URL: http://127.0.0.1:11021
        PUBLISHABLE_KEY: pk
`
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	m, f, err := store.NewStore(path).Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if f != store.FormatYAML {
		t.Fatalf("format %q", f)
	}
	checks := map[string]string{
		"PRESENTATION__API__PORT":              "8080",
		"INTERNAL__CHAT__ENABLED":              "true",
		"INTERNAL__CHAT__DEFAULT_MODEL":        "deepseek/flash",
		"EXTERNALS__SUPABASE__URL":             "http://127.0.0.1:11021",
		"EXTERNALS__SUPABASE__PUBLISHABLE_KEY": "pk",
	}
	for k, want := range checks {
		if m[k] != want {
			t.Fatalf("%s = %q want %q (%#v)", k, m[k], want, m)
		}
	}
}

func TestLoad_jsonNested(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.json")
	raw := `# FORMAT: JSON
{
  "INTERNAL": {
    "SESSIONS": {
      "ENABLED": true,
      "MAX": 10
    }
  },
  "FOO": "bar"
}
`
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	m, f, err := store.NewStore(path).Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if f != store.FormatJSON {
		t.Fatalf("format %q", f)
	}
	if m["FOO"] != "bar" || m["INTERNAL__SESSIONS__ENABLED"] != "true" || m["INTERNAL__SESSIONS__MAX"] != "10" {
		t.Fatalf("%#v", m)
	}
}

func TestRevealModes(t *testing.T) {
	long := "sk-secret-value-1234567890-abcdefgh"
	m := store.Map{
		"FOO":         "bar",
		"SECRET_LONG": long,
	}
	got := store.Lookup(m, []string{"FOO", "MISSING", "SECRET_LONG"}, store.RevealTrust)
	if !got["FOO"].Defined || got["FOO"].Value != "bar" {
		t.Fatalf("FOO %#v", got["FOO"])
	}
	if got["MISSING"].Defined {
		t.Fatal("MISSING should be undefined")
	}
	mask := store.Lookup(m, []string{"FOO"}, store.RevealMask)
	if mask["FOO"].Value != "***" {
		t.Fatalf("mask %#v", mask["FOO"])
	}
	peek := store.Lookup(m, []string{"FOO", "SECRET_LONG"}, store.RevealPeek)
	if peek["FOO"].Value != "***" {
		t.Fatalf("peek short %#v", peek["FOO"])
	}
	want := long[:4] + "***" + long[len(long)-4:]
	if peek["SECRET_LONG"].Value != want {
		t.Fatalf("peek long %#v want %q", peek["SECRET_LONG"].Value, want)
	}
}

func TestSet_dotenvPreservesComments(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	raw := "# FORMAT: DOTENV\n# header comment\nFOO=old # trailing\n\n# middle\nBAR=1\n"
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	s := store.NewStore(path)
	if err := s.Set(context.Background(), map[string]string{
		"FOO":                       "new",
		"INTERNAL:SESSIONS:ENABLED": "true",
	}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if !strings.Contains(text, "# header comment") || !strings.Contains(text, "# middle") {
		t.Fatalf("lost comments:\n%s", text)
	}
	if !strings.Contains(text, "# trailing") {
		t.Fatalf("lost trailing:\n%s", text)
	}
	if !strings.Contains(text, "INTERNAL__SESSIONS__ENABLED=true") {
		t.Fatalf("missing key:\n%s", text)
	}
	m, _, err := s.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if m["FOO"] != "new" || m["BAR"] != "1" || m["INTERNAL__SESSIONS__ENABLED"] != "true" {
		t.Fatalf("%#v", m)
	}
}

func TestSet_jsonRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.json")
	raw := "# FORMAT: JSON\n{\n  \"A\": \"1\"\n}\n"
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	s := store.NewStore(path)
	if err := s.Set(context.Background(), map[string]string{
		"A":               "2",
		"INTERNAL:NESTED": "x",
	}); err != nil {
		t.Fatal(err)
	}
	m, f, err := s.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if f != store.FormatJSON {
		t.Fatalf("format %q", f)
	}
	if m["A"] != "2" || m["INTERNAL__NESTED"] != "x" {
		t.Fatalf("%#v", m)
	}
}

func TestSet_createsMissingDotenv(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	s := store.NewStore(path)
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
	data, _ := os.ReadFile(path)
	if strings.Contains(string(data), "FORMAT:") {
		t.Fatalf("new dotenv should omit header:\n%s", data)
	}
}

func TestRead_missing(t *testing.T) {
	s := store.NewStore(filepath.Join(t.TempDir(), "x.env"))
	_, err := s.Read(context.Background(), "FOO")
	if err == nil {
		t.Fatal("expected error")
	}
	if _, ok := err.(*store.ErrNotDefined); !ok {
		t.Fatalf("want ErrNotDefined, got %T %v", err, err)
	}
}

func TestRead_exactBytes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	s := store.NewStore(path)
	if err := s.Set(context.Background(), map[string]string{"MULTI": "a\nb"}); err != nil {
		t.Fatal(err)
	}
	v, err := s.Read(context.Background(), "MULTI")
	if err != nil {
		t.Fatal(err)
	}
	if v != "a\nb" {
		t.Fatalf("got %q", v)
	}
}

func TestKeys_sorted(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	s := store.NewStore(path)
	_ = s.Set(context.Background(), map[string]string{"B": "1", "A": "2"})
	keys, err := s.Keys(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 2 || keys[0] != "A" || keys[1] != "B" {
		t.Fatalf("%v", keys)
	}
}

func TestSet_quoteRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	_ = os.WriteFile(path, []byte("X=1\n"), 0o600)
	s := store.NewStore(path)
	want := "a\x00b\nc"
	if err := s.Set(context.Background(), map[string]string{"X": want}); err != nil {
		t.Fatal(err)
	}
	m, _, err := s.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if m["X"] != want {
		t.Fatalf("got %q want %q", m["X"], want)
	}
}

func TestSet_yamlUpdate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.yaml")
	raw := `# FORMAT: YAML
INTERNAL:
    SESSIONS:
        ENABLED: "false"
        MAX_MINUTES_TIMEOUT: "10"
`
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	s := store.NewStore(path)
	if err := s.Set(context.Background(), map[string]string{
		"INTERNAL:SESSIONS:MAX_MINUTES_TIMEOUT": "360",
	}); err != nil {
		t.Fatal(err)
	}
	m, _, err := s.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if m["INTERNAL__SESSIONS__MAX_MINUTES_TIMEOUT"] != "360" {
		t.Fatalf("%#v", m)
	}
	if m["INTERNAL__SESSIONS__ENABLED"] != "false" {
		t.Fatalf("ENABLED clobbered: %#v", m)
	}
}

func TestPathSegments(t *testing.T) {
	got := store.PathSegments("INTERNAL__CHAT__MAX")
	if len(got) != 3 || got[0] != "INTERNAL" || got[2] != "MAX" {
		t.Fatalf("%v", got)
	}
	if store.KeyFromPath([]string{"internal", "chat"}) != "INTERNAL__CHAT" {
		t.Fatal(store.KeyFromPath([]string{"internal", "chat"}))
	}
}
