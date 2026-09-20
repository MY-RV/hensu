package cli_test

import (
	"context"
	"errors"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/my-rv/hensu"
	"github.com/my-rv/hensu/internal/cli"
	"github.com/my-rv/hensu/internal/cli/command"
)

// Coverage of the agent-facing contract: masked output, explicit raw grants,
// and a way to hand values to a process instead of to a transcript.

const secret = "sk-live-0123456789abcdefghij"

// envApp builds an App whose ambient reveal mode comes from mode ("" = unset).
func envApp(t *testing.T, cwd, mode string) *cli.App {
	t.Helper()
	app, _, _ := newApp(t, cwd, "")
	app.LookupEnv = func(k string) (string, bool) {
		if k == command.RevealEnv {
			return mode, mode != ""
		}
		return "", false
	}
	return app
}

func fixture(t *testing.T) (cwd, path string) {
	t.Helper()
	cwd = t.TempDir()
	path = filepath.Join(cwd, ".env")
	writeFile(t, path, "TOKEN="+secret+"\nDEBUG=true\nPORT=8080\n")
	return cwd, path
}

func TestAmbientReveal_maskAppliesToGet(t *testing.T) {
	cwd, path := fixture(t)
	app := envApp(t, cwd, "mask")
	var out strings.Builder
	app.Stdout = &out
	if err := app.Run([]string{"-f", path, "get", "TOKEN"}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), secret) {
		t.Fatalf("get leaked the value: %q", out.String())
	}
	if !strings.Contains(out.String(), "***") {
		t.Fatalf("get did not mask: %q", out.String())
	}
}

func TestAmbientReveal_invalidValueIsAnErrorAndLeaksNothing(t *testing.T) {
	cwd, path := fixture(t)
	app := envApp(t, cwd, "loud")
	var out, errOut strings.Builder
	app.Stdout = &out
	app.Stderr = &errOut
	err := app.Run([]string{"-f", path, "get", "TOKEN"})
	if err == nil {
		t.Fatal("a typo in HENSU_REVEAL must not fall back to raw output")
	}
	var inv *hensu.InvalidArgument
	if !errors.As(err, &inv) {
		t.Fatalf("want InvalidArgument (exit 2), got %T: %v", err, err)
	}
	if strings.Contains(out.String()+errOut.String(), secret) {
		t.Fatalf("leaked while erroring: stdout=%q stderr=%q", out.String(), errOut.String())
	}
}

func TestRead_refusesUnlessTrust(t *testing.T) {
	cwd, path := fixture(t)
	app := envApp(t, cwd, "")
	var out strings.Builder
	app.Stdout = &out
	err := app.Run([]string{"-f", path, "read", "TOKEN"})
	if err == nil || strings.Contains(out.String(), secret) {
		t.Fatalf("read must refuse under a floor: err=%v out=%q", err, out.String())
	}
	var inv *hensu.InvalidArgument
	if !errors.As(err, &inv) {
		t.Fatalf("want InvalidArgument (exit 2), got %T", err)
	}
	if !strings.Contains(err.Error(), "read prints the raw value") {
		t.Fatalf("the refusal should explain its grant: %v", err)
	}
	app = envApp(t, cwd, "trust")
	out.Reset()
	app.Stdout = &out
	if err := app.Run([]string{"-f", path, "read", "TOKEN"}); err != nil {
		t.Fatal(err)
	}
	if out.String() != secret {
		t.Fatalf("trust read got %q", out.String())
	}
}

type capturedRun struct {
	name string
	args []string
	env  []string
}

func execApp(t *testing.T, cwd, mode string, parent []string, code int) (*cli.App, *capturedRun, *strings.Builder, *strings.Builder) {
	t.Helper()
	app := envApp(t, cwd, mode)
	var out, errOut strings.Builder
	app.Stdout = &out
	app.Stderr = &errOut
	got := &capturedRun{}
	app.Environ = func() []string { return parent }
	app.Runner = func(_ context.Context, name string, args, env []string, _ io.Reader, _, _ io.Writer) error {
		got.name, got.args, got.env = name, args, env
		if code != 0 {
			return &command.ExitError{Cmd: name, Code: code}
		}
		return nil
	}
	return app, got, &out, &errOut
}

func envValue(env []string, key string) (string, bool) {
	for _, kv := range env {
		if strings.HasPrefix(kv, key+"=") {
			return strings.TrimPrefix(kv, key+"="), true
		}
	}
	return "", false
}

func TestExec_handsValuesToTheProcessNotTheCaller(t *testing.T) {
	cwd, path := fixture(t)
	app, got, out, _ := execApp(t, cwd, "mask", []string{"PATH=/usr/bin", "TOKEN=stale"}, 0)
	if err := app.Run([]string{"-f", path, "exec", "--", "myapp", "serve"}); err != nil {
		t.Fatal(err)
	}
	if got.name != "myapp" || len(got.args) != 1 || got.args[0] != "serve" {
		t.Fatalf("command: %q %v", got.name, got.args)
	}
	if v, _ := envValue(got.env, "TOKEN"); v != secret {
		t.Fatalf("the file must win over a stale inherited value, got %q", v)
	}
	if v, _ := envValue(got.env, "PATH"); v != "/usr/bin" {
		t.Fatalf("inherited env lost: %q", v)
	}
	if strings.Contains(out.String(), secret) {
		t.Fatalf("exec printed a value: %q", out.String())
	}
}

func TestExec_skipsNamesThatAreNotEnvironmentVariables(t *testing.T) {
	cwd := t.TempDir()
	path := filepath.Join(cwd, "conf.yaml")
	writeFile(t, path, "# FORMAT: YAML\nok: 1\n\"not a name\": 2\n")
	app, got, _, errOut := execApp(t, cwd, "", nil, 0)
	if err := app.Run([]string{"-f", path, "exec", "myapp"}); err != nil {
		t.Fatal(err)
	}
	if _, ok := envValue(got.env, "OK"); !ok {
		t.Fatalf("valid key dropped: %v", got.env)
	}
	for _, kv := range got.env {
		if strings.Contains(kv, "NOT A NAME") {
			t.Fatalf("smuggled an invalid name into the child: %q", kv)
		}
	}
	if !strings.Contains(errOut.String(), "NOT A NAME") {
		t.Fatalf("the skip should be reported: %q", errOut.String())
	}
}

func TestExec_forwardsTheChildExitCode(t *testing.T) {
	cwd, path := fixture(t)
	app, _, _, _ := execApp(t, cwd, "", nil, 3)
	err := app.Run([]string{"-f", path, "exec", "myapp"})
	var xe *cli.ExitError
	if !errors.As(err, &xe) || xe.Code != 3 {
		t.Fatalf("want exit code 3, got %T %v", err, err)
	}
}

func TestExec_needsACommand(t *testing.T) {
	cwd, path := fixture(t)
	app, _, _, _ := execApp(t, cwd, "", nil, 0)
	err := app.Run([]string{"-f", path, "exec", "--"})
	var inv *hensu.InvalidArgument
	if !errors.As(err, &inv) {
		t.Fatalf("want InvalidArgument (exit 2), got %T %v", err, err)
	}
}
