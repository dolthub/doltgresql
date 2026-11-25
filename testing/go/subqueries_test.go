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
			Name: "IN subquery unnests into a semi join",
			SetUpScript: []string{
				`CREATE TABLE parent (id INT PRIMARY KEY, name TEXT NOT NULL);`,
				`CREATE TABLE child (id INT NOT NULL, parent_id INT NOT NULL REFERENCES parent (id), n INT NOT NULL, PRIMARY KEY (id, parent_id));`,
				`CREATE TABLE ids (v BIGINT);`,
				`INSERT INTO parent (id, name) SELECT g, 'parent ' || g FROM generate_series(1, 5) AS g;`,
				`INSERT INTO child (id, parent_id, n) SELECT i, p, i * p FROM generate_series(1, 1000) AS i, generate_series(1, 5) AS p;`,
				`INSERT INTO ids VALUES (42), (42), (999), (NULL);`,
				`ANALYZE child;`,
				`ANALYZE parent;`,
				`ANALYZE ids;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `EXPLAIN SELECT * FROM child WHERE id IN (SELECT v FROM (VALUES (42), (4242)) AS ids(v));`,
					Expected: []sql.Row{
						{"Project"},
						{" ├─ columns: [child.id, child.parent_id, child.n]"},
						{" └─ LookupJoin"},
						{"     ├─ Distinct"},
						{"     │   └─ Values() as ids"},
						{"     │       ├─ Row("},
						{"     │       │  42)"},
						{"     │       └─ Row("},
						{"     │          4242)"},
						{"     └─ IndexedTableAccess(child)"},
						{"         ├─ index: [child.id,child.parent_id]"},
						{"         └─ keys: ids.v"},
					},
				},
				{
					Query: `EXPLAIN SELECT * FROM child WHERE id IN (SELECT v FROM ids);`,
					Expected: []sql.Row{
						{"Project"},
						{" ├─ columns: [child.id, child.parent_id, child.n]"},
						{" └─ LookupJoin"},
						{"     ├─ Distinct"},
						{"     │   └─ Table"},
						{"     │       └─ name: ids"},
						{"     └─ IndexedTableAccess(child)"},
						{"         ├─ index: [child.id,child.parent_id]"},
						{"         └─ keys: ids.v"},
					},
				},
				{
					Query: `EXPLAIN SELECT * FROM child WHERE parent_id IN (SELECT id FROM parent WHERE id < 2);`,
					Expected: []sql.Row{
						{"Project"},
						{" ├─ columns: [child.id, child.parent_id, child.n]"},
						{" └─ LookupJoin"},
						{"     ├─ Distinct"},
						{"     │   └─ Project"},
						{"     │       ├─ columns: [parent.id]"},
						{"     │       └─ IndexedTableAccess(parent)"},
						{"     │           ├─ index: [parent.id]"},
						{"     │           └─ filters: [{(NULL, 2)}]"},
						{"     └─ IndexedTableAccess(child)"},
						{"         ├─ index: [child.parent_id]"},
						{"         └─ keys: parent.id"},
					},
				},
				{
					Query: `EXPLAIN SELECT * FROM child WHERE id NOT IN (SELECT v FROM ids);`,
					Expected: []sql.Row{
						{"Filter"},
						{" ├─ NOT child.id IN ((select  v from ids))"},
						{" └─ Table"},
						{"     └─ name: child"},
					},
				},
				{
					Query: `SELECT * FROM child WHERE id IN (SELECT v FROM (VALUES (42), (4242)) AS ids(v)) ORDER BY id, parent_id;`,
					Expected: []sql.Row{
						{42, 1, 42},
						{42, 2, 84},
						{42, 3, 126},
						{42, 4, 168},
						{42, 5, 210},
					},
				},
				{
					Query: `SELECT * FROM child WHERE id IN (SELECT v FROM ids) ORDER BY id, parent_id;`,
					Expected: []sql.Row{
						{42, 1, 42},
						{42, 2, 84},
						{42, 3, 126},
						{42, 4, 168},
						{42, 5, 210},
						{999, 1, 999},
						{999, 2, 1998},
						{999, 3, 2997},
						{999, 4, 3996},
						{999, 5, 4995},
					},
				},
				{
					Query:    `SELECT count(*) FROM child WHERE parent_id IN (SELECT id FROM parent WHERE id < 2);`,
					Expected: []sql.Row{{1000}},
				},
				{
					Query:    `SELECT count(*) FROM child WHERE id NOT IN (SELECT v FROM ids);`,
					Expected: []sql.Row{{0}},
				},
				{
					Query:    `SELECT count(*) FROM child WHERE id NOT IN (SELECT v FROM ids WHERE v IS NOT NULL);`,
					Expected: []sql.Row{{4990}},
				},
				{
					Query: `SELECT * FROM child WHERE id IN (SELECT v FROM (VALUES (42.0), (43.5)) AS ids(v)) ORDER BY id, parent_id;`,
					Expected: []sql.Row{
						{42, 1, 42},
						{42, 2, 84},
						{42, 3, 126},
						{42, 4, 168},
						{42, 5, 210},
					},
				},
				{
					Query: `SELECT * FROM child WHERE id IN (SELECT v FROM (VALUES (3000000000), (7)) AS ids(v)) ORDER BY id, parent_id;`,
					Expected: []sql.Row{
						{7, 1, 7},
						{7, 2, 14},
						{7, 3, 21},
						{7, 4, 28},
						{7, 5, 35},
					},
				},
				{
					Query: `SELECT * FROM child WHERE id IN (SELECT v FROM (VALUES (3000000000), (7)) AS ids(v) WHERE v > 0) ORDER BY id, parent_id;`,
					Expected: []sql.Row{
						{7, 1, 7},
						{7, 2, 14},
						{7, 3, 21},
						{7, 4, 28},
						{7, 5, 35},
					},
				},
				{
					Query: `SELECT * FROM child WHERE id IN (SELECT v FROM ids ORDER BY v LIMIT 1) ORDER BY id, parent_id;`,
					Expected: []sql.Row{
						{42, 1, 42},
						{42, 2, 84},
						{42, 3, 126},
						{42, 4, 168},
						{42, 5, 210},
					},
				},
				{
					Query: `SELECT * FROM child WHERE id IN (SELECT max(v) FROM ids) ORDER BY id, parent_id;`,
					Expected: []sql.Row{
						{999, 1, 999},
						{999, 2, 1998},
						{999, 3, 2997},
						{999, 4, 3996},
						{999, 5, 4995},
					},
				},
				{
					Query: `SELECT * FROM child WHERE id IN (SELECT v + 1 FROM ids) ORDER BY id, parent_id;`,
					Expected: []sql.Row{
						{43, 1, 43},
						{43, 2, 86},
						{43, 3, 129},
						{43, 4, 172},
						{43, 5, 215},
						{1000, 1, 1000},
						{1000, 2, 2000},
						{1000, 3, 3000},
						{1000, 4, 4000},
						{1000, 5, 5000},
					},
				},
				{
					Query: `SELECT * FROM child WHERE id IN (SELECT v FROM ids UNION SELECT 7::bigint) AND parent_id = 1 ORDER BY id, parent_id;`,
					Expected: []sql.Row{
						{7, 1, 7},
						{42, 1, 42},
						{999, 1, 999},
					},
				},
				{
					Query: `SELECT * FROM child WHERE n IN (SELECT id * 10 FROM parent) ORDER BY id, parent_id;`,
					Expected: []sql.Row{
						{2, 5, 10},
						{4, 5, 20},
						{5, 2, 10},
						{5, 4, 20},
						{6, 5, 30},
						{8, 5, 40},
						{10, 1, 10},
						{10, 2, 20},
						{10, 3, 30},
						{10, 4, 40},
						{10, 5, 50},
						{15, 2, 30},
						{20, 1, 20},
						{20, 2, 40},
						{25, 2, 50},
						{30, 1, 30},
						{40, 1, 40},
						{50, 1, 50},
					},
				},
				{
					Query: `SELECT * FROM child WHERE id IN (SELECT v FROM ids WHERE child.parent_id = 1) ORDER BY id, parent_id;`,
					Expected: []sql.Row{
						{42, 1, 42},
						{999, 1, 999},
					},
				},
				{
					Query:    `SELECT count(*) FROM child WHERE id NOT IN (SELECT v FROM ids WHERE v IS NOT NULL AND child.parent_id = 1);`,
					Expected: []sql.Row{{4998}},
				},
				{
					Query:    `SELECT count(*) FROM child WHERE id NOT IN (SELECT v FROM ids WHERE child.parent_id = 1);`,
					Expected: []sql.Row{{4000}},
				},
				{
					Query:       `SELECT * FROM child WHERE id IN (SELECT v FROM (VALUES ('42'::text)) AS ids(v));`,
					ExpectedErr: "operator does not exist",
				},
				{
					Query: `SELECT id, id IN (SELECT v FROM ids) FROM child WHERE parent_id = 1 AND id IN (1, 42) ORDER BY id;`,
					Expected: []sql.Row{
						{1, nil},
						{42, "t"},
					},
				},
				{
					Query: `SELECT * FROM child WHERE id IN (SELECT v FROM ids) AND parent_id IN (SELECT id FROM parent WHERE name = 'parent 3') ORDER BY id;`,
					Expected: []sql.Row{
						{42, 3, 126},
						{999, 3, 2997},
					},
				},
				{
					Query: `SELECT * FROM child WHERE id IN (SELECT v FROM ids) OR n = 5 ORDER BY id, parent_id;`,
					Expected: []sql.Row{
						{1, 5, 5},
						{5, 1, 5},
						{42, 1, 42},
						{42, 2, 84},
						{42, 3, 126},
						{42, 4, 168},
						{42, 5, 210},
						{999, 1, 999},
						{999, 2, 1998},
						{999, 3, 2997},
						{999, 4, 3996},
						{999, 5, 4995},
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
