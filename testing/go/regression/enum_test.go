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

func TestEnum(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_enum)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_enum,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TYPE rainbow AS ENUM ('red', 'orange', 'yellow', 'green', 'blue', 'purple');`,
			},
			{
				Statement: `SELECT COUNT(*) FROM pg_enum WHERE enumtypid = 'rainbow'::regtype;`,
				Results:   []sql.Row{{6}},
			},
			{
				Statement: `SELECT 'red'::rainbow;`,
				Results:   []sql.Row{{`red`}},
			},
			{
				Statement:   `SELECT 'mauve'::rainbow;`,
				ErrorString: `invalid input value for enum rainbow: "mauve"`,
			},
			{
				Statement: `CREATE TYPE planets AS ENUM ( 'venus', 'earth', 'mars' );`,
			},
			{
				Statement: `SELECT enumlabel, enumsortorder
FROM pg_enum
WHERE enumtypid = 'planets'::regtype
ORDER BY 2;`,
				Results: []sql.Row{{`venus`, 1}, {`earth`, 2}, {`mars`, 3}},
			},
			{
				Statement: `ALTER TYPE planets ADD VALUE 'uranus';`,
			},
			{
				Statement: `SELECT enumlabel, enumsortorder
FROM pg_enum
WHERE enumtypid = 'planets'::regtype
ORDER BY 2;`,
				Results: []sql.Row{{`venus`, 1}, {`earth`, 2}, {`mars`, 3}, {`uranus`, 4}},
			},
			{
				Statement: `ALTER TYPE planets ADD VALUE 'mercury' BEFORE 'venus';`,
			},
			{
				Statement: `ALTER TYPE planets ADD VALUE 'saturn' BEFORE 'uranus';`,
			},
			{
				Statement: `ALTER TYPE planets ADD VALUE 'jupiter' AFTER 'mars';`,
			},
			{
				Statement: `ALTER TYPE planets ADD VALUE 'neptune' AFTER 'uranus';`,
			},
			{
				Statement: `SELECT enumlabel, enumsortorder
FROM pg_enum
WHERE enumtypid = 'planets'::regtype
ORDER BY 2;`,
				Results: []sql.Row{{`mercury`, 0}, {`venus`, 1}, {`earth`, 2}, {`mars`, 3}, {`jupiter`, 3.25}, {`saturn`, 3.5}, {`uranus`, 4}, {`neptune`, 5}},
			},
			{
				Statement: `SELECT enumlabel, enumsortorder
