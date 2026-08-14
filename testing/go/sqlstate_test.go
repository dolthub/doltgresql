// Copyright 2026 Dolthub, Inc.
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
)

// TestSQLStateCodes asserts the SQLSTATE reported for the error classes that castSQLError maps.
// Reporting the proper SQLSTATE is more than cosmetic: errors that reach a client with an XX-class
// (internal error) code are treated as critical failures by some clients — Npgsql, for example,
// closes the connection on any XX-class error — and error-handling code in applications and ORMs
// dispatches on these codes.
func TestSQLStateCodes(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "class 22 data exception",
			Assertions: []ScriptTestAssertion{
				{
					Query:           "SELECT 1/0",
					ExpectedErr:     "division by zero",
					ExpectedErrCode: "22012",
				},
				{
					Query:           "SELECT 5 % 0",
					ExpectedErr:     "division by zero",
					ExpectedErrCode: "22012",
				},
				{
					Query:           "SELECT 'abc'::int4",
					ExpectedErr:     "invalid input syntax for type int4",
					ExpectedErrCode: "22P02",
				},
				{
					Query:           "SELECT (2000000000)::int4 + (2000000000)::int4",
					ExpectedErr:     "integer out of range",
					ExpectedErrCode: "22003",
				},
				{
					Query:           "SELECT acos(2.0)",
					ExpectedErr:     "input is out of range",
					ExpectedErrCode: "22003",
				},
			},
		},
		{
			Name: "class 23 integrity constraint violation",
			SetUpScript: []string{
				"CREATE TABLE constraints_t (id INT PRIMARY KEY, u INT UNIQUE, n INT NOT NULL, c INT CHECK (c > 0))",
				"INSERT INTO constraints_t VALUES (1, 1, 1, 1)",
				"CREATE TABLE fk_parent (id INT PRIMARY KEY)",
				"CREATE TABLE fk_child (id INT PRIMARY KEY, pid INT REFERENCES fk_parent(id))",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:           "INSERT INTO constraints_t VALUES (1, 2, 2, 2)",
					ExpectedErr:     "duplicate primary key given",
					ExpectedErrCode: "23505",
				},
				{
					Query:           "INSERT INTO constraints_t VALUES (2, 1, 2, 2)",
					ExpectedErr:     "duplicate unique key given",
					ExpectedErrCode: "23505",
				},
				{
					Query:           "INSERT INTO constraints_t VALUES (2, 2, NULL, 2)",
					ExpectedErr:     "non-nullable",
					ExpectedErrCode: "23502",
				},
				{
					Query:           "INSERT INTO constraints_t VALUES (2, 2, 2, -1)",
					ExpectedErr:     "Check constraint",
					ExpectedErrCode: "23514",
				},
				{
					Query:           "INSERT INTO fk_child VALUES (1, 99)",
					ExpectedErr:     "Foreign key violation",
					ExpectedErrCode: "23503",
				},
			},
		},
		{
			Name: "class 42 undefined and duplicate objects",
			SetUpScript: []string{
				"CREATE TABLE existing_t (id INT PRIMARY KEY)",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:           "SELEC 1",
					ExpectedErr:     "syntax error",
					ExpectedErrCode: "42601",
				},
				{
					Query:           "SELECT * FROM no_such_table",
					ExpectedErr:     "table not found",
					ExpectedErrCode: "42P01",
				},
				{
					Query:           "SELECT no_such_col FROM existing_t",
					ExpectedErr:     "could not be found",
					ExpectedErrCode: "42703",
				},
				{
					Query:           "SELECT no_such_function(1)",
					ExpectedErr:     "not found",
					ExpectedErrCode: "42883",
				},
				{
					Query:           "CREATE TABLE existing_t (id INT PRIMARY KEY)",
					ExpectedErr:     "already exists",
					ExpectedErrCode: "42P07",
				},
				{
					Query:           "SELECT 1::no_such_type",
					ExpectedErr:     "unable to resolve type",
					ExpectedErrCode: "42704",
				},
			},
		},
		{
			Name: "class 3D undefined database",
			Assertions: []ScriptTestAssertion{
				{
					Query:           "SELECT * FROM no_such_db.public.tbl",
					ExpectedErr:     "database not found",
					ExpectedErrCode: "3D000",
				},
			},
		},
		{
			Name: "class 21 cardinality violation",
			SetUpScript: []string{
				"CREATE TABLE two_rows (id INT PRIMARY KEY)",
				"INSERT INTO two_rows VALUES (1), (2)",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:           "SELECT (SELECT id FROM two_rows)",
					ExpectedErr:     "more than 1 row",
					ExpectedErrCode: "21000",
				},
			},
		},
	})
}
