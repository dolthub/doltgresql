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
fn test_drop_table() {
    run_scripts(&[
        ScriptTest {
            name: "DROP TABLE on table type on column",
            set_up_script: &[
                "CREATE TABLE test1 (pk INT4 PRIMARY KEY, v1 TEXT);",
                "CREATE TABLE test2 (v1 test1);",
                "INSERT INTO test2 VALUES (ROW(1, 'abc')::test1);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "DROP TABLE test1;",
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop table test1 because other objects depend on it", detail: "column v1 of table test2 depends on type test1", hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE test2;",
                    expected: Expected::Tag("DROP TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE test1;",
                    expected: Expected::Tag("DROP TABLE"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "DROP TABLE on table type on function parameter",
            set_up_script: &[
                "CREATE TABLE test (pk INT4 PRIMARY KEY, v1 TEXT);",
                "CREATE FUNCTION example_func(t test) RETURNS INT4 AS $$ BEGIN RETURN t.pk * 2; END; $$ LANGUAGE plpgsql;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "DROP TABLE test;",
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop table test because other objects depend on it", detail: "function example_func(test) depends on type test", hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP FUNCTION example_func(test);",
                    expected: Expected::Tag("DROP FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE test;",
                    expected: Expected::Tag("DROP TABLE"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "DROP TABLE on table type on procedure parameter",
            set_up_script: &[
                "CREATE TABLE test1 (pk INT4 PRIMARY KEY, v1 TEXT);",
                "CREATE TABLE test2 (v1 INT4);",
                "CREATE PROCEDURE example_proc(input test1) AS $$ BEGIN INSERT INTO test2 VALUES (input.pk); END; $$ LANGUAGE plpgsql;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "DROP TABLE test1;",
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop table test1 because other objects depend on it", detail: "function example_proc(test1) depends on type test1", hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP PROCEDURE example_proc(test1);",
                    expected: Expected::Tag("DROP PROCEDURE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE test1;",
                    expected: Expected::Tag("DROP TABLE"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "DROP TABLE on table type on column concurrent",
            set_up_script: &[
                "CREATE TABLE test1 (pk INT4 PRIMARY KEY, v1 TEXT);",
                "CREATE TABLE test2 (v1 test1);",
                "INSERT INTO test2 VALUES (ROW(1, 'abc')::test1);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "DROP TABLE test1;",
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop table test1 because other objects depend on it", detail: "column v1 of table test2 depends on type test1", hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE test1, test2;",
                    expected: Expected::Tag("DROP TABLE"),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_drop_table_cascade() {
    run_scripts(&[
        ScriptTest {
            name: "DROP TABLE CASCADE drops foreign keys referencing the table",
            set_up_script: &[
                "CREATE TABLE parent (pk INT4 PRIMARY KEY, v1 TEXT);",
                "CREATE TABLE child (pk INT4 PRIMARY KEY, parent_pk INT4, CONSTRAINT child_parent_fk FOREIGN KEY (parent_pk) REFERENCES parent (pk));",
                "INSERT INTO parent VALUES (1, 'one');",
                "INSERT INTO child VALUES (10, 1);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "DROP TABLE parent;",
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop table parent because other objects depend on it", detail: "constraint child_parent_fk on table child depends on table parent", hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE parent RESTRICT;",
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop table parent because other objects depend on it", detail: "constraint child_parent_fk on table child depends on table parent", hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE parent CASCADE;",
                    expected: Expected::Tag("DROP TABLE"),
                    notices: &[Diagnostic { code: "00000", message: "drop cascades to constraint child_parent_fk on table child", ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM parent;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "parent" does not exist"#, position: 15, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM child;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("parent_pk", INT4)],
                        rows: &[
                            &[T("10"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO child VALUES (11, 99);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM child ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("parent_pk", INT4)],
                        rows: &[
                            &[T("10"), T("1")],
                            &[T("11"), T("99")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "DROP TABLE RESTRICT succeeds when nothing depends on the table",
            set_up_script: &[
                "CREATE TABLE test (pk INT4 PRIMARY KEY);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "DROP TABLE test RESTRICT;",
                    expected: Expected::Tag("DROP TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "test" does not exist"#, position: 15, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "DROP TABLE CASCADE with no dependencies",
            set_up_script: &[
                "CREATE TABLE test (pk INT4 PRIMARY KEY);",
                "INSERT INTO test VALUES (1);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "DROP TABLE test CASCADE;",
                    expected: Expected::Tag("DROP TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "test" does not exist"#, position: 15, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "DROP TABLE CASCADE with a self-referential foreign key",
            set_up_script: &[
                "CREATE TABLE test (pk INT4 PRIMARY KEY, parent_pk INT4, CONSTRAINT test_self_fk FOREIGN KEY (parent_pk) REFERENCES test (pk));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "DROP TABLE test CASCADE;",
                    expected: Expected::Tag("DROP TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "test" does not exist"#, position: 15, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "DROP TABLE CASCADE with multiple tables",
            set_up_script: &[
                "CREATE TABLE parent (pk INT4 PRIMARY KEY);",
                "CREATE TABLE middle (pk INT4 PRIMARY KEY, parent_pk INT4, CONSTRAINT middle_parent_fk FOREIGN KEY (parent_pk) REFERENCES parent (pk));",
                "CREATE TABLE child (pk INT4 PRIMARY KEY, middle_pk INT4, CONSTRAINT child_middle_fk FOREIGN KEY (middle_pk) REFERENCES middle (pk));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "DROP TABLE parent, middle CASCADE;",
                    expected: Expected::Tag("DROP TABLE"),
                    notices: &[Diagnostic { code: "00000", message: "drop cascades to constraint child_middle_fk on table child", ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM parent;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "parent" does not exist"#, position: 15, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM middle;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "middle" does not exist"#, position: 15, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO child VALUES (1, 99);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "DROP TABLE IF EXISTS CASCADE",
            set_up_script: &[
                "CREATE TABLE parent (pk INT4 PRIMARY KEY);",
                "CREATE TABLE child (pk INT4 PRIMARY KEY, parent_pk INT4, CONSTRAINT child_parent_fk FOREIGN KEY (parent_pk) REFERENCES parent (pk));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "DROP TABLE IF EXISTS doesnotexist CASCADE;",
                    expected: Expected::Tag("DROP TABLE"),
                    notices: &[Diagnostic { code: "00000", message: r#"table "doesnotexist" does not exist, skipping"#, ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE doesnotexist CASCADE;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"table "doesnotexist" does not exist"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE IF EXISTS doesnotexist, parent CASCADE;",
                    expected: Expected::Tag("DROP TABLE"),
                    notices: &[Diagnostic { code: "00000", message: r#"table "doesnotexist" does not exist, skipping"#, ..N }, Diagnostic { code: "00000", message: "drop cascades to constraint child_parent_fk on table child", ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM parent;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "parent" does not exist"#, position: 15, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO child VALUES (1, 99);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "DROP TABLE CASCADE with schema-qualified names",
            set_up_script: &[
                "CREATE SCHEMA sch1;",
                "CREATE SCHEMA sch2;",
                "CREATE TABLE sch1.parent (pk INT4 PRIMARY KEY);",
                "CREATE TABLE sch2.child (pk INT4 PRIMARY KEY, parent_pk INT4, CONSTRAINT child_parent_fk FOREIGN KEY (parent_pk) REFERENCES sch1.parent (pk));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "DROP TABLE sch1.parent;",
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop table sch1.parent because other objects depend on it", detail: "constraint child_parent_fk on table sch2.child depends on table sch1.parent", hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE sch1.parent CASCADE;",
                    expected: Expected::Tag("DROP TABLE"),
                    notices: &[Diagnostic { code: "00000", message: "drop cascades to constraint child_parent_fk on table sch2.child", ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM sch1.parent;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "sch1.parent" does not exist"#, position: 15, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO sch2.child VALUES (1, 99);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "DROP TABLE CASCADE resolves tables on the search path",
            set_up_script: &[
                "CREATE SCHEMA sch1;",
                "SET search_path TO sch1;",
                "CREATE TABLE parent (pk INT4 PRIMARY KEY);",
                "CREATE TABLE child (pk INT4 PRIMARY KEY, parent_pk INT4, CONSTRAINT child_parent_fk FOREIGN KEY (parent_pk) REFERENCES parent (pk));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "DROP TABLE parent CASCADE;",
                    expected: Expected::Tag("DROP TABLE"),
                    notices: &[Diagnostic { code: "00000", message: "drop cascades to constraint child_parent_fk on table child", ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM sch1.parent;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "sch1.parent" does not exist"#, position: 15, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO sch1.child VALUES (1, 99);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "DROP TABLE CASCADE drops sequences owned by the table",
            set_up_script: &[
                "CREATE TABLE test (pk SERIAL PRIMARY KEY, v1 TEXT);",
                "INSERT INTO test (v1) VALUES ('one');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT nextval('test_pk_seq');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE test CASCADE;",
                    expected: Expected::Tag("DROP TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test_pk_seq');",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "test_pk_seq" does not exist"#, position: 16, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "DROP TABLE CASCADE with a dependent view",
            set_up_script: &[
                "CREATE TABLE test (pk INT4 PRIMARY KEY, v1 TEXT);",
                "INSERT INTO test VALUES (1, 'one');",
                "CREATE VIEW test_view AS SELECT * FROM test;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "DROP TABLE test CASCADE;",
                    expected: Expected::Tag("DROP TABLE"),
                    notices: &[Diagnostic { code: "00000", message: "drop cascades to view test_view", ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test_view;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "test_view" does not exist"#, position: 15, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "DROP TABLE CASCADE drops transitively dependent views only",
            set_up_script: &[
                "CREATE TABLE test (pk INT4 PRIMARY KEY, v1 TEXT);",
                "CREATE TABLE other (pk INT4 PRIMARY KEY);",
                "INSERT INTO test VALUES (1, 'one');",
                "INSERT INTO other VALUES (7);",
                "CREATE VIEW test_view AS SELECT * FROM test;",
                "CREATE VIEW test_view_view AS SELECT pk FROM test_view;",
                "CREATE VIEW other_view AS SELECT * FROM other;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "DROP TABLE test CASCADE;",
                    expected: Expected::Tag("DROP TABLE"),
                    notices: &[Diagnostic { code: "00000", message: "drop cascades to 2 other objects", detail: r#"drop cascades to view test_view
drop cascades to view test_view_view"#, ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test_view;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "test_view" does not exist"#, position: 15, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test_view_view;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "test_view_view" does not exist"#, position: 15, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM other_view;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
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
            name: "DROP TABLE CASCADE does not drop views resolving to a same-named table in another schema",
            set_up_script: &[
                "CREATE SCHEMA sch1;",
                "CREATE TABLE test (pk INT4 PRIMARY KEY);",
                "CREATE TABLE sch1.test (pk INT4 PRIMARY KEY);",
                "INSERT INTO sch1.test VALUES (3);",
                "CREATE VIEW qualified_view AS SELECT * FROM sch1.test;",
                "CREATE VIEW unqualified_view AS SELECT * FROM test;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "DROP TABLE test CASCADE;",
                    expected: Expected::Tag("DROP TABLE"),
                    notices: &[Diagnostic { code: "00000", message: "drop cascades to view unqualified_view", ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM qualified_view;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM unqualified_view;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "unqualified_view" does not exist"#, position: 15, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "DROP TABLE CASCADE drops columns using the table's row type",
            set_up_script: &[
                "CREATE TABLE test1 (pk INT4 PRIMARY KEY, v1 TEXT);",
                "CREATE TABLE test2 (pk INT4 PRIMARY KEY, v1 test1);",
                "INSERT INTO test2 VALUES (1, ROW(2, 'abc')::test1);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "DROP TABLE test1;",
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop table test1 because other objects depend on it", detail: "column v1 of table test2 depends on type test1", hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE test1 CASCADE;",
                    expected: Expected::Tag("DROP TABLE"),
                    notices: &[Diagnostic { code: "00000", message: "drop cascades to column v1 of table test2", ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test2;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
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
            name: "DROP TABLE CASCADE drops functions and procedures using the table's row type",
            set_up_script: &[
                "CREATE TABLE test (pk INT4 PRIMARY KEY, v1 TEXT);",
                "CREATE FUNCTION dependent_func(t test) RETURNS INT4 AS $$ BEGIN RETURN t.pk * 2; END; $$ LANGUAGE plpgsql;",
                "CREATE FUNCTION unrelated_func(v INT4) RETURNS INT4 AS $$ BEGIN RETURN v + 1; END; $$ LANGUAGE plpgsql;",
                "CREATE PROCEDURE dependent_proc(input test) AS $$ BEGIN END; $$ LANGUAGE plpgsql;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "DROP TABLE test;",
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop table test because other objects depend on it", detail: r#"function dependent_func(test) depends on type test
function dependent_proc(test) depends on type test"#, hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE test CASCADE;",
                    expected: Expected::Tag("DROP TABLE"),
                    notices: &[Diagnostic { code: "00000", message: "drop cascades to 2 other objects", detail: r#"drop cascades to function dependent_func(test)
drop cascades to function dependent_proc(test)"#, ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT dependent_func(NULL);",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function dependent_func(unknown) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT unrelated_func(1);",
                    expected: Expected::Rows {
                        columns: &[Column("unrelated_func", INT4)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CALL dependent_proc(NULL);",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "procedure dependent_proc(unknown) does not exist", hint: "No procedure matches the given name and argument types. You might need to add explicit type casts.", position: 6, ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}