FROM pg_enum
WHERE enumtypid = 'planets'::regtype
ORDER BY enumlabel::planets;`,
				Results: []sql.Row{{`mercury`, 0}, {`venus`, 1}, {`earth`, 2}, {`mars`, 3}, {`jupiter`, 3.25}, {`saturn`, 3.5}, {`uranus`, 4}, {`neptune`, 5}},
			},
			{
				Statement: `ALTER TYPE planets ADD VALUE
  'plutoplutoplutoplutoplutoplutoplutoplutoplutoplutoplutoplutoplutopluto';`,
				ErrorString: `invalid enum label "plutoplutoplutoplutoplutoplutoplutoplutoplutoplutoplutoplutoplutopluto"`,
			},
			{
				Statement:   `ALTER TYPE planets ADD VALUE 'pluto' AFTER 'zeus';`,
				ErrorString: `"zeus" is not an existing enum label`,
			},
			{
				Statement:   `ALTER TYPE planets ADD VALUE 'mercury';`,
				ErrorString: `enum label "mercury" already exists`,
			},
			{
				Statement: `ALTER TYPE planets ADD VALUE IF NOT EXISTS 'mercury';`,
			},
			{
				Statement: `SELECT enum_last(NULL::planets);`,
				Results:   []sql.Row{{`neptune`}},
			},
			{
				Statement: `ALTER TYPE planets ADD VALUE IF NOT EXISTS 'pluto';`,
			},
			{
				Statement: `SELECT enum_last(NULL::planets);`,
				Results:   []sql.Row{{`pluto`}},
			},
			{
				Statement: `create type insenum as enum ('L1', 'L2');`,
			},
			{
				Statement: `alter type insenum add value 'i1' before 'L2';`,
			},
			{
				Statement: `alter type insenum add value 'i2' before 'L2';`,
			},
			{
				Statement: `alter type insenum add value 'i3' before 'L2';`,
			},
			{
				Statement: `alter type insenum add value 'i4' before 'L2';`,
			},
			{
				Statement: `alter type insenum add value 'i5' before 'L2';`,
			},
			{
				Statement: `alter type insenum add value 'i6' before 'L2';`,
			},
			{
				Statement: `alter type insenum add value 'i7' before 'L2';`,
			},
			{
				Statement: `alter type insenum add value 'i8' before 'L2';`,
			},
			{
				Statement: `alter type insenum add value 'i9' before 'L2';`,
			},
			{
				Statement: `alter type insenum add value 'i10' before 'L2';`,
			},
			{
				Statement: `alter type insenum add value 'i11' before 'L2';`,
			},
			{
				Statement: `alter type insenum add value 'i12' before 'L2';`,
			},
			{
				Statement: `alter type insenum add value 'i13' before 'L2';`,
			},
			{
				Statement: `alter type insenum add value 'i14' before 'L2';`,
			},
			{
				Statement: `alter type insenum add value 'i15' before 'L2';`,
			},
			{
				Statement: `alter type insenum add value 'i16' before 'L2';`,
			},
			{
				Statement: `alter type insenum add value 'i17' before 'L2';`,
			},
			{
				Statement: `alter type insenum add value 'i18' before 'L2';`,
			},
			{
				Statement: `alter type insenum add value 'i19' before 'L2';`,
			},
			{
				Statement: `alter type insenum add value 'i20' before 'L2';`,
			},
			{
				Statement: `alter type insenum add value 'i21' before 'L2';`,
			},
			{
				Statement: `alter type insenum add value 'i22' before 'L2';`,
			},
			{
				Statement: `alter type insenum add value 'i23' before 'L2';`,
			},
			{
				Statement: `alter type insenum add value 'i24' before 'L2';`,
			},
			{
				Statement: `alter type insenum add value 'i25' before 'L2';`,
			},
			{
				Statement: `alter type insenum add value 'i26' before 'L2';`,
			},
			{
				Statement: `alter type insenum add value 'i27' before 'L2';`,
			},
			{
				Statement: `alter type insenum add value 'i28' before 'L2';`,
			},
			{
				Statement: `alter type insenum add value 'i29' before 'L2';`,
			},
			{
				Statement: `alter type insenum add value 'i30' before 'L2';`,
			},
			{
				Statement: `SELECT enumlabel,
       case when enumsortorder > 20 then null else enumsortorder end as so
