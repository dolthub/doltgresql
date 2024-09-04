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

func TestTxid(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_txid)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_txid,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `select '12:13:'::txid_snapshot;`,
				Results:   []sql.Row{{`12:13:`}},
			},
			{
				Statement: `select '12:18:14,16'::txid_snapshot;`,
				Results:   []sql.Row{{`12:18:14,16`}},
			},
			{
				Statement: `select '12:16:14,14'::txid_snapshot;`,
				Results:   []sql.Row{{`12:16:14`}},
			},
			{
				Statement:   `select '31:12:'::txid_snapshot;`,
				ErrorString: `invalid input syntax for type pg_snapshot: "31:12:"`,
			},
			{
				Statement:   `select '0:1:'::txid_snapshot;`,
				ErrorString: `invalid input syntax for type pg_snapshot: "0:1:"`,
			},
			{
				Statement:   `select '12:13:0'::txid_snapshot;`,
				ErrorString: `invalid input syntax for type pg_snapshot: "12:13:0"`,
			},
			{
				Statement:   `select '12:16:14,13'::txid_snapshot;`,
				ErrorString: `invalid input syntax for type pg_snapshot: "12:16:14,13"`,
			},
			{
				Statement: `create temp table snapshot_test (
	nr	integer,
	snap	txid_snapshot
);`,
			},
			{
				Statement: `insert into snapshot_test values (1, '12:13:');`,
			},
			{
				Statement: `insert into snapshot_test values (2, '12:20:13,15,18');`,
			},
			{
				Statement: `insert into snapshot_test values (3, '100001:100009:100005,100007,100008');`,
			},
			{
				Statement: `insert into snapshot_test values (4, '100:150:101,102,103,104,105,106,107,108,109,110,111,112,113,114,115,116,117,118,119,120,121,122,123,124,125,126,127,128,129,130,131');`,
			},
			{
				Statement: `select snap from snapshot_test order by nr;`,
				Results:   []sql.Row{{`12:13:`}, {`12:20:13,15,18`}, {`100001:100009:100005,100007,100008`}, {`100:150:101,102,103,104,105,106,107,108,109,110,111,112,113,114,115,116,117,118,119,120,121,122,123,124,125,126,127,128,129,130,131`}},
			},
			{
				Statement: `select  txid_snapshot_xmin(snap),
	txid_snapshot_xmax(snap),
	txid_snapshot_xip(snap)
from snapshot_test order by nr;`,
				Results: []sql.Row{{12, 20, 13}, {12, 20, 15}, {12, 20, 18}, {100001, 100009, 100005}, {100001, 100009, 100007}, {100001, 100009, 100008}, {100, 150, 101}, {100, 150, 102}, {100, 150, 103}, {100, 150, 104}, {100, 150, 105}, {100, 150, 106}, {100, 150, 107}, {100, 150, 108}, {100, 150, 109}, {100, 150, 110}, {100, 150, 111}, {100, 150, 112}, {100, 150, 113}, {100, 150, 114}, {100, 150, 115}, {100, 150, 116}, {100, 150, 117}, {100, 150, 118}, {100, 150, 119}, {100, 150, 120}, {100, 150, 121}, {100, 150, 122}, {100, 150, 123}, {100, 150, 124}, {100, 150, 125}, {100, 150, 126}, {100, 150, 127}, {100, 150, 128}, {100, 150, 129}, {100, 150, 130}, {100, 150, 131}},
			},
			{
				Statement: `select id, txid_visible_in_snapshot(id, snap)
from snapshot_test, generate_series(11, 21) id
where nr = 2;`,
				Results: []sql.Row{{11, true}, {12, true}, {13, false}, {14, true}, {15, false}, {16, true}, {17, true}, {18, false}, {19, true}, {20, false}, {21, false}},
			},
			{
				Statement: `select id, txid_visible_in_snapshot(id, snap)
from snapshot_test, generate_series(90, 160) id
where nr = 4;`,
				Results: []sql.Row{{90, true}, {91, true}, {92, true}, {93, true}, {94, true}, {95, true}, {96, true}, {97, true}, {98, true}, {99, true}, {100, true}, {101, false}, {102, false}, {103, false}, {104, false}, {105, false}, {106, false}, {107, false}, {108, false}, {109, false}, {110, false}, {111, false}, {112, false}, {113, false}, {114, false}, {115, false}, {116, false}, {117, false}, {118, false}, {119, false}, {120, false}, {121, false}, {122, false}, {123, false}, {124, false}, {125, false}, {126, false}, {127, false}, {128, false}, {129, false}, {130, false}, {131, false}, {132, true}, {133, true}, {134, true}, {135, true}, {136, true}, {137, true}, {138, true}, {139, true}, {140, true}, {141, true}, {142, true}, {143, true}, {144, true}, {145, true}, {146, true}, {147, true}, {148, true}, {149, true}, {150, false}, {151, false}, {152, false}, {153, false}, {154, false}, {155, false}, {156, false}, {157, false}, {158, false}, {159, false}, {160, false}},
			},
			{
				Statement: `select txid_current() >= txid_snapshot_xmin(txid_current_snapshot());`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select txid_visible_in_snapshot(txid_current(), txid_current_snapshot());`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select txid_snapshot '1000100010001000:1000100010001100:1000100010001012,1000100010001013';`,
				Results:   []sql.Row{{`1000100010001000:1000100010001100:1000100010001012,1000100010001013`}},
			},
			{
				Statement: `select txid_visible_in_snapshot('1000100010001012', '1000100010001000:1000100010001100:1000100010001012,1000100010001013');`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select txid_visible_in_snapshot('1000100010001015', '1000100010001000:1000100010001100:1000100010001012,1000100010001013');`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT txid_snapshot '1:9223372036854775807:3';`,
				Results:   []sql.Row{{`1:9223372036854775807:3`}},
			},
			{
				Statement:   `SELECT txid_snapshot '1:9223372036854775808:3';`,
				ErrorString: `invalid input syntax for type pg_snapshot: "1:9223372036854775808:3"`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SELECT txid_current_if_assigned() IS NULL;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT txid_current() \gset
SELECT txid_current_if_assigned() IS NOT DISTINCT FROM BIGINT :'txid_current';`,
				Results: []sql.Row{{true}},
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SELECT txid_current() AS committed \gset
COMMIT;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SELECT txid_current() AS rolledback \gset
ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SELECT txid_current() AS inprogress \gset
SELECT txid_status(:committed) AS committed;`,
				Results: []sql.Row{{`committed`}},
			},
			{
				Statement: `SELECT txid_status(:rolledback) AS rolledback;`,
				Results:   []sql.Row{{`aborted`}},
			},
			{
				Statement: `SELECT txid_status(:inprogress) AS inprogress;`,
				Results:   []sql.Row{{`in progress`}},
			},
			{
				Statement: `SELECT txid_status(1); -- BootstrapTransactionId is always committed`,
				Results:   []sql.Row{{`committed`}},
			},
			{
				Statement: `SELECT txid_status(2); -- FrozenTransactionId is always committed`,
				Results:   []sql.Row{{`committed`}},
			},
			{
				Statement: `SELECT txid_status(3); -- in regress testing FirstNormalTransactionId will always be behind oldestXmin`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `CREATE FUNCTION test_future_xid_status(bigint)
RETURNS void
LANGUAGE plpgsql
AS
$$
BEGIN
  PERFORM txid_status($1);`,
			},
			{
				Statement: `  RAISE EXCEPTION 'didn''t ERROR at xid in the future as expected';`,
			},
			{
				Statement: `EXCEPTION
  WHEN invalid_parameter_value THEN
    RAISE NOTICE 'Got expected error for xid in the future';`,
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `SELECT test_future_xid_status(:inprogress + 10000);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `ROLLBACK;`,
			},
		},
	})
}
