// Copyright 2023 Dolthub, Inc.
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

func TestRegressions(t *testing.T) {
	RunScripts(t, []ScriptTest{
		// {
		// 	Name:        "nullif",
		// 	SetUpScript: []string{},
		// 	Assertions: []ScriptTestAssertion{
		// 		{
		// 			Query:    "select nullif(1, 1);",
		// 			Expected: []sql.Row{{nil}},
		// 		},
		// 		{
		// 			Query:    "select nullif('', null);",
		// 			Expected: []sql.Row{{""}},
		// 		},
		// 		{
		// 			Query:    "select nullif(10, 'a');",
		// 			Expected: []sql.Row{{10}},
		// 		},
		// 	},
		// },
		// {
		// 	Name:        "coalesce",
		// 	SetUpScript: []string{},
		// 	Assertions: []ScriptTestAssertion{
		// 		{
		// 			Query:    "select coalesce(null + 5, 100);",
		// 			Expected: []sql.Row{{100.0}}, // TODO: this should be an integer
		// 		},
		// 		{
		// 			Query:    "select coalesce(null, null, 'abc');",
		// 			Expected: []sql.Row{{"abc"}},
		// 		},
		// 		{
		// 			Query:    "select coalesce(null, null);",
		// 			Expected: []sql.Row{{nil}},
		// 		},
		// 	},
		// },
		{
			Name:        "case / when",
			SetUpScript: []string{},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT\n" +
						"  CASE\n" +
						"    WHEN 1 = 1 THEN 'One is equal to One'\n" +
						"    ELSE 'One is not equal to One'\n" +
						"  END AS result;",
					Expected: []sql.Row{{"One is equal to One"}},
				},
				{
					Query: "SELECT\n" +
						"  CASE\n" +
						"    WHEN NULL IS NULL THEN 'NULL is NULL'\n" +
						"    ELSE 'NULL is not NULL'\n" +
						"  END AS result;",
					Expected: []sql.Row{{"NULL is NULL"}},
				},
			},
		},
		{
			Name: "ALL / DISTINCT in functions",
			SetUpScript: []string{
				"create table t1 (pk int);",
				"insert into t1 values (1), (2), (3), (1);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "select all count(*) from t1;",
					Expected: []sql.Row{{4}},
				},
				{
					Query:    "select all count(distinct pk) from t1;",
					Expected: []sql.Row{{3}},
				},
				{
					Query:    "select all count(all pk) from t1;",
					Expected: []sql.Row{{4}},
				},
			},
		},
		{
			Name: "cross joins",
			SetUpScript: []string{
				"create table t1 (pk1 int);",
				"create table t2 (pk2 int);",
				"insert into t1 values (1), (2);",
				"insert into t2 values (3), (4);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "select * from t1 cross join t2 order by pk1, pk2;",
					Expected: []sql.Row{
						{1, 3},
						{1, 4},
						{2, 3},
						{2, 4},
					},
				},
				{
					Query: "select * from t1, t2 order by pk1, pk2;",
					Expected: []sql.Row{
						{1, 3},
						{1, 4},
						{2, 3},
						{2, 4},
					},
				},
			},
		},
		{
			Skip: true,
			Name: "casting null as integer",
			SetUpScript: []string{
				`CREATE TABLE tab0(pk INTEGER PRIMARY KEY, col0 INTEGER, col1 FLOAT, col2 TEXT, col3 INTEGER, col4 FLOAT, col5 TEXT);`,
				`INSERT INTO tab0 VALUES (0,698,169.42,'apdbu',431,316.15,'sqvis'), (1,538,676.36,'fuqeu',514,685.97,'bgwrq');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT ALL + 58 FROM tab0 WHERE NULL NOT BETWEEN + 71 * CAST ( NULL AS INTEGER ) AND col4",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "addition expression in prepared statement",
			SetUpScript: []string{
				`CREATE TABLE t1(x INTEGER);`,
				`CREATE TABLE t2(y INTEGER PRIMARY KEY);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT 1 IN (SELECT x+y FROM t1, t2);`,
					Expected: []sql.Row{{0}},
				},
			},
		},
		{
			Skip: true,
			Name: "casting from float64 to int64 and float32",
			SetUpScript: []string{
				`CREATE TABLE tab0(pk INTEGER PRIMARY KEY, col0 INTEGER, col1 FLOAT, col2 TEXT, col3 INTEGER, col4 FLOAT, col5 TEXT);`,
				`INSERT INTO tab0 VALUES (0,698,169.42,'apdbu',431,316.15,'sqvis'), (1,538,676.36,'fuqeu',514,685.97,'bgwrq'), (2,90,205.26,'yrrzx',123,836.88,'kpuhc'), 
(3,620,864.8,'myrdv',877,820.98,'oxkuv'), (4,754,677.3,'iofrg',67,665.49,'bzqba'), (5,107,710.19,'lhfro',286,504.28,'kwwsg'), (6,904,193.16,'eozui',48,698.55,'ejyzs'), 
(7,606,650.64,'ovmce',417,962.43,'dvkbh'), (8,535,18.11,'ijika',630,489.63,'hpnyu'), (9,501,776.40,'cvygg',725,75.5,'etlyv');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT ALL * FROM tab0 cor0 WHERE ( + CAST ( - col4 AS INTEGER ) ) IN ( - col1, + col3, col3 );",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT * FROM tab0 WHERE - - col0 * + - col4 >= ( + CAST ( col1 AS REAL ) );",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Skip: true,
			Name: "typecheck fails to detect doltgres types in GMS ",
			SetUpScript: []string{
				`CREATE TABLE tab0(col0 INTEGER, col1 INTEGER, col2 INTEGER);`,
				`INSERT INTO tab0 VALUES (97,1,99), (15,81,47), (87,21,10);`,
				`CREATE TABLE tab1(col0 INTEGER, col1 INTEGER, col2 INTEGER);`,
				`INSERT INTO tab1 VALUES (51,14,96), (85,5,59), (91,47,68);`,
				`CREATE TABLE tab2(col0 INTEGER, col1 INTEGER, col2 INTEGER);`,
				`INSERT INTO tab2 VALUES(64,77,40), (75,67,58), (46,51,23);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT ALL + col2 * + ( ( 43 ) ) - 47 + + col1 * CAST ( - ( + 63 ) / col0 AS INTEGER ) AS col0 FROM tab1 AS cor0;",
					Expected: []sql.Row{{4067}, {2490}, {2877}},
				},
				{
					Query:    "SELECT - COUNT ( * ) - 26 * + 96 AS col2 FROM tab0 WHERE + 2 * col0 NOT BETWEEN 33 * CAST ( + 7 / 91 AS REAL ) AND 52;",
					Expected: []sql.Row{{-2498}},
				},
				{
					Query:    "SELECT CAST ( + CAST ( 73 AS REAL ) AS INTEGER ) * 62 FROM tab2 WHERE NULL IS NULL;",
					Expected: []sql.Row{{4526}, {4526}, {4526}},
				},
			},
		},
	})
}
