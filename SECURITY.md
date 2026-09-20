# Security Policy

## Supported versions

Hensu is pre-1.0. Only the latest released `v0.x` tag receives fixes.

## Reporting a vulnerability

Please report privately, not as a public issue:

- GitHub → **Security** → **Report a vulnerability** (private advisory) on https://github.com/MY-RV/hensu

Include what you did, what happened, and what you expected. A proof of concept helps; a working exploit is not required.

Expect an acknowledgement within a week. Fixes ship as a new `v0.x.y` tag, with the advisory published once users have had a chance to update.

## Scope

In scope: anything that leaks a config value through a path documented as masked, that writes the wrong file, that corrupts a config file, or that lets a tampered binary pass the self-update checksum gate.

Out of scope: everything under "Out of scope" in [docs/security.md](./docs/security.md) — in particular, that a caller with shell access can read config files directly. Hensu narrows the reasons to do so; it is not a sandbox.
