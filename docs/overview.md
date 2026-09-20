# Overview

**Hensu** is machine-local config as a **single file with one reader**: `hensu` reads and writes one dotenv, YAML or JSON file and always hands back the same flat `KEY=value` map.

Seeing a value in full is a permission here, not the normal state. Values come back clipped unless someone asks for more.

Product name: **Hensu**. CLI, module path, and file name stay lowercase `hensu` (Go / Unix convention).

## Mental model

| Invocation | Meaning |
|------------|---------|
| `hensu keys` | What is configured — names only |
| `hensu get PORT` | That key, clipped (`peek`, the default) |
| `hensu -r trust get PORT` | That key, in full — someone asked |
| `hensu exec -- go run .` | The program gets the values; the caller does not |
| `hensu` | Usage. There is no bulk read |

## Shape (v0.1)

- File: `./.env` by default, anything else via `-f`
- Formats: `DOTENV`, `SHEXPORT`, `YAML`, `JSON`, declared by a `# FORMAT:` first line
- Reveal: `peek` (default), `mask`, `trust` — one axis, one flag
- Output: JSON, and only JSON
- Writes: atomic, locked between processes, mode `0600`, comments preserved

## Two audiences

1. **CLI users** — install a binary, point it at a config file, ask it questions
2. **Go library users** — `import "github.com/my-rv/hensu"` and drive the `Store` in-process

## Why it exists

Origin story: [EN](./story.md) · [ES](./story.es.md).

## Source of truth

Behavioral rules: [contract.md](./contract.md). What we promise to ship: [roadmap.md](./roadmap.md). What it does not defend: [security.md](./security.md). Layout (contributors): [dev/architecture.md](./dev/architecture.md).
