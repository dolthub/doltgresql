# External PostgreSQL/prolly compatibility verification

Use the existing script harness with a fresh server cluster for every script:

```
DOLTGRES_TEST_SERVER_KIND=postgres \
DOLTGRES_TEST_SERVER_BINARY=/absolute/path/to/postgres \
DOLTGRES_TEST_PROFILE=prolly \
go test ./testing/go -count=1 -run '^(TestExternalServerCleanup|TestExternalPostgresIsolation|TestProllyVersionCompatibility|TestInsert|TestDelete|TestSelect|TestUpdate|TestUpdateAssignmentSemantics)$' -v
```

`initdb` must be alongside `postgres`. The harness initializes a private UTF-8/C
cluster with trust authentication, listens on the selected loopback port, and
sets default_table_access_method=dolt_prolly_table so ordinary test CREATE TABLE
statements exercise the new AM. Each script has fresh catalogs, table files,
manifest history, roles and transaction state. Shutdown precedes directory
removal; initialization failures also remove their private directory. No existing
cluster or user database is modified. PostgreSQL mode is explicitly selected;
Doltgres external binary startup retains its original command-line flags.

The profile is opt-in. `prolly_profile.go` contains exact named exclusions and
reasons for not-yet-implemented functionality or pre-existing PostgreSQL/Doltgres
behavior differences. No query is automatically skipped based on an error. Run
without the profile to audit the exclusions. Original assertion expectations are
not rewritten. Subtest selection now occurs before server startup, so filtered
scripts do not create unused clusters.

`TestProllyVersionCompatibility` isolates implemented operations from the broader
Doltgres scenarios that require checkout, full schema capture, status, or other
unimplemented APIs. Run the identical test without the external variables to
validate the expectations against the in-process Doltgres engine. It checks named
command outputs, staging versus live contents, empty-commit flags, branch/tag
resolution, history and diff. Hash-returning mutation statements execute directly;
subsequent history assertions verify their effects. This honors the approved
standalone-command rule without assuming Dolt's 32-character hash format.

The reference currently ignores dolt_log's -n option and does not match the
qualified table filter in the tested historical diff case. Shared tests therefore
use SQL LIMIT and an unqualified unambiguous table name; native TAP tests retain
coverage of our explicit limit and qualified-filter behavior. These differences
are recorded, not treated as cross-engine passes for those argument forms.

This is a bounded compatibility pass, not certification of all Doltgres features
or row-level concurrency. ON CONFLICT, row locks, SSI, branch switching and full
schema versioning remain outside the current milestone.

Add `TestTypes` to the run expression to include the built-in type suite. Its
explicit exclusions cover missing index operator classes and the approved
rejection of persisted OID/XID identifiers. The remaining excluded assertions
were reproduced against the same PostgreSQL build using heap storage:

```
DOLTGRES_TEST_SERVER_KIND=postgres \
DOLTGRES_TEST_SERVER_BINARY=/absolute/path/to/postgres \
DOLTGRES_TEST_TABLE_AM=heap \
go test ./testing/go -count=1 -run '^TestTypes$' -v
```

The heap control is expected to fail some Doltgres-specific casts, messages and
built-in behavior expectations. It distinguishes those differences from prolly
codec/index defects. All type assertions that failed after successful setup on
prolly also failed on heap in this verification pass. Do not use the prolly
profile when auditing the control, since that would hide the differences.

## Bootstrap history discrepancy

Doltgres records an extra CREATE DATABASE commit as an artifact of the Dolt
storage library. This is not desired behavior for the PostgreSQL fork, which
creates only the repository initialization commit before user commits. The
shared compatibility test accepts either three or four total commits after two
user commits and separately verifies both user commits in order. Do not add
synthetic history to the fork to equalize these counts. Existing histories are
not rewritten by the correction.

### Named-type and nested SQL checkpoint

`catalog-context.sh` now includes `TestBranchBuiltinResolution`, the original
`TestSelectFromFunctions`, 19 original operator scenarios, and 16 original
set-returning-function scenarios. The new activated-branch scenario passes on
both engines: qualified/unqualified bootstrap functions, explicit operator
qualification, native overloads and array helpers consume independent branch
inputs. Native PostgreSQL TAP tests separately exercise canonical composite
names and nested SQL execution, including arguments from historical catalogs.
Those tests use shared bootstrap helper routines; versioned user routines are
not implemented.

An exploratory run of five complete function/operator suites exposed 11 failing
scenarios. The same 11 fail on the preceding fork build, before this checkpoint:
unknown-function/operator error wording; numeric modulo infinity expectations;
JSONB comparison ordering; case-insensitive non-ASCII regex under the harness's
C locale; generate_series NaN error wording; and a partition-ancestor scenario
whose setup requires unsupported foreign keys. Their original assertions remain
unchanged. The routine gate selects complete passing scenarios explicitly;
it does not assert that the full five suites pass or that these are all catalog
consumer gaps.

### Procedural execution and typed parameters

The gate also runs the fork-specific immutable-composite extended-protocol test,
six original parameter-binding tests, and 18 original PL record scenarios.
The remaining FOR-loop record scenario expects `int4` rather than PostgreSQL's
`integer` in an error; the whole-record scenario expects an error for an
unassigned whole record where this PostgreSQL version returns NULL. Their original
assertions are unchanged and excluded as whole scenarios. Existing native
PL/pgSQL regression tests run separately, including native record semantics.
