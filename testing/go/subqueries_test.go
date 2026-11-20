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
		{
			Name: "subquery equality",
			SetUpScript: []string{
				`CREATE TABLE test (id INT, c varchar);`,
				`INSERT INTO test VALUES (1, 'a'), (2, 'b'), (3, 'b');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM test WHERE id = (SELECT id from test where id = 2);`,
					Expected: []sql.Row{{int32(2), "b"}},
				},
				{
					Skip:     true, // panic in equality func
					Query:    `SELECT (SELECT id from test where id = 2) = (SELECT id from test where id = 2);`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Skip:     true, // panic in equality func
					Query:    `SELECT (SELECT c from test where id = 2) = (SELECT c from test where id = 3);`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Skip:     true, // panic in equality func
					Query:    `SELECT (SELECT c from test where id = 1) = (SELECT c from test where id = 2);`,
					Expected: []sql.Row{{"f"}},
				},
			},
		},
		{
			Name: "array flatten",
			SetUpScript: []string{
				`CREATE TABLE test (id INT, c varchar);`,
				`INSERT INTO test VALUES (1, 'a'), (2, 'b'), (3, 'c');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT ARRAY(SELECT id FROM test order by 1);`,
					Expected: []sql.Row{
						{"{1,2,3}"},
					},
				},
				{
					Query: `SELECT ARRAY(SELECT c FROM test order by id limit 1);`,
					Expected: []sql.Row{
						{"{a}"},
					},
				},
				{
					Query: `SELECT ARRAY(SELECT c FROM test order by id desc);`,
					Expected: []sql.Row{
						{"{c,b,a}"},
					},
				},
				{
					Query:       `SELECT ARRAY(SELECT id, id FROM test order by 1);`,
					ExpectedErr: "only a single column subquery is supported",
				},
				{
					Query: `SELECT array_to_string(ARRAY(SELECT id FROM test order by 1), ',')`,
					Expected: []sql.Row{
						{"1,2,3"},
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

func TestExistSubquery(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "basic case",
			SetUpScript: []string{
				`CREATE TABLE test (id INT PRIMARY KEY);`,
				`INSERT INTO test VALUES (1), (3), (2);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT * FROM test WHERE EXISTS (SELECT 123);`,
					Expected: []sql.Row{
						{1},
						{2},
						{3},
					},
				},
				{
					Query:    `SELECT * FROM test WHERE NOT EXISTS (SELECT 123);`,
					Expected: []sql.Row{},
				},
				{
					Query: `SELECT 123 WHERE EXISTS (SELECT * FROM test);`,
					Expected: []sql.Row{
						{123},
					},
				},
				{
					Query:    `SELECT 123 WHERE EXISTS (SELECT * FROM test WHERE id > 10);`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT 123 WHERE NOT EXISTS (SELECT * FROM test);`,
					Expected: []sql.Row{},
				},
				{
					Query: `SELECT 123 WHERE NOT EXISTS (SELECT * FROM test WHERE id > 10);`,
					Expected: []sql.Row{
						{123},
					},
				},
			},
		},
		{
			Name: "correlated EXISTS decorrelates into a semi join",
			SetUpScript: []string{
				`CREATE TABLE a (id INT PRIMARY KEY, x INT);`,
				`CREATE TABLE b (id INT PRIMARY KEY, x INT);`,
				`INSERT INTO a VALUES (1,1),(2,2),(3,3);`,
				`INSERT INTO b VALUES (1,1),(2,2);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `EXPLAIN SELECT * FROM a WHERE EXISTS (SELECT 1 FROM b WHERE a.x = b.x);`,
					Expected: []sql.Row{
						{"SemiJoin"},
						{" ├─ a.x = b.x"},
						{" ├─ Table"},
						{" │   └─ name: a"},
						{" └─ Table"},
						{"     ├─ name: b"},
						{"     └─ columns: [x]"},
					},
				},
				{
					Query: `SELECT * FROM a WHERE EXISTS (SELECT 1 FROM b WHERE a.x = b.x) ORDER BY id;`,
					Expected: []sql.Row{
						{1, 1},
						{2, 2},
					},
				},
				{
					Query: `EXPLAIN SELECT * FROM a WHERE NOT EXISTS (SELECT 1 FROM b WHERE a.x = b.x);`,
					Expected: []sql.Row{
						{"Project"},
						{" ├─ columns: [a.id, a.x]"},
						{" └─ Filter"},
						{"     ├─ 1 IS NULL"},
						{"     └─ LeftOuterJoin"},
						{"         ├─ a.x = b.x"},
						{"         ├─ Table"},
						{"         │   └─ name: a"},
						{"         └─ Project"},
						{"             ├─ columns: [b.x, 1]"},
						{"             └─ Table"},
						{"                 ├─ name: b"},
						{"                 └─ columns: [x]"},
					},
				},
				{
					Query: `SELECT * FROM a WHERE NOT EXISTS (SELECT 1 FROM b WHERE a.x = b.x) ORDER BY id;`,
					Expected: []sql.Row{
						{3, 3},
					},
				},
			},
		},
	})
}

