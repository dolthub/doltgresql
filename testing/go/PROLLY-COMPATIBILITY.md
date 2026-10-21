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
