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

func TestTidscan(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_tidscan)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_tidscan,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TABLE tidscan(id integer);`,
			},
			{
				Statement: `INSERT INTO tidscan VALUES (1), (2), (3);`,
			},
			{
				Statement: `SELECT ctid, * FROM tidscan;`,
				Results:   []sql.Row{{`(0,1)`, 1}, {`(0,2)`, 2}, {`(0,3)`, 3}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT ctid, * FROM tidscan WHERE ctid = '(0,1)';`,
				Results: []sql.Row{{`Tid Scan on tidscan`}, {`TID Cond: (ctid = '(0,1)'::tid)`}},
			},
			{
				Statement: `SELECT ctid, * FROM tidscan WHERE ctid = '(0,1)';`,
				Results:   []sql.Row{{`(0,1)`, 1}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT ctid, * FROM tidscan WHERE '(0,1)' = ctid;`,
				Results: []sql.Row{{`Tid Scan on tidscan`}, {`TID Cond: ('(0,1)'::tid = ctid)`}},
			},
			{
				Statement: `SELECT ctid, * FROM tidscan WHERE '(0,1)' = ctid;`,
				Results:   []sql.Row{{`(0,1)`, 1}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT ctid, * FROM tidscan WHERE ctid = '(0,2)' OR '(0,1)' = ctid;`,
				Results: []sql.Row{{`Tid Scan on tidscan`}, {`TID Cond: ((ctid = '(0,2)'::tid) OR ('(0,1)'::tid = ctid))`}},
			},
			{
				Statement: `SELECT ctid, * FROM tidscan WHERE ctid = '(0,2)' OR '(0,1)' = ctid;`,
				Results:   []sql.Row{{`(0,1)`, 1}, {`(0,2)`, 2}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT ctid, * FROM tidscan WHERE ctid = ANY(ARRAY['(0,1)', '(0,2)']::tid[]);`,
				Results: []sql.Row{{`Tid Scan on tidscan`}, {`TID Cond: (ctid = ANY ('{"(0,1)","(0,2)"}'::tid[]))`}},
			},
			{
				Statement: `SELECT ctid, * FROM tidscan WHERE ctid = ANY(ARRAY['(0,1)', '(0,2)']::tid[]);`,
				Results:   []sql.Row{{`(0,1)`, 1}, {`(0,2)`, 2}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT ctid, * FROM tidscan WHERE ctid != ANY(ARRAY['(0,1)', '(0,2)']::tid[]);`,
				Results: []sql.Row{{`Seq Scan on tidscan`}, {`Filter: (ctid <> ANY ('{"(0,1)","(0,2)"}'::tid[]))`}},
			},
			{
				Statement: `SELECT ctid, * FROM tidscan WHERE ctid != ANY(ARRAY['(0,1)', '(0,2)']::tid[]);`,
				Results:   []sql.Row{{`(0,1)`, 1}, {`(0,2)`, 2}, {`(0,3)`, 3}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT ctid, * FROM tidscan
WHERE (id = 3 AND ctid IN ('(0,2)', '(0,3)')) OR (ctid = '(0,1)' AND id = 1);`,
				Results: []sql.Row{{`Tid Scan on tidscan`}, {`TID Cond: ((ctid = ANY ('{"(0,2)","(0,3)"}'::tid[])) OR (ctid = '(0,1)'::tid))`}, {`Filter: (((id = 3) AND (ctid = ANY ('{"(0,2)","(0,3)"}'::tid[]))) OR ((ctid = '(0,1)'::tid) AND (id = 1)))`}},
			},
			{
				Statement: `SELECT ctid, * FROM tidscan
WHERE (id = 3 AND ctid IN ('(0,2)', '(0,3)')) OR (ctid = '(0,1)' AND id = 1);`,
				Results: []sql.Row{{`(0,1)`, 1}, {`(0,3)`, 3}},
			},
			{
				Statement: `SET enable_hashjoin TO off;  -- otherwise hash join might win`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.ctid, t1.*, t2.ctid, t2.*
FROM tidscan t1 JOIN tidscan t2 ON t1.ctid = t2.ctid WHERE t1.id = 1;`,
				Results: []sql.Row{{`Nested Loop`}, {`->  Seq Scan on tidscan t1`}, {`Filter: (id = 1)`}, {`->  Tid Scan on tidscan t2`}, {`TID Cond: (ctid = t1.ctid)`}},
			},
			{
				Statement: `SELECT t1.ctid, t1.*, t2.ctid, t2.*
FROM tidscan t1 JOIN tidscan t2 ON t1.ctid = t2.ctid WHERE t1.id = 1;`,
				Results: []sql.Row{{`(0,1)`, 1, `(0,1)`, 1}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.ctid, t1.*, t2.ctid, t2.*
FROM tidscan t1 LEFT JOIN tidscan t2 ON t1.ctid = t2.ctid WHERE t1.id = 1;`,
				Results: []sql.Row{{`Nested Loop Left Join`}, {`->  Seq Scan on tidscan t1`}, {`Filter: (id = 1)`}, {`->  Tid Scan on tidscan t2`}, {`TID Cond: (t1.ctid = ctid)`}},
			},
			{
				Statement: `SELECT t1.ctid, t1.*, t2.ctid, t2.*
FROM tidscan t1 LEFT JOIN tidscan t2 ON t1.ctid = t2.ctid WHERE t1.id = 1;`,
				Results: []sql.Row{{`(0,1)`, 1, `(0,1)`, 1}},
			},
			{
				Statement: `RESET enable_hashjoin;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `DECLARE c CURSOR FOR
SELECT ctid, * FROM tidscan WHERE ctid = ANY(ARRAY['(0,1)', '(0,2)']::tid[]);`,
			},
			{
				Statement: `FETCH ALL FROM c;`,
				Results:   []sql.Row{{`(0,1)`, 1}, {`(0,2)`, 2}},
			},
			{
				Statement: `FETCH BACKWARD 1 FROM c;`,
				Results:   []sql.Row{{`(0,2)`, 2}},
			},
			{
				Statement: `FETCH FIRST FROM c;`,
				Results:   []sql.Row{{`(0,1)`, 1}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `DECLARE c CURSOR FOR SELECT ctid, * FROM tidscan;`,
			},
			{
				Statement: `FETCH NEXT FROM c; -- skip one row`,
				Results:   []sql.Row{{`(0,1)`, 1}},
			},
			{
				Statement: `FETCH NEXT FROM c;`,
				Results:   []sql.Row{{`(0,2)`, 2}},
			},
			{
				Statement: `EXPLAIN (ANALYZE, COSTS OFF, SUMMARY OFF, TIMING OFF)
UPDATE tidscan SET id = -id WHERE CURRENT OF c RETURNING *;`,
				Results: []sql.Row{{`Update on tidscan (actual rows=1 loops=1)`}, {`->  Tid Scan on tidscan (actual rows=1 loops=1)`}, {`TID Cond: CURRENT OF c`}},
			},
			{
				Statement: `FETCH NEXT FROM c;`,
				Results:   []sql.Row{{`(0,3)`, 3}},
			},
			{
				Statement: `EXPLAIN (ANALYZE, COSTS OFF, SUMMARY OFF, TIMING OFF)
UPDATE tidscan SET id = -id WHERE CURRENT OF c RETURNING *;`,
				Results: []sql.Row{{`Update on tidscan (actual rows=1 loops=1)`}, {`->  Tid Scan on tidscan (actual rows=1 loops=1)`}, {`TID Cond: CURRENT OF c`}},
			},
			{
				Statement: `SELECT * FROM tidscan;`,
				Results:   []sql.Row{{1}, {-2}, {-3}},
			},
			{
				Statement: `FETCH NEXT FROM c;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `EXPLAIN (ANALYZE, COSTS OFF, SUMMARY OFF, TIMING OFF)
UPDATE tidscan SET id = -id WHERE CURRENT OF c RETURNING *;`,
				ErrorString: `cursor "c" is not positioned on a row`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM tenk1 t1 JOIN tenk1 t2 ON t1.ctid = t2.ctid;`,
				Results: []sql.Row{{`Aggregate`}, {`->  Hash Join`}, {`Hash Cond: (t1.ctid = t2.ctid)`}, {`->  Seq Scan on tenk1 t1`}, {`->  Hash`}, {`->  Seq Scan on tenk1 t2`}},
			},
			{
				Statement: `SELECT count(*) FROM tenk1 t1 JOIN tenk1 t2 ON t1.ctid = t2.ctid;`,
				Results:   []sql.Row{{10000}},
			},
			{
				Statement: `SET enable_hashjoin TO off;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM tenk1 t1 JOIN tenk1 t2 ON t1.ctid = t2.ctid;`,
				Results: []sql.Row{{`Aggregate`}, {`->  Merge Join`}, {`Merge Cond: (t1.ctid = t2.ctid)`}, {`->  Sort`}, {`Sort Key: t1.ctid`}, {`->  Seq Scan on tenk1 t1`}, {`->  Sort`}, {`Sort Key: t2.ctid`}, {`->  Seq Scan on tenk1 t2`}},
			},
			{
				Statement: `SELECT count(*) FROM tenk1 t1 JOIN tenk1 t2 ON t1.ctid = t2.ctid;`,
				Results:   []sql.Row{{10000}},
			},
			{
				Statement: `RESET enable_hashjoin;`,
			},
			{
				Statement: `BEGIN ISOLATION LEVEL SERIALIZABLE;`,
			},
			{
				Statement: `SELECT * FROM tidscan WHERE ctid = '(0,1)';`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT locktype, mode FROM pg_locks WHERE pid = pg_backend_pid() AND mode = 'SIReadLock';`,
				Results:   []sql.Row{{`tuple`, `SIReadLock`}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `DROP TABLE tidscan;`,
			},
		},
	})
}
