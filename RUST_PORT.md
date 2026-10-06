# Doltgres Rust port — scope and plan

Working document for the experimental Rust rewrite. It is removed when the port is finished, since the
end state contains only Rust code.

- Worktree branch: `daylon/rust-port`
- Pinned Go baseline: `96c7b6ba630cd0862c9e0a74714b5b12c04626f7` (local tag `rust-port-base`)
- Toolchain: stable Rust via Homebrew `rustup`

## Goals

1. Client I/O identical to the Go version: wire messages, rows, errors and SQLSTATEs, notices, tags.
2. Storage byte-identical and bidirectional: the same operations produce the same chunks and hashes, and
   Go and Rust read, write, push, and pull each other's repositories and remotes, including every legacy
   format the Go version can still read.
3. Performance equal to or better than Go on the sysbench workloads in `scripts/quick_sysbench.sh`.
4. The ported script tests (`testing/go`, extensions) assert real Postgres 15 output, and the Rust server
   must pass all of them, including assertions the Go server fails. The other suites (regression replay,
   sqllogictest, dump imports, enginetest Dolt sets) are gated on Go parity: everything the Go version
   passes must pass, and the rest are stretch goals.

## Constraints

- Every crate we write carries `#![forbid(unsafe_code)]`, enforced by the workspace `unsafe_code = "forbid"`
  lint. Any genuinely required `unsafe` is raised with the owner first.
- Third-party crates are allowed, including C-linked ones (e.g. ICU), when we understand their impact.
  Popular, long-lived staples are preferred over young pure-Rust alternatives.
- Internals need not resemble Doltgres, go-mysql-server, Vitess, or Dolt.
- The Go code stays in the worktree as a reference and test oracle, and is deleted only once the port is
  completely finished.
- Targets macOS, Linux, and Windows, matching the Go CI.

## Tests

All Go tests become Rust tests. Data fixtures stay as data. Bats suites and other-language client tests
stay as they are, since they drive the `doltgres` binary.

### Included (logic-confirming)

| Suite | Size | Go baseline |
|---|---|---|
| `testing/go` script, wire, flow, transaction, replication tests | ~118 files, ~12.5k assertions | 244 `Skip: true` |
| `testing/go/extensions` (pgvector, uuid-ossp) | 6 files, ~900 assertions | 52 skips |
| Postgres regression replay | 205 files, 42,090 statements | 20,608 pass (48.96%), 10 files fully pass |
| Dump imports (`TestImportingDumps`, unskipped) | 103 dumps | 45 pass, 46 fail, 12 hang (3-minute timeout) |
| sqllogictest | 5,675,180 tests | 99.317% ok (README, v1.0.0) |
| Enginetests: Dolt version-control script sets only | merge, conflicts, revert, reset, branch, tag, stash, commit, rm, history, diff functions, ... | currently passing queries |
| bats (`testing/bats`) | 87 tests | 4 skipped |
| Client-language tests | 30 clients | requires Docker |
| Version compatibility bats | 76 tests | v0.56.x |
| go-sql-server-driver scenarios | 22 Go tests, 99 YAML cases | cluster, remotesapi, TLS, GC, large values |

EXPLAIN assertions check internal behavior (index use, join strategy, node presence). They become plan
property assertions that the Rust server reports directly, rather than matching go-mysql-server's plan
text.

### Excluded

- `testing/generation` (syntax-only).
- Go unit tests of internal APIs in `server/`, `core/`, `utils/`, `postgres/`, `servercfg/`, and
  `testing/dataloader`. Their observable requirements (legacy formats, serialization) are covered by
  compatibility testing.
- Enginetest sets with MySQL semantics.

### Operational scope (all in)

Dolt cluster replication, failover, remotesapi, JWT, metrics; remote backends (file, AWS, GCS, OCI,
HTTP); Postgres logical replication; auto-GC with identical logs and file layout; startup integrity
check; the `admin` corruption report and repair tool.

### Expectations come from Postgres

The Go tests partly encode go-mysql-server and Dolt behavior (error text especially), and their
normalization hides type mistakes such as int4 versus int8. The Rust tests instead encode what Postgres
15 actually sends:

- Every Go script is replayed against a real Postgres 15 with fresh state per script, and its output is
  recorded: exact value text, column names and type OIDs, command tags, notices, and errors.
- Errors assert severity, SQLSTATE, message, detail, hint, position, and the schema, table, column, and
  constraint name fields. A code plus message substring is the fallback only where an exact match is
  prohibitive. Postgres' File, Line, and Routine fields are ignored.
- Rows are compared in order only when the query has ORDER BY; otherwise they are compared as a multiset.
- Assertions Postgres cannot run because they depend on Dolt features (dolt_* functions, tables, and
  procedures, AS OF, branches, merges, conflicts) take the Go server's output, reviewed against the Go
  test's intent.
- Implementation-defined values use explicit matchers (any OID, any PID, any timestamp) that still check
  type and shape. Deliberate Doltgres differences (version strings, Doltgres-only catalog entries) keep
  the Go server's value and are documented as Doltgres-specific.

Where a Go test's input does not work in Postgres, the input is changed minimally and the generated test says so
in a `// Changed from the Go test:` comment:

- Setup that Postgres rejects only because Doltgres is lenient is rewritten to valid Postgres (for example
  `gen_random_uuid()` without `public.`, `jsonb` where `json` cannot be indexed, a quoted mixed-case schema).
- Setup that a test expects to be rejected (unknown types, non-boolean trigger conditions, schema-qualified type
  aliases) becomes assertions, since Postgres rejects it at creation instead of at use.
- PL/pgSQL loops over an uninitialized variable, which never end in Postgres, initialize it, and a separate
  assertion checks that an uninitialized variable stays NULL.
- Features beyond Postgres 15 (JSON_TABLE) take their expectations from Postgres 17.

Tests whose Go version is custom code over a single pgx connection (the binding tests, application settings) are
ported by recording the exact bytes pgx sent and replaying them as wire conversations.

### Harness

The Rust harness spawns a fresh `doltgres` binary per script (path from an environment variable) on a
free port with a temporary data directory, matching the Go suite's server-per-script isolation. It sends
exactly the protocol messages pgx sends (statement cache names, DescribeExec versus CacheStatement modes,
result format codes, parameter encodings), proven by diffing frontend message transcripts recorded from
the Go suite and the Rust suite. It can also target a real Postgres, which is how expectations are
captured and re-verified.

## Phases

0. Cargo workspace with the unsafe lint; wire codec (done, verified against pgproto3); pgx-equivalent
   client and harness; Postgres expectation capture; port of every Go test.
1. Storage read path: hashes, NBS table files, journal, archives, manifest, flatbuffers messages, prolly
   trees, commit graph. Verified by dumping Go-written repositories identically.
2. Storage write path, byte-identical: same chunks and hashes as Go for the same operations.
3. Wire protocol, parser, catalog, and a minimal engine, enough for the smoke tests.
4. Breadth: types, functions, operators, DDL, DML, pg_catalog, PL/pgSQL, triggers, sequences, auth.
5. Version control: branches, commits, merge, conflicts, diff, remotes, backups, GC, cluster replication.
6. Operational features, logical replication, admin tool, and performance work against sysbench.
7. Remove the Go code.

## Baseline artifacts

- `testing/go/regression/out/results.trackers`: per-statement regression results of the Go baseline.
- `testing/go/regression/out/baseline-report.txt`: readable form of the above.
- `testing/go/regression/out/dumps-status.txt` and `skipped-dumps.txt`: per-dump import results.

These live under a gitignored directory and are moved into the Rust test tree once it exists.
