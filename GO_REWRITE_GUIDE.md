# Rewriting Doltgres from scratch: the playbook

This document records how the experimental Rust rewrite of Doltgres (branch `daylon/rust-port`) was carried out, step
by step, so that another agent can run the same experiment in Go. The aim of such a run would be a fair comparison:
the same starting point, the same constraints, the same tests, and the same order of work, with only the language
changing. Nothing here has been started in Go. Read `RUST_PORT.md` (the scope and plan) and `GO_FINDINGS.md`
(surprises in the existing Go implementation) alongside this file.

## 1. What the rewrite is

Doltgres is Dolt (Git for data) with a Postgres front end. The existing implementation parses Postgres SQL, converts
it to a Vitess (MySQL) AST, and runs it through go-mysql-server (GMS) over Dolt's storage. The rewrite replaces all
of that, Vitess, GMS, and Dolt included, with a single new implementation that:

1. Speaks the Postgres wire protocol exactly as the Go server does (rows, errors with SQLSTATEs, notices, tags).
2. Reads and writes the same on-disk repositories and remotes as the Go server, in both directions, including every
   legacy format the Go server can still read. A Go server must keep working on a repository the rewrite touched,
   and the reverse.
3. Is faster than the Go server on every tested workload and data size.
4. Matches real Postgres 15 wherever Postgres and Go differ, and keeps every Doltgres-only feature, in a Postgres
   style.

### The starting point and timeline

- The Rust rewrite branched from `main` at `96c7b6ba630cd0862c9e0a74714b5b12c04626f7` (local tag `rust-port-base`)
  on 2026-10-05.
- It reached a draft PR with CI pointed at the new server on 2026-10-08, after about 250 commits over 9 agent
  sessions (each session lasting until its context had been compacted several times). The order of work is below.
- For a fair comparison, a Go run should branch from the same commit (or record the commit it used), then rebase onto
  a later `main` and port what changed, as the Rust run did in its CI phase.

### Results to compare against (as of 2026-10-09)

| Measure | Go server on `main` | Rust rewrite |
| --- | --- | --- |
| Ported script tests (`testing/go`, PG15 expectations) | fails many by design | 616 of 616 test functions pass |
| Postgres regression replay (42,090 statements) | 20,729 (49.25%) | 31,585 (75.04%) |
| Postgres 18 regression suite (51,577 statements, 2026-10-10) | 17,921 (34.75%) | 31,031 (60.16%) |
| Dump imports (103 dumps) | 45 | 45 |
| sqllogictest | 99.317% | above Go |
| Bats, client-language, compatibility, driver suites | pass | pass |
| sysbench (one thread, tps) point_select / read_only | 10,522 / 504 | 20,603 / 1,284 |
| 119-query complex benchmark, geometric mean time | 1.00 | about 0.28 |

The Rust run reached its goal of 75% on the regression replay on 2026-10-09. The next goal is sysbench throughput
within 20% of Postgres 15's: on Linux CI, one thread, the Rust server reaches 48-82% of Postgres' tps on reads and
24-38% on writes.

## 2. Ground rules that shaped every step

These came from the owner and should carry over unchanged, apart from swapping "safe Rust" for its Go equivalent.

- **Language safety.** Rust: `#![forbid(unsafe_code)]` in every crate, through a workspace lint. Third-party crates
  may contain unsafe code, and C libraries (libpg_query, zstd, libxml2, ICU) are allowed when well understood. The Go
  equivalent: no `unsafe` package and no `//go:linkname` in your own code, cgo only through well-known libraries.
- **No dependency on the old stack.** Nothing from Vitess, GMS, Dolt, or the existing Doltgres Go code is imported.
  They are reference material and a test oracle only. In Go this temptation is much stronger, since you could just
  import Dolt's `store` packages. Don't, or the comparison measures nothing. Popular, maintained libraries are fine
  (the Rust run used pg_query, rustls, zstd, snap, fancy-regex, libxml2, yaml-rust2, tonic, and reqwest).
