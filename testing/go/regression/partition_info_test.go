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

func TestPartitionInfo(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_partition_info)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_partition_info,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `SELECT * FROM pg_partition_tree(NULL);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT * FROM pg_partition_tree(0);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT * FROM pg_partition_ancestors(NULL);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT * FROM pg_partition_ancestors(0);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT pg_partition_root(NULL);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT pg_partition_root(0);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `CREATE TABLE ptif_test (a int, b int) PARTITION BY range (a);`,
			},
			{
				Statement: `CREATE TABLE ptif_test0 PARTITION OF ptif_test
  FOR VALUES FROM (minvalue) TO (0) PARTITION BY list (b);`,
			},
			{
				Statement: `CREATE TABLE ptif_test01 PARTITION OF ptif_test0 FOR VALUES IN (1);`,
			},
			{
				Statement: `CREATE TABLE ptif_test1 PARTITION OF ptif_test
  FOR VALUES FROM (0) TO (100) PARTITION BY list (b);`,
			},
			{
				Statement: `CREATE TABLE ptif_test11 PARTITION OF ptif_test1 FOR VALUES IN (1);`,
			},
			{
				Statement: `CREATE TABLE ptif_test2 PARTITION OF ptif_test
  FOR VALUES FROM (100) TO (200);`,
			},
			{
				Statement: `CREATE TABLE ptif_test3 PARTITION OF ptif_test
  FOR VALUES FROM (200) TO (maxvalue) PARTITION BY list (b);`,
			},
			{
				Statement: `SELECT pg_partition_root('ptif_test');`,
				Results:   []sql.Row{{`ptif_test`}},
			},
			{
				Statement: `SELECT pg_partition_root('ptif_test0');`,
				Results:   []sql.Row{{`ptif_test`}},
			},
			{
				Statement: `SELECT pg_partition_root('ptif_test01');`,
				Results:   []sql.Row{{`ptif_test`}},
			},
			{
				Statement: `SELECT pg_partition_root('ptif_test3');`,
				Results:   []sql.Row{{`ptif_test`}},
			},
			{
				Statement: `CREATE INDEX ptif_test_index ON ONLY ptif_test (a);`,
			},
			{
				Statement: `CREATE INDEX ptif_test0_index ON ONLY ptif_test0 (a);`,
			},
			{
				Statement: `ALTER INDEX ptif_test_index ATTACH PARTITION ptif_test0_index;`,
			},
			{
				Statement: `CREATE INDEX ptif_test01_index ON ptif_test01 (a);`,
			},
			{
				Statement: `ALTER INDEX ptif_test0_index ATTACH PARTITION ptif_test01_index;`,
			},
			{
				Statement: `CREATE INDEX ptif_test1_index ON ONLY ptif_test1 (a);`,
			},
			{
				Statement: `ALTER INDEX ptif_test_index ATTACH PARTITION ptif_test1_index;`,
			},
			{
				Statement: `CREATE INDEX ptif_test11_index ON ptif_test11 (a);`,
			},
			{
				Statement: `ALTER INDEX ptif_test1_index ATTACH PARTITION ptif_test11_index;`,
			},
			{
				Statement: `CREATE INDEX ptif_test2_index ON ptif_test2 (a);`,
			},
			{
				Statement: `ALTER INDEX ptif_test_index ATTACH PARTITION ptif_test2_index;`,
			},
			{
				Statement: `CREATE INDEX ptif_test3_index ON ptif_test3 (a);`,
			},
			{
				Statement: `ALTER INDEX ptif_test_index ATTACH PARTITION ptif_test3_index;`,
			},
			{
				Statement: `SELECT pg_partition_root('ptif_test_index');`,
				Results:   []sql.Row{{`ptif_test_index`}},
			},
			{
				Statement: `SELECT pg_partition_root('ptif_test0_index');`,
				Results:   []sql.Row{{`ptif_test_index`}},
			},
			{
				Statement: `SELECT pg_partition_root('ptif_test01_index');`,
				Results:   []sql.Row{{`ptif_test_index`}},
			},
			{
				Statement: `SELECT pg_partition_root('ptif_test3_index');`,
				Results:   []sql.Row{{`ptif_test_index`}},
			},
			{
				Statement: `SELECT relid, parentrelid, level, isleaf
  FROM pg_partition_tree('ptif_test');`,
				Results: []sql.Row{{`ptif_test`, ``, 0, false}, {`ptif_test0`, `ptif_test`, 1, false}, {`ptif_test1`, `ptif_test`, 1, false}, {`ptif_test2`, `ptif_test`, 1, true}, {`ptif_test3`, `ptif_test`, 1, false}, {`ptif_test01`, `ptif_test0`, 2, true}, {`ptif_test11`, `ptif_test1`, 2, true}},
			},
			{
				Statement: `SELECT relid, parentrelid, level, isleaf
  FROM pg_partition_tree('ptif_test0') p
  JOIN pg_class c ON (p.relid = c.oid);`,
				Results: []sql.Row{{`ptif_test0`, `ptif_test`, 0, false}, {`ptif_test01`, `ptif_test0`, 1, true}},
			},
			{
				Statement: `SELECT relid, parentrelid, level, isleaf
  FROM pg_partition_tree('ptif_test01') p
  JOIN pg_class c ON (p.relid = c.oid);`,
				Results: []sql.Row{{`ptif_test01`, `ptif_test0`, 0, true}},
			},
			{
				Statement: `SELECT relid, parentrelid, level, isleaf
  FROM pg_partition_tree('ptif_test3') p
  JOIN pg_class c ON (p.relid = c.oid);`,
				Results: []sql.Row{{`ptif_test3`, `ptif_test`, 0, false}},
			},
			{
				Statement: `SELECT * FROM pg_partition_ancestors('ptif_test01');`,
				Results:   []sql.Row{{`ptif_test01`}, {`ptif_test0`}, {`ptif_test`}},
			},
			{
				Statement: `SELECT * FROM pg_partition_ancestors('ptif_test');`,
				Results:   []sql.Row{{`ptif_test`}},
			},
			{
				Statement: `SELECT relid, parentrelid, level, isleaf
  FROM pg_partition_tree(pg_partition_root('ptif_test01')) p
  JOIN pg_class c ON (p.relid = c.oid);`,
				Results: []sql.Row{{`ptif_test`, ``, 0, false}, {`ptif_test0`, `ptif_test`, 1, false}, {`ptif_test1`, `ptif_test`, 1, false}, {`ptif_test2`, `ptif_test`, 1, true}, {`ptif_test3`, `ptif_test`, 1, false}, {`ptif_test01`, `ptif_test0`, 2, true}, {`ptif_test11`, `ptif_test1`, 2, true}},
			},
			{
				Statement: `SELECT relid, parentrelid, level, isleaf
  FROM pg_partition_tree('ptif_test_index');`,
				Results: []sql.Row{{`ptif_test_index`, ``, 0, false}, {`ptif_test0_index`, `ptif_test_index`, 1, false}, {`ptif_test1_index`, `ptif_test_index`, 1, false}, {`ptif_test2_index`, `ptif_test_index`, 1, true}, {`ptif_test3_index`, `ptif_test_index`, 1, false}, {`ptif_test01_index`, `ptif_test0_index`, 2, true}, {`ptif_test11_index`, `ptif_test1_index`, 2, true}},
			},
			{
				Statement: `SELECT relid, parentrelid, level, isleaf
  FROM pg_partition_tree('ptif_test0_index') p
  JOIN pg_class c ON (p.relid = c.oid);`,
				Results: []sql.Row{{`ptif_test0_index`, `ptif_test_index`, 0, false}, {`ptif_test01_index`, `ptif_test0_index`, 1, true}},
			},
			{
				Statement: `SELECT relid, parentrelid, level, isleaf
  FROM pg_partition_tree('ptif_test01_index') p
  JOIN pg_class c ON (p.relid = c.oid);`,
				Results: []sql.Row{{`ptif_test01_index`, `ptif_test0_index`, 0, true}},
			},
			{
				Statement: `SELECT relid, parentrelid, level, isleaf
  FROM pg_partition_tree('ptif_test3_index') p
  JOIN pg_class c ON (p.relid = c.oid);`,
				Results: []sql.Row{{`ptif_test3_index`, `ptif_test_index`, 0, false}},
			},
			{
				Statement: `SELECT relid, parentrelid, level, isleaf
  FROM pg_partition_tree(pg_partition_root('ptif_test01_index')) p
  JOIN pg_class c ON (p.relid = c.oid);`,
				Results: []sql.Row{{`ptif_test_index`, ``, 0, false}, {`ptif_test0_index`, `ptif_test_index`, 1, false}, {`ptif_test1_index`, `ptif_test_index`, 1, false}, {`ptif_test2_index`, `ptif_test_index`, 1, true}, {`ptif_test3_index`, `ptif_test_index`, 1, false}, {`ptif_test01_index`, `ptif_test0_index`, 2, true}, {`ptif_test11_index`, `ptif_test1_index`, 2, true}},
			},
			{
				Statement: `SELECT * FROM pg_partition_ancestors('ptif_test01_index');`,
				Results:   []sql.Row{{`ptif_test01_index`}, {`ptif_test0_index`}, {`ptif_test_index`}},
			},
			{
				Statement: `SELECT * FROM pg_partition_ancestors('ptif_test_index');`,
				Results:   []sql.Row{{`ptif_test_index`}},
			},
			{
				Statement: `DROP TABLE ptif_test;`,
			},
			{
				Statement: `CREATE TABLE ptif_normal_table(a int);`,
			},
			{
				Statement: `SELECT relid, parentrelid, level, isleaf
  FROM pg_partition_tree('ptif_normal_table');`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT * FROM pg_partition_ancestors('ptif_normal_table');`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT pg_partition_root('ptif_normal_table');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `DROP TABLE ptif_normal_table;`,
			},
			{
				Statement: `CREATE VIEW ptif_test_view AS SELECT 1;`,
			},
			{
				Statement: `CREATE MATERIALIZED VIEW ptif_test_matview AS SELECT 1;`,
			},
			{
				Statement: `CREATE TABLE ptif_li_parent ();`,
			},
			{
				Statement: `CREATE TABLE ptif_li_child () INHERITS (ptif_li_parent);`,
			},
			{
				Statement: `SELECT * FROM pg_partition_tree('ptif_test_view');`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT * FROM pg_partition_tree('ptif_test_matview');`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT * FROM pg_partition_tree('ptif_li_parent');`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT * FROM pg_partition_tree('ptif_li_child');`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT * FROM pg_partition_ancestors('ptif_test_view');`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT * FROM pg_partition_ancestors('ptif_test_matview');`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT * FROM pg_partition_ancestors('ptif_li_parent');`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT * FROM pg_partition_ancestors('ptif_li_child');`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT pg_partition_root('ptif_test_view');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT pg_partition_root('ptif_test_matview');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT pg_partition_root('ptif_li_parent');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT pg_partition_root('ptif_li_child');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `DROP VIEW ptif_test_view;`,
			},
			{
				Statement: `DROP MATERIALIZED VIEW ptif_test_matview;`,
			},
			{
				Statement: `DROP TABLE ptif_li_parent, ptif_li_child;`,
			},
		},
	})
}
