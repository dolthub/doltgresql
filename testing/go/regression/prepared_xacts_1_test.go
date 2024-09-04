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

func TestPreparedXacts1(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_prepared_xacts_1)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_prepared_xacts_1,
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
				Statement:   `PREPARE TRANSACTION 'foo1';`,
				ErrorString: `prepared transactions are disabled`,
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
				Statement:   `ROLLBACK PREPARED 'foo1';`,
				ErrorString: `prepared transaction with identifier "foo1" does not exist`,
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
				Statement:   `PREPARE TRANSACTION 'foo2';`,
				ErrorString: `prepared transactions are disabled`,
			},
			{
				Statement: `SELECT * FROM pxtest1;`,
				Results:   []sql.Row{{`aaa`}},
			},
			{
				Statement:   `COMMIT PREPARED 'foo2';`,
				ErrorString: `prepared transaction with identifier "foo2" does not exist`,
			},
			{
				Statement: `SELECT * FROM pxtest1;`,
				Results:   []sql.Row{{`aaa`}},
			},
			{
				Statement: `BEGIN TRANSACTION ISOLATION LEVEL SERIALIZABLE;`,
			},
			{
				Statement: `UPDATE pxtest1 SET foobar = 'eee' WHERE foobar = 'ddd';`,
			},
			{
				Statement: `SELECT * FROM pxtest1;`,
				Results:   []sql.Row{{`aaa`}},
			},
			{
				Statement:   `PREPARE TRANSACTION 'foo3';`,
				ErrorString: `prepared transactions are disabled`,
			},
			{
				Statement: `SELECT gid FROM pg_prepared_xacts;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `BEGIN TRANSACTION ISOLATION LEVEL SERIALIZABLE;`,
			},
			{
				Statement: `INSERT INTO pxtest1 VALUES ('fff');`,
			},
			{
				Statement:   `PREPARE TRANSACTION 'foo3';`,
				ErrorString: `prepared transactions are disabled`,
			},
			{
				Statement: `SELECT * FROM pxtest1;`,
				Results:   []sql.Row{{`aaa`}},
			},
			{
				Statement:   `ROLLBACK PREPARED 'foo3';`,
				ErrorString: `prepared transaction with identifier "foo3" does not exist`,
			},
			{
				Statement: `SELECT * FROM pxtest1;`,
				Results:   []sql.Row{{`aaa`}},
			},
			{
				Statement: `BEGIN TRANSACTION ISOLATION LEVEL SERIALIZABLE;`,
			},
			{
				Statement: `UPDATE pxtest1 SET foobar = 'eee' WHERE foobar = 'ddd';`,
			},
			{
				Statement: `SELECT * FROM pxtest1;`,
				Results:   []sql.Row{{`aaa`}},
			},
			{
				Statement:   `PREPARE TRANSACTION 'foo4';`,
				ErrorString: `prepared transactions are disabled`,
			},
			{
				Statement: `SELECT gid FROM pg_prepared_xacts;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `BEGIN TRANSACTION ISOLATION LEVEL SERIALIZABLE;`,
			},
			{
				Statement: `SELECT * FROM pxtest1;`,
				Results:   []sql.Row{{`aaa`}},
			},
			{
				Statement: `INSERT INTO pxtest1 VALUES ('fff');`,
			},
			{
				Statement:   `PREPARE TRANSACTION 'foo5';`,
				ErrorString: `prepared transactions are disabled`,
			},
			{
				Statement: `SELECT gid FROM pg_prepared_xacts;`,
				Results:   []sql.Row{},
			},
			{
				Statement:   `ROLLBACK PREPARED 'foo4';`,
				ErrorString: `prepared transaction with identifier "foo4" does not exist`,
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
				ErrorString: `prepared transactions are disabled`,
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
				Statement:   `PREPARE TRANSACTION 'regress-one';`,
				ErrorString: `prepared transactions are disabled`,
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
				Statement:   `PREPARE TRANSACTION 'regress-two';`,
				ErrorString: `prepared transactions are disabled`,
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
				Results:   []sql.Row{},
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `lock table pxtest3 in access share mode nowait;`,
			},
			{
				Statement: `rollback;`,
			},
			{
				Statement: `\c -
SELECT gid FROM pg_prepared_xacts;`,
				Results: []sql.Row{},
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `lock table pxtest3 in access share mode nowait;`,
			},
			{
				Statement: `rollback;`,
			},
			{
				Statement:   `COMMIT PREPARED 'regress-one';`,
				ErrorString: `prepared transaction with identifier "regress-one" does not exist`,
			},
			{
				Statement: `\d pxtest2
SELECT * FROM pxtest2;`,
				ErrorString: `relation "pxtest2" does not exist`,
			},
			{
				Statement: `SELECT gid FROM pg_prepared_xacts;`,
				Results:   []sql.Row{},
			},
			{
				Statement:   `COMMIT PREPARED 'regress-two';`,
				ErrorString: `prepared transaction with identifier "regress-two" does not exist`,
			},
			{
				Statement: `SELECT * FROM pxtest3;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT gid FROM pg_prepared_xacts;`,
				Results:   []sql.Row{},
			},
			{
				Statement:   `DROP TABLE pxtest2;`,
				ErrorString: `table "pxtest2" does not exist`,
			},
			{
				Statement: `DROP TABLE pxtest3;  -- will still be there if prepared xacts are disabled`,
			},
			{
				Statement:   `DROP TABLE pxtest4;`,
				ErrorString: `table "pxtest4" does not exist`,
			},
		},
	})
}
