# Updating

```bash
hensu --update-check    # is there a newer release?
hensu --update          # replace this binary with it
```

Both are context flags, like `--version`: they never open your config file.

```console
$ hensu --update-check
current: 0.1.0
latest:  0.1.1
update available

$ hensu --update
current: 0.1.0
latest:  0.1.1
downloading https://github.com/MY-RV/hensu/releases/download/v0.1.1/hensu_0.1.1_darwin_arm64 → /usr/local/bin/hensu
updated (hensu_0.1.1_darwin_arm64 verified against checksums.txt)
```

## What "verified" means

A self-updater replaces the binary that handles your secrets. If it installed whatever bytes came back from the network, it would be a worse hole than anything Hensu closes. So it does not.

Before anything is written, `--update` fetches `checksums.txt` from the same release, finds the line for the asset it is about to install, and compares the SHA-256 of what it downloaded. Three ways that ends badly, and all three stop the update with your existing binary untouched:

| Situation | Result |
|-----------|--------|
| The release publishes no `checksums.txt` | Refused — nothing is verified, so nothing is installed |
| The asset is not listed in it | Refused — the file exists but says nothing about this download |
| The digest does not match | Refused, and the mismatch is printed |

The download is also capped, so a hostile or broken server cannot fill your disk while you wait.

## Where it looks

GitHub Releases for `MY-RV/hensu`. Point it somewhere else — a mirror, or a test server — with:

```bash
HENSU_RELEASES_API=https://example.internal/api/repos/me/hensu hensu --update-check
```

It prefers the bare binary asset (`hensu_<version>_<os>_<arch>`) over the archives, because a self-update has nowhere to unpack a tarball to.

## When not to use it

If you installed through Homebrew or Scoop, update through them instead:

```bash
brew upgrade --cask hensu
scoop update hensu
```

`--update` would overwrite the binary those tools manage, and then their bookkeeping and reality would disagree. Same for a binary installed by `go install`: re-run `go install` instead.

`--update` is for the case it was built for — you downloaded a binary, put it on your `$PATH`, and there is no package manager in the loop.

## Verifying by hand

Every release also carries signed build provenance, which is stronger than a checksum because it says *where the bytes were built*, not just that they did not change in transit:

```bash
gh attestation verify hensu_0.1.1_darwin_arm64 --repo MY-RV/hensu
sha256sum -c checksums.txt --ignore-missing
```

See [distribution.md](../distribution.md) for the channels and [release.md](../release.md) for how the artifacts are produced.
