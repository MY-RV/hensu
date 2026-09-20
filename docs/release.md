# Release

## Versioning

| What | Where |
|------|--------|
| Binary / module | `hensu.Version` (ldflags `-X github.com/my-rv/hensu.Version=…`) |
| Git tag | `vX.Y.Z` |

See [versioning.md](./dev/versioning.md).

## Local (no upload)

```bash
./scripts/release-local.sh
```

Output: `dist/hensu_<ver>_<os>_<arch>` + `SHA256SUMS`.

## Prerequisites (one time)

1. Public `MY-RV/hensu` repo with Actions enabled.
2. Secret `TAP_GITHUB_TOKEN`: a PAT with write permission to `MY-RV/homebrew-tap` and `MY-RV/scoop-bucket`. The workflow `GITHUB_TOKEN` does not work — it does not cross repos. Without that secret the release still goes out; it simply does not update the tap or bucket, and `brew` users stay on the previous version.

`ci` runs `goreleaser check` on every push, so a broken `.goreleaser.yaml` is seen long before the tag.

## GitHub

1. `godo ci` green on `main` (on the remote, not only locally)
2. CHANGELOG with today's date
3. Annotated tag `vX.Y.Z` and push the tag
4. **release** workflow: gate (`godo ci`) → GoReleaser → archives, bare binaries and `checksums.txt` → Homebrew cask (`MY-RV/homebrew-tap`, `Casks/` directory) + Scoop bucket → provenance attestation
5. Users: `brew`, `scoop`, direct download, `go install …@vX.Y.Z`, or `hensu --update`

## Verify a published artifact

```bash
gh attestation verify hensu_0.1.0_darwin_arm64 --repo MY-RV/hensu
sha256sum -c checksums.txt --ignore-missing
```

`hensu --update` performs checksum verification itself: it downloads `checksums.txt` from the release, compares SHA-256, and only then replaces the executable. A release without `checksums.txt` is rejected.

## After the release

- Test a fresh install: `go install github.com/my-rv/hensu/cmd/hensu@vX.Y.Z`
- `brew install --cask MY-RV/tap/hensu` on a clean machine (macOS)
- pkg.go.dev: `curl 'https://proxy.golang.org/github.com/my-rv/hensu/@v/vX.Y.Z.info'`
