package _go

import (
	"os"
	"strings"
)

// Explicit, opt-in exclusions; never skip a query because it failed. Keep the
// ordinary Doltgres suite unchanged. Entries distinguish missing AM features
// from expected PostgreSQL behavior that differs from Doltgres's oracle.
func prollyScriptSkip(name string) string {
	if os.Getenv("DOLTGRES_TEST_PROFILE") != "prolly" {
		return ""
	}
	return map[string]string{
		"TestDoltMerge/Merge_without_conflicts": "mixed schema slice: array PK, sequences, generated columns, triggers and routines; isolated data equivalents in TestDoltMergeData",
		"TestDoltMerge/Fast_forward":            "mixed schema slice: array PK, defaults, FK, generated columns and routines; isolated data equivalent in TestDoltMergeData",
		"TestDoltMerge/--no-ff":                 "mixed schema slice: array PK, defaults, FK, generated columns and routines; isolated data equivalent in TestDoltMergeData",
		"TestDoltMerge/Merge_with_nested_parentheses_in_default,_generated,_and_check_expressions":    "not implemented: schema capture with default/generated/check expressions",
		"TestMerge/merge_with_check_expressions_and_column_defaults":                                  "not implemented: schema merges, ALTER constraints and default capture",
		"TestMerge/merge_with_unique_constraints_and_foreign_keys":                                    "not implemented: foreign keys and schema merges; native tests cover data unique collisions",
		"TestMerge/merge_a_branch_that_created_a_type":                                                "not implemented: branch-specific schemas and schema merge",
		"TestDoltDiff/Single_commit":                                                                  "mixed schema slice: array PK, serial, defaults, FK, generated columns, triggers and routines; see TestPostgresDoltDiffFunctions",
		"TestDoltDiffStat/Single_commit":                                                              "mixed schema slice: array PK, serial, defaults, FK, generated columns, triggers and routines; see TestPostgresDoltDiffFunctions",
		"TestDoltDiffSummary/Single_commit":                                                           "mixed schema slice: array PK, serial, defaults, FK, generated columns, triggers and routines; see TestPostgresDoltDiffFunctions",
		"TestDoltFunctionSmokeTests/smoke_test_select_dolt_merge":                                     "not implemented: schema merges (the merged branch creates a table)",
		"TestDoltFunctionSmokeTests/smoke_test_select_dolt_merge_dirty_working_set,_different_tables": "not implemented: schema merges (the merged branch creates a table)",
		"TestDoltAdd/Add_all_using_dot":                                                               "mixed schema slice: array PK, serial, defaults, FK, generated columns, triggers and routines; see TestPostgresWorkspaceCommands",
		"TestDoltAdd/Add_all_using_-A":                                                                "mixed schema slice: array PK, serial, defaults, FK, generated columns, triggers and routines; see TestPostgresWorkspaceCommands",
		"TestDoltAdd/Add_all_individually":                                                            "mixed schema slice: array PK, serial, defaults, FK, generated columns, triggers and routines; see TestPostgresWorkspaceCommands",
		"TestDoltBranch/All_branch_options":                                                           "mixed schema slice: array PK, serial, defaults, FK, generated columns, triggers and routines; see TestPostgresWorkspaceCommands",
		"TestDoltCheckout/All_checkout_options":                                                       "mixed schema slice: array PK, serial, defaults, FK, generated columns, triggers and routines; see TestPostgresWorkspaceCommands",
		"TestDoltCherryPick/Smoke_test":                                                               "mixed schema slice: array PK, serial, defaults, FK, generated columns, triggers and routines; see TestPostgresWorkspaceCommands",
		"TestDoltClean/Clean_all":                                                                     "mixed schema slice: array PK, serial, defaults, FK, generated columns, triggers and routines; see TestPostgresWorkspaceCommands",
		"TestDoltClean/Clean_all_except_one":                                                          "mixed schema slice: array PK, serial, defaults, FK, generated columns, triggers and routines; see TestPostgresWorkspaceCommands",
		"TestDoltClean/Clean_all_by_name":                                                             "mixed schema slice: array PK, serial, defaults, FK, generated columns, triggers and routines; see TestPostgresWorkspaceCommands",
		"TestDoltClean/Dry_run":                                                                       "mixed schema slice: array PK, serial, defaults, FK, generated columns, triggers and routines; see TestPostgresWorkspaceCommands",
		"TestDoltReset/Hard":                                                                          "mixed schema slice: array PK, serial, defaults, FK, generated columns, triggers and routines; see TestPostgresWorkspaceCommands",
		"TestDoltReset/Soft":                                                                          "mixed schema slice: array PK, serial, defaults, FK, generated columns, triggers and routines; see TestPostgresWorkspaceCommands",
		"TestDoltRevert/Revert_a_single_commit":                                                       "mixed schema slice: array PK, serial, defaults, FK, generated columns, triggers and routines; see TestPostgresWorkspaceCommands",
		"TestDoltRM/Cached_only":                                                                      "mixed schema slice: array PK, serial, defaults, FK, generated columns, triggers and routines; see TestPostgresWorkspaceCommands",
		"TestDoltConflictsResolve/Simple_conflicts_resolve_all_using_--ours":                          "approved contract: version-control commands are standalone statements outside transaction blocks; see TestPostgresMergeConflictStatus",
		"TestDoltConflictsResolve/Simple_conflicts_resolve_all_using_--theirs":                        "approved contract: version-control commands are standalone statements outside transaction blocks; see TestPostgresMergeConflictStatus",
		"TestDoltConflictsResolve/Simple_conflicts_resolve_table_using_--ours":                        "approved contract: version-control commands are standalone statements outside transaction blocks; see TestPostgresMergeConflictStatus",
		"TestDoltPreviewMergeConflicts/Preview_conflicts":                                             "mixed schema slice: array PK, serial, generated columns, triggers and routines; see TestPostgresDoltTables and the preview smoke tests",
		"TestUserSpaceDoltTables/dolt_conflicts_with_tablename":                                       "approved contract: version-control commands are standalone statements outside transaction blocks; see TestPostgresDoltTables",
		"TestUserSpaceDoltTables/dolt_constraint_violations":                                          "approved contract: version-control commands are standalone statements outside transaction blocks; see TestPostgresDoltTables",
		"TestUserSpaceDoltTables/dolt_constraint_violations_with_tablename":                           "approved contract: version-control commands are standalone statements outside transaction blocks; see TestPostgresDoltTables",
		"TestUserSpaceDoltTables/dolt_schema_conflicts":                                               "not implemented: schema merges (ALTER TABLE on branches and merging schema changes)",
		"TestUserSpaceDoltTables/dolt_conflicts":                                                      "approved contract: version-control commands are standalone statements outside transaction blocks; see TestPostgresDoltTables",

		"TestStatsAggregates/aggregate_functions_over_dolt_statistics":                      "not implemented: the prolly access method rejects ANALYZE, so dolt_statistics has no rows for versioned tables",
		"TestStatsUsage/dolt_statistics_contains_reasonable_values_after_ANALYZE":           "not implemented: the prolly access method rejects ANALYZE, so dolt_statistics has no rows for versioned tables",
		"TestStatsUsage/ANALYZE_refreshes_statistics_after_data_changes":                    "not implemented: the prolly access method rejects ANALYZE, so dolt_statistics has no rows for versioned tables",
		"TestStatsUsage/join_planner_puts_small_table_on_the_scan_side_of_a_lookup_join":    "not implemented: the prolly access method rejects ANALYZE; PostgreSQL plans are not Doltgres plans",
		"TestStatsUsage/join_planner_builds_hash_table_from_the_small_table_in_a_hash_join": "not implemented: the prolly access method rejects ANALYZE; PostgreSQL plans are not Doltgres plans",
		"TestStatsUsage/planner_uses_cardinality_to_select_more_selective_index":            "not implemented: the prolly access method rejects ANALYZE; PostgreSQL plans are not Doltgres plans",
		"TestStatsUsage/join_planner_uses_histograms_to_reorder_a_filtered_lookup_join":     "not implemented: the prolly access method rejects ANALYZE; PostgreSQL plans are not Doltgres plans",
		"TestTypes/Bit_key": "not implemented: index operator class for this type",
		"TestTypes/Character_comparisons_ignore_trailing_spaces":                           "not implemented: index operator class for this type",
		"TestTypes/Character_key":                                                          "not implemented: index operator class for this type",
		"TestTypes/Character_varying_key_with_mixed_short_and_long_values":                 "not implemented: index operator class for this type",
		"TestTypes/Character_varying_type,_no_length,_as_primary_key":                      "not implemented: index operator class for this type",
		"TestTypes/Character_varying_type_as_primary_key":                                  "not implemented: index operator class for this type",
		"TestTypes/Date_key":                                                               "not implemented: index operator class for this type",
		"TestTypes/Double_precision_key":                                                   "not implemented: index operator class for this type",
		"TestTypes/Interval_key":                                                           "not implemented: index operator class for this type",
		"TestTypes/JSON_key":                                                               "not implemented: index operator class for this type",
		"TestTypes/Name_key":                                                               "not implemented: index operator class for this type",
		"TestTypes/Numeric_key":                                                            "not implemented: index operator class for this type",
		"TestTypes/Oid_array_type":                                                         "intentional portability restriction: persisted OID/XID identifiers",
		"TestTypes/Oid_type":                                                               "intentional portability restriction: persisted OID/XID identifiers",
		"TestTypes/Oid_type,_explicit_casts":                                               "intentional portability restriction: persisted OID/XID identifiers",
		"TestTypes/Oidvector_array_type":                                                   "intentional portability restriction: persisted OID/XID identifiers",
		"TestTypes/Oidvector_type":                                                         "intentional portability restriction: persisted OID/XID identifiers",
		"TestTypes/Real_key":                                                               "not implemented: index operator class for this type",
		"TestTypes/Time_without_time_zone_key":                                             "not implemented: index operator class for this type",
		"TestTypes/Uuid_key":                                                               "not implemented: index operator class for this type",
		"TestTypes/Xid_array_type":                                                         "intentional portability restriction: persisted OID/XID identifiers",
		"TestTypes/Xid_type":                                                               "intentional portability restriction: persisted OID/XID identifiers",
		"TestTypes/Xid_type,_explicit_casts":                                               "intentional portability restriction: persisted OID/XID identifiers",
		"TestInsert/on_conflict_clause":                                                    "not implemented: speculative insertion / ON CONFLICT",
		"TestInsert/conditional_on_conflict_update":                                        "not implemented: speculative insertion / ON CONFLICT",
		"TestInsert/on_conflict_update_returning_with_check_constraint":                    "not implemented: speculative insertion / ON CONFLICT",
		"TestInsert/on_conflict_do_nothing_only_ignores_uniqueness_conflicts":              "not implemented: foreign keys / ALTER TABLE",
		"TestInsert/insert_on_conflict_do_nothing_returning":                               "not implemented: speculative insertion / ON CONFLICT",
		"TestInsert/default_values_compatibility":                                          "not implemented: temporary tables",
		"TestUpdate/update_keeps_GENERATED_ALWAYS_AS_IDENTITY_values":                      "not implemented: foreign keys / ALTER TABLE",
		"TestUpdate/UPDATE_with_join_on_table_with_trigger":                                "not implemented: foreign keys / ALTER TABLE",
		"TestUpdateAssignmentSemantics/assignments_through_foreign_key_and_check_handlers": "not implemented: foreign keys / ALTER TABLE",
	}[name]
}

