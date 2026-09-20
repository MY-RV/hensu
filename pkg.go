// Package hensu is machine-local config: format-agnostic get/set/dump
// over key-value, YAML, or JSON files.
//
// This root package is a stable public facade. Implementation lives under
// internal/store; the CLI entrypoint is cmd/hensu.
//
// CLI: github.com/my-rv/hensu/cmd/hensu
package hensu

import "github.com/my-rv/hensu/internal/store"

// Version identifies the module/CLI. Override at link time:
//
//	-ldflags "-X github.com/my-rv/hensu.Version=1.2.3"
var Version = "0.1.0"

type (
	Format          = store.Format
	Map             = store.Map
	Entry           = store.Entry
	Reveal          = store.Reveal
	Document        = store.Document
	Store           = store.Store
	StoreOption     = store.StoreOption
	FileSystem      = store.FileSystem
	OSFileSystem    = store.OSFileSystem
	MemFileSystem   = store.MemFileSystem
	FormatCodec     = store.FormatCodec
	CodecRegistry   = store.CodecRegistry
	ErrNotDefined   = store.ErrNotDefined
	InvalidArgument = store.InvalidArgument
	ParseError      = store.ParseError
)

const (
	FormatDotenv   = store.FormatDotenv
	FormatShexport = store.FormatShexport
	FormatYAML     = store.FormatYAML
	FormatJSON     = store.FormatJSON

	RevealPeek  = store.RevealPeek
	RevealMask  = store.RevealMask
	RevealTrust = store.RevealTrust

	ShortValueMax = store.ShortValueMax
)

var (
	ErrEmptyKey     = store.ErrEmptyKey
	ErrExpectedKeys = store.ErrExpectedKeys
)

var (
	NewStore          = store.NewStore
	WithFileSystem    = store.WithFileSystem
	WithRegistry      = store.WithRegistry
	DefaultPath       = store.DefaultPath
	NewMemFileSystem  = store.NewMemFileSystem
	ParseDocument     = store.ParseDocument
	SerializeDocument = store.SerializeDocument
	CanonicalKey      = store.CanonicalKey
	PathSegments      = store.PathSegments
	KeyFromPath       = store.KeyFromPath
	NormalizeMap      = store.NormalizeMap
	Lookup            = store.Lookup
)
