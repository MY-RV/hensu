package store

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Format identifies on-disk serialization.
type Format string

const (
	FormatDotenv   Format = "DOTENV"
	FormatShexport Format = "SHEXPORT"
	FormatYAML     Format = "YAML"
	FormatJSON     Format = "JSON"
)

// Valid reports whether f is a known Format.
func (f Format) Valid() bool {
	switch f {
	case FormatDotenv, FormatShexport, FormatYAML, FormatJSON:
		return true
	default:
		return false
	}
}

// Map is the canonical flat KEY=value store (keys already canonical).
type Map map[string]string

// Entry is one get/def/spy result. When Defined is false, JSON value is null.
type Entry struct {
	Value   string `json:"-"`
	Defined bool   `json:"defined"`
}

// MarshalJSON implements the contract shape {value, defined}.
func (e Entry) MarshalJSON() ([]byte, error) {
	if !e.Defined {
		return json.Marshal(struct {
			Value   any  `json:"value"`
			Defined bool `json:"defined"`
		}{Value: nil, Defined: false})
	}
	return json.Marshal(struct {
		Value   string `json:"value"`
		Defined bool   `json:"defined"`
	}{Value: e.Value, Defined: true})
}

// UnmarshalJSON accepts the contract shape.
func (e *Entry) UnmarshalJSON(data []byte) error {
	var raw struct {
		Value   any  `json:"value"`
		Defined bool `json:"defined"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	e.Defined = raw.Defined
	if !raw.Defined || raw.Value == nil {
		e.Value = ""
		return nil
	}
	switch v := raw.Value.(type) {
	case string:
		e.Value = v
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		e.Value = string(b)
	}
	return nil
}

// Reveal selects how much of a defined value may be printed.
//
// The zero value is RevealPeek, so a caller that sets nothing still cannot
// print a secret whole. Raw output is a grant, and grants are asked for.
type Reveal int

const (
	RevealPeek  Reveal = iota // tip for long values, "***" for short ones
	RevealMask                // "***" when non-empty
	RevealTrust               // raw
)

// ShortValueMax is the max rune length fully hidden under RevealPeek.
const ShortValueMax = 16

var (
	// ErrEmptyKey is returned when a key is blank after canonicalization.
	ErrEmptyKey = errors.New("empty key")
	// ErrExpectedKeys is returned when get/def/spy receive no keys.
	ErrExpectedKeys = errors.New("expected one or more keys")
)

// ErrNotDefined is returned by Read when the key is missing.
type ErrNotDefined struct {
	Key string
}

func (e *ErrNotDefined) Error() string {
	return fmt.Sprintf("%s not defined", e.Key)
}

// InvalidArgument is a caller input error (CLI args, empty set JSON, etc.).
type InvalidArgument struct {
	Msg string
}

func (e *InvalidArgument) Error() string { return e.Msg }

// ParseError wraps document/codec parse failures.
type ParseError struct {
	Msg string
	Err error
}

func (e *ParseError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Msg, e.Err)
	}
	return e.Msg
}

func (e *ParseError) Unwrap() error { return e.Err }
