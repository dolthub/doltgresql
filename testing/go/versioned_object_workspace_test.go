// Copyright 2026 Dolthub, Inc.
// Licensed under the Apache License, Version 2.0.

package _go

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

// These cases intentionally create no seed table. Routines must be versioned
// from their first definition, including when the database has no row storage.
func TestVersionedRoutineWorkspace(t *testing.T) {
	if testTarget(t) != targetPostgresFork {
		t.Skip("native routine signature aliases and routine checkout/reset extensions")
	}
	RunScripts(t, []ScriptTest{
		{
			Name: "stage one overload and restore it independently",
			SetUpScript: []string{
				`CREATE FUNCTION f(i integer) RETURNS integer LANGUAGE sql AS $$SELECT i + 1$$`,
				`CREATE FUNCTION f(i text) RETURNS text LANGUAGE sql AS $$SELECT i || '!'$$`,
			},
			Assertions: []ScriptTestAssertion{
				{Query: `SELECT count(*) FROM dolt_status`, Expected: []sql.Row{{2}}},
				{Query: `SELECT dolt_add('f(integer)')`, Expected: []sql.Row{{0}}},
				{Query: `SELECT count(*) FROM dolt_status WHERE staged`, Expected: []sql.Row{{1}}},
				{Query: `SELECT count(*) FROM dolt_status WHERE NOT staged`, Expected: []sql.Row{{1}}},
				{Query: `SELECT dolt_commit('-m','one overload')`, SkipResultsCheck: true},
				{Query: `SELECT f(1), f('a'::text)`, Expected: []sql.Row{{2, "a!"}}},
				{Query: `CREATE OR REPLACE FUNCTION f(i integer) RETURNS integer LANGUAGE sql AS $$SELECT i + 10$$`},
				{Query: `SELECT dolt_checkout('f(integer)')`, SkipResultsCheck: true},
				{Query: `SELECT f(1), f('a'::text)`, Expected: []sql.Row{{2, "a!"}}},
				{Query: `SELECT dolt_clean('f(text)')`, Expected: []sql.Row{{0}}},
				{Query: `SELECT count(*) FROM dolt_status`, Expected: []sql.Row{{0}}},
				{Query: `SELECT f(1)`, Expected: []sql.Row{{2}}},
			},
		},
		{
			Name: "routine-only branches merge disjoint definitions",
			SetUpScript: []string{
				`CREATE FUNCTION base_fn() RETURNS integer LANGUAGE sql AS $$SELECT 1$$`,
				`SELECT dolt_commit('-Am','base')`,
				`SELECT dolt_checkout('-b','other')`,
				`CREATE FUNCTION other_fn() RETURNS integer LANGUAGE sql AS $$SELECT 2$$`,
				`SELECT dolt_commit('-Am','other')`,
				`SELECT dolt_checkout('main')`,
				`CREATE FUNCTION main_fn() RETURNS integer LANGUAGE sql AS $$SELECT 3$$`,
				`SELECT dolt_commit('-Am','main')`,
			},
			Assertions: []ScriptTestAssertion{
				{Query: `SELECT dolt_merge('other')`, SkipResultsCheck: true},
				{Query: `SELECT base_fn(), other_fn(), main_fn()`, Expected: []sql.Row{{1, 2, 3}}},
				{Query: `SELECT count(*) FROM dolt_status`, Expected: []sql.Row{{0}}},
				{Query: `SELECT dolt_rm('other_fn()')`, Expected: []sql.Row{{0}}},
				{Query: `SELECT count(*) FROM dolt_status WHERE staged AND status='deleted'`, Expected: []sql.Row{{1}}},
				{Query: `SELECT dolt_reset('other_fn()')`, Expected: []sql.Row{{0}}},
				{Query: `SELECT count(*) FROM dolt_status WHERE NOT staged AND status='deleted'`, Expected: []sql.Row{{1}}},
				{Query: `SELECT dolt_checkout('other_fn()')`, SkipResultsCheck: true},
				{Query: `SELECT other_fn()`, Expected: []sql.Row{{2}}},
			},
		},
	})
}

func TestVersionedRoutineDiffSummary(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "routine-only edits participate in summary and tracked commits",
			SetUpScript: []string{
				`CREATE FUNCTION f() RETURNS integer LANGUAGE sql AS $$SELECT 1$$`,
				`SELECT dolt_commit('-Am','routine base')`,
				`CREATE OR REPLACE FUNCTION f() RETURNS integer LANGUAGE sql AS $$SELECT 2$$`,
			},
			Assertions: []ScriptTestAssertion{
				{Query: `SELECT data_change, schema_change FROM dolt_diff_summary('HEAD','WORKING','f()')`, Expected: []sql.Row{{"t", "f"}}},
				{Query: `SELECT count(*) FROM dolt_diff_summary('HEAD','WORKING')`, Expected: []sql.Row{{1}}},
				{Query: `SELECT dolt_commit('-am','tracked routine')`, SkipResultsCheck: true},
				{Query: `SELECT count(*) FROM dolt_status`, Expected: []sql.Row{{0}}},
				{Query: `SELECT f()`, Expected: []sql.Row{{2}}},
			},
		},
	})
}

