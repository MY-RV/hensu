# Distribution

Where a published `hensu` comes from, and which channel to trust.

## Channels

| Channel | Command |
|---------|---------|
| Go | `go install github.com/my-rv/hensu/cmd/hensu@latest` |
| Homebrew (macOS) | `brew install --cask MY-RV/tap/hensu` ([tap](https://github.com/MY-RV/homebrew-tap)) |
| Scoop (Windows) | `scoop bucket add my-rv https://github.com/MY-RV/scoop-bucket` then `scoop install hensu` |
| Binaries | [Releases](https://github.com/MY-RV/hensu/releases) — archives, bare binaries, `checksums.txt` |
| Self-update | `hensu --update` |

Homebrew ships pre-built binaries as **casks**, and casks are macOS-only. On Linux use Go or the release binary.

## Homes

| Repo | |
|------|--|
| [MY-RV/hensu](https://github.com/MY-RV/hensu) | Source, releases, advisories |
| [MY-RV/homebrew-tap](https://github.com/MY-RV/homebrew-tap) | `Casks/hensu.rb`, written by GoReleaser |
| [MY-RV/scoop-bucket](https://github.com/MY-RV/scoop-bucket) | `hensu.json`, written by GoReleaser |

The fully-qualified Homebrew name avoids clashing with unrelated taps.

## Verifying what you got

```bash
sha256sum -c checksums.txt --ignore-missing
gh attestation verify hensu_0.2.0_darwin_arm64 --repo MY-RV/hensu
```

Every release artifact carries signed build provenance. `hensu --update` does the checksum half on its own and refuses a release that publishes none — see [install.md](./install.md).

## Not channels

No `curl | sh` installer, and no package in Homebrew core or Scoop Main. Both would mean asking people to trust something this project cannot yet vouch for.
