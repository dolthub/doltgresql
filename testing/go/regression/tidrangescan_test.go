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

func TestTidrangescan(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_tidrangescan)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_tidrangescan,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `SET enable_seqscan TO off;`,
			},
			{
				Statement: `CREATE TABLE tidrangescan(id integer, data text);`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT ctid FROM tidrangescan WHERE ctid < '(1, 0)';`,
				Results: []sql.Row{{`Tid Range Scan on tidrangescan`}, {`TID Cond: (ctid < '(1,0)'::tid)`}},
			},
			{
				Statement: `SELECT ctid FROM tidrangescan WHERE ctid < '(1, 0)';`,
				Results:   []sql.Row{},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT ctid FROM tidrangescan WHERE ctid > '(9, 0)';`,
				Results: []sql.Row{{`Tid Range Scan on tidrangescan`}, {`TID Cond: (ctid > '(9,0)'::tid)`}},
			},
			{
				Statement: `SELECT ctid FROM tidrangescan WHERE ctid > '(9, 0)';`,
				Results:   []sql.Row{},
			},
			{
				Statement: `INSERT INTO tidrangescan SELECT i,repeat('x', 100) FROM generate_series(1,200) AS s(i);`,
			},
			{
				Statement: `DELETE FROM tidrangescan
WHERE substring(ctid::text FROM ',(\d+)\)')::integer > 10 OR substring(ctid::text FROM '\((\d+),')::integer > 2;`,
			},
			{
				Statement: `VACUUM tidrangescan;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT ctid FROM tidrangescan WHERE ctid < '(1,0)';`,
				Results: []sql.Row{{`Tid Range Scan on tidrangescan`}, {`TID Cond: (ctid < '(1,0)'::tid)`}},
			},
			{
				Statement: `SELECT ctid FROM tidrangescan WHERE ctid < '(1,0)';`,
				Results:   []sql.Row{{`(0,1)`}, {`(0,2)`}, {`(0,3)`}, {`(0,4)`}, {`(0,5)`}, {`(0,6)`}, {`(0,7)`}, {`(0,8)`}, {`(0,9)`}, {`(0,10)`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT ctid FROM tidrangescan WHERE ctid <= '(1,5)';`,
				Results: []sql.Row{{`Tid Range Scan on tidrangescan`}, {`TID Cond: (ctid <= '(1,5)'::tid)`}},
			},
			{
				Statement: `SELECT ctid FROM tidrangescan WHERE ctid <= '(1,5)';`,
				Results:   []sql.Row{{`(0,1)`}, {`(0,2)`}, {`(0,3)`}, {`(0,4)`}, {`(0,5)`}, {`(0,6)`}, {`(0,7)`}, {`(0,8)`}, {`(0,9)`}, {`(0,10)`}, {`(1,1)`}, {`(1,2)`}, {`(1,3)`}, {`(1,4)`}, {`(1,5)`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT ctid FROM tidrangescan WHERE ctid < '(0,0)';`,
				Results: []sql.Row{{`Tid Range Scan on tidrangescan`}, {`TID Cond: (ctid < '(0,0)'::tid)`}},
			},
			{
				Statement: `SELECT ctid FROM tidrangescan WHERE ctid < '(0,0)';`,
				Results:   []sql.Row{},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT ctid FROM tidrangescan WHERE ctid > '(2,8)';`,
				Results: []sql.Row{{`Tid Range Scan on tidrangescan`}, {`TID Cond: (ctid > '(2,8)'::tid)`}},
			},
			{
				Statement: `SELECT ctid FROM tidrangescan WHERE ctid > '(2,8)';`,
				Results:   []sql.Row{{`(2,9)`}, {`(2,10)`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT ctid FROM tidrangescan WHERE '(2,8)' < ctid;`,
				Results: []sql.Row{{`Tid Range Scan on tidrangescan`}, {`TID Cond: ('(2,8)'::tid < ctid)`}},
			},
			{
				Statement: `SELECT ctid FROM tidrangescan WHERE '(2,8)' < ctid;`,
				Results:   []sql.Row{{`(2,9)`}, {`(2,10)`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT ctid FROM tidrangescan WHERE ctid >= '(2,8)';`,
				Results: []sql.Row{{`Tid Range Scan on tidrangescan`}, {`TID Cond: (ctid >= '(2,8)'::tid)`}},
			},
			{
				Statement: `SELECT ctid FROM tidrangescan WHERE ctid >= '(2,8)';`,
				Results:   []sql.Row{{`(2,8)`}, {`(2,9)`}, {`(2,10)`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT ctid FROM tidrangescan WHERE ctid >= '(100,0)';`,
				Results: []sql.Row{{`Tid Range Scan on tidrangescan`}, {`TID Cond: (ctid >= '(100,0)'::tid)`}},
			},
			{
				Statement: `SELECT ctid FROM tidrangescan WHERE ctid >= '(100,0)';`,
				Results:   []sql.Row{},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT ctid FROM tidrangescan WHERE ctid > '(1,4)' AND '(1,7)' >= ctid;`,
				Results: []sql.Row{{`Tid Range Scan on tidrangescan`}, {`TID Cond: ((ctid > '(1,4)'::tid) AND ('(1,7)'::tid >= ctid))`}},
			},
			{
				Statement: `SELECT ctid FROM tidrangescan WHERE ctid > '(1,4)' AND '(1,7)' >= ctid;`,
				Results:   []sql.Row{{`(1,5)`}, {`(1,6)`}, {`(1,7)`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT ctid FROM tidrangescan WHERE '(1,7)' >= ctid AND ctid > '(1,4)';`,
				Results: []sql.Row{{`Tid Range Scan on tidrangescan`}, {`TID Cond: (('(1,7)'::tid >= ctid) AND (ctid > '(1,4)'::tid))`}},
			},
			{
				Statement: `SELECT ctid FROM tidrangescan WHERE '(1,7)' >= ctid AND ctid > '(1,4)';`,
				Results:   []sql.Row{{`(1,5)`}, {`(1,6)`}, {`(1,7)`}},
			},
			{
				Statement: `SELECT ctid FROM tidrangescan WHERE ctid > '(0,65535)' AND ctid < '(1,0)' LIMIT 1;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT ctid FROM tidrangescan WHERE ctid < '(0,0)' LIMIT 1;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT ctid FROM tidrangescan WHERE ctid > '(4294967295,65535)';`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT ctid FROM tidrangescan WHERE ctid < '(0,0)';`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT ctid FROM tidrangescan WHERE ctid >= (SELECT NULL::tid);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t.ctid,t2.c FROM tidrangescan t,
LATERAL (SELECT count(*) c FROM tidrangescan t2 WHERE t2.ctid <= t.ctid) t2
WHERE t.ctid < '(1,0)';`,
				Results: []sql.Row{{`Nested Loop`}, {`->  Tid Range Scan on tidrangescan t`}, {`TID Cond: (ctid < '(1,0)'::tid)`}, {`->  Aggregate`}, {`->  Tid Range Scan on tidrangescan t2`}, {`TID Cond: (ctid <= t.ctid)`}},
			},
			{
				Statement: `SELECT t.ctid,t2.c FROM tidrangescan t,
LATERAL (SELECT count(*) c FROM tidrangescan t2 WHERE t2.ctid <= t.ctid) t2
WHERE t.ctid < '(1,0)';`,
				Results: []sql.Row{{`(0,1)`, 1}, {`(0,2)`, 2}, {`(0,3)`, 3}, {`(0,4)`, 4}, {`(0,5)`, 5}, {`(0,6)`, 6}, {`(0,7)`, 7}, {`(0,8)`, 8}, {`(0,9)`, 9}, {`(0,10)`, 10}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
DECLARE c SCROLL CURSOR FOR SELECT ctid FROM tidrangescan WHERE ctid < '(1,0)';`,
				Results: []sql.Row{{`Tid Range Scan on tidrangescan`}, {`TID Cond: (ctid < '(1,0)'::tid)`}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `DECLARE c SCROLL CURSOR FOR SELECT ctid FROM tidrangescan WHERE ctid < '(1,0)';`,
			},
			{
				Statement: `FETCH NEXT c;`,
				Results:   []sql.Row{{`(0,1)`}},
			},
			{
				Statement: `FETCH NEXT c;`,
				Results:   []sql.Row{{`(0,2)`}},
			},
			{
				Statement: `FETCH PRIOR c;`,
				Results:   []sql.Row{{`(0,1)`}},
			},
			{
				Statement: `FETCH FIRST c;`,
				Results:   []sql.Row{{`(0,1)`}},
			},
			{
				Statement: `FETCH LAST c;`,
				Results:   []sql.Row{{`(0,10)`}},
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `DROP TABLE tidrangescan;`,
			},
			{
				Statement: `RESET enable_seqscan;`,
			},
		},
	})
}
