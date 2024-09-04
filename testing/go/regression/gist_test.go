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

func TestGist(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_gist)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_gist,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `create table gist_point_tbl(id int4, p point);`,
			},
			{
				Statement: `create index gist_pointidx on gist_point_tbl using gist(p);`,
			},
			{
				Statement: `create index gist_pointidx2 on gist_point_tbl using gist(p) with (buffering = on, fillfactor=50);`,
			},
			{
				Statement: `create index gist_pointidx3 on gist_point_tbl using gist(p) with (buffering = off);`,
			},
			{
				Statement: `create index gist_pointidx4 on gist_point_tbl using gist(p) with (buffering = auto);`,
			},
			{
				Statement: `drop index gist_pointidx2, gist_pointidx3, gist_pointidx4;`,
			},
			{
				Statement:   `create index gist_pointidx5 on gist_point_tbl using gist(p) with (buffering = invalid_value);`,
				ErrorString: `invalid value for enum option "buffering": invalid_value`,
			},
			{
				Statement:   `create index gist_pointidx5 on gist_point_tbl using gist(p) with (fillfactor=9);`,
				ErrorString: `value 9 out of bounds for option "fillfactor"`,
			},
			{
				Statement:   `create index gist_pointidx5 on gist_point_tbl using gist(p) with (fillfactor=101);`,
				ErrorString: `value 101 out of bounds for option "fillfactor"`,
			},
			{
				Statement: `insert into gist_point_tbl (id, p)
select g,        point(g*10, g*10) from generate_series(1, 10000) g;`,
			},
			{
				Statement: `insert into gist_point_tbl (id, p)
select g+100000, point(g*10+1, g*10+1) from generate_series(1, 10000) g;`,
			},
			{
				Statement: `delete from gist_point_tbl where id % 2 = 1;`,
			},
			{
				Statement: `delete from gist_point_tbl where id > 5000;`,
			},
			{
				Statement: `vacuum analyze gist_point_tbl;`,
			},
			{
				Statement: `alter index gist_pointidx SET (fillfactor = 40);`,
			},
			{
				Statement: `reindex index gist_pointidx;`,
			},
			{
				Statement: `create table gist_tbl (b box, p point, c circle);`,
			},
			{
				Statement: `insert into gist_tbl
select box(point(0.05*i, 0.05*i), point(0.05*i, 0.05*i)),
       point(0.05*i, 0.05*i),
       circle(point(0.05*i, 0.05*i), 1.0)
from generate_series(0,10000) as i;`,
			},
			{
				Statement: `vacuum analyze gist_tbl;`,
			},
			{
				Statement: `set enable_seqscan=off;`,
			},
			{
				Statement: `set enable_bitmapscan=off;`,
			},
			{
				Statement: `set enable_indexonlyscan=on;`,
			},
			{
				Statement: `create index gist_tbl_point_index on gist_tbl using gist (p);`,
			},
			{
				Statement: `explain (costs off)
select p from gist_tbl where p <@ box(point(0,0), point(0.5, 0.5));`,
				Results: []sql.Row{{`Index Only Scan using gist_tbl_point_index on gist_tbl`}, {`Index Cond: (p <@ '(0.5,0.5),(0,0)'::box)`}},
			},
			{
				Statement: `select p from gist_tbl where p <@ box(point(0,0), point(0.5, 0.5));`,
				Results:   []sql.Row{{`(0,0)`}, {`(0.05,0.05)`}, {`(0.1,0.1)`}, {`(0.15,0.15)`}, {`(0.2,0.2)`}, {`(0.25,0.25)`}, {`(0.3,0.3)`}, {`(0.35,0.35)`}, {`(0.4,0.4)`}, {`(0.45,0.45)`}, {`(0.5,0.5)`}},
			},
			{
				Statement: `explain (costs off)
select p from gist_tbl where p <@ box(point(0,0), point(0.5, 0.5))
order by p <-> point(0.201, 0.201);`,
				Results: []sql.Row{{`Index Only Scan using gist_tbl_point_index on gist_tbl`}, {`Index Cond: (p <@ '(0.5,0.5),(0,0)'::box)`}, {`Order By: (p <-> '(0.201,0.201)'::point)`}},
			},
			{
				Statement: `select p from gist_tbl where p <@ box(point(0,0), point(0.5, 0.5))
order by p <-> point(0.201, 0.201);`,
				Results: []sql.Row{{`(0.2,0.2)`}, {`(0.25,0.25)`}, {`(0.15,0.15)`}, {`(0.3,0.3)`}, {`(0.1,0.1)`}, {`(0.35,0.35)`}, {`(0.05,0.05)`}, {`(0.4,0.4)`}, {`(0,0)`}, {`(0.45,0.45)`}, {`(0.5,0.5)`}},
			},
			{
				Statement: `explain (costs off)
select p from gist_tbl where p <@ box(point(0,0), point(0.5, 0.5))
order by point(0.101, 0.101) <-> p;`,
				Results: []sql.Row{{`Index Only Scan using gist_tbl_point_index on gist_tbl`}, {`Index Cond: (p <@ '(0.5,0.5),(0,0)'::box)`}, {`Order By: (p <-> '(0.101,0.101)'::point)`}},
			},
			{
				Statement: `select p from gist_tbl where p <@ box(point(0,0), point(0.5, 0.5))
order by point(0.101, 0.101) <-> p;`,
				Results: []sql.Row{{`(0.1,0.1)`}, {`(0.15,0.15)`}, {`(0.05,0.05)`}, {`(0.2,0.2)`}, {`(0,0)`}, {`(0.25,0.25)`}, {`(0.3,0.3)`}, {`(0.35,0.35)`}, {`(0.4,0.4)`}, {`(0.45,0.45)`}, {`(0.5,0.5)`}},
			},
			{
				Statement: `explain (costs off)
select p from
  (values (box(point(0,0), point(0.5,0.5))),
          (box(point(0.5,0.5), point(0.75,0.75))),
          (box(point(0.8,0.8), point(1.0,1.0)))) as v(bb)
cross join lateral
  (select p from gist_tbl where p <@ bb order by p <-> bb[0] limit 2) ss;`,
				Results: []sql.Row{{`Nested Loop`}, {`->  Values Scan on "*VALUES*"`}, {`->  Limit`}, {`->  Index Only Scan using gist_tbl_point_index on gist_tbl`}, {`Index Cond: (p <@ "*VALUES*".column1)`}, {`Order By: (p <-> ("*VALUES*".column1)[0])`}},
			},
			{
				Statement: `select p from
  (values (box(point(0,0), point(0.5,0.5))),
          (box(point(0.5,0.5), point(0.75,0.75))),
          (box(point(0.8,0.8), point(1.0,1.0)))) as v(bb)
cross join lateral
  (select p from gist_tbl where p <@ bb order by p <-> bb[0] limit 2) ss;`,
				Results: []sql.Row{{`(0.5,0.5)`}, {`(0.45,0.45)`}, {`(0.75,0.75)`}, {`(0.7,0.7)`}, {`(1,1)`}, {`(0.95,0.95)`}},
			},
			{
				Statement: `drop index gist_tbl_point_index;`,
			},
			{
				Statement: `create index gist_tbl_box_index on gist_tbl using gist (b);`,
			},
			{
				Statement: `explain (costs off)
select b from gist_tbl where b <@ box(point(5,5), point(6,6));`,
				Results: []sql.Row{{`Index Only Scan using gist_tbl_box_index on gist_tbl`}, {`Index Cond: (b <@ '(6,6),(5,5)'::box)`}},
			},
			{
				Statement: `select b from gist_tbl where b <@ box(point(5,5), point(6,6));`,
				Results:   []sql.Row{{`(5,5),(5,5)`}, {`(5.05,5.05),(5.05,5.05)`}, {`(5.1,5.1),(5.1,5.1)`}, {`(5.15,5.15),(5.15,5.15)`}, {`(5.2,5.2),(5.2,5.2)`}, {`(5.25,5.25),(5.25,5.25)`}, {`(5.3,5.3),(5.3,5.3)`}, {`(5.35,5.35),(5.35,5.35)`}, {`(5.4,5.4),(5.4,5.4)`}, {`(5.45,5.45),(5.45,5.45)`}, {`(5.5,5.5),(5.5,5.5)`}, {`(5.55,5.55),(5.55,5.55)`}, {`(5.6,5.6),(5.6,5.6)`}, {`(5.65,5.65),(5.65,5.65)`}, {`(5.7,5.7),(5.7,5.7)`}, {`(5.75,5.75),(5.75,5.75)`}, {`(5.8,5.8),(5.8,5.8)`}, {`(5.85,5.85),(5.85,5.85)`}, {`(5.9,5.9),(5.9,5.9)`}, {`(5.95,5.95),(5.95,5.95)`}, {`(6,6),(6,6)`}},
			},
			{
				Statement: `explain (costs off)
select b from gist_tbl where b <@ box(point(5,5), point(6,6))
order by b <-> point(5.2, 5.91);`,
				Results: []sql.Row{{`Index Only Scan using gist_tbl_box_index on gist_tbl`}, {`Index Cond: (b <@ '(6,6),(5,5)'::box)`}, {`Order By: (b <-> '(5.2,5.91)'::point)`}},
			},
			{
				Statement: `select b from gist_tbl where b <@ box(point(5,5), point(6,6))
order by b <-> point(5.2, 5.91);`,
				Results: []sql.Row{{`(5.55,5.55),(5.55,5.55)`}, {`(5.6,5.6),(5.6,5.6)`}, {`(5.5,5.5),(5.5,5.5)`}, {`(5.65,5.65),(5.65,5.65)`}, {`(5.45,5.45),(5.45,5.45)`}, {`(5.7,5.7),(5.7,5.7)`}, {`(5.4,5.4),(5.4,5.4)`}, {`(5.75,5.75),(5.75,5.75)`}, {`(5.35,5.35),(5.35,5.35)`}, {`(5.8,5.8),(5.8,5.8)`}, {`(5.3,5.3),(5.3,5.3)`}, {`(5.85,5.85),(5.85,5.85)`}, {`(5.25,5.25),(5.25,5.25)`}, {`(5.9,5.9),(5.9,5.9)`}, {`(5.2,5.2),(5.2,5.2)`}, {`(5.95,5.95),(5.95,5.95)`}, {`(5.15,5.15),(5.15,5.15)`}, {`(6,6),(6,6)`}, {`(5.1,5.1),(5.1,5.1)`}, {`(5.05,5.05),(5.05,5.05)`}, {`(5,5),(5,5)`}},
			},
			{
				Statement: `explain (costs off)
select b from gist_tbl where b <@ box(point(5,5), point(6,6))
order by point(5.2, 5.91) <-> b;`,
				Results: []sql.Row{{`Index Only Scan using gist_tbl_box_index on gist_tbl`}, {`Index Cond: (b <@ '(6,6),(5,5)'::box)`}, {`Order By: (b <-> '(5.2,5.91)'::point)`}},
			},
			{
				Statement: `select b from gist_tbl where b <@ box(point(5,5), point(6,6))
order by point(5.2, 5.91) <-> b;`,
				Results: []sql.Row{{`(5.55,5.55),(5.55,5.55)`}, {`(5.6,5.6),(5.6,5.6)`}, {`(5.5,5.5),(5.5,5.5)`}, {`(5.65,5.65),(5.65,5.65)`}, {`(5.45,5.45),(5.45,5.45)`}, {`(5.7,5.7),(5.7,5.7)`}, {`(5.4,5.4),(5.4,5.4)`}, {`(5.75,5.75),(5.75,5.75)`}, {`(5.35,5.35),(5.35,5.35)`}, {`(5.8,5.8),(5.8,5.8)`}, {`(5.3,5.3),(5.3,5.3)`}, {`(5.85,5.85),(5.85,5.85)`}, {`(5.25,5.25),(5.25,5.25)`}, {`(5.9,5.9),(5.9,5.9)`}, {`(5.2,5.2),(5.2,5.2)`}, {`(5.95,5.95),(5.95,5.95)`}, {`(5.15,5.15),(5.15,5.15)`}, {`(6,6),(6,6)`}, {`(5.1,5.1),(5.1,5.1)`}, {`(5.05,5.05),(5.05,5.05)`}, {`(5,5),(5,5)`}},
			},
			{
				Statement: `drop index gist_tbl_box_index;`,
			},
			{
				Statement: `create index gist_tbl_multi_index on gist_tbl using gist (p, c);`,
			},
			{
				Statement: `explain (costs off)
select p, c from gist_tbl
where p <@ box(point(5,5), point(6, 6));`,
				Results: []sql.Row{{`Index Scan using gist_tbl_multi_index on gist_tbl`}, {`Index Cond: (p <@ '(6,6),(5,5)'::box)`}},
			},
			{
				Statement: `select b, p from gist_tbl
where b <@ box(point(4.5, 4.5), point(5.5, 5.5))
and p <@ box(point(5,5), point(6, 6));`,
				Results: []sql.Row{{`(5,5),(5,5)`, `(5,5)`}, {`(5.05,5.05),(5.05,5.05)`, `(5.05,5.05)`}, {`(5.1,5.1),(5.1,5.1)`, `(5.1,5.1)`}, {`(5.15,5.15),(5.15,5.15)`, `(5.15,5.15)`}, {`(5.2,5.2),(5.2,5.2)`, `(5.2,5.2)`}, {`(5.25,5.25),(5.25,5.25)`, `(5.25,5.25)`}, {`(5.3,5.3),(5.3,5.3)`, `(5.3,5.3)`}, {`(5.35,5.35),(5.35,5.35)`, `(5.35,5.35)`}, {`(5.4,5.4),(5.4,5.4)`, `(5.4,5.4)`}, {`(5.45,5.45),(5.45,5.45)`, `(5.45,5.45)`}, {`(5.5,5.5),(5.5,5.5)`, `(5.5,5.5)`}},
			},
			{
				Statement: `drop index gist_tbl_multi_index;`,
			},
			{
				Statement: `create index gist_tbl_multi_index on gist_tbl using gist (circle(p,1), p);`,
			},
			{
				Statement: `explain (verbose, costs off)
select circle(p,1) from gist_tbl
where p <@ box(point(5, 5), point(5.3, 5.3));`,
				Results: []sql.Row{{`Index Only Scan using gist_tbl_multi_index on public.gist_tbl`}, {`Output: circle(p, '1'::double precision)`}, {`Index Cond: (gist_tbl.p <@ '(5.3,5.3),(5,5)'::box)`}},
			},
			{
				Statement: `select circle(p,1) from gist_tbl
where p <@ box(point(5, 5), point(5.3, 5.3));`,
				Results: []sql.Row{{`<(5,5),1>`}, {`<(5.05,5.05),1>`}, {`<(5.1,5.1),1>`}, {`<(5.15,5.15),1>`}, {`<(5.2,5.2),1>`}, {`<(5.25,5.25),1>`}, {`<(5.3,5.3),1>`}},
			},
			{
				Statement: `explain (verbose, costs off)
select p from gist_tbl where circle(p,1) @> circle(point(0,0),0.95);`,
				Results: []sql.Row{{`Index Only Scan using gist_tbl_multi_index on public.gist_tbl`}, {`Output: p`}, {`Index Cond: ((circle(gist_tbl.p, '1'::double precision)) @> '<(0,0),0.95>'::circle)`}},
			},
			{
				Statement: `select p from gist_tbl where circle(p,1) @> circle(point(0,0),0.95);`,
				Results:   []sql.Row{{`(0,0)`}},
			},
			{
				Statement: `explain (verbose, costs off)
select count(*) from gist_tbl;`,
				Results: []sql.Row{{`Aggregate`}, {`Output: count(*)`}, {`->  Index Only Scan using gist_tbl_multi_index on public.gist_tbl`}},
			},
			{
				Statement: `select count(*) from gist_tbl;`,
				Results:   []sql.Row{{10001}},
			},
			{
				Statement: `explain (verbose, costs off)
select p from gist_tbl order by circle(p,1) <-> point(0,0) limit 1;`,
				Results: []sql.Row{{`Limit`}, {`Output: p, ((circle(p, '1'::double precision) <-> '(0,0)'::point))`}, {`->  Index Only Scan using gist_tbl_multi_index on public.gist_tbl`}, {`Output: p, (circle(p, '1'::double precision) <-> '(0,0)'::point)`}, {`Order By: ((circle(gist_tbl.p, '1'::double precision)) <-> '(0,0)'::point)`}},
			},
			{
				Statement:   `select p from gist_tbl order by circle(p,1) <-> point(0,0) limit 1;`,
				ErrorString: `lossy distance functions are not supported in index-only scans`,
			},
			{
				Statement: `reset enable_seqscan;`,
			},
			{
				Statement: `reset enable_bitmapscan;`,
			},
			{
				Statement: `reset enable_indexonlyscan;`,
			},
			{
				Statement: `drop table gist_tbl;`,
			},
			{
				Statement: `create unlogged table gist_tbl (b box);`,
			},
			{
				Statement: `create index gist_tbl_box_index on gist_tbl using gist (b);`,
			},
			{
				Statement: `insert into gist_tbl
  select box(point(0.05*i, 0.05*i)) from generate_series(0,10) as i;`,
			},
			{
				Statement: `drop table gist_tbl;`,
			},
		},
	})
}
