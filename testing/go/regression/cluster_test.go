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

func TestCluster(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_cluster)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_cluster,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TABLE clstr_tst_s (rf_a SERIAL PRIMARY KEY,
	b INT);`,
			},
			{
				Statement: `CREATE TABLE clstr_tst (a SERIAL PRIMARY KEY,
	b INT,
	c TEXT,
	d TEXT,
	CONSTRAINT clstr_tst_con FOREIGN KEY (b) REFERENCES clstr_tst_s);`,
			},
			{
				Statement: `CREATE INDEX clstr_tst_b ON clstr_tst (b);`,
			},
			{
				Statement: `CREATE INDEX clstr_tst_c ON clstr_tst (c);`,
			},
			{
				Statement: `CREATE INDEX clstr_tst_c_b ON clstr_tst (c,b);`,
			},
			{
				Statement: `CREATE INDEX clstr_tst_b_c ON clstr_tst (b,c);`,
			},
			{
				Statement: `INSERT INTO clstr_tst_s (b) VALUES (0);`,
			},
			{
				Statement: `INSERT INTO clstr_tst_s (b) SELECT b FROM clstr_tst_s;`,
			},
			{
				Statement: `INSERT INTO clstr_tst_s (b) SELECT b FROM clstr_tst_s;`,
			},
			{
				Statement: `INSERT INTO clstr_tst_s (b) SELECT b FROM clstr_tst_s;`,
			},
			{
				Statement: `INSERT INTO clstr_tst_s (b) SELECT b FROM clstr_tst_s;`,
			},
			{
				Statement: `INSERT INTO clstr_tst_s (b) SELECT b FROM clstr_tst_s;`,
			},
			{
				Statement: `CREATE TABLE clstr_tst_inh () INHERITS (clstr_tst);`,
			},
			{
				Statement: `INSERT INTO clstr_tst (b, c) VALUES (11, 'once');`,
			},
			{
				Statement: `INSERT INTO clstr_tst (b, c) VALUES (10, 'diez');`,
			},
			{
				Statement: `INSERT INTO clstr_tst (b, c) VALUES (31, 'treinta y uno');`,
			},
			{
				Statement: `INSERT INTO clstr_tst (b, c) VALUES (22, 'veintidos');`,
			},
			{
				Statement: `INSERT INTO clstr_tst (b, c) VALUES (3, 'tres');`,
			},
			{
				Statement: `INSERT INTO clstr_tst (b, c) VALUES (20, 'veinte');`,
			},
			{
				Statement: `INSERT INTO clstr_tst (b, c) VALUES (23, 'veintitres');`,
			},
			{
				Statement: `INSERT INTO clstr_tst (b, c) VALUES (21, 'veintiuno');`,
			},
			{
				Statement: `INSERT INTO clstr_tst (b, c) VALUES (4, 'cuatro');`,
			},
			{
				Statement: `INSERT INTO clstr_tst (b, c) VALUES (14, 'catorce');`,
			},
			{
				Statement: `INSERT INTO clstr_tst (b, c) VALUES (2, 'dos');`,
			},
			{
				Statement: `INSERT INTO clstr_tst (b, c) VALUES (18, 'dieciocho');`,
			},
			{
				Statement: `INSERT INTO clstr_tst (b, c) VALUES (27, 'veintisiete');`,
			},
			{
				Statement: `INSERT INTO clstr_tst (b, c) VALUES (25, 'veinticinco');`,
			},
			{
				Statement: `INSERT INTO clstr_tst (b, c) VALUES (13, 'trece');`,
			},
			{
				Statement: `INSERT INTO clstr_tst (b, c) VALUES (28, 'veintiocho');`,
			},
			{
				Statement: `INSERT INTO clstr_tst (b, c) VALUES (32, 'treinta y dos');`,
			},
			{
				Statement: `INSERT INTO clstr_tst (b, c) VALUES (5, 'cinco');`,
			},
			{
				Statement: `INSERT INTO clstr_tst (b, c) VALUES (29, 'veintinueve');`,
			},
			{
				Statement: `INSERT INTO clstr_tst (b, c) VALUES (1, 'uno');`,
			},
			{
				Statement: `INSERT INTO clstr_tst (b, c) VALUES (24, 'veinticuatro');`,
			},
			{
				Statement: `INSERT INTO clstr_tst (b, c) VALUES (30, 'treinta');`,
			},
			{
				Statement: `INSERT INTO clstr_tst (b, c) VALUES (12, 'doce');`,
			},
			{
				Statement: `INSERT INTO clstr_tst (b, c) VALUES (17, 'diecisiete');`,
			},
			{
				Statement: `INSERT INTO clstr_tst (b, c) VALUES (9, 'nueve');`,
			},
			{
				Statement: `INSERT INTO clstr_tst (b, c) VALUES (19, 'diecinueve');`,
			},
			{
				Statement: `INSERT INTO clstr_tst (b, c) VALUES (26, 'veintiseis');`,
			},
			{
				Statement: `INSERT INTO clstr_tst (b, c) VALUES (15, 'quince');`,
			},
			{
				Statement: `INSERT INTO clstr_tst (b, c) VALUES (7, 'siete');`,
			},
			{
				Statement: `INSERT INTO clstr_tst (b, c) VALUES (16, 'dieciseis');`,
			},
			{
				Statement: `INSERT INTO clstr_tst (b, c) VALUES (8, 'ocho');`,
			},
			{
				Statement: `INSERT INTO clstr_tst (b, c, d) VALUES (6, 'seis', repeat('xyzzy', 100000));`,
			},
			{
				Statement: `CLUSTER clstr_tst_c ON clstr_tst;`,
			},
			{
				Statement: `SELECT a,b,c,substring(d for 30), length(d) from clstr_tst;`,
				Results:   []sql.Row{{10, 14, `catorce`, ``, ``}, {18, 5, `cinco`, ``, ``}, {9, 4, `cuatro`, ``, ``}, {26, 19, `diecinueve`, ``, ``}, {12, 18, `dieciocho`, ``, ``}, {30, 16, `dieciseis`, ``, ``}, {24, 17, `diecisiete`, ``, ``}, {2, 10, `diez`, ``, ``}, {23, 12, `doce`, ``, ``}, {11, 2, `dos`, ``, ``}, {25, 9, `nueve`, ``, ``}, {31, 8, `ocho`, ``, ``}, {1, 11, `once`, ``, ``}, {28, 15, `quince`, ``, ``}, {32, 6, `seis`, `xyzzyxyzzyxyzzyxyzzyxyzzyxyzzy`, 500000}, {29, 7, `siete`, ``, ``}, {15, 13, `trece`, ``, ``}, {22, 30, `treinta`, ``, ``}, {17, 32, `treinta y dos`, ``, ``}, {3, 31, `treinta y uno`, ``, ``}, {5, 3, `tres`, ``, ``}, {20, 1, `uno`, ``, ``}, {6, 20, `veinte`, ``, ``}, {14, 25, `veinticinco`, ``, ``}, {21, 24, `veinticuatro`, ``, ``}, {4, 22, `veintidos`, ``, ``}, {19, 29, `veintinueve`, ``, ``}, {16, 28, `veintiocho`, ``, ``}, {27, 26, `veintiseis`, ``, ``}, {13, 27, `veintisiete`, ``, ``}, {7, 23, `veintitres`, ``, ``}, {8, 21, `veintiuno`, ``, ``}},
			},
			{
				Statement: `SELECT a,b,c,substring(d for 30), length(d) from clstr_tst ORDER BY a;`,
				Results:   []sql.Row{{1, 11, `once`, ``, ``}, {2, 10, `diez`, ``, ``}, {3, 31, `treinta y uno`, ``, ``}, {4, 22, `veintidos`, ``, ``}, {5, 3, `tres`, ``, ``}, {6, 20, `veinte`, ``, ``}, {7, 23, `veintitres`, ``, ``}, {8, 21, `veintiuno`, ``, ``}, {9, 4, `cuatro`, ``, ``}, {10, 14, `catorce`, ``, ``}, {11, 2, `dos`, ``, ``}, {12, 18, `dieciocho`, ``, ``}, {13, 27, `veintisiete`, ``, ``}, {14, 25, `veinticinco`, ``, ``}, {15, 13, `trece`, ``, ``}, {16, 28, `veintiocho`, ``, ``}, {17, 32, `treinta y dos`, ``, ``}, {18, 5, `cinco`, ``, ``}, {19, 29, `veintinueve`, ``, ``}, {20, 1, `uno`, ``, ``}, {21, 24, `veinticuatro`, ``, ``}, {22, 30, `treinta`, ``, ``}, {23, 12, `doce`, ``, ``}, {24, 17, `diecisiete`, ``, ``}, {25, 9, `nueve`, ``, ``}, {26, 19, `diecinueve`, ``, ``}, {27, 26, `veintiseis`, ``, ``}, {28, 15, `quince`, ``, ``}, {29, 7, `siete`, ``, ``}, {30, 16, `dieciseis`, ``, ``}, {31, 8, `ocho`, ``, ``}, {32, 6, `seis`, `xyzzyxyzzyxyzzyxyzzyxyzzyxyzzy`, 500000}},
			},
			{
				Statement: `SELECT a,b,c,substring(d for 30), length(d) from clstr_tst ORDER BY b;`,
				Results:   []sql.Row{{20, 1, `uno`, ``, ``}, {11, 2, `dos`, ``, ``}, {5, 3, `tres`, ``, ``}, {9, 4, `cuatro`, ``, ``}, {18, 5, `cinco`, ``, ``}, {32, 6, `seis`, `xyzzyxyzzyxyzzyxyzzyxyzzyxyzzy`, 500000}, {29, 7, `siete`, ``, ``}, {31, 8, `ocho`, ``, ``}, {25, 9, `nueve`, ``, ``}, {2, 10, `diez`, ``, ``}, {1, 11, `once`, ``, ``}, {23, 12, `doce`, ``, ``}, {15, 13, `trece`, ``, ``}, {10, 14, `catorce`, ``, ``}, {28, 15, `quince`, ``, ``}, {30, 16, `dieciseis`, ``, ``}, {24, 17, `diecisiete`, ``, ``}, {12, 18, `dieciocho`, ``, ``}, {26, 19, `diecinueve`, ``, ``}, {6, 20, `veinte`, ``, ``}, {8, 21, `veintiuno`, ``, ``}, {4, 22, `veintidos`, ``, ``}, {7, 23, `veintitres`, ``, ``}, {21, 24, `veinticuatro`, ``, ``}, {14, 25, `veinticinco`, ``, ``}, {27, 26, `veintiseis`, ``, ``}, {13, 27, `veintisiete`, ``, ``}, {16, 28, `veintiocho`, ``, ``}, {19, 29, `veintinueve`, ``, ``}, {22, 30, `treinta`, ``, ``}, {3, 31, `treinta y uno`, ``, ``}, {17, 32, `treinta y dos`, ``, ``}},
			},
			{
				Statement: `SELECT a,b,c,substring(d for 30), length(d) from clstr_tst ORDER BY c;`,
				Results:   []sql.Row{{10, 14, `catorce`, ``, ``}, {18, 5, `cinco`, ``, ``}, {9, 4, `cuatro`, ``, ``}, {26, 19, `diecinueve`, ``, ``}, {12, 18, `dieciocho`, ``, ``}, {30, 16, `dieciseis`, ``, ``}, {24, 17, `diecisiete`, ``, ``}, {2, 10, `diez`, ``, ``}, {23, 12, `doce`, ``, ``}, {11, 2, `dos`, ``, ``}, {25, 9, `nueve`, ``, ``}, {31, 8, `ocho`, ``, ``}, {1, 11, `once`, ``, ``}, {28, 15, `quince`, ``, ``}, {32, 6, `seis`, `xyzzyxyzzyxyzzyxyzzyxyzzyxyzzy`, 500000}, {29, 7, `siete`, ``, ``}, {15, 13, `trece`, ``, ``}, {22, 30, `treinta`, ``, ``}, {17, 32, `treinta y dos`, ``, ``}, {3, 31, `treinta y uno`, ``, ``}, {5, 3, `tres`, ``, ``}, {20, 1, `uno`, ``, ``}, {6, 20, `veinte`, ``, ``}, {14, 25, `veinticinco`, ``, ``}, {21, 24, `veinticuatro`, ``, ``}, {4, 22, `veintidos`, ``, ``}, {19, 29, `veintinueve`, ``, ``}, {16, 28, `veintiocho`, ``, ``}, {27, 26, `veintiseis`, ``, ``}, {13, 27, `veintisiete`, ``, ``}, {7, 23, `veintitres`, ``, ``}, {8, 21, `veintiuno`, ``, ``}},
			},
			{
				Statement: `INSERT INTO clstr_tst_inh VALUES (0, 100, 'in child table');`,
			},
			{
				Statement: `SELECT a,b,c,substring(d for 30), length(d) from clstr_tst;`,
				Results:   []sql.Row{{10, 14, `catorce`, ``, ``}, {18, 5, `cinco`, ``, ``}, {9, 4, `cuatro`, ``, ``}, {26, 19, `diecinueve`, ``, ``}, {12, 18, `dieciocho`, ``, ``}, {30, 16, `dieciseis`, ``, ``}, {24, 17, `diecisiete`, ``, ``}, {2, 10, `diez`, ``, ``}, {23, 12, `doce`, ``, ``}, {11, 2, `dos`, ``, ``}, {25, 9, `nueve`, ``, ``}, {31, 8, `ocho`, ``, ``}, {1, 11, `once`, ``, ``}, {28, 15, `quince`, ``, ``}, {32, 6, `seis`, `xyzzyxyzzyxyzzyxyzzyxyzzyxyzzy`, 500000}, {29, 7, `siete`, ``, ``}, {15, 13, `trece`, ``, ``}, {22, 30, `treinta`, ``, ``}, {17, 32, `treinta y dos`, ``, ``}, {3, 31, `treinta y uno`, ``, ``}, {5, 3, `tres`, ``, ``}, {20, 1, `uno`, ``, ``}, {6, 20, `veinte`, ``, ``}, {14, 25, `veinticinco`, ``, ``}, {21, 24, `veinticuatro`, ``, ``}, {4, 22, `veintidos`, ``, ``}, {19, 29, `veintinueve`, ``, ``}, {16, 28, `veintiocho`, ``, ``}, {27, 26, `veintiseis`, ``, ``}, {13, 27, `veintisiete`, ``, ``}, {7, 23, `veintitres`, ``, ``}, {8, 21, `veintiuno`, ``, ``}, {0, 100, `in child table`, ``, ``}},
			},
			{
				Statement:   `INSERT INTO clstr_tst (b, c) VALUES (1111, 'this should fail');`,
				ErrorString: `insert or update on table "clstr_tst" violates foreign key constraint "clstr_tst_con"`,
			},
			{
				Statement: `SELECT conname FROM pg_constraint WHERE conrelid = 'clstr_tst'::regclass
