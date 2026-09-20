# CLI reference

```
hensu [-f path] [-r mode] <command>
```

Context flags go **before** the command keyword; they are not scanned after it.

| Flag | |
|------|--|
| `-f, --file <path>` | Config file (default `./.env`; relative to cwd) |
| `-r, --reveal <mode>` | `peek` \| `mask` \| `trust` (default `peek`) |
| `--help` | Usage on stderr |
| `--version` | Version on stdout |
| `--update` | Replace the binary with the latest release (checksum verified) |
| `--update-check` | Report whether an update exists; download nothing |

Both spellings take `=` too: `--file=path`, `-f=path`, `--reveal=mask`, `-r=mask`.

| Command | Output |
|---------|--------|
| `get KEY [KEY...]` | JSON `{ KEY: {value, defined} }`, values per the reveal mode |
| `keys` | JSON array of defined key names, sorted |
| `read KEY` | Exact bytes of one value, no JSON, no trailing newline. Requires `-r trust` |
| `set KEY VALUE` | Writes one key (`VALUE` may contain spaces) |
| `set --json '{...}'` | Writes many keys from a JSON object |
| `set --json -` | Writes many keys from JSON on stdin |
| `exec [--] CMD [ARGS...]` | Runs `CMD` with the config in its environment |

No command → usage on stderr, exit 2. There is no bulk read.

| Exit | |
|------|--|
| `0` | ok |
| `2` | usage / invalid arguments (no command, `read` without `trust`, bad `HENSU_REVEAL`) |
| `1` | runtime error (missing key, parse, I/O) |
| `n` | `exec`: the child's exit code, verbatim |

| Environment | |
|-------------|--|
| `HENSU_REVEAL` | Default reveal mode; an unknown value is an error |
| `HENSU_RELEASES_API` | Override the releases API for `--update` (tests, mirrors) |

Normative detail: [contract.md](../contract.md).
