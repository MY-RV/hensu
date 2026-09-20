# Exec

`hensu exec [--] CMD [ARGS...]` runs a command with the config already in its environment.

```bash
hensu exec -- go run .
hensu exec -- ./server --port 8080
hensu -f config/prod.env exec -- ./migrate up
```

This is the way out of the reveal modes that does not involve giving up. When a value is genuinely needed, it is almost never needed by whoever is typing — it is needed by a program. So the program gets it, and the person does not.

The `--` is optional. It matters when the command has flags that would otherwise look like Hensu's:

```bash
hensu exec -- ls -f          # -f goes to ls
hensu exec ls -f             # also ls, because flags stop at the verb
```

Context flags still have to come before `exec`, like every other command: `hensu -f x.env exec -- ...`, never `hensu exec -f x.env -- ...`.

## How the environment is built

The child gets the parent's environment with the config layered over it.

Say the shell has `PORT=3000` and `PATH=/usr/bin`, and the config file says:

```
PORT=8080
TOKEN=sk-live-0123456789abcdef
```

The child sees:

| Variable | Value | Why |
|----------|-------|-----|
| `PATH` | `/usr/bin` | inherited, config says nothing about it |
| `PORT` | `8080` | **config wins** over the inherited value |
| `TOKEN` | `sk-live-…` | from the config |

The file winning is the whole point. A stale inherited value that silently beats the file is the bug Hensu exists to remove — you edit the config, nothing changes, and you lose an afternoon.

## Keys that cannot be environment variables

An environment variable name has to match `[A-Za-z_][A-Za-z0-9_]*`. A YAML mapping can hold keys that do not — `"my key": 1`, `"a.b": 2`. Those are skipped, and the skip is announced on stderr, naming the key and never the value:

```console
$ hensu -f config.yaml exec -- ./server
hensu: exec: skipping "MY KEY" (not a valid environment variable name)
listening on :8080
```

Skipping rather than failing keeps one odd key from blocking a deploy; announcing it rather than swallowing it keeps the surprise short.

## Exit codes and signals

The child's exit code becomes Hensu's, with no `hensu:` prefix bolted on — the child already said whatever it had to say:

```console
$ hensu exec -- sh -c 'exit 3'
$ echo $?
3
```

If a signal kills the child there is no exit status to forward, so it surfaces as an ordinary Hensu error with exit 1 rather than an invented number.

stdin, stdout and stderr pass straight through, so interactive programs, pipes and progress bars all behave.

## Replacing `source`

The shell idiom this replaces:

```bash
set -a && source "$ENV_FILE" && set +a && ./server
```

`source` does not read the file, it **executes** it. A config file containing `$(rm -rf /tmp/x)` runs that command, as whoever ran the script. It is a config file being treated as a program because there was nothing better around.

```bash
hensu -f "$ENV_FILE" exec -- ./server
```

Hensu parses the file and never evaluates it. It also reads YAML and JSON, and accepts `export ` prefixes that some consumers — systemd's `EnvironmentFile=`, for one — do not.

## In a container or a unit file

```dockerfile
ENTRYPOINT ["hensu", "-f", "/etc/app/config.env", "exec", "--", "/usr/local/bin/app"]
```

```ini
ExecStart=/usr/local/bin/hensu -f /etc/app/config.env exec -- /usr/local/bin/app
```

One caveat worth knowing before you put it in a unit file: `exec` forks and waits, so the supervisor sees Hensu as the main process and your program as its child. With a `KillMode` that only signals the main PID, that is not what you want. Making `exec` replace its own process image on Unix is on the [roadmap](../roadmap.md) and is not done yet.

## What it does not do

`hensu exec -- printenv` prints everything. That is not a hole to plug — someone who wants the values and can run commands will get them, with or without Hensu.

What `exec` removes is the *need* to dump a config in order to run something with it. See [security.md](../security.md) for where that line is drawn.
