# Hensu

**Hensu** 変数 (*hensū*, variable — 変 change, 数 value). Machine-local config: one file, read and written as a flat `KEY=value` map, whether it is dotenv, YAML or JSON.

A config file holds secrets, and everything reads it now — you, a script, a model. So seeing a value in full is a **permission** here, not the normal state: values come back clipped unless someone types `-r trust`, and there is no bulk read at all. Full story: [EN](./docs/story.md) · [ES](./docs/story.es.md).

```bash
hensu keys                # what is configured
hensu get PORT            # {"PORT": {"value": "***", "defined": true}}
hensu exec -- go run .    # the process gets the values; you do not
```

## Install

```bash
go install github.com/my-rv/hensu/cmd/hensu@latest
```

```bash
brew install --cask MY-RV/tap/hensu
```

Binaries (no Go): [docs/install.md](./docs/install.md). Scoop: [docs/distribution.md](./docs/distribution.md).

## Library

```go
import "github.com/my-rv/hensu"

s := hensu.NewStore(".env")
entries, err := s.Get(ctx, []string{"PORT"}, hensu.RevealPeek)
```

## Docs

[Getting started](./docs/getting-started.md) · [Guides](./docs/README.md) · [Contract](./docs/contract.md) · [Security](./docs/security.md) · [Roadmap](./docs/roadmap.md)

## Develop

```bash
go install github.com/my-rv/godo/cmd/godo@latest
godo ci
```

[Contributing](./CONTRIBUTING.md) · [Security](./SECURITY.md) · [MIT](./LICENSE)
