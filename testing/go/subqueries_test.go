package _go

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestSubqueries(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Subselect",
			SetUpScript: []string{
				`CREATE TABLE test (id INT);`,
				`INSERT INTO test VALUES (1), (3), (2);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT * FROM test WHERE id = (SELECT 2);`,
					Expected: []sql.Row{
						{int32(2)},
					},
				},
				{
					Query: `SELECT *, (SELECT id from test where id = 2) FROM test order by id;`,
					Expected: []sql.Row{
						{1, 2},
						{2, 2},
						{3, 2},
					},
				},
				{
					Query: `SELECT *, (SELECT id from test t2 where t2.id = test.id) FROM test order by id;`,
					Expected: []sql.Row{
						{1, 1},
						{2, 2},
						{3, 3},
					},
				},
			},
		},
		{
			Name: "IN",
			SetUpScript: []string{
				`CREATE TABLE test (id INT);`,
				`INSERT INTO test VALUES (1), (3), (2);`,

				`CREATE TABLE test2 (id INT, test_id INT, txt text);`,
				`INSERT INTO test2 VALUES (1, 1, 'foo'), (2, 10, 'bar'), (3, 2, 'baz');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM test WHERE id IN (SELECT * FROM test WHERE id = 2);`,
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:    `SELECT * FROM test WHERE id IN (SELECT id FROM test WHERE id = 3);`,
					Expected: []sql.Row{{int32(3)}},
				},
				{
					Query:    `SELECT * FROM test WHERE id IN (SELECT * FROM test WHERE id > 0);`,
					Expected: []sql.Row{{int32(1)}, {int32(3)}, {int32(2)}},
				},
				{
					Query:    `SELECT * FROM test2 WHERE test_id IN (SELECT * FROM test WHERE id = 2);`,
					Expected: []sql.Row{{int32(3), int32(2), "baz"}},
				},
				{
					Query: `SELECT * FROM test2 WHERE test_id IN (SELECT * FROM test WHERE id > 0);`,
					Expected: []sql.Row{
						{int32(1), int32(1), "foo"},
						{int32(3), int32(2), "baz"},
					},
				},
				{
					Query: `SELECT id FROM test2 WHERE (2, 10) IN (SELECT id, test_id FROM test2 WHERE id > 0);`,
					Skip:  true, // won't pass until we have a doltgres tuple type to match against for equality funcs
					Expected: []sql.Row{
						{1}, {2}, {3},
					},
				},
				{
					Query: `SELECT id FROM test2 WHERE (id, test_id) IN (SELECT id, test_id FROM test2 WHERE id > 0);`,
					Skip:  true, // won't pass until we have a doltgres tuple type to match against for equality funcs
					Expected: []sql.Row{
						{2},
					},
				},
			},
		},
	})
}

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
