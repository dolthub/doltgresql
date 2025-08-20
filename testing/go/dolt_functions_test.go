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

func TestDoltAdd(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Add all using dot",
			SetUpScript: []string{
				"CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
				"CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
				"CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
				"CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
				"CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
				"CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
				"CREATE TABLE t_unique (v1 INT4 UNIQUE);",
				"CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
				"CREATE TABLE t_trigger (v1 INT8);",
				"CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
				"CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
				"CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
				"INSERT INTO t_simple VALUES (1), (2), (3);",
				"INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
				"INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
				"INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
				"INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
				"INSERT INTO t_checked VALUES (1), (50), (100);",
				"INSERT INTO t_fk_parent VALUES (10), (20), (30);",
				"INSERT INTO t_fk_child VALUES (10), (20);",
				"INSERT INTO t_unique VALUES (7), (8), (9);",
				"INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
				"INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
				"INSERT INTO t_default_func (v2) VALUES (1), (2);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT COUNT(*) FROM dolt_status WHERE staged = 'f';",
					Expected: []sql.Row{{16}},
				},
				{
					Query:    "SELECT DOLT_ADD('.');",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    "SELECT COUNT(*) FROM dolt_status WHERE staged = 't';",
					Expected: []sql.Row{{16}},
				},
			},
		},
		{
			Name: "Add all using -A",
			SetUpScript: []string{
				"CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
				"CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
				"CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
				"CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
				"CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
				"CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
				"CREATE TABLE t_unique (v1 INT4 UNIQUE);",
				"CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
				"CREATE TABLE t_trigger (v1 INT8);",
				"CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
				"CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
				"CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
				"INSERT INTO t_simple VALUES (1), (2), (3);",
				"INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
				"INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
				"INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
				"INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
				"INSERT INTO t_checked VALUES (1), (50), (100);",
				"INSERT INTO t_fk_parent VALUES (10), (20), (30);",
				"INSERT INTO t_fk_child VALUES (10), (20);",
				"INSERT INTO t_unique VALUES (7), (8), (9);",
				"INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
				"INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
				"INSERT INTO t_default_func (v2) VALUES (1), (2);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT COUNT(*) FROM dolt_status WHERE staged = 'f';",
					Expected: []sql.Row{{16}},
				},
				{
					Query:    "SELECT DOLT_ADD('-A');",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    "SELECT COUNT(*) FROM dolt_status WHERE staged = 't';",
					Expected: []sql.Row{{16}},
				},
			},
		},
		{
			Name: "Add all individually",
			SetUpScript: []string{
				"CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
				"CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
				"CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
				"CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
				"CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
				"CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
				"CREATE TABLE t_unique (v1 INT4 UNIQUE);",
				"CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
				"CREATE TABLE t_trigger (v1 INT8);",
				"CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
				"CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
				"CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
				"INSERT INTO t_simple VALUES (1), (2), (3);",
				"INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
				"INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
				"INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
				"INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
				"INSERT INTO t_checked VALUES (1), (50), (100);",
				"INSERT INTO t_fk_parent VALUES (10), (20), (30);",
				"INSERT INTO t_fk_child VALUES (10), (20);",
				"INSERT INTO t_unique VALUES (7), (8), (9);",
				"INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
				"INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
				"INSERT INTO t_default_func (v2) VALUES (1), (2);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT COUNT(*) FROM dolt_status WHERE staged = 'f';",
					Expected: []sql.Row{{16}},
				},
				{
					Query: "SELECT DOLT_ADD('t_simple','t_composite','t_array','t_serial','t_default_simple'," +
						"'t_checked','t_fk_parent','t_fk_child','t_unique','t_generated','t_trigger','t_default_func');",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    "SELECT COUNT(*) FROM dolt_status WHERE staged = 't';",
					Expected: []sql.Row{{12}},
				},
				{
					Query:    "SELECT DOLT_ADD('f_trigger()','f_default()');",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    "SELECT COUNT(*) FROM dolt_status WHERE staged = 't';",
					Expected: []sql.Row{{14}},
				},
				{
					Query:    "SELECT DOLT_ADD('t_serial_pk_seq','t_trigger.trig_trigger');",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    "SELECT COUNT(*) FROM dolt_status WHERE staged = 't';",
					Expected: []sql.Row{{16}},
				},
			},
		},
	})
}

func TestDoltBranch(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "All branch options",
			SetUpScript: []string{
				"CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
				"CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
				"CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
				"CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
				"CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
				"CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
				"CREATE TABLE t_unique (v1 INT4 UNIQUE);",
				"CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
				"CREATE TABLE t_trigger (v1 INT8);",
				"CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
				"CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
				"CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
				"INSERT INTO t_simple VALUES (1), (2), (3);",
				"INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
				"INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
				"INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
				"INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
				"INSERT INTO t_checked VALUES (1), (50), (100);",
				"INSERT INTO t_fk_parent VALUES (10), (20), (30);",
				"INSERT INTO t_fk_child VALUES (10), (20);",
				"INSERT INTO t_unique VALUES (7), (8), (9);",
				"INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
				"INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
				"INSERT INTO t_default_func (v2) VALUES (1), (2);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT DOLT_ADD('-A');",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    "SELECT length(DOLT_COMMIT('-m', 'initial')::text) = 34;",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT DOLT_BRANCH('original');",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    "INSERT INTO t_simple VALUES (4);",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT DOLT_ADD('-A');",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    "SELECT length(DOLT_COMMIT('-m', 'initial')::text) = 34;",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT DOLT_BRANCH('-c', 'main', 'copy');",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    "SELECT DOLT_BRANCH('-c', '-f', 'original', 'forcecopy');",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    `SELECT * FROM t_simple;`,
					Expected: []sql.Row{{1}, {2}, {3}, {4}},
				},
				{
					Query:    `SELECT DOLT_CHECKOUT('main')`,
					Expected: []sql.Row{{`{0,"Already on branch 'main'"}`}},
				},
				{
					Query:    `SELECT DOLT_CHECKOUT('original')`,
					Expected: []sql.Row{{`{0,"Switched to branch 'original'"}`}},
				},
				{
					Query:    `SELECT * FROM t_simple;`,
					Expected: []sql.Row{{1}, {2}, {3}},
				},
				{
					Query:    `SELECT DOLT_CHECKOUT('copy')`,
					Expected: []sql.Row{{`{0,"Switched to branch 'copy'"}`}},
				},
				{
					Query:    `SELECT * FROM t_simple;`,
					Expected: []sql.Row{{1}, {2}, {3}, {4}},
				},
				{
					Query:    `SELECT DOLT_CHECKOUT('forcecopy')`,
					Expected: []sql.Row{{`{0,"Switched to branch 'forcecopy'"}`}},
				},
				{
					Query:    `SELECT * FROM t_simple;`,
					Expected: []sql.Row{{1}, {2}, {3}},
				},
				{
					Query:       `SELECT DOLT_BRANCH('-d', 'forcecopy')`,
					ExpectedErr: `Cannot delete checked out branch 'forcecopy'`,
				},
				{
					Query:    `SELECT DOLT_BRANCH('-d', 'original')`,
					Expected: []sql.Row{{`{0}`}},
				},
				{
					Query:    `SELECT DOLT_BRANCH('-m','copy','renamedcopy')`,
					Expected: []sql.Row{{`{0}`}},
				},
				{
					Query:       `SELECT DOLT_CHECKOUT('original')`,
					ExpectedErr: `'original' did not match any table`,
				},
				{
					Query:       `SELECT DOLT_CHECKOUT('copy')`,
					ExpectedErr: `'copy' did not match any table`,
				},
				{
					Query:    `SELECT DOLT_CHECKOUT('renamedcopy')`,
					Expected: []sql.Row{{`{0,"Switched to branch 'renamedcopy'"}`}},
				},
				{
					Query:    `SELECT * FROM t_simple;`,
					Expected: []sql.Row{{1}, {2}, {3}, {4}},
				},
			},
		},
	})
}

func TestDoltBranchStatus(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Smoke test",
			SetUpScript: []string{
				"CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
				"INSERT INTO t_simple VALUES (1), (2), (3);",
				"SELECT DOLT_ADD('-A');",
				"SELECT length(DOLT_COMMIT('-m', 'initial')::text) = 34;",
				"SELECT DOLT_BRANCH('original');",
				"INSERT INTO t_simple VALUES (4);",
				"SELECT DOLT_ADD('-A');",
				"SELECT length(DOLT_COMMIT('-m', 'initial')::text) = 34;",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT * FROM DOLT_BRANCH_STATUS('main', 'original');",
					Expected: []sql.Row{{"original", Numeric("0"), Numeric("1")}},
				},
				{
					Query:    "SELECT * FROM DOLT_BRANCH_STATUS('original', 'main');",
					Expected: []sql.Row{{"main", Numeric("1"), Numeric("0")}},
				},
			},
		},
	})
}

func TestDoltCheckout(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "All checkout options",
			SetUpScript: []string{
				"CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
				"CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
				"CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
				"CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
				"CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
				"CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
				"CREATE TABLE t_unique (v1 INT4 UNIQUE);",
				"CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
				"CREATE TABLE t_trigger (v1 INT8);",
				"CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
				"CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
				"CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
				"INSERT INTO t_simple VALUES (1), (2), (3);",
				"INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
				"INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
				"INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
				"INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
				"INSERT INTO t_checked VALUES (1), (50), (100);",
				"INSERT INTO t_fk_parent VALUES (10), (20), (30);",
				"INSERT INTO t_fk_child VALUES (10), (20);",
				"INSERT INTO t_unique VALUES (7), (8), (9);",
				"INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
				"INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
				"INSERT INTO t_default_func (v2) VALUES (1), (2);",
				"SELECT DOLT_ADD('-A');",
				"SELECT length(DOLT_COMMIT('-m', 'initial')::text) = 34;",
				"SELECT DOLT_BRANCH('original');",
				"INSERT INTO t_simple VALUES (4);",
				"INSERT INTO t_composite VALUES (3, 100);",
				"INSERT INTO t_array VALUES (ARRAY['stu']);",
				"INSERT INTO t_serial VALUES (DEFAULT);",
				"INSERT INTO t_default_simple VALUES (98, 8);",
				"INSERT INTO t_checked VALUES (99);",
				"INSERT INTO t_fk_parent VALUES (40);",
				"INSERT INTO t_fk_child VALUES (30);",
				"INSERT INTO t_unique VALUES (10);",
				"INSERT INTO t_generated (v1) VALUES (11);",
				"CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 33; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_trigger (v1) VALUES (2);",
				"CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 34; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_default_func (v2) VALUES (3);",
				"SELECT DOLT_ADD('-A');",
				"SELECT length(DOLT_COMMIT('-m', 'initial')::text) = 34;",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT DOLT_CHECKOUT('-b', 'checkoutbranch');",
					Expected: []sql.Row{{`{0,"Switched to branch 'checkoutbranch'"}`}},
				},
				{
					Query:    `SELECT * FROM t_simple;`,
					Expected: []sql.Row{{1}, {2}, {3}, {4}},
				},
				{
					Query:    "SELECT DOLT_CHECKOUT('checkoutbranch');",
					Expected: []sql.Row{{`{0,"Already on branch 'checkoutbranch'"}`}},
				},
				{
					Query:    `SELECT DOLT_CHECKOUT('main')`,
					Expected: []sql.Row{{`{0,"Switched to branch 'main'"}`}},
				},
				{
					Query:    `SELECT * FROM t_simple;`,
					Expected: []sql.Row{{1}, {2}, {3}, {4}},
				},
				{
					Query:    `SELECT DOLT_CHECKOUT('original')`,
					Expected: []sql.Row{{`{0,"Switched to branch 'original'"}`}},
				},
				{
					Query:    `SELECT * FROM t_simple;`,
					Expected: []sql.Row{{1}, {2}, {3}},
				},
			},
		},
	})
}

