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
fn test_create_procedure_language_sql() {
    run_scripts(&[
        ScriptTest {
            name: "procedure with insert returning",
            set_up_script: &[
                r#"CREATE TABLE public.games (
    id bigint NOT NULL,
    game_id character varying(4) NOT NULL,
    host_connection_id character varying(50) NOT NULL
);"#,
                "CREATE SEQUENCE public.games_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;",
                "ALTER SEQUENCE public.games_id_seq OWNED BY public.games.id;",
                "ALTER TABLE ONLY public.games ALTER COLUMN id SET DEFAULT nextval('public.games_id_seq'::regclass);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE PROCEDURE public.add(INOUT new_host_connection_id character varying)
    LANGUAGE sql
    AS $$
	INSERT INTO public.games (
		game_id,
		host_connection_id
	)
	VALUES (2222, new_host_connection_id)
	RETURNING game_id;
$$;"#,
                    expected: Expected::Tag("CREATE PROCEDURE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CALL add('f')",
                    expected: Expected::Rows {
                        columns: &[Column("new_host_connection_id", VARCHAR)],
                        rows: &[
                            &[T("2222")],
                        ],
                        tag: "CALL",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, game_id, host_connection_id FROM games",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT8), Column("game_id", VARCHAR), Column("host_connection_id", VARCHAR)],
                        rows: &[
                            &[T("1"), T("2222"), T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE PROCEDURE public.create_game(INOUT new_host_connection_id character varying)
    LANGUAGE sql
    AS $$
	WITH new_game_id_holder (new_game_id) AS (
		SELECT n.random_number
		FROM (
			SELECT LPAD(FLOOR(random() * 10000)::varchar, 4, '0') AS random_number
			FROM generate_series(1, (SELECT COUNT(*) FROM public.games) + 10)
		) AS n
		LEFT OUTER JOIN 
			public.games AS g on g.game_id = n.random_number
		WHERE g.id IS NULL
		LIMIT 1
	)
	INSERT INTO public.games (
		game_id,
		host_connection_id
	)
	VALUES ( 
		(SELECT new_game_id FROM new_game_id_holder),
		new_host_connection_id
	)
	RETURNING game_id;
$$;"#,
                    expected: Expected::Tag("CREATE PROCEDURE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CALL create_game('d')",
                    expected: Expected::Tag("CALL"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, host_connection_id FROM games",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT8), Column("host_connection_id", VARCHAR)],
                        rows: &[
                            &[T("1"), T("f")],
                            &[T("2"), T("d")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "procedure with default expression in parameter",
            set_up_script: &[
                "CREATE TABLE cp_test (a int, b text);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE OR REPLACE PROCEDURE ptest5(a int, b text, c int default 100)
							LANGUAGE SQL
							AS $$
								INSERT INTO cp_test VALUES(a, b);
								INSERT INTO cp_test VALUES(c, b);
							$$;"#,
                    expected: Expected::Tag("CREATE PROCEDURE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CALL ptest5(10, 'Hello', 20);",
                    expected: Expected::Tag("CALL"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM cp_test",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", TEXT)],
                        rows: &[
                            &[T("10"), T("Hello")],
                            &[T("20"), T("Hello")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CALL ptest5(50, 'Bye');",
                    expected: Expected::Tag("CALL"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM cp_test",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", TEXT)],
                        rows: &[
                            &[T("10"), T("Hello")],
                            &[T("20"), T("Hello")],
                            &[T("50"), T("Bye")],
                            &[T("100"), T("Bye")],
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
