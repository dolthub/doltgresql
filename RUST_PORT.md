# Doltgres Rust port — scope and plan

Working document for the experimental Rust rewrite. The Go code stays in the repository alongside the Rust code, so
that GitHub shows the Rust changes on their own; continuous integration moves to the Rust server.

- Worktree branch: `daylon/rust-port`
- Pinned Go baseline: `96c7b6ba630cd0862c9e0a74714b5b12c04626f7` (local tag `rust-port-base`)
- Toolchain: stable Rust via Homebrew `rustup`

## Goals

1. Client I/O identical to the Go version: wire messages, rows, errors and SQLSTATEs, notices, tags.
2. Storage bidirectional: Go and Rust read, write, push, and pull each other's repositories and remotes, including
   every legacy format the Go version can still read, with the same behavior (merges and diffs included). Bytes and
   hashes may differ from Go's wherever readers cannot tell, so internals follow the fastest design rather than Go's.
3. Performance better than Go's on every tested workload and data size: DoltHub's published sysbench tests and
   TPC-C, a benchmark of complex queries, several threads, and databases of a gigabyte and more.
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
- The Go code stays in the repository as a reference and test oracle, and is not deleted.
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
  `testing/dataloader`, and `TestDropRoleCleansPersistedAuthorizationReferences`, which calls the auth package
  directly. Their observable requirements (legacy formats, serialization) are covered by compatibility testing.
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

Status as of 2026-10-08. The untracked `HANDOFF.md` holds the exact position and the approved plan in detail.

0. (Done) Cargo workspace with the unsafe lint; wire codec (verified against pgproto3); pgx-equivalent client and
   harness; Postgres expectation capture; port of every Go test.
1. (Done) Storage read path: hashes, NBS table files, journal, archives, manifest, flatbuffers messages, prolly
   trees, commit graph. Verified by dumping Go-written repositories identically.
2. (Done) Storage write path: Go reads everything Rust writes and the reverse. Bytes and hashes matched Go's at
   first, and now may differ wherever readers cannot tell.
3. (Done) Wire protocol, parser, catalog, and a minimal engine. The parser is `pg_query` (libpg_query, Postgres' own
   grammar, currently 17.7), and the engine is our own row engine built for OLTP over prolly trees. Syntax newer than
   Postgres 15 is accepted; a statement fails only when what it needs is unsupported.
4. (Done) Breadth: types, functions, operators, DDL, DML, pg_catalog, PL/pgSQL, triggers, sequences, auth.
5. (Done) Version control: branches, commits, merge, conflicts, diff, remotes, backups, GC, archives, cluster
   replication, apart from persisted statistics and column-level schema merges.
6. (Done) Operational features, logical replication, the admin tool, and performance: faster than Go on DoltHub's
   published sysbench tests, TPC-C, and a benchmark of 115 complex queries, at one and four threads, on small data,
   at a gigabyte, and at three gigabytes. Sessions share each database, reads run concurrently, merges read only what
   changed, and garbage collection copies without blocking writers.
7. The planner (rule-based choices with adaptive joins, catalog index scans, statistics-driven join order), the
   remaining test failures, fixes for Go bugs that tests encode, and the git and ssh remotes.
8. Continuous integration against the Rust server, with comparisons between Go on `main` and Rust on this branch,
   and a rebase onto the latest `main` that ports what changed in Go since. The Go code stays in the repository.
After Phase 8 the work is continuous rather than phased: keep CI green, raise the Postgres regression replay to at
least 75%, close quick compatibility gaps (missing functions, casts, and the like), and move performance toward
Postgres 15's speed rather than Go's, without losing compatibility. `GO_REWRITE_GUIDE.md` records how this rewrite was
done, for a possible Go rewrite to compare against.

## Phase 0 status

Done:

