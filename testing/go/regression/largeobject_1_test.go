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

func TestLargeobject1(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_largeobject_1)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_largeobject_1,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `\getenv abs_srcdir PG_ABS_SRCDIR
\getenv abs_builddir PG_ABS_BUILDDIR
SET bytea_output TO escape;`,
			},
			{
				Statement: `CREATE ROLE regress_lo_user;`,
			},
			{
				Statement: `SELECT lo_create(42);`,
				Results:   []sql.Row{{42}},
			},
			{
				Statement: `ALTER LARGE OBJECT 42 OWNER TO regress_lo_user;`,
			},
			{
				Statement: `GRANT SELECT ON LARGE OBJECT 42 TO public;`,
			},
			{
				Statement: `COMMENT ON LARGE OBJECT 42 IS 'the ultimate answer';`,
			},
			{
				Statement: `\lo_list
               Large objects
 ID |      Owner      |     Description     
----+-----------------+---------------------
 42 | regress_lo_user | the ultimate answer
(1 row)
\lo_list+
                                  Large objects
 ID |      Owner      |         Access privileges          |     Description     
----+-----------------+------------------------------------+---------------------
 42 | regress_lo_user | regress_lo_user=rw/regress_lo_user+| the ultimate answer
    |                 | =r/regress_lo_user                 | 
(1 row)
\lo_unlink 42
\dl
      Large objects
 ID | Owner | Description 
----+-------+-------------
(0 rows)
CREATE TABLE lotest_stash_values (loid oid, fd integer);`,
			},
			{
				Statement: `INSERT INTO lotest_stash_values (loid) SELECT lo_creat(42);`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `UPDATE lotest_stash_values SET fd = lo_open(loid, CAST(x'20000' | x'40000' AS integer));`,
			},
			{
				Statement: `SELECT lowrite(fd, '
I wandered lonely as a cloud
That floats on high o''er vales and hills,
When all at once I saw a crowd,
A host, of golden daffodils;`,
			},
			{
				Statement: `Beside the lake, beneath the trees,
Fluttering and dancing in the breeze.
Continuous as the stars that shine
And twinkle on the milky way,
They stretched in never-ending line
Along the margin of a bay:
Ten thousand saw I at a glance,
Tossing their heads in sprightly dance.
The waves beside them danced; but they
Out-did the sparkling waves in glee:
A poet could not but be gay,
In such a jocund company:
I gazed--and gazed--but little thought
What wealth the show to me had brought:
For oft, when on my couch I lie
In vacant or in pensive mood,
They flash upon that inward eye
Which is the bliss of solitude;`,
			},
			{
				Statement: `And then my heart with pleasure fills,
And dances with the daffodils.
         -- William Wordsworth
') FROM lotest_stash_values;`,
				Results: []sql.Row{{848}},
			},
			{
				Statement: `SELECT lo_close(fd) FROM lotest_stash_values;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `SELECT lo_from_bytea(0, lo_get(loid)) AS newloid FROM lotest_stash_values
\gset
COMMENT ON LARGE OBJECT :newloid IS 'I Wandered Lonely as a Cloud';`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `UPDATE lotest_stash_values SET fd=lo_open(loid, CAST(x'20000' | x'40000' AS integer));`,
			},
			{
				Statement: `SELECT lo_lseek(fd, 104, 0) FROM lotest_stash_values;`,
				Results:   []sql.Row{{104}},
			},
			{
				Statement: `SELECT loread(fd, 28) FROM lotest_stash_values;`,
				Results:   []sql.Row{{`A host, of golden daffodils;`}},
			},
			{
				Statement: `SELECT lo_lseek(fd, -19, 1) FROM lotest_stash_values;`,
				Results:   []sql.Row{{113}},
			},
			{
				Statement: `SELECT lowrite(fd, 'n') FROM lotest_stash_values;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT lo_tell(fd) FROM lotest_stash_values;`,
				Results:   []sql.Row{{114}},
			},
			{
				Statement: `SELECT lo_lseek(fd, -744, 2) FROM lotest_stash_values;`,
				Results:   []sql.Row{{104}},
			},
			{
				Statement: `SELECT loread(fd, 28) FROM lotest_stash_values;`,
				Results:   []sql.Row{{`A host, on golden daffodils;`}},
			},
			{
				Statement: `SELECT lo_close(fd) FROM lotest_stash_values;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SELECT lo_open(loid, x'40000'::int) from lotest_stash_values;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `ABORT;`,
			},
			{
				Statement: `\set filename :abs_builddir '/results/invalid/path'
\set dobody 'DECLARE loid oid; BEGIN '
\set dobody :dobody 'SELECT tbl.loid INTO loid FROM lotest_stash_values tbl; '
\set dobody :dobody 'PERFORM lo_export(loid, ' :'filename' '); '
\set dobody :dobody 'EXCEPTION WHEN UNDEFINED_FILE THEN '
\set dobody :dobody 'RAISE NOTICE ''could not open file, as expected''; END'
DO :'dobody';`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `UPDATE lotest_stash_values SET fd=lo_open(loid, CAST(x'20000' | x'40000' AS integer));`,
			},
			{
				Statement: `SELECT lo_truncate(fd, 11) FROM lotest_stash_values;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT loread(fd, 15) FROM lotest_stash_values;`,
				Results:   []sql.Row{{`\012I wandered`}},
			},
			{
				Statement: `SELECT lo_truncate(fd, 10000) FROM lotest_stash_values;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT loread(fd, 10) FROM lotest_stash_values;`,
				Results:   []sql.Row{{`\000\000\000\000\000\000\000\000\000\000`}},
			},
			{
				Statement: `SELECT lo_lseek(fd, 0, 2) FROM lotest_stash_values;`,
				Results:   []sql.Row{{10000}},
			},
			{
				Statement: `SELECT lo_tell(fd) FROM lotest_stash_values;`,
				Results:   []sql.Row{{10000}},
			},
			{
				Statement: `SELECT lo_truncate(fd, 5000) FROM lotest_stash_values;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT lo_lseek(fd, 0, 2) FROM lotest_stash_values;`,
				Results:   []sql.Row{{5000}},
			},
			{
				Statement: `SELECT lo_tell(fd) FROM lotest_stash_values;`,
				Results:   []sql.Row{{5000}},
			},
			{
				Statement: `SELECT lo_close(fd) FROM lotest_stash_values;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `UPDATE lotest_stash_values SET fd = lo_open(loid, CAST(x'20000' | x'40000' AS integer));`,
			},
			{
				Statement: `SELECT lo_lseek64(fd, 4294967296, 0) FROM lotest_stash_values;`,
				Results:   []sql.Row{{4294967296}},
			},
			{
				Statement: `SELECT lowrite(fd, 'offset:4GB') FROM lotest_stash_values;`,
				Results:   []sql.Row{{10}},
			},
			{
				Statement: `SELECT lo_tell64(fd) FROM lotest_stash_values;`,
				Results:   []sql.Row{{4294967306}},
			},
			{
				Statement: `SELECT lo_lseek64(fd, -10, 1) FROM lotest_stash_values;`,
				Results:   []sql.Row{{4294967296}},
			},
			{
				Statement: `SELECT lo_tell64(fd) FROM lotest_stash_values;`,
				Results:   []sql.Row{{4294967296}},
			},
			{
				Statement: `SELECT loread(fd, 10) FROM lotest_stash_values;`,
				Results:   []sql.Row{{`offset:4GB`}},
			},
			{
				Statement: `SELECT lo_truncate64(fd, 5000000000) FROM lotest_stash_values;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT lo_lseek64(fd, 0, 2) FROM lotest_stash_values;`,
				Results:   []sql.Row{{5000000000}},
			},
			{
				Statement: `SELECT lo_tell64(fd) FROM lotest_stash_values;`,
				Results:   []sql.Row{{5000000000}},
			},
			{
				Statement: `SELECT lo_truncate64(fd, 3000000000) FROM lotest_stash_values;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT lo_lseek64(fd, 0, 2) FROM lotest_stash_values;`,
				Results:   []sql.Row{{3000000000}},
			},
			{
				Statement: `SELECT lo_tell64(fd) FROM lotest_stash_values;`,
				Results:   []sql.Row{{3000000000}},
			},
			{
				Statement: `SELECT lo_close(fd) FROM lotest_stash_values;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `SELECT lo_unlink(loid) from lotest_stash_values;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `TRUNCATE lotest_stash_values;`,
			},
			{
				Statement: `\set filename :abs_srcdir '/data/tenk.data'
INSERT INTO lotest_stash_values (loid) SELECT lo_import(:'filename');`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `UPDATE lotest_stash_values SET fd=lo_open(loid, CAST(x'20000' | x'40000' AS integer));`,
			},
			{
				Statement: `SELECT lo_lseek(fd, 0, 2) FROM lotest_stash_values;`,
				Results:   []sql.Row{{680800}},
			},
			{
				Statement: `SELECT lo_lseek(fd, 2030, 0) FROM lotest_stash_values;`,
				Results:   []sql.Row{{2030}},
			},
			{
				Statement: `SELECT loread(fd, 36) FROM lotest_stash_values;`,
				Results:   []sql.Row{{`44\011144\0111144\0114144\0119144\01188\01189\011SNAAAA\011F`}},
			},
			{
				Statement: `SELECT lo_tell(fd) FROM lotest_stash_values;`,
				Results:   []sql.Row{{2066}},
			},
			{
				Statement: `SELECT lo_lseek(fd, -26, 1) FROM lotest_stash_values;`,
				Results:   []sql.Row{{2040}},
			},
			{
				Statement: `SELECT lowrite(fd, 'abcdefghijklmnop') FROM lotest_stash_values;`,
				Results:   []sql.Row{{16}},
			},
			{
				Statement: `SELECT lo_lseek(fd, 2030, 0) FROM lotest_stash_values;`,
				Results:   []sql.Row{{2030}},
			},
			{
				Statement: `SELECT loread(fd, 36) FROM lotest_stash_values;`,
				Results:   []sql.Row{{`44\011144\011114abcdefghijklmnop9\011SNAAAA\011F`}},
			},
			{
				Statement: `SELECT lo_close(fd) FROM lotest_stash_values;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `\set filename :abs_builddir '/results/lotest.txt'
SELECT lo_export(loid, :'filename') FROM lotest_stash_values;`,
				Results: []sql.Row{{1}},
			},
			{
				Statement: `\lo_import :filename
\set newloid :LASTOID
\set filename :abs_builddir '/results/lotest2.txt'
\lo_export :newloid :filename
SELECT pageno, data FROM pg_largeobject WHERE loid = (SELECT loid from lotest_stash_values)
EXCEPT
SELECT pageno, data FROM pg_largeobject WHERE loid = :newloid;`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT lo_unlink(loid) FROM lotest_stash_values;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `TRUNCATE lotest_stash_values;`,
			},
			{
				Statement: `\lo_unlink :newloid
\set filename :abs_builddir '/results/lotest.txt'
\lo_import :filename
\set newloid_1 :LASTOID
SELECT lo_from_bytea(0, lo_get(:newloid_1)) AS newloid_2
\gset
SELECT md5(lo_get(:newloid_1)) = md5(lo_get(:newloid_2));`,
				Results: []sql.Row{{true}},
			},
			{
				Statement: `SELECT lo_get(:newloid_1, 0, 20);`,
				Results:   []sql.Row{{`8800\0110\0110\0110\0110\0110\0110\011800`}},
			},
			{
				Statement: `SELECT lo_get(:newloid_1, 10, 20);`,
				Results:   []sql.Row{{`\0110\0110\0110\011800\011800\0113800\011`}},
			},
			{
				Statement: `SELECT lo_put(:newloid_1, 5, decode('afafafaf', 'hex'));`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT lo_get(:newloid_1, 0, 20);`,
				Results:   []sql.Row{{`8800\011\257\257\257\2570\0110\0110\0110\011800`}},
			},
			{
				Statement: `SELECT lo_put(:newloid_1, 4294967310, 'foo');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement:   `SELECT lo_get(:newloid_1);`,
				ErrorString: `large object read request is too large`,
			},
			{
				Statement: `SELECT lo_get(:newloid_1, 4294967294, 100);`,
				Results:   []sql.Row{{`\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000foo`}},
			},
			{
				Statement: `\lo_unlink :newloid_1
\lo_unlink :newloid_2
SELECT lo_from_bytea(0, E'\\xdeadbeef') AS newloid
\gset
SET bytea_output TO hex;`,
			},
			{
				Statement: `SELECT lo_get(:newloid);`,
				Results:   []sql.Row{{`\xdeadbeef`}},
			},
			{
				Statement: `SELECT lo_create(2121);`,
				Results:   []sql.Row{{2121}},
			},
			{
				Statement: `COMMENT ON LARGE OBJECT 2121 IS 'testing comments';`,
			},
			{
				Statement: `DROP TABLE lotest_stash_values;`,
			},
			{
				Statement: `DROP ROLE regress_lo_user;`,
			},
		},
	})
}
