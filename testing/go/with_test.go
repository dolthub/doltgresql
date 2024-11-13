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

func TestWithStatements(t *testing.T) {
	RunScripts(t, WithStatementTests)
}

var WithStatementTests = []ScriptTest{
	{
		Name: "basic values statements",
		SetUpScript: []string{
			"create table t (i int primary key);",
			"insert into t values (1), (2), (3);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "with cte as (select 1) select * from cte;",
				Expected: []sql.Row{
					{1},
				},
			},
			{
				Query: "with cte as (select 1, 2, 3 union select 4, 5, 6) select * from cte;",
				Expected: []sql.Row{
					{1, 2, 3},
					{4, 5, 6},
				},
			},
			{
				Query: "with cte as (values (1)) select * from cte;",
				Expected: []sql.Row{
					{1},
				},
			},
			{
				Query: "with cte as (values (1, 2, 3) union values (4, 5, 6)) select * from cte;",
				Expected: []sql.Row{
					{1, 2, 3},
					{4, 5, 6},
				},
			},
			{
				Query: "with cte as (select 1, 2, 3 union values (4, 5, 6)) select * from cte;",
				Expected: []sql.Row{
					{1, 2, 3},
					{4, 5, 6},
				},
			},
			{
				Query: "with recursive cte(x) as (select 1 union all select x + 1 from cte) select * from cte limit 5;",
				Expected: []sql.Row{
					{1},
					{2},
					{3},
					{4},
					{5},
				},
			},
		},
	},
}
