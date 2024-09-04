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
		Skip: true,
		SetUpScript: []string{
			"CREATE DATABASE db2;",
			"USE db2;", // TODO: not a real postgres statement
			"create schema schema2;",
			"use postgres",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "CREATE TABLE db2.schema2.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
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
				Expected: []sql.Row{{`"$user", public`}},
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
		Name: "with branches",
		SetUpScript: []string{
			"CREATE SCHEMA myschema",
			"SET search_path = 'myschema'",
			"CREATE TABLE mytbl (pk BIGINT PRIMARY KEY, v1 BIGINT);",
			"set dolt_show_branch_databases to 1;", // TODO: Use `use db/branch` instead of dolt_checkout for these tests and remove this
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
				Skip:  true, // TODO: revision database not supported yet
				Query: "SELECT schema_name FROM information_schema.schemata WHERE catalog_name = 'postgres/main';",
				Expected: []sql.Row{
					{"myschema"},
					{"pg_catalog"},
					{"public"},
					{"information_schema"},
				},
			},
			{
				Query: "SELECT schema_name FROM information_schema.schemata WHERE catalog_name = 'postgres';",
				Expected: []sql.Row{
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
				Query: "SELECT schemaname, tablename FROM pg_catalog.pg_tables WHERE schemaname != 'pg_catalog';",
				Expected: []sql.Row{
					{"myschema", "mytbl"},
				},
			},
			{
				Query:    "SELECT dolt_checkout('-b', 'newbranch')",
				Expected: []sql.Row{{"{0,\"Switched to branch 'newbranch'\"}"}},
			},
			{
				Skip:     true, // TODO: ERROR: no schema has been selected to create in
				Query:    "CREATE TABLE mytbl2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
				Expected: []sql.Row{},
			},
			{
				Query:    "CREATE SCHEMA newbranchschema;",
				Expected: []sql.Row{},
			},
			{
				Query:    "SET search_path = 'newbranchschema';",
				Expected: []sql.Row{},
			},
			{
				Query:    "CREATE TABLE mytbl2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
				Expected: []sql.Row{},
			},
			{
				Query: "SELECT current_schemas(true)",
				Expected: []sql.Row{
					{"{pg_catalog,newbranchschema}"},
				},
			},
			{
				Skip:  true, // TODO: pg_catalog and public are not showing up
				Query: "SELECT schema_name FROM information_schema.schemata WHERE catalog_name = 'postgres/newbranch';",
				Expected: []sql.Row{
					{"newbranchschema"},
					{"pg_catalog"},
					{"public"},
					{"information_schema"},
				},
			},
			{
				Skip:  true, // TODO: Why are pg_catalog and public not showing up?
				Query: "SELECT schema_name FROM information_schema.schemata WHERE catalog_name = 'postgres';",
				Expected: []sql.Row{
					{"newbranchschema"},
					{"pg_catalog"},
					{"public"},
					{"information_schema"},
				},
			},
			{
				Skip:  true, // TODO: Getting error "database schema not found: pg_catalog"
				Query: "SELECT schemaname, tablename FROM pg_catalog.pg_tables WHERE schemaname != 'pg_catalog';",
				Expected: []sql.Row{
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
