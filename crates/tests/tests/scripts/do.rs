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
fn test_do_statements() {
    run_scripts(&[
        ScriptTest {
            name: "empty block",
            assertions: &[
                ScriptTestAssertion {
                    query: "DO $$ BEGIN NULL; END $$;",
                    expected: Expected::Tag("DO"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "line comments in subqueries",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"DO $$
BEGIN
  IF EXISTS (SELECT 1 -- comment
  ) THEN
    NULL;
  END IF;
END $$;"#,
                    expected: Expected::Tag("DO"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"DO $$
DECLARE v INT := 7;
BEGIN
  IF EXISTS (SELECT 1 -- v must not be substituted in this comment
             WHERE v = 7) THEN
    NULL;
  ELSE
    RAISE EXCEPTION 'expected true';
  END IF;
  IF NOT EXISTS (SELECT 1 -- comment before a false condition
                 WHERE v = 8) THEN
    NULL;
  ELSE
    RAISE EXCEPTION 'expected false';
  END IF;
  v := (SELECT v -- assignment subquery
        ) + 1;
  IF v <> 8 THEN
    RAISE EXCEPTION 'incorrect assignment';
  END IF;
END $$;"#,
                    expected: Expected::Tag("DO"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "declarations and mutations",
            set_up_script: &[
                "CREATE TABLE do_values (v INT PRIMARY KEY)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "DO $$ DECLARE i INT := 1; BEGIN INSERT INTO do_values VALUES (i); END $$;",
                    expected: Expected::Tag("DO"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM do_values",
                    expected: Expected::Rows {
                        columns: &[Column("v", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "explicit language",
            set_up_script: &[
                "CREATE TABLE do_language (v INT)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "DO LANGUAGE plpgsql $$ BEGIN INSERT INTO do_language VALUES (1); END $$;",
                    expected: Expected::Tag("DO"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DO $$ BEGIN INSERT INTO do_language VALUES (2); END $$ LANGUAGE plpgsql;",
                    expected: Expected::Tag("DO"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM do_language ORDER BY v",
                    expected: Expected::Rows {
                        columns: &[Column("v", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "PostgreSQL documentation example",
            set_up_script: &[
                "CREATE ROLE webuser",
                "CREATE TABLE do_source (v INT)",
                "CREATE VIEW do_view AS SELECT * FROM do_source",
                "CREATE TABLE do_audit (view_name TEXT)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"DO $$DECLARE r record;
BEGIN
    FOR r IN SELECT table_schema, table_name FROM information_schema.tables
             WHERE table_type = 'VIEW' AND table_schema = 'public'
    LOOP
        EXECUTE 'GRANT ALL ON ' || quote_ident(r.table_schema) || '.' || quote_ident(r.table_name) || ' TO webuser';
        INSERT INTO do_audit VALUES (r.table_name);
    END LOOP;
END$$;"#,
                    expected: Expected::Tag("DO"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM do_audit",
                    expected: Expected::Rows {
                        columns: &[Column("view_name", TEXT)],
                        rows: &[
                            &[T("do_view")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "unsupported languages",
            assertions: &[
                ScriptTestAssertion {
                    query: "DO LANGUAGE sql 'SELECT 1';",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"language "sql" does not support inline code execution"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DO LANGUAGE missing_language 'BEGIN NULL; END';",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"language "missing_language" does not exist"#, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "dynamic command expression errors and bindings",
            set_up_script: &[
                "CREATE TABLE do_dynamic (v INT)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "DO $$ BEGIN EXECUTE NULL; END $$;",
                    expected: Expected::Error(Diagnostic { code: "22004", message: "query string argument of EXECUTE is null", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"DO $$ DECLARE
a TEXT := 'I'; b TEXT := 'N'; c TEXT := 'S'; d TEXT := 'E'; e TEXT := 'R';
f TEXT := 'T'; g TEXT := ' '; h TEXT := 'O'; i TEXT := ''; j TEXT := ''; k TEXT := '10';
BEGIN EXECUTE a || b || c || d || e || f || g || a || b || f || h || i || j || g || 'do_dynamic VALUES (' || k || ')'; END $$;"#,
                    expected: Expected::Tag("DO"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM do_dynamic",
                    expected: Expected::Rows {
                        columns: &[Column("v", INT4)],
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
            name: "statement failure is atomic",
            set_up_script: &[
                "CREATE TABLE do_atomic (v INT PRIMARY KEY)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "DO $$ BEGIN INSERT INTO do_atomic VALUES (1); RAISE EXCEPTION 'boom'; END $$;",
                    expected: Expected::Error(Diagnostic { code: "P0001", message: "boom", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM do_atomic",
                    expected: Expected::Rows {
                        columns: &[Column("v", INT4)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "dynamic execute using expressions",
            set_up_script: &[
                "CREATE TABLE do_using (v INT)",
                "CREATE TABLE do_repeated_using (total INT, value TEXT)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"DO $$ DECLARE r RECORD; __dynamic_using_0__ INT := 40; __dynamic_using_1__ RECORD; BEGIN
EXECUTE 'INSERT INTO do_using VALUES ($1), ($2), ($3)' USING 1 + 1, length('abc'), NULL;
EXECUTE 'SELECT $1 AS v' INTO r USING 5 + 1;
INSERT INTO do_using VALUES (r.v);
EXECUTE 'INSERT INTO do_using VALUES ($1), ($2)' USING __dynamic_using_0__ + 1, __dynamic_using_0__ + 2;
EXECUTE 'SELECT $1' INTO __dynamic_using_0__ USING 43;
EXECUTE 'SELECT $1 AS v' INTO __dynamic_using_1__ USING 44;
INSERT INTO do_using VALUES (__dynamic_using_0__), (__dynamic_using_1__.v);
END $$;"#,
                    expected: Expected::Error(Diagnostic { code: "42804", message: r#"column "v" is of type integer but expression is of type text"#, hint: "You will need to rewrite or cast the expression.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"DO $$ DECLARE total INT; value TEXT; r RECORD; BEGIN
EXECUTE 'SELECT $1 + $1, $2::text' INTO total, value USING 6, NULL;
INSERT INTO do_repeated_using VALUES (total, value);
EXECUTE 'SELECT $1 + $1 AS total, $2::text AS value' INTO r USING 7, NULL;
INSERT INTO do_repeated_using VALUES (r.total, r.value);
EXECUTE 'INSERT INTO do_using VALUES ($1), ($1), ($10), ($10)' USING 1, 2, 3, 4, 5, 6, 7, 8, 9, 10;
END $$;"#,
                    expected: Expected::Tag("DO"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM do_repeated_using ORDER BY total",
                    expected: Expected::Rows {
                        columns: &[Column("total", INT4), Column("value", TEXT)],
                        rows: &[
                            &[T("12"), Null],
                            &[T("14"), Null],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v, count(*) FROM do_using WHERE v IN (1, 10) GROUP BY v ORDER BY v",
                    expected: Expected::Rows {
                        columns: &[Column("v", INT4), Column("count", INT8)],
                        rows: &[
                            &[T("1"), T("2")],
                            &[T("10"), T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM do_using WHERE v IS NOT NULL ORDER BY v",
                    expected: Expected::Rows {
                        columns: &[Column("v", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("1")],
                            &[T("10")],
                            &[T("10")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM do_using WHERE v IS NULL",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("0")],
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
