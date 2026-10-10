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
fn test_psql_commands() {
    run_scripts(&[
        ScriptTest {
            name: "operator keyword",
            assertions: &[
                ScriptTestAssertion {
                    query: "select 1 OPERATOR(pg_catalog.+) 1",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select 1 OPERATOR(PG_CATALOG.+) 1",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select 1 OPERATOR(myschema.+) 1",
                    expected: Expected::Error(Diagnostic { code: "3F000", message: r#"schema "myschema" does not exist"#, position: 10, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select 1 OPERATOR(pg_catalog.<) 1",
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
                    query: "select 1 OPERATOR(myschema.<) 1",
                    expected: Expected::Error(Diagnostic { code: "3F000", message: r#"schema "myschema" does not exist"#, position: 10, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select 1 OPERATOR(pg_catalog.<=) 1",
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
                    query: "select 1 OPERATOR(pg_catalog.=) 1",
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
                    query: "select 'hello' OPERATOR(pg_catalog.~) 'hello';",
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
            name: r#"\dt tablename"#,
            set_up_script: &[
                "CREATE TABLE test_table (id INT PRIMARY KEY, name TEXT);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT n.nspname as "Schema",   c.relname as "Name",   CASE c.relkind WHEN 'r' THEN 'table' WHEN 'v' THEN 'view' WHEN 'm' THEN 'materialized view' WHEN 'i' THEN 'index' WHEN 'S' THEN 'sequence' WHEN 't' THEN 'TOAST table' WHEN 'f' THEN 'foreign table' WHEN 'p' THEN 'partitioned table' WHEN 'I' THEN 'partitioned index' END as "Type",   pg_catalog.pg_get_userbyid(c.relowner) as "Owner" FROM pg_catalog.pg_class c      LEFT JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace      LEFT JOIN pg_catalog.pg_am am ON am.oid = c.relam WHERE c.relkind IN ('r','p','t','s','')   AND c.relname OPERATOR(pg_catalog.~) '^(test_table)$' COLLATE pg_catalog.default   AND pg_catalog.pg_table_is_visible(c.oid) ORDER BY 1,2;"#,
                    expected: Expected::Rows {
                        columns: &[Column("Schema", NAME), Column("Name", NAME), Column("Type", TEXT), Column("Owner", NAME)],
                        rows: &[
                            &[T("public"), T("test_table"), T("table"), T("postgres")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"\df"#,
            set_up_script: &[
                "CREATE FUNCTION add_two(a int, b int) RETURNS int LANGUAGE sql AS 'SELECT a + b';",
                "CREATE PROCEDURE noop_proc() LANGUAGE sql AS $$ SELECT 1 $$;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT n.nspname as "Schema",
  p.proname as "Name",
  pg_catalog.pg_get_function_result(p.oid) as "Result data type",
  pg_catalog.pg_get_function_arguments(p.oid) as "Argument data types",
 CASE p.prokind
  WHEN 'a' THEN 'agg'
  WHEN 'w' THEN 'window'
  WHEN 'p' THEN 'proc'
  ELSE 'func'
 END as "Type"
FROM pg_catalog.pg_proc p
     LEFT JOIN pg_catalog.pg_namespace n ON n.oid = p.pronamespace
WHERE pg_catalog.pg_function_is_visible(p.oid)
      AND n.nspname <> 'pg_catalog'
      AND n.nspname <> 'information_schema'
ORDER BY 1, 2, 4;"#,
                    expected: Expected::Rows {
                        columns: &[Column("Schema", NAME), Column("Name", NAME), Column("Result data type", TEXT), Column("Argument data types", TEXT), Column("Type", TEXT)],
                        rows: &[
                            &[T("public"), T("add_two"), T("integer"), T("a integer, b integer"), T("func")],
                            &[T("public"), T("noop_proc"), Null, T(""), T("proc")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"\d tablename triggers"#,
            set_up_script: &[
                "CREATE TABLE test_table (id INT PRIMARY KEY, name TEXT);",
                "CREATE FUNCTION trig_fn() RETURNS trigger AS $$ BEGIN RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "CREATE TRIGGER test_trig BEFORE INSERT ON test_table FOR EACH ROW EXECUTE FUNCTION trig_fn();",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT t.tgname, pg_catalog.pg_get_triggerdef(t.oid, true), t.tgenabled, t.tgisinternal,
  CASE WHEN t.tgparentid != 0 THEN
    (SELECT u.tgrelid::pg_catalog.regclass
     FROM pg_catalog.pg_trigger AS u,
          pg_catalog.pg_partition_ancestors(t.tgrelid) WITH ORDINALITY AS a(relid, depth)
     WHERE u.tgname = t.tgname AND u.tgrelid = a.relid
           AND u.tgparentid = 0
     ORDER BY a.depth LIMIT 1)
  END AS parent
FROM pg_catalog.pg_trigger t
WHERE t.tgrelid = 'test_table'::regclass AND (NOT t.tgisinternal OR (t.tgisinternal AND t.tgenabled = 'D'))
ORDER BY 1;"#,
                    expected: Expected::Rows {
                        columns: &[Column("tgname", NAME), Column("pg_get_triggerdef", TEXT), Column("tgenabled", CHAR), Column("tgisinternal", BOOL), Column("parent", REGCLASS)],
                        rows: &[
                            &[T("test_trig"), T("CREATE TRIGGER test_trig BEFORE INSERT ON test_table FOR EACH ROW EXECUTE FUNCTION trig_fn()"), T("O"), T("f"), Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"\d tablename"#,
            set_up_script: &[
                "CREATE TABLE test_table (id INT PRIMARY KEY, name TEXT);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT pol.polname, pol.polpermissive,
       CASE WHEN pol.polroles = '{0}' THEN NULL ELSE pg_catalog.array_to_string(array(select rolname from pg_catalog.pg_roles where oid = any (pol.polroles) order by 1),',') END,
       pg_catalog.pg_get_expr(pol.polqual, pol.polrelid),
       pg_catalog.pg_get_expr(pol.polwithcheck, pol.polrelid),
       CASE pol.polcmd
           WHEN 'r' THEN 'SELECT'
           WHEN 'a' THEN 'INSERT'
           WHEN 'w' THEN 'UPDATE'
           WHEN 'd' THEN 'DELETE'
           END AS cmd
FROM pg_catalog.pg_policy pol
WHERE pol.polrelid = '4131846889' ORDER BY 1;"#,
                    expected: Expected::Rows {
                        columns: &[Column("polname", NAME), Column("polpermissive", BOOL), Column("array_to_string", TEXT), Column("pg_get_expr", TEXT), Column("pg_get_expr", TEXT), Column("cmd", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT oid, stxrelid::pg_catalog.regclass,
stxnamespace::pg_catalog.regnamespace::pg_catalog.text AS nsp,
 stxname, pg_catalog.pg_get_statisticsobjdef_columns(oid) AS columns,
          'd' = any(stxkind) AS ndist_enabled,
          'f' = any(stxkind) AS deps_enabled,
          'm' = any(stxkind) AS mcv_enabled,
          stxstattarget FROM pg_catalog.pg_statistic_ext 
                        WHERE stxrelid = '4131846889' ORDER BY nsp, stxname;"#,
                    expected: Expected::Rows {
                        columns: &[Column("oid", OID), Column("stxrelid", REGCLASS), Column("nsp", TEXT), Column("stxname", NAME), Column("columns", TEXT), Column("ndist_enabled", BOOL), Column("deps_enabled", BOOL), Column("mcv_enabled", BOOL), Column("stxstattarget", INT2)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT pubname
, NULL
, NULL
FROM pg_catalog.pg_publication p
JOIN pg_catalog.pg_publication_namespace pn ON p.oid = pn.pnpubid
JOIN pg_catalog.pg_class pc ON pc.relnamespace = pn.pnnspid
WHERE pc.oid ='4131846889' and pg_catalog.pg_relation_is_publishable('4131846889')
UNION
SELECT pubname
, pg_get_expr(pr.prqual, c.oid)
, (CASE WHEN pr.prattrs IS NOT NULL THEN
(SELECT string_agg(attname, ', ')
FROM pg_catalog.generate_series(0, pg_catalog.array_upper(pr.prattrs::pg_catalog.int2[], 1)) s,
pg_catalog.pg_attribute
WHERE attrelid = pr.prrelid AND attnum = prattrs[s])
ELSE NULL END) FROM pg_catalog.pg_publication p
JOIN pg_catalog.pg_publication_rel pr ON p.oid = pr.prpubid
JOIN pg_catalog.pg_class c ON c.oid = pr.prrelid
WHERE pr.prrelid = '4131846889'
UNION
SELECT pubname
, NULL
, NULL
FROM pg_catalog.pg_publication p
WHERE p.puballtables AND pg_catalog.pg_relation_is_publishable('4131846889')
ORDER BY 1;"#,
                    expected: Expected::Rows {
                        columns: &[Column("pubname", NAME), Column("?column?", TEXT), Column("?column?", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}