- **The Go code stays.** The new implementation lives beside it (the Rust run used `crates/` and `third_party/`), so
  that the PR's diff shows only the new code. CI moves to the new server at the end. In a Go rewrite, put the new
  module in its own directory (for example `rewrite/`) with its own `go.mod`.
- **Behavior, not internals.** Client-visible input and output must match. Storage must be readable both ways with the
  same behavior, but bytes, hashes, chunk boundaries, and journal records may differ wherever no reader can tell.
  Early in the Rust run, storage had to be byte-identical; this was relaxed on day 2, once byte identity had been
  proven and started blocking faster designs. Proving byte identity first was still worth it: it is how every format
  was learned exactly.
- **Postgres wins.** Where Go diverges from Postgres 15 (error codes, messages, generated constraint names, case
  folding, catalog contents), the rewrite matches Postgres. Expectations that encoded Go's behavior get rewritten, and
  each divergence that is surprising or hard to fix in Go is logged in `GO_FINDINGS.md`.
- **Doltgres-only statements stay, Postgres-flavored.** `DESCRIBE`, `SHOW TABLES`, `SHOW CREATE TABLE`, `USE`,
  `AS OF`, `CREATE USER IF NOT EXISTS`, and the like are kept, but their output is modeled on psql (`\d`, `\dt`, and
  so on) rather than on MySQL.
- **Storage format extensions.** Postgres metadata that Go's formats lack goes into optional fields written only when
  present, so an old Go reader fails only on data that uses the new feature. When no such field fits, encode it Go's
  way and record the gap.
- **Tests.** Every check becomes a kept test. Expectations come from a real Postgres 15, recorded by tooling and never
  written by hand. A known divergence stays as a skipped test with a reason, never deleted. New tests go after every
  existing test in a file.
- **Process.** Commit locally as you go (one capitalized past-tense subject line, Oxford commas, no body, no agent
  attribution). Never push until told to. No subagents. Keep a handoff file with the exact position, and hand off to a
  fresh session after three context compactions.
- **Autonomy.** After the owner approved the plan, the run worked without asking questions: it made decisions,
  recorded them in the handoff file, and kept going, with a cron job every ten minutes to resume after interruptions.

## 3. The order of work

Each phase below lists what was built, how it was verified, and what to watch for in Go. Do the phases in this order.
The order matters: the test harness comes before the server, and storage comes before SQL.

### Phase 0: tests first (day 1)

Before writing any server code, port every test, with expectations taken from real Postgres. This is the most
valuable step, since every later phase is driven by these failures.

1. **Wire codec.** Write a Postgres wire protocol codec for both directions. Verify it byte for byte against
   `pgproto3` (from pgx) on golden messages.
2. **A pgx-equivalent client.** The Go suites use pgx, and their outcomes depend on exactly what pgx sends: statement
   cache names, describe-then-execute versus cached statements, result format codes, and parameter encodings. The
   Rust run wrote a client that sends exactly the same messages, and proved it by recording the frontend traffic of
   the Go suite and the new suite and diffing the transcripts (2,022 scripts identical, 21 deliberate differences).
   *In Go:* use pgx itself as the test client. This is a real advantage of Go and is fair to take, but record that
   the Rust run had to build its own client.
3. **A binary decoder** for every type's binary format, cross-checked against Postgres' own output.
4. **The script runner** (`crates/harness`). It spawns a fresh server binary per script on a free port with a
   temporary data directory, taking the binary from an environment variable (`DOLTGRES_TEST_TARGET=doltgres:<path>`
   or a Postgres target). It compares rows in order only under ORDER BY, and otherwise as a multiset. Errors are
   compared on severity, SQLSTATE, message, detail, hint, position, and the schema, table, column, and constraint
   fields. Add matchers for values that cannot be fixed (any OID, any PID, any timestamp, floats within 2 ULPs for
   libm-dependent functions) that still check type and shape.