// Dolt counts each changed routine as one object, not its individual properties.
func TestVersionedObjectDiffStat(t *testing.T) {
	RunScripts(t, []ScriptTest{{
		Name: "routine additions modifications and deletions",
		SetUpScript: []string{
			"CREATE FUNCTION f() RETURNS integer LANGUAGE sql AS 'SELECT 1'",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SELECT table_name, rows_added, rows_deleted, rows_modified FROM dolt_diff_stat('HEAD','WORKING','f()')",
				Expected: []sql.Row{{"public.f()", 1, 0, 0}},
			},
			{Query: "SELECT dolt_commit('-Am','base')", SkipResultsCheck: true},
			{Query: "CREATE OR REPLACE FUNCTION f() RETURNS integer LANGUAGE sql AS 'SELECT 2'"},
			{
				Query:    "SELECT table_name, rows_added, rows_deleted, rows_modified FROM dolt_diff_stat('HEAD','WORKING')",
				Expected: []sql.Row{{"public.f()", 0, 0, 1}},
			},
			{
				Query:    "SELECT count(*) FROM dolt_diff_stat('HEAD','HEAD','f()')",
				Expected: []sql.Row{{0}},
			},
			{
				// Dolt omits object totals, so rows_unmodified becomes -1.
				// Preserve this interface exactly on both implementations.
				Query:    "SELECT rows_unmodified, old_row_count, new_row_count, cells_modified FROM dolt_diff_stat('HEAD','WORKING','f()')",
				Expected: []sql.Row{{-1, 0, 0, 0}},
			},
			{Query: "DROP FUNCTION f()"},
			{
				Query:    "SELECT table_name, rows_added, rows_deleted, rows_modified FROM dolt_diff_stat('HEAD','WORKING','f()')",
				Expected: []sql.Row{{"public.f()", 0, 1, 0}},
			},
			{
				Query:    "SELECT rows_unmodified, old_row_count, new_row_count FROM dolt_diff_stat('HEAD','WORKING','f()')",
				Expected: []sql.Row{{-1, 0, 0}},
			},
		},
	}})
}

// Doltgres rejects detailed routine dolt_diff calls. These assertions specify
// the fork's read-only canonical property view, not conflict-table DML.
func TestVersionedObjectPropertyDiff(t *testing.T) {
	if testTarget(t) != targetPostgresFork {
		t.Skip("canonical schema property diffs are a fork extension")
	}
	RunScripts(t, []ScriptTest{{
		Name: "routine properties across working staged and retained revisions",
		SetUpScript: []string{
			"CREATE FUNCTION f() RETURNS integer LANGUAGE sql AS 'SELECT 1'",
			"SELECT dolt_commit('-Am','base')",
			"CREATE OR REPLACE FUNCTION f() RETURNS integer LANGUAGE sql AS 'SELECT 2'",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SELECT from_property, from_value #>> '{}', to_value #>> '{}', diff_type FROM dolt_diff('HEAD','WORKING','f()')",
				Expected: []sql.Row{{"prosrc", "SELECT 1", "SELECT 2", "modified"}},
			},
			{
				Query:    "SELECT from_commit, to_commit, from_commit_date IS NOT NULL, to_commit_date IS NULL FROM dolt_diff('HEAD..WORKING','f()')",
				Expected: []sql.Row{{"HEAD", "WORKING", "t", "t"}},
			},
			{Query: "SELECT dolt_add('f()')", Expected: []sql.Row{{0}}},
			{Query: "CREATE OR REPLACE FUNCTION f() RETURNS integer LANGUAGE sql AS 'SELECT 3'"},
			{
				Query:    "SELECT from_value #>> '{}', to_value #>> '{}' FROM dolt_diff('STAGED','WORKING','f()')",
				Expected: []sql.Row{{"SELECT 2", "SELECT 3"}},
			},
			{Query: "SELECT dolt_commit('-Am','three')", SkipResultsCheck: true},
			{
				Query:    "SELECT from_value #>> '{}', to_value #>> '{}' FROM dolt_diff('HEAD~1','HEAD','f()')",
				Expected: []sql.Row{{"SELECT 1", "SELECT 3"}},
			},
			{Query: "DROP FUNCTION f()"},
			{
				Query:    "SELECT from_property, from_value #>> '{}', to_property IS NULL, diff_type FROM dolt_diff('HEAD','WORKING','f()') WHERE from_property='prosrc'",
				Expected: []sql.Row{{"prosrc", "SELECT 3", "t", "removed"}},
			},
			{
				Query:    "SELECT to_property, to_value #>> '{}', from_property IS NULL, diff_type FROM dolt_diff('WORKING','HEAD','f()') WHERE to_property='prosrc'",
				Expected: []sql.Row{{"prosrc", "SELECT 3", "t", "added"}},
			},
		},
	}})
}
