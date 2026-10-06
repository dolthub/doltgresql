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

use harness::oid::*;
use harness::pgx::Time;
use harness::plan::PlanFact;
use harness::script::Cell::{Any, Null, Text as T};
use harness::script::{A, BindVar, Column, Diagnostic, E, Expected, Flow, N, S, ScriptTest, ScriptTestAssertion, USER_DEFINED, run_scripts, run_scripts_repeated};
use harness::wire::{Datum, F, Field, Fields, PGX_STARTUP, Receive, Send, Step, W, WireTest, run_wire_tests};

#[test]
fn test_plpgsql_record_into() {
    run_scripts(&[
        ScriptTest {
            name: "RECORD declaration default",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_record_default() RETURNS text LANGUAGE plpgsql AS $$
DECLARE r RECORD := ROW(1, 'a');
BEGIN RETURN r::text; END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_record_default();",
                    expected: Expected::Rows {
                        columns: &[Column("f_record_default", TEXT)],
                        rows: &[
                            &[T("(1,a)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "RECORD declaration default from NEW",
            set_up_script: &[
                "CREATE TABLE src (id int, note text);",
                "CREATE TABLE res (id int, note text);",
                r#"CREATE FUNCTION trg_record_default() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE whole RECORD := NEW;
BEGIN
	INSERT INTO res VALUES (whole.id, whole.note);
	RETURN NEW;
END; $$;"#,
                "CREATE TRIGGER t_record_default AFTER INSERT ON src FOR EACH ROW EXECUTE FUNCTION trg_record_default();",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO src VALUES (1, 'a');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, note FROM res;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("note", TEXT)],
                        rows: &[
                            &[T("1"), T("a")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "RECORD declaration default referencing an earlier variable",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_record_default_local() RETURNS text LANGUAGE plpgsql AS $$
DECLARE n int := 7; r RECORD := ROW(n, 'a');
BEGIN RETURN r::text; END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_record_default_local();",
                    expected: Expected::Rows {
                        columns: &[Column("f_record_default_local", TEXT)],
                        rows: &[
                            &[T("(7,a)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "RECORD declaration default referencing parameters",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_record_default_param(n int, s text) RETURNS text LANGUAGE plpgsql AS $$
DECLARE r RECORD := ROW(n * 2, s);
BEGIN RETURN r::text || '|' || r.f1 || '|' || r.f2; END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_record_default_param(1, 'x'), f_record_default_param(2, 'y');",
                    expected: Expected::Rows {
                        columns: &[Column("f_record_default_param", TEXT), Column("f_record_default_param", TEXT)],
                        rows: &[
                            &[T("(2,x)|2|x"), T("(4,y)|4|y")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "RECORD declaration default referencing an outer block's variable",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_record_default_outer() RETURNS text LANGUAGE plpgsql AS $$
DECLARE n int := 7;
BEGIN
	DECLARE r RECORD := ROW(n, 'a');
	BEGIN RETURN r::text; END;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_record_default_outer();",
                    expected: Expected::Rows {
                        columns: &[Column("f_record_default_outer", TEXT)],
                        rows: &[
                            &[T("(7,a)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "RECORD declaration default referencing a shadowing variable",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_record_default_shadow() RETURNS text LANGUAGE plpgsql AS $$
DECLARE n int := 1;
BEGIN
	DECLARE n int := 2; r RECORD := ROW(n);
	BEGIN RETURN r::text; END;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_record_default_shadow();",
                    expected: Expected::Rows {
                        columns: &[Column("f_record_default_shadow", TEXT)],
                        rows: &[
                            &[T("(2)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "variable declaration default referencing an earlier RECORD",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_variable_default_record() RETURNS text LANGUAGE plpgsql AS $$
DECLARE r RECORD := ROW(1, 'a'); t text := r::text; m int := r.f1 + 1;
BEGIN RETURN t || '|' || m; END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_variable_default_record();",
                    expected: Expected::Rows {
                        columns: &[Column("f_variable_default_record", TEXT)],
                        rows: &[
                            &[T("(1,a)|2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "RECORD declaration default from an earlier RECORD",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_record_default_record() RETURNS text LANGUAGE plpgsql AS $$
DECLARE a RECORD := ROW(1, 'x'::text); b RECORD := a;
BEGIN RETURN b::text || '|' || b.f2; END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_record_default_record();",
                    expected: Expected::Rows {
                        columns: &[Column("f_record_default_record", TEXT)],
                        rows: &[
                            &[T("(1,x)|x")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "RECORD and variable declaration defaults interleaved",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_record_default_chain() RETURNS text LANGUAGE plpgsql AS $$
DECLARE n int := 1; r RECORD := ROW(n); m int := n + 1; s RECORD := ROW(n, m);
BEGIN RETURN r::text || s::text; END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_record_default_chain();",
                    expected: Expected::Rows {
                        columns: &[Column("f_record_default_chain", TEXT)],
                        rows: &[
                            &[T("(1)(1,2)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "RECORD declaration default referencing a later variable",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_record_default_later() RETURNS text LANGUAGE plpgsql AS $$
DECLARE r RECORD := ROW(n); n int := 1;
BEGIN RETURN r::text; END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_record_default_later();",
                    expected: Expected::Error(Diagnostic { code: "42703", message: r#"column "n" does not exist"#, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "RECORD declaration default from NEW and a variable",
            set_up_script: &[
                "CREATE TABLE src (id int, note text);",
                "CREATE TABLE res (id int, note text);",
                r#"CREATE FUNCTION trg_record_default_var() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE bump int := 100; r RECORD := ROW(NEW.id + bump, NEW.note || '!');
BEGIN
	INSERT INTO res VALUES (r.f1, r.f2);
	RETURN NEW;
END; $$;"#,
                "CREATE TRIGGER t_record_default_var AFTER INSERT ON src FOR EACH ROW EXECUTE FUNCTION trg_record_default_var();",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO src VALUES (1, 'a');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, note FROM res;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("note", TEXT)],
                        rows: &[
                            &[T("101"), T("a!")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "SELECT INTO a RECORD variable",
            set_up_script: &[
                "CREATE TABLE k (id int, name text, amt numeric);",
                "INSERT INTO k VALUES (1, 'a', 10.5), (2, 'b', 20.25);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_repro() RETURNS int LANGUAGE plpgsql AS $$
DECLARE r RECORD;
BEGIN SELECT id INTO r FROM k LIMIT 1; RETURN 1; END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_repro();",
                    expected: Expected::Rows {
                        columns: &[Column("f_repro", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_field_text(p int) RETURNS text LANGUAGE plpgsql AS $$
DECLARE r RECORD;
BEGIN
	SELECT id, name INTO r FROM k WHERE id = p;
	RETURN r.name;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_field_text(2);",
                    expected: Expected::Rows {
                        columns: &[Column("f_field_text", TEXT)],
                        rows: &[
                            &[T("b")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_field_int(p int) RETURNS int LANGUAGE plpgsql AS $$
DECLARE r RECORD;
BEGIN
	SELECT id, name INTO r FROM k WHERE id = p;
	RETURN r.id;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_field_int(2);",
                    expected: Expected::Rows {
                        columns: &[Column("f_field_int", INT4)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_star() RETURNS numeric LANGUAGE plpgsql AS $$
DECLARE r RECORD;
BEGIN
	SELECT * INTO r FROM k WHERE id = 1;
	RETURN r.amt;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_star();",
                    expected: Expected::Rows {
                        columns: &[Column("f_star", NUMERIC)],
                        rows: &[
                            &[T("10.5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_agg() RETURNS bigint LANGUAGE plpgsql AS $$
DECLARE r RECORD;
BEGIN
	SELECT count(*) AS c INTO r FROM k;
	RETURN r.c;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_agg();",
                    expected: Expected::Rows {
                        columns: &[Column("f_agg", INT8)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_nomatch() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE r RECORD;
BEGIN
	SELECT id, name INTO r FROM k WHERE id = 999;
	RETURN r.id IS NULL;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_nomatch();",
                    expected: Expected::Rows {
                        columns: &[Column("f_nomatch", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_reshape() RETURNS text LANGUAGE plpgsql AS $$
DECLARE r RECORD;
BEGIN
	SELECT id, name INTO r FROM k WHERE id = 1;
	SELECT name AS other INTO r FROM k WHERE id = 2;
	RETURN r.other;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_reshape();",
                    expected: Expected::Rows {
                        columns: &[Column("f_reshape", TEXT)],
                        rows: &[
                            &[T("b")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "INSERT RETURNING INTO a RECORD variable",
            set_up_script: &[
                "CREATE TABLE k (id int, name text);",
                "INSERT INTO k VALUES (1, 'a');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_insert_returning() RETURNS text LANGUAGE plpgsql AS $$
DECLARE r RECORD;
BEGIN
	INSERT INTO k VALUES (3, 'c') RETURNING id, name INTO r;
	RETURN r.name;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_insert_returning();",
                    expected: Expected::Rows {
                        columns: &[Column("f_insert_returning", TEXT)],
                        rows: &[
                            &[T("c")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, name FROM k ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("name", TEXT)],
                        rows: &[
                            &[T("1"), T("a")],
                            &[T("3"), T("c")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "EXECUTE INTO a RECORD variable",
            set_up_script: &[
                "CREATE TABLE k (id int, name text);",
                "INSERT INTO k VALUES (1, 'a'), (2, 'b');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_exec() RETURNS text LANGUAGE plpgsql AS $$
DECLARE r RECORD;
BEGIN
	EXECUTE 'SELECT id, name FROM k WHERE id = 2' INTO r;
	RETURN r.name;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_exec();",
                    expected: Expected::Rows {
                        columns: &[Column("f_exec", TEXT)],
                        rows: &[
                            &[T("b")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "trigger function using SELECT INTO a RECORD variable",
            set_up_script: &[
                "CREATE TABLE t (id int primary key, v int);",
                r#"CREATE FUNCTION trg_guard() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE blocker RECORD;
BEGIN
	SELECT id, v INTO blocker FROM t WHERE v > NEW.v LIMIT 1;
	IF blocker.id IS NOT NULL THEN
		RAISE EXCEPTION 'blocked by row %', blocker.id;
	END IF;
	RETURN NEW;
END; $$;"#,
                "CREATE TRIGGER trg BEFORE INSERT ON t FOR EACH ROW EXECUTE FUNCTION trg_guard();",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (1, 10);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (2, 5);",
                    expected: Expected::Error(Diagnostic { code: "P0001", message: "blocked by row 1", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, v FROM t ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", INT4)],
                        rows: &[
                            &[T("1"), T("10")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "RECORD fields written as quoted identifiers",
            set_up_script: &[
                r#"CREATE TABLE k3 ("id" int, "book_date" date);"#,
                "INSERT INTO k3 VALUES (7, '2026-01-02'), (9, '2026-03-04');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_quoted() RETURNS int LANGUAGE plpgsql AS $$
DECLARE blocker RECORD;
BEGIN
	SELECT b."id", b."book_date" INTO blocker FROM k3 b ORDER BY b."book_date" LIMIT 1;
	IF blocker."id" IS NOT NULL THEN
		RETURN blocker."id";
	END IF;
	RETURN -1;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_quoted();",
                    expected: Expected::Rows {
                        columns: &[Column("f_quoted", INT4)],
                        rows: &[
                            &[T("7")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_raise() RETURNS void LANGUAGE plpgsql AS $$
DECLARE r RECORD;
BEGIN
	SELECT "id" INTO r FROM k3 ORDER BY "id" LIMIT 1;
	RAISE EXCEPTION 'saw %', r."id";
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_raise();",
                    expected: Expected::Error(Diagnostic { code: "P0001", message: "saw 7", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "FOR..IN..SELECT over a RECORD variable",
            set_up_script: &[
                "CREATE TABLE k4 (id int, name text);",
                "INSERT INTO k4 VALUES (1, 'a'), (2, 'b'), (3, 'c');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_forloop() RETURNS text LANGUAGE plpgsql AS $$
DECLARE r RECORD; acc text := '';
BEGIN
	FOR r IN SELECT id, name FROM k4 ORDER BY id LOOP
		acc := acc || r.id || r.name;
	END LOOP;
	RETURN acc;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_forloop();",
                    expected: Expected::Rows {
                        columns: &[Column("f_forloop", TEXT)],
                        rows: &[
                            &[T("1a2b3c")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "SELECT INTO a RECORD from a derived table",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_derived() RETURNS numeric LANGUAGE plpgsql AS $$
DECLARE fig RECORD;
BEGIN
	SELECT * INTO fig FROM (SELECT 12.5::numeric AS computed_balance, 3::bigint AS line_count) f;
	RETURN fig.computed_balance;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_derived();",
                    expected: Expected::Rows {
                        columns: &[Column("f_derived", NUMERIC)],
                        rows: &[
                            &[T("12.5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "errors accessing RECORD fields",
            set_up_script: &[
                "CREATE TABLE k (id int, name text);",
                "INSERT INTO k VALUES (1, 'a');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_badfield() RETURNS text LANGUAGE plpgsql AS $$
DECLARE r RECORD;
BEGIN
	SELECT id INTO r FROM k WHERE id = 1;
	RETURN r.nope;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_badfield();",
                    expected: Expected::Error(Diagnostic { code: "42703", message: r#"record "r" has no field "nope""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_unassigned() RETURNS text LANGUAGE plpgsql AS $$
DECLARE r RECORD;
BEGIN
	RETURN r.id;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_unassigned();",
                    expected: Expected::Error(Diagnostic { code: "55000", message: r#"record "r" is not assigned yet"#, detail: "The tuple structure of a not-yet-assigned record is indeterminate.", ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_plpgsql_whole_record_reference() {
    run_scripts(&[
        ScriptTest {
            name: "a RECORD variable referenced as a whole",
            set_up_script: &[
                "CREATE TABLE k (id int, name text);",
                "INSERT INTO k VALUES (1, 'a');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_jsonb() RETURNS text LANGUAGE plpgsql AS $$
DECLARE r RECORD;
BEGIN
	SELECT * INTO r FROM k WHERE id = 1;
	RETURN to_jsonb(r)::text;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_jsonb();",
                    expected: Expected::Rows {
                        columns: &[Column("f_jsonb", TEXT)],
                        rows: &[
                            &[T(r#"{"id": 1, "name": "a"}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_text() RETURNS text LANGUAGE plpgsql AS $$
DECLARE r RECORD;
BEGIN
	SELECT id, name AS renamed INTO r FROM k WHERE id = 1;
	RETURN r::text;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_text();",
                    expected: Expected::Rows {
                        columns: &[Column("f_text", TEXT)],
                        rows: &[
                            &[T("(1,a)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_renamed() RETURNS text LANGUAGE plpgsql AS $$
DECLARE r RECORD;
BEGIN
	SELECT id, name AS renamed INTO r FROM k WHERE id = 1;
	RETURN to_jsonb(r)::text;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_renamed();",
                    expected: Expected::Rows {
                        columns: &[Column("f_renamed", TEXT)],
                        rows: &[
                            &[T(r#"{"id": 1, "renamed": "a"}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_unassigned_whole() RETURNS text LANGUAGE plpgsql AS $$
DECLARE r RECORD;
BEGIN
	RETURN to_jsonb(r)::text;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_unassigned_whole();",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
            ],
            ..S
        },
    ]);
}
