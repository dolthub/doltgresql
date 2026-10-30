// Copyright 2026 Dolthub, Inc.
// Licensed under the Apache License, Version 2.0.
package _go

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestPostgresInheritanceDDL(t *testing.T) {
	if testTarget(t) != targetPostgres && testTarget(t) != targetPostgresFork {
		t.Skip("native PostgreSQL recursive inheritance semantics")
	}
	RunScripts(t, []ScriptTest{{
		Name: "recursive DDL preserves local child columns and constraints",
		SetUpScript: []string{
			"CREATE TABLE parent(a int CONSTRAINT positive CHECK(a>0))",
			"CREATE TABLE child(b text) INHERITS(parent)",
			"CREATE TABLE grandchild(c int) INHERITS(child)",
			"INSERT INTO parent VALUES(1)",
			"INSERT INTO child VALUES(2,'child')",
			"INSERT INTO grandchild VALUES(3,'grandchild',4)",
			"CREATE INDEX child_a ON child(a)",
		},
		Assertions: []ScriptTestAssertion{
			{Query: "SELECT a FROM parent ORDER BY a", Expected: []sql.Row{{1}, {2}, {3}}},
			{Query: "SELECT a FROM ONLY parent", Expected: []sql.Row{{1}}},
			{Query: "ALTER TABLE child DROP CONSTRAINT positive", ExpectedErrCode: "42P16"},
			{Query: "ALTER TABLE parent ADD COLUMN n int DEFAULT 7", Expected: []sql.Row{}},
			{Query: "SELECT a,b,c,n FROM grandchild", Expected: []sql.Row{{3, "grandchild", 4, 7}}},
			{Query: "ALTER TABLE parent ALTER a TYPE bigint USING a+10", Expected: []sql.Row{}},
			{Query: "SELECT a,n FROM parent ORDER BY a", Expected: []sql.Row{{11, 7}, {12, 7}, {13, 7}}},
			{Query: "ALTER TABLE parent RENAME a TO id", Expected: []sql.Row{}},
			{Query: "SELECT id,b,c FROM grandchild", Expected: []sql.Row{{13, "grandchild", 4}}},
			{Query: "CREATE VIEW inherited_view AS SELECT id FROM grandchild", Expected: []sql.Row{}},
			{Query: "ALTER TABLE parent DROP COLUMN id CASCADE", Expected: []sql.Row{}},
			{Query: "SELECT b,c,n FROM grandchild", Expected: []sql.Row{{"grandchild", 4, 7}}},
			{Query: "SELECT count(*) FROM pg_class WHERE relname='inherited_view'", Expected: []sql.Row{{0}}},
		},
	}})
}

func TestPostgresForkVersionedInheritance(t *testing.T) {
	if testTarget(t) != targetPostgresFork {
		t.Skip("branch-aware inherited storage")
	}
	RunScripts(t, []ScriptTest{{
		Name: "merge inherited definitions and data while retaining history",
		SetUpScript: []string{
			"CREATE TABLE parent(a int)", "INSERT INTO parent VALUES(1)",
			"SELECT dolt_commit('-Am','base')", "SELECT dolt_branch('feature')",
			"SELECT dolt_checkout('feature')", "CREATE TABLE child(b text) INHERITS(parent)",
			"INSERT INTO child VALUES(2,'feature')", "SELECT dolt_commit('-Am','child')",
			"SELECT dolt_checkout('main')", "CREATE TABLE independent(v int)",
			"SELECT dolt_commit('-Am','diverge')",
		},
		Assertions: []ScriptTestAssertion{
			{Query: "SELECT a FROM parent", Expected: []sql.Row{{1}}},
			{Query: "SELECT conflicts FROM dolt_merge('feature')", Expected: []sql.Row{{0}}},
			{Query: "SELECT a FROM parent ORDER BY a", Expected: []sql.Row{{1}, {2}}},
			{Query: "ALTER TABLE parent ADD COLUMN n int DEFAULT 9", Expected: []sql.Row{}},
			{Query: "SELECT a,b,n FROM child", Expected: []sql.Row{{2, "feature", 9}}},
			{Query: "SELECT p.a,c.n FROM parent AS OF 'feature' p CROSS JOIN child c ORDER BY 1", Expected: []sql.Row{{1, 9}, {2, 9}}},
			{Query: "SELECT a,b FROM child AS OF 'feature'", Expected: []sql.Row{{2, "feature"}}},
		},
	}})
}
