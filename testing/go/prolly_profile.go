package _go

import "os"

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
		"TestDoltMerge/Merge_with_nested_parentheses_in_default,_generated,_and_check_expressions": "not implemented: schema capture with default/generated/check expressions",
		"TestMerge/merge_with_check_expressions_and_column_defaults":                               "not implemented: schema merges, ALTER constraints and default capture",
		"TestMerge/merge_with_unique_constraints_and_foreign_keys":                                 "not implemented: foreign keys and schema merges; native tests cover data unique collisions",
		"TestMerge/merge_a_branch_that_created_a_type":                                             "not implemented: branch-specific schemas and schema merge",

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

func prollyAssertionSkip(query string) string {
	if os.Getenv("DOLTGRES_TEST_PROFILE") != "prolly" {
		return ""
	}
	return map[string]string{
		"SELECT (-1)::bit, (-2)::bit, (-5)::bit(2), (-6::int4)::bit(2);":         "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"INSERT INTO t_char VALUES (6, 7);":                                      "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"INSERT INTO t_char VALUES (6, true);":                                   "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"SELECT true::\"char\";":                                                 "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"SELECT 100000::bigint::\"char\";":                                       "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"SELECT id, v1::int, v1::text FROM t_char WHERE id < 10;":                "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"INSERT INTO t_interval VALUES (3, 7);":                                  "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"INSERT INTO t_interval VALUES (3, true);":                               "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"SELECT v1::xid FROM t_name WHERE id=1;":                                 "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"INSERT INTO t_real VALUES (3, 1.0e100);":                                "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"INSERT INTO t_real VALUES (3, 1.0e100::numeric);":                       "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"SELECT ('1' || repeat('0', 320))::numeric::float8;":                     "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"SELECT 'doesnotexist.public.testing'::regclass;":                        "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"SELECT '\"Testing2\"'::regclass;":                                       "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"SELECT '\"integer\"'::regtype;":                                         "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"SELECT '\"integer'::regtype;":                                           "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"SELECT 'integer\"'::regtype;":                                           "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"SELECT '\"\"integer'::regtype;":                                         "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"SELECT NULL::int2[]::int2vector;":                                       "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"SELECT ARRAY[]::int2[]::int2vector;":                                    "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"SELECT (ARRAY[1::int2]::int2[])::int2vector;":                           "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"SELECT v::int2vector FROM t_int2_array;":                                "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"SELECT (ARRAY[1::int2] || ARRAY[2::int2])::int2vector;":                 "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"CREATE TABLE t_tsvector (id INTEGER primary key, v1 TSVECTOR);":         "PostgreSQL/prolly supports tsvector; Doltgres test expects unsupported",
		"SELECT '<a/>'::xml::int;":                                               "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"SELECT 1::xml;":                                                         "PostgreSQL built-in/cast/error behavior differs; reproduced with heap control",
		"INSERT INTO t (j) VALUES (5), (6), (7) RETURNING i, doesnotexist":       "PostgreSQL error wording differs (undefined column)",
		"INSERT INTO t (j) VALUES (5), (6), (7) RETURNING i, doesnotexist(j)":    "PostgreSQL error wording differs (undefined function)",
		"select 1 limit 18446744073709551615":                                    "PostgreSQL error wording differs (bigint out of range)",
		"select 1 limit -18446744073709551616":                                   "PostgreSQL error wording differs (bigint out of range)",
		"SELECT DISTINCT ON(v3) v1 FROM test2;":                                  "unordered DISTINCT ON does not specify which representative row wins",
		"SELECT relid FROM pg_partition_ancestors('ft');":                        "PostgreSQL returns no ancestors for an unpartitioned table; Doltgres returns itself",
		"SELECT pg_partition_ancestors.relid FROM pg_partition_ancestors('ft');": "PostgreSQL returns no ancestors for an unpartitioned table; Doltgres returns itself",
		"SELECT x FROM pg_partition_ancestors('ft') AS x;":                       "PostgreSQL returns no ancestors for an unpartitioned table; Doltgres returns itself",
		"SELECT relid FROM ft, pg_partition_ancestors('ft');":                    "PostgreSQL returns no ancestors for an unpartitioned table; Doltgres returns itself",
	}[query]
}
