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
						Query:            "alter table child DROP constraint child_test_pk_fkey;",
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
			{
				Name: "foreign key default naming",
				SetUpScript: []string{
					"CREATE TABLE webhooks (id varchar not null, id2 int8, primary key (id));",
					"CREATE UNIQUE INDEX idx1 on webhooks(id, id2);",
					"CREATE TABLE t33 (id varchar not null, webhook_id_fk varchar not null, webhook_id2_fk int8, foreign key (webhook_id_fk) references webhooks(id), foreign key (webhook_id_fk, webhook_id2_fk) references webhooks(id, id2), primary key (id));",
				},
				Assertions: []ScriptTestAssertion{
					{
						Query: "SELECT conname AS constraint_name FROM pg_constraint WHERE conrelid = 't33'::regclass  AND contype = 'f';",
						Expected: []sql.Row{
							{"t33_webhook_id_fk_fkey"},
							{"t33_webhook_id_fk_webhook_id2_fk_fkey"},
						},
					},
					{
						Query:    "ALTER TABLE t33 DROP CONSTRAINT t33_webhook_id_fk_fkey;",
						Expected: []sql.Row{},
					},
				},
			},
			{
				Name: "foreign key default naming, name collision ",
				SetUpScript: []string{
					"CREATE TABLE parent (id varchar not null primary key);",
					"CREATE TABLE child (id varchar primary key, constraint t33_webhook_id_fk_fkey foreign key (id) references parent(id));",
					"CREATE TABLE webhooks (id varchar not null, id2 int8, primary key (id));",
					"CREATE TABLE t33 (id varchar not null, webhook_id_fk varchar not null, foreign key (webhook_id_fk) references webhooks(id), primary key (id));",
				},
				Assertions: []ScriptTestAssertion{
					{
						Query: "SELECT conname AS constraint_name FROM pg_constraint WHERE conrelid = 't33'::regclass  AND contype = 'f';",
						Expected: []sql.Row{
							{"t33_webhook_id_fk_fkey1"},
						},
					},
					{
						Query:    "ALTER TABLE t33 DROP CONSTRAINT t33_webhook_id_fk_fkey1;",
						Expected: []sql.Row{},
					},
				},
			},
			{
				Name: "foreign key default naming, in column definition",
				SetUpScript: []string{
					"CREATE TABLE webhooks (id varchar not null, primary key (id));",
					"CREATE TABLE t33 (id varchar not null, webhook_id_fk varchar not null references webhooks(id), primary key (id));",
				},
				Assertions: []ScriptTestAssertion{
					{
						Query: "SELECT conname AS constraint_name FROM pg_constraint WHERE conrelid = 't33'::regclass  AND contype = 'f';",
						Expected: []sql.Row{
							{"t33_webhook_id_fk_fkey"},
						},
					},
					{
						Query:    "ALTER TABLE t33 DROP CONSTRAINT t33_webhook_id_fk_fkey;",
						Expected: []sql.Row{},
					},
				},
			},
			{
				Name: "foreign key custom naming",
				SetUpScript: []string{
					"CREATE TABLE webhooks (id VARCHAR NOT NULL, PRIMARY KEY (id));",
					"CREATE TABLE t33 (id VARCHAR NOT NULL, webhook_id_fk VARCHAR NOT NULL, CONSTRAINT foo1 FOREIGN KEY (webhook_id_fk) REFERENCES webhooks(id), PRIMARY KEY (id));",
				},
				Assertions: []ScriptTestAssertion{
					{
						Query:    "SELECT conname AS constraint_name FROM pg_constraint WHERE conrelid = 't33'::regclass AND contype = 'f';",
						Expected: []sql.Row{{"foo1"}},
					},
					{
						Query:    "ALTER TABLE t33 DROP CONSTRAINT foo1;",
						Expected: []sql.Row{},
					},
				},
			},
			{
				Name: "foreign key default naming, added through alter table",
				SetUpScript: []string{
					"CREATE TABLE webhooks (id varchar not null, primary key (id));",
					"CREATE TABLE t33 (id varchar not null, webhook_id_fk varchar not null, primary key (id));",
					"ALTER TABLE t33 ADD FOREIGN KEY (webhook_id_fk) REFERENCES webhooks(id);",
				},
				Assertions: []ScriptTestAssertion{
					{
						Query: "SELECT conname AS constraint_name FROM pg_constraint WHERE conrelid = 't33'::regclass  AND contype = 'f';",
						Expected: []sql.Row{
							{"t33_webhook_id_fk_fkey"},
						},
					},
					{
						Query:    "ALTER TABLE t33 DROP CONSTRAINT t33_webhook_id_fk_fkey;",
						Expected: []sql.Row{},
					},
				},
			},
		},
	)
}
