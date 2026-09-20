# Roadmap

What `v0.1` promises, what it does not, and what would have to be true for `v1.0`.

## Promised in v0.1

- The CLI contract in [contract.md](./contract.md): flags, commands, reveal modes, exit codes, key canonicalization, the `# FORMAT:` header.
- `peek` as the default, and `trust` as the only way to raw output.
- Atomic, locked, comment-preserving writes at mode `0600`.
- Checksum-verified `--update`, and signed provenance on every release artifact.

## Not promised in v0.1

- A stable Go API. The facade is small on purpose, but `v0.x` may still move it.
- Windows beyond "it builds and the tests pass": the quoting and locking paths there get far less real use than the Unix ones.
- YAML comment round-trip in every shape — it is best-effort, and says so.
- Anything resembling a vault: encryption at rest, rotation, remote backends.

## Candidates after v0.1

Ordered by how much evidence there is that someone needs them, not by appeal.

- `exec` replacing its own process image (`exec(2)`) on Unix instead of forking, so a supervisor sees one PID and signals land natively.
- Reading a value from stdin on `set`, so rotating a secret does not put it in shell history or in `ps`.
- Comparing a config against a declared contract (`.env.example` and friends) to answer "is anything missing?" without printing values.

None of these is committed. Each needs a caller that actually wants it first.

## v1.0

When the CLI contract has survived real use long enough that changing it would be a bigger cost than living with it, and the Go API has stopped moving. Not before.

[Versioning](./dev/versioning.md).
