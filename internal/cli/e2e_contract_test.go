package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/my-rv/hensu"
	"github.com/my-rv/hensu/internal/cli"
)

// E2E coverage of docs/contract.md — CLI → Store → codecs.

func newApp(t *testing.T, cwd string, stdin string) (*cli.App, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	app := cli.New()
	app.Stdout = &stdout
	app.Stderr = &stderr
	app.Stdin = strings.NewReader(stdin)
	app.Getwd = func() (string, error) { return cwd, nil }
	app.LookupEnv = func(string) (string, bool) { return "", false }
	return app, &stdout, &stderr
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestE2E_versionAndHelp(t *testing.T) {
	cwd := t.TempDir()
	app, out, _ := newApp(t, cwd, "")
	if err := app.Run([]string{"--version"}); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out.String()) != hensu.Version {
		t.Fatalf("version: %q", out.String())
	}
	app, _, errBuf := newApp(t, cwd, "")
	if err := app.Run([]string{"--help"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(errBuf.String(), "usage: hensu") {
		t.Fatalf("help: %q", errBuf.String())
	}
}

func TestE2E_contextFlagsBeforeCommand(t *testing.T) {
	cwd := t.TempDir()
	path := filepath.Join(cwd, "a.env")
	writeFile(t, path, "FOO=bar\n")

	for _, args := range [][]string{
		{"-f", path, "-r", "trust", "read", "FOO"},
		{"--file", path, "--reveal", "trust", "read", "FOO"},
		{"--file=" + path, "--reveal=trust", "read", "FOO"},
	} {
		app, out, _ := newApp(t, cwd, "")
		if err := app.Run(args); err != nil {
			t.Fatalf("%v: %v", args, err)
		}
		if out.String() != "bar" {
			t.Fatalf("%v: got %q", args, out.String())
		}
	}

	// Flags after the keyword are not scanned as context — they become command args.
	app, _, _ := newApp(t, cwd, "")
	if err := app.Run([]string{"keys", "-f", path}); err == nil {
		t.Fatal("expected error when -f follows the command")
	}
}

func TestE2E_defaultPathIsCwdDotEnv(t *testing.T) {
	cwd := t.TempDir()
	writeFile(t, filepath.Join(cwd, ".env"), "FROM_DEFAULT=1\n")
	app, out, _ := newApp(t, cwd, "")
	if err := app.Run([]string{"-r", "trust", "read", "FROM_DEFAULT"}); err != nil {
		t.Fatal(err)
	}
	if out.String() != "1" {
		t.Fatalf("got %q", out.String())
	}
}

func TestE2E_relativeFileResolvedFromCwd(t *testing.T) {
	cwd := t.TempDir()
	writeFile(t, filepath.Join(cwd, "cfg", "x.env"), "REL=ok\n")
	app, out, _ := newApp(t, cwd, "")
	if err := app.Run([]string{"-f", "cfg/x.env", "-r", "trust", "read", "REL"}); err != nil {
		t.Fatal(err)
	}
	if out.String() != "ok" {
		t.Fatalf("got %q", out.String())
	}
}

func TestE2E_getRevealModes(t *testing.T) {
	cwd := t.TempDir()
	path := filepath.Join(cwd, ".env")
	long := "sk-secret-value-1234567890-abcdefgh"
	writeFile(t, path, "FOO=bar\nEMPTY=\nSECRET="+long+"\n")

	app, out, _ := newApp(t, cwd, "")
	if err := app.Run([]string{"-f", path, "get", "FOO", "MISSING", "EMPTY", "SECRET"}); err != nil {
		t.Fatal(err)
	}
	var get map[string]struct {
		Value   any  `json:"value"`
		Defined bool `json:"defined"`
	}
	if err := json.Unmarshal(out.Bytes(), &get); err != nil {
		t.Fatal(err)
	}
	if !get["FOO"].Defined || get["FOO"].Value != "***" {
		t.Fatalf("get FOO %#v", get["FOO"])
	}
	if get["MISSING"].Defined || get["MISSING"].Value != nil {
		t.Fatalf("get MISSING %#v", get["MISSING"])
	}
	if !get["EMPTY"].Defined || get["EMPTY"].Value != "" {
		t.Fatalf("get EMPTY %#v", get["EMPTY"])
	}
	want := long[:4] + "***" + long[len(long)-4:]
	if get["SECRET"].Value != want {
		t.Fatalf("default peek long %#v want %q", get["SECRET"].Value, want)
	}

	app, out, _ = newApp(t, cwd, "")
	if err := app.Run([]string{"-f", path, "-r", "mask", "get", "FOO", "EMPTY"}); err != nil {
		t.Fatal(err)
	}
	var mask map[string]struct {
		Value   any  `json:"value"`
		Defined bool `json:"defined"`
	}
	if err := json.Unmarshal(out.Bytes(), &mask); err != nil {
		t.Fatal(err)
	}
	if mask["FOO"].Value != "***" || mask["EMPTY"].Value != "" {
		t.Fatalf("mask %#v", mask)
	}

	app, out, _ = newApp(t, cwd, "")
	if err := app.Run([]string{"-f", path, "-r", "trust", "get", "FOO", "SECRET"}); err != nil {
		t.Fatal(err)
	}
	var trust map[string]struct {
		Value   any  `json:"value"`
		Defined bool `json:"defined"`
	}
	if err := json.Unmarshal(out.Bytes(), &trust); err != nil {
		t.Fatal(err)
	}
	if trust["FOO"].Value != "bar" || trust["SECRET"].Value != long {
		t.Fatalf("trust %#v", trust)
	}
}

func TestE2E_readKeysAndSet(t *testing.T) {
	cwd := t.TempDir()
	path := filepath.Join(cwd, ".env")
	writeFile(t, path, "B=2\nA=1\n")

	app, out, _ := newApp(t, cwd, "")
	if err := app.Run([]string{"-f", path, "keys"}); err != nil {
		t.Fatal(err)
	}
	var keys []string
	if err := json.Unmarshal(out.Bytes(), &keys); err != nil {
		t.Fatal(err)
	}
	if len(keys) != 2 || keys[0] != "A" || keys[1] != "B" {
		t.Fatalf("keys %v", keys)
	}

	app, out, _ = newApp(t, cwd, "")
	if err := app.Run([]string{"-f", path, "-r", "trust", "read", "A"}); err != nil {
		t.Fatal(err)
	}
	if out.String() != "1" {
		t.Fatalf("read %q", out.String())
	}

	app, _, _ = newApp(t, cwd, "")
	if err := app.Run([]string{"-f", path, "set", "MULTI", "hello world"}); err != nil {
		t.Fatal(err)
	}
	app, out, _ = newApp(t, cwd, "")
	if err := app.Run([]string{"-f", path, "-r", "trust", "read", "MULTI"}); err != nil {
		t.Fatal(err)
	}
	if out.String() != "hello world" {
		t.Fatalf("set spaces %q", out.String())
	}

	app, _, _ = newApp(t, cwd, "")
	if err := app.Run([]string{"-f", path, "set", "--json", `{"X":2,"Y":"z"}`}); err != nil {
		t.Fatal(err)
	}
	app, _, _ = newApp(t, cwd, `{"Z":true}`)
	if err := app.Run([]string{"-f", path, "set", "--json", "-"}); err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]string{"X": "2", "Y": "z", "Z": "true"} {
		app, out, _ = newApp(t, cwd, "")
		if err := app.Run([]string{"-f", path, "-r", "trust", "read", key}); err != nil {
			t.Fatal(err)
		}
		if out.String() != want {
			t.Fatalf("%s = %q want %q", key, out.String(), want)
		}
	}
}

func TestE2E_formatsAndMissingFile(t *testing.T) {
	cwd := t.TempDir()

	// Missing file → empty on read / keys.
	missing := filepath.Join(cwd, "missing.env")
	app, out, _ := newApp(t, cwd, "")
	if err := app.Run([]string{"-f", missing, "keys"}); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out.String()) != "[]" {
		t.Fatalf("missing keys %q", out.String())
	}

	// set creates DOTENV without FORMAT header.
	created := filepath.Join(cwd, "new.env")
	app, _, _ = newApp(t, cwd, "")
	if err := app.Run([]string{"-f", created, "set", "FOO", "bar"}); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(created)
	if strings.Contains(string(raw), "FORMAT:") {
		t.Fatalf("new file should omit header:\n%s", raw)
	}

	// DOTENV without header.
	dotenv := filepath.Join(cwd, "plain.env")
	writeFile(t, dotenv, "FOO=1\n# c\nBAR=2\n")
	app, out, _ = newApp(t, cwd, "")
	if err := app.Run([]string{"-f", dotenv, "get", "FOO"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `"value": "***"`) {
		t.Fatalf("%s", out.String())
	}

	// Comment-preserving set (DOTENV).
	app, _, _ = newApp(t, cwd, "")
	if err := app.Run([]string{"-f", dotenv, "set", "FOO", "9"}); err != nil {
		t.Fatal(err)
	}
	raw, _ = os.ReadFile(dotenv)
	if !strings.Contains(string(raw), "# c") {
		t.Fatalf("lost comment:\n%s", raw)
	}

	// SHEXPORT write uses export.
	shex := filepath.Join(cwd, "sh.env")
	writeFile(t, shex, "# FORMAT: SHEXPORT\nexport A=\"1\"\n")
	app, _, _ = newApp(t, cwd, "")
	if err := app.Run([]string{"-f", shex, "set", "B", "2"}); err != nil {
		t.Fatal(err)
	}
	raw, _ = os.ReadFile(shex)
	if !strings.Contains(string(raw), "export B=") {
		t.Fatalf("shexport:\n%s", raw)
	}

	// YAML nested ↔ flat with :
	yamlPath := filepath.Join(cwd, "cfg.yaml")
	writeFile(t, yamlPath, `# FORMAT: YAML
INTERNAL:
    SESSIONS:
        FOO: old
`)
	app, _, _ = newApp(t, cwd, "")
	if err := app.Run([]string{"-f", yamlPath, "set", "INTERNAL:SESSIONS:FOO", "new"}); err != nil {
		t.Fatal(err)
	}
	app, out, _ = newApp(t, cwd, "")
	if err := app.Run([]string{"-f", yamlPath, "-r", "trust", "read", "INTERNAL__SESSIONS__FOO"}); err != nil {
		t.Fatal(err)
	}
	if out.String() != "new" {
		t.Fatalf("yaml %q", out.String())
	}

	// JSON nested.
	jsonPath := filepath.Join(cwd, "cfg.json")
	writeFile(t, jsonPath, `# FORMAT: JSON
{"INTERNAL":{"SESSIONS":{"N":1}},"A":"x"}
`)
	app, out, _ = newApp(t, cwd, "")
	if err := app.Run([]string{"-f", jsonPath, "-r", "trust", "read", "INTERNAL:SESSIONS:N"}); err != nil {
		t.Fatal(err)
	}
	if out.String() != "1" {
		t.Fatalf("json nested %q", out.String())
	}
}

func TestE2E_errorsExitNonZero(t *testing.T) {
	cwd := t.TempDir()
	path := filepath.Join(cwd, ".env")
	writeFile(t, path, "FOO=1\n")

	cases := []struct {
		name  string
		args  []string
		stdin string
	}{
		{"unknown", []string{"nope"}, ""},
		{"read missing", []string{"-f", path, "read", "NOPE"}, ""},
		{"read arity", []string{"-f", path, "read"}, ""},
		{"get empty", []string{"-f", path, "get"}, ""},
		{"set bad", []string{"-f", path, "set"}, ""},
		{"set json empty", []string{"-f", path, "set", "--json", "{}"}, ""},
		{"bare -f", []string{"-f"}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app, _, _ := newApp(t, cwd, tc.stdin)
			if err := app.Run(tc.args); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestE2E_memFileSystemInjection(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	fs := hensu.NewMemFileSystem()
	app, _, _ := newApp(t, dir, "")
	app.NewStore = func(p string) *hensu.Store {
		return hensu.NewStore(p, hensu.WithFileSystem(fs))
	}
	if err := app.Run([]string{"-f", path, "set", "INTERNAL:FOO", "bar baz"}); err != nil {
		t.Fatal(err)
	}
	app2, out2, _ := newApp(t, dir, "")
	app2.NewStore = app.NewStore
	if err := app2.Run([]string{"-f", path, "-r", "trust", "read", "INTERNAL:FOO"}); err != nil {
		t.Fatal(err)
	}
	if out2.String() != "bar baz" {
		t.Fatalf("got %q", out2.String())
	}
}

func TestE2E_libraryAPIMatchesContract(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	s := hensu.NewStore(path)
	m, f, err := s.Load(context.Background())
	if err != nil || len(m) != 0 || f != hensu.FormatDotenv {
		t.Fatalf("missing load %#v %q %v", m, f, err)
	}
	if hensu.CanonicalKey("internal:sessions:foo") != "INTERNAL__SESSIONS__FOO" {
		t.Fatal(hensu.CanonicalKey("internal:sessions:foo"))
	}
	segs := hensu.PathSegments("INTERNAL__SESSIONS__FOO")
	if len(segs) != 3 {
		t.Fatalf("%v", segs)
	}
	if err := s.Set(context.Background(), map[string]string{"internal:a": "1", "B": "2"}); err != nil {
		t.Fatal(err)
	}
	got, err := s.Get(context.Background(), []string{"INTERNAL:A", "MISSING"}, hensu.RevealTrust)
	if err != nil || !got["INTERNAL__A"].Defined || got["INTERNAL__A"].Value != "1" {
		t.Fatalf("%#v %v", got, err)
	}
	v, err := s.Read(context.Background(), "B")
	if err != nil || v != "2" {
		t.Fatal(v, err)
	}
	keys, err := s.Keys(context.Background())
	if err != nil || len(keys) != 2 {
		t.Fatal(keys, err)
	}
}
