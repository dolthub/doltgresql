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
			"CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
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
			"CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
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
		Focus: true,
		Assertions: []ScriptTestAssertion{
			{
				Query:    "create schema mySchema",
				Expected: []sql.Row{},
			},
			{
				Query: "CREATE TABLE mySchema.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
			},
			{
				Query: "SELECT * FROM mySchema.test;",
				Expected: []sql.Row{
					{1, 1},
					{2, 2},
				},
			},
		},
	},
	{
		Name: "create new database and new schema",
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
