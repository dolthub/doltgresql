# go-sql-server-driver tests (Doltgres port)

These tests are ported from Dolt's
`integration-tests/go-sql-server-driver`. They exercise a real `doltgres`
server process over the Postgres wire protocol, driven either by the YAML test
definition format (`tests/*.yaml`, run via `testdef.go`) or by standalone Go
tests that use the `driver` package directly.

## Running

The tests locate the `doltgres` binary via the `DOLTGRES_BIN_PATH` environment
variable, falling back to `doltgres` on the `PATH`.

```
DOLTGRES_BIN_PATH=/path/to/doltgres go test ./...
```

## Differences from the Dolt suite

The port preserves the structure and intent of the Dolt tests, translating only
what the Postgres dialect and wire protocol, and the doltgres binary, require:

* **Server CLI.** `doltgres` has a minimal CLI (`--config`, `--data-dir`) and
  *no* `-P`/`-l`/`--max-connections` flags. The listener port can only be set
  via the config file, so the test framework (`MakeServer` in `testdef.go`)
  generates/merges a config file that injects the dynamic port, host, and a
  unique unix socket. Ported YAML expresses server settings via a config file
  (referenced with `--config`) rather than via Dolt CLI flags.

* **Data-dir model.** Doltgres serves databases out of a data-dir; each "repo"
  is a database subdirectory. The server runs from the store (data-dir) with
  `--data-dir=.`, and `with_files` are written relative to the store directory.
  Databases are initialized by briefly running a server (there is no
  `dolt init` equivalent).

* **Wire protocol.** Connections use the pgx stdlib driver. Default user is
  `postgres`, password `password`.

* **SQL dialect.** MySQL syntax is translated to Postgres (e.g.
  `AUTO_INCREMENT` -> `SERIAL`/`GENERATED`, backtick identifiers -> unquoted or
  double-quoted, `int`/`varchar(n)` are fine, `select ... from dual` ->
  `select ...`).

* **DOLT procedures.** Dolt's `CALL DOLT_*(...)` / `SELECT DOLT_*()` become
  `SELECT dolt_*(...)` in Doltgres. Result columns default to the function name
  (e.g. `dolt_add`), and the value is rendered as a Postgres array, e.g.
  `{0}`.

* **Result columns.** Postgres lowercases unquoted identifiers and uses the
  expression text / alias for computed columns; expected `columns:` and `rows:`
  are adjusted accordingly. Prefer explicit `AS` aliases.

* **Cluster / remotesapi replication.** Doltgres does not yet implement Dolt's
  cluster replication or the remotes API. Config structs for these are mirrored
  into doltgres so the config files parse, but the tests that depend on the
  feature are marked `skip:` with a reason until the feature lands.

When a ported test does not pass yet (missing feature or behavioral
difference), it is marked with `skip:` (YAML) or `t.Skip(...)` (Go) with a short
reason, per the porting instructions.
