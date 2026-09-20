# API reference

Module `github.com/my-rv/hensu`. The root package is a facade over `internal/store`; these are the only exported names, and they are what the pre-1.0 promise covers.

## Store

```go
func NewStore(path string, opts ...StoreOption) *Store
func DefaultPath() string   // ".env"
```

| Method | |
|--------|--|
| `Path() string` | The file this Store was built for (immutable) |
| `Load(ctx) (Map, Format, error)` | Whole file as a flat map + its format |
| `Get(ctx, keys []string, mode Reveal) (map[string]Entry, error)` | Entries for named keys |
| `Read(ctx, key string) (string, error)` | One raw value |
| `Keys(ctx) ([]string, error)` | Sorted canonical names |
| `Set(ctx, updates map[string]string) error` | Atomic, locked, comment-preserving write |

Options: `WithFileSystem(FileSystem)`, `WithRegistry(CodecRegistry)`.

## Values

| Name | |
|------|--|
| `Map` | `map[string]string`, canonical keys |
| `Entry` | `{Value string; Defined bool}`, JSON `{value, defined}` |
| `Reveal` | `RevealPeek` (zero value), `RevealMask`, `RevealTrust` |
| `ShortValueMax` | `16` — runes at or below which `peek` shows nothing |
| `Lookup(m Map, keys []string, mode Reveal) map[string]Entry` | `Get` without a file |

## Keys

`CanonicalKey`, `PathSegments`, `KeyFromPath`, `NormalizeMap`.

## Documents and formats

`Format` (`FormatDotenv`, `FormatShexport`, `FormatYAML`, `FormatJSON`), `Document`, `ParseDocument`, `SerializeDocument`, `FormatCodec`, `CodecRegistry`.

## Filesystem

`FileSystem`, `OSFileSystem`, `MemFileSystem`, `NewMemFileSystem`.

## Errors

`ErrEmptyKey`, `ErrExpectedKeys` (sentinels, use `errors.Is`); `*ErrNotDefined`, `*InvalidArgument`, `*ParseError` (types, use `errors.As`).

## Version

`Version` — set at link time with `-ldflags "-X github.com/my-rv/hensu.Version=0.1.0"`.

Worked examples: [guide/library.md](../guide/library.md).