func TestInSubqueryUnnesting(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "IN subquery unnests into an indexed semi join",
			SetUpScript: []string{
				`CREATE TABLE parent (id int PRIMARY KEY, name text NOT NULL);`,
				`CREATE TABLE child (id int NOT NULL, parent_id int NOT NULL REFERENCES parent (id), n int NOT NULL, PRIMARY KEY (id, parent_id));`,
				`INSERT INTO parent (id, name) SELECT g, 'parent ' || g FROM generate_series(1, 5) AS g;`,
				`INSERT INTO child (id, parent_id, n) SELECT i, p, i * p FROM generate_series(1, 200) AS i, generate_series(1, 5) AS p;`,
				`ANALYZE child;`,
				`ANALYZE parent;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `EXPLAIN SELECT * FROM child WHERE id IN (SELECT v FROM (VALUES (42), (142)) AS ids(v));`,
					Expected: []sql.Row{
						{"Project"},
						{" ├─ columns: [child.id, child.parent_id, child.n]"},
						{" └─ LookupJoin"},
						{"     ├─ Distinct"},
						{"     │   └─ Values() as ids"},
						{"     │       ├─ Row("},
						{"     │       │  42)"},
						{"     │       └─ Row("},
						{"     │          142)"},
						{"     └─ IndexedTableAccess(child)"},
						{"         ├─ index: [child.id,child.parent_id]"},
						{"         └─ keys: ids.v"},
					},
				},
				{
					Query:    `SELECT * FROM child WHERE id IN (SELECT v FROM (VALUES (42), (142)) AS ids(v)) ORDER BY id, parent_id;`,
					Expected: []sql.Row{{42, 1, 42}, {42, 2, 84}, {42, 3, 126}, {42, 4, 168}, {42, 5, 210}, {142, 1, 142}, {142, 2, 284}, {142, 3, 426}, {142, 4, 568}, {142, 5, 710}},
				},
				{
					Query:    `SELECT * FROM child WHERE id IN (SELECT v FROM (VALUES (42), (42), (142), (142)) AS ids(v)) ORDER BY id, parent_id;`,
					Expected: []sql.Row{{42, 1, 42}, {42, 2, 84}, {42, 3, 126}, {42, 4, 168}, {42, 5, 210}, {142, 1, 142}, {142, 2, 284}, {142, 3, 426}, {142, 4, 568}, {142, 5, 710}},
				},
				{
					Query:    `SELECT count(*) FROM child WHERE id NOT IN (SELECT v FROM (VALUES (42), (142)) AS ids(v));`,
					Expected: []sql.Row{{990}},
				},
				{
					Query: `EXPLAIN SELECT * FROM parent WHERE id IN (SELECT parent_id FROM child WHERE n = 40);`,
					Expected: []sql.Row{
						{"SemiLookupJoin"},
						{" ├─ Table"},
						{" │   └─ name: parent"},
						{" └─ Project"},
						{"     ├─ columns: [child.parent_id]"},
						{"     └─ Filter"},
						{"         ├─ child.n = 40"},
						{"         └─ IndexedTableAccess(child)"},
						{"             ├─ index: [child.parent_id]"},
						{"             └─ keys: parent.id"},
					},
				},
				{
					Query:    `SELECT * FROM parent WHERE id IN (SELECT parent_id FROM child WHERE n = 40) ORDER BY id;`,
					Expected: []sql.Row{{1, "parent 1"}, {2, "parent 2"}, {4, "parent 4"}, {5, "parent 5"}},
				},
			},
		},
		{
			Name: "unnested IN subquery semantics",
			SetUpScript: []string{
				`CREATE TABLE a (id INT PRIMARY KEY, x INT, t TEXT);`,
				`CREATE TABLE b (id INT PRIMARY KEY, x INT, big BIGINT, num NUMERIC, t VARCHAR(10));`,
				`CREATE TABLE empty (x INT);`,
				`INSERT INTO a VALUES (1, 1, 'one'), (2, 2, 'two'), (3, 3, 'three'), (4, NULL, NULL);`,
				`INSERT INTO b VALUES (1, 1, 1, 1.0, 'one'), (2, 1, 2, 2.5, 'two'), (3, NULL, NULL, NULL, NULL);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT id FROM a WHERE x IN (SELECT x FROM b) ORDER BY id;`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT id FROM a WHERE x IN (SELECT x FROM b WHERE x IS NOT NULL) ORDER BY id;`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT id FROM a WHERE x NOT IN (SELECT x FROM b) ORDER BY id;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT id FROM a WHERE x NOT IN (SELECT x FROM b WHERE x IS NOT NULL) ORDER BY id;`,
					Expected: []sql.Row{{2}, {3}},
				},
				{
					Query:    `SELECT id FROM a WHERE x NOT IN (SELECT x FROM empty) ORDER BY id;`,
					Expected: []sql.Row{{1}, {2}, {3}, {4}},
				},
				{
					Query:    `SELECT id FROM a WHERE x IN (SELECT big FROM b) ORDER BY id;`,
					Expected: []sql.Row{{1}, {2}},
				},
				{
					Query:    `SELECT id FROM a WHERE x IN (SELECT num FROM b) ORDER BY id;`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT id FROM a WHERE t IN (SELECT t FROM b) ORDER BY id;`,
					Expected: []sql.Row{{1}, {2}},
				},
				{
					Query:    `SELECT id FROM a WHERE x IN (SELECT b.x FROM b WHERE b.id = a.id) ORDER BY id;`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT id FROM a WHERE x IN (SELECT x FROM b) AND id IN (SELECT id FROM b) ORDER BY id;`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT id, x IN (SELECT x FROM b) FROM a ORDER BY id;`,
					Expected: []sql.Row{{1, "t"}, {2, nil}, {3, nil}, {4, nil}},
				},
				{
					Query:    `SELECT id FROM a WHERE x IN (SELECT x FROM b) OR id = 3 ORDER BY id;`,
					Expected: []sql.Row{{1}, {3}},
				},
				{
					Query:    `CREATE TABLE keyless (x INT, t TEXT);`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO keyless VALUES (1, 'one'), (1, 'one'), (2, 'two'), (2, 'two'), (3, 'three');`,
					Expected: []sql.Row{},
				},
				{
					Query:    `UPDATE keyless SET t = 'z' WHERE x IN (SELECT big FROM b);`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM keyless ORDER BY x, t;`,
					Expected: []sql.Row{{1, "z"}, {1, "z"}, {2, "z"}, {2, "z"}, {3, "three"}},
				},
				{
					Query:    `UPDATE a SET t = 'z' WHERE x IN (SELECT big FROM b);`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM a ORDER BY id;`,
					Expected: []sql.Row{{1, 1, "z"}, {2, 2, "z"}, {3, 3, "three"}, {4, nil, nil}},
				},
			},
		},
	})
}