func TestDoltCherryPick(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Smoke test",
			SetUpScript: []string{
				"CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
				"CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
				"CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
				"CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
				"CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
				"CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
				"CREATE TABLE t_unique (v1 INT4 UNIQUE);",
				"CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
				"CREATE TABLE t_trigger (v1 INT8);",
				"CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
				"CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
				"CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
				"INSERT INTO t_simple VALUES (1), (2), (3);",
				"INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
				"INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
				"INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
				"INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
				"INSERT INTO t_checked VALUES (1), (50), (100);",
				"INSERT INTO t_fk_parent VALUES (10), (20), (30);",
				"INSERT INTO t_fk_child VALUES (10), (20);",
				"INSERT INTO t_unique VALUES (7), (8), (9);",
				"INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
				"INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
				"INSERT INTO t_default_func (v2) VALUES (1), (2);",
				"SELECT DOLT_ADD('-A');",
				"SELECT length(DOLT_COMMIT('-m', 'initial')::text) = 34;",
				"SELECT DOLT_BRANCH('original');",
				"INSERT INTO t_simple VALUES (4);",
				"INSERT INTO t_composite VALUES (3, 100);",
				"INSERT INTO t_array VALUES (ARRAY['stu']);",
				"INSERT INTO t_serial VALUES (DEFAULT);",
				"INSERT INTO t_default_simple VALUES (98, 8);",
				"INSERT INTO t_checked VALUES (99);",
				"INSERT INTO t_fk_parent VALUES (40);",
				"INSERT INTO t_fk_child VALUES (30);",
				"INSERT INTO t_unique VALUES (10);",
				"INSERT INTO t_generated (v1) VALUES (11);",
				"CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 33; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_trigger (v1) VALUES (2);",
				"CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 34; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_default_func (v2) VALUES (3);",
				"SELECT DOLT_ADD('-A');",
				"SELECT length(DOLT_COMMIT('-m', 'initial')::text) = 34;",
				`SELECT DOLT_CHECKOUT('original')`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM t_simple;`,
					Expected: []sql.Row{{1}, {2}, {3}},
				},
				{ // This returns a hash, so we need to take the consistent substring portion for testing
					Query:    `SELECT substring(DOLT_CHERRY_PICK('main')::text, 34);`,
					Expected: []sql.Row{{",0,0,0}"}},
				},
				{
					Query:    `SELECT * FROM t_simple;`,
					Expected: []sql.Row{{1}, {2}, {3}, {4}},
				},
			},
		},
	})
}

func TestDoltClean(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Clean all",
			SetUpScript: []string{
				"CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
				"CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
				"CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
				"CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
				"CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
				"CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
				"CREATE TABLE t_unique (v1 INT4 UNIQUE);",
				"CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
				"CREATE TABLE t_trigger (v1 INT8);",
				"CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
				"CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
				"CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
				"INSERT INTO t_simple VALUES (1), (2), (3);",
				"INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
				"INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
				"INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
				"INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
				"INSERT INTO t_checked VALUES (1), (50), (100);",
				"INSERT INTO t_fk_parent VALUES (10), (20), (30);",
				"INSERT INTO t_fk_child VALUES (10), (20);",
				"INSERT INTO t_unique VALUES (7), (8), (9);",
				"INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
				"INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
				"INSERT INTO t_default_func (v2) VALUES (1), (2);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT COUNT(*) FROM dolt_status;",
					Expected: []sql.Row{{16}},
				},
				{
					Query:    "SELECT DOLT_CLEAN();",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    "SELECT COUNT(*) FROM dolt_status;",
					Expected: []sql.Row{{0}},
				},
			},
		},
		{
			Name: "Clean all except one",
			SetUpScript: []string{
				"CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
				"CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
				"CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
				"CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
				"CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
				"CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
				"CREATE TABLE t_unique (v1 INT4 UNIQUE);",
				"CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
				"CREATE TABLE t_trigger (v1 INT8);",
				"CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
				"CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
				"CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
				"INSERT INTO t_simple VALUES (1), (2), (3);",
				"INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
				"INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
				"INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
				"INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
				"INSERT INTO t_checked VALUES (1), (50), (100);",
				"INSERT INTO t_fk_parent VALUES (10), (20), (30);",
				"INSERT INTO t_fk_child VALUES (10), (20);",
				"INSERT INTO t_unique VALUES (7), (8), (9);",
				"INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
				"INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
				"INSERT INTO t_default_func (v2) VALUES (1), (2);",
				"SELECT DOLT_ADD('t_simple');",
				"SELECT length(DOLT_COMMIT('-m', 'initial')::text) = 34;",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT COUNT(*) FROM dolt_status;",
					Expected: []sql.Row{{15}},
				},
				{
					Query:    "SELECT DOLT_CLEAN();",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    "SELECT COUNT(*) FROM dolt_status;",
					Expected: []sql.Row{{0}},
				},
				{
					Query:    `SELECT * FROM t_simple;`,
					Expected: []sql.Row{{1}, {2}, {3}},
				},
			},
		},
		{
			Name: "Clean all by name",
			SetUpScript: []string{
				"CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
				"CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
				"CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
				"CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
				"CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
				"CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
				"CREATE TABLE t_unique (v1 INT4 UNIQUE);",
				"CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
				"CREATE TABLE t_trigger (v1 INT8);",
				"CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
				"CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
				"CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
				"INSERT INTO t_simple VALUES (1), (2), (3);",
				"INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
				"INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
				"INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
				"INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
				"INSERT INTO t_checked VALUES (1), (50), (100);",
				"INSERT INTO t_fk_parent VALUES (10), (20), (30);",
				"INSERT INTO t_fk_child VALUES (10), (20);",
				"INSERT INTO t_unique VALUES (7), (8), (9);",
				"INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
				"INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
				"INSERT INTO t_default_func (v2) VALUES (1), (2);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT COUNT(*) FROM dolt_status;",
					Expected: []sql.Row{{16}},
				},
				{
					Query: "SELECT DOLT_CLEAN('t_simple','t_composite','t_array','t_serial','t_default_simple'," +
						"'t_checked','t_fk_parent','t_fk_child','t_unique','t_generated','t_trigger','t_default_func'," +
						"'f_trigger()','f_default()');",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    "SELECT COUNT(*) FROM dolt_status;",
					Expected: []sql.Row{{0}},
				},
			},
		},
		{
			Name: "Dry run",
			SetUpScript: []string{
				"CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
				"CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
				"CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
				"CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
				"CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
				"CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
				"CREATE TABLE t_unique (v1 INT4 UNIQUE);",
				"CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
				"CREATE TABLE t_trigger (v1 INT8);",
				"CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
				"CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
				"CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
				"INSERT INTO t_simple VALUES (1), (2), (3);",
				"INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
				"INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
				"INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
				"INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
				"INSERT INTO t_checked VALUES (1), (50), (100);",
				"INSERT INTO t_fk_parent VALUES (10), (20), (30);",
				"INSERT INTO t_fk_child VALUES (10), (20);",
				"INSERT INTO t_unique VALUES (7), (8), (9);",
				"INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
				"INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
				"INSERT INTO t_default_func (v2) VALUES (1), (2);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT COUNT(*) FROM dolt_status;",
					Expected: []sql.Row{{16}},
				},
				{
					Query:    "SELECT DOLT_CLEAN('--dry-run');",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    "SELECT COUNT(*) FROM dolt_status;",
					Expected: []sql.Row{{16}},
				},
			},
		},
	})
}

func TestDoltCommit(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Stage all then commit",
			SetUpScript: []string{
				"CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
				"CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
				"CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
				"CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
				"CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
				"CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
				"CREATE TABLE t_unique (v1 INT4 UNIQUE);",
				"CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
				"CREATE TABLE t_trigger (v1 INT8);",
				"CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
				"CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
				"CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
				"INSERT INTO t_simple VALUES (1), (2), (3);",
				"INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
				"INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
				"INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
				"INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
				"INSERT INTO t_checked VALUES (1), (50), (100);",
				"INSERT INTO t_fk_parent VALUES (10), (20), (30);",
				"INSERT INTO t_fk_child VALUES (10), (20);",
				"INSERT INTO t_unique VALUES (7), (8), (9);",
				"INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
				"INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
				"INSERT INTO t_default_func (v2) VALUES (1), (2);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT COUNT(*) FROM dolt_status;",
					Expected: []sql.Row{{16}},
				},
				{
					Query:    "SELECT DOLT_ADD('-A');",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    "SELECT length(DOLT_COMMIT('-m', 'initial')::text) = 34;",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT COUNT(*) FROM dolt_status;",
					Expected: []sql.Row{{0}},
				},
			},
		},
		{
			Name: "Commit with staging all",
			SetUpScript: []string{
				"CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
				"CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
				"CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
				"CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
				"CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
				"CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
				"CREATE TABLE t_unique (v1 INT4 UNIQUE);",
				"CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
				"CREATE TABLE t_trigger (v1 INT8);",
				"CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
				"CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
				"CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
				"INSERT INTO t_simple VALUES (1), (2), (3);",
				"INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
				"INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
				"INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
				"INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
				"INSERT INTO t_checked VALUES (1), (50), (100);",
				"INSERT INTO t_fk_parent VALUES (10), (20), (30);",
				"INSERT INTO t_fk_child VALUES (10), (20);",
				"INSERT INTO t_unique VALUES (7), (8), (9);",
				"INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
				"INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
				"INSERT INTO t_default_func (v2) VALUES (1), (2);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT COUNT(*) FROM dolt_status;",
					Expected: []sql.Row{{16}},
				},
				{
					Query:    "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 34;",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT COUNT(*) FROM dolt_status;",
					Expected: []sql.Row{{0}},
				},
			},
		},
		{
			Name: "Commit with staging modifications",
			SetUpScript: []string{
				"CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
				"CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
				"CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
				"CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
				"CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
				"CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
				"CREATE TABLE t_unique (v1 INT4 UNIQUE);",
				"CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
				"CREATE TABLE t_trigger (v1 INT8);",
				"CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
				"CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
				"CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
				"INSERT INTO t_simple VALUES (1), (2), (3);",
				"INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
				"INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
				"INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
				"INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
				"INSERT INTO t_checked VALUES (1), (50), (100);",
				"INSERT INTO t_fk_parent VALUES (10), (20), (30);",
				"INSERT INTO t_fk_child VALUES (10), (20);",
				"INSERT INTO t_unique VALUES (7), (8), (9);",
				"INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
				"INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
				"INSERT INTO t_default_func (v2) VALUES (1), (2);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT COUNT(*) FROM dolt_status;",
					Expected: []sql.Row{{16}},
				},
				{
					Query:    "SELECT DOLT_ADD('t_simple','t_composite','t_array');",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    "SELECT length(DOLT_COMMIT('-m', 'initial')::text) = 34;",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT COUNT(*) FROM dolt_status;",
					Expected: []sql.Row{{13}},
				},
				{
					Query:    "INSERT INTO t_simple VALUES (4);",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO t_composite VALUES (3, 100);",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO t_array VALUES (ARRAY['stu']);",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT COUNT(*) FROM dolt_status;",
					Expected: []sql.Row{{16}},
				},
				{
					Query:    "SELECT length(DOLT_COMMIT('-a', '-m', 'initial')::text) = 34;",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT COUNT(*) FROM dolt_status;",
					Expected: []sql.Row{{13}},
				},
			},
		},
		{
			Name: "Allow and skip empty",
			SetUpScript: []string{
				"CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
				"INSERT INTO t_simple VALUES (1), (2), (3);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 34;",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:       "SELECT DOLT_COMMIT('-m', 'should_error');",
					ExpectedErr: "nothing to commit",
				},
				{
					Query:    "SELECT DOLT_COMMIT('--skip-empty', '-m', 'should_error');",
					Expected: []sql.Row{{nil}},
				},
				{
					Query:    "SELECT length(DOLT_COMMIT('--allow-empty', '-m', 'initial')::text) = 34;",
					Expected: []sql.Row{{"t"}},
				},
			},
		},
	})
}

func TestDoltConflictsResolve(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Resolve all using --ours",
			Skip: true, // TODO: attempting a merge with generated columns causes a panic
			SetUpScript: []string{
				"CREATE TABLE t_simple (pk INT4 PRIMARY KEY, v1 INT4);",
				"CREATE TABLE t_composite (pk1 INT2, pk2 INT8, v1 INT4, PRIMARY KEY(pk1, pk2));",
				"CREATE TABLE t_array (pk TEXT[] PRIMARY KEY, v1 INT4);",
				"CREATE TABLE t_serial (pk SERIAL PRIMARY KEY, v1 INT4);",
				"CREATE TABLE t_generated (pk INT8 PRIMARY KEY, v1 INT4, v2 INT8 GENERATED ALWAYS AS (pk * 1000) STORED);",
				"CREATE TABLE t_trigger (pk INT4 PRIMARY KEY, v1 INT8);",
				"CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
				"CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
				"CREATE TABLE t_default_func (pk INT4 PRIMARY KEY, v1 INT8 DEFAULT f_default(), v2 INT8);",
				"INSERT INTO t_simple VALUES (1, 1);",
				"INSERT INTO t_composite VALUES (1, 1, 1);",
				"INSERT INTO t_array VALUES (ARRAY['abc'], 1);",
				"INSERT INTO t_serial VALUES (DEFAULT, 1);",
				"INSERT INTO t_generated (pk, v1) VALUES (1, 1);",
				"INSERT INTO t_trigger VALUES (1, 1);",
				"INSERT INTO t_default_func (pk, v2) VALUES (1, 1);",
				"SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 34;",
				"SELECT DOLT_BRANCH('other');",
				"INSERT INTO t_simple VALUES (2, 2);",
				"INSERT INTO t_composite VALUES (2, 2, 2);",
				"INSERT INTO t_array VALUES (ARRAY['def'], 2);",
				"INSERT INTO t_serial VALUES (DEFAULT, 2);",
				"INSERT INTO t_generated (pk, v1) VALUES (2, 2);",
				"CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 33; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_trigger VALUES (2, 2);",
				"CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 34; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_default_func (pk, v2) VALUES (2, 2);",
				"SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 34;",
				"SELECT DOLT_CHECKOUT('other');",
				"INSERT INTO t_simple VALUES (2, 3);",
				"INSERT INTO t_composite VALUES (2, 2, 3);",
				"INSERT INTO t_array VALUES (ARRAY['def'], 3);",
				"INSERT INTO t_serial VALUES (DEFAULT, 3);",
				"INSERT INTO t_generated (pk, v1) VALUES (2, 3);",
				"CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 34; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_trigger VALUES (2, 3);",
				"CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 35; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_default_func (pk, v2) VALUES (2, 3);",
				"SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 34;",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT DOLT_MERGE('main');",
					Expected: []sql.Row{{}},
				},
				// TODO: finish adding these
			},
		},
		{
			Name: "Resolve all using --theirs",
			Skip: true, // TODO: attempting a merge with generated columns causes a panic
			SetUpScript: []string{
				"CREATE TABLE t_simple (pk INT4 PRIMARY KEY, v1 INT4);",
				"CREATE TABLE t_composite (pk1 INT2, pk2 INT8, v1 INT4, PRIMARY KEY(pk1, pk2));",
				"CREATE TABLE t_array (pk TEXT[] PRIMARY KEY, v1 INT4);",
				"CREATE TABLE t_serial (pk SERIAL PRIMARY KEY, v1 INT4);",
				"CREATE TABLE t_generated (pk INT8 PRIMARY KEY, v1 INT4, v2 INT8 GENERATED ALWAYS AS (pk * 1000) STORED);",
				"CREATE TABLE t_trigger (pk INT4 PRIMARY KEY, v1 INT8);",
				"CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
				"CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
				"CREATE TABLE t_default_func (pk INT4 PRIMARY KEY, v1 INT8 DEFAULT f_default(), v2 INT8);",
				"INSERT INTO t_simple VALUES (1, 1);",
				"INSERT INTO t_composite VALUES (1, 1, 1);",
				"INSERT INTO t_array VALUES (ARRAY['abc'], 1);",
				"INSERT INTO t_serial VALUES (DEFAULT, 1);",
				"INSERT INTO t_generated (pk, v1) VALUES (1, 1);",
				"INSERT INTO t_trigger VALUES (1, 1);",
				"INSERT INTO t_default_func (pk, v2) VALUES (1, 1);",
				"SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 34;",
				"SELECT DOLT_BRANCH('other');",
				"INSERT INTO t_simple VALUES (2, 2);",
				"INSERT INTO t_composite VALUES (2, 2, 2);",
				"INSERT INTO t_array VALUES (ARRAY['def'], 2);",
				"INSERT INTO t_serial VALUES (DEFAULT, 2);",
				"INSERT INTO t_generated (pk, v1) VALUES (2, 2);",
				"CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 33; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_trigger VALUES (2, 2);",
				"CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 34; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_default_func (pk, v2) VALUES (2, 2);",
				"SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 34;",
				"SELECT DOLT_CHECKOUT('other');",
				"INSERT INTO t_simple VALUES (2, 3);",
				"INSERT INTO t_composite VALUES (2, 2, 3);",
				"INSERT INTO t_array VALUES (ARRAY['def'], 3);",
				"INSERT INTO t_serial VALUES (DEFAULT, 3);",
				"INSERT INTO t_generated (pk, v1) VALUES (2, 3);",
				"CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 34; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_trigger VALUES (2, 3);",
				"CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 35; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_default_func (pk, v2) VALUES (2, 3);",
				"SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 34;",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT DOLT_MERGE('main');",
					Expected: []sql.Row{{}},
				},
				// TODO: finish adding these
			},
		},
		{
			Name: "Resolve individual items",
			Skip: true, // TODO: attempting a merge with generated columns causes a panic
			SetUpScript: []string{
				"CREATE TABLE t_simple (pk INT4 PRIMARY KEY, v1 INT4);",
				"CREATE TABLE t_composite (pk1 INT2, pk2 INT8, v1 INT4, PRIMARY KEY(pk1, pk2));",
				"CREATE TABLE t_array (pk TEXT[] PRIMARY KEY, v1 INT4);",
				"CREATE TABLE t_serial (pk SERIAL PRIMARY KEY, v1 INT4);",
				"CREATE TABLE t_generated (pk INT8 PRIMARY KEY, v1 INT4, v2 INT8 GENERATED ALWAYS AS (pk * 1000) STORED);",
				"CREATE TABLE t_trigger (pk INT4 PRIMARY KEY, v1 INT8);",
				"CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
				"CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
				"CREATE TABLE t_default_func (pk INT4 PRIMARY KEY, v1 INT8 DEFAULT f_default(), v2 INT8);",
				"INSERT INTO t_simple VALUES (1, 1);",
				"INSERT INTO t_composite VALUES (1, 1, 1);",
				"INSERT INTO t_array VALUES (ARRAY['abc'], 1);",
				"INSERT INTO t_serial VALUES (DEFAULT, 1);",
				"INSERT INTO t_generated (pk, v1) VALUES (1, 1);",
				"INSERT INTO t_trigger VALUES (1, 1);",
				"INSERT INTO t_default_func (pk, v2) VALUES (1, 1);",
				"SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 34;",
				"SELECT DOLT_BRANCH('other');",
				"INSERT INTO t_simple VALUES (2, 2);",
				"INSERT INTO t_composite VALUES (2, 2, 2);",
				"INSERT INTO t_array VALUES (ARRAY['def'], 2);",
				"INSERT INTO t_serial VALUES (DEFAULT, 2);",
				"INSERT INTO t_generated (pk, v1) VALUES (2, 2);",
				"CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 33; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_trigger VALUES (2, 2);",
				"CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 34; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_default_func (pk, v2) VALUES (2, 2);",
				"SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 34;",
				"SELECT DOLT_CHECKOUT('other');",
				"INSERT INTO t_simple VALUES (2, 3);",
				"INSERT INTO t_composite VALUES (2, 2, 3);",
				"INSERT INTO t_array VALUES (ARRAY['def'], 3);",
				"INSERT INTO t_serial VALUES (DEFAULT, 3);",
				"INSERT INTO t_generated (pk, v1) VALUES (2, 3);",
				"CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 34; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_trigger VALUES (2, 3);",
				"CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 35; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_default_func (pk, v2) VALUES (2, 3);",
				"SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 34;",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT DOLT_MERGE('main');",
					Expected: []sql.Row{{16}},
				},
				// TODO: finish adding these
			},
		},
	})
}

func TestDoltDiff(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Single commit",
			SetUpScript: []string{
				"CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
				"CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
				"CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
				"CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
				"CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
				"CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
				"CREATE TABLE t_unique (v1 INT4 UNIQUE);",
				"CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
				"CREATE TABLE t_trigger (v1 INT8);",
				"CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
				"CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
				"CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
				"INSERT INTO t_simple VALUES (1), (2), (3);",
				"INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
				"INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
				"INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
				"INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
				"INSERT INTO t_checked VALUES (1), (50), (100);",
				"INSERT INTO t_fk_parent VALUES (10), (20), (30);",
				"INSERT INTO t_fk_child VALUES (10), (20);",
				"INSERT INTO t_unique VALUES (7), (8), (9);",
				"INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
				"INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
				"INSERT INTO t_default_func (v2) VALUES (1), (2);",
				"SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 34;",
				"SELECT DOLT_BRANCH('original');",
				"INSERT INTO t_simple VALUES (4);",
				"INSERT INTO t_composite VALUES (3, 100);",
				"INSERT INTO t_array VALUES (ARRAY['stu']);",
				"INSERT INTO t_serial VALUES (DEFAULT);",
				"INSERT INTO t_default_simple VALUES (98, 8);",
				"INSERT INTO t_checked VALUES (99);",
				"INSERT INTO t_fk_parent VALUES (40);",
				"INSERT INTO t_fk_child VALUES (30);",
				"INSERT INTO t_unique VALUES (10);",
				"INSERT INTO t_generated (v1) VALUES (11);",
				"CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 33; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_trigger (v1) VALUES (2);",
				"CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 34; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_default_func (v2) VALUES (3);",
				"SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 34;",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT from_pk FROM DOLT_DIFF('main', 'original', 't_simple');",
					Expected: []sql.Row{{4}},
				},
				{
					Query:    "SELECT from_pk1, from_pk2 FROM DOLT_DIFF('main', 'original', 't_composite');",
					Expected: []sql.Row{{3, 100}},
				},
				{
					Query:    "SELECT from_pk FROM DOLT_DIFF('main', 'original', 't_array');",
					Expected: []sql.Row{{`{stu}`}},
				},
				{
					Query:    "SELECT from_pk FROM DOLT_DIFF('main', 'original', 't_serial');",
					Expected: []sql.Row{{4}},
				},
				{
					Query:    "SELECT from_v1, from_v2 FROM DOLT_DIFF('main', 'original', 't_default_simple');",
					Expected: []sql.Row{{98, 8}},
				},
				{
					Query:    "SELECT from_v1 FROM DOLT_DIFF('main', 'original', 't_checked');",
					Expected: []sql.Row{{Numeric("99")}},
				},
				{
					Query:    "SELECT from_pk FROM DOLT_DIFF('main', 'original', 't_fk_parent');",
					Expected: []sql.Row{{40}},
				},
				{
					Query:    "SELECT from_pk FROM DOLT_DIFF('main', 'original', 't_fk_child');",
					Expected: []sql.Row{{30}},
				},
				{
					Query:    "SELECT from_v1 FROM DOLT_DIFF('main', 'original', 't_unique');",
					Expected: []sql.Row{{10}},
				},
				{
					Query:    "SELECT from_v1, from_v2 FROM DOLT_DIFF('main', 'original', 't_generated');",
					Expected: []sql.Row{{11, 11000}},
				},
				{
					Query:    "SELECT from_v1 FROM DOLT_DIFF('main', 'original', 't_trigger');",
					Expected: []sql.Row{{66}},
				},
				{
					Query:    "SELECT from_v1, from_v2 FROM DOLT_DIFF('main', 'original', 't_default_func');",
					Expected: []sql.Row{{34, 3}},
				},
				{
					Query:    "SELECT * FROM DOLT_DIFF('main', 'original', 'f_trigger()');",
					Skip:     true, // TODO: need to implement for functions
					Expected: []sql.Row{{}},
				},
				{
					Query:    "SELECT * FROM DOLT_DIFF('main', 'original', 'f_default()');",
					Skip:     true, // TODO: need to implement for functions
					Expected: []sql.Row{{}},
				},
				{
					Query:    "SELECT * FROM DOLT_DIFF('main', 'original', 't_trigger.trig_trigger');",
					Skip:     true, // TODO: need to implement for triggers
					Expected: []sql.Row{{}},
				},
				{
					Query:    "SELECT * FROM DOLT_DIFF('main', 'original', 't_serial_pk_seq');",
					Skip:     true, // TODO: need to implement for sequences
					Expected: []sql.Row{{}},
				},
			},
		},
	})
}

func TestDoltDiffStat(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Single commit",
			SetUpScript: []string{
				"CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
				"CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
				"CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
				"CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
				"CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
				"CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
				"CREATE TABLE t_unique (v1 INT4 UNIQUE);",
				"CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
				"CREATE TABLE t_trigger (v1 INT8);",
				"CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
				"CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
				"CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
				"INSERT INTO t_simple VALUES (1), (2), (3);",
				"INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
				"INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
				"INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
				"INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
				"INSERT INTO t_checked VALUES (1), (50), (100);",
				"INSERT INTO t_fk_parent VALUES (10), (20), (30);",
				"INSERT INTO t_fk_child VALUES (10), (20);",
				"INSERT INTO t_unique VALUES (7), (8), (9);",
				"INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
				"INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
				"INSERT INTO t_default_func (v2) VALUES (1), (2);",
				"SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 34;",
				"SELECT DOLT_BRANCH('original');",
				"INSERT INTO t_simple VALUES (4);",
				"INSERT INTO t_composite VALUES (3, 100);",
				"INSERT INTO t_array VALUES (ARRAY['stu']);",
				"INSERT INTO t_serial VALUES (DEFAULT);",
				"INSERT INTO t_default_simple VALUES (98, 8);",
				"INSERT INTO t_checked VALUES (99);",
				"INSERT INTO t_fk_parent VALUES (40);",
				"INSERT INTO t_fk_child VALUES (30);",
				"INSERT INTO t_unique VALUES (10);",
				"INSERT INTO t_generated (v1) VALUES (11);",
				"CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 33; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_trigger (v1) VALUES (2);",
				"CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 34; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_default_func (v2) VALUES (3);",
				"SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 34;",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT COUNT(*) FROM DOLT_DIFF_STAT('main', 'original');",
					Expected: []sql.Row{{15}},
				},
				{
					Query:    "SELECT table_name FROM DOLT_DIFF_STAT('main', 'original', 't_simple');",
					Expected: []sql.Row{{"public.t_simple"}},
				},
				{
					Query:    "SELECT table_name FROM DOLT_DIFF_STAT('main', 'original', 't_composite');",
					Expected: []sql.Row{{"public.t_composite"}},
				},
				{
					Query:    "SELECT table_name FROM DOLT_DIFF_STAT('main', 'original', 't_array');",
					Expected: []sql.Row{{"public.t_array"}},
				},
				{
					Query:    "SELECT table_name FROM DOLT_DIFF_STAT('main', 'original', 't_serial');",
					Expected: []sql.Row{{"public.t_serial"}},
				},
				{
					Query:    "SELECT table_name FROM DOLT_DIFF_STAT('main', 'original', 't_default_simple');",
					Expected: []sql.Row{{"public.t_default_simple"}},
				},
				{
					Query:    "SELECT table_name FROM DOLT_DIFF_STAT('main', 'original', 't_checked');",
					Expected: []sql.Row{{"public.t_checked"}},
				},
				{
					Query:    "SELECT table_name FROM DOLT_DIFF_STAT('main', 'original', 't_fk_parent');",
					Expected: []sql.Row{{"public.t_fk_parent"}},
				},
				{
					Query:    "SELECT table_name FROM DOLT_DIFF_STAT('main', 'original', 't_fk_child');",
					Expected: []sql.Row{{"public.t_fk_child"}},
				},
				{
					Query:    "SELECT table_name FROM DOLT_DIFF_STAT('main', 'original', 't_unique');",
					Expected: []sql.Row{{"public.t_unique"}},
				},
				{
					Query:    "SELECT table_name FROM DOLT_DIFF_STAT('main', 'original', 't_generated');",
					Expected: []sql.Row{{"public.t_generated"}},
				},
				{
					Query:    "SELECT table_name FROM DOLT_DIFF_STAT('main', 'original', 't_trigger');",
					Expected: []sql.Row{{"public.t_trigger"}},
				},
				{
					Query:    "SELECT table_name FROM DOLT_DIFF_STAT('main', 'original', 't_default_func');",
					Expected: []sql.Row{{"public.t_default_func"}},
				},
				{
					Query:    "SELECT table_name FROM DOLT_DIFF_STAT('main', 'original', 'f_trigger()');",
					Expected: []sql.Row{{"public.f_trigger()"}},
				},
				{
					Query:    "SELECT table_name FROM DOLT_DIFF_STAT('main', 'original', 'f_default()');",
					Expected: []sql.Row{{"public.f_default()"}},
				},
				{
					Query:    "SELECT table_name FROM DOLT_DIFF_STAT('main', 'original', 't_serial_pk_seq');",
					Expected: []sql.Row{{"public.t_serial_pk_seq"}},
				},
			},
		},
	})
}

func TestDoltDiffSummary(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Single commit",
			SetUpScript: []string{
				"CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
				"CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
				"CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
				"CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
				"CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
				"CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
				"CREATE TABLE t_unique (v1 INT4 UNIQUE);",
				"CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
				"CREATE TABLE t_trigger (v1 INT8);",
				"CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
				"CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
				"CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
				"INSERT INTO t_simple VALUES (1), (2), (3);",
				"INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
				"INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
				"INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
				"INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
				"INSERT INTO t_checked VALUES (1), (50), (100);",
				"INSERT INTO t_fk_parent VALUES (10), (20), (30);",
				"INSERT INTO t_fk_child VALUES (10), (20);",
				"INSERT INTO t_unique VALUES (7), (8), (9);",
				"INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
				"INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
				"INSERT INTO t_default_func (v2) VALUES (1), (2);",
				"SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 34;",
				"SELECT DOLT_BRANCH('original');",
				"INSERT INTO t_simple VALUES (4);",
				"INSERT INTO t_composite VALUES (3, 100);",
				"INSERT INTO t_array VALUES (ARRAY['stu']);",
				"INSERT INTO t_serial VALUES (DEFAULT);",
				"INSERT INTO t_default_simple VALUES (98, 8);",
				"INSERT INTO t_checked VALUES (99);",
				"INSERT INTO t_fk_parent VALUES (40);",
				"INSERT INTO t_fk_child VALUES (30);",
				"INSERT INTO t_unique VALUES (10);",
				"INSERT INTO t_generated (v1) VALUES (11);",
				"CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 33; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_trigger (v1) VALUES (2);",
				"CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 34; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_default_func (v2) VALUES (3);",
				"SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 34;",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT COUNT(*) FROM DOLT_DIFF_SUMMARY('main', 'original');",
					Expected: []sql.Row{{15}},
				},
				{
					Query:    "SELECT from_table_name FROM DOLT_DIFF_SUMMARY('main', 'original', 't_simple');",
					Expected: []sql.Row{{"public.t_simple"}},
				},
				{
					Query:    "SELECT from_table_name FROM DOLT_DIFF_SUMMARY('main', 'original', 't_composite');",
					Expected: []sql.Row{{"public.t_composite"}},
				},
				{
					Query:    "SELECT from_table_name FROM DOLT_DIFF_SUMMARY('main', 'original', 't_array');",
					Expected: []sql.Row{{"public.t_array"}},
				},
				{
					Query:    "SELECT from_table_name FROM DOLT_DIFF_SUMMARY('main', 'original', 't_serial');",
					Expected: []sql.Row{{"public.t_serial"}},
				},
				{
					Query:    "SELECT from_table_name FROM DOLT_DIFF_SUMMARY('main', 'original', 't_default_simple');",
					Expected: []sql.Row{{"public.t_default_simple"}},
				},
				{
					Query:    "SELECT from_table_name FROM DOLT_DIFF_SUMMARY('main', 'original', 't_checked');",
					Expected: []sql.Row{{"public.t_checked"}},
				},
				{
					Query:    "SELECT from_table_name FROM DOLT_DIFF_SUMMARY('main', 'original', 't_fk_parent');",
					Expected: []sql.Row{{"public.t_fk_parent"}},
				},
				{
					Query:    "SELECT from_table_name FROM DOLT_DIFF_SUMMARY('main', 'original', 't_fk_child');",
					Expected: []sql.Row{{"public.t_fk_child"}},
				},
				{
					Query:    "SELECT from_table_name FROM DOLT_DIFF_SUMMARY('main', 'original', 't_unique');",
					Expected: []sql.Row{{"public.t_unique"}},
				},
				{
					Query:    "SELECT from_table_name FROM DOLT_DIFF_SUMMARY('main', 'original', 't_generated');",
					Expected: []sql.Row{{"public.t_generated"}},
				},
				{
					Query:    "SELECT from_table_name FROM DOLT_DIFF_SUMMARY('main', 'original', 't_trigger');",
					Expected: []sql.Row{{"public.t_trigger"}},
				},
				{
					Query:    "SELECT from_table_name FROM DOLT_DIFF_SUMMARY('main', 'original', 't_default_func');",
					Expected: []sql.Row{{"public.t_default_func"}},
				},
				{
					Query:    "SELECT from_table_name FROM DOLT_DIFF_SUMMARY('main', 'original', 'f_trigger()');",
					Expected: []sql.Row{{"public.f_trigger()"}},
				},
				{
					Query:    "SELECT from_table_name FROM DOLT_DIFF_SUMMARY('main', 'original', 'f_default()');",
					Expected: []sql.Row{{"public.f_default()"}},
				},
				{
					Query:    "SELECT from_table_name FROM DOLT_DIFF_SUMMARY('main', 'original', 't_serial_pk_seq');",
					Expected: []sql.Row{{"public.t_serial_pk_seq"}},
				},
			},
		},
	})
}

func TestDoltGC(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Full",
			SetUpScript: []string{
				"CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
				"CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
				"CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
				"CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
				"CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
				"CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
				"CREATE TABLE t_unique (v1 INT4 UNIQUE);",
				"CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
				"CREATE TABLE t_trigger (v1 INT8);",
				"CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
				"CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
				"CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
				"INSERT INTO t_simple VALUES (1), (2), (3);",
				"INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
				"INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
				"INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
				"INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
				"INSERT INTO t_checked VALUES (1), (50), (100);",
				"INSERT INTO t_fk_parent VALUES (10), (20), (30);",
				"INSERT INTO t_fk_child VALUES (10), (20);",
				"INSERT INTO t_unique VALUES (7), (8), (9);",
				"INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
				"INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
				"INSERT INTO t_default_func (v2) VALUES (1), (2);",
				"SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 34;",
				"SELECT DOLT_BRANCH('original');",
				"INSERT INTO t_simple VALUES (4);",
				"INSERT INTO t_composite VALUES (3, 100);",
				"INSERT INTO t_array VALUES (ARRAY['stu']);",
				"INSERT INTO t_serial VALUES (DEFAULT);",
				"INSERT INTO t_default_simple VALUES (98, 8);",
				"INSERT INTO t_checked VALUES (99);",
				"INSERT INTO t_fk_parent VALUES (40);",
				"INSERT INTO t_fk_child VALUES (30);",
				"INSERT INTO t_unique VALUES (10);",
				"INSERT INTO t_generated (v1) VALUES (11);",
				"CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 33; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_trigger (v1) VALUES (2);",
				"CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 34; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_default_func (v2) VALUES (3);",
				"SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 34;",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT DOLT_GC();",
					Expected: []sql.Row{{"{0}"}},
				},
			},
		},
		{
			Name: "Shallow",
			SetUpScript: []string{
				"CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
				"CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
				"CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
				"CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
				"CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
				"CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
				"CREATE TABLE t_unique (v1 INT4 UNIQUE);",
				"CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
				"CREATE TABLE t_trigger (v1 INT8);",
				"CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
				"CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
				"CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
				"INSERT INTO t_simple VALUES (1), (2), (3);",
				"INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
				"INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
				"INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
				"INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
				"INSERT INTO t_checked VALUES (1), (50), (100);",
				"INSERT INTO t_fk_parent VALUES (10), (20), (30);",
				"INSERT INTO t_fk_child VALUES (10), (20);",
				"INSERT INTO t_unique VALUES (7), (8), (9);",
				"INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
				"INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
				"INSERT INTO t_default_func (v2) VALUES (1), (2);",
				"SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 34;",
				"SELECT DOLT_BRANCH('original');",
				"INSERT INTO t_simple VALUES (4);",
				"INSERT INTO t_composite VALUES (3, 100);",
				"INSERT INTO t_array VALUES (ARRAY['stu']);",
				"INSERT INTO t_serial VALUES (DEFAULT);",
				"INSERT INTO t_default_simple VALUES (98, 8);",
				"INSERT INTO t_checked VALUES (99);",
				"INSERT INTO t_fk_parent VALUES (40);",
				"INSERT INTO t_fk_child VALUES (30);",
				"INSERT INTO t_unique VALUES (10);",
				"INSERT INTO t_generated (v1) VALUES (11);",
				"CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 33; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_trigger (v1) VALUES (2);",
				"CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 34; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_default_func (v2) VALUES (3);",
				"SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 34;",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT DOLT_GC('--shallow');",
					Expected: []sql.Row{{"{0}"}},
				},
			},
		},
	})
}

func TestDoltLog(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Smoke test",
			SetUpScript: []string{
				"CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
				"INSERT INTO t_simple VALUES (1), (2), (3);",
				"SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 34;",
				"SELECT DOLT_BRANCH('original');",
				"INSERT INTO t_simple VALUES (4);",
				"SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 34;",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT COUNT(*) FROM DOLT_LOG('main');",
					Expected: []sql.Row{{4}},
				},
				{
					Query:    "SELECT COUNT(*) FROM DOLT_LOG('original');",
					Expected: []sql.Row{{3}},
				},
			},
		},
	})
}

func TestDoltMerge(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Merge without conflicts",
			SetUpScript: []string{
				"CREATE TABLE t_simple (pk INT4 PRIMARY KEY, v1 INT4);",
				"CREATE TABLE t_composite (pk1 INT2, pk2 INT8, v1 INT4, PRIMARY KEY(pk1, pk2));",
				"CREATE TABLE t_array (pk TEXT[] PRIMARY KEY, v1 INT4);",
				"CREATE TABLE t_serial (pk SERIAL PRIMARY KEY, v1 INT4);",
				"CREATE TABLE t_generated (pk INT8 PRIMARY KEY, v1 INT4, v2 INT8 GENERATED ALWAYS AS (pk * 1000) STORED);",
				"CREATE TABLE t_trigger (pk INT4 PRIMARY KEY, v1 INT8);",
				"CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
				"CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
				"CREATE TABLE t_default_func (pk INT4 PRIMARY KEY, v1 INT8 DEFAULT f_default(), v2 INT8);",
				"INSERT INTO t_simple VALUES (1, 1);",
				"INSERT INTO t_composite VALUES (1, 1, 1);",
				"INSERT INTO t_array VALUES (ARRAY['abc'], 1);",
				"INSERT INTO t_serial VALUES (DEFAULT, 1);",
				"INSERT INTO t_generated (pk, v1) VALUES (1, 1);",
				"INSERT INTO t_trigger VALUES (1, 1);",
				"INSERT INTO t_default_func (pk, v2) VALUES (1, 1);",
				"SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 34;",
				"SELECT DOLT_BRANCH('other');",
				"INSERT INTO t_simple VALUES (2, 2);",
				"INSERT INTO t_composite VALUES (2, 2, 2);",
				"INSERT INTO t_array VALUES (ARRAY['def'], 2);",
				"INSERT INTO t_serial VALUES (DEFAULT, 2);",
				"INSERT INTO t_generated (pk, v1) VALUES (2, 2);",
				"CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 33; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_trigger VALUES (2, 2);",
				"CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 34; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_default_func (pk, v2) VALUES (2, 2);",
				"SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 34;",
				"SELECT DOLT_CHECKOUT('other');",
				"INSERT INTO t_simple VALUES (3, 3);",
				"INSERT INTO t_composite VALUES (3, 2, 3);",
				"INSERT INTO t_array VALUES (ARRAY['dfe'], 3);",
				"INSERT INTO t_serial VALUES (4, 3);",
				"INSERT INTO t_generated (pk, v1) VALUES (3, 3);",
				"INSERT INTO t_trigger VALUES (3, 3);",
				"CREATE OR REPLACE FUNCTION f_default() RETURNS INT4 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_default_func (pk, v2) VALUES (3, 3);",
				"SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 34;",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT strpos(DOLT_MERGE('main')::text, '0,0,\"merge successful\"') > 1;",
					Expected: []sql.Row{{"t"}},
				},
			},
		},
		{
			Name: "Merge with conflicts",
			Skip: true, // TODO: attempting a merge with generated columns causes a panic
			SetUpScript: []string{
				"CREATE TABLE t_simple (pk INT4 PRIMARY KEY, v1 INT4);",
				"CREATE TABLE t_composite (pk1 INT2, pk2 INT8, v1 INT4, PRIMARY KEY(pk1, pk2));",
				"CREATE TABLE t_array (pk TEXT[] PRIMARY KEY, v1 INT4);",
				"CREATE TABLE t_serial (pk SERIAL PRIMARY KEY, v1 INT4);",
				"CREATE TABLE t_generated (pk INT8 PRIMARY KEY, v1 INT4, v2 INT8 GENERATED ALWAYS AS (pk * 1000) STORED);",
				"CREATE TABLE t_trigger (pk INT4 PRIMARY KEY, v1 INT8);",
				"CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
				"CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
				"CREATE TABLE t_default_func (pk INT4 PRIMARY KEY, v1 INT8 DEFAULT f_default(), v2 INT8);",
				"INSERT INTO t_simple VALUES (1, 1);",
				"INSERT INTO t_composite VALUES (1, 1, 1);",
				"INSERT INTO t_array VALUES (ARRAY['abc'], 1);",
				"INSERT INTO t_serial VALUES (DEFAULT, 1);",
				"INSERT INTO t_generated (pk, v1) VALUES (1, 1);",
				"INSERT INTO t_trigger VALUES (1, 1);",
				"INSERT INTO t_default_func (pk, v2) VALUES (1, 1);",
				"SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 34;",
				"SELECT DOLT_BRANCH('other');",
				"INSERT INTO t_simple VALUES (2, 2);",
				"INSERT INTO t_composite VALUES (2, 2, 2);",
				"INSERT INTO t_array VALUES (ARRAY['def'], 2);",
				"INSERT INTO t_serial VALUES (DEFAULT, 2);",
				"INSERT INTO t_generated (pk, v1) VALUES (2, 2);",
				"CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 33; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_trigger VALUES (2, 2);",
				"CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 34; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_default_func (pk, v2) VALUES (2, 2);",
				"SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 34;",
				"SELECT DOLT_CHECKOUT('other');",
				"INSERT INTO t_simple VALUES (2, 3);",
				"INSERT INTO t_composite VALUES (2, 2, 3);",
				"INSERT INTO t_array VALUES (ARRAY['def'], 3);",
				"INSERT INTO t_serial VALUES (DEFAULT, 3);",
				"INSERT INTO t_generated (pk, v1) VALUES (2, 3);",
				"CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 34; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_trigger VALUES (2, 3);",
				"CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 35; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_default_func (pk, v2) VALUES (2, 3);",
				"SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 34;",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT DOLT_MERGE('main');",
					Expected: []sql.Row{{}},
				},
			},
		},
		{
			Name: "Fast forward",
			SetUpScript: []string{
				"CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
				"CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
				"CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
				"CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
				"CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
				"CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
				"CREATE TABLE t_unique (v1 INT4 UNIQUE);",
				"CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
				"CREATE TABLE t_trigger (v1 INT8);",
				"CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
				"CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
				"CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
				"INSERT INTO t_simple VALUES (1), (2), (3);",
				"INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
				"INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
				"INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
				"INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
				"INSERT INTO t_checked VALUES (1), (50), (100);",
				"INSERT INTO t_fk_parent VALUES (10), (20), (30);",
				"INSERT INTO t_fk_child VALUES (10), (20);",
				"INSERT INTO t_unique VALUES (7), (8), (9);",
				"INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
				"INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
				"INSERT INTO t_default_func (v2) VALUES (1), (2);",
				"SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 34;",
				"SELECT DOLT_BRANCH('original');",
				"INSERT INTO t_simple VALUES (4);",
				"INSERT INTO t_composite VALUES (3, 100);",
				"INSERT INTO t_array VALUES (ARRAY['stu']);",
				"INSERT INTO t_serial VALUES (DEFAULT);",
				"INSERT INTO t_default_simple VALUES (98, 8);",
				"INSERT INTO t_checked VALUES (99);",
				"INSERT INTO t_fk_parent VALUES (40);",
				"INSERT INTO t_fk_child VALUES (30);",
				"INSERT INTO t_unique VALUES (10);",
				"INSERT INTO t_generated (v1) VALUES (11);",
				"CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 33; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_trigger (v1) VALUES (2);",
				"CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 34; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_default_func (v2) VALUES (3);",
				"SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 34;",
				"SELECT DOLT_CHECKOUT('original');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT strpos(DOLT_MERGE('main')::text, '1,0,\"merge successful\"') > 1;",
					Expected: []sql.Row{{"t"}},
				},
			},
		},
		{
			Name: "--no-ff",
			SetUpScript: []string{
				"CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
				"CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
				"CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
				"CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
				"CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
				"CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
				"CREATE TABLE t_unique (v1 INT4 UNIQUE);",
				"CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
				"CREATE TABLE t_trigger (v1 INT8);",
				"CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
				"CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
				"CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
				"INSERT INTO t_simple VALUES (1), (2), (3);",
				"INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
				"INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
				"INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
				"INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
				"INSERT INTO t_checked VALUES (1), (50), (100);",
				"INSERT INTO t_fk_parent VALUES (10), (20), (30);",
				"INSERT INTO t_fk_child VALUES (10), (20);",
				"INSERT INTO t_unique VALUES (7), (8), (9);",
				"INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
				"INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
				"INSERT INTO t_default_func (v2) VALUES (1), (2);",
				"SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 34;",
				"SELECT DOLT_BRANCH('original');",
				"INSERT INTO t_simple VALUES (4);",
				"INSERT INTO t_composite VALUES (3, 100);",
				"INSERT INTO t_array VALUES (ARRAY['stu']);",
				"INSERT INTO t_serial VALUES (DEFAULT);",
				"INSERT INTO t_default_simple VALUES (98, 8);",
				"INSERT INTO t_checked VALUES (99);",
				"INSERT INTO t_fk_parent VALUES (40);",
				"INSERT INTO t_fk_child VALUES (30);",
				"INSERT INTO t_unique VALUES (10);",
				"INSERT INTO t_generated (v1) VALUES (11);",
				"CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 33; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_trigger (v1) VALUES (2);",
				"CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 34; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_default_func (v2) VALUES (3);",
				"SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 34;",
				"SELECT DOLT_CHECKOUT('original');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT strpos(DOLT_MERGE('main', '--no-ff', '-m', 'merge_commit')::text, 'merge successful') > 1;",
					Expected: []sql.Row{{"t"}},
				},
			},
		},
	})
}

func TestDoltPreviewMergeConflicts(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Preview conflicts",
			SetUpScript: []string{
				"CREATE TABLE t_simple (pk INT4 PRIMARY KEY, v1 INT4);",
				"CREATE TABLE t_composite (pk1 INT2, pk2 INT8, v1 INT4, PRIMARY KEY(pk1, pk2));",
				"CREATE TABLE t_array (pk TEXT[] PRIMARY KEY, v1 INT4);",
				"CREATE TABLE t_serial (pk SERIAL PRIMARY KEY, v1 INT4);",
				"CREATE TABLE t_generated (pk INT8 PRIMARY KEY, v1 INT4, v2 INT8 GENERATED ALWAYS AS (pk * 1000) STORED);",
				"CREATE TABLE t_trigger (pk INT4 PRIMARY KEY, v1 INT8);",
				"CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
				"CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
				"CREATE TABLE t_default_func (pk INT4 PRIMARY KEY, v1 INT8 DEFAULT f_default(), v2 INT8);",
				"INSERT INTO t_simple VALUES (1, 1);",
				"INSERT INTO t_composite VALUES (1, 1, 1);",
				"INSERT INTO t_array VALUES (ARRAY['abc'], 1);",
				"INSERT INTO t_serial VALUES (DEFAULT, 1);",
				"INSERT INTO t_generated (pk, v1) VALUES (1, 1);",
				"INSERT INTO t_trigger VALUES (1, 1);",
				"INSERT INTO t_default_func (pk, v2) VALUES (1, 1);",
				"SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 34;",
				"SELECT DOLT_BRANCH('other');",
				"INSERT INTO t_simple VALUES (2, 2);",
				"INSERT INTO t_composite VALUES (2, 2, 2);",
				"INSERT INTO t_array VALUES (ARRAY['def'], 2);",
				"INSERT INTO t_serial VALUES (DEFAULT, 2);",
				"INSERT INTO t_generated (pk, v1) VALUES (2, 2);",
				"CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 33; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_trigger VALUES (2, 2);",
				"CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 34; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_default_func (pk, v2) VALUES (2, 2);",
				"SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 34;",
				"SELECT DOLT_CHECKOUT('other');",
				"INSERT INTO t_simple VALUES (2, 3);",
				"INSERT INTO t_composite VALUES (2, 2, 3);",
				"INSERT INTO t_array VALUES (ARRAY['def'], 3);",
				"INSERT INTO t_serial VALUES (DEFAULT, 3);",
				"INSERT INTO t_generated (pk, v1) VALUES (2, 3);",
				"CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 34; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_trigger VALUES (2, 3);",
				"CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 35; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_default_func (pk, v2) VALUES (2, 3);",
				"SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 34;",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT COUNT(*) FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'other', 't_simple');",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT COUNT(*) FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'other', 't_composite');",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT COUNT(*) FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'other', 't_array');",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT COUNT(*) FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'other', 't_serial');",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT COUNT(*) FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'other', 't_generated');",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT COUNT(*) FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'other', 't_trigger');",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT COUNT(*) FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'other', 't_default_func');",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT COUNT(*) FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'other', 'f_default()');",
					Skip:     true, // TODO: implement for functions
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT COUNT(*) FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'other', 'f_trigger()');",
					Skip:     true, // TODO: implement for functions
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT COUNT(*) FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'other', 't_serial_pk_seq');",
					Skip:     true, // TODO: implement for sequences
					Expected: []sql.Row{{1}},
				},
			},
		},
	})
}

func TestDoltPreviewMergeConflictsSummary(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Preview summary",
			Skip: true, // TODO: attempting to preview with interpreted functions causes a panic
			SetUpScript: []string{
				"CREATE TABLE t_simple (pk INT4 PRIMARY KEY, v1 INT4);",
				"CREATE TABLE t_composite (pk1 INT2, pk2 INT8, v1 INT4, PRIMARY KEY(pk1, pk2));",
				"CREATE TABLE t_array (pk TEXT[] PRIMARY KEY, v1 INT4);",
				"CREATE TABLE t_serial (pk SERIAL PRIMARY KEY, v1 INT4);",
				"CREATE TABLE t_generated (pk INT8 PRIMARY KEY, v1 INT4, v2 INT8 GENERATED ALWAYS AS (pk * 1000) STORED);",
				"CREATE TABLE t_trigger (pk INT4 PRIMARY KEY, v1 INT8);",
				"CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
				"CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
				"CREATE TABLE t_default_func (pk INT4 PRIMARY KEY, v1 INT8 DEFAULT f_default(), v2 INT8);",
				"INSERT INTO t_simple VALUES (1, 1);",
				"INSERT INTO t_composite VALUES (1, 1, 1);",
				"INSERT INTO t_array VALUES (ARRAY['abc'], 1);",
				"INSERT INTO t_serial VALUES (DEFAULT, 1);",
				"INSERT INTO t_generated (pk, v1) VALUES (1, 1);",
				"INSERT INTO t_trigger VALUES (1, 1);",
				"INSERT INTO t_default_func (pk, v2) VALUES (1, 1);",
				"SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 34;",
				"SELECT DOLT_BRANCH('other');",
				"INSERT INTO t_simple VALUES (2, 2);",
				"INSERT INTO t_composite VALUES (2, 2, 2);",
				"INSERT INTO t_array VALUES (ARRAY['def'], 2);",
				"INSERT INTO t_serial VALUES (DEFAULT, 2);",
				"INSERT INTO t_generated (pk, v1) VALUES (2, 2);",
				"CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 33; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_trigger VALUES (2, 2);",
				"CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 34; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_default_func (pk, v2) VALUES (2, 2);",
				"SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 34;",
				"SELECT DOLT_CHECKOUT('other');",
				"INSERT INTO t_simple VALUES (2, 3);",
				"INSERT INTO t_composite VALUES (2, 2, 3);",
				"INSERT INTO t_array VALUES (ARRAY['def'], 3);",
				"INSERT INTO t_serial VALUES (DEFAULT, 3);",
				"INSERT INTO t_generated (pk, v1) VALUES (2, 3);",
				"CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 34; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_trigger VALUES (2, 3);",
				"CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 35; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_default_func (pk, v2) VALUES (2, 3);",
				"SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 34;",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT COUNT(*) FROM DOLT_PREVIEW_MERGE_CONFLICTS_SUMMARY('main', 'other');",
					Expected: []sql.Row{{9}},
				},
			},
		},
	})
}

func TestDoltQueryDiff(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Smoke test",
			Skip: true, // TODO: AS OF seems to not be implemented yet
			SetUpScript: []string{
				"CREATE TABLE t_simple (pk INT4 PRIMARY KEY, v1 INT4);",
				"SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 34;",
				"SELECT DOLT_BRANCH('other');",
				"INSERT INTO t_simple VALUES (1, 1), (2, 1);",
				"SELECT length(DOLT_COMMIT('-A', '-m', 'next')::text) = 34;",
				"SELECT DOLT_CHECKOUT('other');",
				"INSERT INTO t_simple VALUES (1, 2), (2, 2);",
				"SELECT length(DOLT_COMMIT('-A', '-m', 'next')::text) = 34;",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT * FROM DOLT_QUERY_DIFF('SELECT * FROM t_simple AS OF main', 'SELECT * FROM t_simple AS OF other');",
					Expected: []sql.Row{{}},
				},
			},
		},
	})
}

func TestDoltReset(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Hard",
			SetUpScript: []string{
				"CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
				"CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
				"CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
				"CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
				"CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
				"CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
				"CREATE TABLE t_unique (v1 INT4 UNIQUE);",
				"CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
				"CREATE TABLE t_trigger (v1 INT8);",
				"CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
				"CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
				"CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
				"INSERT INTO t_simple VALUES (1), (2), (3);",
				"INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
				"INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
				"INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
				"INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
				"INSERT INTO t_checked VALUES (1), (50), (100);",
				"INSERT INTO t_fk_parent VALUES (10), (20), (30);",
				"INSERT INTO t_fk_child VALUES (10), (20);",
				"INSERT INTO t_unique VALUES (7), (8), (9);",
				"INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
				"INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
				"INSERT INTO t_default_func (v2) VALUES (1), (2);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT COUNT(*) FROM dolt_status;",
					Expected: []sql.Row{{16}},
				},
				{
					Query:    "SELECT DOLT_RESET('--hard', 'HEAD~1');",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    "SELECT * FROM dolt_status;",
					Skip:     true, // TODO: need to implement root object support
					Expected: []sql.Row{{0}},
				},
			},
		},
		{
			Name: "Soft",
			SetUpScript: []string{
				"CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
				"CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
				"CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
				"CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
				"CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
				"CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
				"CREATE TABLE t_unique (v1 INT4 UNIQUE);",
				"CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
				"CREATE TABLE t_trigger (v1 INT8);",
				"CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
				"CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
				"CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
				"INSERT INTO t_simple VALUES (1), (2), (3);",
				"INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
				"INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
				"INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
				"INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
				"INSERT INTO t_checked VALUES (1), (50), (100);",
				"INSERT INTO t_fk_parent VALUES (10), (20), (30);",
				"INSERT INTO t_fk_child VALUES (10), (20);",
				"INSERT INTO t_unique VALUES (7), (8), (9);",
				"INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
				"INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
				"INSERT INTO t_default_func (v2) VALUES (1), (2);",
				"SELECT DOLT_ADD('-A');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT COUNT(*) FROM dolt_status WHERE staged = 't';",
					Expected: []sql.Row{{16}},
				},
				{
					Query:    "SELECT DOLT_RESET('--soft', 'HEAD');",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    "SELECT COUNT(*) FROM dolt_status WHERE staged = 'f';",
					Expected: []sql.Row{{16}},
				},
			},
		},
	})
}

func TestDoltRevert(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Revert a single commit",
			SetUpScript: []string{
				"CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
				"CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
				"CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
				"CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
				"CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
				"CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
				"CREATE TABLE t_unique (v1 INT4 UNIQUE);",
				"CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
				"CREATE TABLE t_trigger (v1 INT8);",
				"CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
				"CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
				"CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
				"INSERT INTO t_simple VALUES (1), (2), (3);",
				"INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
				"INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
				"INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
				"INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
				"INSERT INTO t_checked VALUES (1), (50), (100);",
				"INSERT INTO t_fk_parent VALUES (10), (20), (30);",
				"INSERT INTO t_fk_child VALUES (10), (20);",
				"INSERT INTO t_unique VALUES (7), (8), (9);",
				"INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
				"INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
				"INSERT INTO t_default_func (v2) VALUES (1), (2);",
				"SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 34;",
				"INSERT INTO t_simple VALUES (4);",
				"INSERT INTO t_composite VALUES (3, 100);",
				"INSERT INTO t_array VALUES (ARRAY['stu']);",
				"INSERT INTO t_serial VALUES (DEFAULT);",
				"INSERT INTO t_default_simple VALUES (98, 8);",
				"INSERT INTO t_checked VALUES (99);",
				"INSERT INTO t_fk_parent VALUES (40);",
				"INSERT INTO t_fk_child VALUES (30);",
				"INSERT INTO t_unique VALUES (10);",
				"INSERT INTO t_generated (v1) VALUES (11);",
				"CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 33; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_trigger (v1) VALUES (2);",
				"CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 34; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_default_func (v2) VALUES (3);",
				"SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 34;",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM t_simple;`,
					Expected: []sql.Row{{1}, {2}, {3}, {4}},
				},
				{
					Query:    `SELECT f_default();`,
					Expected: []sql.Row{{34}},
				},
				{
					Query:    "SELECT DOLT_REVERT('HEAD');",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    `SELECT * FROM t_simple;`,
					Expected: []sql.Row{{1}, {2}, {3}},
				},
				{
					Query:    `SELECT f_default();`,
					Expected: []sql.Row{{33}},
				},
			},
		},
	})
}

