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
fn test_create_aggregate() {
    run_scripts(&[
        ScriptTest {
            name: "CREATE AGGREGATE with SFUNC and STYPE",
            set_up_script: &[
                r#"CREATE FUNCTION agg_sum_step(state int4, val int4) RETURNS int4
					AS $$ SELECT state + val $$ LANGUAGE SQL;"#,
                "CREATE TABLE agg_sum_vals (pk int4 PRIMARY KEY, grp text, v int4);",
                "INSERT INTO agg_sum_vals VALUES (1, 'a', 10), (2, 'a', 20), (3, 'b', 5), (4, 'b', NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE AGGREGATE agg_sum (int4) (SFUNC = agg_sum_step, STYPE = int4, INITCOND = '0');",
                    expected: Expected::Tag("CREATE AGGREGATE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT agg_sum(v) FROM agg_sum_vals WHERE grp = 'a';",
                    expected: Expected::Rows {
                        columns: &[Column("agg_sum", INT4)],
                        rows: &[
                            &[T("30")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT grp, agg_sum(v) FROM agg_sum_vals GROUP BY grp ORDER BY grp;",
                    expected: Expected::Rows {
                        columns: &[Column("grp", TEXT), Column("agg_sum", INT4)],
                        rows: &[
                            &[T("a"), T("30")],
                            &[T("b"), Null],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT agg_sum(v) FROM agg_sum_vals WHERE pk = 0;",
                    expected: Expected::Rows {
                        columns: &[Column("agg_sum", INT4)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT aggkind, agginitval, aggfinalfn::text, aggcombinefn::text FROM pg_aggregate WHERE aggtransfn::text = 'agg_sum_step';",
                    expected: Expected::Rows {
                        columns: &[Column("aggkind", CHAR), Column("agginitval", TEXT), Column("aggfinalfn", TEXT), Column("aggcombinefn", TEXT)],
                        rows: &[
                            &[T("n"), T("0"), T("-"), T("-")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT proname, prokind FROM pg_proc WHERE proname = 'agg_sum';",
                    expected: Expected::Rows {
                        columns: &[Column("proname", NAME), Column("prokind", CHAR)],
                        rows: &[
                            &[T("agg_sum"), T("a")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CREATE AGGREGATE with STRICT transition function and no INITCOND",
            set_up_script: &[
                r#"CREATE FUNCTION agg_larger_step(state int4, val int4) RETURNS int4
					AS $$ SELECT CASE WHEN state > val THEN state ELSE val END $$ LANGUAGE SQL STRICT;"#,
                "CREATE TABLE agg_larger_vals (pk int4 PRIMARY KEY, v int4);",
                "INSERT INTO agg_larger_vals VALUES (1, 3), (2, NULL), (3, 8), (4, 5);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE AGGREGATE agg_larger (int4) (SFUNC = agg_larger_step, STYPE = int4);",
                    expected: Expected::Tag("CREATE AGGREGATE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT agg_larger(v) FROM agg_larger_vals;",
                    expected: Expected::Rows {
                        columns: &[Column("agg_larger", INT4)],
                        rows: &[
                            &[T("8")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT agg_larger(v) FROM agg_larger_vals WHERE pk = 2;",
                    expected: Expected::Rows {
                        columns: &[Column("agg_larger", INT4)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT agg_larger(v) FROM agg_larger_vals WHERE pk = 0;",
                    expected: Expected::Rows {
                        columns: &[Column("agg_larger", INT4)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT agginitval IS NULL FROM pg_aggregate WHERE aggtransfn::text = 'agg_larger_step';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
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
            name: "CREATE AGGREGATE with FINALFUNC",
            set_up_script: &[
                r#"CREATE FUNCTION agg_charcount_step(state int4, val text) RETURNS int4
					AS $$ SELECT state + length(val) $$ LANGUAGE SQL;"#,
                r#"CREATE FUNCTION agg_charcount_final(state int4) RETURNS text
					AS $$ SELECT 'chars: ' || state $$ LANGUAGE SQL;"#,
                "CREATE TABLE agg_charcount_vals (pk int4 PRIMARY KEY, v text);",
                "INSERT INTO agg_charcount_vals VALUES (1, 'ab'), (2, 'cde'), (3, 'f');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE AGGREGATE agg_charcount (text) (SFUNC = agg_charcount_step, STYPE = int4,
						FINALFUNC = agg_charcount_final, INITCOND = '0');"#,
                    expected: Expected::Tag("CREATE AGGREGATE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT agg_charcount(v) FROM agg_charcount_vals;",
                    expected: Expected::Rows {
                        columns: &[Column("agg_charcount", TEXT)],
                        rows: &[
                            &[T("chars: 6")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT agg_charcount(v) FROM agg_charcount_vals WHERE pk = 0;",
                    expected: Expected::Rows {
                        columns: &[Column("agg_charcount", TEXT)],
                        rows: &[
                            &[T("chars: 0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT aggfinalfn::text FROM pg_aggregate WHERE aggtransfn::text = 'agg_charcount_step';",
                    expected: Expected::Rows {
                        columns: &[Column("aggfinalfn", TEXT)],
                        rows: &[
                            &[T("agg_charcount_final")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CREATE AGGREGATE with COMBINEFUNC",
            set_up_script: &[
                r#"CREATE FUNCTION agg_combined_step(state int8, val int4) RETURNS int8
					AS $$ SELECT state + val $$ LANGUAGE SQL;"#,
                r#"CREATE FUNCTION agg_combined_merge(s1 int8, s2 int8) RETURNS int8
					AS $$ SELECT s1 + s2 $$ LANGUAGE SQL;"#,
                "CREATE TABLE agg_combined_vals (pk int4 PRIMARY KEY, v int4);",
                "INSERT INTO agg_combined_vals VALUES (1, 10), (2, 20), (3, 5);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE AGGREGATE agg_combined (int4) (SFUNC = agg_combined_step, STYPE = int8,
						COMBINEFUNC = agg_combined_merge, INITCOND = '0');"#,
                    expected: Expected::Tag("CREATE AGGREGATE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT agg_combined(v) FROM agg_combined_vals;",
                    expected: Expected::Rows {
                        columns: &[Column("agg_combined", INT8)],
                        rows: &[
                            &[T("35")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT aggcombinefn::text FROM pg_aggregate WHERE aggtransfn::text = 'agg_combined_step';",
                    expected: Expected::Rows {
                        columns: &[Column("aggcombinefn", TEXT)],
                        rows: &[
                            &[T("agg_combined_merge")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CREATE AGGREGATE with multiple arguments",
            set_up_script: &[
                r#"CREATE FUNCTION agg_wsum_step(state int8, val int4, weight int4) RETURNS int8
					AS $$ SELECT state + (val * weight) $$ LANGUAGE SQL;"#,
                "CREATE TABLE agg_wsum_vals (pk int4 PRIMARY KEY, v int4, w int4);",
                "INSERT INTO agg_wsum_vals VALUES (1, 10, 1), (2, 20, 2), (3, 30, 3);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE AGGREGATE agg_wsum (int4, int4) (SFUNC = agg_wsum_step, STYPE = int8, INITCOND = '0');",
                    expected: Expected::Tag("CREATE AGGREGATE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT agg_wsum(v, w) FROM agg_wsum_vals;",
                    expected: Expected::Rows {
                        columns: &[Column("agg_wsum", INT8)],
                        rows: &[
                            &[T("140")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CREATE AGGREGATE in a custom schema",
            set_up_script: &[
                "CREATE SCHEMA agg_nsp;",
                r#"CREATE FUNCTION agg_nsp_step(state int4, val int4) RETURNS int4
					AS $$ SELECT state + 1 $$ LANGUAGE SQL;"#,
                "CREATE TABLE agg_nsp_vals (pk int4 PRIMARY KEY, v int4);",
                "INSERT INTO agg_nsp_vals VALUES (1, 10), (2, 20), (3, NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE AGGREGATE agg_nsp.agg_rows (int4) (SFUNC = agg_nsp_step, STYPE = int4, INITCOND = '0');",
                    expected: Expected::Tag("CREATE AGGREGATE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT agg_nsp.agg_rows(v) FROM agg_nsp_vals;",
                    expected: Expected::Rows {
                        columns: &[Column("agg_rows", INT4)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT n.nspname FROM pg_proc p JOIN pg_namespace n ON n.oid = p.pronamespace WHERE p.proname = 'agg_rows';",
                    expected: Expected::Rows {
                        columns: &[Column("nspname", NAME)],
                        rows: &[
                            &[T("agg_nsp")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT agg_rows(v) FROM agg_nsp_vals;",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function agg_rows(integer) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CREATE OR REPLACE AGGREGATE",
            set_up_script: &[
                r#"CREATE FUNCTION agg_replace_step(state int4, val int4) RETURNS int4
					AS $$ SELECT state + val $$ LANGUAGE SQL;"#,
                "CREATE TABLE agg_replace_vals (pk int4 PRIMARY KEY, v int4);",
                "INSERT INTO agg_replace_vals VALUES (1, 10), (2, 20);",
                "CREATE AGGREGATE agg_replace (int4) (SFUNC = agg_replace_step, STYPE = int4, INITCOND = '0');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT agg_replace(v) FROM agg_replace_vals;",
                    expected: Expected::Rows {
                        columns: &[Column("agg_replace", INT4)],
                        rows: &[
                            &[T("30")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE OR REPLACE AGGREGATE agg_replace (int4) (SFUNC = agg_replace_step, STYPE = int4, INITCOND = '100');",
                    expected: Expected::Tag("CREATE AGGREGATE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT agg_replace(v) FROM agg_replace_vals;",
                    expected: Expected::Rows {
                        columns: &[Column("agg_replace", INT4)],
                        rows: &[
                            &[T("130")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT agginitval FROM pg_aggregate WHERE aggtransfn::text = 'agg_replace_step';",
                    expected: Expected::Rows {
                        columns: &[Column("agginitval", TEXT)],
                        rows: &[
                            &[T("100")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "DROP AGGREGATE smoke test",
            set_up_script: &[
                r#"CREATE FUNCTION agg_drop_step(state int4, val int4) RETURNS int4
					AS $$ SELECT state + val $$ LANGUAGE SQL;"#,
                "CREATE TABLE agg_drop_vals (pk int4 PRIMARY KEY, v int4);",
                "INSERT INTO agg_drop_vals VALUES (1, 10), (2, 20);",
                "CREATE AGGREGATE agg_drop (int4) (SFUNC = agg_drop_step, STYPE = int4, INITCOND = '0');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT agg_drop(v) FROM agg_drop_vals;",
                    expected: Expected::Rows {
                        columns: &[Column("agg_drop", INT4)],
                        rows: &[
                            &[T("30")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP AGGREGATE agg_drop(int4);",
                    expected: Expected::Tag("DROP AGGREGATE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT agg_drop(v) FROM agg_drop_vals;",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function agg_drop(integer) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'agg_drop');",
                    expected: Expected::Rows {
                        columns: &[Column("exists", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE AGGREGATE agg_drop (int4) (SFUNC = agg_drop_step, STYPE = int4, INITCOND = '0');",
                    expected: Expected::Tag("CREATE AGGREGATE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT agg_drop(v) FROM agg_drop_vals;",
                    expected: Expected::Rows {
                        columns: &[Column("agg_drop", INT4)],
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
            name: "CREATE AGGREGATE validation",
            set_up_script: &[
                r#"CREATE FUNCTION agg_valid_step(state int4, val int4) RETURNS int4
					AS $$ SELECT state + val $$ LANGUAGE SQL;"#,
                r#"CREATE FUNCTION agg_one_arg(state int4) RETURNS int4
					AS $$ SELECT state $$ LANGUAGE SQL;"#,
                r#"CREATE FUNCTION agg_wrong_ret(state int4, val int4) RETURNS text
					AS $$ SELECT 'x' $$ LANGUAGE SQL;"#,
                r#"CREATE FUNCTION agg_taken(a int4) RETURNS int4
					AS $$ SELECT a $$ LANGUAGE SQL;"#,
                "CREATE AGGREGATE agg_valid (int4) (SFUNC = agg_valid_step, STYPE = int4);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE AGGREGATE agg_bad1 (int4) (SFUNC = agg_missing_step, STYPE = int4);",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function agg_missing_step(integer, integer) does not exist", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE AGGREGATE agg_bad2 (int4) (SFUNC = agg_one_arg, STYPE = int4);",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function agg_one_arg(integer, integer) does not exist", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE AGGREGATE agg_bad3 (int4) (SFUNC = agg_wrong_ret, STYPE = int4);",
                    expected: Expected::Error(Diagnostic { code: "42804", message: "return type of transition function agg_wrong_ret is not integer", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE AGGREGATE agg_bad4 (int4) (SFUNC = agg_valid_step, STYPE = int4, FINALFUNC = agg_missing_final);",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function agg_missing_final(integer) does not exist", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE AGGREGATE agg_valid (int4) (SFUNC = agg_valid_step, STYPE = int4);",
                    expected: Expected::Error(Diagnostic { code: "42723", message: r#"function "agg_valid" already exists with same argument types"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE AGGREGATE agg_taken (int4) (SFUNC = agg_valid_step, STYPE = int4);",
                    expected: Expected::Error(Diagnostic { code: "42723", message: r#"function "agg_taken" already exists with same argument types"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE AGGREGATE agg_bad5 (OUT x int4) (SFUNC = agg_valid_step, STYPE = int4);",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "aggregates cannot have output arguments", position: 28, ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}