5. **The porting tool** (`crates/goport`). It is the backbone of the test port:
   - *Dump:* a hook added to the Go test framework (`add_dump_hooks.py`, enabled by `DUMP_TESTS_FILE`) writes every
     Go script test as JSON instead of running it.
   - *Capture:* replay every script against a real Postgres 15 (fresh database per script, from a template) twice,
     and against the Go server twice. Values that differ between two runs of the same server become matchers.
   - *Generate:* emit Rust test modules. Expectations come from Postgres for everything Postgres can run. Assertions
     that need Dolt features (dolt_* functions, branches, merges, AS OF) take the Go server's output, reviewed
     against the test's intent and marked `// Doltgres-specific`. EXPLAIN assertions become plan-property checks
     ("uses an index range on t_idx", "is a hash join") rather than GMS plan text.
   - *Overrides:* where a Go test's input doesn't work in Postgres (Doltgres leniency, `public.gen_random_uuid()`,
     loops over uninitialized variables), change the input minimally through an overrides file, and have the
     generated test say `// Changed from the Go test:`.
   - *Record:* a subcommand that turns a plain SQL file (`-- name:` lines start scripts) into a recorded test against
     PG15. Every later fix adds a kept test this way (`testing/go/regression/out/kept_sql/*.sql`, then
     `goport record`, then `append_kept.py` to append it to the right module).
   - *Merge:* when Go tests change on `main`, regenerate, then merge per script (old generation, new generation,
     committed version) so hand fixes and overrides survive (`merge_scripts.py`).
6. **The other suites**, each ported as a runner that reproduces the Go runner's results exactly against the Go
   binary before it is ever pointed at the new server:
   - Postgres regression replay (`crates/regression`): 205 files, 42,090 statements, emulating pgx's cell decoding,
     so that it reproduces the Go replay's result for every statement against the Go binary.
   - sqllogictest (`crates/logictest`): sends what pgx v4's database/sql driver sends, scans values the way the Go
     harness does, and matches the Go runner record for record.
   - Dump imports: 103 dumps through psql, with a proxy that serves psql's reconnects concurrently (the Go test hung on
     12 of them).
   - Dolt's version-control enginetest scripts (183 scripts, 4,275 statements), replayed over the simple protocol,
     with expectations from the Go binary.
   - The go-sql-server-driver YAML scenarios and Go tests (cluster, remotesapi, TLS, GC, large values).
   - Hand ports of the few Go tests that are custom code (SSL, missing database, startup time zone). Tests written as
     custom code over one pgx connection were ported by recording the exact bytes pgx sent and replaying them.
   - Left as they are, because they drive the binary: bats, the client-language tests, and the compatibility bats.
7. **Excluded:** syntax-only generation tests, unit tests of Go internal APIs, and enginetest sets with MySQL
   semantics.
8. **Round trip:** every Postgres-sourced assertion passes against a real Postgres 15 before the port moves on.

Result: about 2,300 scripts and 200 wire conversations, with 11,823 Postgres-sourced and 2,023 Go-sourced assertions.

### Phase 1: storage read path (day 1)

Learn every on-disk format by reading it, and prove each reader against Go-written data.

1. Build fixtures with the Go server (`build_store_fixtures.sh` runs SQL files from `storefixture_sql/`): one per
   feature (journal, GC, archives, large maps, conflicts, rich types, wide rows, merge and rebase states, schemas,
   blobs at each level boundary, empty), plus databases written by old releases (Doltgres 0.50, 0.56, 0.57, 1.0) and
   archives written by Dolt (versions 1 and 2). Doltgres 0.18 is out of scope since the Go server can't read it.
2. Write small Go "oracle" programs that use Dolt's and Doltgres' own readers to print, for each fixture, every chunk
   (`oracle.txt`), the whole object graph with row digests (`graph.txt`), and the addresses each chunk refers to in
   WalkAddrs order (`refs.txt`). Keep them outside the shipped tree (the Rust run used the gitignored
   `testing/go/regression/out/`).
