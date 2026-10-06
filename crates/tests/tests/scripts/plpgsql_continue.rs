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
fn test_plpgsql_continue() {
    run_scripts(&[
        ScriptTest {
            name: "CONTINUE WHEN in an integer FOR loop",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_fori_when() RETURNS int LANGUAGE plpgsql AS $$
DECLARE i int; n int := 0; guard int := 0;
BEGIN
	FOR i IN 1..5 LOOP
		guard := guard + 1;
		IF guard > 20 THEN RETURN -99; END IF;
		CONTINUE WHEN i % 2 = 0;
		n := n + i;
	END LOOP;
	RETURN n;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_fori_when();",
                    expected: Expected::Rows {
                        columns: &[Column("f_fori_when", INT4)],
                        rows: &[
                            &[T("9")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "bare CONTINUE in an integer FOR loop",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_fori_bare() RETURNS int LANGUAGE plpgsql AS $$
DECLARE i int; n int := 0; guard int := 0;
BEGIN
	FOR i IN 1..5 LOOP
		guard := guard + 1;
		IF guard > 20 THEN RETURN -99; END IF;
		IF i = 3 THEN
			CONTINUE;
		END IF;
		n := n + i;
	END LOOP;
	RETURN n;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_fori_bare();",
                    expected: Expected::Rows {
                        columns: &[Column("f_fori_bare", INT4)],
                        rows: &[
                            &[T("12")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CONTINUE preserves the integer FOR loop variable",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_fori_var() RETURNS text LANGUAGE plpgsql AS $$
DECLARE i int; acc text := ''; guard int := 0;
BEGIN
	FOR i IN 1..4 LOOP
		guard := guard + 1;
		IF guard > 20 THEN RETURN 'guard'; END IF;
		CONTINUE WHEN i = 2;
		acc := acc || i::text;
	END LOOP;
	RETURN acc;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_fori_var();",
                    expected: Expected::Rows {
                        columns: &[Column("f_fori_var", TEXT)],
                        rows: &[
                            &[T("134")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CONTINUE as the last statement of an integer FOR loop body",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_fori_last() RETURNS int LANGUAGE plpgsql AS $$
DECLARE i int; n int := 0; guard int := 0;
BEGIN
	FOR i IN 1..4 LOOP
		guard := guard + 1;
		IF guard > 20 THEN RETURN -99; END IF;
		n := n + i;
		CONTINUE WHEN i > 0;
	END LOOP;
	RETURN n;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_fori_last();",
                    expected: Expected::Rows {
                        columns: &[Column("f_fori_last", INT4)],
                        rows: &[
                            &[T("10")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CONTINUE in an integer FOR loop with BY",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_fori_by() RETURNS int LANGUAGE plpgsql AS $$
DECLARE i int; n int := 0; guard int := 0;
BEGIN
	FOR i IN 1..10 BY 3 LOOP
		guard := guard + 1;
		IF guard > 20 THEN RETURN -99; END IF;
		CONTINUE WHEN i = 4;
		n := n + i;
	END LOOP;
	RETURN n;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_fori_by();",
                    expected: Expected::Rows {
                        columns: &[Column("f_fori_by", INT4)],
                        rows: &[
                            &[T("18")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CONTINUE in a REVERSE integer FOR loop with BY",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_fori_reverse() RETURNS int LANGUAGE plpgsql AS $$
DECLARE i int; n int := 0; guard int := 0;
BEGIN
	FOR i IN REVERSE 10..1 BY 3 LOOP
		guard := guard + 1;
		IF guard > 20 THEN RETURN -99; END IF;
		CONTINUE WHEN i = 7;
		n := n + i;
	END LOOP;
	RETURN n;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_fori_reverse();",
                    expected: Expected::Rows {
                        columns: &[Column("f_fori_reverse", INT4)],
                        rows: &[
                            &[T("15")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CONTINUE and EXIT in the same integer FOR loop",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_fori_exit() RETURNS int LANGUAGE plpgsql AS $$
DECLARE i int; n int := 0; guard int := 0;
BEGIN
	FOR i IN 1..10 LOOP
		guard := guard + 1;
		IF guard > 25 THEN RETURN -99; END IF;
		CONTINUE WHEN i % 2 = 0;
		EXIT WHEN i > 6;
		n := n + i;
	END LOOP;
	RETURN n;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_fori_exit();",
                    expected: Expected::Rows {
                        columns: &[Column("f_fori_exit", INT4)],
                        rows: &[
                            &[T("9")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "bare CONTINUE in a nested integer FOR loop",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_fori_nested() RETURNS int LANGUAGE plpgsql AS $$
DECLARE i int; j int; n int := 0; guard int := 0;
BEGIN
	FOR i IN 1..3 LOOP
		FOR j IN 1..3 LOOP
			guard := guard + 1;
			IF guard > 30 THEN RETURN -99; END IF;
			CONTINUE WHEN j = 2;
			n := n + 1;
		END LOOP;
	END LOOP;
	RETURN n;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_fori_nested();",
                    expected: Expected::Rows {
                        columns: &[Column("f_fori_nested", INT4)],
                        rows: &[
                            &[T("6")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "labelled CONTINUE of an outer integer FOR loop",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_fori_label() RETURNS int LANGUAGE plpgsql AS $$
DECLARE i int; j int; n int := 0; guard int := 0;
BEGIN
	<<outer_loop>>
	FOR i IN 1..3 LOOP
		FOR j IN 1..3 LOOP
			guard := guard + 1;
			IF guard > 30 THEN RETURN -99; END IF;
			CONTINUE outer_loop WHEN j = 2;
			n := n + (i * 10 + j);
		END LOOP;
	END LOOP;
	RETURN n;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_fori_label();",
                    expected: Expected::Rows {
                        columns: &[Column("f_fori_label", INT4)],
                        rows: &[
                            &[T("63")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CONTINUE in a WHILE loop",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_while() RETURNS int LANGUAGE plpgsql AS $$
DECLARE i int := 0; n int := 0; guard int := 0;
BEGIN
	WHILE i < 5 LOOP
		i := i + 1;
		guard := guard + 1;
		IF guard > 20 THEN RETURN -99; END IF;
		CONTINUE WHEN i = 3;
		n := n + i;
	END LOOP;
	RETURN n;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_while();",
                    expected: Expected::Rows {
                        columns: &[Column("f_while", INT4)],
                        rows: &[
                            &[T("12")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "labelled CONTINUE of an outer WHILE loop",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_while_label() RETURNS int LANGUAGE plpgsql AS $$
DECLARE i int := 0; j int; n int := 0; guard int := 0;
BEGIN
	<<outer_loop>>
	WHILE i < 3 LOOP
		i := i + 1;
		FOR j IN 1..3 LOOP
			guard := guard + 1;
			IF guard > 30 THEN RETURN -99; END IF;
			CONTINUE outer_loop WHEN j = 2;
			n := n + (i * 10 + j);
		END LOOP;
	END LOOP;
	RETURN n;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_while_label();",
                    expected: Expected::Rows {
                        columns: &[Column("f_while_label", INT4)],
                        rows: &[
                            &[T("63")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CONTINUE in a plain LOOP",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_loop() RETURNS int LANGUAGE plpgsql AS $$
DECLARE i int := 0; n int := 0; guard int := 0;
BEGIN
	LOOP
		i := i + 1;
		EXIT WHEN i > 5;
		guard := guard + 1;
		IF guard > 20 THEN RETURN -99; END IF;
		CONTINUE WHEN i = 3;
		n := n + i;
	END LOOP;
	RETURN n;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_loop();",
                    expected: Expected::Rows {
                        columns: &[Column("f_loop", INT4)],
                        rows: &[
                            &[T("12")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CONTINUE WHEN in a FOR..IN..SELECT loop",
            set_up_script: &[
                "CREATE TABLE c1 (id int, val int);",
                "INSERT INTO c1 VALUES (1, 10), (2, 20), (3, 30);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_fors_when() RETURNS int LANGUAGE plpgsql AS $$
DECLARE r RECORD; n int := 0; guard int := 0;
BEGIN
	FOR r IN SELECT id, val FROM c1 ORDER BY id LOOP
		guard := guard + 1;
		IF guard > 20 THEN RETURN -99; END IF;
		CONTINUE WHEN r.id = 2;
		n := n + r.val;
	END LOOP;
	RETURN n;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_fors_when();",
                    expected: Expected::Rows {
                        columns: &[Column("f_fors_when", INT4)],
                        rows: &[
                            &[T("40")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "bare CONTINUE in a FOR..IN..SELECT loop",
            set_up_script: &[
                "CREATE TABLE c2 (id int, val int);",
                "INSERT INTO c2 VALUES (1, 10), (2, 20), (3, 30);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_fors_bare() RETURNS int LANGUAGE plpgsql AS $$
DECLARE r RECORD; n int := 0; guard int := 0;
BEGIN
	FOR r IN SELECT id, val FROM c2 ORDER BY id LOOP
		guard := guard + 1;
		IF guard > 20 THEN RETURN -99; END IF;
		IF r.id = 3 THEN
			CONTINUE;
		END IF;
		n := n + r.val;
	END LOOP;
	RETURN n;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_fors_bare();",
                    expected: Expected::Rows {
                        columns: &[Column("f_fors_bare", INT4)],
                        rows: &[
                            &[T("30")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CONTINUE and EXIT in the same FOR..IN..SELECT loop",
            set_up_script: &[
                "CREATE TABLE c3 (id int, val int);",
                "INSERT INTO c3 VALUES (1, 10), (2, 20), (3, 30), (4, 40), (5, 50);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_fors_exit() RETURNS int LANGUAGE plpgsql AS $$
DECLARE r RECORD; n int := 0; guard int := 0;
BEGIN
	FOR r IN SELECT id, val FROM c3 ORDER BY id LOOP
		guard := guard + 1;
		IF guard > 20 THEN RETURN -99; END IF;
		CONTINUE WHEN r.id = 2;
		EXIT WHEN r.id = 4;
		n := n + r.val;
	END LOOP;
	RETURN n;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_fors_exit();",
                    expected: Expected::Rows {
                        columns: &[Column("f_fors_exit", INT4)],
                        rows: &[
                            &[T("40")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "labelled CONTINUE of an outer FOR..IN..SELECT loop",
            set_up_script: &[
                "CREATE TABLE c4 (id int, val int);",
                "INSERT INTO c4 VALUES (1, 10), (2, 20), (3, 30);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_fors_label() RETURNS int LANGUAGE plpgsql AS $$
DECLARE r RECORD; s RECORD; n int := 0; guard int := 0;
BEGIN
	<<outer_loop>>
	FOR r IN SELECT id, val FROM c4 ORDER BY id LOOP
		FOR s IN SELECT id FROM c4 ORDER BY id LOOP
			guard := guard + 1;
			IF guard > 30 THEN RETURN -99; END IF;
			CONTINUE outer_loop WHEN s.id = 2;
			n := n + (r.id * 10 + s.id);
		END LOOP;
	END LOOP;
	RETURN n;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_fors_label();",
                    expected: Expected::Rows {
                        columns: &[Column("f_fors_label", INT4)],
                        rows: &[
                            &[T("63")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "labelled CONTINUE of an outer integer FOR loop from a FOR..IN..SELECT loop",
            set_up_script: &[
                "CREATE TABLE c5 (id int);",
                "INSERT INTO c5 VALUES (1), (2), (3);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_mixed_label() RETURNS int LANGUAGE plpgsql AS $$
DECLARE i int; r RECORD; n int := 0; guard int := 0;
BEGIN
	<<outer_loop>>
	FOR i IN 1..3 LOOP
		FOR r IN SELECT id FROM c5 ORDER BY id LOOP
			guard := guard + 1;
			IF guard > 30 THEN RETURN -99; END IF;
			CONTINUE outer_loop WHEN r.id = 2;
			n := n + (i * 10 + r.id);
		END LOOP;
	END LOOP;
	RETURN n;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_mixed_label();",
                    expected: Expected::Rows {
                        columns: &[Column("f_mixed_label", INT4)],
                        rows: &[
                            &[T("63")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CONTINUE as the last statement of a FOR..IN..SELECT loop body",
            set_up_script: &[
                "CREATE TABLE c6 (id int, val int);",
                "INSERT INTO c6 VALUES (1, 10), (2, 20), (3, 30);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_fors_last() RETURNS int LANGUAGE plpgsql AS $$
DECLARE r RECORD; n int := 0; guard int := 0;
BEGIN
	FOR r IN SELECT id, val FROM c6 ORDER BY id LOOP
		guard := guard + 1;
		IF guard > 20 THEN RETURN -99; END IF;
		n := n + r.val;
		CONTINUE WHEN r.id > 0;
	END LOOP;
	RETURN n;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_fors_last();",
                    expected: Expected::Rows {
                        columns: &[Column("f_fors_last", INT4)],
                        rows: &[
                            &[T("60")],
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
