package command_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/my-rv/hensu"
	"github.com/my-rv/hensu/internal/cli/command"
)

func TestParseSetArgs_edges(t *testing.T) {
	_, err := command.ParseSetArgs(nil, strings.NewReader(""))
	var inv *hensu.InvalidArgument
	if !errors.As(err, &inv) {
		t.Fatalf("%v", err)
	}

	_, err = command.ParseSetArgs([]string{"--json"}, strings.NewReader(""))
	if !errors.As(err, &inv) {
		t.Fatalf("%v", err)
	}

	_, err = command.ParseSetArgs([]string{"--json", "{}"}, strings.NewReader(""))
	if !errors.As(err, &inv) {
		t.Fatalf("%v", err)
	}

	_, err = command.ParseSetArgs([]string{"--json", "{"}, strings.NewReader(""))
	var pe *hensu.ParseError
	if !errors.As(err, &pe) {
		t.Fatalf("%T %v", err, err)
	}

	m, err := command.ParseSetArgs([]string{"--json", "-"}, strings.NewReader(`{"A":1,"B":true}`))
	if err != nil {
		t.Fatal(err)
	}
	if m["A"] != "1" || m["B"] != "true" {
		t.Fatalf("%#v", m)
	}

	m, err = command.ParseSetArgs([]string{"K", "v", "with", "spaces"}, strings.NewReader(""))
	if err != nil || m["K"] != "v with spaces" {
		t.Fatalf("%#v %v", m, err)
	}
}