3. Write the readers, in dependency order, and make each print the same text as the oracles: hashes and manifests,
   NBS table files, the chunk journal (with Go's recovery and data-loss rules), archives with zstd dictionaries, the
   old and new generations, flatbuffers messages, prolly tree nodes and tuples, blobs, the commit graph, Doltgres root
   values, and every root object (sequences, types, functions, triggers, extensions, procedures, casts, operators,
   aggregates, conflicts), including serialized column types.
4. Rust note: `flatc`'s Rust output uses `unsafe`, so the Rust run wrote a bounds-checked flatbuffers reader by hand.
   *In Go:* generated flatbuffers code is fine, but generate it yourself from Dolt's `.fbs` schemas rather than
   importing Dolt's `gen/fb/serial` package.

### Phase 2: storage write path (days 1 and 2)

1. Port Dolt's flatbuffers builder, then writers of every message type in Go's build order. Verify that every fixture
   message rewrites to identical bytes, including layouts that Go's in-place edits leave behind.
2. Port the chunker, the blob builder, tree serializers, and Dolt's ApplyMutations. Verify that every fixture tree
   rebuilds identically, and that random edits build the same trees as building from empty.
3. Chunk records (snappy, exactly as `golang/snappy` compresses), table files, manifests and lock hashes, journal
   records, and a journal writer whose index file Go accepts.
4. A journaling store with a memtable, commits to the journal, a manifest rewritten only when files change, a LOCK
   file, and dangling-reference checks.
5. Database creation that writes the journal Go writes for a new database, record for record. Learn each write path
   by building a tiny fixture with the Go server and dumping its journal in write order (`journaldump`).
6. Verify interoperability both ways: Go opens what the new code wrote, and the new code opens what Go wrote.

Watch for, in Go: Dolt's chunker calls `math.Expm1`, and the Go compiler fuses multiply-adds into FMA instructions on
arm64 but not on amd64. For 25 of 16,385 inputs, chunk boundaries therefore differ between architectures. The Rust run
always rounds as amd64 does, which was verified exhaustively. A Go rewrite must stop the compiler from fusing, with an
explicit `float64(x * y)` conversion around each product, and verify all 16,385 inputs on both architectures. Also:
Go's table file index sort is unstable, GC is nondeterministic in its old-generation file order, and the journal index
keys chunks by 16 of their 20 address bytes (see `GO_FINDINGS.md`).

### Phase 3: server, parser, minimal engine (day 2)

1. Parser: `pg_query` (libpg_query, Postgres' own grammar, currently 18.6). The Rust run vendored it to expose
   syntax error cursor positions (`third_party/pg_query`). *In Go:* `pganalyze/pg_query_go` is the same library
   through cgo. Accept newer syntax, and fail a statement only when what it needs is unsupported.
2. A small extension parser, tried when pg_query fails, for the Doltgres-only syntax (`USE db/branch`, `AS OF`,
   `CREATE ... IF NOT EXISTS` for users and roles, `SET name = expression`). It cuts the extra text out and replaces it
   with spaces, so every error position still points into the original query.
3. The server: SCRAM-SHA-256, the startup handshake (Postgres 15's full ParameterStatus list, `server_version`
   15.17), the simple protocol, then the extended protocol (Parse, Bind, Describe, Execute, Close, Sync, Flush), one
   thread per connection.
4. The engine: the built-in type catalog (generated from the Go server's serialized types by a `typeoracle`
   program), transactions over Dolt working sets, CREATE TABLE, INSERT, UPDATE, DELETE, single-table SELECT.
5. Decision: an own row engine built for OLTP over prolly trees, not a general analytics engine.

### Phase 4: breadth (days 2 and 3)

Work through the scripts suite's failures by weight: run the suite with a failures file
(`DOLTGRES_FAILURES_FILE`), group the failures (`analyze_rust_failures.py`), and fix the biggest group next. The order
the Rust run took: numerics and adaptive values, settings, the function framework with overload resolution, the
planner (joins, subqueries, grouping, set operations, VALUES), defaults and checks, datetimes, arrays, Dolt's core
procedures and system tables, indexes and unique constraints, sequences, ALTER TABLE, views, WITH and recursion,
window functions, pattern matching, JSON, RETURNING and ON CONFLICT, foreign keys, the system catalogs with Postgres'
real OIDs, roles and privileges, functions, procedures, PL/pgSQL, triggers, enum, composite, and domain types, the
remaining scalar types, extensions (uuid-ossp and pgvector, emulated in code), COPY, to_char and friends, xml, and so
on. The commit log lists each step.

Practices that paid off:

- For every Doltgres root object (functions, types, triggers, and so on), store exactly what Go stores. PL/pgSQL, for
  example, is stored as Go's interpreter operations, so the new server compiles bodies into Go's operation format and
  runs that format, which keeps functions portable between the servers.
- After each write path, check Go-readability by writing with the new server and querying with the Go server.
- Measure the suite after each step. The Rust scripts suite went from 12% to 47% of assertions in one day of breadth
  work, then to 100%.

### Phase 5: version control (day 3)

Branches, commits, merges (three-way, with conflicts and constraint violations), diffs and history tables, revert,
cherry-pick, rebase, stash, remotes (file, then http through a remotes API server, then cloud blob stores, then git and
ssh), backups, garbage collection (generational, archives, conjoin, automatic GC), and cluster replication. Verify
each against the Go server's behavior on the same statements, and check that mixed Go and new-server histories merge
cleanly. Column tags must be generated exactly as Dolt does (the Rust run ported Go's `math/rand`), since merges
match columns by tag.

### Phase 6: operations and performance (days 3 and 4)

1. Operational features: the startup integrity check, automatic GC, TLS, read-only mode, logical replication from a
   Postgres primary, and the admin corruption report and repair tool.
2. Performance tooling (`crates/bench`): a corpus of 119 named queries over generated tables. It runs any set of
   servers side by side, saves and loads baselines (Go and PG15 baselines were saved once and loaded afterwards), and
   a profiling script samples the server and summarizes hot functions. Also sysbench, TPC-C, several threads, and
   databases of 1 and 3 GB (`testing/go/regression/out/tools/scale.sh`).
3. What made the difference, in the order it was done: a streaming executor (row iterators for every plan node);
   decoding only the needed columns; index range seeks in both directions; dropping filters that index ranges answer
   exactly; incremental aggregates; per-thread caches of decoded table definitions; counting rows from tree counts;
   reusing row buffers; writing result values straight into DataRow messages; adaptive joins (lookups by primary key
   until hashing is cheaper, then hash joins, plus anti and semi joins); syncing the journal outside the database lock
   so concurrent commits share one sync; databases shared across sessions with internal locks and concurrent reads;
   garbage collection that copies from a snapshot without blocking writers.
4. Lesson: do large structural refactors (the concurrency model) before point optimizations that they could make
   irrelevant.

### Phase 7: planner and the long tail (day 4)

A rule-based planner over row estimates from index trees (not a port of GMS's memo coster): lookups through any
index, hash joins, semi and anti joins for EXISTS, `lookup_join` hints, catalog index scans, and EXPLAIN in Postgres'
text format plus a Doltgres `Index Ranges:` line. Then the remaining suite failures to zero, and fixes of Go bugs that
Doltgres-specific tests encode (rewrite the expectation, log the bug in `GO_FINDINGS.md`).

### Phase 8: CI and catching up with `main` (day 4)

1. Point every CI workflow at the new server: build it, run its suites, keep the bats, client-language, and
   compatibility workflows driving the new binary. The comparison workflows (regression replay, sysbench) build Go
   from `main` and the new server from the branch, and post the difference on the PR.
2. Push a backup of the branch, rebase onto the latest `main`, and port every behavior change, new test, and fix that
   landed in Go since the branch point. Regenerate the ported tests with goport and merge them per script.
3. Fix CI on every platform (Linux, macOS, Windows). Issues the Rust run hit that a Go run will hit too: Windows paths
   in Git Bash (`cygpath -w`), libm digits that differ between macOS and glibc, OS error text, and CI caches shared
   across runner images. Bats and client tests that encoded Go-only output were rewritten to Postgres' output, with
   each rewrite logged.
4. Run the version-compatibility suite (`integration-tests/compatibility/runner.sh`) locally before pushing. It
   downloads real old releases, creates repositories with them, and checks both directions. It caught several things
   no other suite did: empty root object addresses written as zeros by old releases, the old text-column encodings
   (`STRING_ADDR` and friends), index order fields that old releases can't parse, and array type versions.

### After Phase 8: continuous work

No more phase names. Priorities: keep CI green, raise the regression replay (goal at least 75%), close compatibility
gaps (missing functions, casts, and so on; `testing/go/regression/out/tools/missing_functions.sh` lists the built-in functions Postgres has and
the server lacks), and move performance toward Postgres 15's.

On 2026-10-10 the owner moved the target from Postgres 15 to Postgres 18, since the rewrite is meant to replace the
Go server. The recorded replay gave way in CI to Postgres 18's own regression suite, run directly against both servers
(`crates/regress`, `scripts/install_regress_suite.sh`). The runner feeds each test's script to psql 18 exactly as
pg_regress does: the same flags, environment variables, and database settings. It splits psql's echoed output into
one unit per line that sends statements, using psql's own lexing rules (quotes, dollar quotes, comments,
`BEGIN ATOMIC` bodies, COPY data, the ECHO variable, and skipped empty lines). Then it compares each unit with every
alternative expected file. Validate a runner like this against a real Postgres first: it should score 100%, and this
one does, on all 51,577 statements. The first run found crashes and hangs that the replay never reached: a
self-deadlock on the auth lock, a Flush that never delivered a queued error, and an unanswered FunctionCall.

Moving the kept script tests to Postgres 18 expectations was done in three passes:
- Run the whole suite against real PG15 and real PG18 and diff the failures. The failures that only PG18 has are the
  expectations to change.
- Re-record the Postgres-sourced assertions from PG18, leaving the ones marked Doltgres-specific alone. Then make the
  server match.

The traps:
- Re-recording turns OID matchers (`Oid(16385)`) into literal text, so restore them, and re-record every OID-bearing
  assertion of a touched script so that one numbering is used throughout.
- The harness's own server settings (`fsync=off`, socket paths) leak into recorded values.
- An installed extension's version must be the emulated one (pgvector 0.8.6, not the 0.8.7 that Homebrew ships).
- Scripts that run as other users, bind parameters, or read test data files cannot be recorded mechanically. Compare
  PG18's and the server's results for those by hand.

The catalogs, settings, and built-in rows were regenerated from a fresh PG18 cluster, and the server now reports
version 18.6.

The loop that drives this work:

1. Replay the whole regression suite, dump the trackers, and group the failures by their normalized error message
   (quoted names and numbers replaced) across all files. Pick the largest coherent cluster, not the worst file: one
   missing type family (ranges and multiranges) failed about 850 statements across several files.
2. Implement the feature by porting Postgres' own C semantics (for ranges, `rangetypes.c` and `multirangetypes.c`:
   bound comparison, canonical forms, parse and quote rules, and the exact error codes and details).
3. While building, diff each batch of statements between a fresh server and a real Postgres 15 with a small script
   (`rq.sh`: start the server on a temporary directory, run the SQL through psql against both, diff the outputs).
   Error codes need psql's `\set VERBOSITY verbose` or a recorded test to show.
4. Replay just the affected files plus `test_setup`, which creates the `regression` database the other files need
   (`replay_some.sh <name> <files...>`, which prints the per-file results and the top remaining errors).
5. Record kept tests from SQL files on PG15, run the whole scripts suite, and commit.

Fixing one feature often exposes a general gap behind it. Ranges needed polymorphic user-defined functions (Postgres
resolves `anyelement`, `anyarray`, and `anyrange` per call, rejects at CREATE a polymorphic result no input can
decide, and skips body analysis for polymorphic SQL functions), which in turn exposed that PL/pgSQL's `$N`
references never reached the Nth parameter.

Large C modules are worth porting whole rather than piecemeal: `geo_ops.c` (every geometric function, operator, and
cast, with its 1e-6 comparison tolerance and float overflow checks), `formatting.c`'s numeric part (`to_char` and
`to_number`), the text search parser's state tables, and the Snowball stemmer each passed hundreds of statements at
once, because the regression files test their corner cases exhaustively. Keep C's quirks, such as output that ends
at the first NUL byte. Kept tests should avoid results whose last digits come from the platform's math library (sin,
cos, and similar), since those differ between Linux, macOS, and Windows. Input parsers are worth the same treatment:
an ad hoc interval parser kept failing edge cases until `datetime.c`'s ParseDateTime, DecodeInterval, and
DecodeISO8601Interval were ported as they are (fields read right to left, overflow-checked accumulation, the
256-byte work buffer and 25-field limits), which passed 51 more statements of `interval` alone. The hash support
functions (`hashint4`, `hash_numeric`, `jsonb_hash`, `hash_array`, and the rest, with their seeded forms) are small
once lookup3 is ported, and they let the regression's consistency checks pass with Postgres' exact values.

