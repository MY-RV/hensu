# Documentation

Product docs for using and shipping **Hensu** 変数 (CLI/module: `hensu`).

| Doc | |
|-----|--|
| [Overview](./overview.md) | What Hensu is |
| [Why Hensu exists](./story.md) | Origin story and the name (EN) |
| [Por qué existe Hensu](./story.es.md) | Misma historia (ES) |
| [Getting started](./getting-started.md) | Install + first read |
| [Install](./install.md) | Binaries, Go, Homebrew, self-update |
| [Distribution](./distribution.md) | Releases and package channels |
| [Release](./release.md) | Tags + GoReleaser (for publishers) |
| [Roadmap](./roadmap.md) | Promises for v0.1 / post-v0.1 / v1.0 |
| [Contract](./contract.md) | Normative CLI + file behavior |
| [Security](./security.md) | Threat model, and what is out of scope |

## Guides

One topic per file (copy-paste examples):

| Doc | |
|-----|--|
| [Reveal](./guide/reveal.md) | `peek` / `mask` / `trust`, and why peek is the default |
| [Exec](./guide/exec.md) | Give a value to a process, not to a transcript |
| [Set](./guide/set.md) | Writing: one key, many keys, and what survives |
| [Updating](./guide/update.md) | `--update`, and why it verifies before it writes |
| [Agents](./guide/agents.md) | Wiring Hensu into an agent harness |
| [Keys](./guide/keys.md) | `:` ≡ `__`, canonical names, nesting |
| [Formats](./guide/formats.md) | `# FORMAT:` header, DOTENV / SHEXPORT / YAML / JSON |
| [Library](./guide/library.md) | Embed the `Store` in Go |

## Reference

| Doc | |
|-----|--|
| [CLI reference](./reference/cli.md) | Flag / command tables |
| [API reference](./reference/api.md) | Public Go surface |

## Contributors

Code/layout/contribution detail: [`docs/dev/`](./dev/README.md). Also root [`CONTRIBUTING.md`](../CONTRIBUTING.md) and [`SECURITY.md`](../SECURITY.md).

Contract wins over guides when they disagree.
