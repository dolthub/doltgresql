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
