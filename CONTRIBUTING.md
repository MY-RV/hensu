# Contributing

## Before a change

1. Read [docs/contract.md](./docs/contract.md) — behavior changes need a contract update in the same PR.
2. Keep the cut: thin facade + `internal/store`; CLI talks to the facade only.

## Loop

```bash
godo ci
# or
go test -race ./...
go build -o hensu ./cmd/hensu
./hensu --version
```

## PR shape

- One concern per PR.
- Tests for the behavior you touch (happy path **and** the edge that failed or could fail).
- No drive-by refactors.

## Code

See [docs/standards.md](./docs/dev/standards.md).
