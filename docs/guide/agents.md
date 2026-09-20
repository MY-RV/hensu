# Agents

An agent asked whether `DATABASE_URL` is configured has one obvious move: open the file. The question was small — is it set? — and the answer arrives whole, into a transcript, a session log, and a model's context. Three places it does not leave.

Hensu splits the useful question from the dangerous answer, and it does it by default, so nothing depends on the agent choosing well.

## Nothing to configure

`peek` is the default mode. A caller that sets no flag and no variable already cannot print a secret whole:

```console
$ hensu get DATABASE_URL PORT MISSING
{
  "DATABASE_URL": { "value": "post***/app", "defined": true  },
  "PORT":         { "value": "***",         "defined": true  },
  "MISSING":      { "value": null,          "defined": false }
}
```

That is enough to answer the three questions that actually come up — is it set, is it empty, is it the value I expect — without transcribing any of them.

## No bulk read

`hensu` with no command prints usage, not the config. There is no verb that dumps everything. To see a value you name its key, and even then it arrives clipped.

That is the difference between a tool an agent can drain in one call and one it has to ask specific questions of.

| Question | Command | What comes back |
|----------|---------|-----------------|
| What keys exist? | `hensu keys` | names only |
| Is this one set? | `hensu get KEY` | `***` or a tip, plus `defined` |
| Does the app run with it? | `hensu exec -- go run .` | the program's own output |
| I need the value itself | `hensu -r trust read KEY` | the value, because someone asked |

## exec is the honest way out

When a value is genuinely required, it is almost never required by the agent — it is required by a process the agent is starting:

```bash
hensu exec -- go run .
hensu -f source/service/presentation/service/.env exec -- go test ./...
```

The program gets the real values, the agent sees what the program prints. Full rules in [exec.md](./exec.md).

## Wiring a harness

The default is already safe, so this is optional. Set it when you want the stricter mode everywhere, with no flag on any call:

```bash
export HENSU_REVEAL=mask
```

Then in the repo, tell the agent what to reach for. Something like this, in `AGENTS.md` or `CLAUDE.md`:

```markdown
## Config

Do not read `.env` or any config file directly.

- `hensu keys` — which keys exist
- `hensu get KEY` — whether a key is set (the value comes back clipped)
- `hensu exec -- CMD` — run something with the config in its environment

If you need a value in full, ask for it with `-r trust` and say why.
```

## What an agent can and cannot do

| | Without asking | Asking explicitly |
|---|---|---|
| List key names | yes | — |
| See a value clipped | yes | — |
| See a value in full | no | `-r trust` |
| Read exact bytes | no | `-r trust read` |
| Run a program with the values | yes | — |
| Dump the whole file | no | there is no such command |

"Asking explicitly" means the word is in the command line — in the history, in the session log, greppable tomorrow.

## This is not a sandbox

An agent with shell access can run `cat .env` and see everything. So can `hensu exec -- printenv`. No CLI prevents that, and pretending otherwise would be the most dangerous thing in this document.

What changes is narrower and still worth having:

1. **The convenient path stopped being the leaky one.** The obvious command now returns something safe, so the common case no longer costs you a secret.
2. **A leak became an act.** Someone typed `trust`, or typed `cat`. Both are visible after the fact.
3. **It does not rely on anyone remembering.** The default is the careful one, so a forgetful caller — human or model — is not a leak.

Threat model, including what is out of scope: [security.md](../security.md).
