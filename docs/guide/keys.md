# Keys

Hensu keeps one canonical spelling for every key, so the same value is reachable whether the file is flat or nested, and whichever separator you feel like typing.

## Canonical form

Uppercase, segments joined by `__`:

```
INTERNAL__SESSIONS__SECRET
```

That is what `keys` prints, what a child process sees under `exec`, and what ends up on disk in a flat format.

## `:` is sugar for `__`

These are the same key:

```bash
hensu get INTERNAL:SESSIONS:SECRET
hensu get INTERNAL__SESSIONS__SECRET
hensu get internal:sessions:secret
```

Colons are easier to type and read; `__` is what an environment variable can actually be called. Lowercase is folded up, surrounding whitespace is trimmed, and repeated separators collapse — `A____B` and `A::B` both land on `A__B`.

The sugar applies everywhere a key is accepted: `get`, `read`, `set`, and the keys inside `set --json`.

## Nesting

`__` is the boundary between levels in YAML and JSON. This file:

```yaml
# FORMAT: YAML
internal:
  sessions:
    secret: s3cr3t
    ttl: 3600
```

is exactly this map:

```
INTERNAL__SESSIONS__SECRET=s3cr3t
INTERNAL__SESSIONS__TTL=3600
```

and writing through a path writes into the tree, not a flat key beside it:

```bash
hensu -f config.yaml set INTERNAL:SESSIONS:TTL 7200
```

Going the other way, a flat `A__B=1` in a dotenv file is just a key with underscores in it — flattening is a property of the nested formats, not something Hensu imposes on flat ones.

## Keys that cannot be environment variables

Canonicalization uppercases and swaps separators; it does not rename. A YAML mapping can hold `"my key"` or `"a.b"`, and those stay as they are — readable through `get`, listed by `keys`.

They only become a problem at the one place an actual environment variable is required: `exec` skips them, with a warning naming the key. See [exec.md](./exec.md).

## In the library

```go
hensu.CanonicalKey("internal:sessions:foo")     // "INTERNAL__SESSIONS__FOO"
hensu.PathSegments("INTERNAL__SESSIONS__FOO")   // []string{"INTERNAL","SESSIONS","FOO"}
hensu.KeyFromPath([]string{"internal", "foo"})  // "INTERNAL__FOO"
hensu.NormalizeMap(map[string]string{"a:b": "1"}) // {"A__B": "1"}
```

An empty key — or one that is empty after trimming — is `ErrEmptyKey` rather than a silently dropped write.

## Pitfalls

Two keys that differ only in case are the same key: `Port` and `PORT` collide, and the last one parsed wins. If a file has both, that is a bug in the file, and Hensu will not invent a way to keep them apart.
