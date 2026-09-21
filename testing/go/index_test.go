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

	"github.com/dolthub/doltgresql/testing/go/testdata"
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
					Query: "explain SELECT * FROM test WHERE v1 = 2 ORDER BY pk;",
					Expected: []sql.Row{
						{"Sort(test.pk ASC)"},
						{" └─ IndexedTableAccess(test)"},
						{"     ├─ index: [test.v1]"},
						{"     ├─ filters: [{[2, 2]}]"},
						{"     └─ columns: [pk v1]"},
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
					Query: "explain SELECT * FROM test WHERE v1 > 2 ORDER BY pk;",
					Expected: []sql.Row{
						{"Sort(test.pk ASC)"},
						{" └─ IndexedTableAccess(test)"},
						{"     ├─ index: [test.v1]"},
						{"     ├─ filters: [{(2, ∞)}]"},
						{"     └─ columns: [pk v1]"},
					},
				},
				{
					Query: "SELECT * FROM test WHERE (v1 > 3 OR v1 < 2) AND v1 <> 5 ORDER BY pk;",
					Expected: []sql.Row{
						{11, 1},
						{14, 4}},
				},
				{
					Query: "explain SELECT * FROM test WHERE (v1 > 3 OR v1 < 2) AND v1 <> 5 ORDER BY pk;",
					Expected: []sql.Row{
						{"Sort(test.pk ASC)"},
						{" └─ IndexedTableAccess(test)"},
						{"     ├─ index: [test.v1]"},
						{"     ├─ filters: [{(NULL, 2)}, {(3, 5)}, {(5, ∞)}]"},
						{"     └─ columns: [pk v1]"},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 = 2 OR v1 = 4 ORDER BY pk;",
					Expected: []sql.Row{
						{12, 2},
						{14, 4},
					},
				},
				{
					Query: "explain SELECT * FROM test WHERE v1 = 2 OR v1 = 4 ORDER BY pk;",
					Expected: []sql.Row{
						{"Sort(test.pk ASC)"},
						{" └─ IndexedTableAccess(test)"},
						{"     ├─ index: [test.v1]"},
						{"     ├─ filters: [{[2, 2]}, {[4, 4]}]"},
						{"     └─ columns: [pk v1]"},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 IN (2, 4) ORDER BY pk;",
					Expected: []sql.Row{
						{12, 2},
						{14, 4},
					},
				},
				{
					Query: "explain SELECT * FROM test WHERE v1 IN (2, 4) ORDER BY pk;",
					Expected: []sql.Row{
						{"Sort(test.pk ASC)"},
						{" └─ IndexedTableAccess(test)"},
						{"     ├─ index: [test.v1]"},
						{"     ├─ filters: [{[2, 2]}, {[4, 4]}]"},
						{"     └─ columns: [pk v1]"},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 NOT IN (2, 4) ORDER BY pk;",
					Expected: []sql.Row{
						{11, 1},
						{13, 3},
						{15, 5},
					},
				},
				{
					Query: "explain SELECT * FROM test WHERE v1 NOT IN (2, 4) ORDER BY pk;",
					Expected: []sql.Row{
						{"Sort(test.pk ASC)"},
						{" └─ IndexedTableAccess(test)"},
						{"     ├─ index: [test.v1]"},
						{"     ├─ filters: [{(NULL, 2)}, {(2, 4)}, {(4, ∞)}]"},
						{"     └─ columns: [pk v1]"},
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
					Query: "explain SELECT * FROM test WHERE v1 >= 4 ORDER BY pk;",
					Expected: []sql.Row{
						{"Sort(test.pk ASC)"},
						{" └─ IndexedTableAccess(test)"},
						{"     ├─ index: [test.v1]"},
						{"     ├─ filters: [{[4, ∞)}]"},
						{"     └─ columns: [pk v1]"},
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
					Query: "SELECT * FROM test WHERE v1 > 't' OR v1 < 'f' ORDER BY pk;",
					Expected: []sql.Row{
						{11, "eleven"},
						{12, "twelve"},
						{13, "thirteen"},
					},
				},
				{
					Query: "explain SELECT * FROM test WHERE v1 > 't' OR v1 < 'f' ORDER BY pk;",
					Expected: []sql.Row{
						{"Sort(test.pk ASC)"},
						{" └─ IndexedTableAccess(test)"},
						{"     ├─ index: [test.pk,test.v1]"},
						{"     ├─ filters: [{[NULL, ∞), (NULL, f)}, {[NULL, ∞), (t, ∞)}]"},
						{"     └─ columns: [pk v1]"},
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
			Name: "Covering Composite Index",
			SetUpScript: []string{
				"CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT, v2 BIGINT);",
				"INSERT INTO test VALUES (13, 3, 23), (11, 1, 21), (15, 5, 25), (12, 2, 22), (14, 4, 24), (16, 2, 25);",
				"CREATE INDEX v1_v2_idx ON test(v1, v2);",
				"CREATE TABLE jointable (v3 bigint, v4 bigint)",
				"INSERT INTO jointable VALUES (2, 22)",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM test WHERE v1 = 2 AND v2 = 22 ORDER BY pk;",
					Expected: []sql.Row{
						{12, 2, 22},
					},
				},
				{
					Query: "explain SELECT * FROM test WHERE v1 = 2 AND v2 = 22 ORDER BY pk;",
					Expected: []sql.Row{
						{"Sort(test.pk ASC)"},
						{" └─ IndexedTableAccess(test)"},
						{"     ├─ index: [test.v1,test.v2]"},
						{"     ├─ filters: [{[2, 2], [22, 22]}]"},
						{"     └─ columns: [pk v1 v2]"},
					},
				},
				{
					Query: "select /*+ lookup_join(jointable, test) */ HINT * from test join jointable on test.v1 = jointable.v3 and test.v2 = 22 order by 1",
					Expected: []sql.Row{
						{12, 2, 22, 2, 22},
					},
				},
				{
					Query: "explain select * from test join jointable on test.v1 = jointable.v3 and test.v2 = 22 order by 1",
					Expected: []sql.Row{
						{"InnerJoin"},
						{" ├─ test.v1 = jointable.v3"},
						{" ├─ Filter"},
						{" │   ├─ test.v2 = 22"},
						{" │   └─ IndexedTableAccess(test)"},
						{" │       ├─ index: [test.pk]"},
						{" │       ├─ filters: [{[NULL, ∞)}]"},
						{" │       └─ columns: [pk v1 v2]"},
						{" └─ Table"},
						{"     ├─ name: jointable"},
						{"     └─ columns: [v3 v4]"},
					},
				},
				{
					Query: "select * from test join jointable on test.v1 = jointable.v3 and test.v2 = jointable.v4 order by 1",
					Expected: []sql.Row{
						{12, 2, 22, 2, 22},
					},
				},
				{
					Query: "explain select * from test join jointable on test.v1 = jointable.v3 and test.v2 = jointable.v4 order by 1",
					Expected: []sql.Row{
						{"InnerJoin"},
						{" ├─ (test.v1 = jointable.v3 AND test.v2 = jointable.v4)"},
						{" ├─ IndexedTableAccess(test)"},
						{" │   ├─ index: [test.pk]"},
						{" │   ├─ filters: [{[NULL, ∞)}]"},
						{" │   └─ columns: [pk v1 v2]"},
						{" └─ Table"},
						{"     ├─ name: jointable"},
						{"     └─ columns: [v3 v4]"},
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
			// TODO: lookups when the join key is specified by a subquery
			Name: "Covering Composite Index join, different types",
			SetUpScript: []string{
				"CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 smallint, v2 smallint);",
				"INSERT INTO test VALUES (13, 3, 23), (11, 1, 21), (15, 5, 25), (12, 2, 22), (14, 4, 24), (16, 2, 25);",
				"CREATE INDEX v1_v2_idx ON test(v1, v2);",
				"CREATE TABLE jointable (v3 bigint, v4 bigint)",
				"INSERT INTO jointable VALUES (2, 22)",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "select /*+ lookup_join(jointable, test) */ HINT * from test join jointable on test.v1 = jointable.v3 and test.v2 = 22 order by 1",
					Expected: []sql.Row{
						{12, 2, 22, 2, 22},
					},
				},
				{
					// TODO: Unskip once matched filter expressions are removed from filter nodes
					//  https://github.com/dolthub/dolt/issues/11231
					Skip:  true,
					Query: "explain select /*+ lookup_join(jointable, test) */ HINT * from test join jointable on test.v1 = jointable.v3 and test.v2 = 22 order by 1",
					Expected: []sql.Row{
						{"Project"},
						{" ├─ columns: [test.pk, test.v1, test.v2, jointable.v3, jointable.v4]"},
						{" └─ Sort(test.pk ASC)"},
						{"     └─ LookupJoin"},
						{"         ├─ Table"},
						{"         │   ├─ name: jointable"},
						{"         │   └─ columns: [v3 v4]"},
						{"         └─ IndexedTableAccess(test)"},
						{"             ├─ index: [test.v1,test.v2]"},
						{"             ├─ columns: [pk v1 v2]"},
						{"             └─ keys: jointable.v3, 22"},
					},
				},
				{
					Query: "select /*+ lookup_join(jointable, test) */ HINT * from test join jointable on test.v1 = jointable.v3 and test.v2 = jointable.v4 order by 1",
					Expected: []sql.Row{
						{12, 2, 22, 2, 22},
					},
				},
				{
					Query: "explain select /*+ lookup_join(jointable, test) */ HINT * from test join jointable on test.v1 = jointable.v3 and test.v2 = jointable.v4 order by 1",
					Expected: []sql.Row{
						{"Project"},
						{" ├─ columns: [test.pk, test.v1, test.v2, jointable.v3, jointable.v4]"},
						{" └─ Sort(test.pk ASC)"},
						{"     └─ LookupJoin"},
						{"         ├─ Table"},
						{"         │   ├─ name: jointable"},
						{"         │   └─ columns: [v3 v4]"},
						{"         └─ IndexedTableAccess(test)"},
						{"             ├─ index: [test.v1,test.v2]"},
						{"             ├─ columns: [pk v1 v2]"},
						{"             └─ keys: jointable.v3, jointable.v4"},
					},
				},
			},
		},
		{
			Name: "Covering Composite Index join, different types out of range",
			SetUpScript: []string{
				"CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 smallint, v2 smallint);",
				// The zero value in the last row is important because it catches an error mode in index lookup creation failure
				"INSERT INTO test VALUES (13, 3, 23), (11, 1, 21), (14, 0, 22)",
				"CREATE INDEX v1_v2_idx ON test(v1, v2);",
				"CREATE TABLE jointable (v3 bigint, v4 bigint)",
				"INSERT INTO jointable VALUES (2147483648, 2147483649), (1, 21)",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "select /*+ lookup_join(jointable, test) */ HINT * from test join jointable on test.v1 = jointable.v3 and test.v2 = 22 order by 1",
					Expected: []sql.Row{},
				},
				{
					Query: "select /*+ lookup_join(jointable, test) */ HINT * from test join jointable on test.v1 = jointable.v3 and test.v2 = 21 order by 1",
					Expected: []sql.Row{
						{11, 1, 21, 1, 21},
					},
				},
				{
					Query: "explain select * from test join jointable on test.v1 = jointable.v3 and test.v2 = 22 order by 1",
					Expected: []sql.Row{
						{"InnerJoin"},
						{" ├─ test.v1 = jointable.v3"},
						{" ├─ Filter"},
						{" │   ├─ test.v2 = 22"},
						{" │   └─ IndexedTableAccess(test)"},
						{" │       ├─ index: [test.pk]"},
						{" │       ├─ filters: [{[NULL, ∞)}]"},
						{" │       └─ columns: [pk v1 v2]"},
						{" └─ Table"},
						{"     ├─ name: jointable"},
						{"     └─ columns: [v3 v4]"},
					},
				},
			},
		},
		{
			Name: "Covering Composite Index join, subquery",
			SetUpScript: []string{
				"CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 smallint, v2 smallint);",
				"INSERT INTO test VALUES (13, 3, 23), (11, 1, 21), (14, 0, 22)",
				"CREATE INDEX v1_v2_idx ON test(v1, v2);",
				"CREATE TABLE jointable (v3 bigint, v4 bigint)",
				"INSERT INTO jointable VALUES (2, 22), (1, 21), (2147483648, 22)",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "select /*+ lookup_join(sq, test) */ HINT * from test join " +
						"(select * from jointable) sq " +
						"on test.v1 = sq.v3 and test.v2 = sq.v4 order by 1",
					Expected: []sql.Row{
						{11, 1, 21, 1, 21},
					},
				},
				{
					Query: "explain select * from test join (select * from jointable) sq on test.v1 = sq.v3 and test.v2 = sq.v4 order by 1",
					Expected: []sql.Row{
						{"InnerJoin"},
						{" ├─ (test.v1 = sq.v3 AND test.v2 = sq.v4)"},
						{" ├─ IndexedTableAccess(test)"},
						{" │   ├─ index: [test.pk]"},
						{" │   ├─ filters: [{[NULL, ∞)}]"},
						{" │   └─ columns: [pk v1 v2]"},
						{" └─ TableAlias(sq)"},
						{"     └─ Table"},
						{"         ├─ name: jointable"},
						{"         └─ columns: [v3 v4]"},
					},
				},
				{
					Query: "explain select /*+ lookup_join(sq, test) */ HINT * from test join (select * from jointable) sq on test.v1 = sq.v3 and test.v2 = sq.v4 order by 1",
					Expected: []sql.Row{
						{"Project"},
						{" ├─ columns: [test.pk, test.v1, test.v2, sq.v3, sq.v4]"},
						{" └─ Sort(test.pk ASC)"},
						{"     └─ LookupJoin"},
						{"         ├─ TableAlias(sq)"},
						{"         │   └─ Table"},
						{"         │       ├─ name: jointable"},
						{"         │       └─ columns: [v3 v4]"},
						{"         └─ IndexedTableAccess(test)"},
						{"             ├─ index: [test.v1,test.v2]"},
						{"             ├─ columns: [pk v1 v2]"},
						{"             └─ keys: sq.v3, sq.v4"},
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
					Query: "explain SELECT * FROM test WHERE v1 IN (2, '3', 4) ORDER BY v1;",
					Expected: []sql.Row{
						{"IndexedTableAccess(test)"},
						{" ├─ index: [test.v1]"},
						{" ├─ filters: [{[2, 2]}, {[3, 3]}, {[4, 4]}]"},
						{" └─ columns: [pk v1 v2]"},
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
					Query:       "CREATE INDEX v1_idx2 ON test using hash (v1);",
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
				{
					Query: "explain SELECT * FROM test WHERE v1 BETWEEN 3 AND 5 OR v1 BETWEEN 7 AND 9 order by 1;",
					Expected: []sql.Row{
						{"Sort(test.pk ASC)"},
						{" └─ IndexedTableAccess(test)"},
						{"     ├─ index: [test.v1]"},
						{"     ├─ filters: [{[3, 5]}, {[7, 9]}]"},
						{"     └─ columns: [pk v1]"},
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
		{
			Name: "Index names must be unique across all relation types",
			SetUpScript: []string{
				"CREATE TABLE t1 (pk int PRIMARY KEY, v1 int);",
				"CREATE TABLE t2 (pk int PRIMARY KEY, v1 int);",
				"CREATE TABLE tbl1 (pk int PRIMARY KEY, v1 int);",
				"CREATE SEQUENCE seq1;",
				"CREATE VIEW view1 AS SELECT pk FROM t1;",
				"CREATE INDEX idx1 ON t1 (v1);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "CREATE INDEX idx_unique ON t1 (v1);",
					Expected: []sql.Row{},
				},
				{
					Query:    "CREATE INDEX IF NOT EXISTS idx_unique ON t1 (v1);",
					Expected: []sql.Row{},
				},
				{
					Query:       "CREATE INDEX idx_unique ON t2 (v1);",
					ExpectedErr: `relation "idx_unique" already exists`,
				},
				{
					Query:    "CREATE INDEX IF NOT EXISTS idx_unique ON t2 (v1);",
					Expected: []sql.Row{},
				},
				{
					Query:       "CREATE INDEX tbl1 ON t2 (v1);",
					ExpectedErr: `relation "tbl1" already exists`,
				},
				{
					Query:    "CREATE INDEX IF NOT EXISTS tbl1 ON t2 (v1);",
					Expected: []sql.Row{},
				},
				{
					Query:       "CREATE INDEX seq1 ON t2 (v1);",
					ExpectedErr: `relation "seq1" already exists`,
				},
				{
					Query:    "CREATE INDEX IF NOT EXISTS seq1 ON t2 (v1);",
					Expected: []sql.Row{},
				},
				{
					Query:       "CREATE INDEX view1 ON t2 (v1);",
					ExpectedErr: `relation "view1" already exists`,
				},
				{
					Query:    "CREATE INDEX IF NOT EXISTS view1 ON t2 (v1);",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "DROP INDEX",
			SetUpScript: []string{
				"CREATE TABLE t (pk int PRIMARY KEY, v1 int);",
				"CREATE INDEX v1_idx ON t (v1);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "DROP INDEX v1_idx;",
					Expected: []sql.Row{},
				},
				{
					Query:       "DROP INDEX v1_idx;",
					ExpectedErr: "unable to find index",
				},
				{
					Query:       "DROP INDEX no_such_index;",
					ExpectedErr: "unable to find index",
				},
			},
		},
		{
			Name: "DROP INDEX IF EXISTS",
			SetUpScript: []string{
				"CREATE TABLE t (pk int PRIMARY KEY, v1 int);",
				"CREATE INDEX v1_idx ON t (v1);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "DROP INDEX IF EXISTS v1_idx;",
					Expected: []sql.Row{},
				},
				{
					Query:    "DROP INDEX IF EXISTS v1_idx;",
					Expected: []sql.Row{},
				},
				{
					Query:    "DROP INDEX IF EXISTS no_such_index;",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "DROP INDEX removes index from query plan",
			SetUpScript: []string{
				"CREATE TABLE t (pk int PRIMARY KEY, v1 int);",
				"INSERT INTO t VALUES (1, 10), (2, 20), (3, 30);",
				"CREATE INDEX v1_idx ON t (v1);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "EXPLAIN SELECT * FROM t WHERE v1 = 20;",
					Expected: []sql.Row{
						{"IndexedTableAccess(t)"},
						{" ├─ index: [t.v1]"},
						{" ├─ filters: [{[20, 20]}]"},
						{" └─ columns: [pk v1]"},
					},
				},
				{
					Query:    "DROP INDEX v1_idx;",
					Expected: []sql.Row{},
				},
				{
					Query: "EXPLAIN SELECT * FROM t WHERE v1 = 20;",
					Expected: []sql.Row{
						{"Filter"},
						{" ├─ t.v1 = 20"},
						{" └─ Table"},
						{"     ├─ name: t"},
						{"     └─ columns: [pk v1]"},
					},
				},
			},
		},
		{
			Name: "DROP INDEX is case-insensitive on index name",
			SetUpScript: []string{
				"CREATE TABLE t (pk int PRIMARY KEY, v1 int);",
				`CREATE INDEX "idx_Mixed" ON t (v1);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `DROP INDEX "IDX_MIXED";`,
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "ALTER INDEX RENAME TO",
			SetUpScript: []string{
				"CREATE TABLE t (pk int PRIMARY KEY, v1 int);",
				"INSERT INTO t VALUES (1, 10), (2, 20), (3, 30);",
				"CREATE INDEX v1_idx ON t (v1);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "ALTER INDEX v1_idx RENAME TO v1_renamed_idx;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT indexname FROM pg_indexes WHERE tablename = 't' AND indexname <> 't_pkey';",
					Expected: []sql.Row{{"v1_renamed_idx"}},
				},
				{
					// The renamed index is still used in query plans
					Query: "EXPLAIN SELECT * FROM t WHERE v1 = 20;",
					Expected: []sql.Row{
						{"IndexedTableAccess(t)"},
						{" ├─ index: [t.v1]"},
						{" ├─ filters: [{[20, 20]}]"},
						{" └─ columns: [pk v1]"},
					},
				},
				{
					Query:    "SELECT * FROM t WHERE v1 = 20;",
					Expected: []sql.Row{{2, 20}},
				},
				{
					// The old name is gone
					Query:       "ALTER INDEX v1_idx RENAME TO something_else;",
					ExpectedErr: `relation "v1_idx" does not exist`,
				},
				{
					// The new name is usable by DROP INDEX
					Query:    "DROP INDEX v1_renamed_idx;",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "ALTER INDEX RENAME TO with schema-qualified name",
			SetUpScript: []string{
				"CREATE SCHEMA myschema;",
				"CREATE TABLE myschema.t (pk int PRIMARY KEY, v1 int);",
				"CREATE INDEX v1_idx ON myschema.t (v1);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "ALTER INDEX myschema.v1_idx RENAME TO v1_renamed_idx;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT indexname FROM pg_indexes WHERE schemaname = 'myschema' AND indexname <> 't_pkey';",
					Expected: []sql.Row{{"v1_renamed_idx"}},
				},
				{
					// The schema is not on the search path, so the unqualified name can't be found
					Query:       "ALTER INDEX v1_renamed_idx RENAME TO another_name;",
					ExpectedErr: `relation "v1_renamed_idx" does not exist`,
				},
				{
					// A schema that doesn't contain the index errors
					Query:       "ALTER INDEX public.v1_renamed_idx RENAME TO another_name;",
					ExpectedErr: `relation "v1_renamed_idx" does not exist`,
				},
			},
		},
		{
			Name: "ALTER INDEX RENAME TO error cases",
			SetUpScript: []string{
				"CREATE TABLE t (pk int PRIMARY KEY, v1 int, v2 int);",
				"CREATE INDEX v1_idx ON t (v1);",
				"CREATE INDEX v2_idx ON t (v2);",
				"CREATE TABLE t2 (pk int PRIMARY KEY, v1 int);",
				"CREATE INDEX t2_v1_idx ON t2 (v1);",
				"CREATE VIEW myview AS SELECT pk FROM t;",
				"CREATE SEQUENCE myseq;",
				"CREATE SCHEMA otherschema;",
				"CREATE TABLE otherschema.t3 (pk int PRIMARY KEY);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       "ALTER INDEX no_such_index RENAME TO new_name;",
					ExpectedErr: `relation "no_such_index" does not exist`,
				},
				{
					Query:    "ALTER INDEX IF EXISTS no_such_index RENAME TO new_name;",
					Expected: []sql.Row{},
				},
				{
					// Renaming to a name that's already in use on the same table errors
					Query:       "ALTER INDEX v1_idx RENAME TO v2_idx;",
					ExpectedErr: `relation "v2_idx" already exists`,
				},
				{
					// All relations in a schema share a namespace, so an index on another table conflicts too
					Query:       "ALTER INDEX v1_idx RENAME TO t2_v1_idx;",
					ExpectedErr: `relation "t2_v1_idx" already exists`,
				},
				{
					// ... as does a table name
					Query:       "ALTER INDEX v1_idx RENAME TO t2;",
					ExpectedErr: `relation "t2" already exists`,
				},
				{
					// ... a view name
					Query:       "ALTER INDEX v1_idx RENAME TO myview;",
					ExpectedErr: `relation "myview" already exists`,
				},
				{
					// ... a sequence name
					Query:       "ALTER INDEX v1_idx RENAME TO myseq;",
					ExpectedErr: `relation "myseq" already exists`,
				},
				{
					// ... and the index's own current name
					Query:       "ALTER INDEX v1_idx RENAME TO v1_idx;",
					ExpectedErr: `relation "v1_idx" already exists`,
				},
				{
					// Relations in other schemas don't conflict
					Query:    "ALTER INDEX v1_idx RENAME TO t3;",
					Expected: []sql.Row{},
				},
				{
					Query:    "ALTER INDEX t3 RENAME TO v1_idx;",
					Expected: []sql.Row{},
				},
				{
					Query:    "ALTER INDEX IF EXISTS v1_idx RENAME TO v1_renamed_idx;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT indexname FROM pg_indexes WHERE tablename = 't' AND indexname <> 't_pkey' ORDER BY indexname;",
					Expected: []sql.Row{{"v1_renamed_idx"}, {"v2_idx"}},
				},
			},
		},
		{
			Name: "partial index",
			SetUpScript: []string{
				`CREATE TABLE user_sessions (
    session_id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE
);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "CREATE UNIQUE INDEX idx_one_active_session_per_user ON user_sessions (user_id) WHERE is_active = TRUE;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT indexdef FROM pg_indexes WHERE indexname = 'idx_one_active_session_per_user';",
					Expected: []sql.Row{{"CREATE UNIQUE INDEX idx_one_active_session_per_user ON public.user_sessions USING btree (user_id) WHERE (user_sessions.is_active = true)"}},
				},
				{
					Query:    "INSERT INTO user_sessions (user_id, is_active) VALUES (42, true);",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO user_sessions (user_id, is_active) VALUES (99, true);",
					Expected: []sql.Row{},
				},
				{
					Query:       "INSERT INTO user_sessions (user_id, is_active) VALUES (42, true);",
					ExpectedErr: "duplicate unique key given",
				},
				{
					// succeeds because is_active is false
					Query:    "INSERT INTO user_sessions (user_id, is_active) VALUES (42, false);",
					Expected: []sql.Row{},
				},
				{
					// succeeds because is_active is false
					Query:    "INSERT INTO user_sessions (user_id, is_active) VALUES (42, false);",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT * FROM user_sessions;",
					Expected: []sql.Row{{1, 42, "t"}, {2, 99, "t"}, {4, 42, "f"}, {5, 42, "f"}},
				},
				{
					Query:    "SELECT * FROM user_sessions WHERE user_id = 42;",
					Expected: []sql.Row{{1, 42, "t"}, {4, 42, "f"}, {5, 42, "f"}},
				},
				{
					Query:    "SELECT is_active FROM user_sessions WHERE user_id = 42;",
					Expected: []sql.Row{{"t"}, {"f"}, {"f"}},
				},
				{
					Query:    "SELECT count(*) FROM user_sessions WHERE user_id = 42;",
					Expected: []sql.Row{{3}},
				},
				{
					Query:       "UPDATE user_sessions SET is_active = true WHERE user_id = 42 AND is_active = false;",
					ExpectedErr: "duplicate unique key given",
				},
			},
		},
		{
			Name: "partial index on keyless table",
			SetUpScript: []string{
				"CREATE TABLE t (a INT, b INT);",
				"INSERT INTO t VALUES (1, 1), (2, 2), (3, 3);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "CREATE INDEX idx_partial ON t (a) WHERE a > 1;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT indexdef FROM pg_indexes WHERE indexname = 'idx_partial';",
					Expected: []sql.Row{{"CREATE INDEX idx_partial ON public.t USING btree (a) WHERE (t.a > 1)"}},
				},
				{
					Query: "EXPLAIN SELECT * FROM t WHERE a > 1;",
					Expected: []sql.Row{
						{"IndexedTableAccess(t)"},
						{" ├─ index: [t.a,t.a > 1]"},
						{" ├─ filters: [{(1, ∞)}]"},
						{" └─ columns: [a b]"},
					},
				},
				{
					Query: "EXPLAIN SELECT * FROM t WHERE a > 0;",
					Expected: []sql.Row{
						{"Filter"},
						{" ├─ t.a > 0"},
						{" └─ Table"},
						{"     ├─ name: t"},
						{"     └─ columns: [a b]"},
					},
				},
				{
					Query:    "INSERT INTO t VALUES (0, 0);",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO t VALUES (5, 5);",
					Expected: []sql.Row{},
				},
				{
					Query:    "CREATE UNIQUE INDEX idx_uniq_partial ON t (a) WHERE a > 2;",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO t VALUES (1, 99);",
					Expected: []sql.Row{},
				},
				{
					Query:       "INSERT INTO t VALUES (3, 99);",
					ExpectedErr: "duplicate unique key given",
				},
			},
		},
		{
			// The predicate is stored as text and re-compiled by a query built around the table name. That name is
			// unqualified today, so it is resolved against the session's search_path rather than the schema the index
			// actually lives in.
			Name: "partial index on a table whose schema is not on the search path",
			SetUpScript: []string{
				"CREATE TABLE public.cg (id int PRIMARY KEY, org_id int, code text, deleted_at timestamptz);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:            "SELECT pg_catalog.set_config('search_path', '', false);",
					SkipResultsCheck: true,
				},
				{
					Query:    "CREATE UNIQUE INDEX uq_cg ON public.cg USING btree (org_id, code) WHERE ((code IS NOT NULL) AND (deleted_at IS NULL));",
					Expected: []sql.Row{},
				},
				{
					// A second connection, so the writer state is compiled fresh rather than served from the cache
					// the CREATE INDEX above populated.
					Query:            "SELECT pg_catalog.set_config('search_path', '', false);",
					SkipResultsCheck: true,
					Username:         "postgres",
					Password:         "password",
				},
				{
					Query:    "INSERT INTO public.cg (id, org_id, code) VALUES (1, 1, 'a');",
					Expected: []sql.Row{},
					Username: "postgres",
					Password: "password",
				},
				{
					Query:       "INSERT INTO public.cg (id, org_id, code) VALUES (2, 1, 'a');",
					ExpectedErr: "duplicate unique key given",
					Username:    "postgres",
					Password:    "password",
				},
			},
		},
		{
			// Same cause as above, but silent: the predicate compiles against s2.cg, whose columns sit at different
			// positions, so the unique index is simply not enforced.
			Name: "partial index predicate does not resolve to a same-named table in another schema",
			SetUpScript: []string{
				"CREATE SCHEMA s2;",
				"CREATE TABLE public.cg (id int PRIMARY KEY, org_id int, code text, deleted_at text);",
				"CREATE UNIQUE INDEX uq_cg ON public.cg USING btree (org_id, code) WHERE ((code IS NOT NULL) AND (deleted_at IS NULL));",
				"CREATE TABLE s2.cg (deleted_at text, code text, org_id int, id int PRIMARY KEY);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:            "SET search_path TO 's2';",
					SkipResultsCheck: true,
					Username:         "postgres",
					Password:         "password",
				},
				{
					Query:    "INSERT INTO public.cg (id, org_id, code, deleted_at) VALUES (1, 1, 'a', NULL);",
					Expected: []sql.Row{},
					Username: "postgres",
					Password: "password",
				},
				{
					Query:       "INSERT INTO public.cg (id, org_id, code, deleted_at) VALUES (2, 1, 'a', NULL);",
					ExpectedErr: "duplicate unique key given",
					Username:    "postgres",
					Password:    "password",
				},
			},
		},
		{
			Name: "index naming: unnamed index uses table_col_idx convention",
			SetUpScript: []string{
				"CREATE TABLE t1 (pk INT PRIMARY KEY, a INT, b INT);",
				"CREATE INDEX ON t1 (a);",
				"CREATE TABLE t2 (pk INT PRIMARY KEY, a INT);",
				"CREATE UNIQUE INDEX ON t2 (a);",
				"CREATE TABLE t3 (pk INT PRIMARY KEY, a INT UNIQUE);",
				"CREATE TABLE t4 (pk INT PRIMARY KEY, a INT, b INT);",
				"CREATE INDEX ON t4 (a, b);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT indexname FROM pg_indexes WHERE tablename = 't1' AND indexname <> 't1_pkey' ORDER BY indexname;",
					Expected: []sql.Row{{"t1_a_idx"}},
				},
				{
					Query:    "SELECT indexname FROM pg_indexes WHERE tablename = 't2' AND indexname <> 't2_pkey' ORDER BY indexname;",
					Expected: []sql.Row{{"t2_a_key"}},
				},
				{
					Query:    "SELECT indexname FROM pg_indexes WHERE tablename = 't3' AND indexname <> 't3_pkey' ORDER BY indexname;",
					Expected: []sql.Row{{"t3_a_key"}},
				},
				{
					Query:    "SELECT indexname FROM pg_indexes WHERE tablename = 't4' AND indexname <> 't4_pkey' ORDER BY indexname;",
					Expected: []sql.Row{{"t4_a_b_idx"}},
				},
			},
		},
		{
			Name: "index naming: collision appends numeric suffix",
			SetUpScript: []string{
				"CREATE TABLE t5 (pk INT PRIMARY KEY, a INT);",
				"CREATE INDEX t5_a_idx ON t5 (a);",
				"CREATE INDEX ON t5 (a);",
				"CREATE TABLE t6_a_idx (pk INT PRIMARY KEY);",
				"CREATE TABLE t6 (pk INT PRIMARY KEY, a INT);",
				"CREATE INDEX ON t6 (a);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT indexname FROM pg_indexes WHERE tablename = 't5' AND indexname NOT IN ('t5_pkey', 't5_a_idx') ORDER BY indexname;",
					Expected: []sql.Row{{"t5_a_idx1"}},
				},
				{
					Query:    "SELECT indexname FROM pg_indexes WHERE tablename = 't6' AND indexname <> 't6_pkey' ORDER BY indexname;",
					Expected: []sql.Row{{"t6_a_idx1"}},
				},
			},
		},
	})
}

func TestIndexColumnOptions(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "operator classes and descending columns", // https://github.com/dolthub/doltgresql/issues/3229
			SetUpScript: []string{
				"CREATE TABLE index_warning_repro (id BIGINT PRIMARY KEY, owner_id BIGINT, name VARCHAR(100), created_at TIMESTAMPTZ);",
				"CREATE INDEX index_warning_pattern ON index_warning_repro (name varchar_pattern_ops);",
				"CREATE INDEX index_warning_newest ON index_warning_repro (owner_id, created_at DESC, id DESC);",
				"INSERT INTO index_warning_repro VALUES (1, 1, 'b', '2024-01-01 00:00:00+00'), (2, 1, 'a', '2024-01-03 00:00:00+00'), (3, 1, 'c', NULL), (4, 2, 'd', '2024-01-02 00:00:00+00'), (5, 1, 'e', '2024-01-02 00:00:00+00');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT indexname, indexdef FROM pg_indexes WHERE tablename = 'index_warning_repro' ORDER BY indexname;",
					Expected: []sql.Row{
						{"index_warning_newest", "CREATE INDEX index_warning_newest ON public.index_warning_repro USING btree (owner_id, created_at DESC, id DESC)"},
						{"index_warning_pattern", "CREATE INDEX index_warning_pattern ON public.index_warning_repro USING btree (name varchar_pattern_ops)"},
						{"index_warning_repro_pkey", "CREATE UNIQUE INDEX index_warning_repro_pkey ON public.index_warning_repro USING btree (id)"},
					},
				},
				{
					Query:    "SELECT pg_get_indexdef('index_warning_newest'::regclass);",
					Expected: []sql.Row{{"CREATE INDEX index_warning_newest ON public.index_warning_repro USING btree (owner_id, created_at DESC, id DESC)"}},
				},
				{
					Query:    "SELECT array_to_string(indoption, ',') FROM pg_index WHERE indexrelid = 'index_warning_newest'::regclass;",
					Expected: []sql.Row{{"0,3,3"}},
				},
				{ // Postgres assigns operator class OIDs during initdb, Doltgres uses the fixed OIDs from core/id
					Query:    "SELECT array_to_string(indclass, ',') FROM pg_index WHERE indexrelid = 'index_warning_newest'::regclass;",
					Expected: []sql.Row{{"15010,15020,15010"}},
				},
				{
					Query:    "SELECT array_to_string(indoption, ',') || ' ' || array_to_string(indclass, ',') FROM pg_index WHERE indexrelid = 'index_warning_pattern'::regclass;",
					Expected: []sql.Row{{"0 15027"}},
				},
				{
					Query:    "SELECT id FROM index_warning_repro WHERE owner_id = 1 ORDER BY created_at DESC NULLS LAST, id DESC;",
					Expected: []sql.Row{{2}, {5}, {1}, {3}},
				},
				{
					Query:    "SELECT id FROM index_warning_repro WHERE owner_id = 1 ORDER BY created_at DESC, id DESC;",
					Expected: []sql.Row{{3}, {2}, {5}, {1}},
				},
				{
					Query:    "SELECT id FROM index_warning_repro WHERE owner_id = 1 AND created_at < '2024-01-03 00:00:00+00' ORDER BY id;",
					Expected: []sql.Row{{1}, {5}},
				},
				{
					Query:    "SELECT id FROM index_warning_repro WHERE owner_id = 1 AND created_at IS NULL;",
					Expected: []sql.Row{{3}},
				},
				{
					Query:    "SELECT id FROM index_warning_repro WHERE owner_id = 1 AND created_at >= '2024-01-02 00:00:00+00' ORDER BY id;",
					Expected: []sql.Row{{2}, {5}},
				},
				{
					Query:    "SELECT id FROM index_warning_repro WHERE owner_id = 1 AND created_at IS NOT NULL ORDER BY id;",
					Expected: []sql.Row{{1}, {2}, {5}},
				},
				{
					Query:    "SELECT id FROM index_warning_repro WHERE owner_id = 1 AND created_at = '2024-01-02 00:00:00+00' AND id > 1;",
					Expected: []sql.Row{{5}},
				},
				{
					Query:    "SELECT id FROM index_warning_repro WHERE name > 'b' ORDER BY id;",
					Expected: []sql.Row{{3}, {4}, {5}},
				},
				{
					Query:    "SELECT id FROM index_warning_repro WHERE name = 'c';",
					Expected: []sql.Row{{3}},
				},
				{
					Query:    "SELECT id FROM index_warning_repro WHERE name LIKE 'a%';",
					Expected: []sql.Row{{2}},
				},
				{
					Query: "CREATE INDEX ok_default ON index_warning_repro (name text_ops);",
				},
				{
					Query: "CREATE INDEX ok_nondefault ON index_warning_repro (name varchar_ops);",
				},
				{
					Query: "SELECT indexname, indexdef FROM pg_indexes WHERE indexname IN ('ok_default', 'ok_nondefault') ORDER BY indexname;",
					Expected: []sql.Row{
						{"ok_default", "CREATE INDEX ok_default ON public.index_warning_repro USING btree (name)"},
						{"ok_nondefault", "CREATE INDEX ok_nondefault ON public.index_warning_repro USING btree (name varchar_ops)"},
					},
				},
				{
					Query: "CREATE INDEX ok_qualified ON index_warning_repro (name pg_catalog.text_pattern_ops);",
				},
				{
					Query: "CREATE INDEX ok_expr ON index_warning_repro ((lower(name)) text_pattern_ops);",
				},
				{
					Query: "SELECT indexname, indexdef FROM pg_indexes WHERE indexname IN ('ok_qualified', 'ok_expr') ORDER BY indexname;",
					Expected: []sql.Row{
						{"ok_expr", "CREATE INDEX ok_expr ON public.index_warning_repro USING btree ((lower(name)) text_pattern_ops)"},
						{"ok_qualified", "CREATE INDEX ok_qualified ON public.index_warning_repro USING btree (name text_pattern_ops)"},
					},
				},
				{
					Query:    "SELECT array_to_string(indclass, ',') FROM pg_index WHERE indexrelid = 'ok_expr'::regclass;",
					Expected: []sql.Row{{"15026"}},
				},
				{
					Query:    "SELECT id FROM index_warning_repro WHERE lower(name) LIKE 'a%';",
					Expected: []sql.Row{{2}},
				},
				{
					Query: "DROP INDEX ok_expr;",
				},
				{
					Query: "CREATE INDEX ok_expr_default ON index_warning_repro ((lower(name)) text_ops);",
				},
				{
					Query:    "SELECT indexdef FROM pg_indexes WHERE indexname = 'ok_expr_default';",
					Expected: []sql.Row{{"CREATE INDEX ok_expr_default ON public.index_warning_repro USING btree ((lower(name)))"}},
				},
			},
		},
		{
			Name: "operator class errors",
			SetUpScript: []string{
				"CREATE TABLE opclass_errors (id BIGINT PRIMARY KEY, owner_id BIGINT, name VARCHAR(100));",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       "CREATE INDEX bad1 ON opclass_errors (id text_pattern_ops);",
					ExpectedErr: `operator class "text_pattern_ops" does not accept data type bigint`,
				},
				{
					Query:       "CREATE INDEX bad2 ON opclass_errors (name nonexistent_ops);",
					ExpectedErr: `operator class "nonexistent_ops" does not exist for access method "btree"`,
				},
				{
					Query:       "CREATE INDEX bad3 ON opclass_errors (name varchar_pattern_ops (foo = 1));",
					ExpectedErr: "operator class varchar_pattern_ops has no options",
				},
				{
					Query:       "CREATE INDEX bad4 ON opclass_errors (owner_id int4_ops);",
					ExpectedErr: `operator class "int4_ops" does not accept data type bigint`,
				},
				{
					Query:       "CREATE INDEX bad5 ON opclass_errors (name public.text_pattern_ops);",
					ExpectedErr: `operator class "public.text_pattern_ops" does not exist for access method "btree"`,
				},
				{
					Query:       "CREATE INDEX bad6 ON opclass_errors ((id + 1) text_pattern_ops);",
					ExpectedErr: `operator class "text_pattern_ops" does not accept data type bigint`,
				},
				{
					Query:    "SELECT count(*) FROM pg_indexes WHERE tablename = 'opclass_errors';",
					Expected: []sql.Row{{1}},
				},
			},
		},
		{
			Name: "descending indexes on assorted types",
			SetUpScript: []string{
				"CREATE TABLE dt (pk INT PRIMARY KEY, i2 SMALLINT, i8 BIGINT, n NUMERIC(10,2), f8 DOUBLE PRECISION, t TEXT, v VARCHAR(20), c CHAR(3), d DATE, ts TIMESTAMP, tz TIMESTAMPTZ, b BOOLEAN, u UUID);",
				"CREATE INDEX dt_i2 ON dt (i2 DESC);",
				"CREATE INDEX dt_i8 ON dt (i8 DESC);",
				"CREATE INDEX dt_n ON dt (n DESC);",
				"CREATE INDEX dt_f8 ON dt (f8 DESC);",
				"CREATE INDEX dt_t ON dt (t DESC);",
				"CREATE INDEX dt_v ON dt (v DESC);",
				"CREATE INDEX dt_c ON dt (c DESC);",
				"CREATE INDEX dt_d ON dt (d DESC);",
				"CREATE INDEX dt_ts ON dt (ts DESC);",
				"CREATE INDEX dt_tz ON dt (tz DESC);",
				"CREATE INDEX dt_b ON dt (b DESC);",
				"CREATE INDEX dt_u ON dt (u DESC);",
				"INSERT INTO dt VALUES (1, 1, 100, 1.50, 1.5, 'b', 'bb', 'b', '2024-01-02', '2024-01-02 10:00:00', '2024-01-02 10:00:00+00', true, '00000000-0000-0000-0000-000000000002');",
				"INSERT INTO dt VALUES (2, -3, -100, -2.25, -2.5, 'a', 'aa', 'a', '2024-01-01', '2024-01-01 10:00:00', '2024-01-01 10:00:00+00', false, '00000000-0000-0000-0000-000000000001');",
				"INSERT INTO dt VALUES (3, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL);",
				"INSERT INTO dt VALUES (4, 7, 9223372036854775807, 1000.00, 0, 'c', 'cc', 'c', '2024-01-03', '2024-01-03 10:00:00', '2024-01-03 10:00:00+00', true, '00000000-0000-0000-0000-000000000003');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT indexname, indexdef FROM pg_indexes WHERE tablename = 'dt' ORDER BY indexname;",
					Expected: []sql.Row{
						{"dt_b", "CREATE INDEX dt_b ON public.dt USING btree (b DESC)"},
						{"dt_c", "CREATE INDEX dt_c ON public.dt USING btree (c DESC)"},
						{"dt_d", "CREATE INDEX dt_d ON public.dt USING btree (d DESC)"},
						{"dt_f8", "CREATE INDEX dt_f8 ON public.dt USING btree (f8 DESC)"},
						{"dt_i2", "CREATE INDEX dt_i2 ON public.dt USING btree (i2 DESC)"},
						{"dt_i8", "CREATE INDEX dt_i8 ON public.dt USING btree (i8 DESC)"},
						{"dt_n", "CREATE INDEX dt_n ON public.dt USING btree (n DESC)"},
						{"dt_pkey", "CREATE UNIQUE INDEX dt_pkey ON public.dt USING btree (pk)"},
						{"dt_t", "CREATE INDEX dt_t ON public.dt USING btree (t DESC)"},
						{"dt_ts", "CREATE INDEX dt_ts ON public.dt USING btree (ts DESC)"},
						{"dt_tz", "CREATE INDEX dt_tz ON public.dt USING btree (tz DESC)"},
						{"dt_u", "CREATE INDEX dt_u ON public.dt USING btree (u DESC)"},
						{"dt_v", "CREATE INDEX dt_v ON public.dt USING btree (v DESC)"},
					},
				},
				{
					Query: "EXPLAIN SELECT pk FROM dt ORDER BY i2 DESC;",
					Expected: []sql.Row{
						{"Project"},
						{" ├─ columns: [dt.pk]"},
						{" └─ IndexedTableAccess(dt)"},
						{"     ├─ index: [dt.i2 DESC]"},
						{"     ├─ filters: [{[NULL, ∞)}]"},
						{"     └─ columns: [pk i2]"},
					},
				},
				{
					Query:    "SELECT pk FROM dt ORDER BY i2 DESC;",
					Expected: []sql.Row{{3}, {4}, {1}, {2}},
				},
				{
					Query: "EXPLAIN SELECT pk FROM dt ORDER BY i2;",
					Expected: []sql.Row{
						{"Project"},
						{" ├─ columns: [dt.pk]"},
						{" └─ IndexedTableAccess(dt)"},
						{"     ├─ index: [dt.i2 DESC]"},
						{"     ├─ filters: [{[NULL, ∞)}]"},
						{"     ├─ columns: [pk i2]"},
						{"     └─ reverse: true"},
					},
				},
				{
					Query:    "SELECT pk FROM dt ORDER BY i2;",
					Expected: []sql.Row{{2}, {1}, {4}, {3}},
				},
				{
					Query:    "SELECT pk FROM dt WHERE i2 > 0 ORDER BY i2 DESC;",
					Expected: []sql.Row{{4}, {1}},
				},
				{
					Query:    "SELECT pk FROM dt WHERE i2 <= 1 ORDER BY i2;",
					Expected: []sql.Row{{2}, {1}},
				},
				{
					Query:    "SELECT pk FROM dt ORDER BY i8 DESC;",
					Expected: []sql.Row{{3}, {4}, {1}, {2}},
				},
				{
					Query:    "SELECT pk FROM dt WHERE i8 >= 100 ORDER BY i8 DESC;",
					Expected: []sql.Row{{4}, {1}},
				},
				{
					Query:    "SELECT pk FROM dt ORDER BY n DESC;",
					Expected: []sql.Row{{3}, {4}, {1}, {2}},
				},
				{
					Query:    "SELECT pk FROM dt WHERE n BETWEEN -5 AND 2 ORDER BY n DESC;",
					Expected: []sql.Row{{1}, {2}},
				},
				{
					Query:    "SELECT pk FROM dt ORDER BY f8 DESC;",
					Expected: []sql.Row{{3}, {1}, {4}, {2}},
				},
				{
					Query:    "SELECT pk FROM dt WHERE f8 < 1 ORDER BY f8;",
					Expected: []sql.Row{{2}, {4}},
				},
				{
					Query: "EXPLAIN SELECT pk FROM dt WHERE t >= 'b' ORDER BY t DESC;",
					Expected: []sql.Row{
						{"Project"},
						{" ├─ columns: [dt.pk]"},
						{" └─ IndexedTableAccess(dt)"},
						{"     ├─ index: [dt.t DESC]"},
						{"     ├─ filters: [{[b, ∞)}]"},
						{"     └─ columns: [pk t]"},
					},
				},
				{
					Query:    "SELECT pk FROM dt ORDER BY t DESC;",
					Expected: []sql.Row{{3}, {4}, {1}, {2}},
				},
				{
					Query:    "SELECT pk FROM dt WHERE t >= 'b' ORDER BY t DESC;",
					Expected: []sql.Row{{4}, {1}},
				},
				{
					Query:    "SELECT pk FROM dt WHERE t < 'c' ORDER BY t;",
					Expected: []sql.Row{{2}, {1}},
				},
				{
					Query:    "SELECT pk FROM dt ORDER BY v DESC;",
					Expected: []sql.Row{{3}, {4}, {1}, {2}},
				},
				{
					Query:    "SELECT pk FROM dt WHERE v = 'aa';",
					Expected: []sql.Row{{2}},
				},
				{
					Query:    "SELECT pk FROM dt ORDER BY c DESC;",
					Expected: []sql.Row{{3}, {4}, {1}, {2}},
				},
				{
					Query:    "SELECT pk FROM dt WHERE c > 'a' ORDER BY c DESC;",
					Expected: []sql.Row{{4}, {1}},
				},
				{
					Query:    "SELECT pk FROM dt ORDER BY d DESC;",
					Expected: []sql.Row{{3}, {4}, {1}, {2}},
				},
				{
					Query:    "SELECT pk FROM dt WHERE d >= '2024-01-02' ORDER BY d DESC;",
					Expected: []sql.Row{{4}, {1}},
				},
				{
					Query: "EXPLAIN SELECT pk FROM dt WHERE ts < '2024-01-03' ORDER BY ts DESC;",
					Expected: []sql.Row{
						{"Project"},
						{" ├─ columns: [dt.pk]"},
						{" └─ IndexedTableAccess(dt)"},
						{"     ├─ index: [dt.ts DESC]"},
						{"     ├─ filters: [{(NULL, 2024-01-03 00:00:00 +0000 UTC)}]"},
						{"     └─ columns: [pk ts]"},
					},
				},
				{
					Query:    "SELECT pk FROM dt ORDER BY ts DESC;",
					Expected: []sql.Row{{3}, {4}, {1}, {2}},
				},
				{
					Query:    "SELECT pk FROM dt WHERE ts < '2024-01-03' ORDER BY ts DESC;",
					Expected: []sql.Row{{1}, {2}},
				},
				{
					Query:    "SELECT pk FROM dt ORDER BY tz DESC;",
					Expected: []sql.Row{{3}, {4}, {1}, {2}},
				},
				{
					Query:    "SELECT pk FROM dt WHERE tz = '2024-01-01 10:00:00+00';",
					Expected: []sql.Row{{2}},
				},
				{
					Query:    "SELECT pk FROM dt WHERE b = true ORDER BY pk;",
					Expected: []sql.Row{{1}, {4}},
				},
				{
					Query:    "SELECT pk FROM dt WHERE b IS NULL;",
					Expected: []sql.Row{{3}},
				},
				{
					Query: "EXPLAIN SELECT pk FROM dt WHERE u > '00000000-0000-0000-0000-000000000001' ORDER BY u DESC;",
					Expected: []sql.Row{
						{"Project"},
						{" ├─ columns: [dt.pk]"},
						{" └─ IndexedTableAccess(dt)"},
						{"     ├─ index: [dt.u DESC]"},
						{"     ├─ filters: [{(00000000-0000-0000-0000-000000000001, ∞)}]"},
						{"     └─ columns: [pk u]"},
					},
				},
				{
					Query:    "SELECT pk FROM dt ORDER BY u DESC;",
					Expected: []sql.Row{{3}, {4}, {1}, {2}},
				},
				{
					Query:    "SELECT pk FROM dt WHERE u > '00000000-0000-0000-0000-000000000001' ORDER BY u DESC;",
					Expected: []sql.Row{{4}, {1}},
				},
				{
					Query: "EXPLAIN SELECT pk FROM dt ORDER BY t DESC NULLS LAST;",
					Expected: []sql.Row{
						{"Project"},
						{" ├─ columns: [dt.pk]"},
						{" └─ Sort(dt.t DESC)"},
						{"     └─ Table"},
						{"         ├─ name: dt"},
						{"         └─ columns: [pk t]"},
					},
				},
				{
					Query:    "SELECT pk FROM dt ORDER BY t DESC NULLS LAST;",
					Expected: []sql.Row{{4}, {1}, {2}, {3}},
				},
				{
					Query:    "SELECT pk FROM dt ORDER BY t ASC NULLS FIRST;",
					Expected: []sql.Row{{3}, {2}, {1}, {4}},
				},
			},
		},
		{
			Name: "mixed column orderings",
			SetUpScript: []string{
				"CREATE TABLE mo (pk INT PRIMARY KEY, a INT, b INT, c TEXT);",
				"CREATE INDEX mo_ad_b ON mo (a DESC, b);",
				"CREATE INDEX mo_a_bd ON mo (a, b DESC);",
				"CREATE INDEX mo_ad_bd ON mo (a DESC, b DESC);",
				"CREATE INDEX mo_anl_cnf ON mo (a DESC NULLS LAST, c ASC NULLS FIRST);",
				"INSERT INTO mo VALUES (1, 1, 1, 'x'), (2, 1, 2, 'y'), (3, 2, 1, NULL), (4, NULL, 3, 'z'), (5, 3, NULL, 'w'), (6, 2, 2, 'x'), (7, NULL, NULL, NULL);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT indexname, indexdef FROM pg_indexes WHERE tablename = 'mo' ORDER BY indexname;",
					Expected: []sql.Row{
						{"mo_a_bd", "CREATE INDEX mo_a_bd ON public.mo USING btree (a, b DESC)"},
						{"mo_ad_b", "CREATE INDEX mo_ad_b ON public.mo USING btree (a DESC, b)"},
						{"mo_ad_bd", "CREATE INDEX mo_ad_bd ON public.mo USING btree (a DESC, b DESC)"},
						{"mo_anl_cnf", "CREATE INDEX mo_anl_cnf ON public.mo USING btree (a DESC NULLS LAST, c NULLS FIRST)"},
						{"mo_pkey", "CREATE UNIQUE INDEX mo_pkey ON public.mo USING btree (pk)"},
					},
				},
				{
					Query: "SELECT indexrelid::regclass, array_to_string(indoption, ',') FROM pg_index WHERE indrelid = 'mo'::regclass ORDER BY indexrelid::regclass::text;",
					Expected: []sql.Row{
						{"mo_a_bd", "0,3"},
						{"mo_ad_b", "3,0"},
						{"mo_ad_bd", "3,3"},
						{"mo_anl_cnf", "1,2"},
						{"mo_pkey", "0"},
					},
				},
				{
					Query: "EXPLAIN SELECT pk FROM mo ORDER BY a DESC, b;",
					Expected: []sql.Row{
						{"Project"},
						{" ├─ columns: [mo.pk]"},
						{" └─ IndexedTableAccess(mo)"},
						{"     ├─ index: [mo.a,mo.b DESC]"},
						{"     ├─ filters: [{[NULL, ∞), [NULL, ∞)}]"},
						{"     ├─ columns: [pk a b]"},
						{"     └─ reverse: true"},
					},
				},
				{
					Query:    "SELECT pk FROM mo ORDER BY a DESC, b;",
					Expected: []sql.Row{{4}, {7}, {5}, {3}, {6}, {1}, {2}},
				},
				{
					Query: "EXPLAIN SELECT pk FROM mo ORDER BY a, b DESC;",
					Expected: []sql.Row{
						{"Project"},
						{" ├─ columns: [mo.pk]"},
						{" └─ IndexedTableAccess(mo)"},
						{"     ├─ index: [mo.a,mo.b DESC]"},
						{"     ├─ filters: [{[NULL, ∞), [NULL, ∞)}]"},
						{"     └─ columns: [pk a b]"},
					},
				},
				{
					Query:    "SELECT pk FROM mo ORDER BY a, b DESC;",
					Expected: []sql.Row{{2}, {1}, {6}, {3}, {5}, {7}, {4}},
				},
				{
					Query: "EXPLAIN SELECT pk FROM mo ORDER BY a DESC, b DESC;",
					Expected: []sql.Row{
						{"Project"},
						{" ├─ columns: [mo.pk]"},
						{" └─ IndexedTableAccess(mo)"},
						{"     ├─ index: [mo.a DESC,mo.b DESC]"},
						{"     ├─ filters: [{[NULL, ∞), [NULL, ∞)}]"},
						{"     └─ columns: [pk a b]"},
					},
				},
				{
					Query:    "SELECT pk FROM mo ORDER BY a DESC, b DESC;",
					Expected: []sql.Row{{7}, {4}, {5}, {6}, {3}, {2}, {1}},
				},
				{
					Query: "EXPLAIN SELECT pk FROM mo ORDER BY a, b;",
					Expected: []sql.Row{
						{"Project"},
						{" ├─ columns: [mo.pk]"},
						{" └─ IndexedTableAccess(mo)"},
						{"     ├─ index: [mo.a DESC,mo.b DESC]"},
						{"     ├─ filters: [{[NULL, ∞), [NULL, ∞)}]"},
						{"     ├─ columns: [pk a b]"},
						{"     └─ reverse: true"},
					},
				},
				{
					Query:    "SELECT pk FROM mo ORDER BY a, b;",
					Expected: []sql.Row{{1}, {2}, {3}, {6}, {5}, {4}, {7}},
				},
				{
					Query: "EXPLAIN SELECT pk FROM mo ORDER BY a DESC NULLS LAST, c NULLS FIRST;",
					Expected: []sql.Row{
						{"Project"},
						{" ├─ columns: [mo.pk]"},
						{" └─ IndexedTableAccess(mo)"},
						{"     ├─ index: [mo.a DESC,mo.c]"},
						{"     ├─ filters: [{[NULL, ∞), [NULL, ∞)}]"},
						{"     └─ columns: [pk a c]"},
					},
				},
				{
					Query:    "SELECT pk FROM mo ORDER BY a DESC NULLS LAST, c NULLS FIRST;",
					Expected: []sql.Row{{5}, {3}, {6}, {1}, {2}, {7}, {4}},
				},
				{
					Query: "EXPLAIN SELECT pk FROM mo ORDER BY a NULLS FIRST, c DESC NULLS LAST;",
					Expected: []sql.Row{
						{"Project"},
						{" ├─ columns: [mo.pk]"},
						{" └─ IndexedTableAccess(mo)"},
						{"     ├─ index: [mo.a DESC,mo.c]"},
						{"     ├─ filters: [{[NULL, ∞), [NULL, ∞)}]"},
						{"     ├─ columns: [pk a c]"},
						{"     └─ reverse: true"},
					},
				},
				{
					Query:    "SELECT pk FROM mo ORDER BY a NULLS FIRST, c DESC NULLS LAST;",
					Expected: []sql.Row{{4}, {7}, {2}, {1}, {6}, {3}, {5}},
				},
				{
					Query:    "SELECT pk FROM mo WHERE a = 2 ORDER BY b DESC;",
					Expected: []sql.Row{{6}, {3}},
				},
				{
					Query:    "SELECT pk FROM mo WHERE a = 2 ORDER BY b;",
					Expected: []sql.Row{{3}, {6}},
				},
				{
					Query: "EXPLAIN SELECT pk FROM mo WHERE a = 1 AND b > 1;",
					Expected: []sql.Row{
						{"Project"},
						{" ├─ columns: [mo.pk]"},
						{" └─ IndexedTableAccess(mo)"},
						{"     ├─ index: [mo.a,mo.b DESC]"},
						{"     ├─ filters: [{[1, 1], (1, ∞)}]"},
						{"     └─ columns: [pk a b]"},
					},
				},
				{
					Query:    "SELECT pk FROM mo WHERE a = 1 AND b > 1;",
					Expected: []sql.Row{{2}},
				},
				{
					Query:    "SELECT pk FROM mo WHERE a >= 2 ORDER BY a DESC, b DESC LIMIT 2;",
					Expected: []sql.Row{{5}, {6}},
				},
				{
					Query: "EXPLAIN SELECT pk FROM mo WHERE a < 3 ORDER BY a DESC, b;",
					Expected: []sql.Row{
						{"Project"},
						{" ├─ columns: [mo.pk]"},
						{" └─ IndexedTableAccess(mo)"},
						{"     ├─ index: [mo.a,mo.b DESC]"},
						{"     ├─ filters: [{(NULL, 3), [NULL, ∞)}]"},
						{"     ├─ columns: [pk a b]"},
						{"     └─ reverse: true"},
					},
				},
				{
					Query:    "SELECT pk FROM mo WHERE a < 3 ORDER BY a DESC, b;",
					Expected: []sql.Row{{3}, {6}, {1}, {2}},
				},
				{
					Query:    "SELECT pk FROM mo WHERE a IS NULL ORDER BY b;",
					Expected: []sql.Row{{4}, {7}},
				},
				{
					Query:    "SELECT pk FROM mo WHERE a IS NULL ORDER BY b DESC;",
					Expected: []sql.Row{{7}, {4}},
				},
				{
					Query:    "SELECT pk FROM mo WHERE a IS NOT NULL ORDER BY a DESC, b DESC;",
					Expected: []sql.Row{{5}, {6}, {3}, {2}, {1}},
				},
				{
					Query:    "SELECT pk FROM mo WHERE a IN (1, 3) ORDER BY a DESC, b;",
					Expected: []sql.Row{{5}, {1}, {2}},
				},
				{
					Query: "EXPLAIN SELECT pk FROM mo WHERE a = 1 AND b IS NULL;",
					Expected: []sql.Row{
						{"Project"},
						{" ├─ columns: [mo.pk]"},
						{" └─ IndexedTableAccess(mo)"},
						{"     ├─ index: [mo.a,mo.b DESC]"},
						{"     ├─ filters: [{[1, 1], [NULL, NULL]}]"},
						{"     └─ columns: [pk a b]"},
					},
				},
				{
					Query:    "SELECT pk FROM mo WHERE a = 1 AND b IS NULL;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT pk FROM mo WHERE a = 3 AND b IS NULL;",
					Expected: []sql.Row{{5}},
				},
				{
					Query:    "SELECT a, count(*) FROM mo GROUP BY a ORDER BY a DESC;",
					Expected: []sql.Row{{nil, 2}, {3, 1}, {2, 2}, {1, 2}},
				},
				{
					Query:    "SELECT max(a), min(a) FROM mo;",
					Expected: []sql.Row{{3, 1}},
				},
				{
					Query:    "SELECT max(b) FROM mo WHERE a = 1;",
					Expected: []sql.Row{{2}},
				},
				{
					Query: "UPDATE mo SET a = 5 WHERE pk = 1;",
				},
				{
					Query:    "SELECT pk FROM mo ORDER BY a DESC, b;",
					Expected: []sql.Row{{4}, {7}, {1}, {5}, {3}, {6}, {2}},
				},
				{
					Query: "DELETE FROM mo WHERE a = 2 AND b = 2;",
				},
				{
					Query:    "SELECT pk FROM mo WHERE a = 2 ORDER BY b DESC;",
					Expected: []sql.Row{{3}},
				},
				{
					Query:    "SELECT pk FROM mo ORDER BY a, b DESC;",
					Expected: []sql.Row{{2}, {3}, {5}, {1}, {7}, {4}},
				},
			},
		},
		{
			Name: "NaN in float indexes",
			SetUpScript: []string{
				"CREATE TABLE fl (pk INT PRIMARY KEY, f4 REAL, f8 DOUBLE PRECISION);",
				"CREATE INDEX fl_f4 ON fl (f4);",
				"CREATE INDEX fl_f8 ON fl (f8 DESC);",
				"INSERT INTO fl VALUES (1, 1.5, 1.5), (2, 'NaN', 'NaN'), (3, NULL, NULL), (4, -2.5, -2.5), (5, 'Infinity', 'Infinity'), (6, '-Infinity', '-Infinity'), (7, 0, 0);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "EXPLAIN SELECT pk FROM fl WHERE f4 = 'NaN';",
					Expected: []sql.Row{
						{"Project"},
						{" ├─ columns: [fl.pk]"},
						{" └─ IndexedTableAccess(fl)"},
						{"     ├─ index: [fl.f4]"},
						{"     ├─ filters: [{[NaN, NaN]}]"},
						{"     └─ columns: [pk f4]"},
					},
				},
				{
					Query:    "SELECT pk FROM fl WHERE f4 = 'NaN';",
					Expected: []sql.Row{{2}},
				},
				{
					Query: "EXPLAIN SELECT pk FROM fl WHERE f8 = 'NaN';",
					Expected: []sql.Row{
						{"Project"},
						{" ├─ columns: [fl.pk]"},
						{" └─ IndexedTableAccess(fl)"},
						{"     ├─ index: [fl.f8 DESC]"},
						{"     ├─ filters: [{[NaN, NaN]}]"},
						{"     └─ columns: [pk f8]"},
					},
				},
				{
					Query:    "SELECT pk FROM fl WHERE f8 = 'NaN';",
					Expected: []sql.Row{{2}},
				},
				{
					Query:    "SELECT pk FROM fl WHERE f4 > 1 ORDER BY f4;",
					Expected: []sql.Row{{1}, {5}, {2}},
				},
				{
					Query:    "SELECT pk FROM fl WHERE f8 > 1 ORDER BY f8 DESC;",
					Expected: []sql.Row{{2}, {5}, {1}},
				},
				{
					Query:    "SELECT pk FROM fl ORDER BY f4;",
					Expected: []sql.Row{{6}, {4}, {7}, {1}, {5}, {2}, {3}},
				},
				{
					Query:    "SELECT pk FROM fl ORDER BY f4 DESC;",
					Expected: []sql.Row{{3}, {2}, {5}, {1}, {7}, {4}, {6}},
				},
				{
					Query:    "SELECT pk FROM fl ORDER BY f8 DESC;",
					Expected: []sql.Row{{3}, {2}, {5}, {1}, {7}, {4}, {6}},
				},
				{
					Query: "EXPLAIN SELECT pk FROM fl ORDER BY f8;",
					Expected: []sql.Row{
						{"Project"},
						{" ├─ columns: [fl.pk]"},
						{" └─ IndexedTableAccess(fl)"},
						{"     ├─ index: [fl.f8 DESC]"},
						{"     ├─ filters: [{[NULL, ∞)}]"},
						{"     ├─ columns: [pk f8]"},
						{"     └─ reverse: true"},
					},
				},
				{
					Query:    "SELECT pk FROM fl ORDER BY f8;",
					Expected: []sql.Row{{6}, {4}, {7}, {1}, {5}, {2}, {3}},
				},
				{
					Query:    "SELECT pk FROM fl WHERE f8 < 0 ORDER BY f8 DESC;",
					Expected: []sql.Row{{4}, {6}},
				},
				{
					Query:    "SELECT pk FROM fl WHERE f8 >= 'Infinity' ORDER BY pk;",
					Expected: []sql.Row{{2}, {5}},
				},
				{
					Query:    "SELECT pk FROM fl WHERE f4 <= 'Infinity' ORDER BY f4;",
					Expected: []sql.Row{{6}, {4}, {7}, {1}, {5}},
				},
				{
					Query:    "SELECT pk FROM fl WHERE f4 = 'NaN' AND f8 = 'NaN';",
					Expected: []sql.Row{{2}},
				},
				{
					Query:    "SELECT pk, f4 = 'NaN', f8 <> 'NaN' FROM fl WHERE pk = 2;",
					Expected: []sql.Row{{2, "t", "f"}},
				},
			},
		},
		{
			Name: "unique descending indexes",
			SetUpScript: []string{
				"CREATE TABLE ud (pk INT PRIMARY KEY, a INT, b TEXT);",
				"CREATE UNIQUE INDEX ud_a ON ud (a DESC);",
				"CREATE UNIQUE INDEX ud_ba ON ud (b DESC NULLS LAST, a);",
				"INSERT INTO ud VALUES (1, 1, 'x'), (2, 2, 'y'), (3, NULL, 'z'), (4, NULL, NULL);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       "INSERT INTO ud VALUES (5, 1, 'q');",
					ExpectedErr: "duplicate unique key given",
				},
				{
					Query: "INSERT INTO ud VALUES (5, NULL, 'q');",
				},
				{
					Query: "INSERT INTO ud VALUES (6, 3, 'x');",
				},
				{
					Query: "INSERT INTO ud VALUES (7, 4, 'x');",
				},
				{
					Query:       "UPDATE ud SET a = 3 WHERE pk = 7;",
					ExpectedErr: "duplicate unique key given",
				},
				{
					Query:       "UPDATE ud SET a = 2 WHERE pk = 1;",
					ExpectedErr: "duplicate unique key given",
				},
				{
					Query: "UPDATE ud SET a = 9 WHERE pk = 1;",
				},
				{
					Query: "EXPLAIN SELECT pk FROM ud WHERE a IS NOT NULL ORDER BY a DESC;",
					Expected: []sql.Row{
						{"Project"},
						{" ├─ columns: [ud.pk]"},
						{" └─ IndexedTableAccess(ud)"},
						{"     ├─ index: [ud.a DESC]"},
						{"     ├─ filters: [{(NULL, ∞)}]"},
						{"     └─ columns: [pk a]"},
					},
				},
				{
					Query:    "SELECT pk FROM ud WHERE a IS NOT NULL ORDER BY a DESC;",
					Expected: []sql.Row{{1}, {7}, {6}, {2}},
				},
				{
					Query:    "SELECT pk FROM ud WHERE a IS NOT NULL ORDER BY a;",
					Expected: []sql.Row{{2}, {6}, {7}, {1}},
				},
				{
					Query:    "SELECT pk FROM ud WHERE b = 'x' ORDER BY a;",
					Expected: []sql.Row{{6}, {7}, {1}},
				},
				{
					Query:    "SELECT pk FROM ud WHERE b IS NULL;",
					Expected: []sql.Row{{4}},
				},
				{
					Query:    "SELECT pk FROM ud WHERE b > 'x' ORDER BY b DESC, a;",
					Expected: []sql.Row{{3}, {2}},
				},
				{
					Query:    "SELECT pk FROM ud WHERE a = 9;",
					Expected: []sql.Row{{1}},
				},
				{
					Query: "DELETE FROM ud WHERE a = 9;",
				},
				{
					Query:    "SELECT pk FROM ud WHERE a = 9;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT pk FROM ud WHERE a IS NOT NULL ORDER BY a DESC;",
					Expected: []sql.Row{{7}, {6}, {2}},
				},
				{
					Query: "INSERT INTO ud VALUES (8, 9, 'r');",
				},
				{
					Query:    "SELECT pk FROM ud WHERE a = 9;",
					Expected: []sql.Row{{8}},
				},
			},
		},
		{
			Name: "descending index on a table without a primary key",
			SetUpScript: []string{
				"CREATE TABLE kl (a INT, b INT);",
				"CREATE INDEX kl_ab ON kl (a DESC, b DESC);",
				"INSERT INTO kl VALUES (1, 1), (1, 2), (2, 1), (NULL, 5), (1, 1), (2, NULL);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "EXPLAIN SELECT a, b FROM kl ORDER BY a DESC, b DESC;",
					Expected: []sql.Row{
						{"IndexedTableAccess(kl)"},
						{" ├─ index: [kl.a DESC,kl.b DESC]"},
						{" ├─ filters: [{[NULL, ∞), [NULL, ∞)}]"},
						{" └─ columns: [a b]"},
					},
				},
				{
					Query:    "SELECT a, b FROM kl ORDER BY a DESC, b DESC;",
					Expected: []sql.Row{{nil, 5}, {2, nil}, {2, 1}, {1, 2}, {1, 1}, {1, 1}},
				},
				{
					Query: "EXPLAIN SELECT a, b FROM kl ORDER BY a, b;",
					Expected: []sql.Row{
						{"IndexedTableAccess(kl)"},
						{" ├─ index: [kl.a DESC,kl.b DESC]"},
						{" ├─ filters: [{[NULL, ∞), [NULL, ∞)}]"},
						{" ├─ columns: [a b]"},
						{" └─ reverse: true"},
					},
				},
				{
					Query:    "SELECT a, b FROM kl ORDER BY a, b;",
					Expected: []sql.Row{{1, 1}, {1, 1}, {1, 2}, {2, 1}, {2, nil}, {nil, 5}},
				},
				{
					Query:    "SELECT a, b FROM kl WHERE a = 1 ORDER BY b DESC;",
					Expected: []sql.Row{{1, 2}, {1, 1}, {1, 1}},
				},
				{
					Query:    "SELECT a, b FROM kl WHERE a IS NULL;",
					Expected: []sql.Row{{nil, 5}},
				},
				{
					Query:    "SELECT a, b FROM kl WHERE a = 2 AND b IS NULL;",
					Expected: []sql.Row{{2, nil}},
				},
				{
					Query: "DELETE FROM kl WHERE a = 1 AND b = 1;",
				},
				{
					Query:    "SELECT a, b FROM kl ORDER BY a DESC, b DESC;",
					Expected: []sql.Row{{nil, 5}, {2, nil}, {2, 1}, {1, 2}},
				},
			},
		},
		{
			Name: "descending expression and operator class indexes",
			SetUpScript: []string{
				"CREATE TABLE ex (id BIGINT PRIMARY KEY, name VARCHAR(100), created_at TIMESTAMPTZ);",
				"CREATE INDEX ex_lower_desc ON ex ((lower(name)) DESC);",
				"INSERT INTO ex VALUES (1, 'Bob', '2024-01-01 00:00:00+00'), (2, 'alice', '2024-01-03 00:00:00+00'), (3, 'Carol', NULL), (4, NULL, '2024-01-02 00:00:00+00');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT indexdef FROM pg_indexes WHERE indexname = 'ex_lower_desc';",
					Expected: []sql.Row{{"CREATE INDEX ex_lower_desc ON public.ex USING btree ((lower(name)) DESC)"}},
				},
				{
					Query:    "SELECT id FROM ex ORDER BY lower(name) DESC;",
					Expected: []sql.Row{{4}, {3}, {1}, {2}},
				},
				{
					Query:    "SELECT id FROM ex ORDER BY lower(name);",
					Expected: []sql.Row{{2}, {1}, {3}, {4}},
				},
				{
					Query:    "SELECT id FROM ex WHERE lower(name) = 'bob';",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT id FROM ex WHERE lower(name) > 'b' ORDER BY lower(name) DESC;",
					Expected: []sql.Row{{3}, {1}},
				},
				{
					Query: "DROP INDEX ex_lower_desc;",
				},
				{
					Query: "CREATE INDEX ex_pattern ON ex (name varchar_pattern_ops DESC);",
				},
				{
					Query:    "SELECT indexdef FROM pg_indexes WHERE indexname = 'ex_pattern';",
					Expected: []sql.Row{{"CREATE INDEX ex_pattern ON public.ex USING btree (name varchar_pattern_ops DESC)"}},
				},
				{
					Query:    "SELECT array_to_string(indoption, ',') || ' ' || array_to_string(indclass, ',') FROM pg_index WHERE indexrelid = 'ex_pattern'::regclass;",
					Expected: []sql.Row{{"3 15027"}},
				},
				{
					Query: "EXPLAIN SELECT id FROM ex WHERE name LIKE 'C%';",
					Expected: []sql.Row{
						{"Project"},
						{" ├─ columns: [ex.id]"},
						{" └─ Filter"},
						{"     ├─ ex.name LIKE 'C%'"},
						{"     └─ IndexedTableAccess(ex)"},
						{"         ├─ index: [ex.name DESC]"},
						{"         ├─ filters: [{[C, D)}]"},
						{"         └─ columns: [id name]"},
					},
				},
				{
					Query:    "SELECT id FROM ex WHERE name LIKE 'C%';",
					Expected: []sql.Row{{3}},
				},
				{
					Query:    "SELECT id FROM ex WHERE name LIKE 'a%';",
					Expected: []sql.Row{{2}},
				},
			},
		},
		{
			Name: "LIKE prefix uses an index",
			SetUpScript: []string{
				"CREATE TABLE like_idx (pk INT PRIMARY KEY, t TEXT, v VARCHAR(20));",
				"CREATE INDEX like_idx_t ON like_idx (t text_pattern_ops);",
				"CREATE INDEX like_idx_v ON like_idx (v varchar_pattern_ops);",
				`INSERT INTO like_idx VALUES (1, 'abc', 'abc'), (2, 'abd', 'abd'), (3, 'ab', 'ab'), (4, 'b', 'b'), (5, 'a_c', 'a_c'), (6, NULL, NULL), (7, 'ABC', 'ABC'), (8, 'ac', 'ac'), (9, 'a%c', 'a%c'), (10, 'a\c', 'a\c');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "EXPLAIN SELECT pk FROM like_idx WHERE t LIKE 'ab%';",
					Expected: []sql.Row{
						{"Project"},
						{" ├─ columns: [like_idx.pk]"},
						{" └─ Filter"},
						{"     ├─ like_idx.t LIKE 'ab%'"},
						{"     └─ IndexedTableAccess(like_idx)"},
						{"         ├─ index: [like_idx.t]"},
						{"         ├─ filters: [{[ab, ac)}]"},
						{"         └─ columns: [pk t]"},
					},
				},
				{
					Query: "EXPLAIN SELECT pk FROM like_idx WHERE v LIKE 'ab%';",
					Expected: []sql.Row{
						{"Project"},
						{" ├─ columns: [like_idx.pk]"},
						{" └─ Filter"},
						{"     ├─ like_idx.v LIKE 'ab%'"},
						{"     └─ IndexedTableAccess(like_idx)"},
						{"         ├─ index: [like_idx.v]"},
						{"         ├─ filters: [{[ab, ac)}]"},
						{"         └─ columns: [pk v]"},
					},
				},
				{
					Query: "EXPLAIN SELECT pk FROM like_idx WHERE t LIKE 'ab%' OR t LIKE 'b%';",
					Expected: []sql.Row{
						{"Project"},
						{" ├─ columns: [like_idx.pk]"},
						{" └─ Filter"},
						{"     ├─ (like_idx.t LIKE 'ab%' OR like_idx.t LIKE 'b%')"},
						{"     └─ Table"},
						{"         ├─ name: like_idx"},
						{"         └─ columns: [pk t]"},
					},
				},
				{
					Query: "EXPLAIN SELECT pk FROM like_idx WHERE t LIKE '%bc';",
					Expected: []sql.Row{
						{"Project"},
						{" ├─ columns: [like_idx.pk]"},
						{" └─ Filter"},
						{"     ├─ like_idx.t LIKE '%bc'"},
						{"     └─ Table"},
						{"         ├─ name: like_idx"},
						{"         └─ columns: [pk t]"},
					},
				},
				{
					Query: "EXPLAIN SELECT pk FROM like_idx WHERE t NOT LIKE 'ab%';",
					Expected: []sql.Row{
						{"Project"},
						{" ├─ columns: [like_idx.pk]"},
						{" └─ Filter"},
						{"     ├─ (NOT(like_idx.t LIKE 'ab%'))"},
						{"     └─ Table"},
						{"         ├─ name: like_idx"},
						{"         └─ columns: [pk t]"},
					},
				},
				{
					Query:    "SELECT pk FROM like_idx WHERE t LIKE 'ab%' ORDER BY pk;",
					Expected: []sql.Row{{1}, {2}, {3}},
				},
				{
					Query:    "SELECT pk FROM like_idx WHERE v LIKE 'ab%' ORDER BY pk;",
					Expected: []sql.Row{{1}, {2}, {3}},
				},
				{
					Query:    "SELECT pk FROM like_idx WHERE t LIKE 'a_c%' ORDER BY pk;",
					Expected: []sql.Row{{1}, {5}, {9}, {10}},
				},
				{
					Query:    `SELECT pk FROM like_idx WHERE t LIKE 'a\_c' ORDER BY pk;`,
					Expected: []sql.Row{{5}},
				},
				{
					Query:    `SELECT pk FROM like_idx WHERE t LIKE 'a\%%' ORDER BY pk;`,
					Expected: []sql.Row{{9}},
				},
				{
					Query:    "SELECT pk FROM like_idx WHERE t LIKE '%bc' ORDER BY pk;",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT pk FROM like_idx WHERE t LIKE 'ab' ORDER BY pk;",
					Expected: []sql.Row{{3}},
				},
				{
					Query:    "SELECT pk FROM like_idx WHERE t NOT LIKE 'ab%' ORDER BY pk;",
					Expected: []sql.Row{{4}, {5}, {7}, {8}, {9}, {10}},
				},
				{
					Query:    "SELECT pk FROM like_idx WHERE NOT (t LIKE 'ab%') ORDER BY pk;",
					Expected: []sql.Row{{4}, {5}, {7}, {8}, {9}, {10}},
				},
				{
					Query:    "SELECT pk FROM like_idx WHERE t LIKE 'A%' ORDER BY pk;",
					Expected: []sql.Row{{7}},
				},
				{
					Query:    "SELECT pk FROM like_idx WHERE t LIKE 'ab%' OR t LIKE 'b%' ORDER BY pk;",
					Expected: []sql.Row{{1}, {2}, {3}, {4}},
				},
				{
					Query:    "SELECT pk FROM like_idx WHERE t LIKE 'ab%' AND pk > 1 ORDER BY pk;",
					Expected: []sql.Row{{2}, {3}},
				},
				{
					Query:    "SELECT pk FROM like_idx WHERE t LIKE 'a%' AND t LIKE '%d' ORDER BY pk;",
					Expected: []sql.Row{{2}},
				},
				{
					Query:    "SELECT count(*) FROM like_idx WHERE t LIKE 'ab%';",
					Expected: []sql.Row{{3}},
				},
				{
					Query: "UPDATE like_idx SET v = 'zz' WHERE t LIKE 'ab%';",
				},
				{
					Query:    "SELECT pk, v FROM like_idx WHERE v LIKE 'z%' ORDER BY pk;",
					Expected: []sql.Row{{1, "zz"}, {2, "zz"}, {3, "zz"}},
				},
				{
					Query: "DELETE FROM like_idx WHERE v LIKE 'zz%';",
				},
				{
					Query:    "SELECT pk FROM like_idx ORDER BY pk;",
					Expected: []sql.Row{{4}, {5}, {6}, {7}, {8}, {9}, {10}},
				},
			},
		},
		{
			Name: "NULLS FIRST and NULLS LAST",
			SetUpScript: []string{
				"CREATE TABLE nulls_idx (pk INT PRIMARY KEY, v INT);",
				"CREATE INDEX nulls_idx_first ON nulls_idx (v NULLS FIRST);",
				"CREATE INDEX nulls_idx_desc_last ON nulls_idx (v DESC NULLS LAST);",
				"CREATE UNIQUE INDEX nulls_idx_unique_desc ON nulls_idx (v DESC, pk);",
				"INSERT INTO nulls_idx VALUES (1, 3), (2, NULL), (3, 1), (4, NULL), (5, 2);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT indexname, indexdef FROM pg_indexes WHERE tablename = 'nulls_idx' ORDER BY indexname;",
					Expected: []sql.Row{
						{"nulls_idx_desc_last", "CREATE INDEX nulls_idx_desc_last ON public.nulls_idx USING btree (v DESC NULLS LAST)"},
						{"nulls_idx_first", "CREATE INDEX nulls_idx_first ON public.nulls_idx USING btree (v NULLS FIRST)"},
						{"nulls_idx_pkey", "CREATE UNIQUE INDEX nulls_idx_pkey ON public.nulls_idx USING btree (pk)"},
						{"nulls_idx_unique_desc", "CREATE UNIQUE INDEX nulls_idx_unique_desc ON public.nulls_idx USING btree (v DESC, pk)"},
					},
				},
				{
					Query: "SELECT indexrelid::regclass, array_to_string(indoption, ',') FROM pg_index WHERE indrelid = 'nulls_idx'::regclass ORDER BY indexrelid::regclass::text;",
					Expected: []sql.Row{
						{"nulls_idx_desc_last", "1"},
						{"nulls_idx_first", "2"},
						{"nulls_idx_pkey", "0"},
						{"nulls_idx_unique_desc", "3,0"},
					},
				},
				{
					Query:    "SELECT pk FROM nulls_idx WHERE v IS NULL ORDER BY pk;",
					Expected: []sql.Row{{2}, {4}},
				},
				{
					Query:    "SELECT pk FROM nulls_idx WHERE v > 1 ORDER BY pk;",
					Expected: []sql.Row{{1}, {5}},
				},
				{
					Query:    "SELECT pk FROM nulls_idx WHERE v <= 2 ORDER BY pk;",
					Expected: []sql.Row{{3}, {5}},
				},
				{
					Query:    "SELECT pk FROM nulls_idx WHERE v IS NOT NULL ORDER BY pk;",
					Expected: []sql.Row{{1}, {3}, {5}},
				},
				{
					Query:    "SELECT pk FROM nulls_idx WHERE v = 2 OR v IS NULL ORDER BY pk;",
					Expected: []sql.Row{{2}, {4}, {5}},
				},
				{
					Query:    "SELECT pk FROM nulls_idx WHERE v BETWEEN 1 AND 2 ORDER BY pk;",
					Expected: []sql.Row{{3}, {5}},
				},
				{
					Query:    "SELECT pk, v FROM nulls_idx ORDER BY v ASC NULLS FIRST, pk;",
					Expected: []sql.Row{{2, nil}, {4, nil}, {3, 1}, {5, 2}, {1, 3}},
				},
				{
					Query:    "SELECT pk, v FROM nulls_idx ORDER BY v DESC NULLS LAST, pk;",
					Expected: []sql.Row{{1, 3}, {5, 2}, {3, 1}, {2, nil}, {4, nil}},
				},
				{
					Query:    "SELECT pk, v FROM nulls_idx ORDER BY v, pk;",
					Expected: []sql.Row{{3, 1}, {5, 2}, {1, 3}, {2, nil}, {4, nil}},
				},
				{
					Query:    "SELECT pk, v FROM nulls_idx ORDER BY v DESC, pk;",
					Expected: []sql.Row{{2, nil}, {4, nil}, {1, 3}, {5, 2}, {3, 1}},
				},
			},
		},
		{
			Name: "descending index satisfies ORDER BY",
			SetUpScript: []string{
				"CREATE TABLE sorted_idx (pk INT PRIMARY KEY, a INT NOT NULL, b INT NOT NULL);",
				"CREATE INDEX sorted_idx_ab ON sorted_idx (a, b DESC);",
				"INSERT INTO sorted_idx VALUES (1, 1, 1), (2, 1, 3), (3, 2, 2), (4, 1, 2);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT pk FROM sorted_idx WHERE a = 1 ORDER BY b DESC;",
					Expected: []sql.Row{{2}, {4}, {1}},
				},
				{
					Query:    "SELECT pk FROM sorted_idx ORDER BY a ASC, b DESC;",
					Expected: []sql.Row{{2}, {4}, {1}, {3}},
				},
				{
					Query:    "SELECT pk FROM sorted_idx ORDER BY a DESC, b ASC;",
					Expected: []sql.Row{{3}, {1}, {4}, {2}},
				},
			},
		},
		{
			Name: "unique constraints and plain indexes use the Postgres default order",
			SetUpScript: []string{
				"CREATE TABLE uq (pk INT PRIMARY KEY, a INT UNIQUE, b INT NOT NULL, c INT, d INT NOT NULL, CONSTRAINT uq_b UNIQUE (b), CONSTRAINT uq_c UNIQUE (c));",
				"CREATE INDEX uq_c_first ON uq (c NULLS FIRST);",
				"CREATE INDEX uq_c_plain ON uq (c);",
				"CREATE INDEX uq_d_first ON uq (d NULLS FIRST);",
				"ALTER TABLE uq ADD COLUMN e INT UNIQUE;",
				"INSERT INTO uq VALUES (1, 1, 1, NULL, 1, 1), (2, NULL, 2, 2, 2, NULL), (3, 3, 3, 1, 3, 3);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       "INSERT INTO uq VALUES (4, 4, 4, 4, 4, 3);",
					ExpectedErr: "duplicate unique key",
				},
				{
					Query: "SELECT indexname, indexdef FROM pg_indexes WHERE tablename = 'uq' ORDER BY indexname;",
					Expected: []sql.Row{
						{"uq_a_key", "CREATE UNIQUE INDEX uq_a_key ON public.uq USING btree (a)"},
						{"uq_b", "CREATE UNIQUE INDEX uq_b ON public.uq USING btree (b)"},
						{"uq_c", "CREATE UNIQUE INDEX uq_c ON public.uq USING btree (c)"},
						{"uq_c_first", "CREATE INDEX uq_c_first ON public.uq USING btree (c NULLS FIRST)"},
						{"uq_c_plain", "CREATE INDEX uq_c_plain ON public.uq USING btree (c)"},
						{"uq_d_first", "CREATE INDEX uq_d_first ON public.uq USING btree (d NULLS FIRST)"},
						{"uq_e_key", "CREATE UNIQUE INDEX uq_e_key ON public.uq USING btree (e)"},
						{"uq_pkey", "CREATE UNIQUE INDEX uq_pkey ON public.uq USING btree (pk)"},
					},
				},
				{
					Query: "SELECT indexrelid::regclass, array_to_string(indoption, ',') FROM pg_index WHERE indrelid = 'uq'::regclass ORDER BY indexrelid::regclass::text;",
					Expected: []sql.Row{
						{"uq_a_key", "0"},
						{"uq_b", "0"},
						{"uq_c", "0"},
						{"uq_c_first", "2"},
						{"uq_c_plain", "0"},
						{"uq_d_first", "2"},
						{"uq_e_key", "0"},
						{"uq_pkey", "0"},
					},
				},
				{
					Query:    "SELECT pk FROM uq ORDER BY c;",
					Expected: []sql.Row{{3}, {2}, {1}},
				},
				{
					Query:    "SELECT pk FROM uq ORDER BY c NULLS FIRST;",
					Expected: []sql.Row{{1}, {3}, {2}},
				},
				{
					Query:    "SELECT pk FROM uq ORDER BY a;",
					Expected: []sql.Row{{1}, {3}, {2}},
				},
				{
					Query:    "SELECT pk FROM uq ORDER BY e;",
					Expected: []sql.Row{{1}, {3}, {2}},
				},
				{
					Query:    "SELECT pk FROM uq WHERE c IS NULL;",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT pk FROM uq WHERE a IS NULL;",
					Expected: []sql.Row{{2}},
				},
				{
					Query:    "SELECT pk FROM uq WHERE a = 3;",
					Expected: []sql.Row{{3}},
				},
				{
					Query:    "SELECT pk FROM uq WHERE e = 3;",
					Expected: []sql.Row{{3}},
				},
			},
		},
	})
}