// prollyAssertionSkip matches either the bare query (any scenario) or the
// full subtest name of one assertion (scenario-specific), ignoring the #NN
// suffix Go adds to repeated subtest names.
func prollyAssertionSkip(fullName, query string) string {
	if os.Getenv("DOLTGRES_TEST_PROFILE") != "prolly" {
		return ""
	}
	if i := strings.LastIndex(fullName, "#"); i > 0 && strings.Trim(fullName[i+1:], "0123456789") == "" {
		fullName = fullName[:i]
	}
	if reason := prollyScopedAssertionSkip[fullName]; reason != "" {
		return reason
	}
	return map[string]string{
		"SELECT (-1)::bit, (-2)::bit, (-5)::bit(2), (-6::int4)::bit(2);":                                "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"INSERT INTO t_char VALUES (6, 7);":                                                             "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"INSERT INTO t_char VALUES (6, true);":                                                          "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"SELECT true::\"char\";":                                                                        "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"SELECT 100000::bigint::\"char\";":                                                              "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"SELECT id, v1::int, v1::text FROM t_char WHERE id < 10;":                                       "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"INSERT INTO t_interval VALUES (3, 7);":                                                         "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"INSERT INTO t_interval VALUES (3, true);":                                                      "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"SELECT v1::xid FROM t_name WHERE id=1;":                                                        "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"INSERT INTO t_real VALUES (3, 1.0e100);":                                                       "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"INSERT INTO t_real VALUES (3, 1.0e100::numeric);":                                              "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"SELECT ('1' || repeat('0', 320))::numeric::float8;":                                            "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"SELECT 'doesnotexist.public.testing'::regclass;":                                               "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"SELECT '\"Testing2\"'::regclass;":                                                              "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"SELECT '\"integer\"'::regtype;":                                                                "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"SELECT '\"integer'::regtype;":                                                                  "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"SELECT 'integer\"'::regtype;":                                                                  "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"SELECT '\"\"integer'::regtype;":                                                                "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"SELECT NULL::int2[]::int2vector;":                                                              "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"SELECT ARRAY[]::int2[]::int2vector;":                                                           "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"SELECT (ARRAY[1::int2]::int2[])::int2vector;":                                                  "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"SELECT v::int2vector FROM t_int2_array;":                                                       "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"SELECT (ARRAY[1::int2] || ARRAY[2::int2])::int2vector;":                                        "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"CREATE TABLE t_tsvector (id INTEGER primary key, v1 TSVECTOR);":                                "PostgreSQL/prolly supports tsvector; Doltgres test expects unsupported",
		"SELECT '<a/>'::xml::int;":                                                                      "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"SELECT 1::xml;":                                                                                "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"INSERT INTO t (j) VALUES (5), (6), (7) RETURNING i, doesnotexist":                              "PostgreSQL error wording differs (undefined column)",
		"INSERT INTO t (j) VALUES (5), (6), (7) RETURNING i, doesnotexist(j)":                           "PostgreSQL error wording differs (undefined function)",
		"select 1 limit 18446744073709551615":                                                           "PostgreSQL error wording differs (bigint out of range)",
		"select 1 limit -18446744073709551616":                                                          "PostgreSQL error wording differs (bigint out of range)",
		"SELECT DISTINCT ON(v3) v1 FROM test2;":                                                         "unordered DISTINCT ON does not specify which representative row wins",
		"SELECT relid FROM pg_partition_ancestors('ft');":                                               "PostgreSQL returns no ancestors for an unpartitioned table; Doltgres returns itself",
		"SELECT pg_partition_ancestors.relid FROM pg_partition_ancestors('ft');":                        "PostgreSQL returns no ancestors for an unpartitioned table; Doltgres returns itself",
		"SELECT x FROM pg_partition_ancestors('ft') AS x;":                                              "PostgreSQL returns no ancestors for an unpartitioned table; Doltgres returns itself",
		"SELECT relid FROM ft, pg_partition_ancestors('ft');":                                           "PostgreSQL returns no ancestors for an unpartitioned table; Doltgres returns itself",
		"EXPLAIN SELECT to_id FROM dolt_diff_test WHERE to_commit = '0123456789abcdefghij0123456789ab'": "engine-specific plan output; PostgreSQL evaluates commit filters against the generated table function",
		"EXPLAIN SELECT to_id FROM dolt_diff_test WHERE from_commit = '0123456789abcdefghij0123456789ab' AND to_commit = 'ab0123456789abcdefghij0123456789'": "engine-specific plan output; PostgreSQL evaluates commit filters against the generated table function",
		"EXPLAIN SELECT to_id FROM dolt_diff_test WHERE to_commit < '0123456789abcdefghij0123456789ab'":                                                      "engine-specific plan output; PostgreSQL evaluates commit filters against the generated table function",
		"EXPLAIN SELECT to_id FROM dolt_diff_test WHERE to_commit >= '0123456789abcdefghij0123456789ab'":                                                     "engine-specific plan output; PostgreSQL evaluates commit filters against the generated table function",
		"SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS_SUMMARY('main', 'branch1') ORDER BY 'table'":                                                             "PostgreSQL rejects a string constant in ORDER BY; the unordered form is checked by the other assertions",
		"DESCRIBE dolt.\"status\"":                                                               "DESCRIBE is MySQL syntax; PostgreSQL uses catalog queries",
		"DESCRIBE dolt.status":                                                                   "DESCRIBE is MySQL syntax; PostgreSQL uses catalog queries",
		"DESCRIBE dolt_status":                                                                   "DESCRIBE is MySQL syntax; PostgreSQL uses catalog queries",
		"SELECT count(*) FROM dolt.commits":                                                      "bootstrap history: the fork records no CREATE DATABASE commit; see TestPostgresDoltTables",
		"SELECT count(*) FROM dolt_commits":                                                      "bootstrap history: the fork records no CREATE DATABASE commit; see TestPostgresDoltTables",
		"SELECT count(*) FROM commits":                                                           "bootstrap history: the fork records no CREATE DATABASE commit; see TestPostgresDoltTables",
		"SELECT dolt.commits.message FROM dolt.commits":                                          "bootstrap history: the fork records no CREATE DATABASE commit; see TestPostgresDoltTables",
		"SELECT dolt_commits.message FROM dolt_commits":                                          "bootstrap history: the fork records no CREATE DATABASE commit; see TestPostgresDoltTables",
		"SELECT count(*) FROM dolt.commit_ancestors":                                             "bootstrap history: the fork records no CREATE DATABASE commit; see TestPostgresDoltTables",
		"SELECT count(*) FROM dolt_commit_ancestors":                                             "bootstrap history: the fork records no CREATE DATABASE commit; see TestPostgresDoltTables",
		"SELECT count(*) FROM commit_ancestors":                                                  "bootstrap history: the fork records no CREATE DATABASE commit; see TestPostgresDoltTables",
		"SELECT dolt.commit_ancestors.parent_index FROM dolt.commit_ancestors":                   "bootstrap history: the fork records no CREATE DATABASE commit; see TestPostgresDoltTables",
		"SELECT dolt_commit_ancestors.parent_index FROM dolt_commit_ancestors":                   "bootstrap history: the fork records no CREATE DATABASE commit; see TestPostgresDoltTables",
		"SELECT count(*) FROM dolt.log":                                                          "bootstrap history: the fork records no CREATE DATABASE commit; see TestPostgresDoltTables",
		"SELECT count(*) FROM dolt_log":                                                          "bootstrap history: the fork records no CREATE DATABASE commit; see TestPostgresDoltTables",
		"SELECT COUNT(*) FROM DOLT_LOG('main');":                                                 "bootstrap history: the fork records no CREATE DATABASE commit; see TestPostgresWorkspaceCommands",
		"SELECT COUNT(*) FROM DOLT_LOG('original');":                                             "bootstrap history: the fork records no CREATE DATABASE commit; see TestPostgresWorkspaceCommands",
		"SELECT count(*) FROM log":                                                               "bootstrap history: the fork records no CREATE DATABASE commit; see TestPostgresDoltTables",
		"SELECT * FROM dolt_diff_summary('main', 'WORKING', 'docs')":                             "not implemented: dolt_docs is database-local configuration, not a versioned table",
		"SELECT diff_type, from_doc_name, to_doc_name FROM dolt_diff('main', 'WORKING', 'docs')": "not implemented: dolt_docs is database-local configuration, not a versioned table",
	}[query]
}

