package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/my-rv/hensu"
	"github.com/my-rv/hensu/internal/cli"
)

func main() {
	if err := cli.New().Run(os.Args[1:]); err != nil {
		// A child of `hensu exec` already said whatever it had to say; forward
		// its status without prefixing hensu's own noise.
		var xe *cli.ExitError
		if errors.As(err, &xe) {
			os.Exit(xe.Code)
		}
		fmt.Fprintf(os.Stderr, "hensu: %v\n", err)
		os.Exit(exitCode(err))
	}
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	var xe *cli.ExitError
	if errors.As(err, &xe) {
		return xe.Code
	}
	var inv *hensu.InvalidArgument
	if errors.As(err, &inv) {
		return 2
	}
	if errors.Is(err, hensu.ErrEmptyKey) || errors.Is(err, hensu.ErrExpectedKeys) {
		return 2
	}
	var nd *hensu.ErrNotDefined
	if errors.As(err, &nd) {
		return 1
	}
	var pe *hensu.ParseError
	if errors.As(err, &pe) {
		return 1
	}
	return 1
}
