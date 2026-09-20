# Standards — Hensu

Short rules. If it is not here, prefer the smallest change that matches existing code.

## Language / module

- Go, module `github.com/my-rv/hensu`.
- `cmd/hensu/main.go` is entrypoint only — no business logic.
- Public import is `github.com/my-rv/hensu` (`pkg.go` facade).
- Domain lives in `internal/store`; codecs/locks under `internal/`.
- Do not import `internal/store` from outside this module (CLI uses the facade).

## Behavior

- Contract in `docs/contract.md` wins over implementation guesses.
- Canonical keys: `:` ≡ `__`, uppercase.
- Missing store file → empty map (DOTENV), not an error on load.
- Typed errors (`ErrNotDefined`, `InvalidArgument`, `ParseError`) stay stable enough for CLI exit mapping.

## Tests

- La raíz tiene sólo `pkg.go` + `pkg_test.go` (`package hensu_test`): los `Example*` de la API pública y un smoke test del facade. Los tests de dominio van en `internal/store` como `package store_test`.
- Table or focused cases for edges (missing keys, empty set JSON, format headers).
- Do not only test the happy path.
- `godo ci` / `go test -race ./...` should stay green. La puerta vive en `godo.yaml`; no hay Makefile.

## Docs

- Design docs live in `docs/`.
- README stays short; link out.
- No feature in code without a contract line if it changes CLI or file semantics.

## Style

- No new dependency without a clear need.
- Avoid pattern theater: codec registry is fine; extra layers are not.
