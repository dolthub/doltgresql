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

func TestTablesample(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_tablesample)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_tablesample,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TABLE test_tablesample (id int, name text) WITH (fillfactor=10);`,
			},
			{
				Statement: `INSERT INTO test_tablesample
  SELECT i, repeat(i::text, 200) FROM generate_series(0, 9) s(i);`,
			},
			{
				Statement: `SELECT t.id FROM test_tablesample AS t TABLESAMPLE SYSTEM (50) REPEATABLE (0);`,
				Results:   []sql.Row{{3}, {4}, {5}, {6}, {7}, {8}},
			},
			{
				Statement: `SELECT id FROM test_tablesample TABLESAMPLE SYSTEM (100.0/11) REPEATABLE (0);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT id FROM test_tablesample TABLESAMPLE SYSTEM (50) REPEATABLE (0);`,
				Results:   []sql.Row{{3}, {4}, {5}, {6}, {7}, {8}},
			},
			{
				Statement: `SELECT id FROM test_tablesample TABLESAMPLE BERNOULLI (50) REPEATABLE (0);`,
				Results:   []sql.Row{{4}, {5}, {6}, {7}, {8}},
			},
			{
				Statement: `SELECT id FROM test_tablesample TABLESAMPLE BERNOULLI (5.5) REPEATABLE (0);`,
				Results:   []sql.Row{{7}},
			},
			{
				Statement: `SELECT count(*) FROM test_tablesample TABLESAMPLE SYSTEM (100);`,
				Results:   []sql.Row{{10}},
			},
			{
				Statement: `SELECT count(*) FROM test_tablesample TABLESAMPLE SYSTEM (100) REPEATABLE (1+2);`,
				Results:   []sql.Row{{10}},
			},
			{
				Statement: `SELECT count(*) FROM test_tablesample TABLESAMPLE SYSTEM (100) REPEATABLE (0.4);`,
				Results:   []sql.Row{{10}},
			},
			{
				Statement: `CREATE VIEW test_tablesample_v1 AS
  SELECT id FROM test_tablesample TABLESAMPLE SYSTEM (10*2) REPEATABLE (2);`,
			},
			{
				Statement: `CREATE VIEW test_tablesample_v2 AS
  SELECT id FROM test_tablesample TABLESAMPLE SYSTEM (99);`,
			},
			{
				Statement: `\d+ test_tablesample_v1
                     View "public.test_tablesample_v1"
 Column |  Type   | Collation | Nullable | Default | Storage | Description 
--------+---------+-----------+----------+---------+---------+-------------
 id     | integer |           |          |         | plain   | 
View definition:
 SELECT test_tablesample.id
   FROM test_tablesample TABLESAMPLE system ((10 * 2)) REPEATABLE (2);`,
			},
			{
				Statement: `\d+ test_tablesample_v2
                     View "public.test_tablesample_v2"
 Column |  Type   | Collation | Nullable | Default | Storage | Description 
--------+---------+-----------+----------+---------+---------+-------------
 id     | integer |           |          |         | plain   | 
