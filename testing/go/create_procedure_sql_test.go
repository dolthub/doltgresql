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

package _go

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestCreateProcedureLanguageSql(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "procedure with insert returning",
			SetUpScript: []string{
				`CREATE TABLE public.games (
    id bigint NOT NULL,
    game_id character varying(4) NOT NULL,
    host_connection_id character varying(50) NOT NULL
);`,
				`CREATE SEQUENCE public.games_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;`,
				`ALTER SEQUENCE public.games_id_seq OWNED BY public.games.id;`,
				`ALTER TABLE ONLY public.games ALTER COLUMN id SET DEFAULT nextval('public.games_id_seq'::regclass);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE PROCEDURE public.add(INOUT new_host_connection_id character varying)
    LANGUAGE sql
    AS $$
	INSERT INTO public.games (
		game_id,
		host_connection_id
	)
	VALUES (2222, new_host_connection_id)
	RETURNING game_id;
$$;`,
					Expected: []sql.Row{},
				},
				{
					SkipResultsCheck: true, // TODO: need fix for returning results
					Query:            `CALL add('f')`,
					Expected:         []sql.Row{{"2222"}},
				},
				{
					Query:    `SELECT id, game_id, host_connection_id FROM games`,
					Expected: []sql.Row{{1, "2222", "f"}},
				},
				{
					Query: `CREATE PROCEDURE public.create_game(INOUT new_host_connection_id character varying)
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
$$;`,
					Expected: []sql.Row{},
				},
				{
					Query: `CALL create_game('d')`,
				},
				{
					Query:    `SELECT id, host_connection_id FROM games`,
					Expected: []sql.Row{{1, "f"}, {2, "d"}},
				},
			},
		},
	})
}
