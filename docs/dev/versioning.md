# Versioning

Module: `github.com/my-rv/hensu`

## SemVer

- **`v0.x`**: public API may change without a major bump. Prefer pinning to a commit or minor version in production consumers.
- **`v1.0.0` and later**: follow [Go module semantic import versioning](https://go.dev/doc/modules/version-numbers). Breaking changes require a new major module path (`/v2`, …).

## Stability today (v0.1)

| Surface | Stability |
|---------|-----------|
| CLI verbs / exit codes / file FORMAT contract | Intentional; changes require contract update |
| `Store` method set + `context.Context` first arg | Evolving in v0 |
| `FileSystem` / `CodecRegistry` | Evolving in v0 |
| Typed errors (`ErrNotDefined`, `InvalidArgument`, …) | Evolving in v0 |

See [CHANGELOG.md](../../CHANGELOG.md).
