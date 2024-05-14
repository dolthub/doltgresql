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

var CommandTagTests = []ScriptTest{
	{
		Name: "set",
		Assertions: []ScriptTestAssertion{
			{
				Query:       "SET extra_float_digits = 3",
				ExpectedTag: "SET",
			},
		},
	},
	{
		Name: "show",
		Assertions: []ScriptTestAssertion{
			{
				Query:       "SHOW extra_float_digits",
				ExpectedTag: "SHOW",
			},
		},
	},
	{
		Name: "create database",
		Assertions: []ScriptTestAssertion{
			{
				Query:       "CREATE DATABASE mydb",
				ExpectedTag: "CREATE DATABASE",
			},
		},
	},
	{
		Name: "insert",
		SetUpScript: []string{
			"CREATE TABLE table0 (id int, name text)",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:       "INSERT INTO table0 VALUES (1,'Dolt'), (2,'Doltgres'), (3,'DoltHub')",
				ExpectedTag: "INSERT 0 3",
			},
			{
				Query:    "SELECT * FROM table0 order by id",
				Expected: []sql.Row{{1, "Dolt"}, {2, "Doltgres"}, {3, "DoltHub"}},
			},
			{
				Query:       "SELECT * FROM table0",
				ExpectedTag: "SELECT 3",
			},
		},
	},
	{
		Name: "update",
		SetUpScript: []string{
			"CREATE TABLE table0 (id int, name text)",
			"INSERT INTO table0 VALUES (1,'Dolt'), (2,'Doltgres'), (3,'DoltHub')",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:       "UPDATE table0 SET id = 4 WHERE name = 'Doltgres'",
				ExpectedTag: "UPDATE 1",
			},
			{
				Query:    "SELECT * FROM table0 order by id",
				Expected: []sql.Row{{1, "Dolt"}, {3, "DoltHub"}, {4, "Doltgres"}},
			},
			{
				Query:       "SELECT * FROM table0 WHERE name <> 'Dolt'",
				ExpectedTag: "SELECT 2",
			},
		},
	},
	{
		Name: "delete",
		SetUpScript: []string{
			"CREATE TABLE table0 (id int, name text)",
			"INSERT INTO table0 VALUES (1,'Dolt'), (2,'Doltgres'), (3,'DoltHub')",
		},
		Assertions: []ScriptTestAssertion{

			{
				Query:       "DELETE FROM table0",
				ExpectedTag: "DELETE 3",
			},
			{
				Query:    "SELECT * FROM table0 order by id",
				Expected: []sql.Row{},
			},
			{
				Query:       "SELECT * FROM table0",
				ExpectedTag: "SELECT 0",
			},
		},
	},
}

func TestServerMessages(t *testing.T) {
	RunScripts(t, CommandTagTests)
}
