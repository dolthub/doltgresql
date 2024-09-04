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

func TestSelect(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_select)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_select,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `SELECT * FROM onek
   WHERE onek.unique1 < 10
   ORDER BY onek.unique1;`,
				Results: []sql.Row{{0, 998, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, `AAAAAA`, `KMBAAA`, `OOOOxx`}, {1, 214, 1, 1, 1, 1, 1, 1, 1, 1, 1, 2, 3, `BAAAAA`, `GIAAAA`, `OOOOxx`}, {2, 326, 0, 2, 2, 2, 2, 2, 2, 2, 2, 4, 5, `CAAAAA`, `OMAAAA`, `OOOOxx`}, {3, 431, 1, 3, 3, 3, 3, 3, 3, 3, 3, 6, 7, `DAAAAA`, `PQAAAA`, `VVVVxx`}, {4, 833, 0, 0, 4, 4, 4, 4, 4, 4, 4, 8, 9, `EAAAAA`, `BGBAAA`, `HHHHxx`}, {5, 541, 1, 1, 5, 5, 5, 5, 5, 5, 5, 10, 11, `FAAAAA`, `VUAAAA`, `HHHHxx`}, {6, 978, 0, 2, 6, 6, 6, 6, 6, 6, 6, 12, 13, `GAAAAA`, `QLBAAA`, `OOOOxx`}, {7, 647, 1, 3, 7, 7, 7, 7, 7, 7, 7, 14, 15, `HAAAAA`, `XYAAAA`, `VVVVxx`}, {8, 653, 0, 0, 8, 8, 8, 8, 8, 8, 8, 16, 17, `IAAAAA`, `DZAAAA`, `HHHHxx`}, {9, 49, 1, 1, 9, 9, 9, 9, 9, 9, 9, 18, 19, `JAAAAA`, `XBAAAA`, `HHHHxx`}},
			},
			{
				Statement: `SELECT onek.unique1, onek.stringu1 FROM onek
   WHERE onek.unique1 < 20
   ORDER BY unique1 using >;`,
				Results: []sql.Row{{19, `TAAAAA`}, {18, `SAAAAA`}, {17, `RAAAAA`}, {16, `QAAAAA`}, {15, `PAAAAA`}, {14, `OAAAAA`}, {13, `NAAAAA`}, {12, `MAAAAA`}, {11, `LAAAAA`}, {10, `KAAAAA`}, {9, `JAAAAA`}, {8, `IAAAAA`}, {7, `HAAAAA`}, {6, `GAAAAA`}, {5, `FAAAAA`}, {4, `EAAAAA`}, {3, `DAAAAA`}, {2, `CAAAAA`}, {1, `BAAAAA`}, {0, `AAAAAA`}},
			},
			{
				Statement: `SELECT onek.unique1, onek.stringu1 FROM onek
   WHERE onek.unique1 > 980
   ORDER BY stringu1 using <;`,
				Results: []sql.Row{{988, `AMAAAA`}, {989, `BMAAAA`}, {990, `CMAAAA`}, {991, `DMAAAA`}, {992, `EMAAAA`}, {993, `FMAAAA`}, {994, `GMAAAA`}, {995, `HMAAAA`}, {996, `IMAAAA`}, {997, `JMAAAA`}, {998, `KMAAAA`}, {999, `LMAAAA`}, {981, `TLAAAA`}, {982, `ULAAAA`}, {983, `VLAAAA`}, {984, `WLAAAA`}, {985, `XLAAAA`}, {986, `YLAAAA`}, {987, `ZLAAAA`}},
			},
			{
				Statement: `SELECT onek.unique1, onek.string4 FROM onek
   WHERE onek.unique1 > 980
   ORDER BY string4 using <, unique1 using >;`,
				Results: []sql.Row{{999, `AAAAxx`}, {995, `AAAAxx`}, {983, `AAAAxx`}, {982, `AAAAxx`}, {981, `AAAAxx`}, {998, `HHHHxx`}, {997, `HHHHxx`}, {993, `HHHHxx`}, {990, `HHHHxx`}, {986, `HHHHxx`}, {996, `OOOOxx`}, {991, `OOOOxx`}, {988, `OOOOxx`}, {987, `OOOOxx`}, {985, `OOOOxx`}, {994, `VVVVxx`}, {992, `VVVVxx`}, {989, `VVVVxx`}, {984, `VVVVxx`}},
			},
			{
				Statement: `SELECT onek.unique1, onek.string4 FROM onek
   WHERE onek.unique1 > 980
   ORDER BY string4 using >, unique1 using <;`,
				Results: []sql.Row{{984, `VVVVxx`}, {989, `VVVVxx`}, {992, `VVVVxx`}, {994, `VVVVxx`}, {985, `OOOOxx`}, {987, `OOOOxx`}, {988, `OOOOxx`}, {991, `OOOOxx`}, {996, `OOOOxx`}, {986, `HHHHxx`}, {990, `HHHHxx`}, {993, `HHHHxx`}, {997, `HHHHxx`}, {998, `HHHHxx`}, {981, `AAAAxx`}, {982, `AAAAxx`}, {983, `AAAAxx`}, {995, `AAAAxx`}, {999, `AAAAxx`}},
			},
			{
				Statement: `SELECT onek.unique1, onek.string4 FROM onek
   WHERE onek.unique1 < 20
   ORDER BY unique1 using >, string4 using <;`,
				Results: []sql.Row{{19, `OOOOxx`}, {18, `VVVVxx`}, {17, `HHHHxx`}, {16, `OOOOxx`}, {15, `VVVVxx`}, {14, `AAAAxx`}, {13, `OOOOxx`}, {12, `AAAAxx`}, {11, `OOOOxx`}, {10, `AAAAxx`}, {9, `HHHHxx`}, {8, `HHHHxx`}, {7, `VVVVxx`}, {6, `OOOOxx`}, {5, `HHHHxx`}, {4, `HHHHxx`}, {3, `VVVVxx`}, {2, `OOOOxx`}, {1, `OOOOxx`}, {0, `OOOOxx`}},
			},
			{
				Statement: `SELECT onek.unique1, onek.string4 FROM onek
   WHERE onek.unique1 < 20
   ORDER BY unique1 using <, string4 using >;`,
				Results: []sql.Row{{0, `OOOOxx`}, {1, `OOOOxx`}, {2, `OOOOxx`}, {3, `VVVVxx`}, {4, `HHHHxx`}, {5, `HHHHxx`}, {6, `OOOOxx`}, {7, `VVVVxx`}, {8, `HHHHxx`}, {9, `HHHHxx`}, {10, `AAAAxx`}, {11, `OOOOxx`}, {12, `AAAAxx`}, {13, `OOOOxx`}, {14, `AAAAxx`}, {15, `VVVVxx`}, {16, `OOOOxx`}, {17, `HHHHxx`}, {18, `VVVVxx`}, {19, `OOOOxx`}},
			},
			{
				Statement: `ANALYZE onek2;`,
			},
			{
				Statement: `SET enable_seqscan TO off;`,
			},
			{
				Statement: `SET enable_bitmapscan TO off;`,
			},
			{
				Statement: `SET enable_sort TO off;`,
			},
			{
				Statement: `SELECT onek2.* FROM onek2 WHERE onek2.unique1 < 10;`,
				Results:   []sql.Row{{0, 998, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, `AAAAAA`, `KMBAAA`, `OOOOxx`}, {1, 214, 1, 1, 1, 1, 1, 1, 1, 1, 1, 2, 3, `BAAAAA`, `GIAAAA`, `OOOOxx`}, {2, 326, 0, 2, 2, 2, 2, 2, 2, 2, 2, 4, 5, `CAAAAA`, `OMAAAA`, `OOOOxx`}, {3, 431, 1, 3, 3, 3, 3, 3, 3, 3, 3, 6, 7, `DAAAAA`, `PQAAAA`, `VVVVxx`}, {4, 833, 0, 0, 4, 4, 4, 4, 4, 4, 4, 8, 9, `EAAAAA`, `BGBAAA`, `HHHHxx`}, {5, 541, 1, 1, 5, 5, 5, 5, 5, 5, 5, 10, 11, `FAAAAA`, `VUAAAA`, `HHHHxx`}, {6, 978, 0, 2, 6, 6, 6, 6, 6, 6, 6, 12, 13, `GAAAAA`, `QLBAAA`, `OOOOxx`}, {7, 647, 1, 3, 7, 7, 7, 7, 7, 7, 7, 14, 15, `HAAAAA`, `XYAAAA`, `VVVVxx`}, {8, 653, 0, 0, 8, 8, 8, 8, 8, 8, 8, 16, 17, `IAAAAA`, `DZAAAA`, `HHHHxx`}, {9, 49, 1, 1, 9, 9, 9, 9, 9, 9, 9, 18, 19, `JAAAAA`, `XBAAAA`, `HHHHxx`}},
			},
			{
				Statement: `SELECT onek2.unique1, onek2.stringu1 FROM onek2
    WHERE onek2.unique1 < 20
    ORDER BY unique1 using >;`,
				Results: []sql.Row{{19, `TAAAAA`}, {18, `SAAAAA`}, {17, `RAAAAA`}, {16, `QAAAAA`}, {15, `PAAAAA`}, {14, `OAAAAA`}, {13, `NAAAAA`}, {12, `MAAAAA`}, {11, `LAAAAA`}, {10, `KAAAAA`}, {9, `JAAAAA`}, {8, `IAAAAA`}, {7, `HAAAAA`}, {6, `GAAAAA`}, {5, `FAAAAA`}, {4, `EAAAAA`}, {3, `DAAAAA`}, {2, `CAAAAA`}, {1, `BAAAAA`}, {0, `AAAAAA`}},
			},
			{
				Statement: `SELECT onek2.unique1, onek2.stringu1 FROM onek2
   WHERE onek2.unique1 > 980;`,
				Results: []sql.Row{{981, `TLAAAA`}, {982, `ULAAAA`}, {983, `VLAAAA`}, {984, `WLAAAA`}, {985, `XLAAAA`}, {986, `YLAAAA`}, {987, `ZLAAAA`}, {988, `AMAAAA`}, {989, `BMAAAA`}, {990, `CMAAAA`}, {991, `DMAAAA`}, {992, `EMAAAA`}, {993, `FMAAAA`}, {994, `GMAAAA`}, {995, `HMAAAA`}, {996, `IMAAAA`}, {997, `JMAAAA`}, {998, `KMAAAA`}, {999, `LMAAAA`}},
			},
			{
				Statement: `RESET enable_seqscan;`,
			},
			{
				Statement: `RESET enable_bitmapscan;`,
			},
			{
				Statement: `RESET enable_sort;`,
			},
			{
				Statement: `SELECT p.name, p.age FROM person* p;`,
				Results:   []sql.Row{{`mike`, 40}, {`joe`, 20}, {`sally`, 34}, {`sandra`, 19}, {`alex`, 30}, {`sue`, 50}, {`denise`, 24}, {`sarah`, 88}, {`teresa`, 38}, {`nan`, 28}, {`leah`, 68}, {`wendy`, 78}, {`melissa`, 28}, {`joan`, 18}, {`mary`, 8}, {`jane`, 58}, {`liza`, 38}, {`jean`, 28}, {`jenifer`, 38}, {`juanita`, 58}, {`susan`, 78}, {`zena`, 98}, {`martie`, 88}, {`chris`, 78}, {`pat`, 18}, {`zola`, 58}, {`louise`, 98}, {`edna`, 18}, {`bertha`, 88}, {`sumi`, 38}, {`koko`, 88}, {`gina`, 18}, {`rean`, 48}, {`sharon`, 78}, {`paula`, 68}, {`julie`, 68}, {`belinda`, 38}, {`karen`, 48}, {`carina`, 58}, {`diane`, 18}, {`esther`, 98}, {`trudy`, 88}, {`fanny`, 8}, {`carmen`, 78}, {`lita`, 25}, {`pamela`, 48}, {`sandy`, 38}, {`trisha`, 88}, {`uma`, 78}, {`velma`, 68}, {`sharon`, 25}, {`sam`, 30}, {`bill`, 20}, {`fred`, 28}, {`larry`, 60}, {`jeff`, 23}, {`cim`, 30}, {`linda`, 19}},
			},
			{
				Statement: `SELECT p.name, p.age FROM person* p ORDER BY age using >, name;`,
				Results:   []sql.Row{{`esther`, 98}, {`louise`, 98}, {`zena`, 98}, {`bertha`, 88}, {`koko`, 88}, {`martie`, 88}, {`sarah`, 88}, {`trisha`, 88}, {`trudy`, 88}, {`carmen`, 78}, {`chris`, 78}, {`sharon`, 78}, {`susan`, 78}, {`uma`, 78}, {`wendy`, 78}, {`julie`, 68}, {`leah`, 68}, {`paula`, 68}, {`velma`, 68}, {`larry`, 60}, {`carina`, 58}, {`jane`, 58}, {`juanita`, 58}, {`zola`, 58}, {`sue`, 50}, {`karen`, 48}, {`pamela`, 48}, {`rean`, 48}, {`mike`, 40}, {`belinda`, 38}, {`jenifer`, 38}, {`liza`, 38}, {`sandy`, 38}, {`sumi`, 38}, {`teresa`, 38}, {`sally`, 34}, {`alex`, 30}, {`cim`, 30}, {`sam`, 30}, {`fred`, 28}, {`jean`, 28}, {`melissa`, 28}, {`nan`, 28}, {`lita`, 25}, {`sharon`, 25}, {`denise`, 24}, {`jeff`, 23}, {`bill`, 20}, {`joe`, 20}, {`linda`, 19}, {`sandra`, 19}, {`diane`, 18}, {`edna`, 18}, {`gina`, 18}, {`joan`, 18}, {`pat`, 18}, {`fanny`, 8}, {`mary`, 8}},
			},
			{
				Statement: `select foo from (select 1 offset 0) as foo;`,
				Results:   []sql.Row{{`(1)`}},
			},
			{
				Statement: `select foo from (select null offset 0) as foo;`,
				Results:   []sql.Row{{`()`}},
			},
			{
				Statement: `select foo from (select 'xyzzy',1,null offset 0) as foo;`,
				Results:   []sql.Row{{`(xyzzy,1,)`}},
			},
			{
				Statement: `select * from onek, (values(147, 'RFAAAA'), (931, 'VJAAAA')) as v (i, j)
    WHERE onek.unique1 = v.i and onek.stringu1 = v.j;`,
				Results: []sql.Row{{147, 0, 1, 3, 7, 7, 7, 47, 147, 147, 147, 14, 15, `RFAAAA`, `AAAAAA`, `AAAAxx`, 147, `RFAAAA`}, {931, 1, 1, 3, 1, 11, 1, 31, 131, 431, 931, 2, 3, `VJAAAA`, `BAAAAA`, `HHHHxx`, 931, `VJAAAA`}},
			},
			{
				Statement: `select * from onek,
  (values ((select i from
    (values(10000), (2), (389), (1000), (2000), ((select 10029))) as foo(i)
    order by i asc limit 1))) bar (i)
  where onek.unique1 = bar.i;`,
				Results: []sql.Row{{2, 326, 0, 2, 2, 2, 2, 2, 2, 2, 2, 4, 5, `CAAAAA`, `OMAAAA`, `OOOOxx`, 2}},
			},
			{
				Statement: `select * from onek
    where (unique1,ten) in (values (1,1), (20,0), (99,9), (17,99))
    order by unique1;`,
				Results: []sql.Row{{1, 214, 1, 1, 1, 1, 1, 1, 1, 1, 1, 2, 3, `BAAAAA`, `GIAAAA`, `OOOOxx`}, {20, 306, 0, 0, 0, 0, 0, 20, 20, 20, 20, 0, 1, `UAAAAA`, `ULAAAA`, `OOOOxx`}, {99, 101, 1, 3, 9, 19, 9, 99, 99, 99, 99, 18, 19, `VDAAAA`, `XDAAAA`, `HHHHxx`}},
			},
			{
				Statement: `VALUES (1,2), (3,4+4), (7,77.7);`,
				Results:   []sql.Row{{1, 2}, {3, 8}, {7, 77.7}},
			},
			{
				Statement: `VALUES (1,2), (3,4+4), (7,77.7)
UNION ALL
SELECT 2+2, 57
UNION ALL
TABLE int8_tbl;`,
				Results: []sql.Row{{1, 2}, {3, 8}, {7, 77.7}, {4, 57}, {123, 456}, {123, 4567890123456789}, {4567890123456789, 123}, {4567890123456789, 4567890123456789}, {4567890123456789, -4567890123456789}},
			},
			{
				Statement: `CREATE TEMP TABLE nocols();`,
			},
			{
				Statement: `INSERT INTO nocols DEFAULT VALUES;`,
			},
			{
				Statement: `SELECT * FROM nocols n, LATERAL (VALUES(n.*)) v;`,
			},
			{
				Statement: `(1 row)
CREATE TEMP TABLE foo (f1 int);`,
			},
			{
				Statement: `INSERT INTO foo VALUES (42),(3),(10),(7),(null),(null),(1);`,
			},
			{
				Statement: `SELECT * FROM foo ORDER BY f1;`,
				Results:   []sql.Row{{1}, {3}, {7}, {10}, {42}, {``}, {``}},
			},
			{
				Statement: `SELECT * FROM foo ORDER BY f1 ASC;	-- same thing`,
				Results:   []sql.Row{{1}, {3}, {7}, {10}, {42}, {``}, {``}},
			},
			{
				Statement: `SELECT * FROM foo ORDER BY f1 NULLS FIRST;`,
				Results:   []sql.Row{{``}, {``}, {1}, {3}, {7}, {10}, {42}},
			},
			{
				Statement: `SELECT * FROM foo ORDER BY f1 DESC;`,
				Results:   []sql.Row{{``}, {``}, {42}, {10}, {7}, {3}, {1}},
			},
			{
				Statement: `SELECT * FROM foo ORDER BY f1 DESC NULLS LAST;`,
				Results:   []sql.Row{{42}, {10}, {7}, {3}, {1}, {``}, {``}},
			},
			{
				Statement: `CREATE INDEX fooi ON foo (f1);`,
			},
			{
				Statement: `SET enable_sort = false;`,
			},
			{
				Statement: `SELECT * FROM foo ORDER BY f1;`,
				Results:   []sql.Row{{1}, {3}, {7}, {10}, {42}, {``}, {``}},
			},
			{
				Statement: `SELECT * FROM foo ORDER BY f1 NULLS FIRST;`,
				Results:   []sql.Row{{``}, {``}, {1}, {3}, {7}, {10}, {42}},
			},
			{
				Statement: `SELECT * FROM foo ORDER BY f1 DESC;`,
				Results:   []sql.Row{{``}, {``}, {42}, {10}, {7}, {3}, {1}},
			},
			{
				Statement: `SELECT * FROM foo ORDER BY f1 DESC NULLS LAST;`,
				Results:   []sql.Row{{42}, {10}, {7}, {3}, {1}, {``}, {``}},
			},
			{
				Statement: `DROP INDEX fooi;`,
			},
			{
				Statement: `CREATE INDEX fooi ON foo (f1 DESC);`,
			},
			{
				Statement: `SELECT * FROM foo ORDER BY f1;`,
				Results:   []sql.Row{{1}, {3}, {7}, {10}, {42}, {``}, {``}},
			},
			{
				Statement: `SELECT * FROM foo ORDER BY f1 NULLS FIRST;`,
				Results:   []sql.Row{{``}, {``}, {1}, {3}, {7}, {10}, {42}},
			},
			{
				Statement: `SELECT * FROM foo ORDER BY f1 DESC;`,
				Results:   []sql.Row{{``}, {``}, {42}, {10}, {7}, {3}, {1}},
			},
			{
				Statement: `SELECT * FROM foo ORDER BY f1 DESC NULLS LAST;`,
				Results:   []sql.Row{{42}, {10}, {7}, {3}, {1}, {``}, {``}},
			},
			{
				Statement: `DROP INDEX fooi;`,
			},
			{
				Statement: `CREATE INDEX fooi ON foo (f1 DESC NULLS LAST);`,
			},
			{
				Statement: `SELECT * FROM foo ORDER BY f1;`,
				Results:   []sql.Row{{1}, {3}, {7}, {10}, {42}, {``}, {``}},
			},
			{
				Statement: `SELECT * FROM foo ORDER BY f1 NULLS FIRST;`,
				Results:   []sql.Row{{``}, {``}, {1}, {3}, {7}, {10}, {42}},
			},
			{
				Statement: `SELECT * FROM foo ORDER BY f1 DESC;`,
				Results:   []sql.Row{{``}, {``}, {42}, {10}, {7}, {3}, {1}},
			},
			{
				Statement: `SELECT * FROM foo ORDER BY f1 DESC NULLS LAST;`,
				Results:   []sql.Row{{42}, {10}, {7}, {3}, {1}, {``}, {``}},
			},
			{
				Statement: `explain (costs off)
select * from onek2 where unique2 = 11 and stringu1 = 'ATAAAA';`,
				Results: []sql.Row{{`Index Scan using onek2_u2_prtl on onek2`}, {`Index Cond: (unique2 = 11)`}, {`Filter: (stringu1 = 'ATAAAA'::name)`}},
			},
			{
				Statement: `select * from onek2 where unique2 = 11 and stringu1 = 'ATAAAA';`,
				Results:   []sql.Row{{494, 11, 0, 2, 4, 14, 4, 94, 94, 494, 494, 8, 9, `ATAAAA`, `LAAAAA`, `VVVVxx`}},
			},
			{
				Statement: `explain (costs off, analyze on, timing off, summary off)
select * from onek2 where unique2 = 11 and stringu1 = 'ATAAAA';`,
				Results: []sql.Row{{`Index Scan using onek2_u2_prtl on onek2 (actual rows=1 loops=1)`}, {`Index Cond: (unique2 = 11)`}, {`Filter: (stringu1 = 'ATAAAA'::name)`}},
			},
			{
				Statement: `explain (costs off)
select unique2 from onek2 where unique2 = 11 and stringu1 = 'ATAAAA';`,
				Results: []sql.Row{{`Index Scan using onek2_u2_prtl on onek2`}, {`Index Cond: (unique2 = 11)`}, {`Filter: (stringu1 = 'ATAAAA'::name)`}},
			},
			{
				Statement: `select unique2 from onek2 where unique2 = 11 and stringu1 = 'ATAAAA';`,
				Results:   []sql.Row{{11}},
			},
			{
				Statement: `explain (costs off)
select * from onek2 where unique2 = 11 and stringu1 < 'B';`,
				Results: []sql.Row{{`Index Scan using onek2_u2_prtl on onek2`}, {`Index Cond: (unique2 = 11)`}},
			},
			{
				Statement: `select * from onek2 where unique2 = 11 and stringu1 < 'B';`,
				Results:   []sql.Row{{494, 11, 0, 2, 4, 14, 4, 94, 94, 494, 494, 8, 9, `ATAAAA`, `LAAAAA`, `VVVVxx`}},
			},
			{
				Statement: `explain (costs off)
select unique2 from onek2 where unique2 = 11 and stringu1 < 'B';`,
				Results: []sql.Row{{`Index Only Scan using onek2_u2_prtl on onek2`}, {`Index Cond: (unique2 = 11)`}},
			},
			{
				Statement: `select unique2 from onek2 where unique2 = 11 and stringu1 < 'B';`,
				Results:   []sql.Row{{11}},
			},
			{
				Statement: `explain (costs off)
select unique2 from onek2 where unique2 = 11 and stringu1 < 'B' for update;`,
				Results: []sql.Row{{`LockRows`}, {`->  Index Scan using onek2_u2_prtl on onek2`}, {`Index Cond: (unique2 = 11)`}, {`Filter: (stringu1 < 'B'::name)`}},
			},
			{
				Statement: `select unique2 from onek2 where unique2 = 11 and stringu1 < 'B' for update;`,
				Results:   []sql.Row{{11}},
			},
			{
				Statement: `explain (costs off)
select unique2 from onek2 where unique2 = 11 and stringu1 < 'C';`,
				Results: []sql.Row{{`Seq Scan on onek2`}, {`Filter: ((stringu1 < 'C'::name) AND (unique2 = 11))`}},
			},
			{
				Statement: `select unique2 from onek2 where unique2 = 11 and stringu1 < 'C';`,
				Results:   []sql.Row{{11}},
			},
			{
				Statement: `SET enable_indexscan TO off;`,
			},
			{
				Statement: `explain (costs off)
select unique2 from onek2 where unique2 = 11 and stringu1 < 'B';`,
				Results: []sql.Row{{`Bitmap Heap Scan on onek2`}, {`Recheck Cond: ((unique2 = 11) AND (stringu1 < 'B'::name))`}, {`->  Bitmap Index Scan on onek2_u2_prtl`}, {`Index Cond: (unique2 = 11)`}},
			},
			{
				Statement: `select unique2 from onek2 where unique2 = 11 and stringu1 < 'B';`,
				Results:   []sql.Row{{11}},
			},
			{
				Statement: `RESET enable_indexscan;`,
			},
			{
				Statement: `explain (costs off)
select unique1, unique2 from onek2
  where (unique2 = 11 or unique1 = 0) and stringu1 < 'B';`,
				Results: []sql.Row{{`Bitmap Heap Scan on onek2`}, {`Recheck Cond: (((unique2 = 11) AND (stringu1 < 'B'::name)) OR (unique1 = 0))`}, {`Filter: (stringu1 < 'B'::name)`}, {`->  BitmapOr`}, {`->  Bitmap Index Scan on onek2_u2_prtl`}, {`Index Cond: (unique2 = 11)`}, {`->  Bitmap Index Scan on onek2_u1_prtl`}, {`Index Cond: (unique1 = 0)`}},
			},
			{
				Statement: `select unique1, unique2 from onek2
  where (unique2 = 11 or unique1 = 0) and stringu1 < 'B';`,
				Results: []sql.Row{{494, 11}, {0, 998}},
			},
			{
				Statement: `explain (costs off)
select unique1, unique2 from onek2
  where (unique2 = 11 and stringu1 < 'B') or unique1 = 0;`,
				Results: []sql.Row{{`Bitmap Heap Scan on onek2`}, {`Recheck Cond: (((unique2 = 11) AND (stringu1 < 'B'::name)) OR (unique1 = 0))`}, {`->  BitmapOr`}, {`->  Bitmap Index Scan on onek2_u2_prtl`}, {`Index Cond: (unique2 = 11)`}, {`->  Bitmap Index Scan on onek2_u1_prtl`}, {`Index Cond: (unique1 = 0)`}},
			},
			{
				Statement: `select unique1, unique2 from onek2
  where (unique2 = 11 and stringu1 < 'B') or unique1 = 0;`,
				Results: []sql.Row{{494, 11}, {0, 998}},
			},
			{
				Statement: `SELECT 1 AS x ORDER BY x;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `create function sillysrf(int) returns setof int as
  'values (1),(10),(2),($1)' language sql immutable;`,
			},
			{
				Statement: `select sillysrf(42);`,
				Results:   []sql.Row{{1}, {10}, {2}, {42}},
			},
			{
				Statement: `select sillysrf(-1) order by 1;`,
				Results:   []sql.Row{{-1}, {1}, {2}, {10}},
			},
			{
				Statement: `drop function sillysrf(int);`,
			},
			{
				Statement: `select * from (values (2),(null),(1)) v(k) where k = k order by k;`,
				Results:   []sql.Row{{1}, {2}},
			},
			{
				Statement: `select * from (values (2),(null),(1)) v(k) where k = k;`,
				Results:   []sql.Row{{2}, {1}},
			},
			{
				Statement: `create table list_parted_tbl (a int,b int) partition by list (a);`,
			},
			{
				Statement: `create table list_parted_tbl1 partition of list_parted_tbl
  for values in (1) partition by list(b);`,
			},
			{
				Statement: `explain (costs off) select * from list_parted_tbl;`,
				Results:   []sql.Row{{`Result`}, {`One-Time Filter: false`}},
			},
			{
				Statement: `drop table list_parted_tbl;`,
			},
		},
	})
}