FROM pg_enum
WHERE enumtypid = 'insenum'::regtype
ORDER BY enumsortorder;`,
				Results: []sql.Row{{`L1`, 1}, {`i1`, 2}, {`i2`, 3}, {`i3`, 4}, {`i4`, 5}, {`i5`, 6}, {`i6`, 7}, {`i7`, 8}, {`i8`, 9}, {`i9`, 10}, {`i10`, 11}, {`i11`, 12}, {`i12`, 13}, {`i13`, 14}, {`i14`, 15}, {`i15`, 16}, {`i16`, 17}, {`i17`, 18}, {`i18`, 19}, {`i19`, 20}, {`i20`, ``}, {`i21`, ``}, {`i22`, ``}, {`i23`, ``}, {`i24`, ``}, {`i25`, ``}, {`i26`, ``}, {`i27`, ``}, {`i28`, ``}, {`i29`, ``}, {`i30`, ``}, {`L2`, ``}},
			},
			{
				Statement: `CREATE TABLE enumtest (col rainbow);`,
			},
			{
				Statement: `INSERT INTO enumtest values ('red'), ('orange'), ('yellow'), ('green');`,
			},
			{
				Statement: `COPY enumtest FROM stdin;`,
			},
			{
				Statement: `SELECT * FROM enumtest;`,
				Results:   []sql.Row{{`red`}, {`orange`}, {`yellow`}, {`green`}, {`blue`}, {`purple`}},
			},
			{
				Statement: `SELECT * FROM enumtest WHERE col = 'orange';`,
				Results:   []sql.Row{{`orange`}},
			},
			{
				Statement: `SELECT * FROM enumtest WHERE col <> 'orange' ORDER BY col;`,
				Results:   []sql.Row{{`red`}, {`yellow`}, {`green`}, {`blue`}, {`purple`}},
			},
			{
				Statement: `SELECT * FROM enumtest WHERE col > 'yellow' ORDER BY col;`,
				Results:   []sql.Row{{`green`}, {`blue`}, {`purple`}},
			},
			{
				Statement: `SELECT * FROM enumtest WHERE col >= 'yellow' ORDER BY col;`,
				Results:   []sql.Row{{`yellow`}, {`green`}, {`blue`}, {`purple`}},
			},
			{
				Statement: `SELECT * FROM enumtest WHERE col < 'green' ORDER BY col;`,
				Results:   []sql.Row{{`red`}, {`orange`}, {`yellow`}},
			},
			{
				Statement: `SELECT * FROM enumtest WHERE col <= 'green' ORDER BY col;`,
				Results:   []sql.Row{{`red`}, {`orange`}, {`yellow`}, {`green`}},
			},
			{
				Statement: `SELECT 'red'::rainbow::text || 'hithere';`,
				Results:   []sql.Row{{`redhithere`}},
			},
			{
				Statement: `SELECT 'red'::text::rainbow = 'red'::rainbow;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT min(col) FROM enumtest;`,
				Results:   []sql.Row{{`red`}},
			},
			{
				Statement: `SELECT max(col) FROM enumtest;`,
				Results:   []sql.Row{{`purple`}},
			},
			{
				Statement: `SELECT max(col) FROM enumtest WHERE col < 'green';`,
				Results:   []sql.Row{{`yellow`}},
			},
			{
				Statement: `SET enable_seqscan = off;`,
			},
			{
				Statement: `SET enable_bitmapscan = off;`,
			},
			{
				Statement: `CREATE UNIQUE INDEX enumtest_btree ON enumtest USING btree (col);`,
			},
			{
				Statement: `SELECT * FROM enumtest WHERE col = 'orange';`,
				Results:   []sql.Row{{`orange`}},
			},
			{
				Statement: `SELECT * FROM enumtest WHERE col <> 'orange' ORDER BY col;`,
				Results:   []sql.Row{{`red`}, {`yellow`}, {`green`}, {`blue`}, {`purple`}},
			},
			{
				Statement: `SELECT * FROM enumtest WHERE col > 'yellow' ORDER BY col;`,
				Results:   []sql.Row{{`green`}, {`blue`}, {`purple`}},
			},
			{
				Statement: `SELECT * FROM enumtest WHERE col >= 'yellow' ORDER BY col;`,
				Results:   []sql.Row{{`yellow`}, {`green`}, {`blue`}, {`purple`}},
			},
			{
				Statement: `SELECT * FROM enumtest WHERE col < 'green' ORDER BY col;`,
				Results:   []sql.Row{{`red`}, {`orange`}, {`yellow`}},
			},
			{
				Statement: `SELECT * FROM enumtest WHERE col <= 'green' ORDER BY col;`,
				Results:   []sql.Row{{`red`}, {`orange`}, {`yellow`}, {`green`}},
			},
			{
				Statement: `SELECT min(col) FROM enumtest;`,
				Results:   []sql.Row{{`red`}},
			},
			{
				Statement: `SELECT max(col) FROM enumtest;`,
				Results:   []sql.Row{{`purple`}},
			},
			{
				Statement: `SELECT max(col) FROM enumtest WHERE col < 'green';`,
				Results:   []sql.Row{{`yellow`}},
			},
			{
				Statement: `DROP INDEX enumtest_btree;`,
			},
			{
				Statement: `CREATE INDEX enumtest_hash ON enumtest USING hash (col);`,
			},
			{
				Statement: `SELECT * FROM enumtest WHERE col = 'orange';`,
				Results:   []sql.Row{{`orange`}},
			},
			{
				Statement: `DROP INDEX enumtest_hash;`,
			},
			{
				Statement: `RESET enable_seqscan;`,
			},
			{
				Statement: `RESET enable_bitmapscan;`,
			},
			{
				Statement: `CREATE DOMAIN rgb AS rainbow CHECK (VALUE IN ('red', 'green', 'blue'));`,
			},
			{
				Statement: `SELECT 'red'::rgb;`,
				Results:   []sql.Row{{`red`}},
			},
			{
				Statement:   `SELECT 'purple'::rgb;`,
				ErrorString: `value for domain rgb violates check constraint "rgb_check"`,
			},
			{
				Statement:   `SELECT 'purple'::rainbow::rgb;`,
				ErrorString: `value for domain rgb violates check constraint "rgb_check"`,
			},
			{
				Statement: `DROP DOMAIN rgb;`,
			},
			{
				Statement: `SELECT '{red,green,blue}'::rainbow[];`,
				Results:   []sql.Row{{`{red,green,blue}`}},
			},
			{
				Statement: `SELECT ('{red,green,blue}'::rainbow[])[2];`,
				Results:   []sql.Row{{`green`}},
			},
			{
				Statement: `SELECT 'red' = ANY ('{red,green,blue}'::rainbow[]);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT 'yellow' = ANY ('{red,green,blue}'::rainbow[]);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT 'red' = ALL ('{red,green,blue}'::rainbow[]);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT 'red' = ALL ('{red,red}'::rainbow[]);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT enum_first(NULL::rainbow);`,
				Results:   []sql.Row{{`red`}},
			},
			{
				Statement: `SELECT enum_last('green'::rainbow);`,
				Results:   []sql.Row{{`purple`}},
			},
			{
				Statement: `SELECT enum_range(NULL::rainbow);`,
				Results:   []sql.Row{{`{red,orange,yellow,green,blue,purple}`}},
			},
			{
				Statement: `SELECT enum_range('orange'::rainbow, 'green'::rainbow);`,
				Results:   []sql.Row{{`{orange,yellow,green}`}},
			},
			{
				Statement: `SELECT enum_range(NULL, 'green'::rainbow);`,
				Results:   []sql.Row{{`{red,orange,yellow,green}`}},
			},
			{
				Statement: `SELECT enum_range('orange'::rainbow, NULL);`,
				Results:   []sql.Row{{`{orange,yellow,green,blue,purple}`}},
			},
			{
				Statement: `SELECT enum_range(NULL::rainbow, NULL);`,
				Results:   []sql.Row{{`{red,orange,yellow,green,blue,purple}`}},
			},
			{
				Statement: `CREATE FUNCTION echo_me(anyenum) RETURNS text AS $$
