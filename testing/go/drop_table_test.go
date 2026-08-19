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

func TestDropTableCascade(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "DROP TABLE CASCADE drops foreign keys referencing the table",
			SetUpScript: []string{
				`CREATE TABLE parent (pk INT4 PRIMARY KEY, v1 TEXT);`,
				`CREATE TABLE child (pk INT4 PRIMARY KEY, parent_pk INT4, CONSTRAINT child_parent_fk FOREIGN KEY (parent_pk) REFERENCES parent (pk));`,
				`INSERT INTO parent VALUES (1, 'one');`,
				`INSERT INTO child VALUES (10, 1);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					// Without CASCADE, dropping a table that a foreign key references is an error
					Query:       `DROP TABLE parent;`,
					ExpectedErr: "cannot drop table",
				},
				{
					// RESTRICT is the default behavior, so it errors the same way
					Query:       `DROP TABLE parent RESTRICT;`,
					ExpectedErr: "cannot drop table",
				},
				{
					Query:    `DROP TABLE parent CASCADE;`,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT * FROM parent;`,
					ExpectedErr: "not found",
				},
				{
					// The child table sticks around, only its foreign key constraint was dropped
					Query:    `SELECT * FROM child;`,
					Expected: []sql.Row{{10, 1}},
				},
				{
					// The foreign key constraint no longer exists, so this insert succeeds
					Query:    `INSERT INTO child VALUES (11, 99);`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM child ORDER BY pk;`,
					Expected: []sql.Row{{10, 1}, {11, 99}},
				},
			},
		},
		{
			Name: "DROP TABLE RESTRICT succeeds when nothing depends on the table",
			SetUpScript: []string{
				`CREATE TABLE test (pk INT4 PRIMARY KEY);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `DROP TABLE test RESTRICT;`,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT * FROM test;`,
					ExpectedErr: "not found",
				},
			},
		},
		{
			Name: "DROP TABLE CASCADE with no dependencies",
			SetUpScript: []string{
				`CREATE TABLE test (pk INT4 PRIMARY KEY);`,
				`INSERT INTO test VALUES (1);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `DROP TABLE test CASCADE;`,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT * FROM test;`,
					ExpectedErr: "not found",
				},
			},
		},
		{
			Name: "DROP TABLE CASCADE with a self-referential foreign key",
			SetUpScript: []string{
				`CREATE TABLE test (pk INT4 PRIMARY KEY, parent_pk INT4, CONSTRAINT test_self_fk FOREIGN KEY (parent_pk) REFERENCES test (pk));`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `DROP TABLE test CASCADE;`,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT * FROM test;`,
					ExpectedErr: "not found",
				},
			},
		},
		{
			Name: "DROP TABLE CASCADE with multiple tables",
			SetUpScript: []string{
				`CREATE TABLE parent (pk INT4 PRIMARY KEY);`,
				`CREATE TABLE middle (pk INT4 PRIMARY KEY, parent_pk INT4, CONSTRAINT middle_parent_fk FOREIGN KEY (parent_pk) REFERENCES parent (pk));`,
				`CREATE TABLE child (pk INT4 PRIMARY KEY, middle_pk INT4, CONSTRAINT child_middle_fk FOREIGN KEY (middle_pk) REFERENCES middle (pk));`,
			},
			Assertions: []ScriptTestAssertion{
				{
					// parent is listed before middle, which references it; middle's foreign key on parent is handled
					// by dropping both tables, while child's foreign key on middle is dropped by the cascade
					Query:    `DROP TABLE parent, middle CASCADE;`,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT * FROM parent;`,
					ExpectedErr: "not found",
				},
				{
					Query:       `SELECT * FROM middle;`,
					ExpectedErr: "not found",
				},
				{
					Query:    `INSERT INTO child VALUES (1, 99);`,
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "DROP TABLE IF EXISTS CASCADE",
			SetUpScript: []string{
				`CREATE TABLE parent (pk INT4 PRIMARY KEY);`,
				`CREATE TABLE child (pk INT4 PRIMARY KEY, parent_pk INT4, CONSTRAINT child_parent_fk FOREIGN KEY (parent_pk) REFERENCES parent (pk));`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `DROP TABLE IF EXISTS doesnotexist CASCADE;`,
					Expected: []sql.Row{},
				},
				{
					Query:       `DROP TABLE doesnotexist CASCADE;`,
					ExpectedErr: `table "doesnotexist" does not exist`,
				},
				{
					Query:    `DROP TABLE IF EXISTS doesnotexist, parent CASCADE;`,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT * FROM parent;`,
					ExpectedErr: "not found",
				},
				{
					Query:    `INSERT INTO child VALUES (1, 99);`,
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "DROP TABLE CASCADE with schema-qualified names",
			SetUpScript: []string{
				`CREATE SCHEMA sch1;`,
				`CREATE SCHEMA sch2;`,
				`CREATE TABLE sch1.parent (pk INT4 PRIMARY KEY);`,
				`CREATE TABLE sch2.child (pk INT4 PRIMARY KEY, parent_pk INT4, CONSTRAINT child_parent_fk FOREIGN KEY (parent_pk) REFERENCES sch1.parent (pk));`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       `DROP TABLE sch1.parent;`,
					ExpectedErr: "cannot drop table",
				},
				{
					Query:    `DROP TABLE sch1.parent CASCADE;`,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT * FROM sch1.parent;`,
					ExpectedErr: "not found",
				},
				{
					Query:    `INSERT INTO sch2.child VALUES (1, 99);`,
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "DROP TABLE CASCADE resolves tables on the search path",
			SetUpScript: []string{
				`CREATE SCHEMA sch1;`,
				`SET search_path TO sch1;`,
				`CREATE TABLE parent (pk INT4 PRIMARY KEY);`,
				`CREATE TABLE child (pk INT4 PRIMARY KEY, parent_pk INT4, CONSTRAINT child_parent_fk FOREIGN KEY (parent_pk) REFERENCES parent (pk));`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `DROP TABLE parent CASCADE;`,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT * FROM sch1.parent;`,
					ExpectedErr: "not found",
				},
				{
					Query:    `INSERT INTO sch1.child VALUES (1, 99);`,
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "DROP TABLE CASCADE drops sequences owned by the table",
			SetUpScript: []string{
				`CREATE TABLE test (pk SERIAL PRIMARY KEY, v1 TEXT);`,
				`INSERT INTO test (v1) VALUES ('one');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT nextval('test_pk_seq');`,
					Expected: []sql.Row{{2}},
				},
				{
					Query:    `DROP TABLE test CASCADE;`,
					Expected: []sql.Row{},
				},
				{
					// The sequence backing the SERIAL column was dropped along with the table
					Query:       `SELECT nextval('test_pk_seq');`,
					ExpectedErr: "does not exist",
				},
			},
		},
		{
			Name: "DROP TABLE CASCADE with a dependent view",
			SetUpScript: []string{
				`CREATE TABLE test (pk INT4 PRIMARY KEY, v1 TEXT);`,
				`INSERT INTO test VALUES (1, 'one');`,
				`CREATE VIEW test_view AS SELECT * FROM test;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					// TODO: view dependencies on tables are not yet tracked, so CASCADE cannot drop the dependent
					//  view; the table drop succeeds and leaves the view behind (which errors when used), matching
					//  the behavior of DROP TABLE without CASCADE
					Query:    `DROP TABLE test CASCADE;`,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT * FROM test_view;`,
					ExpectedErr: "references invalid table",
				},
			},
		},
	})
}
