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

func TestAmutils(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_amutils)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_amutils,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup, RegressionFileName_geometry, RegressionFileName_create_index_spgist, RegressionFileName_hash_index, RegressionFileName_brin},
		Statements: []RegressionFileStatement{
			{
				Statement: `select prop,
       pg_indexam_has_property(a.oid, prop) as "AM",
       pg_index_has_property('onek_hundred'::regclass, prop) as "Index",
       pg_index_column_has_property('onek_hundred'::regclass, 1, prop) as "Column"
  from pg_am a,
       unnest(array['asc', 'desc', 'nulls_first', 'nulls_last',
                    'orderable', 'distance_orderable', 'returnable',
                    'search_array', 'search_nulls',
                    'clusterable', 'index_scan', 'bitmap_scan',
                    'backward_scan',
                    'can_order', 'can_unique', 'can_multi_col',
                    'can_exclude', 'can_include',
                    'bogus']::text[])
         with ordinality as u(prop,ord)
 where a.amname = 'btree'
 order by ord;`,
				Results: []sql.Row{{`asc`, ``, ``, true}, {`desc`, ``, ``, false}, {`nulls_first`, ``, ``, false}, {`nulls_last`, ``, ``, true}, {`orderable`, ``, ``, true}, {`distance_orderable`, ``, ``, false}, {`returnable`, ``, ``, true}, {`search_array`, ``, ``, true}, {`search_nulls`, ``, ``, true}, {`clusterable`, ``, true, ``}, {`index_scan`, ``, true, ``}, {`bitmap_scan`, ``, true, ``}, {`backward_scan`, ``, true, ``}, {`can_order`, true, ``, ``}, {`can_unique`, true, ``, ``}, {`can_multi_col`, true, ``, ``}, {`can_exclude`, true, ``, ``}, {`can_include`, true, ``, ``}, {`bogus`, ``, ``, ``}},
			},
			{
				Statement: `select prop,
       pg_indexam_has_property(a.oid, prop) as "AM",
       pg_index_has_property('gcircleind'::regclass, prop) as "Index",
       pg_index_column_has_property('gcircleind'::regclass, 1, prop) as "Column"
  from pg_am a,
       unnest(array['asc', 'desc', 'nulls_first', 'nulls_last',
                    'orderable', 'distance_orderable', 'returnable',
                    'search_array', 'search_nulls',
                    'clusterable', 'index_scan', 'bitmap_scan',
                    'backward_scan',
                    'can_order', 'can_unique', 'can_multi_col',
                    'can_exclude', 'can_include',
                    'bogus']::text[])
         with ordinality as u(prop,ord)
 where a.amname = 'gist'
 order by ord;`,
				Results: []sql.Row{{`asc`, ``, ``, false}, {`desc`, ``, ``, false}, {`nulls_first`, ``, ``, false}, {`nulls_last`, ``, ``, false}, {`orderable`, ``, ``, false}, {`distance_orderable`, ``, ``, true}, {`returnable`, ``, ``, false}, {`search_array`, ``, ``, false}, {`search_nulls`, ``, ``, true}, {`clusterable`, ``, true, ``}, {`index_scan`, ``, true, ``}, {`bitmap_scan`, ``, true, ``}, {`backward_scan`, ``, false, ``}, {`can_order`, false, ``, ``}, {`can_unique`, false, ``, ``}, {`can_multi_col`, true, ``, ``}, {`can_exclude`, true, ``, ``}, {`can_include`, true, ``, ``}, {`bogus`, ``, ``, ``}},
			},
			{
				Statement: `select prop,
       pg_index_column_has_property('onek_hundred'::regclass, 1, prop) as btree,
       pg_index_column_has_property('hash_i4_index'::regclass, 1, prop) as hash,
       pg_index_column_has_property('gcircleind'::regclass, 1, prop) as gist,
       pg_index_column_has_property('sp_radix_ind'::regclass, 1, prop) as spgist_radix,
       pg_index_column_has_property('sp_quad_ind'::regclass, 1, prop) as spgist_quad,
       pg_index_column_has_property('botharrayidx'::regclass, 1, prop) as gin,
       pg_index_column_has_property('brinidx'::regclass, 1, prop) as brin
  from unnest(array['asc', 'desc', 'nulls_first', 'nulls_last',
                    'orderable', 'distance_orderable', 'returnable',
                    'search_array', 'search_nulls',
                    'bogus']::text[])
         with ordinality as u(prop,ord)
 order by ord;`,
				Results: []sql.Row{{`asc`, true, false, false, false, false, false, false}, {`desc`, false, false, false, false, false, false, false}, {`nulls_first`, false, false, false, false, false, false, false}, {`nulls_last`, true, false, false, false, false, false, false}, {`orderable`, true, false, false, false, false, false, false}, {`distance_orderable`, false, false, true, false, true, false, false}, {`returnable`, true, false, false, true, true, false, false}, {`search_array`, true, false, false, false, false, false, false}, {`search_nulls`, true, false, true, true, true, false, true}, {`bogus`, ``, ``, ``, ``, ``, ``, ``}},
			},
			{
				Statement: `select prop,
       pg_index_has_property('onek_hundred'::regclass, prop) as btree,
       pg_index_has_property('hash_i4_index'::regclass, prop) as hash,
       pg_index_has_property('gcircleind'::regclass, prop) as gist,
       pg_index_has_property('sp_radix_ind'::regclass, prop) as spgist,
       pg_index_has_property('botharrayidx'::regclass, prop) as gin,
       pg_index_has_property('brinidx'::regclass, prop) as brin
  from unnest(array['clusterable', 'index_scan', 'bitmap_scan',
                    'backward_scan',
                    'bogus']::text[])
         with ordinality as u(prop,ord)
 order by ord;`,
				Results: []sql.Row{{`clusterable`, true, false, true, false, false, false}, {`index_scan`, true, true, true, true, false, false}, {`bitmap_scan`, true, true, true, true, true, true}, {`backward_scan`, true, true, false, false, false, false}, {`bogus`, ``, ``, ``, ``, ``, ``}},
			},
			{
				Statement: `select amname, prop, pg_indexam_has_property(a.oid, prop) as p
  from pg_am a,
       unnest(array['can_order', 'can_unique', 'can_multi_col',
                    'can_exclude', 'can_include', 'bogus']::text[])
         with ordinality as u(prop,ord)
 where amtype = 'i'
 order by amname, ord;`,
				Results: []sql.Row{{`brin`, `can_order`, false}, {`brin`, `can_unique`, false}, {`brin`, `can_multi_col`, true}, {`brin`, `can_exclude`, false}, {`brin`, `can_include`, false}, {`brin`, `bogus`, ``}, {`btree`, `can_order`, true}, {`btree`, `can_unique`, true}, {`btree`, `can_multi_col`, true}, {`btree`, `can_exclude`, true}, {`btree`, `can_include`, true}, {`btree`, `bogus`, ``}, {`gin`, `can_order`, false}, {`gin`, `can_unique`, false}, {`gin`, `can_multi_col`, true}, {`gin`, `can_exclude`, false}, {`gin`, `can_include`, false}, {`gin`, `bogus`, ``}, {`gist`, `can_order`, false}, {`gist`, `can_unique`, false}, {`gist`, `can_multi_col`, true}, {`gist`, `can_exclude`, true}, {`gist`, `can_include`, true}, {`gist`, `bogus`, ``}, {`hash`, `can_order`, false}, {`hash`, `can_unique`, false}, {`hash`, `can_multi_col`, false}, {`hash`, `can_exclude`, true}, {`hash`, `can_include`, false}, {`hash`, `bogus`, ``}, {`spgist`, `can_order`, false}, {`spgist`, `can_unique`, false}, {`spgist`, `can_multi_col`, false}, {`spgist`, `can_exclude`, true}, {`spgist`, `can_include`, true}, {`spgist`, `bogus`, ``}},
			},
			{
				Statement: `CREATE TEMP TABLE foo (f1 int, f2 int, f3 int, f4 int);`,
			},
			{
				Statement: `CREATE INDEX fooindex ON foo (f1 desc, f2 asc, f3 nulls first, f4 nulls last);`,
			},
			{
				Statement: `select col, prop, pg_index_column_has_property(o, col, prop)
  from (values ('fooindex'::regclass)) v1(o),
       (values (1,'orderable'),(2,'asc'),(3,'desc'),
               (4,'nulls_first'),(5,'nulls_last'),
               (6, 'bogus')) v2(idx,prop),
       generate_series(1,4) col
 order by col, idx;`,
				Results: []sql.Row{{1, `orderable`, true}, {1, `asc`, false}, {1, `desc`, true}, {1, `nulls_first`, true}, {1, `nulls_last`, false}, {1, `bogus`, ``}, {2, `orderable`, true}, {2, `asc`, true}, {2, `desc`, false}, {2, `nulls_first`, false}, {2, `nulls_last`, true}, {2, `bogus`, ``}, {3, `orderable`, true}, {3, `asc`, true}, {3, `desc`, false}, {3, `nulls_first`, true}, {3, `nulls_last`, false}, {3, `bogus`, ``}, {4, `orderable`, true}, {4, `asc`, true}, {4, `desc`, false}, {4, `nulls_first`, false}, {4, `nulls_last`, true}, {4, `bogus`, ``}},
			},
			{
				Statement: `CREATE INDEX foocover ON foo (f1) INCLUDE (f2,f3);`,
			},
			{
				Statement: `select col, prop, pg_index_column_has_property(o, col, prop)
  from (values ('foocover'::regclass)) v1(o),
       (values (1,'orderable'),(2,'asc'),(3,'desc'),
               (4,'nulls_first'),(5,'nulls_last'),
               (6,'distance_orderable'),(7,'returnable'),
               (8, 'bogus')) v2(idx,prop),
       generate_series(1,3) col
 order by col, idx;`,
				Results: []sql.Row{{1, `orderable`, true}, {1, `asc`, true}, {1, `desc`, false}, {1, `nulls_first`, false}, {1, `nulls_last`, true}, {1, `distance_orderable`, false}, {1, `returnable`, true}, {1, `bogus`, ``}, {2, `orderable`, false}, {2, `asc`, ``}, {2, `desc`, ``}, {2, `nulls_first`, ``}, {2, `nulls_last`, ``}, {2, `distance_orderable`, false}, {2, `returnable`, true}, {2, `bogus`, ``}, {3, `orderable`, false}, {3, `asc`, ``}, {3, `desc`, ``}, {3, `nulls_first`, ``}, {3, `nulls_last`, ``}, {3, `distance_orderable`, false}, {3, `returnable`, true}, {3, `bogus`, ``}},
			},
		},
	})
}