func TestDoltRM(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Cached only",
			SetUpScript: []string{
				"CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
				"CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
				"CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
				"CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
				"CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
				"CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
				"CREATE TABLE t_unique (v1 INT4 UNIQUE);",
				"CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
				"CREATE TABLE t_trigger (v1 INT8);",
				"CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
				"CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
				"CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
				"INSERT INTO t_simple VALUES (1), (2), (3);",
				"INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
				"INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
				"INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
				"INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
				"INSERT INTO t_checked VALUES (1), (50), (100);",
				"INSERT INTO t_fk_parent VALUES (10), (20), (30);",
				"INSERT INTO t_fk_child VALUES (10), (20);",
				"INSERT INTO t_unique VALUES (7), (8), (9);",
				"INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
				"INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
				"INSERT INTO t_default_func (v2) VALUES (1), (2);",
				"SELECT DOLT_ADD('-A');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT COUNT(*) FROM dolt_status WHERE staged = 't';",
					Expected: []sql.Row{{16}},
				},
				{
					Query: "SELECT DOLT_RM('--cached', 't_simple','t_composite','t_array','t_serial','t_default_simple'," +
						"'t_checked','t_fk_parent','t_fk_child','t_unique','t_generated','t_trigger','t_default_func'," +
						"'f_trigger()','f_default()','t_serial_pk_seq','t_trigger.trig_trigger');",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    "SELECT COUNT(*) FROM dolt_status WHERE staged = 'f';",
					Expected: []sql.Row{{16}},
				},
			},
		},
		// TODO: figure out how the non-cached version is supposed to work, as the example doesn't even work
	})
}

