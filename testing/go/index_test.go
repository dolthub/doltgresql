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
					Query: "DELETE FROM test WHERE v1 = 'twelve'",
					SkipResultsCheck: true,
				},
				{
					Query: "SELECT * FROM test WHERE v1 = 'twelve' ORDER BY pk;",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "String primary key ordering",
			Skip: true, // string primary key ordering is broken
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
			Name: "Covering Composite Index",
			SetUpScript: []string{
				"CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT, v2 BIGINT);",
				"INSERT INTO test VALUES (13, 3, 23), (11, 1, 21), (15, 5, 25), (12, 2, 22), (14, 4, 24);",
				"CREATE INDEX v1_idx ON test(v1, v2);",
			},
			Assertions: []ScriptTestAssertion{
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
	})
}
