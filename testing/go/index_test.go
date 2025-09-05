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

	"github.com/dolthub/doltgresql/testing/go/testdata"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestBasicIndexing(t *testing.T) {
	RunScripts(t, []ScriptTest{

		{
			Name: "Covering Index",
			SetUpScript: []string{
				"CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
				"INSERT INTO test VALUES (13, 3), (11, 1), (15, 5), (12, 2), (14, 4);",
				"CREATE INDEX v1_idx ON test(v1);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM test WHERE v1 = 2 ORDER BY pk;",
					Expected: []sql.Row{
						{12, 2},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 > 2 ORDER BY pk;",
					Expected: []sql.Row{
						{13, 3},
						{14, 4},
						{15, 5},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 >= 4 ORDER BY pk;",
					Expected: []sql.Row{
						{14, 4},
						{15, 5},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 < 3 ORDER BY pk;",
					Expected: []sql.Row{
						{11, 1},
						{12, 2},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 <= 3 ORDER BY pk;",
					Expected: []sql.Row{
						{11, 1},
						{12, 2},
						{13, 3},
					},
				},
			},
		},
		{
			Name: "Covering string Index",
			SetUpScript: []string{
				"CREATE TABLE test (pk bigint PRIMARY KEY, v1 varchar(10));",
				"INSERT INTO test VALUES (13, 'thirteen'), (11, 'eleven'), (15, 'fifteen'), (12, 'twelve'), (14, 'fourteen');",
				"CREATE UNIQUE INDEX v1_idx ON test(v1);",
				"CREATE INDEX v1_pk_idx ON test(v1, pk);",
				"CREATE INDEX pk_v1_idx ON test(pk, v1);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM test WHERE v1 = 'twelve' ORDER BY pk;",
					Expected: []sql.Row{
						{12, "twelve"},
					},
				},
				{
					Query:            "DELETE FROM test WHERE v1 = 'twelve'",
					SkipResultsCheck: true,
				},
				{
					Query:    "SELECT * FROM test WHERE v1 = 'twelve' ORDER BY pk;",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "String primary key ordering",
			SetUpScript: []string{
				"create table t (s varchar(5) primary key);",
				"insert into t values ('foo');",
				"insert into t values ('bar');",
				"insert into t values ('baz');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "select * from t order by s;",
					Expected: []sql.Row{{"bar"}, {"baz"}, {"foo"}},
				},
			},
		},
		{
			Name: "Unique Covering Index",
			SetUpScript: []string{
				"CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
				"INSERT INTO test VALUES (13, 3), (11, 1), (15, 5), (12, 2), (14, 4);",
				"CREATE unique INDEX v1_idx ON test(v1);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM test WHERE v1 > 2 ORDER BY pk;",
					Expected: []sql.Row{
						{13, 3},
						{14, 4},
						{15, 5},
					},
				},
				{
					Query:       "insert into test values (16, 3);",
					ExpectedErr: "duplicate unique key given",
				},
			},
		},
		{
			Name:  "Covering Composite Index",
			Focus: true,
			SetUpScript: []string{
				"CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT, v2 BIGINT);",
				"INSERT INTO test VALUES (13, 3, 23), (11, 1, 21), (15, 5, 25), (12, 2, 22), (14, 4, 24);",
				"CREATE INDEX v1_idx ON test(v1, v2);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "EXPLAIN SELECT * FROM test WHERE v1 = 2 AND v2 = 22 ORDER BY pk;",
					Expected: []sql.Row{
						{12, 2, 22},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 = 2 AND v2 = 22 ORDER BY pk;",
					Expected: []sql.Row{
						{12, 2, 22},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 > 2 AND v2 = 24 ORDER BY pk;",
					Expected: []sql.Row{
						{14, 4, 24},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 >= 4 AND v2 = 25 ORDER BY pk;",
					Expected: []sql.Row{
						{15, 5, 25},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 < 3 AND v2 = 21 ORDER BY pk;",
					Expected: []sql.Row{
						{11, 1, 21},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 <= 3 AND v2 = 22 ORDER BY pk;",
					Expected: []sql.Row{
						{12, 2, 22},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 = 2 AND v2 < 23 ORDER BY pk;",
					Expected: []sql.Row{
						{12, 2, 22},
					},
				},
				{
					Query:    "SELECT * FROM test WHERE v1 = 2 AND v2 < 22 ORDER BY pk;",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test WHERE v1 > 2 AND v2 < 25 ORDER BY pk;",
					Expected: []sql.Row{
						{13, 3, 23},
						{14, 4, 24},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 >= 4 AND v2 <= 24 ORDER BY pk;",
					Expected: []sql.Row{
						{14, 4, 24},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 < 3 AND v2 < 22 ORDER BY pk;",
					Expected: []sql.Row{
						{11, 1, 21},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 <= 3 AND v2 < 23 ORDER BY pk;",
					Expected: []sql.Row{
						{11, 1, 21},
						{12, 2, 22},
					},
				},
			},
		},
		{
			Name: "Covering Index Multiple AND",
			SetUpScript: []string{
				"CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
				"INSERT INTO test VALUES (13, 3), (11, 1), (15, 5), (12, 2), (14, 4);",
				"CREATE INDEX v1_idx ON test(v1);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT * FROM test WHERE v1 = 2 AND v1 = '3' ORDER BY pk;",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test WHERE v1 > 2 AND v1 > '3' ORDER BY pk;",
					Expected: []sql.Row{
						{14, 4},
						{15, 5},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 >= 3 AND v1 <= 4.0 ORDER BY pk;",
					Expected: []sql.Row{
						{13, 3},
						{14, 4},
					},
				},
				{
					Query:    "SELECT * FROM test WHERE v1 < 3 AND v1 > 3::float8 ORDER BY pk;",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test WHERE v1 <= 3 AND v1 = 1 ORDER BY pk;",
					Expected: []sql.Row{
						{11, 1},
					},
				},
			},
		},
		{
			Name: "Covering Index BETWEEN",
			SetUpScript: []string{
				"CREATE TABLE test (pk FLOAT8 PRIMARY KEY, v1 FLOAT8);",
				"INSERT INTO test VALUES (13, 3), (11, 1), (17, 7);",
				"CREATE INDEX v1_idx ON test(v1);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM test WHERE v1 BETWEEN 1 AND 4 ORDER BY pk;",
					Expected: []sql.Row{
						{float64(11), float64(1)},
						{float64(13), float64(3)},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 BETWEEN 2 AND 4 ORDER BY pk;",
					Expected: []sql.Row{
						{float64(13), float64(3)},
					},
				},
				{
					Query:    "SELECT * FROM test WHERE v1 BETWEEN 4 AND 2 ORDER BY pk;",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test WHERE v1 BETWEEN SYMMETRIC 1 AND 4 ORDER BY pk;",
					Expected: []sql.Row{
						{float64(11), float64(1)},
						{float64(13), float64(3)},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 BETWEEN SYMMETRIC 2 AND 4 ORDER BY pk;",
					Expected: []sql.Row{
						{float64(13), float64(3)},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 BETWEEN SYMMETRIC 4 AND 2 ORDER BY pk;",
					Expected: []sql.Row{
						{float64(13), float64(3)},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 NOT BETWEEN 1 AND 4 ORDER BY pk;",
					Expected: []sql.Row{
						{float64(17), float64(7)},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 NOT BETWEEN 2 AND 4 ORDER BY pk;",
					Expected: []sql.Row{
						{float64(11), float64(1)},
						{float64(17), float64(7)},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 NOT BETWEEN 4 AND 2 ORDER BY pk;",
					Expected: []sql.Row{
						{float64(11), float64(1)},
						{float64(13), float64(3)},
						{float64(17), float64(7)},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 NOT BETWEEN SYMMETRIC 1 AND 4 ORDER BY pk;",
					Expected: []sql.Row{
						{float64(17), float64(7)},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 NOT BETWEEN SYMMETRIC 2 AND 4 ORDER BY pk;",
					Expected: []sql.Row{
						{float64(11), float64(1)},
						{float64(17), float64(7)},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 NOT BETWEEN SYMMETRIC 4 AND 2 ORDER BY pk;",
					Expected: []sql.Row{
						{float64(11), float64(1)},
						{float64(17), float64(7)},
					},
				},
			},
		},

		{
			Name: "Covering Index IN",
			SetUpScript: []string{
				"CREATE TABLE test(pk INT4 PRIMARY KEY, v1 INT4, v2 INT4);",
				"INSERT INTO test VALUES (1, 1, 1), (2, 2, 2), (3, 3, 3), (4, 4, 4), (5, 5, 5);",
				"CREATE INDEX v1_idx ON test(v1);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM test WHERE v1 IN (2, '3', 4) ORDER BY v1;",
					Expected: []sql.Row{
						{2, 2, 2},
						{3, 3, 3},
						{4, 4, 4},
					},
				},
				{
					Query:    "CREATE INDEX v2_idx ON test(v2);",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test WHERE v2 IN (2, '3', 4) ORDER BY v1;",
					Expected: []sql.Row{
						{2, 2, 2},
						{3, 3, 3},
						{4, 4, 4},
					},
				},
			},
		},
		{
			Name: "Non-Covering Index",
			SetUpScript: []string{
				"CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT, v2 BIGINT);",
				"INSERT INTO test VALUES (13, 3, 23), (11, 1, 21), (15, 5, 25), (12, 2, 22), (14, 4, 24);",
				"CREATE INDEX v1_idx ON test(v1);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM test WHERE v1 = 2 ORDER BY pk;",
					Expected: []sql.Row{
						{12, 2, 22},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 > 2 ORDER BY pk;",
					Expected: []sql.Row{
						{13, 3, 23},
						{14, 4, 24},
						{15, 5, 25},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 >= 4 ORDER BY pk;",
					Expected: []sql.Row{
						{14, 4, 24},
						{15, 5, 25},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 < 3 ORDER BY pk;",
					Expected: []sql.Row{
						{11, 1, 21},
						{12, 2, 22},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 <= 3 ORDER BY pk;",
					Expected: []sql.Row{
						{11, 1, 21},
						{12, 2, 22},
						{13, 3, 23},
					},
				},
			},
		},
		{
			Name: "Unique Non-Covering Index",
			SetUpScript: []string{
				"CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT, v2 BIGINT);",
				"INSERT INTO test VALUES (13, 3, 23), (11, 1, 21), (15, 5, 25), (12, 2, 22), (14, 4, 24);",
				"CREATE UNIQUE INDEX v1_idx ON test(v1);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM test WHERE v1 > 2 ORDER BY pk;",
					Expected: []sql.Row{
						{13, 3, 23},
						{14, 4, 24},
						{15, 5, 25},
					},
				},
				{
					Query:       "insert into test values (16, 3, 23);",
					ExpectedErr: "duplicate unique key given",
				},
			},
		},
		{
			Name: "Non-Covering Composite Index",
			SetUpScript: []string{
				"CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT, v2 BIGINT, v3 BIGINT);",
				"INSERT INTO test VALUES (13, 3, 23, 33), (11, 1, 21, 31), (15, 5, 25, 35), (12, 2, 22, 32), (14, 4, 24, 34);",
				"CREATE INDEX v1_idx ON test(v1, v2);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM test WHERE v1 = 2 AND v2 = 22 ORDER BY pk;",
					Expected: []sql.Row{
						{12, 2, 22, 32},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 > 2 AND v2 = 24 ORDER BY pk;",
					Expected: []sql.Row{
						{14, 4, 24, 34},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 >= 4 AND v2 = 25 ORDER BY pk;",
					Expected: []sql.Row{
						{15, 5, 25, 35},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 < 3 AND v2 = 21 ORDER BY pk;",
					Expected: []sql.Row{
						{11, 1, 21, 31},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 <= 3 AND v2 = 22 ORDER BY pk;",
					Expected: []sql.Row{
						{12, 2, 22, 32},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 = 2 AND v2 < 23 ORDER BY pk;",
					Expected: []sql.Row{
						{12, 2, 22, 32},
					},
				},
				{
					Query:    "SELECT * FROM test WHERE v1 = 2 AND v2 < 22 ORDER BY pk;",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test WHERE v1 > 2 AND v2 < 25 ORDER BY pk;",
					Expected: []sql.Row{
						{13, 3, 23, 33},
						{14, 4, 24, 34},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 >= 4 AND v2 <= 24 ORDER BY pk;",
					Expected: []sql.Row{
						{14, 4, 24, 34},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 < 3 AND v2 < 22 ORDER BY pk;",
					Expected: []sql.Row{
						{11, 1, 21, 31},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 <= 3 AND v2 < 23 ORDER BY pk;",
					Expected: []sql.Row{
						{11, 1, 21, 31},
						{12, 2, 22, 32},
					},
				},
			},
		},
		{
			Name: "Unique Non-Covering Composite Index",
			SetUpScript: []string{
				"CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT, v2 BIGINT, v3 BIGINT);",
				"INSERT INTO test VALUES (13, 3, 23, 33), (11, 1, 21, 31), (15, 5, 25, 35), (12, 2, 22, 32), (14, 4, 24, 34);",
				"CREATE UNIQUE INDEX v1_idx ON test(v1, v2);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM test WHERE v1 < 3 AND v2 = 21 ORDER BY pk;",
					Expected: []sql.Row{
						{11, 1, 21, 31},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 <= 3 AND v2 < 23 ORDER BY pk;",
					Expected: []sql.Row{
						{11, 1, 21, 31},
						{12, 2, 22, 32},
					},
				},
				{
					Query:       "insert into test values (16, 3, 23, 33);",
					ExpectedErr: "duplicate unique key given",
				},
			},
		},
		{
			Name: "Keyless Index",
			SetUpScript: []string{
				"CREATE TABLE test (v0 BIGINT, v1 BIGINT, v2 BIGINT);",
				"INSERT INTO test VALUES (13, 3, 23), (11, 1, 21), (15, 5, 25), (12, 2, 22), (14, 4, 24);",
				"CREATE INDEX v1_idx ON test(v1);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM test WHERE v1 = 2 ORDER BY v0;",
					Expected: []sql.Row{
						{12, 2, 22},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 > 2 ORDER BY v0;",
					Expected: []sql.Row{
						{13, 3, 23},
						{14, 4, 24},
						{15, 5, 25},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 >= 4 ORDER BY v0;",
					Expected: []sql.Row{
						{14, 4, 24},
						{15, 5, 25},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 < 3 ORDER BY v0;",
					Expected: []sql.Row{
						{11, 1, 21},
						{12, 2, 22},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 <= 3 ORDER BY v0;",
					Expected: []sql.Row{
						{11, 1, 21},
						{12, 2, 22},
						{13, 3, 23},
					},
				},
			},
		},
		{
			Name: "Unique Keyless Index",
			SetUpScript: []string{
				"CREATE TABLE test (v0 BIGINT, v1 BIGINT, v2 BIGINT);",
				"INSERT INTO test VALUES (13, 3, 23), (11, 1, 21), (15, 5, 25), (12, 2, 22), (14, 4, 24);",
				"CREATE UNIQUE INDEX v1_idx ON test(v1);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM test WHERE v1 = 2 ORDER BY v0;",
					Expected: []sql.Row{
						{12, 2, 22},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 > 2 ORDER BY v0;",
					Expected: []sql.Row{
						{13, 3, 23},
						{14, 4, 24},
						{15, 5, 25},
					},
				},
				{
					Query:       "INSERT INTO test VALUES (16, 3, 23);",
					ExpectedErr: "duplicate unique key given",
				},
			},
		},
		{
			Name: "Keyless Composite Index",
			SetUpScript: []string{
				"CREATE TABLE test (v0 BIGINT, v1 BIGINT, v2 BIGINT, v3 BIGINT);",
				"INSERT INTO test VALUES (13, 3, 23, 33), (11, 1, 21, 31), (15, 5, 25, 35), (12, 2, 22, 32), (14, 4, 24, 34);",
				"CREATE INDEX v1_idx ON test(v1, v2);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM test WHERE v1 = 2 AND v2 = 22 ORDER BY v0;",
					Expected: []sql.Row{
						{12, 2, 22, 32},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 > 2 AND v2 = 24 ORDER BY v0;",
					Expected: []sql.Row{
						{14, 4, 24, 34},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 >= 4 AND v2 = 25 ORDER BY v0;",
					Expected: []sql.Row{
						{15, 5, 25, 35},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 < 3 AND v2 = 21 ORDER BY v0;",
					Expected: []sql.Row{
						{11, 1, 21, 31},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 <= 3 AND v2 = 22 ORDER BY v0;",
					Expected: []sql.Row{
						{12, 2, 22, 32},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 = 2 AND v2 < 23 ORDER BY v0;",
					Expected: []sql.Row{
						{12, 2, 22, 32},
					},
				},
				{
					Query:    "SELECT * FROM test WHERE v1 = 2 AND v2 < 22 ORDER BY v0;",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test WHERE v1 > 2 AND v2 < 25 ORDER BY v0;",
					Expected: []sql.Row{
						{13, 3, 23, 33},
						{14, 4, 24, 34},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 >= 4 AND v2 <= 24 ORDER BY v0;",
					Expected: []sql.Row{
						{14, 4, 24, 34},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 < 3 AND v2 < 22 ORDER BY v0;",
					Expected: []sql.Row{
						{11, 1, 21, 31},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 <= 3 AND v2 < 23 ORDER BY v0;",
					Expected: []sql.Row{
						{11, 1, 21, 31},
						{12, 2, 22, 32},
					},
				},
			},
		},
		{
			Name: "Unique Keyless Composite Index",
			SetUpScript: []string{
				"CREATE TABLE test (v0 BIGINT, v1 BIGINT, v2 BIGINT, v3 BIGINT);",
				"INSERT INTO test VALUES (13, 3, 23, 33), (11, 1, 21, 31), (15, 5, 25, 35), (12, 2, 22, 32), (14, 4, 24, 34);",
				"CREATE UNIQUE INDEX v1_idx ON test(v1, v2);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM test WHERE v1 = 2 AND v2 < 23 ORDER BY v0;",
					Expected: []sql.Row{
						{12, 2, 22, 32},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 <= 3 AND v2 < 23 ORDER BY v0;",
					Expected: []sql.Row{
						{11, 1, 21, 31},
						{12, 2, 22, 32},
					},
				},
				{
					Query:       "insert into test values (16, 3, 23, 33);",
					ExpectedErr: "duplicate unique key given",
				},
			},
		},
		{
			Name: "Indexed Join Covering Indexes",
			SetUpScript: []string{
				"CREATE TABLE test1 (pk BIGINT PRIMARY KEY, v1 BIGINT, v2 BIGINT);",
				"CREATE TABLE test2 (pk BIGINT PRIMARY KEY, v1 BIGINT, v2 BIGINT);",
				"INSERT INTO test1 VALUES (13, 3, 23), (11, 1, 21), (15, 5, 25), (12, 2, 22), (14, 4, 24);",
				"INSERT INTO test2 VALUES (33, 3, 43), (31, 1, 41), (35, 5, 45), (32, 2, 42), (37, 7, 47);",
				"CREATE INDEX v1_idx ON test1(v1);",
				"CREATE INDEX v2_idx ON test2(v1);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT t1.pk, t2.pk FROM test1 t1 JOIN test2 t2 ON t1.v1 = t2.v1 ORDER BY t1.v1;",
					Expected: []sql.Row{
						{11, 31},
						{12, 32},
						{13, 33},
						{15, 35},
					},
				},
				{
					Query: "SELECT t1.pk, t2.pk FROM test1 t1, test2 t2 WHERE t1.v1 = t2.v1 ORDER BY t1.v1;",
					Expected: []sql.Row{
						{11, 31},
						{12, 32},
						{13, 33},
						{15, 35},
					},
				},
			},
		},
		{
			Name: "Unsupported options",
			SetUpScript: []string{
				"CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 varchar);",
			},
			Assertions: []ScriptTestAssertion{
				{
					// ignored warning-generating unsupported options
					Query: "CREATE INDEX v1_idx ON test(v1 varchar_pattern_ops) WITH (storage_opt1 = foo) TABLESPACE tablespace_name;",
				},
				{
					Query:       "CREATE INDEX v1_idx2 ON test( (concat(v1, v1)) ) ;",
					ExpectedErr: "not yet supported",
				},
				{
					Query:       "CREATE INDEX v1_idx2 ON test using hash (v1);",
					ExpectedErr: "not yet supported",
				},
				{
					Query:       "CREATE INDEX v1_idx2 ON test(v1) WHERE v1 > 100;",
					ExpectedErr: "not yet supported",
				},
				{
					Query:       "CREATE INDEX v1_idx2 ON test(v1) INCLUDE (pk);",
					ExpectedErr: "not yet supported",
				},
			},
		},
		{
			Name: "multi column int index",
			SetUpScript: []string{
				`CREATE TABLE test (pk INT4 PRIMARY KEY, a int, b int);`,
				`ALTER TABLE test ADD CONSTRAINT uniqIdx UNIQUE (a, b);`,
				`INSERT INTO test VALUES (1, 1, 2);`,
				`insert into test values (2, 1, 3)`,
				`insert into test values (3, 2, 2);`,
				`insert into test values (4, 3, 1);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT pk FROM test WHERE a = 2 and b = 2;`,
					Expected: []sql.Row{
						{3},
					},
				},
				{
					Query: `SELECT pk FROM test WHERE a > 1`,
					Expected: []sql.Row{
						{3},
						{4},
					},
				},
				{
					Query: `SELECT pk FROM test WHERE a = 2 and b < 3`,
					Expected: []sql.Row{
						{3},
					},
				},
				{
					Query: `SELECT pk FROM test WHERE a > 2 and b < 3`,
					Expected: []sql.Row{
						{4},
					},
				},
				{
					Query: `SELECT pk FROM test WHERE a > 2 and b < 2`,
					Expected: []sql.Row{
						{4},
					},
				},
				{
					Query:    `SELECT pk FROM test WHERE a > 3 and b < 2`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT pk FROM test WHERE a > 3 and b < 2`,
					Expected: []sql.Row{},
				},
				{
					Query: `SELECT pk FROM test WHERE a > 1 and b > 1`,
					Expected: []sql.Row{
						{3},
					},
				},
				{
					Query: `SELECT pk FROM test WHERE a > 1 and b = 1`,
					Expected: []sql.Row{
						{4},
					},
				},
				{
					Query: `SELECT pk FROM test WHERE a < 3 and b > 0 order by 1`,
					Expected: []sql.Row{
						{1},
						{2},
						{3},
					},
				},
				{
					Query: `SELECT pk FROM test WHERE a > 1 and a < 3 order by 1`,
					Expected: []sql.Row{
						{3},
					},
				},
				{
					Query: `SELECT pk FROM test WHERE a > 1 and a < 3 order by 1`,
					Expected: []sql.Row{
						{3},
					},
				},
				{
					Query: `SELECT pk FROM test WHERE a > 1 and b > 1 order by 1`,
					Expected: []sql.Row{
						{3},
					},
				},
			},
		},
		{
			Name: "multi column int index, part 2",
			SetUpScript: []string{
				`CREATE TABLE test (pk INT4 PRIMARY KEY, a int, b int);`,
				`ALTER TABLE test ADD CONSTRAINT uniqIdx UNIQUE (a, b);`,
				`INSERT INTO test VALUES (1, 1, 2);`,
				`insert into test values (2, 1, 3)`,
				`insert into test values (3, 2, 2);`,
				`insert into test values (4, 2, 3);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT pk FROM test WHERE a = 2 and b = 2;`,
					Expected: []sql.Row{
						{3},
					},
				},
				{
					Query: `SELECT pk FROM test WHERE a = 2 and b = 3;`,
					Expected: []sql.Row{
						{4},
					},
				},
			},
		},
		{
			Name: "multi column int index, reverse traversal",
			SetUpScript: []string{
				`CREATE TABLE test (pk INT4 PRIMARY KEY, a int, b int);`,
				`ALTER TABLE test ADD CONSTRAINT uniqIdx UNIQUE (a, b);`,
				`INSERT INTO test VALUES (1, 1, 1);`,
				`insert into test values (2, 1, 3)`,
				`insert into test values (3, 2, 2);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT pk FROM test WHERE a < 3 and b = 2 order by a desc, b desc;`,
					Expected: []sql.Row{
						{3},
					},
				},
				{
					Query: `SELECT pk FROM test WHERE a < 2 and b = 3 order by a desc, b desc;`,
					Expected: []sql.Row{
						{2},
					},
				},
				{
					Query: `SELECT pk FROM test WHERE a < 2 and b < 10 order by a desc, b desc;`,
					Expected: []sql.Row{
						{2},
						{1},
					},
				},
			},
		},
		{
			Name: "Unique index varchar",
			SetUpScript: []string{
				`CREATE TABLE test (pk INT4 PRIMARY KEY, v1 varchar(100), v2 varchar(100));`,
				`ALTER TABLE test ADD CONSTRAINT uniqIdx UNIQUE (v1, v2);`,
				`INSERT INTO test VALUES (1, 'a', 'b');`,
				`insert into test values (2, 'a', 'u')`,
				`insert into test values (3, 'c', 'c');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT pk FROM test WHERE (v1 = 'c' AND v2 = 'c');`,
					Expected: []sql.Row{
						{3},
					},
				},
			},
		},
		{
			Name: "unique index select",
			SetUpScript: []string{
				`CREATE TABLE "django_content_type" ("id" integer NOT NULL PRIMARY KEY GENERATED BY DEFAULT AS IDENTITY, "name" varchar(100) NOT NULL, "app_label" varchar(100) NOT NULL, "model" varchar(100) NOT NULL);`,
				`ALTER TABLE "django_content_type" ADD CONSTRAINT "django_content_type_app_label_model_76bd3d3b_uniq" UNIQUE ("app_label", "model");`,
				`ALTER TABLE "django_content_type" ALTER COLUMN "name" DROP NOT NULL;`,
				`ALTER TABLE "django_content_type" DROP COLUMN "name" CASCADE;`,
				`INSERT INTO "django_content_type" ("app_label", "model") VALUES ('auth', 'permission'), ('auth', 'group'), ('auth', 'user') RETURNING "django_content_type"."id";`,
				`INSERT INTO "django_content_type" ("app_label", "model") VALUES ('contenttypes', 'contenttype') RETURNING "django_content_type"."id";`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT "django_content_type"."id", "django_content_type"."app_label", "django_content_type"."model" FROM "django_content_type" WHERE ("django_content_type"."app_label" = 'auth' AND "django_content_type"."model" = 'permission') LIMIT 21;`,
					Expected: []sql.Row{
						{1, "auth", "permission"},
					},
				},
				{
					Query: `SELECT "django_content_type"."id", "django_content_type"."app_label", "django_content_type"."model" FROM "django_content_type" WHERE ("django_content_type"."app_label" = 'auth' AND "django_content_type"."model" = 'group') LIMIT 21;`,
					Expected: []sql.Row{
						{2, "auth", "group"},
					},
				},
				{
					Query: `SELECT "django_content_type"."id", "django_content_type"."app_label", "django_content_type"."model" FROM "django_content_type" WHERE ("django_content_type"."app_label" = 'auth' AND "django_content_type"."model" = 'user') LIMIT 21;`,
					Expected: []sql.Row{
						{3, "auth", "user"},
					},
				},
				{
					Query: `SELECT "django_content_type"."id", "django_content_type"."app_label", "django_content_type"."model" FROM "django_content_type" WHERE ("django_content_type"."app_label" = 'contenttypes' AND "django_content_type"."model" = 'contenttype') LIMIT 21;`,
					Expected: []sql.Row{
						{4, "contenttypes", "contenttype"},
					},
				},
			},
		},
		{
			Name: "Proper range AND + OR handling",
			SetUpScript: []string{
				"CREATE TABLE test(pk INTEGER PRIMARY KEY, v1 INTEGER);",
				"INSERT INTO test VALUES (1, 1),  (2, 3),  (3, 5),  (4, 7),  (5, 9);",
				"CREATE INDEX v1_idx ON test(v1);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM test WHERE v1 BETWEEN 3 AND 5 OR v1 BETWEEN 7 AND 9;",
					Expected: []sql.Row{
						{2, 3},
						{3, 5},
						{4, 7},
						{5, 9},
					},
				},
			},
		},
		{
			Name: "Performance Regression Test #1",
			SetUpScript: []string{
				"CREATE TABLE sbtest1(id SERIAL, k INTEGER DEFAULT '0' NOT NULL, c CHAR(120) DEFAULT '' NOT NULL, pad CHAR(60) DEFAULT '' NOT NULL, PRIMARY KEY (id))",
				testdata.INDEX_PERFORMANCE_REGRESSION_INSERTS,
				"CREATE INDEX k_1 ON sbtest1(k)",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT id, k FROM sbtest1 WHERE k BETWEEN 3708 AND 3713 OR k BETWEEN 5041 AND 5046;",
					Expected: []sql.Row{
						{2, 5041},
						{18, 5041},
						{57, 5046},
						{58, 5044},
						{79, 5045},
						{80, 5041},
						{81, 5045},
						{107, 5041},
						{113, 5044},
						{153, 5043},
						{167, 5043},
						{187, 5044},
						{210, 5046},
						{213, 5046},
						{216, 5041},
						{222, 5045},
						{238, 5043},
						{265, 5042},
						{269, 5046},
						{279, 5045},
						{295, 5042},
						{298, 5045},
						{309, 5044},
						{324, 3710},
						{348, 5042},
						{353, 5045},
						{374, 5045},
						{390, 5042},
						{400, 5045},
						{430, 5045},
						{445, 5044},
						{476, 5046},
						{496, 5045},
						{554, 5042},
						{565, 5043},
						{566, 5045},
						{571, 5046},
						{573, 5046},
						{582, 5043},
					},
				},
			},
		},
	})
}