func TestDoltSchemaDiff(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Single commit",
			SetUpScript: []string{
				"CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
				"CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
				"CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
				"CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
				"CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
				"CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
				"CREATE TABLE t_unique (v1 INT4 UNIQUE);",
				"CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
				"CREATE TABLE t_trigger (v1 INT8);",
				"CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
				"CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
				"CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
				"INSERT INTO t_simple VALUES (1), (2), (3);",
				"INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
				"INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
				"INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
				"INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
				"INSERT INTO t_checked VALUES (1), (50), (100);",
				"INSERT INTO t_fk_parent VALUES (10), (20), (30);",
				"INSERT INTO t_fk_child VALUES (10), (20);",
				"INSERT INTO t_unique VALUES (7), (8), (9);",
				"INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
				"INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
				"INSERT INTO t_default_func (v2) VALUES (1), (2);",
				"SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 34;",
				"SELECT DOLT_BRANCH('original');",
				"INSERT INTO t_simple VALUES (4);",
				"INSERT INTO t_composite VALUES (3, 100);",
				"INSERT INTO t_array VALUES (ARRAY['stu']);",
				"INSERT INTO t_serial VALUES (DEFAULT);",
				"INSERT INTO t_default_simple VALUES (98, 8);",
				"INSERT INTO t_checked VALUES (99);",
				"INSERT INTO t_fk_parent VALUES (40);",
				"INSERT INTO t_fk_child VALUES (30);",
				"INSERT INTO t_unique VALUES (10);",
				"INSERT INTO t_generated (v1) VALUES (11);",
				"CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 33; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_trigger (v1) VALUES (2);",
				"CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 34; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_default_func (v2) VALUES (3);",
				"ALTER TABLE t_simple RENAME COLUMN pk TO rcol;",
				"ALTER TABLE t_composite RENAME COLUMN pk1 TO rcol;",
				"ALTER TABLE t_array RENAME COLUMN pk TO rcol;",
				"ALTER TABLE t_serial RENAME COLUMN pk TO rcol;",
				"ALTER TABLE t_default_simple RENAME COLUMN v1 TO rcol;",
				"ALTER TABLE t_generated RENAME COLUMN v1 TO rcol;",
				"SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 34;",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT COUNT(*) FROM DOLT_SCHEMA_DIFF('main', 'original');",
					Expected: []sql.Row{{6}},
				},
			},
		},
	})
}

