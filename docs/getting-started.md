# Getting started

## 1. Install

With Go:

```bash
go install github.com/my-rv/hensu/cmd/hensu@latest
```

macOS, without Go:

```bash
brew install --cask MY-RV/tap/hensu
```

Other channels and bare binaries: [install.md](./install.md).

```console
$ hensu --version
0.2.0
```

## 2. Point it at a file

Default is `./.env` in the current directory. Anything else goes in `-f`:

```bash
hensu -f source/service/presentation/service/.env keys
```

Given this `.env`:

```
PORT=8080
INTERNAL__SESSIONS__SECRET=s3cret-value-0123456789
```

## 3. Ask what is configured

```console
$ hensu keys
[ "INTERNAL__SESSIONS__SECRET", "PORT" ]

$ hensu get PORT INTERNAL:SESSIONS:SECRET MISSING
{
  "INTERNAL__SESSIONS__SECRET": { "value": "s3cr***6789", "defined": true },
  "MISSING":                    { "value": null,          "defined": false },
  "PORT":                       { "value": "***",         "defined": true }
}
```

Values come back clipped: that is `peek`, the default. Note `hensu` on its own prints usage, not the config — there is no bulk read.

## 4. Run something with the config

```console
$ hensu exec -- go run .
listening on :8080
```

The process gets the real values. You never saw them.

## 5. When you really need the value

```console
$ hensu -r trust read PORT
8080
```

`read` refuses without `-r trust`, because printing bytes verbatim is its whole job.

## 6. Write

```bash
hensu set PORT 9090
echo '{"A":"1","B":"2"}' | hensu set --json -
```

Comments and layout survive; the write is atomic and mode `0600`.

## Next

[Reveal](./guide/reveal.md) · [Exec](./guide/exec.md) · [Agents](./guide/agents.md) · [Contract](./contract.md)
