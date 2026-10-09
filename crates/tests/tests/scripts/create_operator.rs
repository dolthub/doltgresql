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
fn test_create_operator() {
    run_scripts(&[
        ScriptTest {
            name: "CREATE OPERATOR with LEFTARG, RIGHTARG, and FUNCTION",
            set_up_script: &[
                r#"CREATE FUNCTION op_int_dist(a int4, b int4) RETURNS int4
					AS $$ SELECT abs(a - b) $$ LANGUAGE SQL;"#,
                "CREATE TABLE op_points (pk int4 PRIMARY KEY, v int4);",
                "INSERT INTO op_points VALUES (1, 4), (2, 9), (3, 15);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE OPERATOR <-> (LEFTARG = int4, RIGHTARG = int4, FUNCTION = op_int_dist);",
                    expected: Expected::Tag("CREATE OPERATOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 3 <-> 10;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
                        rows: &[
                            &[T("7")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 10 <-> 3;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
                        rows: &[
                            &[T("7")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (NULL::int4 <-> 3) IS NULL;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pk, v <-> 10 FROM op_points ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("?column?", INT4)],
                        rows: &[
                            &[T("1"), T("6")],
                            &[T("2"), T("1")],
                            &[T("3"), T("5")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pk FROM op_points WHERE v <-> 10 < 3 ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT oprname, oprkind, oprcanhash, oprcanmerge, oprleft::regtype::text, oprright::regtype::text, oprresult::regtype::text
						FROM pg_operator WHERE oprcode::text = 'op_int_dist';"#,
                    expected: Expected::Rows {
                        columns: &[Column("oprname", NAME), Column("oprkind", CHAR), Column("oprcanhash", BOOL), Column("oprcanmerge", BOOL), Column("oprleft", TEXT), Column("oprright", TEXT), Column("oprresult", TEXT)],
                        rows: &[
                            &[T("<->"), T("b"), T("f"), T("f"), T("integer"), T("integer"), T("integer")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT n.nspname FROM pg_operator o JOIN pg_namespace n ON n.oid = o.oprnamespace WHERE o.oprcode::text = 'op_int_dist';",
                    expected: Expected::Rows {
                        columns: &[Column("nspname", NAME)],
                        rows: &[
                            &[T("public")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CREATE OPERATOR with COMMUTATOR and NEGATOR",
            set_up_script: &[
                r#"CREATE FUNCTION op_len_eq(a text, b text) RETURNS boolean
					AS $$ SELECT length(a) = length(b) $$ LANGUAGE SQL;"#,
                r#"CREATE FUNCTION op_len_ne(a text, b text) RETURNS boolean
					AS $$ SELECT length(a) <> length(b) $$ LANGUAGE SQL;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE OPERATOR <=> (LEFTARG = text, RIGHTARG = text, FUNCTION = op_len_eq, COMMUTATOR = <=>);",
                    expected: Expected::Tag("CREATE OPERATOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE OPERATOR <~> (LEFTARG = text, RIGHTARG = text, FUNCTION = op_len_ne, COMMUTATOR = <~>, NEGATOR = <=>);",
                    expected: Expected::Tag("CREATE OPERATOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'abc' <=> 'xyz';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'abc' <=> 'wxyz';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'abc' <~> 'wxyz';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT c.oprname FROM pg_operator o JOIN pg_operator c ON c.oid = o.oprcom WHERE o.oprcode::text = 'op_len_eq';",
                    expected: Expected::Rows {
                        columns: &[Column("oprname", NAME)],
                        rows: &[
                            &[T("<=>")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT n.oprname FROM pg_operator o JOIN pg_operator n ON n.oid = o.oprnegate WHERE o.oprcode::text = 'op_len_ne';",
                    expected: Expected::Rows {
                        columns: &[Column("oprname", NAME)],
                        rows: &[
                            &[T("<=>")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT n.oprname FROM pg_operator o JOIN pg_operator n ON n.oid = o.oprnegate WHERE o.oprcode::text = 'op_len_eq';",
                    expected: Expected::Rows {
                        columns: &[Column("oprname", NAME)],
                        rows: &[
                            &[T("<~>")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CREATE OPERATOR with HASHES and MERGES",
            set_up_script: &[
                r#"CREATE FUNCTION op_ci_eq(a text, b text) RETURNS boolean
					AS $$ SELECT lower(a) = lower(b) $$ LANGUAGE SQL;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE OPERATOR <%> (LEFTARG = text, RIGHTARG = text, FUNCTION = op_ci_eq, COMMUTATOR = <%>, HASHES, MERGES);",
                    expected: Expected::Tag("CREATE OPERATOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'ABC' <%> 'abc';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT oprcanhash, oprcanmerge FROM pg_operator WHERE oprcode::text = 'op_ci_eq';",
                    expected: Expected::Rows {
                        columns: &[Column("oprcanhash", BOOL), Column("oprcanmerge", BOOL)],
                        rows: &[
                            &[T("t"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CREATE OPERATOR with composite operands",
            set_up_script: &[
                "CREATE TABLE op_pair (x int4, y int4);",
                r#"CREATE FUNCTION op_pair_add(a op_pair, b op_pair) RETURNS op_pair
					AS $$ SELECT ROW((a).x + (b).x, (a).y + (b).y)::op_pair $$ LANGUAGE SQL;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE OPERATOR <+> (LEFTARG = op_pair, RIGHTARG = op_pair, FUNCTION = op_pair_add);",
                    expected: Expected::Tag("CREATE OPERATOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 2)::op_pair <+> ROW(3, 4)::op_pair;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", USER_DEFINED)],
                        rows: &[
                            &[T("(4,6)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT oprleft::regtype::text, oprright::regtype::text, oprresult::regtype::text
						FROM pg_operator WHERE oprcode::text = 'op_pair_add';"#,
                    expected: Expected::Rows {
                        columns: &[Column("oprleft", TEXT), Column("oprright", TEXT), Column("oprresult", TEXT)],
                        rows: &[
                            &[T("op_pair"), T("op_pair"), T("op_pair")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CREATE OPERATOR with composite operands on table rows in aggregate",
            set_up_script: &[
                "CREATE TABLE op_pair (x int4, y int4);",
                r#"CREATE FUNCTION op_pair_add(a op_pair, b op_pair) RETURNS op_pair
			AS $$ SELECT ROW((a).x + (b).x, (a).y + (b).y)::op_pair $$ LANGUAGE SQL;"#,
                "CREATE OPERATOR <+> (LEFTARG = op_pair, RIGHTARG = op_pair, FUNCTION = op_pair_add);",
                "INSERT INTO op_pair VALUES (1, 2), (3, 4), (5, 6);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT sum(((p <+> ROW(10, 20)::op_pair)).x), sum(((p <+> ROW(10, 20)::op_pair)).y)
			       FROM op_pair p;"#,
                    expected: Expected::Rows {
                        columns: &[Column("sum", INT8), Column("sum", INT8)],
                        rows: &[
                            &[T("39"), T("72")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CREATE OPERATOR with mixed operand types",
            set_up_script: &[
                r#"CREATE FUNCTION op_repeat(a text, b int4) RETURNS text
					AS $$ SELECT repeat(a, b) $$ LANGUAGE SQL;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE OPERATOR <#> (LEFTARG = text, RIGHTARG = int4, FUNCTION = op_repeat);",
                    expected: Expected::Tag("CREATE OPERATOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'ab' <#> 3;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT)],
                        rows: &[
                            &[T("ababab")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 3 <#> 'ab';",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "operator does not exist: integer <#> unknown", hint: "No operator matches the given name and argument types. You might need to add explicit type casts.", position: 10, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "DROP OPERATOR smoke test",
            set_up_script: &[
                r#"CREATE FUNCTION op_drop_dist(a int4, b int4) RETURNS int4
					AS $$ SELECT abs(a - b) $$ LANGUAGE SQL;"#,
                "CREATE OPERATOR <-> (LEFTARG = int4, RIGHTARG = int4, FUNCTION = op_drop_dist);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT 3 <-> 10;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
                        rows: &[
                            &[T("7")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP OPERATOR <-> (int4, int4);",
                    expected: Expected::Tag("DROP OPERATOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 3 <-> 10;",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "operator does not exist: integer <-> integer", hint: "No operator matches the given name and argument types. You might need to add explicit type casts.", position: 10, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXISTS (SELECT 1 FROM pg_operator WHERE oprcode::text = 'op_drop_dist');",
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
                    query: "CREATE OPERATOR <-> (LEFTARG = int4, RIGHTARG = int4, FUNCTION = op_drop_dist);",
                    expected: Expected::Tag("CREATE OPERATOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 3 <-> 10;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
                        rows: &[
                            &[T("7")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CREATE OPERATOR validation",
            set_up_script: &[
                r#"CREATE FUNCTION op_valid_dist(a int4, b int4) RETURNS int4
					AS $$ SELECT abs(a - b) $$ LANGUAGE SQL;"#,
                r#"CREATE FUNCTION op_one_arg(a int4) RETURNS int4
					AS $$ SELECT a $$ LANGUAGE SQL;"#,
                "CREATE OPERATOR <-> (LEFTARG = int4, RIGHTARG = int4, FUNCTION = op_valid_dist);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE OPERATOR <=> (LEFTARG = int4, RIGHTARG = int4);",
                    expected: Expected::Error(Diagnostic { code: "42P13", message: "operator function must be specified", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE OPERATOR <=> (FUNCTION = op_valid_dist);",
                    expected: Expected::Error(Diagnostic { code: "42P13", message: "operator argument types must be specified", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE OPERATOR <=> (LEFTARG = int4, FUNCTION = op_valid_dist);",
                    expected: Expected::Error(Diagnostic { code: "42P13", message: "operator right argument type must be specified", detail: "Postfix operators are not supported.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE OPERATOR <=> (LEFTARG = int4, RIGHTARG = int4, FUNCTION = op_missing_fn);",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function op_missing_fn(integer, integer) does not exist", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE OPERATOR <=> (LEFTARG = int4, RIGHTARG = int4, FUNCTION = op_one_arg);",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function op_one_arg(integer, integer) does not exist", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE OPERATOR <-> (LEFTARG = int4, RIGHTARG = int4, FUNCTION = op_valid_dist);",
                    expected: Expected::Error(Diagnostic { code: "42723", message: "operator <-> already exists", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CREATE OPERATOR prefix operators, options, and DROP OPERATOR errors",
            set_up_script: &[
                r#"CREATE FUNCTION op_prefix_neg(b int4) RETURNS int4
					AS $$ SELECT -b $$ LANGUAGE SQL;"#,
                r#"CREATE FUNCTION op_prefix_sub(a int4, b int4) RETURNS int4
					AS $$ SELECT a - b $$ LANGUAGE SQL;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "DROP OPERATOR <-> (int4, int4);",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "operator does not exist: integer <-> integer", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP OPERATOR IF EXISTS <-> (int4, int4);",
                    expected: Expected::Tag("DROP OPERATOR"),
                    notices: &[Diagnostic { code: "00000", message: "operator <-> does not exist, skipping", ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE OPERATOR <-> (RIGHTARG = int4, FUNCTION = op_prefix_neg, COMMUTATOR = <->);",
                    expected: Expected::Error(Diagnostic { code: "42P13", message: "only binary operators can have commutators", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE OPERATOR <-> (RIGHTARG = int4, FUNCTION = op_prefix_neg, HASHES);",
                    expected: Expected::Error(Diagnostic { code: "42P13", message: "only binary operators can hash", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE OPERATOR <-> (LEFTARG = int4, RIGHTARG = int4, FUNCTION = op_prefix_sub, NEGATOR = <>);",
                    expected: Expected::Error(Diagnostic { code: "42P13", message: "only boolean operators can have negators", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE OPERATOR <-> (LEFTARG = int4, RIGHTARG = int4, PROCEDURE = op_prefix_sub, BOGUS);",
                    expected: Expected::Tag("CREATE OPERATOR"),
                    notices: &[Diagnostic { severity: "WARNING", code: "42601", message: r#"operator attribute "bogus" not recognized"#, ..E }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE OPERATOR <-> (RIGHTARG = int4, FUNCTION = op_prefix_neg);",
                    expected: Expected::Tag("CREATE OPERATOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT <-> 5, 7 <-> 5;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4), Column("?column?", INT4)],
                        rows: &[
                            &[T("-5"), T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT oprkind, oprleft, oprright, oprresult, oprowner, oprcode::text, oprrest::text, oprjoin::text FROM pg_operator WHERE oprname = '<->' AND oprnamespace = 'public'::regnamespace ORDER BY oprkind;",
                    expected: Expected::Rows {
                        columns: &[Column("oprkind", CHAR), Column("oprleft", OID), Column("oprright", OID), Column("oprresult", OID), Column("oprowner", OID), Column("oprcode", TEXT), Column("oprrest", TEXT), Column("oprjoin", TEXT)],
                        rows: &[
                            &[T("b"), T("23"), T("23"), T("23"), T("10"), T("op_prefix_sub"), T("-"), T("-")],
                            &[T("l"), T("0"), T("23"), T("23"), T("10"), T("op_prefix_neg"), T("-"), T("-")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP OPERATOR <-> (none, int4);",
                    expected: Expected::Tag("DROP OPERATOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP OPERATOR <-> (int4, int4);",
                    expected: Expected::Tag("DROP OPERATOR"),
                    ..A
                },
            ],
            ..S
        },
    ]);
}
