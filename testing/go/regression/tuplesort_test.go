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

func TestTuplesort(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_tuplesort)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_tuplesort,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `SET max_parallel_maintenance_workers = 0;`,
			},
			{
				Statement: `SET max_parallel_workers = 0;`,
			},
			{
				Statement: `CREATE TEMP TABLE abbrev_abort_uuids (
    id serial not null,
    abort_increasing uuid,
    abort_decreasing uuid,
    noabort_increasing uuid,
    noabort_decreasing uuid);`,
			},
			{
				Statement: `INSERT INTO abbrev_abort_uuids (abort_increasing, abort_decreasing, noabort_increasing, noabort_decreasing)
    SELECT
        ('00000000-0000-0000-0000-'||to_char(g.i, '000000000000FM'))::uuid abort_increasing,
        ('00000000-0000-0000-0000-'||to_char(20000 - g.i, '000000000000FM'))::uuid abort_decreasing,
        (to_char(g.i % 10009, '00000000FM')||'-0000-0000-0000-'||to_char(g.i, '000000000000FM'))::uuid noabort_increasing,
        (to_char(((20000 - g.i) % 10009), '00000000FM')||'-0000-0000-0000-'||to_char(20000 - g.i, '000000000000FM'))::uuid noabort_decreasing
    FROM generate_series(0, 20000, 1) g(i);`,
			},
			{
				Statement: `INSERT INTO abbrev_abort_uuids(id) VALUES(0);`,
			},
			{
				Statement: `INSERT INTO abbrev_abort_uuids DEFAULT VALUES;`,
			},
			{
				Statement: `INSERT INTO abbrev_abort_uuids DEFAULT VALUES;`,
			},
			{
				Statement: `INSERT INTO abbrev_abort_uuids (abort_increasing, abort_decreasing, noabort_increasing, noabort_decreasing)
    SELECT abort_increasing, abort_decreasing, noabort_increasing, noabort_decreasing
    FROM abbrev_abort_uuids
    WHERE (id < 10 OR id > 19990) AND id % 3 = 0 AND abort_increasing is not null;`,
			},
			{
				Statement: `----
----
SELECT abort_increasing, abort_decreasing FROM abbrev_abort_uuids ORDER BY abort_increasing OFFSET 20000 - 4;`,
				Results: []sql.Row{{`00000000-0000-0000-0000-000000019992`, `00000000-0000-0000-0000-000000000008`}, {`00000000-0000-0000-0000-000000019993`, `00000000-0000-0000-0000-000000000007`}, {`00000000-0000-0000-0000-000000019994`, `00000000-0000-0000-0000-000000000006`}, {`00000000-0000-0000-0000-000000019994`, `00000000-0000-0000-0000-000000000006`}, {`00000000-0000-0000-0000-000000019995`, `00000000-0000-0000-0000-000000000005`}, {`00000000-0000-0000-0000-000000019996`, `00000000-0000-0000-0000-000000000004`}, {`00000000-0000-0000-0000-000000019997`, `00000000-0000-0000-0000-000000000003`}, {`00000000-0000-0000-0000-000000019997`, `00000000-0000-0000-0000-000000000003`}, {`00000000-0000-0000-0000-000000019998`, `00000000-0000-0000-0000-000000000002`}, {`00000000-0000-0000-0000-000000019999`, `00000000-0000-0000-0000-000000000001`}, {`00000000-0000-0000-0000-000000020000`, `00000000-0000-0000-0000-000000000000`}, {`00000000-0000-0000-0000-000000020000`, `00000000-0000-0000-0000-000000000000`}, {``, ``}, {``, ``}, {``, ``}},
			},
			{
				Statement: `SELECT abort_increasing, abort_decreasing FROM abbrev_abort_uuids ORDER BY abort_decreasing NULLS FIRST OFFSET 20000 - 4;`,
				Results:   []sql.Row{{`00000000-0000-0000-0000-000000000011`, `00000000-0000-0000-0000-000000019989`}, {`00000000-0000-0000-0000-000000000010`, `00000000-0000-0000-0000-000000019990`}, {`00000000-0000-0000-0000-000000000009`, `00000000-0000-0000-0000-000000019991`}, {`00000000-0000-0000-0000-000000000008`, `00000000-0000-0000-0000-000000019992`}, {`00000000-0000-0000-0000-000000000008`, `00000000-0000-0000-0000-000000019992`}, {`00000000-0000-0000-0000-000000000007`, `00000000-0000-0000-0000-000000019993`}, {`00000000-0000-0000-0000-000000000006`, `00000000-0000-0000-0000-000000019994`}, {`00000000-0000-0000-0000-000000000005`, `00000000-0000-0000-0000-000000019995`}, {`00000000-0000-0000-0000-000000000005`, `00000000-0000-0000-0000-000000019995`}, {`00000000-0000-0000-0000-000000000004`, `00000000-0000-0000-0000-000000019996`}, {`00000000-0000-0000-0000-000000000003`, `00000000-0000-0000-0000-000000019997`}, {`00000000-0000-0000-0000-000000000002`, `00000000-0000-0000-0000-000000019998`}, {`00000000-0000-0000-0000-000000000002`, `00000000-0000-0000-0000-000000019998`}, {`00000000-0000-0000-0000-000000000001`, `00000000-0000-0000-0000-000000019999`}, {`00000000-0000-0000-0000-000000000000`, `00000000-0000-0000-0000-000000020000`}},
			},
			{
				Statement: `SELECT noabort_increasing, noabort_decreasing FROM abbrev_abort_uuids ORDER BY noabort_increasing OFFSET 20000 - 4;`,
				Results:   []sql.Row{{`00009997-0000-0000-0000-000000009997`, `00010003-0000-0000-0000-000000010003`}, {`00009998-0000-0000-0000-000000009998`, `00010002-0000-0000-0000-000000010002`}, {`00009999-0000-0000-0000-000000009999`, `00010001-0000-0000-0000-000000010001`}, {`00010000-0000-0000-0000-000000010000`, `00010000-0000-0000-0000-000000010000`}, {`00010001-0000-0000-0000-000000010001`, `00009999-0000-0000-0000-000000009999`}, {`00010002-0000-0000-0000-000000010002`, `00009998-0000-0000-0000-000000009998`}, {`00010003-0000-0000-0000-000000010003`, `00009997-0000-0000-0000-000000009997`}, {`00010004-0000-0000-0000-000000010004`, `00009996-0000-0000-0000-000000009996`}, {`00010005-0000-0000-0000-000000010005`, `00009995-0000-0000-0000-000000009995`}, {`00010006-0000-0000-0000-000000010006`, `00009994-0000-0000-0000-000000009994`}, {`00010007-0000-0000-0000-000000010007`, `00009993-0000-0000-0000-000000009993`}, {`00010008-0000-0000-0000-000000010008`, `00009992-0000-0000-0000-000000009992`}, {``, ``}, {``, ``}, {``, ``}},
			},
			{
				Statement: `SELECT noabort_increasing, noabort_decreasing FROM abbrev_abort_uuids ORDER BY noabort_decreasing NULLS FIRST OFFSET 20000 - 4;`,
				Results:   []sql.Row{{`00010006-0000-0000-0000-000000010006`, `00009994-0000-0000-0000-000000009994`}, {`00010005-0000-0000-0000-000000010005`, `00009995-0000-0000-0000-000000009995`}, {`00010004-0000-0000-0000-000000010004`, `00009996-0000-0000-0000-000000009996`}, {`00010003-0000-0000-0000-000000010003`, `00009997-0000-0000-0000-000000009997`}, {`00010002-0000-0000-0000-000000010002`, `00009998-0000-0000-0000-000000009998`}, {`00010001-0000-0000-0000-000000010001`, `00009999-0000-0000-0000-000000009999`}, {`00010000-0000-0000-0000-000000010000`, `00010000-0000-0000-0000-000000010000`}, {`00009999-0000-0000-0000-000000009999`, `00010001-0000-0000-0000-000000010001`}, {`00009998-0000-0000-0000-000000009998`, `00010002-0000-0000-0000-000000010002`}, {`00009997-0000-0000-0000-000000009997`, `00010003-0000-0000-0000-000000010003`}, {`00009996-0000-0000-0000-000000009996`, `00010004-0000-0000-0000-000000010004`}, {`00009995-0000-0000-0000-000000009995`, `00010005-0000-0000-0000-000000010005`}, {`00009994-0000-0000-0000-000000009994`, `00010006-0000-0000-0000-000000010006`}, {`00009993-0000-0000-0000-000000009993`, `00010007-0000-0000-0000-000000010007`}, {`00009992-0000-0000-0000-000000009992`, `00010008-0000-0000-0000-000000010008`}},
			},
			{
				Statement: `SELECT abort_increasing, noabort_increasing FROM abbrev_abort_uuids ORDER BY abort_increasing LIMIT 5;`,
				Results:   []sql.Row{{`00000000-0000-0000-0000-000000000000`, `00000000-0000-0000-0000-000000000000`}, {`00000000-0000-0000-0000-000000000001`, `00000001-0000-0000-0000-000000000001`}, {`00000000-0000-0000-0000-000000000002`, `00000002-0000-0000-0000-000000000002`}, {`00000000-0000-0000-0000-000000000002`, `00000002-0000-0000-0000-000000000002`}, {`00000000-0000-0000-0000-000000000003`, `00000003-0000-0000-0000-000000000003`}},
			},
			{
				Statement: `SELECT abort_increasing, noabort_increasing FROM abbrev_abort_uuids ORDER BY noabort_increasing NULLS FIRST LIMIT 5;`,
				Results:   []sql.Row{{``, ``}, {``, ``}, {``, ``}, {`00000000-0000-0000-0000-000000000000`, `00000000-0000-0000-0000-000000000000`}, {`00000000-0000-0000-0000-000000010009`, `00000000-0000-0000-0000-000000010009`}},
			},
			{
				Statement: `----
----
CREATE INDEX abbrev_abort_uuids__noabort_increasing_idx ON abbrev_abort_uuids (noabort_increasing);`,
			},
			{
				Statement: `CREATE INDEX abbrev_abort_uuids__noabort_decreasing_idx ON abbrev_abort_uuids (noabort_decreasing);`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT id, noabort_increasing, noabort_decreasing FROM abbrev_abort_uuids ORDER BY noabort_increasing LIMIT 5;`,
				Results: []sql.Row{{`Limit`}, {`->  Index Scan using abbrev_abort_uuids__noabort_increasing_idx on abbrev_abort_uuids`}},
			},
			{
				Statement: `SELECT id, noabort_increasing, noabort_decreasing FROM abbrev_abort_uuids ORDER BY noabort_increasing LIMIT 5;`,
				Results:   []sql.Row{{1, `00000000-0000-0000-0000-000000000000`, `00009991-0000-0000-0000-000000020000`}, {10010, `00000000-0000-0000-0000-000000010009`, `00009991-0000-0000-0000-000000009991`}, {2, `00000001-0000-0000-0000-000000000001`, `00009990-0000-0000-0000-000000019999`}, {10011, `00000001-0000-0000-0000-000000010010`, `00009990-0000-0000-0000-000000009990`}, {3, `00000002-0000-0000-0000-000000000002`, `00009989-0000-0000-0000-000000019998`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT id, noabort_increasing, noabort_decreasing FROM abbrev_abort_uuids ORDER BY noabort_decreasing LIMIT 5;`,
				Results: []sql.Row{{`Limit`}, {`->  Index Scan using abbrev_abort_uuids__noabort_decreasing_idx on abbrev_abort_uuids`}},
			},
			{
				Statement: `SELECT id, noabort_increasing, noabort_decreasing FROM abbrev_abort_uuids ORDER BY noabort_decreasing LIMIT 5;`,
				Results:   []sql.Row{{20001, `00009991-0000-0000-0000-000000020000`, `00000000-0000-0000-0000-000000000000`}, {20010, `00009991-0000-0000-0000-000000020000`, `00000000-0000-0000-0000-000000000000`}, {9992, `00009991-0000-0000-0000-000000009991`, `00000000-0000-0000-0000-000000010009`}, {20000, `00009990-0000-0000-0000-000000019999`, `00000001-0000-0000-0000-000000000001`}, {9991, `00009990-0000-0000-0000-000000009990`, `00000001-0000-0000-0000-000000010010`}},
			},
			{
				Statement: `CREATE INDEX abbrev_abort_uuids__abort_increasing_idx ON abbrev_abort_uuids (abort_increasing);`,
			},
			{
				Statement: `CREATE INDEX abbrev_abort_uuids__abort_decreasing_idx ON abbrev_abort_uuids (abort_decreasing);`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT id, abort_increasing, abort_decreasing FROM abbrev_abort_uuids ORDER BY abort_increasing LIMIT 5;`,
				Results: []sql.Row{{`Limit`}, {`->  Index Scan using abbrev_abort_uuids__abort_increasing_idx on abbrev_abort_uuids`}},
			},
			{
				Statement: `SELECT id, abort_increasing, abort_decreasing FROM abbrev_abort_uuids ORDER BY abort_increasing LIMIT 5;`,
				Results:   []sql.Row{{1, `00000000-0000-0000-0000-000000000000`, `00000000-0000-0000-0000-000000020000`}, {2, `00000000-0000-0000-0000-000000000001`, `00000000-0000-0000-0000-000000019999`}, {3, `00000000-0000-0000-0000-000000000002`, `00000000-0000-0000-0000-000000019998`}, {20004, `00000000-0000-0000-0000-000000000002`, `00000000-0000-0000-0000-000000019998`}, {4, `00000000-0000-0000-0000-000000000003`, `00000000-0000-0000-0000-000000019997`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT id, abort_increasing, abort_decreasing FROM abbrev_abort_uuids ORDER BY abort_decreasing LIMIT 5;`,
				Results: []sql.Row{{`Limit`}, {`->  Index Scan using abbrev_abort_uuids__abort_decreasing_idx on abbrev_abort_uuids`}},
			},
			{
				Statement: `SELECT id, abort_increasing, abort_decreasing FROM abbrev_abort_uuids ORDER BY abort_decreasing LIMIT 5;`,
				Results:   []sql.Row{{20001, `00000000-0000-0000-0000-000000020000`, `00000000-0000-0000-0000-000000000000`}, {20010, `00000000-0000-0000-0000-000000020000`, `00000000-0000-0000-0000-000000000000`}, {20000, `00000000-0000-0000-0000-000000019999`, `00000000-0000-0000-0000-000000000001`}, {19999, `00000000-0000-0000-0000-000000019998`, `00000000-0000-0000-0000-000000000002`}, {19998, `00000000-0000-0000-0000-000000019997`, `00000000-0000-0000-0000-000000000003`}},
			},
			{
				Statement: `----
----
BEGIN;`,
			},
			{
				Statement: `SET LOCAL enable_indexscan = false;`,
			},
			{
				Statement: `CLUSTER abbrev_abort_uuids USING abbrev_abort_uuids__abort_increasing_idx;`,
			},
			{
				Statement: `SELECT id, abort_increasing, abort_decreasing, noabort_increasing, noabort_decreasing
FROM abbrev_abort_uuids
ORDER BY ctid LIMIT 5;`,
				Results: []sql.Row{{1, `00000000-0000-0000-0000-000000000000`, `00000000-0000-0000-0000-000000020000`, `00000000-0000-0000-0000-000000000000`, `00009991-0000-0000-0000-000000020000`}, {2, `00000000-0000-0000-0000-000000000001`, `00000000-0000-0000-0000-000000019999`, `00000001-0000-0000-0000-000000000001`, `00009990-0000-0000-0000-000000019999`}, {3, `00000000-0000-0000-0000-000000000002`, `00000000-0000-0000-0000-000000019998`, `00000002-0000-0000-0000-000000000002`, `00009989-0000-0000-0000-000000019998`}, {20004, `00000000-0000-0000-0000-000000000002`, `00000000-0000-0000-0000-000000019998`, `00000002-0000-0000-0000-000000000002`, `00009989-0000-0000-0000-000000019998`}, {4, `00000000-0000-0000-0000-000000000003`, `00000000-0000-0000-0000-000000019997`, `00000003-0000-0000-0000-000000000003`, `00009988-0000-0000-0000-000000019997`}},
			},
			{
				Statement: `SELECT id, abort_increasing, abort_decreasing, noabort_increasing, noabort_decreasing
FROM abbrev_abort_uuids
ORDER BY ctid DESC LIMIT 5;`,
				Results: []sql.Row{{0, ``, ``, ``, ``}, {20002, ``, ``, ``, ``}, {20003, ``, ``, ``, ``}, {20001, `00000000-0000-0000-0000-000000020000`, `00000000-0000-0000-0000-000000000000`, `00009991-0000-0000-0000-000000020000`, `00000000-0000-0000-0000-000000000000`}, {20010, `00000000-0000-0000-0000-000000020000`, `00000000-0000-0000-0000-000000000000`, `00009991-0000-0000-0000-000000020000`, `00000000-0000-0000-0000-000000000000`}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SET LOCAL enable_indexscan = false;`,
			},
			{
				Statement: `CLUSTER abbrev_abort_uuids USING abbrev_abort_uuids__abort_decreasing_idx;`,
			},
			{
				Statement: `SELECT id, abort_increasing, abort_decreasing, noabort_increasing, noabort_decreasing
FROM abbrev_abort_uuids
ORDER BY ctid LIMIT 5;`,
				Results: []sql.Row{{20010, `00000000-0000-0000-0000-000000020000`, `00000000-0000-0000-0000-000000000000`, `00009991-0000-0000-0000-000000020000`, `00000000-0000-0000-0000-000000000000`}, {20001, `00000000-0000-0000-0000-000000020000`, `00000000-0000-0000-0000-000000000000`, `00009991-0000-0000-0000-000000020000`, `00000000-0000-0000-0000-000000000000`}, {20000, `00000000-0000-0000-0000-000000019999`, `00000000-0000-0000-0000-000000000001`, `00009990-0000-0000-0000-000000019999`, `00000001-0000-0000-0000-000000000001`}, {19999, `00000000-0000-0000-0000-000000019998`, `00000000-0000-0000-0000-000000000002`, `00009989-0000-0000-0000-000000019998`, `00000002-0000-0000-0000-000000000002`}, {20009, `00000000-0000-0000-0000-000000019997`, `00000000-0000-0000-0000-000000000003`, `00009988-0000-0000-0000-000000019997`, `00000003-0000-0000-0000-000000000003`}},
			},
			{
				Statement: `SELECT id, abort_increasing, abort_decreasing, noabort_increasing, noabort_decreasing
FROM abbrev_abort_uuids
ORDER BY ctid DESC LIMIT 5;`,
				Results: []sql.Row{{0, ``, ``, ``, ``}, {20002, ``, ``, ``, ``}, {20003, ``, ``, ``, ``}, {1, `00000000-0000-0000-0000-000000000000`, `00000000-0000-0000-0000-000000020000`, `00000000-0000-0000-0000-000000000000`, `00009991-0000-0000-0000-000000020000`}, {2, `00000000-0000-0000-0000-000000000001`, `00000000-0000-0000-0000-000000019999`, `00000001-0000-0000-0000-000000000001`, `00009990-0000-0000-0000-000000019999`}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SET LOCAL enable_indexscan = false;`,
			},
			{
				Statement: `CLUSTER abbrev_abort_uuids USING abbrev_abort_uuids__noabort_increasing_idx;`,
			},
			{
				Statement: `SELECT id, abort_increasing, abort_decreasing, noabort_increasing, noabort_decreasing
FROM abbrev_abort_uuids
ORDER BY ctid LIMIT 5;`,
				Results: []sql.Row{{1, `00000000-0000-0000-0000-000000000000`, `00000000-0000-0000-0000-000000020000`, `00000000-0000-0000-0000-000000000000`, `00009991-0000-0000-0000-000000020000`}, {10010, `00000000-0000-0000-0000-000000010009`, `00000000-0000-0000-0000-000000009991`, `00000000-0000-0000-0000-000000010009`, `00009991-0000-0000-0000-000000009991`}, {2, `00000000-0000-0000-0000-000000000001`, `00000000-0000-0000-0000-000000019999`, `00000001-0000-0000-0000-000000000001`, `00009990-0000-0000-0000-000000019999`}, {10011, `00000000-0000-0000-0000-000000010010`, `00000000-0000-0000-0000-000000009990`, `00000001-0000-0000-0000-000000010010`, `00009990-0000-0000-0000-000000009990`}, {20004, `00000000-0000-0000-0000-000000000002`, `00000000-0000-0000-0000-000000019998`, `00000002-0000-0000-0000-000000000002`, `00009989-0000-0000-0000-000000019998`}},
			},
			{
				Statement: `SELECT id, abort_increasing, abort_decreasing, noabort_increasing, noabort_decreasing
FROM abbrev_abort_uuids
ORDER BY ctid DESC LIMIT 5;`,
				Results: []sql.Row{{0, ``, ``, ``, ``}, {20002, ``, ``, ``, ``}, {20003, ``, ``, ``, ``}, {10009, `00000000-0000-0000-0000-000000010008`, `00000000-0000-0000-0000-000000009992`, `00010008-0000-0000-0000-000000010008`, `00009992-0000-0000-0000-000000009992`}, {10008, `00000000-0000-0000-0000-000000010007`, `00000000-0000-0000-0000-000000009993`, `00010007-0000-0000-0000-000000010007`, `00009993-0000-0000-0000-000000009993`}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SET LOCAL enable_indexscan = false;`,
			},
			{
				Statement: `CLUSTER abbrev_abort_uuids USING abbrev_abort_uuids__noabort_decreasing_idx;`,
			},
			{
				Statement: `SELECT id, abort_increasing, abort_decreasing, noabort_increasing, noabort_decreasing
FROM abbrev_abort_uuids
ORDER BY ctid LIMIT 5;`,
				Results: []sql.Row{{20010, `00000000-0000-0000-0000-000000020000`, `00000000-0000-0000-0000-000000000000`, `00009991-0000-0000-0000-000000020000`, `00000000-0000-0000-0000-000000000000`}, {20001, `00000000-0000-0000-0000-000000020000`, `00000000-0000-0000-0000-000000000000`, `00009991-0000-0000-0000-000000020000`, `00000000-0000-0000-0000-000000000000`}, {9992, `00000000-0000-0000-0000-000000009991`, `00000000-0000-0000-0000-000000010009`, `00009991-0000-0000-0000-000000009991`, `00000000-0000-0000-0000-000000010009`}, {20000, `00000000-0000-0000-0000-000000019999`, `00000000-0000-0000-0000-000000000001`, `00009990-0000-0000-0000-000000019999`, `00000001-0000-0000-0000-000000000001`}, {9991, `00000000-0000-0000-0000-000000009990`, `00000000-0000-0000-0000-000000010010`, `00009990-0000-0000-0000-000000009990`, `00000001-0000-0000-0000-000000010010`}},
			},
			{
				Statement: `SELECT id, abort_increasing, abort_decreasing, noabort_increasing, noabort_decreasing
FROM abbrev_abort_uuids
ORDER BY ctid DESC LIMIT 5;`,
				Results: []sql.Row{{0, ``, ``, ``, ``}, {20003, ``, ``, ``, ``}, {20002, ``, ``, ``, ``}, {9993, `00000000-0000-0000-0000-000000009992`, `00000000-0000-0000-0000-000000010008`, `00009992-0000-0000-0000-000000009992`, `00010008-0000-0000-0000-000000010008`}, {9994, `00000000-0000-0000-0000-000000009993`, `00000000-0000-0000-0000-000000010007`, `00009993-0000-0000-0000-000000009993`, `00010007-0000-0000-0000-000000010007`}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `----
----
BEGIN;`,
			},
			{
				Statement: `SET LOCAL enable_indexscan = false;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF) DECLARE c SCROLL CURSOR FOR SELECT noabort_decreasing FROM abbrev_abort_uuids ORDER BY noabort_decreasing;`,
				Results:   []sql.Row{{`Sort`}, {`Sort Key: noabort_decreasing`}, {`->  Seq Scan on abbrev_abort_uuids`}},
			},
			{
				Statement: `DECLARE c SCROLL CURSOR FOR SELECT noabort_decreasing FROM abbrev_abort_uuids ORDER BY noabort_decreasing;`,
			},
			{
				Statement: `FETCH NEXT FROM c;`,
				Results:   []sql.Row{{`00000000-0000-0000-0000-000000000000`}},
			},
			{
				Statement: `FETCH NEXT FROM c;`,
				Results:   []sql.Row{{`00000000-0000-0000-0000-000000000000`}},
			},
			{
				Statement: `FETCH BACKWARD FROM c;`,
				Results:   []sql.Row{{`00000000-0000-0000-0000-000000000000`}},
			},
			{
				Statement: `FETCH BACKWARD FROM c;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `FETCH BACKWARD FROM c;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `FETCH BACKWARD FROM c;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `FETCH NEXT FROM c;`,
				Results:   []sql.Row{{`00000000-0000-0000-0000-000000000000`}},
			},
			{
				Statement: `FETCH LAST FROM c;`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `FETCH BACKWARD FROM c;`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `FETCH NEXT FROM c;`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `FETCH NEXT FROM c;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `FETCH NEXT FROM c;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `FETCH BACKWARD FROM c;`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `FETCH NEXT FROM c;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SET LOCAL enable_indexscan = false;`,
			},
			{
				Statement: `SET LOCAL work_mem = '100kB';`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF) DECLARE c SCROLL CURSOR FOR SELECT noabort_decreasing FROM abbrev_abort_uuids ORDER BY noabort_decreasing;`,
				Results:   []sql.Row{{`Sort`}, {`Sort Key: noabort_decreasing`}, {`->  Seq Scan on abbrev_abort_uuids`}},
			},
			{
				Statement: `DECLARE c SCROLL CURSOR FOR SELECT noabort_decreasing FROM abbrev_abort_uuids ORDER BY noabort_decreasing;`,
			},
			{
				Statement: `FETCH NEXT FROM c;`,
				Results:   []sql.Row{{`00000000-0000-0000-0000-000000000000`}},
			},
			{
				Statement: `FETCH NEXT FROM c;`,
				Results:   []sql.Row{{`00000000-0000-0000-0000-000000000000`}},
			},
			{
				Statement: `FETCH BACKWARD FROM c;`,
				Results:   []sql.Row{{`00000000-0000-0000-0000-000000000000`}},
			},
			{
				Statement: `FETCH BACKWARD FROM c;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `FETCH BACKWARD FROM c;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `FETCH BACKWARD FROM c;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `FETCH NEXT FROM c;`,
				Results:   []sql.Row{{`00000000-0000-0000-0000-000000000000`}},
			},
			{
				Statement: `FETCH LAST FROM c;`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `FETCH BACKWARD FROM c;`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `FETCH NEXT FROM c;`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `FETCH NEXT FROM c;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `FETCH NEXT FROM c;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `FETCH BACKWARD FROM c;`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `FETCH NEXT FROM c;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `----
---
SELECT
    -- fixed-width by-value datum
    (array_agg(id ORDER BY id DESC NULLS FIRST))[0:5],
    -- fixed-width by-ref datum
    (array_agg(abort_increasing ORDER BY abort_increasing DESC NULLS LAST))[0:5],
    -- variable-width datum
    (array_agg(id::text ORDER BY id::text DESC NULLS LAST))[0:5],
    -- fixed width by-value datum tuplesort
    percentile_disc(0.99) WITHIN GROUP (ORDER BY id),
    -- ensure state is shared
    percentile_disc(0.01) WITHIN GROUP (ORDER BY id),
    -- fixed width by-ref datum tuplesort
    percentile_disc(0.8) WITHIN GROUP (ORDER BY abort_increasing),
    -- variable width by-ref datum tuplesort
    percentile_disc(0.2) WITHIN GROUP (ORDER BY id::text),
    -- multi-column tuplesort
    rank('00000000-0000-0000-0000-000000000000', '2', '2') WITHIN GROUP (ORDER BY noabort_increasing, id, id::text)
FROM (
    SELECT * FROM abbrev_abort_uuids
    UNION ALL
    SELECT NULL, NULL, NULL, NULL, NULL) s;`,
				Results: []sql.Row{{`{NULL,20010,20009,20008,20007}`, `{00000000-0000-0000-0000-000000020000,00000000-0000-0000-0000-000000020000,00000000-0000-0000-0000-000000019999,00000000-0000-0000-0000-000000019998,00000000-0000-0000-0000-000000019997}`, `{9999,9998,9997,9996,9995}`, 19810, 200, `00000000-0000-0000-0000-000000016003`, 136, 2}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SET LOCAL work_mem = '100kB';`,
			},
			{
				Statement: `SELECT
    (array_agg(id ORDER BY id DESC NULLS FIRST))[0:5],
    (array_agg(abort_increasing ORDER BY abort_increasing DESC NULLS LAST))[0:5],
    (array_agg(id::text ORDER BY id::text DESC NULLS LAST))[0:5],
    percentile_disc(0.99) WITHIN GROUP (ORDER BY id),
    percentile_disc(0.01) WITHIN GROUP (ORDER BY id),
    percentile_disc(0.8) WITHIN GROUP (ORDER BY abort_increasing),
    percentile_disc(0.2) WITHIN GROUP (ORDER BY id::text),
    rank('00000000-0000-0000-0000-000000000000', '2', '2') WITHIN GROUP (ORDER BY noabort_increasing, id, id::text)
FROM (
    SELECT * FROM abbrev_abort_uuids
    UNION ALL
    SELECT NULL, NULL, NULL, NULL, NULL) s;`,
				Results: []sql.Row{{`{NULL,20010,20009,20008,20007}`, `{00000000-0000-0000-0000-000000020000,00000000-0000-0000-0000-000000020000,00000000-0000-0000-0000-000000019999,00000000-0000-0000-0000-000000019998,00000000-0000-0000-0000-000000019997}`, `{9999,9998,9997,9996,9995}`, 19810, 200, `00000000-0000-0000-0000-000000016003`, 136, 2}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `----
---
CREATE TEMP TABLE test_mark_restore(col1 int, col2 int, col12 int);`,
			},
			{
				Statement: `INSERT INTO test_mark_restore(col1, col2, col12)
   SELECT a.i, b.i, a.i * b.i FROM generate_series(1, 500) a(i), generate_series(1, 5) b(i);`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SET LOCAL enable_nestloop = off;`,
			},
			{
				Statement: `SET LOCAL enable_hashjoin = off;`,
			},
			{
				Statement: `SET LOCAL enable_material = off;`,
			},
			{
				Statement: `SELECT $$
    SELECT col12, count(distinct a.col1), count(distinct a.col2), count(distinct b.col1), count(distinct b.col2), count(*)
    FROM test_mark_restore a
        JOIN test_mark_restore b USING(col12)
    GROUP BY 1
    HAVING count(*) > 1
    ORDER BY 2 DESC, 1 DESC, 3 DESC, 4 DESC, 5 DESC, 6 DESC
    LIMIT 10
$$ AS qry \gset
EXPLAIN (COSTS OFF) :qry;`,
				Results: []sql.Row{{`Limit`}, {`->  Sort`}, {`Sort Key: (count(DISTINCT a.col1)) DESC, a.col12 DESC, (count(DISTINCT a.col2)) DESC, (count(DISTINCT b.col1)) DESC, (count(DISTINCT b.col2)) DESC, (count(*)) DESC`}, {`->  GroupAggregate`}, {`Group Key: a.col12`}, {`Filter: (count(*) > 1)`}, {`->  Merge Join`}, {`Merge Cond: (a.col12 = b.col12)`}, {`->  Sort`}, {`Sort Key: a.col12 DESC`}, {`->  Seq Scan on test_mark_restore a`}, {`->  Sort`}, {`Sort Key: b.col12 DESC`}, {`->  Seq Scan on test_mark_restore b`}},
			},
			{
				Statement: `:qry;`,
				Results:   []sql.Row{{480, 5, 5, 5, 5, 25}, {420, 5, 5, 5, 5, 25}, {360, 5, 5, 5, 5, 25}, {300, 5, 5, 5, 5, 25}, {240, 5, 5, 5, 5, 25}, {180, 5, 5, 5, 5, 25}, {120, 5, 5, 5, 5, 25}, {60, 5, 5, 5, 5, 25}, {960, 4, 4, 4, 4, 16}, {900, 4, 4, 4, 4, 16}},
			},
			{
				Statement: `SET LOCAL work_mem = '100kB';`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF) :qry;`,
				Results:   []sql.Row{{`Limit`}, {`->  Sort`}, {`Sort Key: (count(DISTINCT a.col1)) DESC, a.col12 DESC, (count(DISTINCT a.col2)) DESC, (count(DISTINCT b.col1)) DESC, (count(DISTINCT b.col2)) DESC, (count(*)) DESC`}, {`->  GroupAggregate`}, {`Group Key: a.col12`}, {`Filter: (count(*) > 1)`}, {`->  Merge Join`}, {`Merge Cond: (a.col12 = b.col12)`}, {`->  Sort`}, {`Sort Key: a.col12 DESC`}, {`->  Seq Scan on test_mark_restore a`}, {`->  Sort`}, {`Sort Key: b.col12 DESC`}, {`->  Seq Scan on test_mark_restore b`}},
			},
			{
				Statement: `:qry;`,
				Results:   []sql.Row{{480, 5, 5, 5, 5, 25}, {420, 5, 5, 5, 5, 25}, {360, 5, 5, 5, 5, 25}, {300, 5, 5, 5, 5, 25}, {240, 5, 5, 5, 5, 25}, {180, 5, 5, 5, 5, 25}, {120, 5, 5, 5, 5, 25}, {60, 5, 5, 5, 5, 25}, {960, 4, 4, 4, 4, 16}, {900, 4, 4, 4, 4, 16}},
			},
			{
				Statement: `COMMIT;`,
			},
		},
	})
}
