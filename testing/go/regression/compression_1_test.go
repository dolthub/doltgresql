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

func TestCompression1(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_compression_1)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_compression_1,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `\set HIDE_TOAST_COMPRESSION false
SET default_toast_compression = 'pglz';`,
			},
			{
				Statement: `CREATE TABLE cmdata(f1 text COMPRESSION pglz);`,
			},
			{
				Statement: `CREATE INDEX idx ON cmdata(f1);`,
			},
			{
				Statement: `INSERT INTO cmdata VALUES(repeat('1234567890', 1000));`,
			},
			{
				Statement: `\d+ cmdata
                                        Table "public.cmdata"
 Column | Type | Collation | Nullable | Default | Storage  | Compression | Stats target | Description 
--------+------+-----------+----------+---------+----------+-------------+--------------+-------------
 f1     | text |           |          |         | extended | pglz        |              | 
Indexes:
    "idx" btree (f1)
CREATE TABLE cmdata1(f1 TEXT COMPRESSION lz4);`,
				ErrorString: `compression method lz4 not supported`,
			},
			{
				Statement:   `INSERT INTO cmdata1 VALUES(repeat('1234567890', 1004));`,
				ErrorString: `relation "cmdata1" does not exist`,
			},
			{
				Statement: `\d+ cmdata1
SELECT pg_column_compression(f1) FROM cmdata;`,
				Results: []sql.Row{{`pglz`}},
			},
			{
				Statement:   `SELECT pg_column_compression(f1) FROM cmdata1;`,
				ErrorString: `relation "cmdata1" does not exist`,
			},
			{
				Statement: `SELECT SUBSTR(f1, 200, 5) FROM cmdata;`,
				Results:   []sql.Row{{01234}},
			},
			{
				Statement:   `SELECT SUBSTR(f1, 2000, 50) FROM cmdata1;`,
				ErrorString: `relation "cmdata1" does not exist`,
			},
			{
				Statement: `SELECT * INTO cmmove1 FROM cmdata;`,
			},
			{
				Statement: `\d+ cmmove1
                                        Table "public.cmmove1"
 Column | Type | Collation | Nullable | Default | Storage  | Compression | Stats target | Description 
--------+------+-----------+----------+---------+----------+-------------+--------------+-------------
 f1     | text |           |          |         | extended |             |              | 
SELECT pg_column_compression(f1) FROM cmmove1;`,
				Results: []sql.Row{{`pglz`}},
			},
			{
				Statement: `CREATE TABLE cmmove3(f1 text COMPRESSION pglz);`,
			},
			{
				Statement: `INSERT INTO cmmove3 SELECT * FROM cmdata;`,
			},
			{
				Statement:   `INSERT INTO cmmove3 SELECT * FROM cmdata1;`,
				ErrorString: `relation "cmdata1" does not exist`,
			},
			{
				Statement: `SELECT pg_column_compression(f1) FROM cmmove3;`,
				Results:   []sql.Row{{`pglz`}},
			},
			{
				Statement:   `CREATE TABLE cmdata2 (LIKE cmdata1 INCLUDING COMPRESSION);`,
				ErrorString: `relation "cmdata1" does not exist`,
			},
			{
				Statement: `\d+ cmdata2
DROP TABLE cmdata2;`,
				ErrorString: `table "cmdata2" does not exist`,
			},
			{
				Statement:   `CREATE TABLE cmdata2 (f1 int COMPRESSION pglz);`,
				ErrorString: `column data type integer does not support compression`,
			},
			{
				Statement: `CREATE TABLE cmmove2(f1 text COMPRESSION pglz);`,
			},
			{
				Statement: `INSERT INTO cmmove2 VALUES (repeat('1234567890', 1004));`,
			},
			{
				Statement: `SELECT pg_column_compression(f1) FROM cmmove2;`,
				Results:   []sql.Row{{`pglz`}},
			},
			{
				Statement:   `UPDATE cmmove2 SET f1 = cmdata1.f1 FROM cmdata1;`,
				ErrorString: `relation "cmdata1" does not exist`,
			},
			{
				Statement: `SELECT pg_column_compression(f1) FROM cmmove2;`,
				Results:   []sql.Row{{`pglz`}},
			},
			{
				Statement: `CREATE OR REPLACE FUNCTION large_val() RETURNS TEXT LANGUAGE SQL AS
'select array_agg(md5(g::text))::text from generate_series(1, 256) g';`,
			},
			{
				Statement: `CREATE TABLE cmdata2 (f1 text COMPRESSION pglz);`,
			},
			{
				Statement: `INSERT INTO cmdata2 SELECT large_val() || repeat('a', 4000);`,
			},
			{
				Statement: `SELECT pg_column_compression(f1) FROM cmdata2;`,
				Results:   []sql.Row{{`pglz`}},
			},
			{
				Statement:   `INSERT INTO cmdata1 SELECT large_val() || repeat('a', 4000);`,
				ErrorString: `relation "cmdata1" does not exist`,
			},
			{
				Statement:   `SELECT pg_column_compression(f1) FROM cmdata1;`,
				ErrorString: `relation "cmdata1" does not exist`,
			},
			{
				Statement:   `SELECT SUBSTR(f1, 200, 5) FROM cmdata1;`,
				ErrorString: `relation "cmdata1" does not exist`,
			},
			{
				Statement: `SELECT SUBSTR(f1, 200, 5) FROM cmdata2;`,
				Results:   []sql.Row{{`8f14e`}},
			},
			{
				Statement: `DROP TABLE cmdata2;`,
			},
			{
				Statement: `CREATE TABLE cmdata2 (f1 int);`,
			},
			{
				Statement: `\d+ cmdata2
                                         Table "public.cmdata2"
 Column |  Type   | Collation | Nullable | Default | Storage | Compression | Stats target | Description 
--------+---------+-----------+----------+---------+---------+-------------+--------------+-------------
 f1     | integer |           |          |         | plain   |             |              | 
ALTER TABLE cmdata2 ALTER COLUMN f1 TYPE varchar;`,
			},
			{
				Statement: `\d+ cmdata2
                                              Table "public.cmdata2"
 Column |       Type        | Collation | Nullable | Default | Storage  | Compression | Stats target | Description 
--------+-------------------+-----------+----------+---------+----------+-------------+--------------+-------------
 f1     | character varying |           |          |         | extended |             |              | 
ALTER TABLE cmdata2 ALTER COLUMN f1 TYPE int USING f1::integer;`,
			},
			{
				Statement: `\d+ cmdata2
                                         Table "public.cmdata2"
 Column |  Type   | Collation | Nullable | Default | Storage | Compression | Stats target | Description 
--------+---------+-----------+----------+---------+---------+-------------+--------------+-------------
 f1     | integer |           |          |         | plain   |             |              | 
ALTER TABLE cmdata2 ALTER COLUMN f1 TYPE varchar;`,
			},
			{
				Statement: `ALTER TABLE cmdata2 ALTER COLUMN f1 SET COMPRESSION pglz;`,
			},
			{
				Statement: `\d+ cmdata2
                                              Table "public.cmdata2"
 Column |       Type        | Collation | Nullable | Default | Storage  | Compression | Stats target | Description 
--------+-------------------+-----------+----------+---------+----------+-------------+--------------+-------------
 f1     | character varying |           |          |         | extended | pglz        |              | 
ALTER TABLE cmdata2 ALTER COLUMN f1 SET STORAGE plain;`,
			},
			{
				Statement: `\d+ cmdata2
                                              Table "public.cmdata2"
 Column |       Type        | Collation | Nullable | Default | Storage | Compression | Stats target | Description 
--------+-------------------+-----------+----------+---------+---------+-------------+--------------+-------------
 f1     | character varying |           |          |         | plain   | pglz        |              | 
INSERT INTO cmdata2 VALUES (repeat('123456789', 800));`,
			},
			{
				Statement: `SELECT pg_column_compression(f1) FROM cmdata2;`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement:   `CREATE MATERIALIZED VIEW compressmv(x) AS SELECT * FROM cmdata1;`,
				ErrorString: `relation "cmdata1" does not exist`,
			},
			{
				Statement: `\d+ compressmv
SELECT pg_column_compression(f1) FROM cmdata1;`,
				ErrorString: `relation "cmdata1" does not exist`,
			},
			{
				Statement:   `SELECT pg_column_compression(x) FROM compressmv;`,
				ErrorString: `relation "compressmv" does not exist`,
			},
			{
				Statement:   `CREATE TABLE cmpart(f1 text COMPRESSION lz4) PARTITION BY HASH(f1);`,
				ErrorString: `compression method lz4 not supported`,
			},
			{
				Statement:   `CREATE TABLE cmpart1 PARTITION OF cmpart FOR VALUES WITH (MODULUS 2, REMAINDER 0);`,
				ErrorString: `relation "cmpart" does not exist`,
			},
			{
				Statement: `CREATE TABLE cmpart2(f1 text COMPRESSION pglz);`,
			},
			{
				Statement:   `ALTER TABLE cmpart ATTACH PARTITION cmpart2 FOR VALUES WITH (MODULUS 2, REMAINDER 1);`,
				ErrorString: `relation "cmpart" does not exist`,
			},
			{
				Statement:   `INSERT INTO cmpart VALUES (repeat('123456789', 1004));`,
				ErrorString: `relation "cmpart" does not exist`,
			},
			{
				Statement:   `INSERT INTO cmpart VALUES (repeat('123456789', 4004));`,
				ErrorString: `relation "cmpart" does not exist`,
			},
			{
				Statement:   `SELECT pg_column_compression(f1) FROM cmpart1;`,
				ErrorString: `relation "cmpart1" does not exist`,
			},
			{
				Statement: `SELECT pg_column_compression(f1) FROM cmpart2;`,
				Results:   []sql.Row{},
			},
			{
				Statement:   `CREATE TABLE cminh() INHERITS(cmdata, cmdata1);`,
				ErrorString: `relation "cmdata1" does not exist`,
			},
			{
				Statement:   `CREATE TABLE cminh(f1 TEXT COMPRESSION lz4) INHERITS(cmdata);`,
				ErrorString: `column "f1" has a compression method conflict`,
			},
			{
				Statement:   `SET default_toast_compression = '';`,
				ErrorString: `invalid value for parameter "default_toast_compression": ""`,
			},
			{
				Statement:   `SET default_toast_compression = 'I do not exist compression';`,
				ErrorString: `invalid value for parameter "default_toast_compression": "I do not exist compression"`,
			},
			{
				Statement:   `SET default_toast_compression = 'lz4';`,
				ErrorString: `invalid value for parameter "default_toast_compression": "lz4"`,
			},
			{
				Statement: `SET default_toast_compression = 'pglz';`,
			},
			{
				Statement:   `ALTER TABLE cmdata ALTER COLUMN f1 SET COMPRESSION lz4;`,
				ErrorString: `compression method lz4 not supported`,
			},
			{
				Statement: `INSERT INTO cmdata VALUES (repeat('123456789', 4004));`,
			},
			{
				Statement: `\d+ cmdata
                                        Table "public.cmdata"
 Column | Type | Collation | Nullable | Default | Storage  | Compression | Stats target | Description 
--------+------+-----------+----------+---------+----------+-------------+--------------+-------------
 f1     | text |           |          |         | extended | pglz        |              | 
Indexes:
    "idx" btree (f1)
SELECT pg_column_compression(f1) FROM cmdata;`,
				Results: []sql.Row{{`pglz`}, {`pglz`}},
			},
			{
				Statement: `ALTER TABLE cmdata2 ALTER COLUMN f1 SET COMPRESSION default;`,
			},
			{
				Statement: `\d+ cmdata2
                                              Table "public.cmdata2"
 Column |       Type        | Collation | Nullable | Default | Storage | Compression | Stats target | Description 
--------+-------------------+-----------+----------+---------+---------+-------------+--------------+-------------
 f1     | character varying |           |          |         | plain   |             |              | 
ALTER MATERIALIZED VIEW compressmv ALTER COLUMN x SET COMPRESSION lz4;`,
				ErrorString: `relation "compressmv" does not exist`,
			},
			{
				Statement: `\d+ compressmv
ALTER TABLE cmpart1 ALTER COLUMN f1 SET COMPRESSION pglz;`,
				ErrorString: `relation "cmpart1" does not exist`,
			},
			{
				Statement:   `ALTER TABLE cmpart2 ALTER COLUMN f1 SET COMPRESSION lz4;`,
				ErrorString: `compression method lz4 not supported`,
			},
			{
				Statement:   `INSERT INTO cmpart VALUES (repeat('123456789', 1004));`,
				ErrorString: `relation "cmpart" does not exist`,
			},
			{
				Statement:   `INSERT INTO cmpart VALUES (repeat('123456789', 4004));`,
				ErrorString: `relation "cmpart" does not exist`,
			},
			{
				Statement:   `SELECT pg_column_compression(f1) FROM cmpart1;`,
				ErrorString: `relation "cmpart1" does not exist`,
			},
			{
				Statement: `SELECT pg_column_compression(f1) FROM cmpart2;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT pg_column_compression(f1) FROM cmdata;`,
				Results:   []sql.Row{{`pglz`}, {`pglz`}},
			},
			{
				Statement: `VACUUM FULL cmdata;`,
			},
			{
				Statement: `SELECT pg_column_compression(f1) FROM cmdata;`,
				Results:   []sql.Row{{`pglz`}, {`pglz`}},
			},
			{
				Statement: `DROP TABLE cmdata2;`,
			},
			{
				Statement:   `CREATE TABLE cmdata2 (f1 TEXT COMPRESSION pglz, f2 TEXT COMPRESSION lz4);`,
				ErrorString: `compression method lz4 not supported`,
			},
			{
				Statement:   `CREATE UNIQUE INDEX idx1 ON cmdata2 ((f1 || f2));`,
				ErrorString: `relation "cmdata2" does not exist`,
			},
			{
				Statement: `INSERT INTO cmdata2 VALUES((SELECT array_agg(md5(g::TEXT))::TEXT FROM
generate_series(1, 50) g), VERSION());`,
				ErrorString: `relation "cmdata2" does not exist`,
			},
			{
				Statement: `SELECT length(f1) FROM cmdata;`,
				Results:   []sql.Row{{10000}, {36036}},
			},
			{
				Statement:   `SELECT length(f1) FROM cmdata1;`,
				ErrorString: `relation "cmdata1" does not exist`,
			},
			{
				Statement: `SELECT length(f1) FROM cmmove1;`,
				Results:   []sql.Row{{10000}},
			},
			{
				Statement: `SELECT length(f1) FROM cmmove2;`,
				Results:   []sql.Row{{10040}},
			},
			{
				Statement: `SELECT length(f1) FROM cmmove3;`,
				Results:   []sql.Row{{10000}},
			},
			{
				Statement:   `CREATE TABLE badcompresstbl (a text COMPRESSION I_Do_Not_Exist_Compression); -- fails`,
				ErrorString: `invalid compression method "i_do_not_exist_compression"`,
			},
			{
				Statement: `CREATE TABLE badcompresstbl (a text);`,
			},
			{
				Statement:   `ALTER TABLE badcompresstbl ALTER a SET COMPRESSION I_Do_Not_Exist_Compression; -- fails`,
				ErrorString: `invalid compression method "i_do_not_exist_compression"`,
			},
			{
				Statement: `DROP TABLE badcompresstbl;`,
			},
			{
				Statement: `\set HIDE_TOAST_COMPRESSION true`,
			},
		},
	})
}
