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

import "testing"

func TestForeignKeys(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "simple foreign key",
			SetUpScript: []string{
				`CREATE TABLE parent (a INT PRIMARY KEY, b int)`,
				`CREATE TABLE child (a INT PRIMARY KEY, b INT, FOREIGN KEY (b) REFERENCES parent(a))`,
				`INSERT INTO parent VALUES (1, 1)`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "INSERT INTO child VALUES (1, 1)",
				},
				{
					Query: "INSERT INTO child VALUES (2, 1)",
				},
				{
					Query: "INSERT INTO child VALUES (2, 2)",
					ExpectedErr: "Foreign key violation",
				},
			},
		},
		{
			Name: "foreign key with dolt_commit",
			Focus: true,
			SetUpScript: []string{
				"create table test (pk int, \"value\" int, primary key(pk));",
				"CREATE TABLE test_info (id int, info varchar(255), test_pk int, primary key(id), foreign key (test_pk) references test(pk))",
			},
		},
	})
}
			