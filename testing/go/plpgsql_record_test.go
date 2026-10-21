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
			Name: "RECORD declaration default",
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION f_record_default() RETURNS text LANGUAGE plpgsql AS $$
DECLARE r RECORD := ROW(1, 'a');
BEGIN RETURN r::text; END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_record_default();`,
					Expected: []sql.Row{{"(1,a)"}},
				},
			},
		},
		{
			Name: "RECORD declaration default from NEW",
			SetUpScript: []string{
				`CREATE TABLE src (id int, note text);`,
				`CREATE TABLE res (id int, note text);`,
				`CREATE FUNCTION trg_record_default() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE whole RECORD := NEW;
BEGIN
	INSERT INTO res VALUES (whole.id, whole.note);
	RETURN NEW;
END; $$;`,
				`CREATE TRIGGER t_record_default AFTER INSERT ON src FOR EACH ROW EXECUTE FUNCTION trg_record_default();`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `INSERT INTO src VALUES (1, 'a');`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT id, note FROM res;`,
					Expected: []sql.Row{{1, "a"}},
				},
			},
		},
		{
			Name: "RECORD declaration default referencing an earlier variable",
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION f_record_default_local() RETURNS text LANGUAGE plpgsql AS $$
DECLARE n int := 7; r RECORD := ROW(n, 'a');
BEGIN RETURN r::text; END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_record_default_local();`,
					Expected: []sql.Row{{"(7,a)"}},
				},
			},
		},
		{
			Name: "RECORD declaration default referencing parameters",
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION f_record_default_param(n int, s text) RETURNS text LANGUAGE plpgsql AS $$
DECLARE r RECORD := ROW(n * 2, s);
BEGIN RETURN r::text || '|' || r.f1 || '|' || r.f2; END; $$;`,
					Expected: []sql.Row{},
				},
				{
					// The default is evaluated anew on each call.
					Query:    `SELECT f_record_default_param(1, 'x'), f_record_default_param(2, 'y');`,
					Expected: []sql.Row{{"(2,x)|2|x", "(4,y)|4|y"}},
				},
			},
		},
		{
			Name: "RECORD declaration default referencing an outer block's variable",
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION f_record_default_outer() RETURNS text LANGUAGE plpgsql AS $$
DECLARE n int := 7;
BEGIN
	DECLARE r RECORD := ROW(n, 'a');
	BEGIN RETURN r::text; END;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_record_default_outer();`,
					Expected: []sql.Row{{"(7,a)"}},
				},
			},
		},
		{
			Name: "RECORD declaration default referencing a shadowing variable",
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION f_record_default_shadow() RETURNS text LANGUAGE plpgsql AS $$
DECLARE n int := 1;
BEGIN
	DECLARE n int := 2; r RECORD := ROW(n);
	BEGIN RETURN r::text; END;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_record_default_shadow();`,
					Expected: []sql.Row{{"(2)"}},
				},
			},
		},
		{
			Name: "variable declaration default referencing an earlier RECORD",
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION f_variable_default_record() RETURNS text LANGUAGE plpgsql AS $$
DECLARE r RECORD := ROW(1, 'a'); t text := r::text; m int := r.f1 + 1;
BEGIN RETURN t || '|' || m; END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_variable_default_record();`,
					Expected: []sql.Row{{"(1,a)|2"}},
				},
			},
		},
		{
			Name: "RECORD declaration default from an earlier RECORD",
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION f_record_default_record() RETURNS text LANGUAGE plpgsql AS $$
DECLARE a RECORD := ROW(1, 'x'::text); b RECORD := a;
BEGIN RETURN b::text || '|' || b.f2; END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_record_default_record();`,
					Expected: []sql.Row{{"(1,x)|x"}},
				},
			},
		},
		{
			Name: "RECORD and variable declaration defaults interleaved",
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION f_record_default_chain() RETURNS text LANGUAGE plpgsql AS $$
DECLARE n int := 1; r RECORD := ROW(n); m int := n + 1; s RECORD := ROW(n, m);
BEGIN RETURN r::text || s::text; END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_record_default_chain();`,
					Expected: []sql.Row{{"(1)(1,2)"}},
				},
			},
		},
		{
			Name: "RECORD declaration default referencing a later variable",
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION f_record_default_later() RETURNS text LANGUAGE plpgsql AS $$
DECLARE r RECORD := ROW(n); n int := 1;
BEGIN RETURN r::text; END; $$;`,
					Expected: []sql.Row{},
				},
				{
					// PostgreSQL: column "n" does not exist
					Query:       `SELECT f_record_default_later();`,
					ExpectedErr: `column "n"`,
				},
			},
		},
		{
			Name: "RECORD declaration default from NEW and a variable",
			SetUpScript: []string{
				`CREATE TABLE src (id int, note text);`,
				`CREATE TABLE res (id int, note text);`,
				`CREATE FUNCTION trg_record_default_var() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE bump int := 100; r RECORD := ROW(NEW.id + bump, NEW.note || '!');
BEGIN
	INSERT INTO res VALUES (r.f1, r.f2);
	RETURN NEW;
END; $$;`,
				`CREATE TRIGGER t_record_default_var AFTER INSERT ON src FOR EACH ROW EXECUTE FUNCTION trg_record_default_var();`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `INSERT INTO src VALUES (1, 'a');`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT id, note FROM res;`,
					Expected: []sql.Row{{101, "a!"}},
				},
			},
		},
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
				{
					Query: `CREATE FUNCTION f_for_scalar_loop() RETURNS text LANGUAGE plpgsql AS $$
DECLARE v int; result text := '';
BEGIN
	FOR v IN SELECT 1 LOOP
		result := result || v;
	END LOOP;
	RETURN result;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_for_scalar_loop();`,
					Expected: []sql.Row{{"1"}},
				},
				{
					// Scalar FOR targets use assignment casts, just like SELECT ... INTO targets.
					Query: `CREATE FUNCTION f_for_scalar_text_loop() RETURNS int LANGUAGE plpgsql AS $$
DECLARE v int; result int := 0;
BEGIN
	FOR v IN SELECT '7'::text LOOP
		result := result + v;
	END LOOP;
	RETURN result;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_for_scalar_text_loop();`,
					Expected: []sql.Row{{7}},
				},
				{
					Query: `CREATE FUNCTION f_for_scalar_invalid_loop() RETURNS int LANGUAGE plpgsql AS $$
DECLARE v int;
BEGIN
	FOR v IN SELECT 'not-an-integer'::text LOOP
		RETURN v;
	END LOOP;
	RETURN 0;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:           `SELECT f_for_scalar_invalid_loop();`,
					ExpectedErr:     `invalid input syntax for type int4`,
					ExpectedErrCode: `22P02`,
				},
				{
					// The conversion error is a regular SQL error; this uses the same client connection.
					Query:    `SELECT f_for_scalar_loop();`,
					Expected: []sql.Row{{"1"}},
				},
				{
					Query: `CREATE FUNCTION f_for_variable_list_loop() RETURNS text LANGUAGE plpgsql AS $$
DECLARE a int; b int; result text := '';
BEGIN
	FOR a, b IN SELECT 1, 2 LOOP
		result := result || a || ':' || b;
	END LOOP;
	RETURN result;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_for_variable_list_loop();`,
					Expected: []sql.Row{{"1:2"}},
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

// TestPlpgsqlWholeRecordReference covers referencing a RECORD variable as a whole rather than a field of it,
// which is what passing it to a function such as to_jsonb() does.
func TestPlpgsqlWholeRecordReference(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "a RECORD variable referenced as a whole",
			SetUpScript: []string{
				`CREATE TABLE k (id int, name text);`,
				`INSERT INTO k VALUES (1, 'a');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION f_jsonb() RETURNS text LANGUAGE plpgsql AS $$
DECLARE r RECORD;
BEGIN
	SELECT * INTO r FROM k WHERE id = 1;
	RETURN to_jsonb(r)::text;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_jsonb();`,
					Expected: []sql.Row{{`{"id": 1, "name": "a"}`}},
				},
				{
					Query: `CREATE FUNCTION f_text() RETURNS text LANGUAGE plpgsql AS $$
DECLARE r RECORD;
BEGIN
	SELECT id, name AS renamed INTO r FROM k WHERE id = 1;
	RETURN r::text;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_text();`,
					Expected: []sql.Row{{`(1,a)`}},
				},
				{
					Query: `CREATE FUNCTION f_renamed() RETURNS text LANGUAGE plpgsql AS $$
DECLARE r RECORD;
BEGIN
	SELECT id, name AS renamed INTO r FROM k WHERE id = 1;
	RETURN to_jsonb(r)::text;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_renamed();`,
					Expected: []sql.Row{{`{"id": 1, "renamed": "a"}`}},
				},
				{
					Query: `CREATE FUNCTION f_unassigned_whole() RETURNS text LANGUAGE plpgsql AS $$
DECLARE r RECORD;
BEGIN
	RETURN to_jsonb(r)::text;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT f_unassigned_whole();`,
					ExpectedErr: `record "r" is not assigned yet`,
				},
			},
		},
	})
}
