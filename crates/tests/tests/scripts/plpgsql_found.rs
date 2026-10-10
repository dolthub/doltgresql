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
fn test_plpgsql_found() {
    run_scripts(&[
        ScriptTest {
            name: "FOUND starts out false",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_init() RETURNS boolean LANGUAGE plpgsql AS $$
BEGIN RETURN FOUND; END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_init();",
                    expected: Expected::Rows {
                        columns: &[Column("f_init", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "SELECT INTO sets FOUND",
            set_up_script: &[
                "CREATE TABLE k (id int, nm text);",
                "INSERT INTO k VALUES (1, 'a'), (2, 'b');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_hit() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE v int;
BEGIN SELECT id INTO v FROM k WHERE id = 1; RETURN FOUND; END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_hit();",
                    expected: Expected::Rows {
                        columns: &[Column("f_hit", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_miss() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE v int;
BEGIN SELECT id INTO v FROM k WHERE id = 99; RETURN FOUND; END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_miss();",
                    expected: Expected::Rows {
                        columns: &[Column("f_miss", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_lower() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE v int;
BEGIN SELECT id INTO v FROM k WHERE id = 99; RETURN found; END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_lower();",
                    expected: Expected::Rows {
                        columns: &[Column("f_lower", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_rec_miss() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE r RECORD;
BEGIN SELECT id INTO r FROM k WHERE id = 99; RETURN FOUND; END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_rec_miss();",
                    expected: Expected::Rows {
                        columns: &[Column("f_rec_miss", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_rec_hit() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE r RECORD;
BEGIN SELECT id INTO r FROM k WHERE id = 1; RETURN FOUND; END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_rec_hit();",
                    expected: Expected::Rows {
                        columns: &[Column("f_rec_hit", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_guard(p int) RETURNS text LANGUAGE plpgsql AS $$
DECLARE r RECORD;
BEGIN
	SELECT id, nm INTO r FROM k WHERE id = p;
	IF NOT FOUND THEN RETURN 'absent'; END IF;
	RETURN r.nm;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_guard(2);",
                    expected: Expected::Rows {
                        columns: &[Column("f_guard", TEXT)],
                        rows: &[
                            &[T("b")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_guard(99);",
                    expected: Expected::Rows {
                        columns: &[Column("f_guard", TEXT)],
                        rows: &[
                            &[T("absent")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "statements that leave FOUND alone",
            set_up_script: &[
                "CREATE TABLE k (id int);",
                "INSERT INTO k VALUES (1);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_assign() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE v int;
BEGIN SELECT id INTO v FROM k WHERE id = 1; v := 42; RETURN FOUND; END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_assign();",
                    expected: Expected::Rows {
                        columns: &[Column("f_assign", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_exec_hit() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE v int;
BEGIN EXECUTE 'SELECT id FROM k WHERE id = 1' INTO v; RETURN FOUND; END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_exec_hit();",
                    expected: Expected::Rows {
                        columns: &[Column("f_exec_hit", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_ddl() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE v int;
BEGIN
	SELECT id INTO v FROM k WHERE id = 1;
	CREATE TABLE f_ddl_scratch (a int);
	RETURN FOUND;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_ddl();",
                    expected: Expected::Rows {
                        columns: &[Column("f_ddl", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_ddl_miss() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE v int;
BEGIN
	SELECT id INTO v FROM k WHERE id = 99;
	CREATE TABLE f_ddl_scratch2 (a int);
	RETURN FOUND;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_ddl_miss();",
                    expected: Expected::Rows {
                        columns: &[Column("f_ddl_miss", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_cte_dml() RETURNS boolean LANGUAGE plpgsql AS $$
BEGIN
	WITH src AS (SELECT 99 AS id) INSERT INTO k SELECT id FROM src;
	RETURN FOUND;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_cte_dml();",
                    expected: Expected::Rows {
                        columns: &[Column("f_cte_dml", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "PERFORM and data-modifying statements set FOUND",
            set_up_script: &[
                "CREATE TABLE k (id int);",
                "INSERT INTO k VALUES (1), (2), (3);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_perf_hit() RETURNS boolean LANGUAGE plpgsql AS $$
BEGIN PERFORM 1 FROM k WHERE id = 1; RETURN FOUND; END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_perf_hit();",
                    expected: Expected::Rows {
                        columns: &[Column("f_perf_hit", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_perf_miss() RETURNS boolean LANGUAGE plpgsql AS $$
BEGIN PERFORM 1 FROM k WHERE id = 99; RETURN FOUND; END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_perf_miss();",
                    expected: Expected::Rows {
                        columns: &[Column("f_perf_miss", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_del_miss() RETURNS boolean LANGUAGE plpgsql AS $$
BEGIN DELETE FROM k WHERE id = 99; RETURN FOUND; END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_del_miss();",
                    expected: Expected::Rows {
                        columns: &[Column("f_del_miss", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_del_hit() RETURNS boolean LANGUAGE plpgsql AS $$
BEGIN DELETE FROM k WHERE id = 3; RETURN FOUND; END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_del_hit();",
                    expected: Expected::Rows {
                        columns: &[Column("f_del_hit", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_upd_miss() RETURNS boolean LANGUAGE plpgsql AS $$
BEGIN UPDATE k SET id = id WHERE id = 99; RETURN FOUND; END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_upd_miss();",
                    expected: Expected::Rows {
                        columns: &[Column("f_upd_miss", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_ins_plain() RETURNS boolean LANGUAGE plpgsql AS $$
BEGIN INSERT INTO k VALUES (10); RETURN FOUND; END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_ins_plain();",
                    expected: Expected::Rows {
                        columns: &[Column("f_ins_plain", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_ins_returning() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE v int;
BEGIN INSERT INTO k VALUES (11) RETURNING id INTO v; RETURN FOUND; END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_ins_returning();",
                    expected: Expected::Rows {
                        columns: &[Column("f_ins_returning", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "FOR..IN..SELECT sets FOUND when the loop exits",
            set_up_script: &[
                "CREATE TABLE k (id int);",
                "INSERT INTO k VALUES (1), (2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_for_hit() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE r RECORD;
BEGIN FOR r IN SELECT id FROM k LOOP END LOOP; RETURN FOUND; END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_for_hit();",
                    expected: Expected::Rows {
                        columns: &[Column("f_for_hit", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_for_miss() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE r RECORD;
BEGIN FOR r IN SELECT id FROM k WHERE id = 99 LOOP END LOOP; RETURN FOUND; END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_for_miss();",
                    expected: Expected::Rows {
                        columns: &[Column("f_for_miss", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_in_loop() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE r RECORD; res boolean;
BEGIN FOR r IN SELECT id FROM k LOOP res := FOUND; END LOOP; RETURN res; END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_in_loop();",
                    expected: Expected::Rows {
                        columns: &[Column("f_in_loop", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_for_exit() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE r RECORD; n int := 0;
BEGIN
	FOR r IN SELECT id FROM k ORDER BY id LOOP n := n + 1; EXIT; END LOOP;
	RETURN FOUND;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_for_exit();",
                    expected: Expected::Rows {
                        columns: &[Column("f_for_exit", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_for_exit_when() RETURNS int LANGUAGE plpgsql AS $$
DECLARE r RECORD; n int := 0;
BEGIN
	FOR r IN SELECT id FROM k ORDER BY id LOOP
		n := n + 1;
		EXIT WHEN r.id = 1;
	END LOOP;
	IF NOT FOUND THEN RETURN -1; END IF;
	RETURN n;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_for_exit_when();",
                    expected: Expected::Rows {
                        columns: &[Column("f_for_exit_when", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_labelled_exit() RETURNS int LANGUAGE plpgsql AS $$
DECLARE r RECORD; s RECORD; n int := 0;
BEGIN
	<<outer>>
	FOR r IN SELECT id FROM k ORDER BY id LOOP
		FOR s IN SELECT id FROM k ORDER BY id LOOP
			n := n + 1;
			EXIT outer;
		END LOOP;
	END LOOP;
	IF NOT FOUND THEN RETURN -1; END IF;
	RETURN n;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_labelled_exit();",
                    expected: Expected::Rows {
                        columns: &[Column("f_labelled_exit", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_reentered_loop() RETURNS int LANGUAGE plpgsql AS $$
DECLARE r RECORD; i int; n int := 0;
BEGIN
	FOR i IN 1..2 LOOP
		FOR r IN SELECT id FROM k ORDER BY id LOOP n := n + 1; END LOOP;
	END LOOP;
	RETURN n;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_reentered_loop();",
                    expected: Expected::Rows {
                        columns: &[Column("f_reentered_loop", INT4)],
                        rows: &[
                            &[T("4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "integer FOR loops set FOUND when they exit",
            set_up_script: &[
                "CREATE TABLE k (id int);",
                "INSERT INTO k VALUES (1);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_fori_hit() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE i int;
BEGIN FOR i IN 1..3 LOOP END LOOP; RETURN FOUND; END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_fori_hit();",
                    expected: Expected::Rows {
                        columns: &[Column("f_fori_hit", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_fori_miss() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE i int;
BEGIN FOR i IN 1..0 LOOP END LOOP; RETURN FOUND; END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_fori_miss();",
                    expected: Expected::Rows {
                        columns: &[Column("f_fori_miss", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_fori_overwrites() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE i int; v int;
BEGIN
	SELECT id INTO v FROM k WHERE id = 1;
	FOR i IN 1..0 LOOP END LOOP;
	RETURN FOUND;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_fori_overwrites();",
                    expected: Expected::Rows {
                        columns: &[Column("f_fori_overwrites", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_fori_reverse() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE i int;
BEGIN FOR i IN REVERSE 3..1 LOOP END LOOP; RETURN FOUND; END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_fori_reverse();",
                    expected: Expected::Rows {
                        columns: &[Column("f_fori_reverse", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_fori_exit() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE i int; n int := 0;
BEGIN FOR i IN 1..3 LOOP n := n + 1; EXIT; END LOOP; RETURN FOUND; END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_fori_exit();",
                    expected: Expected::Rows {
                        columns: &[Column("f_fori_exit", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_fori_in_loop() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE i int; res boolean;
BEGIN FOR i IN 1..2 LOOP res := FOUND; END LOOP; RETURN res; END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_fori_in_loop();",
                    expected: Expected::Rows {
                        columns: &[Column("f_fori_in_loop", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_fori_nested() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE i int; j int;
BEGIN
	FOR i IN 1..2 LOOP
		FOR j IN 1..0 LOOP END LOOP;
	END LOOP;
	RETURN FOUND;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_fori_nested();",
                    expected: Expected::Rows {
                        columns: &[Column("f_fori_nested", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "WHILE and plain LOOP leave FOUND alone",
            set_up_script: &[
                "CREATE TABLE k (id int);",
                "INSERT INTO k VALUES (1);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_while_keeps() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE v int; i int := 0;
BEGIN
	SELECT id INTO v FROM k WHERE id = 1;
	WHILE i < 3 LOOP i := i + 1; END LOOP;
	RETURN FOUND;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_while_keeps();",
                    expected: Expected::Rows {
                        columns: &[Column("f_while_keeps", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_while_no_body() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE v int;
BEGIN
	SELECT id INTO v FROM k WHERE id = 99;
	WHILE false LOOP NULL; END LOOP;
	RETURN FOUND;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_while_no_body();",
                    expected: Expected::Rows {
                        columns: &[Column("f_while_no_body", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_loop_keeps() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE v int;
BEGIN
	SELECT id INTO v FROM k WHERE id = 1;
	LOOP EXIT; END LOOP;
	RETURN FOUND;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_loop_keeps();",
                    expected: Expected::Rows {
                        columns: &[Column("f_loop_keeps", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "RETURN QUERY sets FOUND",
            set_up_script: &[
                "CREATE TABLE k (id int);",
                "INSERT INTO k VALUES (1), (2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_rq_miss() RETURNS SETOF int LANGUAGE plpgsql AS $$
BEGIN
	RETURN QUERY SELECT id FROM k WHERE id = 99;
	IF NOT FOUND THEN RAISE EXCEPTION 'no rows'; END IF;
	RAISE EXCEPTION 'had rows';
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM f_rq_miss();",
                    expected: Expected::Error(Diagnostic { code: "P0001", message: "no rows", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_rq_hit() RETURNS SETOF int LANGUAGE plpgsql AS $$
BEGIN
	RETURN QUERY SELECT id FROM k ORDER BY id;
	IF NOT FOUND THEN RAISE EXCEPTION 'no rows'; END IF;
	RAISE EXCEPTION 'had rows';
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM f_rq_hit();",
                    expected: Expected::Error(Diagnostic { code: "P0001", message: "had rows", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "trigger combining a RECORD target with a NOT FOUND guard",
            set_up_script: &[
                r#"CREATE TABLE reporting_exception (
	id int PRIMARY KEY, org_id int, register_key text, entry_key text,
	version int, superseded_by_id int);"#,
                r#"CREATE FUNCTION resc() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE
	successor RECORD;
BEGIN
	IF NEW."superseded_by_id" IS NULL THEN
		RETURN NULL;
	END IF;

	SELECT "register_key", "entry_key", "version"
	  INTO successor
	  FROM "reporting_exception"
	 WHERE "id" = NEW."superseded_by_id"
	   AND "org_id" = NEW."org_id";

	IF NOT FOUND THEN
		RETURN NULL;
	END IF;

	IF successor."version" <= NEW."version" THEN
		RAISE EXCEPTION 'reporting_exception %: successor must carry a HIGHER version (this row is version %, successor is version %)',
			NEW."id", NEW."version", successor."version";
	END IF;

	IF successor."register_key" IS DISTINCT FROM NEW."register_key"
	   OR successor."entry_key" IS DISTINCT FROM NEW."entry_key" THEN
		RAISE EXCEPTION 'reporting_exception %: successor must supersede the SAME entry', NEW."id";
	END IF;

	RETURN NULL;
END; $$;"#,
                "CREATE TRIGGER trg AFTER INSERT ON reporting_exception FOR EACH ROW EXECUTE FUNCTION resc();",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO reporting_exception VALUES (1, 7, 'rk', 'ek', 1, NULL);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO reporting_exception VALUES (2, 7, 'rk', 'ek', 2, 999);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO reporting_exception VALUES (3, 7, 'rk', 'ek', 9, NULL);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO reporting_exception VALUES (4, 7, 'rk', 'ek', 5, 3);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO reporting_exception VALUES (5, 7, 'rk', 'ek', 50, 3);",
                    expected: Expected::Error(Diagnostic { code: "P0001", message: "reporting_exception 5: successor must carry a HIGHER version (this row is version 50, successor is version 9)", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO reporting_exception VALUES (6, 7, 'other', 'ek', 1, 3);",
                    expected: Expected::Error(Diagnostic { code: "P0001", message: "reporting_exception 6: successor must supersede the SAME entry", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM reporting_exception ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                            &[T("3")],
                            &[T("4")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "FOUND survives a nested block, and can be shadowed",
            set_up_script: &[
                "CREATE TABLE k (id int);",
                "INSERT INTO k VALUES (1);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_nested() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE v int;
BEGIN
	BEGIN SELECT id INTO v FROM k WHERE id = 1; END;
	RETURN FOUND;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_nested();",
                    expected: Expected::Rows {
                        columns: &[Column("f_nested", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_shadow() RETURNS int LANGUAGE plpgsql AS $$
DECLARE found int := 5;
BEGIN RETURN found; END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_shadow();",
                    expected: Expected::Rows {
                        columns: &[Column("f_shadow", INT4)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_param(found boolean) RETURNS boolean LANGUAGE plpgsql AS $$
BEGIN RETURN FOUND; END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_param(true);",
                    expected: Expected::Rows {
                        columns: &[Column("f_param", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "a labelled CONTINUE reports the inner loop's FOUND",
            set_up_script: &[
                "CREATE TABLE cf (id int);",
                "INSERT INTO cf VALUES (1), (2), (3);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_cont_fori() RETURNS text LANGUAGE plpgsql AS $$
DECLARE i int; j int; acc text := '';
BEGIN
	<<outer_loop>>
	FOR i IN 1..3 LOOP
		acc := acc || CASE WHEN FOUND THEN 't' ELSE 'f' END;
		FOR j IN 1..3 LOOP
			CONTINUE outer_loop WHEN j = 1;
		END LOOP;
	END LOOP;
	RETURN acc;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_cont_fori();",
                    expected: Expected::Rows {
                        columns: &[Column("f_cont_fori", TEXT)],
                        rows: &[
                            &[T("ftt")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_cont_fors() RETURNS text LANGUAGE plpgsql AS $$
DECLARE i int; r RECORD; acc text := '';
BEGIN
	<<outer_loop>>
	FOR i IN 1..3 LOOP
		acc := acc || CASE WHEN FOUND THEN 't' ELSE 'f' END;
		FOR r IN SELECT id FROM cf ORDER BY id LOOP
			CONTINUE outer_loop WHEN r.id = 1;
		END LOOP;
	END LOOP;
	RETURN acc;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_cont_fors();",
                    expected: Expected::Rows {
                        columns: &[Column("f_cont_fors", TEXT)],
                        rows: &[
                            &[T("ftt")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_cont_empty() RETURNS text LANGUAGE plpgsql AS $$
DECLARE i int; r RECORD; acc text := '';
BEGIN
	<<outer_loop>>
	FOR i IN 1..3 LOOP
		acc := acc || CASE WHEN FOUND THEN 't' ELSE 'f' END;
		FOR r IN SELECT id FROM cf WHERE id < 0 LOOP
			NULL;
		END LOOP;
		CONTINUE outer_loop;
	END LOOP;
	RETURN acc;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_cont_empty();",
                    expected: Expected::Rows {
                        columns: &[Column("f_cont_empty", TEXT)],
                        rows: &[
                            &[T("fff")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}
