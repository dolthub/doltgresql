// Copyright 2024 Dolthub, Inc.
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

package regression

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestCreateMisc(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_create_misc)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_create_misc,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TABLE a_star (
	class		char,
	a 			int4
);`,
			},
			{
				Statement: `CREATE TABLE b_star (
	b 			text
) INHERITS (a_star);`,
			},
			{
				Statement: `CREATE TABLE c_star (
	c 			name
) INHERITS (a_star);`,
			},
			{
				Statement: `CREATE TABLE d_star (
	d 			float8
) INHERITS (b_star, c_star);`,
			},
			{
				Statement: `CREATE TABLE e_star (
	e 			int2
) INHERITS (c_star);`,
			},
			{
				Statement: `CREATE TABLE f_star (
	f 			polygon
) INHERITS (e_star);`,
			},
			{
				Statement: `INSERT INTO a_star (class, a) VALUES ('a', 1);`,
			},
			{
				Statement: `INSERT INTO a_star (class, a) VALUES ('a', 2);`,
			},
			{
				Statement: `INSERT INTO a_star (class) VALUES ('a');`,
			},
			{
				Statement: `INSERT INTO b_star (class, a, b) VALUES ('b', 3, 'mumble'::text);`,
			},
			{
				Statement: `INSERT INTO b_star (class, a) VALUES ('b', 4);`,
			},
			{
				Statement: `INSERT INTO b_star (class, b) VALUES ('b', 'bumble'::text);`,
			},
			{
				Statement: `INSERT INTO b_star (class) VALUES ('b');`,
			},
			{
				Statement: `INSERT INTO c_star (class, a, c) VALUES ('c', 5, 'hi mom'::name);`,
			},
			{
				Statement: `INSERT INTO c_star (class, a) VALUES ('c', 6);`,
			},
			{
				Statement: `INSERT INTO c_star (class, c) VALUES ('c', 'hi paul'::name);`,
			},
			{
				Statement: `INSERT INTO c_star (class) VALUES ('c');`,
			},
			{
				Statement: `INSERT INTO d_star (class, a, b, c, d)
   VALUES ('d', 7, 'grumble'::text, 'hi sunita'::name, '0.0'::float8);`,
			},
			{
				Statement: `INSERT INTO d_star (class, a, b, c)
   VALUES ('d', 8, 'stumble'::text, 'hi koko'::name);`,
			},
			{
				Statement: `INSERT INTO d_star (class, a, b, d)
   VALUES ('d', 9, 'rumble'::text, '1.1'::float8);`,
			},
			{
				Statement: `INSERT INTO d_star (class, a, c, d)
   VALUES ('d', 10, 'hi kristin'::name, '10.01'::float8);`,
			},
			{
				Statement: `INSERT INTO d_star (class, b, c, d)
   VALUES ('d', 'crumble'::text, 'hi boris'::name, '100.001'::float8);`,
			},
			{
				Statement: `INSERT INTO d_star (class, a, b)
   VALUES ('d', 11, 'fumble'::text);`,
			},
			{
				Statement: `INSERT INTO d_star (class, a, c)
   VALUES ('d', 12, 'hi avi'::name);`,
			},
			{
				Statement: `INSERT INTO d_star (class, a, d)
   VALUES ('d', 13, '1000.0001'::float8);`,
			},
			{
				Statement: `INSERT INTO d_star (class, b, c)
   VALUES ('d', 'tumble'::text, 'hi andrew'::name);`,
			},
			{
				Statement: `INSERT INTO d_star (class, b, d)
   VALUES ('d', 'humble'::text, '10000.00001'::float8);`,
			},
			{
				Statement: `INSERT INTO d_star (class, c, d)
   VALUES ('d', 'hi ginger'::name, '100000.000001'::float8);`,
			},
			{
				Statement: `INSERT INTO d_star (class, a) VALUES ('d', 14);`,
			},
			{
				Statement: `INSERT INTO d_star (class, b) VALUES ('d', 'jumble'::text);`,
			},
			{
				Statement: `INSERT INTO d_star (class, c) VALUES ('d', 'hi jolly'::name);`,
			},
			{
				Statement: `INSERT INTO d_star (class, d) VALUES ('d', '1000000.0000001'::float8);`,
			},
			{
				Statement: `INSERT INTO d_star (class) VALUES ('d');`,
			},
			{
				Statement: `INSERT INTO e_star (class, a, c, e)
   VALUES ('e', 15, 'hi carol'::name, '-1'::int2);`,
			},
			{
				Statement: `INSERT INTO e_star (class, a, c)
   VALUES ('e', 16, 'hi bob'::name);`,
			},
			{
				Statement: `INSERT INTO e_star (class, a, e)
   VALUES ('e', 17, '-2'::int2);`,
			},
			{
				Statement: `INSERT INTO e_star (class, c, e)
   VALUES ('e', 'hi michelle'::name, '-3'::int2);`,
			},
			{
				Statement: `INSERT INTO e_star (class, a)
   VALUES ('e', 18);`,
			},
			{
				Statement: `INSERT INTO e_star (class, c)
   VALUES ('e', 'hi elisa'::name);`,
			},
			{
				Statement: `INSERT INTO e_star (class, e)
   VALUES ('e', '-4'::int2);`,
			},
			{
				Statement: `INSERT INTO f_star (class, a, c, e, f)
   VALUES ('f', 19, 'hi claire'::name, '-5'::int2, '(1,3),(2,4)'::polygon);`,
			},
			{
				Statement: `INSERT INTO f_star (class, a, c, e)
   VALUES ('f', 20, 'hi mike'::name, '-6'::int2);`,
			},
			{
				Statement: `INSERT INTO f_star (class, a, c, f)
   VALUES ('f', 21, 'hi marcel'::name, '(11,44),(22,55),(33,66)'::polygon);`,
			},
			{
				Statement: `INSERT INTO f_star (class, a, e, f)
   VALUES ('f', 22, '-7'::int2, '(111,555),(222,666),(333,777),(444,888)'::polygon);`,
			},
			{
				Statement: `INSERT INTO f_star (class, c, e, f)
   VALUES ('f', 'hi keith'::name, '-8'::int2,
	   '(1111,3333),(2222,4444)'::polygon);`,
			},
			{
				Statement: `INSERT INTO f_star (class, a, c)
   VALUES ('f', 24, 'hi marc'::name);`,
			},
			{
				Statement: `INSERT INTO f_star (class, a, e)
   VALUES ('f', 25, '-9'::int2);`,
			},
			{
				Statement: `INSERT INTO f_star (class, a, f)
   VALUES ('f', 26, '(11111,33333),(22222,44444)'::polygon);`,
			},
			{
				Statement: `INSERT INTO f_star (class, c, e)
   VALUES ('f', 'hi allison'::name, '-10'::int2);`,
			},
			{
				Statement: `INSERT INTO f_star (class, c, f)
   VALUES ('f', 'hi jeff'::name,
           '(111111,333333),(222222,444444)'::polygon);`,
			},
			{
				Statement: `INSERT INTO f_star (class, e, f)
   VALUES ('f', '-11'::int2, '(1111111,3333333),(2222222,4444444)'::polygon);`,
			},
			{
				Statement: `INSERT INTO f_star (class, a) VALUES ('f', 27);`,
			},
			{
				Statement: `INSERT INTO f_star (class, c) VALUES ('f', 'hi carl'::name);`,
			},
			{
				Statement: `INSERT INTO f_star (class, e) VALUES ('f', '-12'::int2);`,
			},
			{
				Statement: `INSERT INTO f_star (class, f)
   VALUES ('f', '(11111111,33333333),(22222222,44444444)'::polygon);`,
			},
			{
				Statement: `INSERT INTO f_star (class) VALUES ('f');`,
			},
			{
				Statement: `ANALYZE a_star;`,
			},
			{
				Statement: `ANALYZE b_star;`,
			},
			{
				Statement: `ANALYZE c_star;`,
			},
			{
				Statement: `ANALYZE d_star;`,
			},
			{
				Statement: `ANALYZE e_star;`,
			},
			{
				Statement: `ANALYZE f_star;`,
			},
			{
				Statement: `SELECT * FROM a_star*;`,
				Results:   []sql.Row{{`a`, 1}, {`a`, 2}, {`a`, ``}, {`b`, 3}, {`b`, 4}, {`b`, ``}, {`b`, ``}, {`c`, 5}, {`c`, 6}, {`c`, ``}, {`c`, ``}, {`d`, 7}, {`d`, 8}, {`d`, 9}, {`d`, 10}, {`d`, ``}, {`d`, 11}, {`d`, 12}, {`d`, 13}, {`d`, ``}, {`d`, ``}, {`d`, ``}, {`d`, 14}, {`d`, ``}, {`d`, ``}, {`d`, ``}, {`d`, ``}, {`e`, 15}, {`e`, 16}, {`e`, 17}, {`e`, ``}, {`e`, 18}, {`e`, ``}, {`e`, ``}, {false, 19}, {false, 20}, {false, 21}, {false, 22}, {false, ``}, {false, 24}, {false, 25}, {false, 26}, {false, ``}, {false, ``}, {false, ``}, {false, 27}, {false, ``}, {false, ``}, {false, ``}, {false, ``}},
			},
			{
				Statement: `SELECT *
   FROM b_star* x
   WHERE x.b = text 'bumble' or x.a < 3;`,
				Results: []sql.Row{{`b`, ``, `bumble`}},
			},
			{
				Statement: `SELECT class, a
   FROM c_star* x
   WHERE x.c ~ text 'hi';`,
				Results: []sql.Row{{`c`, 5}, {`c`, ``}, {`d`, 7}, {`d`, 8}, {`d`, 10}, {`d`, ``}, {`d`, 12}, {`d`, ``}, {`d`, ``}, {`d`, ``}, {`e`, 15}, {`e`, 16}, {`e`, ``}, {`e`, ``}, {false, 19}, {false, 20}, {false, 21}, {false, ``}, {false, 24}, {false, ``}, {false, ``}, {false, ``}},
			},
			{
				Statement: `SELECT class, b, c
   FROM d_star* x
   WHERE x.a < 100;`,
				Results: []sql.Row{{`d`, `grumble`, `hi sunita`}, {`d`, `stumble`, `hi koko`}, {`d`, `rumble`, ``}, {`d`, ``, `hi kristin`}, {`d`, `fumble`, ``}, {`d`, ``, `hi avi`}, {`d`, ``, ``}, {`d`, ``, ``}},
			},
			{
				Statement: `SELECT class, c FROM e_star* x WHERE x.c NOTNULL;`,
				Results:   []sql.Row{{`e`, `hi carol`}, {`e`, `hi bob`}, {`e`, `hi michelle`}, {`e`, `hi elisa`}, {false, `hi claire`}, {false, `hi mike`}, {false, `hi marcel`}, {false, `hi keith`}, {false, `hi marc`}, {false, `hi allison`}, {false, `hi jeff`}, {false, `hi carl`}},
			},
			{
				Statement: `SELECT * FROM f_star* x WHERE x.c ISNULL;`,
				Results:   []sql.Row{{false, 22, ``, -7, `((111,555),(222,666),(333,777),(444,888))`}, {false, 25, ``, -9, ``}, {false, 26, ``, ``, `((11111,33333),(22222,44444))`}, {false, ``, ``, -11, `((1111111,3333333),(2222222,4444444))`}, {false, 27, ``, ``, ``}, {false, ``, ``, -12, ``}, {false, ``, ``, ``, `((11111111,33333333),(22222222,44444444))`}, {false, ``, ``, ``, ``}},
			},
			{
				Statement: `SELECT sum(a) FROM a_star*;`,
				Results:   []sql.Row{{355}},
			},
			{
				Statement: `SELECT class, sum(a) FROM a_star* GROUP BY class ORDER BY class;`,
				Results:   []sql.Row{{`a`, 3}, {`b`, 7}, {`c`, 11}, {`d`, 84}, {`e`, 66}, {false, 184}},
			},
			{
				Statement: `ALTER TABLE f_star RENAME COLUMN f TO ff;`,
			},
			{
				Statement: `ALTER TABLE e_star* RENAME COLUMN e TO ee;`,
			},
			{
				Statement: `ALTER TABLE d_star* RENAME COLUMN d TO dd;`,
			},
			{
				Statement: `ALTER TABLE c_star* RENAME COLUMN c TO cc;`,
			},
			{
				Statement: `ALTER TABLE b_star* RENAME COLUMN b TO bb;`,
			},
			{
				Statement: `ALTER TABLE a_star* RENAME COLUMN a TO aa;`,
			},
			{
				Statement: `SELECT class, aa
   FROM a_star* x
   WHERE aa ISNULL;`,
				Results: []sql.Row{{`a`, ``}, {`b`, ``}, {`b`, ``}, {`c`, ``}, {`c`, ``}, {`d`, ``}, {`d`, ``}, {`d`, ``}, {`d`, ``}, {`d`, ``}, {`d`, ``}, {`d`, ``}, {`d`, ``}, {`e`, ``}, {`e`, ``}, {`e`, ``}, {false, ``}, {false, ``}, {false, ``}, {false, ``}, {false, ``}, {false, ``}, {false, ``}, {false, ``}},
			},
			{
				Statement: `ALTER TABLE a_star RENAME COLUMN aa TO foo;`,
			},
			{
				Statement: `SELECT class, foo
   FROM a_star* x
   WHERE x.foo >= 2;`,
				Results: []sql.Row{{`a`, 2}, {`b`, 3}, {`b`, 4}, {`c`, 5}, {`c`, 6}, {`d`, 7}, {`d`, 8}, {`d`, 9}, {`d`, 10}, {`d`, 11}, {`d`, 12}, {`d`, 13}, {`d`, 14}, {`e`, 15}, {`e`, 16}, {`e`, 17}, {`e`, 18}, {false, 19}, {false, 20}, {false, 21}, {false, 22}, {false, 24}, {false, 25}, {false, 26}, {false, 27}},
			},
			{
				Statement: `ALTER TABLE a_star RENAME COLUMN foo TO aa;`,
			},
			{
				Statement: `SELECT *
   from a_star*
   WHERE aa < 1000;`,
				Results: []sql.Row{{`a`, 1}, {`a`, 2}, {`b`, 3}, {`b`, 4}, {`c`, 5}, {`c`, 6}, {`d`, 7}, {`d`, 8}, {`d`, 9}, {`d`, 10}, {`d`, 11}, {`d`, 12}, {`d`, 13}, {`d`, 14}, {`e`, 15}, {`e`, 16}, {`e`, 17}, {`e`, 18}, {false, 19}, {false, 20}, {false, 21}, {false, 22}, {false, 24}, {false, 25}, {false, 26}, {false, 27}},
			},
			{
				Statement: `ALTER TABLE f_star ADD COLUMN f int4;`,
			},
			{
				Statement: `UPDATE f_star SET f = 10;`,
			},
			{
				Statement: `ALTER TABLE e_star* ADD COLUMN e int4;`,
			},
			{
				Statement: `SELECT * FROM e_star*;`,
				Results:   []sql.Row{{`e`, 15, `hi carol`, -1, ``}, {`e`, 16, `hi bob`, ``, ``}, {`e`, 17, ``, -2, ``}, {`e`, ``, `hi michelle`, -3, ``}, {`e`, 18, ``, ``, ``}, {`e`, ``, `hi elisa`, ``, ``}, {`e`, ``, ``, -4, ``}, {false, 19, `hi claire`, -5, ``}, {false, 20, `hi mike`, -6, ``}, {false, 21, `hi marcel`, ``, ``}, {false, 22, ``, -7, ``}, {false, ``, `hi keith`, -8, ``}, {false, 24, `hi marc`, ``, ``}, {false, 25, ``, -9, ``}, {false, 26, ``, ``, ``}, {false, ``, `hi allison`, -10, ``}, {false, ``, `hi jeff`, ``, ``}, {false, ``, ``, -11, ``}, {false, 27, ``, ``, ``}, {false, ``, `hi carl`, ``, ``}, {false, ``, ``, -12, ``}, {false, ``, ``, ``, ``}, {false, ``, ``, ``, ``}},
			},
			{
				Statement: `ALTER TABLE a_star* ADD COLUMN a text;`,
			},
			{
				Statement: `SELECT relname, reltoastrelid <> 0 AS has_toast_table
   FROM pg_class
   WHERE oid::regclass IN ('a_star', 'c_star')
   ORDER BY 1;`,
				Results: []sql.Row{{`a_star`, true}, {`c_star`, true}},
			},
			{
				Statement: `SELECT class, aa, a FROM a_star*;`,
				Results:   []sql.Row{{`a`, 1, ``}, {`a`, 2, ``}, {`a`, ``, ``}, {`b`, 3, ``}, {`b`, 4, ``}, {`b`, ``, ``}, {`b`, ``, ``}, {`c`, 5, ``}, {`c`, 6, ``}, {`c`, ``, ``}, {`c`, ``, ``}, {`d`, 7, ``}, {`d`, 8, ``}, {`d`, 9, ``}, {`d`, 10, ``}, {`d`, ``, ``}, {`d`, 11, ``}, {`d`, 12, ``}, {`d`, 13, ``}, {`d`, ``, ``}, {`d`, ``, ``}, {`d`, ``, ``}, {`d`, 14, ``}, {`d`, ``, ``}, {`d`, ``, ``}, {`d`, ``, ``}, {`d`, ``, ``}, {`e`, 15, ``}, {`e`, 16, ``}, {`e`, 17, ``}, {`e`, ``, ``}, {`e`, 18, ``}, {`e`, ``, ``}, {`e`, ``, ``}, {false, 19, ``}, {false, 20, ``}, {false, 21, ``}, {false, 22, ``}, {false, ``, ``}, {false, 24, ``}, {false, 25, ``}, {false, 26, ``}, {false, ``, ``}, {false, ``, ``}, {false, ``, ``}, {false, 27, ``}, {false, ``, ``}, {false, ``, ``}, {false, ``, ``}, {false, ``, ``}},
			},
		},
	})
}
