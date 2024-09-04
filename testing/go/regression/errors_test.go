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

func TestErrors(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_errors)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_errors,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `select 1;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `select;`,
			},
			{
				Statement: `(1 row)
select * from nonesuch;`,
				ErrorString: `relation "nonesuch" does not exist`,
			},
			{
				Statement:   `select nonesuch from pg_database;`,
				ErrorString: `column "nonesuch" does not exist`,
			},
			{
				Statement:   `select distinct from pg_database;`,
				ErrorString: `syntax error at or near "from"`,
			},
			{
				Statement:   `select * from pg_database where nonesuch = pg_database.datname;`,
				ErrorString: `column "nonesuch" does not exist`,
			},
			{
				Statement:   `select * from pg_database where pg_database.datname = nonesuch;`,
				ErrorString: `column "nonesuch" does not exist`,
			},
			{
				Statement:   `select distinct on (foobar) * from pg_database;`,
				ErrorString: `column "foobar" does not exist`,
			},
			{
				Statement:   `select null from pg_database group by datname for update;`,
				ErrorString: `FOR UPDATE is not allowed with GROUP BY clause`,
			},
			{
				Statement:   `select null from pg_database group by grouping sets (()) for update;`,
				ErrorString: `FOR UPDATE is not allowed with GROUP BY clause`,
			},
			{
				Statement:   `delete from;`,
				ErrorString: `syntax error at or near ";"`,
			},
			{
				Statement:   `delete from nonesuch;`,
				ErrorString: `relation "nonesuch" does not exist`,
			},
			{
				Statement:   `drop table;`,
				ErrorString: `syntax error at or near ";"`,
			},
			{
				Statement:   `drop table nonesuch;`,
				ErrorString: `table "nonesuch" does not exist`,
			},
			{
				Statement:   `alter table rename;`,
				ErrorString: `syntax error at or near ";"`,
			},
			{
				Statement:   `alter table nonesuch rename to newnonesuch;`,
				ErrorString: `relation "nonesuch" does not exist`,
			},
			{
				Statement:   `alter table nonesuch rename to stud_emp;`,
				ErrorString: `relation "nonesuch" does not exist`,
			},
			{
				Statement:   `alter table stud_emp rename to student;`,
				ErrorString: `relation "student" already exists`,
			},
			{
				Statement:   `alter table stud_emp rename to stud_emp;`,
				ErrorString: `relation "stud_emp" already exists`,
			},
			{
				Statement:   `alter table nonesuchrel rename column nonesuchatt to newnonesuchatt;`,
				ErrorString: `relation "nonesuchrel" does not exist`,
			},
			{
				Statement:   `alter table emp rename column nonesuchatt to newnonesuchatt;`,
				ErrorString: `column "nonesuchatt" does not exist`,
			},
			{
				Statement:   `alter table emp rename column salary to manager;`,
				ErrorString: `column "manager" of relation "stud_emp" already exists`,
			},
			{
				Statement:   `alter table emp rename column salary to ctid;`,
				ErrorString: `column name "ctid" conflicts with a system column name`,
			},
			{
				Statement: `abort;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `create aggregate newavg2 (sfunc = int4pl,
			  basetype = int4,
			  stype = int4,
			  finalfunc = int2um,
			  initcond = '0');`,
				ErrorString: `function int2um(integer) does not exist`,
			},
			{
				Statement: `create aggregate newcnt1 (sfunc = int4inc,
			  stype = int4,
			  initcond = '0');`,
				ErrorString: `aggregate input type must be specified`,
			},
			{
				Statement:   `drop index;`,
				ErrorString: `syntax error at or near ";"`,
			},
			{
				Statement:   `drop index 314159;`,
				ErrorString: `syntax error at or near "314159"`,
			},
			{
				Statement:   `drop index nonesuch;`,
				ErrorString: `index "nonesuch" does not exist`,
			},
			{
				Statement:   `drop aggregate;`,
				ErrorString: `syntax error at or near ";"`,
			},
			{
				Statement:   `drop aggregate newcnt1;`,
				ErrorString: `syntax error at or near ";"`,
			},
			{
				Statement:   `drop aggregate 314159 (int);`,
				ErrorString: `syntax error at or near "314159"`,
			},
			{
				Statement:   `drop aggregate newcnt (nonesuch);`,
				ErrorString: `type "nonesuch" does not exist`,
			},
			{
				Statement:   `drop aggregate nonesuch (int4);`,
				ErrorString: `aggregate nonesuch(integer) does not exist`,
			},
			{
				Statement:   `drop aggregate newcnt (float4);`,
				ErrorString: `aggregate newcnt(real) does not exist`,
			},
			{
				Statement:   `drop function ();`,
				ErrorString: `syntax error at or near "("`,
			},
			{
				Statement:   `drop function 314159();`,
				ErrorString: `syntax error at or near "314159"`,
			},
			{
				Statement:   `drop function nonesuch();`,
				ErrorString: `function nonesuch() does not exist`,
			},
			{
				Statement:   `drop type;`,
				ErrorString: `syntax error at or near ";"`,
			},
			{
				Statement:   `drop type 314159;`,
				ErrorString: `syntax error at or near "314159"`,
			},
			{
				Statement:   `drop type nonesuch;`,
				ErrorString: `type "nonesuch" does not exist`,
			},
			{
				Statement:   `drop operator;`,
				ErrorString: `syntax error at or near ";"`,
			},
			{
				Statement:   `drop operator equals;`,
				ErrorString: `syntax error at or near ";"`,
			},
			{
				Statement:   `drop operator ===;`,
				ErrorString: `syntax error at or near ";"`,
			},
			{
				Statement:   `drop operator int4, int4;`,
				ErrorString: `syntax error at or near ","`,
			},
			{
				Statement:   `drop operator (int4, int4);`,
				ErrorString: `syntax error at or near "("`,
			},
			{
				Statement:   `drop operator === ();`,
				ErrorString: `syntax error at or near ")"`,
			},
			{
				Statement:   `drop operator === (int4);`,
				ErrorString: `missing argument`,
			},
			{
				Statement:   `drop operator === (int4, int4);`,
				ErrorString: `operator does not exist: integer === integer`,
			},
			{
				Statement:   `drop operator = (nonesuch);`,
				ErrorString: `missing argument`,
			},
			{
				Statement:   `drop operator = ( , int4);`,
				ErrorString: `syntax error at or near ","`,
			},
			{
				Statement:   `drop operator = (nonesuch, int4);`,
				ErrorString: `type "nonesuch" does not exist`,
			},
			{
				Statement:   `drop operator = (int4, nonesuch);`,
				ErrorString: `type "nonesuch" does not exist`,
			},
			{
				Statement:   `drop operator = (int4, );`,
				ErrorString: `syntax error at or near ")"`,
			},
			{
				Statement:   `drop rule;`,
				ErrorString: `syntax error at or near ";"`,
			},
			{
				Statement:   `drop rule 314159;`,
				ErrorString: `syntax error at or near "314159"`,
			},
			{
				Statement:   `drop rule nonesuch on noplace;`,
				ErrorString: `relation "noplace" does not exist`,
			},
			{
				Statement:   `drop tuple rule nonesuch;`,
				ErrorString: `syntax error at or near "tuple"`,
			},
			{
				Statement:   `drop instance rule nonesuch on noplace;`,
				ErrorString: `syntax error at or near "instance"`,
			},
			{
				Statement:   `drop rewrite rule nonesuch;`,
				ErrorString: `syntax error at or near "rewrite"`,
			},
			{
				Statement:   `select 1/0;`,
				ErrorString: `division by zero`,
			},
			{
				Statement:   `select 1::int8/0;`,
				ErrorString: `division by zero`,
			},
			{
				Statement:   `select 1/0::int8;`,
				ErrorString: `division by zero`,
			},
			{
				Statement:   `select 1::int2/0;`,
				ErrorString: `division by zero`,
			},
			{
				Statement:   `select 1/0::int2;`,
				ErrorString: `division by zero`,
			},
			{
				Statement:   `select 1::numeric/0;`,
				ErrorString: `division by zero`,
			},
			{
				Statement:   `select 1/0::numeric;`,
				ErrorString: `division by zero`,
			},
			{
				Statement:   `select 1::float8/0;`,
				ErrorString: `division by zero`,
			},
			{
				Statement:   `select 1/0::float8;`,
				ErrorString: `division by zero`,
			},
			{
				Statement:   `select 1::float4/0;`,
				ErrorString: `division by zero`,
			},
			{
				Statement:   `select 1/0::float4;`,
				ErrorString: `division by zero`,
			},
			{
				Statement:   `xxx;`,
				ErrorString: `syntax error at or near "xxx"`,
			},
			{
				Statement:   `CREATE foo;`,
				ErrorString: `syntax error at or near "foo"`,
			},
			{
				Statement:   `CREATE TABLE ;`,
				ErrorString: `syntax error at or near ";"`,
			},
			{
				Statement: `CREATE TABLE
\g
ERROR:  syntax error at end of input
LINE 1: CREATE TABLE
                    ^
INSERT INTO foo VALUES(123) foo;`,
				ErrorString: `syntax error at or near "foo"`,
			},
			{
				Statement: `INSERT INTO 123
VALUES(123);`,
				ErrorString: `syntax error at or near "123"`,
			},
			{
				Statement: `INSERT INTO foo
VALUES(123) 123
;`,
				ErrorString: `syntax error at or near "123"`,
			},
			{
				Statement: `CREATE TABLE foo
  (id INT4 UNIQUE NOT NULL, id2 TEXT NOT NULL PRIMARY KEY,
	id3 INTEGER NOT NUL,
   id4 INT4 UNIQUE NOT NULL, id5 TEXT UNIQUE NOT NULL);`,
				ErrorString: `syntax error at or near "NUL"`,
			},
			{
				Statement: `CREATE TABLE foo(id INT4 UNIQUE NOT NULL, id2 TEXT NOT NULL PRIMARY KEY, id3 INTEGER NOT NUL,
id4 INT4 UNIQUE NOT NULL, id5 TEXT UNIQUE NOT NULL);`,
				ErrorString: `syntax error at or near "NUL"`,
			},
			{
				Statement: `CREATE TABLE foo(
id3 INTEGER NOT NUL, id4 INT4 UNIQUE NOT NULL, id5 TEXT UNIQUE NOT NULL, id INT4 UNIQUE NOT NULL, id2 TEXT NOT NULL PRIMARY KEY);`,
				ErrorString: `syntax error at or near "NUL"`,
			},
			{
				Statement:   `CREATE TABLE foo(id INT4 UNIQUE NOT NULL, id2 TEXT NOT NULL PRIMARY KEY, id3 INTEGER NOT NUL, id4 INT4 UNIQUE NOT NULL, id5 TEXT UNIQUE NOT NULL);`,
				ErrorString: `syntax error at or near "NUL"`,
			},
			{
				Statement: `CREATE
TEMPORARY
TABLE
foo(id INT4 UNIQUE NOT NULL, id2 TEXT NOT NULL PRIMARY KEY, id3 INTEGER NOT NUL,
id4 INT4
UNIQUE
NOT
NULL,
id5 TEXT
UNIQUE
NOT
NULL)
;`,
				ErrorString: `syntax error at or near "NUL"`,
			},
			{
				Statement: `CREATE
TEMPORARY
TABLE
foo(
id3 INTEGER NOT NUL, id4 INT4 UNIQUE NOT NULL, id5 TEXT UNIQUE NOT NULL, id INT4 UNIQUE NOT NULL, id2 TEXT NOT NULL PRIMARY KEY)
;`,
				ErrorString: `syntax error at or near "NUL"`,
			},
			{
				Statement: `CREATE
TEMPORARY
TABLE
foo
(id
INT4
UNIQUE NOT NULL, idx INT4 UNIQUE NOT NULL, idy INT4 UNIQUE NOT NULL, id2 TEXT NOT NULL PRIMARY KEY, id3 INTEGER NOT NUL, id4 INT4 UNIQUE NOT NULL, id5 TEXT UNIQUE NOT NULL,
idz INT4 UNIQUE NOT NULL,
idv INT4 UNIQUE NOT NULL);`,
				ErrorString: `syntax error at or near "NUL"`,
			},
			{
				Statement: `CREATE
TEMPORARY
TABLE
foo
(id
INT4
UNIQUE
NOT
NULL
,
idm
INT4
UNIQUE
NOT
NULL,
idx INT4 UNIQUE NOT NULL, idy INT4 UNIQUE NOT NULL, id2 TEXT NOT NULL PRIMARY KEY, id3 INTEGER NOT NUL, id4 INT4 UNIQUE NOT NULL, id5 TEXT UNIQUE NOT NULL,
idz INT4 UNIQUE NOT NULL,
idv
INT4
UNIQUE
NOT
NULL);`,
				ErrorString: `syntax error at or near "NUL"`,
			},
		},
	})
}
