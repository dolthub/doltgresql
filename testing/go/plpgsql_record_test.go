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

	"github.com/dolthub/go-mysql-server/sql"
)

// TestPlpgsqlRecordInto covers using a variable declared as RECORD as the INTO target of a
// SQL statement (SELECT/INSERT ... RETURNING/EXECUTE). All expectations were verified against
// PostgreSQL 16.
func TestPlpgsqlRecordInto(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "SELECT INTO a RECORD variable",
			SetUpScript: []string{
				`CREATE TABLE k (id int, name text, amt numeric);`,
				`INSERT INTO k VALUES (1, 'a', 10.5), (2, 'b', 20.25);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					// The record target is never read, so only the conversion has to succeed.
					Query: `CREATE FUNCTION f_repro() RETURNS int LANGUAGE plpgsql AS $$
DECLARE r RECORD;
BEGIN SELECT id INTO r FROM k LIMIT 1; RETURN 1; END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_repro();`,
					Expected: []sql.Row{{1}},
				},
				{
					Query: `CREATE FUNCTION f_field_text(p int) RETURNS text LANGUAGE plpgsql AS $$
DECLARE r RECORD;
BEGIN
	SELECT id, name INTO r FROM k WHERE id = p;
	RETURN r.name;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_field_text(2);`,
					Expected: []sql.Row{{"b"}},
				},
				{
					Query: `CREATE FUNCTION f_field_int(p int) RETURNS int LANGUAGE plpgsql AS $$
DECLARE r RECORD;
BEGIN
	SELECT id, name INTO r FROM k WHERE id = p;
	RETURN r.id;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_field_int(2);`,
					Expected: []sql.Row{{2}},
				},
				{
					// SELECT * INTO a RECORD picks up every column of the table.
					Query: `CREATE FUNCTION f_star() RETURNS numeric LANGUAGE plpgsql AS $$
DECLARE r RECORD;
BEGIN
	SELECT * INTO r FROM k WHERE id = 1;
	RETURN r.amt;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_star();`,
					Expected: []sql.Row{{Numeric("10.5")}},
				},
				{
					// Field names come from the query's output column names, not from any table.
					Query: `CREATE FUNCTION f_agg() RETURNS bigint LANGUAGE plpgsql AS $$
DECLARE r RECORD;
BEGIN
	SELECT count(*) AS c INTO r FROM k;
	RETURN r.c;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_agg();`,
					Expected: []sql.Row{{2}},
				},
				{
					// When the query matches no rows, every field of the record is NULL.
					Query: `CREATE FUNCTION f_nomatch() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE r RECORD;
BEGIN
	SELECT id, name INTO r FROM k WHERE id = 999;
	RETURN r.id IS NULL;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_nomatch();`,
					Expected: []sql.Row{{"t"}},
				},
				{
					// A RECORD takes the shape of whatever was last assigned to it.
					Query: `CREATE FUNCTION f_reshape() RETURNS text LANGUAGE plpgsql AS $$
DECLARE r RECORD;
BEGIN
	SELECT id, name INTO r FROM k WHERE id = 1;
	SELECT name AS other INTO r FROM k WHERE id = 2;
	RETURN r.other;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_reshape();`,
					Expected: []sql.Row{{"b"}},
				},
			},
		},
		{
			Name: "INSERT RETURNING INTO a RECORD variable",
			SetUpScript: []string{
				`CREATE TABLE k (id int, name text);`,
				`INSERT INTO k VALUES (1, 'a');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION f_insert_returning() RETURNS text LANGUAGE plpgsql AS $$
DECLARE r RECORD;
BEGIN
	INSERT INTO k VALUES (3, 'c') RETURNING id, name INTO r;
	RETURN r.name;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_insert_returning();`,
					Expected: []sql.Row{{"c"}},
				},
				{
					Query:    `SELECT id, name FROM k ORDER BY id;`,
					Expected: []sql.Row{{1, "a"}, {3, "c"}},
				},
			},
		},
		{
			Name: "EXECUTE INTO a RECORD variable",
			SetUpScript: []string{
				`CREATE TABLE k (id int, name text);`,
				`INSERT INTO k VALUES (1, 'a'), (2, 'b');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION f_exec() RETURNS text LANGUAGE plpgsql AS $$
DECLARE r RECORD;
BEGIN
	EXECUTE 'SELECT id, name FROM k WHERE id = 2' INTO r;
	RETURN r.name;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_exec();`,
					Expected: []sql.Row{{"b"}},
				},
			},
		},
		{
			Name: "trigger function using SELECT INTO a RECORD variable",
			SetUpScript: []string{
				`CREATE TABLE t (id int primary key, v int);`,
				`CREATE FUNCTION trg_guard() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE blocker RECORD;
BEGIN
	SELECT id, v INTO blocker FROM t WHERE v > NEW.v LIMIT 1;
	IF blocker.id IS NOT NULL THEN
		RAISE EXCEPTION 'blocked by row %', blocker.id;
	END IF;
	RETURN NEW;
END; $$;`,
				`CREATE TRIGGER trg BEFORE INSERT ON t FOR EACH ROW EXECUTE FUNCTION trg_guard();`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `INSERT INTO t VALUES (1, 10);`,
					Expected: []sql.Row{},
				},
				{
					Query:       `INSERT INTO t VALUES (2, 5);`,
					ExpectedErr: `blocked by row 1`,
				},
				{
					Query:    `SELECT id, v FROM t ORDER BY id;`,
					Expected: []sql.Row{{1, 10}},
				},
			},
		},
		{
			Name: "RECORD fields written as quoted identifiers",
			SetUpScript: []string{
				`CREATE TABLE k3 ("id" int, "book_date" date);`,
				`INSERT INTO k3 VALUES (7, '2026-01-02'), (9, '2026-03-04');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION f_quoted() RETURNS int LANGUAGE plpgsql AS $$
DECLARE blocker RECORD;
BEGIN
	SELECT b."id", b."book_date" INTO blocker FROM k3 b ORDER BY b."book_date" LIMIT 1;
	IF blocker."id" IS NOT NULL THEN
		RETURN blocker."id";
	END IF;
	RETURN -1;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_quoted();`,
					Expected: []sql.Row{{7}},
				},
				{
					// A record field read through RAISE resolves the same way as one read through a query.
					Query: `CREATE FUNCTION f_raise() RETURNS void LANGUAGE plpgsql AS $$
DECLARE r RECORD;
BEGIN
	SELECT "id" INTO r FROM k3 ORDER BY "id" LIMIT 1;
	RAISE EXCEPTION 'saw %', r."id";
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT f_raise();`,
					ExpectedErr: `saw 7`,
				},
			},
		},
		{
			Name: "FOR..IN..SELECT over a RECORD variable",
			SetUpScript: []string{
				`CREATE TABLE k4 (id int, name text);`,
				`INSERT INTO k4 VALUES (1, 'a'), (2, 'b'), (3, 'c');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					// The existing FOR..IN..SELECT coverage loops over an empty result, so the record is never
					// actually assigned. This reads fields off it on every iteration.
					Query: `CREATE FUNCTION f_forloop() RETURNS text LANGUAGE plpgsql AS $$
DECLARE r RECORD; acc text := '';
BEGIN
	FOR r IN SELECT id, name FROM k4 ORDER BY id LOOP
		acc := acc || r.id || r.name;
	END LOOP;
	RETURN acc;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_forloop();`,
					Expected: []sql.Row{{"1a2b3c"}},
				},
			},
		},
		{
			Name: "SELECT INTO a RECORD from a derived table",
			Assertions: []ScriptTestAssertion{
				{
					// The record's shape comes from the query's output columns, which here belong to no table.
					Query: `CREATE FUNCTION f_derived() RETURNS numeric LANGUAGE plpgsql AS $$
DECLARE fig RECORD;
BEGIN
	SELECT * INTO fig FROM (SELECT 12.5::numeric AS computed_balance, 3::bigint AS line_count) f;
	RETURN fig.computed_balance;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_derived();`,
					Expected: []sql.Row{{Numeric("12.5")}},
				},
			},
		},
		{
			Name: "errors accessing RECORD fields",
			SetUpScript: []string{
				`CREATE TABLE k (id int, name text);`,
				`INSERT INTO k VALUES (1, 'a');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION f_badfield() RETURNS text LANGUAGE plpgsql AS $$
DECLARE r RECORD;
BEGIN
	SELECT id INTO r FROM k WHERE id = 1;
	RETURN r.nope;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT f_badfield();`,
					ExpectedErr: `record "r" has no field "nope"`,
				},
				{
					Query: `CREATE FUNCTION f_unassigned() RETURNS text LANGUAGE plpgsql AS $$
DECLARE r RECORD;
BEGIN
	RETURN r.id;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT f_unassigned();`,
					ExpectedErr: `record "r" is not assigned yet`,
				},
			},
		},
	})
}
