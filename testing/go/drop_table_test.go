// Copyright 2025 Dolthub, Inc.
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

func TestDropTable(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "DROP TABLE on table type on column",
			SetUpScript: []string{
				`CREATE TABLE test1 (pk INT4 PRIMARY KEY, v1 TEXT);`,
				`CREATE TABLE test2 (v1 test1);`,
				`INSERT INTO test2 VALUES (ROW(1, 'abc')::test1);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       `DROP TABLE test1;`,
					ExpectedErr: "cannot drop table test1 because other objects depend on it\ncolumn v1 of table test2 depends on type test1",
				},
				{
					Query:    `DROP TABLE test2;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `DROP TABLE test1;`,
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "DROP TABLE on table type on function parameter",
			SetUpScript: []string{
				`CREATE TABLE test (pk INT4 PRIMARY KEY, v1 TEXT);`,
				`CREATE FUNCTION example_func(t test) RETURNS INT4 AS $$ BEGIN RETURN t.pk * 2; END; $$ LANGUAGE plpgsql;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       `DROP TABLE test;`,
					ExpectedErr: "cannot drop table test because other objects depend on it\nfunction example_func(test) depends on type test",
				},
				{
					Query:    `DROP FUNCTION example_func(test);`,
					Expected: []sql.Row{},
				},
				{
					Query:    `DROP TABLE test;`,
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "DROP TABLE on table type on procedure parameter",
			SetUpScript: []string{
				`CREATE TABLE test1 (pk INT4 PRIMARY KEY, v1 TEXT);`,
				`CREATE TABLE test2 (v1 INT4);`,
				`CREATE PROCEDURE example_proc(input test1) AS $$ BEGIN INSERT INTO test2 VALUES (input.pk); END; $$ LANGUAGE plpgsql;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       `DROP TABLE test1;`,
					ExpectedErr: "cannot drop table test1 because other objects depend on it\nfunction example_proc(test1) depends on type test1",
				},
				{
					Query:    `DROP PROCEDURE example_proc(test1);`,
					Expected: []sql.Row{},
				},
				{
					Query:    `DROP TABLE test1;`,
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "DROP TABLE on table type on column concurrent",
			SetUpScript: []string{
				`CREATE TABLE test1 (pk INT4 PRIMARY KEY, v1 TEXT);`,
				`CREATE TABLE test2 (v1 test1);`,
				`INSERT INTO test2 VALUES (ROW(1, 'abc')::test1);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       `DROP TABLE test1;`,
					ExpectedErr: "cannot drop table test1 because other objects depend on it\ncolumn v1 of table test2 depends on type test1",
				},
				{
					Query:    `DROP TABLE test1, test2;`,
					Expected: []sql.Row{},
				},
			},
		},
	})
}
