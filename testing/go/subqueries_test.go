package _go

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestSubqueryJoins(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "subquery join",
			SetUpScript: []string{
				"CREATE TABLE t1 (a int primary key);",
				"CREATE TABLE t2 (b int primary key);",
				"INSERT INTO t1 VALUES (1), (2), (3);",
				"INSERT INTO t2 VALUES (2), (3), (4);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT
s1.a FROM (SELECT a from t1) s1
INNER JOIN t2 q1
ON q1.b = s1.a
ORDER BY 1;`,
					Expected: []sql.Row{
						{2},
						{3},
					},
				},
			},
		},
		{
			Name: "subquery join with aliased column",
			SetUpScript: []string{
				"CREATE TABLE t1 (a int primary key);",
				"CREATE TABLE t2 (b int primary key);",
				"INSERT INTO t1 VALUES (1), (2), (3);",
				"INSERT INTO t2 VALUES (2), (3), (4);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT
s1.c FROM (SELECT a as c from t1) s1
INNER JOIN t2 q1
ON q1.b = s1.c
ORDER BY 1;`,
					Expected: []sql.Row{
						{2},
						{3},
					},
				},
			},
		},
		{
			Name: "subquery join with column renames",
			SetUpScript: []string{
				"CREATE TABLE t1 (a int primary key, b int);",
				"CREATE TABLE t2 (c int primary key);",
				"INSERT INTO t1 VALUES (1,10), (2,20), (3,30);",
				"INSERT INTO t2 VALUES (2), (3), (4);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT
s1.d FROM (SELECT b as f, a as g from t1) s1(d,e)
INNER JOIN t2 q1
ON q1.c = s1.e
ORDER BY 1;`,
					Expected: []sql.Row{
						{20},
						{30},
					},
				},
			},
		},
	})
}