Once the replay passed 75%, the work turned to sysbench against Postgres 15. Profile first: on macOS, 90% of a
write was the commit's F_FULLFSYNC, which Go's `File.Sync` and Rust's `sync_data` both issue, while Postgres' default
`wal_sync_method` there (`open_datasync`) never flushes the drive's cache. Commits now sync with a plain `fsync` on
macOS, a deliberate choice to match Postgres' durability there (Linux uses fdatasync either way). After that the
costs were CPU: binding a table's defaults re-parsed their stored text on every statement (now cached per thread),
SHA-512 chunk addresses (the ARMv8 instructions need `sha2`'s `asm` feature), prolly tree edits, and pg_query's
protobuf round trip for every statement. That round trip was most of a point select, and the C parser itself little
of it: protobuf-c sized and packed every `Node` (a oneof over several hundred node types) by walking all of its
fields. A patch to the vendored protobuf-c found a Node's one set field by id instead. libpg_query 18 replaced
protobuf-c with upb, whose encoder has the same cost (41% of a parse) and takes the same patch. pg_query_go parses
through the same protobuf path, so a Go rewrite that uses it needs the same patch.

The planner then moved toward Postgres' own (the user asked to port as much of Postgres' analyzer as is reasonable).
A first round hand-wrote Postgres' rules onto the existing planner: subquery pull-up, reduce_outer_joins, eqjoinsel,
dynamic programming over join orders, and column statistics computed as analyze.c's compute_scalar_stats does
(sampled evenly through the tree's subtree counts, so plans stay deterministic). The user rejected that approach:
"If you're finding gaps, then that sounds like you didn't port Postgres' analyzer over, as they would have already
identified any gaps in their analyzer." Port the optimizer itself, file by file (`crates/sql/src/optimizer/`, one
module per C file, keeping Postgres' function names): build Postgres' Query (range table, join tree, Vars) from the
binder's unplanned FROM plan, then run prepjointree.c, initsplan.c, allpaths.c, indxpath.c, joinrels.c, joinpath.c,
pathnode.c, costsize.c, clausesel.c and selfuncs.c over it, and createplan.c back into the executor's plan. Keep
Postgres' cost formulas and default settings; only the inputs change: a table's pages are those its rows would fill
in Postgres' heap, Dolt's primary index is a clustered index whose scans fetch nothing else, a covering secondary
index scan is an index-only scan, and an index's correlation comes from the statistics sample read in primary key
order. Doltgres' existing index scan chooser and lookup joins supply the index paths, which take Postgres' costs.
Stand the port beside the old planner behind a switch (`DOLTGRES_PG_PLANNER`) and grow it until it plans everything
the old one does, measuring it against Postgres' own EXPLAIN output on the regression files. Expect plan-shape tests
written for go-mysql-server's choices to disagree: Postgres reads tiny tables in full rather than through an index.
In Go, this planner would sit in front of go-mysql-server rather than inside it.

The first port followed Postgres 15; the user then asked for the latest released optimizer (PG 18), since later
releases fixed what earlier ones got wrong, and for everything that fits to be ported. Keep an inventory of every
function of src/backend/optimizer (with geqo, selfuncs.c, and analyze.c's statistics), each marked done, not
applicable with the reason, or pending, and never call the port finished while any is pending. PG 16 and later
change the core shapes, so port those first and together: Vars carry the outer joins that can make them NULL
(`varnullingrels`), outer joins get range table indexes that appear in relation sets, PlaceHolderVars keep a pulled-up
subquery's expressions below the outer joins that NULL them, equivalence classes absorb mergejoinable equalities (so
index lookups must ask them for join clauses), outer-join clauses are cloned for each order in which the joins
commute, and query_planner starts over after each join it removes (useless left joins, unique semijoins, PG 18's self
joins on a unique key). Doltgres has no operator families, so a btree family is the default one of the compared type,
and an expression whose type the planner cannot see simply takes no part in equivalence classes. In Rust the planner's
shared, mutable nodes (RestrictInfos, SpecialJoinInfos, classes and members, pathkeys) live in vectors on the
PlannerInfo and are referred to by index, and Vars are ids into a per-statement table, interned so that Postgres'
`equal()` becomes id equality. Dolt has no tuple IDs for bitmap scans, so a bitmap is the sorted set of primary keys
that its index scans find, BitmapAnd and BitmapOr merge those sets, and the heap scan looks each key up in the primary
index in key order; a parameterized bitmap scan is the inner side of a lateral nested loop and builds its ranges from
each outer row. A merge join runs as a hash join over its sorted inputs, which keeps the outer order that is all its
pathkeys promise, and no join path is parameterized, since only a nested loop's inner side reads outer rows.

## 4. Habits and tooling worth copying

- **A handoff file** (untracked) with an "Exact position" section at the top: the last commit, what is in progress,
  the newest failures file, and the next steps. Rewrite it at every handoff. Keep the approved plan, the rules, and
  every decision made without the owner in it.
- **Scratch tooling in a gitignored directory** (`testing/go/regression/out/`), since session scratch space doesn't
  survive: oracles, fixture builders, comparison scripts, baselines, emulators for cloud remotes, and saved suite
  results. These files exist only in the original machine's worktree. The scripts this document names (the dump hook,
  `merge_scripts.py`, `append_kept.py`, `analyze_rust_failures.py`, the oracles, and `tools/`) are small, and the
  descriptions here are enough to rewrite them.
