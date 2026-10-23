#!/usr/bin/env bash
# Run from any directory. Each scenario gets the harness's fresh server state.
set -euo pipefail
: "${DOLTGRES_TEST_SERVER_BINARY:?Set to the absolute path of the fork postgres binary}"
export DOLTGRES_TEST_SERVER_KIND=postgres
export DOLTGRES_TEST_PROFILE=prolly
export GOMAXPROCS=${GOMAXPROCS:-4}
cd "$(dirname "$0")/../.."
compat_binary=$(mktemp /tmp/catalog-compat.XXXXXX)
trap 'rm -f "$compat_binary"' EXIT
go test -c -o "$compat_binary" ./testing/go
# Version semantics and activated-catalog gates, plus native SQL/catalog inputs.
"$compat_binary" -test.v -test.timeout=15m -test.run '^(TestPostgresRevisionCatalogViews|TestBranchSearchPath|TestBranchCatalogIntrospection|TestBranchCatalogAlter|TestAsOf|TestDoltMergeData|TestPostgresMergeConflictStatus|TestBranchCatalogCreate|TestBranchCatalogDrop|TestBranchCatalogSameName|TestCatalogHistory|TestProllyVersionCompatibility|TestNativeVersionedQueries|TestNativeHistoricalCursors|TestRevisionWorkingSets|TestInfoSchemaTables|TestInfoSchemaPgCharMaxLength|TestWithStatements|TestWithRecursive)$'
# Preserve original tests; exclude whole scenarios requiring unimplemented DDL,
# detached-HEAD USE syntax, fixed OIDs or engine-specific system inventories.
"$compat_binary" -test.v -test.timeout=5m -test.run '^TestInfoSchemaRevisionDb$/^info_schema_changes_with_dolt_checkout$'
"$compat_binary" -test.v -test.timeout=5m -test.run '^TestPgIndex$/^pg_index_indoption_has_one_entry_per_key_column$'
"$compat_binary" -test.v -test.timeout=5m -test.run '^TestInfoSchemaColumns$/^(xml_columns|literal_column_defaults|cast_column_defaults)$'
# Relation-name and schema-path scenarios with portable result/error assertions.
"$compat_binary" -test.v -test.timeout=5m -test.run '^TestSchemas$/^(search_path_returns_user_table_before_public_table|public_schema_qualifier|public_schema_qualifier,_multiple_tables|public_db_and_schema_qualifier|insert,_update,_delete_with_schema)$'
