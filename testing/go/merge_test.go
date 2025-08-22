// Copyright 2025 Dolthub, Inc.
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

func TestMerge(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "simple merge",
			SetUpScript: []string{
				"CREATE TABLE t1 (a INT, b INT, c INT, PRIMARY KEY (a))",
				"CREATE TABLE t2 (a INT, b INT, c INT, PRIMARY KEY (a))",
				"INSERT INTO t1 VALUES (1, 2, 3), (4, 5, 6), (7, 8, 9)",
				"INSERT INTO t2 VALUES (1, 2, 3), (4, 5, 6), (7, 8, 9)",
				"SELECT DOLT_COMMIT('-Am', 'intial commit')",
				"SELECT DOLT_BRANCH('branch1')",
				"SELECT DOLT_BRANCH('branch2')",
				"SELECT DOLT_CHECKOUT('branch1')",
				"INSERT INTO t1 VALUES (10, 11, 12)",
				"INSERT INTO t2 VALUES (10, 11, 12)",
				"SELECT DOLT_COMMIT('-Am', 'added 10')",
				"SELECT DOLT_CHECKOUT('branch2')",
				"INSERT INTO t1 VALUES (20, 21, 22)",
				"INSERT INTO t2 VALUES (20, 21, 22)",
				"SELECT DOLT_COMMIT('-Am', 'added 20')",
				"SELECT DOLT_CHECKOUT('main')",
				"SELECT DOLT_MERGE('branch1')",
				"SELECT DOLT_MERGE('branch2')",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM t1",
					Expected: []sql.Row{
						{1, 2, 3},
						{4, 5, 6},
						{7, 8, 9},
						{10, 11, 12},
						{20, 21, 22},
					},
				},
				{
					Query: "SELECT * FROM t2",
					Expected: []sql.Row{
						{1, 2, 3},
						{4, 5, 6},
						{7, 8, 9},
						{10, 11, 12},
						{20, 21, 22},
					},
				},
			},
		},
		{
			Name: "merge with check expressions and column defaults",
			SetUpScript: []string{
				"CREATE TABLE t1 (a INT, b timestamptz default '2020-01-01 00:00:00'::timestamptz, PRIMARY KEY (a))",
				"ALTER TABLE t1 ADD CONSTRAINT check_b CHECK (b >= '2020-01-01 00:00:00'::timestamptz)",
				"INSERT INTO t1 VALUES (1, '2020-01-02 00:00:00'), (2, '2020-01-03 00:00:00')",
				"SELECT DOLT_COMMIT('-Am', 'intial commit')",
				"SELECT DOLT_BRANCH('branch1')",
				"SELECT DOLT_BRANCH('branch2')",
				"SELECT DOLT_CHECKOUT('branch1')",
				"INSERT INTO t1 VALUES (3, '2020-01-04 00:00:00')",
				"SELECT DOLT_COMMIT('-Am', 'added 3')",
				"SELECT DOLT_CHECKOUT('branch2')",
				"INSERT INTO t1 VALUES (4, '2020-01-05 00:00:00')",
				"SELECT DOLT_COMMIT('-Am', 'added 4')",
				"SELECT DOLT_CHECKOUT('main')",
				"SELECT DOLT_MERGE('branch1')",
				"SELECT DOLT_MERGE('branch2')",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM t1 order by a",
					Expected: []sql.Row{
						{1, "2020-01-02 00:00:00-08"},
						{2, "2020-01-03 00:00:00-08"},
						{3, "2020-01-04 00:00:00-08"},
						{4, "2020-01-05 00:00:00-08"},
					},
				},
				{
					// make sure the check constraint is still there
					Query:       "INSERT INTO t1 VALUES (5, '2019-12-31 00:00:00')",
					ExpectedErr: "Check constraint",
				},
			},
		},
		{
			Name: "merge with unique constraints and foreign keys",
			SetUpScript: []string{
				"CREATE TABLE t1 (a INT, b INT, PRIMARY KEY (a), unique (b))",
				"CREATE TABLE t2 (a INT, b INT, PRIMARY KEY (a), foreign key (b) references t1(b))",
				"INSERT INTO t1 VALUES (1, 2), (4, 5), (7, 8)",
				"INSERT INTO t2 VALUES (1, 2), (4, 5), (7, 8)",
				"SELECT DOLT_COMMIT('-Am', 'intial commit')",
				"SELECT DOLT_BRANCH('branch1')",
				"SELECT DOLT_BRANCH('branch2')",
				"SELECT DOLT_CHECKOUT('branch1')",
				"INSERT INTO t1 VALUES (10, 11)",
				"INSERT INTO t2 VALUES (10, 11)",
				"SELECT DOLT_COMMIT('-Am', 'added 10')",
				"SELECT DOLT_CHECKOUT('branch2')",
				"INSERT INTO t1 VALUES (20, 21)",
				"INSERT INTO t2 VALUES (20, 21)",
				"SELECT DOLT_COMMIT('-Am', 'added 20')",
				"SELECT DOLT_CHECKOUT('main')",
				"SELECT DOLT_MERGE('branch1')",
				"SELECT DOLT_MERGE('branch2')",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM t1 order by a",
					Expected: []sql.Row{
						{1, 2},
						{4, 5},
						{7, 8},
						{10, 11},
						{20, 21},
					},
				},
				{
					// make sure the unique constraint is still there
					Query:       "INSERT INTO t1 VALUES (100, 2)",
					ExpectedErr: "unique key",
				},
				{
					// make sure the foreign key constraint is still there
					Query:       "INSERT INTO t2 VALUES (100, 200)",
					ExpectedErr: "Foreign key violation",
				},
			},
		},
	})
}
