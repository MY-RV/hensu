# Library: embedding the Store

```go
import "github.com/my-rv/hensu"
```

The root package is a thin facade over the implementation. Everything the CLI does, a Go program can do directly, and nothing the CLI does is private to it.

## A whole program

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/my-rv/hensu"
)

func main() {
	ctx := context.Background()
	s := hensu.NewStore(".env")

	keys, err := s.Keys(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("configured:", keys)

	entries, err := s.Get(ctx, []string{"PORT", "MISSING"}, hensu.RevealPeek)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(entries["PORT"].Defined, entries["PORT"].Value) // true ***
	fmt.Println(entries["MISSING"].Defined)                     // false

	if err := s.Set(ctx, map[string]string{"PORT": "8080"}); err != nil {
		log.Fatal(err)
	}
}
```

## Reading

```go
m, format, err := s.Load(ctx)          // the whole file, flat, plus the format it was in
v, err := s.Read(ctx, "INTERNAL:FOO")  // one raw value
keys, err := s.Keys(ctx)               // sorted canonical names, no values
```

`Load` gives you raw values — the library has no `trust` gate, because in a Go program the caller and the reader are the same thing. The gate exists in the CLI, where they are not.

## Reveal in the library

`Get` returns the entries the CLI prints, masked by the mode you pass:

```go
entries, _ := s.Get(ctx, []string{"TOKEN"}, hensu.RevealPeek)  // "sk-l***cdef"
entries, _ = s.Get(ctx, []string{"TOKEN"}, hensu.RevealMask)   // "***"
entries, _ = s.Get(ctx, []string{"TOKEN"}, hensu.RevealTrust)  // raw
```

`RevealPeek` is the zero value, so `var mode hensu.Reveal` is already the careful one — a struct field you forgot to fill cannot leak.

To mask values you already hold, without touching a file:

```go
entries := hensu.Lookup(myMap, []string{"TOKEN"}, hensu.RevealPeek)
```

## Writing

```go
err := s.Set(ctx, map[string]string{
	"PORT":                "8080",
	"INTERNAL:SESSIONS:X": "1",  // canonicalized for you
})
```

Atomic, mode `0600`, serialized against other processes, comments preserved. See [set.md](./set.md) — the library and the CLI go through exactly the same path.

## Context

Every method takes a `context.Context`, and it is honored: cancellation is checked while acquiring the cross-process write lock and between IO steps, so a cancelled `Set` returns instead of blocking behind another writer.

```go
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()
err := s.Set(ctx, updates)
```

## Testing without a disk

```go
fs := hensu.NewMemFileSystem()
s := hensu.NewStore("/app/.env", hensu.WithFileSystem(fs))
```

Same semantics, no temp directories, no cleanup. `WithRegistry` swaps the codec set the same way when you want to test one format in isolation.

## Errors worth matching

| Error | Meaning | How to match |
|-------|---------|--------------|
| `ErrEmptyKey` | A key was blank after canonicalization | `errors.Is` |
| `ErrExpectedKeys` | `Get` was called with no keys | `errors.Is` |
| `*ErrNotDefined` | `Read` found no such key | `errors.As` (carries `Key`) |
| `*InvalidArgument` | Caller input was wrong | `errors.As` (carries `Msg`) |
| `*ParseError` | The file could not be decoded | `errors.As` (wraps the cause) |

```go
var nd *hensu.ErrNotDefined
if errors.As(err, &nd) {
	log.Printf("%s is not configured", nd.Key)
}
```

## Stability

Pre-1.0: this surface can still move. It is small on purpose so that when it moves, the diff is small too. Full list in [reference/api.md](../reference/api.md); what is promised and what is not in [roadmap.md](../roadmap.md).
