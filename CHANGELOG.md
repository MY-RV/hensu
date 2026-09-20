# Changelog

## [0.2.0] — unreleased

### Security
- **`--update` verifies SHA-256 against the release `checksums.txt` before
  replacing the running binary.** It previously installed whatever bytes the
  download returned. A release without checksums, an asset missing from the
  file, or a digest mismatch now aborts without touching the executable.

### Changed
- **No bulk read.** `hensu` with no command prints usage and exits 2 instead of dumping the
  config; to see a value you name its key.
- **JSON is the only output encoding.** The `json` and `yaml` verbs and the
  `export KEY=value` dump are gone, along with `Store.DumpExport` / `DumpJSON` / `DumpYAML`.
  YAML and JSON remain first-class as *file* formats, which is where they matter.
- `update.Client.Latest` returns a `Candidate` (asset + checksums).
- Public API is a thin facade (`pkg.go`); domain lives in `internal/store` (godo-style layout).
- Canonical Go module layout (`cmd/hensu`, `internal/`, import `github.com/my-rv/hensu`).
- `Store` and `FileSystem` methods take `context.Context` (cancelable lock acquire / IO checks).
- CLI exit codes: `0` ok, `2` invalid usage, `1` runtime error.
- File locking always uses a stable sidecar `<path>.hensu-lock` (correct with atomic rename).
- GoReleaser + tag-triggered GitHub release workflow.

### Added
- **`peek` is the default reveal mode**, and the zero value of `Reveal`. A caller that
  configures nothing still cannot print a secret whole: long values come back as
  `abcd***wxyz`, short ones as `***`. Seeing a value in full is a grant — `-r trust` — and
  grants are asked for, by a human or a model alike.
- `-r` / `--reveal` with its own vocabulary (`peek` / `mask` / `trust`) replacing the old
  `get` / `def` / `spy` verbs. One axis, one spelling: the strictness lattice the three verbs
  needed is gone with them.
- `hensu exec [--] CMD [ARGS...]` — runs a command with the config in its environment
  (config wins over inherited values), forwards the child exit code, and skips keys that are
  not valid environment variable names. The reveal mode is not propagated to children.
- `read KEY` requires `-r trust`.
- `HENSU_REVEAL` sets the mode for a session; an unparseable value is an error, never a
  silent fallback to raw output.
- Cross-process `Set` locking, durable atomic writes (`Sync` + dir sync).
- Typed errors, godoc examples, fuzz tests, benchmarks, GitHub Actions CI (incl. staticcheck).
- Apache-2.0 LICENSE, docs (contract, security, versioning, architecture, install, release).
- `scripts/release-local.sh` for multi-arch artifacts without upload.
- CONTRIBUTING.md.
- CLI self-update: `--update` / `--update-check` (`internal/update`, `HENSU_RELEASES_API`).
- `godo.yaml` dogfood scripts (`test`/`vet`/`build`/`check`/`ci`/`release-local`).
- GoReleaser publishes bare binaries (for `--update`) in addition to archives.
- Homebrew cask and Scoop manifest published by GoReleaser (`MY-RV/homebrew-tap`
  under `Casks/`, `MY-RV/scoop-bucket`); skipped rather than failed when the PAT
  is absent. Casks are macOS-only — Linux installs via `go install` or the
  release binary.
- Signed build provenance attestations for every released artifact.
- `goreleaser check` in CI, so a broken release config fails on push, not at tag.
- `SECURITY.md` (private vulnerability reporting) and `docs/agents.md`.

### Notes
- Pre-1.0: APIs may still change. Treat `v0.x` as evolving.
