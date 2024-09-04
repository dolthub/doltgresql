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

func TestSpgist(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_spgist)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_spgist,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `create table spgist_point_tbl(id int4, p point);`,
			},
			{
				Statement: `create index spgist_point_idx on spgist_point_tbl using spgist(p) with (fillfactor = 75);`,
			},
			{
				Statement: `insert into spgist_point_tbl (id, p)
select g, point(g*10, g*10) from generate_series(1, 10) g;`,
			},
			{
				Statement: `delete from spgist_point_tbl where id < 5;`,
			},
			{
				Statement: `vacuum spgist_point_tbl;`,
			},
			{
				Statement: `insert into spgist_point_tbl (id, p)
select g,      point(g*10, g*10) from generate_series(1, 10000) g;`,
			},
			{
				Statement: `insert into spgist_point_tbl (id, p)
select g+100000, point(g*10+1, g*10+1) from generate_series(1, 10000) g;`,
			},
			{
				Statement: `delete from spgist_point_tbl where id % 2 = 1;`,
			},
			{
				Statement: `delete from spgist_point_tbl where id < 10000;`,
			},
			{
				Statement: `vacuum spgist_point_tbl;`,
			},
			{
				Statement: `create table spgist_box_tbl(id serial, b box);`,
			},
			{
				Statement: `insert into spgist_box_tbl(b)
select box(point(i,j),point(i+s,j+s))
  from generate_series(1,100,5) i,
       generate_series(1,100,5) j,
       generate_series(1,10) s;`,
			},
			{
				Statement: `create index spgist_box_idx on spgist_box_tbl using spgist (b);`,
			},
			{
				Statement: `select count(*)
  from (values (point(5,5)),(point(8,8)),(point(12,12))) v(p)
 where exists(select * from spgist_box_tbl b where b.b && box(v.p,v.p));`,
				Results: []sql.Row{{3}},
			},
			{
				Statement: `create table spgist_text_tbl(id int4, t text);`,
			},
			{
				Statement: `create index spgist_text_idx on spgist_text_tbl using spgist(t);`,
			},
			{
				Statement: `insert into spgist_text_tbl (id, t)
select g, 'f' || repeat('o', 100) || g from generate_series(1, 10000) g
union all
select g, 'baaaaaaaaaaaaaar' || g from generate_series(1, 1000) g;`,
			},
			{
				Statement: `insert into spgist_text_tbl (id, t)
select -g, 'f' || repeat('o', 100-g) || 'surprise' from generate_series(1, 100) g;`,
			},
			{
				Statement:   `create index spgist_point_idx2 on spgist_point_tbl using spgist(p) with (fillfactor = 9);`,
				ErrorString: `value 9 out of bounds for option "fillfactor"`,
			},
			{
				Statement:   `create index spgist_point_idx2 on spgist_point_tbl using spgist(p) with (fillfactor = 101);`,
				ErrorString: `value 101 out of bounds for option "fillfactor"`,
			},
			{
				Statement: `alter index spgist_point_idx set (fillfactor = 90);`,
			},
			{
				Statement: `reindex index spgist_point_idx;`,
			},
			{
				Statement: `create domain spgist_text as varchar;`,
			},
			{
				Statement: `create table spgist_domain_tbl (f1 spgist_text);`,
			},
			{
				Statement: `create index spgist_domain_idx on spgist_domain_tbl using spgist(f1);`,
			},
			{
				Statement: `insert into spgist_domain_tbl values('fee'), ('fi'), ('fo'), ('fum');`,
			},
			{
				Statement: `explain (costs off)
select * from spgist_domain_tbl where f1 = 'fo';`,
				Results: []sql.Row{{`Bitmap Heap Scan on spgist_domain_tbl`}, {`Recheck Cond: ((f1)::text = 'fo'::text)`}, {`->  Bitmap Index Scan on spgist_domain_idx`}, {`Index Cond: ((f1)::text = 'fo'::text)`}},
			},
			{
				Statement: `select * from spgist_domain_tbl where f1 = 'fo';`,
				Results:   []sql.Row{{`fo`}},
			},
			{
				Statement: `create unlogged table spgist_unlogged_tbl(id serial, b box);`,
			},
			{
				Statement: `create index spgist_unlogged_idx on spgist_unlogged_tbl using spgist (b);`,
			},
			{
				Statement: `insert into spgist_unlogged_tbl(b)
select box(point(i,j))
  from generate_series(1,100,5) i,
       generate_series(1,10,5) j;`,
			},
		},
	})
}
