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
						Query:       "INSERT INTO child VALUES (2, 2)",
						ExpectedErr: "Foreign key violation",
					},
				},
			},
			{
				Name: "text foreign key",
				SetUpScript: []string{
					`CREATE TABLE parent (a text PRIMARY KEY, b int)`,
					`CREATE TABLE child (a INT PRIMARY KEY, b text, FOREIGN KEY (b) REFERENCES parent(a))`,
					`INSERT INTO parent VALUES ('a', 1)`,
				},
				Assertions: []ScriptTestAssertion{
					{
						Query: "INSERT INTO child VALUES (1, 'a')",
					},
					{
						Query: "INSERT INTO child VALUES (2, 'a')",
					},
					{
						Query:       "INSERT INTO child VALUES (3, 'b')",
						ExpectedErr: "Foreign key violation",
					},
				},
			},
			{
				Name: "type compatibility",
				Focus: true,
				SetUpScript: []string{
					`CREATE TABLE child (
      id uuid DEFAULT public.gen_random_uuid() NOT NULL,
      name character varying NOT NULL
  );`,
					`CREATE TABLE parent (
      value text NOT NULL,
      comment text
  );`,
				},
				Assertions: []ScriptTestAssertion{
					{
						Query: `ALTER TABLE ONLY child
      ADD CONSTRAINT name_fkey FOREIGN KEY (name) REFERENCES parent(value) ON UPDATE RESTRICT ON DELETE RESTRICT;`,
					},
				},
			},
			{
				Name: "foreign key with dolt_add, dolt_commit",
				SetUpScript: []string{
					"create table test (pk int, \"value\" int, primary key(pk));",
					"CREATE TABLE test_info (id int, info varchar(255), test_pk int, primary key(id), foreign key (test_pk) references test(pk))",
					"INSERT INTO test VALUES (0, 0), (1, 1), (2,2)",
					"SELECT DOLT_ADD('.')",
				},
				Assertions: []ScriptTestAssertion{
					{
						Query: "SELECT * FROM dolt.status",
						Expected: []sql.Row{
							{"public.test", "t", "new table"},
							{"public.test_info", "t", "new table"},
						},
					},
					{
						Query:            "SELECT dolt_commit('-am', 'new tables')",
						SkipResultsCheck: true,
					},
					{
						Query:    "SELECT * FROM dolt.status",
						Expected: []sql.Row{},
					},
					{
						Query:    "SELECT * FROM dolt_schema_diff('HEAD', 'WORKING', 'test_info')",
						Expected: []sql.Row{},
					},
					{
						Query:    "INSERT INTO test_info VALUES (2, 'two', 2)",
						Expected: []sql.Row{},
					},
					{
						Query:       "INSERT INTO test_info VALUES (3, 'three', 3)",
						ExpectedErr: "Foreign key violation",
					},
					{
						Query: "SELECT * FROM test_info",
						Expected: []sql.Row{
							{2, "two", 2},
						},
					},
				},
			},
			{
				Name: "foreign key with explicit schema",
				SetUpScript: []string{
					"create table parent (pk int, \"value\" int, primary key(pk));",
					"CREATE TABLE child (id int, info varchar(255), test_pk int, primary key(id), foreign key (test_pk) references public.parent(pk))",
					"INSERT INTO parent VALUES (0, 0), (1, 1), (2,2)",
					"SELECT DOLT_ADD('.')",
				},
				Assertions: []ScriptTestAssertion{
					{
						Query: "SELECT * FROM dolt.status",
						Expected: []sql.Row{
							{"public.child", "t", "new table"},
							{"public.parent", "t", "new table"},
						},
					},
					{
						Query:            "SELECT dolt_commit('-am', 'new tables')",
						SkipResultsCheck: true,
					},
					{
						Query:    "SELECT * FROM dolt.status",
						Expected: []sql.Row{},
					},
					{
						Query:    "SELECT * FROM dolt_schema_diff('HEAD', 'WORKING', 'child')",
						Expected: []sql.Row{},
					},
					{
						Query:    "INSERT INTO child VALUES (2, 'two', 2)",
						Expected: []sql.Row{},
					},
					{
						Query:       "INSERT INTO child VALUES (3, 'three', 3)",
						ExpectedErr: "Foreign key violation",
					},
					{
						Query: "SELECT * FROM child",
						Expected: []sql.Row{
							{2, "two", 2},
						},
					},
				},
			},
			{
				Name: "foreign key in another schema with search path",
				SetUpScript: []string{
					"create schema parent",
					"create schema child",
					"create schema fake",
					"call dolt_commit('-Am', 'create schemas')",
					"set search_path to 'parent, child'",
					`create table parent.parent (pk int, val int, primary key(pk));`,
					`create table fake.parent (pk int, val int, primary key(pk));`,
					"CREATE TABLE child.child (id int, info varchar(255), test_pk int, primary key(id), foreign key (test_pk) references parent(pk))",
					"INSERT INTO parent VALUES (0, 0), (1, 1), (2,2)",
					"SELECT DOLT_ADD('.')",
				},
				Assertions: []ScriptTestAssertion{
					{
						Query: "SELECT * FROM dolt_status",
						Expected: []sql.Row{
							{"child.child", "t", "new table"},
							{"fake.parent", "t", "new table"},
							{"parent.parent", "t", "new table"},
						},
					},
					{
						Query:            "SELECT dolt_commit('-am', 'new tables')",
						SkipResultsCheck: true,
					},
					{
						Query:    "SELECT * FROM dolt_status",
						Expected: []sql.Row{},
					},
					{
						Query:    "SELECT * FROM dolt_schema_diff('HEAD', 'WORKING', 'child')",
						Expected: []sql.Row{},
					},
					{
						Query:    "INSERT INTO child VALUES (2, 'two', 2)",
						Expected: []sql.Row{},
					},
					{
						Query:       "INSERT INTO child VALUES (3, 'three', 3)",
						ExpectedErr: "Foreign key violation",
					},
					{
						Query: "SELECT * FROM child.child",
						Expected: []sql.Row{
							{2, "two", 2},
						},
					},
				},
			},
			{
				Name: "foreign key in another schema with search path, parent table not on search path",
				SetUpScript: []string{
					"create schema parent",
					"create schema child",
					"create schema fake",
					"call dolt_commit('-Am', 'create schemas')",
					"set search_path to 'child, fake'",
					`create table parent.parent (pk int, val int, primary key(pk));`,
					`create table fake.parent (pk int, val int, primary key(pk));`,
					"CREATE TABLE child.child (id int, info varchar(255), test_pk int, primary key(id), foreign key (test_pk) references parent.parent(pk))",
					"INSERT INTO parent.parent VALUES (0, 0), (1, 1), (2,2)",
					"SELECT DOLT_ADD('.')",
				},
				Assertions: []ScriptTestAssertion{
					{
						Query: "SELECT * FROM dolt.status",
						Expected: []sql.Row{
							{"child.child", "t", "new table"},
							{"fake.parent", "t", "new table"},
							{"parent.parent", "t", "new table"},
						},
					},
					{
						Query:            "SELECT dolt_commit('-am', 'new tables')",
						SkipResultsCheck: true,
					},
					{
						Query:    "SELECT * FROM dolt_status",
						Expected: []sql.Row{},
					},
					{
						Query:    "SELECT * FROM dolt_schema_diff('HEAD', 'WORKING', 'child')",
						Expected: []sql.Row{},
					},
					{
						Query:    "INSERT INTO child VALUES (2, 'two', 2)",
						Expected: []sql.Row{},
					},
					{
						Query:       "INSERT INTO child VALUES (3, 'three', 3)",
						ExpectedErr: "Foreign key violation",
					},
					{
						Query: "SELECT * FROM child.child",
						Expected: []sql.Row{
							{2, "two", 2},
						},
					},
				},
			},
			{
				Name: "foreign key in another schema, no search path",
				SetUpScript: []string{
					"create schema parent",
					"create schema child",
					"create schema fake",
					"call dolt_commit('-Am', 'create schemas')",
					`create table parent.parent (pk int, val int, primary key(pk));`,
					`create table fake.parent (pk int, val int, primary key(pk));`,
					"CREATE TABLE child.child (id int, info varchar(255), test_pk int, primary key(id), foreign key (test_pk) references parent.parent(pk))",
					"INSERT INTO parent.parent VALUES (0, 0), (1, 1), (2,2)",
					"SELECT DOLT_ADD('.')",
				},
				Assertions: []ScriptTestAssertion{
					{
						Query: "SELECT * FROM dolt_status",
						Expected: []sql.Row{
							{"child.child", "t", "new table"},
							{"fake.parent", "t", "new table"},
							{"parent.parent", "t", "new table"},
						},
					},
					{
						Query:            "SELECT dolt_commit('-am', 'new tables')",
						SkipResultsCheck: true,
					},
					{
						Query:    "SELECT * FROM dolt.status",
						Expected: []sql.Row{},
					},
					{
						Query:    "SELECT * FROM dolt_schema_diff('HEAD', 'WORKING', 'child')",
						Expected: []sql.Row{},
					},
					{
						Query:    "INSERT INTO child.child VALUES (2, 'two', 2)",
						Expected: []sql.Row{},
					},
					{
						Query:       "INSERT INTO child.child VALUES (3, 'three', 3)",
						ExpectedErr: "Foreign key violation",
					},
					{
						Query: "SELECT * FROM child.child",
						Expected: []sql.Row{
							{2, "two", 2},
						},
					},
				},
			},
			{
				Name: "add foreign key in another schema on search path",
				SetUpScript: []string{
					"create schema parent",
					"create schema child",
					"create schema fake",
					"call dolt_commit('-Am', 'create schemas')",
					"set search_path to 'child, parent'",
					`create table parent.parent (pk int, val int, primary key(pk));`,
					`create table fake.parent (pk int, val int, primary key(pk));`,
					"CREATE TABLE child.child (id int, info varchar(255), test_pk int, primary key(id))",
					"INSERT INTO parent.parent VALUES (0, 0), (1, 1), (2,2)",
					"SELECT DOLT_COMMIT('-Am', 'new tables')",
				},
				Assertions: []ScriptTestAssertion{
					{
						Query:    "INSERT INTO child.child VALUES (2, 'two', 2)",
						Expected: []sql.Row{},
					},
					{
						Query:            "ALTER TABLE child ADD FOREIGN KEY (test_pk) REFERENCES parent(pk)",
						SkipResultsCheck: true,
					},
					{
						Query:       "INSERT INTO child VALUES (3, 'three', 3)",
						ExpectedErr: "Foreign key violation",
					},
					{
						Query: "SELECT * FROM child",
						Expected: []sql.Row{
							{2, "two", 2},
						},
					},
				},
			},
			{
				Name: "add foreign key in another schema, parent table not on search path",
				SetUpScript: []string{
					"create schema parent",
					"create schema child",
					"create schema fake",
					"call dolt_commit('-Am', 'create schemas')",
					"set search_path to 'child, fake'",
					`create table parent.parent (pk int, val int, primary key(pk));`,
					`create table fake.parent (pk int, val int, primary key(pk));`,
					"CREATE TABLE child.child (id int, info varchar(255), test_pk int, primary key(id))",
					"INSERT INTO parent.parent VALUES (0, 0), (1, 1), (2,2)",
					"SELECT DOLT_COMMIT('-Am', 'new tables')",
				},
				Assertions: []ScriptTestAssertion{
					{
						Query:    "INSERT INTO child.child VALUES (2, 'two', 2)",
						Expected: []sql.Row{},
					},
					{
						Query:            "ALTER TABLE child ADD FOREIGN KEY (test_pk) REFERENCES parent.parent(pk)",
						SkipResultsCheck: true,
					},
					{
						Query:       "INSERT INTO child VALUES (3, 'three', 3)",
						ExpectedErr: "Foreign key violation",
					},
					{
						Query: "SELECT * FROM child",
						Expected: []sql.Row{
							{2, "two", 2},
						},
					},
				},
			},
			{
				Name: "add foreign key in another schema, no search path",
				SetUpScript: []string{
					"create schema parent",
					"create schema child",
					"create schema fake",
					"call dolt_commit('-Am', 'create schemas')",
					`create table parent.parent (pk int, val int, primary key(pk));`,
					`create table fake.parent (pk int, val int, primary key(pk));`,
					"CREATE TABLE child.child (id int, info varchar(255), test_pk int, primary key(id))",
					"INSERT INTO parent.parent VALUES (0, 0), (1, 1), (2,2)",
					"SELECT DOLT_COMMIT('-Am', 'new tables')",
				},
				Assertions: []ScriptTestAssertion{
					{
						Query:    "INSERT INTO child.child VALUES (2, 'two', 2)",
						Expected: []sql.Row{},
					},
					{
						Query:            "ALTER TABLE child.child ADD FOREIGN KEY (test_pk) REFERENCES parent.parent(pk)",
						SkipResultsCheck: true,
					},
					{
						Query:       "INSERT INTO child.child VALUES (3, 'three', 3)",
						ExpectedErr: "Foreign key violation",
					},
					{
						Query: "SELECT * FROM child.child",
						Expected: []sql.Row{
							{2, "two", 2},
						},
					},
				},
			},
			{
				Name: "drop foreign key in another schema, on search path",
				SetUpScript: []string{
					"create schema parent",
					"create schema child",
					"create schema fake",
					"call dolt_commit('-Am', 'create schemas')",
					"set search_path to 'child, parent'",
					`create table parent.parent (pk int, val int, primary key(pk));`,
					`create table fake.parent (pk int, val int, primary key(pk));`,
					"CREATE TABLE child.child (id int, info varchar(255), test_pk int, primary key(id))",
					"INSERT INTO parent.parent VALUES (0, 0), (1, 1), (2,2)",
					"SELECT DOLT_COMMIT('-Am', 'new tables')",
					"INSERT INTO child.child VALUES (2, 'two', 2)",
					"ALTER TABLE child.child ADD FOREIGN KEY (test_pk) REFERENCES parent(pk)",
				},
				Assertions: []ScriptTestAssertion{
					{
						Query:       "INSERT INTO child.child VALUES (3, 'three', 3)",
						ExpectedErr: "Foreign key violation",
					},
					{
						Query:            "alter table child DROP constraint child_ibfk_1",
						SkipResultsCheck: true,
					},
					{
						Query:    "INSERT INTO child.child VALUES (3, 'three', 3)",
						Expected: []sql.Row{},
					},
				},
			},
			{
				Name: "drop foreign key in another schema, no search path",
				Skip: true, // not getting the explicit schema name passed to the node
				SetUpScript: []string{
					"create schema parent",
					"create schema child",
					"create schema fake",
					"call dolt_commit('-Am', 'create schemas')",
					`create table parent.parent (pk int, val int, primary key(pk));`,
					`create table fake.parent (pk int, val int, primary key(pk));`,
					"CREATE TABLE child.child (id int, info varchar(255), test_pk int, primary key(id))",
					"INSERT INTO parent.parent VALUES (0, 0), (1, 1), (2,2)",
					"SELECT DOLT_COMMIT('-Am', 'new tables')",
					"INSERT INTO child.child VALUES (2, 'two', 2)",
					"ALTER TABLE child.child ADD FOREIGN KEY (test_pk) REFERENCES parent.parent(pk)",
				},
				Assertions: []ScriptTestAssertion{
					{
						Query:       "INSERT INTO child.child VALUES (3, 'three', 3)",
						ExpectedErr: "Foreign key violation",
					},
					{
						Query:            "alter table child.child DROP constraint child_ibfk_1",
						SkipResultsCheck: true,
					},
					{
						Query:    "INSERT INTO child.child VALUES (3, 'three', 3)",
						Expected: []sql.Row{},
					},
				},
			},
		},
	)
}
