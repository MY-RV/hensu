# Contract — Hensu

Machine config: read and write **one** file. Format-agnostic → flat map.

## CLI

```
hensu [-f path] [-r mode] <command>
```

Context flags go **before** the keyword. They are not scanned after the command.

| Context flag | |
|--------------|--|
| `-f, --file <path>` | Config file (default `./.env`) |
| `-r, --reveal <mode>` | `peek` \| `mask` \| `trust` (default `peek`) |
| `--help` | Help |
| `--version` | Version |
| `--update` | Downloads the latest binary from GitHub Releases (verifies checksum) |
| `--update-check` | Reports whether an update is available (does not download) |

Without `--file` → `./.env` (cwd). Relative paths are resolved from cwd.

`--update` / `--update-check` are context flags (like `--version`): they do not open the config file. API override (tests / mirror): `HENSU_RELEASES_API`. Expected assets: bare binary `hensu_<ver>_<os>_<arch>` plus `checksums.txt`. `--update` verifies SHA-256 against `checksums.txt` before replacing the binary. A release without `checksums.txt`, an asset not listed there, or a digest that does not match → aborts without touching the executable.

## Commands

| Command | |
|---------|--|
| `get KEY [KEY...]` | JSON `{ KEY: {value, defined} }` — value according to the mode |
| `keys` | JSON array of defined names (sorted, without values) |
| `read KEY` | one key; exact bytes to stdout (no JSON, no extra newline). **Requires `-r trust`** |
| `set KEY VALUE` | set one key (`VALUE` may contain spaces) |
| `set --json '{...}'` | set many keys from a JSON object |
| `set --json -` | set many keys from JSON on stdin |
| `exec [--] CMD [ARGS...]` | runs `CMD` with the config in its environment |

**No command**: prints help to stderr and exits with **2**. There is no bulk read: to see a value, its key must be named.

JSON is the only output encoding. Whatever needs another form gets it downstream.

## Reveal

| Mode | Defined value |
|------|----------------|
| `peek` | runes `≤ 16` → `"***"`; otherwise → `trim[:4] + "***" + trim[-4:]` (by runes, not bytes) |
| `mask` | `"***"` if non-empty; `""` if empty |
| `trust` | raw |

**`peek` is the default**, and it is the zero value of the type. Seeing a full value is a grant, and grants are requested — whoever is reading, human or model.

Precedence: `-r` (flag) > `HENSU_REVEAL` (env) > `peek`.

`HENSU_REVEAL` with an unknown value is an error (exit 2), never a silent fallback: a fallback would print precisely what the variable exists to hide.

`read` with a mode other than `trust` → exit 2, naming `-r trust` and `hensu exec`.

## Exec

`hensu exec [--] CMD [ARGS...]` runs `CMD` with the config in its environment.

- Child env = inherited env + config on top (**the config wins**: a stale inherited value is exactly the bug Hensu removes).
- Keys that are not valid variable names (`[A-Za-z_][A-Za-z0-9_]*`) are omitted, with a stderr warning that names the key (never the value).
- The reveal mode does **not** propagate to the child: the default is already safe, and a grant given to this invocation is not a grant for its descendants.
- stdin/stdout/stderr pass through untouched; Hensu does not print values.
- The child's exit code is Hensu's exit code (without a `hensu:` prefix). If a signal kills the child there is no status to forward → normal error, exit 1.

## File

Default: `./.env`

Formats (optional header):

```
# FORMAT: DOTENV | SHEXPORT | YAML | JSON
```

Without a header → `DOTENV`.

| Format | Shape |
|--------|--------|
| `DOTENV` | key-value (`KEY=value`; optional `export` per line) |
| `SHEXPORT` | key-value with `export KEY=value` on write |
| `YAML` | nested mapping ↔ flat with `__` |
| `JSON` | object (flat or nested) ↔ flat with `__` |

Semantics:

- Format-agnostic read → flat `KEY=value` map (strings)
- `set` preserves file comments and format (DOTENV/SHEXPORT/YAML); JSON is rewritten indented
- Missing file → empty map (read); `set` creates in `DOTENV` without a header
- Values are always strings in the canonical map
- `set` serializes writers (lock per file), writes atomically and with mode `0600`

## Keys

- Canonical flat: `INTERNAL__SESSIONS__FOO`
- Lookup and load: `:` ≡ `__` (`INTERNAL:SESSIONS:FOO` ≡ `INTERNAL__SESSIONS__FOO`)
- Nested YAML/JSON: flatten/unflatten with `__`
- Canonical keys are uppercase

## Exit

| Code | Meaning |
|------|---------|
| `0` | ok |
| `2` | usage / invalid arguments (includes no command, `read` without `trust`, invalid `HENSU_REVEAL`) |
| `1` | runtime error (missing key, parse, I/O, etc.) |
| `n` | `exec`: the child process exit code, as-is |

## Limitations (honest)

- YAML comment round-trip is best-effort (`yaml.v3`); canonical values win on conflict.
- `set` coordinates writers via a sibling lock file `<path>.hensu-lock` (mode `0600`, stable inode). Required because writes use atomic rename — flock on the data file would not survive replace.
- Concurrent `set` is serialized; readers may observe the previous file contents until rename completes.
- JSON `set` rewrites the body (no comments in JSON).
- Hensu is not a sandbox: whoever has a shell can read the file directly. See [security.md](./security.md).

## Library (Go)

Module and import path: `github.com/my-rv/hensu`. The CLI lives in `cmd/hensu` — `go install github.com/my-rv/hensu/cmd/hensu@latest`.

Minimal public API:

- `CanonicalKey`, `PathSegments`, `KeyFromPath`, `NormalizeMap`
- `Format` + `FormatCodec` / `CodecRegistry`
- `ParseDocument` / `SerializeDocument`
- `Store` (all methods take `context.Context`): `Path`, `Load`, `Get`, `Read`, `Keys`, `Set`
- `Reveal` (`RevealPeek`, `RevealMask`, `RevealTrust`), `Lookup`, `ShortValueMax`
- `FileSystem` / `MemFileSystem` / `WithFileSystem` / `WithRegistry`
- Errors: `ErrEmptyKey`, `ErrExpectedKeys`, `*ErrNotDefined`, `*InvalidArgument`, `*ParseError`
- `Version` (overridable via `-ldflags -X github.com/my-rv/hensu.Version=…`)

Also: [security.md](./security.md), [guide/agents.md](./guide/agents.md), [dev/versioning.md](./dev/versioning.md).
