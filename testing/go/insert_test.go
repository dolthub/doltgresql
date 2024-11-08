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

func TestInsert(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "simple insert",
			SetUpScript: []string{
				"CREATE TABLE mytable (id INT PRIMARY KEY, name TEXT)",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "INSERT INTO mytable (id, name) VALUES (1, 'hello')",
					SkipResultsCheck: true,
				},
				{
					Query: "INSERT INTO mytable (ID, naME) VALUES (2, 'world')",
					SkipResultsCheck: true,
				},
				{
					Query: "SELECT * FROM mytable order by id",
					Expected: []sql.Row{
						{1, "hello"},
						{2, "world"},
					},
				},
			},
		},
		{
			Name: "keyless insert",
			SetUpScript: []string{
				"CREATE TABLE mytable (id INT, name TEXT)",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "INSERT INTO mytable (id, name) VALUES (1, 'hello')",
					SkipResultsCheck: true,
				},
				{
					Query: "INSERT INTO mytable (ID, naME) VALUES (2, 'world')",
					SkipResultsCheck: true,
				},
				{
					Query: "SELECT * FROM mytable order by id",
					Expected: []sql.Row{
						{1, "hello"},
						{2, "world"},
					},
				},
			},
		},
		{
			Name: "on conflict clause",
			SetUpScript: []string{
				"CREATE TABLE mytable (id INT primary key, name TEXT)",
				"create table t2 (id int primary key, c1 text, c2 text)",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "INSERT INTO mytable (id, name) VALUES (1, 'hello')",
					SkipResultsCheck: true,
				},
				{
					Query: "INSERT INTO mytable (ID, naME) VALUES (2, 'world')",
					SkipResultsCheck: true,
				},
				{
					Query: "INSERT INTO mytable (ID, naME) VALUES (1, 'world') ON CONFLICT (id) DO UPDATE SET name = 'world'",
					SkipResultsCheck: true,
				},
				{
					Query: "INSERT INTO mytable (ID, naME) VALUES (2, 'hello') ON CONFLICT (id) DO UPDATE SET name = 'conflict'",
					SkipResultsCheck: true,
				},
				{
					Query: "INSERT INTO mytable (ID, naME) VALUES (1, 'not inserted') ON CONFLICT (id) DO NOTHING",
				},
				{
					Query: "SELECT * FROM mytable order by id",
					Expected: []sql.Row{
						{1, "world"},
						{2, "conflict"},
					},
				},
				{
					Query: "INSERT INTO mytable (ID, naME) VALUES (1, 'hello') ON CONFLICT (id) DO UPDATE set name = concat('new', name)",
				},
				{
					Query: "SELECT * FROM mytable order by id",
					Expected: []sql.Row{
						{1, "newworld"},
						{2, "conflict"},
					},
				},
				{
					Query: "INSERT INTO t2 (id, c1, c2) VALUES (1, 'hello', 'world'), (2, 'world', 'hello')",
					SkipResultsCheck: true,
				},
				{
					Query: "INSERT INTO t2 (id, c1, c2) VALUES (1, 'hello', 'world') ON CONFLICT (id) DO UPDATE SET c1 = 'conflict', c2 = c1",
					SkipResultsCheck: true,
				},
				{
					Query: "INSERT INTO t2 (id, c1, c2) VALUES (2, 'hello', 'world') ON CONFLICT (id) DO UPDATE SET c2 = c1",
					SkipResultsCheck: true,
				},
				{
					Query: "SELECT * FROM t2 order by id",
					Expected: []sql.Row{
						{1, "conflict", "conflict"},
						{2, "world", "world"},
					},
				},
			},
		},
	})
}