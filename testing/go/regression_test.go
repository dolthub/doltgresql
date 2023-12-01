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
		{
			Name:        "nullif",
			SetUpScript: []string{},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "select nullif(1, 1);",
					Expected: []sql.Row{{nil}},
				},
				{
					Query:    "select nullif('', null);",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "select nullif(10, 'a');",
					Expected: []sql.Row{{10}},
				},
			},
		},
		{
			Name:        "coalesce",
			SetUpScript: []string{},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "select coalesce(null + 5, 100);",
					Expected: []sql.Row{{100.0}}, // TODO: this should be an integer
				},
				{
					Query:    "select coalesce(null, null, 'abc');",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "select coalesce(null, null);",
					Expected: []sql.Row{{nil}},
				},
			},
		},
		{
			Name:        "case / when",
			SetUpScript: []string{},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT\n" +
						"  CASE\n" +
						"    WHEN 1 = 1 THEN 'One is equal to One'\n" +
						"    ELSE 'One is not equal to One'\n" +
						"  END AS result;",
					Expected: []sql.Row{{"One is equal to One"}},
				},
				{
					Query:    "SELECT\n" +
						"  CASE\n" +
						"    WHEN NULL IS NULL THEN 'NULL is NULL'\n" +
						"    ELSE 'NULL is not NULL'\n" +
						"  END AS result;",
					Expected: []sql.Row{{"NULL is NULL"}},
				},
			},
		},
		{
			Name:        "ALL / DISTINCT in functions",
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
			Name:        "cross joins",
			SetUpScript: []string{
				"create table t1 (pk int);",
				"create table t2 (pk int);",
				"insert into t1 values (1), (2);",
				"insert into t2 values (3), (4);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "select count(*) from t1 cross join t2;",
					Expected: []sql.Row{{4}},
				},
				{
					Query:    "select count(*) from t1, t2;",
					Expected: []sql.Row{{4}},
				},
			},
		},
	})
}

