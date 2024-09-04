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

func TestCreateIndexSpgist(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_create_index_spgist)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_create_index_spgist,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TABLE quad_point_tbl AS
    SELECT point(unique1,unique2) AS p FROM tenk1;`,
			},
			{
				Statement: `INSERT INTO quad_point_tbl
    SELECT '(333.0,400.0)'::point FROM generate_series(1,1000);`,
			},
			{
				Statement: `INSERT INTO quad_point_tbl VALUES (NULL), (NULL), (NULL);`,
			},
			{
				Statement: `CREATE INDEX sp_quad_ind ON quad_point_tbl USING spgist (p);`,
			},
			{
				Statement: `CREATE TABLE kd_point_tbl AS SELECT * FROM quad_point_tbl;`,
			},
			{
				Statement: `CREATE INDEX sp_kd_ind ON kd_point_tbl USING spgist (p kd_point_ops);`,
			},
			{
				Statement: `CREATE TABLE radix_text_tbl AS
    SELECT name AS t FROM road WHERE name !~ '^[0-9]';`,
			},
			{
				Statement: `INSERT INTO radix_text_tbl
    SELECT 'P0123456789abcdef' FROM generate_series(1,1000);`,
			},
			{
				Statement: `INSERT INTO radix_text_tbl VALUES ('P0123456789abcde');`,
			},
			{
				Statement: `INSERT INTO radix_text_tbl VALUES ('P0123456789abcdefF');`,
			},
			{
				Statement: `CREATE INDEX sp_radix_ind ON radix_text_tbl USING spgist (t);`,
			},
			{
				Statement: `SET enable_seqscan = ON;`,
			},
			{
				Statement: `SET enable_indexscan = OFF;`,
			},
			{
				Statement: `SET enable_bitmapscan = OFF;`,
			},
			{
				Statement: `SELECT count(*) FROM quad_point_tbl WHERE p IS NULL;`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement: `SELECT count(*) FROM quad_point_tbl WHERE p IS NOT NULL;`,
				Results:   []sql.Row{{11000}},
			},
			{
				Statement: `SELECT count(*) FROM quad_point_tbl;`,
				Results:   []sql.Row{{11003}},
			},
			{
				Statement: `SELECT count(*) FROM quad_point_tbl WHERE p <@ box '(200,200,1000,1000)';`,
				Results:   []sql.Row{{1057}},
			},
			{
				Statement: `SELECT count(*) FROM quad_point_tbl WHERE box '(200,200,1000,1000)' @> p;`,
				Results:   []sql.Row{{1057}},
			},
			{
				Statement: `SELECT count(*) FROM quad_point_tbl WHERE p << '(5000, 4000)';`,
				Results:   []sql.Row{{6000}},
			},
			{
				Statement: `SELECT count(*) FROM quad_point_tbl WHERE p >> '(5000, 4000)';`,
				Results:   []sql.Row{{4999}},
			},
			{
				Statement: `SELECT count(*) FROM quad_point_tbl WHERE p <<| '(5000, 4000)';`,
				Results:   []sql.Row{{5000}},
			},
			{
				Statement: `SELECT count(*) FROM quad_point_tbl WHERE p |>> '(5000, 4000)';`,
				Results:   []sql.Row{{5999}},
			},
			{
				Statement: `SELECT count(*) FROM quad_point_tbl WHERE p ~= '(4585, 365)';`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `CREATE TEMP TABLE quad_point_tbl_ord_seq1 AS
SELECT row_number() OVER (ORDER BY p <-> '0,0') n, p <-> '0,0' dist, p
FROM quad_point_tbl;`,
			},
			{
				Statement: `CREATE TEMP TABLE quad_point_tbl_ord_seq2 AS
SELECT row_number() OVER (ORDER BY p <-> '0,0') n, p <-> '0,0' dist, p
FROM quad_point_tbl WHERE p <@ box '(200,200,1000,1000)';`,
			},
			{
				Statement: `CREATE TEMP TABLE quad_point_tbl_ord_seq3 AS
