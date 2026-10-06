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
fn test_user_space_dolt_tables() {
    run_scripts(&[
        ScriptTest {
            name: "dolt branches",
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT name FROM dolt.branches",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("main")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT name FROM dolt_branches",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("main")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT name FROM public.dolt_branches",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("main")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT branches.name FROM dolt.branches",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("main")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt.branches.name FROM dolt.branches",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("main")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_branches.name FROM dolt_branches",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("main")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM public.branches",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: branches", ..E }),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM branches",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: branches", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE branches (id INT PRIMARY KEY)",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO branches VALUES (1)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM branches",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT name FROM dolt.branches",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("main")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE SCHEMA dolt",
                    expected: Expected::Error(Diagnostic { code: "42P06", message: "can't create schema dolt; schema exists", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path = 'dolt'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT name FROM branches",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("main")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM public.branches",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path = 'public'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM branches",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path = public, dolt",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM branches",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM BRANCHES",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"SELECT "dolt_branches"."name" FROM "dolt_branches" WHERE "dolt_branches"."name" IN ('main') ORDER BY "dolt_branches"."name" DESC LIMIT 21;"#,
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("main")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT name FROM dolt.branches WHERE name IN ('main')",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("main")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT name FROM dolt.branches WHERE name IN ('main', 'nonexistent')",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("main")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT name FROM dolt.branches WHERE name NOT IN ('nonexistent')",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("main")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "dolt blame with tablename",
            set_up_script: &[
                "CREATE TABLE test (id INT PRIMARY KEY)",
                "INSERT INTO test VALUES (1)",
                "SELECT dolt_commit('-Am', 'test commit', '--author', 'John Doe <johndoe@example.com>')",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"WITH sorted_diffs_by_pk
									AS (SELECT
													"to_id",
													to_commit,
													to_commit_date,
													diff_type,
													ROW_NUMBER() OVER (
															PARTITION BY coalesce("to_id", "from_id")
															ORDER BY coalesce(to_commit_date, from_commit_date) DESC
													) row_num
											FROM "dolt_diff_test"
										)
									SELECT
											sd."to_id" AS "id",
											dl.committer,
											dl.email,
											dl.message
									FROM
											sorted_diffs_by_pk as sd,
											dolt_log as dl
									WHERE
											dl.commit_hash = sd.to_commit
											and sd.row_num = 1
											and sd.diff_type <> 'removed'
									ORDER BY
													sd."to_id" ASC;"#,
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("committer", TEXT), Column("email", TEXT), Column("message", TEXT)],
                        rows: &[
                            &[T("1"), T("John Doe"), T("johndoe@example.com"), T("test commit")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT id, committer FROM dolt_blame_test",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"at or near "`": syntax error"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT id, committer FROM public.dolt_blame_test",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"at or near "`": syntax error"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_blame_test.id FROM public.dolt_blame_test",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"at or near "`": syntax error"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM other.dolt_blame_test",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: dolt_blame_test", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE SCHEMA newschema",
                    expected: Expected::Tag("CREATE SCHEMA"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SET search_path = 'newschema'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE TABLE test_sch (id INT PRIMARY KEY)",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO test_sch VALUES (11)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_commit('-Am', 'add test_sch')",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_commit", TEXT)],
                        rows: &[
                            &[T("urfg5vsfrg707j3vfdv39o1k2dnhn7pr")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT id FROM newschema.dolt_blame_test_sch",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"at or near "`": syntax error"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT id, committer FROM public.dolt_blame_test",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"at or near "`": syntax error"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "dolt column diff",
            set_up_script: &[
                "CREATE TABLE test (id INT PRIMARY KEY)",
                "SELECT dolt_commit('-Am', 'test commit')",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name, column_name FROM dolt.column_diff",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("column_name", TEXT)],
                        rows: &[
                            &[T("public.test"), T("id")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name, column_name FROM dolt_column_diff",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("column_name", TEXT)],
                        rows: &[
                            &[T("public.test"), T("id")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt.column_diff.table_name FROM dolt.column_diff",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT)],
                        rows: &[
                            &[T("public.test")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_column_diff.table_name, dolt_column_diff.column_name FROM dolt_column_diff",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("column_name", TEXT)],
                        rows: &[
                            &[T("public.test"), T("id")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM public.column_diff",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: column_diff", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM column_diff",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: column_diff", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE TABLE column_diff (id INT PRIMARY KEY)",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO column_diff VALUES (1)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM column_diff",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name, column_name FROM dolt.column_diff WHERE table_name = 'public.test'",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("column_name", TEXT)],
                        rows: &[
                            &[T("public.test"), T("id")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SET search_path = 'dolt'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name, column_name FROM column_diff WHERE table_name = 'public.test'",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("column_name", TEXT)],
                        rows: &[
                            &[T("public.test"), T("id")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM public.column_diff",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SET search_path = 'public'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM column_diff",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SET search_path = public, dolt",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM column_diff",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM COLUMN_DIFF",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
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
            name: "dolt commit ancestors",
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM dolt.commit_ancestors",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM dolt_commit_ancestors",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt.commit_ancestors.parent_index FROM dolt.commit_ancestors",
                    expected: Expected::Rows {
                        columns: &[Column("parent_index", INT4)],
                        rows: &[
                            &[T("0")],
                            &[T("0")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_commit_ancestors.parent_index FROM dolt_commit_ancestors",
                    expected: Expected::Rows {
                        columns: &[Column("parent_index", INT4)],
                        rows: &[
                            &[T("0")],
                            &[T("0")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM public.commit_ancestors",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: commit_ancestors", ..E }),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM commit_ancestors",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: commit_ancestors", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE commit_ancestors (id INT PRIMARY KEY)",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO commit_ancestors VALUES (1)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM commit_ancestors",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM dolt.commit_ancestors",
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
                    query: "SET search_path = 'dolt'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM commit_ancestors",
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
                    query: "SELECT * FROM public.commit_ancestors",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path = 'public'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM commit_ancestors",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path = public, dolt",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM commit_ancestors",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM COMMIT_ANCESTORS",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
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
            name: "dolt commit diff with tablename",
            set_up_script: &[
                "CREATE TABLE test (id INT PRIMARY KEY)",
                "INSERT INTO test VALUES (10)",
                "SELECT dolt_commit('-Am', 'test commit 1')",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_id, to_id, diff_type FROM dolt_commit_diff_test WHERE from_commit=HASHOF('HEAD^1') AND to_commit=HASHOF('HEAD')",
                    expected: Expected::Rows {
                        columns: &[Column("from_id", INT4), Column("to_id", INT4), Column("diff_type", TEXT)],
                        rows: &[
                            &[Null, T("10"), T("added")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_id, to_id, diff_type FROM public.dolt_commit_diff_test WHERE from_commit=HASHOF('HEAD^1') AND to_commit=HASHOF('HEAD')",
                    expected: Expected::Rows {
                        columns: &[Column("from_id", INT4), Column("to_id", INT4), Column("diff_type", TEXT)],
                        rows: &[
                            &[Null, T("10"), T("added")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_commit_diff_test.to_id FROM public.dolt_commit_diff_test WHERE from_commit=HASHOF('HEAD^1') AND to_commit=HASHOF('HEAD')",
                    expected: Expected::Rows {
                        columns: &[Column("to_id", INT4)],
                        rows: &[
                            &[T("10")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM other.dolt_commit_diff_test",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: dolt_commit_diff_test", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM public.dolt_commit_diff_none",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: dolt_commit_diff_none", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE SCHEMA newschema",
                    expected: Expected::Tag("CREATE SCHEMA"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SET search_path = 'newschema'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE TABLE test_sch (id INT PRIMARY KEY)",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO test_sch VALUES (11)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_commit('-Am', 'add test_sch')",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_id, to_id, diff_type FROM newschema.dolt_commit_diff_test_sch WHERE from_commit=HASHOF('HEAD^1') AND to_commit=HASHOF('HEAD')",
                    expected: Expected::Rows {
                        columns: &[Column("from_id", INT4), Column("to_id", INT4), Column("diff_type", TEXT)],
                        rows: &[
                            &[Null, T("11"), T("added")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_id, to_id, diff_type FROM dolt_commit_diff_test_sch WHERE from_commit=HASHOF('HEAD^1') AND to_commit=HASHOF('HEAD')",
                    expected: Expected::Rows {
                        columns: &[Column("from_id", INT4), Column("to_id", INT4), Column("diff_type", TEXT)],
                        rows: &[
                            &[Null, T("11"), T("added")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_id, to_id, diff_type FROM dolt_commit_diff_test WHERE from_commit=HASHOF('HEAD^1') AND to_commit=HASHOF('HEAD')",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: dolt_commit_diff_test", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT to_id, diff_type FROM public.dolt_commit_diff_test WHERE from_commit=HASHOF('HEAD^2') AND to_commit=HASHOF('HEAD^1')",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "invalid ancestor spec", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT to_id FROM public.dolt_commit_diff_test_sch WHERE from_commit=HASHOF('HEAD^2') AND to_commit=HASHOF('HEAD^1')",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: dolt_commit_diff_test_sch", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT to_id, diff_type FROM newschema.dolt_commit_diff_test WHERE from_commit=HASHOF('HEAD^1') AND to_commit=HASHOF('HEAD')",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: dolt_commit_diff_test", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE TABLE test (id INT PRIMARY KEY)",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES (12)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_commit('-Am', 'add test')",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_id, to_id, diff_type FROM newschema.dolt_commit_diff_test WHERE from_commit=HASHOF('HEAD~1') AND to_commit=HASHOF('HEAD')",
                    expected: Expected::Rows {
                        columns: &[Column("from_id", INT4), Column("to_id", INT4), Column("diff_type", TEXT)],
                        rows: &[
                            &[Null, T("12"), T("added")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_id, to_id, diff_type FROM dolt_commit_diff_test WHERE from_commit=HASHOF('HEAD~1') AND to_commit=HASHOF('HEAD')",
                    expected: Expected::Rows {
                        columns: &[Column("from_id", INT4), Column("to_id", INT4), Column("diff_type", TEXT)],
                        rows: &[
                            &[Null, T("12"), T("added")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_id, to_id, diff_type FROM public.dolt_commit_diff_test WHERE from_commit=HASHOF('HEAD~3') AND to_commit=HASHOF('HEAD~2')",
                    expected: Expected::Rows {
                        columns: &[Column("from_id", INT4), Column("to_id", INT4), Column("diff_type", TEXT)],
                        rows: &[
                            &[Null, T("10"), T("added")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "dolt commits",
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM dolt.commits",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM dolt_commits",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt.commits.message FROM dolt.commits",
                    expected: Expected::Rows {
                        columns: &[Column("message", TEXT)],
                        rows: &[
                            &[T("CREATE DATABASE")],
                            &[T("Initialize data repository")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_commits.message FROM dolt_commits",
                    expected: Expected::Rows {
                        columns: &[Column("message", TEXT)],
                        rows: &[
                            &[T("CREATE DATABASE")],
                            &[T("Initialize data repository")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM public.commits",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: commits", ..E }),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM commits",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: commits", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE commits (id INT PRIMARY KEY)",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO commits VALUES (1)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM commits",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM dolt.commits",
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
                    query: "SET search_path = 'dolt'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM commits",
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
                    query: "SELECT * FROM public.commits",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path = 'public'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM commits",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path = public, dolt",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM commits",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM COMMITS",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
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
            name: "dolt conflicts",
            set_up_script: &[
                "START TRANSACTION",
                "CREATE TABLE test (id INT PRIMARY KEY, col1 TEXT)",
                "SELECT dolt_commit('-Am', 'first commit')",
                "SELECT dolt_branch('b1')",
                "SELECT dolt_checkout('-b', 'b2')",
                "INSERT INTO test VALUES (1, 'a')",
                "SELECT dolt_commit('-Am', 'commit b2')",
                "SELECT dolt_checkout('b1')",
                "INSERT INTO test VALUES (1, 'b')",
                "SELECT dolt_commit('-Am', 'commit b1')",
                "SELECT dolt_checkout('main')",
                "SELECT dolt_merge('b1')",
                "SELECT dolt_merge('b2')",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt.conflicts",
                    expected: Expected::Rows {
                        columns: &[Column("table", TEXT), Column("num_conflicts", NUMERIC)],
                        rows: &[
                            &[T("test"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_conflicts",
                    expected: Expected::Rows {
                        columns: &[Column("table", TEXT), Column("num_conflicts", NUMERIC)],
                        rows: &[
                            &[T("test"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt.conflicts.table FROM dolt.conflicts",
                    expected: Expected::Rows {
                        columns: &[Column("table", TEXT)],
                        rows: &[
                            &[T("test")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_conflicts.table FROM dolt_conflicts",
                    expected: Expected::Rows {
                        columns: &[Column("table", TEXT)],
                        rows: &[
                            &[T("test")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SAVEPOINT probe",
                    expected: Expected::Tag("SAVEPOINT"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM public.conflicts",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: conflicts", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "ROLLBACK TO probe",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM conflicts",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: conflicts", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "ROLLBACK TO probe",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE TABLE conflicts (id INT PRIMARY KEY)",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO conflicts VALUES (1)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM conflicts",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt.conflicts",
                    expected: Expected::Rows {
                        columns: &[Column("table", TEXT), Column("num_conflicts", NUMERIC)],
                        rows: &[
                            &[T("test"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SET search_path = 'dolt'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM conflicts",
                    expected: Expected::Rows {
                        columns: &[Column("table", TEXT), Column("num_conflicts", NUMERIC)],
                        rows: &[
                            &[T("test"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM public.conflicts",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SET search_path = 'public'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM conflicts",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SET search_path = public, dolt",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM conflicts",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM CONFLICTS",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
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
            name: "dolt conflicts with tablename",
            set_up_script: &[
                "START TRANSACTION",
                "CREATE TABLE test (id INT PRIMARY KEY, col1 TEXT)",
                "SELECT dolt_commit('-Am', 'first commit')",
                "SELECT dolt_branch('b1')",
                "SELECT dolt_checkout('-b', 'b2')",
                "INSERT INTO test VALUES (1, 'a')",
                "SELECT dolt_commit('-Am', 'commit b2')",
                "SELECT dolt_checkout('b1')",
                "INSERT INTO test VALUES (1, 'b')",
                "SELECT dolt_commit('-Am', 'commit b1')",
                "SELECT dolt_checkout('main')",
                "SELECT dolt_merge('b1')",
                "SELECT dolt_merge('b2')",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT base_id, base_col1, our_id, our_col1, their_id, their_col1 FROM dolt_conflicts_test",
                    expected: Expected::Rows {
                        columns: &[Column("base_id", INT4), Column("base_col1", TEXT), Column("our_id", INT4), Column("our_col1", TEXT), Column("their_id", INT4), Column("their_col1", TEXT)],
                        rows: &[
                            &[Null, Null, T("1"), T("b"), T("1"), T("a")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT our_col1, their_col1 FROM public.dolt_conflicts_test",
                    expected: Expected::Rows {
                        columns: &[Column("our_col1", TEXT), Column("their_col1", TEXT)],
                        rows: &[
                            &[T("b"), T("a")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_conflicts_test.their_col1 FROM public.dolt_conflicts_test",
                    expected: Expected::Rows {
                        columns: &[Column("their_col1", TEXT)],
                        rows: &[
                            &[T("a")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SAVEPOINT probe",
                    expected: Expected::Tag("SAVEPOINT"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM other.dolt_conflicts_test",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: dolt_conflicts_test", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "ROLLBACK TO probe",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM public.dolt_conflicts_none",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: dolt_conflicts_none", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "ROLLBACK TO probe",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "DELETE FROM public.dolt_conflicts_test",
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT base_id, base_col1, our_id, our_col1, their_id, their_col1 FROM dolt_conflicts_test",
                    expected: Expected::Rows {
                        columns: &[Column("base_id", INT4), Column("base_col1", TEXT), Column("our_id", INT4), Column("our_col1", TEXT), Column("their_id", INT4), Column("their_col1", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE SCHEMA newschema",
                    expected: Expected::Tag("CREATE SCHEMA"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SET search_path = 'newschema'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE TABLE test_sch (id INT PRIMARY KEY)",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO test_sch VALUES (11)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_commit('-Am', 'add test_sch')",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM newschema.dolt_conflicts_test_sch",
                    expected: Expected::Rows {
                        columns: &[Column("from_root_ish", TEXT), Column("base_id", INT4), Column("our_id", INT4), Column("our_diff_type", TEXT), Column("their_id", INT4), Column("their_diff_type", TEXT), Column("dolt_conflict_id", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_conflicts_test_sch",
                    expected: Expected::Rows {
                        columns: &[Column("from_root_ish", TEXT), Column("base_id", INT4), Column("our_id", INT4), Column("our_diff_type", TEXT), Column("their_id", INT4), Column("their_diff_type", TEXT), Column("dolt_conflict_id", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SAVEPOINT probe",
                    expected: Expected::Tag("SAVEPOINT"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_conflicts_test",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: dolt_conflicts_test", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "ROLLBACK TO probe",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM public.dolt_conflicts_test",
                    expected: Expected::Rows {
                        columns: &[Column("from_root_ish", TEXT), Column("base_id", INT4), Column("base_col1", TEXT), Column("our_id", INT4), Column("our_col1", TEXT), Column("our_diff_type", TEXT), Column("their_id", INT4), Column("their_col1", TEXT), Column("their_diff_type", TEXT), Column("dolt_conflict_id", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT id FROM public.dolt_conflicts_test_sch",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: dolt_conflicts_test_sch", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "ROLLBACK TO probe",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM newschema.dolt_conflicts_test",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: dolt_conflicts_test", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "ROLLBACK TO probe",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE TABLE test (id INT PRIMARY KEY)",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES (12)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_commit('-Am', 'add test')",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM newschema.dolt_conflicts_test",
                    expected: Expected::Rows {
                        columns: &[Column("from_root_ish", TEXT), Column("base_id", INT4), Column("our_id", INT4), Column("our_diff_type", TEXT), Column("their_id", INT4), Column("their_diff_type", TEXT), Column("dolt_conflict_id", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_conflicts_test",
                    expected: Expected::Rows {
                        columns: &[Column("from_root_ish", TEXT), Column("base_id", INT4), Column("our_id", INT4), Column("our_diff_type", TEXT), Column("their_id", INT4), Column("their_diff_type", TEXT), Column("dolt_conflict_id", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM public.dolt_conflicts_test",
                    expected: Expected::Rows {
                        columns: &[Column("from_root_ish", TEXT), Column("base_id", INT4), Column("base_col1", TEXT), Column("our_id", INT4), Column("our_col1", TEXT), Column("our_diff_type", TEXT), Column("their_id", INT4), Column("their_col1", TEXT), Column("their_diff_type", TEXT), Column("dolt_conflict_id", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "dolt constraint violations",
            set_up_script: &[
                "CREATE TABLE otherTable (pk int primary key);",
                "CREATE TABLE test (pk int primary key, col1 int unique);",
                "SELECT dolt_commit('-Am', 'initial commit');",
                "SELECT dolt_branch('branch1');",
                "INSERT INTO test (pk, col1) VALUES (1, 1);",
                "SELECT dolt_commit('-am', 'insert on main');",
                "SELECT dolt_checkout('branch1');",
                "INSERT INTO test (pk, col1) VALUES (2, 1);",
                "SELECT dolt_commit('-am', 'insert on branch1');",
                "START TRANSACTION",
                "SELECT dolt_merge('main', '--squash')",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt.constraint_violations",
                    expected: Expected::Rows {
                        columns: &[Column("table", TEXT), Column("num_violations", NUMERIC)],
                        rows: &[
                            &[T("test"), T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_constraint_violations",
                    expected: Expected::Rows {
                        columns: &[Column("table", TEXT), Column("num_violations", NUMERIC)],
                        rows: &[
                            &[T("test"), T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt.constraint_violations.table FROM dolt.constraint_violations",
                    expected: Expected::Rows {
                        columns: &[Column("table", TEXT)],
                        rows: &[
                            &[T("test")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_constraint_violations.table FROM dolt_constraint_violations",
                    expected: Expected::Rows {
                        columns: &[Column("table", TEXT)],
                        rows: &[
                            &[T("test")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SAVEPOINT probe",
                    expected: Expected::Tag("SAVEPOINT"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM public.constraint_violations",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: constraint_violations", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "ROLLBACK TO probe",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM constraint_violations",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: constraint_violations", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "ROLLBACK TO probe",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE TABLE constraint_violations (id INT PRIMARY KEY)",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO constraint_violations VALUES (1)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM constraint_violations",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt.constraint_violations",
                    expected: Expected::Rows {
                        columns: &[Column("table", TEXT), Column("num_violations", NUMERIC)],
                        rows: &[
                            &[T("test"), T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SET search_path = 'dolt'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM constraint_violations",
                    expected: Expected::Rows {
                        columns: &[Column("table", TEXT), Column("num_violations", NUMERIC)],
                        rows: &[
                            &[T("test"), T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM public.constraint_violations",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SET search_path = 'public'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM constraint_violations",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SET search_path = public, dolt",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM constraint_violations",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM CONSTRAINT_VIOLATIONS",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
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
            name: "dolt constraint violations with tablename",
            set_up_script: &[
                "CREATE TABLE otherTable (pk int primary key);",
                "CREATE TABLE test (pk int primary key, col1 int unique);",
                "SELECT dolt_commit('-Am', 'initial commit');",
                "SELECT dolt_branch('branch1');",
                "INSERT INTO test (pk, col1) VALUES (1, 1);",
                "SELECT dolt_commit('-am', 'insert on main');",
                "SELECT dolt_checkout('branch1');",
                "INSERT INTO test (pk, col1) VALUES (2, 1);",
                "SELECT dolt_commit('-am', 'insert on branch1');",
                "START TRANSACTION",
                "SELECT dolt_merge('main', '--squash')",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT violation_type, pk, col1, violation_info FROM dolt_constraint_violations_test",
                    expected: Expected::Rows {
                        columns: &[Column("violation_type", VARCHAR), Column("pk", INT4), Column("col1", INT4), Column("violation_info", JSON)],
                        rows: &[
                            &[T("unique index"), T("1"), T("1"), T(r#"{"Columns":["col1"],"Name":"test_col1_key"}"#)],
                            &[T("unique index"), T("2"), T("1"), T(r#"{"Columns":["col1"],"Name":"test_col1_key"}"#)],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT violation_type, pk, col1, violation_info FROM public.dolt_constraint_violations_test",
                    expected: Expected::Rows {
                        columns: &[Column("violation_type", VARCHAR), Column("pk", INT4), Column("col1", INT4), Column("violation_info", JSON)],
                        rows: &[
                            &[T("unique index"), T("1"), T("1"), T(r#"{"Columns":["col1"],"Name":"test_col1_key"}"#)],
                            &[T("unique index"), T("2"), T("1"), T(r#"{"Columns":["col1"],"Name":"test_col1_key"}"#)],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM public.dolt_constraint_violations_test WHERE violation_type = 'foreign key'",
                    expected: Expected::Rows {
                        columns: &[Column("from_root_ish", TEXT), Column("violation_type", VARCHAR), Column("pk", INT4), Column("col1", INT4), Column("violation_info", JSON)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_constraint_violations_test.violation_type FROM public.dolt_constraint_violations_test",
                    expected: Expected::Rows {
                        columns: &[Column("violation_type", VARCHAR)],
                        rows: &[
                            &[T("unique index")],
                            &[T("unique index")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SAVEPOINT probe",
                    expected: Expected::Tag("SAVEPOINT"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM other.dolt_constraint_violations_test",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: dolt_constraint_violations_test", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "ROLLBACK TO probe",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM public.dolt_constraint_violations_none",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: dolt_constraint_violations_none", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "ROLLBACK TO probe",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "DELETE FROM public.dolt_constraint_violations_test",
                    expected: Expected::Tag("DELETE 2"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_constraint_violations_test",
                    expected: Expected::Rows {
                        columns: &[Column("from_root_ish", TEXT), Column("violation_type", VARCHAR), Column("pk", INT4), Column("col1", INT4), Column("violation_info", JSON)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE SCHEMA newschema",
                    expected: Expected::Tag("CREATE SCHEMA"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SET search_path = 'newschema'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE TABLE test_sch (id INT PRIMARY KEY)",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO test_sch VALUES (11)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_commit('-Am', 'add test_sch')",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM newschema.dolt_constraint_violations_test_sch",
                    expected: Expected::Rows {
                        columns: &[Column("from_root_ish", TEXT), Column("violation_type", VARCHAR), Column("id", INT4), Column("violation_info", JSON)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_constraint_violations_test_sch",
                    expected: Expected::Rows {
                        columns: &[Column("from_root_ish", TEXT), Column("violation_type", VARCHAR), Column("id", INT4), Column("violation_info", JSON)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SAVEPOINT probe",
                    expected: Expected::Tag("SAVEPOINT"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_constraint_violations_test",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: dolt_constraint_violations_test", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "ROLLBACK TO probe",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM public.dolt_constraint_violations_test",
                    expected: Expected::Rows {
                        columns: &[Column("from_root_ish", TEXT), Column("violation_type", VARCHAR), Column("pk", INT4), Column("col1", INT4), Column("violation_info", JSON)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT id FROM public.dolt_constraint_violations_test_sch",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: dolt_constraint_violations_test_sch", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "ROLLBACK TO probe",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM newschema.dolt_constraint_violations_test",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: dolt_constraint_violations_test", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "ROLLBACK TO probe",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE TABLE test (id INT PRIMARY KEY)",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES (12)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_commit('-Am', 'add test')",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM newschema.dolt_constraint_violations_test",
                    expected: Expected::Rows {
                        columns: &[Column("from_root_ish", TEXT), Column("violation_type", VARCHAR), Column("id", INT4), Column("violation_info", JSON)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_constraint_violations_test",
                    expected: Expected::Rows {
                        columns: &[Column("from_root_ish", TEXT), Column("violation_type", VARCHAR), Column("id", INT4), Column("violation_info", JSON)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM public.dolt_constraint_violations_test",
                    expected: Expected::Rows {
                        columns: &[Column("from_root_ish", TEXT), Column("violation_type", VARCHAR), Column("pk", INT4), Column("col1", INT4), Column("violation_info", JSON)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "dolt docs",
            set_up_script: &[
                "INSERT INTO dolt.docs values ('README.md', 'testing')",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt.docs",
                    expected: Expected::Rows {
                        columns: &[Column("doc_name", TEXT), Column("doc_text", TEXT)],
                        rows: &[
                            &[T("README.md"), T("testing")],
                            &[T("AGENT.md"), T(r#"# AGENT.md - Dolt Database Operations Guide

This file provides guidance for AI agents working with Dolt databases to maximize productivity and follow best practices.

## Quick Start

Dolt is "Git for Data" - a SQL database with version control capabilities. All Git commands have Dolt equivalents:
- `git add` → `dolt add`  
- `git commit` → `dolt commit`
- `git branch` → `dolt branch`
- `git merge` → `dolt merge`
- `git diff` → `dolt diff`

For help and documentation on commands, you can run `dolt --help` and `dolt <command> --help`.

## Essential Dolt CLI Commands

### Repository Operations
```bash
# Initialize new database
dolt init

# Clone existing database
dolt clone <remote-url>

# Show current status
dolt status

# View commit history
dolt log
```

### Branch Management
```bash
# List branches
dolt branch

# Create new branch
dolt branch <branch-name>

# Switch branches
dolt checkout <branch-name>

# Create and switch to new branch
dolt checkout -b <branch-name>
```

### Checkout Behavior with Running SQL Servers
- `dolt checkout` on the CLI only affects the shell process that runs the command. When a `dolt sql-server` is running, existing SQL connections keep their current branch until they explicitly switch.
- Each SQL session (CLI `dolt sql`, MySQL client, application connection) maintains its own active branch. Run `CALL dolt_checkout('<branch>');` at the beginning of every session or scripted block to ensure you are on the correct branch.
- Chain branch changes inside scripts: start with `CALL dolt_checkout('<branch>');`, then run your queries. Do not assume a previous checkout persists for new connections.
- When automating, include the checkout in the same transaction / session context where the data changes execute.
- A good way to make sure a `dolt sql` session connects to the br1 branch for instance is `dolt --branch br1 sql`.

### Data Operations
```bash
# Stage changes
dolt add <table-name>
dolt add .  # stage all changes

# Commit changes
dolt commit -m "commit message"

# View differences
dolt diff
dolt diff <table-name>
dolt diff <branch1> <branch2>

# Merge branches
dolt merge <branch-name>
```

## Starting and Connecting to Dolt SQL Server

### Start SQL Server
```bash
# Start server on default port (3306)
dolt sql-server

# Start on specific port
dolt sql-server --port=3307

# Start with specific host
dolt sql-server --host=0.0.0.0 --port=3307

# Start in background
dolt sql-server --port=3307 &
```

### Connecting to SQL Server
```bash
# Connect with dolt sql command
dolt sql

# Connect with mysql client
mysql -h 127.0.0.1 -P 3306 -u root

# Connect with specific database
mysql -h 127.0.0.1 -P 3306 -u root -D <database-name>
```

## Dolt Testing with dolt_test System Table

### Unit Testing with dolt_test

The dolt_test system table provides a powerful way to create and run unit tests for your database. This is the preferred method for testing data integrity, business rules, and schema validation.

#### Creating Tests

Tests are created by inserting rows into the `dolt_tests` system table:

```sql
-- Create a simple test
INSERT INTO `dolt_tests` VALUES (
    'test_user_count', 
    'validation', 
    'SELECT COUNT(*) as user_count FROM users;', 
    'row_count',
    '>',
    '0'
);

-- Create a test with expected result
INSERT INTO `dolt_tests` VALUES (
    'test_valid_emails', 
    'validation', 
    'SELECT COUNT(*) FROM users WHERE email NOT LIKE "%@%";', 
    'row_count',
    '==',
    '0'
);

-- Create a schema validation test
INSERT INTO `dolt_tests` VALUES (
    'test_users_schema', 
    'schema', 
    'DESCRIBE users;', 
    'row_count',
    '>=',
    '5'
);
```

#### Test Structure

Each test row contains:
- test_name: Unique identifier for the test
- test_group: Optional grouping for tests (e.g., 'validation', 'schema', 'integration')
- test_query: SQL query to execute
- assertion_type: Type of assertion ('expected_rows', 'expected_columns', 'expected_single_value')
- assertion_comparator: Comparison operator ('==', '>', '<', '>=', '<=', '!=')
- assertion_value: Expected value for comparison

#### Running Tests

```sql
-- Run all tests
SELECT * FROM dolt_test_run();

-- Run specific test
SELECT * FROM dolt_test_run('test_user_count');

-- Run tests with filtering
SELECT * FROM dolt_test_run() WHERE test_name LIKE 'test_user%' AND status != 'PASS';
```

#### Test Result Interpretation

The dolt_test_run() function returns:
- test_name: Name of the test
- status: PASS, FAIL, or ERROR
- actual_result: Actual query result
- expected_result: Expected result
- message: Additional details

#### Advanced Testing Examples

```sql
-- Test data integrity
INSERT INTO `dolt_tests` VALUES (
    'test_no_orphaned_orders', 
    'integrity', 
    'SELECT COUNT(*) FROM orders o LEFT JOIN users u ON o.user_id = u.id WHERE u.id IS NULL;', 
    'row_count',
    '==',
    '0'
);

-- Test business rules
INSERT INTO `dolt_tests` VALUES (
    'test_positive_prices', 
    'business_rules', 
    'SELECT COUNT(*) FROM products WHERE price <= 0;', 
    'row_count',
    '==',
    '0'
);

-- Test complex relationships
INSERT INTO `dolt_tests` VALUES (
    'test_order_totals', 
    'integrity', 
    'SELECT COUNT(*) FROM orders o JOIN order_items oi ON o.id = oi.order_id GROUP BY o.id HAVING SUM(oi.quantity * oi.price) != o.total;', 
    'row_count',
    '==',
    '0'
);
```

### Dolt CI for DoltHub Integration

Dolt CI is specifically designed for running tests on DoltHub when pull requests are created. Use this only for tests you want to run automatically on DoltHub.

#### Prerequisites for DoltHub CI
- Requires Dolt v1.43.14 or later
- Must initialize CI capabilities: `dolt ci init`
- Workflows defined in YAML files

#### Available CI Commands
```bash
# Initialize CI capabilities
dolt ci init

# List available workflows
dolt ci ls

# View workflow details
dolt ci view <workflow-name>

# View specific job in workflow
dolt ci view <workflow-name> <job-name>

# Run workflow locally (for testing before DoltHub)
dolt ci run <workflow-name>
```

#### Creating CI Workflows for DoltHub

Create workflow files that will run on DoltHub when pull requests are opened:

```yaml
name: doltHub validation workflow
on:
  push:
    branches:
      - master
      - main
jobs:
  - name: validate schema
    steps:
      - name: check required tables exist
        saved_query_name: show_tables
        expected_rows: ">= 3"
      
      - name: validate user data
        saved_query_name: user_count_check
        expected_columns: "== 1"
        expected_rows: "> 0"
  
  - name: data integrity checks
    steps:
      - name: check email format
        saved_query_name: valid_emails
        expected_rows: "== 0"  # No invalid emails
```

### Best Practices for Testing

1. **Use dolt_test for Unit Testing**
   - Create tests for data validation
   - Test business rules and constraints
   - Validate schema changes
   - Run tests frequently during development

2. **Use Dolt CI for DoltHub Integration**
   - Only for tests that should run on pull requests
   - Focus on integration and deployment validation
   - Test against production-like data

3. **Create Comprehensive Test Suites**
   - Test data integrity constraints
   - Validate business rules
   - Check schema requirements
   - Verify data relationships

4. **Version Control Your Tests**
   - Commit test definitions to repository
   - Track changes to test configuration
   - Use branches for test development

## System Tables for Version Control

Dolt exposes version control operations through system tables accessible via SQL:

### Core System Tables
```sql
-- View commit history
SELECT * FROM dolt_log;

-- Check current status
SELECT * FROM dolt_status;

-- View branch information
SELECT * FROM dolt_branches;

-- See table diffs
SELECT * FROM dolt_diff_<table_name>;

-- View schema changes
SELECT * FROM dolt_schema_diff;

-- Check conflicts during merge
SELECT * FROM dolt_conflicts_<table_name>;

-- View commit metadata
SELECT * FROM dolt_commits;
```

### Version Control Operations via SQL

When working in SQL sessions, you can execute version control operations using stored procedures:

```sql
-- Stage and commit changes
CALL dolt_add('.');
CALL dolt_commit('-m', 'commit message');

-- Branch operations
CALL dolt_branch('<branch_name>');
CALL dolt_checkout('<branch_name>');
CALL dolt_merge('<branch_name>');
```

**Note:** Use CLI commands (`dolt add`, `dolt commit`, etc.) for most operations. SQL procedures are useful when already in a SQL session.

### Advanced System Tables
```sql
-- View remotes
SELECT * FROM dolt_remotes;

-- Check merge conflicts
SELECT * FROM dolt_conflicts;

-- View statistics
SELECT * FROM dolt_statistics;

-- See ignored tables
SELECT * FROM dolt_ignore;
```

## CLI vs SQL Approach

**Prefer CLI commands for:**
- Version control operations (add, commit, branch, merge)
- Repository management (init, clone, push, pull)
- Conflict resolution
- Status checking and history viewing

**Use SQL for:**
- Data queries and analysis
- Complex data transformations
- Examining system tables (dolt_log, dolt_status, etc.)
- When already in an active SQL session

## Schema Design Recommendations

### Use UUID Keys Instead of Auto-Increment

For Dolt's version control features, use UUID primary keys instead of auto-increment:

```sql
-- Recommended
CREATE TABLE users (
    id varchar(36) default(uuid()) primary key,
    name varchar(255)
);

-- Avoid auto-increment with Dolt
-- id int auto_increment primary key
```

**Benefits:**
- Prevents merge conflicts across branches and database clones
- Automatic generation with default(uuid())
- Works seamlessly in distributed environments

## Best Practices for Agents

### 1. Always Work on Feature Branches
```bash
# Create feature branch before making changes
dolt checkout -b feature/agent-changes

# Make changes on feature branch
dolt sql -q "INSERT INTO users VALUES (1, 'Alice');"

# Stage and commit
dolt add .
dolt commit -m "Add new user Alice"

# Switch back to main to merge
dolt checkout main
dolt merge feature/agent-changes
```

### 2. Use SQL for Data Operations, CLI for Version Control
```bash
# Use dolt sql for data changes
dolt sql -q "INSERT INTO users VALUES (1, 'Alice');"
dolt sql -q "UPDATE products SET price = price * 1.1 WHERE category = 'electronics';"

# Check status and commit using CLI
dolt status
dolt add .
dolt commit -m "Update user and product data"
```

### 3. Validate Changes with System Tables
```sql
-- Before major operations, check current state
SELECT * FROM dolt_status;
SELECT * FROM dolt_branches;

-- After changes, verify with diffs
SELECT * FROM dolt_diff_users;
SELECT * FROM dolt_schema_diff;
```

### 4. Use dolt_test for Data Validation
Create tests to validate:
- Data integrity after changes
- Schema compatibility
- Business rule compliance
- Cross-table relationships

### 5. Handle Conflicts Gracefully
```bash
# Check for conflicts using CLI
dolt conflicts cat <table_name>
dolt conflicts resolve --ours <table_name>
dolt conflicts resolve --theirs <table_name>

# Or use SQL to examine conflicts
dolt sql -q "SELECT * FROM dolt_conflicts_<table_name>;"
```

## Common Workflow Examples

### Data Migration Workflow
```bash
# Create migration branch
dolt checkout -b migration/update-schema

# Apply schema changes via SQL
dolt sql -q "ALTER TABLE users ADD COLUMN email VARCHAR(255);"

# Create validation tests
dolt sql -q "INSERT INTO `dolt_tests` VALUES ('test_users_schema', 'schema', 'DESCRIBE users;', 'row_count', '>=', '6');"
dolt sql -q "INSERT INTO `dolt_tests` VALUES ('test_email_column', 'schema', 'SELECT COUNT(*) FROM users WHERE email IS NULL;', 'row_count', '>=', '0');"

# Run tests to validate changes
dolt sql -q "SELECT * FROM dolt_test_run();"

# Stage and commit
dolt add .
dolt commit -m "Add email column to users table"

# Merge back
dolt checkout main
dolt merge migration/update-schema
```

### Data Analysis Workflow
```bash
# Create analysis branch
dolt checkout -b analysis/user-behavior

# Create analysis tables via SQL
dolt sql -q "CREATE TABLE user_metrics AS 
            SELECT user_id, COUNT(*) as actions 
            FROM user_actions 
            GROUP BY user_id;"

# Create tests to validate analysis
dolt sql -q "INSERT INTO `dolt_tests` VALUES ('test_metrics_created', 'analysis', 'SELECT COUNT(*) FROM user_metrics;', 'row_count', '>', '0');"
dolt sql -q "INSERT INTO `dolt_tests` VALUES ('test_metrics_integrity', 'integrity', 'SELECT COUNT(*) FROM user_metrics um LEFT JOIN users u ON um.user_id = u.id WHERE u.id IS NULL;', 'row_count', '==', '0');"

# Run tests to validate analysis
dolt sql -q "SELECT * FROM dolt_test_run();"

# Stage and commit using CLI
dolt add user_metrics
dolt commit -m "Add user behavior analysis"
```

## Integration with External Tools

### Database Clients
Most MySQL clients work with Dolt:
- MySQL Workbench
- phpMyAdmin  
- DataGrip
- DBeaver

### Backup and Sync
```bash
# Push to remote
dolt push origin main

# Pull changes
dolt pull origin main

# Clone for backup
dolt clone <remote-url> backup-location
```

This guide enables agents to leverage Dolt's unique version control capabilities while maintaining data integrity and following collaborative development practices."#)],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt.docs.doc_name FROM dolt.docs",
                    expected: Expected::Rows {
                        columns: &[Column("doc_name", TEXT)],
                        rows: &[
                            &[T("README.md")],
                            &[T("AGENT.md")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM public.docs",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: docs", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM docs",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: docs", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_diff_summary('main', 'WORKING')",
                    expected: Expected::Rows {
                        columns: &[Column("from_table_name", TEXT), Column("to_table_name", TEXT), Column("diff_type", TEXT), Column("data_change", BOOL), Column("schema_change", BOOL)],
                        rows: &[
                            &[T(""), T("dolt.docs"), T("added"), T("t"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_diff_summary('main', 'WORKING', 'docs')",
                    expected: Expected::Rows {
                        columns: &[Column("from_table_name", TEXT), Column("to_table_name", TEXT), Column("diff_type", TEXT), Column("data_change", BOOL), Column("schema_change", BOOL)],
                        rows: &[
                            &[T(""), T("dolt.docs"), T("added"), T("t"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_diff_summary('main', 'WORKING', 'dolt_docs')",
                    expected: Expected::Rows {
                        columns: &[Column("from_table_name", TEXT), Column("to_table_name", TEXT), Column("diff_type", TEXT), Column("data_change", BOOL), Column("schema_change", BOOL)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_diff_summary('main', 'WORKING', 'dolt.docs')",
                    expected: Expected::Rows {
                        columns: &[Column("from_table_name", TEXT), Column("to_table_name", TEXT), Column("diff_type", TEXT), Column("data_change", BOOL), Column("schema_change", BOOL)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_diff_summary('main', 'WORKING', 'docs')",
                    expected: Expected::Rows {
                        columns: &[Column("from_table_name", TEXT), Column("to_table_name", TEXT), Column("diff_type", TEXT), Column("data_change", BOOL), Column("schema_change", BOOL)],
                        rows: &[
                            &[T(""), T("dolt.docs"), T("added"), T("t"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT diff_type, from_doc_name, to_doc_name FROM dolt_diff('main', 'WORKING', 'docs')",
                    expected: Expected::Rows {
                        columns: &[Column("diff_type", TEXT), Column("from_doc_name", TEXT), Column("to_doc_name", TEXT)],
                        rows: &[
                            &[T("added"), Null, T("README.md")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT diff_type, from_doc_name, to_doc_name FROM dolt_diff('main', 'WORKING', 'docs')",
                    expected: Expected::Rows {
                        columns: &[Column("diff_type", TEXT), Column("from_doc_name", TEXT), Column("to_doc_name", TEXT)],
                        rows: &[
                            &[T("added"), Null, T("README.md")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE TABLE docs (id INT PRIMARY KEY)",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO docs VALUES (1)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM docs",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT doc_name FROM dolt.docs",
                    expected: Expected::Rows {
                        columns: &[Column("doc_name", TEXT)],
                        rows: &[
                            &[T("README.md")],
                            &[T("AGENT.md")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SET search_path = 'dolt'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT doc_name FROM docs",
                    expected: Expected::Rows {
                        columns: &[Column("doc_name", TEXT)],
                        rows: &[
                            &[T("README.md")],
                            &[T("AGENT.md")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM public.docs",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SET search_path = 'public'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM docs",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SET search_path = public, dolt",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM docs",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM DOCS",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SET search_path = 'public'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "DELETE FROM dolt.docs WHERE doc_name = 'README.md'",
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt.docs",
                    expected: Expected::Rows {
                        columns: &[Column("doc_name", TEXT), Column("doc_text", TEXT)],
                        rows: &[
                            &[T("AGENT.md"), T(r#"# AGENT.md - Dolt Database Operations Guide

This file provides guidance for AI agents working with Dolt databases to maximize productivity and follow best practices.

## Quick Start

Dolt is "Git for Data" - a SQL database with version control capabilities. All Git commands have Dolt equivalents:
- `git add` → `dolt add`  
- `git commit` → `dolt commit`
- `git branch` → `dolt branch`
- `git merge` → `dolt merge`
- `git diff` → `dolt diff`

For help and documentation on commands, you can run `dolt --help` and `dolt <command> --help`.

## Essential Dolt CLI Commands

### Repository Operations
```bash
# Initialize new database
dolt init

# Clone existing database
dolt clone <remote-url>

# Show current status
dolt status

# View commit history
dolt log
```

### Branch Management
```bash
# List branches
dolt branch

# Create new branch
dolt branch <branch-name>

# Switch branches
dolt checkout <branch-name>

# Create and switch to new branch
dolt checkout -b <branch-name>
```

### Checkout Behavior with Running SQL Servers
- `dolt checkout` on the CLI only affects the shell process that runs the command. When a `dolt sql-server` is running, existing SQL connections keep their current branch until they explicitly switch.
- Each SQL session (CLI `dolt sql`, MySQL client, application connection) maintains its own active branch. Run `CALL dolt_checkout('<branch>');` at the beginning of every session or scripted block to ensure you are on the correct branch.
- Chain branch changes inside scripts: start with `CALL dolt_checkout('<branch>');`, then run your queries. Do not assume a previous checkout persists for new connections.
- When automating, include the checkout in the same transaction / session context where the data changes execute.
- A good way to make sure a `dolt sql` session connects to the br1 branch for instance is `dolt --branch br1 sql`.

### Data Operations
```bash
# Stage changes
dolt add <table-name>
dolt add .  # stage all changes

# Commit changes
dolt commit -m "commit message"

# View differences
dolt diff
dolt diff <table-name>
dolt diff <branch1> <branch2>

# Merge branches
dolt merge <branch-name>
```

## Starting and Connecting to Dolt SQL Server

### Start SQL Server
```bash
# Start server on default port (3306)
dolt sql-server

# Start on specific port
dolt sql-server --port=3307

# Start with specific host
dolt sql-server --host=0.0.0.0 --port=3307

# Start in background
dolt sql-server --port=3307 &
```

### Connecting to SQL Server
```bash
# Connect with dolt sql command
dolt sql

# Connect with mysql client
mysql -h 127.0.0.1 -P 3306 -u root

# Connect with specific database
mysql -h 127.0.0.1 -P 3306 -u root -D <database-name>
```

## Dolt Testing with dolt_test System Table

### Unit Testing with dolt_test

The dolt_test system table provides a powerful way to create and run unit tests for your database. This is the preferred method for testing data integrity, business rules, and schema validation.

#### Creating Tests

Tests are created by inserting rows into the `dolt_tests` system table:

```sql
-- Create a simple test
INSERT INTO `dolt_tests` VALUES (
    'test_user_count', 
    'validation', 
    'SELECT COUNT(*) as user_count FROM users;', 
    'row_count',
    '>',
    '0'
);

-- Create a test with expected result
INSERT INTO `dolt_tests` VALUES (
    'test_valid_emails', 
    'validation', 
    'SELECT COUNT(*) FROM users WHERE email NOT LIKE "%@%";', 
    'row_count',
    '==',
    '0'
);

-- Create a schema validation test
INSERT INTO `dolt_tests` VALUES (
    'test_users_schema', 
    'schema', 
    'DESCRIBE users;', 
    'row_count',
    '>=',
    '5'
);
```

#### Test Structure

Each test row contains:
- test_name: Unique identifier for the test
- test_group: Optional grouping for tests (e.g., 'validation', 'schema', 'integration')
- test_query: SQL query to execute
- assertion_type: Type of assertion ('expected_rows', 'expected_columns', 'expected_single_value')
- assertion_comparator: Comparison operator ('==', '>', '<', '>=', '<=', '!=')
- assertion_value: Expected value for comparison

#### Running Tests

```sql
-- Run all tests
SELECT * FROM dolt_test_run();

-- Run specific test
SELECT * FROM dolt_test_run('test_user_count');

-- Run tests with filtering
SELECT * FROM dolt_test_run() WHERE test_name LIKE 'test_user%' AND status != 'PASS';
```

#### Test Result Interpretation

The dolt_test_run() function returns:
- test_name: Name of the test
- status: PASS, FAIL, or ERROR
- actual_result: Actual query result
- expected_result: Expected result
- message: Additional details

#### Advanced Testing Examples

```sql
-- Test data integrity
INSERT INTO `dolt_tests` VALUES (
    'test_no_orphaned_orders', 
    'integrity', 
    'SELECT COUNT(*) FROM orders o LEFT JOIN users u ON o.user_id = u.id WHERE u.id IS NULL;', 
    'row_count',
    '==',
    '0'
);

-- Test business rules
INSERT INTO `dolt_tests` VALUES (
    'test_positive_prices', 
    'business_rules', 
    'SELECT COUNT(*) FROM products WHERE price <= 0;', 
    'row_count',
    '==',
    '0'
);

-- Test complex relationships
INSERT INTO `dolt_tests` VALUES (
    'test_order_totals', 
    'integrity', 
    'SELECT COUNT(*) FROM orders o JOIN order_items oi ON o.id = oi.order_id GROUP BY o.id HAVING SUM(oi.quantity * oi.price) != o.total;', 
    'row_count',
    '==',
    '0'
);
```

### Dolt CI for DoltHub Integration

Dolt CI is specifically designed for running tests on DoltHub when pull requests are created. Use this only for tests you want to run automatically on DoltHub.

#### Prerequisites for DoltHub CI
- Requires Dolt v1.43.14 or later
- Must initialize CI capabilities: `dolt ci init`
- Workflows defined in YAML files

#### Available CI Commands
```bash
# Initialize CI capabilities
dolt ci init

# List available workflows
dolt ci ls

# View workflow details
dolt ci view <workflow-name>

# View specific job in workflow
dolt ci view <workflow-name> <job-name>

# Run workflow locally (for testing before DoltHub)
dolt ci run <workflow-name>
```

#### Creating CI Workflows for DoltHub

Create workflow files that will run on DoltHub when pull requests are opened:

```yaml
name: doltHub validation workflow
on:
  push:
    branches:
      - master
      - main
jobs:
  - name: validate schema
    steps:
      - name: check required tables exist
        saved_query_name: show_tables
        expected_rows: ">= 3"
      
      - name: validate user data
        saved_query_name: user_count_check
        expected_columns: "== 1"
        expected_rows: "> 0"
  
  - name: data integrity checks
    steps:
      - name: check email format
        saved_query_name: valid_emails
        expected_rows: "== 0"  # No invalid emails
```

### Best Practices for Testing

1. **Use dolt_test for Unit Testing**
   - Create tests for data validation
   - Test business rules and constraints
   - Validate schema changes
   - Run tests frequently during development

2. **Use Dolt CI for DoltHub Integration**
   - Only for tests that should run on pull requests
   - Focus on integration and deployment validation
   - Test against production-like data

3. **Create Comprehensive Test Suites**
   - Test data integrity constraints
   - Validate business rules
   - Check schema requirements
   - Verify data relationships

4. **Version Control Your Tests**
   - Commit test definitions to repository
   - Track changes to test configuration
   - Use branches for test development

## System Tables for Version Control

Dolt exposes version control operations through system tables accessible via SQL:

### Core System Tables
```sql
-- View commit history
SELECT * FROM dolt_log;

-- Check current status
SELECT * FROM dolt_status;

-- View branch information
SELECT * FROM dolt_branches;

-- See table diffs
SELECT * FROM dolt_diff_<table_name>;

-- View schema changes
SELECT * FROM dolt_schema_diff;

-- Check conflicts during merge
SELECT * FROM dolt_conflicts_<table_name>;

-- View commit metadata
SELECT * FROM dolt_commits;
```

### Version Control Operations via SQL

When working in SQL sessions, you can execute version control operations using stored procedures:

```sql
-- Stage and commit changes
CALL dolt_add('.');
CALL dolt_commit('-m', 'commit message');

-- Branch operations
CALL dolt_branch('<branch_name>');
CALL dolt_checkout('<branch_name>');
CALL dolt_merge('<branch_name>');
```

**Note:** Use CLI commands (`dolt add`, `dolt commit`, etc.) for most operations. SQL procedures are useful when already in a SQL session.

### Advanced System Tables
```sql
-- View remotes
SELECT * FROM dolt_remotes;

-- Check merge conflicts
SELECT * FROM dolt_conflicts;

-- View statistics
SELECT * FROM dolt_statistics;

-- See ignored tables
SELECT * FROM dolt_ignore;
```

## CLI vs SQL Approach

**Prefer CLI commands for:**
- Version control operations (add, commit, branch, merge)
- Repository management (init, clone, push, pull)
- Conflict resolution
- Status checking and history viewing

**Use SQL for:**
- Data queries and analysis
- Complex data transformations
- Examining system tables (dolt_log, dolt_status, etc.)
- When already in an active SQL session

## Schema Design Recommendations

### Use UUID Keys Instead of Auto-Increment

For Dolt's version control features, use UUID primary keys instead of auto-increment:

```sql
-- Recommended
CREATE TABLE users (
    id varchar(36) default(uuid()) primary key,
    name varchar(255)
);

-- Avoid auto-increment with Dolt
-- id int auto_increment primary key
```

**Benefits:**
- Prevents merge conflicts across branches and database clones
- Automatic generation with default(uuid())
- Works seamlessly in distributed environments

## Best Practices for Agents

### 1. Always Work on Feature Branches
```bash
# Create feature branch before making changes
dolt checkout -b feature/agent-changes

# Make changes on feature branch
dolt sql -q "INSERT INTO users VALUES (1, 'Alice');"

# Stage and commit
dolt add .
dolt commit -m "Add new user Alice"

# Switch back to main to merge
dolt checkout main
dolt merge feature/agent-changes
```

### 2. Use SQL for Data Operations, CLI for Version Control
```bash
# Use dolt sql for data changes
dolt sql -q "INSERT INTO users VALUES (1, 'Alice');"
dolt sql -q "UPDATE products SET price = price * 1.1 WHERE category = 'electronics';"

# Check status and commit using CLI
dolt status
dolt add .
dolt commit -m "Update user and product data"
```

### 3. Validate Changes with System Tables
```sql
-- Before major operations, check current state
SELECT * FROM dolt_status;
SELECT * FROM dolt_branches;

-- After changes, verify with diffs
SELECT * FROM dolt_diff_users;
SELECT * FROM dolt_schema_diff;
```

### 4. Use dolt_test for Data Validation
Create tests to validate:
- Data integrity after changes
- Schema compatibility
- Business rule compliance
- Cross-table relationships

### 5. Handle Conflicts Gracefully
```bash
# Check for conflicts using CLI
dolt conflicts cat <table_name>
dolt conflicts resolve --ours <table_name>
dolt conflicts resolve --theirs <table_name>

# Or use SQL to examine conflicts
dolt sql -q "SELECT * FROM dolt_conflicts_<table_name>;"
```

## Common Workflow Examples

### Data Migration Workflow
```bash
# Create migration branch
dolt checkout -b migration/update-schema

# Apply schema changes via SQL
dolt sql -q "ALTER TABLE users ADD COLUMN email VARCHAR(255);"

# Create validation tests
dolt sql -q "INSERT INTO `dolt_tests` VALUES ('test_users_schema', 'schema', 'DESCRIBE users;', 'row_count', '>=', '6');"
dolt sql -q "INSERT INTO `dolt_tests` VALUES ('test_email_column', 'schema', 'SELECT COUNT(*) FROM users WHERE email IS NULL;', 'row_count', '>=', '0');"

# Run tests to validate changes
dolt sql -q "SELECT * FROM dolt_test_run();"

# Stage and commit
dolt add .
dolt commit -m "Add email column to users table"

# Merge back
dolt checkout main
dolt merge migration/update-schema
```

### Data Analysis Workflow
```bash
# Create analysis branch
dolt checkout -b analysis/user-behavior

# Create analysis tables via SQL
dolt sql -q "CREATE TABLE user_metrics AS 
            SELECT user_id, COUNT(*) as actions 
            FROM user_actions 
            GROUP BY user_id;"

# Create tests to validate analysis
dolt sql -q "INSERT INTO `dolt_tests` VALUES ('test_metrics_created', 'analysis', 'SELECT COUNT(*) FROM user_metrics;', 'row_count', '>', '0');"
dolt sql -q "INSERT INTO `dolt_tests` VALUES ('test_metrics_integrity', 'integrity', 'SELECT COUNT(*) FROM user_metrics um LEFT JOIN users u ON um.user_id = u.id WHERE u.id IS NULL;', 'row_count', '==', '0');"

# Run tests to validate analysis
dolt sql -q "SELECT * FROM dolt_test_run();"

# Stage and commit using CLI
dolt add user_metrics
dolt commit -m "Add user behavior analysis"
```

## Integration with External Tools

### Database Clients
Most MySQL clients work with Dolt:
- MySQL Workbench
- phpMyAdmin  
- DataGrip
- DBeaver

### Backup and Sync
```bash
# Push to remote
dolt push origin main

# Pull changes
dolt pull origin main

# Clone for backup
dolt clone <remote-url> backup-location
```

This guide enables agents to leverage Dolt's unique version control capabilities while maintaining data integrity and following collaborative development practices."#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "DELETE FROM dolt_docs WHERE doc_name = 'README.md'",
                    expected: Expected::Tag("DELETE 0"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "dolt diff",
            set_up_script: &[
                "CREATE TABLE test (id INT PRIMARY KEY)",
                "SELECT dolt_commit('-Am', 'test commit')",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name FROM dolt.diff",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT)],
                        rows: &[
                            &[T("public.test")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name, committer, email, message, data_change, schema_change FROM dolt.diff",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("committer", TEXT), Column("email", TEXT), Column("message", TEXT), Column("data_change", BOOL), Column("schema_change", BOOL)],
                        rows: &[
                            &[T("public.test"), T("postgres"), T("postgres@127.0.0.1"), T("test commit"), T("f"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name, data_change, schema_change FROM dolt.diff WHERE data_change=false",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("data_change", BOOL), Column("schema_change", BOOL)],
                        rows: &[
                            &[T("public.test"), T("f"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name, data_change, schema_change FROM dolt.diff WHERE schema_change=false",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("data_change", BOOL), Column("schema_change", BOOL)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name FROM dolt_diff",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT)],
                        rows: &[
                            &[T("public.test")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt.diff.table_name FROM dolt.diff",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT)],
                        rows: &[
                            &[T("public.test")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_diff.table_name FROM dolt_diff",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT)],
                        rows: &[
                            &[T("public.test")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM public.diff",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: diff", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM diff",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: diff", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE TABLE diff (id INT PRIMARY KEY)",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO diff VALUES (1)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM diff",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name FROM dolt.diff WHERE table_name = 'public.test'",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT)],
                        rows: &[
                            &[T("public.test")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SET search_path = 'dolt'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name FROM diff WHERE table_name = 'public.test'",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT)],
                        rows: &[
                            &[T("public.test")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM public.diff",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SET search_path = 'public'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM diff",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SET search_path = public, dolt",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM diff",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM DIFF",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
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
            name: "dolt diff with tablename",
            set_up_script: &[
                "CREATE TABLE test (id INT PRIMARY KEY)",
                "INSERT INTO test VALUES (10)",
                "SELECT dolt_commit('-Am', 'test commit 1')",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_id, to_id, diff_type FROM dolt_diff_test WHERE to_commit=HASHOF('HEAD')",
                    expected: Expected::Rows {
                        columns: &[Column("from_id", INT4), Column("to_id", INT4), Column("diff_type", TEXT)],
                        rows: &[
                            &[Null, T("10"), T("added")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_id, to_id, diff_type FROM doLt_DIff_tEst WHERE to_commit=HASHOF('HEAD')",
                    expected: Expected::Rows {
                        columns: &[Column("from_id", INT4), Column("to_id", INT4), Column("diff_type", TEXT)],
                        rows: &[
                            &[Null, T("10"), T("added")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_id, to_id, diff_type FROM public.dolt_diff_test WHERE to_commit=HASHOF('HEAD')",
                    expected: Expected::Rows {
                        columns: &[Column("from_id", INT4), Column("to_id", INT4), Column("diff_type", TEXT)],
                        rows: &[
                            &[Null, T("10"), T("added")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_id, to_id, diff_type FROM public.doLt_DIff_tEst WHERE to_commit=HASHOF('HEAD')",
                    expected: Expected::Rows {
                        columns: &[Column("from_id", INT4), Column("to_id", INT4), Column("diff_type", TEXT)],
                        rows: &[
                            &[Null, T("10"), T("added")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_diff_test.to_id FROM public.dolt_diff_test WHERE to_commit=HASHOF('HEAD')",
                    expected: Expected::Rows {
                        columns: &[Column("to_id", INT4)],
                        rows: &[
                            &[T("10")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM other.dolt_diff_test",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: dolt_diff_test", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM public.dolt_diff_none",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: dolt_diff_none", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_diff_none",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: dolt_diff_none", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE SCHEMA newschema",
                    expected: Expected::Tag("CREATE SCHEMA"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SET search_path = 'newschema'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE TABLE test_sch (id INT PRIMARY KEY)",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO test_sch VALUES (11)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_commit('-Am', 'add test_sch')",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_id, to_id, diff_type FROM newschema.dolt_diff_test_sch WHERE  to_commit=HASHOF('HEAD')",
                    expected: Expected::Rows {
                        columns: &[Column("from_id", INT4), Column("to_id", INT4), Column("diff_type", TEXT)],
                        rows: &[
                            &[Null, T("11"), T("added")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_id, to_id, diff_type FROM dolt_diff_test_sch WHERE to_commit=HASHOF('HEAD')",
                    expected: Expected::Rows {
                        columns: &[Column("from_id", INT4), Column("to_id", INT4), Column("diff_type", TEXT)],
                        rows: &[
                            &[Null, T("11"), T("added")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_id, to_id, diff_type FROM dolt_diff_test WHERE to_commit=HASHOF('HEAD')",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: dolt_diff_test", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_id, to_id, diff_type FROM public.dolt_diff_test WHERE to_commit=HASHOF('HEAD^1')",
                    expected: Expected::Rows {
                        columns: &[Column("from_id", INT4), Column("to_id", INT4), Column("diff_type", TEXT)],
                        rows: &[
                            &[Null, T("10"), T("added")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT to_id FROM public.dolt_diff_test_sch WHERE to_commit=HASHOF('HEAD^1')",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: dolt_diff_test_sch", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT to_id FROM newschema.dolt_diff_test WHERE to_commit=HASHOF('HEAD')",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: dolt_diff_test", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE TABLE test (id INT PRIMARY KEY)",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES (12)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_commit('-Am', 'add test')",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_id, to_id, diff_type FROM newschema.dolt_diff_test WHERE  to_commit=HASHOF('HEAD')",
                    expected: Expected::Rows {
                        columns: &[Column("from_id", INT4), Column("to_id", INT4), Column("diff_type", TEXT)],
                        rows: &[
                            &[Null, T("12"), T("added")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_id, to_id, diff_type FROM dolt_diff_test WHERE to_commit=HASHOF('HEAD')",
                    expected: Expected::Rows {
                        columns: &[Column("from_id", INT4), Column("to_id", INT4), Column("diff_type", TEXT)],
                        rows: &[
                            &[Null, T("12"), T("added")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_id, to_id, diff_type FROM public.dolt_diff_test WHERE to_commit=HASHOF('HEAD~2')",
                    expected: Expected::Rows {
                        columns: &[Column("from_id", INT4), Column("to_id", INT4), Column("diff_type", TEXT)],
                        rows: &[
                            &[Null, T("10"), T("added")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SET search_path = newschema, public",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_id, to_id, diff_type FROM dolt_diff_test WHERE to_commit=HASHOF('HEAD')",
                    expected: Expected::Rows {
                        columns: &[Column("from_id", INT4), Column("to_id", INT4), Column("diff_type", TEXT)],
                        rows: &[
                            &[Null, T("12"), T("added")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SET search_path = public, newschema",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_id, to_id, diff_type FROM dolt_diff_test WHERE to_commit=HASHOF('HEAD~2')",
                    expected: Expected::Rows {
                        columns: &[Column("from_id", INT4), Column("to_id", INT4), Column("diff_type", TEXT)],
                        rows: &[
                            &[Null, T("10"), T("added")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "dolt diff with tablename commit index lookups",
            set_up_script: &[
                "CREATE TABLE test (id INT PRIMARY KEY, val INT)",
                "INSERT INTO test VALUES (1, 1)",
                "SELECT dolt_commit('-Am', 'commit 1')",
                "INSERT INTO test VALUES (2, 2)",
                "SELECT dolt_commit('-Am', 'commit 2')",
                "UPDATE test SET val = 3 WHERE id = 1",
                "SELECT dolt_commit('-Am', 'commit 3')",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_id, to_id, to_val, diff_type FROM dolt_diff_test WHERE from_commit = HASHOF('HEAD~1') AND to_commit = HASHOF('HEAD')",
                    expected: Expected::Rows {
                        columns: &[Column("from_id", INT4), Column("to_id", INT4), Column("to_val", INT4), Column("diff_type", TEXT)],
                        rows: &[
                            &[T("1"), T("1"), T("3"), T("modified")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_id, to_id, to_val, diff_type FROM dolt_diff_test WHERE from_commit = HASHOF('HEAD~2') AND to_commit = HASHOF('HEAD~1')",
                    expected: Expected::Rows {
                        columns: &[Column("from_id", INT4), Column("to_id", INT4), Column("to_val", INT4), Column("diff_type", TEXT)],
                        rows: &[
                            &[Null, T("2"), T("2"), T("added")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_id, to_id, to_val, diff_type FROM dolt_diff_test WHERE to_commit = HASHOF('HEAD')",
                    expected: Expected::Rows {
                        columns: &[Column("from_id", INT4), Column("to_id", INT4), Column("to_val", INT4), Column("diff_type", TEXT)],
                        rows: &[
                            &[T("1"), T("1"), T("3"), T("modified")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_id, to_id, to_val, diff_type FROM dolt_diff_test WHERE from_commit = HASHOF('HEAD~1')",
                    expected: Expected::Rows {
                        columns: &[Column("from_id", INT4), Column("to_id", INT4), Column("to_val", INT4), Column("diff_type", TEXT)],
                        rows: &[
                            &[T("1"), T("1"), T("3"), T("modified")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_id, to_id, to_val, diff_type FROM dolt_diff_test WHERE to_commit IN (HASHOF('HEAD'), HASHOF('HEAD~1')) ORDER BY to_id",
                    expected: Expected::Rows {
                        columns: &[Column("from_id", INT4), Column("to_id", INT4), Column("to_val", INT4), Column("diff_type", TEXT)],
                        rows: &[
                            &[T("1"), T("1"), T("3"), T("modified")],
                            &[Null, T("2"), T("2"), T("added")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_id, to_id, to_val, diff_type FROM dolt_diff_test WHERE to_commit NOT IN (HASHOF('HEAD'), 'WORKING') ORDER BY to_id",
                    expected: Expected::Rows {
                        columns: &[Column("from_id", INT4), Column("to_id", INT4), Column("to_val", INT4), Column("diff_type", TEXT)],
                        rows: &[
                            &[Null, T("1"), T("1"), T("added")],
                            &[Null, T("2"), T("2"), T("added")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "EXPLAIN SELECT to_id FROM dolt_diff_test WHERE to_commit = '0123456789abcdefghij0123456789ab'",
                    expected: Expected::Plan(&[PlanFact::IndexScan { table: "dolt_diff_test", columns: &["to_commit"], ranges: "[{[0123456789abcdefghij0123456789ab, 0123456789abcdefghij0123456789ab]}]" }]),
                    ..A
                },
                ScriptTestAssertion {
                    query: "EXPLAIN SELECT to_id FROM dolt_diff_test WHERE from_commit = '0123456789abcdefghij0123456789ab' AND to_commit = 'ab0123456789abcdefghij0123456789'",
                    expected: Expected::Plan(&[PlanFact::IndexScan { table: "dolt_diff_test", columns: &["from_commit"], ranges: "[{[0123456789abcdefghij0123456789ab, 0123456789abcdefghij0123456789ab]}]" }]),
                    ..A
                },
                ScriptTestAssertion {
                    query: "EXPLAIN SELECT to_id FROM dolt_diff_test WHERE to_commit < '0123456789abcdefghij0123456789ab'",
                    expected: Expected::Plan(&[PlanFact::FullScan { table: "dolt_diff_test" }]),
                    ..A
                },
                ScriptTestAssertion {
                    query: "EXPLAIN SELECT to_id FROM dolt_diff_test WHERE to_commit >= '0123456789abcdefghij0123456789ab'",
                    expected: Expected::Plan(&[PlanFact::FullScan { table: "dolt_diff_test" }]),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES (4, 4)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_id, to_id, to_val, diff_type FROM dolt_diff_test WHERE to_commit = 'WORKING'",
                    expected: Expected::Rows {
                        columns: &[Column("from_id", INT4), Column("to_id", INT4), Column("to_val", INT4), Column("diff_type", TEXT)],
                        rows: &[
                            &[Null, T("4"), T("4"), T("added")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "dolt_commit_diff subselect",
            set_up_script: &[
                "CREATE TABLE bug6 (id integer PRIMARY KEY, v text);",
                "INSERT INTO bug6 VALUES (1, 'a');",
                "SELECT dolt_add('-A');",
                "SELECT dolt_commit('--all', '--message', 'base', '--author', 'A <a@example.com>');",
                "UPDATE bug6 SET v = 'b' WHERE id = 1;",
                "SELECT dolt_add('-A');",
                "SELECT dolt_commit('--all', '--message', 'change', '--author', 'A <a@example.com>');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"SELECT to_id FROM dolt_commit_diff_bug6                                                                   
WHERE to_commit = (SELECT commit_hash FROM dolt.log ORDER BY date DESC LIMIT 1)                         
  AND from_commit = (SELECT commit_hash FROM dolt.log ORDER BY date DESC OFFSET 1 LIMIT 1);"#,
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "error querying table dolt_commit_diff_bug6: dolt_commit_diff_* tables must be filtered to a single 'to_commit'", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"SELECT to_id FROM dolt_commit_diff_bug6                                                                   
WHERE to_commit = dolt_hashof('HEAD')                         
  AND from_commit = dolt_hashof('HEAD~');"#,
                    expected: Expected::Rows {
                        columns: &[Column("to_id", INT4)],
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
            name: "dolt history with tablename",
            set_up_script: &[
                "CREATE TABLE test (id INT PRIMARY KEY)",
                "INSERT INTO test VALUES (10)",
                "SELECT dolt_commit('-Am', 'test commit', '--author', 'John Doe <johndoe@example.com>')",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT id, committer FROM dolt_history_test",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("committer", TEXT)],
                        rows: &[
                            &[T("10"), T("John Doe")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT id, committer FROM public.dolt_history_test",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("committer", TEXT)],
                        rows: &[
                            &[T("10"), T("John Doe")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_history_test.id FROM public.dolt_history_test",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("10")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM other.dolt_history_test",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: dolt_history_test", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM public.dolt_history_none",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: dolt_history_none", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE SCHEMA newschema",
                    expected: Expected::Tag("CREATE SCHEMA"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SET search_path = 'newschema'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE TABLE test_sch (id INT PRIMARY KEY)",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO test_sch VALUES (11)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_commit('-Am', 'add test_sch', '--author', 'Another Doe <adoe@example.com>')",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT id, committer FROM newschema.dolt_history_test_sch",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("committer", TEXT)],
                        rows: &[
                            &[T("11"), T("Another Doe")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT id, committer FROM dolt_history_test_sch",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("committer", TEXT)],
                        rows: &[
                            &[T("11"), T("Another Doe")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT id, committer FROM dolt_history_test",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: dolt_history_test", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT id, committer FROM public.dolt_history_test order by id, committer",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("committer", TEXT)],
                        rows: &[
                            &[T("10"), T("Another Doe")],
                            &[T("10"), T("John Doe")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT id FROM public.dolt_history_test_sch",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: dolt_history_test_sch", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT id, committer FROM newschema.dolt_history_test",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: dolt_history_test", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE TABLE test (id INT PRIMARY KEY)",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES (12)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_commit('-Am', 'add test')",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT id, committer FROM newschema.dolt_history_test",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("committer", TEXT)],
                        rows: &[
                            &[T("12"), T("postgres")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT id, committer FROM dolt_history_test",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("committer", TEXT)],
                        rows: &[
                            &[T("12"), T("postgres")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT id, committer FROM public.dolt_history_test order by id, committer",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("committer", TEXT)],
                        rows: &[
                            &[T("10"), T("Another Doe")],
                            &[T("10"), T("John Doe")],
                            &[T("10"), T("postgres")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SET search_path = newschema, public",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT id, committer FROM dolt_history_test",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("committer", TEXT)],
                        rows: &[
                            &[T("12"), T("postgres")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "dolt ignore",
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_ignore",
                    expected: Expected::Rows {
                        columns: &[Column("pattern", TEXT), Column("ignored", BOOL)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO dolt_ignore VALUES ('generated_*', true), ('generated_exception', false)",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_ignore",
                    expected: Expected::Rows {
                        columns: &[Column("pattern", TEXT), Column("ignored", BOOL)],
                        rows: &[
                            &[T("generated_*"), T("t")],
                            &[T("generated_exception"), T("f")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_ignore WHERE ignored=false",
                    expected: Expected::Rows {
                        columns: &[Column("pattern", TEXT), Column("ignored", BOOL)],
                        rows: &[
                            &[T("generated_exception"), T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM public.dolt_ignore",
                    expected: Expected::Rows {
                        columns: &[Column("pattern", TEXT), Column("ignored", BOOL)],
                        rows: &[
                            &[T("generated_*"), T("t")],
                            &[T("generated_exception"), T("f")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_ignore.pattern FROM public.dolt_ignore",
                    expected: Expected::Rows {
                        columns: &[Column("pattern", TEXT)],
                        rows: &[
                            &[T("generated_*")],
                            &[T("generated_exception")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT name FROM other.dolt_ignore",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: dolt_ignore", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_diff_summary('main', 'WORKING')",
                    expected: Expected::Rows {
                        columns: &[Column("from_table_name", TEXT), Column("to_table_name", TEXT), Column("diff_type", TEXT), Column("data_change", BOOL), Column("schema_change", BOOL)],
                        rows: &[
                            &[T(""), T("public.dolt_ignore"), T("added"), T("t"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT diff_type, from_pattern, to_pattern FROM dolt_diff('main', 'WORKING', 'dolt_ignore')",
                    expected: Expected::Rows {
                        columns: &[Column("diff_type", TEXT), Column("from_pattern", TEXT), Column("to_pattern", TEXT)],
                        rows: &[
                            &[T("added"), Null, T("generated_*")],
                            &[T("added"), Null, T("generated_exception")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE foo (pk int);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE generated_foo (pk int);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE generated_exception (pk int);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_add('-A');",
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
                    query: "SELECT * FROM dolt_status;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[
                            &[T("public.dolt_ignore"), T("t"), T("new table")],
                            &[T("public.foo"), T("t"), T("new table")],
                            &[T("public.generated_exception"), T("t"), T("new table")],
                            &[T("public.generated_foo"), T("f"), T("new table")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SCHEMA newschema",
                    expected: Expected::Tag("CREATE SCHEMA"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO newschema.dolt_ignore VALUES ('test_*', true)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path = 'newschema'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_ignore",
                    expected: Expected::Rows {
                        columns: &[Column("pattern", TEXT), Column("ignored", BOOL)],
                        rows: &[
                            &[T("test_*"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO dolt_ignore VALUES ('generated_exception', true)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_ignore",
                    expected: Expected::Rows {
                        columns: &[Column("pattern", TEXT), Column("ignored", BOOL)],
                        rows: &[
                            &[T("generated_exception"), T("t")],
                            &[T("test_*"), T("t")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM newschema.dolt_ignore",
                    expected: Expected::Rows {
                        columns: &[Column("pattern", TEXT), Column("ignored", BOOL)],
                        rows: &[
                            &[T("generated_exception"), T("t")],
                            &[T("test_*"), T("t")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM public.dolt_ignore",
                    expected: Expected::Rows {
                        columns: &[Column("pattern", TEXT), Column("ignored", BOOL)],
                        rows: &[
                            &[T("generated_*"), T("t")],
                            &[T("generated_exception"), T("f")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_diff_summary('main', 'WORKING', 'dolt_ignore')",
                    expected: Expected::Rows {
                        columns: &[Column("from_table_name", TEXT), Column("to_table_name", TEXT), Column("diff_type", TEXT), Column("data_change", BOOL), Column("schema_change", BOOL)],
                        rows: &[
                            &[T(""), T("newschema.dolt_ignore"), T("added"), T("t"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT pattern FROM public.dolt_ignore",
                    expected: Expected::Rows {
                        columns: &[Column("pattern", TEXT)],
                        rows: &[
                            &[T("generated_*")],
                            &[T("generated_exception")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE foo (pk int);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE test_foo (pk int);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE generated_foo (pk int);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE generated_exception (pk int);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_add('-A');",
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
                    query: "SELECT * FROM dolt_status ORDER BY table_name;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[
                            &[T("newschema"), T("t"), T("new schema")],
                            &[T("newschema.dolt_ignore"), T("t"), T("new table")],
                            &[T("newschema.foo"), T("t"), T("new table")],
                            &[T("newschema.generated_exception"), T("f"), T("new table")],
                            &[T("newschema.generated_foo"), T("t"), T("new table")],
                            &[T("newschema.test_foo"), T("f"), T("new table")],
                            &[T("public.dolt_ignore"), T("t"), T("new table")],
                            &[T("public.foo"), T("t"), T("new table")],
                            &[T("public.generated_exception"), T("t"), T("new table")],
                            &[T("public.generated_foo"), T("f"), T("new table")],
                        ],
                        tag: "SELECT 10",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "dolt log",
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM dolt.log",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM dolt_log",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM dolt.log WHERE message IN ('Initialize data repository')",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM public.log",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: log", ..E }),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM log",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: log", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE log (id INT PRIMARY KEY)",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO log VALUES (1)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM log",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM dolt.log",
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
                    query: "SET search_path = 'dolt'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM log",
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
                    query: "SELECT * FROM public.log",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path = 'public'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM log",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
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
            name: "dolt merge status",
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT is_merging FROM dolt.merge_status",
                    expected: Expected::Rows {
                        columns: &[Column("is_merging", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT is_merging FROM dolt.merge_status WHERE is_merging=true",
                    expected: Expected::Rows {
                        columns: &[Column("is_merging", BOOL)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT is_merging FROM dolt_merge_status",
                    expected: Expected::Rows {
                        columns: &[Column("is_merging", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt.merge_status.is_merging FROM dolt.merge_status",
                    expected: Expected::Rows {
                        columns: &[Column("is_merging", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_merge_status.is_merging FROM dolt_merge_status",
                    expected: Expected::Rows {
                        columns: &[Column("is_merging", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM public.merge_status",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: merge_status", ..E }),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM merge_status",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: merge_status", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE merge_status (id INT PRIMARY KEY)",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO merge_status VALUES (1)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM merge_status",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT is_merging FROM dolt.merge_status",
                    expected: Expected::Rows {
                        columns: &[Column("is_merging", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path = 'dolt'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT is_merging FROM merge_status",
                    expected: Expected::Rows {
                        columns: &[Column("is_merging", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM public.merge_status",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path = 'public'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM merge_status",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path = public, dolt",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM merge_status",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM MERGE_STATUS",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
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
            name: "dolt status",
            set_up_script: &[
                "CREATE TABLE t (id INT PRIMARY KEY)",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt.status",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[
                            &[T("public.t"), T("f"), T("new table")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_status",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[
                            &[T("public.t"), T("f"), T("new table")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"DESCRIBE dolt."status""#,
                    expected: Expected::Rows {
                        columns: &[Column("Field", TEXT), Column("Type", TEXT), Column("Null", TEXT), Column("Key", TEXT), Column("Default", TEXT), Column("Extra", TEXT)],
                        rows: &[
                            &[T("table_name"), T("text"), T("NO"), T("PRI"), Null, T("")],
                            &[T("staged"), T("boolean"), T("NO"), T("PRI"), Null, T("")],
                            &[T("status"), T("text"), T("NO"), T("PRI"), Null, T("")],
                        ],
                        tag: "EXPLAIN",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "DESCRIBE dolt.status",
                    expected: Expected::Rows {
                        columns: &[Column("Field", TEXT), Column("Type", TEXT), Column("Null", TEXT), Column("Key", TEXT), Column("Default", TEXT), Column("Extra", TEXT)],
                        rows: &[
                            &[T("table_name"), T("text"), T("NO"), T("PRI"), Null, T("")],
                            &[T("staged"), T("boolean"), T("NO"), T("PRI"), Null, T("")],
                            &[T("status"), T("text"), T("NO"), T("PRI"), Null, T("")],
                        ],
                        tag: "EXPLAIN",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "DESCRIBE dolt_status",
                    expected: Expected::Rows {
                        columns: &[Column("Field", TEXT), Column("Type", TEXT), Column("Null", TEXT), Column("Key", TEXT), Column("Default", TEXT), Column("Extra", TEXT)],
                        rows: &[
                            &[T("table_name"), T("text"), T("NO"), T("PRI"), Null, T("")],
                            &[T("staged"), T("boolean"), T("NO"), T("PRI"), Null, T("")],
                            &[T("status"), T("text"), T("NO"), T("PRI"), Null, T("")],
                        ],
                        tag: "EXPLAIN",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt.status WHERE staged=true",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt.status.table_name FROM dolt.status",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT)],
                        rows: &[
                            &[T("public.t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_status.table_name FROM dolt_status",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT)],
                        rows: &[
                            &[T("public.t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM public.status",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: status", ..E }),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM status",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: status", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE status (id INT PRIMARY KEY)",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO status VALUES (1)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM status",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name FROM dolt.status",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT)],
                        rows: &[
                            &[T("public.status")],
                            &[T("public.t")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path = 'dolt'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name FROM status",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT)],
                        rows: &[
                            &[T("public.status")],
                            &[T("public.t")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM public.status",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path = 'public'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM status",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path = public, dolt",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM status",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM STATUS",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
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
            name: "dolt tags",
            set_up_script: &[
                "SELECT dolt_tag('v1')",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT tag_name FROM dolt.tags",
                    expected: Expected::Rows {
                        columns: &[Column("tag_name", TEXT)],
                        rows: &[
                            &[T("v1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT tag_name FROM dolt_tags",
                    expected: Expected::Rows {
                        columns: &[Column("tag_name", TEXT)],
                        rows: &[
                            &[T("v1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt.tags.tag_name FROM dolt.tags",
                    expected: Expected::Rows {
                        columns: &[Column("tag_name", TEXT)],
                        rows: &[
                            &[T("v1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_tags.tag_name FROM dolt_tags",
                    expected: Expected::Rows {
                        columns: &[Column("tag_name", TEXT)],
                        rows: &[
                            &[T("v1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM public.tags",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: tags", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM tags",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: tags", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE TABLE tags (id INT PRIMARY KEY)",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO tags VALUES (1)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM tags",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT tag_name FROM dolt.tags",
                    expected: Expected::Rows {
                        columns: &[Column("tag_name", TEXT)],
                        rows: &[
                            &[T("v1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE SCHEMA dolt",
                    expected: Expected::Error(Diagnostic { code: "42P06", message: "can't create schema dolt; schema exists", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SET search_path = 'dolt'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT tag_name FROM tags",
                    expected: Expected::Rows {
                        columns: &[Column("tag_name", TEXT)],
                        rows: &[
                            &[T("v1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM public.tags",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SET search_path = 'public'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM tags",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SET search_path = public, dolt",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM tags",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM TAGS",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
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
            name: "dolt procedures",
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_procedures",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT), Column("create_stmt", TEXT), Column("created_at", TIMESTAMP), Column("modified_at", TIMESTAMP), Column("sql_mode", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM public.dolt_procedures",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT), Column("create_stmt", TEXT), Column("created_at", TIMESTAMP), Column("modified_at", TIMESTAMP), Column("sql_mode", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_procedures.name FROM public.dolt_procedures",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT name FROM other.dolt_procedures",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: dolt_procedures", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SCHEMA newschema",
                    expected: Expected::Tag("CREATE SCHEMA"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path = 'newschema'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM newschema.dolt_procedures",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT), Column("create_stmt", TEXT), Column("created_at", TIMESTAMP), Column("modified_at", TIMESTAMP), Column("sql_mode", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT name FROM dolt_procedures",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT name FROM public.dolt_procedures",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path = newschema, public",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT name FROM dolt_procedures",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "dolt rebase",
            set_up_script: &[
                "create table t (pk int primary key);",
                "select dolt_commit('-Am', 'creating table t');",
                "select dolt_branch('branch1');",
                "insert into t values (0);",
                "select dolt_commit('-am', 'inserting row 0');",
                "select dolt_checkout('branch1');",
                "insert into t values (1);",
                "select dolt_commit('-am', 'inserting row 1');",
                "insert into t values (2);",
                "select dolt_commit('-am', 'inserting row 2');",
                "insert into t values (3);",
                "select dolt_commit('-am', 'inserting row 3');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select message from dolt_log;",
                    expected: Expected::Rows {
                        columns: &[Column("message", TEXT)],
                        rows: &[
                            &[T("inserting row 3")],
                            &[T("inserting row 2")],
                            &[T("inserting row 1")],
                            &[T("creating table t")],
                            &[T("CREATE DATABASE")],
                            &[T("Initialize data repository")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_rebase('-i', 'main');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_rebase", RECORD)],
                        rows: &[
                            &[T(r#"(0,"interactive rebase started on branch dolt_rebase_branch1; adjust the rebase plan in the dolt_rebase table, then continue rebasing by calling dolt_rebase('--continue')")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select rebase_order, action, commit_message from dolt_rebase order by rebase_order;",
                    expected: Expected::Rows {
                        columns: &[Column("rebase_order", FLOAT4), Column("action", VARCHAR), Column("commit_message", TEXT)],
                        rows: &[
                            &[T("1"), T("pick"), T("inserting row 1")],
                            &[T("2"), T("pick"), T("inserting row 2")],
                            &[T("3"), T("pick"), T("inserting row 3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select rebase_order, action, commit_message from dolt.rebase order by rebase_order;",
                    expected: Expected::Rows {
                        columns: &[Column("rebase_order", FLOAT4), Column("action", VARCHAR), Column("commit_message", TEXT)],
                        rows: &[
                            &[T("1"), T("pick"), T("inserting row 1")],
                            &[T("2"), T("pick"), T("inserting row 2")],
                            &[T("3"), T("pick"), T("inserting row 3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select rebase.commit_message from dolt.rebase order by rebase_order;",
                    expected: Expected::Rows {
                        columns: &[Column("commit_message", TEXT)],
                        rows: &[
                            &[T("inserting row 1")],
                            &[T("inserting row 2")],
                            &[T("inserting row 3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_rebase.commit_message from dolt_rebase order by rebase_order;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: dolt_rebase", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM public.rebase",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: rebase", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM rebase",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: rebase", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE TABLE rebase (id INT PRIMARY KEY)",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO rebase VALUES (1)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM rebase",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT commit_message FROM dolt.rebase",
                    expected: Expected::Rows {
                        columns: &[Column("commit_message", TEXT)],
                        rows: &[
                            &[T("inserting row 1")],
                            &[T("inserting row 2")],
                            &[T("inserting row 3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE SCHEMA dolt",
                    expected: Expected::Error(Diagnostic { code: "42P06", message: "can't create schema dolt; schema exists", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SET search_path = 'dolt'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT commit_message FROM rebase",
                    expected: Expected::Rows {
                        columns: &[Column("commit_message", TEXT)],
                        rows: &[
                            &[T("inserting row 1")],
                            &[T("inserting row 2")],
                            &[T("inserting row 3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM public.rebase",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SET search_path = 'public'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM rebase",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SET search_path = public, dolt",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM rebase",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM REBASE",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "DROP TABLE public.rebase;",
                    expected: Expected::Tag("DROP TABLE"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "update dolt.rebase set action='reword', commit_message='insert rows' where rebase_order=1;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "update dolt.rebase set action='drop' where rebase_order=2;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "update dolt_rebase set action='fixup' where rebase_order=3;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select rebase_order, action, commit_message from dolt_rebase order by rebase_order;",
                    expected: Expected::Rows {
                        columns: &[Column("rebase_order", FLOAT4), Column("action", VARCHAR), Column("commit_message", TEXT)],
                        rows: &[
                            &[T("1"), T("reword"), T("insert rows")],
                            &[T("2"), T("drop"), T("inserting row 2")],
                            &[T("3"), T("fixup"), T("inserting row 3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select rebase_order, action, commit_message from dolt.rebase order by rebase_order;",
                    expected: Expected::Rows {
                        columns: &[Column("rebase_order", FLOAT4), Column("action", VARCHAR), Column("commit_message", TEXT)],
                        rows: &[
                            &[T("1"), T("reword"), T("insert rows")],
                            &[T("2"), T("drop"), T("inserting row 2")],
                            &[T("3"), T("fixup"), T("inserting row 3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_rebase('--continue');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_rebase", RECORD)],
                        rows: &[
                            &[T(r#"(0,"Successfully rebased and updated refs/heads/branch1")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select message from dolt_log;",
                    expected: Expected::Rows {
                        columns: &[Column("message", TEXT)],
                        rows: &[
                            &[T("insert rows")],
                            &[T("inserting row 0")],
                            &[T("creating table t")],
                            &[T("CREATE DATABASE")],
                            &[T("Initialize data repository")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select * from dolt_rebase;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: dolt_rebase", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select * from dolt.rebase;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: rebase", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "dolt remote branches",
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT name FROM dolt.remote_branches",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT name FROM dolt_remote_branches",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt.remote_branches.name FROM dolt.remote_branches",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_remote_branches.name FROM dolt_remote_branches",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM public.remote_branches",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: remote_branches", ..E }),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM remote_branches",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: remote_branches", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE remote_branches (id INT PRIMARY KEY)",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO remote_branches VALUES (1)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM remote_branches",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT name FROM dolt.remote_branches",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path = 'dolt'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT name FROM remote_branches",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM public.remote_branches",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path = 'public'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM remote_branches",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path = public, dolt",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM remote_branches",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM REMOTE_BRANCHES",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
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
            name: "dolt remotes",
            set_up_script: &[
                "SELECT dolt_remote('add', 'origin', 'https://doltremoteapi.dolthub.com/dolthub/test')",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT name FROM dolt.remotes",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("origin")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT name FROM dolt_remotes",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("origin")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt.remotes.name FROM dolt.remotes",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("origin")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_remotes.name FROM dolt_remotes",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("origin")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT name FROM dolt.remotes WHERE name IN ('origin')",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("origin")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM public.remotes",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: remotes", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM remotes",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: remotes", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE TABLE remotes (id INT PRIMARY KEY)",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO remotes VALUES (1)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM remotes",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT name FROM dolt.remotes",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("origin")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SET search_path = 'dolt'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT name FROM remotes",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("origin")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM public.remotes",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SET search_path = 'public'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM remotes",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SET search_path = public, dolt",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM remotes",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM REMOTES",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
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
            name: "dolt schema conflicts",
            set_up_script: &[
                "CREATE TABLE test (pk int primary key, c0 varchar(20))",
                "SELECT dolt_commit('-Am', 'added table t')",
                "SELECT dolt_checkout('-b', 'other')",
                "ALTER TABLE test ALTER COLUMN c0 TYPE int",
                "SELECT dolt_commit('-am', 'altered t on branch other')",
                "SELECT dolt_checkout('main')",
                "ALTER TABLE test ALTER COLUMN c0 TYPE date",
                "SELECT dolt_commit('-am', 'altered t on branch main')",
                "START TRANSACTION",
                "SELECT dolt_merge('other')",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name FROM dolt.schema_conflicts",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT)],
                        rows: &[
                            &[T("test")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name FROM dolt_schema_conflicts",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT)],
                        rows: &[
                            &[T("test")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt.schema_conflicts.table_name FROM dolt.schema_conflicts",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT)],
                        rows: &[
                            &[T("test")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_schema_conflicts.table_name FROM dolt_schema_conflicts",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT)],
                        rows: &[
                            &[T("test")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SAVEPOINT probe",
                    expected: Expected::Tag("SAVEPOINT"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM public.schema_conflicts",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: schema_conflicts", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "ROLLBACK TO probe",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM schema_conflicts",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: schema_conflicts", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "ROLLBACK TO probe",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE TABLE schema_conflicts (id INT PRIMARY KEY)",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO schema_conflicts VALUES (1)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM schema_conflicts",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name FROM dolt.schema_conflicts",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT)],
                        rows: &[
                            &[T("test")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SET search_path = 'dolt'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name FROM schema_conflicts",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT)],
                        rows: &[
                            &[T("test")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM public.schema_conflicts",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SET search_path = 'public'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM schema_conflicts",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SET search_path = public, dolt",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM schema_conflicts",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM SCHEMA_CONFLICTS",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
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
            name: "dolt schemas",
            set_up_script: &[
                "create view myView as select 2 + 2",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_schemas",
                    expected: Expected::Rows {
                        columns: &[Column("type", TEXT), Column("name", TEXT), Column("fragment", TEXT), Column("extra", JSON), Column("sql_mode", TEXT)],
                        rows: &[
                            &[T("view"), T("myview"), T("create view myView as select 2 + 2"), T(r#"{"Bytes":"eyJDcmVhdGVkQXQiOjB9"}"#), T("ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM public.dolt_schemas",
                    expected: Expected::Rows {
                        columns: &[Column("type", TEXT), Column("name", TEXT), Column("fragment", TEXT), Column("extra", JSON), Column("sql_mode", TEXT)],
                        rows: &[
                            &[T("view"), T("myview"), T("create view myView as select 2 + 2"), T(r#"{"Bytes":"eyJDcmVhdGVkQXQiOjB9"}"#), T("ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_schemas.name FROM public.dolt_schemas",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("myview")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM public.myview",
                    expected: Expected::Rows {
                        columns: &[Column("2 + 2", INT4)],
                        rows: &[
                            &[T("4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT name FROM other.dolt_schemas",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: dolt_schemas", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_diff_summary('main', 'WORKING')",
                    expected: Expected::Rows {
                        columns: &[Column("from_table_name", TEXT), Column("to_table_name", TEXT), Column("diff_type", TEXT), Column("data_change", BOOL), Column("schema_change", BOOL)],
                        rows: &[
                            &[T(""), T("public.dolt_schemas"), T("added"), T("t"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_diff_summary('main', 'WORKING', 'dolt_schemas')",
                    expected: Expected::Rows {
                        columns: &[Column("from_table_name", TEXT), Column("to_table_name", TEXT), Column("diff_type", TEXT), Column("data_change", BOOL), Column("schema_change", BOOL)],
                        rows: &[
                            &[T(""), T("public.dolt_schemas"), T("added"), T("t"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_diff_summary('main', 'WORKING', 'dolt_schemas')",
                    expected: Expected::Rows {
                        columns: &[Column("from_table_name", TEXT), Column("to_table_name", TEXT), Column("diff_type", TEXT), Column("data_change", BOOL), Column("schema_change", BOOL)],
                        rows: &[
                            &[T(""), T("public.dolt_schemas"), T("added"), T("t"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT diff_type, from_name, to_name FROM dolt_diff('main', 'WORKING', 'dolt_schemas')",
                    expected: Expected::Rows {
                        columns: &[Column("diff_type", TEXT), Column("from_name", TEXT), Column("to_name", TEXT)],
                        rows: &[
                            &[T("added"), Null, T("myview")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT diff_type, from_name, to_name FROM dolt_diff('main', 'WORKING', 'dolt_schemas')",
                    expected: Expected::Rows {
                        columns: &[Column("diff_type", TEXT), Column("from_name", TEXT), Column("to_name", TEXT)],
                        rows: &[
                            &[T("added"), Null, T("myview")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SCHEMA newschema",
                    expected: Expected::Tag("CREATE SCHEMA"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path = 'newschema'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM myview",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: myview", ..E }),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM public.myview",
                    expected: Expected::Rows {
                        columns: &[Column("2 + 2", INT4)],
                        rows: &[
                            &[T("4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW testView AS SELECT 1 + 1",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM newschema.dolt_schemas",
                    expected: Expected::Rows {
                        columns: &[Column("type", TEXT), Column("name", TEXT), Column("fragment", TEXT), Column("extra", JSON), Column("sql_mode", TEXT)],
                        rows: &[
                            &[T("view"), T("testview"), T("CREATE VIEW testView AS SELECT 1 + 1"), T(r#"{"Bytes":"eyJDcmVhdGVkQXQiOjB9"}"#), T("ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT name FROM dolt_schemas",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("testview")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT table_schema, table_name FROM information_schema.views",
                    expected: Expected::Rows {
                        columns: &[Column("table_schema", VARCHAR), Column("table_name", VARCHAR)],
                        rows: &[
                            &[T("newschema"), T("testview")],
                            &[T("public"), T("myview")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_diff_summary('main', 'WORKING', 'dolt_schemas')",
                    expected: Expected::Rows {
                        columns: &[Column("from_table_name", TEXT), Column("to_table_name", TEXT), Column("diff_type", TEXT), Column("data_change", BOOL), Column("schema_change", BOOL)],
                        rows: &[
                            &[T(""), T("newschema.dolt_schemas"), T("added"), T("t"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_diff_summary('main', 'WORKING', 'public.dolt_schemas')",
                    expected: Expected::Rows {
                        columns: &[Column("from_table_name", TEXT), Column("to_table_name", TEXT), Column("diff_type", TEXT), Column("data_change", BOOL), Column("schema_change", BOOL)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT name FROM public.dolt_schemas",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("myview")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "DROP VIEW myView",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "the view postgres.myview does not exist", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP VIEW public.myView",
                    expected: Expected::Tag("DROP VIEW"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT name FROM public.dolt_schemas",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "create view public.myNewView as select 3 + 3",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT name FROM public.dolt_schemas",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("mynewview")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT name FROM dolt_schemas",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("testview")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path = newschema, public",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT name FROM dolt_schemas",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("testview")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_diff_summary('main', 'WORKING', 'dolt_schemas')",
                    expected: Expected::Rows {
                        columns: &[Column("from_table_name", TEXT), Column("to_table_name", TEXT), Column("diff_type", TEXT), Column("data_change", BOOL), Column("schema_change", BOOL)],
                        rows: &[
                            &[T(""), T("newschema.dolt_schemas"), T("added"), T("t"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path = 'public'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW testView AS SELECT 4 + 4",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT name, fragment FROM dolt_schemas",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT), Column("fragment", TEXT)],
                        rows: &[
                            &[T("mynewview"), T("create view public.myNewView as select 3 + 3")],
                            &[T("testview"), T("CREATE VIEW testView AS SELECT 4 + 4")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT name, fragment FROM newschema.dolt_schemas",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT), Column("fragment", TEXT)],
                        rows: &[
                            &[T("testview"), T("CREATE VIEW testView AS SELECT 1 + 1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT name, fragment FROM dolt_schemas",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT), Column("fragment", TEXT)],
                        rows: &[
                            &[T("mynewview"), T("create view public.myNewView as select 3 + 3")],
                            &[T("testview"), T("CREATE VIEW testView AS SELECT 4 + 4")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "DROP VIEW IF EXISTS noexist.testView",
                    expected: Expected::Tag("DROP VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP VIEW IF EXISTS newschema.testView",
                    expected: Expected::Tag("DROP VIEW"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT name FROM newschema.dolt_schemas",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT name FROM dolt_schemas",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("mynewview")],
                            &[T("testview")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "dolt workspace with tablename",
            set_up_script: &[
                "CREATE TABLE test (id INT PRIMARY KEY)",
                "INSERT INTO test VALUES (10)",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT id, staged, from_id, to_id FROM dolt_workspace_test",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT8), Column("staged", BOOL), Column("from_id", INT4), Column("to_id", INT4)],
                        rows: &[
                            &[T("0"), T("f"), Null, T("10")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT id, staged, from_id, to_id FROM public.dolt_workspace_test",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT8), Column("staged", BOOL), Column("from_id", INT4), Column("to_id", INT4)],
                        rows: &[
                            &[T("0"), T("f"), Null, T("10")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_workspace_test.id FROM public.dolt_workspace_test",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM other.dolt_workspace_test",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: dolt_workspace_test", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM public.dolt_workspace_none",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT8), Column("staged", BOOL)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SCHEMA newschema",
                    expected: Expected::Tag("CREATE SCHEMA"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path = 'newschema'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE test_sch (id INT PRIMARY KEY)",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test_sch VALUES (11)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_add('test_sch')",
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
                    query: "SELECT id, staged, from_id, to_id FROM newschema.dolt_workspace_test_sch",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT8), Column("staged", BOOL), Column("from_id", INT4), Column("to_id", INT4)],
                        rows: &[
                            &[T("0"), T("t"), Null, T("11")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT id, staged, from_id, to_id FROM dolt_workspace_test_sch",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT8), Column("staged", BOOL), Column("from_id", INT4), Column("to_id", INT4)],
                        rows: &[
                            &[T("0"), T("t"), Null, T("11")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT id, staged, from_id, to_id FROM dolt_workspace_test_sch WHERE staged=true",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT8), Column("staged", BOOL), Column("from_id", INT4), Column("to_id", INT4)],
                        rows: &[
                            &[T("0"), T("t"), Null, T("11")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT id, staged, from_id, to_id FROM dolt_workspace_test_sch WHERE staged=false",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT8), Column("staged", BOOL), Column("from_id", INT4), Column("to_id", INT4)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_workspace_test",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT8), Column("staged", BOOL)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT id, staged, from_id, to_id FROM public.dolt_workspace_test",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT8), Column("staged", BOOL), Column("from_id", INT4), Column("to_id", INT4)],
                        rows: &[
                            &[T("0"), T("f"), Null, T("10")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM public.dolt_workspace_test_sch",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT8), Column("staged", BOOL)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM newschema.dolt_workspace_test",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT8), Column("staged", BOOL)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE test (id INT PRIMARY KEY)",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES (12)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT id, staged, from_id, to_id FROM newschema.dolt_workspace_test",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT8), Column("staged", BOOL), Column("from_id", INT4), Column("to_id", INT4)],
                        rows: &[
                            &[T("0"), T("f"), Null, T("12")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT id, staged, from_id, to_id FROM dolt_workspace_test",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT8), Column("staged", BOOL), Column("from_id", INT4), Column("to_id", INT4)],
                        rows: &[
                            &[T("0"), T("f"), Null, T("12")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT id, staged, from_id, to_id FROM public.dolt_workspace_test",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT8), Column("staged", BOOL), Column("from_id", INT4), Column("to_id", INT4)],
                        rows: &[
                            &[T("0"), T("f"), Null, T("10")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path = newschema, public",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT id, staged, from_id, to_id FROM dolt_workspace_test",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT8), Column("staged", BOOL), Column("from_id", INT4), Column("to_id", INT4)],
                        rows: &[
                            &[T("0"), T("f"), Null, T("12")],
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
