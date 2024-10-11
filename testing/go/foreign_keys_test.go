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

func TestForeignKeys(t *testing.T) {
	RunScripts(
		t,
		[]ScriptTest{
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
				Name:  "foreign key with dolt_add, dolt_commit",
				Focus: true,
				SetUpScript: []string{
					"create table test (pk int, \"value\" int, primary key(pk));",
					"CREATE TABLE test_info (id int, info varchar(255), test_pk int, primary key(id), foreign key (test_pk) references test(pk))",
					"INSERT INTO test VALUES (0, 0), (1, 1), (2,2)",
					"SELECT DOLT_ADD('.')",
				},
				Assertions: []ScriptTestAssertion{
					{
						Query: "SELECT * FROM dolt_status",
						Expected: []sql.Row{
							{"public.test", 1, "new table"},
							{"public.test_info", 1, "new table"},
						},
					},
					{
						Query: "SELECT dolt_commit('-am', 'new tables')",
						SkipResultsCheck: true,
					},
					{
						Query: "SELECT * FROM dolt_status",
						Expected: []sql.Row{},
					},
					{
						Query: "SELECT * FROM dolt_schema_diff('HEAD', 'WORKING', 'test_info')",
						Expected: []sql.Row{
							{
								"public.test_info",
								"public.test_info",
								"CREATE TABLE `test_info` (\n" +
									"  `id` integer NOT NULL,\n" +
									"  `info` varchar(255),\n" +
									"  `test_pk` integer,\n" +
									"  PRIMARY KEY (`id`),\n" +
									"  KEY `test_pk` (`test_pk`)\n" +
									") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_bin;",
									"CREATE TABLE `test_info` (\n" +
									"  `id` integer NOT NULL,\n" +
									"  `info` varchar(255),\n" +
									"  `test_pk` integer,\n" +
									"  PRIMARY KEY (`id`),\n" +
									"  KEY `test_pk` (`test_pk`),\n" +
									"  CONSTRAINT `test_info_ibfk_1` FOREIGN KEY (`test_pk`) REFERENCES `test` (`pk`) ON DELETE NO ACTION ON UPDATE NO ACTION\n" +
									") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_bin;",
							},
						},
					},
				},
			},
		},
	)
}
			