SELECT row_number() OVER (ORDER BY p <-> '333,400') n, p <-> '333,400' dist, p
FROM quad_point_tbl WHERE p IS NOT NULL;`,
			},
			{
				Statement: `SELECT count(*) FROM radix_text_tbl WHERE t = 'P0123456789abcdef';`,
				Results:   []sql.Row{{1000}},
			},
			{
				Statement: `SELECT count(*) FROM radix_text_tbl WHERE t = 'P0123456789abcde';`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT count(*) FROM radix_text_tbl WHERE t = 'P0123456789abcdefF';`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT count(*) FROM radix_text_tbl WHERE t <    'Aztec                         Ct  ';`,
				Results:   []sql.Row{{272}},
			},
			{
				Statement: `SELECT count(*) FROM radix_text_tbl WHERE t ~<~  'Aztec                         Ct  ';`,
				Results:   []sql.Row{{272}},
			},
			{
				Statement: `SELECT count(*) FROM radix_text_tbl WHERE t <=   'Aztec                         Ct  ';`,
				Results:   []sql.Row{{273}},
			},
			{
				Statement: `SELECT count(*) FROM radix_text_tbl WHERE t ~<=~ 'Aztec                         Ct  ';`,
				Results:   []sql.Row{{273}},
			},
			{
				Statement: `SELECT count(*) FROM radix_text_tbl WHERE t =    'Aztec                         Ct  ';`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT count(*) FROM radix_text_tbl WHERE t =    'Worth                         St  ';`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `SELECT count(*) FROM radix_text_tbl WHERE t >=   'Worth                         St  ';`,
				Results:   []sql.Row{{50}},
			},
			{
				Statement: `SELECT count(*) FROM radix_text_tbl WHERE t ~>=~ 'Worth                         St  ';`,
				Results:   []sql.Row{{50}},
			},
			{
				Statement: `SELECT count(*) FROM radix_text_tbl WHERE t >    'Worth                         St  ';`,
				Results:   []sql.Row{{48}},
			},
			{
				Statement: `SELECT count(*) FROM radix_text_tbl WHERE t ~>~  'Worth                         St  ';`,
				Results:   []sql.Row{{48}},
			},
			{
				Statement: `SELECT count(*) FROM radix_text_tbl WHERE t ^@  'Worth';`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `SET enable_seqscan = OFF;`,
			},
			{
				Statement: `SET enable_indexscan = ON;`,
			},
			{
				Statement: `SET enable_bitmapscan = OFF;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM quad_point_tbl WHERE p IS NULL;`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Only Scan using sp_quad_ind on quad_point_tbl`}, {`Index Cond: (p IS NULL)`}},
			},
			{
				Statement: `SELECT count(*) FROM quad_point_tbl WHERE p IS NULL;`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM quad_point_tbl WHERE p IS NOT NULL;`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Only Scan using sp_quad_ind on quad_point_tbl`}, {`Index Cond: (p IS NOT NULL)`}},
			},
			{
				Statement: `SELECT count(*) FROM quad_point_tbl WHERE p IS NOT NULL;`,
				Results:   []sql.Row{{11000}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM quad_point_tbl;`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Only Scan using sp_quad_ind on quad_point_tbl`}},
			},
			{
				Statement: `SELECT count(*) FROM quad_point_tbl;`,
				Results:   []sql.Row{{11003}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM quad_point_tbl WHERE p <@ box '(200,200,1000,1000)';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Only Scan using sp_quad_ind on quad_point_tbl`}, {`Index Cond: (p <@ '(1000,1000),(200,200)'::box)`}},
			},
			{
				Statement: `SELECT count(*) FROM quad_point_tbl WHERE p <@ box '(200,200,1000,1000)';`,
				Results:   []sql.Row{{1057}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM quad_point_tbl WHERE box '(200,200,1000,1000)' @> p;`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Only Scan using sp_quad_ind on quad_point_tbl`}, {`Index Cond: (p <@ '(1000,1000),(200,200)'::box)`}},
			},
			{
				Statement: `SELECT count(*) FROM quad_point_tbl WHERE box '(200,200,1000,1000)' @> p;`,
				Results:   []sql.Row{{1057}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM quad_point_tbl WHERE p << '(5000, 4000)';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Only Scan using sp_quad_ind on quad_point_tbl`}, {`Index Cond: (p << '(5000,4000)'::point)`}},
			},
			{
				Statement: `SELECT count(*) FROM quad_point_tbl WHERE p << '(5000, 4000)';`,
				Results:   []sql.Row{{6000}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM quad_point_tbl WHERE p >> '(5000, 4000)';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Only Scan using sp_quad_ind on quad_point_tbl`}, {`Index Cond: (p >> '(5000,4000)'::point)`}},
			},
			{
				Statement: `SELECT count(*) FROM quad_point_tbl WHERE p >> '(5000, 4000)';`,
				Results:   []sql.Row{{4999}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM quad_point_tbl WHERE p <<| '(5000, 4000)';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Only Scan using sp_quad_ind on quad_point_tbl`}, {`Index Cond: (p <<| '(5000,4000)'::point)`}},
			},
			{
				Statement: `SELECT count(*) FROM quad_point_tbl WHERE p <<| '(5000, 4000)';`,
				Results:   []sql.Row{{5000}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM quad_point_tbl WHERE p |>> '(5000, 4000)';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Only Scan using sp_quad_ind on quad_point_tbl`}, {`Index Cond: (p |>> '(5000,4000)'::point)`}},
			},
			{
				Statement: `SELECT count(*) FROM quad_point_tbl WHERE p |>> '(5000, 4000)';`,
				Results:   []sql.Row{{5999}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM quad_point_tbl WHERE p ~= '(4585, 365)';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Only Scan using sp_quad_ind on quad_point_tbl`}, {`Index Cond: (p ~= '(4585,365)'::point)`}},
			},
			{
				Statement: `SELECT count(*) FROM quad_point_tbl WHERE p ~= '(4585, 365)';`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT row_number() OVER (ORDER BY p <-> '0,0') n, p <-> '0,0' dist, p
FROM quad_point_tbl;`,
				Results: []sql.Row{{`WindowAgg`}, {`->  Index Only Scan using sp_quad_ind on quad_point_tbl`}, {`Order By: (p <-> '(0,0)'::point)`}},
			},
			{
				Statement: `CREATE TEMP TABLE quad_point_tbl_ord_idx1 AS
SELECT row_number() OVER (ORDER BY p <-> '0,0') n, p <-> '0,0' dist, p
FROM quad_point_tbl;`,
			},
			{
				Statement: `SELECT * FROM quad_point_tbl_ord_seq1 seq FULL JOIN quad_point_tbl_ord_idx1 idx
ON seq.n = idx.n
WHERE seq.dist IS DISTINCT FROM idx.dist;`,
				Results: []sql.Row{},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT row_number() OVER (ORDER BY p <-> '0,0') n, p <-> '0,0' dist, p
FROM quad_point_tbl WHERE p <@ box '(200,200,1000,1000)';`,
				Results: []sql.Row{{`WindowAgg`}, {`->  Index Only Scan using sp_quad_ind on quad_point_tbl`}, {`Index Cond: (p <@ '(1000,1000),(200,200)'::box)`}, {`Order By: (p <-> '(0,0)'::point)`}},
			},
			{
				Statement: `CREATE TEMP TABLE quad_point_tbl_ord_idx2 AS
SELECT row_number() OVER (ORDER BY p <-> '0,0') n, p <-> '0,0' dist, p
FROM quad_point_tbl WHERE p <@ box '(200,200,1000,1000)';`,
			},
			{
				Statement: `SELECT * FROM quad_point_tbl_ord_seq2 seq FULL JOIN quad_point_tbl_ord_idx2 idx
ON seq.n = idx.n
WHERE seq.dist IS DISTINCT FROM idx.dist;`,
				Results: []sql.Row{},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT row_number() OVER (ORDER BY p <-> '333,400') n, p <-> '333,400' dist, p
FROM quad_point_tbl WHERE p IS NOT NULL;`,
				Results: []sql.Row{{`WindowAgg`}, {`->  Index Only Scan using sp_quad_ind on quad_point_tbl`}, {`Index Cond: (p IS NOT NULL)`}, {`Order By: (p <-> '(333,400)'::point)`}},
			},
			{
				Statement: `CREATE TEMP TABLE quad_point_tbl_ord_idx3 AS
SELECT row_number() OVER (ORDER BY p <-> '333,400') n, p <-> '333,400' dist, p
FROM quad_point_tbl WHERE p IS NOT NULL;`,
			},
			{
				Statement: `SELECT * FROM quad_point_tbl_ord_seq3 seq FULL JOIN quad_point_tbl_ord_idx3 idx
ON seq.n = idx.n
WHERE seq.dist IS DISTINCT FROM idx.dist;`,
				Results: []sql.Row{},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM kd_point_tbl WHERE p <@ box '(200,200,1000,1000)';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Only Scan using sp_kd_ind on kd_point_tbl`}, {`Index Cond: (p <@ '(1000,1000),(200,200)'::box)`}},
			},
			{
				Statement: `SELECT count(*) FROM kd_point_tbl WHERE p <@ box '(200,200,1000,1000)';`,
				Results:   []sql.Row{{1057}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM kd_point_tbl WHERE box '(200,200,1000,1000)' @> p;`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Only Scan using sp_kd_ind on kd_point_tbl`}, {`Index Cond: (p <@ '(1000,1000),(200,200)'::box)`}},
			},
			{
				Statement: `SELECT count(*) FROM kd_point_tbl WHERE box '(200,200,1000,1000)' @> p;`,
				Results:   []sql.Row{{1057}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM kd_point_tbl WHERE p << '(5000, 4000)';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Only Scan using sp_kd_ind on kd_point_tbl`}, {`Index Cond: (p << '(5000,4000)'::point)`}},
			},
			{
				Statement: `SELECT count(*) FROM kd_point_tbl WHERE p << '(5000, 4000)';`,
				Results:   []sql.Row{{6000}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM kd_point_tbl WHERE p >> '(5000, 4000)';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Only Scan using sp_kd_ind on kd_point_tbl`}, {`Index Cond: (p >> '(5000,4000)'::point)`}},
			},
			{
				Statement: `SELECT count(*) FROM kd_point_tbl WHERE p >> '(5000, 4000)';`,
				Results:   []sql.Row{{4999}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM kd_point_tbl WHERE p <<| '(5000, 4000)';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Only Scan using sp_kd_ind on kd_point_tbl`}, {`Index Cond: (p <<| '(5000,4000)'::point)`}},
			},
			{
				Statement: `SELECT count(*) FROM kd_point_tbl WHERE p <<| '(5000, 4000)';`,
				Results:   []sql.Row{{5000}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM kd_point_tbl WHERE p |>> '(5000, 4000)';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Only Scan using sp_kd_ind on kd_point_tbl`}, {`Index Cond: (p |>> '(5000,4000)'::point)`}},
			},
			{
				Statement: `SELECT count(*) FROM kd_point_tbl WHERE p |>> '(5000, 4000)';`,
				Results:   []sql.Row{{5999}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM kd_point_tbl WHERE p ~= '(4585, 365)';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Only Scan using sp_kd_ind on kd_point_tbl`}, {`Index Cond: (p ~= '(4585,365)'::point)`}},
			},
			{
				Statement: `SELECT count(*) FROM kd_point_tbl WHERE p ~= '(4585, 365)';`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT row_number() OVER (ORDER BY p <-> '0,0') n, p <-> '0,0' dist, p
FROM kd_point_tbl;`,
				Results: []sql.Row{{`WindowAgg`}, {`->  Index Only Scan using sp_kd_ind on kd_point_tbl`}, {`Order By: (p <-> '(0,0)'::point)`}},
			},
			{
				Statement: `CREATE TEMP TABLE kd_point_tbl_ord_idx1 AS
SELECT row_number() OVER (ORDER BY p <-> '0,0') n, p <-> '0,0' dist, p
FROM kd_point_tbl;`,
			},
			{
				Statement: `SELECT * FROM quad_point_tbl_ord_seq1 seq FULL JOIN kd_point_tbl_ord_idx1 idx
ON seq.n = idx.n
WHERE seq.dist IS DISTINCT FROM idx.dist;`,
				Results: []sql.Row{},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT row_number() OVER (ORDER BY p <-> '0,0') n, p <-> '0,0' dist, p
FROM kd_point_tbl WHERE p <@ box '(200,200,1000,1000)';`,
				Results: []sql.Row{{`WindowAgg`}, {`->  Index Only Scan using sp_kd_ind on kd_point_tbl`}, {`Index Cond: (p <@ '(1000,1000),(200,200)'::box)`}, {`Order By: (p <-> '(0,0)'::point)`}},
			},
			{
				Statement: `CREATE TEMP TABLE kd_point_tbl_ord_idx2 AS
SELECT row_number() OVER (ORDER BY p <-> '0,0') n, p <-> '0,0' dist, p
FROM kd_point_tbl WHERE p <@ box '(200,200,1000,1000)';`,
			},
			{
				Statement: `SELECT * FROM quad_point_tbl_ord_seq2 seq FULL JOIN kd_point_tbl_ord_idx2 idx
ON seq.n = idx.n
WHERE seq.dist IS DISTINCT FROM idx.dist;`,
				Results: []sql.Row{},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT row_number() OVER (ORDER BY p <-> '333,400') n, p <-> '333,400' dist, p
FROM kd_point_tbl WHERE p IS NOT NULL;`,
				Results: []sql.Row{{`WindowAgg`}, {`->  Index Only Scan using sp_kd_ind on kd_point_tbl`}, {`Index Cond: (p IS NOT NULL)`}, {`Order By: (p <-> '(333,400)'::point)`}},
			},
			{
				Statement: `CREATE TEMP TABLE kd_point_tbl_ord_idx3 AS
SELECT row_number() OVER (ORDER BY p <-> '333,400') n, p <-> '333,400' dist, p
FROM kd_point_tbl WHERE p IS NOT NULL;`,
			},
			{
				Statement: `SELECT * FROM quad_point_tbl_ord_seq3 seq FULL JOIN kd_point_tbl_ord_idx3 idx
ON seq.n = idx.n
WHERE seq.dist IS DISTINCT FROM idx.dist;`,
				Results: []sql.Row{},
			},
			{
				Statement: `SET extra_float_digits = 0;`,
			},
			{
				Statement: `CREATE INDEX ON quad_point_tbl_ord_seq1 USING spgist(p) INCLUDE(dist);`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT p, dist FROM quad_point_tbl_ord_seq1 ORDER BY p <-> '0,0' LIMIT 10;`,
				Results: []sql.Row{{`Limit`}, {`->  Index Only Scan using quad_point_tbl_ord_seq1_p_dist_idx on quad_point_tbl_ord_seq1`}, {`Order By: (p <-> '(0,0)'::point)`}},
			},
			{
				Statement: `SELECT p, dist FROM quad_point_tbl_ord_seq1 ORDER BY p <-> '0,0' LIMIT 10;`,
				Results:   []sql.Row{{`(59,21)`, 62.6258732474047}, {`(88,104)`, 136.235090927411}, {`(39,143)`, 148.222805262888}, {`(139,160)`, 211.945747775227}, {`(209,38)`, 212.42645786248}, {`(157,156)`, 221.325552072055}, {`(175,150)`, 230.488611432322}, {`(236,34)`, 238.436574375661}, {`(263,28)`, 264.486294540946}, {`(322,53)`, 326.33265236565}},
			},
			{
				Statement: `RESET extra_float_digits;`,
			},
			{
				Statement: `SELECT (SELECT p FROM kd_point_tbl ORDER BY p <-> pt, p <-> '0,0' LIMIT 1)
FROM (VALUES (point '1,2'), (NULL), ('1234,5678')) pts(pt);`,
				Results: []sql.Row{{`(59,21)`}, {`(59,21)`}, {`(1239,5647)`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM radix_text_tbl WHERE t = 'P0123456789abcdef';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Only Scan using sp_radix_ind on radix_text_tbl`}, {`Index Cond: (t = 'P0123456789abcdef'::text)`}},
			},
			{
				Statement: `SELECT count(*) FROM radix_text_tbl WHERE t = 'P0123456789abcdef';`,
				Results:   []sql.Row{{1000}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM radix_text_tbl WHERE t = 'P0123456789abcde';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Only Scan using sp_radix_ind on radix_text_tbl`}, {`Index Cond: (t = 'P0123456789abcde'::text)`}},
			},
			{
				Statement: `SELECT count(*) FROM radix_text_tbl WHERE t = 'P0123456789abcde';`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM radix_text_tbl WHERE t = 'P0123456789abcdefF';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Only Scan using sp_radix_ind on radix_text_tbl`}, {`Index Cond: (t = 'P0123456789abcdefF'::text)`}},
			},
			{
				Statement: `SELECT count(*) FROM radix_text_tbl WHERE t = 'P0123456789abcdefF';`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM radix_text_tbl WHERE t <    'Aztec                         Ct  ';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Only Scan using sp_radix_ind on radix_text_tbl`}, {`Index Cond: (t < 'Aztec                         Ct  '::text)`}},
			},
			{
				Statement: `SELECT count(*) FROM radix_text_tbl WHERE t <    'Aztec                         Ct  ';`,
				Results:   []sql.Row{{272}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM radix_text_tbl WHERE t ~<~  'Aztec                         Ct  ';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Only Scan using sp_radix_ind on radix_text_tbl`}, {`Index Cond: (t ~<~ 'Aztec                         Ct  '::text)`}},
			},
			{
				Statement: `SELECT count(*) FROM radix_text_tbl WHERE t ~<~  'Aztec                         Ct  ';`,
				Results:   []sql.Row{{272}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM radix_text_tbl WHERE t <=   'Aztec                         Ct  ';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Only Scan using sp_radix_ind on radix_text_tbl`}, {`Index Cond: (t <= 'Aztec                         Ct  '::text)`}},
			},
			{
				Statement: `SELECT count(*) FROM radix_text_tbl WHERE t <=   'Aztec                         Ct  ';`,
				Results:   []sql.Row{{273}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM radix_text_tbl WHERE t ~<=~ 'Aztec                         Ct  ';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Only Scan using sp_radix_ind on radix_text_tbl`}, {`Index Cond: (t ~<=~ 'Aztec                         Ct  '::text)`}},
			},
			{
				Statement: `SELECT count(*) FROM radix_text_tbl WHERE t ~<=~ 'Aztec                         Ct  ';`,
				Results:   []sql.Row{{273}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM radix_text_tbl WHERE t =    'Aztec                         Ct  ';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Only Scan using sp_radix_ind on radix_text_tbl`}, {`Index Cond: (t = 'Aztec                         Ct  '::text)`}},
			},
			{
				Statement: `SELECT count(*) FROM radix_text_tbl WHERE t =    'Aztec                         Ct  ';`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM radix_text_tbl WHERE t =    'Worth                         St  ';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Only Scan using sp_radix_ind on radix_text_tbl`}, {`Index Cond: (t = 'Worth                         St  '::text)`}},
			},
			{
				Statement: `SELECT count(*) FROM radix_text_tbl WHERE t =    'Worth                         St  ';`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM radix_text_tbl WHERE t >=   'Worth                         St  ';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Only Scan using sp_radix_ind on radix_text_tbl`}, {`Index Cond: (t >= 'Worth                         St  '::text)`}},
			},
			{
				Statement: `SELECT count(*) FROM radix_text_tbl WHERE t >=   'Worth                         St  ';`,
				Results:   []sql.Row{{50}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM radix_text_tbl WHERE t ~>=~ 'Worth                         St  ';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Only Scan using sp_radix_ind on radix_text_tbl`}, {`Index Cond: (t ~>=~ 'Worth                         St  '::text)`}},
			},
			{
				Statement: `SELECT count(*) FROM radix_text_tbl WHERE t ~>=~ 'Worth                         St  ';`,
				Results:   []sql.Row{{50}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM radix_text_tbl WHERE t >    'Worth                         St  ';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Only Scan using sp_radix_ind on radix_text_tbl`}, {`Index Cond: (t > 'Worth                         St  '::text)`}},
			},
			{
				Statement: `SELECT count(*) FROM radix_text_tbl WHERE t >    'Worth                         St  ';`,
				Results:   []sql.Row{{48}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM radix_text_tbl WHERE t ~>~  'Worth                         St  ';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Only Scan using sp_radix_ind on radix_text_tbl`}, {`Index Cond: (t ~>~ 'Worth                         St  '::text)`}},
			},
			{
				Statement: `SELECT count(*) FROM radix_text_tbl WHERE t ~>~  'Worth                         St  ';`,
				Results:   []sql.Row{{48}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM radix_text_tbl WHERE t ^@	 'Worth';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Only Scan using sp_radix_ind on radix_text_tbl`}, {`Index Cond: (t ^@ 'Worth'::text)`}},
			},
			{
				Statement: `SELECT count(*) FROM radix_text_tbl WHERE t ^@	 'Worth';`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM radix_text_tbl WHERE starts_with(t, 'Worth');`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Only Scan using sp_radix_ind on radix_text_tbl`}, {`Index Cond: (t ^@ 'Worth'::text)`}, {`Filter: starts_with(t, 'Worth'::text)`}},
			},
			{
				Statement: `SELECT count(*) FROM radix_text_tbl WHERE starts_with(t, 'Worth');`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `SET enable_seqscan = OFF;`,
			},
			{
				Statement: `SET enable_indexscan = OFF;`,
			},
			{
				Statement: `SET enable_bitmapscan = ON;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM quad_point_tbl WHERE p IS NULL;`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on quad_point_tbl`}, {`Recheck Cond: (p IS NULL)`}, {`->  Bitmap Index Scan on sp_quad_ind`}, {`Index Cond: (p IS NULL)`}},
			},
			{
				Statement: `SELECT count(*) FROM quad_point_tbl WHERE p IS NULL;`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM quad_point_tbl WHERE p IS NOT NULL;`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on quad_point_tbl`}, {`Recheck Cond: (p IS NOT NULL)`}, {`->  Bitmap Index Scan on sp_quad_ind`}, {`Index Cond: (p IS NOT NULL)`}},
			},
			{
				Statement: `SELECT count(*) FROM quad_point_tbl WHERE p IS NOT NULL;`,
				Results:   []sql.Row{{11000}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM quad_point_tbl;`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on quad_point_tbl`}, {`->  Bitmap Index Scan on sp_quad_ind`}},
			},
			{
				Statement: `SELECT count(*) FROM quad_point_tbl;`,
				Results:   []sql.Row{{11003}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM quad_point_tbl WHERE p <@ box '(200,200,1000,1000)';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on quad_point_tbl`}, {`Recheck Cond: (p <@ '(1000,1000),(200,200)'::box)`}, {`->  Bitmap Index Scan on sp_quad_ind`}, {`Index Cond: (p <@ '(1000,1000),(200,200)'::box)`}},
			},
			{
				Statement: `SELECT count(*) FROM quad_point_tbl WHERE p <@ box '(200,200,1000,1000)';`,
				Results:   []sql.Row{{1057}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM quad_point_tbl WHERE box '(200,200,1000,1000)' @> p;`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on quad_point_tbl`}, {`Recheck Cond: ('(1000,1000),(200,200)'::box @> p)`}, {`->  Bitmap Index Scan on sp_quad_ind`}, {`Index Cond: (p <@ '(1000,1000),(200,200)'::box)`}},
			},
			{
				Statement: `SELECT count(*) FROM quad_point_tbl WHERE box '(200,200,1000,1000)' @> p;`,
				Results:   []sql.Row{{1057}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM quad_point_tbl WHERE p << '(5000, 4000)';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on quad_point_tbl`}, {`Recheck Cond: (p << '(5000,4000)'::point)`}, {`->  Bitmap Index Scan on sp_quad_ind`}, {`Index Cond: (p << '(5000,4000)'::point)`}},
			},
			{
				Statement: `SELECT count(*) FROM quad_point_tbl WHERE p << '(5000, 4000)';`,
				Results:   []sql.Row{{6000}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM quad_point_tbl WHERE p >> '(5000, 4000)';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on quad_point_tbl`}, {`Recheck Cond: (p >> '(5000,4000)'::point)`}, {`->  Bitmap Index Scan on sp_quad_ind`}, {`Index Cond: (p >> '(5000,4000)'::point)`}},
			},
			{
				Statement: `SELECT count(*) FROM quad_point_tbl WHERE p >> '(5000, 4000)';`,
				Results:   []sql.Row{{4999}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM quad_point_tbl WHERE p <<| '(5000, 4000)';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on quad_point_tbl`}, {`Recheck Cond: (p <<| '(5000,4000)'::point)`}, {`->  Bitmap Index Scan on sp_quad_ind`}, {`Index Cond: (p <<| '(5000,4000)'::point)`}},
			},
			{
				Statement: `SELECT count(*) FROM quad_point_tbl WHERE p <<| '(5000, 4000)';`,
				Results:   []sql.Row{{5000}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM quad_point_tbl WHERE p |>> '(5000, 4000)';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on quad_point_tbl`}, {`Recheck Cond: (p |>> '(5000,4000)'::point)`}, {`->  Bitmap Index Scan on sp_quad_ind`}, {`Index Cond: (p |>> '(5000,4000)'::point)`}},
			},
			{
				Statement: `SELECT count(*) FROM quad_point_tbl WHERE p |>> '(5000, 4000)';`,
				Results:   []sql.Row{{5999}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM quad_point_tbl WHERE p ~= '(4585, 365)';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on quad_point_tbl`}, {`Recheck Cond: (p ~= '(4585,365)'::point)`}, {`->  Bitmap Index Scan on sp_quad_ind`}, {`Index Cond: (p ~= '(4585,365)'::point)`}},
			},
			{
				Statement: `SELECT count(*) FROM quad_point_tbl WHERE p ~= '(4585, 365)';`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM kd_point_tbl WHERE p <@ box '(200,200,1000,1000)';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on kd_point_tbl`}, {`Recheck Cond: (p <@ '(1000,1000),(200,200)'::box)`}, {`->  Bitmap Index Scan on sp_kd_ind`}, {`Index Cond: (p <@ '(1000,1000),(200,200)'::box)`}},
			},
			{
				Statement: `SELECT count(*) FROM kd_point_tbl WHERE p <@ box '(200,200,1000,1000)';`,
				Results:   []sql.Row{{1057}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM kd_point_tbl WHERE box '(200,200,1000,1000)' @> p;`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on kd_point_tbl`}, {`Recheck Cond: ('(1000,1000),(200,200)'::box @> p)`}, {`->  Bitmap Index Scan on sp_kd_ind`}, {`Index Cond: (p <@ '(1000,1000),(200,200)'::box)`}},
			},
			{
				Statement: `SELECT count(*) FROM kd_point_tbl WHERE box '(200,200,1000,1000)' @> p;`,
				Results:   []sql.Row{{1057}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM kd_point_tbl WHERE p << '(5000, 4000)';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on kd_point_tbl`}, {`Recheck Cond: (p << '(5000,4000)'::point)`}, {`->  Bitmap Index Scan on sp_kd_ind`}, {`Index Cond: (p << '(5000,4000)'::point)`}},
			},
			{
				Statement: `SELECT count(*) FROM kd_point_tbl WHERE p << '(5000, 4000)';`,
				Results:   []sql.Row{{6000}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM kd_point_tbl WHERE p >> '(5000, 4000)';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on kd_point_tbl`}, {`Recheck Cond: (p >> '(5000,4000)'::point)`}, {`->  Bitmap Index Scan on sp_kd_ind`}, {`Index Cond: (p >> '(5000,4000)'::point)`}},
			},
			{
				Statement: `SELECT count(*) FROM kd_point_tbl WHERE p >> '(5000, 4000)';`,
				Results:   []sql.Row{{4999}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM kd_point_tbl WHERE p <<| '(5000, 4000)';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on kd_point_tbl`}, {`Recheck Cond: (p <<| '(5000,4000)'::point)`}, {`->  Bitmap Index Scan on sp_kd_ind`}, {`Index Cond: (p <<| '(5000,4000)'::point)`}},
			},
			{
				Statement: `SELECT count(*) FROM kd_point_tbl WHERE p <<| '(5000, 4000)';`,
				Results:   []sql.Row{{5000}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM kd_point_tbl WHERE p |>> '(5000, 4000)';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on kd_point_tbl`}, {`Recheck Cond: (p |>> '(5000,4000)'::point)`}, {`->  Bitmap Index Scan on sp_kd_ind`}, {`Index Cond: (p |>> '(5000,4000)'::point)`}},
			},
			{
				Statement: `SELECT count(*) FROM kd_point_tbl WHERE p |>> '(5000, 4000)';`,
				Results:   []sql.Row{{5999}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM kd_point_tbl WHERE p ~= '(4585, 365)';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on kd_point_tbl`}, {`Recheck Cond: (p ~= '(4585,365)'::point)`}, {`->  Bitmap Index Scan on sp_kd_ind`}, {`Index Cond: (p ~= '(4585,365)'::point)`}},
			},
			{
				Statement: `SELECT count(*) FROM kd_point_tbl WHERE p ~= '(4585, 365)';`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM radix_text_tbl WHERE t = 'P0123456789abcdef';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on radix_text_tbl`}, {`Recheck Cond: (t = 'P0123456789abcdef'::text)`}, {`->  Bitmap Index Scan on sp_radix_ind`}, {`Index Cond: (t = 'P0123456789abcdef'::text)`}},
			},
			{
				Statement: `SELECT count(*) FROM radix_text_tbl WHERE t = 'P0123456789abcdef';`,
				Results:   []sql.Row{{1000}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM radix_text_tbl WHERE t = 'P0123456789abcde';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on radix_text_tbl`}, {`Recheck Cond: (t = 'P0123456789abcde'::text)`}, {`->  Bitmap Index Scan on sp_radix_ind`}, {`Index Cond: (t = 'P0123456789abcde'::text)`}},
			},
			{
				Statement: `SELECT count(*) FROM radix_text_tbl WHERE t = 'P0123456789abcde';`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM radix_text_tbl WHERE t = 'P0123456789abcdefF';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on radix_text_tbl`}, {`Recheck Cond: (t = 'P0123456789abcdefF'::text)`}, {`->  Bitmap Index Scan on sp_radix_ind`}, {`Index Cond: (t = 'P0123456789abcdefF'::text)`}},
			},
			{
				Statement: `SELECT count(*) FROM radix_text_tbl WHERE t = 'P0123456789abcdefF';`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM radix_text_tbl WHERE t <    'Aztec                         Ct  ';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on radix_text_tbl`}, {`Recheck Cond: (t < 'Aztec                         Ct  '::text)`}, {`->  Bitmap Index Scan on sp_radix_ind`}, {`Index Cond: (t < 'Aztec                         Ct  '::text)`}},
			},
			{
				Statement: `SELECT count(*) FROM radix_text_tbl WHERE t <    'Aztec                         Ct  ';`,
				Results:   []sql.Row{{272}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM radix_text_tbl WHERE t ~<~  'Aztec                         Ct  ';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on radix_text_tbl`}, {`Recheck Cond: (t ~<~ 'Aztec                         Ct  '::text)`}, {`->  Bitmap Index Scan on sp_radix_ind`}, {`Index Cond: (t ~<~ 'Aztec                         Ct  '::text)`}},
			},
			{
				Statement: `SELECT count(*) FROM radix_text_tbl WHERE t ~<~  'Aztec                         Ct  ';`,
				Results:   []sql.Row{{272}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM radix_text_tbl WHERE t <=   'Aztec                         Ct  ';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on radix_text_tbl`}, {`Recheck Cond: (t <= 'Aztec                         Ct  '::text)`}, {`->  Bitmap Index Scan on sp_radix_ind`}, {`Index Cond: (t <= 'Aztec                         Ct  '::text)`}},
			},
			{
				Statement: `SELECT count(*) FROM radix_text_tbl WHERE t <=   'Aztec                         Ct  ';`,
				Results:   []sql.Row{{273}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM radix_text_tbl WHERE t ~<=~ 'Aztec                         Ct  ';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on radix_text_tbl`}, {`Recheck Cond: (t ~<=~ 'Aztec                         Ct  '::text)`}, {`->  Bitmap Index Scan on sp_radix_ind`}, {`Index Cond: (t ~<=~ 'Aztec                         Ct  '::text)`}},
			},
			{
				Statement: `SELECT count(*) FROM radix_text_tbl WHERE t ~<=~ 'Aztec                         Ct  ';`,
				Results:   []sql.Row{{273}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM radix_text_tbl WHERE t =    'Aztec                         Ct  ';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on radix_text_tbl`}, {`Recheck Cond: (t = 'Aztec                         Ct  '::text)`}, {`->  Bitmap Index Scan on sp_radix_ind`}, {`Index Cond: (t = 'Aztec                         Ct  '::text)`}},
			},
			{
				Statement: `SELECT count(*) FROM radix_text_tbl WHERE t =    'Aztec                         Ct  ';`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM radix_text_tbl WHERE t =    'Worth                         St  ';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on radix_text_tbl`}, {`Recheck Cond: (t = 'Worth                         St  '::text)`}, {`->  Bitmap Index Scan on sp_radix_ind`}, {`Index Cond: (t = 'Worth                         St  '::text)`}},
			},
			{
				Statement: `SELECT count(*) FROM radix_text_tbl WHERE t =    'Worth                         St  ';`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM radix_text_tbl WHERE t >=   'Worth                         St  ';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on radix_text_tbl`}, {`Recheck Cond: (t >= 'Worth                         St  '::text)`}, {`->  Bitmap Index Scan on sp_radix_ind`}, {`Index Cond: (t >= 'Worth                         St  '::text)`}},
			},
			{
				Statement: `SELECT count(*) FROM radix_text_tbl WHERE t >=   'Worth                         St  ';`,
				Results:   []sql.Row{{50}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM radix_text_tbl WHERE t ~>=~ 'Worth                         St  ';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on radix_text_tbl`}, {`Recheck Cond: (t ~>=~ 'Worth                         St  '::text)`}, {`->  Bitmap Index Scan on sp_radix_ind`}, {`Index Cond: (t ~>=~ 'Worth                         St  '::text)`}},
			},
			{
				Statement: `SELECT count(*) FROM radix_text_tbl WHERE t ~>=~ 'Worth                         St  ';`,
				Results:   []sql.Row{{50}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM radix_text_tbl WHERE t >    'Worth                         St  ';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on radix_text_tbl`}, {`Recheck Cond: (t > 'Worth                         St  '::text)`}, {`->  Bitmap Index Scan on sp_radix_ind`}, {`Index Cond: (t > 'Worth                         St  '::text)`}},
			},
			{
				Statement: `SELECT count(*) FROM radix_text_tbl WHERE t >    'Worth                         St  ';`,
				Results:   []sql.Row{{48}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM radix_text_tbl WHERE t ~>~  'Worth                         St  ';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on radix_text_tbl`}, {`Recheck Cond: (t ~>~ 'Worth                         St  '::text)`}, {`->  Bitmap Index Scan on sp_radix_ind`}, {`Index Cond: (t ~>~ 'Worth                         St  '::text)`}},
			},
			{
				Statement: `SELECT count(*) FROM radix_text_tbl WHERE t ~>~  'Worth                         St  ';`,
				Results:   []sql.Row{{48}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM radix_text_tbl WHERE t ^@	 'Worth';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on radix_text_tbl`}, {`Recheck Cond: (t ^@ 'Worth'::text)`}, {`->  Bitmap Index Scan on sp_radix_ind`}, {`Index Cond: (t ^@ 'Worth'::text)`}},
			},
			{
				Statement: `SELECT count(*) FROM radix_text_tbl WHERE t ^@	 'Worth';`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM radix_text_tbl WHERE starts_with(t, 'Worth');`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on radix_text_tbl`}, {`Filter: starts_with(t, 'Worth'::text)`}, {`->  Bitmap Index Scan on sp_radix_ind`}, {`Index Cond: (t ^@ 'Worth'::text)`}},
			},
			{
				Statement: `SELECT count(*) FROM radix_text_tbl WHERE starts_with(t, 'Worth');`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `RESET enable_seqscan;`,
			},
			{
				Statement: `RESET enable_indexscan;`,
			},
			{
				Statement: `RESET enable_bitmapscan;`,
			},
		},
	})
}
