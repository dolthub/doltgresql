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

var SchemaTests = []ScriptTest{
	{
		Name: "implicit schema with index gets created in public schema",
		SetUpScript: []string{
			"create table employees (id int, last_name varchar(255), first_name varchar(255), primary key(id));",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "INSERT INTO employees VALUES (1, 'John', 'Doe'), (2, 'Jane', 'Doe');",
			},
			{
				Query: "SELECT * FROM employees;",
				Expected: []sql.Row{
					{1, "John", "Doe"},
					{2, "Jane", "Doe"},
				},
			},
			{
				Query: "SELECT * FROM public.employees;",
				Expected: []sql.Row{
					{1, "John", "Doe"},
					{2, "Jane", "Doe"},
				},
			},
		},
	},
	{
		Name: "table gets created in public schema by default",
		SetUpScript: []string{
			"CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "INSERT INTO public.test VALUES (1, 1), (2, 2);",
			},
			{
				Query: "SELECT * FROM public.test;",
				Expected: []sql.Row{
					{1, 1},
					{2, 2},
				},
			},
		},
	},
	{
		Name: "table creation respects search_path",
		SetUpScript: []string{
			"create schema postgres",
			// "$user" comes before public by default
			"CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "INSERT INTO postgres.test VALUES (1, 1), (2, 2);",
			},
			{
				Query: "SELECT * FROM postgres.test;",
				Expected: []sql.Row{
					{1, 1},
					{2, 2},
				},
			},
		},
	},
	{
		Name: "drop table",
		SetUpScript: []string{
			"CREATE TABLE t1 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
			"CREATE TABLE t2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
			"CREATE TABLE t3 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
			"CREATE TABLE t4 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "INSERT INTO t1 VALUES (1, 1), (2, 2);",
			},
			{
				Query: "INSERT INTO t2 VALUES (3, 3), (4, 4);",
			},
			{
				Query: "INSERT INTO t3 VALUES (1, 1), (2, 2);",
			},
			{
				Query: "INSERT INTO t4 VALUES (3, 3), (4, 4);",
			},
			{
				Query: "drop table t1",
			},
			{
				Query: "drop table public.t2",
			},
			{
				Query: "drop table if exists t3",
			},
			{
				Query: "drop table if exists public.t4",
			},
			{
				Query:       "INSERT INTO t1 VALUES (1, 1), (2, 2);",
				ExpectedErr: "not found",
			},
			{
				Query:       "INSERT INTO t2 VALUES (3, 3), (4, 4);",
				ExpectedErr: "not found",
			},
			{
				Query:       "INSERT INTO t3 VALUES (1, 1), (2, 2);",
				ExpectedErr: "not found",
			},
			{
				Query:       "INSERT INTO t4 VALUES (3, 3), (4, 4);",
				ExpectedErr: "not found",
			},
			{
				Query: "drop table if exists t1",
			},
			{
				Query:       "drop table t1",
				ExpectedErr: "not found",
			},
		},
	},
	{
		Name: "alter table",
		Skip: true, // alter table not supported yet
		SetUpScript: []string{
			"CREATE TABLE t1 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
			"CREATE TABLE t2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "alter table t1 add column v2 BIGINT;",
			},
			{
				Query: "alter table public.t2 add column v2 BIGINT;",
			},
			{
				Query: "INSERT INTO t1 VALUES (5, 5, 5), (6, 6, 6);",
			},
			{
				Query: "INSERT INTO public.t2 VALUES (7, 7, 7), (8, 8, 8);",
			},
		},
	},
	{
		Name: "table creation fails with no schema available",
		SetUpScript: []string{
			"set search_path to ''",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:       "CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
				ExpectedErr: "no schema has been selected to create in",
			},
			{
				Query: `set search_path to '"$user"'`,
			},
			{
				// only available schema doesn't exist
				Query:       "CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
				ExpectedErr: "no schema has been selected to create in",
			},
			{
				Query: `create schema postgres`,
			},
			{
				Query: "CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
			},
		},
	},
	{
		Name: "search path returns user table before public table",
		SetUpScript: []string{
			"create schema postgres",
			"CREATE TABLE public.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "INSERT INTO public.test VALUES (1, 1);",
			},
			{
				Query: "SELECT * FROM test;",
				Expected: []sql.Row{
					{1, 1},
				},
			},
			{
				Query: "CREATE TABLE postgres.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
			},
			{
				Query:    "INSERT INTO postgres.test VALUES (2, 2);",
				Expected: []sql.Row{},
			},
			{
				Query: "SELECT * FROM test;",
				Expected: []sql.Row{
					{2, 2},
				},
			},
		},
	},
	{
		Name: "empty search path will not resolve tables",
		SetUpScript: []string{
			"CREATE TABLE public.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
			"set search_path to ''",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "INSERT INTO public.test VALUES (1, 1);",
			},
			{
				Query:       "SELECT * FROM test;",
				ExpectedErr: "table not found",
			},
			{
				Query: "SELECT * FROM public.test;",
				Expected: []sql.Row{
					{1, 1},
				},
			},
		},
	},
	{
		Name: "public schema qualifier",
		SetUpScript: []string{
			"CREATE TABLE public.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "INSERT INTO public.test VALUES (1, 1), (2, 2);",
				Expected: []sql.Row{},
			},
			{
				Query: "SELECT * FROM public.test;",
				Expected: []sql.Row{
					{1, 1},
					{2, 2},
				},
			},
		},
	},
	{
		Name: "public schema qualifier, multiple tables",
		SetUpScript: []string{
			"CREATE TABLE public.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
			"CREATE TABLE public.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "INSERT INTO public.test VALUES (1, 1), (2, 2);",
				Expected: []sql.Row{},
			},
			{
				Query:    "INSERT INTO public.test2 VALUES (3, 3), (4, 4);",
				Expected: []sql.Row{},
			},
			{
				Query: "SELECT * FROM public.test;",
				Expected: []sql.Row{
					{1, 1},
					{2, 2},
				},
			},
			{
				Query: "SELECT * FROM public.test2;",
				Expected: []sql.Row{
					{3, 3},
					{4, 4},
				},
			},
		},
	},
	{
		Name: "public db and schema qualifier",
		SetUpScript: []string{
			"CREATE TABLE postgres.public.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "INSERT INTO postgres.public.test VALUES (1, 1), (2, 2);",
				Expected: []sql.Row{},
			},
			{
				Query: "SELECT * FROM postgres.public.test;",
				Expected: []sql.Row{
					{1, 1},
					{2, 2},
				},
			},
		},
	},
	{
		Name: "create new schema",
		Assertions: []ScriptTestAssertion{
			{
				Query: "create schema mySchema",
			},
			{
				Query: "create schema otherSchema",
			},
			{
				Query: "CREATE TABLE mySchema.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
			},
			{
				Query: "insert into mySchema.test values (1,1), (2,2)",
			},
			{
				Query: "CREATE TABLE otherSchema.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
			},
			{
				Query: "insert into otherSchema.test values (3,3), (4,4)",
			},
			{
				Query: "SELECT * FROM mySchema.test;",
				Expected: []sql.Row{
					{1, 1},
					{2, 2},
				},
			},
			{
				Query: "SELECT * FROM otherSchema.test;",
				Expected: []sql.Row{
					{3, 3},
					{4, 4},
				},
			},
		},
	},
	{
		Name: "schema already exists",
		Assertions: []ScriptTestAssertion{
			{
				Query: "create schema mySchema",
			},
			{
				Query:       "create schema MYSCHEMA",
				ExpectedErr: "can't create schema myschema; schema exists",
			},
		},
	},
	{
		Name: "create new schema (diff order)",
		Assertions: []ScriptTestAssertion{
			{
				Query: "create schema mySchema",
			},
			{
				Query: "create schema otherSchema",
			},
			{
				Query: "CREATE TABLE mySchema.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
			},
			{
				Query:    "insert into mySchema.test values (1,1), (2,2)",
				Expected: []sql.Row{},
			},
			{
				Query: "SELECT * FROM mySchema.test;",
				Expected: []sql.Row{
					{1, 1},
					{2, 2},
				},
			},
			{
				Query: "CREATE TABLE otherSchema.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
			},
			{
				Query:    "insert into otherSchema.test values (3,3), (4,4)",
				Expected: []sql.Row{},
			},
			{
				Query: "SELECT * FROM otherSchema.test;",
				Expected: []sql.Row{
					{3, 3},
					{4, 4},
				},
			},
		},
	},
	{
		Name: "insert, update, delete with schema",
		Assertions: []ScriptTestAssertion{
			{
				Query: "create schema mySchema",
			},
			{
				Query: "create schema otherSchema",
			},
			{
				Query: "CREATE TABLE mySchema.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
			},
			{
				Query: "CREATE TABLE otherSchema.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
			},
			{
				Query: "insert into mySchema.test values (1,1), (2,2)",
			},
			{
				Query:    "insert into otherSchema.test values (3,3), (4,4)",
				Expected: []sql.Row{},
			},
			{
				Query: "update mySchema.test set v1 = 3 where pk = 1",
			},
			{
				Query: "update otherSchema.test set v1 = 4 where pk = 3",
			},
			{
				Query: "delete from mySchema.test where pk = 2",
			},
			{
				Query: "delete from otherSchema.test where pk = 4",
			},
			{
				Query: "SELECT * FROM mySchema.test;",
				Expected: []sql.Row{
					{1, 3},
				},
			},
			{
				Query: "SELECT * FROM otherSchema.test;",
				Expected: []sql.Row{
					{3, 4},
				},
			},
		},
	},
	{
		Name: "schema does not exist",
		Assertions: []ScriptTestAssertion{
			{
				Query: "create schema mySchema",
			},
			{
				Query: "CREATE TABLE mySchema.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
			},
			{
				Query:       "CREATE TABLE otherSchema.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
				ExpectedErr: "database schema not found",
			},
			{
				Query: "insert into mySchema.test values (1,1), (2,2)",
			},
			{
				Query:       "insert into otherSchema.test values (3,3), (4,4)",
				ExpectedErr: "database schema not found",
			},
			{
				Query: "SELECT * FROM mySchema.test;",
				Expected: []sql.Row{
					{1, 1},
					{2, 2},
				},
			},
			{
				Query:       "SELECT * FROM otherSchema.test;",
				ExpectedErr: "database schema not found",
			},
		},
	},
	{
		Name: "create new database and new schema",
		SetUpScript: []string{
			"CREATE DATABASE db2;",
			"USE db2;",
			"create schema schema2;",
			"use postgres",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "CREATE TABLE db2.schema2.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
			},
			{
				Query: "INSERT INTO db2.schema2.test VALUES (1, 1), (2, 2);",
			},
			{
				Query: "SELECT * FROM db2.schema2.test;",
				Expected: []sql.Row{
					{1, 1},
					{2, 2},
				},
			},
		},
	},
	{
		Name: "non default schema qualifier",
		SetUpScript: []string{
			"CREATE SCHEMA myschema",
			"SET search_path = 'myschema'",
			"CREATE TABLE mytbl (pk BIGINT PRIMARY KEY, v1 BIGINT);",
			"INSERT INTO mytbl VALUES (1, 1), (2, 2)",
			"CREATE VIEW myvw AS SELECT pk+3, v1 from mytbl",
			"SET search_path TO DEFAULT",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW search_path",
				Expected: []sql.Row{{`"$user", public,`}},
			},
			{
				Query:       "SELECT * FROM mytbl;",
				ExpectedErr: "table not found",
			},
			{
				Query:       "SELECT * FROM myvw;",
				ExpectedErr: "table not found",
			},
			{
				Query: "SELECT * FROM myschema.mytbl;",
				Expected: []sql.Row{
					{1, 1},
					{2, 2},
				},
			},
			{
				Query: "SELECT * FROM myschema.myvw;",
				Expected: []sql.Row{
					{4, 1},
					{5, 2},
				},
			},
		},
	},
	{
		Name: "add new table in new schema, commit, status",
		SetUpScript: []string{
			"CREATE SCHEMA myschema",
			"Create table myschema.mytbl (pk BIGINT PRIMARY KEY, v1 BIGINT);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM dolt.status;",
				Expected: []sql.Row{
					{"myschema.mytbl", "f", "new table"},
					{"myschema", "f", "new schema"},
				},
			},
			{
				Query: "select dolt_add('.')",
				Expected: []sql.Row{
					{"{0}"},
				},
			},
			{
				Query: "SELECT * FROM dolt.status;",
				Expected: []sql.Row{
					{"myschema.mytbl", "t", "new table"},
					{"myschema", "t", "new schema"},
				},
			},
			{
				Query:            "select dolt_commit('-m', 'new table in new schema')",
				SkipResultsCheck: true,
			},
			{
				Query:    "SELECT * FROM dolt.status;",
				Expected: []sql.Row{},
			},
			{
				Query: "select message from dolt.log order by date desc limit 1",
				Expected: []sql.Row{
					{"new table in new schema"},
				},
			},
		},
	},
	{
		Name: "add new table in new schema, commit -Am",
		SetUpScript: []string{
			"CREATE SCHEMA myschema",
			"Create table myschema.mytbl (pk BIGINT PRIMARY KEY, v1 BIGINT);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM dolt.status;",
				Expected: []sql.Row{
					{"myschema.mytbl", "f", "new table"},
					{"myschema", "f", "new schema"},
				},
			},
			{
				Query:            "select dolt_commit('-Am', 'new table in new schema')",
				SkipResultsCheck: true,
			},
			{
				Query:    "SELECT * FROM dolt.status;",
				Expected: []sql.Row{},
			},
			{
				Query: "select message from dolt.log order by date desc limit 1",
				Expected: []sql.Row{
					{"new table in new schema"},
				},
			},
		},
	},
	{
		Name: "merge new table in new schema",
		SetUpScript: []string{
			"call dolt_checkout('-b', 'branch1')",
			"CREATE SCHEMA branchschema",
			"Create table branchschema.mytbl (pk BIGINT PRIMARY KEY, v1 BIGINT);",
			"INSERT INTO branchschema.mytbl VALUES (1, 1), (2, 2)",
			"Create table branchschema.mytbl2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
			"INSERT INTO branchschema.mytbl2 VALUES (3, 3), (4, 4)",
			"select dolt_commit('-Am', 'new table in new schema')",
			"call dolt_checkout('main')",
			"create schema mainschema",
			"create table mainschema.maintable (pk BIGINT PRIMARY KEY, v1 BIGINT);",
			"insert into mainschema.maintable values (5, 5), (6, 6)",
			"select dolt_commit('-Am', 'new table in main')",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "call dolt_merge('branch1')",
			},
			{
				Query: "SELECT * from mainschema.maintable",
				Expected: []sql.Row{
					{5, 5},
					{6, 6},
				},
			},
			{
				Query: "SELECT * from branchschema.mytbl",
				Expected: []sql.Row{
					{1, 1},
					{2, 2},
				},
			},
			{
				Query: "SELECT * from branchschema.mytbl2",
				Expected: []sql.Row{
					{3, 3},
					{4, 4},
				},
			},
		},
	},
	{
		Name: "create new schema with no tables, add and commit",
		SetUpScript: []string{
			"CREATE SCHEMA myschema",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SELECT * FROM dolt.status;",
				Expected: []sql.Row{{"myschema", "f", "new schema"}},
			},
			{
				Query: "select dolt_add('.')",
				Expected: []sql.Row{
					{"{0}"},
				},
			},
			{
				Query:    "SELECT * FROM dolt.status;",
				Expected: []sql.Row{{"myschema", "t", "new schema"}},
			},
			{
				Query:            "select dolt_commit('-m', 'new schema')",
				SkipResultsCheck: true,
			},
			{
				Query:    "SELECT * FROM dolt.status;",
				Expected: []sql.Row{},
			},
			{
				Query: "select message from dolt.log order by date desc limit 1",
				Expected: []sql.Row{
					{"new schema"},
				},
			},
		},
	},
	{
		Name: "USE branches",
		SetUpScript: []string{
			`USE "postgres/main"`,
			"CREATE SCHEMA myschema",
			"SET search_path = 'myschema'",
			"CREATE TABLE mytbl (pk BIGINT PRIMARY KEY, v1 BIGINT);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SELECT active_branch();",
				Expected: []sql.Row{{"main"}},
			},
			{
				Query: "SELECT current_schemas(true);",
				Expected: []sql.Row{
					{"{pg_catalog,myschema}"},
				},
			},
			{
				Query: "SELECT catalog_name, schema_name FROM information_schema.schemata;",
				Expected: []sql.Row{
					{"postgres", "dolt"},
					{"postgres", "myschema"},
					{"postgres", "pg_catalog"},
					{"postgres", "public"},
					{"postgres", "information_schema"},
					{"postgres/main", "dolt"},
					{"postgres/main", "myschema"},
					{"postgres/main", "pg_catalog"},
					{"postgres/main", "public"},
					{"postgres/main", "information_schema"}},
			},
			{
				Query: "SELECT schema_name FROM information_schema.schemata WHERE catalog_name = 'postgres/main';",
				Expected: []sql.Row{
					{"dolt"},
					{"myschema"},
					{"pg_catalog"},
					{"public"},
					{"information_schema"},
				},
			},
			{
				Query: "SELECT schema_name FROM information_schema.schemata WHERE catalog_name = 'postgres';",
				Expected: []sql.Row{
					{"dolt"},
					{"myschema"},
					{"pg_catalog"},
					{"public"},
					{"information_schema"},
				},
			},
			{
				Query:    "SELECT * FROM mytbl;",
				Expected: []sql.Row{},
			},
			{
				Query: "SELECT schemaname, tablename FROM pg_catalog.pg_tables WHERE schemaname != 'pg_catalog' AND schemaname != 'information_schema';",
				Expected: []sql.Row{
					{"myschema", "mytbl"},
				},
			},
			{
				Query: "SELECT * FROM dolt.status;",
				Expected: []sql.Row{
					{"myschema.mytbl", "f", "new table"},
					{"myschema", "f", "new schema"},
				},
			},
			{
				Query:            "SELECT dolt_commit('-A', '-m', 'Add mytbl');",
				SkipResultsCheck: true,
			},
			{
				Query:    "SELECT * FROM dolt.status;",
				Expected: []sql.Row{},
			},
			{
				Query:    "SELECT dolt_branch('newbranch')",
				Expected: []sql.Row{{"{0}"}},
			},
			{
				Query:    "USE 'postgres/newbranch'",
				Expected: []sql.Row{},
			},
			{
				Query: "SELECT current_schemas(true);",
				Expected: []sql.Row{
					{"{pg_catalog,myschema}"},
				},
			},
			{
				Query:    "CREATE SCHEMA newbranchschema;",
				Expected: []sql.Row{},
			},
			{
				Query:    "SET search_path TO 'newbranchschema'",
				Expected: []sql.Row{},
			},
			{
				Query: "SELECT current_schemas(true);",
				Expected: []sql.Row{
					{"{pg_catalog,newbranchschema}"},
				},
			},
			{
				Query:    "CREATE TABLE mytbl2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
				Expected: []sql.Row{},
			},
			{
				Query: "SELECT schema_name FROM information_schema.schemata WHERE catalog_name = 'postgres/newbranch';",
				Expected: []sql.Row{
					{"dolt"},
					{"myschema"},
					{"newbranchschema"},
					{"pg_catalog"},
					{"public"},
					{"information_schema"},
				},
			},
			{
				Query: "SELECT schema_name FROM information_schema.schemata WHERE catalog_name = 'postgres';",
				Expected: []sql.Row{
					{"dolt"},
					{"myschema"},
					{"pg_catalog"},
					{"public"},
					{"information_schema"},
				},
			},
			{
				Query: "SELECT schemaname, tablename FROM pg_catalog.pg_tables WHERE schemaname != 'pg_catalog' AND schemaname != 'information_schema';",
				Expected: []sql.Row{
					{"myschema", "mytbl"},
					{"newbranchschema", "mytbl2"},
				},
			},
		},
	},
	// More tests:
	// * alter table statements, when they work better
	// * AS OF (when supported)
	// * revision qualifiers
	// * drop schema
	// * more statement types
	// * INSERT INTO schema1 SELECT FROM schema2
	// * Subqueries accessing different schemas in the same SELECT
	// * Joins across schemas
	// * Table names matching schema names. For example, test1.test1, test1.test2, test2.test1, test2.test2
}

func TestSchemas(t *testing.T) {
	RunScripts(t, SchemaTests)
}