- `pgproto`: wire codec, byte-identical to pgproto3 on golden messages.
- `harness`: pgx-equivalent client, binary decoder (cross-checked against Postgres' own output), script runner,
  wire conversation runner, plan facts, server launcher, recording and structured failure output.
- `goport` (temporary): dumps the Go tests, captures them against Postgres and the Go server, and generates Rust.
- `crates/tests`: 2,292 scripts and 198 wire conversations ported. Expectations come from Postgres 15 for 11,823
  assertions, from Postgres 17 for JSON_TABLE, and from the Go server for 2,023 Dolt-dependent ones (marked
  `// Doltgres-specific`). 105 EXPLAIN assertions are plan facts. Values that vary between two Postgres runs are
  `Any`, and a Go error message that varies between runs is matched by its common leading lines. The server port is
  `{PORT}`, and each script's temporary directory is `{TEMPDIR}`.
- Hand ports (`crates/tests/tests/*.rs` beside `scripts/`): SSL, the missing-database connection and the invalid
  startup timezone pass on Postgres 15 and the Go binary. The Dolt backup and remote suites (59 scripts) go through
  goport, with their temporary directories as `{TEMPDIR}` and `{NEWDIR:name}`, and pass on the Go binary.
- Round trip: every Postgres-sourced assertion passes against a real Postgres 15.
- Client traffic: the Rust suite sends the same frontend messages as the pgx recordings of the Go suite for 2,022
  scripts. The 21 that differ are deliberate input changes and testdata paths.
- `regression`: the Postgres regression replay, with pgx's cell decoding emulated so that rows match exactly when the
  Go replay's rows match. Every distinct cell of the recordings and of the Go server's responses decodes like the Go
  replay (a kept fixture test). Against the Go binary it gives the Go replay's result for all 42,090 statements,
  apart from one `now()` comparison that also varies between Go runs. The one known difference: a timestamp in the
  machine's local zone prints a numeric zone name, which only changes error text.
- `logictest`: the sqllogictest runner, sending what pgx v4's database/sql driver sends and scanning values the way
  the Go harness does. On a test file covering every result path, its log matches the Go runner's record for record,
  messages included. The full corpus has since been run with both runners.
- Dump imports (`crates/tests/tests/dumps.rs`, ignored by default like the Go test): the same 45 of 103 dumps pass as
  under the Go test. The proxy serves psql's connections concurrently, so the 12 dumps that hung the Go test by
  reconnecting now run and fail on real server errors.

- Enginetest Dolt version-control sets (`doltgres_engine` module): every script that passes in the Go run of the
  15 version-control enginetests (183 scripts, 4,275 statements) is replayed statement by statement over the simple
  protocol, as the Go harness sends them. Expectations come from the Go binary, cross-checked against each statement's
  outcome in the Go test process. 7 `dolt_help` statements take the Go test process's rows, since the Go binary has
  no help text, so they fail on the Go binary.
- `driver`: the go-sql-server-driver runner, with pgx v5's database/sql value conversion and TLS modes. All 99 YAML
  cases and every Go test (large values, type diversity, wide tables, GC, auto GC, concurrency, metrics auth) are
  ported, one test per subtest. Against the Go binary, each passes where the Go run passes, and each Go skip is an
  ignored test with the same reason. TestStatsGCConcurrency's skip is stale: it passes in Go and Rust once unskipped.

Remaining: 36 assertions that neither Postgres nor the Go server can produce (the Go suite skips them too; in 7 the
Go server panics), which stay skipped.

## Phase 1 status

Done (read path):

- `store`: hashes, manifests, NBS table files, the chunk journal (with Go's recovery and data loss rules), archives
  versions 1 to 3 with zstd dictionaries, and the old and new generations.
- `serial`: a bounds-checked flatbuffers reader, since `flatc`'s Rust output uses `unsafe`, with views of the store
  root, commits, tags, working sets with merge and rebase state, Doltgres root values, tables, schemas, foreign keys,
  stashes, and tree nodes.
- `prolly`: tree nodes of every kind, leaf walks, tuples, and blob trees.
- `objects`: Doltgres root objects (sequences, types, functions, triggers, extensions, procedures, casts, operators,
  aggregates, conflicts) and serialized types, including column types.
- `doltdb`: a reader of the whole object graph from the refs down to row tuples, with out-of-band values resolved.
- Verification: Go oracles that use Dolt's and Doltgres' own readers (kept locally in `testing/go/regression/out`)
  print the chunks and the object graph, including row digests, of fixture databases. The Rust readers match them
  exactly on 16 fixtures: databases from the current Go server covering every message type, multi-level maps, merges
  with table and root object conflicts, stashes, rebases, constraints and index options, and blobs at each level
  boundary; databases from Doltgres 0.50, 0.56, 0.57, and 1.0; and
  version 1 and 2 archives written by Dolt. Doltgres 0.18 databases are out of scope, since the Go server cannot read
  them, and 0.52 to 0.56 lose chunks in their own GC.

Since done with later phases: values by type, auth and branch control files, vector index nodes, and runs against
gigabyte databases. Persisted statistics remain.

## Phase 2 status

Done (write path), each checked against every chunk of the Go-written fixtures:

- `serial`: a port of Dolt's flatbuffers builder, and writers of store roots, commits, tags, working sets, stashes,
  tables, schemas, foreign keys, and Doltgres root values in Go's build order. Every message rewrites identically,
  including the layouts that Go's in-place edits leave behind (a root object field added after the others, an
  auto-increment value set to zero).
- `prolly`: serializers of every tree node kind; the chunker, which rebuilds every tree node from its leaf items; the
  blob builder, which rebuilds every blob, including the single-child roots Go builds when a blob fills its levels
  exactly; and cursors with in-place edits (a port of Dolt's ApplyMutations), which build the same trees as building
  from empty over random edits to trees of three and more levels.
- `objects`: serializers of every root object, which write current versions as Go does.
- `store`: chunk records (the Rust `snap` crate compresses every chunk exactly as `golang/snappy` did), table files
  (every one rewrites identically, name included), manifests and their lock hashes, and journal records (every one
  re-encodes identically). libzstd is pinned to 1.5.6, the version Dolt's gozstd bundles.
- The journal writer replays every fixture journal into identical bytes, and keeps an index file that Go accepts
  (checked by having Go open a Rust-written journal of 45,000 chunks over several index batches).
- The journaling store (`JournalStore`) puts chunks through a memtable, commits roots to the journal, rewrites the
  manifest only when its files change, trues up the manifest's root on open, locks out other processes, and refuses
  dangling references. Go reads databases that it wrote and Rust extended.
- `serial::walk`: the addresses each message refers to, in the order Dolt's WalkAddrs visits them, matching Go on
  every chunk of every fixture.

Findings:

- Go fuses multiply-adds into FMA instructions on arm64 but not on amd64, so Dolt's `math.Expm1`, which decides
  where nodes end, rounds differently on the two for 25 of the 16,385 possible inputs. A boundary decision flips only
  when a key's hash falls within an ulp of the threshold, so this is very rare, but the same data can chunk
  differently on the two architectures. Either shape reads correctly everywhere, but identical data can then have
  different hashes. The Rust chunker always rounds as Go does on amd64, checked exhaustively against Go, so it is
  deterministic across platforms and matches Go on amd64. Go on arm64 can still differ from it in these rare cases.
- Table file indexes sort records by prefix with Go's unstable sort, so two chunks whose 8-byte prefixes collide can
  land in either order. The Rust writer keeps them in insertion order.
- Go's GC is not deterministic: on two copies of one database, it writes the same chunks to each generation, and the
  same new-generation table file, but old-generation files whose chunk order differs. GC is checked by the chunks
  of each generation, not by file bytes.
- Dolt's journal iterator reports the chunks it found through the index file by the first 16 bytes of their
  addresses, padded with zeros. Comparisons with Go's iteration go by each chunk's data hash.

Since done with later phases: GC and the writable old generation, conjoining, archive writing, and the auth and branch
control files. Persisted statistics remain.

## Findings in Go

`GO_FINDINGS.md` records surprising behavior in the Go implementation, and problems that would be very hard to change
there, as evidence for the port beyond performance. Add to it as they are found.

## Baseline artifacts

- `testing/go/regression/out/results.trackers`: per-statement regression results of the Go baseline.
- `testing/go/regression/out/baseline-report.txt`: readable form of the above.
- `testing/go/regression/out/dumps-status.txt` and `skipped-dumps.txt`: per-dump import results.

These live under a gitignored directory; the Rust suites keep their own results.
