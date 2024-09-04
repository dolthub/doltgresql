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

func TestMacaddr(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_macaddr)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_macaddr,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TABLE macaddr_data (a int, b macaddr);`,
			},
			{
				Statement: `INSERT INTO macaddr_data VALUES (1, '08:00:2b:01:02:03');`,
			},
			{
				Statement: `INSERT INTO macaddr_data VALUES (2, '08-00-2b-01-02-03');`,
			},
			{
				Statement: `INSERT INTO macaddr_data VALUES (3, '08002b:010203');`,
			},
			{
				Statement: `INSERT INTO macaddr_data VALUES (4, '08002b-010203');`,
			},
			{
				Statement: `INSERT INTO macaddr_data VALUES (5, '0800.2b01.0203');`,
			},
			{
				Statement: `INSERT INTO macaddr_data VALUES (6, '0800-2b01-0203');`,
			},
			{
				Statement: `INSERT INTO macaddr_data VALUES (7, '08002b010203');`,
			},
			{
				Statement:   `INSERT INTO macaddr_data VALUES (8, '0800:2b01:0203'); -- invalid`,
				ErrorString: `invalid input syntax for type macaddr: "0800:2b01:0203"`,
			},
			{
				Statement:   `INSERT INTO macaddr_data VALUES (9, 'not even close'); -- invalid`,
				ErrorString: `invalid input syntax for type macaddr: "not even close"`,
			},
			{
				Statement: `INSERT INTO macaddr_data VALUES (10, '08:00:2b:01:02:04');`,
			},
			{
				Statement: `INSERT INTO macaddr_data VALUES (11, '08:00:2b:01:02:02');`,
			},
			{
				Statement: `INSERT INTO macaddr_data VALUES (12, '08:00:2a:01:02:03');`,
			},
			{
				Statement: `INSERT INTO macaddr_data VALUES (13, '08:00:2c:01:02:03');`,
			},
			{
				Statement: `INSERT INTO macaddr_data VALUES (14, '08:00:2a:01:02:04');`,
			},
			{
				Statement: `SELECT * FROM macaddr_data;`,
				Results:   []sql.Row{{1, `08:00:2b:01:02:03`}, {2, `08:00:2b:01:02:03`}, {3, `08:00:2b:01:02:03`}, {4, `08:00:2b:01:02:03`}, {5, `08:00:2b:01:02:03`}, {6, `08:00:2b:01:02:03`}, {7, `08:00:2b:01:02:03`}, {10, `08:00:2b:01:02:04`}, {11, `08:00:2b:01:02:02`}, {12, `08:00:2a:01:02:03`}, {13, `08:00:2c:01:02:03`}, {14, `08:00:2a:01:02:04`}},
			},
			{
				Statement: `CREATE INDEX macaddr_data_btree ON macaddr_data USING btree (b);`,
			},
			{
				Statement: `CREATE INDEX macaddr_data_hash ON macaddr_data USING hash (b);`,
			},
			{
				Statement: `SELECT a, b, trunc(b) FROM macaddr_data ORDER BY 2, 1;`,
				Results:   []sql.Row{{12, `08:00:2a:01:02:03`, `08:00:2a:00:00:00`}, {14, `08:00:2a:01:02:04`, `08:00:2a:00:00:00`}, {11, `08:00:2b:01:02:02`, `08:00:2b:00:00:00`}, {1, `08:00:2b:01:02:03`, `08:00:2b:00:00:00`}, {2, `08:00:2b:01:02:03`, `08:00:2b:00:00:00`}, {3, `08:00:2b:01:02:03`, `08:00:2b:00:00:00`}, {4, `08:00:2b:01:02:03`, `08:00:2b:00:00:00`}, {5, `08:00:2b:01:02:03`, `08:00:2b:00:00:00`}, {6, `08:00:2b:01:02:03`, `08:00:2b:00:00:00`}, {7, `08:00:2b:01:02:03`, `08:00:2b:00:00:00`}, {10, `08:00:2b:01:02:04`, `08:00:2b:00:00:00`}, {13, `08:00:2c:01:02:03`, `08:00:2c:00:00:00`}},
			},
			{
				Statement: `SELECT b <  '08:00:2b:01:02:04' FROM macaddr_data WHERE a = 1; -- true`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT b >  '08:00:2b:01:02:04' FROM macaddr_data WHERE a = 1; -- false`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT b >  '08:00:2b:01:02:03' FROM macaddr_data WHERE a = 1; -- false`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT b <= '08:00:2b:01:02:04' FROM macaddr_data WHERE a = 1; -- true`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT b >= '08:00:2b:01:02:04' FROM macaddr_data WHERE a = 1; -- false`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT b =  '08:00:2b:01:02:03' FROM macaddr_data WHERE a = 1; -- true`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT b <> '08:00:2b:01:02:04' FROM macaddr_data WHERE a = 1; -- true`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT b <> '08:00:2b:01:02:03' FROM macaddr_data WHERE a = 1; -- false`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT ~b                       FROM macaddr_data;`,
				Results:   []sql.Row{{`f7:ff:d4:fe:fd:fc`}, {`f7:ff:d4:fe:fd:fc`}, {`f7:ff:d4:fe:fd:fc`}, {`f7:ff:d4:fe:fd:fc`}, {`f7:ff:d4:fe:fd:fc`}, {`f7:ff:d4:fe:fd:fc`}, {`f7:ff:d4:fe:fd:fc`}, {`f7:ff:d4:fe:fd:fb`}, {`f7:ff:d4:fe:fd:fd`}, {`f7:ff:d5:fe:fd:fc`}, {`f7:ff:d3:fe:fd:fc`}, {`f7:ff:d5:fe:fd:fb`}},
			},
			{
				Statement: `SELECT  b & '00:00:00:ff:ff:ff' FROM macaddr_data;`,
				Results:   []sql.Row{{`00:00:00:01:02:03`}, {`00:00:00:01:02:03`}, {`00:00:00:01:02:03`}, {`00:00:00:01:02:03`}, {`00:00:00:01:02:03`}, {`00:00:00:01:02:03`}, {`00:00:00:01:02:03`}, {`00:00:00:01:02:04`}, {`00:00:00:01:02:02`}, {`00:00:00:01:02:03`}, {`00:00:00:01:02:03`}, {`00:00:00:01:02:04`}},
			},
			{
				Statement: `SELECT  b | '01:02:03:04:05:06' FROM macaddr_data;`,
				Results:   []sql.Row{{`09:02:2b:05:07:07`}, {`09:02:2b:05:07:07`}, {`09:02:2b:05:07:07`}, {`09:02:2b:05:07:07`}, {`09:02:2b:05:07:07`}, {`09:02:2b:05:07:07`}, {`09:02:2b:05:07:07`}, {`09:02:2b:05:07:06`}, {`09:02:2b:05:07:06`}, {`09:02:2b:05:07:07`}, {`09:02:2f:05:07:07`}, {`09:02:2b:05:07:06`}},
			},
			{
				Statement: `DROP TABLE macaddr_data;`,
			},
		},
	})
}
