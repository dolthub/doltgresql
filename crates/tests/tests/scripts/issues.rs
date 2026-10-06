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
fn test_issues() {
    run_scripts(&[
        ScriptTest {
            name: "Issue #25: double-quoted args to dolt functions treated as identifiers",
            set_up_script: &[
                "create table tbl (pk int);",
                "insert into tbl values (1);",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"select dolt_add(".");"#,
                    expected: Expected::Error(Diagnostic { code: "42703", message: r#"column "." could not be found in any table in scope"#, ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_add('.');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_add", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"select dolt_commit("-m", "look ma");"#,
                    expected: Expected::Error(Diagnostic { code: "42703", message: r#"column "-m" could not be found in any table in scope"#, ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select length(dolt_commit('-m', 'look ma')::text);",
                    expected: Expected::Rows {
                        columns: &[Column("length", INT4)],
                        rows: &[
                            &[T("32")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"select dolt_branch("br1");"#,
                    expected: Expected::Error(Diagnostic { code: "42703", message: r#"column "br1" could not be found in any table in scope"#, ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_branch('br1');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_branch", INT8)],
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
        ScriptTest {
            name: "Issue #2030: Drizzle relational query fails with BigInt error",
            set_up_script: &[
                r#"CREATE TABLE sub_entities (
  project_id VARCHAR(256) NOT NULL,
  entity_id  VARCHAR(256) NOT NULL,
  id         VARCHAR(256) NOT NULL,
  name       VARCHAR(256) NOT NULL,
  PRIMARY KEY (project_id, entity_id, id)
);
"#,
                r#"
CREATE TABLE entities (
  project_id              VARCHAR(256) NOT NULL,
  id                      VARCHAR(256) NOT NULL,
  name                    VARCHAR(256) NOT NULL,
  default_sub_entity_id   VARCHAR(256),
  PRIMARY KEY (project_id, id)
);
"#,
                r#"
CREATE TABLE conversations (
  id                 VARCHAR(256) NOT NULL,
  tenant_id          VARCHAR(256) NOT NULL,
  project_id         VARCHAR(256) NOT NULL,
  active_sub_agent_id VARCHAR(256) NOT NULL,
  PRIMARY KEY (tenant_id, project_id, id)
);
"#,
                r#"INSERT INTO sub_entities (project_id, entity_id, id, name) VALUES
  ('projectA', 'entityA', 'subA1', 'Sub-Entity A1'),
  ('projectA', 'entityB', 'subB1', 'Sub-Entity B1');
"#,
                r#"INSERT INTO entities (project_id, id, name, default_sub_entity_id) VALUES
  ('projectA', 'entityA', 'Entity A', 'subA1'),
  ('projectA', 'entityB', 'Entity B', 'subB1');
"#,
                r#"INSERT INTO conversations (tenant_id, project_id, id, active_sub_agent_id) VALUES
  ('tenant1', 'projectA', 'conv1', 'subA1'),
  ('tenant1', 'projectA', 'conv2', 'subB1');
"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"select
  "entities"."project_id",
  "entities"."id",
  "entities"."name",
  "entities"."default_sub_entity_id",
  "entities_defaultSubEntity"."data" as "defaultSubEntity"
from "entities" "entities"
left join lateral (
  select json_build_array(
           "entities_defaultSubEntity"."project_id",
           "entities_defaultSubEntity"."entity_id",
           "entities_defaultSubEntity"."id",
           "entities_defaultSubEntity"."name"
         ) as "data"
  from (
    select * from "sub_entities" "entities_defaultSubEntity"
    where "entities_defaultSubEntity"."id" = "entities"."default_sub_entity_id"
    limit $1
  ) "entities_defaultSubEntity"
) "entities_defaultSubEntity" on true
where ("entities"."project_id" = $2 and "entities"."id" = $3)
limit $4"#,
                    bind_vars: &[BindVar::Int64(1), BindVar::Str("projectA"), BindVar::Str("entityA"), BindVar::Int64(1)],
                    expected: Expected::Rows {
                        columns: &[Column("project_id", VARCHAR), Column("id", VARCHAR), Column("name", VARCHAR), Column("default_sub_entity_id", VARCHAR), Column("defaultSubEntity", JSON)],
                        rows: &[
                            &[T("projectA"), T("entityA"), T("Entity A"), T("subA1"), T(r#"["projectA", "entityA", "subA1", "Sub-Entity A1"]"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"select
  "entities"."project_id",
  "entities"."id",
  "entities"."name",
  "entities"."default_sub_entity_id",
  "entities_defaultSubEntity"."data" as "defaultSubEntity"
from "entities" "entities"
left join lateral (
  select json_build_array(
           "entities_defaultSubEntity"."project_id",
           "entities_defaultSubEntity"."entity_id",
           "entities_defaultSubEntity"."id",
           "entities_defaultSubEntity"."name"
         ) as "data"
  from (
    select * from "sub_entities" "entities_defaultSubEntity"
    where "entities_defaultSubEntity"."id" = "entities"."default_sub_entity_id"
    limit 1
  ) "entities_defaultSubEntity"
) "entities_defaultSubEntity" on true
where ("entities"."project_id" = 'projectA' and "entities"."id" = 'entityA')
limit 1"#,
                    expected: Expected::Rows {
                        columns: &[Column("project_id", VARCHAR), Column("id", VARCHAR), Column("name", VARCHAR), Column("default_sub_entity_id", VARCHAR), Column("defaultSubEntity", JSON)],
                        rows: &[
                            &[T("projectA"), T("entityA"), T("Entity A"), T("subA1"), T(r#"["projectA", "entityA", "subA1", "Sub-Entity A1"]"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Issue #2049: bad control character in JSON string literal",
            set_up_script: &[
                r#"CREATE TABLE jsonb_test (id VARCHAR(256) NOT NULL PRIMARY KEY, "jsonbColumn" JSONB);"#,
                r#"INSERT INTO jsonb_test VALUES ('test', '{"test": "value\n"}');"#,
                r#"INSERT INTO jsonb_test VALUES ('test2', '{"test": "value\t"}');"#,
                r#"INSERT INTO jsonb_test VALUES ('test3', '{"test": "value\r"}');"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM jsonb_test;",
                    expected: Expected::Rows {
                        columns: &[Column("id", VARCHAR), Column("jsonbColumn", JSONB)],
                        rows: &[
                            &[T("test"), T(r#"{"test": "value\n"}"#)],
                            &[T("test2"), T(r#"{"test": "value\t"}"#)],
                            &[T("test3"), T(r#"{"test": "value\r"}"#)],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Issue #2197 Part 1: bugs with table composite types",
            set_up_script: &[
                "CREATE TABLE t1 (a INT, b VARCHAR(3));",
                "CREATE TABLE t2(id SERIAL, t1 t1);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO t2(t1) VALUES (ROW(1, 'abc'));",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t2;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("t1", USER_DEFINED)],
                        rows: &[
                            &[T("1"), T("(1,abc)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t2(t1) VALUES (ROW('a', 'def'));",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type integer: "a""#, position: 32, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t2(t1) VALUES (ROW(true, 'def'));",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type record to t1", detail: "Cannot cast type boolean to integer in column 1.", position: 32, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t2(t1) VALUES (ROW(2, 'def', 'ghi'));",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type record to t1", detail: "Input has too many columns.", position: 28, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t2(t1) VALUES (ROW(2));",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type record to t1", detail: "Input has too few columns.", position: 28, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Issue #2197 Part 2: bugs with table composite types",
            set_up_script: &[
                "CREATE TABLE t1a (a INT4, b VARCHAR(3));",
                "CREATE TABLE t1b (a INT4 NOT NULL, b VARCHAR(3) NOT NULL);",
                "CREATE TABLE t2 (id SERIAL, t1a t1a, t1b t1b);",
                "INSERT INTO t2 (t1a) VALUES (ROW(1, 'abc'));",
                "INSERT INTO t2 (t1b) VALUES (ROW(1, 'abc'));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t2;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("t1a", USER_DEFINED), Column("t1b", USER_DEFINED)],
                        rows: &[
                            &[T("1"), T("(1,abc)"), Null],
                            &[T("2"), Null, T("(1,abc)")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t1a ADD COLUMN c VARCHAR(10);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t1b ADD COLUMN c VARCHAR(10) NOT NULL;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t2 ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("t1a", USER_DEFINED), Column("t1b", USER_DEFINED)],
                        rows: &[
                            &[T("1"), T("(1,abc,)"), Null],
                            &[T("2"), Null, T("(1,abc,)")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t1a DROP COLUMN b;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t1b DROP COLUMN b;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t2 ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("t1a", USER_DEFINED), Column("t1b", USER_DEFINED)],
                        rows: &[
                            &[T("1"), T("(1,)"), Null],
                            &[T("2"), Null, T("(1,)")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t1a VALUES (2, 'def');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t1b VALUES (3, 'xyzzy');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t2 (t1a) SELECT ROW(a,c)::t1a FROM t1a;",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t2 (t1b) SELECT ROW(a,c)::t1b FROM t1b;",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t2 ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("t1a", USER_DEFINED), Column("t1b", USER_DEFINED)],
                        rows: &[
                            &[T("1"), T("(1,)"), Null],
                            &[T("2"), Null, T("(1,)")],
                            &[T("3"), T("(2,def)"), Null],
                            &[T("4"), Null, T("(3,xyzzy)")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ((t1a).@1), ((t1b).@2) FROM t2 ORDER BY id;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "@""#, position: 15, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t2 SET t1a=ROW((t1a).a+100, (t1a).c)::t1a WHERE length(t1a::text) > 0;",
                    expected: Expected::Tag("UPDATE 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t2 SET t1b=ROW((t1b).@1+100, (t1b).@2)::t1b WHERE length(t1b::text) > 0;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "@""#, position: 29, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t2 ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("t1a", USER_DEFINED), Column("t1b", USER_DEFINED)],
                        rows: &[
                            &[T("1"), T("(101,)"), Null],
                            &[T("2"), Null, T("(1,)")],
                            &[T("3"), T("(102,def)"), Null],
                            &[T("4"), Null, T("(3,xyzzy)")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (id).a FROM t2;",
                    expected: Expected::Error(Diagnostic { code: "42809", message: "column notation .a applied to type integer, which is not a composite type", position: 9, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (t1a).g FROM t2;",
                    expected: Expected::Error(Diagnostic { code: "42703", message: r#"column "g" not found in data type t1a"#, position: 9, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (t1a).@0 FROM t2;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "@""#, position: 14, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (t1a).@3 FROM t2;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "@""#, position: 14, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t1a ADD COLUMN d VARCHAR(10) DEFAULT 'abc';",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"cannot alter table "t1a" because column "t2.t1a" uses its row type"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t1a ADD COLUMN d VARCHAR(10);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t1a DROP COLUMN c;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t2 ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("t1a", USER_DEFINED), Column("t1b", USER_DEFINED)],
                        rows: &[
                            &[T("1"), T("(101,)"), Null],
                            &[T("2"), Null, T("(1,)")],
                            &[T("3"), T("(102,)"), Null],
                            &[T("4"), Null, T("(3,xyzzy)")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Issue #2299: nil pointer panic using DEFAULT on ENUM column",
            set_up_script: &[
                "CREATE TYPE team_role AS ENUM ('admin', 'editor', 'member');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE users (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), role team_role NOT NULL DEFAULT 'member');",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO users (role) VALUES (DEFAULT);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT role FROM users;",
                    expected: Expected::Rows {
                        columns: &[Column("role", USER_DEFINED)],
                        rows: &[
                            &[T("member")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Issue #2307: SELECT EXISTS returns INT2 instead of BOOL",
            set_up_script: &[
                "CREATE TABLE test (pk INT4);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT EXISTS(SELECT 1 FROM pg_catalog.pg_tables WHERE tablename = 'test');",
                    expected: Expected::Rows {
                        columns: &[Column("exists", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT NOT EXISTS(SELECT 1 FROM pg_catalog.pg_tables WHERE tablename = 'test');",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
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
            name: "Issue #2548: timezone parser rejects valid time zone offset formats",
            set_up_script: &[
                "CREATE TABLE test (pk INT4 PRIMARY KEY, v1 TIMESTAMP WITH TIME ZONE);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SET TimeZone = 'UTC-01:00';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES (1, '2026-04-15 10:11:12');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET TimeZone = 'UTC-03:00';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES (2, '2026-04-15 10:11:12');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (SELECT v1 FROM test WHERE pk = 2) - (SELECT v1 FROM test WHERE pk = 1);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INTERVAL)],
                        rows: &[
                            &[T("-02:00:00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Issue #2604: dolt_merge syntax error with unique index and DEFAULT",
            set_up_script: &[
                "CREATE TABLE t (id INT PRIMARY KEY, a TEXT, b TEXT DEFAULT 'x');",
                "CREATE UNIQUE INDEX idx_t_a ON t(a);",
                "SELECT dolt_add('-A');",
                "SELECT dolt_commit('-m', 'schema');",
                "SELECT dolt_branch('f', 'main');",
                "SELECT dolt_checkout('f');",
                "INSERT INTO t (id, a) VALUES (1, 'feat');",
                "SELECT dolt_add('-A');",
                "SELECT dolt_commit('-m', 'feat');",
                "SELECT dolt_checkout('main');",
                "INSERT INTO t (id, a) VALUES (2, 'main');",
                "SELECT dolt_add('-A');",
                "SELECT dolt_commit('-m', 'main');",
                "SELECT dolt_checkout('f');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT length(dolt_merge('main')::text) = 57;",
                    expected: Expected::Rows {
                        columns: &[Column("length = 57", BOOL)],
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
            name: "Issue #3116: dolt.branches.dirty does not have boolean output",
            set_up_script: &[
                "CREATE TABLE t3116 (id INT PRIMARY KEY);",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dirty FROM dolt.branches;",
                    expected: Expected::Rows {
                        columns: &[Column("dirty", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dirty FROM dolt.branches WHERE name = 'main';",
                    expected: Expected::Rows {
                        columns: &[Column("dirty", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dirty FROM dolt.branches ORDER BY name;",
                    expected: Expected::Rows {
                        columns: &[Column("dirty", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dirty FROM dolt.branches WHERE dirty = true;",
                    expected: Expected::Rows {
                        columns: &[Column("dirty", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT b.dirty FROM dolt.branches b ORDER BY b.name;",
                    expected: Expected::Rows {
                        columns: &[Column("dirty", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT staged FROM dolt.status ORDER BY table_name;",
                    expected: Expected::Rows {
                        columns: &[Column("staged", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT data_change, schema_change FROM dolt.diff ORDER BY table_name;",
                    expected: Expected::Rows {
                        columns: &[Column("data_change", BOOL), Column("schema_change", BOOL)],
                        rows: &[
                            &[T("f"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_commit('-Am', 'commit for issue 3116');",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dirty FROM dolt.branches WHERE name = 'main';",
                    expected: Expected::Rows {
                        columns: &[Column("dirty", BOOL)],
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
            name: "Issue #3138: WITH ORDINALITY in correlated subquery internal error",
            set_up_script: &[
                "CREATE TABLE bug16_parent (id integer PRIMARY KEY);",
                "CREATE TABLE bug16_child  (id integer PRIMARY KEY, parent_id integer REFERENCES bug16_parent(id));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT c.conname,
       array(SELECT colid FROM unnest(c.conkey) WITH ORDINALITY cols(colid, arridx) ORDER BY cols.arridx)
FROM pg_constraint c JOIN pg_class cl ON c.conrelid = cl.oid WHERE cl.relname = 'bug16_child' ORDER BY c.conname;"#,
                    expected: Expected::Rows {
                        columns: &[Column("conname", NAME), Column("array", INT2_ARRAY)],
                        rows: &[
                            &[T("bug16_child_parent_id_fkey"), T("{2}")],
                            &[T("bug16_child_pkey"), T("{1}")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT colid, arridx FROM unnest(ARRAY[1,2]) WITH ORDINALITY cols(colid, arridx);",
                    expected: Expected::Rows {
                        columns: &[Column("colid", INT4), Column("arridx", INT8)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Issue #3097: bind parameters described as unknown OID, breaking pgx",
            set_up_script: &[
                "CREATE TABLE g_bool (id INT4 PRIMARY KEY, flag BOOLEAN);",
                "INSERT INTO g_bool VALUES (1, true), (2, false), (3, NULL);",
                "CREATE TABLE g_arr (id INT4 PRIMARY KEY, v INT4, vals INT4[]);",
                "INSERT INTO g_arr VALUES (1, 1, ARRAY[1,2,3]), (2, 2, ARRAY[4,5,6]);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM g_bool WHERE flag IS DISTINCT FROM $1;",
                    bind_vars: &[BindVar::Bool(true)],
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM g_bool WHERE flag IS NOT DISTINCT FROM $1;",
                    bind_vars: &[BindVar::Bool(true)],
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM g_arr WHERE $1 = ANY(vals);",
                    bind_vars: &[BindVar::Int32(2)],
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM g_arr WHERE $1 = ANY(SELECT v FROM g_arr);",
                    bind_vars: &[BindVar::Int32(2)],
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM g_arr WHERE vals[$1] = 2;",
                    bind_vars: &[BindVar::Int32(2)],
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM g_arr WHERE coalesce($1, v) = 1;",
                    bind_vars: &[BindVar::Null],
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM g_arr WHERE greatest($1, v) = 2;",
                    bind_vars: &[BindVar::Int32(0)],
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM g_arr WHERE least($1, v) = 1;",
                    bind_vars: &[BindVar::Int32(5)],
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
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
            name: "Issue #3111 (set-returning functions and max1Row spooling)",
            set_up_script: &[
                "CREATE TABLE bug14 (a integer, b integer);",
                "CREATE INDEX bug14_ab ON bug14 (a, b);",
                "CREATE TABLE arrtbl (pk integer PRIMARY KEY, arr integer[]);",
                "CREATE UNIQUE INDEX arrtbl_pk ON arrtbl (pk);",
                "INSERT INTO arrtbl VALUES (1, '{10,20,30}'), (2, '{40}');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT unnest(indkey) FROM pg_index WHERE indexrelid = 'bug14_ab'::regclass;",
                    expected: Expected::Rows {
                        columns: &[Column("unnest", INT2)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT unnest(arr) FROM arrtbl WHERE pk = 1;",
                    expected: Expected::Rows {
                        columns: &[Column("unnest", INT4)],
                        rows: &[
                            &[T("10")],
                            &[T("20")],
                            &[T("30")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT generate_series(1, 3) FROM arrtbl WHERE pk = 2;",
                    expected: Expected::Rows {
                        columns: &[Column("generate_series", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                            &[T("3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Issue #3083: partial index on a non-empty table",
            set_up_script: &[
                "CREATE TABLE p (id int PRIMARY KEY, flag boolean NOT NULL DEFAULT false);",
                "INSERT INTO p (id) VALUES (1);",
                "CREATE TABLE t (id int PRIMARY KEY, k int NOT NULL, flag boolean NOT NULL DEFAULT false);",
                "INSERT INTO t (id, k, flag) VALUES (1, 10, false), (2, 20, true), (3, 30, false);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE INDEX p_flagged ON p (id) WHERE flag;",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM p WHERE flag AND id > 0 ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO p (id, flag) VALUES (2, true);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM p WHERE flag AND id > 0 ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX t_partial ON t (k) WHERE flag;",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE UNIQUE INDEX t_partial_uniq ON t (k) WHERE flag;",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Issue #3333: CHECK constraint calling a function",
            set_up_script: &[
                "CREATE TABLE t3333_issue (z TEXT CHECK (regexp_like(z::text, '^[0-9]+$')));",
                "CREATE TABLE t3333 (z TEXT PRIMARY KEY CHECK (z ~ '^[0-9]+$'), y TEXT CONSTRAINT y_chk CHECK (regexp_like(y, '^[a-z]+$')));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO t3333_issue VALUES ('12345');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t3333_issue VALUES ('abc');",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"new row for relation "t3333_issue" violates check constraint "t3333_issue_z_check""#, detail: "Failing row contains (abc).", schema: "public", table: "t3333_issue", constraint: "t3333_issue_z_check", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT z FROM t3333_issue;",
                    expected: Expected::Rows {
                        columns: &[Column("z", TEXT)],
                        rows: &[
                            &[T("12345")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t3333 VALUES ('123', 'abc');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t3333 VALUES ('12a', 'abc');",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"new row for relation "t3333" violates check constraint "t3333_z_check""#, detail: "Failing row contains (12a, abc).", schema: "public", table: "t3333", constraint: "t3333_z_check", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t3333 VALUES ('124', 'ABC');",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"new row for relation "t3333" violates check constraint "y_chk""#, detail: "Failing row contains (124, ABC).", schema: "public", table: "t3333", constraint: "y_chk", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT check_clause FROM information_schema.check_constraints WHERE constraint_name = 'y_chk';",
                    expected: Expected::Rows {
                        columns: &[Column("check_clause", VARCHAR)],
                        rows: &[
                            &[T("(regexp_like(y, '^[a-z]+$'::text))")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t3333;",
                    expected: Expected::Rows {
                        columns: &[Column("z", TEXT), Column("y", TEXT)],
                        rows: &[
                            &[T("123"), T("abc")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Issue #3366 (multi-array unnest)",
            set_up_script: &[
                "CREATE TABLE bulk_example (id integer PRIMARY KEY, label text);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM unnest(ARRAY[1, 2]::integer[], ARRAY[3, 4]::integer[]);",
                    expected: Expected::Rows {
                        columns: &[Column("unnest", INT4), Column("unnest", INT4)],
                        rows: &[
                            &[T("1"), T("3")],
                            &[T("2"), T("4")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO bulk_example (id, label) SELECT * FROM unnest(ARRAY[1, 2]::integer[], ARRAY['a', 'b']::text[]);",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM bulk_example ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("label", TEXT)],
                        rows: &[
                            &[T("1"), T("a")],
                            &[T("2"), T("b")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO bulk_example (id, label) SELECT * FROM unnest(ARRAY[1, 3]::integer[], ARRAY['c', 'd']::text[]) ON CONFLICT (id) DO UPDATE SET label = EXCLUDED.label RETURNING *;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("label", TEXT)],
                        rows: &[
                            &[T("1"), T("c")],
                            &[T("3"), T("d")],
                        ],
                        tag: "INSERT 0 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM bulk_example ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("label", TEXT)],
                        rows: &[
                            &[T("1"), T("c")],
                            &[T("2"), T("b")],
                            &[T("3"), T("d")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}
