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

func TestMisc(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_misc)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_misc,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `\getenv abs_srcdir PG_ABS_SRCDIR
\getenv abs_builddir PG_ABS_BUILDDIR
\getenv libdir PG_LIBDIR
\getenv dlsuffix PG_DLSUFFIX
\set regresslib :libdir '/regress' :dlsuffix
CREATE FUNCTION overpaid(emp)
   RETURNS bool
   AS :'regresslib'
   LANGUAGE C STRICT;`,
			},
			{
				Statement: `CREATE FUNCTION reverse_name(name)
   RETURNS name
   AS :'regresslib'
   LANGUAGE C STRICT;`,
			},
			{
				Statement: `UPDATE onek
   SET unique1 = onek.unique1 + 1;`,
			},
			{
				Statement: `UPDATE onek
   SET unique1 = onek.unique1 - 1;`,
			},
			{
				Statement: `SELECT two, stringu1, ten, string4
   INTO TABLE tmp
   FROM onek;`,
			},
			{
				Statement: `UPDATE tmp
   SET stringu1 = reverse_name(onek.stringu1)
   FROM onek
   WHERE onek.stringu1 = 'JBAAAA' and
	  onek.stringu1 = tmp.stringu1;`,
			},
			{
				Statement: `UPDATE tmp
   SET stringu1 = reverse_name(onek2.stringu1)
   FROM onek2
   WHERE onek2.stringu1 = 'JCAAAA' and
	  onek2.stringu1 = tmp.stringu1;`,
			},
			{
				Statement: `DROP TABLE tmp;`,
			},
			{
				Statement: `\set filename :abs_builddir '/results/onek.data'
COPY onek TO :'filename';`,
			},
			{
				Statement: `CREATE TEMP TABLE onek_copy (LIKE onek);`,
			},
			{
				Statement: `COPY onek_copy FROM :'filename';`,
			},
			{
				Statement: `SELECT * FROM onek EXCEPT ALL SELECT * FROM onek_copy;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT * FROM onek_copy EXCEPT ALL SELECT * FROM onek;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `\set filename :abs_builddir '/results/stud_emp.data'
COPY BINARY stud_emp TO :'filename';`,
			},
			{
				Statement: `CREATE TEMP TABLE stud_emp_copy (LIKE stud_emp);`,
			},
			{
				Statement: `COPY BINARY stud_emp_copy FROM :'filename';`,
			},
			{
				Statement: `SELECT * FROM stud_emp_copy;`,
				Results:   []sql.Row{{`jeff`, 23, `(8,7.7)`, 600, `sharon`, 3.5, ``}, {`cim`, 30, `(10.5,4.7)`, 400, ``, 3.4, ``}, {`linda`, 19, `(0.9,6.1)`, 100, ``, 2.9, ``}},
			},
			{
				Statement: `CREATE TABLE hobbies_r (
	name		text,
	person 		text
);`,
			},
			{
				Statement: `CREATE TABLE equipment_r (
	name 		text,
	hobby		text
);`,
			},
			{
				Statement: `INSERT INTO hobbies_r (name, person)
   SELECT 'posthacking', p.name
   FROM person* p
   WHERE p.name = 'mike' or p.name = 'jeff';`,
			},
			{
				Statement: `INSERT INTO hobbies_r (name, person)
   SELECT 'basketball', p.name
   FROM person p
   WHERE p.name = 'joe' or p.name = 'sally';`,
			},
			{
				Statement: `INSERT INTO hobbies_r (name) VALUES ('skywalking');`,
			},
			{
				Statement: `INSERT INTO equipment_r (name, hobby) VALUES ('advil', 'posthacking');`,
			},
			{
				Statement: `INSERT INTO equipment_r (name, hobby) VALUES ('peet''s coffee', 'posthacking');`,
			},
			{
				Statement: `INSERT INTO equipment_r (name, hobby) VALUES ('hightops', 'basketball');`,
			},
			{
				Statement: `INSERT INTO equipment_r (name, hobby) VALUES ('guts', 'skywalking');`,
			},
			{
				Statement: `CREATE FUNCTION hobbies(person)
   RETURNS setof hobbies_r
   AS 'select * from hobbies_r where person = $1.name'
   LANGUAGE SQL;`,
			},
			{
				Statement: `CREATE FUNCTION hobby_construct(text, text)
   RETURNS hobbies_r
   AS 'select $1 as name, $2 as hobby'
   LANGUAGE SQL;`,
			},
			{
				Statement: `CREATE FUNCTION hobby_construct_named(name text, hobby text)
   RETURNS hobbies_r
   AS 'select name, hobby'
   LANGUAGE SQL;`,
			},
			{
				Statement: `CREATE FUNCTION hobbies_by_name(hobbies_r.name%TYPE)
   RETURNS hobbies_r.person%TYPE
   AS 'select person from hobbies_r where name = $1'
   LANGUAGE SQL;`,
			},
			{
				Statement: `CREATE FUNCTION equipment(hobbies_r)
   RETURNS setof equipment_r
   AS 'select * from equipment_r where hobby = $1.name'
   LANGUAGE SQL;`,
			},
			{
				Statement: `CREATE FUNCTION equipment_named(hobby hobbies_r)
   RETURNS setof equipment_r
   AS 'select * from equipment_r where equipment_r.hobby = equipment_named.hobby.name'
   LANGUAGE SQL;`,
			},
			{
				Statement: `CREATE FUNCTION equipment_named_ambiguous_1a(hobby hobbies_r)
   RETURNS setof equipment_r
   AS 'select * from equipment_r where hobby = equipment_named_ambiguous_1a.hobby.name'
   LANGUAGE SQL;`,
			},
			{
				Statement: `CREATE FUNCTION equipment_named_ambiguous_1b(hobby hobbies_r)
   RETURNS setof equipment_r
   AS 'select * from equipment_r where equipment_r.hobby = hobby.name'
   LANGUAGE SQL;`,
			},
			{
				Statement: `CREATE FUNCTION equipment_named_ambiguous_1c(hobby hobbies_r)
   RETURNS setof equipment_r
   AS 'select * from equipment_r where hobby = hobby.name'
   LANGUAGE SQL;`,
			},
			{
				Statement: `CREATE FUNCTION equipment_named_ambiguous_2a(hobby text)
   RETURNS setof equipment_r
   AS 'select * from equipment_r where hobby = equipment_named_ambiguous_2a.hobby'
   LANGUAGE SQL;`,
			},
			{
				Statement: `CREATE FUNCTION equipment_named_ambiguous_2b(hobby text)
   RETURNS setof equipment_r
   AS 'select * from equipment_r where equipment_r.hobby = hobby'
   LANGUAGE SQL;`,
			},
			{
				Statement: `SELECT p.name, name(p.hobbies) FROM ONLY person p;`,
				Results:   []sql.Row{{`mike`, `posthacking`}, {`joe`, `basketball`}, {`sally`, `basketball`}},
			},
			{
				Statement: `SELECT p.name, name(p.hobbies) FROM person* p;`,
				Results:   []sql.Row{{`mike`, `posthacking`}, {`joe`, `basketball`}, {`sally`, `basketball`}, {`jeff`, `posthacking`}},
			},
			{
				Statement: `SELECT DISTINCT hobbies_r.name, name(hobbies_r.equipment) FROM hobbies_r
  ORDER BY 1,2;`,
				Results: []sql.Row{{`basketball`, `hightops`}, {`posthacking`, `advil`}, {`posthacking`, `peet's coffee`}, {`skywalking`, `guts`}},
			},
			{
				Statement: `SELECT hobbies_r.name, (hobbies_r.equipment).name FROM hobbies_r;`,
				Results:   []sql.Row{{`posthacking`, `advil`}, {`posthacking`, `peet's coffee`}, {`posthacking`, `advil`}, {`posthacking`, `peet's coffee`}, {`basketball`, `hightops`}, {`basketball`, `hightops`}, {`skywalking`, `guts`}},
			},
			{
				Statement: `SELECT p.name, name(p.hobbies), name(equipment(p.hobbies)) FROM ONLY person p;`,
				Results:   []sql.Row{{`mike`, `posthacking`, `advil`}, {`mike`, `posthacking`, `peet's coffee`}, {`joe`, `basketball`, `hightops`}, {`sally`, `basketball`, `hightops`}},
			},
			{
				Statement: `SELECT p.name, name(p.hobbies), name(equipment(p.hobbies)) FROM person* p;`,
				Results:   []sql.Row{{`mike`, `posthacking`, `advil`}, {`mike`, `posthacking`, `peet's coffee`}, {`joe`, `basketball`, `hightops`}, {`sally`, `basketball`, `hightops`}, {`jeff`, `posthacking`, `advil`}, {`jeff`, `posthacking`, `peet's coffee`}},
			},
			{
				Statement: `SELECT name(equipment(p.hobbies)), p.name, name(p.hobbies) FROM ONLY person p;`,
				Results:   []sql.Row{{`advil`, `mike`, `posthacking`}, {`peet's coffee`, `mike`, `posthacking`}, {`hightops`, `joe`, `basketball`}, {`hightops`, `sally`, `basketball`}},
			},
			{
				Statement: `SELECT (p.hobbies).equipment.name, p.name, name(p.hobbies) FROM person* p;`,
				Results:   []sql.Row{{`advil`, `mike`, `posthacking`}, {`peet's coffee`, `mike`, `posthacking`}, {`hightops`, `joe`, `basketball`}, {`hightops`, `sally`, `basketball`}, {`advil`, `jeff`, `posthacking`}, {`peet's coffee`, `jeff`, `posthacking`}},
			},
			{
				Statement: `SELECT (p.hobbies).equipment.name, name(p.hobbies), p.name FROM ONLY person p;`,
				Results:   []sql.Row{{`advil`, `posthacking`, `mike`}, {`peet's coffee`, `posthacking`, `mike`}, {`hightops`, `basketball`, `joe`}, {`hightops`, `basketball`, `sally`}},
			},
			{
				Statement: `SELECT name(equipment(p.hobbies)), name(p.hobbies), p.name FROM person* p;`,
				Results:   []sql.Row{{`advil`, `posthacking`, `mike`}, {`peet's coffee`, `posthacking`, `mike`}, {`hightops`, `basketball`, `joe`}, {`hightops`, `basketball`, `sally`}, {`advil`, `posthacking`, `jeff`}, {`peet's coffee`, `posthacking`, `jeff`}},
			},
			{
				Statement: `SELECT name(equipment(hobby_construct(text 'skywalking', text 'mer')));`,
				Results:   []sql.Row{{`guts`}},
			},
			{
				Statement: `SELECT name(equipment(hobby_construct_named(text 'skywalking', text 'mer')));`,
				Results:   []sql.Row{{`guts`}},
			},
			{
				Statement: `SELECT name(equipment_named(hobby_construct_named(text 'skywalking', text 'mer')));`,
				Results:   []sql.Row{{`guts`}},
			},
			{
				Statement: `SELECT name(equipment_named_ambiguous_1a(hobby_construct_named(text 'skywalking', text 'mer')));`,
				Results:   []sql.Row{{`guts`}},
			},
			{
				Statement: `SELECT name(equipment_named_ambiguous_1b(hobby_construct_named(text 'skywalking', text 'mer')));`,
				Results:   []sql.Row{{`guts`}},
			},
			{
				Statement: `SELECT name(equipment_named_ambiguous_1c(hobby_construct_named(text 'skywalking', text 'mer')));`,
				Results:   []sql.Row{{`guts`}},
			},
			{
				Statement: `SELECT name(equipment_named_ambiguous_2a(text 'skywalking'));`,
				Results:   []sql.Row{{`guts`}},
			},
			{
				Statement: `SELECT name(equipment_named_ambiguous_2b(text 'skywalking'));`,
				Results:   []sql.Row{{`advil`}, {`peet's coffee`}, {`hightops`}, {`guts`}},
			},
			{
				Statement: `SELECT hobbies_by_name('basketball');`,
				Results:   []sql.Row{{`joe`}},
			},
			{
				Statement: `SELECT name, overpaid(emp.*) FROM emp;`,
				Results:   []sql.Row{{`sharon`, true}, {`sam`, true}, {`bill`, true}, {`jeff`, false}, {`cim`, false}, {`linda`, false}},
			},
			{
				Statement: `SELECT * FROM equipment(ROW('skywalking', 'mer'));`,
				Results:   []sql.Row{{`guts`, `skywalking`}},
			},
			{
				Statement: `SELECT name(equipment(ROW('skywalking', 'mer')));`,
				Results:   []sql.Row{{`guts`}},
			},
			{
				Statement: `SELECT *, name(equipment(h.*)) FROM hobbies_r h;`,
				Results:   []sql.Row{{`posthacking`, `mike`, `advil`}, {`posthacking`, `mike`, `peet's coffee`}, {`posthacking`, `jeff`, `advil`}, {`posthacking`, `jeff`, `peet's coffee`}, {`basketball`, `joe`, `hightops`}, {`basketball`, `sally`, `hightops`}, {`skywalking`, ``, `guts`}},
			},
			{
				Statement: `SELECT *, (equipment(CAST((h.*) AS hobbies_r))).name FROM hobbies_r h;`,
				Results:   []sql.Row{{`posthacking`, `mike`, `advil`}, {`posthacking`, `mike`, `peet's coffee`}, {`posthacking`, `jeff`, `advil`}, {`posthacking`, `jeff`, `peet's coffee`}, {`basketball`, `joe`, `hightops`}, {`basketball`, `sally`, `hightops`}, {`skywalking`, ``, `guts`}},
			},
		},
	})
}
