// Copyright 2026 Dolthub, Inc.
// Licensed under the Apache License, Version 2.0.
package _go

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

// Native DDL compatibility is checked by TestCreateAggregate, TestCreateOperator
// and TestCasts. These assertions cover the fork's catalog publication contract.
func TestPostgresForkVersionedSchemaObjects(t *testing.T) {
	if testTarget(t) != targetPostgresFork {
		t.Skip("native branch catalog integration")
	}
	RunScripts(t, []ScriptTest{{
		Name: "schema objects merge without leaking definitions between branches",
		SetUpScript: []string{
			"CREATE TABLE t(v int)", "INSERT INTO t VALUES (1),(2),(3)",
			"SELECT dolt_commit('-Am','base')", "SELECT dolt_branch('feature')",
			"SELECT dolt_checkout('feature')",
			"CREATE AGGREGATE total(int) (SFUNC=int4pl, STYPE=int4, INITCOND='0')",
			"CREATE FUNCTION eq_int(int,int) RETURNS boolean LANGUAGE sql IMMUTABLE AS 'SELECT $1=$2'",
			"CREATE OPERATOR === (LEFTARG=int4, RIGHTARG=int4, FUNCTION=eq_int)",
			"CREATE TYPE source AS (v int)", "CREATE TYPE target AS (v int)",
			"CREATE FUNCTION cv(source) RETURNS target LANGUAGE sql IMMUTABLE AS 'SELECT ROW(($1).v+1)::target'",
			"CREATE CAST (source AS target) WITH FUNCTION cv(source)",
			"CREATE VIEW summary AS SELECT total(v) AS n FROM t",
			"SELECT dolt_commit('-Am','objects')", "SELECT dolt_checkout('main')",
			"CREATE TABLE separate(v int)", "SELECT dolt_commit('-Am','diverge')",
		},
		Assertions: []ScriptTestAssertion{
			{Query: "SELECT count(*) FROM pg_proc WHERE proname='total'", Expected: []sql.Row{{int64(0)}}},
			{Query: "SELECT count(*) FROM pg_operator WHERE oprname='==='", Expected: []sql.Row{{int64(0)}}},
			{Query: "SELECT conflicts FROM dolt_merge('feature')", Expected: []sql.Row{{int64(0)}}},
			{Query: "SELECT n, 4 === 4, (ROW(7)::source::target).v FROM summary", Expected: []sql.Row{{6, "t", 8}}},
			{Query: "PREPARE branch_operator AS SELECT 4 === 4", Expected: []sql.Row{}},
			{Query: "EXECUTE branch_operator", Expected: []sql.Row{{"t"}}},
			{Query: "SELECT status FROM dolt_checkout('feature')", Expected: []sql.Row{{int64(0)}}},
			{Query: "CREATE OR REPLACE FUNCTION eq_int(int,int) RETURNS boolean LANGUAGE sql IMMUTABLE AS 'SELECT $1<>$2'", Expected: []sql.Row{}},
			{Query: "EXECUTE branch_operator", Expected: []sql.Row{{"f"}}},
			{Query: "SELECT status FROM dolt_checkout('main')", Expected: []sql.Row{{int64(0)}}},
			{Query: "EXECUTE branch_operator", Expected: []sql.Row{{"t"}}},
			{Query: "DROP FUNCTION eq_int(int,int)", ExpectedErrCode: "2BP01"},
			{Query: "DROP AGGREGATE total(int)", ExpectedErrCode: "2BP01"},
		},
	}})
}

func TestPostgresForkSchemaObjectWorkspace(t *testing.T) {
	if testTarget(t) != targetPostgresFork {
		t.Skip("native aggregate, operator and cast versioning")
	}
	RunScripts(t, []ScriptTest{{
		Name: "stage schema objects by the identities reported in status",
		SetUpScript: []string{
			"CREATE TYPE source AS (v int)", "CREATE TYPE target AS (v int)",
			"CREATE FUNCTION cv(source) RETURNS target LANGUAGE sql IMMUTABLE AS 'SELECT ROW(($1).v)::target'",
			"SELECT dolt_commit('-Am','base')",
			"CREATE AGGREGATE total(int) (SFUNC=int4pl, STYPE=int4, INITCOND='0')",
			"CREATE OPERATOR === (LEFTARG=int4, RIGHTARG=int4, FUNCTION=int4eq)",
			"CREATE CAST (source AS target) WITH FUNCTION cv(source)",
		},
		Assertions: []ScriptTestAssertion{
			{Query: "SELECT count(*) FROM dolt_status WHERE NOT staged", Expected: []sql.Row{{3}}},
			{Query: "SELECT dolt_add('public.total(int)','public.===(pg_catalog.int4,pg_catalog.int4)','(public.source AS public.target)')", Expected: []sql.Row{{0}}},
			{Query: "SELECT count(*) FROM dolt_status WHERE staged", Expected: []sql.Row{{3}}},
			{Query: "SELECT dolt_commit('-m','schema objects')", SkipResultsCheck: true},
			{Query: "SELECT count(*) FROM dolt_diff_summary('HEAD~1','HEAD')", Expected: []sql.Row{{3}}},
			{Query: "SELECT count(*) FROM dolt_diff_stat('HEAD~1','HEAD')", Expected: []sql.Row{{3}}},
			{Query: "SELECT count(*) FROM dolt_status", Expected: []sql.Row{{0}}},
		},
	}})
}
