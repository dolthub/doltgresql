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

func TestUpdate(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "simple update",
			SetUpScript: []string{
				"CREATE TABLE t1 (a INT PRIMARY KEY, b INT)",
				"INSERT INTO t1 VALUES (1, 2), (2, 3), (3, 4)",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "UPDATE t1 SET b = 5 WHERE a = 2",
				},
				{
					Query: "SELECT * FROM t1 where a =  2",
					Expected: []sql.Row{
						{2, 5},
					},
				},
			},
		},
		{
			Name: "update to default",
			SetUpScript: []string{
				"create table t (i int default 10, j varchar(128) default (concat('abc', 'def')));",
				"insert into t values (100, 'a'), (200, 'b');",
				"create table t2 (i int);",
				"insert into t2 values (1), (2), (3);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "update t set i = default where i = 100;",
				},
				{
					Query: "select * from t order by i",
					Expected: []sql.Row{
						{10, "a"},
						{200, "b"},
					},
				},
				{
					Query: "update t set j = default where i = 200;",
				},
				{
					Query: "select * from t order by i",
					Expected: []sql.Row{
						{10, "a"},
						{200, "abcdef"},
					},
				},
				{
					Query: "update t set i = default, j = default;",
				},
				{
					Query: "select * from t order by i",
					Expected: []sql.Row{
						{10, "abcdef"},
						{10, "abcdef"},
					},
				},
				{
					Query: "update t2 set i = default",
					Skip:  true, // UPDATE: non-Doltgres type found in source
				},
				{
					Query: "select * from t2",
					Skip:  true, // skipped because of above
					Expected: []sql.Row{
						{nil},
						{nil},
						{nil},
					},
				},
			},
		},
	})
}