ORDER BY 1;`,
				Results: []sql.Row{{`clstr_tst_con`}, {`clstr_tst_pkey`}},
			},
			{
				Statement: `SELECT relname, relkind,
    EXISTS(SELECT 1 FROM pg_class WHERE oid = c.reltoastrelid) AS hastoast
FROM pg_class c WHERE relname LIKE 'clstr_tst%' ORDER BY relname;`,
				Results: []sql.Row{{`clstr_tst`, `r`, true}, {`clstr_tst_a_seq`, `S`, false}, {`clstr_tst_b`, `i`, false}, {`clstr_tst_b_c`, `i`, false}, {`clstr_tst_c`, `i`, false}, {`clstr_tst_c_b`, `i`, false}, {`clstr_tst_inh`, `r`, true}, {`clstr_tst_pkey`, `i`, false}, {`clstr_tst_s`, `r`, false}, {`clstr_tst_s_pkey`, `i`, false}, {`clstr_tst_s_rf_a_seq`, `S`, false}},
			},
			{
				Statement: `SELECT pg_class.relname FROM pg_index, pg_class, pg_class AS pg_class_2
WHERE pg_class.oid=indexrelid
	AND indrelid=pg_class_2.oid
	AND pg_class_2.relname = 'clstr_tst'
	AND indisclustered;`,
				Results: []sql.Row{{`clstr_tst_c`}},
			},
			{
				Statement: `ALTER TABLE clstr_tst CLUSTER ON clstr_tst_b_c;`,
			},
			{
				Statement: `SELECT pg_class.relname FROM pg_index, pg_class, pg_class AS pg_class_2
