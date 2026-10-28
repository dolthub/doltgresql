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
