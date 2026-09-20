# Security

Hensu reads and writes **machine-local config files** that often contain secrets (API keys, tokens, DB URLs). It is not a vault.

## Threat model (in scope)

| Asset | Risk | Mitigation in Hensu |
|-------|------|---------------------|
| Config file contents | Local user/process read | Files written with mode `0600` |
| Config values | Leaking into logs, transcripts, model context | `peek` by default; no bulk read; `exec` |
| Concurrent writers | Lost updates / torn files | Sidecar flock (`*.hensu-lock`) + atomic rename |
| Path confusion | Writing the wrong file | Explicit `--file`; default is cwd `./.env` only |
| Self-update channel | Tampered or wrong binary replacing Hensu | SHA-256 checked against the release `checksums.txt`; no checksums → refuse |
| Lock sidecar | Extra file next to config | Expected; mode `0600`; do not delete while processes may wait |

## The default is not to show

Seeing a full value is a **grant**, not the normal state:

| Mode | A defined value is printed |
|------|------------------------------|
| `peek` (default) | `abcd***wxyz` if it exceeds `ShortValueMax` (16 runes); `***` otherwise |
| `mask` | `***` |
| `trust` | raw |

- `peek` is the zero value of the `Reveal` type: a caller that configures nothing does not leak either. There is nothing to remember.
- `read`, whose only job is to print exact bytes, requires `-r trust`.
- `hensu` with no command returns help, not the config: there is no bulk read.
- Precedence `-r` > `HENSU_REVEAL` > `peek`. An invalid `HENSU_REVEAL` is an error: a fallback to raw would print exactly what the variable exists to hide.
- `exec` does not propagate the mode to its children. The default is already safe, and a grant given to an invocation is not a grant for its descendants.

What `peek` does intentionally reveal: the four runes at each end of a long value. It is enough to confirm identity ("yes, it is the key I expected") without transcribing the secret, and it is the same trimming Hensu's ancestor already used. For a short value it shows nothing, because in a short value four runes are almost everything.

## Out of scope

- Encryption at rest
- Secret rotation / remote sync (Doppler, 1Password, cloud KMS)
- Multi-user ACL beyond OS file permissions
- Protecting against a privileged local attacker
- **Containing a caller that already has a shell.** An agent, a script or a person who can run `cat .env` or `hensu exec -- printenv` sees everything. Hensu removes the *reason* to do that and makes doing it explicit; it cannot and does not prevent it.

## Operator expectations

1. Keep config files off shared/world-readable volumes.
2. Do not commit real `.env` files; use examples without secrets.
3. `HENSU_REVEAL=mask` in agent harnesses, CI jobs, and shared shells, if you want stricter than the default.
4. Reach for `hensu exec` instead of dumping config into a command line — an argument is visible in `ps` and in shell history; an environment variable is at least limited to the process and its children.
5. Treat output from `-r trust` and `read` as sensitive: it is the only output that carries full values.

## Reporting a vulnerability

See [SECURITY.md](../SECURITY.md) at the repository root.
