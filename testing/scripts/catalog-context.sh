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
"$compat_binary" -test.v -test.timeout=15m -test.run '^(TestPostgresCatalogTypedParameters|TestBranchBuiltinResolution|TestSelectFromFunctions|TestPostgresRevisionCatalogViews|TestBranchSearchPath|TestBranchCatalogIntrospection|TestBranchCatalogAlter|TestAsOf|TestDoltMergeData|TestPostgresMergeConflictStatus|TestBranchCatalogCreate|TestBranchCatalogDrop|TestBranchCatalogSameName|TestCatalogHistory|TestProllyVersionCompatibility|TestNativeVersionedQueries|TestNativeHistoricalCursors|TestRevisionWorkingSets|TestInfoSchemaTables|TestInfoSchemaPgCharMaxLength|TestWithStatements|TestWithRecursive)$'
# Preserve original tests; exclude whole scenarios requiring unimplemented DDL,
# detached-HEAD USE syntax, fixed OIDs or engine-specific system inventories.
"$compat_binary" -test.v -test.timeout=5m -test.run '^TestInfoSchemaRevisionDb$/^info_schema_changes_with_dolt_checkout$'
"$compat_binary" -test.v -test.timeout=5m -test.run '^TestPgIndex$/^pg_index_indoption_has_one_entry_per_key_column$'
"$compat_binary" -test.v -test.timeout=5m -test.run '^TestInfoSchemaColumns$/^(xml_columns|literal_column_defaults|cast_column_defaults)$'
# Relation-name and schema-path scenarios with portable result/error assertions.
"$compat_binary" -test.v -test.timeout=5m -test.run '^TestSchemas$/^(search_path_returns_user_table_before_public_table|public_schema_qualifier|public_schema_qualifier,_multiple_tables|public_db_and_schema_qualifier|insert,_update,_delete_with_schema)$'

# Native overload and SRF cases. Excluded scenarios retain their original
# assertions: numeric infinity/JSONB ordering and regex locale differ; error
# wording differs; partition ancestors requires unsupported foreign keys.
"$compat_binary" -test.v -test.timeout=5m -test.run '^TestOperators$/^(Addition|Subtraction|Multiplication|Division|Shift_Left|Shift_Right|Equal|Not_Equal_Standard_Syntax_.*|Not_Equal_Alternate_Syntax_.*|Bit_And|Bit_Or|Bit_Xor|Negate|Unary_Plus|Binary_JSON|Table_Columns|Concatenate|ARRAY|json_extract_text_operator_used_in_generated_column_default)$'
"$compat_binary" -test.v -test.timeout=5m -test.run '^TestSetReturningFunctions$/^(generate_series|aliased_SRF_alongside_scalar_columns_with_ORDER_BY|generate_series_as_table_function_with_column_alias|table_function_WITH_ORDINALITY|set-returning_function_as_join_operand|generate_series_as_table_function_used_in_pgJDBC-style_enum_catalog_query|limit,_offset,_sort|generate_series_with_table|generate_subscripts|generate_subscripts_as_table_function_with_scalar_alias|generate_subscripts_with_join|generate_subscripts_and_generate_series_combined|generate_subscripts_on_0-indexed_array_types|set_generation_with_other_func_calls|generate_series_as_table_function_and_projection|insert_with_set_returning_function)$'

# Native extended-protocol inference and procedural consumers, without requiring
# versioned CREATE FUNCTION in their setup.
"$compat_binary" -test.v -test.timeout=5m -test.run '^(TestBindingWithOidZero|TestBindingInsertSelectWithOidZero|TestBindingInt4ToBigintWithUntypedNull|TestBindingWhereClauseInt4ToBigintWithUntypedNull|TestBindingUpdateInt4ToBigintWithUntypedNull|TestBindingWithTextArray)$'

# Keep the native integer error wording and whole-record NULL behavior; the two
# Doltgres scenarios expecting different behavior remain outside this gate.
"$compat_binary" -test.v -test.timeout=5m -test.run '^TestPlpgsqlRecordInto$/^(RECORD_.*|variable_declaration_default_referencing_an_earlier_RECORD|SELECT_INTO_.*|INSERT_RETURNING_INTO_a_RECORD_variable|EXECUTE_INTO_a_RECORD_variable|trigger_function_using_SELECT_INTO_a_RECORD_variable|errors_accessing_RECORD_fields)$'
