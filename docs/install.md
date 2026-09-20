# Install — Hensu

## With Go

```bash
go install github.com/my-rv/hensu/cmd/hensu@latest
```

Local build with injected version:

```bash
go build -ldflags "-X github.com/my-rv/hensu.Version=0.1.0" -o hensu ./cmd/hensu
hensu --version
```

## Homebrew (macOS)

```bash
brew install --cask MY-RV/tap/hensu
```

Homebrew distributes pre-compiled binaries as a **cask**, and casks are macOS-only. On Linux: `go install`, the release binary, or Scoop does not apply.

## Scoop (Windows)

```powershell
scoop bucket add my-rv https://github.com/MY-RV/scoop-bucket
scoop install hensu
```

Tap and bucket are updated by GoReleaser on every release; available since `v0.1.0`.

## Without Go (GitHub Releases)

1. Open [Releases](https://github.com/my-rv/hensu/releases) and download the asset for your OS/arch:
   - archive: `hensu_<ver>_darwin_arm64.tar.gz` (or `.zip` on Windows), **or**
   - bare binary: `hensu_<ver>_darwin_arm64` (same naming as `scripts/release-local.sh`)
   - checksums in the release
2. Extract if needed, `chmod +x hensu`, move to `$PATH`.
3. Verify: `hensu --version`.
4. Verify provenance (optional): `gh attestation verify hensu_<ver>_<os>_<arch> --repo MY-RV/hensu`
5. Update later: `hensu --update` / `hensu --update-check`  
   (requires public Releases; override: `HENSU_RELEASES_API`). `--update` compares SHA-256 against the release `checksums.txt` before replacing the binary.

## Local artifacts (no upload)

```bash
./scripts/release-local.sh
# o, si tenés godo en PATH:
godo release-local
```

Generates `dist/` + `SHA256SUMS` + bare binaries with ldflags. Does not upload to GitHub.
