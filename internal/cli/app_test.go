package cli_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/my-rv/hensu"
	"github.com/my-rv/hensu/internal/cli"
	"github.com/my-rv/hensu/internal/cli/command"
)

func TestParseGlobalFlags_fileBeforeCommand(t *testing.T) {
	opts, rest, err := cli.ParseGlobalFlags([]string{"-f", "/tmp/a.env", "keys"})
	if err != nil || opts.File != "/tmp/a.env" || len(rest) != 1 || rest[0] != "keys" {
		t.Fatalf("short: file=%q rest=%v err=%v", opts.File, rest, err)
	}
	opts, rest, err = cli.ParseGlobalFlags([]string{"--file", "/tmp/b.env", "read", "FOO"})
	if err != nil || opts.File != "/tmp/b.env" || len(rest) != 2 || rest[0] != "read" {
		t.Fatalf("long: file=%q rest=%v err=%v", opts.File, rest, err)
	}
	opts, rest, err = cli.ParseGlobalFlags([]string{"--file=/tmp/c.env", "get", "FOO"})
	if err != nil || opts.File != "/tmp/c.env" || rest[0] != "get" {
		t.Fatalf("eq: file=%q rest=%v err=%v", opts.File, rest, err)
	}
	opts, rest, err = cli.ParseGlobalFlags([]string{"-f=/tmp/d.env", "get", "FOO"})
	if err != nil || opts.File != "/tmp/d.env" || rest[0] != "get" {
		t.Fatalf("short eq: file=%q rest=%v err=%v", opts.File, rest, err)
	}
	opts, rest, err = cli.ParseGlobalFlags([]string{"keys", "-f", "/tmp/nope.env"})
	if err != nil || opts.File != "" || rest[0] != "keys" {
		t.Fatalf("after cmd: file=%q rest=%v err=%v", opts.File, rest, err)
	}
	_, _, err = cli.ParseGlobalFlags([]string{"-f"})
	if err == nil {
		t.Fatal("expected error for bare -f")
	}
}

func TestParseGlobalFlags_reveal(t *testing.T) {
	opts, rest, err := cli.ParseGlobalFlags([]string{"--reveal", "mask", "keys"})
	if err != nil || !opts.RevealFlag || opts.Reveal != hensu.RevealMask || rest[0] != "keys" {
		t.Fatalf("opts=%+v rest=%v err=%v", opts, rest, err)
	}
	opts, _, err = cli.ParseGlobalFlags([]string{"--reveal=trust", "get", "FOO"})
	if err != nil || opts.Reveal != hensu.RevealTrust {
		t.Fatalf("opts=%+v err=%v", opts, err)
	}
	opts, _, err = cli.ParseGlobalFlags([]string{"-r=peek", "get", "FOO"})
	if err != nil || opts.Reveal != hensu.RevealPeek {
		t.Fatalf("opts=%+v err=%v", opts, err)
	}
	opts, _, err = cli.ParseGlobalFlags([]string{"-r", "trust", "get", "FOO"})
	if err != nil || opts.Reveal != hensu.RevealTrust {
		t.Fatalf("opts=%+v err=%v", opts, err)
	}
	if _, _, err := cli.ParseGlobalFlags([]string{"--reveal", "loud"}); err == nil {
		t.Fatal("expected error for unknown mode")
	}
	if _, _, err := cli.ParseGlobalFlags([]string{"--reveal"}); err == nil {
		t.Fatal("expected error for bare --reveal")
	}
}

func TestRun_noCommandPrintsUsageAndReturnsInvalidArgument(t *testing.T) {
	var stderr bytes.Buffer
	app := cli.New()
	app.Stderr = &stderr
	app.LookupEnv = func(string) (string, bool) { return "", false }
	err := app.Run(nil)
	var inv *hensu.InvalidArgument
	if !errors.As(err, &inv) {
		t.Fatalf("want InvalidArgument, got %T %v", err, err)
	}
	if !strings.Contains(stderr.String(), "usage: hensu") {
		t.Fatalf("usage not printed: %q", stderr.String())
	}
}

func TestRun_unknownCommand(t *testing.T) {
	var stderr bytes.Buffer
	app := cli.New()
	app.Stderr = &stderr
	app.LookupEnv = func(string) (string, bool) { return "", false }
	err := app.Run([]string{"list"})
	if err == nil || !strings.Contains(err.Error(), `unknown command "list"`) {
		t.Fatalf("err=%v", err)
	}
}

func TestParseSetJSON(t *testing.T) {
	m, err := command.ParseSetArgs([]string{"--json", `{"A":2,"B":"x"}`}, strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	if m["A"] != "2" || m["B"] != "x" {
		t.Fatalf("%#v", m)
	}
}