BEGIN
RETURN $1::text || 'omg';`,
			},
			{
				Statement: `END
$$ LANGUAGE plpgsql;`,
			},
			{
				Statement: `SELECT echo_me('red'::rainbow);`,
				Results:   []sql.Row{{`redomg`}},
			},
			{
				Statement: `CREATE FUNCTION echo_me(rainbow) RETURNS text AS $$
BEGIN
RETURN $1::text || 'wtf';`,
			},
			{
				Statement: `END
$$ LANGUAGE plpgsql;`,
			},
			{
				Statement: `SELECT echo_me('red'::rainbow);`,
				Results:   []sql.Row{{`redwtf`}},
			},
			{
				Statement: `DROP FUNCTION echo_me(anyenum);`,
			},
			{
				Statement: `SELECT echo_me('red');`,
				Results:   []sql.Row{{`redwtf`}},
			},
			{
				Statement: `DROP FUNCTION echo_me(rainbow);`,
			},
			{
				Statement: `CREATE TABLE enumtest_parent (id rainbow PRIMARY KEY);`,
			},
			{
				Statement: `CREATE TABLE enumtest_child (parent rainbow REFERENCES enumtest_parent);`,
			},
			{
				Statement: `INSERT INTO enumtest_parent VALUES ('red');`,
			},
			{
				Statement: `INSERT INTO enumtest_child VALUES ('red');`,
			},
			{
				Statement:   `INSERT INTO enumtest_child VALUES ('blue');  -- fail`,
				ErrorString: `insert or update on table "enumtest_child" violates foreign key constraint "enumtest_child_parent_fkey"`,
			},
			{
				Statement:   `DELETE FROM enumtest_parent;  -- fail`,
				ErrorString: `update or delete on table "enumtest_parent" violates foreign key constraint "enumtest_child_parent_fkey" on table "enumtest_child"`,
			},
			{
				Statement: `CREATE TYPE bogus AS ENUM('good', 'bad', 'ugly');`,
			},
			{
				Statement:   `CREATE TABLE enumtest_bogus_child(parent bogus REFERENCES enumtest_parent);`,
				ErrorString: `foreign key constraint "enumtest_bogus_child_parent_fkey" cannot be implemented`,
			},
			{
				Statement: `DROP TYPE bogus;`,
			},
			{
				Statement: `ALTER TYPE rainbow RENAME VALUE 'red' TO 'crimson';`,
			},
			{
				Statement: `SELECT enumlabel, enumsortorder