var prollyScopedAssertionSkip = map[string]string{
	"TestUserSpaceDoltTables/dolt_rebase/select_message_from_dolt_log;":                                                                                                                                                                     "repository history starts at dolt_init; there is no CREATE DATABASE commit",
	"TestDoltFunctionSmokeTests/smoke_test_select_dolt_add_and_dolt_commit/select_count(*)_from_dolt.log":                                                                                                                                   "repository history starts at dolt_init; there is no CREATE DATABASE commit",
	"TestUserSpaceDoltTables/dolt_blame_with_tablename/SELECT_schemaname,_viewname_FROM_pg_views_WHERE_viewname_LIKE_'dolt_blame%'_ORDER_BY_1,_2":                                                                                           "not implemented: generated dolt_*_<table> tables are parser rewrites, not catalog relations",
	"TestUserSpaceDoltTables/dolt_blame_with_tablename/SELECT_relname,_relkind_FROM_pg_class_WHERE_relname_LIKE_'dolt_blame%'_ORDER_BY_1":                                                                                                   "not implemented: generated dolt_*_<table> tables are parser rewrites, not catalog relations",
	"TestUserSpaceDoltTables/dolt_blame_with_tablename/SELECT_table_schema,_table_name,_table_type_FROM_information_schema.tables_WHERE_table_name_LIKE_'dolt_blame%'_ORDER_BY_1,_2":                                                        "not implemented: generated dolt_*_<table> tables are parser rewrites, not catalog relations",
	"TestUserSpaceDoltTables/dolt_blame_catalog_columns_outside_search_path/SELECT_a.attname_______FROM_pg_attribute_a_______JOIN_pg_class_c_ON_a.attrelid_=_c.oid_______WHERE_c.relname_=_'dolt_blame_remote'________AND_a.attname_=_'id'": "not implemented: generated dolt_*_<table> tables are parser rewrites, not catalog relations",
	"TestUserSpaceDoltTables/dolt_docs/SELECT_*_FROM_dolt_diff_summary('main',_'WORKING')":                                                                                                                                                  "not implemented: dolt_docs is database-local configuration, not a versioned table",
	"TestUserSpaceDoltTables/dolt_schemas/SELECT_*_FROM_dolt_diff_summary('main',_'WORKING')":                                                                                                                                               "not implemented: views are not captured by version control; dolt_schemas is generated from the catalog",
	"TestUserSpaceDoltTables/dolt_schemas/SELECT_*_FROM_dolt_diff_summary('main',_'WORKING',_'dolt_schemas')":                                                                                                                               "not implemented: views are not captured by version control; dolt_schemas is generated from the catalog",
	"TestUserSpaceDoltTables/dolt_schemas/SELECT_diff_type,_from_name,_to_name_FROM_dolt_diff('main',_'WORKING',_'dolt_schemas')":                                                                                                           "not implemented: views are not captured by version control; dolt_schemas is generated from the catalog",
	"TestUserSpaceDoltTables/dolt_schemas/SELECT_table_schema,_table_name_FROM_information_schema.views":                                                                                                                                    "PostgreSQL lists its catalog views in information_schema.views",
	"TestUserSpaceDoltTables/dolt_schemas/DROP_VIEW_myView":                                                                                                                                                                                 "PostgreSQL error wording differs (undefined view)",
	"TestUserSpaceDoltTables/dolt_schemas/SELECT_name,_fragment_FROM_dolt_schemas":                                                                                                                                                          "PostgreSQL stores a view's parsed definition, not its original text",
	"TestUserSpaceDoltTables/dolt_schemas/SELECT_name,_fragment_FROM_newschema.dolt_schemas":                                                                                                                                                "PostgreSQL stores a view's parsed definition, not its original text",
	"TestCreateViewStatements/Views_that_reference_each_other/SELECT_*_FROM_cyc_v2;":                                                                                                                                                        "PostgreSQL error wording differs (view recursion)",
	"TestCreateViewStatements/Views_that_reference_each_other/SELECT_*_FROM_cyc_v3;":                                                                                                                                                        "PostgreSQL error wording differs (view recursion)",
	"TestCreateViewStatements/Views_that_reference_each_other/SELECT_*_FROM_(SELECT_*_FROM_cyc_v3)_sq;":                                                                                                                                     "PostgreSQL error wording differs (view recursion)",
	"TestCreateViewStatements/Views_that_reference_each_other/SELECT_*_FROM_cyc_self;":                                                                                                                                                      "PostgreSQL error wording differs (view recursion)",
}
