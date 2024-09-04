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

func TestReturning(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_returning)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_returning,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TEMP TABLE foo (f1 serial, f2 text, f3 int default 42);`,
			},
			{
				Statement: `INSERT INTO foo (f2,f3)
  VALUES ('test', DEFAULT), ('More', 11), (upper('more'), 7+9)
  RETURNING *, f1+f3 AS sum;`,
				Results: []sql.Row{{1, `test`, 42, 43}, {2, `More`, 11, 13}, {3, `MORE`, 16, 19}},
			},
			{
				Statement: `SELECT * FROM foo;`,
				Results:   []sql.Row{{1, `test`, 42}, {2, `More`, 11}, {3, `MORE`, 16}},
			},
			{
				Statement: `UPDATE foo SET f2 = lower(f2), f3 = DEFAULT RETURNING foo.*, f1+f3 AS sum13;`,
				Results:   []sql.Row{{1, `test`, 42, 43}, {2, `more`, 42, 44}, {3, `more`, 42, 45}},
			},
			{
				Statement: `SELECT * FROM foo;`,
				Results:   []sql.Row{{1, `test`, 42}, {2, `more`, 42}, {3, `more`, 42}},
			},
			{
				Statement: `DELETE FROM foo WHERE f1 > 2 RETURNING f3, f2, f1, least(f1,f3);`,
				Results:   []sql.Row{{42, `more`, 3, 3}},
			},
			{
				Statement: `SELECT * FROM foo;`,
				Results:   []sql.Row{{1, `test`, 42}, {2, `more`, 42}},
			},
			{
				Statement: `INSERT INTO foo SELECT f1+10, f2, f3+99 FROM foo
  RETURNING *, f1+112 IN (SELECT q1 FROM int8_tbl) AS subplan,
    EXISTS(SELECT * FROM int4_tbl) AS initplan;`,
				Results: []sql.Row{{11, `test`, 141, true, true}, {12, `more`, 141, false, true}},
			},
			{
				Statement: `UPDATE foo SET f3 = f3 * 2
  WHERE f1 > 10
  RETURNING *, f1+112 IN (SELECT q1 FROM int8_tbl) AS subplan,
    EXISTS(SELECT * FROM int4_tbl) AS initplan;`,
				Results: []sql.Row{{11, `test`, 282, true, true}, {12, `more`, 282, false, true}},
			},
			{
				Statement: `DELETE FROM foo
  WHERE f1 > 10
  RETURNING *, f1+112 IN (SELECT q1 FROM int8_tbl) AS subplan,
    EXISTS(SELECT * FROM int4_tbl) AS initplan;`,
				Results: []sql.Row{{11, `test`, 282, true, true}, {12, `more`, 282, false, true}},
			},
			{
				Statement: `UPDATE foo SET f3 = f3*2
  FROM int4_tbl i
  WHERE foo.f1 + 123455 = i.f1
  RETURNING foo.*, i.f1 as "i.f1";`,
				Results: []sql.Row{{1, `test`, 84, 123456}},
			},
			{
				Statement: `SELECT * FROM foo;`,
				Results:   []sql.Row{{2, `more`, 42}, {1, `test`, 84}},
			},
			{
				Statement: `DELETE FROM foo
  USING int4_tbl i
  WHERE foo.f1 + 123455 = i.f1
  RETURNING foo.*, i.f1 as "i.f1";`,
				Results: []sql.Row{{1, `test`, 84, 123456}},
			},
			{
				Statement: `SELECT * FROM foo;`,
				Results:   []sql.Row{{2, `more`, 42}},
			},
			{
				Statement: `CREATE TEMP TABLE foochild (fc int) INHERITS (foo);`,
			},
			{
				Statement: `INSERT INTO foochild VALUES(123,'child',999,-123);`,
			},
			{
				Statement: `ALTER TABLE foo ADD COLUMN f4 int8 DEFAULT 99;`,
			},
			{
				Statement: `SELECT * FROM foo;`,
				Results:   []sql.Row{{2, `more`, 42, 99}, {123, `child`, 999, 99}},
			},
			{
				Statement: `SELECT * FROM foochild;`,
				Results:   []sql.Row{{123, `child`, 999, -123, 99}},
			},
			{
				Statement: `UPDATE foo SET f4 = f4 + f3 WHERE f4 = 99 RETURNING *;`,
				Results:   []sql.Row{{2, `more`, 42, 141}, {123, `child`, 999, 1098}},
			},
			{
				Statement: `SELECT * FROM foo;`,
				Results:   []sql.Row{{2, `more`, 42, 141}, {123, `child`, 999, 1098}},
			},
			{
				Statement: `SELECT * FROM foochild;`,
				Results:   []sql.Row{{123, `child`, 999, -123, 1098}},
			},
			{
				Statement: `UPDATE foo SET f3 = f3*2
  FROM int8_tbl i
  WHERE foo.f1 = i.q2
  RETURNING *;`,
				Results: []sql.Row{{123, `child`, 1998, 1098, 4567890123456789, 123}},
			},
			{
				Statement: `SELECT * FROM foo;`,
				Results:   []sql.Row{{2, `more`, 42, 141}, {123, `child`, 1998, 1098}},
			},
			{
				Statement: `SELECT * FROM foochild;`,
				Results:   []sql.Row{{123, `child`, 1998, -123, 1098}},
			},
			{
				Statement: `DELETE FROM foo
  USING int8_tbl i
  WHERE foo.f1 = i.q2
  RETURNING *;`,
				Results: []sql.Row{{123, `child`, 1998, 1098, 4567890123456789, 123}},
			},
			{
				Statement: `SELECT * FROM foo;`,
				Results:   []sql.Row{{2, `more`, 42, 141}},
			},
			{
				Statement: `SELECT * FROM foochild;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `DROP TABLE foochild;`,
			},
			{
				Statement: `CREATE TEMP VIEW voo AS SELECT f1, f2 FROM foo;`,
			},
			{
				Statement: `CREATE RULE voo_i AS ON INSERT TO voo DO INSTEAD
  INSERT INTO foo VALUES(new.*, 57);`,
			},
			{
				Statement: `INSERT INTO voo VALUES(11,'zit');`,
			},
			{
				Statement:   `INSERT INTO voo VALUES(12,'zoo') RETURNING *, f1*2;`,
				ErrorString: `cannot perform INSERT RETURNING on relation "voo"`,
			},
			{
				Statement: `CREATE OR REPLACE RULE voo_i AS ON INSERT TO voo DO INSTEAD
  INSERT INTO foo VALUES(new.*, 57) RETURNING *;`,
				ErrorString: `RETURNING list has too many entries`,
			},
			{
				Statement: `CREATE OR REPLACE RULE voo_i AS ON INSERT TO voo DO INSTEAD
  INSERT INTO foo VALUES(new.*, 57) RETURNING f1, f2;`,
			},
			{
				Statement: `INSERT INTO voo VALUES(13,'zit2');`,
			},
			{
				Statement: `INSERT INTO voo VALUES(14,'zoo2') RETURNING *;`,
				Results:   []sql.Row{{14, `zoo2`}},
			},
			{
				Statement: `SELECT * FROM foo;`,
				Results:   []sql.Row{{2, `more`, 42, 141}, {11, `zit`, 57, 99}, {13, `zit2`, 57, 99}, {14, `zoo2`, 57, 99}},
			},
			{
				Statement: `SELECT * FROM voo;`,
				Results:   []sql.Row{{2, `more`}, {11, `zit`}, {13, `zit2`}, {14, `zoo2`}},
			},
			{
				Statement: `CREATE OR REPLACE RULE voo_u AS ON UPDATE TO voo DO INSTEAD
  UPDATE foo SET f1 = new.f1, f2 = new.f2 WHERE f1 = old.f1
  RETURNING f1, f2;`,
			},
			{
				Statement: `update voo set f1 = f1 + 1 where f2 = 'zoo2';`,
			},
			{
				Statement: `update voo set f1 = f1 + 1 where f2 = 'zoo2' RETURNING *, f1*2;`,
				Results:   []sql.Row{{16, `zoo2`, 32}},
			},
			{
				Statement: `SELECT * FROM foo;`,
				Results:   []sql.Row{{2, `more`, 42, 141}, {11, `zit`, 57, 99}, {13, `zit2`, 57, 99}, {16, `zoo2`, 57, 99}},
			},
			{
				Statement: `SELECT * FROM voo;`,
				Results:   []sql.Row{{2, `more`}, {11, `zit`}, {13, `zit2`}, {16, `zoo2`}},
			},
			{
				Statement: `CREATE OR REPLACE RULE voo_d AS ON DELETE TO voo DO INSTEAD
  DELETE FROM foo WHERE f1 = old.f1
  RETURNING f1, f2;`,
			},
			{
				Statement: `DELETE FROM foo WHERE f1 = 13;`,
			},
			{
				Statement: `DELETE FROM foo WHERE f2 = 'zit' RETURNING *;`,
				Results:   []sql.Row{{11, `zit`, 57, 99}},
			},
			{
				Statement: `SELECT * FROM foo;`,
				Results:   []sql.Row{{2, `more`, 42, 141}, {16, `zoo2`, 57, 99}},
			},
			{
				Statement: `SELECT * FROM voo;`,
				Results:   []sql.Row{{2, `more`}, {16, `zoo2`}},
			},
			{
				Statement: `CREATE TEMP TABLE joinme (f2j text, other int);`,
			},
			{
				Statement: `INSERT INTO joinme VALUES('more', 12345);`,
			},
			{
				Statement: `INSERT INTO joinme VALUES('zoo2', 54321);`,
			},
			{
				Statement: `INSERT INTO joinme VALUES('other', 0);`,
			},
			{
				Statement: `CREATE TEMP VIEW joinview AS
  SELECT foo.*, other FROM foo JOIN joinme ON (f2 = f2j);`,
			},
			{
				Statement: `SELECT * FROM joinview;`,
				Results:   []sql.Row{{2, `more`, 42, 141, 12345}, {16, `zoo2`, 57, 99, 54321}},
			},
			{
				Statement: `CREATE RULE joinview_u AS ON UPDATE TO joinview DO INSTEAD
  UPDATE foo SET f1 = new.f1, f3 = new.f3
    FROM joinme WHERE f2 = f2j AND f2 = old.f2
    RETURNING foo.*, other;`,
			},
			{
				Statement: `UPDATE joinview SET f1 = f1 + 1 WHERE f3 = 57 RETURNING *, other + 1;`,
				Results:   []sql.Row{{17, `zoo2`, 57, 99, 54321, 54322}},
			},
			{
				Statement: `SELECT * FROM joinview;`,
				Results:   []sql.Row{{2, `more`, 42, 141, 12345}, {17, `zoo2`, 57, 99, 54321}},
			},
			{
				Statement: `SELECT * FROM foo;`,
				Results:   []sql.Row{{2, `more`, 42, 141}, {17, `zoo2`, 57, 99}},
			},
			{
				Statement: `SELECT * FROM voo;`,
				Results:   []sql.Row{{2, `more`}, {17, `zoo2`}},
			},
			{
				Statement: `INSERT INTO foo AS bar DEFAULT VALUES RETURNING *; -- ok`,
				Results:   []sql.Row{{4, ``, 42, 99}},
			},
			{
				Statement:   `INSERT INTO foo AS bar DEFAULT VALUES RETURNING foo.*; -- fails, wrong name`,
				ErrorString: `invalid reference to FROM-clause entry for table "foo"`,
			},
			{
				Statement: `INSERT INTO foo AS bar DEFAULT VALUES RETURNING bar.*; -- ok`,
				Results:   []sql.Row{{5, ``, 42, 99}},
			},
			{
				Statement: `INSERT INTO foo AS bar DEFAULT VALUES RETURNING bar.f3; -- ok`,
				Results:   []sql.Row{{42}},
			},
		},
	})
}
