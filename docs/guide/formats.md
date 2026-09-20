# Formats

One file, four spellings. Hensu reads all of them into the same flat `KEY=value` map, so nothing downstream has to branch on which one you picked.

## Declaring it

First line of the file, optional:

```
# FORMAT: DOTENV
```

No header means `DOTENV`. The header is also how a writer knows what to write back, which is why `set` never silently converts your YAML into a dotenv.

## The four

**DOTENV** — the default, and what `.env` files look like everywhere:

```
# comments survive
PORT=8080
GREETING="hello world"
export LEGACY=1
```

A leading `export ` is tolerated on read, because plenty of files have it. Values may be bare, single-quoted or double-quoted; escapes inside double quotes are unescaped.

**SHEXPORT** — identical on read, but writes every line with `export`, for files meant to be sourced by a shell:

```
# FORMAT: SHEXPORT
export PORT=8080
```

**YAML** — nested, flattened with `__`:

```yaml
# FORMAT: YAML
internal:
  sessions:
    secret: s3cr3t
port: 8080
```

reads as:

```
INTERNAL__SESSIONS__SECRET=s3cr3t
PORT=8080
```

**JSON** — flat or nested, same flattening:

```json
{ "internal": { "sessions": { "secret": "s3cr3t" } }, "port": 8080 }
```

## Everything is a string

`port: 8080` in YAML reads as `"8080"`. A boolean reads as `"true"`. The canonical map holds strings because that is what an environment variable is, and inventing types on the way out would mean guessing on the way back in.

## A missing file is empty, not an error

```console
$ hensu -f /nowhere/.env keys
[]
```

That is deliberate. It lets a script call `hensu exec` unconditionally when the config file is optional, instead of testing for the file first and branching. `read` on a key that is not there is still an error — the file being absent is normal, the key you asked for being absent is not.

## Round trips

| You do | What comes back |
|--------|-----------------|
| `set` on DOTENV/SHEXPORT | Comments, order and blank lines intact |
| `set` on YAML | Structure and nesting intact; comments best-effort |
| `set` on JSON | Same structure, reindented |
| Read YAML, write YAML | Nesting preserved, not flattened onto disk |

The YAML caveat is real and worth stating plainly: comment round-trip goes through `yaml.v3` and is best-effort. When the library and the canonical value disagree, the value wins. If comments in a YAML config are load-bearing for you, keep the config in DOTENV.

See [set.md](./set.md) for what the write itself guarantees, and [keys.md](./keys.md) for the flattening rules.
