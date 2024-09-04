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

func TestLimit(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_limit)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_limit,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `SELECT ''::text AS two, unique1, unique2, stringu1
		FROM onek WHERE unique1 > 50
		ORDER BY unique1 LIMIT 2;`,
				Results: []sql.Row{{``, 51, 76, `ZBAAAA`}, {``, 52, 985, `ACAAAA`}},
			},
			{
				Statement: `SELECT ''::text AS five, unique1, unique2, stringu1
		FROM onek WHERE unique1 > 60
		ORDER BY unique1 LIMIT 5;`,
				Results: []sql.Row{{``, 61, 560, `JCAAAA`}, {``, 62, 633, `KCAAAA`}, {``, 63, 296, `LCAAAA`}, {``, 64, 479, `MCAAAA`}, {``, 65, 64, `NCAAAA`}},
			},
			{
				Statement: `SELECT ''::text AS two, unique1, unique2, stringu1
		FROM onek WHERE unique1 > 60 AND unique1 < 63
		ORDER BY unique1 LIMIT 5;`,
				Results: []sql.Row{{``, 61, 560, `JCAAAA`}, {``, 62, 633, `KCAAAA`}},
			},
			{
				Statement: `SELECT ''::text AS three, unique1, unique2, stringu1
		FROM onek WHERE unique1 > 100
		ORDER BY unique1 LIMIT 3 OFFSET 20;`,
				Results: []sql.Row{{``, 121, 700, `REAAAA`}, {``, 122, 519, `SEAAAA`}, {``, 123, 777, `TEAAAA`}},
			},
			{
				Statement: `SELECT ''::text AS zero, unique1, unique2, stringu1
		FROM onek WHERE unique1 < 50
		ORDER BY unique1 DESC LIMIT 8 OFFSET 99;`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT ''::text AS eleven, unique1, unique2, stringu1
		FROM onek WHERE unique1 < 50
		ORDER BY unique1 DESC LIMIT 20 OFFSET 39;`,
				Results: []sql.Row{{``, 10, 520, `KAAAAA`}, {``, 9, 49, `JAAAAA`}, {``, 8, 653, `IAAAAA`}, {``, 7, 647, `HAAAAA`}, {``, 6, 978, `GAAAAA`}, {``, 5, 541, `FAAAAA`}, {``, 4, 833, `EAAAAA`}, {``, 3, 431, `DAAAAA`}, {``, 2, 326, `CAAAAA`}, {``, 1, 214, `BAAAAA`}, {``, 0, 998, `AAAAAA`}},
			},
			{
				Statement: `SELECT ''::text AS ten, unique1, unique2, stringu1
		FROM onek
		ORDER BY unique1 OFFSET 990;`,
				Results: []sql.Row{{``, 990, 369, `CMAAAA`}, {``, 991, 426, `DMAAAA`}, {``, 992, 363, `EMAAAA`}, {``, 993, 661, `FMAAAA`}, {``, 994, 695, `GMAAAA`}, {``, 995, 144, `HMAAAA`}, {``, 996, 258, `IMAAAA`}, {``, 997, 21, `JMAAAA`}, {``, 998, 549, `KMAAAA`}, {``, 999, 152, `LMAAAA`}},
			},
			{
				Statement: `SELECT ''::text AS five, unique1, unique2, stringu1
		FROM onek
		ORDER BY unique1 OFFSET 990 LIMIT 5;`,
				Results: []sql.Row{{``, 990, 369, `CMAAAA`}, {``, 991, 426, `DMAAAA`}, {``, 992, 363, `EMAAAA`}, {``, 993, 661, `FMAAAA`}, {``, 994, 695, `GMAAAA`}},
			},
			{
				Statement: `SELECT ''::text AS five, unique1, unique2, stringu1
		FROM onek
		ORDER BY unique1 LIMIT 5 OFFSET 900;`,
				Results: []sql.Row{{``, 900, 913, `QIAAAA`}, {``, 901, 931, `RIAAAA`}, {``, 902, 702, `SIAAAA`}, {``, 903, 641, `TIAAAA`}, {``, 904, 793, `UIAAAA`}},
			},
			{
				Statement: `select * from int8_tbl limit (case when random() < 0.5 then null::bigint end);`,
				Results:   []sql.Row{{123, 456}, {123, 4567890123456789}, {4567890123456789, 123}, {4567890123456789, 4567890123456789}, {4567890123456789, -4567890123456789}},
			},
			{
				Statement: `select * from int8_tbl offset (case when random() < 0.5 then null::bigint end);`,
				Results:   []sql.Row{{123, 456}, {123, 4567890123456789}, {4567890123456789, 123}, {4567890123456789, 4567890123456789}, {4567890123456789, -4567890123456789}},
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `declare c1 cursor for select * from int8_tbl limit 10;`,
			},
			{
				Statement: `fetch all in c1;`,
				Results:   []sql.Row{{123, 456}, {123, 4567890123456789}, {4567890123456789, 123}, {4567890123456789, 4567890123456789}, {4567890123456789, -4567890123456789}},
			},
			{
				Statement: `fetch 1 in c1;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `fetch backward 1 in c1;`,
				Results:   []sql.Row{{4567890123456789, -4567890123456789}},
			},
			{
				Statement: `fetch backward all in c1;`,
				Results:   []sql.Row{{4567890123456789, 4567890123456789}, {4567890123456789, 123}, {123, 4567890123456789}, {123, 456}},
			},
			{
				Statement: `fetch backward 1 in c1;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `fetch all in c1;`,
				Results:   []sql.Row{{123, 456}, {123, 4567890123456789}, {4567890123456789, 123}, {4567890123456789, 4567890123456789}, {4567890123456789, -4567890123456789}},
			},
			{
				Statement: `declare c2 cursor for select * from int8_tbl limit 3;`,
			},
			{
				Statement: `fetch all in c2;`,
				Results:   []sql.Row{{123, 456}, {123, 4567890123456789}, {4567890123456789, 123}},
			},
			{
				Statement: `fetch 1 in c2;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `fetch backward 1 in c2;`,
				Results:   []sql.Row{{4567890123456789, 123}},
			},
			{
				Statement: `fetch backward all in c2;`,
				Results:   []sql.Row{{123, 4567890123456789}, {123, 456}},
			},
			{
				Statement: `fetch backward 1 in c2;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `fetch all in c2;`,
				Results:   []sql.Row{{123, 456}, {123, 4567890123456789}, {4567890123456789, 123}},
			},
			{
				Statement: `declare c3 cursor for select * from int8_tbl offset 3;`,
			},
			{
				Statement: `fetch all in c3;`,
				Results:   []sql.Row{{4567890123456789, 4567890123456789}, {4567890123456789, -4567890123456789}},
			},
			{
				Statement: `fetch 1 in c3;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `fetch backward 1 in c3;`,
				Results:   []sql.Row{{4567890123456789, -4567890123456789}},
			},
			{
				Statement: `fetch backward all in c3;`,
				Results:   []sql.Row{{4567890123456789, 4567890123456789}},
			},
			{
				Statement: `fetch backward 1 in c3;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `fetch all in c3;`,
				Results:   []sql.Row{{4567890123456789, 4567890123456789}, {4567890123456789, -4567890123456789}},
			},
			{
				Statement: `declare c4 cursor for select * from int8_tbl offset 10;`,
			},
			{
				Statement: `fetch all in c4;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `fetch 1 in c4;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `fetch backward 1 in c4;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `fetch backward all in c4;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `fetch backward 1 in c4;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `fetch all in c4;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `declare c5 cursor for select * from int8_tbl order by q1 fetch first 2 rows with ties;`,
			},
			{
				Statement: `fetch all in c5;`,
				Results:   []sql.Row{{123, 456}, {123, 4567890123456789}},
			},
			{
				Statement: `fetch 1 in c5;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `fetch backward 1 in c5;`,
				Results:   []sql.Row{{123, 4567890123456789}},
			},
			{
				Statement: `fetch backward 1 in c5;`,
				Results:   []sql.Row{{123, 456}},
			},
			{
				Statement: `fetch all in c5;`,
				Results:   []sql.Row{{123, 4567890123456789}},
			},
			{
				Statement: `fetch backward all in c5;`,
				Results:   []sql.Row{{123, 4567890123456789}, {123, 456}},
			},
			{
				Statement: `fetch all in c5;`,
				Results:   []sql.Row{{123, 456}, {123, 4567890123456789}},
			},
			{
				Statement: `fetch backward all in c5;`,
				Results:   []sql.Row{{123, 4567890123456789}, {123, 456}},
			},
			{
				Statement: `rollback;`,
			},
			{
				Statement: `SELECT
  (SELECT n
     FROM (VALUES (1)) AS x,
          (SELECT n FROM generate_series(1,10) AS n
             ORDER BY n LIMIT 1 OFFSET s-1) AS y) AS z
  FROM generate_series(1,10) AS s;`,
				Results: []sql.Row{{1}, {2}, {3}, {4}, {5}, {6}, {7}, {8}, {9}, {10}},
			},
			{
				Statement: `create temp sequence testseq;`,
			},
			{
				Statement: `explain (verbose, costs off)
select unique1, unique2, nextval('testseq')
  from tenk1 order by unique2 limit 10;`,
				Results: []sql.Row{{`Limit`}, {`Output: unique1, unique2, (nextval('testseq'::regclass))`}, {`->  Index Scan using tenk1_unique2 on public.tenk1`}, {`Output: unique1, unique2, nextval('testseq'::regclass)`}},
			},
			{
				Statement: `select unique1, unique2, nextval('testseq')
  from tenk1 order by unique2 limit 10;`,
				Results: []sql.Row{{8800, 0, 1}, {1891, 1, 2}, {3420, 2, 3}, {9850, 3, 4}, {7164, 4, 5}, {8009, 5, 6}, {5057, 6, 7}, {6701, 7, 8}, {4321, 8, 9}, {3043, 9, 10}},
			},
			{
				Statement: `select currval('testseq');`,
				Results:   []sql.Row{{10}},
			},
			{
				Statement: `explain (verbose, costs off)
select unique1, unique2, nextval('testseq')
  from tenk1 order by tenthous limit 10;`,
				Results: []sql.Row{{`Limit`}, {`Output: unique1, unique2, (nextval('testseq'::regclass)), tenthous`}, {`->  Result`}, {`Output: unique1, unique2, nextval('testseq'::regclass), tenthous`}, {`->  Sort`}, {`Output: unique1, unique2, tenthous`}, {`Sort Key: tenk1.tenthous`}, {`->  Seq Scan on public.tenk1`}, {`Output: unique1, unique2, tenthous`}},
			},
			{
				Statement: `select unique1, unique2, nextval('testseq')
  from tenk1 order by tenthous limit 10;`,
				Results: []sql.Row{{0, 9998, 11}, {1, 2838, 12}, {2, 2716, 13}, {3, 5679, 14}, {4, 1621, 15}, {5, 5557, 16}, {6, 2855, 17}, {7, 8518, 18}, {8, 5435, 19}, {9, 4463, 20}},
			},
			{
				Statement: `select currval('testseq');`,
				Results:   []sql.Row{{20}},
			},
			{
				Statement: `explain (verbose, costs off)
select unique1, unique2, generate_series(1,10)
  from tenk1 order by unique2 limit 7;`,
				Results: []sql.Row{{`Limit`}, {`Output: unique1, unique2, (generate_series(1, 10))`}, {`->  ProjectSet`}, {`Output: unique1, unique2, generate_series(1, 10)`}, {`->  Index Scan using tenk1_unique2 on public.tenk1`}, {`Output: unique1, unique2, two, four, ten, twenty, hundred, thousand, twothousand, fivethous, tenthous, odd, even, stringu1, stringu2, string4`}},
			},
			{
				Statement: `select unique1, unique2, generate_series(1,10)
  from tenk1 order by unique2 limit 7;`,
				Results: []sql.Row{{8800, 0, 1}, {8800, 0, 2}, {8800, 0, 3}, {8800, 0, 4}, {8800, 0, 5}, {8800, 0, 6}, {8800, 0, 7}},
			},
			{
				Statement: `explain (verbose, costs off)
select unique1, unique2, generate_series(1,10)
  from tenk1 order by tenthous limit 7;`,
				Results: []sql.Row{{`Limit`}, {`Output: unique1, unique2, (generate_series(1, 10)), tenthous`}, {`->  ProjectSet`}, {`Output: unique1, unique2, generate_series(1, 10), tenthous`}, {`->  Sort`}, {`Output: unique1, unique2, tenthous`}, {`Sort Key: tenk1.tenthous`}, {`->  Seq Scan on public.tenk1`}, {`Output: unique1, unique2, tenthous`}},
			},
			{
				Statement: `select unique1, unique2, generate_series(1,10)
  from tenk1 order by tenthous limit 7;`,
				Results: []sql.Row{{0, 9998, 1}, {0, 9998, 2}, {0, 9998, 3}, {0, 9998, 4}, {0, 9998, 5}, {0, 9998, 6}, {0, 9998, 7}},
			},
			{
				Statement: `explain (verbose, costs off)
select generate_series(0,2) as s1, generate_series((random()*.1)::int,2) as s2;`,
				Results: []sql.Row{{`ProjectSet`}, {`Output: generate_series(0, 2), generate_series(((random() * '0.1'::double precision))::integer, 2)`}, {`->  Result`}},
			},
			{
				Statement: `select generate_series(0,2) as s1, generate_series((random()*.1)::int,2) as s2;`,
				Results:   []sql.Row{{0, 0}, {1, 1}, {2, 2}},
			},
			{
				Statement: `explain (verbose, costs off)
select generate_series(0,2) as s1, generate_series((random()*.1)::int,2) as s2
order by s2 desc;`,
				Results: []sql.Row{{`Sort`}, {`Output: (generate_series(0, 2)), (generate_series(((random() * '0.1'::double precision))::integer, 2))`}, {`Sort Key: (generate_series(((random() * '0.1'::double precision))::integer, 2)) DESC`}, {`->  ProjectSet`}, {`Output: generate_series(0, 2), generate_series(((random() * '0.1'::double precision))::integer, 2)`}, {`->  Result`}},
			},
			{
				Statement: `select generate_series(0,2) as s1, generate_series((random()*.1)::int,2) as s2
order by s2 desc;`,
				Results: []sql.Row{{2, 2}, {1, 1}, {0, 0}},
			},
			{
				Statement: `explain (verbose, costs off)
select sum(tenthous) as s1, sum(tenthous) + random()*0 as s2
  from tenk1 group by thousand order by thousand limit 3;`,
				Results: []sql.Row{{`Limit`}, {`Output: (sum(tenthous)), (((sum(tenthous))::double precision + (random() * '0'::double precision))), thousand`}, {`->  GroupAggregate`}, {`Output: sum(tenthous), ((sum(tenthous))::double precision + (random() * '0'::double precision)), thousand`}, {`Group Key: tenk1.thousand`}, {`->  Index Only Scan using tenk1_thous_tenthous on public.tenk1`}, {`Output: thousand, tenthous`}},
			},
			{
				Statement: `select sum(tenthous) as s1, sum(tenthous) + random()*0 as s2
  from tenk1 group by thousand order by thousand limit 3;`,
				Results: []sql.Row{{45000, 45000}, {45010, 45010}, {45020, 45020}},
			},
			{
				Statement: `SELECT  thousand
		FROM onek WHERE thousand < 5
		ORDER BY thousand FETCH FIRST 2 ROW WITH TIES;`,
				Results: []sql.Row{{0}, {0}, {0}, {0}, {0}, {0}, {0}, {0}, {0}, {0}},
			},
			{
				Statement: `SELECT  thousand
		FROM onek WHERE thousand < 5
		ORDER BY thousand FETCH FIRST ROWS WITH TIES;`,
				Results: []sql.Row{{0}, {0}, {0}, {0}, {0}, {0}, {0}, {0}, {0}, {0}},
			},
			{
				Statement: `SELECT  thousand
		FROM onek WHERE thousand < 5
		ORDER BY thousand FETCH FIRST 1 ROW WITH TIES;`,
				Results: []sql.Row{{0}, {0}, {0}, {0}, {0}, {0}, {0}, {0}, {0}, {0}},
			},
			{
				Statement: `SELECT  thousand
		FROM onek WHERE thousand < 5
		ORDER BY thousand FETCH FIRST 2 ROW ONLY;`,
				Results: []sql.Row{{0}, {0}},
			},
			{
				Statement: `SELECT  thousand
		FROM onek WHERE thousand < 5
		ORDER BY thousand FETCH FIRST 1 ROW WITH TIES FOR UPDATE SKIP LOCKED;`,
				ErrorString: `SKIP LOCKED and WITH TIES options cannot be used together`,
			},
			{
				Statement: `SELECT ''::text AS two, unique1, unique2, stringu1
		FROM onek WHERE unique1 > 50
		FETCH FIRST 2 ROW WITH TIES;`,
				ErrorString: `WITH TIES cannot be specified without ORDER BY clause`,
			},
			{
				Statement: `CREATE VIEW limit_thousand_v_1 AS SELECT thousand FROM onek WHERE thousand < 995
		ORDER BY thousand FETCH FIRST 5 ROWS WITH TIES OFFSET 10;`,
			},
			{
				Statement: `\d+ limit_thousand_v_1
                      View "public.limit_thousand_v_1"
  Column  |  Type   | Collation | Nullable | Default | Storage | Description 
----------+---------+-----------+----------+---------+---------+-------------
 thousand | integer |           |          |         | plain   | 
View definition:
 SELECT onek.thousand
   FROM onek
  WHERE onek.thousand < 995
  ORDER BY onek.thousand
 OFFSET 10
 FETCH FIRST 5 ROWS WITH TIES;`,
			},
			{
				Statement: `CREATE VIEW limit_thousand_v_2 AS SELECT thousand FROM onek WHERE thousand < 995
		ORDER BY thousand OFFSET 10 FETCH FIRST 5 ROWS ONLY;`,
			},
			{
				Statement: `\d+ limit_thousand_v_2
                      View "public.limit_thousand_v_2"
  Column  |  Type   | Collation | Nullable | Default | Storage | Description 
----------+---------+-----------+----------+---------+---------+-------------
 thousand | integer |           |          |         | plain   | 
View definition:
 SELECT onek.thousand
   FROM onek
  WHERE onek.thousand < 995
  ORDER BY onek.thousand
 OFFSET 10
 LIMIT 5;`,
			},
			{
				Statement: `CREATE VIEW limit_thousand_v_3 AS SELECT thousand FROM onek WHERE thousand < 995
		ORDER BY thousand FETCH FIRST NULL ROWS WITH TIES;		-- fails`,
				ErrorString: `row count cannot be null in FETCH FIRST ... WITH TIES clause`,
			},
			{
				Statement: `CREATE VIEW limit_thousand_v_3 AS SELECT thousand FROM onek WHERE thousand < 995
		ORDER BY thousand FETCH FIRST (NULL+1) ROWS WITH TIES;`,
			},
			{
				Statement: `\d+ limit_thousand_v_3
                      View "public.limit_thousand_v_3"
  Column  |  Type   | Collation | Nullable | Default | Storage | Description 
----------+---------+-----------+----------+---------+---------+-------------
 thousand | integer |           |          |         | plain   | 
View definition:
 SELECT onek.thousand
   FROM onek
  WHERE onek.thousand < 995
  ORDER BY onek.thousand
 FETCH FIRST (NULL::integer + 1) ROWS WITH TIES;`,
			},
			{
				Statement: `CREATE VIEW limit_thousand_v_4 AS SELECT thousand FROM onek WHERE thousand < 995
		ORDER BY thousand FETCH FIRST NULL ROWS ONLY;`,
			},
			{
				Statement: `\d+ limit_thousand_v_4
                      View "public.limit_thousand_v_4"
  Column  |  Type   | Collation | Nullable | Default | Storage | Description 
----------+---------+-----------+----------+---------+---------+-------------
 thousand | integer |           |          |         | plain   | 
View definition:
 SELECT onek.thousand
   FROM onek
  WHERE onek.thousand < 995
  ORDER BY onek.thousand
 LIMIT ALL;`,
			},
		},
	})
}
