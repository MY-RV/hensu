# Architecture — Hensu

**Public facade** (`github.com/my-rv/hensu` via `pkg.go`) + **store** (`internal/store`) + thin **CLI** (`cmd/hensu` → `internal/cli`).

## Package map

| Layer | Path | Responsibility |
|-------|------|----------------|
| Facade | `pkg.go` | Stable `import "github.com/my-rv/hensu"` surface; `Version` (ldflags) |
| Store | `internal/store` | Document parse, codecs adapter, `Store`, FS, keys/reveal |
| Codec | `internal/codec` | DOTENV / SHEXPORT / YAML / JSON body codecs |
| Atomic / lock | `internal/atomicfile`, `internal/filelock` | Safe writes + cross-process lock |
| CLI | `internal/cli`, `internal/cli/command` | Flags → facade `Store`; reveal modes (`command/reveal.go`), `exec` (`command/exec.go`) |
| Update | `internal/update` | Release lookup + checksum-verified self-update |
| Entrypoint | `cmd/hensu` | `os.Exit` only |
| Docs | `docs/` | Contract, security, versioning |

La raíz tiene exactamente dos archivos Go, como godo: `pkg.go` (facade) y `pkg_test.go` (los `Example*` públicos + un smoke test del facade). Todo otro test vive al lado del código que ejercita — los del dominio en `internal/store`.

CLI imports the **facade**, not `internal/store`.
