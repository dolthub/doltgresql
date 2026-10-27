# Server expectation targets

Set `DOLTGRES_TEST_TARGET` independently of the server launcher:

| Value | Expectations |
| --- | --- |
| unset or `doltgres` | Original Doltgres expectations (the default CI behavior). |
| `postgres` | Native PostgreSQL expectations, validated against PostgreSQL 18.6. |
| `postgres-fork` | Native expectations, with explicit fork-specific counterparts where supplied. |

Unknown target names fail the test. The target does not enable the older `DOLTGRES_TEST_PROFILE=prolly` exclusions; leave that variable unset when measuring compatibility. The launcher is still selected by `DOLTGRES_TEST_SERVER_BINARY` and `DOLTGRES_TEST_SERVER_KIND`. Stock PostgreSQL also needs `DOLTGRES_TEST_TABLE_AM=heap`; the fork uses `dolt_prolly_table`.

For example, from the repository root:

```sh
DOLTGRES_TEST_TARGET=postgres \
DOLTGRES_TEST_SERVER_KIND=postgres \
DOLTGRES_TEST_SERVER_BINARY=/path/to/stock/bin/postgres \
DOLTGRES_TEST_TABLE_AM=heap \
go test ./testing/go -run 'TestAnyAllSerialization|TestSchemaVisibilityInquiryFunctions' -count=1 -v
```

Use `DOLTGRES_TEST_TARGET=postgres-fork`, the fork's binary, and `DOLTGRES_TEST_TABLE_AM=dolt_prolly_table` for the fork. Use `doltgres` with an external Doltgres binary, or omit all launcher settings for ordinary in-process Doltgres tests.

## Adding counterparts

Add `Postgres: &ScriptTestAssertion{...}` next to the original assertion. The runner reports both variants as subtests, skips the unselected one, and executes exactly one. An empty counterpart `Query` inherits the original query; **no other expectations, credentials, bind parameters or execution modes are inherited**. Specify them explicitly. A pre-existing unconditional `Skip` remains in force for the counterpart. Focus selection remains in force as well.

`PostgresFork` supplies a third, fork-only assertion. If there is only a `PostgresFork` counterpart, both original and fork assertions are skipped on stock: stock has no version-control API. This does not automatically skip other statements in a version-control script; do not select entire Dolt-function suites for a stock run.

Use a `ScriptTest.Postgres` counterpart when setup or statement sequencing differs. Both scripts get fresh servers and visible selection subtests. This is also supported by `RunTransactionTest`. Wire assertions use `PostgresReceive`, sharing their identical sent messages. Hand-written tests can use `runTargetPair`.

Do not generate expectations from the result being tested, skip after observing a failure, or accept multiple answers merely to turn failures green. Native catalog OIDs must be resolved from this fixture's names or catalog joins. EXPLAIN counterparts validate native JSON plan structure and binding, not Dolt's textual planner tree or machine-dependent costs. Fork-only catalog plans require `DoltCatalogScan` nodes for the relevant aliases. These checks do not assert that the native optimizer must choose an index for a tiny catalog.

## Intentional differences covered here

- Dolt's 20-byte hashes render as 32 base-32 characters; the fork renders SHA-256 as 64 hexadecimal characters. A separate commit-identity test captures the returned hash and checks log/branch agreement and retained revision data. It does not rely on a fixed number of initialization commits.
- User-object OIDs and the information_schema namespace OID are runtime identifiers. PostgreSQL exposes additional built-in routines, types, system columns, and PostgreSQL 18 NOT NULL constraints.
- Native advisory lock functions return void. PostgreSQL retains an empty placeholder for a previously set custom GUC after DISCARD ALL.
- Native JSON preserves input whitespace; XML binary output includes the connection's encoding declaration. The wire harness starts its connection in WIN1252.
- PostgreSQL 18 CSV COPY consumes a literal backslash-dot line as data; native fixture copies omit that legacy marker. Binary COPY accepts EOF without the optional trailer; counterpart tests check all rows and subsequent connection use.
- PL/pgSQL variables default to NULL; native loop fixtures initialize counters explicitly.
- Native DDL deparse formatting, procedure argument modes and SQLSTATEs differ. Unique indexes are not automatically unique constraints.
- Native transaction lock tests use explicit COMMIT and SET LOCAL; Dolt's original implicit version-control commit/cache behavior remains separately tested. The fork's version-control commands require a standalone statement; aggregate/filter wrappers around a mutation are not a valid native alternative. Capture its result first and validate it separately.

This is an incremental compatibility baseline, not a claim that all Doltgres tests now pass on PostgreSQL. Unsupported fork features and mismatches with independently validated native expectations must continue to fail. Broad GUC suites, remaining catalog families, engine-specific syntax and unconverted feature expectations require later passes.

## Full server-suite validation

The inventory runner discovers every top-level server test in `testing/go`,
`testing/go/enginetest`, and `testing/go/extensions`. It rejects newly discovered
unmapped test packages rather than silently omitting them. It excludes regression
replay and non-server unit tests; the go-sql-driver suite is outside this tree.
Existing test-defined skips remain visible. This is full-suite execution, not a
promise that all expectations have native counterparts yet.

Build the three binaries, then run without `--tests`:

```sh
mkdir -p /tmp/compatibility-tests
go test -c -o /tmp/compatibility-tests/go.test ./testing/go
go test -c -o /tmp/compatibility-tests/engine.test ./testing/go/enginetest
go test -c -o /tmp/compatibility-tests/extensions.test ./testing/go/extensions
python3 scripts/compatibility/run_server_suite.py \
  --build-dir /tmp/compatibility-tests --output /tmp/compatibility-results \
  --stock /path/to/stock/bin/postgres --fork /path/to/fork/bin/postgres \
  --doltgres /path/to/doltgres
```

Install pgvector and uuid-ossp in both native server installations before testing
extensions. Engine groups share a fixed port and must not run concurrently across
separate runner invocations. Each group has a 20-minute timeout by default.
The audit's opt-in panic recovery records assertion panics as failures and lets
later assertions run; it never converts a panic to a pass. Logs, JSON events,
binary hashes, and every attempt are retained. A rerun uses a different
`--attempt` name; `--tests` is only for diagnosis after the complete sweep.

Native counterparts now cover boolean SHOW output (`on`/`off` rather than integer
flags), the restart-only archive_mode setting, and large text/bit payload tests.
The latter use SQL-generated native fixtures because PostgreSQL has no LOAD_FILE
and forbids NUL in text. They preserve lengths and value distinctions, comparisons,
updates, and multi-column checks. Doltgres keeps its original fixtures. External
servers receive private copies of testdata and test TLS certificates when requested.
Replication additionally requires a logical-replication primary on port 5432 and
`GITHUB_ACTION=1`; ordinary discovery retains that test's existing environment guard.