WHERE pg_class.oid=indexrelid
	AND indrelid=pg_class_2.oid
	AND pg_class_2.relname = 'clstr_tst'
	AND indisclustered;`,
				Results: []sql.Row{{`clstr_tst_b_c`}},
			},
			{
				Statement: `ALTER TABLE clstr_tst SET WITHOUT CLUSTER;`,
			},
			{
				Statement: `SELECT pg_class.relname FROM pg_index, pg_class, pg_class AS pg_class_2
WHERE pg_class.oid=indexrelid
	AND indrelid=pg_class_2.oid
	AND pg_class_2.relname = 'clstr_tst'
	AND indisclustered;`,
				Results: []sql.Row{},
			},
			{
				Statement: `CLUSTER pg_toast.pg_toast_826 USING pg_toast_826_index;`,
			},
			{
				Statement: `CREATE USER regress_clstr_user;`,
			},
			{
				Statement: `CREATE TABLE clstr_1 (a INT PRIMARY KEY);`,
			},
			{
				Statement: `CREATE TABLE clstr_2 (a INT PRIMARY KEY);`,
			},
			{
				Statement: `CREATE TABLE clstr_3 (a INT PRIMARY KEY);`,
			},
			{
				Statement: `ALTER TABLE clstr_1 OWNER TO regress_clstr_user;`,
			},
			{
				Statement: `ALTER TABLE clstr_3 OWNER TO regress_clstr_user;`,
			},
			{
				Statement: `GRANT SELECT ON clstr_2 TO regress_clstr_user;`,
			},
			{
				Statement: `INSERT INTO clstr_1 VALUES (2);`,
			},
			{
				Statement: `INSERT INTO clstr_1 VALUES (1);`,
			},
			{
				Statement: `INSERT INTO clstr_2 VALUES (2);`,
			},
			{
				Statement: `INSERT INTO clstr_2 VALUES (1);`,
			},
			{
				Statement: `INSERT INTO clstr_3 VALUES (2);`,
			},
			{
				Statement: `INSERT INTO clstr_3 VALUES (1);`,
			},
			{
				Statement:   `CLUSTER clstr_2;`,
				ErrorString: `there is no previously clustered index for table "clstr_2"`,
			},
			{
				Statement: `CLUSTER clstr_1_pkey ON clstr_1;`,
			},
			{
				Statement: `CLUSTER clstr_2 USING clstr_2_pkey;`,
			},
			{
				Statement: `SELECT * FROM clstr_1 UNION ALL
  SELECT * FROM clstr_2 UNION ALL
  SELECT * FROM clstr_3;`,
				Results: []sql.Row{{1}, {2}, {1}, {2}, {2}, {1}},
			},
			{
				Statement: `DELETE FROM clstr_1;`,
			},
			{
				Statement: `DELETE FROM clstr_2;`,
			},
			{
				Statement: `DELETE FROM clstr_3;`,
			},
			{
				Statement: `INSERT INTO clstr_1 VALUES (2);`,
			},
			{
				Statement: `INSERT INTO clstr_1 VALUES (1);`,
			},
			{
				Statement: `INSERT INTO clstr_2 VALUES (2);`,
			},
			{
				Statement: `INSERT INTO clstr_2 VALUES (1);`,
			},
			{
				Statement: `INSERT INTO clstr_3 VALUES (2);`,
			},
			{
				Statement: `INSERT INTO clstr_3 VALUES (1);`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_clstr_user;`,
			},
			{
				Statement: `CLUSTER;`,
			},
			{
				Statement: `SELECT * FROM clstr_1 UNION ALL
  SELECT * FROM clstr_2 UNION ALL
  SELECT * FROM clstr_3;`,
				Results: []sql.Row{{1}, {2}, {2}, {1}, {2}, {1}},
			},
			{
				Statement: `DELETE FROM clstr_1;`,
			},
			{
				Statement: `INSERT INTO clstr_1 VALUES (2);`,
			},
			{
				Statement: `INSERT INTO clstr_1 VALUES (1);`,
			},
			{
				Statement: `CLUSTER clstr_1;`,
			},
			{
				Statement: `SELECT * FROM clstr_1;`,
				Results:   []sql.Row{{1}, {2}},
			},
			{
				Statement: `CREATE TABLE clustertest (key int PRIMARY KEY);`,
			},
			{
				Statement: `INSERT INTO clustertest VALUES (10);`,
			},
			{
				Statement: `INSERT INTO clustertest VALUES (20);`,
			},
			{
				Statement: `INSERT INTO clustertest VALUES (30);`,
			},
			{
				Statement: `INSERT INTO clustertest VALUES (40);`,
			},
			{
				Statement: `INSERT INTO clustertest VALUES (50);`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `UPDATE clustertest SET key = 100 WHERE key = 10;`,
			},
			{
				Statement: `UPDATE clustertest SET key = 35 WHERE key = 40;`,
			},
			{
				Statement: `UPDATE clustertest SET key = 60 WHERE key = 50;`,
			},
			{
				Statement: `UPDATE clustertest SET key = 70 WHERE key = 60;`,
			},
			{
				Statement: `UPDATE clustertest SET key = 80 WHERE key = 70;`,
			},
			{
				Statement: `SELECT * FROM clustertest;`,
				Results:   []sql.Row{{20}, {30}, {100}, {35}, {80}},
			},
			{
				Statement: `CLUSTER clustertest_pkey ON clustertest;`,
			},
			{
				Statement: `SELECT * FROM clustertest;`,
				Results:   []sql.Row{{20}, {30}, {35}, {80}, {100}},
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `SELECT * FROM clustertest;`,
				Results:   []sql.Row{{20}, {30}, {35}, {80}, {100}},
			},
			{
				Statement: `create temp table clstr_temp (col1 int primary key, col2 text);`,
			},
			{
				Statement: `insert into clstr_temp values (2, 'two'), (1, 'one');`,
			},
			{
				Statement: `cluster clstr_temp using clstr_temp_pkey;`,
			},
			{
				Statement: `select * from clstr_temp;`,
				Results:   []sql.Row{{1, `one`}, {2, `two`}},
			},
			{
				Statement: `drop table clstr_temp;`,
			},
			{
				Statement: `RESET SESSION AUTHORIZATION;`,
			},
			{
				Statement: `DROP TABLE clustertest;`,
			},
			{
				Statement: `CREATE TABLE clustertest (f1 int PRIMARY KEY);`,
			},
			{
				Statement: `CLUSTER clustertest USING clustertest_pkey;`,
			},
			{
				Statement: `CLUSTER clustertest;`,
			},
			{
				Statement: `CREATE TABLE clstrpart (a int) PARTITION BY RANGE (a);`,
			},
			{
				Statement: `CREATE TABLE clstrpart1 PARTITION OF clstrpart FOR VALUES FROM (1) TO (10) PARTITION BY RANGE (a);`,
			},
			{
				Statement: `CREATE TABLE clstrpart11 PARTITION OF clstrpart1 FOR VALUES FROM (1) TO (5);`,
			},
			{
				Statement: `CREATE TABLE clstrpart12 PARTITION OF clstrpart1 FOR VALUES FROM (5) TO (10) PARTITION BY RANGE (a);`,
			},
			{
				Statement: `CREATE TABLE clstrpart2 PARTITION OF clstrpart FOR VALUES FROM (10) TO (20);`,
			},
			{
				Statement: `CREATE TABLE clstrpart3 PARTITION OF clstrpart DEFAULT PARTITION BY RANGE (a);`,
			},
			{
				Statement: `CREATE TABLE clstrpart33 PARTITION OF clstrpart3 DEFAULT;`,
			},
			{
				Statement: `CREATE INDEX clstrpart_only_idx ON ONLY clstrpart (a);`,
			},
			{
				Statement:   `CLUSTER clstrpart USING clstrpart_only_idx; -- fails`,
				ErrorString: `cannot cluster on invalid index "clstrpart_only_idx"`,
			},
			{
				Statement: `DROP INDEX clstrpart_only_idx;`,
			},
			{
				Statement: `CREATE INDEX clstrpart_idx ON clstrpart (a);`,
			},
			{
				Statement: `CREATE TEMP TABLE old_cluster_info AS SELECT relname, level, relfilenode, relkind FROM pg_partition_tree('clstrpart'::regclass) AS tree JOIN pg_class c ON c.oid=tree.relid ;`,
			},
			{
				Statement: `CLUSTER clstrpart USING clstrpart_idx;`,
			},
			{
				Statement: `CREATE TEMP TABLE new_cluster_info AS SELECT relname, level, relfilenode, relkind FROM pg_partition_tree('clstrpart'::regclass) AS tree JOIN pg_class c ON c.oid=tree.relid ;`,
			},
			{
				Statement: `SELECT relname, old.level, old.relkind, old.relfilenode = new.relfilenode FROM old_cluster_info AS old JOIN new_cluster_info AS new USING (relname) ORDER BY relname COLLATE "C";`,
				Results:   []sql.Row{{`clstrpart`, 0, `p`, true}, {`clstrpart1`, 1, `p`, true}, {`clstrpart11`, 2, `r`, false}, {`clstrpart12`, 2, `p`, true}, {`clstrpart2`, 1, `r`, false}, {`clstrpart3`, 1, `p`, true}, {`clstrpart33`, 2, `r`, false}},
			},
			{
				Statement: `\d clstrpart
       Partitioned table "public.clstrpart"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
Partition key: RANGE (a)
Indexes:
    "clstrpart_idx" btree (a)
Number of partitions: 3 (Use \d+ to list them.)
CLUSTER clstrpart;`,
				ErrorString: `there is no previously clustered index for table "clstrpart"`,
			},
			{
				Statement:   `ALTER TABLE clstrpart SET WITHOUT CLUSTER;`,
				ErrorString: `cannot mark index clustered in partitioned table`,
			},
			{
				Statement:   `ALTER TABLE clstrpart CLUSTER ON clstrpart_idx;`,
				ErrorString: `cannot mark index clustered in partitioned table`,
			},
			{
				Statement: `DROP TABLE clstrpart;`,
			},
			{
				Statement: `CREATE TABLE ptnowner(i int unique) PARTITION BY LIST (i);`,
			},
			{
				Statement: `CREATE INDEX ptnowner_i_idx ON ptnowner(i);`,
			},
			{
				Statement: `CREATE TABLE ptnowner1 PARTITION OF ptnowner FOR VALUES IN (1);`,
			},
			{
				Statement: `CREATE ROLE regress_ptnowner;`,
			},
			{
				Statement: `CREATE TABLE ptnowner2 PARTITION OF ptnowner FOR VALUES IN (2);`,
			},
			{
				Statement: `ALTER TABLE ptnowner1 OWNER TO regress_ptnowner;`,
			},
			{
				Statement: `ALTER TABLE ptnowner OWNER TO regress_ptnowner;`,
			},
			{
				Statement: `CREATE TEMP TABLE ptnowner_oldnodes AS
  SELECT oid, relname, relfilenode FROM pg_partition_tree('ptnowner') AS tree
  JOIN pg_class AS c ON c.oid=tree.relid;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_ptnowner;`,
			},
			{
				Statement: `CLUSTER ptnowner USING ptnowner_i_idx;`,
			},
			{
				Statement: `RESET SESSION AUTHORIZATION;`,
			},
			{
				Statement: `SELECT a.relname, a.relfilenode=b.relfilenode FROM pg_class a
  JOIN ptnowner_oldnodes b USING (oid) ORDER BY a.relname COLLATE "C";`,
				Results: []sql.Row{{`ptnowner`, true}, {`ptnowner1`, false}, {`ptnowner2`, true}},
			},
			{
				Statement: `DROP TABLE ptnowner;`,
			},
			{
				Statement: `DROP ROLE regress_ptnowner;`,
			},
			{
				Statement: `create table clstr_4 as select * from tenk1;`,
			},
			{
				Statement: `create index cluster_sort on clstr_4 (hundred, thousand, tenthous);`,
			},
			{
				Statement: `set enable_indexscan = off;`,
			},
			{
				Statement: `set maintenance_work_mem = '1MB';`,
			},
			{
				Statement: `cluster clstr_4 using cluster_sort;`,
			},
			{
				Statement: `select * from
(select hundred, lag(hundred) over () as lhundred,
        thousand, lag(thousand) over () as lthousand,
        tenthous, lag(tenthous) over () as ltenthous from clstr_4) ss
where row(hundred, thousand, tenthous) <= row(lhundred, lthousand, ltenthous);`,
				Results: []sql.Row{},
			},
			{
				Statement: `reset enable_indexscan;`,
			},
			{
				Statement: `reset maintenance_work_mem;`,
			},
			{
				Statement: `CREATE TABLE clstr_expression(id serial primary key, a int, b text COLLATE "C");`,
			},
			{
				Statement: `INSERT INTO clstr_expression(a, b) SELECT g.i % 42, 'prefix'||g.i FROM generate_series(1, 133) g(i);`,
			},
			{
				Statement: `CREATE INDEX clstr_expression_minus_a ON clstr_expression ((-a), b);`,
			},
			{
				Statement: `CREATE INDEX clstr_expression_upper_b ON clstr_expression ((upper(b)));`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SET LOCAL enable_seqscan = false;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF) SELECT * FROM clstr_expression WHERE upper(b) = 'PREFIX3';`,
				Results:   []sql.Row{{`Index Scan using clstr_expression_upper_b on clstr_expression`}, {`Index Cond: (upper(b) = 'PREFIX3'::text)`}},
			},
			{
				Statement: `SELECT * FROM clstr_expression WHERE upper(b) = 'PREFIX3';`,
				Results:   []sql.Row{{3, 3, `prefix3`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF) SELECT * FROM clstr_expression WHERE -a = -3 ORDER BY -a, b;`,
				Results:   []sql.Row{{`Index Scan using clstr_expression_minus_a on clstr_expression`}, {`Index Cond: ((- a) = '-3'::integer)`}},
			},
			{
				Statement: `SELECT * FROM clstr_expression WHERE -a = -3 ORDER BY -a, b;`,
				Results:   []sql.Row{{129, 3, `prefix129`}, {3, 3, `prefix3`}, {45, 3, `prefix45`}, {87, 3, `prefix87`}},
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `CLUSTER clstr_expression USING clstr_expression_minus_a;`,
			},
			{
				Statement: `WITH rows AS
  (SELECT ctid, lag(a) OVER (ORDER BY ctid) AS la, a FROM clstr_expression)
SELECT * FROM rows WHERE la < a;`,
				Results: []sql.Row{},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SET LOCAL enable_seqscan = false;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF) SELECT * FROM clstr_expression WHERE upper(b) = 'PREFIX3';`,
				Results:   []sql.Row{{`Index Scan using clstr_expression_upper_b on clstr_expression`}, {`Index Cond: (upper(b) = 'PREFIX3'::text)`}},
			},
			{
				Statement: `SELECT * FROM clstr_expression WHERE upper(b) = 'PREFIX3';`,
				Results:   []sql.Row{{3, 3, `prefix3`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF) SELECT * FROM clstr_expression WHERE -a = -3 ORDER BY -a, b;`,
				Results:   []sql.Row{{`Index Scan using clstr_expression_minus_a on clstr_expression`}, {`Index Cond: ((- a) = '-3'::integer)`}},
			},
			{
				Statement: `SELECT * FROM clstr_expression WHERE -a = -3 ORDER BY -a, b;`,
				Results:   []sql.Row{{129, 3, `prefix129`}, {3, 3, `prefix3`}, {45, 3, `prefix45`}, {87, 3, `prefix87`}},
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `CLUSTER clstr_expression USING clstr_expression_upper_b;`,
			},
			{
				Statement: `WITH rows AS
  (SELECT ctid, lag(b) OVER (ORDER BY ctid) AS lb, b FROM clstr_expression)
SELECT * FROM rows WHERE upper(lb) > upper(b);`,
				Results: []sql.Row{},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SET LOCAL enable_seqscan = false;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF) SELECT * FROM clstr_expression WHERE upper(b) = 'PREFIX3';`,
				Results:   []sql.Row{{`Index Scan using clstr_expression_upper_b on clstr_expression`}, {`Index Cond: (upper(b) = 'PREFIX3'::text)`}},
			},
			{
				Statement: `SELECT * FROM clstr_expression WHERE upper(b) = 'PREFIX3';`,
				Results:   []sql.Row{{3, 3, `prefix3`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF) SELECT * FROM clstr_expression WHERE -a = -3 ORDER BY -a, b;`,
				Results:   []sql.Row{{`Index Scan using clstr_expression_minus_a on clstr_expression`}, {`Index Cond: ((- a) = '-3'::integer)`}},
			},
			{
				Statement: `SELECT * FROM clstr_expression WHERE -a = -3 ORDER BY -a, b;`,
				Results:   []sql.Row{{129, 3, `prefix129`}, {3, 3, `prefix3`}, {45, 3, `prefix45`}, {87, 3, `prefix87`}},
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `DROP TABLE clustertest;`,
			},
			{
				Statement: `DROP TABLE clstr_1;`,
			},
			{
				Statement: `DROP TABLE clstr_2;`,
			},
			{
				Statement: `DROP TABLE clstr_3;`,
			},
			{
				Statement: `DROP TABLE clstr_4;`,
			},
			{
				Statement: `DROP TABLE clstr_expression;`,
			},
			{
				Statement: `DROP USER regress_clstr_user;`,
			},
		},
	})
}
