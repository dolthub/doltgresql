# Catalog context compatibility gate

Run `DOLTGRES_TEST_SERVER_BINARY=/absolute/path/to/postgres testing/scripts/catalog-context.sh`
from this checkout. The installed binary must have its matching share/library
files. Do not rebuild or reinstall that installation during the run.

The script builds a temporary Go test binary and runs explicit, repeatable
selections using the existing `prolly` external-server profile. Every scenario
gets fresh database storage. It covers branch/history/merge/catalog gates plus
stored catalog introspection, CTE resolution and portable existing catalog cases.
The new TestBranchCatalogIntrospection also runs against ordinary Doltgres.

This is not the complete Doltgres suite. Excluded broad catalog scenarios assert
fixed OIDs, exact system inventories or engine-specific EXPLAIN/deparser output,
or require unsupported ALTER constraints, triggers and OID/XID storage. The
selected information-schema XML/default tests exercise preactivation PostgreSQL;
they do not establish branch-aware defaults or generated-column DDL support.
Existing framework profile exclusions are unchanged and remain reported as skips.
Use the preserved original tests for a broader diagnostic run; do not infer that
all unselected tests pass, or that tests with no activation cover revision binding.

The revision-view gate adds TestPostgresRevisionCatalogViews, TestBranchSearchPath,
and five existing TestSchemas scenarios. The view test is explicitly fork-only:
a direct probe of Doltgres returned current-branch metadata for three qualified
view assertions. It is skipped on Doltgres until that provider honors revision
qualifiers. TestBranchSearchPath passes on both implementations. The original
empty-path and non-default-schema scenarios were also probed; their failures were
PostgreSQL-versus-Doltgres error wording, so their assertions remain unchanged and
they are not part of this portable selection. Native tests cover those errors.
