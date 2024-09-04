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

func TestHashIndex(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_hash_index)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_hash_index,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `\getenv abs_srcdir PG_ABS_SRCDIR
CREATE TABLE hash_i4_heap (
	seqno 		int4,
	random 		int4
);`,
			},
			{
				Statement: `CREATE TABLE hash_name_heap (
	seqno 		int4,
	random 		name
);`,
			},
			{
				Statement: `CREATE TABLE hash_txt_heap (
	seqno 		int4,
	random 		text
);`,
			},
			{
				Statement: `CREATE TABLE hash_f8_heap (
	seqno		int4,
	random 		float8
);`,
			},
			{
				Statement: `\set filename :abs_srcdir '/data/hash.data'
COPY hash_i4_heap FROM :'filename';`,
			},
			{
				Statement: `COPY hash_name_heap FROM :'filename';`,
			},
			{
				Statement: `COPY hash_txt_heap FROM :'filename';`,
			},
			{
				Statement: `COPY hash_f8_heap FROM :'filename';`,
			},
			{
				Statement: `ANALYZE hash_i4_heap;`,
			},
			{
				Statement: `ANALYZE hash_name_heap;`,
			},
			{
				Statement: `ANALYZE hash_txt_heap;`,
			},
			{
				Statement: `ANALYZE hash_f8_heap;`,
			},
			{
				Statement: `CREATE INDEX hash_i4_index ON hash_i4_heap USING hash (random int4_ops);`,
			},
			{
				Statement: `CREATE INDEX hash_name_index ON hash_name_heap USING hash (random name_ops);`,
			},
			{
				Statement: `CREATE INDEX hash_txt_index ON hash_txt_heap USING hash (random text_ops);`,
			},
			{
				Statement: `CREATE INDEX hash_f8_index ON hash_f8_heap USING hash (random float8_ops)
  WITH (fillfactor=60);`,
			},
			{
				Statement: `create unique index hash_f8_index_1 on hash_f8_heap(abs(random));`,
			},
			{
				Statement: `create unique index hash_f8_index_2 on hash_f8_heap((seqno + 1), random);`,
			},
			{
				Statement: `create unique index hash_f8_index_3 on hash_f8_heap(random) where seqno > 1000;`,
			},
			{
				Statement: `SELECT * FROM hash_i4_heap
   WHERE hash_i4_heap.random = 843938989;`,
				Results: []sql.Row{{15, 843938989}},
			},
			{
				Statement: `SELECT * FROM hash_i4_heap
   WHERE hash_i4_heap.random = 66766766;`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT * FROM hash_name_heap
   WHERE hash_name_heap.random = '1505703298'::name;`,
				Results: []sql.Row{{9838, 1505703298}},
			},
			{
				Statement: `SELECT * FROM hash_name_heap
   WHERE hash_name_heap.random = '7777777'::name;`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT * FROM hash_txt_heap
   WHERE hash_txt_heap.random = '1351610853'::text;`,
				Results: []sql.Row{{5677, 1351610853}},
			},
			{
				Statement: `SELECT * FROM hash_txt_heap
   WHERE hash_txt_heap.random = '111111112222222233333333'::text;`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT * FROM hash_f8_heap
   WHERE hash_f8_heap.random = '444705537'::float8;`,
				Results: []sql.Row{{7853, 444705537}},
			},
			{
				Statement: `SELECT * FROM hash_f8_heap
   WHERE hash_f8_heap.random = '88888888'::float8;`,
				Results: []sql.Row{},
			},
			{
				Statement: `UPDATE hash_i4_heap
   SET random = 1
   WHERE hash_i4_heap.seqno = 1492;`,
			},
			{
				Statement: `SELECT h.seqno AS i1492, h.random AS i1
   FROM hash_i4_heap h
   WHERE h.random = 1;`,
				Results: []sql.Row{{1492, 1}},
			},
			{
				Statement: `UPDATE hash_i4_heap
   SET seqno = 20000
   WHERE hash_i4_heap.random = 1492795354;`,
			},
			{
				Statement: `SELECT h.seqno AS i20000
   FROM hash_i4_heap h
   WHERE h.random = 1492795354;`,
				Results: []sql.Row{{20000}},
			},
			{
				Statement: `UPDATE hash_name_heap
   SET random = '0123456789abcdef'::name
   WHERE hash_name_heap.seqno = 6543;`,
			},
			{
				Statement: `SELECT h.seqno AS i6543, h.random AS c0_to_f
   FROM hash_name_heap h
   WHERE h.random = '0123456789abcdef'::name;`,
				Results: []sql.Row{{6543, `0123456789abcdef`}},
			},
			{
				Statement: `UPDATE hash_name_heap
   SET seqno = 20000
   WHERE hash_name_heap.random = '76652222'::name;`,
			},
			{
				Statement: `SELECT h.seqno AS emptyset
   FROM hash_name_heap h
   WHERE h.random = '76652222'::name;`,
				Results: []sql.Row{},
			},
			{
				Statement: `UPDATE hash_txt_heap
   SET random = '0123456789abcdefghijklmnop'::text
   WHERE hash_txt_heap.seqno = 4002;`,
			},
			{
				Statement: `SELECT h.seqno AS i4002, h.random AS c0_to_p
   FROM hash_txt_heap h
   WHERE h.random = '0123456789abcdefghijklmnop'::text;`,
				Results: []sql.Row{{4002, `0123456789abcdefghijklmnop`}},
			},
			{
				Statement: `UPDATE hash_txt_heap
   SET seqno = 20000
   WHERE hash_txt_heap.random = '959363399'::text;`,
			},
			{
				Statement: `SELECT h.seqno AS t20000
   FROM hash_txt_heap h
   WHERE h.random = '959363399'::text;`,
				Results: []sql.Row{{20000}},
			},
			{
				Statement: `UPDATE hash_f8_heap
   SET random = '-1234.1234'::float8
   WHERE hash_f8_heap.seqno = 8906;`,
			},
			{
				Statement: `SELECT h.seqno AS i8096, h.random AS f1234_1234
   FROM hash_f8_heap h
   WHERE h.random = '-1234.1234'::float8;`,
				Results: []sql.Row{{8906, -1234.1234}},
			},
			{
				Statement: `UPDATE hash_f8_heap
   SET seqno = 20000
   WHERE hash_f8_heap.random = '488912369'::float8;`,
			},
			{
				Statement: `SELECT h.seqno AS f20000
   FROM hash_f8_heap h
   WHERE h.random = '488912369'::float8;`,
				Results: []sql.Row{{20000}},
			},
			{
				Statement: `CREATE TABLE hash_split_heap (keycol INT);`,
			},
			{
				Statement: `INSERT INTO hash_split_heap SELECT 1 FROM generate_series(1, 500) a;`,
			},
			{
				Statement: `CREATE INDEX hash_split_index on hash_split_heap USING HASH (keycol);`,
			},
			{
				Statement: `INSERT INTO hash_split_heap SELECT 1 FROM generate_series(1, 5000) a;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SET enable_seqscan = OFF;`,
			},
			{
				Statement: `SET enable_bitmapscan = OFF;`,
			},
			{
				Statement: `DECLARE c CURSOR FOR SELECT * from hash_split_heap WHERE keycol = 1;`,
			},
			{
				Statement: `MOVE FORWARD ALL FROM c;`,
			},
			{
				Statement: `MOVE BACKWARD 10000 FROM c;`,
			},
			{
				Statement: `MOVE BACKWARD ALL FROM c;`,
			},
			{
				Statement: `CLOSE c;`,
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `DELETE FROM hash_split_heap WHERE keycol = 1;`,
			},
			{
				Statement: `INSERT INTO hash_split_heap SELECT a/2 FROM generate_series(1, 25000) a;`,
			},
			{
				Statement: `VACUUM hash_split_heap;`,
			},
			{
				Statement: `ALTER INDEX hash_split_index SET (fillfactor = 10);`,
			},
			{
				Statement: `REINDEX INDEX hash_split_index;`,
			},
			{
				Statement: `DROP TABLE hash_split_heap;`,
			},
			{
				Statement: `CREATE TEMP TABLE hash_temp_heap (x int, y int);`,
			},
			{
				Statement: `INSERT INTO hash_temp_heap VALUES (1,1);`,
			},
			{
				Statement: `CREATE INDEX hash_idx ON hash_temp_heap USING hash (x);`,
			},
			{
				Statement: `DROP TABLE hash_temp_heap CASCADE;`,
			},
			{
				Statement: `CREATE TABLE hash_heap_float4 (x float4, y int);`,
			},
			{
				Statement: `INSERT INTO hash_heap_float4 VALUES (1.1,1);`,
			},
			{
				Statement: `CREATE INDEX hash_idx ON hash_heap_float4 USING hash (x);`,
			},
			{
				Statement: `DROP TABLE hash_heap_float4 CASCADE;`,
			},
			{
				Statement: `CREATE INDEX hash_f8_index2 ON hash_f8_heap USING hash (random float8_ops)
	WITH (fillfactor=9);`,
				ErrorString: `value 9 out of bounds for option "fillfactor"`,
			},
			{
				Statement: `CREATE INDEX hash_f8_index2 ON hash_f8_heap USING hash (random float8_ops)
	WITH (fillfactor=101);`,
				ErrorString: `value 101 out of bounds for option "fillfactor"`,
			},
		},
	})
}
