package main

import (
	"errors"
	"testing"

	"github.com/my-rv/hensu"
)

func TestExitCode(t *testing.T) {
	if exitCode(nil) != 0 {
		t.Fatal()
	}
	if exitCode(&hensu.InvalidArgument{Msg: "x"}) != 2 {
		t.Fatal()
	}
	if exitCode(hensu.ErrEmptyKey) != 2 {
		t.Fatal()
	}
	if exitCode(&hensu.ErrNotDefined{Key: "A"}) != 1 {
		t.Fatal()
	}
	if exitCode(&hensu.ParseError{Msg: "p"}) != 1 {
		t.Fatal()
	}
	if exitCode(errors.New("io")) != 1 {
		t.Fatal()
	}
}