View definition:
 SELECT test_tablesample.id
   FROM test_tablesample TABLESAMPLE system (99);`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `DECLARE tablesample_cur SCROLL CURSOR FOR
  SELECT id FROM test_tablesample TABLESAMPLE SYSTEM (50) REPEATABLE (0);`,
			},
			{
				Statement: `FETCH FIRST FROM tablesample_cur;`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement: `FETCH NEXT FROM tablesample_cur;`,
				Results:   []sql.Row{{4}},
			},
			{
				Statement: `FETCH NEXT FROM tablesample_cur;`,
				Results:   []sql.Row{{5}},
			},
			{
				Statement: `SELECT id FROM test_tablesample TABLESAMPLE SYSTEM (50) REPEATABLE (0);`,
				Results:   []sql.Row{{3}, {4}, {5}, {6}, {7}, {8}},
			},
			{
				Statement: `FETCH NEXT FROM tablesample_cur;`,
				Results:   []sql.Row{{6}},
			},
			{
				Statement: `FETCH NEXT FROM tablesample_cur;`,
				Results:   []sql.Row{{7}},
			},
			{
				Statement: `FETCH NEXT FROM tablesample_cur;`,
				Results:   []sql.Row{{8}},
			},
			{
				Statement: `FETCH FIRST FROM tablesample_cur;`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement: `FETCH NEXT FROM tablesample_cur;`,
				Results:   []sql.Row{{4}},
			},
			{
				Statement: `FETCH NEXT FROM tablesample_cur;`,
				Results:   []sql.Row{{5}},
			},
			{
				Statement: `FETCH NEXT FROM tablesample_cur;`,
				Results:   []sql.Row{{6}},
			},
			{
				Statement: `FETCH NEXT FROM tablesample_cur;`,
				Results:   []sql.Row{{7}},
			},
			{
				Statement: `FETCH NEXT FROM tablesample_cur;`,
				Results:   []sql.Row{{8}},
			},
			{
				Statement: `CLOSE tablesample_cur;`,
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
  SELECT id FROM test_tablesample TABLESAMPLE SYSTEM (50) REPEATABLE (2);`,
				Results: []sql.Row{{`Sample Scan on test_tablesample`}, {`Sampling: system ('50'::real) REPEATABLE ('2'::double precision)`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
  SELECT * FROM test_tablesample_v1;`,
				Results: []sql.Row{{`Sample Scan on test_tablesample`}, {`Sampling: system ('20'::real) REPEATABLE ('2'::double precision)`}},
			},
			{
				Statement: `explain (costs off)
  select count(*) from person tablesample bernoulli (100);`,
				Results: []sql.Row{{`Aggregate`}, {`->  Append`}, {`->  Sample Scan on person person_1`}, {`Sampling: bernoulli ('100'::real)`}, {`->  Sample Scan on emp person_2`}, {`Sampling: bernoulli ('100'::real)`}, {`->  Sample Scan on student person_3`}, {`Sampling: bernoulli ('100'::real)`}, {`->  Sample Scan on stud_emp person_4`}, {`Sampling: bernoulli ('100'::real)`}},
			},
			{
				Statement: `select count(*) from person tablesample bernoulli (100);`,
				Results:   []sql.Row{{58}},
			},
			{
				Statement: `select count(*) from person;`,
				Results:   []sql.Row{{58}},
			},
			{
				Statement: `SELECT count(*) FROM test_tablesample TABLESAMPLE bernoulli (('1'::text < '0'::text)::int);`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `select * from
  (values (0),(100)) v(pct),
  lateral (select count(*) from tenk1 tablesample bernoulli (pct)) ss;`,
				Results: []sql.Row{{0, 0}, {100, 10000}},
			},
			{
				Statement: `select * from
  (values (0),(100)) v(pct),
  lateral (select count(*) from tenk1 tablesample system (pct)) ss;`,
				Results: []sql.Row{{0, 0}, {100, 10000}},
			},
			{
				Statement: `explain (costs off)
select pct, count(unique1) from
  (values (0),(100)) v(pct),
  lateral (select * from tenk1 tablesample bernoulli (pct)) ss
  group by pct;`,
				Results: []sql.Row{{`HashAggregate`}, {`Group Key: "*VALUES*".column1`}, {`->  Nested Loop`}, {`->  Values Scan on "*VALUES*"`}, {`->  Sample Scan on tenk1`}, {`Sampling: bernoulli ("*VALUES*".column1)`}},
			},
			{
				Statement: `select pct, count(unique1) from
  (values (0),(100)) v(pct),
  lateral (select * from tenk1 tablesample bernoulli (pct)) ss
  group by pct;`,
				Results: []sql.Row{{100, 10000}},
			},
			{
				Statement: `select pct, count(unique1) from
  (values (0),(100)) v(pct),
  lateral (select * from tenk1 tablesample system (pct)) ss
  group by pct;`,
				Results: []sql.Row{{100, 10000}},
			},
			{
				Statement:   `SELECT id FROM test_tablesample TABLESAMPLE FOOBAR (1);`,
				ErrorString: `tablesample method foobar does not exist`,
			},
			{
				Statement:   `SELECT id FROM test_tablesample TABLESAMPLE SYSTEM (NULL);`,
				ErrorString: `TABLESAMPLE parameter cannot be null`,
			},
			{
				Statement:   `SELECT id FROM test_tablesample TABLESAMPLE SYSTEM (50) REPEATABLE (NULL);`,
				ErrorString: `TABLESAMPLE REPEATABLE parameter cannot be null`,
			},
			{
				Statement:   `SELECT id FROM test_tablesample TABLESAMPLE BERNOULLI (-1);`,
				ErrorString: `sample percentage must be between 0 and 100`,
			},
			{
				Statement:   `SELECT id FROM test_tablesample TABLESAMPLE BERNOULLI (200);`,
				ErrorString: `sample percentage must be between 0 and 100`,
			},
			{
				Statement:   `SELECT id FROM test_tablesample TABLESAMPLE SYSTEM (-1);`,
				ErrorString: `sample percentage must be between 0 and 100`,
			},
			{
				Statement:   `SELECT id FROM test_tablesample TABLESAMPLE SYSTEM (200);`,
				ErrorString: `sample percentage must be between 0 and 100`,
			},
			{
				Statement:   `SELECT id FROM test_tablesample_v1 TABLESAMPLE BERNOULLI (1);`,
				ErrorString: `TABLESAMPLE clause can only be applied to tables and materialized views`,
			},
			{
				Statement:   `INSERT INTO test_tablesample_v1 VALUES(1);`,
				ErrorString: `cannot insert into view "test_tablesample_v1"`,
			},
			{
				Statement: `WITH query_select AS (SELECT * FROM test_tablesample)
SELECT * FROM query_select TABLESAMPLE BERNOULLI (5.5) REPEATABLE (1);`,
				ErrorString: `TABLESAMPLE clause can only be applied to tables and materialized views`,
			},
			{
				Statement:   `SELECT q.* FROM (SELECT * FROM test_tablesample) as q TABLESAMPLE BERNOULLI (5);`,
				ErrorString: `syntax error at or near "TABLESAMPLE"`,
			},
			{
				Statement: `create table parted_sample (a int) partition by list (a);`,
			},
			{
				Statement: `create table parted_sample_1 partition of parted_sample for values in (1);`,
			},
			{
				Statement: `create table parted_sample_2 partition of parted_sample for values in (2);`,
			},
			{
				Statement: `explain (costs off)
  select * from parted_sample tablesample bernoulli (100);`,
				Results: []sql.Row{{`Append`}, {`->  Sample Scan on parted_sample_1`}, {`Sampling: bernoulli ('100'::real)`}, {`->  Sample Scan on parted_sample_2`}, {`Sampling: bernoulli ('100'::real)`}},
			},
			{
				Statement: `drop table parted_sample, parted_sample_1, parted_sample_2;`,
			},
		},
	})
}
