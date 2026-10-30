// Copyright 2026 Dolthub, Inc.
// Licensed under the Apache License, Version 2.0.
package _go

import (
	"github.com/dolthub/go-mysql-server/sql"
	"testing"
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
			{Query: "SELECT n, 4 === 4, (ROW(7)::source::target).v FROM summary", Expected: []sql.Row{{6, true, 8}}},
			{Query: "PREPARE branch_operator AS SELECT 4 === 4", Expected: []sql.Row{}},
			{Query: "EXECUTE branch_operator", Expected: []sql.Row{{true}}},
			{Query: "SELECT status FROM dolt_checkout('feature')", Expected: []sql.Row{{int64(0)}}},
			{Query: "CREATE OR REPLACE FUNCTION eq_int(int,int) RETURNS boolean LANGUAGE sql IMMUTABLE AS 'SELECT $1<>$2'", Expected: []sql.Row{}},
			{Query: "EXECUTE branch_operator", Expected: []sql.Row{{false}}},
			{Query: "SELECT status FROM dolt_checkout('main')", Expected: []sql.Row{{int64(0)}}},
			{Query: "EXECUTE branch_operator", Expected: []sql.Row{{true}}},
			{Query: "DROP FUNCTION eq_int(int,int)", ExpectedErrCode: "2BP01"},
			{Query: "DROP AGGREGATE total(int)", ExpectedErrCode: "2BP01"},
		},
	}})
}
