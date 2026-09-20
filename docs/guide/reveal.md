# Reveal

How much of a value Hensu is allowed to print. One axis, one flag, three settings.

| Mode | A defined value prints as |
|------|---------------------------|
| `peek` | `abcd***wxyz` when longer than 16 runes; `***` when shorter |
| `mask` | `***` |
| `trust` | raw |

`peek` is the default, and it is also the zero value of the type in the library, so code that sets nothing gets the careful behavior rather than the loud one.

## What each mode actually does

Take a file with four kinds of value in it:

```
PORT=8080
EMPTY=
TOKEN=sk-live-0123456789abcdef
NOTE=a short note
```

| Key | `peek` (default) | `mask` | `trust` |
|-----|------------------|--------|---------|
| `PORT` | `"***"` | `"***"` | `"8080"` |
| `EMPTY` | `""` | `""` | `""` |
| `TOKEN` | `"sk-l***cdef"` | `"***"` | `"sk-live-0123456789abcdef"` |
| `NOTE` | `"***"` | `"***"` | `"a short note"` |
| `MISSING` | `null`, `defined: false` | same | same |

Two things to read off that table. An empty value stays empty in every mode — hiding `""` behind `***` would lie about the file. And a missing key is never confused with a hidden one: `defined` tells them apart without showing anything.

## Why peek and not mask

`mask` tells you a value exists. `peek` also tells you *which* value it is.

Four runes at each end is enough to confirm "yes, that is the key I meant to put there" without transcribing the secret — the difference between `sk-l***cdef` and `sk-l***9f2a` is the difference between a five-minute debug and an hour. For anything 16 runes or shorter, four runes at each end would be most of the value, so peek shows nothing at all.

That cut is the same one the tool Hensu came from used, and it is the reason `peek` and not `mask` is the default: it is the strictest mode that is still useful for the question people actually ask.

## Setting it

```bash
hensu -r trust get PORT        # this call only
hensu --reveal=mask get PORT   # same thing, long form
export HENSU_REVEAL=mask       # this shell, this session
```

Precedence is short: the flag beats the variable, and the variable beats the default.

```
-r / --reveal   >   HENSU_REVEAL   >   peek
```

A `HENSU_REVEAL` that does not name a mode is an error, not a fallback:

```console
$ HENSU_REVEAL=loud hensu get PORT
hensu: HENSU_REVEAL="loud" is not a reveal mode (want peek, mask, or trust)
$ echo $?
2
```

Falling back to `peek` would be defensible. Falling back to raw would print exactly what the variable exists to hide. Erroring is the only reading with no bad case.

## Which commands care

| Command | Mode applies? |
|---------|---------------|
| `get` | Yes — this is the one it is for |
| `keys` | No — it prints names, never values |
| `exec` | No — it hands values to a process and prints none |
| `read` | Refuses unless the mode is `trust` |
| `set` | No — it writes, it does not print |

`read` is the interesting one. Masking it would be pointless: its whole job is to put exact bytes on stdout so a script can capture them. So instead of printing something useless it stops and says what to do:

```console
$ hensu read TOKEN
hensu: read prints the raw value; pass -r trust, or hand the value to a process with `hensu exec -- CMD`
$ echo $?
2
```

## In a session someone else drives

A CI job, a shared shell, an agent harness — anywhere the person typing is not the person who will read the output:

```bash
export HENSU_REVEAL=mask
```

From there, `get` masks everything and `read` is closed, without anyone having to remember a flag. The program that genuinely needs values still works, because `exec` was never gated:

```bash
hensu exec -- ./server
```

## It is not a floor

The mode is one setting, not a lattice. The flag beats the variable in both directions, including back down to `trust`. That is deliberate, and it is worth being clear about why: Hensu is not trying to trap someone who already has a shell. It is making sure nobody ends up with a raw secret in their scrollback without having typed the word `trust` first. A cage would be a lie; a deliberate act you can grep for is the honest version.

See [security.md](../security.md) for the threat model, and [agents.md](./agents.md) for the harness wiring.
