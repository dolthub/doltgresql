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

func TestForeignKey(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_foreign_key)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_foreign_key,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TABLE PKTABLE ( ptest1 int PRIMARY KEY, ptest2 text );`,
			},
			{
				Statement: `CREATE TABLE FKTABLE ( ftest1 int REFERENCES PKTABLE MATCH FULL ON DELETE CASCADE ON UPDATE CASCADE, ftest2 int );`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES (1, 'Test1');`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES (2, 'Test2');`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES (3, 'Test3');`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES (4, 'Test4');`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES (5, 'Test5');`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES (1, 2);`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES (2, 3);`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES (3, 4);`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES (NULL, 1);`,
			},
			{
				Statement:   `INSERT INTO FKTABLE VALUES (100, 2);`,
				ErrorString: `insert or update on table "fktable" violates foreign key constraint "fktable_ftest1_fkey"`,
			},
			{
				Statement: `DETAIL:  Key (ftest1)=(100) is not present in table "pktable".
SELECT * FROM FKTABLE;`,
				Results: []sql.Row{{1, 2}, {2, 3}, {3, 4}, {``, 1}},
			},
			{
				Statement: `DELETE FROM PKTABLE WHERE ptest1=1;`,
			},
			{
				Statement: `SELECT * FROM FKTABLE;`,
				Results:   []sql.Row{{2, 3}, {3, 4}, {``, 1}},
			},
			{
				Statement: `UPDATE PKTABLE SET ptest1=1 WHERE ptest1=2;`,
			},
			{
				Statement: `SELECT * FROM FKTABLE;`,
				Results:   []sql.Row{{3, 4}, {``, 1}, {1, 3}},
			},
			{
				Statement: `DROP TABLE FKTABLE;`,
			},
			{
				Statement: `DROP TABLE PKTABLE;`,
			},
			{
				Statement: `CREATE TABLE PKTABLE ( ptest1 int, ptest2 int, ptest3 text, PRIMARY KEY(ptest1, ptest2) );`,
			},
			{
				Statement: `CREATE TABLE FKTABLE ( ftest1 int, ftest2 int, ftest3 int, CONSTRAINT constrname FOREIGN KEY(ftest1, ftest2)
                       REFERENCES PKTABLE MATCH FULL ON DELETE SET NULL ON UPDATE SET NULL);`,
			},
			{
				Statement:   `COMMENT ON CONSTRAINT constrname_wrong ON FKTABLE IS 'fk constraint comment';`,
				ErrorString: `constraint "constrname_wrong" for table "fktable" does not exist`,
			},
			{
				Statement: `COMMENT ON CONSTRAINT constrname ON FKTABLE IS 'fk constraint comment';`,
			},
			{
				Statement: `COMMENT ON CONSTRAINT constrname ON FKTABLE IS NULL;`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES (1, 2, 'Test1');`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES (1, 3, 'Test1-2');`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES (2, 4, 'Test2');`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES (3, 6, 'Test3');`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES (4, 8, 'Test4');`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES (5, 10, 'Test5');`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES (1, 2, 4);`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES (1, 3, 5);`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES (2, 4, 8);`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES (3, 6, 12);`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES (NULL, NULL, 0);`,
			},
			{
				Statement:   `INSERT INTO FKTABLE VALUES (100, 2, 4);`,
				ErrorString: `insert or update on table "fktable" violates foreign key constraint "constrname"`,
			},
			{
				Statement: `DETAIL:  Key (ftest1, ftest2)=(100, 2) is not present in table "pktable".
INSERT INTO FKTABLE VALUES (2, 2, 4);`,
				ErrorString: `insert or update on table "fktable" violates foreign key constraint "constrname"`,
			},
			{
				Statement: `DETAIL:  Key (ftest1, ftest2)=(2, 2) is not present in table "pktable".
INSERT INTO FKTABLE VALUES (NULL, 2, 4);`,
				ErrorString: `insert or update on table "fktable" violates foreign key constraint "constrname"`,
			},
			{
				Statement: `DETAIL:  MATCH FULL does not allow mixing of null and nonnull key values.
INSERT INTO FKTABLE VALUES (1, NULL, 4);`,
				ErrorString: `insert or update on table "fktable" violates foreign key constraint "constrname"`,
			},
			{
				Statement: `DETAIL:  MATCH FULL does not allow mixing of null and nonnull key values.
SELECT * FROM FKTABLE;`,
				Results: []sql.Row{{1, 2, 4}, {1, 3, 5}, {2, 4, 8}, {3, 6, 12}, {``, ``, 0}},
			},
			{
				Statement: `DELETE FROM PKTABLE WHERE ptest1=1 and ptest2=2;`,
			},
			{
				Statement: `SELECT * FROM FKTABLE;`,
				Results:   []sql.Row{{1, 3, 5}, {2, 4, 8}, {3, 6, 12}, {``, ``, 0}, {``, ``, 4}},
			},
			{
				Statement: `DELETE FROM PKTABLE WHERE ptest1=5 and ptest2=10;`,
			},
			{
				Statement: `SELECT * FROM FKTABLE;`,
				Results:   []sql.Row{{1, 3, 5}, {2, 4, 8}, {3, 6, 12}, {``, ``, 0}, {``, ``, 4}},
			},
			{
				Statement: `UPDATE PKTABLE SET ptest1=1 WHERE ptest1=2;`,
			},
			{
				Statement: `SELECT * FROM FKTABLE;`,
				Results:   []sql.Row{{1, 3, 5}, {3, 6, 12}, {``, ``, 0}, {``, ``, 4}, {``, ``, 8}},
			},
			{
				Statement:   `UPDATE FKTABLE SET ftest1 = NULL WHERE ftest1 = 1;`,
				ErrorString: `insert or update on table "fktable" violates foreign key constraint "constrname"`,
			},
			{
				Statement: `DETAIL:  MATCH FULL does not allow mixing of null and nonnull key values.
UPDATE FKTABLE SET ftest1 = 1 WHERE ftest1 = 1;`,
			},
			{
				Statement: `ALTER TABLE PKTABLE ALTER COLUMN ptest1 TYPE bigint;`,
			},
			{
				Statement: `ALTER TABLE FKTABLE ALTER COLUMN ftest1 TYPE bigint;`,
			},
			{
				Statement: `SELECT * FROM PKTABLE;`,
				Results:   []sql.Row{{1, 3, `Test1-2`}, {3, 6, `Test3`}, {4, 8, `Test4`}, {1, 4, `Test2`}},
			},
			{
				Statement: `SELECT * FROM FKTABLE;`,
				Results:   []sql.Row{{3, 6, 12}, {``, ``, 0}, {``, ``, 4}, {``, ``, 8}, {1, 3, 5}},
			},
			{
				Statement: `DROP TABLE PKTABLE CASCADE;`,
			},
			{
				Statement: `DROP TABLE FKTABLE;`,
			},
			{
				Statement: `CREATE TABLE PKTABLE ( ptest1 int, ptest2 int, ptest3 text, PRIMARY KEY(ptest1, ptest2) );`,
			},
			{
				Statement: `CREATE TABLE FKTABLE ( ftest1 int DEFAULT -1, ftest2 int DEFAULT -2, ftest3 int, CONSTRAINT constrname2 FOREIGN KEY(ftest1, ftest2)
                       REFERENCES PKTABLE MATCH FULL ON DELETE SET DEFAULT ON UPDATE SET DEFAULT);`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES (-1, -2, 'The Default!');`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES (1, 2, 'Test1');`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES (1, 3, 'Test1-2');`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES (2, 4, 'Test2');`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES (3, 6, 'Test3');`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES (4, 8, 'Test4');`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES (5, 10, 'Test5');`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES (1, 2, 4);`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES (1, 3, 5);`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES (2, 4, 8);`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES (3, 6, 12);`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES (NULL, NULL, 0);`,
			},
			{
				Statement:   `INSERT INTO FKTABLE VALUES (100, 2, 4);`,
				ErrorString: `insert or update on table "fktable" violates foreign key constraint "constrname2"`,
			},
			{
				Statement: `DETAIL:  Key (ftest1, ftest2)=(100, 2) is not present in table "pktable".
INSERT INTO FKTABLE VALUES (2, 2, 4);`,
				ErrorString: `insert or update on table "fktable" violates foreign key constraint "constrname2"`,
			},
			{
				Statement: `DETAIL:  Key (ftest1, ftest2)=(2, 2) is not present in table "pktable".
INSERT INTO FKTABLE VALUES (NULL, 2, 4);`,
				ErrorString: `insert or update on table "fktable" violates foreign key constraint "constrname2"`,
			},
			{
				Statement: `DETAIL:  MATCH FULL does not allow mixing of null and nonnull key values.
INSERT INTO FKTABLE VALUES (1, NULL, 4);`,
				ErrorString: `insert or update on table "fktable" violates foreign key constraint "constrname2"`,
			},
			{
				Statement: `DETAIL:  MATCH FULL does not allow mixing of null and nonnull key values.
SELECT * FROM FKTABLE;`,
				Results: []sql.Row{{1, 2, 4}, {1, 3, 5}, {2, 4, 8}, {3, 6, 12}, {``, ``, 0}},
			},
			{
				Statement: `DELETE FROM PKTABLE WHERE ptest1=1 and ptest2=2;`,
			},
			{
				Statement: `SELECT * FROM FKTABLE;`,
				Results:   []sql.Row{{1, 3, 5}, {2, 4, 8}, {3, 6, 12}, {``, ``, 0}, {-1, -2, 4}},
			},
			{
				Statement: `DELETE FROM PKTABLE WHERE ptest1=5 and ptest2=10;`,
			},
			{
				Statement: `SELECT * FROM FKTABLE;`,
				Results:   []sql.Row{{1, 3, 5}, {2, 4, 8}, {3, 6, 12}, {``, ``, 0}, {-1, -2, 4}},
			},
			{
				Statement: `UPDATE PKTABLE SET ptest1=1 WHERE ptest1=2;`,
			},
			{
				Statement: `SELECT * FROM FKTABLE;`,
				Results:   []sql.Row{{1, 3, 5}, {3, 6, 12}, {``, ``, 0}, {-1, -2, 4}, {-1, -2, 8}},
			},
			{
				Statement:   `DROP TABLE PKTABLE;`,
				ErrorString: `cannot drop table pktable because other objects depend on it`,
			},
			{
				Statement: `DETAIL:  constraint constrname2 on table fktable depends on table pktable
HINT:  Use DROP ... CASCADE to drop the dependent objects too.
DROP TABLE PKTABLE CASCADE;`,
			},
			{
				Statement: `DROP TABLE FKTABLE;`,
			},
			{
				Statement: `CREATE TABLE PKTABLE ( ptest1 int PRIMARY KEY, ptest2 text );`,
			},
			{
				Statement: `CREATE TABLE FKTABLE ( ftest1 int REFERENCES PKTABLE MATCH FULL, ftest2 int );`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES (1, 'Test1');`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES (2, 'Test2');`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES (3, 'Test3');`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES (4, 'Test4');`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES (5, 'Test5');`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES (1, 2);`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES (2, 3);`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES (3, 4);`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES (NULL, 1);`,
			},
			{
				Statement:   `INSERT INTO FKTABLE VALUES (100, 2);`,
				ErrorString: `insert or update on table "fktable" violates foreign key constraint "fktable_ftest1_fkey"`,
			},
			{
				Statement: `DETAIL:  Key (ftest1)=(100) is not present in table "pktable".
SELECT * FROM FKTABLE;`,
				Results: []sql.Row{{1, 2}, {2, 3}, {3, 4}, {``, 1}},
			},
			{
				Statement: `SELECT * FROM PKTABLE;`,
				Results:   []sql.Row{{1, `Test1`}, {2, `Test2`}, {3, `Test3`}, {4, `Test4`}, {5, `Test5`}},
			},
			{
				Statement:   `DELETE FROM PKTABLE WHERE ptest1=1;`,
				ErrorString: `update or delete on table "pktable" violates foreign key constraint "fktable_ftest1_fkey" on table "fktable"`,
			},
			{
				Statement: `DETAIL:  Key (ptest1)=(1) is still referenced from table "fktable".
DELETE FROM PKTABLE WHERE ptest1=5;`,
			},
			{
				Statement: `SELECT * FROM PKTABLE;`,
				Results:   []sql.Row{{1, `Test1`}, {2, `Test2`}, {3, `Test3`}, {4, `Test4`}},
			},
			{
				Statement:   `UPDATE PKTABLE SET ptest1=0 WHERE ptest1=2;`,
				ErrorString: `update or delete on table "pktable" violates foreign key constraint "fktable_ftest1_fkey" on table "fktable"`,
			},
			{
				Statement: `DETAIL:  Key (ptest1)=(2) is still referenced from table "fktable".
UPDATE PKTABLE SET ptest1=0 WHERE ptest1=4;`,
			},
			{
				Statement: `SELECT * FROM PKTABLE;`,
				Results:   []sql.Row{{1, `Test1`}, {2, `Test2`}, {3, `Test3`}, {0, `Test4`}},
			},
			{
				Statement: `DROP TABLE FKTABLE;`,
			},
			{
				Statement: `DROP TABLE PKTABLE;`,
			},
			{
				Statement: `CREATE TABLE PKTABLE ( ptest1 int, ptest2 int, PRIMARY KEY(ptest1, ptest2) );`,
			},
			{
				Statement: `CREATE TABLE FKTABLE ( ftest1 int, ftest2 int );`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES (1, 2);`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES (1, NULL);`,
			},
			{
				Statement:   `ALTER TABLE FKTABLE ADD FOREIGN KEY(ftest1, ftest2) REFERENCES PKTABLE MATCH FULL;`,
				ErrorString: `insert or update on table "fktable" violates foreign key constraint "fktable_ftest1_ftest2_fkey"`,
			},
			{
				Statement: `DETAIL:  MATCH FULL does not allow mixing of null and nonnull key values.
DROP TABLE FKTABLE;`,
			},
			{
				Statement: `DROP TABLE PKTABLE;`,
			},
			{
				Statement: `CREATE TABLE PKTABLE ( ptest1 int, ptest2 int, ptest3 int, ptest4 text, PRIMARY KEY(ptest1, ptest2, ptest3) );`,
			},
			{
				Statement: `CREATE TABLE FKTABLE ( ftest1 int, ftest2 int, ftest3 int, ftest4 int,  CONSTRAINT constrname3
			FOREIGN KEY(ftest1, ftest2, ftest3) REFERENCES PKTABLE);`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES (1, 2, 3, 'test1');`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES (1, 3, 3, 'test2');`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES (2, 3, 4, 'test3');`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES (2, 4, 5, 'test4');`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES (1, 2, 3, 1);`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES (NULL, 2, 3, 2);`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES (2, NULL, 3, 3);`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES (NULL, 2, 7, 4);`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES (NULL, 3, 4, 5);`,
			},
			{
				Statement:   `INSERT INTO FKTABLE VALUES (1, 2, 7, 6);`,
				ErrorString: `insert or update on table "fktable" violates foreign key constraint "constrname3"`,
			},
			{
				Statement: `DETAIL:  Key (ftest1, ftest2, ftest3)=(1, 2, 7) is not present in table "pktable".
SELECT * from FKTABLE;`,
				Results: []sql.Row{{1, 2, 3, 1}, {``, 2, 3, 2}, {2, ``, 3, 3}, {``, 2, 7, 4}, {``, 3, 4, 5}},
			},
			{
				Statement:   `UPDATE PKTABLE set ptest2=5 where ptest2=2;`,
				ErrorString: `update or delete on table "pktable" violates foreign key constraint "constrname3" on table "fktable"`,
			},
			{
				Statement: `DETAIL:  Key (ptest1, ptest2, ptest3)=(1, 2, 3) is still referenced from table "fktable".
UPDATE PKTABLE set ptest1=1 WHERE ptest2=3;`,
			},
			{
				Statement:   `DELETE FROM PKTABLE where ptest1=1 and ptest2=2 and ptest3=3;`,
				ErrorString: `update or delete on table "pktable" violates foreign key constraint "constrname3" on table "fktable"`,
			},
			{
				Statement: `DETAIL:  Key (ptest1, ptest2, ptest3)=(1, 2, 3) is still referenced from table "fktable".
DELETE FROM PKTABLE where ptest1=2;`,
			},
			{
				Statement: `SELECT * from PKTABLE;`,
				Results:   []sql.Row{{1, 2, 3, `test1`}, {1, 3, 3, `test2`}, {1, 3, 4, `test3`}},
			},
			{
				Statement: `SELECT * from FKTABLE;`,
				Results:   []sql.Row{{1, 2, 3, 1}, {``, 2, 3, 2}, {2, ``, 3, 3}, {``, 2, 7, 4}, {``, 3, 4, 5}},
			},
			{
				Statement: `DROP TABLE FKTABLE;`,
			},
			{
				Statement: `DROP TABLE PKTABLE;`,
			},
			{
				Statement: `CREATE TABLE PKTABLE ( ptest1 int, ptest2 int, ptest3 int, ptest4 text, UNIQUE(ptest1, ptest2, ptest3) );`,
			},
			{
				Statement: `CREATE TABLE FKTABLE ( ftest1 int, ftest2 int, ftest3 int, ftest4 int,  CONSTRAINT constrname3
			FOREIGN KEY(ftest1, ftest2, ftest3) REFERENCES PKTABLE (ptest1, ptest2, ptest3));`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES (1, 2, 3, 'test1');`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES (1, 3, NULL, 'test2');`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES (2, NULL, 4, 'test3');`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES (1, 2, 3, 1);`,
			},
			{
				Statement: `DELETE FROM PKTABLE WHERE ptest1 = 2;`,
			},
			{
				Statement: `SELECT * FROM PKTABLE;`,
				Results:   []sql.Row{{1, 2, 3, `test1`}, {1, 3, ``, `test2`}},
			},
			{
				Statement: `SELECT * FROM FKTABLE;`,
				Results:   []sql.Row{{1, 2, 3, 1}},
			},
			{
				Statement: `DROP TABLE FKTABLE;`,
			},
			{
				Statement: `DROP TABLE PKTABLE;`,
			},
			{
				Statement: `CREATE TABLE PKTABLE ( ptest1 int, ptest2 int, ptest3 int, ptest4 text, PRIMARY KEY(ptest1, ptest2, ptest3) );`,
			},
			{
				Statement: `CREATE TABLE FKTABLE ( ftest1 int, ftest2 int, ftest3 int, ftest4 int,  CONSTRAINT constrname3
			FOREIGN KEY(ftest1, ftest2, ftest3) REFERENCES PKTABLE
			ON DELETE CASCADE ON UPDATE CASCADE);`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES (1, 2, 3, 'test1');`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES (1, 3, 3, 'test2');`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES (2, 3, 4, 'test3');`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES (2, 4, 5, 'test4');`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES (1, 2, 3, 1);`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES (NULL, 2, 3, 2);`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES (2, NULL, 3, 3);`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES (NULL, 2, 7, 4);`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES (NULL, 3, 4, 5);`,
			},
			{
				Statement:   `INSERT INTO FKTABLE VALUES (1, 2, 7, 6);`,
				ErrorString: `insert or update on table "fktable" violates foreign key constraint "constrname3"`,
			},
			{
				Statement: `DETAIL:  Key (ftest1, ftest2, ftest3)=(1, 2, 7) is not present in table "pktable".
SELECT * from FKTABLE;`,
				Results: []sql.Row{{1, 2, 3, 1}, {``, 2, 3, 2}, {2, ``, 3, 3}, {``, 2, 7, 4}, {``, 3, 4, 5}},
			},
			{
				Statement: `UPDATE PKTABLE set ptest2=5 where ptest2=2;`,
			},
			{
				Statement: `UPDATE PKTABLE set ptest1=1 WHERE ptest2=3;`,
			},
			{
				Statement: `SELECT * from PKTABLE;`,
				Results:   []sql.Row{{2, 4, 5, `test4`}, {1, 5, 3, `test1`}, {1, 3, 3, `test2`}, {1, 3, 4, `test3`}},
			},
			{
				Statement: `SELECT * from FKTABLE;`,
				Results:   []sql.Row{{``, 2, 3, 2}, {2, ``, 3, 3}, {``, 2, 7, 4}, {``, 3, 4, 5}, {1, 5, 3, 1}},
			},
			{
				Statement: `DELETE FROM PKTABLE where ptest1=1 and ptest2=5 and ptest3=3;`,
			},
			{
				Statement: `SELECT * from PKTABLE;`,
				Results:   []sql.Row{{2, 4, 5, `test4`}, {1, 3, 3, `test2`}, {1, 3, 4, `test3`}},
			},
			{
				Statement: `SELECT * from FKTABLE;`,
				Results:   []sql.Row{{``, 2, 3, 2}, {2, ``, 3, 3}, {``, 2, 7, 4}, {``, 3, 4, 5}},
			},
			{
				Statement: `DELETE FROM PKTABLE where ptest1=2;`,
			},
			{
				Statement: `SELECT * from PKTABLE;`,
				Results:   []sql.Row{{1, 3, 3, `test2`}, {1, 3, 4, `test3`}},
			},
			{
				Statement: `SELECT * from FKTABLE;`,
				Results:   []sql.Row{{``, 2, 3, 2}, {2, ``, 3, 3}, {``, 2, 7, 4}, {``, 3, 4, 5}},
			},
			{
				Statement: `DROP TABLE FKTABLE;`,
			},
			{
				Statement: `DROP TABLE PKTABLE;`,
			},
			{
				Statement: `CREATE TABLE PKTABLE ( ptest1 int, ptest2 int, ptest3 int, ptest4 text, PRIMARY KEY(ptest1, ptest2, ptest3) );`,
			},
			{
				Statement: `CREATE TABLE FKTABLE ( ftest1 int DEFAULT 0, ftest2 int, ftest3 int, ftest4 int,  CONSTRAINT constrname3
			FOREIGN KEY(ftest1, ftest2, ftest3) REFERENCES PKTABLE
			ON DELETE SET DEFAULT ON UPDATE SET NULL);`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES (1, 2, 3, 'test1');`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES (1, 3, 3, 'test2');`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES (2, 3, 4, 'test3');`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES (2, 4, 5, 'test4');`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES (1, 2, 3, 1);`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES (2, 3, 4, 1);`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES (NULL, 2, 3, 2);`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES (2, NULL, 3, 3);`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES (NULL, 2, 7, 4);`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES (NULL, 3, 4, 5);`,
			},
			{
				Statement:   `INSERT INTO FKTABLE VALUES (1, 2, 7, 6);`,
				ErrorString: `insert or update on table "fktable" violates foreign key constraint "constrname3"`,
			},
			{
				Statement: `DETAIL:  Key (ftest1, ftest2, ftest3)=(1, 2, 7) is not present in table "pktable".
SELECT * from FKTABLE;`,
				Results: []sql.Row{{1, 2, 3, 1}, {2, 3, 4, 1}, {``, 2, 3, 2}, {2, ``, 3, 3}, {``, 2, 7, 4}, {``, 3, 4, 5}},
			},
			{
				Statement: `UPDATE PKTABLE set ptest2=5 where ptest2=2;`,
			},
			{
				Statement: `UPDATE PKTABLE set ptest2=2 WHERE ptest2=3 and ptest1=1;`,
			},
			{
				Statement: `SELECT * from PKTABLE;`,
				Results:   []sql.Row{{2, 3, 4, `test3`}, {2, 4, 5, `test4`}, {1, 5, 3, `test1`}, {1, 2, 3, `test2`}},
			},
			{
				Statement: `SELECT * from FKTABLE;`,
				Results:   []sql.Row{{2, 3, 4, 1}, {``, 2, 3, 2}, {2, ``, 3, 3}, {``, 2, 7, 4}, {``, 3, 4, 5}, {``, ``, ``, 1}},
			},
			{
				Statement: `DELETE FROM PKTABLE where ptest1=2 and ptest2=3 and ptest3=4;`,
			},
			{
				Statement: `SELECT * from PKTABLE;`,
				Results:   []sql.Row{{2, 4, 5, `test4`}, {1, 5, 3, `test1`}, {1, 2, 3, `test2`}},
			},
			{
				Statement: `SELECT * from FKTABLE;`,
				Results:   []sql.Row{{``, 2, 3, 2}, {2, ``, 3, 3}, {``, 2, 7, 4}, {``, 3, 4, 5}, {``, ``, ``, 1}, {0, ``, ``, 1}},
			},
			{
				Statement: `DELETE FROM PKTABLE where ptest2=5;`,
			},
			{
				Statement: `SELECT * from PKTABLE;`,
				Results:   []sql.Row{{2, 4, 5, `test4`}, {1, 2, 3, `test2`}},
			},
			{
				Statement: `SELECT * from FKTABLE;`,
				Results:   []sql.Row{{``, 2, 3, 2}, {2, ``, 3, 3}, {``, 2, 7, 4}, {``, 3, 4, 5}, {``, ``, ``, 1}, {0, ``, ``, 1}},
			},
			{
				Statement: `DROP TABLE FKTABLE;`,
			},
			{
				Statement: `DROP TABLE PKTABLE;`,
			},
			{
				Statement: `CREATE TABLE PKTABLE ( ptest1 int, ptest2 int, ptest3 int, ptest4 text, PRIMARY KEY(ptest1, ptest2, ptest3) );`,
			},
			{
				Statement: `CREATE TABLE FKTABLE ( ftest1 int DEFAULT 0, ftest2 int DEFAULT -1, ftest3 int DEFAULT -2, ftest4 int, CONSTRAINT constrname3
			FOREIGN KEY(ftest1, ftest2, ftest3) REFERENCES PKTABLE
			ON DELETE SET NULL ON UPDATE SET DEFAULT);`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES (1, 2, 3, 'test1');`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES (1, 3, 3, 'test2');`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES (2, 3, 4, 'test3');`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES (2, 4, 5, 'test4');`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES (2, -1, 5, 'test5');`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES (1, 2, 3, 1);`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES (2, 3, 4, 1);`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES (2, 4, 5, 1);`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES (NULL, 2, 3, 2);`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES (2, NULL, 3, 3);`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES (NULL, 2, 7, 4);`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES (NULL, 3, 4, 5);`,
			},
			{
				Statement:   `INSERT INTO FKTABLE VALUES (1, 2, 7, 6);`,
				ErrorString: `insert or update on table "fktable" violates foreign key constraint "constrname3"`,
			},
			{
				Statement: `DETAIL:  Key (ftest1, ftest2, ftest3)=(1, 2, 7) is not present in table "pktable".
SELECT * from FKTABLE;`,
				Results: []sql.Row{{1, 2, 3, 1}, {2, 3, 4, 1}, {2, 4, 5, 1}, {``, 2, 3, 2}, {2, ``, 3, 3}, {``, 2, 7, 4}, {``, 3, 4, 5}},
			},
			{
				Statement:   `UPDATE PKTABLE set ptest2=5 where ptest2=2;`,
				ErrorString: `insert or update on table "fktable" violates foreign key constraint "constrname3"`,
			},
			{
				Statement: `DETAIL:  Key (ftest1, ftest2, ftest3)=(0, -1, -2) is not present in table "pktable".
UPDATE PKTABLE set ptest1=0, ptest2=-1, ptest3=-2 where ptest2=2;`,
			},
			{
				Statement: `UPDATE PKTABLE set ptest2=10 where ptest2=4;`,
			},
			{
				Statement: `UPDATE PKTABLE set ptest2=2 WHERE ptest2=3 and ptest1=1;`,
			},
			{
				Statement: `SELECT * from PKTABLE;`,
				Results:   []sql.Row{{2, 3, 4, `test3`}, {2, -1, 5, `test5`}, {0, -1, -2, `test1`}, {2, 10, 5, `test4`}, {1, 2, 3, `test2`}},
			},
			{
				Statement: `SELECT * from FKTABLE;`,
				Results:   []sql.Row{{2, 3, 4, 1}, {``, 2, 3, 2}, {2, ``, 3, 3}, {``, 2, 7, 4}, {``, 3, 4, 5}, {0, -1, -2, 1}, {0, -1, -2, 1}},
			},
			{
				Statement: `DELETE FROM PKTABLE where ptest1=2 and ptest2=3 and ptest3=4;`,
			},
			{
				Statement: `SELECT * from PKTABLE;`,
				Results:   []sql.Row{{2, -1, 5, `test5`}, {0, -1, -2, `test1`}, {2, 10, 5, `test4`}, {1, 2, 3, `test2`}},
			},
			{
				Statement: `SELECT * from FKTABLE;`,
				Results:   []sql.Row{{``, 2, 3, 2}, {2, ``, 3, 3}, {``, 2, 7, 4}, {``, 3, 4, 5}, {0, -1, -2, 1}, {0, -1, -2, 1}, {``, ``, ``, 1}},
			},
			{
				Statement: `DELETE FROM PKTABLE where ptest2=-1 and ptest3=5;`,
			},
			{
				Statement: `SELECT * from PKTABLE;`,
				Results:   []sql.Row{{0, -1, -2, `test1`}, {2, 10, 5, `test4`}, {1, 2, 3, `test2`}},
			},
			{
				Statement: `SELECT * from FKTABLE;`,
				Results:   []sql.Row{{``, 2, 3, 2}, {2, ``, 3, 3}, {``, 2, 7, 4}, {``, 3, 4, 5}, {0, -1, -2, 1}, {0, -1, -2, 1}, {``, ``, ``, 1}},
			},
			{
				Statement: `DROP TABLE FKTABLE;`,
			},
			{
				Statement: `DROP TABLE PKTABLE;`,
			},
			{
				Statement: `CREATE TABLE PKTABLE (tid int, id int, PRIMARY KEY (tid, id));`,
			},
			{
				Statement:   `CREATE TABLE FKTABLE (tid int, id int, foo int, FOREIGN KEY (tid, id) REFERENCES PKTABLE ON DELETE SET NULL (bar));`,
				ErrorString: `column "bar" referenced in foreign key constraint does not exist`,
			},
			{
				Statement:   `CREATE TABLE FKTABLE (tid int, id int, foo int, FOREIGN KEY (tid, id) REFERENCES PKTABLE ON DELETE SET NULL (foo));`,
				ErrorString: `column "foo" referenced in ON DELETE SET action must be part of foreign key`,
			},
			{
				Statement:   `CREATE TABLE FKTABLE (tid int, id int, foo int, FOREIGN KEY (tid, foo) REFERENCES PKTABLE ON UPDATE SET NULL (foo));`,
				ErrorString: `a column list with SET NULL is only supported for ON DELETE actions`,
			},
			{
				Statement: `CREATE TABLE FKTABLE (
  tid int, id int,
  fk_id_del_set_null int,
  fk_id_del_set_default int DEFAULT 0,
  FOREIGN KEY (tid, fk_id_del_set_null) REFERENCES PKTABLE ON DELETE SET NULL (fk_id_del_set_null),
  FOREIGN KEY (tid, fk_id_del_set_default) REFERENCES PKTABLE ON DELETE SET DEFAULT (fk_id_del_set_default)
);`,
			},
			{
				Statement: `SELECT pg_get_constraintdef(oid) FROM pg_constraint WHERE conrelid = 'fktable'::regclass::oid ORDER BY oid;`,
				Results:   []sql.Row{{`FOREIGN KEY (tid, fk_id_del_set_null) REFERENCES pktable(tid, id) ON DELETE SET NULL (fk_id_del_set_null)`}, {`FOREIGN KEY (tid, fk_id_del_set_default) REFERENCES pktable(tid, id) ON DELETE SET DEFAULT (fk_id_del_set_default)`}},
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES (1, 0), (1, 1), (1, 2);`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES
  (1, 1, 1, NULL),
  (1, 2, NULL, 2);`,
			},
			{
				Statement: `DELETE FROM PKTABLE WHERE id = 1 OR id = 2;`,
			},
			{
				Statement: `SELECT * FROM FKTABLE ORDER BY id;`,
				Results:   []sql.Row{{1, 1, ``, ``}, {1, 2, ``, 0}},
			},
			{
				Statement: `DROP TABLE FKTABLE;`,
			},
			{
				Statement: `DROP TABLE PKTABLE;`,
			},
			{
				Statement: `CREATE TABLE PKTABLE (ptest1 int PRIMARY KEY, someoid oid);`,
			},
			{
				Statement:   `CREATE TABLE FKTABLE_FAIL1 ( ftest1 int, CONSTRAINT fkfail1 FOREIGN KEY (ftest2) REFERENCES PKTABLE);`,
				ErrorString: `column "ftest2" referenced in foreign key constraint does not exist`,
			},
			{
				Statement:   `CREATE TABLE FKTABLE_FAIL2 ( ftest1 int, CONSTRAINT fkfail1 FOREIGN KEY (ftest1) REFERENCES PKTABLE(ptest2));`,
				ErrorString: `column "ptest2" referenced in foreign key constraint does not exist`,
			},
			{
				Statement:   `CREATE TABLE FKTABLE_FAIL3 ( ftest1 int, CONSTRAINT fkfail1 FOREIGN KEY (tableoid) REFERENCES PKTABLE(someoid));`,
				ErrorString: `system columns cannot be used in foreign keys`,
			},
			{
				Statement:   `CREATE TABLE FKTABLE_FAIL4 ( ftest1 oid, CONSTRAINT fkfail1 FOREIGN KEY (ftest1) REFERENCES PKTABLE(tableoid));`,
				ErrorString: `system columns cannot be used in foreign keys`,
			},
			{
				Statement: `DROP TABLE PKTABLE;`,
			},
			{
				Statement: `CREATE TABLE PKTABLE (ptest1 int, ptest2 int, UNIQUE(ptest1, ptest2));`,
			},
			{
				Statement:   `CREATE TABLE FKTABLE_FAIL1 (ftest1 int REFERENCES pktable(ptest1));`,
				ErrorString: `there is no unique constraint matching given keys for referenced table "pktable"`,
			},
			{
				Statement:   `DROP TABLE FKTABLE_FAIL1;`,
				ErrorString: `table "fktable_fail1" does not exist`,
			},
			{
				Statement: `DROP TABLE PKTABLE;`,
			},
			{
				Statement: `CREATE TABLE PKTABLE (ptest1 int PRIMARY KEY);`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES(42);`,
			},
			{
				Statement:   `CREATE TABLE FKTABLE (ftest1 inet REFERENCES pktable);`,
				ErrorString: `foreign key constraint "fktable_ftest1_fkey" cannot be implemented`,
			},
			{
				Statement: `DETAIL:  Key columns "ftest1" and "ptest1" are of incompatible types: inet and integer.
CREATE TABLE FKTABLE (ftest1 inet REFERENCES pktable(ptest1));`,
				ErrorString: `foreign key constraint "fktable_ftest1_fkey" cannot be implemented`,
			},
			{
				Statement: `DETAIL:  Key columns "ftest1" and "ptest1" are of incompatible types: inet and integer.
CREATE TABLE FKTABLE (ftest1 int8 REFERENCES pktable);`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES(42);		-- should succeed`,
			},
			{
				Statement:   `INSERT INTO FKTABLE VALUES(43);		-- should fail`,
				ErrorString: `insert or update on table "fktable" violates foreign key constraint "fktable_ftest1_fkey"`,
			},
			{
				Statement: `DETAIL:  Key (ftest1)=(43) is not present in table "pktable".
UPDATE FKTABLE SET ftest1 = ftest1;	-- should succeed`,
			},
			{
				Statement:   `UPDATE FKTABLE SET ftest1 = ftest1 + 1;	-- should fail`,
				ErrorString: `insert or update on table "fktable" violates foreign key constraint "fktable_ftest1_fkey"`,
			},
			{
				Statement: `DETAIL:  Key (ftest1)=(43) is not present in table "pktable".
DROP TABLE FKTABLE;`,
			},
			{
				Statement:   `CREATE TABLE FKTABLE (ftest1 numeric REFERENCES pktable);`,
				ErrorString: `foreign key constraint "fktable_ftest1_fkey" cannot be implemented`,
			},
			{
				Statement: `DETAIL:  Key columns "ftest1" and "ptest1" are of incompatible types: numeric and integer.
DROP TABLE PKTABLE;`,
			},
			{
				Statement: `CREATE TABLE PKTABLE (ptest1 numeric PRIMARY KEY);`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES(42);`,
			},
			{
				Statement: `CREATE TABLE FKTABLE (ftest1 int REFERENCES pktable);`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES(42);		-- should succeed`,
			},
			{
				Statement:   `INSERT INTO FKTABLE VALUES(43);		-- should fail`,
				ErrorString: `insert or update on table "fktable" violates foreign key constraint "fktable_ftest1_fkey"`,
			},
			{
				Statement: `DETAIL:  Key (ftest1)=(43) is not present in table "pktable".
UPDATE FKTABLE SET ftest1 = ftest1;	-- should succeed`,
			},
			{
				Statement:   `UPDATE FKTABLE SET ftest1 = ftest1 + 1;	-- should fail`,
				ErrorString: `insert or update on table "fktable" violates foreign key constraint "fktable_ftest1_fkey"`,
			},
			{
				Statement: `DETAIL:  Key (ftest1)=(43) is not present in table "pktable".
DROP TABLE FKTABLE;`,
			},
			{
				Statement: `DROP TABLE PKTABLE;`,
			},
			{
				Statement: `CREATE TABLE PKTABLE (ptest1 int, ptest2 inet, PRIMARY KEY(ptest1, ptest2));`,
			},
			{
				Statement:   `CREATE TABLE FKTABLE (ftest1 cidr, ftest2 timestamp, FOREIGN KEY(ftest1, ftest2) REFERENCES pktable);`,
				ErrorString: `foreign key constraint "fktable_ftest1_ftest2_fkey" cannot be implemented`,
			},
			{
				Statement: `DETAIL:  Key columns "ftest1" and "ptest1" are of incompatible types: cidr and integer.
CREATE TABLE FKTABLE (ftest1 cidr, ftest2 timestamp, FOREIGN KEY(ftest1, ftest2) REFERENCES pktable(ptest1, ptest2));`,
				ErrorString: `foreign key constraint "fktable_ftest1_ftest2_fkey" cannot be implemented`,
			},
			{
				Statement: `DETAIL:  Key columns "ftest1" and "ptest1" are of incompatible types: cidr and integer.
CREATE TABLE FKTABLE (ftest1 int, ftest2 inet, FOREIGN KEY(ftest2, ftest1) REFERENCES pktable);`,
				ErrorString: `foreign key constraint "fktable_ftest2_ftest1_fkey" cannot be implemented`,
			},
			{
				Statement: `DETAIL:  Key columns "ftest2" and "ptest1" are of incompatible types: inet and integer.
CREATE TABLE FKTABLE (ftest1 int, ftest2 inet, FOREIGN KEY(ftest2, ftest1) REFERENCES pktable(ptest1, ptest2));`,
				ErrorString: `foreign key constraint "fktable_ftest2_ftest1_fkey" cannot be implemented`,
			},
			{
				Statement: `DETAIL:  Key columns "ftest2" and "ptest1" are of incompatible types: inet and integer.
CREATE TABLE FKTABLE (ftest1 int, ftest2 inet, FOREIGN KEY(ftest1, ftest2) REFERENCES pktable(ptest2, ptest1));`,
				ErrorString: `foreign key constraint "fktable_ftest1_ftest2_fkey" cannot be implemented`,
			},
			{
				Statement: `DETAIL:  Key columns "ftest1" and "ptest2" are of incompatible types: integer and inet.
CREATE TABLE FKTABLE (ftest1 int, ftest2 inet, FOREIGN KEY(ftest2, ftest1) REFERENCES pktable(ptest2, ptest1));`,
			},
			{
				Statement: `DROP TABLE FKTABLE;`,
			},
			{
				Statement: `CREATE TABLE FKTABLE (ftest1 int, ftest2 inet, FOREIGN KEY(ftest1, ftest2) REFERENCES pktable(ptest1, ptest2));`,
			},
			{
				Statement: `DROP TABLE FKTABLE;`,
			},
			{
				Statement: `DROP TABLE PKTABLE;`,
			},
			{
				Statement: `CREATE TABLE PKTABLE (ptest1 int, ptest2 inet, ptest3 int, ptest4 inet, PRIMARY KEY(ptest1, ptest2), FOREIGN KEY(ptest3,
ptest4) REFERENCES pktable(ptest1, ptest2));`,
			},
			{
				Statement: `DROP TABLE PKTABLE;`,
			},
			{
				Statement: `CREATE TABLE PKTABLE (ptest1 int, ptest2 inet, ptest3 int, ptest4 inet, PRIMARY KEY(ptest1, ptest2), FOREIGN KEY(ptest3,
ptest4) REFERENCES pktable);`,
			},
			{
				Statement: `DROP TABLE PKTABLE;`,
			},
			{
				Statement: `CREATE TABLE PKTABLE (ptest1 int, ptest2 inet, ptest3 int, ptest4 inet, PRIMARY KEY(ptest1, ptest2), FOREIGN KEY(ptest3,
ptest4) REFERENCES pktable(ptest2, ptest1));`,
				ErrorString: `foreign key constraint "pktable_ptest3_ptest4_fkey" cannot be implemented`,
			},
			{
				Statement: `DETAIL:  Key columns "ptest3" and "ptest2" are of incompatible types: integer and inet.
CREATE TABLE PKTABLE (ptest1 int, ptest2 inet, ptest3 int, ptest4 inet, PRIMARY KEY(ptest1, ptest2), FOREIGN KEY(ptest4,
ptest3) REFERENCES pktable(ptest1, ptest2));`,
				ErrorString: `foreign key constraint "pktable_ptest4_ptest3_fkey" cannot be implemented`,
			},
			{
				Statement: `DETAIL:  Key columns "ptest4" and "ptest1" are of incompatible types: inet and integer.
CREATE TABLE PKTABLE (ptest1 int, ptest2 inet, ptest3 int, ptest4 inet, PRIMARY KEY(ptest1, ptest2), FOREIGN KEY(ptest4,
ptest3) REFERENCES pktable);`,
				ErrorString: `foreign key constraint "pktable_ptest4_ptest3_fkey" cannot be implemented`,
			},
			{
				Statement: `DETAIL:  Key columns "ptest4" and "ptest1" are of incompatible types: inet and integer.
create table pktable_base (base1 int not null);`,
			},
			{
				Statement: `create table pktable (ptest1 int, primary key(base1), unique(base1, ptest1)) inherits (pktable_base);`,
			},
			{
				Statement: `create table fktable (ftest1 int references pktable(base1));`,
			},
			{
				Statement: `insert into pktable(base1) values (1);`,
			},
			{
				Statement: `insert into pktable(base1) values (2);`,
			},
			{
				Statement:   `insert into fktable(ftest1) values (3);`,
				ErrorString: `insert or update on table "fktable" violates foreign key constraint "fktable_ftest1_fkey"`,
			},
			{
				Statement: `DETAIL:  Key (ftest1)=(3) is not present in table "pktable".
insert into pktable(base1) values (3);`,
			},
			{
				Statement: `insert into fktable(ftest1) values (3);`,
			},
			{
				Statement:   `delete from pktable where base1>2;`,
				ErrorString: `update or delete on table "pktable" violates foreign key constraint "fktable_ftest1_fkey" on table "fktable"`,
			},
			{
				Statement: `DETAIL:  Key (base1)=(3) is still referenced from table "fktable".
update pktable set base1=base1*4;`,
				ErrorString: `update or delete on table "pktable" violates foreign key constraint "fktable_ftest1_fkey" on table "fktable"`,
			},
			{
				Statement: `DETAIL:  Key (base1)=(3) is still referenced from table "fktable".
update pktable set base1=base1*4 where base1<3;`,
			},
			{
				Statement: `delete from pktable where base1>3;`,
			},
			{
				Statement: `drop table fktable;`,
			},
			{
				Statement: `delete from pktable;`,
			},
			{
				Statement: `create table fktable (ftest1 int, ftest2 int, foreign key(ftest1, ftest2) references pktable(base1, ptest1));`,
			},
			{
				Statement: `insert into pktable(base1, ptest1) values (1, 1);`,
			},
			{
				Statement: `insert into pktable(base1, ptest1) values (2, 2);`,
			},
			{
				Statement:   `insert into fktable(ftest1, ftest2) values (3, 1);`,
				ErrorString: `insert or update on table "fktable" violates foreign key constraint "fktable_ftest1_ftest2_fkey"`,
			},
			{
				Statement: `DETAIL:  Key (ftest1, ftest2)=(3, 1) is not present in table "pktable".
insert into pktable(base1,ptest1) values (3, 1);`,
			},
			{
				Statement: `insert into fktable(ftest1, ftest2) values (3, 1);`,
			},
			{
				Statement:   `delete from pktable where base1>2;`,
				ErrorString: `update or delete on table "pktable" violates foreign key constraint "fktable_ftest1_ftest2_fkey" on table "fktable"`,
			},
			{
				Statement: `DETAIL:  Key (base1, ptest1)=(3, 1) is still referenced from table "fktable".
update pktable set base1=base1*4;`,
				ErrorString: `update or delete on table "pktable" violates foreign key constraint "fktable_ftest1_ftest2_fkey" on table "fktable"`,
			},
			{
				Statement: `DETAIL:  Key (base1, ptest1)=(3, 1) is still referenced from table "fktable".
update pktable set base1=base1*4 where base1<3;`,
			},
			{
				Statement: `delete from pktable where base1>3;`,
			},
			{
				Statement: `drop table fktable;`,
			},
			{
				Statement: `drop table pktable;`,
			},
			{
				Statement: `drop table pktable_base;`,
			},
			{
				Statement: `create table pktable_base(base1 int not null, base2 int);`,
			},
			{
				Statement: `create table pktable(ptest1 int, ptest2 int, primary key(base1, ptest1), foreign key(base2, ptest2) references
                                             pktable(base1, ptest1)) inherits (pktable_base);`,
			},
			{
				Statement: `insert into pktable (base1, ptest1, base2, ptest2) values (1, 1, 1, 1);`,
			},
			{
				Statement: `insert into pktable (base1, ptest1, base2, ptest2) values (2, 1, 1, 1);`,
			},
			{
				Statement: `insert into pktable (base1, ptest1, base2, ptest2) values (2, 2, 2, 1);`,
			},
			{
				Statement: `insert into pktable (base1, ptest1, base2, ptest2) values (1, 3, 2, 2);`,
			},
			{
				Statement:   `insert into pktable (base1, ptest1, base2, ptest2) values (2, 3, 3, 2);`,
				ErrorString: `insert or update on table "pktable" violates foreign key constraint "pktable_base2_ptest2_fkey"`,
			},
			{
				Statement: `DETAIL:  Key (base2, ptest2)=(3, 2) is not present in table "pktable".
delete from pktable where base1=2;`,
				ErrorString: `update or delete on table "pktable" violates foreign key constraint "pktable_base2_ptest2_fkey" on table "pktable"`,
			},
			{
				Statement: `DETAIL:  Key (base1, ptest1)=(2, 2) is still referenced from table "pktable".
update pktable set base1=3 where base1=1;`,
				ErrorString: `update or delete on table "pktable" violates foreign key constraint "pktable_base2_ptest2_fkey" on table "pktable"`,
			},
			{
				Statement: `DETAIL:  Key (base1, ptest1)=(1, 1) is still referenced from table "pktable".
delete from pktable where base2=2;`,
			},
			{
				Statement: `delete from pktable where base1=2;`,
			},
			{
				Statement: `drop table pktable;`,
			},
			{
				Statement: `drop table pktable_base;`,
			},
			{
				Statement: `create table pktable_base(base1 int not null);`,
			},
			{
				Statement: `create table pktable(ptest1 inet, primary key(base1, ptest1)) inherits (pktable_base);`,
			},
			{
				Statement:   `create table fktable(ftest1 cidr, ftest2 int[], foreign key (ftest1, ftest2) references pktable);`,
				ErrorString: `foreign key constraint "fktable_ftest1_ftest2_fkey" cannot be implemented`,
			},
			{
				Statement: `DETAIL:  Key columns "ftest1" and "base1" are of incompatible types: cidr and integer.
create table fktable(ftest1 cidr, ftest2 int[], foreign key (ftest1, ftest2) references pktable(base1, ptest1));`,
				ErrorString: `foreign key constraint "fktable_ftest1_ftest2_fkey" cannot be implemented`,
			},
			{
				Statement: `DETAIL:  Key columns "ftest1" and "base1" are of incompatible types: cidr and integer.
create table fktable(ftest1 int, ftest2 inet, foreign key(ftest2, ftest1) references pktable);`,
				ErrorString: `foreign key constraint "fktable_ftest2_ftest1_fkey" cannot be implemented`,
			},
			{
				Statement: `DETAIL:  Key columns "ftest2" and "base1" are of incompatible types: inet and integer.
create table fktable(ftest1 int, ftest2 inet, foreign key(ftest2, ftest1) references pktable(base1, ptest1));`,
				ErrorString: `foreign key constraint "fktable_ftest2_ftest1_fkey" cannot be implemented`,
			},
			{
				Statement: `DETAIL:  Key columns "ftest2" and "base1" are of incompatible types: inet and integer.
create table fktable(ftest1 int, ftest2 inet, foreign key(ftest1, ftest2) references pktable(ptest1, base1));`,
				ErrorString: `foreign key constraint "fktable_ftest1_ftest2_fkey" cannot be implemented`,
			},
			{
				Statement: `DETAIL:  Key columns "ftest1" and "ptest1" are of incompatible types: integer and inet.
drop table pktable;`,
			},
			{
				Statement: `drop table pktable_base;`,
			},
			{
				Statement: `create table pktable_base(base1 int not null, base2 int);`,
			},
			{
				Statement: `create table pktable(ptest1 inet, ptest2 inet[], primary key(base1, ptest1), foreign key(base2, ptest2) references
                                             pktable(base1, ptest1)) inherits (pktable_base);`,
				ErrorString: `foreign key constraint "pktable_base2_ptest2_fkey" cannot be implemented`,
			},
			{
				Statement: `DETAIL:  Key columns "ptest2" and "ptest1" are of incompatible types: inet[] and inet.
create table pktable(ptest1 inet, ptest2 inet, primary key(base1, ptest1), foreign key(base2, ptest2) references
                                             pktable(ptest1, base1)) inherits (pktable_base);`,
				ErrorString: `foreign key constraint "pktable_base2_ptest2_fkey" cannot be implemented`,
			},
			{
				Statement: `DETAIL:  Key columns "base2" and "ptest1" are of incompatible types: integer and inet.
create table pktable(ptest1 inet, ptest2 inet, primary key(base1, ptest1), foreign key(ptest2, base2) references
                                             pktable(base1, ptest1)) inherits (pktable_base);`,
				ErrorString: `foreign key constraint "pktable_ptest2_base2_fkey" cannot be implemented`,
			},
			{
				Statement: `DETAIL:  Key columns "ptest2" and "base1" are of incompatible types: inet and integer.
create table pktable(ptest1 inet, ptest2 inet, primary key(base1, ptest1), foreign key(ptest2, base2) references
                                             pktable(base1, ptest1)) inherits (pktable_base);`,
				ErrorString: `foreign key constraint "pktable_ptest2_base2_fkey" cannot be implemented`,
			},
			{
				Statement: `DETAIL:  Key columns "ptest2" and "base1" are of incompatible types: inet and integer.
drop table pktable;`,
				ErrorString: `table "pktable" does not exist`,
			},
			{
				Statement: `drop table pktable_base;`,
			},
			{
				Statement: `CREATE TABLE pktable (
	id		INT4 PRIMARY KEY,
	other	INT4
);`,
			},
			{
				Statement: `CREATE TABLE fktable (
	id		INT4 PRIMARY KEY,
	fk		INT4 REFERENCES pktable DEFERRABLE
);`,
			},
			{
				Statement:   `INSERT INTO fktable VALUES (5, 10);`,
				ErrorString: `insert or update on table "fktable" violates foreign key constraint "fktable_fk_fkey"`,
			},
			{
				Statement: `DETAIL:  Key (fk)=(10) is not present in table "pktable".
BEGIN;`,
			},
			{
				Statement: `SET CONSTRAINTS ALL DEFERRED;`,
			},
			{
				Statement: `INSERT INTO fktable VALUES (10, 15);`,
			},
			{
				Statement: `INSERT INTO pktable VALUES (15, 0); -- make the FK insert valid`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `DROP TABLE fktable, pktable;`,
			},
			{
				Statement: `CREATE TABLE pktable (
	id		INT4 PRIMARY KEY,
	other	INT4
);`,
			},
			{
				Statement: `CREATE TABLE fktable (
	id		INT4 PRIMARY KEY,
	fk		INT4 REFERENCES pktable DEFERRABLE INITIALLY DEFERRED
);`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `INSERT INTO fktable VALUES (100, 200);`,
			},
			{
				Statement: `INSERT INTO pktable VALUES (200, 500); -- make the FK insert valid`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SET CONSTRAINTS ALL IMMEDIATE;`,
			},
			{
				Statement:   `INSERT INTO fktable VALUES (500, 1000);`,
				ErrorString: `insert or update on table "fktable" violates foreign key constraint "fktable_fk_fkey"`,
			},
			{
				Statement: `DETAIL:  Key (fk)=(1000) is not present in table "pktable".
COMMIT;`,
			},
			{
				Statement: `DROP TABLE fktable, pktable;`,
			},
			{
				Statement: `CREATE TABLE pktable (
	id		INT4 PRIMARY KEY,
	other	INT4
);`,
			},
			{
				Statement: `CREATE TABLE fktable (
	id		INT4 PRIMARY KEY,
	fk		INT4 REFERENCES pktable DEFERRABLE
);`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SET CONSTRAINTS ALL DEFERRED;`,
			},
			{
				Statement: `INSERT INTO fktable VALUES (1000, 2000);`,
			},
			{
				Statement:   `SET CONSTRAINTS ALL IMMEDIATE;`,
				ErrorString: `insert or update on table "fktable" violates foreign key constraint "fktable_fk_fkey"`,
			},
			{
				Statement: `DETAIL:  Key (fk)=(2000) is not present in table "pktable".
INSERT INTO pktable VALUES (2000, 3); -- too late`,
				ErrorString: `current transaction is aborted, commands ignored until end of transaction block`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `DROP TABLE fktable, pktable;`,
			},
			{
				Statement: `CREATE TABLE pktable (
	id		INT4 PRIMARY KEY,
	other	INT4
);`,
			},
			{
				Statement: `CREATE TABLE fktable (
	id		INT4 PRIMARY KEY,
	fk		INT4 REFERENCES pktable DEFERRABLE INITIALLY DEFERRED
);`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `INSERT INTO fktable VALUES (100, 200);`,
			},
			{
				Statement:   `COMMIT;`,
				ErrorString: `insert or update on table "fktable" violates foreign key constraint "fktable_fk_fkey"`,
			},
			{
				Statement: `DETAIL:  Key (fk)=(200) is not present in table "pktable".
DROP TABLE pktable, fktable;`,
			},
			{
				Statement: `CREATE TEMP TABLE pktable (
        id1     INT4 PRIMARY KEY,
        id2     VARCHAR(4) UNIQUE,
        id3     REAL UNIQUE,
        UNIQUE(id1, id2, id3)
);`,
			},
			{
				Statement: `CREATE TEMP TABLE fktable (
        x1      INT4 REFERENCES pktable(id1),
        x2      VARCHAR(4) REFERENCES pktable(id2),
        x3      REAL REFERENCES pktable(id3),
        x4      TEXT,
        x5      INT2
);`,
			},
			{
				Statement: `ALTER TABLE fktable ADD CONSTRAINT fk_2_3
FOREIGN KEY (x2) REFERENCES pktable(id3);`,
				ErrorString: `foreign key constraint "fk_2_3" cannot be implemented`,
			},
			{
				Statement: `DETAIL:  Key columns "x2" and "id3" are of incompatible types: character varying and real.
ALTER TABLE fktable ADD CONSTRAINT fk_2_1
FOREIGN KEY (x2) REFERENCES pktable(id1);`,
				ErrorString: `foreign key constraint "fk_2_1" cannot be implemented`,
			},
			{
				Statement: `DETAIL:  Key columns "x2" and "id1" are of incompatible types: character varying and integer.
ALTER TABLE fktable ADD CONSTRAINT fk_3_1
FOREIGN KEY (x3) REFERENCES pktable(id1);`,
				ErrorString: `foreign key constraint "fk_3_1" cannot be implemented`,
			},
			{
				Statement: `DETAIL:  Key columns "x3" and "id1" are of incompatible types: real and integer.
ALTER TABLE fktable ADD CONSTRAINT fk_1_2
FOREIGN KEY (x1) REFERENCES pktable(id2);`,
				ErrorString: `foreign key constraint "fk_1_2" cannot be implemented`,
			},
			{
				Statement: `DETAIL:  Key columns "x1" and "id2" are of incompatible types: integer and character varying.
ALTER TABLE fktable ADD CONSTRAINT fk_1_3
FOREIGN KEY (x1) REFERENCES pktable(id3);`,
			},
			{
				Statement: `ALTER TABLE fktable ADD CONSTRAINT fk_4_2
FOREIGN KEY (x4) REFERENCES pktable(id2);`,
			},
			{
				Statement: `ALTER TABLE fktable ADD CONSTRAINT fk_5_1
FOREIGN KEY (x5) REFERENCES pktable(id1);`,
			},
			{
				Statement: `ALTER TABLE fktable ADD CONSTRAINT fk_123_123
FOREIGN KEY (x1,x2,x3) REFERENCES pktable(id1,id2,id3);`,
			},
			{
				Statement: `ALTER TABLE fktable ADD CONSTRAINT fk_213_213
FOREIGN KEY (x2,x1,x3) REFERENCES pktable(id2,id1,id3);`,
			},
			{
				Statement: `ALTER TABLE fktable ADD CONSTRAINT fk_253_213
FOREIGN KEY (x2,x5,x3) REFERENCES pktable(id2,id1,id3);`,
			},
			{
				Statement: `ALTER TABLE fktable ADD CONSTRAINT fk_123_231
FOREIGN KEY (x1,x2,x3) REFERENCES pktable(id2,id3,id1);`,
				ErrorString: `foreign key constraint "fk_123_231" cannot be implemented`,
			},
			{
				Statement: `DETAIL:  Key columns "x1" and "id2" are of incompatible types: integer and character varying.
ALTER TABLE fktable ADD CONSTRAINT fk_241_132
FOREIGN KEY (x2,x4,x1) REFERENCES pktable(id1,id3,id2);`,
				ErrorString: `foreign key constraint "fk_241_132" cannot be implemented`,
			},
			{
				Statement: `DETAIL:  Key columns "x2" and "id1" are of incompatible types: character varying and integer.
DROP TABLE pktable, fktable;`,
			},
			{
				Statement: `CREATE TEMP TABLE pktable (
    id int primary key,
    other int
);`,
			},
			{
				Statement: `CREATE TEMP TABLE fktable (
    id int primary key,
    fk int references pktable deferrable initially deferred
);`,
			},
			{
				Statement: `INSERT INTO pktable VALUES (5, 10);`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `INSERT INTO fktable VALUES (0, 20);`,
			},
			{
				Statement: `UPDATE fktable SET id = id + 1;`,
			},
			{
				Statement:   `COMMIT;`,
				ErrorString: `insert or update on table "fktable" violates foreign key constraint "fktable_fk_fkey"`,
			},
			{
				Statement: `DETAIL:  Key (fk)=(20) is not present in table "pktable".
BEGIN;`,
			},
			{
				Statement: `INSERT INTO fktable VALUES (0, 20);`,
			},
			{
				Statement: `SAVEPOINT savept1;`,
			},
			{
				Statement: `UPDATE fktable SET id = id + 1;`,
			},
			{
				Statement:   `COMMIT;`,
				ErrorString: `insert or update on table "fktable" violates foreign key constraint "fktable_fk_fkey"`,
			},
			{
				Statement: `DETAIL:  Key (fk)=(20) is not present in table "pktable".
BEGIN;`,
			},
			{
				Statement: `SAVEPOINT savept1;`,
			},
			{
				Statement: `INSERT INTO fktable VALUES (0, 20);`,
			},
			{
				Statement: `RELEASE SAVEPOINT savept1;`,
			},
			{
				Statement: `UPDATE fktable SET id = id + 1;`,
			},
			{
				Statement:   `COMMIT;`,
				ErrorString: `insert or update on table "fktable" violates foreign key constraint "fktable_fk_fkey"`,
			},
			{
				Statement: `DETAIL:  Key (fk)=(20) is not present in table "pktable".
BEGIN;`,
			},
			{
				Statement: `INSERT INTO fktable VALUES (0, 20);`,
			},
			{
				Statement: `SAVEPOINT savept1;`,
			},
			{
				Statement: `UPDATE fktable SET id = id + 1;`,
			},
			{
				Statement: `ROLLBACK TO savept1;`,
			},
			{
				Statement:   `COMMIT;`,
				ErrorString: `insert or update on table "fktable" violates foreign key constraint "fktable_fk_fkey"`,
			},
			{
				Statement: `DETAIL:  Key (fk)=(20) is not present in table "pktable".
INSERT INTO fktable VALUES (1, 5);`,
			},
			{
				Statement: `ALTER TABLE fktable ALTER CONSTRAINT fktable_fk_fkey DEFERRABLE INITIALLY IMMEDIATE;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement:   `UPDATE pktable SET id = 10 WHERE id = 5;`,
				ErrorString: `update or delete on table "pktable" violates foreign key constraint "fktable_fk_fkey" on table "fktable"`,
			},
			{
				Statement: `DETAIL:  Key (id)=(5) is still referenced from table "fktable".
COMMIT;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement:   `INSERT INTO fktable VALUES (0, 20);`,
				ErrorString: `insert or update on table "fktable" violates foreign key constraint "fktable_fk_fkey"`,
			},
			{
				Statement: `DETAIL:  Key (fk)=(20) is not present in table "pktable".
COMMIT;`,
			},
			{
				Statement: `ALTER TABLE fktable ALTER CONSTRAINT fktable_fk_fkey NOT DEFERRABLE;`,
			},
			{
				Statement:   `ALTER TABLE fktable ALTER CONSTRAINT fktable_fk_fkey NOT DEFERRABLE INITIALLY DEFERRED;`,
				ErrorString: `constraint declared INITIALLY DEFERRED must be DEFERRABLE`,
			},
			{
				Statement: `CREATE TEMP TABLE users (
  id INT PRIMARY KEY,
  name VARCHAR NOT NULL
);`,
			},
			{
				Statement: `INSERT INTO users VALUES (1, 'Jozko');`,
			},
			{
				Statement: `INSERT INTO users VALUES (2, 'Ferko');`,
			},
			{
				Statement: `INSERT INTO users VALUES (3, 'Samko');`,
			},
			{
				Statement: `CREATE TEMP TABLE tasks (
  id INT PRIMARY KEY,
  owner INT REFERENCES users ON UPDATE CASCADE ON DELETE SET NULL,
  worker INT REFERENCES users ON UPDATE CASCADE ON DELETE SET NULL,
  checked_by INT REFERENCES users ON UPDATE CASCADE ON DELETE SET NULL
);`,
			},
			{
				Statement: `INSERT INTO tasks VALUES (1,1,NULL,NULL);`,
			},
			{
				Statement: `INSERT INTO tasks VALUES (2,2,2,NULL);`,
			},
			{
				Statement: `INSERT INTO tasks VALUES (3,3,3,3);`,
			},
			{
				Statement: `SELECT * FROM tasks;`,
				Results:   []sql.Row{{1, 1, ``, ``}, {2, 2, 2, ``}, {3, 3, 3, 3}},
			},
			{
				Statement: `UPDATE users SET id = 4 WHERE id = 3;`,
			},
			{
				Statement: `SELECT * FROM tasks;`,
				Results:   []sql.Row{{1, 1, ``, ``}, {2, 2, 2, ``}, {3, 4, 4, 4}},
			},
			{
				Statement: `DELETE FROM users WHERE id = 4;`,
			},
			{
				Statement: `SELECT * FROM tasks;`,
				Results:   []sql.Row{{1, 1, ``, ``}, {2, 2, 2, ``}, {3, ``, ``, ``}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `UPDATE tasks set id=id WHERE id=2;`,
			},
			{
				Statement: `SELECT * FROM tasks;`,
				Results:   []sql.Row{{1, 1, ``, ``}, {3, ``, ``, ``}, {2, 2, 2, ``}},
			},
			{
				Statement: `DELETE FROM users WHERE id = 2;`,
			},
			{
				Statement: `SELECT * FROM tasks;`,
				Results:   []sql.Row{{1, 1, ``, ``}, {3, ``, ``, ``}, {2, ``, ``, ``}},
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `create temp table selfref (
    a int primary key,
    b int,
    foreign key (b) references selfref (a)
        on update cascade on delete cascade
);`,
			},
			{
				Statement: `insert into selfref (a, b)
values
    (0, 0),
    (1, 1);`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `    update selfref set a = 123 where a = 0;`,
			},
			{
				Statement: `    select a, b from selfref;`,
				Results:   []sql.Row{{1, 1}, {123, 123}},
			},
			{
				Statement: `    update selfref set a = 456 where a = 123;`,
			},
			{
				Statement: `    select a, b from selfref;`,
				Results:   []sql.Row{{1, 1}, {456, 456}},
			},
			{
				Statement: `commit;`,
			},
			{
				Statement: `create temp table defp (f1 int primary key);`,
			},
			{
				Statement: `create temp table defc (f1 int default 0
                        references defp on delete set default);`,
			},
			{
				Statement: `insert into defp values (0), (1), (2);`,
			},
			{
				Statement: `insert into defc values (2);`,
			},
			{
				Statement: `select * from defc;`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `delete from defp where f1 = 2;`,
			},
			{
				Statement: `select * from defc;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement:   `delete from defp where f1 = 0; -- fail`,
				ErrorString: `update or delete on table "defp" violates foreign key constraint "defc_f1_fkey" on table "defc"`,
			},
			{
				Statement: `DETAIL:  Key (f1)=(0) is still referenced from table "defc".
alter table defc alter column f1 set default 1;`,
			},
			{
				Statement: `delete from defp where f1 = 0;`,
			},
			{
				Statement: `select * from defc;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement:   `delete from defp where f1 = 1; -- fail`,
				ErrorString: `update or delete on table "defp" violates foreign key constraint "defc_f1_fkey" on table "defc"`,
			},
			{
				Statement: `DETAIL:  Key (f1)=(1) is still referenced from table "defc".
create temp table pp (f1 int primary key);`,
			},
			{
				Statement: `create temp table cc (f1 int references pp on update no action on delete no action);`,
			},
			{
				Statement: `insert into pp values(12);`,
			},
			{
				Statement: `insert into pp values(11);`,
			},
			{
				Statement: `update pp set f1=f1+1;`,
			},
			{
				Statement: `insert into cc values(13);`,
			},
			{
				Statement: `update pp set f1=f1+1;`,
			},
			{
				Statement:   `update pp set f1=f1+1; -- fail`,
				ErrorString: `update or delete on table "pp" violates foreign key constraint "cc_f1_fkey" on table "cc"`,
			},
			{
				Statement: `DETAIL:  Key (f1)=(13) is still referenced from table "cc".
delete from pp where f1 = 13; -- fail`,
				ErrorString: `update or delete on table "pp" violates foreign key constraint "cc_f1_fkey" on table "cc"`,
			},
			{
				Statement: `DETAIL:  Key (f1)=(13) is still referenced from table "cc".
drop table pp, cc;`,
			},
			{
				Statement: `create temp table pp (f1 int primary key);`,
			},
			{
				Statement: `create temp table cc (f1 int references pp on update restrict on delete restrict);`,
			},
			{
				Statement: `insert into pp values(12);`,
			},
			{
				Statement: `insert into pp values(11);`,
			},
			{
				Statement: `update pp set f1=f1+1;`,
			},
			{
				Statement: `insert into cc values(13);`,
			},
			{
				Statement:   `update pp set f1=f1+1; -- fail`,
				ErrorString: `update or delete on table "pp" violates foreign key constraint "cc_f1_fkey" on table "cc"`,
			},
			{
				Statement: `DETAIL:  Key (f1)=(13) is still referenced from table "cc".
delete from pp where f1 = 13; -- fail`,
				ErrorString: `update or delete on table "pp" violates foreign key constraint "cc_f1_fkey" on table "cc"`,
			},
			{
				Statement: `DETAIL:  Key (f1)=(13) is still referenced from table "cc".
drop table pp, cc;`,
			},
			{
				Statement: `create temp table t1 (a integer primary key, b text);`,
			},
			{
				Statement: `create temp table t2 (a integer primary key, b integer references t1);`,
			},
			{
				Statement: `create rule r1 as on delete to t1 do delete from t2 where t2.b = old.a;`,
			},
			{
				Statement: `explain (costs off) delete from t1 where a = 1;`,
				Results:   []sql.Row{{`Delete on t2`}, {`->  Nested Loop`}, {`->  Index Scan using t1_pkey on t1`}, {`Index Cond: (a = 1)`}, {`->  Seq Scan on t2`}, {`Filter: (b = 1)`}, {``}, {`Delete on t1`}, {`->  Index Scan using t1_pkey on t1`}, {`Index Cond: (a = 1)`}},
			},
			{
				Statement: `delete from t1 where a = 1;`,
			},
			{
				Statement: `create table pktable2 (a int, b int, c int, d int, e int, primary key (d, e));`,
			},
			{
				Statement: `create table fktable2 (d int, e int, foreign key (d, e) references pktable2);`,
			},
			{
				Statement: `insert into pktable2 values (1, 2, 3, 4, 5);`,
			},
			{
				Statement: `insert into fktable2 values (4, 5);`,
			},
			{
				Statement:   `delete from pktable2;`,
				ErrorString: `update or delete on table "pktable2" violates foreign key constraint "fktable2_d_e_fkey" on table "fktable2"`,
			},
			{
				Statement: `DETAIL:  Key (d, e)=(4, 5) is still referenced from table "fktable2".
update pktable2 set d = 5;`,
				ErrorString: `update or delete on table "pktable2" violates foreign key constraint "fktable2_d_e_fkey" on table "fktable2"`,
			},
			{
				Statement: `DETAIL:  Key (d, e)=(4, 5) is still referenced from table "fktable2".
drop table pktable2, fktable2;`,
			},
			{
				Statement: `create table pktable1 (a int primary key);`,
			},
			{
				Statement: `create table pktable2 (a int, b int, primary key (a, b));`,
			},
			{
				Statement: `create table fktable2 (
  a int,
  b int,
  very_very_long_column_name_to_exceed_63_characters int,
  foreign key (very_very_long_column_name_to_exceed_63_characters) references pktable1,
  foreign key (a, very_very_long_column_name_to_exceed_63_characters) references pktable2,
  foreign key (a, very_very_long_column_name_to_exceed_63_characters) references pktable2
);`,
			},
			{
				Statement: `select conname from pg_constraint where conrelid = 'fktable2'::regclass order by conname;`,
				Results:   []sql.Row{{`fktable2_a_very_very_long_column_name_to_exceed_63_charac_fkey1`}, {`fktable2_a_very_very_long_column_name_to_exceed_63_charact_fkey`}, {`fktable2_very_very_long_column_name_to_exceed_63_character_fkey`}},
			},
			{
				Statement: `drop table pktable1, pktable2, fktable2;`,
			},
			{
				Statement: `create table pktable2(f1 int primary key);`,
			},
			{
				Statement: `create table fktable2(f1 int references pktable2 deferrable initially deferred);`,
			},
			{
				Statement: `insert into pktable2 values(1);`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `insert into fktable2 values(1);`,
			},
			{
				Statement: `savepoint x;`,
			},
			{
				Statement: `delete from fktable2;`,
			},
			{
				Statement: `rollback to x;`,
			},
			{
				Statement: `commit;`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `insert into fktable2 values(2);`,
			},
			{
				Statement: `savepoint x;`,
			},
			{
				Statement: `delete from fktable2;`,
			},
			{
				Statement: `rollback to x;`,
			},
			{
				Statement:   `commit; -- fail`,
				ErrorString: `insert or update on table "fktable2" violates foreign key constraint "fktable2_f1_fkey"`,
			},
			{
				Statement: `DETAIL:  Key (f1)=(2) is not present in table "pktable2".
begin;`,
			},
			{
				Statement: `insert into fktable2 values(2);`,
			},
			{
				Statement:   `alter table fktable2 drop constraint fktable2_f1_fkey;`,
				ErrorString: `cannot ALTER TABLE "fktable2" because it has pending trigger events`,
			},
			{
				Statement: `commit;`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `delete from pktable2 where f1 = 1;`,
			},
			{
				Statement:   `alter table fktable2 drop constraint fktable2_f1_fkey;`,
				ErrorString: `cannot ALTER TABLE "pktable2" because it has pending trigger events`,
			},
			{
				Statement: `commit;`,
			},
			{
				Statement: `drop table pktable2, fktable2;`,
			},
			{
				Statement: `create table pktable2 (a float8, b float8, primary key (a, b));`,
			},
			{
				Statement: `create table fktable2 (x float8, y float8, foreign key (x, y) references pktable2 (a, b) on update cascade);`,
			},
			{
				Statement: `insert into pktable2 values ('-0', '-0');`,
			},
			{
				Statement: `insert into fktable2 values ('-0', '-0');`,
			},
			{
				Statement: `select * from pktable2;`,
				Results:   []sql.Row{{-0, -0}},
			},
			{
				Statement: `select * from fktable2;`,
				Results:   []sql.Row{{-0, -0}},
			},
			{
				Statement: `update pktable2 set a = '0' where a = '-0';`,
			},
			{
				Statement: `select * from pktable2;`,
				Results:   []sql.Row{{0, -0}},
			},
			{
				Statement: `select * from fktable2;`,
				Results:   []sql.Row{{0, -0}},
			},
			{
				Statement: `drop table pktable2, fktable2;`,
			},
			{
				Statement: `CREATE TABLE fk_notpartitioned_pk (fdrop1 int, a int, fdrop2 int, b int,
  PRIMARY KEY (a, b));`,
			},
			{
				Statement: `ALTER TABLE fk_notpartitioned_pk DROP COLUMN fdrop1, DROP COLUMN fdrop2;`,
			},
			{
				Statement: `CREATE TABLE fk_partitioned_fk (b int, fdrop1 int, a int) PARTITION BY RANGE (a, b);`,
			},
			{
				Statement: `ALTER TABLE fk_partitioned_fk DROP COLUMN fdrop1;`,
			},
			{
				Statement: `CREATE TABLE fk_partitioned_fk_1 (fdrop1 int, fdrop2 int, a int, fdrop3 int, b int);`,
			},
			{
				Statement: `ALTER TABLE fk_partitioned_fk_1 DROP COLUMN fdrop1, DROP COLUMN fdrop2, DROP COLUMN fdrop3;`,
			},
			{
				Statement: `ALTER TABLE fk_partitioned_fk ATTACH PARTITION fk_partitioned_fk_1 FOR VALUES FROM (0,0) TO (1000,1000);`,
			},
			{
				Statement: `ALTER TABLE fk_partitioned_fk ADD FOREIGN KEY (a, b) REFERENCES fk_notpartitioned_pk;`,
			},
			{
				Statement: `CREATE TABLE fk_partitioned_fk_2 (b int, fdrop1 int, fdrop2 int, a int);`,
			},
			{
				Statement: `ALTER TABLE fk_partitioned_fk_2 DROP COLUMN fdrop1, DROP COLUMN fdrop2;`,
			},
			{
				Statement: `ALTER TABLE fk_partitioned_fk ATTACH PARTITION fk_partitioned_fk_2 FOR VALUES FROM (1000,1000) TO (2000,2000);`,
			},
			{
				Statement: `CREATE TABLE fk_partitioned_fk_3 (fdrop1 int, fdrop2 int, fdrop3 int, fdrop4 int, b int, a int)
  PARTITION BY HASH (a);`,
			},
			{
				Statement: `ALTER TABLE fk_partitioned_fk_3 DROP COLUMN fdrop1, DROP COLUMN fdrop2,
	DROP COLUMN fdrop3, DROP COLUMN fdrop4;`,
			},
			{
				Statement: `CREATE TABLE fk_partitioned_fk_3_0 PARTITION OF fk_partitioned_fk_3 FOR VALUES WITH (MODULUS 5, REMAINDER 0);`,
			},
			{
				Statement: `CREATE TABLE fk_partitioned_fk_3_1 PARTITION OF fk_partitioned_fk_3 FOR VALUES WITH (MODULUS 5, REMAINDER 1);`,
			},
			{
				Statement: `ALTER TABLE fk_partitioned_fk ATTACH PARTITION fk_partitioned_fk_3
  FOR VALUES FROM (2000,2000) TO (3000,3000);`,
			},
			{
				Statement: `ALTER TABLE ONLY fk_partitioned_fk ADD FOREIGN KEY (a, b)
  REFERENCES fk_notpartitioned_pk;`,
				ErrorString: `cannot use ONLY for foreign key on partitioned table "fk_partitioned_fk" referencing relation "fk_notpartitioned_pk"`,
			},
			{
				Statement: `ALTER TABLE fk_partitioned_fk ADD FOREIGN KEY (a, b)
  REFERENCES fk_notpartitioned_pk NOT VALID;`,
				ErrorString: `cannot add NOT VALID foreign key on partitioned table "fk_partitioned_fk" referencing relation "fk_notpartitioned_pk"`,
			},
			{
				Statement: `DETAIL:  This feature is not yet supported on partitioned tables.
INSERT INTO fk_partitioned_fk (a,b) VALUES (500, 501);`,
				ErrorString: `insert or update on table "fk_partitioned_fk_1" violates foreign key constraint "fk_partitioned_fk_a_b_fkey"`,
			},
			{
				Statement: `DETAIL:  Key (a, b)=(500, 501) is not present in table "fk_notpartitioned_pk".
INSERT INTO fk_partitioned_fk_1 (a,b) VALUES (500, 501);`,
				ErrorString: `insert or update on table "fk_partitioned_fk_1" violates foreign key constraint "fk_partitioned_fk_a_b_fkey"`,
			},
			{
				Statement: `DETAIL:  Key (a, b)=(500, 501) is not present in table "fk_notpartitioned_pk".
INSERT INTO fk_partitioned_fk (a,b) VALUES (1500, 1501);`,
				ErrorString: `insert or update on table "fk_partitioned_fk_2" violates foreign key constraint "fk_partitioned_fk_a_b_fkey"`,
			},
			{
				Statement: `DETAIL:  Key (a, b)=(1500, 1501) is not present in table "fk_notpartitioned_pk".
INSERT INTO fk_partitioned_fk_2 (a,b) VALUES (1500, 1501);`,
				ErrorString: `insert or update on table "fk_partitioned_fk_2" violates foreign key constraint "fk_partitioned_fk_a_b_fkey"`,
			},
			{
				Statement: `DETAIL:  Key (a, b)=(1500, 1501) is not present in table "fk_notpartitioned_pk".
INSERT INTO fk_partitioned_fk (a,b) VALUES (2500, 2502);`,
				ErrorString: `insert or update on table "fk_partitioned_fk_3_1" violates foreign key constraint "fk_partitioned_fk_a_b_fkey"`,
			},
			{
				Statement: `DETAIL:  Key (a, b)=(2500, 2502) is not present in table "fk_notpartitioned_pk".
INSERT INTO fk_partitioned_fk_3 (a,b) VALUES (2500, 2502);`,
				ErrorString: `insert or update on table "fk_partitioned_fk_3_1" violates foreign key constraint "fk_partitioned_fk_a_b_fkey"`,
			},
			{
				Statement: `DETAIL:  Key (a, b)=(2500, 2502) is not present in table "fk_notpartitioned_pk".
INSERT INTO fk_partitioned_fk (a,b) VALUES (2501, 2503);`,
				ErrorString: `insert or update on table "fk_partitioned_fk_3_0" violates foreign key constraint "fk_partitioned_fk_a_b_fkey"`,
			},
			{
				Statement: `DETAIL:  Key (a, b)=(2501, 2503) is not present in table "fk_notpartitioned_pk".
INSERT INTO fk_partitioned_fk_3 (a,b) VALUES (2501, 2503);`,
				ErrorString: `insert or update on table "fk_partitioned_fk_3_0" violates foreign key constraint "fk_partitioned_fk_a_b_fkey"`,
			},
			{
				Statement: `DETAIL:  Key (a, b)=(2501, 2503) is not present in table "fk_notpartitioned_pk".
INSERT INTO fk_notpartitioned_pk VALUES (500, 501), (1500, 1501),
  (2500, 2502), (2501, 2503);`,
			},
			{
				Statement: `INSERT INTO fk_partitioned_fk (a,b) VALUES (500, 501);`,
			},
			{
				Statement: `INSERT INTO fk_partitioned_fk (a,b) VALUES (1500, 1501);`,
			},
			{
				Statement: `INSERT INTO fk_partitioned_fk (a,b) VALUES (2500, 2502);`,
			},
			{
				Statement: `INSERT INTO fk_partitioned_fk (a,b) VALUES (2501, 2503);`,
			},
			{
				Statement:   `UPDATE fk_partitioned_fk SET a = a + 1 WHERE a = 2501;`,
				ErrorString: `insert or update on table "fk_partitioned_fk_3_1" violates foreign key constraint "fk_partitioned_fk_a_b_fkey"`,
			},
			{
				Statement: `DETAIL:  Key (a, b)=(2502, 2503) is not present in table "fk_notpartitioned_pk".
INSERT INTO fk_notpartitioned_pk (a,b) VALUES (2502, 2503);`,
			},
			{
				Statement: `UPDATE fk_partitioned_fk SET a = a + 1 WHERE a = 2501;`,
			},
			{
				Statement:   `UPDATE fk_notpartitioned_pk SET b = 502 WHERE a = 500;`,
				ErrorString: `update or delete on table "fk_notpartitioned_pk" violates foreign key constraint "fk_partitioned_fk_a_b_fkey" on table "fk_partitioned_fk"`,
			},
			{
				Statement: `DETAIL:  Key (a, b)=(500, 501) is still referenced from table "fk_partitioned_fk".
UPDATE fk_notpartitioned_pk SET b = 1502 WHERE a = 1500;`,
				ErrorString: `update or delete on table "fk_notpartitioned_pk" violates foreign key constraint "fk_partitioned_fk_a_b_fkey" on table "fk_partitioned_fk"`,
			},
			{
				Statement: `DETAIL:  Key (a, b)=(1500, 1501) is still referenced from table "fk_partitioned_fk".
UPDATE fk_notpartitioned_pk SET b = 2504 WHERE a = 2500;`,
				ErrorString: `update or delete on table "fk_notpartitioned_pk" violates foreign key constraint "fk_partitioned_fk_a_b_fkey" on table "fk_partitioned_fk"`,
			},
			{
				Statement: `DETAIL:  Key (a, b)=(2500, 2502) is still referenced from table "fk_partitioned_fk".
\d fk_notpartitioned_pk
        Table "public.fk_notpartitioned_pk"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           | not null | 
 b      | integer |           | not null | 
Indexes:
    "fk_notpartitioned_pk_pkey" PRIMARY KEY, btree (a, b)
Referenced by:
    TABLE "fk_partitioned_fk" CONSTRAINT "fk_partitioned_fk_a_b_fkey" FOREIGN KEY (a, b) REFERENCES fk_notpartitioned_pk(a, b)
ALTER TABLE fk_partitioned_fk DROP CONSTRAINT fk_partitioned_fk_a_b_fkey;`,
			},
			{
				Statement: `DROP TABLE fk_notpartitioned_pk, fk_partitioned_fk;`,
			},
			{
				Statement: `CREATE TABLE fk_notpartitioned_pk (a INT, PRIMARY KEY(a), CHECK (a > 0));`,
			},
			{
				Statement: `CREATE TABLE fk_partitioned_fk (a INT REFERENCES fk_notpartitioned_pk(a) PRIMARY KEY) PARTITION BY RANGE(a);`,
			},
			{
				Statement: `CREATE TABLE fk_partitioned_fk_1 PARTITION OF fk_partitioned_fk FOR VALUES FROM (MINVALUE) TO (MAXVALUE);`,
			},
			{
				Statement: `INSERT INTO fk_notpartitioned_pk VALUES (1);`,
			},
			{
				Statement: `INSERT INTO fk_partitioned_fk VALUES (1);`,
			},
			{
				Statement: `ALTER TABLE fk_notpartitioned_pk ALTER COLUMN a TYPE bigint;`,
			},
			{
				Statement:   `DELETE FROM fk_notpartitioned_pk WHERE a = 1;`,
				ErrorString: `update or delete on table "fk_notpartitioned_pk" violates foreign key constraint "fk_partitioned_fk_a_fkey" on table "fk_partitioned_fk"`,
			},
			{
				Statement: `DETAIL:  Key (a)=(1) is still referenced from table "fk_partitioned_fk".
DROP TABLE fk_notpartitioned_pk, fk_partitioned_fk;`,
			},
			{
				Statement: `CREATE TABLE fk_notpartitioned_pk (a int, b int, primary key (a, b));`,
			},
			{
				Statement: `CREATE TABLE fk_partitioned_fk (a int default 2501, b int default 142857) PARTITION BY LIST (a);`,
			},
			{
				Statement: `CREATE TABLE fk_partitioned_fk_1 PARTITION OF fk_partitioned_fk FOR VALUES IN (NULL,500,501,502);`,
			},
			{
				Statement: `ALTER TABLE fk_partitioned_fk ADD FOREIGN KEY (a, b)
  REFERENCES fk_notpartitioned_pk MATCH SIMPLE
  ON DELETE SET NULL ON UPDATE SET NULL;`,
			},
			{
				Statement: `CREATE TABLE fk_partitioned_fk_2 PARTITION OF fk_partitioned_fk FOR VALUES IN (1500,1502);`,
			},
			{
				Statement: `CREATE TABLE fk_partitioned_fk_3 (a int, b int);`,
			},
			{
				Statement: `ALTER TABLE fk_partitioned_fk ATTACH PARTITION fk_partitioned_fk_3 FOR VALUES IN (2500,2501,2502,2503);`,
			},
			{
				Statement:   `INSERT INTO fk_partitioned_fk (a, b) VALUES (2502, 2503);`,
				ErrorString: `insert or update on table "fk_partitioned_fk_3" violates foreign key constraint "fk_partitioned_fk_a_b_fkey"`,
			},
			{
				Statement: `DETAIL:  Key (a, b)=(2502, 2503) is not present in table "fk_notpartitioned_pk".
INSERT INTO fk_partitioned_fk_3 (a, b) VALUES (2502, 2503);`,
				ErrorString: `insert or update on table "fk_partitioned_fk_3" violates foreign key constraint "fk_partitioned_fk_a_b_fkey"`,
			},
			{
				Statement: `DETAIL:  Key (a, b)=(2502, 2503) is not present in table "fk_notpartitioned_pk".
INSERT INTO fk_partitioned_fk_3 (a, b) VALUES (2502, NULL);`,
			},
			{
				Statement: `INSERT INTO fk_notpartitioned_pk VALUES (2502, 2503);`,
			},
			{
				Statement: `INSERT INTO fk_partitioned_fk_3 (a, b) VALUES (2502, 2503);`,
			},
			{
				Statement: `INSERT INTO fk_partitioned_fk (a,b) VALUES (NULL, NULL);`,
			},
			{
				Statement: `INSERT INTO fk_notpartitioned_pk VALUES (1, 2);`,
			},
			{
				Statement: `CREATE TABLE fk_partitioned_fk_full (x int, y int) PARTITION BY RANGE (x);`,
			},
			{
				Statement: `CREATE TABLE fk_partitioned_fk_full_1 PARTITION OF fk_partitioned_fk_full DEFAULT;`,
			},
			{
				Statement: `INSERT INTO fk_partitioned_fk_full VALUES (1, NULL);`,
			},
			{
				Statement:   `ALTER TABLE fk_partitioned_fk_full ADD FOREIGN KEY (x, y) REFERENCES fk_notpartitioned_pk MATCH FULL;  -- fails`,
				ErrorString: `insert or update on table "fk_partitioned_fk_full_1" violates foreign key constraint "fk_partitioned_fk_full_x_y_fkey"`,
			},
			{
				Statement: `DETAIL:  MATCH FULL does not allow mixing of null and nonnull key values.
TRUNCATE fk_partitioned_fk_full;`,
			},
			{
				Statement: `ALTER TABLE fk_partitioned_fk_full ADD FOREIGN KEY (x, y) REFERENCES fk_notpartitioned_pk MATCH FULL;`,
			},
			{
				Statement:   `INSERT INTO fk_partitioned_fk_full VALUES (1, NULL);  -- fails`,
				ErrorString: `insert or update on table "fk_partitioned_fk_full_1" violates foreign key constraint "fk_partitioned_fk_full_x_y_fkey"`,
			},
			{
				Statement: `DETAIL:  MATCH FULL does not allow mixing of null and nonnull key values.
DROP TABLE fk_partitioned_fk_full;`,
			},
			{
				Statement: `SELECT tableoid::regclass, a, b FROM fk_partitioned_fk WHERE b IS NULL ORDER BY a;`,
				Results:   []sql.Row{{`fk_partitioned_fk_3`, 2502, ``}, {`fk_partitioned_fk_1`, ``, ``}},
			},
			{
				Statement: `UPDATE fk_notpartitioned_pk SET a = a + 1 WHERE a = 2502;`,
			},
			{
				Statement: `SELECT tableoid::regclass, a, b FROM fk_partitioned_fk WHERE b IS NULL ORDER BY a;`,
				Results:   []sql.Row{{`fk_partitioned_fk_3`, 2502, ``}, {`fk_partitioned_fk_1`, ``, ``}, {`fk_partitioned_fk_1`, ``, ``}},
			},
			{
				Statement: `INSERT INTO fk_partitioned_fk VALUES (2503, 2503);`,
			},
			{
				Statement: `SELECT count(*) FROM fk_partitioned_fk WHERE a IS NULL;`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `DELETE FROM fk_notpartitioned_pk;`,
			},
			{
				Statement: `SELECT count(*) FROM fk_partitioned_fk WHERE a IS NULL;`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement: `ALTER TABLE fk_partitioned_fk DROP CONSTRAINT fk_partitioned_fk_a_b_fkey;`,
			},
			{
				Statement: `ALTER TABLE fk_partitioned_fk ADD FOREIGN KEY (a, b)
  REFERENCES fk_notpartitioned_pk
  ON DELETE SET DEFAULT ON UPDATE SET DEFAULT;`,
			},
			{
				Statement: `INSERT INTO fk_notpartitioned_pk VALUES (2502, 2503);`,
			},
			{
				Statement: `INSERT INTO fk_partitioned_fk_3 (a, b) VALUES (2502, 2503);`,
			},
			{
				Statement:   `UPDATE fk_notpartitioned_pk SET a = 1500 WHERE a = 2502;`,
				ErrorString: `insert or update on table "fk_partitioned_fk_3" violates foreign key constraint "fk_partitioned_fk_a_b_fkey"`,
			},
			{
				Statement: `DETAIL:  Key (a, b)=(2501, 142857) is not present in table "fk_notpartitioned_pk".
INSERT INTO fk_notpartitioned_pk VALUES (2501, 142857);`,
			},
			{
				Statement: `UPDATE fk_notpartitioned_pk SET a = 1500 WHERE a = 2502;`,
			},
			{
				Statement: `SELECT * FROM fk_partitioned_fk WHERE b = 142857;`,
				Results:   []sql.Row{{2501, 142857}},
			},
			{
				Statement: `ALTER TABLE fk_partitioned_fk DROP CONSTRAINT fk_partitioned_fk_a_b_fkey;`,
			},
			{
				Statement: `ALTER TABLE fk_partitioned_fk ADD FOREIGN KEY (a, b)
  REFERENCES fk_notpartitioned_pk
  ON DELETE SET NULL (a);`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `DELETE FROM fk_notpartitioned_pk WHERE b = 142857;`,
			},
			{
				Statement: `SELECT * FROM fk_partitioned_fk WHERE a IS NOT NULL OR b IS NOT NULL ORDER BY a NULLS LAST;`,
				Results:   []sql.Row{{2502, ``}, {``, 142857}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `ALTER TABLE fk_partitioned_fk DROP CONSTRAINT fk_partitioned_fk_a_b_fkey;`,
			},
			{
				Statement: `ALTER TABLE fk_partitioned_fk ADD FOREIGN KEY (a, b)
  REFERENCES fk_notpartitioned_pk
  ON DELETE SET DEFAULT (a);`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `DELETE FROM fk_partitioned_fk;`,
			},
			{
				Statement: `DELETE FROM fk_notpartitioned_pk;`,
			},
			{
				Statement: `INSERT INTO fk_notpartitioned_pk VALUES (500, 100000), (2501, 100000);`,
			},
			{
				Statement: `INSERT INTO fk_partitioned_fk VALUES (500, 100000);`,
			},
			{
				Statement: `DELETE FROM fk_notpartitioned_pk WHERE a = 500;`,
			},
			{
				Statement: `SELECT * FROM fk_partitioned_fk ORDER BY a;`,
				Results:   []sql.Row{{2501, 100000}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `ALTER TABLE fk_partitioned_fk DROP CONSTRAINT fk_partitioned_fk_a_b_fkey;`,
			},
			{
				Statement: `ALTER TABLE fk_partitioned_fk ADD FOREIGN KEY (a, b)
  REFERENCES fk_notpartitioned_pk
  ON DELETE CASCADE ON UPDATE CASCADE;`,
			},
			{
				Statement: `UPDATE fk_notpartitioned_pk SET a = 2502 WHERE a = 2501;`,
			},
			{
				Statement: `SELECT * FROM fk_partitioned_fk WHERE b = 142857;`,
				Results:   []sql.Row{{2502, 142857}},
			},
			{
				Statement: `SELECT * FROM fk_partitioned_fk WHERE b = 142857;`,
				Results:   []sql.Row{{2502, 142857}},
			},
			{
				Statement: `DELETE FROM fk_notpartitioned_pk WHERE b = 142857;`,
			},
			{
				Statement: `SELECT * FROM fk_partitioned_fk WHERE a = 142857;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `DROP TABLE fk_partitioned_fk_2;`,
			},
			{
				Statement: `CREATE TABLE fk_partitioned_fk_2 PARTITION OF fk_partitioned_fk FOR VALUES IN (1500,1502);`,
			},
			{
				Statement: `ALTER TABLE fk_partitioned_fk DETACH PARTITION fk_partitioned_fk_2;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `DROP TABLE fk_partitioned_fk;`,
			},
			{
				Statement: `\d fk_partitioned_fk_2;`,
			},
			{
				Statement: `        Table "public.fk_partitioned_fk_2"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 2501
 b      | integer |           |          | 142857
Foreign-key constraints:
    "fk_partitioned_fk_a_b_fkey" FOREIGN KEY (a, b) REFERENCES fk_notpartitioned_pk(a, b) ON UPDATE CASCADE ON DELETE CASCADE
ROLLBACK;`,
			},
			{
				Statement: `ALTER TABLE fk_partitioned_fk ATTACH PARTITION fk_partitioned_fk_2 FOR VALUES IN (1500,1502);`,
			},
			{
				Statement: `DROP TABLE fk_partitioned_fk_2;`,
			},
			{
				Statement: `CREATE TABLE fk_partitioned_fk_2 (b int, c text, a int,
	FOREIGN KEY (a, b) REFERENCES fk_notpartitioned_pk ON UPDATE CASCADE ON DELETE CASCADE);`,
			},
			{
				Statement: `ALTER TABLE fk_partitioned_fk_2 DROP COLUMN c;`,
			},
			{
				Statement: `ALTER TABLE fk_partitioned_fk ATTACH PARTITION fk_partitioned_fk_2 FOR VALUES IN (1500,1502);`,
			},
			{
				Statement: `\d fk_partitioned_fk_2
        Table "public.fk_partitioned_fk_2"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 b      | integer |           |          | 
 a      | integer |           |          | 
Partition of: fk_partitioned_fk FOR VALUES IN (1500, 1502)
Foreign-key constraints:
    TABLE "fk_partitioned_fk" CONSTRAINT "fk_partitioned_fk_a_b_fkey" FOREIGN KEY (a, b) REFERENCES fk_notpartitioned_pk(a, b) ON UPDATE CASCADE ON DELETE CASCADE
DROP TABLE fk_partitioned_fk_2;`,
			},
			{
				Statement: `CREATE TABLE fk_partitioned_fk_4 (a int, b int, FOREIGN KEY (a, b) REFERENCES fk_notpartitioned_pk(a, b) ON UPDATE CASCADE ON DELETE CASCADE) PARTITION BY RANGE (b, a);`,
			},
			{
				Statement: `CREATE TABLE fk_partitioned_fk_4_1 PARTITION OF fk_partitioned_fk_4 FOR VALUES FROM (1,1) TO (100,100);`,
			},
			{
				Statement: `CREATE TABLE fk_partitioned_fk_4_2 (a int, b int, FOREIGN KEY (a, b) REFERENCES fk_notpartitioned_pk(a, b) ON UPDATE SET NULL);`,
			},
			{
				Statement: `ALTER TABLE fk_partitioned_fk_4 ATTACH PARTITION fk_partitioned_fk_4_2 FOR VALUES FROM (100,100) TO (1000,1000);`,
			},
			{
				Statement: `ALTER TABLE fk_partitioned_fk ATTACH PARTITION fk_partitioned_fk_4 FOR VALUES IN (3500,3502);`,
			},
			{
				Statement: `ALTER TABLE fk_partitioned_fk DETACH PARTITION fk_partitioned_fk_4;`,
			},
			{
				Statement: `ALTER TABLE fk_partitioned_fk ATTACH PARTITION fk_partitioned_fk_4 FOR VALUES IN (3500,3502);`,
			},
			{
				Statement: `\d fk_partitioned_fk_4
  Partitioned table "public.fk_partitioned_fk_4"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 b      | integer |           |          | 
Partition of: fk_partitioned_fk FOR VALUES IN (3500, 3502)
Partition key: RANGE (b, a)
Foreign-key constraints:
    TABLE "fk_partitioned_fk" CONSTRAINT "fk_partitioned_fk_a_b_fkey" FOREIGN KEY (a, b) REFERENCES fk_notpartitioned_pk(a, b) ON UPDATE CASCADE ON DELETE CASCADE
Number of partitions: 2 (Use \d+ to list them.)
\d fk_partitioned_fk_4_1
       Table "public.fk_partitioned_fk_4_1"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 b      | integer |           |          | 
Partition of: fk_partitioned_fk_4 FOR VALUES FROM (1, 1) TO (100, 100)
Foreign-key constraints:
    TABLE "fk_partitioned_fk" CONSTRAINT "fk_partitioned_fk_a_b_fkey" FOREIGN KEY (a, b) REFERENCES fk_notpartitioned_pk(a, b) ON UPDATE CASCADE ON DELETE CASCADE
\d fk_partitioned_fk_4_2
       Table "public.fk_partitioned_fk_4_2"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 b      | integer |           |          | 
Partition of: fk_partitioned_fk_4 FOR VALUES FROM (100, 100) TO (1000, 1000)
Foreign-key constraints:
    "fk_partitioned_fk_4_2_a_b_fkey" FOREIGN KEY (a, b) REFERENCES fk_notpartitioned_pk(a, b) ON UPDATE SET NULL
    TABLE "fk_partitioned_fk" CONSTRAINT "fk_partitioned_fk_a_b_fkey" FOREIGN KEY (a, b) REFERENCES fk_notpartitioned_pk(a, b) ON UPDATE CASCADE ON DELETE CASCADE
CREATE TABLE fk_partitioned_fk_5 (a int, b int,
	FOREIGN KEY (a, b) REFERENCES fk_notpartitioned_pk(a, b) ON UPDATE CASCADE ON DELETE CASCADE DEFERRABLE,
	FOREIGN KEY (a, b) REFERENCES fk_notpartitioned_pk(a, b) MATCH FULL ON UPDATE CASCADE ON DELETE CASCADE)
  PARTITION BY RANGE (a);`,
			},
			{
				Statement: `CREATE TABLE fk_partitioned_fk_5_1 (a int, b int, FOREIGN KEY (a, b) REFERENCES fk_notpartitioned_pk);`,
			},
			{
				Statement: `ALTER TABLE fk_partitioned_fk ATTACH PARTITION fk_partitioned_fk_5 FOR VALUES IN (4500);`,
			},
			{
				Statement: `ALTER TABLE fk_partitioned_fk_5 ATTACH PARTITION fk_partitioned_fk_5_1 FOR VALUES FROM (0) TO (10);`,
			},
			{
				Statement: `ALTER TABLE fk_partitioned_fk DETACH PARTITION fk_partitioned_fk_5;`,
			},
			{
				Statement: `ALTER TABLE fk_partitioned_fk ATTACH PARTITION fk_partitioned_fk_5 FOR VALUES IN (4500);`,
			},
			{
				Statement: `\d fk_partitioned_fk_5
  Partitioned table "public.fk_partitioned_fk_5"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 b      | integer |           |          | 
Partition of: fk_partitioned_fk FOR VALUES IN (4500)
Partition key: RANGE (a)
Foreign-key constraints:
    "fk_partitioned_fk_5_a_b_fkey" FOREIGN KEY (a, b) REFERENCES fk_notpartitioned_pk(a, b) ON UPDATE CASCADE ON DELETE CASCADE DEFERRABLE
    "fk_partitioned_fk_5_a_b_fkey1" FOREIGN KEY (a, b) REFERENCES fk_notpartitioned_pk(a, b) MATCH FULL ON UPDATE CASCADE ON DELETE CASCADE
    TABLE "fk_partitioned_fk" CONSTRAINT "fk_partitioned_fk_a_b_fkey" FOREIGN KEY (a, b) REFERENCES fk_notpartitioned_pk(a, b) ON UPDATE CASCADE ON DELETE CASCADE
Number of partitions: 1 (Use \d+ to list them.)
ALTER TABLE fk_partitioned_fk_5 DETACH PARTITION fk_partitioned_fk_5_1;`,
			},
			{
				Statement: `ALTER TABLE fk_partitioned_fk_5 ATTACH PARTITION fk_partitioned_fk_5_1 FOR VALUES FROM (0) TO (10);`,
			},
			{
				Statement: `\d fk_partitioned_fk_5_1
       Table "public.fk_partitioned_fk_5_1"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 b      | integer |           |          | 
Partition of: fk_partitioned_fk_5 FOR VALUES FROM (0) TO (10)
Foreign-key constraints:
    "fk_partitioned_fk_5_1_a_b_fkey" FOREIGN KEY (a, b) REFERENCES fk_notpartitioned_pk(a, b)
    TABLE "fk_partitioned_fk_5" CONSTRAINT "fk_partitioned_fk_5_a_b_fkey" FOREIGN KEY (a, b) REFERENCES fk_notpartitioned_pk(a, b) ON UPDATE CASCADE ON DELETE CASCADE DEFERRABLE
    TABLE "fk_partitioned_fk_5" CONSTRAINT "fk_partitioned_fk_5_a_b_fkey1" FOREIGN KEY (a, b) REFERENCES fk_notpartitioned_pk(a, b) MATCH FULL ON UPDATE CASCADE ON DELETE CASCADE
    TABLE "fk_partitioned_fk" CONSTRAINT "fk_partitioned_fk_a_b_fkey" FOREIGN KEY (a, b) REFERENCES fk_notpartitioned_pk(a, b) ON UPDATE CASCADE ON DELETE CASCADE
CREATE TABLE fk_partitioned_fk_2 (a int, b int) PARTITION BY RANGE (b);`,
			},
			{
				Statement: `CREATE TABLE fk_partitioned_fk_2_1 PARTITION OF fk_partitioned_fk_2 FOR VALUES FROM (0) TO (1000);`,
			},
			{
				Statement: `CREATE TABLE fk_partitioned_fk_2_2 PARTITION OF fk_partitioned_fk_2 FOR VALUES FROM (1000) TO (2000);`,
			},
			{
				Statement: `INSERT INTO fk_partitioned_fk_2 VALUES (1600, 601), (1600, 1601);`,
			},
			{
				Statement: `ALTER TABLE fk_partitioned_fk ATTACH PARTITION fk_partitioned_fk_2
  FOR VALUES IN (1600);`,
				ErrorString: `insert or update on table "fk_partitioned_fk_2_1" violates foreign key constraint "fk_partitioned_fk_a_b_fkey"`,
			},
			{
				Statement: `DETAIL:  Key (a, b)=(1600, 601) is not present in table "fk_notpartitioned_pk".
INSERT INTO fk_notpartitioned_pk VALUES (1600, 601), (1600, 1601);`,
			},
			{
				Statement: `ALTER TABLE fk_partitioned_fk ATTACH PARTITION fk_partitioned_fk_2
  FOR VALUES IN (1600);`,
			},
			{
				Statement: `create role regress_other_partitioned_fk_owner;`,
			},
			{
				Statement: `grant references on fk_notpartitioned_pk to regress_other_partitioned_fk_owner;`,
			},
			{
				Statement: `set role regress_other_partitioned_fk_owner;`,
			},
			{
				Statement: `create table other_partitioned_fk(a int, b int) partition by list (a);`,
			},
			{
				Statement: `create table other_partitioned_fk_1 partition of other_partitioned_fk
  for values in (2048);`,
			},
			{
				Statement: `insert into other_partitioned_fk
  select 2048, x from generate_series(1,10) x;`,
			},
			{
				Statement: `alter table other_partitioned_fk add foreign key (a, b)
  references fk_notpartitioned_pk(a, b);`,
				ErrorString: `insert or update on table "other_partitioned_fk_1" violates foreign key constraint "other_partitioned_fk_a_b_fkey"`,
			},
			{
				Statement: `DETAIL:  Key (a, b)=(2048, 1) is not present in table "fk_notpartitioned_pk".
reset role;`,
			},
			{
				Statement: `insert into fk_notpartitioned_pk (a, b)
  select 2048, x from generate_series(1,10) x;`,
			},
			{
				Statement: `set role regress_other_partitioned_fk_owner;`,
			},
			{
				Statement: `alter table other_partitioned_fk add foreign key (a, b)
  references fk_notpartitioned_pk(a, b);`,
			},
			{
				Statement: `drop table other_partitioned_fk;`,
			},
			{
				Statement: `reset role;`,
			},
			{
				Statement: `revoke all on fk_notpartitioned_pk from regress_other_partitioned_fk_owner;`,
			},
			{
				Statement: `drop role regress_other_partitioned_fk_owner;`,
			},
			{
				Statement: `CREATE TABLE parted_self_fk (
    id bigint NOT NULL PRIMARY KEY,
    id_abc bigint,
    FOREIGN KEY (id_abc) REFERENCES parted_self_fk(id)
)
PARTITION BY RANGE (id);`,
			},
			{
				Statement: `CREATE TABLE part1_self_fk (
    id bigint NOT NULL PRIMARY KEY,
    id_abc bigint
);`,
			},
			{
				Statement: `ALTER TABLE parted_self_fk ATTACH PARTITION part1_self_fk FOR VALUES FROM (0) TO (10);`,
			},
			{
				Statement: `CREATE TABLE part2_self_fk PARTITION OF parted_self_fk FOR VALUES FROM (10) TO (20);`,
			},
			{
				Statement: `CREATE TABLE part3_self_fk (	-- a partitioned partition
	id bigint NOT NULL PRIMARY KEY,
	id_abc bigint
) PARTITION BY RANGE (id);`,
			},
			{
				Statement: `CREATE TABLE part32_self_fk PARTITION OF part3_self_fk FOR VALUES FROM (20) TO (30);`,
			},
			{
				Statement: `ALTER TABLE parted_self_fk ATTACH PARTITION part3_self_fk FOR VALUES FROM (20) TO (40);`,
			},
			{
				Statement: `CREATE TABLE part33_self_fk (
	id bigint NOT NULL PRIMARY KEY,
	id_abc bigint
);`,
			},
			{
				Statement: `ALTER TABLE part3_self_fk ATTACH PARTITION part33_self_fk FOR VALUES FROM (30) TO (40);`,
			},
			{
				Statement: `SELECT cr.relname, co.conname, co.contype, co.convalidated,
       p.conname AS conparent, p.convalidated, cf.relname AS foreignrel
FROM pg_constraint co
JOIN pg_class cr ON cr.oid = co.conrelid
LEFT JOIN pg_class cf ON cf.oid = co.confrelid
LEFT JOIN pg_constraint p ON p.oid = co.conparentid
WHERE cr.oid IN (SELECT relid FROM pg_partition_tree('parted_self_fk'))
ORDER BY co.contype, cr.relname, co.conname, p.conname;`,
				Results: []sql.Row{{`part1_self_fk`, `parted_self_fk_id_abc_fkey`, false, true, `parted_self_fk_id_abc_fkey`, true, `parted_self_fk`}, {`part2_self_fk`, `parted_self_fk_id_abc_fkey`, false, true, `parted_self_fk_id_abc_fkey`, true, `parted_self_fk`}, {`part32_self_fk`, `parted_self_fk_id_abc_fkey`, false, true, `parted_self_fk_id_abc_fkey`, true, `parted_self_fk`}, {`part33_self_fk`, `parted_self_fk_id_abc_fkey`, false, true, `parted_self_fk_id_abc_fkey`, true, `parted_self_fk`}, {`part3_self_fk`, `parted_self_fk_id_abc_fkey`, false, true, `parted_self_fk_id_abc_fkey`, true, `parted_self_fk`}, {`parted_self_fk`, `parted_self_fk_id_abc_fkey`, false, true, ``, ``, `parted_self_fk`}, {`part1_self_fk`, `part1_self_fk_pkey`, `p`, true, `parted_self_fk_pkey`, true, ``}, {`part2_self_fk`, `part2_self_fk_pkey`, `p`, true, `parted_self_fk_pkey`, true, ``}, {`part32_self_fk`, `part32_self_fk_pkey`, `p`, true, `part3_self_fk_pkey`, true, ``}, {`part33_self_fk`, `part33_self_fk_pkey`, `p`, true, `part3_self_fk_pkey`, true, ``}, {`part3_self_fk`, `part3_self_fk_pkey`, `p`, true, `parted_self_fk_pkey`, true, ``}, {`parted_self_fk`, `parted_self_fk_pkey`, `p`, true, ``, ``, ``}},
			},
			{
				Statement: `ALTER TABLE parted_self_fk DETACH PARTITION part2_self_fk;`,
			},
			{
				Statement: `ALTER TABLE parted_self_fk ATTACH PARTITION part2_self_fk FOR VALUES FROM (10) TO (20);`,
			},
			{
				Statement: `ALTER TABLE parted_self_fk DETACH PARTITION part2_self_fk;`,
			},
			{
				Statement: `ALTER TABLE parted_self_fk ATTACH PARTITION part2_self_fk FOR VALUES FROM (10) TO (20);`,
			},
			{
				Statement: `SELECT cr.relname, co.conname, co.contype, co.convalidated,
       p.conname AS conparent, p.convalidated, cf.relname AS foreignrel
FROM pg_constraint co
JOIN pg_class cr ON cr.oid = co.conrelid
LEFT JOIN pg_class cf ON cf.oid = co.confrelid
LEFT JOIN pg_constraint p ON p.oid = co.conparentid
WHERE cr.oid IN (SELECT relid FROM pg_partition_tree('parted_self_fk'))
ORDER BY co.contype, cr.relname, co.conname, p.conname;`,
				Results: []sql.Row{{`part1_self_fk`, `parted_self_fk_id_abc_fkey`, false, true, `parted_self_fk_id_abc_fkey`, true, `parted_self_fk`}, {`part2_self_fk`, `parted_self_fk_id_abc_fkey`, false, true, `parted_self_fk_id_abc_fkey`, true, `parted_self_fk`}, {`part32_self_fk`, `parted_self_fk_id_abc_fkey`, false, true, `parted_self_fk_id_abc_fkey`, true, `parted_self_fk`}, {`part33_self_fk`, `parted_self_fk_id_abc_fkey`, false, true, `parted_self_fk_id_abc_fkey`, true, `parted_self_fk`}, {`part3_self_fk`, `parted_self_fk_id_abc_fkey`, false, true, `parted_self_fk_id_abc_fkey`, true, `parted_self_fk`}, {`parted_self_fk`, `parted_self_fk_id_abc_fkey`, false, true, ``, ``, `parted_self_fk`}, {`part1_self_fk`, `part1_self_fk_pkey`, `p`, true, `parted_self_fk_pkey`, true, ``}, {`part2_self_fk`, `part2_self_fk_pkey`, `p`, true, `parted_self_fk_pkey`, true, ``}, {`part32_self_fk`, `part32_self_fk_pkey`, `p`, true, `part3_self_fk_pkey`, true, ``}, {`part33_self_fk`, `part33_self_fk_pkey`, `p`, true, `part3_self_fk_pkey`, true, ``}, {`part3_self_fk`, `part3_self_fk_pkey`, `p`, true, `parted_self_fk_pkey`, true, ``}, {`parted_self_fk`, `parted_self_fk_pkey`, `p`, true, ``, ``, ``}},
			},
			{
				Statement: `create schema fkpart0
  create table pkey (a int primary key)
  create table fk_part (a int) partition by list (a)
  create table fk_part_1 partition of fk_part
      (foreign key (a) references fkpart0.pkey) for values in (1)
  create table fk_part_23 partition of fk_part
      (foreign key (a) references fkpart0.pkey) for values in (2, 3)
      partition by list (a)
  create table fk_part_23_2 partition of fk_part_23 for values in (2);`,
			},
			{
				Statement: `alter table fkpart0.fk_part add foreign key (a) references fkpart0.pkey;`,
			},
			{
				Statement: `\d fkpart0.fk_part_1	\\ -- should have only one FK
             Table "fkpart0.fk_part_1"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
Partition of: fkpart0.fk_part FOR VALUES IN (1)
Foreign-key constraints:
    TABLE "fkpart0.fk_part" CONSTRAINT "fk_part_a_fkey" FOREIGN KEY (a) REFERENCES fkpart0.pkey(a)
alter table fkpart0.fk_part_1 drop constraint fk_part_1_a_fkey;`,
				ErrorString: `cannot drop inherited constraint "fk_part_1_a_fkey" of relation "fk_part_1"`,
			},
			{
				Statement: `\d fkpart0.fk_part_23	\\ -- should have only one FK
      Partitioned table "fkpart0.fk_part_23"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
Partition of: fkpart0.fk_part FOR VALUES IN (2, 3)
Partition key: LIST (a)
Foreign-key constraints:
    TABLE "fkpart0.fk_part" CONSTRAINT "fk_part_a_fkey" FOREIGN KEY (a) REFERENCES fkpart0.pkey(a)
Number of partitions: 1 (Use \d+ to list them.)
\d fkpart0.fk_part_23_2	\\ -- should have only one FK
           Table "fkpart0.fk_part_23_2"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
Partition of: fkpart0.fk_part_23 FOR VALUES IN (2)
Foreign-key constraints:
    TABLE "fkpart0.fk_part" CONSTRAINT "fk_part_a_fkey" FOREIGN KEY (a) REFERENCES fkpart0.pkey(a)
alter table fkpart0.fk_part_23 drop constraint fk_part_23_a_fkey;`,
				ErrorString: `cannot drop inherited constraint "fk_part_23_a_fkey" of relation "fk_part_23"`,
			},
			{
				Statement:   `alter table fkpart0.fk_part_23_2 drop constraint fk_part_23_a_fkey;`,
				ErrorString: `cannot drop inherited constraint "fk_part_23_a_fkey" of relation "fk_part_23_2"`,
			},
			{
				Statement: `create table fkpart0.fk_part_4 partition of fkpart0.fk_part for values in (4);`,
			},
			{
				Statement: `\d fkpart0.fk_part_4
             Table "fkpart0.fk_part_4"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
Partition of: fkpart0.fk_part FOR VALUES IN (4)
Foreign-key constraints:
    TABLE "fkpart0.fk_part" CONSTRAINT "fk_part_a_fkey" FOREIGN KEY (a) REFERENCES fkpart0.pkey(a)
alter table fkpart0.fk_part_4 drop constraint fk_part_a_fkey;`,
				ErrorString: `cannot drop inherited constraint "fk_part_a_fkey" of relation "fk_part_4"`,
			},
			{
				Statement: `create table fkpart0.fk_part_56 partition of fkpart0.fk_part
    for values in (5,6) partition by list (a);`,
			},
			{
				Statement: `create table fkpart0.fk_part_56_5 partition of fkpart0.fk_part_56
    for values in (5);`,
			},
			{
				Statement: `\d fkpart0.fk_part_56
      Partitioned table "fkpart0.fk_part_56"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
Partition of: fkpart0.fk_part FOR VALUES IN (5, 6)
Partition key: LIST (a)
Foreign-key constraints:
    TABLE "fkpart0.fk_part" CONSTRAINT "fk_part_a_fkey" FOREIGN KEY (a) REFERENCES fkpart0.pkey(a)
Number of partitions: 1 (Use \d+ to list them.)
alter table fkpart0.fk_part_56 drop constraint fk_part_a_fkey;`,
				ErrorString: `cannot drop inherited constraint "fk_part_a_fkey" of relation "fk_part_56"`,
			},
			{
				Statement:   `alter table fkpart0.fk_part_56_5 drop constraint fk_part_a_fkey;`,
				ErrorString: `cannot drop inherited constraint "fk_part_a_fkey" of relation "fk_part_56_5"`,
			},
			{
				Statement: `create schema fkpart1
  create table pkey (a int primary key)
  create table fk_part (a int) partition by list (a)
  create table fk_part_1 partition of fk_part for values in (1) partition by list (a)
  create table fk_part_1_1 partition of fk_part_1 for values in (1);`,
			},
			{
				Statement: `alter table fkpart1.fk_part add foreign key (a) references fkpart1.pkey;`,
			},
			{
				Statement:   `insert into fkpart1.fk_part values (1);		-- should fail`,
				ErrorString: `insert or update on table "fk_part_1_1" violates foreign key constraint "fk_part_a_fkey"`,
			},
			{
				Statement: `DETAIL:  Key (a)=(1) is not present in table "pkey".
insert into fkpart1.pkey values (1);`,
			},
			{
				Statement: `insert into fkpart1.fk_part values (1);`,
			},
			{
				Statement:   `delete from fkpart1.pkey where a = 1;		-- should fail`,
				ErrorString: `update or delete on table "pkey" violates foreign key constraint "fk_part_a_fkey" on table "fk_part"`,
			},
			{
				Statement: `DETAIL:  Key (a)=(1) is still referenced from table "fk_part".
alter table fkpart1.fk_part detach partition fkpart1.fk_part_1;`,
			},
			{
				Statement: `create table fkpart1.fk_part_1_2 partition of fkpart1.fk_part_1 for values in (2);`,
			},
			{
				Statement:   `insert into fkpart1.fk_part_1 values (2);	-- should fail`,
				ErrorString: `insert or update on table "fk_part_1_2" violates foreign key constraint "fk_part_a_fkey"`,
			},
			{
				Statement: `DETAIL:  Key (a)=(2) is not present in table "pkey".
delete from fkpart1.pkey where a = 1;`,
				ErrorString: `update or delete on table "pkey" violates foreign key constraint "fk_part_a_fkey" on table "fk_part_1"`,
			},
			{
				Statement: `DETAIL:  Key (a)=(1) is still referenced from table "fk_part_1".
create schema fkpart2
  create table pkey (a int primary key)
  create table fk_part (a int, constraint fkey foreign key (a) references fkpart2.pkey) partition by list (a)
  create table fk_part_1 partition of fkpart2.fk_part for values in (1) partition by list (a)
  create table fk_part_1_1 (a int, constraint my_fkey foreign key (a) references fkpart2.pkey);`,
			},
			{
				Statement: `alter table fkpart2.fk_part_1 attach partition fkpart2.fk_part_1_1 for values in (1);`,
			},
			{
				Statement:   `alter table fkpart2.fk_part_1 drop constraint fkey;	-- should fail`,
				ErrorString: `cannot drop inherited constraint "fkey" of relation "fk_part_1"`,
			},
			{
				Statement:   `alter table fkpart2.fk_part_1_1 drop constraint my_fkey;	-- should fail`,
				ErrorString: `cannot drop inherited constraint "my_fkey" of relation "fk_part_1_1"`,
			},
			{
				Statement: `alter table fkpart2.fk_part detach partition fkpart2.fk_part_1;`,
			},
			{
				Statement: `alter table fkpart2.fk_part_1 drop constraint fkey;	-- ok`,
			},
			{
				Statement:   `alter table fkpart2.fk_part_1_1 drop constraint my_fkey;	-- doesn't exist`,
				ErrorString: `constraint "my_fkey" of relation "fk_part_1_1" does not exist`,
			},
			{
				Statement: `create schema fkpart3
  create table pkey (a int primary key)
  create table fk_part (a int, constraint fkey foreign key (a) references fkpart3.pkey deferrable initially immediate) partition by list (a)
  create table fk_part_1 partition of fkpart3.fk_part for values in (1) partition by list (a)
  create table fk_part_1_1 partition of fkpart3.fk_part_1 for values in (1)
  create table fk_part_2 partition of fkpart3.fk_part for values in (2);`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `set constraints fkpart3.fkey deferred;`,
			},
			{
				Statement: `insert into fkpart3.fk_part values (1);`,
			},
			{
				Statement: `insert into fkpart3.pkey values (1);`,
			},
			{
				Statement: `commit;`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `set constraints fkpart3.fkey deferred;`,
			},
			{
				Statement: `delete from fkpart3.pkey;`,
			},
			{
				Statement: `delete from fkpart3.fk_part;`,
			},
			{
				Statement: `commit;`,
			},
			{
				Statement: `drop schema fkpart0, fkpart1, fkpart2, fkpart3 cascade;`,
			},
			{
				Statement: `DETAIL:  drop cascades to table fkpart3.pkey
drop cascades to table fkpart3.fk_part
drop cascades to table fkpart2.pkey
drop cascades to table fkpart2.fk_part
drop cascades to table fkpart2.fk_part_1
drop cascades to table fkpart1.pkey
drop cascades to table fkpart1.fk_part
drop cascades to table fkpart1.fk_part_1
drop cascades to table fkpart0.pkey
drop cascades to table fkpart0.fk_part
CREATE SCHEMA fkpart3;`,
			},
			{
				Statement: `SET search_path TO fkpart3;`,
			},
			{
				Statement: `CREATE TABLE pk (a int PRIMARY KEY) PARTITION BY RANGE (a);`,
			},
			{
				Statement: `CREATE TABLE pk1 PARTITION OF pk FOR VALUES FROM (0) TO (1000);`,
			},
			{
				Statement: `CREATE TABLE pk2 (b int, a int);`,
			},
			{
				Statement: `ALTER TABLE pk2 DROP COLUMN b;`,
			},
			{
				Statement: `ALTER TABLE pk2 ALTER a SET NOT NULL;`,
			},
			{
				Statement: `ALTER TABLE pk ATTACH PARTITION pk2 FOR VALUES FROM (1000) TO (2000);`,
			},
			{
				Statement: `CREATE TABLE fk (a int) PARTITION BY RANGE (a);`,
			},
			{
				Statement: `CREATE TABLE fk1 PARTITION OF fk FOR VALUES FROM (0) TO (750);`,
			},
			{
				Statement: `ALTER TABLE fk ADD FOREIGN KEY (a) REFERENCES pk;`,
			},
			{
				Statement: `CREATE TABLE fk2 (b int, a int) ;`,
			},
			{
				Statement: `ALTER TABLE fk2 DROP COLUMN b;`,
			},
			{
				Statement: `ALTER TABLE fk ATTACH PARTITION fk2 FOR VALUES FROM (750) TO (3500);`,
			},
			{
				Statement: `CREATE TABLE pk3 PARTITION OF pk FOR VALUES FROM (2000) TO (3000);`,
			},
			{
				Statement: `CREATE TABLE pk4 (LIKE pk);`,
			},
			{
				Statement: `ALTER TABLE pk ATTACH PARTITION pk4 FOR VALUES FROM (3000) TO (4000);`,
			},
			{
				Statement: `CREATE TABLE pk5 (c int, b int, a int NOT NULL) PARTITION BY RANGE (a);`,
			},
			{
				Statement: `ALTER TABLE pk5 DROP COLUMN b, DROP COLUMN c;`,
			},
			{
				Statement: `CREATE TABLE pk51 PARTITION OF pk5 FOR VALUES FROM (4000) TO (4500);`,
			},
			{
				Statement: `CREATE TABLE pk52 PARTITION OF pk5 FOR VALUES FROM (4500) TO (5000);`,
			},
			{
				Statement: `ALTER TABLE pk ATTACH PARTITION pk5 FOR VALUES FROM (4000) TO (5000);`,
			},
			{
				Statement: `CREATE TABLE fk3 PARTITION OF fk FOR VALUES FROM (3500) TO (5000);`,
			},
			{
				Statement:   `INSERT into fk VALUES (1);`,
				ErrorString: `insert or update on table "fk1" violates foreign key constraint "fk_a_fkey"`,
			},
			{
				Statement: `DETAIL:  Key (a)=(1) is not present in table "pk".
INSERT into fk VALUES (1000);`,
				ErrorString: `insert or update on table "fk2" violates foreign key constraint "fk_a_fkey"`,
			},
			{
				Statement: `DETAIL:  Key (a)=(1000) is not present in table "pk".
INSERT into fk VALUES (2000);`,
				ErrorString: `insert or update on table "fk2" violates foreign key constraint "fk_a_fkey"`,
			},
			{
				Statement: `DETAIL:  Key (a)=(2000) is not present in table "pk".
INSERT into fk VALUES (3000);`,
				ErrorString: `insert or update on table "fk2" violates foreign key constraint "fk_a_fkey"`,
			},
			{
				Statement: `DETAIL:  Key (a)=(3000) is not present in table "pk".
INSERT into fk VALUES (4000);`,
				ErrorString: `insert or update on table "fk3" violates foreign key constraint "fk_a_fkey"`,
			},
			{
				Statement: `DETAIL:  Key (a)=(4000) is not present in table "pk".
INSERT into fk VALUES (4500);`,
				ErrorString: `insert or update on table "fk3" violates foreign key constraint "fk_a_fkey"`,
			},
			{
				Statement: `DETAIL:  Key (a)=(4500) is not present in table "pk".
INSERT into pk VALUES (1), (1000), (2000), (3000), (4000), (4500);`,
			},
			{
				Statement: `INSERT into fk VALUES (1), (1000), (2000), (3000), (4000), (4500);`,
			},
			{
				Statement:   `DELETE FROM pk WHERE a = 1;`,
				ErrorString: `update or delete on table "pk1" violates foreign key constraint "fk_a_fkey1" on table "fk"`,
			},
			{
				Statement: `DETAIL:  Key (a)=(1) is still referenced from table "fk".
DELETE FROM pk WHERE a = 1000;`,
				ErrorString: `update or delete on table "pk2" violates foreign key constraint "fk_a_fkey2" on table "fk"`,
			},
			{
				Statement: `DETAIL:  Key (a)=(1000) is still referenced from table "fk".
DELETE FROM pk WHERE a = 2000;`,
				ErrorString: `update or delete on table "pk3" violates foreign key constraint "fk_a_fkey3" on table "fk"`,
			},
			{
				Statement: `DETAIL:  Key (a)=(2000) is still referenced from table "fk".
DELETE FROM pk WHERE a = 3000;`,
				ErrorString: `update or delete on table "pk4" violates foreign key constraint "fk_a_fkey4" on table "fk"`,
			},
			{
				Statement: `DETAIL:  Key (a)=(3000) is still referenced from table "fk".
DELETE FROM pk WHERE a = 4000;`,
				ErrorString: `update or delete on table "pk51" violates foreign key constraint "fk_a_fkey6" on table "fk"`,
			},
			{
				Statement: `DETAIL:  Key (a)=(4000) is still referenced from table "fk".
DELETE FROM pk WHERE a = 4500;`,
				ErrorString: `update or delete on table "pk52" violates foreign key constraint "fk_a_fkey7" on table "fk"`,
			},
			{
				Statement: `DETAIL:  Key (a)=(4500) is still referenced from table "fk".
UPDATE pk SET a = 2 WHERE a = 1;`,
				ErrorString: `update or delete on table "pk1" violates foreign key constraint "fk_a_fkey1" on table "fk"`,
			},
			{
				Statement: `DETAIL:  Key (a)=(1) is still referenced from table "fk".
UPDATE pk SET a = 1002 WHERE a = 1000;`,
				ErrorString: `update or delete on table "pk2" violates foreign key constraint "fk_a_fkey2" on table "fk"`,
			},
			{
				Statement: `DETAIL:  Key (a)=(1000) is still referenced from table "fk".
UPDATE pk SET a = 2002 WHERE a = 2000;`,
				ErrorString: `update or delete on table "pk3" violates foreign key constraint "fk_a_fkey3" on table "fk"`,
			},
			{
				Statement: `DETAIL:  Key (a)=(2000) is still referenced from table "fk".
UPDATE pk SET a = 3002 WHERE a = 3000;`,
				ErrorString: `update or delete on table "pk4" violates foreign key constraint "fk_a_fkey4" on table "fk"`,
			},
			{
				Statement: `DETAIL:  Key (a)=(3000) is still referenced from table "fk".
UPDATE pk SET a = 4002 WHERE a = 4000;`,
				ErrorString: `update or delete on table "pk51" violates foreign key constraint "fk_a_fkey6" on table "fk"`,
			},
			{
				Statement: `DETAIL:  Key (a)=(4000) is still referenced from table "fk".
UPDATE pk SET a = 4502 WHERE a = 4500;`,
				ErrorString: `update or delete on table "pk52" violates foreign key constraint "fk_a_fkey7" on table "fk"`,
			},
			{
				Statement: `DETAIL:  Key (a)=(4500) is still referenced from table "fk".
DELETE FROM fk;`,
			},
			{
				Statement: `UPDATE pk SET a = 2 WHERE a = 1;`,
			},
			{
				Statement: `DELETE FROM pk WHERE a = 2;`,
			},
			{
				Statement: `UPDATE pk SET a = 1002 WHERE a = 1000;`,
			},
			{
				Statement: `DELETE FROM pk WHERE a = 1002;`,
			},
			{
				Statement: `UPDATE pk SET a = 2002 WHERE a = 2000;`,
			},
			{
				Statement: `DELETE FROM pk WHERE a = 2002;`,
			},
			{
				Statement: `UPDATE pk SET a = 3002 WHERE a = 3000;`,
			},
			{
				Statement: `DELETE FROM pk WHERE a = 3002;`,
			},
			{
				Statement: `UPDATE pk SET a = 4002 WHERE a = 4000;`,
			},
			{
				Statement: `DELETE FROM pk WHERE a = 4002;`,
			},
			{
				Statement: `UPDATE pk SET a = 4502 WHERE a = 4500;`,
			},
			{
				Statement: `DELETE FROM pk WHERE a = 4502;`,
			},
			{
				Statement: `CREATE SCHEMA fkpart4;`,
			},
			{
				Statement: `SET search_path TO fkpart4;`,
			},
			{
				Statement: `CREATE TABLE droppk (a int PRIMARY KEY) PARTITION BY RANGE (a);`,
			},
			{
				Statement: `CREATE TABLE droppk1 PARTITION OF droppk FOR VALUES FROM (0) TO (1000);`,
			},
			{
				Statement: `CREATE TABLE droppk_d PARTITION OF droppk DEFAULT;`,
			},
			{
				Statement: `CREATE TABLE droppk2 PARTITION OF droppk FOR VALUES FROM (1000) TO (2000)
  PARTITION BY RANGE (a);`,
			},
			{
				Statement: `CREATE TABLE droppk21 PARTITION OF droppk2 FOR VALUES FROM (1000) TO (1400);`,
			},
			{
				Statement: `CREATE TABLE droppk2_d PARTITION OF droppk2 DEFAULT;`,
			},
			{
				Statement: `INSERT into droppk VALUES (1), (1000), (1500), (2000);`,
			},
			{
				Statement: `CREATE TABLE dropfk (a int REFERENCES droppk);`,
			},
			{
				Statement: `INSERT into dropfk VALUES (1), (1000), (1500), (2000);`,
			},
			{
				Statement:   `ALTER TABLE droppk DETACH PARTITION droppk_d;`,
				ErrorString: `removing partition "droppk_d" violates foreign key constraint "dropfk_a_fkey5"`,
			},
			{
				Statement: `DETAIL:  Key (a)=(2000) is still referenced from table "dropfk".
ALTER TABLE droppk2 DETACH PARTITION droppk2_d;`,
				ErrorString: `removing partition "droppk2_d" violates foreign key constraint "dropfk_a_fkey4"`,
			},
			{
				Statement: `DETAIL:  Key (a)=(1500) is still referenced from table "dropfk".
ALTER TABLE droppk DETACH PARTITION droppk1;`,
				ErrorString: `removing partition "droppk1" violates foreign key constraint "dropfk_a_fkey1"`,
			},
			{
				Statement: `DETAIL:  Key (a)=(1) is still referenced from table "dropfk".
ALTER TABLE droppk DETACH PARTITION droppk2;`,
				ErrorString: `removing partition "droppk2" violates foreign key constraint "dropfk_a_fkey2"`,
			},
			{
				Statement: `DETAIL:  Key (a)=(1000) is still referenced from table "dropfk".
ALTER TABLE droppk2 DETACH PARTITION droppk21;`,
				ErrorString: `removing partition "droppk21" violates foreign key constraint "dropfk_a_fkey3"`,
			},
			{
				Statement: `DETAIL:  Key (a)=(1000) is still referenced from table "dropfk".
DROP TABLE droppk_d;`,
				ErrorString: `cannot drop table droppk_d because other objects depend on it`,
			},
			{
				Statement: `DETAIL:  constraint dropfk_a_fkey on table dropfk depends on table droppk_d
HINT:  Use DROP ... CASCADE to drop the dependent objects too.
DROP TABLE droppk2_d;`,
				ErrorString: `cannot drop table droppk2_d because other objects depend on it`,
			},
			{
				Statement: `DETAIL:  constraint dropfk_a_fkey on table dropfk depends on table droppk2_d
HINT:  Use DROP ... CASCADE to drop the dependent objects too.
DROP TABLE droppk1;`,
				ErrorString: `cannot drop table droppk1 because other objects depend on it`,
			},
			{
				Statement: `DETAIL:  constraint dropfk_a_fkey on table dropfk depends on table droppk1
HINT:  Use DROP ... CASCADE to drop the dependent objects too.
DROP TABLE droppk2;`,
				ErrorString: `cannot drop table droppk2 because other objects depend on it`,
			},
			{
				Statement: `DETAIL:  constraint dropfk_a_fkey on table dropfk depends on table droppk2
HINT:  Use DROP ... CASCADE to drop the dependent objects too.
DROP TABLE droppk21;`,
				ErrorString: `cannot drop table droppk21 because other objects depend on it`,
			},
			{
				Statement: `DETAIL:  constraint dropfk_a_fkey on table dropfk depends on table droppk21
HINT:  Use DROP ... CASCADE to drop the dependent objects too.
DELETE FROM dropfk;`,
			},
			{
				Statement:   `DROP TABLE droppk_d;`,
				ErrorString: `cannot drop table droppk_d because other objects depend on it`,
			},
			{
				Statement: `DETAIL:  constraint dropfk_a_fkey on table dropfk depends on table droppk_d
HINT:  Use DROP ... CASCADE to drop the dependent objects too.
DROP TABLE droppk2_d;`,
				ErrorString: `cannot drop table droppk2_d because other objects depend on it`,
			},
			{
				Statement: `DETAIL:  constraint dropfk_a_fkey on table dropfk depends on table droppk2_d
HINT:  Use DROP ... CASCADE to drop the dependent objects too.
DROP TABLE droppk1;`,
				ErrorString: `cannot drop table droppk1 because other objects depend on it`,
			},
			{
				Statement: `DETAIL:  constraint dropfk_a_fkey on table dropfk depends on table droppk1
HINT:  Use DROP ... CASCADE to drop the dependent objects too.
ALTER TABLE droppk2 DETACH PARTITION droppk21;`,
			},
			{
				Statement:   `DROP TABLE droppk2;`,
				ErrorString: `cannot drop table droppk2 because other objects depend on it`,
			},
			{
				Statement: `DETAIL:  constraint dropfk_a_fkey on table dropfk depends on table droppk2
HINT:  Use DROP ... CASCADE to drop the dependent objects too.
CREATE SCHEMA fkpart5;`,
			},
			{
				Statement: `SET search_path TO fkpart5;`,
			},
			{
				Statement: `CREATE TABLE pk (a int PRIMARY KEY) PARTITION BY LIST (a);`,
			},
			{
				Statement: `CREATE TABLE pk1 PARTITION OF pk FOR VALUES IN (1) PARTITION BY LIST (a);`,
			},
			{
				Statement: `CREATE TABLE pk11 PARTITION OF pk1 FOR VALUES IN (1);`,
			},
			{
				Statement: `CREATE TABLE fk (a int) PARTITION BY LIST (a);`,
			},
			{
				Statement: `CREATE TABLE fk1 PARTITION OF fk FOR VALUES IN (1) PARTITION BY LIST (a);`,
			},
			{
				Statement: `CREATE TABLE fk11 PARTITION OF fk1 FOR VALUES IN (1);`,
			},
			{
				Statement: `ALTER TABLE fk ADD FOREIGN KEY (a) REFERENCES pk;`,
			},
			{
				Statement: `CREATE TABLE pk2 PARTITION OF pk FOR VALUES IN (2);`,
			},
			{
				Statement: `CREATE TABLE pk3 (a int NOT NULL) PARTITION BY LIST (a);`,
			},
			{
				Statement: `CREATE TABLE pk31 PARTITION OF pk3 FOR VALUES IN (31);`,
			},
			{
				Statement: `CREATE TABLE pk32 (b int, a int NOT NULL);`,
			},
			{
				Statement: `ALTER TABLE pk32 DROP COLUMN b;`,
			},
			{
				Statement: `ALTER TABLE pk3 ATTACH PARTITION pk32 FOR VALUES IN (32);`,
			},
			{
				Statement: `ALTER TABLE pk ATTACH PARTITION pk3 FOR VALUES IN (31, 32);`,
			},
			{
				Statement: `CREATE TABLE fk2 PARTITION OF fk FOR VALUES IN (2);`,
			},
			{
				Statement: `CREATE TABLE fk3 (b int, a int);`,
			},
			{
				Statement: `ALTER TABLE fk3 DROP COLUMN b;`,
			},
			{
				Statement: `ALTER TABLE fk ATTACH PARTITION fk3 FOR VALUES IN (3);`,
			},
			{
				Statement: `SELECT pg_describe_object('pg_constraint'::regclass, oid, 0), confrelid::regclass,
       CASE WHEN conparentid <> 0 THEN pg_describe_object('pg_constraint'::regclass, conparentid, 0) ELSE 'TOP' END
FROM pg_catalog.pg_constraint
WHERE conrelid IN (SELECT relid FROM pg_partition_tree('fk'))
ORDER BY conrelid::regclass::text, conname;`,
				Results: []sql.Row{{`constraint fk_a_fkey on table fk`, `pk`, `TOP`}, {`constraint fk_a_fkey1 on table fk`, `pk1`, `constraint fk_a_fkey on table fk`}, {`constraint fk_a_fkey2 on table fk`, `pk11`, `constraint fk_a_fkey1 on table fk`}, {`constraint fk_a_fkey3 on table fk`, `pk2`, `constraint fk_a_fkey on table fk`}, {`constraint fk_a_fkey4 on table fk`, `pk3`, `constraint fk_a_fkey on table fk`}, {`constraint fk_a_fkey5 on table fk`, `pk31`, `constraint fk_a_fkey4 on table fk`}, {`constraint fk_a_fkey6 on table fk`, `pk32`, `constraint fk_a_fkey4 on table fk`}, {`constraint fk_a_fkey on table fk1`, `pk`, `constraint fk_a_fkey on table fk`}, {`constraint fk_a_fkey on table fk11`, `pk`, `constraint fk_a_fkey on table fk1`}, {`constraint fk_a_fkey on table fk2`, `pk`, `constraint fk_a_fkey on table fk`}, {`constraint fk_a_fkey on table fk3`, `pk`, `constraint fk_a_fkey on table fk`}},
			},
			{
				Statement: `CREATE TABLE fk4 (LIKE fk);`,
			},
			{
				Statement: `INSERT INTO fk4 VALUES (50);`,
			},
			{
				Statement:   `ALTER TABLE fk ATTACH PARTITION fk4 FOR VALUES IN (50);`,
				ErrorString: `insert or update on table "fk4" violates foreign key constraint "fk_a_fkey"`,
			},
			{
				Statement: `DETAIL:  Key (a)=(50) is not present in table "pk".
CREATE SCHEMA fkpart9;`,
			},
			{
				Statement: `SET search_path TO fkpart9;`,
			},
			{
				Statement: `CREATE TABLE pk (a int PRIMARY KEY) PARTITION BY LIST (a);`,
			},
			{
				Statement: `CREATE TABLE pk1 PARTITION OF pk FOR VALUES IN (1, 2) PARTITION BY LIST (a);`,
			},
			{
				Statement: `CREATE TABLE pk11 PARTITION OF pk1 FOR VALUES IN (1);`,
			},
			{
				Statement: `CREATE TABLE pk3 PARTITION OF pk FOR VALUES IN (3);`,
			},
			{
				Statement: `CREATE TABLE fk (a int REFERENCES pk DEFERRABLE INITIALLY IMMEDIATE);`,
			},
			{
				Statement:   `INSERT INTO fk VALUES (1);		-- should fail`,
				ErrorString: `insert or update on table "fk" violates foreign key constraint "fk_a_fkey"`,
			},
			{
				Statement: `DETAIL:  Key (a)=(1) is not present in table "pk".
BEGIN;`,
			},
			{
				Statement: `SET CONSTRAINTS fk_a_fkey DEFERRED;`,
			},
			{
				Statement: `INSERT INTO fk VALUES (1);`,
			},
			{
				Statement:   `COMMIT;							-- should fail`,
				ErrorString: `insert or update on table "fk" violates foreign key constraint "fk_a_fkey"`,
			},
			{
				Statement: `DETAIL:  Key (a)=(1) is not present in table "pk".
BEGIN;`,
			},
			{
				Statement: `SET CONSTRAINTS fk_a_fkey DEFERRED;`,
			},
			{
				Statement: `INSERT INTO fk VALUES (1);`,
			},
			{
				Statement: `INSERT INTO pk VALUES (1);`,
			},
			{
				Statement: `COMMIT;							-- OK`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SET CONSTRAINTS fk_a_fkey DEFERRED;`,
			},
			{
				Statement: `DELETE FROM pk WHERE a = 1;`,
			},
			{
				Statement: `DELETE FROM fk WHERE a = 1;`,
			},
			{
				Statement: `COMMIT;							-- OK`,
			},
			{
				Statement: `CREATE TABLE pt(f1 int, f2 int, f3 int, PRIMARY KEY(f1,f2));`,
			},
			{
				Statement: `CREATE TABLE ref(f1 int, f2 int, f3 int)
  PARTITION BY list(f1);`,
			},
			{
				Statement: `CREATE TABLE ref1 PARTITION OF ref FOR VALUES IN (1);`,
			},
			{
				Statement: `CREATE TABLE ref2 PARTITION OF ref FOR VALUES in (2);`,
			},
			{
				Statement: `ALTER TABLE ref ADD FOREIGN KEY(f1,f2) REFERENCES pt;`,
			},
			{
				Statement: `ALTER TABLE ref ALTER CONSTRAINT ref_f1_f2_fkey
  DEFERRABLE INITIALLY DEFERRED;`,
			},
			{
				Statement: `INSERT INTO pt VALUES(1,2,3);`,
			},
			{
				Statement: `INSERT INTO ref VALUES(1,2,3);`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `DELETE FROM pt;`,
			},
			{
				Statement: `DELETE FROM ref;`,
			},
			{
				Statement: `ABORT;`,
			},
			{
				Statement: `DROP TABLE pt, ref;`,
			},
			{
				Statement: `CREATE TABLE pt(f1 int, f2 int, f3 int, PRIMARY KEY(f1,f2));`,
			},
			{
				Statement: `CREATE TABLE ref(f1 int, f2 int, f3 int)
  PARTITION BY list(f1);`,
			},
			{
				Statement: `CREATE TABLE ref1_2 PARTITION OF ref FOR VALUES IN (1, 2) PARTITION BY list (f2);`,
			},
			{
				Statement: `CREATE TABLE ref1 PARTITION OF ref1_2 FOR VALUES IN (1);`,
			},
			{
				Statement: `CREATE TABLE ref2 PARTITION OF ref1_2 FOR VALUES IN (2) PARTITION BY list (f2);`,
			},
			{
				Statement: `CREATE TABLE ref22 PARTITION OF ref2 FOR VALUES IN (2);`,
			},
			{
				Statement: `ALTER TABLE ref ADD FOREIGN KEY(f1,f2) REFERENCES pt;`,
			},
			{
				Statement: `INSERT INTO pt VALUES(1,2,3);`,
			},
			{
				Statement: `INSERT INTO ref VALUES(1,2,3);`,
			},
			{
				Statement: `ALTER TABLE ref22 ALTER CONSTRAINT ref_f1_f2_fkey
  DEFERRABLE INITIALLY IMMEDIATE;	-- fails`,
				ErrorString: `cannot alter constraint "ref_f1_f2_fkey" on relation "ref22"`,
			},
			{
				Statement: `DETAIL:  Constraint "ref_f1_f2_fkey" is derived from constraint "ref_f1_f2_fkey" of relation "ref".
HINT:  You may alter the constraint it derives from, instead.
ALTER TABLE ref ALTER CONSTRAINT ref_f1_f2_fkey
  DEFERRABLE INITIALLY DEFERRED;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `DELETE FROM pt;`,
			},
			{
				Statement: `DELETE FROM ref;`,
			},
			{
				Statement: `ABORT;`,
			},
			{
				Statement: `DROP TABLE pt, ref;`,
			},
			{
				Statement: `CREATE TABLE pt(f1 int, f2 int, f3 int, PRIMARY KEY(f1,f2))
  PARTITION BY LIST(f1);`,
			},
			{
				Statement: `CREATE TABLE pt1 PARTITION OF pt FOR VALUES IN (1);`,
			},
			{
				Statement: `CREATE TABLE pt2 PARTITION OF pt FOR VALUES IN (2);`,
			},
			{
				Statement: `CREATE TABLE ref(f1 int, f2 int, f3 int);`,
			},
			{
				Statement: `ALTER TABLE ref ADD FOREIGN KEY(f1,f2) REFERENCES pt;`,
			},
			{
				Statement: `ALTER TABLE ref ALTER CONSTRAINT ref_f1_f2_fkey
  DEFERRABLE INITIALLY DEFERRED;`,
			},
			{
				Statement: `INSERT INTO pt VALUES(1,2,3);`,
			},
			{
				Statement: `INSERT INTO ref VALUES(1,2,3);`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `DELETE FROM pt;`,
			},
			{
				Statement: `DELETE FROM ref;`,
			},
			{
				Statement: `ABORT;`,
			},
			{
				Statement: `DROP TABLE pt, ref;`,
			},
			{
				Statement: `CREATE TABLE pt(f1 int, f2 int, f3 int, PRIMARY KEY(f1,f2))
  PARTITION BY LIST(f1);`,
			},
			{
				Statement: `CREATE TABLE pt1_2 PARTITION OF pt FOR VALUES IN (1, 2) PARTITION BY LIST (f1);`,
			},
			{
				Statement: `CREATE TABLE pt1 PARTITION OF pt1_2 FOR VALUES IN (1);`,
			},
			{
				Statement: `CREATE TABLE pt2 PARTITION OF pt1_2 FOR VALUES IN (2);`,
			},
			{
				Statement: `CREATE TABLE ref(f1 int, f2 int, f3 int);`,
			},
			{
				Statement: `ALTER TABLE ref ADD FOREIGN KEY(f1,f2) REFERENCES pt;`,
			},
			{
				Statement: `ALTER TABLE ref ALTER CONSTRAINT ref_f1_f2_fkey1
  DEFERRABLE INITIALLY DEFERRED;	-- fails`,
				ErrorString: `cannot alter constraint "ref_f1_f2_fkey1" on relation "ref"`,
			},
			{
				Statement: `DETAIL:  Constraint "ref_f1_f2_fkey1" is derived from constraint "ref_f1_f2_fkey" of relation "ref".
HINT:  You may alter the constraint it derives from, instead.
ALTER TABLE ref ALTER CONSTRAINT ref_f1_f2_fkey
  DEFERRABLE INITIALLY DEFERRED;`,
			},
			{
				Statement: `INSERT INTO pt VALUES(1,2,3);`,
			},
			{
				Statement: `INSERT INTO ref VALUES(1,2,3);`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `DELETE FROM pt;`,
			},
			{
				Statement: `DELETE FROM ref;`,
			},
			{
				Statement: `ABORT;`,
			},
			{
				Statement: `DROP TABLE pt, ref;`,
			},
			{
				Statement: `DROP SCHEMA fkpart9 CASCADE;`,
			},
			{
				Statement: `DETAIL:  drop cascades to table pk
drop cascades to table fk
CREATE SCHEMA fkpart6;`,
			},
			{
				Statement: `SET search_path TO fkpart6;`,
			},
			{
				Statement: `CREATE TABLE pk (a int PRIMARY KEY) PARTITION BY RANGE (a);`,
			},
			{
				Statement: `CREATE TABLE pk1 PARTITION OF pk FOR VALUES FROM (1) TO (100) PARTITION BY RANGE (a);`,
			},
			{
				Statement: `CREATE TABLE pk11 PARTITION OF pk1 FOR VALUES FROM (1) TO (50);`,
			},
			{
				Statement: `CREATE TABLE pk12 PARTITION OF pk1 FOR VALUES FROM (50) TO (100);`,
			},
			{
				Statement: `CREATE TABLE fk (a int) PARTITION BY RANGE (a);`,
			},
			{
				Statement: `CREATE TABLE fk1 PARTITION OF fk FOR VALUES FROM (1) TO (100) PARTITION BY RANGE (a);`,
			},
			{
				Statement: `CREATE TABLE fk11 PARTITION OF fk1 FOR VALUES FROM (1) TO (10);`,
			},
			{
				Statement: `CREATE TABLE fk12 PARTITION OF fk1 FOR VALUES FROM (10) TO (100);`,
			},
			{
				Statement: `ALTER TABLE fk ADD FOREIGN KEY (a) REFERENCES pk ON UPDATE CASCADE ON DELETE CASCADE;`,
			},
			{
				Statement: `CREATE TABLE fk_d PARTITION OF fk DEFAULT;`,
			},
			{
				Statement: `INSERT INTO pk VALUES (1);`,
			},
			{
				Statement: `INSERT INTO fk VALUES (1);`,
			},
			{
				Statement: `UPDATE pk SET a = 20;`,
			},
			{
				Statement: `SELECT tableoid::regclass, * FROM fk;`,
				Results:   []sql.Row{{`fk12`, 20}},
			},
			{
				Statement: `DELETE FROM pk WHERE a = 20;`,
			},
			{
				Statement: `SELECT tableoid::regclass, * FROM fk;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `DROP TABLE fk;`,
			},
			{
				Statement: `TRUNCATE TABLE pk;`,
			},
			{
				Statement: `INSERT INTO pk VALUES (20), (50);`,
			},
			{
				Statement: `CREATE TABLE fk (a int) PARTITION BY RANGE (a);`,
			},
			{
				Statement: `CREATE TABLE fk1 PARTITION OF fk FOR VALUES FROM (1) TO (100) PARTITION BY RANGE (a);`,
			},
			{
				Statement: `CREATE TABLE fk11 PARTITION OF fk1 FOR VALUES FROM (1) TO (10);`,
			},
			{
				Statement: `CREATE TABLE fk12 PARTITION OF fk1 FOR VALUES FROM (10) TO (100);`,
			},
			{
				Statement: `ALTER TABLE fk ADD FOREIGN KEY (a) REFERENCES pk ON UPDATE SET NULL ON DELETE SET NULL;`,
			},
			{
				Statement: `CREATE TABLE fk_d PARTITION OF fk DEFAULT;`,
			},
			{
				Statement: `INSERT INTO fk VALUES (20), (50);`,
			},
			{
				Statement: `UPDATE pk SET a = 21 WHERE a = 20;`,
			},
			{
				Statement: `DELETE FROM pk WHERE a = 50;`,
			},
			{
				Statement: `SELECT tableoid::regclass, * FROM fk;`,
				Results:   []sql.Row{{`fk_d`, ``}, {`fk_d`, ``}},
			},
			{
				Statement: `DROP TABLE fk;`,
			},
			{
				Statement: `TRUNCATE TABLE pk;`,
			},
			{
				Statement: `INSERT INTO pk VALUES (20), (30), (50);`,
			},
			{
				Statement: `CREATE TABLE fk (id int, a int DEFAULT 50) PARTITION BY RANGE (a);`,
			},
			{
				Statement: `CREATE TABLE fk1 PARTITION OF fk FOR VALUES FROM (1) TO (100) PARTITION BY RANGE (a);`,
			},
			{
				Statement: `CREATE TABLE fk11 PARTITION OF fk1 FOR VALUES FROM (1) TO (10);`,
			},
			{
				Statement: `CREATE TABLE fk12 PARTITION OF fk1 FOR VALUES FROM (10) TO (100);`,
			},
			{
				Statement: `ALTER TABLE fk ADD FOREIGN KEY (a) REFERENCES pk ON UPDATE SET DEFAULT ON DELETE SET DEFAULT;`,
			},
			{
				Statement: `CREATE TABLE fk_d PARTITION OF fk DEFAULT;`,
			},
			{
				Statement: `INSERT INTO fk VALUES (1, 20), (2, 30);`,
			},
			{
				Statement: `DELETE FROM pk WHERE a = 20 RETURNING *;`,
				Results:   []sql.Row{{20}},
			},
			{
				Statement: `UPDATE pk SET a = 90 WHERE a = 30 RETURNING *;`,
				Results:   []sql.Row{{90}},
			},
			{
				Statement: `SELECT tableoid::regclass, * FROM fk;`,
				Results:   []sql.Row{{`fk12`, 1, 50}, {`fk12`, 2, 50}},
			},
			{
				Statement: `DROP TABLE fk;`,
			},
			{
				Statement: `TRUNCATE TABLE pk;`,
			},
			{
				Statement: `INSERT INTO pk VALUES (20), (30);`,
			},
			{
				Statement: `CREATE TABLE fk (a int DEFAULT 50) PARTITION BY RANGE (a);`,
			},
			{
				Statement: `CREATE TABLE fk1 PARTITION OF fk FOR VALUES FROM (1) TO (100) PARTITION BY RANGE (a);`,
			},
			{
				Statement: `CREATE TABLE fk11 PARTITION OF fk1 FOR VALUES FROM (1) TO (10);`,
			},
			{
				Statement: `CREATE TABLE fk12 PARTITION OF fk1 FOR VALUES FROM (10) TO (100);`,
			},
			{
				Statement: `ALTER TABLE fk ADD FOREIGN KEY (a) REFERENCES pk ON UPDATE RESTRICT ON DELETE RESTRICT;`,
			},
			{
				Statement: `CREATE TABLE fk_d PARTITION OF fk DEFAULT;`,
			},
			{
				Statement: `INSERT INTO fk VALUES (20), (30);`,
			},
			{
				Statement:   `DELETE FROM pk WHERE a = 20;`,
				ErrorString: `update or delete on table "pk11" violates foreign key constraint "fk_a_fkey2" on table "fk"`,
			},
			{
				Statement: `DETAIL:  Key (a)=(20) is still referenced from table "fk".
UPDATE pk SET a = 90 WHERE a = 30;`,
				ErrorString: `update or delete on table "pk" violates foreign key constraint "fk_a_fkey" on table "fk"`,
			},
			{
				Statement: `DETAIL:  Key (a)=(30) is still referenced from table "fk".
SELECT tableoid::regclass, * FROM fk;`,
				Results: []sql.Row{{`fk12`, 20}, {`fk12`, 30}},
			},
			{
				Statement: `DROP TABLE fk;`,
			},
			{
				Statement: `CREATE SCHEMA fkpart7
  CREATE TABLE pkpart (a int) PARTITION BY LIST (a)
  CREATE TABLE pkpart1 PARTITION OF pkpart FOR VALUES IN (1);`,
			},
			{
				Statement: `ALTER TABLE fkpart7.pkpart1 ADD PRIMARY KEY (a);`,
			},
			{
				Statement: `ALTER TABLE fkpart7.pkpart ADD PRIMARY KEY (a);`,
			},
			{
				Statement: `CREATE TABLE fkpart7.fk (a int REFERENCES fkpart7.pkpart);`,
			},
			{
				Statement: `DROP SCHEMA fkpart7 CASCADE;`,
			},
			{
				Statement: `DETAIL:  drop cascades to table fkpart7.pkpart
drop cascades to table fkpart7.fk
CREATE SCHEMA fkpart8
  CREATE TABLE tbl1(f1 int PRIMARY KEY)
  CREATE TABLE tbl2(f1 int REFERENCES tbl1 DEFERRABLE INITIALLY DEFERRED) PARTITION BY RANGE(f1)
  CREATE TABLE tbl2_p1 PARTITION OF tbl2 FOR VALUES FROM (minvalue) TO (maxvalue);`,
			},
			{
				Statement: `INSERT INTO fkpart8.tbl1 VALUES(1);`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `INSERT INTO fkpart8.tbl2 VALUES(1);`,
			},
			{
				Statement:   `ALTER TABLE fkpart8.tbl2 DROP CONSTRAINT tbl2_f1_fkey;`,
				ErrorString: `cannot ALTER TABLE "tbl2_p1" because it has pending trigger events`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `DROP SCHEMA fkpart8 CASCADE;`,
			},
			{
				Statement: `DETAIL:  drop cascades to table fkpart8.tbl1
drop cascades to table fkpart8.tbl2
CREATE SCHEMA fkpart9
  CREATE TABLE pk (a INT PRIMARY KEY) PARTITION BY RANGE (a)
  CREATE TABLE fk (
    fk_a INT REFERENCES pk(a) ON DELETE CASCADE
  )
  CREATE TABLE pk1 PARTITION OF pk FOR VALUES FROM (30) TO (50) PARTITION BY RANGE (a)
  CREATE TABLE pk11 PARTITION OF pk1 FOR VALUES FROM (30) TO (40);`,
			},
			{
				Statement: `INSERT INTO fkpart9.pk VALUES (35);`,
			},
			{
				Statement: `INSERT INTO fkpart9.fk VALUES (35);`,
			},
			{
				Statement: `DELETE FROM fkpart9.pk WHERE a=35;`,
			},
			{
				Statement: `SELECT * FROM fkpart9.pk;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT * FROM fkpart9.fk;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `DROP SCHEMA fkpart9 CASCADE;`,
			},
			{
				Statement: `DETAIL:  drop cascades to table fkpart9.pk
drop cascades to table fkpart9.fk
CREATE SCHEMA fkpart10
  CREATE TABLE tbl1(f1 int PRIMARY KEY) PARTITION BY RANGE(f1)
  CREATE TABLE tbl1_p1 PARTITION OF tbl1 FOR VALUES FROM (minvalue) TO (1)
  CREATE TABLE tbl1_p2 PARTITION OF tbl1 FOR VALUES FROM (1) TO (maxvalue)
  CREATE TABLE tbl2(f1 int REFERENCES tbl1 DEFERRABLE INITIALLY DEFERRED)
  CREATE TABLE tbl3(f1 int PRIMARY KEY) PARTITION BY RANGE(f1)
  CREATE TABLE tbl3_p1 PARTITION OF tbl3 FOR VALUES FROM (minvalue) TO (1)
  CREATE TABLE tbl3_p2 PARTITION OF tbl3 FOR VALUES FROM (1) TO (maxvalue)
  CREATE TABLE tbl4(f1 int REFERENCES tbl3 DEFERRABLE INITIALLY DEFERRED);`,
			},
			{
				Statement: `INSERT INTO fkpart10.tbl1 VALUES (0), (1);`,
			},
			{
				Statement: `INSERT INTO fkpart10.tbl2 VALUES (0), (1);`,
			},
			{
				Statement: `INSERT INTO fkpart10.tbl3 VALUES (-2), (-1), (0);`,
			},
			{
				Statement: `INSERT INTO fkpart10.tbl4 VALUES (-2), (-1);`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `DELETE FROM fkpart10.tbl1 WHERE f1 = 0;`,
			},
			{
				Statement: `UPDATE fkpart10.tbl1 SET f1 = 2 WHERE f1 = 1;`,
			},
			{
				Statement: `INSERT INTO fkpart10.tbl1 VALUES (0), (1);`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `UPDATE fkpart10.tbl1 SET f1 = 3 WHERE f1 = 0;`,
			},
			{
				Statement: `UPDATE fkpart10.tbl3 SET f1 = f1 * -1;`,
			},
			{
				Statement: `INSERT INTO fkpart10.tbl1 VALUES (4);`,
			},
			{
				Statement:   `COMMIT;`,
				ErrorString: `update or delete on table "tbl1" violates foreign key constraint "tbl2_f1_fkey" on table "tbl2"`,
			},
			{
				Statement: `DETAIL:  Key (f1)=(0) is still referenced from table "tbl2".
BEGIN;`,
			},
			{
				Statement: `UPDATE fkpart10.tbl3 SET f1 = f1 * -1;`,
			},
			{
				Statement: `UPDATE fkpart10.tbl3 SET f1 = f1 + 3;`,
			},
			{
				Statement: `UPDATE fkpart10.tbl1 SET f1 = 3 WHERE f1 = 0;`,
			},
			{
				Statement: `INSERT INTO fkpart10.tbl1 VALUES (0);`,
			},
			{
				Statement:   `COMMIT;`,
				ErrorString: `update or delete on table "tbl3" violates foreign key constraint "tbl4_f1_fkey" on table "tbl4"`,
			},
			{
				Statement: `DETAIL:  Key (f1)=(-2) is still referenced from table "tbl4".
BEGIN;`,
			},
			{
				Statement: `UPDATE fkpart10.tbl3 SET f1 = f1 * -1;`,
			},
			{
				Statement: `UPDATE fkpart10.tbl1 SET f1 = 3 WHERE f1 = 0;`,
			},
			{
				Statement: `INSERT INTO fkpart10.tbl1 VALUES (0);`,
			},
			{
				Statement: `INSERT INTO fkpart10.tbl3 VALUES (-2), (-1);`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `CREATE TABLE fkpart10.tbl5(f1 int REFERENCES fkpart10.tbl3);`,
			},
			{
				Statement: `INSERT INTO fkpart10.tbl5 VALUES (-2), (-1);`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement:   `UPDATE fkpart10.tbl3 SET f1 = f1 * -3;`,
				ErrorString: `update or delete on table "tbl3" violates foreign key constraint "tbl5_f1_fkey" on table "tbl5"`,
			},
			{
				Statement: `DETAIL:  Key (f1)=(-2) is still referenced from table "tbl5".
COMMIT;`,
			},
			{
				Statement: `DELETE FROM fkpart10.tbl5;`,
			},
			{
				Statement: `INSERT INTO fkpart10.tbl5 VALUES (0);`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `UPDATE fkpart10.tbl3 SET f1 = f1 * -3;`,
			},
			{
				Statement:   `COMMIT;`,
				ErrorString: `update or delete on table "tbl3" violates foreign key constraint "tbl4_f1_fkey" on table "tbl4"`,
			},
			{
				Statement: `DETAIL:  Key (f1)=(-2) is still referenced from table "tbl4".
DROP SCHEMA fkpart10 CASCADE;`,
			},
			{
				Statement: `DETAIL:  drop cascades to table fkpart10.tbl1
drop cascades to table fkpart10.tbl2
drop cascades to table fkpart10.tbl3
drop cascades to table fkpart10.tbl4
drop cascades to table fkpart10.tbl5
CREATE SCHEMA fkpart11
  CREATE TABLE pk (a INT PRIMARY KEY, b text) PARTITION BY LIST (a)
  CREATE TABLE fk (
    a INT,
    CONSTRAINT fkey FOREIGN KEY (a) REFERENCES pk(a) ON UPDATE CASCADE ON DELETE CASCADE
  )
  CREATE TABLE fk_parted (
    a INT PRIMARY KEY,
    CONSTRAINT fkey FOREIGN KEY (a) REFERENCES pk(a) ON UPDATE CASCADE ON DELETE CASCADE
  ) PARTITION BY LIST (a)
  CREATE TABLE fk_another (
    a INT,
    CONSTRAINT fkey FOREIGN KEY (a) REFERENCES fk_parted (a) ON UPDATE CASCADE ON DELETE CASCADE
  )
  CREATE TABLE pk1 PARTITION OF pk FOR VALUES IN (1, 2) PARTITION BY LIST (a)
  CREATE TABLE pk2 PARTITION OF pk FOR VALUES IN (3)
  CREATE TABLE pk3 PARTITION OF pk FOR VALUES IN (4)
  CREATE TABLE fk1 PARTITION OF fk_parted FOR VALUES IN (1, 2)
  CREATE TABLE fk2 PARTITION OF fk_parted FOR VALUES IN (3)
  CREATE TABLE fk3 PARTITION OF fk_parted FOR VALUES IN (4);`,
			},
			{
				Statement: `CREATE TABLE fkpart11.pk11 (b text, a int NOT NULL);`,
			},
			{
				Statement: `ALTER TABLE fkpart11.pk1 ATTACH PARTITION fkpart11.pk11 FOR VALUES IN (1);`,
			},
			{
				Statement: `CREATE TABLE fkpart11.pk12 (b text, c int, a int NOT NULL);`,
			},
			{
				Statement: `ALTER TABLE fkpart11.pk12 DROP c;`,
			},
			{
				Statement: `ALTER TABLE fkpart11.pk1 ATTACH PARTITION fkpart11.pk12 FOR VALUES IN (2);`,
			},
			{
				Statement: `INSERT INTO fkpart11.pk VALUES (1, 'xxx'), (3, 'yyy');`,
			},
			{
				Statement: `INSERT INTO fkpart11.fk VALUES (1), (3);`,
			},
			{
				Statement: `INSERT INTO fkpart11.fk_parted VALUES (1), (3);`,
			},
			{
				Statement: `INSERT INTO fkpart11.fk_another VALUES (1), (3);`,
			},
			{
				Statement: `UPDATE fkpart11.pk SET a = a + 1 RETURNING tableoid::pg_catalog.regclass, *;`,
				Results:   []sql.Row{{`fkpart11.pk12`, 2, `xxx`}, {`fkpart11.pk3`, 4, `yyy`}},
			},
			{
				Statement: `SELECT tableoid::pg_catalog.regclass, * FROM fkpart11.fk;`,
				Results:   []sql.Row{{`fkpart11.fk`, 2}, {`fkpart11.fk`, 4}},
			},
			{
				Statement: `SELECT tableoid::pg_catalog.regclass, * FROM fkpart11.fk_parted;`,
				Results:   []sql.Row{{`fkpart11.fk1`, 2}, {`fkpart11.fk3`, 4}},
			},
			{
				Statement: `SELECT tableoid::pg_catalog.regclass, * FROM fkpart11.fk_another;`,
				Results:   []sql.Row{{`fkpart11.fk_another`, 2}, {`fkpart11.fk_another`, 4}},
			},
			{
				Statement: `ALTER TABLE fkpart11.fk DROP CONSTRAINT fkey;`,
			},
			{
				Statement: `DELETE FROM fkpart11.fk WHERE a = 4;`,
			},
			{
				Statement: `ALTER TABLE fkpart11.fk ADD CONSTRAINT fkey FOREIGN KEY (a) REFERENCES fkpart11.pk1 (a) ON UPDATE CASCADE ON DELETE CASCADE;`,
			},
			{
				Statement:   `UPDATE fkpart11.pk SET a = a - 1;`,
				ErrorString: `cannot move tuple across partitions when a non-root ancestor of the source partition is directly referenced in a foreign key`,
			},
			{
				Statement: `DETAIL:  A foreign key points to ancestor "pk1" but not the root ancestor "pk".
HINT:  Consider defining the foreign key on table "pk".
UPDATE fkpart11.pk1 SET a = a - 1;`,
			},
			{
				Statement: `SELECT tableoid::pg_catalog.regclass, * FROM fkpart11.pk;`,
				Results:   []sql.Row{{`fkpart11.pk11`, 1, `xxx`}, {`fkpart11.pk3`, 4, `yyy`}},
			},
			{
				Statement: `SELECT tableoid::pg_catalog.regclass, * FROM fkpart11.fk;`,
				Results:   []sql.Row{{`fkpart11.fk`, 1}},
			},
			{
				Statement: `SELECT tableoid::pg_catalog.regclass, * FROM fkpart11.fk_parted;`,
				Results:   []sql.Row{{`fkpart11.fk1`, 1}, {`fkpart11.fk3`, 4}},
			},
			{
				Statement: `SELECT tableoid::pg_catalog.regclass, * FROM fkpart11.fk_another;`,
				Results:   []sql.Row{{`fkpart11.fk_another`, 4}, {`fkpart11.fk_another`, 1}},
			},
			{
				Statement: `ALTER TABLE fkpart11.fk DROP CONSTRAINT fkey;`,
			},
			{
				Statement: `ALTER TABLE fkpart11.fk ADD CONSTRAINT fkey FOREIGN KEY (a) REFERENCES fkpart11.pk11 (a) ON UPDATE CASCADE ON DELETE CASCADE;`,
			},
			{
				Statement: `UPDATE fkpart11.pk SET a = a + 1 WHERE a = 1;`,
			},
			{
				Statement: `SELECT tableoid::pg_catalog.regclass, * FROM fkpart11.fk;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `DROP TABLE fkpart11.fk;`,
			},
			{
				Statement: `CREATE FUNCTION fkpart11.print_row () RETURNS TRIGGER LANGUAGE plpgsql AS $$
  BEGIN
    RAISE NOTICE 'TABLE: %, OP: %, OLD: %, NEW: %', TG_RELNAME, TG_OP, OLD, NEW;`,
			},
			{
				Statement: `    RETURN NULL;`,
			},
			{
				Statement: `  END;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `CREATE TRIGGER trig_upd_pk AFTER UPDATE ON fkpart11.pk FOR EACH ROW EXECUTE FUNCTION fkpart11.print_row();`,
			},
			{
				Statement: `CREATE TRIGGER trig_del_pk AFTER DELETE ON fkpart11.pk FOR EACH ROW EXECUTE FUNCTION fkpart11.print_row();`,
			},
			{
				Statement: `CREATE TRIGGER trig_ins_pk AFTER INSERT ON fkpart11.pk FOR EACH ROW EXECUTE FUNCTION fkpart11.print_row();`,
			},
			{
				Statement: `CREATE CONSTRAINT TRIGGER trig_upd_fk_parted AFTER UPDATE ON fkpart11.fk_parted INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION fkpart11.print_row();`,
			},
			{
				Statement: `CREATE CONSTRAINT TRIGGER trig_del_fk_parted AFTER DELETE ON fkpart11.fk_parted INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION fkpart11.print_row();`,
			},
			{
				Statement: `CREATE CONSTRAINT TRIGGER trig_ins_fk_parted AFTER INSERT ON fkpart11.fk_parted INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION fkpart11.print_row();`,
			},
			{
				Statement: `UPDATE fkpart11.pk SET a = 3 WHERE a = 4;`,
			},
			{
				Statement: `UPDATE fkpart11.pk SET a = 1 WHERE a = 2;`,
			},
			{
				Statement: `DROP SCHEMA fkpart11 CASCADE;`,
			},
			{
				Statement: `DETAIL:  drop cascades to table fkpart11.pk
drop cascades to table fkpart11.fk_parted
drop cascades to table fkpart11.fk_another
drop cascades to function fkpart11.print_row()`,
			},
		},
	})
}
