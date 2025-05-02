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

// TestSelect covers SELECT syntax not covered by our MySQL select tests
func TestSelect(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "select values",
			Assertions: []ScriptTestAssertion{
				{
					Query: "select * from (values(1,'峰哥',18),(2,'王哥',20),(3,'张哥',22));",
					Expected: []sql.Row{
						{1, "峰哥", 18},
						{2, "王哥", 20},
						{3, "张哥", 22},
					},
				},
				{
					Query: "select * from (values(1,'峰哥',18),(2,'王哥',20),(3,'张哥',22)) x(id,name,age);",
					Expected: []sql.Row{
						{1, "峰哥", 18},
						{2, "王哥", 20},
						{3, "张哥", 22},
					},
					Cols: []string{
						"id",
						"name",
						"age",
					},
				},
				{
					Query:    "select * from (values(1,'峰哥',18),(2,'王哥',20),(3,'张哥',22)) x(id,name,age) limit $1;",
					BindVars: []any{2}, // forcing this to use prepared statements
					Expected: []sql.Row{
						{1, "峰哥", 18},
						{2, "王哥", 20},
					},
					Cols: []string{
						"id",
						"name",
						"age",
					},
				},
			},
		},
	})
}
