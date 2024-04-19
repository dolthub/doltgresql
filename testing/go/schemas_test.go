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
				Query:    "create schema mySchema",
			},
			{
				Query:    "create schema otherSchema",
			},
			{
				Query: "CREATE TABLE mySchema.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
			},
			{
				Query: "CREATE TABLE otherSchema.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
			},
			{
				Query:    "insert into mySchema.test values (1,1), (2,2)",
				Expected: []sql.Row{},
			},
			{
				Query:    "insert into otherSchema.test values (3,3), (4,4)",
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
				Query: "SELECT * FROM otherSchema.test;",
				Expected: []sql.Row{
					{3, 3},
					{4, 4},
				},
			},
		},
	},
	{
		Name: "insert, update, delete, alter table with schema",
		Assertions: []ScriptTestAssertion{
			{
				Query:    "create schema mySchema",
			},
			{
				Query:    "create schema otherSchema",
			},
			{
				Query: "CREATE TABLE mySchema.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
			},
			{
				Query: "CREATE TABLE otherSchema.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
			},
			{
				Query:    "insert into mySchema.test values (1,1), (2,2)",
			},
			{
				Query:    "insert into otherSchema.test values (3,3), (4,4)",
				Expected: []sql.Row{},
			},
			{
				Query:    "update mySchema.test set v1 = 3 where pk = 1",
			},
			{
				Query:    "update otherSchema.test set v1 = 4 where pk = 3",
			},
			{
				Query:    "delete from mySchema.test where pk = 2",
			},
			{
				Query:    "delete from otherSchema.test where pk = 4",
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
				Query:    "create schema mySchema",
			},
			{
				Query: "CREATE TABLE mySchema.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
			},
			{
				Query: "CREATE TABLE otherSchema.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
				ExpectedErr: true,
			},
			{
				Query:    "insert into mySchema.test values (1,1), (2,2)",
			},
			{
				Query:    "insert into otherSchema.test values (3,3), (4,4)",
				ExpectedErr: true,
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
				ExpectedErr: true,
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
	// More tests:
	// * table name resolution among different schemas
	// * revision qualifiers
	// * drop tests
	// * table name resolution in more statements
}

func TestSchemas(t *testing.T) {
	RunScripts(t, SchemaTests)
}
