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

func TestConstraints(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_constraints)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_constraints,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `\getenv abs_srcdir PG_ABS_SRCDIR
CREATE TABLE DEFAULT_TBL (i int DEFAULT 100,
	x text DEFAULT 'vadim', f float8 DEFAULT 123.456);`,
			},
			{
				Statement: `INSERT INTO DEFAULT_TBL VALUES (1, 'thomas', 57.0613);`,
			},
			{
				Statement: `INSERT INTO DEFAULT_TBL VALUES (1, 'bruce');`,
			},
			{
				Statement: `INSERT INTO DEFAULT_TBL (i, f) VALUES (2, 987.654);`,
			},
			{
				Statement: `INSERT INTO DEFAULT_TBL (x) VALUES ('marc');`,
			},
			{
				Statement: `INSERT INTO DEFAULT_TBL VALUES (3, null, 1.0);`,
			},
			{
				Statement: `SELECT * FROM DEFAULT_TBL;`,
				Results:   []sql.Row{{1, `thomas`, 57.0613}, {1, `bruce`, 123.456}, {2, `vadim`, 987.654}, {100, `marc`, 123.456}, {3, ``, 1}},
			},
			{
				Statement: `CREATE SEQUENCE DEFAULT_SEQ;`,
			},
			{
				Statement: `CREATE TABLE DEFAULTEXPR_TBL (i1 int DEFAULT 100 + (200-199) * 2,
	i2 int DEFAULT nextval('default_seq'));`,
			},
			{
				Statement: `INSERT INTO DEFAULTEXPR_TBL VALUES (-1, -2);`,
			},
			{
				Statement: `INSERT INTO DEFAULTEXPR_TBL (i1) VALUES (-3);`,
			},
			{
				Statement: `INSERT INTO DEFAULTEXPR_TBL (i2) VALUES (-4);`,
			},
			{
				Statement: `INSERT INTO DEFAULTEXPR_TBL (i2) VALUES (NULL);`,
			},
			{
				Statement: `SELECT * FROM DEFAULTEXPR_TBL;`,
				Results:   []sql.Row{{-1, -2}, {-3, 1}, {102, -4}, {102, ``}},
			},
			{
				Statement:   `CREATE TABLE error_tbl (i int DEFAULT (100, ));`,
				ErrorString: `syntax error at or near ")"`,
			},
			{
				Statement:   `CREATE TABLE error_tbl (b1 bool DEFAULT 1 IN (1, 2));`,
				ErrorString: `syntax error at or near "IN"`,
			},
			{
				Statement: `CREATE TABLE error_tbl (b1 bool DEFAULT (1 IN (1, 2)));`,
			},
			{
				Statement: `DROP TABLE error_tbl;`,
			},
			{
				Statement: `CREATE TABLE CHECK_TBL (x int,
	CONSTRAINT CHECK_CON CHECK (x > 3));`,
			},
			{
				Statement: `INSERT INTO CHECK_TBL VALUES (5);`,
			},
			{
				Statement: `INSERT INTO CHECK_TBL VALUES (4);`,
			},
			{
				Statement:   `INSERT INTO CHECK_TBL VALUES (3);`,
				ErrorString: `new row for relation "check_tbl" violates check constraint "check_con"`,
			},
			{
				Statement:   `INSERT INTO CHECK_TBL VALUES (2);`,
				ErrorString: `new row for relation "check_tbl" violates check constraint "check_con"`,
			},
			{
				Statement: `INSERT INTO CHECK_TBL VALUES (6);`,
			},
			{
				Statement:   `INSERT INTO CHECK_TBL VALUES (1);`,
				ErrorString: `new row for relation "check_tbl" violates check constraint "check_con"`,
			},
			{
				Statement: `SELECT * FROM CHECK_TBL;`,
				Results:   []sql.Row{{5}, {4}, {6}},
			},
			{
				Statement: `CREATE SEQUENCE CHECK_SEQ;`,
			},
			{
				Statement: `CREATE TABLE CHECK2_TBL (x int, y text, z int,
	CONSTRAINT SEQUENCE_CON
	CHECK (x > 3 and y <> 'check failed' and z < 8));`,
			},
			{
				Statement: `INSERT INTO CHECK2_TBL VALUES (4, 'check ok', -2);`,
			},
			{
				Statement:   `INSERT INTO CHECK2_TBL VALUES (1, 'x check failed', -2);`,
				ErrorString: `new row for relation "check2_tbl" violates check constraint "sequence_con"`,
			},
			{
				Statement:   `INSERT INTO CHECK2_TBL VALUES (5, 'z check failed', 10);`,
				ErrorString: `new row for relation "check2_tbl" violates check constraint "sequence_con"`,
			},
			{
				Statement:   `INSERT INTO CHECK2_TBL VALUES (0, 'check failed', -2);`,
				ErrorString: `new row for relation "check2_tbl" violates check constraint "sequence_con"`,
			},
			{
				Statement:   `INSERT INTO CHECK2_TBL VALUES (6, 'check failed', 11);`,
				ErrorString: `new row for relation "check2_tbl" violates check constraint "sequence_con"`,
			},
			{
				Statement: `INSERT INTO CHECK2_TBL VALUES (7, 'check ok', 7);`,
			},
			{
				Statement: `SELECT * from CHECK2_TBL;`,
				Results:   []sql.Row{{4, `check ok`, -2}, {7, `check ok`, 7}},
			},
			{
				Statement: `CREATE SEQUENCE INSERT_SEQ;`,
			},
			{
				Statement: `CREATE TABLE INSERT_TBL (x INT DEFAULT nextval('insert_seq'),
	y TEXT DEFAULT '-NULL-',
	z INT DEFAULT -1 * currval('insert_seq'),
	CONSTRAINT INSERT_TBL_CON CHECK (x >= 3 AND y <> 'check failed' AND x < 8),
	CHECK (x + z = 0));`,
			},
			{
				Statement:   `INSERT INTO INSERT_TBL(x,z) VALUES (2, -2);`,
				ErrorString: `new row for relation "insert_tbl" violates check constraint "insert_tbl_con"`,
			},
			{
				Statement: `SELECT * FROM INSERT_TBL;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT 'one' AS one, nextval('insert_seq');`,
				Results:   []sql.Row{{`one`, 1}},
			},
			{
				Statement:   `INSERT INTO INSERT_TBL(y) VALUES ('Y');`,
				ErrorString: `new row for relation "insert_tbl" violates check constraint "insert_tbl_con"`,
			},
			{
				Statement: `INSERT INTO INSERT_TBL(y) VALUES ('Y');`,
			},
			{
				Statement:   `INSERT INTO INSERT_TBL(x,z) VALUES (1, -2);`,
				ErrorString: `new row for relation "insert_tbl" violates check constraint "insert_tbl_check"`,
			},
			{
				Statement: `INSERT INTO INSERT_TBL(z,x) VALUES (-7,  7);`,
			},
			{
				Statement:   `INSERT INTO INSERT_TBL VALUES (5, 'check failed', -5);`,
				ErrorString: `new row for relation "insert_tbl" violates check constraint "insert_tbl_con"`,
			},
			{
				Statement: `INSERT INTO INSERT_TBL VALUES (7, '!check failed', -7);`,
			},
			{
				Statement: `INSERT INTO INSERT_TBL(y) VALUES ('-!NULL-');`,
			},
			{
				Statement: `SELECT * FROM INSERT_TBL;`,
				Results:   []sql.Row{{3, `Y`, -3}, {7, `-NULL-`, -7}, {7, `!check failed`, -7}, {4, `-!NULL-`, -4}},
			},
			{
				Statement:   `INSERT INTO INSERT_TBL(y,z) VALUES ('check failed', 4);`,
				ErrorString: `new row for relation "insert_tbl" violates check constraint "insert_tbl_check"`,
			},
			{
				Statement:   `INSERT INTO INSERT_TBL(x,y) VALUES (5, 'check failed');`,
				ErrorString: `new row for relation "insert_tbl" violates check constraint "insert_tbl_con"`,
			},
			{
				Statement: `INSERT INTO INSERT_TBL(x,y) VALUES (5, '!check failed');`,
			},
			{
				Statement: `INSERT INTO INSERT_TBL(y) VALUES ('-!NULL-');`,
			},
			{
				Statement: `SELECT * FROM INSERT_TBL;`,
				Results:   []sql.Row{{3, `Y`, -3}, {7, `-NULL-`, -7}, {7, `!check failed`, -7}, {4, `-!NULL-`, -4}, {5, `!check failed`, -5}, {6, `-!NULL-`, -6}},
			},
			{
				Statement: `SELECT 'seven' AS one, nextval('insert_seq');`,
				Results:   []sql.Row{{`seven`, 7}},
			},
			{
				Statement:   `INSERT INTO INSERT_TBL(y) VALUES ('Y');`,
				ErrorString: `new row for relation "insert_tbl" violates check constraint "insert_tbl_con"`,
			},
			{
				Statement: `SELECT 'eight' AS one, currval('insert_seq');`,
				Results:   []sql.Row{{`eight`, 8}},
			},
			{
				Statement: `INSERT INTO INSERT_TBL VALUES (null, null, null);`,
			},
			{
				Statement: `SELECT * FROM INSERT_TBL;`,
				Results:   []sql.Row{{3, `Y`, -3}, {7, `-NULL-`, -7}, {7, `!check failed`, -7}, {4, `-!NULL-`, -4}, {5, `!check failed`, -5}, {6, `-!NULL-`, -6}, {``, ``, ``}},
			},
			{
				Statement: `CREATE TABLE SYS_COL_CHECK_TBL (city text, state text, is_capital bool,
                  altitude int,
                  CHECK (NOT (is_capital AND tableoid::regclass::text = 'sys_col_check_tbl')));`,
			},
			{
				Statement: `INSERT INTO SYS_COL_CHECK_TBL VALUES ('Seattle', 'Washington', false, 100);`,
			},
			{
				Statement:   `INSERT INTO SYS_COL_CHECK_TBL VALUES ('Olympia', 'Washington', true, 100);`,
				ErrorString: `new row for relation "sys_col_check_tbl" violates check constraint "sys_col_check_tbl_check"`,
			},
			{
				Statement: `SELECT *, tableoid::regclass::text FROM SYS_COL_CHECK_TBL;`,
				Results:   []sql.Row{{`Seattle`, `Washington`, false, 100, `sys_col_check_tbl`}},
			},
			{
				Statement: `DROP TABLE SYS_COL_CHECK_TBL;`,
			},
			{
				Statement: `CREATE TABLE SYS_COL_CHECK_TBL (city text, state text, is_capital bool,
                  altitude int,
				  CHECK (NOT (is_capital AND ctid::text = 'sys_col_check_tbl')));`,
				ErrorString: `system column "ctid" reference in check constraint is invalid`,
			},
			{
				Statement: `CREATE TABLE INSERT_CHILD (cx INT default 42,
	cy INT CHECK (cy > x))
	INHERITS (INSERT_TBL);`,
			},
			{
				Statement: `INSERT INTO INSERT_CHILD(x,z,cy) VALUES (7,-7,11);`,
			},
			{
				Statement:   `INSERT INTO INSERT_CHILD(x,z,cy) VALUES (7,-7,6);`,
				ErrorString: `new row for relation "insert_child" violates check constraint "insert_child_check"`,
			},
			{
				Statement:   `INSERT INTO INSERT_CHILD(x,z,cy) VALUES (6,-7,7);`,
				ErrorString: `new row for relation "insert_child" violates check constraint "insert_tbl_check"`,
			},
			{
				Statement:   `INSERT INTO INSERT_CHILD(x,y,z,cy) VALUES (6,'check failed',-6,7);`,
				ErrorString: `new row for relation "insert_child" violates check constraint "insert_tbl_con"`,
			},
			{
				Statement: `SELECT * FROM INSERT_CHILD;`,
				Results:   []sql.Row{{7, `-NULL-`, -7, 42, 11}},
			},
			{
				Statement: `DROP TABLE INSERT_CHILD;`,
			},
			{
				Statement: `CREATE TABLE ATACC1 (TEST INT
	CHECK (TEST > 0) NO INHERIT);`,
			},
			{
				Statement: `CREATE TABLE ATACC2 (TEST2 INT) INHERITS (ATACC1);`,
			},
			{
				Statement: `INSERT INTO ATACC2 (TEST) VALUES (-3);`,
			},
			{
				Statement:   `INSERT INTO ATACC1 (TEST) VALUES (-3);`,
				ErrorString: `new row for relation "atacc1" violates check constraint "atacc1_test_check"`,
			},
			{
				Statement: `DROP TABLE ATACC1 CASCADE;`,
			},
			{
				Statement: `CREATE TABLE ATACC1 (TEST INT, TEST2 INT
	CHECK (TEST > 0), CHECK (TEST2 > 10) NO INHERIT);`,
			},
			{
				Statement: `CREATE TABLE ATACC2 () INHERITS (ATACC1);`,
			},
			{
				Statement:   `INSERT INTO ATACC2 (TEST) VALUES (-3);`,
				ErrorString: `new row for relation "atacc2" violates check constraint "atacc1_test_check"`,
			},
			{
				Statement:   `INSERT INTO ATACC1 (TEST) VALUES (-3);`,
				ErrorString: `new row for relation "atacc1" violates check constraint "atacc1_test_check"`,
			},
			{
				Statement: `INSERT INTO ATACC2 (TEST2) VALUES (3);`,
			},
			{
				Statement:   `INSERT INTO ATACC1 (TEST2) VALUES (3);`,
				ErrorString: `new row for relation "atacc1" violates check constraint "atacc1_test2_check"`,
			},
			{
				Statement: `DROP TABLE ATACC1 CASCADE;`,
			},
			{
				Statement: `DELETE FROM INSERT_TBL;`,
			},
			{
				Statement: `ALTER SEQUENCE INSERT_SEQ RESTART WITH 4;`,
			},
			{
				Statement: `CREATE TEMP TABLE tmp (xd INT, yd TEXT, zd INT);`,
			},
			{
				Statement: `INSERT INTO tmp VALUES (null, 'Y', null);`,
			},
			{
				Statement: `INSERT INTO tmp VALUES (5, '!check failed', null);`,
			},
			{
				Statement: `INSERT INTO tmp VALUES (null, 'try again', null);`,
			},
			{
				Statement: `INSERT INTO INSERT_TBL(y) select yd from tmp;`,
			},
			{
				Statement: `SELECT * FROM INSERT_TBL;`,
				Results:   []sql.Row{{4, `Y`, -4}, {5, `!check failed`, -5}, {6, `try again`, -6}},
			},
			{
				Statement: `INSERT INTO INSERT_TBL SELECT * FROM tmp WHERE yd = 'try again';`,
			},
			{
				Statement: `INSERT INTO INSERT_TBL(y,z) SELECT yd, -7 FROM tmp WHERE yd = 'try again';`,
			},
			{
				Statement:   `INSERT INTO INSERT_TBL(y,z) SELECT yd, -8 FROM tmp WHERE yd = 'try again';`,
				ErrorString: `new row for relation "insert_tbl" violates check constraint "insert_tbl_con"`,
			},
			{
				Statement: `SELECT * FROM INSERT_TBL;`,
				Results:   []sql.Row{{4, `Y`, -4}, {5, `!check failed`, -5}, {6, `try again`, -6}, {``, `try again`, ``}, {7, `try again`, -7}},
			},
			{
				Statement: `DROP TABLE tmp;`,
			},
			{
				Statement: `UPDATE INSERT_TBL SET x = NULL WHERE x = 5;`,
			},
			{
				Statement: `UPDATE INSERT_TBL SET x = 6 WHERE x = 6;`,
			},
			{
				Statement: `UPDATE INSERT_TBL SET x = -z, z = -x;`,
			},
			{
				Statement:   `UPDATE INSERT_TBL SET x = z, z = x;`,
				ErrorString: `new row for relation "insert_tbl" violates check constraint "insert_tbl_con"`,
			},
			{
				Statement: `SELECT * FROM INSERT_TBL;`,
				Results:   []sql.Row{{4, `Y`, -4}, {``, `try again`, ``}, {7, `try again`, -7}, {5, `!check failed`, ``}, {6, `try again`, -6}},
			},
			{
				Statement: `CREATE TABLE COPY_TBL (x INT, y TEXT, z INT,
	CONSTRAINT COPY_CON
	CHECK (x > 3 AND y <> 'check failed' AND x < 7 ));`,
			},
			{
				Statement: `\set filename :abs_srcdir '/data/constro.data'
COPY COPY_TBL FROM :'filename';`,
			},
			{
				Statement: `SELECT * FROM COPY_TBL;`,
				Results:   []sql.Row{{4, `!check failed`, 5}, {6, `OK`, 4}},
			},
			{
				Statement: `\set filename :abs_srcdir '/data/constrf.data'
COPY COPY_TBL FROM :'filename';`,
				ErrorString: `new row for relation "copy_tbl" violates check constraint "copy_con"`,
			},
			{
				Statement: `CONTEXT:  COPY copy_tbl, line 2: "7	check failed	6"
SELECT * FROM COPY_TBL;`,
				Results: []sql.Row{{4, `!check failed`, 5}, {6, `OK`, 4}},
			},
			{
				Statement: `CREATE TABLE PRIMARY_TBL (i int PRIMARY KEY, t text);`,
			},
			{
				Statement: `INSERT INTO PRIMARY_TBL VALUES (1, 'one');`,
			},
			{
				Statement: `INSERT INTO PRIMARY_TBL VALUES (2, 'two');`,
			},
			{
				Statement:   `INSERT INTO PRIMARY_TBL VALUES (1, 'three');`,
				ErrorString: `duplicate key value violates unique constraint "primary_tbl_pkey"`,
			},
			{
				Statement: `INSERT INTO PRIMARY_TBL VALUES (4, 'three');`,
			},
			{
				Statement: `INSERT INTO PRIMARY_TBL VALUES (5, 'one');`,
			},
			{
				Statement:   `INSERT INTO PRIMARY_TBL (t) VALUES ('six');`,
				ErrorString: `null value in column "i" of relation "primary_tbl" violates not-null constraint`,
			},
			{
				Statement: `SELECT * FROM PRIMARY_TBL;`,
				Results:   []sql.Row{{1, `one`}, {2, `two`}, {4, `three`}, {5, `one`}},
			},
			{
				Statement: `DROP TABLE PRIMARY_TBL;`,
			},
			{
				Statement: `CREATE TABLE PRIMARY_TBL (i int, t text,
	PRIMARY KEY(i,t));`,
			},
			{
				Statement: `INSERT INTO PRIMARY_TBL VALUES (1, 'one');`,
			},
			{
				Statement: `INSERT INTO PRIMARY_TBL VALUES (2, 'two');`,
			},
			{
				Statement: `INSERT INTO PRIMARY_TBL VALUES (1, 'three');`,
			},
			{
				Statement: `INSERT INTO PRIMARY_TBL VALUES (4, 'three');`,
			},
			{
				Statement: `INSERT INTO PRIMARY_TBL VALUES (5, 'one');`,
			},
			{
				Statement:   `INSERT INTO PRIMARY_TBL (t) VALUES ('six');`,
				ErrorString: `null value in column "i" of relation "primary_tbl" violates not-null constraint`,
			},
			{
				Statement: `SELECT * FROM PRIMARY_TBL;`,
				Results:   []sql.Row{{1, `one`}, {2, `two`}, {1, `three`}, {4, `three`}, {5, `one`}},
			},
			{
				Statement: `DROP TABLE PRIMARY_TBL;`,
			},
			{
				Statement: `CREATE TABLE UNIQUE_TBL (i int UNIQUE, t text);`,
			},
			{
				Statement: `INSERT INTO UNIQUE_TBL VALUES (1, 'one');`,
			},
			{
				Statement: `INSERT INTO UNIQUE_TBL VALUES (2, 'two');`,
			},
			{
				Statement:   `INSERT INTO UNIQUE_TBL VALUES (1, 'three');`,
				ErrorString: `duplicate key value violates unique constraint "unique_tbl_i_key"`,
			},
			{
				Statement: `INSERT INTO UNIQUE_TBL VALUES (4, 'four');`,
			},
			{
				Statement: `INSERT INTO UNIQUE_TBL VALUES (5, 'one');`,
			},
			{
				Statement: `INSERT INTO UNIQUE_TBL (t) VALUES ('six');`,
			},
			{
				Statement: `INSERT INTO UNIQUE_TBL (t) VALUES ('seven');`,
			},
			{
				Statement: `INSERT INTO UNIQUE_TBL VALUES (5, 'five-upsert-insert') ON CONFLICT (i) DO UPDATE SET t = 'five-upsert-update';`,
			},
			{
				Statement: `INSERT INTO UNIQUE_TBL VALUES (6, 'six-upsert-insert') ON CONFLICT (i) DO UPDATE SET t = 'six-upsert-update';`,
			},
			{
				Statement:   `INSERT INTO UNIQUE_TBL VALUES (1, 'a'), (2, 'b'), (2, 'b') ON CONFLICT (i) DO UPDATE SET t = 'fails';`,
				ErrorString: `ON CONFLICT DO UPDATE command cannot affect row a second time`,
			},
			{
				Statement: `SELECT * FROM UNIQUE_TBL;`,
				Results:   []sql.Row{{1, `one`}, {2, `two`}, {4, `four`}, {``, `six`}, {``, `seven`}, {5, `five-upsert-update`}, {6, `six-upsert-insert`}},
			},
			{
				Statement: `DROP TABLE UNIQUE_TBL;`,
			},
			{
				Statement: `CREATE TABLE UNIQUE_TBL (i int UNIQUE NULLS NOT DISTINCT, t text);`,
			},
			{
				Statement: `INSERT INTO UNIQUE_TBL VALUES (1, 'one');`,
			},
			{
				Statement: `INSERT INTO UNIQUE_TBL VALUES (2, 'two');`,
			},
			{
				Statement:   `INSERT INTO UNIQUE_TBL VALUES (1, 'three');  -- fail`,
				ErrorString: `duplicate key value violates unique constraint "unique_tbl_i_key"`,
			},
			{
				Statement: `INSERT INTO UNIQUE_TBL VALUES (4, 'four');`,
			},
			{
				Statement: `INSERT INTO UNIQUE_TBL VALUES (5, 'one');`,
			},
			{
				Statement: `INSERT INTO UNIQUE_TBL (t) VALUES ('six');`,
			},
			{
				Statement:   `INSERT INTO UNIQUE_TBL (t) VALUES ('seven');  -- fail`,
				ErrorString: `duplicate key value violates unique constraint "unique_tbl_i_key"`,
			},
			{
				Statement: `INSERT INTO UNIQUE_TBL (t) VALUES ('eight') ON CONFLICT DO NOTHING;  -- no-op`,
			},
			{
				Statement: `SELECT * FROM UNIQUE_TBL;`,
				Results:   []sql.Row{{1, `one`}, {2, `two`}, {4, `four`}, {5, `one`}, {``, `six`}},
			},
			{
				Statement: `DROP TABLE UNIQUE_TBL;`,
			},
			{
				Statement: `CREATE TABLE UNIQUE_TBL (i int, t text,
	UNIQUE(i,t));`,
			},
			{
				Statement: `INSERT INTO UNIQUE_TBL VALUES (1, 'one');`,
			},
			{
				Statement: `INSERT INTO UNIQUE_TBL VALUES (2, 'two');`,
			},
			{
				Statement: `INSERT INTO UNIQUE_TBL VALUES (1, 'three');`,
			},
			{
				Statement:   `INSERT INTO UNIQUE_TBL VALUES (1, 'one');`,
				ErrorString: `duplicate key value violates unique constraint "unique_tbl_i_t_key"`,
			},
			{
				Statement: `INSERT INTO UNIQUE_TBL VALUES (5, 'one');`,
			},
			{
				Statement: `INSERT INTO UNIQUE_TBL (t) VALUES ('six');`,
			},
			{
				Statement: `SELECT * FROM UNIQUE_TBL;`,
				Results:   []sql.Row{{1, `one`}, {2, `two`}, {1, `three`}, {5, `one`}, {``, `six`}},
			},
			{
				Statement: `DROP TABLE UNIQUE_TBL;`,
			},
			{
				Statement: `CREATE TABLE unique_tbl (i int UNIQUE DEFERRABLE, t text);`,
			},
			{
				Statement: `INSERT INTO unique_tbl VALUES (0, 'one');`,
			},
			{
				Statement: `INSERT INTO unique_tbl VALUES (1, 'two');`,
			},
			{
				Statement: `INSERT INTO unique_tbl VALUES (2, 'tree');`,
			},
			{
				Statement: `INSERT INTO unique_tbl VALUES (3, 'four');`,
			},
			{
				Statement: `INSERT INTO unique_tbl VALUES (4, 'five');`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement:   `UPDATE unique_tbl SET i = 1 WHERE i = 0;`,
				ErrorString: `duplicate key value violates unique constraint "unique_tbl_i_key"`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `UPDATE unique_tbl SET i = i+1;`,
			},
			{
				Statement: `SELECT * FROM unique_tbl;`,
				Results:   []sql.Row{{1, `one`}, {2, `two`}, {3, `tree`}, {4, `four`}, {5, `five`}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SET CONSTRAINTS unique_tbl_i_key DEFERRED;`,
			},
			{
				Statement: `INSERT INTO unique_tbl VALUES (3, 'three');`,
			},
			{
				Statement: `DELETE FROM unique_tbl WHERE t = 'tree'; -- makes constraint valid again`,
			},
			{
				Statement: `COMMIT; -- should succeed`,
			},
			{
				Statement: `SELECT * FROM unique_tbl;`,
				Results:   []sql.Row{{1, `one`}, {2, `two`}, {4, `four`}, {5, `five`}, {3, `three`}},
			},
			{
				Statement: `ALTER TABLE unique_tbl DROP CONSTRAINT unique_tbl_i_key;`,
			},
			{
				Statement: `ALTER TABLE unique_tbl ADD CONSTRAINT unique_tbl_i_key
	UNIQUE (i) DEFERRABLE INITIALLY DEFERRED;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `INSERT INTO unique_tbl VALUES (1, 'five');`,
			},
			{
				Statement: `INSERT INTO unique_tbl VALUES (5, 'one');`,
			},
			{
				Statement: `UPDATE unique_tbl SET i = 4 WHERE i = 2;`,
			},
			{
				Statement: `UPDATE unique_tbl SET i = 2 WHERE i = 4 AND t = 'four';`,
			},
			{
				Statement: `DELETE FROM unique_tbl WHERE i = 1 AND t = 'one';`,
			},
			{
				Statement: `DELETE FROM unique_tbl WHERE i = 5 AND t = 'five';`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `SELECT * FROM unique_tbl;`,
				Results:   []sql.Row{{3, `three`}, {1, `five`}, {5, `one`}, {4, `two`}, {2, `four`}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `INSERT INTO unique_tbl VALUES (3, 'Three'); -- should succeed for now`,
			},
			{
				Statement:   `COMMIT; -- should fail`,
				ErrorString: `duplicate key value violates unique constraint "unique_tbl_i_key"`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SET CONSTRAINTS ALL IMMEDIATE;`,
			},
			{
				Statement:   `INSERT INTO unique_tbl VALUES (3, 'Three'); -- should fail`,
				ErrorString: `duplicate key value violates unique constraint "unique_tbl_i_key"`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SET CONSTRAINTS ALL DEFERRED;`,
			},
			{
				Statement: `INSERT INTO unique_tbl VALUES (3, 'Three'); -- should succeed for now`,
			},
			{
				Statement:   `SET CONSTRAINTS ALL IMMEDIATE; -- should fail`,
				ErrorString: `duplicate key value violates unique constraint "unique_tbl_i_key"`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `CREATE TABLE parted_uniq_tbl (i int UNIQUE DEFERRABLE) partition by range (i);`,
			},
			{
				Statement: `CREATE TABLE parted_uniq_tbl_1 PARTITION OF parted_uniq_tbl FOR VALUES FROM (0) TO (10);`,
			},
			{
				Statement: `CREATE TABLE parted_uniq_tbl_2 PARTITION OF parted_uniq_tbl FOR VALUES FROM (20) TO (30);`,
			},
			{
				Statement: `SELECT conname, conrelid::regclass FROM pg_constraint
  WHERE conname LIKE 'parted_uniq%' ORDER BY conname;`,
				Results: []sql.Row{{`parted_uniq_tbl_1_i_key`, `parted_uniq_tbl_1`}, {`parted_uniq_tbl_2_i_key`, `parted_uniq_tbl_2`}, {`parted_uniq_tbl_i_key`, `parted_uniq_tbl`}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `INSERT INTO parted_uniq_tbl VALUES (1);`,
			},
			{
				Statement: `SAVEPOINT f;`,
			},
			{
				Statement:   `INSERT INTO parted_uniq_tbl VALUES (1);	-- unique violation`,
				ErrorString: `duplicate key value violates unique constraint "parted_uniq_tbl_1_i_key"`,
			},
			{
				Statement: `ROLLBACK TO f;`,
			},
			{
				Statement: `SET CONSTRAINTS parted_uniq_tbl_i_key DEFERRED;`,
			},
			{
				Statement: `INSERT INTO parted_uniq_tbl VALUES (1);	-- OK now, fail at commit`,
			},
			{
				Statement:   `COMMIT;`,
				ErrorString: `duplicate key value violates unique constraint "parted_uniq_tbl_1_i_key"`,
			},
			{
				Statement: `DROP TABLE parted_uniq_tbl;`,
			},
			{
				Statement: `CREATE TABLE parted_fk_naming (
    id bigint NOT NULL default 1,
    id_abc bigint,
    CONSTRAINT dummy_constr FOREIGN KEY (id_abc)
        REFERENCES parted_fk_naming (id),
    PRIMARY KEY (id)
)
PARTITION BY LIST (id);`,
			},
			{
				Statement: `CREATE TABLE parted_fk_naming_1 (
    id bigint NOT NULL default 1,
    id_abc bigint,
    PRIMARY KEY (id),
    CONSTRAINT dummy_constr CHECK (true)
);`,
			},
			{
				Statement: `ALTER TABLE parted_fk_naming ATTACH PARTITION parted_fk_naming_1 FOR VALUES IN ('1');`,
			},
			{
				Statement: `SELECT conname FROM pg_constraint WHERE conrelid = 'parted_fk_naming_1'::regclass AND contype = 'f';`,
				Results:   []sql.Row{{`parted_fk_naming_1_id_abc_fkey`}},
			},
			{
				Statement: `DROP TABLE parted_fk_naming;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `INSERT INTO unique_tbl VALUES (3, 'Three'); -- should succeed for now`,
			},
			{
				Statement: `UPDATE unique_tbl SET t = 'THREE' WHERE i = 3 AND t = 'Three';`,
			},
			{
				Statement:   `COMMIT; -- should fail`,
				ErrorString: `duplicate key value violates unique constraint "unique_tbl_i_key"`,
			},
			{
				Statement: `SELECT * FROM unique_tbl;`,
				Results:   []sql.Row{{3, `three`}, {1, `five`}, {5, `one`}, {4, `two`}, {2, `four`}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `INSERT INTO unique_tbl VALUES(3, 'tree'); -- should succeed for now`,
			},
			{
				Statement: `UPDATE unique_tbl SET t = 'threex' WHERE t = 'tree';`,
			},
			{
				Statement: `DELETE FROM unique_tbl WHERE t = 'three';`,
			},
			{
				Statement: `SELECT * FROM unique_tbl;`,
				Results:   []sql.Row{{1, `five`}, {5, `one`}, {4, `two`}, {2, `four`}, {3, `threex`}},
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `SELECT * FROM unique_tbl;`,
				Results:   []sql.Row{{1, `five`}, {5, `one`}, {4, `two`}, {2, `four`}, {3, `threex`}},
			},
			{
				Statement: `DROP TABLE unique_tbl;`,
			},
			{
				Statement: `CREATE TABLE circles (
  c1 CIRCLE,
  c2 TEXT,
  EXCLUDE USING gist
    (c1 WITH &&, (c2::circle) WITH &&)
    WHERE (circle_center(c1) <> '(0,0)')
);`,
			},
			{
				Statement: `INSERT INTO circles VALUES('<(0,0), 5>', '<(0,0), 5>');`,
			},
			{
				Statement: `INSERT INTO circles VALUES('<(0,0), 5>', '<(0,0), 4>');`,
			},
			{
				Statement: `INSERT INTO circles VALUES('<(10,10), 10>', '<(0,0), 5>');`,
			},
			{
				Statement:   `INSERT INTO circles VALUES('<(20,20), 10>', '<(0,0), 4>');`,
				ErrorString: `conflicting key value violates exclusion constraint "circles_c1_c2_excl"`,
			},
			{
				Statement: `INSERT INTO circles VALUES('<(20,20), 10>', '<(0,0), 4>')
  ON CONFLICT ON CONSTRAINT circles_c1_c2_excl DO NOTHING;`,
			},
			{
				Statement: `INSERT INTO circles VALUES('<(20,20), 10>', '<(0,0), 4>')
  ON CONFLICT ON CONSTRAINT circles_c1_c2_excl DO UPDATE SET c2 = EXCLUDED.c2;`,
				ErrorString: `ON CONFLICT DO UPDATE not supported with exclusion constraints`,
			},
			{
				Statement: `INSERT INTO circles VALUES('<(20,20), 1>', '<(0,0), 5>');`,
			},
			{
				Statement: `INSERT INTO circles VALUES('<(20,20), 10>', '<(10,10), 5>');`,
			},
			{
				Statement: `ALTER TABLE circles ADD EXCLUDE USING gist
  (c1 WITH &&, (c2::circle) WITH &&);`,
				ErrorString: `could not create exclusion constraint "circles_c1_c2_excl1"`,
			},
			{
				Statement: `REINDEX INDEX circles_c1_c2_excl;`,
			},
			{
				Statement: `DROP TABLE circles;`,
			},
			{
				Statement: `CREATE TABLE deferred_excl (
  f1 int,
  f2 int,
  CONSTRAINT deferred_excl_con EXCLUDE (f1 WITH =) INITIALLY DEFERRED
);`,
			},
			{
				Statement: `INSERT INTO deferred_excl VALUES(1);`,
			},
			{
				Statement: `INSERT INTO deferred_excl VALUES(2);`,
			},
			{
				Statement:   `INSERT INTO deferred_excl VALUES(1); -- fail`,
				ErrorString: `conflicting key value violates exclusion constraint "deferred_excl_con"`,
			},
			{
				Statement:   `INSERT INTO deferred_excl VALUES(1) ON CONFLICT ON CONSTRAINT deferred_excl_con DO NOTHING; -- fail`,
				ErrorString: `ON CONFLICT does not support deferrable unique constraints/exclusion constraints as arbiters`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `INSERT INTO deferred_excl VALUES(2); -- no fail here`,
			},
			{
				Statement:   `COMMIT; -- should fail here`,
				ErrorString: `conflicting key value violates exclusion constraint "deferred_excl_con"`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `INSERT INTO deferred_excl VALUES(3);`,
			},
			{
				Statement: `INSERT INTO deferred_excl VALUES(3); -- no fail here`,
			},
			{
				Statement:   `COMMIT; -- should fail here`,
				ErrorString: `conflicting key value violates exclusion constraint "deferred_excl_con"`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `INSERT INTO deferred_excl VALUES(2, 1); -- no fail here`,
			},
			{
				Statement: `DELETE FROM deferred_excl WHERE f1 = 2 AND f2 IS NULL; -- remove old row`,
			},
			{
				Statement: `UPDATE deferred_excl SET f2 = 2 WHERE f1 = 2;`,
			},
			{
				Statement: `COMMIT; -- should not fail`,
			},
			{
				Statement: `SELECT * FROM deferred_excl;`,
				Results:   []sql.Row{{1, ``}, {2, 2}},
			},
			{
				Statement: `ALTER TABLE deferred_excl DROP CONSTRAINT deferred_excl_con;`,
			},
			{
				Statement: `UPDATE deferred_excl SET f1 = 3;`,
			},
			{
				Statement:   `ALTER TABLE deferred_excl ADD EXCLUDE (f1 WITH =);`,
				ErrorString: `could not create exclusion constraint "deferred_excl_f1_excl"`,
			},
			{
				Statement: `DROP TABLE deferred_excl;`,
			},
			{
				Statement: `CREATE ROLE regress_constraint_comments;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_constraint_comments;`,
			},
			{
				Statement: `CREATE TABLE constraint_comments_tbl (a int CONSTRAINT the_constraint CHECK (a > 0));`,
			},
			{
				Statement: `CREATE DOMAIN constraint_comments_dom AS int CONSTRAINT the_constraint CHECK (value > 0);`,
			},
			{
				Statement: `COMMENT ON CONSTRAINT the_constraint ON constraint_comments_tbl IS 'yes, the comment';`,
			},
			{
				Statement: `COMMENT ON CONSTRAINT the_constraint ON DOMAIN constraint_comments_dom IS 'yes, another comment';`,
			},
			{
				Statement:   `COMMENT ON CONSTRAINT no_constraint ON constraint_comments_tbl IS 'yes, the comment';`,
				ErrorString: `constraint "no_constraint" for table "constraint_comments_tbl" does not exist`,
			},
			{
				Statement:   `COMMENT ON CONSTRAINT no_constraint ON DOMAIN constraint_comments_dom IS 'yes, another comment';`,
				ErrorString: `constraint "no_constraint" for domain constraint_comments_dom does not exist`,
			},
			{
				Statement:   `COMMENT ON CONSTRAINT the_constraint ON no_comments_tbl IS 'bad comment';`,
				ErrorString: `relation "no_comments_tbl" does not exist`,
			},
			{
				Statement:   `COMMENT ON CONSTRAINT the_constraint ON DOMAIN no_comments_dom IS 'another bad comment';`,
				ErrorString: `type "no_comments_dom" does not exist`,
			},
			{
				Statement: `COMMENT ON CONSTRAINT the_constraint ON constraint_comments_tbl IS NULL;`,
			},
			{
				Statement: `COMMENT ON CONSTRAINT the_constraint ON DOMAIN constraint_comments_dom IS NULL;`,
			},
			{
				Statement: `RESET SESSION AUTHORIZATION;`,
			},
			{
				Statement: `CREATE ROLE regress_constraint_comments_noaccess;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_constraint_comments_noaccess;`,
			},
			{
				Statement:   `COMMENT ON CONSTRAINT the_constraint ON constraint_comments_tbl IS 'no, the comment';`,
				ErrorString: `must be owner of relation constraint_comments_tbl`,
			},
			{
				Statement:   `COMMENT ON CONSTRAINT the_constraint ON DOMAIN constraint_comments_dom IS 'no, another comment';`,
				ErrorString: `must be owner of type constraint_comments_dom`,
			},
			{
				Statement: `RESET SESSION AUTHORIZATION;`,
			},
			{
				Statement: `DROP TABLE constraint_comments_tbl;`,
			},
			{
				Statement: `DROP DOMAIN constraint_comments_dom;`,
			},
			{
				Statement: `DROP ROLE regress_constraint_comments;`,
			},
			{
				Statement: `DROP ROLE regress_constraint_comments_noaccess;`,
			},
		},
	})
}