- **Release versus quick builds.** The Rust run kept a fast `quick` profile for tests and iteration, and a full
  optimized release build for suites and benchmarks. Test binaries don't depend on the server crate, so the server can
  be rebuilt without rebuilding the tests.
- **Failures files.** Every suite can write structured failure records, which scripts group and count, so the next
  fix is always the biggest group.
- **Before each commit:** a comment pass (one-sentence doc comments, no in-body comments), the formatter, the
  linter with warnings as errors, and the affected suites. In Go: `gofmt`, `go vet`, and staticcheck.
- **The comparison table.** Keep a running table of every suite's result for Go and for the rewrite, so that progress
  is measurable at any point.

## 5. Mistakes the Rust run made, to avoid

- Running the server without a data directory once created databases in the home directory. Always pass a data
  directory, and set `DOLT_ROOT_PATH` for anything that runs the `dolt` binary.
- Copying a binary over an existing executable on macOS made later launches die with SIGKILL. Remove it first.
- The local Homebrew Postgres owns port 5432, so local runs of tests that hardcode that port fail. They pass in CI.
- Changing a shared lookup function to fix one caller (case-folding PL/pgSQL variable names for old triggers)
  silently changed another caller's behavior. Scope compatibility shims to the exact path that needs them, as Go does.
- Pushing before running every workspace test once let a stale count test fail CI. Run everything locally before a
  push.
- Writing an in-memory default back out (full-length index order vectors) broke old readers. Write optional fields
  only when they differ from what an old writer would have produced.
- Automatic garbage collection queued for the database's exclusive lock while a statement ran. Rust's (and Go's)
  read-write locks hold new readers behind a waiting writer, so one long-running query (here a regression query the
  client had already given up on) stopped every other statement until it ended. Background work that needs a
  database alone should poll for a moment without readers (`try_lock`) instead of queueing.
