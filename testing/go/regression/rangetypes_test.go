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

func TestRangetypes(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_rangetypes)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_rangetypes,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement:   `select ''::textrange;`,
				ErrorString: `malformed range literal: ""`,
			},
			{
				Statement:   `select '-[a,z)'::textrange;`,
				ErrorString: `malformed range literal: "-[a,z)"`,
			},
			{
				Statement:   `select '[a,z) - '::textrange;`,
				ErrorString: `malformed range literal: "[a,z) - "`,
			},
			{
				Statement:   `select '(",a)'::textrange;`,
				ErrorString: `malformed range literal: "(",a)"`,
			},
			{
				Statement:   `select '(,,a)'::textrange;`,
				ErrorString: `malformed range literal: "(,,a)"`,
			},
			{
				Statement:   `select '(),a)'::textrange;`,
				ErrorString: `malformed range literal: "(),a)"`,
			},
			{
				Statement:   `select '(a,))'::textrange;`,
				ErrorString: `malformed range literal: "(a,))"`,
			},
			{
				Statement:   `select '(],a)'::textrange;`,
				ErrorString: `malformed range literal: "(],a)"`,
			},
			{
				Statement:   `select '(a,])'::textrange;`,
				ErrorString: `malformed range literal: "(a,])"`,
			},
			{
				Statement:   `select '[z,a]'::textrange;`,
				ErrorString: `range lower bound must be less than or equal to range upper bound`,
			},
			{
				Statement: `select '  empty  '::textrange;`,
				Results:   []sql.Row{{`empty`}},
			},
			{
				Statement: `select ' ( empty, empty )  '::textrange;`,
				Results:   []sql.Row{{`(" empty"," empty ")`}},
			},
			{
				Statement: `select ' ( " a " " a ", " z " " z " )  '::textrange;`,
				Results:   []sql.Row{{`("  a   a ","  z   z  ")`}},
			},
			{
				Statement: `select '(a,)'::textrange;`,
				Results:   []sql.Row{{`(a,)`}},
			},
			{
				Statement: `select '[,z]'::textrange;`,
				Results:   []sql.Row{{`(,z]`}},
			},
			{
				Statement: `select '[a,]'::textrange;`,
				Results:   []sql.Row{{`[a,)`}},
			},
			{
				Statement: `select '(,)'::textrange;`,
				Results:   []sql.Row{{`(,)`}},
			},
			{
				Statement: `select '[ , ]'::textrange;`,
				Results:   []sql.Row{{`[" "," "]`}},
			},
			{
				Statement: `select '["",""]'::textrange;`,
				Results:   []sql.Row{{`["",""]`}},
			},
			{
				Statement: `select '[",",","]'::textrange;`,
				Results:   []sql.Row{{`[",",","]`}},
			},
			{
				Statement: `select '["\\","\\"]'::textrange;`,
				Results:   []sql.Row{{`["\\","\\"]`}},
			},
			{
				Statement: `select '(\\,a)'::textrange;`,
				Results:   []sql.Row{{`("\\",a)`}},
			},
			{
				Statement: `select '((,z)'::textrange;`,
				Results:   []sql.Row{{`("(",z)`}},
			},
			{
				Statement: `select '([,z)'::textrange;`,
				Results:   []sql.Row{{`("[",z)`}},
			},
			{
				Statement: `select '(!,()'::textrange;`,
				Results:   []sql.Row{{`(!,"(")`}},
			},
			{
				Statement: `select '(!,[)'::textrange;`,
				Results:   []sql.Row{{`(!,"[")`}},
			},
			{
				Statement: `select '[a,a]'::textrange;`,
				Results:   []sql.Row{{`[a,a]`}},
			},
			{
				Statement: `select '[a,a)'::textrange;`,
				Results:   []sql.Row{{`empty`}},
			},
			{
				Statement: `select '(a,a]'::textrange;`,
				Results:   []sql.Row{{`empty`}},
			},
			{
				Statement: `select '(a,a)'::textrange;`,
				Results:   []sql.Row{{`empty`}},
			},
			{
				Statement: `CREATE TABLE numrange_test (nr NUMRANGE);`,
			},
			{
				Statement: `create index numrange_test_btree on numrange_test(nr);`,
			},
			{
				Statement: `INSERT INTO numrange_test VALUES('[,)');`,
			},
			{
				Statement: `INSERT INTO numrange_test VALUES('[3,]');`,
			},
			{
				Statement: `INSERT INTO numrange_test VALUES('[, 5)');`,
			},
			{
				Statement: `INSERT INTO numrange_test VALUES(numrange(1.1, 2.2));`,
			},
			{
				Statement: `INSERT INTO numrange_test VALUES('empty');`,
			},
			{
				Statement: `INSERT INTO numrange_test VALUES(numrange(1.7, 1.7, '[]'));`,
			},
			{
				Statement: `SELECT nr, isempty(nr), lower(nr), upper(nr) FROM numrange_test;`,
				Results:   []sql.Row{{`(,)`, false, ``, ``}, {`[3,)`, false, 3, ``}, {`(,5)`, false, ``, 5}, {`[1.1,2.2)`, false, 1.1, 2.2}, {`empty`, true, ``, ``}, {`[1.7,1.7]`, false, 1.7, 1.7}},
			},
			{
				Statement: `SELECT nr, lower_inc(nr), lower_inf(nr), upper_inc(nr), upper_inf(nr) FROM numrange_test;`,
				Results:   []sql.Row{{`(,)`, false, true, false, true}, {`[3,)`, true, false, false, true}, {`(,5)`, false, true, false, false}, {`[1.1,2.2)`, true, false, false, false}, {`empty`, false, false, false, false}, {`[1.7,1.7]`, true, false, true, false}},
			},
			{
				Statement: `SELECT * FROM numrange_test WHERE range_contains(nr, numrange(1.9,1.91));`,
				Results:   []sql.Row{{`(,)`}, {`(,5)`}, {`[1.1,2.2)`}},
			},
			{
				Statement: `SELECT * FROM numrange_test WHERE nr @> numrange(1.0,10000.1);`,
				Results:   []sql.Row{{`(,)`}},
			},
			{
				Statement: `SELECT * FROM numrange_test WHERE range_contained_by(numrange(-1e7,-10000.1), nr);`,
				Results:   []sql.Row{{`(,)`}, {`(,5)`}},
			},
			{
				Statement: `SELECT * FROM numrange_test WHERE 1.9 <@ nr;`,
				Results:   []sql.Row{{`(,)`}, {`(,5)`}, {`[1.1,2.2)`}},
			},
			{
				Statement: `select * from numrange_test where nr = 'empty';`,
				Results:   []sql.Row{{`empty`}},
			},
			{
				Statement: `select * from numrange_test where nr = '(1.1, 2.2)';`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select * from numrange_test where nr = '[1.1, 2.2)';`,
				Results:   []sql.Row{{`[1.1,2.2)`}},
			},
			{
				Statement: `select * from numrange_test where nr < 'empty';`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select * from numrange_test where nr < numrange(-1000.0, -1000.0,'[]');`,
				Results:   []sql.Row{{`(,)`}, {`(,5)`}, {`empty`}},
			},
			{
				Statement: `select * from numrange_test where nr < numrange(0.0, 1.0,'[]');`,
				Results:   []sql.Row{{`(,)`}, {`(,5)`}, {`empty`}},
			},
			{
				Statement: `select * from numrange_test where nr < numrange(1000.0, 1001.0,'[]');`,
				Results:   []sql.Row{{`(,)`}, {`[3,)`}, {`(,5)`}, {`[1.1,2.2)`}, {`empty`}, {`[1.7,1.7]`}},
			},
			{
				Statement: `select * from numrange_test where nr <= 'empty';`,
				Results:   []sql.Row{{`empty`}},
			},
			{
				Statement: `select * from numrange_test where nr >= 'empty';`,
				Results:   []sql.Row{{`(,)`}, {`[3,)`}, {`(,5)`}, {`[1.1,2.2)`}, {`empty`}, {`[1.7,1.7]`}},
			},
			{
				Statement: `select * from numrange_test where nr > 'empty';`,
				Results:   []sql.Row{{`(,)`}, {`[3,)`}, {`(,5)`}, {`[1.1,2.2)`}, {`[1.7,1.7]`}},
			},
			{
				Statement: `select * from numrange_test where nr > numrange(-1001.0, -1000.0,'[]');`,
				Results:   []sql.Row{{`[3,)`}, {`[1.1,2.2)`}, {`[1.7,1.7]`}},
			},
			{
				Statement: `select * from numrange_test where nr > numrange(0.0, 1.0,'[]');`,
				Results:   []sql.Row{{`[3,)`}, {`[1.1,2.2)`}, {`[1.7,1.7]`}},
			},
			{
				Statement: `select * from numrange_test where nr > numrange(1000.0, 1000.0,'[]');`,
				Results:   []sql.Row{},
			},
			{
				Statement:   `select numrange(2.0, 1.0);`,
				ErrorString: `range lower bound must be less than or equal to range upper bound`,
			},
			{
				Statement: `select numrange(2.0, 3.0) -|- numrange(3.0, 4.0);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select range_adjacent(numrange(2.0, 3.0), numrange(3.1, 4.0));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select range_adjacent(numrange(2.0, 3.0), numrange(3.1, null));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select numrange(2.0, 3.0, '[]') -|- numrange(3.0, 4.0, '()');`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select numrange(1.0, 2.0) -|- numrange(2.0, 3.0,'[]');`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select range_adjacent(numrange(2.0, 3.0, '(]'), numrange(1.0, 2.0, '(]'));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select numrange(1.1, 3.3) <@ numrange(0.1,10.1);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select numrange(0.1, 10.1) <@ numrange(1.1,3.3);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select numrange(1.1, 2.2) - numrange(2.0, 3.0);`,
				Results:   []sql.Row{{`[1.1,2.0)`}},
			},
			{
				Statement: `select numrange(1.1, 2.2) - numrange(2.2, 3.0);`,
				Results:   []sql.Row{{`[1.1,2.2)`}},
			},
			{
				Statement: `select numrange(1.1, 2.2,'[]') - numrange(2.0, 3.0);`,
				Results:   []sql.Row{{`[1.1,2.0)`}},
			},
			{
				Statement: `select range_minus(numrange(10.1,12.2,'[]'), numrange(110.0,120.2,'(]'));`,
				Results:   []sql.Row{{`[10.1,12.2]`}},
			},
			{
				Statement: `select range_minus(numrange(10.1,12.2,'[]'), numrange(0.0,120.2,'(]'));`,
				Results:   []sql.Row{{`empty`}},
			},
			{
				Statement: `select numrange(4.5, 5.5, '[]') && numrange(5.5, 6.5);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select numrange(1.0, 2.0) << numrange(3.0, 4.0);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select numrange(1.0, 3.0,'[]') << numrange(3.0, 4.0,'[]');`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select numrange(1.0, 3.0,'()') << numrange(3.0, 4.0,'()');`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select numrange(1.0, 2.0) >> numrange(3.0, 4.0);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select numrange(3.0, 70.0) &< numrange(6.6, 100.0);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select numrange(1.1, 2.2) < numrange(1.0, 200.2);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select numrange(1.1, 2.2) < numrange(1.1, 1.2);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select numrange(1.0, 2.0) + numrange(2.0, 3.0);`,
				Results:   []sql.Row{{`[1.0,3.0)`}},
			},
			{
				Statement: `select numrange(1.0, 2.0) + numrange(1.5, 3.0);`,
				Results:   []sql.Row{{`[1.0,3.0)`}},
			},
			{
				Statement:   `select numrange(1.0, 2.0) + numrange(2.5, 3.0); -- should fail`,
				ErrorString: `result of range union would not be contiguous`,
			},
			{
				Statement: `select range_merge(numrange(1.0, 2.0), numrange(2.0, 3.0));`,
				Results:   []sql.Row{{`[1.0,3.0)`}},
			},
			{
				Statement: `select range_merge(numrange(1.0, 2.0), numrange(1.5, 3.0));`,
				Results:   []sql.Row{{`[1.0,3.0)`}},
			},
			{
				Statement: `select range_merge(numrange(1.0, 2.0), numrange(2.5, 3.0)); -- shouldn't fail`,
				Results:   []sql.Row{{`[1.0,3.0)`}},
			},
			{
				Statement: `select numrange(1.0, 2.0) * numrange(2.0, 3.0);`,
				Results:   []sql.Row{{`empty`}},
			},
			{
				Statement: `select numrange(1.0, 2.0) * numrange(1.5, 3.0);`,
				Results:   []sql.Row{{`[1.5,2.0)`}},
			},
			{
				Statement: `select numrange(1.0, 2.0) * numrange(2.5, 3.0);`,
				Results:   []sql.Row{{`empty`}},
			},
			{
				Statement: `select range_intersect_agg(nr) from numrange_test;`,
				Results:   []sql.Row{{`empty`}},
			},
			{
				Statement: `select range_intersect_agg(nr) from numrange_test where false;`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select range_intersect_agg(nr) from numrange_test where nr @> 4.0;`,
				Results:   []sql.Row{{`[3,5)`}},
			},
			{
				Statement: `analyze numrange_test;`,
			},
			{
				Statement: `create table numrange_test2(nr numrange);`,
			},
			{
				Statement: `create index numrange_test2_hash_idx on numrange_test2 using hash (nr);`,
			},
			{
				Statement: `INSERT INTO numrange_test2 VALUES('[, 5)');`,
			},
			{
				Statement: `INSERT INTO numrange_test2 VALUES(numrange(1.1, 2.2));`,
			},
			{
				Statement: `INSERT INTO numrange_test2 VALUES(numrange(1.1, 2.2));`,
			},
			{
				Statement: `INSERT INTO numrange_test2 VALUES(numrange(1.1, 2.2,'()'));`,
			},
			{
				Statement: `INSERT INTO numrange_test2 VALUES('empty');`,
			},
			{
				Statement: `select * from numrange_test2 where nr = 'empty'::numrange;`,
				Results:   []sql.Row{{`empty`}},
			},
			{
				Statement: `select * from numrange_test2 where nr = numrange(1.1, 2.2);`,
				Results:   []sql.Row{{`[1.1,2.2)`}, {`[1.1,2.2)`}},
			},
			{
				Statement: `select * from numrange_test2 where nr = numrange(1.1, 2.3);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `set enable_nestloop=t;`,
			},
			{
				Statement: `set enable_hashjoin=f;`,
			},
			{
				Statement: `set enable_mergejoin=f;`,
			},
			{
				Statement: `select * from numrange_test natural join numrange_test2 order by nr;`,
				Results:   []sql.Row{{`empty`}, {`(,5)`}, {`[1.1,2.2)`}, {`[1.1,2.2)`}},
			},
			{
				Statement: `set enable_nestloop=f;`,
			},
			{
				Statement: `set enable_hashjoin=t;`,
			},
			{
				Statement: `set enable_mergejoin=f;`,
			},
			{
				Statement: `select * from numrange_test natural join numrange_test2 order by nr;`,
				Results:   []sql.Row{{`empty`}, {`(,5)`}, {`[1.1,2.2)`}, {`[1.1,2.2)`}},
			},
			{
				Statement: `set enable_nestloop=f;`,
			},
			{
				Statement: `set enable_hashjoin=f;`,
			},
			{
				Statement: `set enable_mergejoin=t;`,
			},
			{
				Statement: `select * from numrange_test natural join numrange_test2 order by nr;`,
				Results:   []sql.Row{{`empty`}, {`(,5)`}, {`[1.1,2.2)`}, {`[1.1,2.2)`}},
			},
			{
				Statement: `set enable_nestloop to default;`,
			},
			{
				Statement: `set enable_hashjoin to default;`,
			},
			{
				Statement: `set enable_mergejoin to default;`,
			},
			{
				Statement: `DROP TABLE numrange_test2;`,
			},
			{
				Statement: `CREATE TABLE textrange_test (tr textrange);`,
			},
			{
				Statement: `create index textrange_test_btree on textrange_test(tr);`,
			},
			{
				Statement: `INSERT INTO textrange_test VALUES('[,)');`,
			},
			{
				Statement: `INSERT INTO textrange_test VALUES('["a",]');`,
			},
			{
				Statement: `INSERT INTO textrange_test VALUES('[,"q")');`,
			},
			{
				Statement: `INSERT INTO textrange_test VALUES(textrange('b', 'g'));`,
			},
			{
				Statement: `INSERT INTO textrange_test VALUES('empty');`,
			},
			{
				Statement: `INSERT INTO textrange_test VALUES(textrange('d', 'd', '[]'));`,
			},
			{
				Statement: `SELECT tr, isempty(tr), lower(tr), upper(tr) FROM textrange_test;`,
				Results:   []sql.Row{{`(,)`, false, ``, ``}, {`[a,)`, false, `a`, ``}, {`(,q)`, false, ``, `q`}, {`[b,g)`, false, `b`, `g`}, {`empty`, true, ``, ``}, {`[d,d]`, false, `d`, `d`}},
			},
			{
				Statement: `SELECT tr, lower_inc(tr), lower_inf(tr), upper_inc(tr), upper_inf(tr) FROM textrange_test;`,
				Results:   []sql.Row{{`(,)`, false, true, false, true}, {`[a,)`, true, false, false, true}, {`(,q)`, false, true, false, false}, {`[b,g)`, true, false, false, false}, {`empty`, false, false, false, false}, {`[d,d]`, true, false, true, false}},
			},
			{
				Statement: `SELECT * FROM textrange_test WHERE range_contains(tr, textrange('f', 'fx'));`,
				Results:   []sql.Row{{`(,)`}, {`[a,)`}, {`(,q)`}, {`[b,g)`}},
			},
			{
				Statement: `SELECT * FROM textrange_test WHERE tr @> textrange('a', 'z');`,
				Results:   []sql.Row{{`(,)`}, {`[a,)`}},
			},
			{
				Statement: `SELECT * FROM textrange_test WHERE range_contained_by(textrange('0','9'), tr);`,
				Results:   []sql.Row{{`(,)`}, {`(,q)`}},
			},
			{
				Statement: `SELECT * FROM textrange_test WHERE 'e'::text <@ tr;`,
				Results:   []sql.Row{{`(,)`}, {`[a,)`}, {`(,q)`}, {`[b,g)`}},
			},
			{
				Statement: `select * from textrange_test where tr = 'empty';`,
				Results:   []sql.Row{{`empty`}},
			},
			{
				Statement: `select * from textrange_test where tr = '("b","g")';`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select * from textrange_test where tr = '["b","g")';`,
				Results:   []sql.Row{{`[b,g)`}},
			},
			{
				Statement: `select * from textrange_test where tr < 'empty';`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select int4range(1, 10, '[]');`,
				Results:   []sql.Row{{`[1,11)`}},
			},
			{
				Statement: `select int4range(1, 10, '[)');`,
				Results:   []sql.Row{{`[1,10)`}},
			},
			{
				Statement: `select int4range(1, 10, '(]');`,
				Results:   []sql.Row{{`[2,11)`}},
			},
			{
				Statement: `select int4range(1, 10, '()');`,
				Results:   []sql.Row{{`[2,10)`}},
			},
			{
				Statement: `select int4range(1, 2, '()');`,
				Results:   []sql.Row{{`empty`}},
			},
			{
				Statement: `select daterange('2000-01-10'::date, '2000-01-20'::date, '[]');`,
				Results:   []sql.Row{{`[01-10-2000,01-21-2000)`}},
			},
			{
				Statement: `select daterange('2000-01-10'::date, '2000-01-20'::date, '[)');`,
				Results:   []sql.Row{{`[01-10-2000,01-20-2000)`}},
			},
			{
				Statement: `select daterange('2000-01-10'::date, '2000-01-20'::date, '(]');`,
				Results:   []sql.Row{{`[01-11-2000,01-21-2000)`}},
			},
			{
				Statement: `select daterange('2000-01-10'::date, '2000-01-20'::date, '()');`,
				Results:   []sql.Row{{`[01-11-2000,01-20-2000)`}},
			},
			{
				Statement: `select daterange('2000-01-10'::date, '2000-01-11'::date, '()');`,
				Results:   []sql.Row{{`empty`}},
			},
			{
				Statement: `select daterange('2000-01-10'::date, '2000-01-11'::date, '(]');`,
				Results:   []sql.Row{{`[01-11-2000,01-12-2000)`}},
			},
			{
				Statement: `select daterange('-infinity'::date, '2000-01-01'::date, '()');`,
				Results:   []sql.Row{{`(-infinity,01-01-2000)`}},
			},
			{
				Statement: `select daterange('-infinity'::date, '2000-01-01'::date, '[)');`,
				Results:   []sql.Row{{`[-infinity,01-01-2000)`}},
			},
			{
				Statement: `select daterange('2000-01-01'::date, 'infinity'::date, '[)');`,
				Results:   []sql.Row{{`[01-01-2000,infinity)`}},
			},
			{
				Statement: `select daterange('2000-01-01'::date, 'infinity'::date, '[]');`,
				Results:   []sql.Row{{`[01-01-2000,infinity]`}},
			},
			{
				Statement: `create table test_range_gist(ir int4range);`,
			},
			{
				Statement: `create index test_range_gist_idx on test_range_gist using gist (ir);`,
			},
			{
				Statement: `insert into test_range_gist select int4range(g, g+10) from generate_series(1,2000) g;`,
			},
			{
				Statement: `insert into test_range_gist select 'empty'::int4range from generate_series(1,500) g;`,
			},
			{
				Statement: `insert into test_range_gist select int4range(g, g+10000) from generate_series(1,1000) g;`,
			},
			{
				Statement: `insert into test_range_gist select 'empty'::int4range from generate_series(1,500) g;`,
			},
			{
				Statement: `insert into test_range_gist select int4range(NULL,g*10,'(]') from generate_series(1,100) g;`,
			},
			{
				Statement: `insert into test_range_gist select int4range(g*10,NULL,'(]') from generate_series(1,100) g;`,
			},
			{
				Statement: `insert into test_range_gist select int4range(g, g+10) from generate_series(1,2000) g;`,
			},
			{
				Statement: `analyze test_range_gist;`,
			},
			{
				Statement: `SET enable_seqscan    = t;`,
			},
			{
				Statement: `SET enable_indexscan  = f;`,
			},
			{
				Statement: `SET enable_bitmapscan = f;`,
			},
			{
				Statement: `select count(*) from test_range_gist where ir @> 'empty'::int4range;`,
				Results:   []sql.Row{{6200}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir = int4range(10,20);`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir @> 10;`,
				Results:   []sql.Row{{130}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir @> int4range(10,20);`,
				Results:   []sql.Row{{111}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir && int4range(10,20);`,
				Results:   []sql.Row{{158}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir <@ int4range(10,50);`,
				Results:   []sql.Row{{1062}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir << int4range(100,500);`,
				Results:   []sql.Row{{189}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir >> int4range(100,500);`,
				Results:   []sql.Row{{3554}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir &< int4range(100,500);`,
				Results:   []sql.Row{{1029}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir &> int4range(100,500);`,
				Results:   []sql.Row{{4794}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir -|- int4range(100,500);`,
				Results:   []sql.Row{{5}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir @> '{}'::int4multirange;`,
				Results:   []sql.Row{{6200}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir @> int4multirange(int4range(10,20), int4range(30,40));`,
				Results:   []sql.Row{{107}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir && '{(10,20),(30,40),(50,60)}'::int4multirange;`,
				Results:   []sql.Row{{271}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir <@ '{(10,30),(40,60),(70,90)}'::int4multirange;`,
				Results:   []sql.Row{{1060}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir << int4multirange(int4range(100,200), int4range(400,500));`,
				Results:   []sql.Row{{189}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir >> int4multirange(int4range(100,200), int4range(400,500));`,
				Results:   []sql.Row{{3554}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir &< int4multirange(int4range(100,200), int4range(400,500));`,
				Results:   []sql.Row{{1029}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir &> int4multirange(int4range(100,200), int4range(400,500));`,
				Results:   []sql.Row{{4794}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir -|- int4multirange(int4range(100,200), int4range(400,500));`,
				Results:   []sql.Row{{5}},
			},
			{
				Statement: `SET enable_seqscan    = f;`,
			},
			{
				Statement: `SET enable_indexscan  = t;`,
			},
			{
				Statement: `SET enable_bitmapscan = f;`,
			},
			{
				Statement: `select count(*) from test_range_gist where ir @> 'empty'::int4range;`,
				Results:   []sql.Row{{6200}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir = int4range(10,20);`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir @> 10;`,
				Results:   []sql.Row{{130}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir @> int4range(10,20);`,
				Results:   []sql.Row{{111}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir && int4range(10,20);`,
				Results:   []sql.Row{{158}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir <@ int4range(10,50);`,
				Results:   []sql.Row{{1062}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir << int4range(100,500);`,
				Results:   []sql.Row{{189}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir >> int4range(100,500);`,
				Results:   []sql.Row{{3554}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir &< int4range(100,500);`,
				Results:   []sql.Row{{1029}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir &> int4range(100,500);`,
				Results:   []sql.Row{{4794}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir -|- int4range(100,500);`,
				Results:   []sql.Row{{5}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir @> '{}'::int4multirange;`,
				Results:   []sql.Row{{6200}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir @> int4multirange(int4range(10,20), int4range(30,40));`,
				Results:   []sql.Row{{107}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir && '{(10,20),(30,40),(50,60)}'::int4multirange;`,
				Results:   []sql.Row{{271}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir <@ '{(10,30),(40,60),(70,90)}'::int4multirange;`,
				Results:   []sql.Row{{1060}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir << int4multirange(int4range(100,200), int4range(400,500));`,
				Results:   []sql.Row{{189}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir >> int4multirange(int4range(100,200), int4range(400,500));`,
				Results:   []sql.Row{{3554}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir &< int4multirange(int4range(100,200), int4range(400,500));`,
				Results:   []sql.Row{{1029}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir &> int4multirange(int4range(100,200), int4range(400,500));`,
				Results:   []sql.Row{{4794}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir -|- int4multirange(int4range(100,200), int4range(400,500));`,
				Results:   []sql.Row{{5}},
			},
			{
				Statement: `drop index test_range_gist_idx;`,
			},
			{
				Statement: `create index test_range_gist_idx on test_range_gist using gist (ir);`,
			},
			{
				Statement: `select count(*) from test_range_gist where ir @> 'empty'::int4range;`,
				Results:   []sql.Row{{6200}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir = int4range(10,20);`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir @> 10;`,
				Results:   []sql.Row{{130}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir @> int4range(10,20);`,
				Results:   []sql.Row{{111}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir && int4range(10,20);`,
				Results:   []sql.Row{{158}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir <@ int4range(10,50);`,
				Results:   []sql.Row{{1062}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir << int4range(100,500);`,
				Results:   []sql.Row{{189}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir >> int4range(100,500);`,
				Results:   []sql.Row{{3554}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir &< int4range(100,500);`,
				Results:   []sql.Row{{1029}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir &> int4range(100,500);`,
				Results:   []sql.Row{{4794}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir -|- int4range(100,500);`,
				Results:   []sql.Row{{5}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir @> '{}'::int4multirange;`,
				Results:   []sql.Row{{6200}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir @> int4multirange(int4range(10,20), int4range(30,40));`,
				Results:   []sql.Row{{107}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir && '{(10,20),(30,40),(50,60)}'::int4multirange;`,
				Results:   []sql.Row{{271}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir <@ '{(10,30),(40,60),(70,90)}'::int4multirange;`,
				Results:   []sql.Row{{1060}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir << int4multirange(int4range(100,200), int4range(400,500));`,
				Results:   []sql.Row{{189}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir >> int4multirange(int4range(100,200), int4range(400,500));`,
				Results:   []sql.Row{{3554}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir &< int4multirange(int4range(100,200), int4range(400,500));`,
				Results:   []sql.Row{{1029}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir &> int4multirange(int4range(100,200), int4range(400,500));`,
				Results:   []sql.Row{{4794}},
			},
			{
				Statement: `select count(*) from test_range_gist where ir -|- int4multirange(int4range(100,200), int4range(400,500));`,
				Results:   []sql.Row{{5}},
			},
			{
				Statement: `create table test_range_spgist(ir int4range);`,
			},
			{
				Statement: `create index test_range_spgist_idx on test_range_spgist using spgist (ir);`,
			},
			{
				Statement: `insert into test_range_spgist select int4range(g, g+10) from generate_series(1,2000) g;`,
			},
			{
				Statement: `insert into test_range_spgist select 'empty'::int4range from generate_series(1,500) g;`,
			},
			{
				Statement: `insert into test_range_spgist select int4range(g, g+10000) from generate_series(1,1000) g;`,
			},
			{
				Statement: `insert into test_range_spgist select 'empty'::int4range from generate_series(1,500) g;`,
			},
			{
				Statement: `insert into test_range_spgist select int4range(NULL,g*10,'(]') from generate_series(1,100) g;`,
			},
			{
				Statement: `insert into test_range_spgist select int4range(g*10,NULL,'(]') from generate_series(1,100) g;`,
			},
			{
				Statement: `insert into test_range_spgist select int4range(g, g+10) from generate_series(1,2000) g;`,
			},
			{
				Statement: `SET enable_seqscan    = t;`,
			},
			{
				Statement: `SET enable_indexscan  = f;`,
			},
			{
				Statement: `SET enable_bitmapscan = f;`,
			},
			{
				Statement: `select count(*) from test_range_spgist where ir @> 'empty'::int4range;`,
				Results:   []sql.Row{{6200}},
			},
			{
				Statement: `select count(*) from test_range_spgist where ir = int4range(10,20);`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `select count(*) from test_range_spgist where ir @> 10;`,
				Results:   []sql.Row{{130}},
			},
			{
				Statement: `select count(*) from test_range_spgist where ir @> int4range(10,20);`,
				Results:   []sql.Row{{111}},
			},
			{
				Statement: `select count(*) from test_range_spgist where ir && int4range(10,20);`,
				Results:   []sql.Row{{158}},
			},
			{
				Statement: `select count(*) from test_range_spgist where ir <@ int4range(10,50);`,
				Results:   []sql.Row{{1062}},
			},
			{
				Statement: `select count(*) from test_range_spgist where ir << int4range(100,500);`,
				Results:   []sql.Row{{189}},
			},
			{
				Statement: `select count(*) from test_range_spgist where ir >> int4range(100,500);`,
				Results:   []sql.Row{{3554}},
			},
			{
				Statement: `select count(*) from test_range_spgist where ir &< int4range(100,500);`,
				Results:   []sql.Row{{1029}},
			},
			{
				Statement: `select count(*) from test_range_spgist where ir &> int4range(100,500);`,
				Results:   []sql.Row{{4794}},
			},
			{
				Statement: `select count(*) from test_range_spgist where ir -|- int4range(100,500);`,
				Results:   []sql.Row{{5}},
			},
			{
				Statement: `SET enable_seqscan    = f;`,
			},
			{
				Statement: `SET enable_indexscan  = t;`,
			},
			{
				Statement: `SET enable_bitmapscan = f;`,
			},
			{
				Statement: `select count(*) from test_range_spgist where ir @> 'empty'::int4range;`,
				Results:   []sql.Row{{6200}},
			},
			{
				Statement: `select count(*) from test_range_spgist where ir = int4range(10,20);`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `select count(*) from test_range_spgist where ir @> 10;`,
				Results:   []sql.Row{{130}},
			},
			{
				Statement: `select count(*) from test_range_spgist where ir @> int4range(10,20);`,
				Results:   []sql.Row{{111}},
			},
			{
				Statement: `select count(*) from test_range_spgist where ir && int4range(10,20);`,
				Results:   []sql.Row{{158}},
			},
			{
				Statement: `select count(*) from test_range_spgist where ir <@ int4range(10,50);`,
				Results:   []sql.Row{{1062}},
			},
			{
				Statement: `select count(*) from test_range_spgist where ir << int4range(100,500);`,
				Results:   []sql.Row{{189}},
			},
			{
				Statement: `select count(*) from test_range_spgist where ir >> int4range(100,500);`,
				Results:   []sql.Row{{3554}},
			},
			{
				Statement: `select count(*) from test_range_spgist where ir &< int4range(100,500);`,
				Results:   []sql.Row{{1029}},
			},
			{
				Statement: `select count(*) from test_range_spgist where ir &> int4range(100,500);`,
				Results:   []sql.Row{{4794}},
			},
			{
				Statement: `select count(*) from test_range_spgist where ir -|- int4range(100,500);`,
				Results:   []sql.Row{{5}},
			},
			{
				Statement: `drop index test_range_spgist_idx;`,
			},
			{
				Statement: `create index test_range_spgist_idx on test_range_spgist using spgist (ir);`,
			},
			{
				Statement: `select count(*) from test_range_spgist where ir @> 'empty'::int4range;`,
				Results:   []sql.Row{{6200}},
			},
			{
				Statement: `select count(*) from test_range_spgist where ir = int4range(10,20);`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `select count(*) from test_range_spgist where ir @> 10;`,
				Results:   []sql.Row{{130}},
			},
			{
				Statement: `select count(*) from test_range_spgist where ir @> int4range(10,20);`,
				Results:   []sql.Row{{111}},
			},
			{
				Statement: `select count(*) from test_range_spgist where ir && int4range(10,20);`,
				Results:   []sql.Row{{158}},
			},
			{
				Statement: `select count(*) from test_range_spgist where ir <@ int4range(10,50);`,
				Results:   []sql.Row{{1062}},
			},
			{
				Statement: `select count(*) from test_range_spgist where ir << int4range(100,500);`,
				Results:   []sql.Row{{189}},
			},
			{
				Statement: `select count(*) from test_range_spgist where ir >> int4range(100,500);`,
				Results:   []sql.Row{{3554}},
			},
			{
				Statement: `select count(*) from test_range_spgist where ir &< int4range(100,500);`,
				Results:   []sql.Row{{1029}},
			},
			{
				Statement: `select count(*) from test_range_spgist where ir &> int4range(100,500);`,
				Results:   []sql.Row{{4794}},
			},
			{
				Statement: `select count(*) from test_range_spgist where ir -|- int4range(100,500);`,
				Results:   []sql.Row{{5}},
			},
			{
				Statement: `explain (costs off)
select ir from test_range_spgist where ir -|- int4range(10,20) order by ir;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: ir`}, {`->  Index Only Scan using test_range_spgist_idx on test_range_spgist`}, {`Index Cond: (ir -|- '[10,20)'::int4range)`}},
			},
			{
				Statement: `select ir from test_range_spgist where ir -|- int4range(10,20) order by ir;`,
				Results:   []sql.Row{{`[20,30)`}, {`[20,30)`}, {`[20,10020)`}},
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
			{
				Statement: `create table test_range_elem(i int4);`,
			},
			{
				Statement: `create index test_range_elem_idx on test_range_elem (i);`,
			},
			{
				Statement: `insert into test_range_elem select i from generate_series(1,100) i;`,
			},
			{
				Statement: `SET enable_seqscan    = f;`,
			},
			{
				Statement: `select count(*) from test_range_elem where i <@ int4range(10,50);`,
				Results:   []sql.Row{{40}},
			},
			{
				Statement: `create index on test_range_elem using spgist(int4range(i,i+10));`,
			},
			{
				Statement: `explain (costs off)
select count(*) from test_range_elem where int4range(i,i+10) <@ int4range(10,30);`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Scan using test_range_elem_int4range_idx on test_range_elem`}, {`Index Cond: (int4range(i, (i + 10)) <@ '[10,30)'::int4range)`}},
			},
			{
				Statement: `select count(*) from test_range_elem where int4range(i,i+10) <@ int4range(10,30);`,
				Results:   []sql.Row{{11}},
			},
			{
				Statement: `RESET enable_seqscan;`,
			},
			{
				Statement: `drop table test_range_elem;`,
			},
			{
				Statement: `create table test_range_excl(
  room int4range,
  speaker int4range,
  during tsrange,
  exclude using gist (room with =, during with &&),
  exclude using gist (speaker with =, during with &&)
);`,
			},
			{
				Statement: `insert into test_range_excl
  values(int4range(123, 123, '[]'), int4range(1, 1, '[]'), '[2010-01-02 10:00, 2010-01-02 11:00)');`,
			},
			{
				Statement: `insert into test_range_excl
  values(int4range(123, 123, '[]'), int4range(2, 2, '[]'), '[2010-01-02 11:00, 2010-01-02 12:00)');`,
			},
			{
				Statement: `insert into test_range_excl
  values(int4range(123, 123, '[]'), int4range(3, 3, '[]'), '[2010-01-02 10:10, 2010-01-02 11:00)');`,
				ErrorString: `conflicting key value violates exclusion constraint "test_range_excl_room_during_excl"`,
			},
			{
				Statement: `insert into test_range_excl
  values(int4range(124, 124, '[]'), int4range(3, 3, '[]'), '[2010-01-02 10:10, 2010-01-02 11:10)');`,
			},
			{
				Statement: `insert into test_range_excl
  values(int4range(125, 125, '[]'), int4range(1, 1, '[]'), '[2010-01-02 10:10, 2010-01-02 11:00)');`,
				ErrorString: `conflicting key value violates exclusion constraint "test_range_excl_speaker_during_excl"`,
			},
			{
				Statement: `select int8range(10000000000::int8, 20000000000::int8,'(]');`,
				Results:   []sql.Row{{`[10000000001,20000000001)`}},
			},
			{
				Statement: `set timezone to '-08';`,
			},
			{
				Statement: `select '[2010-01-01 01:00:00 -05, 2010-01-01 02:00:00 -08)'::tstzrange;`,
				Results:   []sql.Row{{`["Thu Dec 31 22:00:00 2009 -08","Fri Jan 01 02:00:00 2010 -08")`}},
			},
			{
				Statement:   `select '[2010-01-01 01:00:00 -08, 2010-01-01 02:00:00 -05)'::tstzrange;`,
				ErrorString: `range lower bound must be less than or equal to range upper bound`,
			},
			{
				Statement: `set timezone to default;`,
			},
			{
				Statement:   `create type bogus_float8range as range (subtype=float8, subtype_diff=float4mi);`,
				ErrorString: `function float4mi(double precision, double precision) does not exist`,
			},
			{
				Statement: `select '[123.001, 5.e9)'::float8range @> 888.882::float8;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `create table float8range_test(f8r float8range, i int);`,
			},
			{
				Statement: `insert into float8range_test values(float8range(-100.00007, '1.111113e9'), 42);`,
			},
			{
				Statement: `select * from float8range_test;`,
				Results:   []sql.Row{{`[-100.00007,1111113000)`, 42}},
			},
			{
				Statement: `drop table float8range_test;`,
			},
			{
				Statement: `create domain mydomain as int4;`,
			},
			{
				Statement: `create type mydomainrange as range(subtype=mydomain);`,
			},
			{
				Statement: `select '[4,50)'::mydomainrange @> 7::mydomain;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement:   `drop domain mydomain;  -- fail`,
				ErrorString: `cannot drop type mydomain because other objects depend on it`,
			},
			{
				Statement: `drop domain mydomain cascade;`,
			},
			{
				Statement: `create domain restrictedrange as int4range check (upper(value) < 10);`,
			},
			{
				Statement: `select '[4,5)'::restrictedrange @> 7;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement:   `select '[4,50)'::restrictedrange @> 7; -- should fail`,
				ErrorString: `value for domain restrictedrange violates check constraint "restrictedrange_check"`,
			},
			{
				Statement: `drop domain restrictedrange;`,
			},
			{
				Statement: `create type textrange1 as range(subtype=text, collation="C");`,
			},
			{
				Statement: `create type textrange2 as range(subtype=text, collation="C");`,
			},
			{
				Statement:   `select textrange1('a','Z') @> 'b'::text;`,
				ErrorString: `range lower bound must be less than or equal to range upper bound`,
			},
			{
				Statement: `select textrange2('a','z') @> 'b'::text;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `drop type textrange1;`,
			},
			{
				Statement: `drop type textrange2;`,
			},
			{
				Statement: `create function anyarray_anyrange_func(a anyarray, r anyrange)
  returns anyelement as 'select $1[1] + lower($2);' language sql;`,
			},
			{
				Statement: `select anyarray_anyrange_func(ARRAY[1,2], int4range(10,20));`,
				Results:   []sql.Row{{11}},
			},
			{
				Statement:   `select anyarray_anyrange_func(ARRAY[1,2], numrange(10,20));`,
				ErrorString: `function anyarray_anyrange_func(integer[], numrange) does not exist`,
			},
			{
				Statement: `drop function anyarray_anyrange_func(anyarray, anyrange);`,
			},
			{
				Statement: `create function bogus_func(anyelement)
  returns anyrange as 'select int4range(1,10)' language sql;`,
				ErrorString: `cannot determine result data type`,
			},
			{
				Statement: `create function bogus_func(int)
  returns anyrange as 'select int4range(1,10)' language sql;`,
				ErrorString: `cannot determine result data type`,
			},
			{
				Statement: `create function range_add_bounds(anyrange)
  returns anyelement as 'select lower($1) + upper($1)' language sql;`,
			},
			{
				Statement: `select range_add_bounds(int4range(1, 17));`,
				Results:   []sql.Row{{18}},
			},
			{
				Statement: `select range_add_bounds(numrange(1.0001, 123.123));`,
				Results:   []sql.Row{{124.1231}},
			},
			{
				Statement: `create function rangetypes_sql(q anyrange, b anyarray, out c anyelement)
  as $$ select upper($1) + $2[1] $$
  language sql;`,
			},
			{
				Statement: `select rangetypes_sql(int4range(1,10), ARRAY[2,20]);`,
				Results:   []sql.Row{{12}},
			},
			{
				Statement:   `select rangetypes_sql(numrange(1,10), ARRAY[2,20]);  -- match failure`,
				ErrorString: `function rangetypes_sql(numrange, integer[]) does not exist`,
			},
			{
				Statement: `create function anycompatiblearray_anycompatiblerange_func(a anycompatiblearray, r anycompatiblerange)
  returns anycompatible as 'select $1[1] + lower($2);' language sql;`,
			},
			{
				Statement: `select anycompatiblearray_anycompatiblerange_func(ARRAY[1,2], int4range(10,20));`,
				Results:   []sql.Row{{11}},
			},
			{
				Statement: `select anycompatiblearray_anycompatiblerange_func(ARRAY[1,2], numrange(10,20));`,
				Results:   []sql.Row{{11}},
			},
			{
				Statement:   `select anycompatiblearray_anycompatiblerange_func(ARRAY[1.1,2], int4range(10,20));`,
				ErrorString: `function anycompatiblearray_anycompatiblerange_func(numeric[], int4range) does not exist`,
			},
			{
				Statement: `drop function anycompatiblearray_anycompatiblerange_func(anycompatiblearray, anycompatiblerange);`,
			},
			{
				Statement: `create function bogus_func(anycompatible)
  returns anycompatiblerange as 'select int4range(1,10)' language sql;`,
				ErrorString: `cannot determine result data type`,
			},
			{
				Statement: `select ARRAY[numrange(1.1, 1.2), numrange(12.3, 155.5)];`,
				Results:   []sql.Row{{`{"[1.1,1.2)","[12.3,155.5)"}`}},
			},
			{
				Statement: `create table i8r_array (f1 int, f2 int8range[]);`,
			},
			{
				Statement: `insert into i8r_array values (42, array[int8range(1,10), int8range(2,20)]);`,
			},
			{
				Statement: `select * from i8r_array;`,
				Results:   []sql.Row{{42, `{"[1,10)","[2,20)"}`}},
			},
			{
				Statement: `drop table i8r_array;`,
			},
			{
				Statement: `create type arrayrange as range (subtype=int4[]);`,
			},
			{
				Statement: `select arrayrange(ARRAY[1,2], ARRAY[2,1]);`,
				Results:   []sql.Row{{`["{1,2}","{2,1}")`}},
			},
			{
				Statement:   `select arrayrange(ARRAY[2,1], ARRAY[1,2]);  -- fail`,
				ErrorString: `range lower bound must be less than or equal to range upper bound`,
			},
			{
				Statement: `select array[1,1] <@ arrayrange(array[1,2], array[2,1]);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select array[1,3] <@ arrayrange(array[1,2], array[2,1]);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `create type two_ints as (a int, b int);`,
			},
			{
				Statement: `create type two_ints_range as range (subtype = two_ints);`,
			},
			{
				Statement: `select *, row_to_json(upper(t)) as u from
  (values (two_ints_range(row(1,2), row(3,4))),
          (two_ints_range(row(5,6), row(7,8)))) v(t);`,
				Results: []sql.Row{{`["(1,2)","(3,4)")`, `{"a":3,"b":4}`}, {`["(5,6)","(7,8)")`, `{"a":7,"b":8}`}},
			},
			{
				Statement:   `alter type two_ints add attribute c two_ints_range;`,
				ErrorString: `composite type two_ints cannot be made a member of itself`,
			},
			{
				Statement: `drop type two_ints cascade;`,
			},
			{
				Statement: `create type cashrange as range (subtype = money);`,
			},
			{
				Statement: `set enable_sort = off;  -- try to make it pick a hash setop implementation`,
			},
			{
				Statement: `select '(2,5)'::cashrange except select '(5,6)'::cashrange;`,
				Results:   []sql.Row{{`($2.00,$5.00)`}},
			},
			{
				Statement: `reset enable_sort;`,
			},
			{
				Statement: `create function outparam_succeed(i anyrange, out r anyrange, out t text)
  as $$ select $1, 'foo'::text $$ language sql;`,
			},
			{
				Statement: `select * from outparam_succeed(int4range(1,2));`,
				Results:   []sql.Row{{`[1,2)`, `foo`}},
			},
			{
				Statement: `create function outparam2_succeed(r anyrange, out lu anyarray, out ul anyarray)
  as $$ select array[lower($1), upper($1)], array[upper($1), lower($1)] $$
  language sql;`,
			},
			{
				Statement: `select * from outparam2_succeed(int4range(1,11));`,
				Results:   []sql.Row{{`{1,11}`, `{11,1}`}},
			},
			{
				Statement: `create function outparam_succeed2(i anyrange, out r anyarray, out t text)
  as $$ select ARRAY[upper($1)], 'foo'::text $$ language sql;`,
			},
			{
				Statement: `select * from outparam_succeed2(int4range(int4range(1,2)));`,
				Results:   []sql.Row{{`{2}`, `foo`}},
			},
			{
				Statement: `create function inoutparam_succeed(out i anyelement, inout r anyrange)
  as $$ select upper($1), $1 $$ language sql;`,
			},
			{
				Statement: `select * from inoutparam_succeed(int4range(1,2));`,
				Results:   []sql.Row{{2, `[1,2)`}},
			},
			{
				Statement: `create function table_succeed(r anyrange)
  returns table(l anyelement, u anyelement)
  as $$ select lower($1), upper($1) $$
  language sql;`,
			},
			{
				Statement: `select * from table_succeed(int4range(1,11));`,
				Results:   []sql.Row{{1, 11}},
			},
			{
				Statement: `create function outparam_fail(i anyelement, out r anyrange, out t text)
  as $$ select '[1,10]', 'foo' $$ language sql;`,
				ErrorString: `cannot determine result data type`,
			},
			{
				Statement: `create function inoutparam_fail(inout i anyelement, out r anyrange)
  as $$ select $1, '[1,10]' $$ language sql;`,
				ErrorString: `cannot determine result data type`,
			},
			{
				Statement: `create function table_fail(i anyelement) returns table(i anyelement, r anyrange)
  as $$ select $1, '[1,10]' $$ language sql;`,
				ErrorString: `cannot determine result data type`,
			},
		},
	})
}