func TestDoltStash(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Push and pop all untracked",
			Skip: true, // TODO: errors with "dolt_procedures: unsupported type int", which seems like the Doltgres->Dolt layer is erroring
			SetUpScript: []string{
				"CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
				"CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
				"CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
				"CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
				"CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
				"CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
				"CREATE TABLE t_unique (v1 INT4 UNIQUE);",
				"CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
				"CREATE TABLE t_trigger (v1 INT8);",
				"CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
				"CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
				"CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
				"INSERT INTO t_simple VALUES (1), (2), (3);",
				"INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
				"INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
				"INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
				"INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
				"INSERT INTO t_checked VALUES (1), (50), (100);",
				"INSERT INTO t_fk_parent VALUES (10), (20), (30);",
				"INSERT INTO t_fk_child VALUES (10), (20);",
				"INSERT INTO t_unique VALUES (7), (8), (9);",
				"INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
				"INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
				"INSERT INTO t_default_func (v2) VALUES (1), (2);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT COUNT(*) FROM dolt_status;",
					Expected: []sql.Row{{16}},
				},
				{
					Query:    "SELECT DOLT_STASH('push', 'dgstash', '--all');",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    "SELECT COUNT(*) FROM dolt_status;",
					Expected: []sql.Row{{0}},
				},
				{
					Query:    "SELECT DOLT_STASH('pop', 'dgstash');",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    "SELECT COUNT(*) FROM dolt_status;",
					Expected: []sql.Row{{16}},
				},
			},
		},
		{
			Name: "Push and pop tracked",
			Skip: true, // TODO: errors with "dolt_procedures: unsupported type int", which seems like the Doltgres->Dolt layer is erroring
			SetUpScript: []string{
				"CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
				"CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
				"CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
				"CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
				"CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
				"CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
				"CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
				"CREATE TABLE t_unique (v1 INT4 UNIQUE);",
				"CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
				"CREATE TABLE t_trigger (v1 INT8);",
				"CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
				"CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
				"CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
				"SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 34;",
				"INSERT INTO t_simple VALUES (1), (2), (3);",
				"INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
				"INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
				"INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
				"INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
				"INSERT INTO t_checked VALUES (1), (50), (100);",
				"INSERT INTO t_fk_parent VALUES (10), (20), (30);",
				"INSERT INTO t_fk_child VALUES (10), (20);",
				"INSERT INTO t_unique VALUES (7), (8), (9);",
				"INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
				"CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 33; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
				"CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 34; END; $$ LANGUAGE plpgsql;",
				"INSERT INTO t_default_func (v2) VALUES (1), (2);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT COUNT(*) FROM dolt_status;",
					Expected: []sql.Row{{15}},
				},
				{
					Query:    "SELECT DOLT_STASH('push', 'dgstash');",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    "SELECT COUNT(*) FROM dolt_status;",
					Expected: []sql.Row{{0}},
				},
				{
					Query:    "SELECT DOLT_STASH('pop', 'dgstash');",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    "SELECT COUNT(*) FROM dolt_status;",
					Expected: []sql.Row{{15}},
				},
			},
		},
	})
}

