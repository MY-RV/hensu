# Repository settings

What to turn on once, after the repo is public. None of this is enforced by code, so it is written down instead.

## Branch protection on `main`

| Setting | Value | Why |
|---------|-------|-----|
| Require a pull request before merging | on | The CI gate runs on the PR, not after the fact |
| Require status checks to pass | `test`, `release-config` | Both jobs in `ci.yml`; a red gate must block the merge |
| Require branches to be up to date | on | Two green PRs can still break `main` when merged blind |
| Require linear history | on | Tags point at a commit, not a merge blob |
| Allow force pushes | off | A published tag must keep meaning what it meant |
| Allow deletions | off | — |

The `release-config` check is the one people forget: it runs `goreleaser check`, so a broken release config fails on a pull request instead of at tag time, when it would take the release down with it.

## Tags

Protect the `v*` pattern so a published version cannot be moved. Re-pointing a tag that someone already downloaded breaks `go install` checksums for everyone and cannot be undone by pushing again — ship `v0.1.1` instead.

## Actions

The release workflow needs `contents: write` (declared in the workflow). Settings → Actions → General → Workflow permissions must allow that.

Add `TAP_GITHUB_TOKEN` as a repository secret — a PAT that can write to `MY-RV/homebrew-tap` and `MY-RV/scoop-bucket`. The built-in `GITHUB_TOKEN` cannot reach another repository. Without the secret the release still publishes; it just skips the cask and the manifest, leaving package users on the previous version.

## Security

Enable private vulnerability reporting: Settings → Code security → Private vulnerability reporting. [SECURITY.md](../../SECURITY.md) points people at it, so it needs to exist before someone follows the instructions.

## Issues

Both templates live in `.github/ISSUE_TEMPLATE/`. The bug one asks people not to paste real values, which matters more here than in most projects.
