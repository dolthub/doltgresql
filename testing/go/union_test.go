// Copyright 2024 Dolthub, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package _go

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestUnion(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "union tests",
			SetUpScript: []string{
				`CREATE TABLE t1 (i INT PRIMARY KEY);`,
				`CREATE TABLE t2 (j INT PRIMARY KEY);`,
				`INSERT INTO t1 VALUES (1), (2), (3);`,
				`INSERT INTO t2 VALUES (2), (3), (4);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT * FROM t1 UNION SELECT * FROM t2;`,
					Expected: []sql.Row{
						{1},
						{2},
						{3},
						{4},
					},
				},
				{
					Query: `SELECT 123 UNION SELECT 456;`,
					Expected: []sql.Row{
						{123},
						{456},
					},
				},
				{
					Query: `SELECT * FROM (VALUES (123), (456)) a UNION SELECT * FROM (VALUES (456), (789)) b;`,
					Expected: []sql.Row{
						{123},
						{456},
						{789},
					},
				},
			},
		},
	})
}

func TestIntersect(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "intersect tests",
			SetUpScript: []string{
				`CREATE TABLE t1 (i INT PRIMARY KEY);`,
				`CREATE TABLE t2 (j INT PRIMARY KEY);`,
				`INSERT INTO t1 VALUES (1), (2), (3);`,
				`INSERT INTO t2 VALUES (2), (3), (4);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT * FROM t1 INTERSECT SELECT * FROM t2;`,
					Expected: []sql.Row{
						{2},
						{3},
					},
				},
				{
					Query:    `SELECT 123 INTERSECT SELECT 456;`,
					Expected: []sql.Row{},
				},
				{
					Query: `SELECT * FROM (VALUES (123), (456)) a INTERSECT SELECT * FROM (VALUES (456), (789)) b;`,
					Expected: []sql.Row{
						{456},
					},
				},
			},
		},
	})
}

func TestExcept(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "except tests",
			SetUpScript: []string{
				`CREATE TABLE t1 (i INT PRIMARY KEY);`,
				`CREATE TABLE t2 (j INT PRIMARY KEY);`,
				`INSERT INTO t1 VALUES (1), (2), (3);`,
				`INSERT INTO t2 VALUES (2), (3), (4);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT * FROM t1 EXCEPT SELECT * FROM t2;`,
					Expected: []sql.Row{
						{1},
					},
				},
				{
					Query: `SELECT 123 EXCEPT SELECT 456;`,
					Expected: []sql.Row{
						{123},
					},
				},
				{
					Query: `SELECT * FROM (VALUES (123), (456)) a EXCEPT SELECT * FROM (VALUES (456), (789)) b;`,
					Expected: []sql.Row{
						{123},
					},
				},
			},
		},
	})
}

func TestSetOperationTypes(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "set operations convert columns to a common type",
			SetUpScript: []string{
				`CREATE TABLE setop_nums (id INT PRIMARY KEY, d NUMERIC(10,2), f DOUBLE PRECISION, s SMALLINT, b BIGINT);`,
				`INSERT INTO setop_nums VALUES (1, 1.00, 1, 1, 1), (2, 2.50, 2.5, 2, 20), (5, 5.00, 5, 5, 50);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT pg_typeof(x) FROM (SELECT id AS x FROM setop_nums UNION SELECT d FROM setop_nums) s LIMIT 1;`,
					Expected: []sql.Row{{"numeric"}},
				},
				{
					Query:    `SELECT d FROM setop_nums UNION ALL SELECT id FROM setop_nums WHERE id = 2 ORDER BY 1;`,
					Expected: []sql.Row{{Numeric("1.00")}, {Numeric("2")}, {Numeric("2.50")}, {Numeric("5.00")}},
				},
				{
					Query:    `SELECT id FROM setop_nums UNION SELECT f FROM setop_nums ORDER BY 1;`,
					Expected: []sql.Row{{1.0}, {2.0}, {2.5}, {5.0}},
				},
				{
					Query:    `SELECT pg_typeof(x) FROM (SELECT id AS x FROM setop_nums UNION SELECT f FROM setop_nums) s LIMIT 1;`,
					Expected: []sql.Row{{"double precision"}},
				},
				{
					Query:    `SELECT s FROM setop_nums UNION SELECT b FROM setop_nums ORDER BY 1;`,
					Expected: []sql.Row{{1}, {2}, {5}, {20}, {50}},
				},
				{
					Query:    `SELECT pg_typeof(x) FROM (SELECT s AS x FROM setop_nums UNION SELECT b FROM setop_nums) s LIMIT 1;`,
					Expected: []sql.Row{{"bigint"}},
				},
				{
					Query:    `SELECT * FROM (VALUES (1), (2.5)) v(n) UNION SELECT id FROM setop_nums ORDER BY 1;`,
					Expected: []sql.Row{{Numeric("1")}, {Numeric("2")}, {Numeric("2.5")}, {Numeric("5")}},
				},
				{
					Query:    `SELECT pg_typeof(x) FROM (SELECT 'x'::varchar(3) AS x UNION SELECT 'yy'::varchar(5)) s LIMIT 1;`,
					Expected: []sql.Row{{"character varying"}},
				},
				{
					Query:       `SELECT 1 UNION SELECT 'a';`,
					ExpectedErr: `invalid input syntax for type int4: "a"`,
				},
				{
					// TODO: these need dolthub/go-mysql-server#4068, numerics with a different scale hash differently
					Skip:     true,
					Query:    `SELECT 5 UNION SELECT 5.0;`,
					Expected: []sql.Row{{Numeric("5")}},
				},
				{
					Skip:     true,
					Query:    `SELECT COUNT(*) FROM (SELECT id FROM setop_nums INTERSECT SELECT d FROM setop_nums) s;`,
					Expected: []sql.Row{{2}},
				},
				{
					Skip:     true,
					Query:    `SELECT id FROM setop_nums EXCEPT SELECT d FROM setop_nums;`,
					Expected: []sql.Row{{Numeric("2")}},
				},
				{
					Skip:     true,
					Query:    `SELECT COUNT(*) FROM (SELECT id FROM setop_nums UNION SELECT d FROM setop_nums) s;`,
					Expected: []sql.Row{{4}},
				},
			},
		},
	})
}
