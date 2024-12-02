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

func TestSetOperations(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Test intersect",
			SetUpScript: []string{
				"create table b (m int, n int);",
				"insert into b values (1,2), (1,3), (3,4);",
				"create table c (m int, n int);",
				"insert into c values (1,3), (1,3), (3,4);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "select * from b intersect select * from c order by 1,2;",
					Expected: []sql.Row{
						{1, 3},
						{3, 4},
					},
				},
				{
					Query: "(table b order by m limit 1 offset 1) intersect (table c order by m limit 1);",
					Expected: []sql.Row{
						{1, 3},
					},
				},
			},
		},
		{
			Name: "Test union",
			SetUpScript: []string{
				"create table b (m int, n int);",
				"insert into b values (1,2), (1,3), (3,4);",
				"create table c (m int, n int);",
				"insert into c values (1,3), (1,3), (3,4);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "select * from b union select * from c order by 1,2;",
					Expected: []sql.Row{
						{1, 2},
						{1, 3},
						{3, 4},
					},
				},
				{
					Query: "select * from b union all select * from c order by 1,2;",
					Expected: []sql.Row{
						{1, 2},
						{1, 3},
						{1, 3},
						{1, 3},
						{3, 4},
						{3, 4},
					},
				},
				{
					Query: "(table b order by m limit 1 offset 1) union (table c order by m limit 1);",
					Expected: []sql.Row{
						{1, 3},
					},
				},
				{
					Query: "(table b order by m limit 1 offset 1) union all (table c order by m limit 1);",
					Expected: []sql.Row{
						{1, 3},
						{1, 3},
					},
				},
			},
		},
		{
			Name: "Test except",
			SetUpScript: []string{
				"create table b (m int, n int);",
				"insert into b values (1,2), (1,3), (3,4);",
				"create table c (m int, n int);",
				"insert into c values (1,3), (1,3), (3,4);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "select * from b except select * from c order by 1,2;",
					Expected: []sql.Row{
						{1, 2},
					},
				},
				{
					Query: "(table b order by m limit 1 offset 1) except (table c order by m limit 1);",
					Expected: []sql.Row{},
				},
			},
		},
	})
}
