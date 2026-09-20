# Set

`set` is the only command that writes. It takes one key, or many at once, and leaves the rest of the file exactly as it found it.

```bash
hensu set PORT 8080
hensu set GREETING hello world              # spaces are fine, no quoting needed
hensu set INTERNAL:SESSIONS:SECRET s3cr3t   # ':' is the usual key sugar
```

Everything after the key is the value, joined by single spaces. That is why `hello world` works without quotes — but quote anyway when the shell would otherwise eat something, like `*` or `$HOME`.

## Many keys at once

```bash
hensu set --json '{"PORT": "8080", "DEBUG": "true"}'
```

Or from stdin, which is what you want when the values come from somewhere else and should not appear in your shell history or in `ps`:

```bash
some-command --emit-json | hensu set --json -
```

JSON types are written as their literal text — `8080`, `true`, `null` becomes empty — because the canonical map is strings all the way down. An empty object is an error rather than a silent no-op.

## What survives a write

| File | What `set` preserves |
|------|----------------------|
| `DOTENV` | Comments, blank lines, key order. Only touched lines change |
| `SHEXPORT` | Same, and keeps writing `export KEY=value` |
| `YAML` | Structure and nesting; comments best-effort |
| `JSON` | Structure, reindented — JSON has nowhere to keep comments |

A file like this:

```
# database
DB_HOST=localhost      # local only
DB_PORT=5432
```

after `hensu set DB_PORT 6543` is still:

```
# database
DB_HOST=localhost      # local only
DB_PORT=6543
```

New keys are appended. A missing file is created as `DOTENV` with no header.

## Safety of the write itself

Three things happen that you do not have to ask for:

- **Atomic.** The new contents go to a temporary file in the same directory, get flushed to disk, and are renamed over the original. A crash mid-write leaves the old file intact, never a half-written one.
- **Locked.** Writers are serialized across processes through a sibling lock file, `<path>.hensu-lock`. Two `set` calls at the same moment cannot lose each other's change.
- **`0600`.** The file ends up readable only by its owner, whatever it was before.

The lock file is expected to sit next to the config. Leave it there; deleting it while something might be waiting on it is how you get the race back.

Readers are not blocked, and a reader that starts before the rename completes sees the previous contents — never a torn file.

## Nested formats

Keys are canonical, so setting through a path writes into the right place:

```console
$ cat config.yaml
# FORMAT: YAML
internal:
  sessions:
    secret: old

$ hensu -f config.yaml set INTERNAL:SESSIONS:SECRET new
$ cat config.yaml
# FORMAT: YAML
internal:
  sessions:
    secret: new
```

See [keys.md](./keys.md) for how `:` and `__` map onto nesting, and [formats.md](./formats.md) for the header.

## What set does not do

It does not print the value back, in any mode. It does not delete keys — there is no `unset` yet, and removing a key means editing the file. And it does not validate: Hensu has no idea whether `8080` is a sensible port, only that it is what you asked to store.