FROM pg_enum
WHERE enumtypid = 'rainbow'::regtype
ORDER BY 2;`,
				Results: []sql.Row{{`crimson`, 1}, {`orange`, 2}, {`yellow`, 3}, {`green`, 4}, {`blue`, 5}, {`purple`, 6}},
			},
			{
				Statement:   `ALTER TYPE rainbow RENAME VALUE 'red' TO 'crimson';`,
				ErrorString: `"red" is not an existing enum label`,
			},
			{
				Statement:   `ALTER TYPE rainbow RENAME VALUE 'blue' TO 'green';`,
				ErrorString: `enum label "green" already exists`,
			},
			{
				Statement: `CREATE TYPE bogus AS ENUM('good');`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `ALTER TYPE bogus ADD VALUE 'new';`,
			},
			{
				Statement: `SAVEPOINT x;`,
			},
			{
				Statement:   `SELECT 'new'::bogus;  -- unsafe`,
				ErrorString: `unsafe use of new value "new" of enum type bogus`,
			},
			{
				Statement: `ROLLBACK TO x;`,
			},
			{
				Statement: `SELECT enum_first(null::bogus);  -- safe`,
				Results:   []sql.Row{{`good`}},
			},
			{
				Statement:   `SELECT enum_last(null::bogus);  -- unsafe`,
				ErrorString: `unsafe use of new value "new" of enum type bogus`,
			},
			{
				Statement: `ROLLBACK TO x;`,
			},
			{
				Statement:   `SELECT enum_range(null::bogus);  -- unsafe`,
				ErrorString: `unsafe use of new value "new" of enum type bogus`,
			},
			{
				Statement: `ROLLBACK TO x;`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `SELECT 'new'::bogus;  -- now safe`,
				Results:   []sql.Row{{`new`}},
			},
			{
				Statement: `SELECT enumlabel, enumsortorder
FROM pg_enum
WHERE enumtypid = 'bogus'::regtype
ORDER BY 2;`,
				Results: []sql.Row{{`good`, 1}, {`new`, 2}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `ALTER TYPE bogus RENAME TO bogon;`,
			},
			{
				Statement: `ALTER TYPE bogon ADD VALUE 'bad';`,
			},
			{
				Statement:   `SELECT 'bad'::bogon;`,
				ErrorString: `unsafe use of new value "bad" of enum type bogon`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `ALTER TYPE bogus RENAME VALUE 'good' to 'bad';`,
			},
			{
				Statement: `SELECT 'bad'::bogus;`,
				Results:   []sql.Row{{`bad`}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `DROP TYPE bogus;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `CREATE TYPE bogus AS ENUM('good','bad','ugly');`,
			},
			{
				Statement: `ALTER TYPE bogus RENAME TO bogon;`,
			},
			{
				Statement: `select enum_range(null::bogon);`,
				Results:   []sql.Row{{`{good,bad,ugly}`}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `CREATE TYPE bogus AS ENUM('good');`,
			},
			{
				Statement: `ALTER TYPE bogus RENAME TO bogon;`,
			},
			{
				Statement: `ALTER TYPE bogon ADD VALUE 'bad';`,
			},
			{
				Statement: `ALTER TYPE bogon ADD VALUE 'ugly';`,
			},
			{
				Statement:   `select enum_range(null::bogon);  -- fails`,
				ErrorString: `unsafe use of new value "bad" of enum type bogon`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `DROP TABLE enumtest_child;`,
			},
			{
				Statement: `DROP TABLE enumtest_parent;`,
			},
			{
				Statement: `DROP TABLE enumtest;`,
			},
			{
				Statement: `DROP TYPE rainbow;`,
			},
			{
				Statement: `SELECT COUNT(*) FROM pg_type WHERE typname = 'rainbow';`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT * FROM pg_enum WHERE NOT EXISTS
  (SELECT 1 FROM pg_type WHERE pg_type.oid = enumtypid);`,
				Results: []sql.Row{},
			},
		},
	})
}
