// Copyright 2023 Dolthub, Inc.
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

func TestSmokeTests(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Simple statements",
			SetUpScript: []string{
				"CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "CREATE TABLE test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO test VALUES (1, 1), (2, 2);",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO test2 VALUES (3, 3), (4, 4);",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{1, 1},
						{2, 2},
					},
				},
				{
					Query: "SELECT * FROM test2;",
					Expected: []sql.Row{
						{3, 3},
						{4, 4},
					},
				},
				{
					Query: "SELECT test2.pk FROM test2;",
					Expected: []sql.Row{
						{3},
						{4},
					},
				},
				{
					Query: "SELECT * FROM test ORDER BY 1 LIMIT 1 OFFSET 1;",
					Expected: []sql.Row{
						{2, 2},
					},
				},
			},
		},
		{
			Name: "Insert statements",
			SetUpScript: []string{
				"CREATE TABLE test (pk INT8 PRIMARY KEY, v1 INT4, v2 INT2);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "INSERT INTO test VALUES (1, 2, 3);",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO test (v1, pk) VALUES (5, 4);",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO test (pk, v2) SELECT pk + 5, v2 + 10 FROM test WHERE v2 IS NOT NULL;",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{1, 2, 3},
						{4, 5, nil},
						{6, nil, 13},
					},
				},
			},
		},
		{
			Name: "Update statements",
			SetUpScript: []string{
				"CREATE TABLE test (pk INT8 PRIMARY KEY, v1 INT4, v2 INT2);",
				"INSERT INTO test VALUES (1, 2, 3), (4, 5, 6), (7, 8, 9);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "UPDATE test SET v2 = 10;",
					Expected: []sql.Row{},
				},
				{
					Query:    "UPDATE test SET v1 = pk + v2;",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{1, 11, 10},
						{4, 14, 10},
						{7, 17, 10},
					},
				},
				{
					Query:    "UPDATE test SET pk = subquery.val FROM (SELECT 22 as val) AS subquery WHERE pk >= 7;",
					Skip:     true, // FROM not yet supported
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test;",
					Skip:  true, // Above query doesn't run yet
					Expected: []sql.Row{
						{1, 11, 10},
						{4, 14, 10},
						{22, 17, 10},
					},
				},
			},
		},
		{
			Name: "Delete statements",
			SetUpScript: []string{
				"CREATE TABLE test (pk INT8 PRIMARY KEY, v1 INT4, v2 INT2);",
				"INSERT INTO test VALUES (1, 1, 1), (2, 3, 4), (5, 7, 9);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "DELETE FROM test WHERE v2 = 9;",
					Expected: []sql.Row{},
				},
				{
					Query:    "DELETE FROM test WHERE v1 = pk;",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{2, 3, 4},
					},
				},
			},
		},
		{
			Name: "Dolt Getting Started example", /* https://docs.dolthub.com/introduction/getting-started/database */
			SetUpScript: []string{
				"create table employees (id int, last_name varchar(255), first_name varchar(255), primary key(id));",
				"create table teams (id int, team_name varchar(255), primary key(id));",
				"create table employees_teams(team_id int, employee_id int, primary key(team_id, employee_id), foreign key (team_id) references teams(id), foreign key (employee_id) references employees(id));",
				"call dolt_add('teams', 'employees', 'employees_teams');",
				"call dolt_commit('-m', 'Created initial schema');",
				"insert into employees values (0, 'Sehn', 'Tim'), (1, 'Hendriks', 'Brian'), (2, 'Son','Aaron'), (3, 'Fitzgerald', 'Brian');",
				"insert into teams values (0, 'Engineering'), (1, 'Sales');",
				"insert into employees_teams(employee_id, team_id) values (0,0), (1,0), (2,0), (0,1), (3,1);",
				"call dolt_commit('-am', 'Populated tables with data');",
				"call dolt_checkout('-b','modifications');",
				"update employees SET first_name='Timothy' where first_name='Tim';",
				"insert INTO employees (id, first_name, last_name) values (4,'Daylon', 'Wilkins');",
				"insert into employees_teams(team_id, employee_id) values (0,4);",
				"delete from employees_teams where employee_id=0 and team_id=1;",
				"call dolt_commit('-am', 'Modifications on a branch')",
				"call dolt_checkout('modifications');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "select to_last_name, to_first_name, to_id, to_commit, from_last_name, from_first_name," +
						"from_id, from_commit, diff_type from dolt_diff('main', 'modifications', 'employees');",
					Expected: []sql.Row{
						{"Sehn", "Timothy", 0, "modifications", "Sehn", "Tim", 0, "main", "modified"},
						{"Wilkins", "Daylon", 4, "modifications", nil, nil, nil, "main", "added"},
					},
				},
			},
		},
		{
			Name: "Boolean results",
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT 1 IN (2);",
					Expected: []sql.Row{
						{"f"},
					},
				},
				{
					Query: "SELECT 2 IN (2);",
					Expected: []sql.Row{
						{"t"},
					},
				},
			},
		},
		{
			Name: "Commit and diff across branches",
			SetUpScript: []string{
				"CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
				"INSERT INTO test VALUES (1, 1), (2, 2);",
				"CALL DOLT_ADD('-A');",
				"CALL DOLT_COMMIT('-m', 'initial commit');",
				"CALL DOLT_BRANCH('other');",
				"UPDATE test SET v1 = 3;",
				"CALL DOLT_ADD('-A');",
				"CALL DOLT_COMMIT('-m', 'commit main');",
				"CALL DOLT_CHECKOUT('other');",
				"UPDATE test SET v1 = 4 WHERE pk = 2;",
				"CALL DOLT_ADD('-A');",
				"CALL DOLT_COMMIT('-m', 'commit other');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:            "CALL DOLT_CHECKOUT('main');",
					SkipResultsCheck: true,
				},
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{1, 3},
						{2, 3},
					},
				},
				{
					Query:            "CALL DOLT_CHECKOUT('other');",
					SkipResultsCheck: true,
				},
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{1, 1},
						{2, 4},
					},
				},
				{
					Query: "SELECT from_pk, to_pk, from_v1, to_v1 FROM dolt_diff_test;",
					Expected: []sql.Row{
						{2, 2, 2, 4},
						{nil, 1, nil, 1},
						{nil, 2, nil, 2},
					},
				},
			},
		},
		{
			Name: "ARRAY expression",
			SetUpScript: []string{
				"CREATE TABLE test1 (id INTEGER primary key, v1 BOOLEAN);",
				"INSERT INTO test1 VALUES (1, 'true'), (2, 'false');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT ARRAY[v1]::boolean[] FROM test1 ORDER BY id;",
					Expected: []sql.Row{
						{"{t}"},
						{"{f}"},
					},
				},
				{
					Query: "SELECT ARRAY[v1] FROM test1 ORDER BY id;",
					Expected: []sql.Row{
						{"{t}"},
						{"{f}"},
					},
				},
				{
					Query: "SELECT ARRAY[v1, true, v1] FROM test1 ORDER BY id;",
					Expected: []sql.Row{
						{"{t,t,t}"},
						{"{f,t,f}"},
					},
				},
				{
					Query: "SELECT ARRAY[1::float8, 2::numeric];",
					Expected: []sql.Row{
						{"{1,2}"},
					},
				},
				{
					Query: "SELECT ARRAY[1::float8, NULL];",
					Expected: []sql.Row{
						{"{1,NULL}"},
					},
				},
				{
					Query: "SELECT ARRAY[1::int2, 2::int4, 3::int8]::varchar[];",
					Expected: []sql.Row{
						{"{1,2,3}"},
					},
				},
				{
					Query:       "SELECT ARRAY[1::int8]::int;",
					ExpectedErr: "cast from `bigint[]` to `integer` does not exist",
				},
				{
					Query:       "SELECT ARRAY[1::int8, 2::varchar];",
					ExpectedErr: "ARRAY types cannot be matched",
				},
			},
		},
		{
			Name: "Array casting",
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT '{true,false,true}'::boolean[];`,
					Expected: []sql.Row{
						{`{t,f,t}`},
					},
				},
				{
					Query: `SELECT '{"\\x68656c6c6f", "\\x776f726c64", "\\x6578616d706c65"}'::bytea[]::text[];`,
					Expected: []sql.Row{
						{`{"\x68656c6c6f","\x776f726c64","\x6578616d706c65"}`},
					},
				},
				{
					Query: `SELECT '{"abcd", "efgh", "ijkl"}'::char(3)[];`,
					Expected: []sql.Row{
						{`{abc,efg,ijk}`},
					},
				},
				{
					Query: `SELECT '{"2020-02-03", "2020-04-05", "2020-06-06"}'::date[];`,
					Expected: []sql.Row{
						{`{2020-02-03,2020-04-05,2020-06-06}`},
					},
				},
				{
					Query: `SELECT '{1.25,2.5,3.75}'::float4[];`,
					Expected: []sql.Row{
						{`{1.25,2.5,3.75}`},
					},
				},
				{
					Query: `SELECT '{4.25,5.5,6.75}'::float8[];`,
					Expected: []sql.Row{
						{`{4.25,5.5,6.75}`},
					},
				},
				{
					Query: `SELECT '{1,2,3}'::int2[];`,
					Expected: []sql.Row{
						{`{1,2,3}`},
					},
				},
				{
					Query: `SELECT '{4,5,6}'::int4[];`,
					Expected: []sql.Row{
						{`{4,5,6}`},
					},
				},
				{
					Query: `SELECT '{7,8,9}'::int8[];`,
					Expected: []sql.Row{
						{`{7,8,9}`},
					},
				},
				{
					Query: `SELECT '{"{\"a\":\"val1\"}", "{\"b\":\"value2\"}", "{\"c\": \"object_value3\"}"}'::json[];`,
					Expected: []sql.Row{
						{`{"{\"a\":\"val1\"}","{\"b\":\"value2\"}","{\"c\": \"object_value3\"}"}`},
					},
				},
				{
					Query: `SELECT '{"{\"d\":\"val1\"}", "{\"e\":\"value2\"}", "{\"f\": \"object_value3\"}"}'::jsonb[];`,
					Expected: []sql.Row{
						{`{"{\"d\": \"val1\"}","{\"e\": \"value2\"}","{\"f\": \"object_value3\"}"}`},
					},
				},
				{
					Query: `SELECT '{"the", "legendary", "formula"}'::name[];`,
					Expected: []sql.Row{
						{`{the,legendary,formula}`},
					},
				},
				{
					Query: `SELECT '{10.01,20.02,30.03}'::numeric[];`,
					Expected: []sql.Row{
						{`{10.01,20.02,30.03}`},
					},
				},
				{
					Query: `SELECT '{1,10,100}'::oid[];`,
					Expected: []sql.Row{
						{`{1,10,100}`},
					},
				},
				{
					Query: `SELECT '{"this", "is", "some", "text"}'::text[], '{text,without,quotes}'::text[], '{null,NULL,"NULL","quoted"}'::text[];`,
					Expected: []sql.Row{
						{`{this,is,some,text}`, `{text,without,quotes}`, `{NULL,NULL,"NULL",quoted}`},
					},
				},
				{
					Query: `SELECT '{"12:12:13", "14:14:15", "16:16:17"}'::time[];`,
					Expected: []sql.Row{
						{`{12:12:13,14:14:15,16:16:17}`},
					},
				},
				{
					Query: `SELECT '{"2020-02-03 12:13:14", "2020-04-05 15:16:17", "2020-06-06 18:19:20"}'::timestamp[];`,
					Expected: []sql.Row{
						{`{"2020-02-03 12:13:14","2020-04-05 15:16:17","2020-06-06 18:19:20"}`},
					},
				},
				{
					Query: `SELECT '{"3920fd79-7b53-437c-b647-d450b58b4532", "a594c217-4c63-4669-96ec-40eed180b7cf", "4367b70d-8d8b-4969-a1aa-bf59536455fb"}'::uuid[];`,
					Expected: []sql.Row{
						{`{3920fd79-7b53-437c-b647-d450b58b4532,a594c217-4c63-4669-96ec-40eed180b7cf,4367b70d-8d8b-4969-a1aa-bf59536455fb}`},
					},
				},
				{
					Query: `SELECT '{"somewhere", "over", "the", "rainbow"}'::varchar(5)[];`,
					Expected: []sql.Row{
						{`{somew,over,the,rainb}`},
					},
				},
				{
					Query: `SELECT '{1,2,3}'::xid[];`,
					Expected: []sql.Row{
						{`{1,2,3}`},
					},
				},
				{
					Query:       `SELECT '{"abc""","def"}'::text[];`,
					ExpectedErr: "malformed",
				},
				{
					Query:       `SELECT '{a,b,c'::text[];`,
					ExpectedErr: "malformed",
				},
				{
					Query:       `SELECT 'a,b,c}'::text[];`,
					ExpectedErr: "malformed",
				},
				{
					Query:       `SELECT '{"a,b,c}'::text[];`,
					ExpectedErr: "malformed",
				},
				{
					Query:       `SELECT '{a",b,c}'::text[];`,
					ExpectedErr: "malformed",
				},
				{
					Query:       `SELECT '{a,b,"c}'::text[];`,
					ExpectedErr: "malformed",
				},
				{
					Query:       `SELECT '{a,b,c"}'::text[];`,
					ExpectedErr: "malformed",
				},
			},
		},
		{
			Name: "BETWEEN",
			SetUpScript: []string{
				"CREATE TABLE test (v1 FLOAT8);",
				"INSERT INTO test VALUES (1), (3), (7);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM test WHERE v1 BETWEEN 1 AND 4 ORDER BY v1;",
					Expected: []sql.Row{
						{float64(1)},
						{float64(3)},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 BETWEEN 2 AND 4 ORDER BY v1;",
					Expected: []sql.Row{
						{float64(3)},
					},
				},
				{
					Query:    "SELECT * FROM test WHERE v1 BETWEEN 4 AND 2 ORDER BY v1;",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test WHERE v1 BETWEEN SYMMETRIC 1 AND 4 ORDER BY v1;",
					Expected: []sql.Row{
						{float64(1)},
						{float64(3)},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 BETWEEN SYMMETRIC 2 AND 4 ORDER BY v1;",
					Expected: []sql.Row{
						{float64(3)},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 BETWEEN SYMMETRIC 4 AND 2 ORDER BY v1;",
					Expected: []sql.Row{
						{float64(3)},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 NOT BETWEEN 1 AND 4 ORDER BY v1;",
					Expected: []sql.Row{
						{float64(7)},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 NOT BETWEEN 2 AND 4 ORDER BY v1;",
					Expected: []sql.Row{
						{float64(1)},
						{float64(7)},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 NOT BETWEEN 4 AND 2 ORDER BY v1;",
					Expected: []sql.Row{
						{float64(1)},
						{float64(3)},
						{float64(7)},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 NOT BETWEEN SYMMETRIC 1 AND 4 ORDER BY v1;",
					Expected: []sql.Row{
						{float64(7)},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 NOT BETWEEN SYMMETRIC 2 AND 4 ORDER BY v1;",
					Expected: []sql.Row{
						{float64(1)},
						{float64(7)},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 NOT BETWEEN SYMMETRIC 4 AND 2 ORDER BY v1;",
					Expected: []sql.Row{
						{float64(1)},
						{float64(7)},
					},
				},
			},
		},
		{
			Name: "IN",
			SetUpScript: []string{
				"CREATE TABLE test(v1 INT4, v2 INT4);",
				"INSERT INTO test VALUES (1, 1), (2, 2), (3, 3), (4, 4), (5, 5);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM test WHERE v1 IN (2, '3', 4) ORDER BY v1;",
					Expected: []sql.Row{
						{2, 2},
						{3, 3},
						{4, 4},
					},
				},
				{
					Query:    "CREATE INDEX v2_idx ON test(v2);",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test WHERE v2 IN (2, '3', 4) ORDER BY v1;",
					Expected: []sql.Row{
						{2, 2},
						{3, 3},
						{4, 4},
					},
				},
			},
		},
		{
			Name: "SUM",
			SetUpScript: []string{
				"CREATE TABLE test(pk SERIAL PRIMARY KEY, v1 INT4);",
				"INSERT INTO test (v1) VALUES (1), (2), (3), (4), (5);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT SUM(v1) FROM test WHERE v1 BETWEEN 3 AND 5;",
					Expected: []sql.Row{
						{12.0},
					},
				},
				{
					Query:    "CREATE INDEX v1_idx ON test(v1);",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT SUM(v1) FROM test WHERE v1 BETWEEN 3 AND 5;",
					Expected: []sql.Row{
						{12.0},
					},
				},
			},
		},
		{
			Name: "Empty statement",
			Assertions: []ScriptTestAssertion{
				{
					Query:    ";",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "Unsupported MySQL statements",
			Assertions: []ScriptTestAssertion{
				{
					Query:       "SHOW CREATE TABLE;",
					ExpectedErr: "syntax error",
				},
			},
		},
	})
}
