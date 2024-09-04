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
	"fmt"
	"testing"
)

func TestTestSetup(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_test_setup)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_test_setup,
		DependsOn:          nil,
		Statements: []RegressionFileStatement{
			{Statement: `SET synchronous_commit = on;`},
			{Statement: `GRANT ALL ON SCHEMA public TO public;`},
			{Statement: `CREATE TABLE CHAR_TBL(f1 char(4));`},
			{Statement: `INSERT INTO CHAR_TBL (f1) VALUES
  ('a'),
  ('ab'),
  ('abcd'),
  ('abcd    ');`},
			{Statement: `VACUUM CHAR_TBL;`},
			{Statement: `CREATE TABLE FLOAT8_TBL(f1 float8);`},
			{Statement: `INSERT INTO FLOAT8_TBL(f1) VALUES
  ('0.0'),
  ('-34.84'),
  ('-1004.30'),
  ('-1.2345678901234e+200'),
  ('-1.2345678901234e-200');`},
			{Statement: `VACUUM FLOAT8_TBL;`},
			{Statement: `CREATE TABLE INT2_TBL(f1 int2);`},
			{Statement: `INSERT INTO INT2_TBL(f1) VALUES
  ('0   '),
  ('  1234 '),
  ('    -1234'),
  ('32767'),  -- largest and smallest values
  ('-32767');`},
			{Statement: `VACUUM INT2_TBL;`},
			{Statement: `CREATE TABLE INT4_TBL(f1 int4);`},
			{Statement: `INSERT INTO INT4_TBL(f1) VALUES
  ('   0  '),
  ('123456     '),
  ('    -123456'),
  ('2147483647'),  -- largest and smallest values
  ('-2147483647');`},
			{Statement: `VACUUM INT4_TBL;`},
			{Statement: `CREATE TABLE INT8_TBL(q1 int8, q2 int8);`},
			{Statement: `INSERT INTO INT8_TBL VALUES
  ('  123   ','  456'),
  ('123   ','4567890123456789'),
  ('4567890123456789','123'),
  (+4567890123456789,'4567890123456789'),
  ('+4567890123456789','-4567890123456789');`},
			{Statement: `VACUUM INT8_TBL;`},
			{Statement: `CREATE TABLE POINT_TBL(f1 point);`},
			{Statement: `INSERT INTO POINT_TBL(f1) VALUES
  ('(0.0,0.0)'),
  ('(-10.0,0.0)'),
  ('(-3.0,4.0)'),
  ('(5.1, 34.5)'),
  ('(-5.0,-12.0)'),
  ('(1e-300,-1e-300)'),  -- To underflow
  ('(1e+300,Inf)'),  -- To overflow
  ('(Inf,1e+300)'),  -- Transposed
  (' ( Nan , NaN ) '),
  ('10.0,10.0');`},
			{Statement: `CREATE TABLE TEXT_TBL (f1 text);`},
			{Statement: `INSERT INTO TEXT_TBL VALUES
  ('doh!'),
  ('hi de ho neighbor');`},
			{Statement: `VACUUM TEXT_TBL;`},
			{Statement: `CREATE TABLE VARCHAR_TBL(f1 varchar(4));`},
			{Statement: `INSERT INTO VARCHAR_TBL (f1) VALUES
  ('a'),
  ('ab'),
  ('abcd'),
  ('abcd    ');`},
			{Statement: `VACUUM VARCHAR_TBL;`},
			{Statement: `CREATE TABLE onek (
	unique1		int4,
	unique2		int4,
	two			int4,
	four		int4,
	ten			int4,
	twenty		int4,
	hundred		int4,
	thousand	int4,
	twothousand	int4,
	fivethous	int4,
	tenthous	int4,
	odd			int4,
	even		int4,
	stringu1	name,
	stringu2	name,
	string4		name
);`},
			{Statement: fmt.Sprintf(`COPY onek FROM '%s';`, GetDataFolder().GetAbsolutePath("onek.data"))},
			{Statement: `VACUUM ANALYZE onek;`},
			{Statement: `CREATE TABLE onek2 AS SELECT * FROM onek;`},
			{Statement: `VACUUM ANALYZE onek2;`},
			{Statement: `CREATE TABLE tenk1 (
	unique1		int4,
	unique2		int4,
	two			int4,
	four		int4,
	ten			int4,
	twenty		int4,
	hundred		int4,
	thousand	int4,
	twothousand	int4,
	fivethous	int4,
	tenthous	int4,
	odd			int4,
	even		int4,
	stringu1	name,
	stringu2	name,
	string4		name
);`},
			{Statement: fmt.Sprintf(`COPY tenk1 FROM '%s';`, GetDataFolder().GetAbsolutePath("tenk.data"))},
			{Statement: `VACUUM ANALYZE tenk1;`},
			{Statement: `CREATE TABLE tenk2 AS SELECT * FROM tenk1;`},
			{Statement: `VACUUM ANALYZE tenk2;`},
			{Statement: `CREATE TABLE person (
	name 		text,
	age			int4,
	location 	point
);`},
			{Statement: fmt.Sprintf(`COPY person FROM '%s';`, GetDataFolder().GetAbsolutePath("person.data"))},
			{Statement: `VACUUM ANALYZE person;`},
			{Statement: `CREATE TABLE emp (
	salary 		int4,
	manager 	name
) INHERITS (person);`},
			{Statement: fmt.Sprintf(`COPY emp FROM '%s';`, GetDataFolder().GetAbsolutePath("emp.data"))},
			{Statement: `VACUUM ANALYZE emp;`},
			{Statement: `CREATE TABLE student (
	gpa 		float8
) INHERITS (person);`},
			{Statement: fmt.Sprintf(`COPY student FROM '%s';`, GetDataFolder().GetAbsolutePath("student.data"))},
			{Statement: `VACUUM ANALYZE student;`},
			{Statement: `CREATE TABLE stud_emp (
	percent 	int4
) INHERITS (emp, student);`},
			{Statement: fmt.Sprintf(`COPY stud_emp FROM '%s';`, GetDataFolder().GetAbsolutePath("stud_emp.data"))},
			{Statement: `VACUUM ANALYZE stud_emp;`},
			{Statement: `CREATE TABLE road (
	name		text,
	thepath 	path
);`},
			{Statement: fmt.Sprintf(`COPY road FROM '%s';`, GetDataFolder().GetAbsolutePath("streets.data"))},
			{Statement: `VACUUM ANALYZE road;`},
			{Statement: `CREATE TABLE ihighway () INHERITS (road);`},
			{Statement: `INSERT INTO ihighway
   SELECT *
   FROM ONLY road
   WHERE name ~ 'I- .*';`},
			{Statement: `VACUUM ANALYZE ihighway;`},
			{Statement: `CREATE TABLE shighway (
	surface		text
) INHERITS (road);`},
			{Statement: `INSERT INTO shighway
   SELECT *, 'asphalt'
   FROM ONLY road
   WHERE name ~ 'State Hwy.*';`},
			{Statement: `VACUUM ANALYZE shighway;`},
			{Statement: `create type stoplight as enum ('red', 'yellow', 'green');`},
			{Statement: `create type float8range as range (subtype = float8, subtype_diff = float8mi);`},
			{Statement: `create type textrange as range (subtype = text, collation = "C");`},
			{Statement: `CREATE FUNCTION binary_coercible(oid, oid)
    RETURNS bool
    AS 'regresslib', 'binary_coercible'
    LANGUAGE C STRICT STABLE PARALLEL SAFE;`}, /*TODO: 'regresslib' has the function implementations that we need to implement ourselves*/
			{Statement: `CREATE FUNCTION ttdummy ()
    RETURNS trigger
    AS 'regresslib'
    LANGUAGE C;`},
			{Statement: `CREATE FUNCTION get_columns_length(oid[])
    RETURNS int
    AS 'regresslib'
    LANGUAGE C STRICT STABLE PARALLEL SAFE;`},
			{Statement: `create function part_hashint4_noop(value int4, seed int8)
    returns int8 as $$
    select value + seed;
    $$ language sql strict immutable parallel safe;`},
			{Statement: `create operator class part_test_int4_ops for type int4 using hash as
    operator 1 =,
    function 2 part_hashint4_noop(int4, int8);`},
			{Statement: `create function part_hashtext_length(value text, seed int8)
    returns int8 as $$
    select length(coalesce(value, ''))::int8
    $$ language sql strict immutable parallel safe;`},
			{Statement: `create operator class part_test_text_ops for type text using hash as
    operator 1 =,
    function 2 part_hashtext_length(text, int8);`},
		},
	})
}
