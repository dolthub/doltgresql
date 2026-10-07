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
fn test_set_constraints() {
    run_scripts(&[
        ScriptTest {
            name: "SET CONSTRAINTS",
            set_up_script: &[
                "CREATE TABLE parent_example (id INTEGER PRIMARY KEY, u INTEGER CONSTRAINT uniq_u UNIQUE, c INTEGER CONSTRAINT chk CHECK (c > 0));",
                "CREATE TABLE child_example (id INTEGER PRIMARY KEY, p2 INTEGER CONSTRAINT nd_fk REFERENCES parent_example(id));",
                "CREATE SCHEMA s2;",
                "CREATE TABLE s2.t2 (id INTEGER PRIMARY KEY);",
                "CREATE DOMAIN dom AS INTEGER CONSTRAINT dom_chk CHECK (VALUE > 0);",
                "CREATE TABLE named_pk_example (id INTEGER CONSTRAINT named_pk PRIMARY KEY);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SET CONSTRAINTS ALL IMMEDIATE;",
                    expected: Expected::Tag("SET CONSTRAINTS"),
                    notices: &[Diagnostic { severity: "WARNING", code: "25P01", message: "SET CONSTRAINTS can only be used in transaction blocks", ..E }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET CONSTRAINTS ALL DEFERRED;",
                    expected: Expected::Tag("SET CONSTRAINTS"),
                    notices: &[Diagnostic { severity: "WARNING", code: "25P01", message: "SET CONSTRAINTS can only be used in transaction blocks", ..E }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET CONSTRAINTS nope IMMEDIATE;",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"constraint "nope" does not exist"#, ..E }),
                    notices: &[Diagnostic { severity: "WARNING", code: "25P01", message: "SET CONSTRAINTS can only be used in transaction blocks", ..E }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET CONSTRAINTS ALL IMMEDIATE;",
                    expected: Expected::Tag("SET CONSTRAINTS"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET CONSTRAINTS ALL DEFERRED;",
                    expected: Expected::Tag("SET CONSTRAINTS"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET CONSTRAINTS nd_fk, parent_example_pkey, uniq_u, chk, dom_chk IMMEDIATE;",
                    expected: Expected::Tag("SET CONSTRAINTS"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET CONSTRAINTS ND_FK IMMEDIATE;",
                    expected: Expected::Tag("SET CONSTRAINTS"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET CONSTRAINTS public.nd_fk IMMEDIATE;",
                    expected: Expected::Tag("SET CONSTRAINTS"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET CONSTRAINTS postgres.public.nd_fk IMMEDIATE;",
                    expected: Expected::Tag("SET CONSTRAINTS"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET CONSTRAINTS s2.t2_pkey IMMEDIATE;",
                    expected: Expected::Tag("SET CONSTRAINTS"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET CONSTRAINTS t2_pkey IMMEDIATE;",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"constraint "t2_pkey" does not exist"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ROLLBACK;",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SET CONSTRAINTS "ND_FK" IMMEDIATE;"#,
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"constraint "ND_FK" does not exist"#, ..E }),
                    notices: &[Diagnostic { severity: "WARNING", code: "25P01", message: "SET CONSTRAINTS can only be used in transaction blocks", ..E }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET CONSTRAINTS public.t2_pkey IMMEDIATE;",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"constraint "t2_pkey" does not exist"#, ..E }),
                    notices: &[Diagnostic { severity: "WARNING", code: "25P01", message: "SET CONSTRAINTS can only be used in transaction blocks", ..E }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET CONSTRAINTS nd_fk, nope IMMEDIATE;",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"constraint "nope" does not exist"#, ..E }),
                    notices: &[Diagnostic { severity: "WARNING", code: "25P01", message: "SET CONSTRAINTS can only be used in transaction blocks", ..E }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET CONSTRAINTS nosuch.nd_fk IMMEDIATE;",
                    expected: Expected::Error(Diagnostic { code: "3F000", message: r#"schema "nosuch" does not exist"#, ..E }),
                    notices: &[Diagnostic { severity: "WARNING", code: "25P01", message: "SET CONSTRAINTS can only be used in transaction blocks", ..E }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET CONSTRAINTS otherdb.public.nd_fk IMMEDIATE;",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"cross-database references are not implemented: "otherdb.public.nd_fk""#, ..E }),
                    notices: &[Diagnostic { severity: "WARNING", code: "25P01", message: "SET CONSTRAINTS can only be used in transaction blocks", ..E }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET CONSTRAINTS nd_fk DEFERRED;",
                    expected: Expected::Error(Diagnostic { code: "42809", message: r#"constraint "nd_fk" is not deferrable"#, ..E }),
                    notices: &[Diagnostic { severity: "WARNING", code: "25P01", message: "SET CONSTRAINTS can only be used in transaction blocks", ..E }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET CONSTRAINTS dom_chk DEFERRED;",
                    expected: Expected::Error(Diagnostic { code: "42809", message: r#"constraint "dom_chk" is not deferrable"#, ..E }),
                    notices: &[Diagnostic { severity: "WARNING", code: "25P01", message: "SET CONSTRAINTS can only be used in transaction blocks", ..E }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET CONSTRAINTS nd_fk, nope DEFERRED;",
                    expected: Expected::Error(Diagnostic { code: "42809", message: r#"constraint "nd_fk" is not deferrable"#, ..E }),
                    notices: &[Diagnostic { severity: "WARNING", code: "25P01", message: "SET CONSTRAINTS can only be used in transaction blocks", ..E }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path = s2, public;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET CONSTRAINTS t2_pkey, nd_fk IMMEDIATE;",
                    expected: Expected::Tag("SET CONSTRAINTS"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "COMMIT;",
                    expected: Expected::Tag("COMMIT"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET CONSTRAINTS named_pk IMMEDIATE;",
                    expected: Expected::Tag("SET CONSTRAINTS"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ROLLBACK;",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET CONSTRAINTS a.b.c.d IMMEDIATE;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "improper qualified name (too many dotted names): a.b.c.d", position: 17, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET CONSTRAINTS ALL DEFERRED;",
                    expected: Expected::Tag("SET CONSTRAINTS"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO child_example VALUES (1, 99);",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "child_example" violates foreign key constraint "nd_fk""#, detail: r#"Key (p2)=(99) is not present in table "parent_example"."#, schema: "public", table: "child_example", constraint: "nd_fk", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ROLLBACK;",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "SET CONSTRAINTS with deferred foreign keys",
            set_up_script: &[
                "CREATE TABLE parent_example (id INTEGER PRIMARY KEY);",
                "CREATE TABLE child_example (id INTEGER PRIMARY KEY, parent_id INTEGER CONSTRAINT child_parent_fk REFERENCES parent_example(id) DEFERRABLE INITIALLY DEFERRED);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO child_example VALUES (1, 10);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO parent_example VALUES (10);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "COMMIT;",
                    expected: Expected::Tag("COMMIT"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO child_example VALUES (2, 20);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "COMMIT;",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "child_example" violates foreign key constraint "child_parent_fk""#, detail: r#"Key (parent_id)=(20) is not present in table "parent_example"."#, schema: "public", table: "child_example", constraint: "child_parent_fk", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO child_example VALUES (3, 30);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET CONSTRAINTS child_parent_fk IMMEDIATE;",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "child_example" violates foreign key constraint "child_parent_fk""#, detail: r#"Key (parent_id)=(30) is not present in table "parent_example"."#, schema: "public", table: "child_example", constraint: "child_parent_fk", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ROLLBACK;",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET CONSTRAINTS ALL IMMEDIATE;",
                    expected: Expected::Tag("SET CONSTRAINTS"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO child_example VALUES (4, 40);",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "child_example" violates foreign key constraint "child_parent_fk""#, detail: r#"Key (parent_id)=(40) is not present in table "parent_example"."#, schema: "public", table: "child_example", constraint: "child_parent_fk", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ROLLBACK;",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM child_example;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("parent_id", INT4)],
                        rows: &[
                            &[T("1"), T("10")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET CONSTRAINTS child_parent_fk DEFERRED;",
                    expected: Expected::Tag("SET CONSTRAINTS"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET CONSTRAINTS child_parent_fk IMMEDIATE;",
                    expected: Expected::Tag("SET CONSTRAINTS"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ROLLBACK;",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "SET CONSTRAINTS with deferrable constraints",
            set_up_script: &[
                "CREATE TABLE p (id INTEGER PRIMARY KEY);",
                "CREATE TABLE fk_imm (id INTEGER PRIMARY KEY, pid INTEGER CONSTRAINT fk_imm_fk REFERENCES p(id) DEFERRABLE INITIALLY IMMEDIATE);",
                "CREATE TABLE u_def (id INTEGER PRIMARY KEY, v INTEGER CONSTRAINT u_def_u UNIQUE DEFERRABLE INITIALLY DEFERRED);",
                "CREATE TABLE pk_def (id INTEGER CONSTRAINT pk_def_pk PRIMARY KEY DEFERRABLE, v INTEGER);",
                "INSERT INTO u_def VALUES (1, 1), (2, 2);",
                "INSERT INTO pk_def VALUES (1, 1), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO fk_imm VALUES (1, 10);",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "fk_imm" violates foreign key constraint "fk_imm_fk""#, detail: r#"Key (pid)=(10) is not present in table "p"."#, schema: "public", table: "fk_imm", constraint: "fk_imm_fk", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ROLLBACK;",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET CONSTRAINTS fk_imm_fk DEFERRED;",
                    expected: Expected::Tag("SET CONSTRAINTS"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO fk_imm VALUES (1, 10);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO p VALUES (10);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "COMMIT;",
                    expected: Expected::Tag("COMMIT"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET CONSTRAINTS ALL DEFERRED;",
                    expected: Expected::Tag("SET CONSTRAINTS"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO fk_imm VALUES (2, 20);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO p VALUES (20);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "COMMIT;",
                    expected: Expected::Tag("COMMIT"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM fk_imm ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("pid", INT4)],
                        rows: &[
                            &[T("1"), T("10")],
                            &[T("2"), T("20")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE u_def SET v = 2 WHERE id = 1;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE u_def SET v = 1 WHERE id = 2;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "COMMIT;",
                    expected: Expected::Tag("COMMIT"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM u_def ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", INT4)],
                        rows: &[
                            &[T("1"), T("2")],
                            &[T("2"), T("1")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE u_def SET v = 1 WHERE id = 1;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "COMMIT;",
                    expected: Expected::Error(Diagnostic { code: "23505", message: r#"duplicate key value violates unique constraint "u_def_u""#, detail: "Key (v)=(1) already exists.", schema: "public", table: "u_def", constraint: "u_def_u", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE u_def SET v = 1 WHERE id = 1;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET CONSTRAINTS u_def_u IMMEDIATE;",
                    expected: Expected::Error(Diagnostic { code: "23505", message: r#"duplicate key value violates unique constraint "u_def_u""#, detail: "Key (v)=(1) already exists.", schema: "public", table: "u_def", constraint: "u_def_u", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ROLLBACK;",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET CONSTRAINTS pk_def_pk DEFERRED;",
                    expected: Expected::Tag("SET CONSTRAINTS"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE pk_def SET id = 2 WHERE v = 1;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE pk_def SET id = 1 WHERE v = 2;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "COMMIT;",
                    expected: Expected::Tag("COMMIT"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM pk_def ORDER BY v;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", INT4)],
                        rows: &[
                            &[T("2"), T("1")],
                            &[T("1"), T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE pk_def SET id = id + 10;",
                    expected: Expected::Tag("UPDATE 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM pk_def ORDER BY v;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", INT4)],
                        rows: &[
                            &[T("12"), T("1")],
                            &[T("11"), T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "SET CONSTRAINTS with deferred parent deletions",
            set_up_script: &[
                "CREATE TABLE dp (id INTEGER PRIMARY KEY);",
                "CREATE TABLE dc (id INTEGER PRIMARY KEY, pid INTEGER CONSTRAINT dc_fk REFERENCES dp(id) DEFERRABLE INITIALLY DEFERRED);",
                "CREATE TABLE du (id INTEGER PRIMARY KEY, v INTEGER CONSTRAINT du_v UNIQUE DEFERRABLE);",
                "INSERT INTO dp VALUES (1);",
                "INSERT INTO dc VALUES (1, 1);",
                "INSERT INTO du VALUES (1, 1), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM dp WHERE id = 1;",
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO dp VALUES (1);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "COMMIT;",
                    expected: Expected::Tag("COMMIT"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM dp WHERE id = 1;",
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "COMMIT;",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"update or delete on table "dp" violates foreign key constraint "dc_fk" on table "dc""#, detail: r#"Key (id)=(1) is still referenced from table "dc"."#, schema: "public", table: "dc", constraint: "dc_fk", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO dc VALUES (2, 5);",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "dc" violates foreign key constraint "dc_fk""#, detail: r#"Key (pid)=(5) is not present in table "dp"."#, schema: "public", table: "dc", constraint: "dc_fk", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO dc VALUES (3, 7);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM dc WHERE id = 3;",
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "COMMIT;",
                    expected: Expected::Tag("COMMIT"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM dc ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("pid", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE du SET v = v + 1;",
                    expected: Expected::Tag("UPDATE 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM du ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", INT4)],
                        rows: &[
                            &[T("1"), T("2")],
                            &[T("2"), T("3")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT indimmediate FROM pg_index WHERE indexrelid = 'du_v'::regclass;",
                    expected: Expected::Rows {
                        columns: &[Column("indimmediate", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT constraint_name, is_deferrable, initially_deferred FROM information_schema.table_constraints WHERE table_name IN ('dc', 'du') AND constraint_type <> 'CHECK' ORDER BY constraint_name;",
                    expected: Expected::Rows {
                        columns: &[Column("constraint_name", NAME), Column("is_deferrable", VARCHAR), Column("initially_deferred", VARCHAR)],
                        rows: &[
                            &[T("dc_fk"), T("YES"), T("YES")],
                            &[T("dc_pkey"), T("NO"), T("NO")],
                            &[T("du_pkey"), T("NO"), T("NO")],
                            &[T("du_v"), T("YES"), T("NO")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}