func TestDoltTag(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Smoke test",
			SetUpScript: []string{
				"CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
				"INSERT INTO t_simple VALUES (1), (2), (3);",
				"SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 34;",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT DOLT_TAG('tagged_commit', 'HEAD');",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:       "SELECT DOLT_CHECKOUT('tagged_commit');",
					ExpectedErr: "detached head",
				},
			},
		},
	})
}

func TestDoltFunctionSmokeTests(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "smoke test select dolt_add and dolt_commit",
			SetUpScript: []string{
				"CREATE TABLE t1 (pk int primary key);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "select dolt_add('.')",
					Expected: []sql.Row{
						{"{0}"},
					},
				},
				{
					Query:            "select dolt_commit('-am', 'new table')",
					SkipResultsCheck: true,
				},
				{
					Query: "select count(*) from dolt.log",
					Expected: []sql.Row{
						{3}, // initial commit, CREATE DATABASE commit, CREATE TABLE commit
					},
				},
				{
					Query: "select message from dolt.log order by date desc limit 1",
					Expected: []sql.Row{
						{"new table"},
					},
				},
			},
		},
		{
			Name: "smoke test select dolt_merge",
			SetUpScript: []string{
				"CREATE TABLE t1 (pk int primary key);",
				"SELECT DOLT_COMMIT('-Am', 'new table');",
				"SELECT DOLT_CHECKOUT('-b', 'new-branch');",
				"CREATE TABLE t2 (pk int primary key);",
				"SELECT DOLT_COMMIT('-Am', 'new table on new branch');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:            "SELECT DOLT_MERGE_BASE('main', 'new-branch');",
					SkipResultsCheck: true,
				},
				{
					Query: "SELECT DOLT_CHECKOUT('main');",
					Expected: []sql.Row{
						{"{0,\"Switched to branch 'main'\"}"},
					},
				},
				{
					Query: "select count(*) from dolt.log",
					Expected: []sql.Row{
						{3}, // initial commit, CREATE DATABASE commit, CREATE TABLE commit
					},
				},
				{
					Query:            "SELECT DOLT_MERGE('new-branch', '--no-ff', '-m', 'merge new-branch into main');",
					SkipResultsCheck: true,
				},
				{
					Query: "select count(*) from dolt.log",
					Expected: []sql.Row{
						{5}, // initial commit, CREATE DATABASE commit, CREATE TABLE t1 commit, new CREATE TABLE t2 commit, merge commit
					},
				},
			},
		},
		{
			Name: "smoke test select dolt_merge dirty working set, same table",
			SetUpScript: []string{
				"CREATE TABLE t1 (pk int primary key);",
				"SELECT DOLT_COMMIT('-Am', 'new table');",
				"INSERT INTO t1 VALUES (1);",
				"SELECT DOLT_CHECKOUT('-b', 'new-branch');",
				"INSERT INTO t1 VALUES (2);",
				"SELECT DOLT_COMMIT('-Am', 'new row on new branch');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:            "SELECT DOLT_MERGE_BASE('main', 'new-branch');",
					SkipResultsCheck: true,
				},
				{
					Query: "SELECT DOLT_CHECKOUT('main');",
					Expected: []sql.Row{
						{"{0,\"Switched to branch 'main'\"}"},
					},
				},
				{
					Query: "SELECT * FROM dolt.status",
					Expected: []sql.Row{
						{"public.t1", "f", "modified"},
					},
				},
				{
					Query:       "SELECT DOLT_MERGE('new-branch', '--no-ff', '-m', 'merge new-branch into main');",
					ExpectedErr: "error: local changes would be stomped by merge",
				},
				{
					Query: "SELECT * FROM dolt.status",
					Expected: []sql.Row{
						{"public.t1", "f", "modified"},
					},
				},
			},
		},
		{
			Name: "smoke test select dolt_merge dirty working set, different tables",
			SetUpScript: []string{
				"CREATE TABLE t1 (pk int primary key);",
				"SELECT DOLT_COMMIT('-Am', 'new table');",
				"INSERT INTO t1 VALUES (1);",
				"SELECT DOLT_CHECKOUT('-b', 'new-branch');",
				"CREATE TABLE t2 (pk int primary key);",
				"SELECT DOLT_COMMIT('-Am', 'new row on new branch');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:            "SELECT DOLT_MERGE_BASE('main', 'new-branch');",
					SkipResultsCheck: true,
				},
				{
					Query: "SELECT DOLT_CHECKOUT('main');",
					Expected: []sql.Row{
						{"{0,\"Switched to branch 'main'\"}"},
					},
				},
				{
					Query: "SELECT * FROM dolt.status",
					Expected: []sql.Row{
						{"public.t1", "f", "modified"},
					},
				},
				{
					Query:            "SELECT DOLT_MERGE('new-branch', '--no-ff', '-m', 'merge new-branch into main');",
					SkipResultsCheck: true,
				},
				{
					Query: "SELECT * FROM dolt.status",
					Expected: []sql.Row{
						{"public.t1", "f", "modified"},
					},
				},
			},
		},
		{
			Name: "smoke test select dolt_reset",
			SetUpScript: []string{
				"CREATE TABLE t1 (pk int primary key);",
				"INSERT INTO t1 VALUES (1);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM dolt.status;",
					Expected: []sql.Row{
						{"public.t1", "f", "new table"},
					},
				},
				{
					Query:    "SELECT DOLT_ADD('t1');",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query: "SELECT * FROM dolt.status;",
					Expected: []sql.Row{
						{"public.t1", "t", "new table"},
					},
				},
				{
					Query:    "SELECT DOLT_RESET('t1');",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query: "SELECT * FROM dolt.status;",
					Expected: []sql.Row{
						{"public.t1", "f", "new table"},
					},
				},
			},
		},
		{
			Name: "smoke test select dolt_clean",
			SetUpScript: []string{
				"CREATE TABLE t1 (pk int primary key);",
				"INSERT INTO t1 VALUES (1);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM dolt.status;",
					Expected: []sql.Row{
						{"public.t1", "f", "new table"},
					},
				},
				{
					Skip:     true, // TODO: function dolt_clean() does not exist
					Query:    "SELECT DOLT_CLEAN();",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    "SELECT DOLT_CLEAN('t1');",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    "SELECT * FROM dolt.status;",
					Expected: []sql.Row{},
				},
				{
					Query:    "CREATE TABLE t1 (pk int primary key);",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM dolt.status;",
					Expected: []sql.Row{
						{"public.t1", "f", "new table"},
					},
				},
				{
					Skip:     true, // TODO: function dolt_clean() does not exist
					Query:    "SELECT DOLT_CLEAN();",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Skip:     true,
					Query:    "SELECT * FROM dolt.status;",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "smoke test select dolt_checkout(table)",
			SetUpScript: []string{
				"CREATE TABLE t1 (pk int primary key);",
				"INSERT INTO t1 VALUES (1);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM dolt.status;",
					Expected: []sql.Row{
						{"public.t1", "f", "new table"},
					},
				},
				{
					Query:    "SELECT DOLT_CHECKOUT('t1');",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    "SELECT * FROM dolt.status;",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "smoke test select dolt diff functions and tables",
			SetUpScript: []string{
				"CREATE TABLE t1 (pk int primary key);",
				"INSERT INTO t1 VALUES (1);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM dolt_diff_stat('HEAD', 'WORKING')",
					Expected: []sql.Row{
						{"public.t1", 0, 1, 0, 0, 1, 0, 0, 0, 1, 0, 1},
					},
				},
				{
					Query: "SELECT * FROM dolt_diff_stat('HEAD', 'WORKING', 't1')",
					Expected: []sql.Row{
						{"public.t1", 0, 1, 0, 0, 1, 0, 0, 0, 1, 0, 1},
					},
				},
				{
					Query: "SELECT * FROM dolt_diff_summary('HEAD', 'WORKING')",
					Expected: []sql.Row{
						{"", "public.t1", "added", 1, 1},
					},
				},
				{
					Query: "SELECT * FROM dolt_diff_summary('HEAD', 'WORKING', 't1')",
					Expected: []sql.Row{
						{"", "public.t1", "added", 1, 1},
					},
				},
				{
					Query: "SELECT diff_type, from_pk, to_pk FROM dolt_diff('HEAD', 'WORKING', 't1')",
					Expected: []sql.Row{
						{"added", nil, 1},
					},
				},
				{
					Query: "SELECT diff_type, from_pk, to_pk FROM dolt_diff('HEAD', 'WORKING', 't1')",
					Expected: []sql.Row{
						{"added", nil, 1},
					},
				},
				{
					Skip:  true, // TODO: dolt_commit_diff_* tables must be filtered to a single 'to_commit'
					Query: "SELECT diff_type, from_pk, to_pk FROM dolt_commit_diff_t1 WHERE to_commit=HASHOF('main') AND from_commit='WORKING'",
					Expected: []sql.Row{
						{"added", nil, 1},
					},
				},
				{
					Query: "SELECT * FROM dolt.diff",
					Expected: []sql.Row{
						{"WORKING", "public.t1", nil, nil, nil, nil, "t", "t"},
					},
				},
				{
					Query: "SELECT statement_order, table_name, diff_type, statement FROM dolt_patch('HEAD', 'WORKING')",
					Expected: []sql.Row{
						{Numeric("1"), "public.t1", "schema", "CREATE TABLE \"t1\" (\n  \"pk\" integer NOT NULL,\n  PRIMARY KEY (\"pk\")\n);"},
						{Numeric("2"), "public.t1", "data", "INSERT INTO \"t1\" (\"pk\") VALUES (1);"},
					},
				},
				{
					Query: "SELECT statement_order, table_name, diff_type, statement FROM dolt_patch('HEAD', 'WORKING', 't1')",
					Expected: []sql.Row{
						{Numeric("1"), "public.t1", "schema", "CREATE TABLE \"t1\" (\n  \"pk\" integer NOT NULL,\n  PRIMARY KEY (\"pk\")\n);"},
						{Numeric("2"), "public.t1", "data", "INSERT INTO \"t1\" (\"pk\") VALUES (1);"},
					},
				},
				{
					Query: "SELECT * FROM dolt_schema_diff('HEAD', 'WORKING')",
					Expected: []sql.Row{
						{"", "public.t1", "", "CREATE TABLE \"t1\" (\n  \"pk\" integer NOT NULL,\n  PRIMARY KEY (\"pk\")\n);"},
					},
				},
				{
					Query: "SELECT * FROM dolt_schema_diff('HEAD', 'WORKING', 't1')",
					Expected: []sql.Row{
						{"", "public.t1", "", "CREATE TABLE \"t1\" (\n  \"pk\" integer NOT NULL,\n  PRIMARY KEY (\"pk\")\n);"},
					},
				},
				{
					Skip:  true, // ERROR: table not found: t1
					Query: "SELECT * FROM dolt_query_diff('select * from t1 as of main', 'select * from t1')",
					Expected: []sql.Row{
						{"", "t1", "added", 1, 1},
					},
				},
			},
		},
		{
			Name: "smoke test select dolt diff functions and tables for multiple schemas",
			SetUpScript: []string{
				"CREATE TABLE t1 (pk int primary key);",
				"INSERT INTO t1 VALUES (1);",
				"CREATE SCHEMA testschema;",
				"CREATE TABLE testschema.t2 (pk int primary key);",
				"INSERT INTO testschema.t2 VALUES (1);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM dolt.status;",
					Expected: []sql.Row{
						{"public.t1", "f", "new table"},
						{"testschema.t2", "f", "new table"},
						{"testschema", "f", "new schema"},
					},
				},
				{
					Query: "SELECT * FROM dolt_diff_stat('HEAD', 'WORKING')",
					Expected: []sql.Row{
						{"public.t1", 0, 1, 0, 0, 1, 0, 0, 0, 1, 0, 1},
						{"testschema.t2", 0, 1, 0, 0, 1, 0, 0, 0, 1, 0, 1},
					},
				},
				{
					Query: "SELECT * FROM dolt_diff_stat('HEAD', 'WORKING', 't1')",
					Expected: []sql.Row{
						{"public.t1", 0, 1, 0, 0, 1, 0, 0, 0, 1, 0, 1},
					},
				},
				{
					Query: "SELECT * FROM dolt_diff_stat('HEAD', 'WORKING', 't2')",
					Expected: []sql.Row{
						{"testschema.t2", 0, 1, 0, 0, 1, 0, 0, 0, 1, 0, 1},
					},
				},
				{
					Query: "SELECT * FROM dolt_diff_summary('HEAD', 'WORKING')",
					Expected: []sql.Row{
						{"", "public.t1", "added", 1, 1},
						{"", "testschema.t2", "added", 1, 1},
					},
				},
				{
					Query: "SELECT * FROM dolt_diff_summary('HEAD', 'WORKING', 't1')",
					Expected: []sql.Row{
						{"", "public.t1", "added", 1, 1},
					},
				},
				{
					Query: "SELECT * FROM dolt_diff_summary('HEAD', 'WORKING', 't2')",
					Expected: []sql.Row{
						{"", "testschema.t2", "added", 1, 1},
					},
				},
				{
					Query: "SELECT diff_type, from_pk, to_pk FROM dolt_diff('HEAD', 'WORKING', 't1')",
					Expected: []sql.Row{
						{"added", nil, 1},
					},
				},
				{
					Query: "SELECT diff_type, from_pk, to_pk FROM dolt_diff('HEAD', 'WORKING', 't2')",
					Expected: []sql.Row{
						{"added", nil, 1},
					},
				},
				{
					Skip:  true, // TODO: dolt_commit_diff_* tables must be filtered to a single 'to_commit'
					Query: "SELECT diff_type, from_pk, to_pk FROM dolt_commit_diff_t1 WHERE to_commit=HASHOF('main') AND from_commit='WORKING'",
					Expected: []sql.Row{
						{"added", nil, 1},
					},
				},
				{
					Query: "SELECT * FROM dolt.diff",
					Expected: []sql.Row{
						{"WORKING", "public.t1", nil, nil, nil, nil, "t", "t"},
						{"WORKING", "testschema.t2", nil, nil, nil, nil, "t", "t"},
					},
				},
				{
					Query: "SELECT statement_order, table_name, diff_type, statement FROM dolt_patch('HEAD', 'WORKING')",
					Expected: []sql.Row{
						{Numeric("1"), "public.t1", "schema", "CREATE TABLE \"t1\" (\n  \"pk\" integer NOT NULL,\n  PRIMARY KEY (\"pk\")\n);"},
						{Numeric("2"), "public.t1", "data", "INSERT INTO \"t1\" (\"pk\") VALUES (1);"},
						{Numeric("3"), "testschema.t2", "schema", "CREATE TABLE \"t2\" (\n  \"pk\" integer NOT NULL,\n  PRIMARY KEY (\"pk\")\n);"},
						{Numeric("4"), "testschema.t2", "data", "INSERT INTO \"t2\" (\"pk\") VALUES (1);"},
					},
				},
				{
					Query: "SELECT statement_order, table_name, diff_type, statement FROM dolt_patch('HEAD', 'WORKING', 't1')",
					Expected: []sql.Row{
						{Numeric("1"), "public.t1", "schema", "CREATE TABLE \"t1\" (\n  \"pk\" integer NOT NULL,\n  PRIMARY KEY (\"pk\")\n);"},
						{Numeric("2"), "public.t1", "data", "INSERT INTO \"t1\" (\"pk\") VALUES (1);"},
					},
				},
				{
					Query: "SELECT statement_order, table_name, diff_type, statement FROM dolt_patch('HEAD', 'WORKING', 't2')",
					Expected: []sql.Row{
						{Numeric("1"), "testschema.t2", "schema", "CREATE TABLE \"t2\" (\n  \"pk\" integer NOT NULL,\n  PRIMARY KEY (\"pk\")\n);"},
						{Numeric("2"), "testschema.t2", "data", "INSERT INTO \"t2\" (\"pk\") VALUES (1);"},
					},
				},
				{
					Query: "SELECT * FROM dolt_schema_diff('HEAD', 'WORKING')",
					Expected: []sql.Row{
						{"", "public.t1", "", "CREATE TABLE \"t1\" (\n  \"pk\" integer NOT NULL,\n  PRIMARY KEY (\"pk\")\n);"},
						{"", "testschema.t2", "", "CREATE TABLE \"t2\" (\n  \"pk\" integer NOT NULL,\n  PRIMARY KEY (\"pk\")\n);"},
					},
				},
				{
					Query: "SELECT * FROM dolt_schema_diff('HEAD', 'WORKING', 't1')",
					Expected: []sql.Row{
						{"", "public.t1", "", "CREATE TABLE \"t1\" (\n  \"pk\" integer NOT NULL,\n  PRIMARY KEY (\"pk\")\n);"},
					},
				},
				{
					Query: "SELECT * FROM dolt_schema_diff('HEAD', 'WORKING', 't2')",
					Expected: []sql.Row{
						{"", "testschema.t2", "", "CREATE TABLE \"t2\" (\n  \"pk\" integer NOT NULL,\n  PRIMARY KEY (\"pk\")\n);"},
					},
				},
				{
					Skip:  true, // ERROR: table not found: t1
					Query: "SELECT * FROM dolt_query_diff('select * from t1 as of main', 'select * from t1')",
					Expected: []sql.Row{
						{"", "public.t1", "added", 1, 1},
					},
				},
				{
					Skip:  true, // ERROR: table not found: t2
					Query: "SELECT * FROM dolt_query_diff('select * from t2 as of main', 'select * from t2')",
					Expected: []sql.Row{
						{"", "public.t1", "added", 1, 1},
					},
				},
			},
		},
		{
			Name: "DOLT_PREVIEW_MERGE_CONFLICTS basic functionality",
			SetUpScript: []string{
				"CREATE TABLE t1 (pk int primary key, c1 int);",
				"INSERT INTO t1 VALUES (1, 10), (2, 20);",
				"SELECT DOLT_COMMIT('-Am', 'initial commit');",
				"SELECT DOLT_CHECKOUT('-b', 'branch1');",
				"UPDATE t1 SET c1 = 100 WHERE pk = 1;",
				"SELECT DOLT_COMMIT('-am', 'update on branch1');",
				"SELECT DOLT_CHECKOUT('main');",
				"UPDATE t1 SET c1 = 200 WHERE pk = 1;",
				"SELECT DOLT_COMMIT('-am', 'update on main');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS_SUMMARY('main', 'branch1')",
					Expected: []sql.Row{
						{"public.t1", Numeric("1"), Numeric("0")},
					},
				},
				{
					Query: "SELECT base_pk, base_c1, our_pk, our_c1, our_diff_type, their_pk, their_c1, their_diff_type FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', 't1')",
					Expected: []sql.Row{
						{1, 10, 1, 200, "modified", 1, 100, "modified"},
					},
				},
			},
		},
		{
			Name: "DOLT_PREVIEW_MERGE_CONFLICTS with no conflicts",
			SetUpScript: []string{
				"CREATE TABLE t1 (pk int primary key, c1 int);",
				"INSERT INTO t1 VALUES (1, 10), (2, 20);",
				"SELECT DOLT_COMMIT('-Am', 'initial commit');",
				"SELECT DOLT_CHECKOUT('-b', 'branch1');",
				"INSERT INTO t1 VALUES (3, 30);",
				"SELECT DOLT_COMMIT('-am', 'insert on branch1');",
				"SELECT DOLT_CHECKOUT('main');",
				"INSERT INTO t1 VALUES (4, 40);",
				"SELECT DOLT_COMMIT('-am', 'insert on main');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS_SUMMARY('main', 'branch1')",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT base_pk, base_c1, our_pk, our_c1, our_diff_type, their_pk, their_c1, their_diff_type FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', 't1')",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "DOLT_PREVIEW_MERGE_CONFLICTS with multiple tables",
			SetUpScript: []string{
				"CREATE TABLE t1 (pk int primary key, c1 int);",
				"CREATE TABLE t2 (pk int primary key, c1 varchar(20));",
				"INSERT INTO t1 VALUES (1, 10);",
				"INSERT INTO t2 VALUES (1, 'initial');",
				"SELECT DOLT_COMMIT('-Am', 'initial commit');",
				"SELECT DOLT_CHECKOUT('-b', 'branch1');",
				"UPDATE t1 SET c1 = 100 WHERE pk = 1;",
				"UPDATE t2 SET c1 = 'branch1' WHERE pk = 1;",
				"SELECT DOLT_COMMIT('-am', 'updates on branch1');",
				"SELECT DOLT_CHECKOUT('main');",
				"UPDATE t1 SET c1 = 200 WHERE pk = 1;",
				"UPDATE t2 SET c1 = 'main' WHERE pk = 1;",
				"SELECT DOLT_COMMIT('-am', 'updates on main');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS_SUMMARY('main', 'branch1') ORDER BY 'table'",
					Expected: []sql.Row{
						{"public.t1", Numeric("1"), Numeric("0")},
						{"public.t2", Numeric("1"), Numeric("0")},
					},
				},
				{
					Query: "SELECT COUNT(*) FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', 't1')",
					Expected: []sql.Row{
						{1},
					},
				},
				{
					Query: "SELECT COUNT(*) FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', 't2')",
					Expected: []sql.Row{
						{1},
					},
				},
			},
		},
		{
			Name: "DOLT_PREVIEW_MERGE_CONFLICTS with schema conflicts",
			SetUpScript: []string{
				"CREATE TABLE t1 (pk int primary key, c1 int);",
				"INSERT INTO t1 VALUES (1, 10);",
				"SELECT DOLT_COMMIT('-Am', 'initial commit');",
				"SELECT DOLT_CHECKOUT('-b', 'branch1');",
				"ALTER TABLE t1 ADD COLUMN c2 varchar(50);",
				"SELECT DOLT_COMMIT('-am', 'add column on branch1');",
				"SELECT DOLT_CHECKOUT('main');",
				"ALTER TABLE t1 ADD COLUMN c2 int;",
				"SELECT DOLT_COMMIT('-am', 'add same column different type on main');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS_SUMMARY('main', 'branch1')",
					Expected: []sql.Row{
						{"public.t1", nil, Numeric("1")},
					},
				},
				{
					Query:       "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', 't1')",
					ExpectedErr: "schema conflicts found: 1",
				},
			},
		},
		{
			Name: "DOLT_PREVIEW_MERGE_CONFLICTS with multiple schemas",
			SetUpScript: []string{
				"CREATE SCHEMA test_schema;",
				"CREATE TABLE t1 (pk int primary key, c1 int);",
				"CREATE TABLE test_schema.t2 (pk int primary key, c1 int);",
				"INSERT INTO t1 VALUES (1, 10);",
				"INSERT INTO test_schema.t2 VALUES (1, 20);",
				"SELECT DOLT_COMMIT('-Am', 'initial commit');",
				"SELECT DOLT_CHECKOUT('-b', 'branch1');",
				"UPDATE t1 SET c1 = 100 WHERE pk = 1;",
				"UPDATE test_schema.t2 SET c1 = 200 WHERE pk = 1;",
				"SELECT DOLT_COMMIT('-am', 'updates on branch1');",
				"SELECT DOLT_CHECKOUT('main');",
				"UPDATE t1 SET c1 = 300 WHERE pk = 1;",
				"UPDATE test_schema.t2 SET c1 = 400 WHERE pk = 1;",
				"SELECT DOLT_COMMIT('-am', 'updates on main');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS_SUMMARY('main', 'branch1') ORDER BY 'table'",
					Expected: []sql.Row{
						{"public.t1", Numeric("1"), Numeric("0")},
						{"test_schema.t2", Numeric("1"), Numeric("0")},
					},
				},
				{
					Query: "SELECT COUNT(*) FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', 't1')",
					Expected: []sql.Row{
						{1},
					},
				},
				{
					Query: "SELECT COUNT(*) FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', 't2')",
					Expected: []sql.Row{
						{1},
					},
				},
			},
		},
		{
			Name: "DOLT_PREVIEW_MERGE_CONFLICTS with multiple schemas, same name",
			SetUpScript: []string{
				"CREATE SCHEMA test_schema;",
				"CREATE TABLE t1 (pk int primary key, c1 int);",
				"CREATE TABLE test_schema.t1 (pk int primary key, c2 int);",
				"INSERT INTO t1 VALUES (1, 10);",
				"INSERT INTO test_schema.t1 VALUES (1, 20);",
				"SELECT DOLT_COMMIT('-Am', 'initial commit');",
				"SELECT DOLT_CHECKOUT('-b', 'branch1');",
				"UPDATE t1 SET c1 = 100 WHERE pk = 1;",
				"UPDATE test_schema.t1 SET c2 = 200 WHERE pk = 1;",
				"SELECT DOLT_COMMIT('-am', 'updates on branch1');",
				"SELECT DOLT_CHECKOUT('main');",
				"UPDATE t1 SET c1 = 300 WHERE pk = 1;",
				"UPDATE test_schema.t1 SET c2 = 400 WHERE pk = 1;",
				"SELECT DOLT_COMMIT('-am', 'updates on main');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS_SUMMARY('main', 'branch1') ORDER BY 'table'",
					Expected: []sql.Row{
						{"public.t1", Numeric("1"), Numeric("0")},
						{"test_schema.t1", Numeric("1"), Numeric("0")},
					},
				},
				{
					Query: "SELECT base_c1 FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', 't1')",
					Expected: []sql.Row{
						{10},
					},
				},
				{
					Query:       "SELECT base_c2 FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', 't1')",
					ExpectedErr: "column \"base_c2\" could not be found in any table in scope",
				},
				{
					Query:    "SET search_path TO test_schema;",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT base_c2 FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', 't1')",
					Expected: []sql.Row{
						{20},
					},
				},
				{
					Query:       "SELECT base_c1 FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', 't1')",
					ExpectedErr: "column \"base_c1\" could not be found in any table in scope",
				},
			},
		},
		{
			Name: "DOLT_PREVIEW_MERGE_CONFLICTS error cases",
			SetUpScript: []string{
				"CREATE TABLE t1 (pk int primary key, c1 int);",
				"SELECT DOLT_COMMIT('-Am', 'initial commit');",
				"SELECT DOLT_CHECKOUT('-b', 'branch1');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS_SUMMARY('nonexistent-branch', 'main')",
					ExpectedErr: "branch not found: nonexistent-branch",
				},
				{
					Query:       "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS_SUMMARY('main', 'branch1', 'table')",
					ExpectedErr: "function 'dolt_preview_merge_conflicts_summary' expected 2 arguments, 3 received",
				},
				{
					Query:       "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS_SUMMARY('main', 'nonexistent-branch')",
					ExpectedErr: "branch not found: nonexistent-branch",
				},
				{
					Query:       "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS_SUMMARY('', 'main')",
					ExpectedErr: "branch name cannot be empty",
				},
				{
					Query:       "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS_SUMMARY('main', '')",
					ExpectedErr: "branch name cannot be empty",
				},
				{
					Query:       "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS_SUMMARY(NULL, 'main')",
					ExpectedErr: "Invalid argument to dolt_preview_merge_conflicts_summary: NULL",
				},
				{
					Query:       "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS_SUMMARY('main', NULL)",
					ExpectedErr: "Invalid argument to dolt_preview_merge_conflicts_summary: NULL",
				},
				{
					Query:       "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS('nonexistent-branch', 'main', 't1')",
					ExpectedErr: "branch not found: nonexistent-branch",
				},
				{
					Query:       "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'nonexistent-branch', 't1')",
					ExpectedErr: "branch not found: nonexistent-branch",
				},
				{
					Query:       "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1')",
					ExpectedErr: "function 'dolt_preview_merge_conflicts' expected 3 arguments, 2 received",
				},
				{
					Query:       "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', 't1', 'extra')",
					ExpectedErr: "function 'dolt_preview_merge_conflicts' expected 3 arguments, 4 received",
				},
				{
					Query:       "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS('', 'main', 't1')",
					ExpectedErr: "string is not a valid branch or hash",
				},
				{
					Query:       "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', '', 't1')",
					ExpectedErr: "string is not a valid branch or hash",
				},
				{
					Query:       "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS(NULL, 'main', 't1')",
					ExpectedErr: "Invalid argument to dolt_preview_merge_conflicts: NULL",
				},
				{
					Query:       "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', NULL, 't1')",
					ExpectedErr: "Invalid argument to dolt_preview_merge_conflicts: NULL",
				},
				{
					Query:       "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', NULL)",
					ExpectedErr: "Invalid argument to dolt_preview_merge_conflicts: NULL",
				},
				{
					Query:       "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', 'nonexistent_table')",
					ExpectedErr: "table not found: public.nonexistent_table",
				},
				{
					Query:       "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', '')",
					ExpectedErr: "table name cannot be empty",
				},
			},
		},
	})
}
