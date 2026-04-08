// Copyright 2025 Dolthub, Inc.
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

func TestCreateFunctionsLanguageSQL(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name:        "unnamed parameter",
			SetUpScript: []string{},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `CREATE FUNCTION alt_func1(int) RETURNS int LANGUAGE sql AS 'SELECT $1 + 1';`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT alt_func1(3);`,
					Expected: []sql.Row{{4}},
				},
			},
		},
		{
			Name:        "named parameter",
			SetUpScript: []string{},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `CREATE FUNCTION alt_func1(x int) RETURNS int LANGUAGE sql AS 'SELECT x + 1';`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT alt_func1(3);`,
					Expected: []sql.Row{{4}},
				},
				{
					Query:    `CREATE FUNCTION sub_numbers(x int, y int) RETURNS int LANGUAGE sql AS 'SELECT y - x';`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT sub_numbers(1, 2);`,
					Expected: []sql.Row{{1}},
				},
			},
		},
		{
			Name:        "unknown functions",
			SetUpScript: []string{},
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION get_grade_description(score INT)
							RETURNS TEXT
							LANGUAGE SQL
							AS $$
								SELECT
									CASE
										WHEN score >= 90 THEN 'Excellent'
										WHEN score >= 75 THEN 'Good'
										WHEN score >= 50 THEN 'Average'
									ELSE 'Fail'
									END;
							$$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT get_grade_description(92);`,
					Expected: []sql.Row{{"Excellent"}},
				},
				{
					Query:    `SELECT get_grade_description(65);`,
					Expected: []sql.Row{{"Average"}},
				},
			},
		},
		{
			Name:        "nested functions",
			SetUpScript: []string{},
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION calculate_double_sum(x INT, y INT)
							RETURNS INT
							LANGUAGE SQL
							AS $$
								SELECT add_numbers(x, y) * 2;
							$$;`,
					// TODO: error message should be:  function add_numbers(integer, integer) does not exist
					ExpectedErr: "function: 'add_numbers' not found",
				},
				{
					Query:    `CREATE FUNCTION add_numbers(int, int) RETURNS int LANGUAGE sql AS 'SELECT $1 + $2';`,
					Expected: []sql.Row{},
				},
				{
					Query: `CREATE FUNCTION calculate_double_sum(x INT, y INT)
							RETURNS INT
							LANGUAGE SQL
							AS $$
								SELECT add_numbers(x, y) * 2;
							$$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT calculate_double_sum(1, 2);`,
					Expected: []sql.Row{{6}},
				},
			},
		},
		{
			Name: "function returning multiple rows",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `CREATE FUNCTION gen(a int) RETURNS SETOF INT LANGUAGE SQL AS $$ SELECT generate_series(1, a) $$ STABLE;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM gen(3);`,
					Expected: []sql.Row{{1}, {2}, {3}},
				},
			},
		},
		{
			Name: "function with create or replace view",
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION public.sp_build_view_bathymetry_layer() RETURNS void
							LANGUAGE sql
							AS $$
								CREATE OR REPLACE VIEW public.view_bathymetry_layer AS
								SELECT 1;
							$$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT public.sp_build_view_bathymetry_layer()`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * from view_bathymetry_layer`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT public.sp_build_view_bathymetry_layer()`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * from view_bathymetry_layer`,
					Expected: []sql.Row{{1}},
				},
			},
		},
		{
			Name: "function with update ... returning",
			SetUpScript: []string{
				`CREATE TYPE public.tax_job_state AS ENUM (
					'sched',
					'busy',
					'final',
					'help'
				);`,
				`CREATE TABLE public.tax_job (
					id bigint NOT NULL,
					state public.tax_job_state NOT NULL,
					created timestamp NOT NULL,
					modified timestamp NOT NULL,
					scheduled timestamp,
					worker text,
					processor text,
					ext_id text,
					data jsonb,
					gross integer,
					notes text[],
					ops jsonb,
					CONSTRAINT tax_job_check CHECK ((NOT ((state = 'sched'::public.tax_job_state) AND (scheduled IS NULL)))),
					CONSTRAINT tax_job_check1 CHECK ((NOT ((state = 'busy'::public.tax_job_state) AND (worker IS NULL))))
				);`,
				`INSERT INTO tax_job (id, state, created, modified, scheduled, worker, processor, ext_id, data) VALUES (1, 'sched', '2025-05-05 05:05:05', '2025-05-05 05:05:05', '2025-05-05 05:05:05', 'worker', 'processor', 'ext_id', NULL)`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION public.tax_job_take(arg_worker text) RETURNS SETOF public.tax_job
								LANGUAGE sql
								AS '
								UPDATE
									tax_job
								SET
									state = ''busy'',
									worker = arg_worker
								WHERE
									state = ''sched''
									AND scheduled <= CURRENT_TIMESTAMP
								RETURNING
									*;
							';`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT public.tax_job_take('worker')`,
					Expected: []sql.Row{{`(1,busy,"2025-05-05 05:05:05","2025-05-05 05:05:05","2025-05-05 05:05:05",worker,processor,ext_id,,,,)`}},
				},
				{
					Query: `INSERT INTO tax_job (id, state, created, modified, scheduled, worker, processor, ext_id, data) VALUES (2, 'sched', '2025-05-05 05:05:06', '2025-05-05 05:05:06', '2025-05-05 05:05:06', 'worker', 'processor', 'ext_id', NULL), (3, 'sched', '2025-05-05 05:05:07', '2025-05-05 05:05:07', '2025-05-05 05:05:07', 'worker', 'processor', 'ext_id', NULL)`,
				},
				{
					Query: `SELECT public.tax_job_take('worker')`,
					Expected: []sql.Row{
						{`(2,busy,"2025-05-05 05:05:06","2025-05-05 05:05:06","2025-05-05 05:05:06",worker,processor,ext_id,,,,)`},
						{`(3,busy,"2025-05-05 05:05:07","2025-05-05 05:05:07","2025-05-05 05:05:07",worker,processor,ext_id,,,,)`},
					},
				},
			},
		},
		{
			Name: "function with delete",
			SetUpScript: []string{
				`CREATE TABLE test (id bigint NOT NULL, state text NOT NULL);`,
				`INSERT INTO test VALUES (1, 'sched'), (2, 'busy');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION d(w text) RETURNS bigint
								LANGUAGE sql
								AS '
								DELETE FROM test
								WHERE
									state = w
								RETURNING
									id;
							';`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM test;`,
					Expected: []sql.Row{{1, "sched"}, {2, "busy"}},
				},
				{
					Query:    `SELECT d('sched');`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT * FROM test;`,
					Expected: []sql.Row{{2, "busy"}},
				},
			},
		},
		{
			Name: "multiple statements in function",
			SetUpScript: []string{
				`CREATE TABLE test (id int);`,
				`INSERT INTO test VALUES (1), (2), (3);`,
				`CREATE VIEW test1 AS SELECT * FROM test WHERE id = 1;`,
				`CREATE VIEW test2 AS SELECT * FROM test WHERE id = 2;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION drop_views() RETURNS void
								LANGUAGE sql
								AS $$
							DROP VIEW test1;
							DROP VIEW test2;
							$$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM test1`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT * FROM test2`,
					Expected: []sql.Row{{2}},
				},
				{
					Query:    `SELECT drop_views();`,
					Expected: []sql.Row{{nil}},
				},
				{
					Query:       `SELECT * FROM test1`,
					ExpectedErr: `not found`,
				},
				{
					Query:       `SELECT * FROM test2`,
					ExpectedErr: `not found`,
				},
			},
		},
		{
			Name: "function with default expression in parameter",
			SetUpScript: []string{
				`CREATE TABLE cp_test (a int, b text);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE OR REPLACE FUNCTION dfunc(e int, d text, f int default 100)
							 RETURNS int LANGUAGE SQL
							AS $$
								INSERT INTO cp_test VALUES(e+f, d);
								SELECT a FROM cp_test WHERE b = d;
							$$;`,
					Expected: []sql.Row{},
				},
				{
					Query: `CREATE OR REPLACE FUNCTION dfunc(e int, f int default 100)
							 RETURNS int LANGUAGE SQL
							AS $$
								INSERT INTO cp_test VALUES(e+f, 'seconddfunc');
								SELECT a FROM cp_test WHERE b = 'seconddfunc';
							$$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM dfunc(10, 'Hello', 20);`,
					Expected: []sql.Row{{30}},
				},
				{
					Query:    `SELECT * FROM cp_test`,
					Expected: []sql.Row{{30, "Hello"}},
				},
				{
					Query:    `SELECT * FROM dfunc(50, 'Bye');`,
					Expected: []sql.Row{{150}},
				},
				{
					Query:    `SELECT * FROM cp_test`,
					Expected: []sql.Row{{30, "Hello"}, {150, "Bye"}},
				},
				{
					Query:    `SELECT dfunc(2, 'After');`,
					Expected: []sql.Row{{102}},
				},
				{
					Query:    `SELECT * FROM cp_test`,
					Expected: []sql.Row{{30, "Hello"}, {150, "Bye"}, {102, "After"}},
				},
				{
					Query: `CREATE OR REPLACE FUNCTION dfunc(e int, f text default '100')
							 RETURNS int LANGUAGE SQL
							AS $$
								INSERT INTO cp_test VALUES(e, f);
								SELECT a FROM cp_test WHERE b = f;
							$$;`,
					Expected: []sql.Row{},
				},
				{
					// TODO: the error message should be "function dfunc(integer) is not unique"
					Query:       `SELECT dfunc(50);`,
					ExpectedErr: `does not exist`,
				},
			},
		},
		{
			Name:        "use sql statements in BEGIN ATOMIC ... END in sql_body",
			SetUpScript: []string{},
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION match_default() RETURNS jsonb
            LANGUAGE sql
            BEGIN ATOMIC 
				SELECT jsonb_build_object('k', 6, 'm', 2048, 'include_original', true, 'tokenizer', json_build_object('kind', 'ngram', 'token_length', 3), 'token_filters', json_build_array(json_build_object('kind', 'downcase'))) AS jsonb_build_object; 
			END;`,
					Expected: []sql.Row{},
				},
				{
					Skip:     true, // TODO support json_build_object() function
					Query:    `SELECT public.match_default();`,
					Expected: []sql.Row{{`{"k": 6, "m": 2048, "tokenizer": {"kind": "ngram", "token_length": 3}, "token_filters": [{"kind": "downcase"}], "include_original": true}`}},
				},
				{
					Query: `CREATE FUNCTION select1() RETURNS int
            LANGUAGE sql
            BEGIN ATOMIC 
				SELECT 1; 
			END;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT select1();`,
					Expected: []sql.Row{{1}},
				},
			},
		},
		{
			Name:        "use RETURN in BEGIN ATOMIC ... END in sql_body",
			SetUpScript: []string{},
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION return1() RETURNS text
            LANGUAGE sql
            BEGIN ATOMIC 
				RETURN 1::text || 'one'; 
			END;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT return1();`,
					Expected: []sql.Row{{"1one"}},
				},
			},
		},
	})
}
