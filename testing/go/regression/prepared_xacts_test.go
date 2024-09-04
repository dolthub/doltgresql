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

func TestPreparedXacts(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_prepared_xacts)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_prepared_xacts,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TABLE pxtest1 (foobar VARCHAR(10));`,
			},
			{
				Statement: `INSERT INTO pxtest1 VALUES ('aaa');`,
			},
			{
				Statement: `BEGIN TRANSACTION ISOLATION LEVEL SERIALIZABLE;`,
			},
			{
				Statement: `UPDATE pxtest1 SET foobar = 'bbb' WHERE foobar = 'aaa';`,
			},
			{
				Statement: `SELECT * FROM pxtest1;`,
				Results:   []sql.Row{{`bbb`}},
			},
			{
				Statement: `PREPARE TRANSACTION 'foo1';`,
			},
			{
				Statement: `SELECT * FROM pxtest1;`,
				Results:   []sql.Row{{`aaa`}},
			},
			{
				Statement: `SELECT gid FROM pg_prepared_xacts;`,
				Results:   []sql.Row{{`foo1`}},
			},
			{
				Statement: `ROLLBACK PREPARED 'foo1';`,
			},
			{
				Statement: `SELECT * FROM pxtest1;`,
				Results:   []sql.Row{{`aaa`}},
			},
			{
				Statement: `SELECT gid FROM pg_prepared_xacts;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `BEGIN TRANSACTION ISOLATION LEVEL SERIALIZABLE;`,
			},
			{
				Statement: `INSERT INTO pxtest1 VALUES ('ddd');`,
			},
			{
				Statement: `SELECT * FROM pxtest1;`,
				Results:   []sql.Row{{`aaa`}, {`ddd`}},
			},
			{
				Statement: `PREPARE TRANSACTION 'foo2';`,
			},
			{
				Statement: `SELECT * FROM pxtest1;`,
				Results:   []sql.Row{{`aaa`}},
			},
			{
				Statement: `COMMIT PREPARED 'foo2';`,
			},
			{
				Statement: `SELECT * FROM pxtest1;`,
				Results:   []sql.Row{{`aaa`}, {`ddd`}},
			},
			{
				Statement: `BEGIN TRANSACTION ISOLATION LEVEL SERIALIZABLE;`,
			},
			{
				Statement: `UPDATE pxtest1 SET foobar = 'eee' WHERE foobar = 'ddd';`,
			},
			{
				Statement: `SELECT * FROM pxtest1;`,
				Results:   []sql.Row{{`aaa`}, {`eee`}},
			},
			{
				Statement: `PREPARE TRANSACTION 'foo3';`,
			},
			{
				Statement: `SELECT gid FROM pg_prepared_xacts;`,
				Results:   []sql.Row{{`foo3`}},
			},
			{
				Statement: `BEGIN TRANSACTION ISOLATION LEVEL SERIALIZABLE;`,
			},
			{
				Statement: `INSERT INTO pxtest1 VALUES ('fff');`,
			},
			{
				Statement:   `PREPARE TRANSACTION 'foo3';`,
				ErrorString: `transaction identifier "foo3" is already in use`,
			},
			{
				Statement: `SELECT * FROM pxtest1;`,
				Results:   []sql.Row{{`aaa`}, {`ddd`}},
			},
			{
				Statement: `ROLLBACK PREPARED 'foo3';`,
			},
			{
				Statement: `SELECT * FROM pxtest1;`,
				Results:   []sql.Row{{`aaa`}, {`ddd`}},
			},
			{
				Statement: `BEGIN TRANSACTION ISOLATION LEVEL SERIALIZABLE;`,
			},
			{
				Statement: `UPDATE pxtest1 SET foobar = 'eee' WHERE foobar = 'ddd';`,
			},
			{
				Statement: `SELECT * FROM pxtest1;`,
				Results:   []sql.Row{{`aaa`}, {`eee`}},
			},
			{
				Statement: `PREPARE TRANSACTION 'foo4';`,
			},
			{
				Statement: `SELECT gid FROM pg_prepared_xacts;`,
				Results:   []sql.Row{{`foo4`}},
			},
			{
				Statement: `BEGIN TRANSACTION ISOLATION LEVEL SERIALIZABLE;`,
			},
			{
				Statement: `SELECT * FROM pxtest1;`,
				Results:   []sql.Row{{`aaa`}, {`ddd`}},
			},
			{
				Statement:   `INSERT INTO pxtest1 VALUES ('fff');`,
				ErrorString: `could not serialize access due to read/write dependencies among transactions`,
			},
			{
				Statement: `PREPARE TRANSACTION 'foo5';`,
			},
			{
				Statement: `SELECT gid FROM pg_prepared_xacts;`,
				Results:   []sql.Row{{`foo4`}},
			},
			{
				Statement: `ROLLBACK PREPARED 'foo4';`,
			},
			{
				Statement: `SELECT gid FROM pg_prepared_xacts;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `DROP TABLE pxtest1;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SELECT pg_advisory_lock(1);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT pg_advisory_xact_lock_shared(1);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement:   `PREPARE TRANSACTION 'foo6';  -- fails`,
				ErrorString: `cannot PREPARE while holding both session-level and transaction-level locks on the same object`,
			},
			{
				Statement: `BEGIN TRANSACTION ISOLATION LEVEL SERIALIZABLE;`,
			},
			{
				Statement: `  CREATE TABLE pxtest2 (a int);`,
			},
			{
				Statement: `  INSERT INTO pxtest2 VALUES (1);`,
			},
			{
				Statement: `  SAVEPOINT a;`,
			},
			{
				Statement: `    INSERT INTO pxtest2 VALUES (2);`,
			},
			{
				Statement: `  ROLLBACK TO a;`,
			},
			{
				Statement: `  SAVEPOINT b;`,
			},
			{
				Statement: `  INSERT INTO pxtest2 VALUES (3);`,
			},
			{
				Statement: `PREPARE TRANSACTION 'regress-one';`,
			},
			{
				Statement: `CREATE TABLE pxtest3(fff int);`,
			},
			{
				Statement: `BEGIN TRANSACTION ISOLATION LEVEL SERIALIZABLE;`,
			},
			{
				Statement: `  DROP TABLE pxtest3;`,
			},
			{
				Statement: `  CREATE TABLE pxtest4 (a int);`,
			},
			{
				Statement: `  INSERT INTO pxtest4 VALUES (1);`,
			},
			{
				Statement: `  INSERT INTO pxtest4 VALUES (2);`,
			},
			{
				Statement: `  DECLARE foo CURSOR FOR SELECT * FROM pxtest4;`,
			},
			{
				Statement: `  -- Fetch 1 tuple, keeping the cursor open
  FETCH 1 FROM foo;`,
				Results: []sql.Row{{1}},
			},
			{
				Statement: `PREPARE TRANSACTION 'regress-two';`,
			},
			{
				Statement:   `FETCH 1 FROM foo;`,
				ErrorString: `cursor "foo" does not exist`,
			},
			{
				Statement:   `SELECT * FROM pxtest2;`,
				ErrorString: `relation "pxtest2" does not exist`,
			},
			{
				Statement: `SELECT gid FROM pg_prepared_xacts;`,
				Results:   []sql.Row{{`regress-one`}, {`regress-two`}},
			},
			{
				Statement: `begin;`,
			},
			{
				Statement:   `lock table pxtest3 in access share mode nowait;`,
				ErrorString: `could not obtain lock on relation "pxtest3"`,
			},
			{
				Statement: `rollback;`,
			},
			{
				Statement: `\c -
SELECT gid FROM pg_prepared_xacts;`,
				Results: []sql.Row{{`regress-one`}, {`regress-two`}},
			},
			{
				Statement: `begin;`,
			},
			{
				Statement:   `lock table pxtest3 in access share mode nowait;`,
				ErrorString: `could not obtain lock on relation "pxtest3"`,
			},
			{
				Statement: `rollback;`,
			},
			{
				Statement: `COMMIT PREPARED 'regress-one';`,
			},
			{
				Statement: `\d pxtest2
              Table "public.pxtest2"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
SELECT * FROM pxtest2;`,
				Results: []sql.Row{{1}, {3}},
			},
			{
				Statement: `SELECT gid FROM pg_prepared_xacts;`,
				Results:   []sql.Row{{`regress-two`}},
			},
			{
				Statement: `COMMIT PREPARED 'regress-two';`,
			},
			{
				Statement:   `SELECT * FROM pxtest3;`,
				ErrorString: `relation "pxtest3" does not exist`,
			},
			{
				Statement: `SELECT gid FROM pg_prepared_xacts;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `DROP TABLE pxtest2;`,
			},
			{
				Statement:   `DROP TABLE pxtest3;  -- will still be there if prepared xacts are disabled`,
				ErrorString: `table "pxtest3" does not exist`,
			},
			{
				Statement: `DROP TABLE pxtest4;`,
			},
		},
	})
}
