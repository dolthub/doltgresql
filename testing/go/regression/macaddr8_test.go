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

func TestMacaddr8(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_macaddr8)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_macaddr8,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `SELECT '08:00:2b:01:02:03     '::macaddr8;`,
				Results:   []sql.Row{{`08:00:2b:ff:fe:01:02:03`}},
			},
			{
				Statement: `SELECT '    08:00:2b:01:02:03     '::macaddr8;`,
				Results:   []sql.Row{{`08:00:2b:ff:fe:01:02:03`}},
			},
			{
				Statement: `SELECT '    08:00:2b:01:02:03'::macaddr8;`,
				Results:   []sql.Row{{`08:00:2b:ff:fe:01:02:03`}},
			},
			{
				Statement: `SELECT '08:00:2b:01:02:03:04:05     '::macaddr8;`,
				Results:   []sql.Row{{`08:00:2b:01:02:03:04:05`}},
			},
			{
				Statement: `SELECT '    08:00:2b:01:02:03:04:05     '::macaddr8;`,
				Results:   []sql.Row{{`08:00:2b:01:02:03:04:05`}},
			},
			{
				Statement: `SELECT '    08:00:2b:01:02:03:04:05'::macaddr8;`,
				Results:   []sql.Row{{`08:00:2b:01:02:03:04:05`}},
			},
			{
				Statement:   `SELECT '123    08:00:2b:01:02:03'::macaddr8; -- invalid`,
				ErrorString: `invalid input syntax for type macaddr8: "123    08:00:2b:01:02:03"`,
			},
			{
				Statement:   `SELECT '08:00:2b:01:02:03  123'::macaddr8; -- invalid`,
				ErrorString: `invalid input syntax for type macaddr8: "08:00:2b:01:02:03  123"`,
			},
			{
				Statement:   `SELECT '123    08:00:2b:01:02:03:04:05'::macaddr8; -- invalid`,
				ErrorString: `invalid input syntax for type macaddr8: "123    08:00:2b:01:02:03:04:05"`,
			},
			{
				Statement:   `SELECT '08:00:2b:01:02:03:04:05  123'::macaddr8; -- invalid`,
				ErrorString: `invalid input syntax for type macaddr8: "08:00:2b:01:02:03:04:05  123"`,
			},
			{
				Statement:   `SELECT '08:00:2b:01:02:03:04:05:06:07'::macaddr8; -- invalid`,
				ErrorString: `invalid input syntax for type macaddr8: "08:00:2b:01:02:03:04:05:06:07"`,
			},
			{
				Statement:   `SELECT '08-00-2b-01-02-03-04-05-06-07'::macaddr8; -- invalid`,
				ErrorString: `invalid input syntax for type macaddr8: "08-00-2b-01-02-03-04-05-06-07"`,
			},
			{
				Statement:   `SELECT '08002b:01020304050607'::macaddr8; -- invalid`,
				ErrorString: `invalid input syntax for type macaddr8: "08002b:01020304050607"`,
			},
			{
				Statement:   `SELECT '08002b01020304050607'::macaddr8; -- invalid`,
				ErrorString: `invalid input syntax for type macaddr8: "08002b01020304050607"`,
			},
			{
				Statement:   `SELECT '0z002b0102030405'::macaddr8; -- invalid`,
				ErrorString: `invalid input syntax for type macaddr8: "0z002b0102030405"`,
			},
			{
				Statement:   `SELECT '08002b010203xyza'::macaddr8; -- invalid`,
				ErrorString: `invalid input syntax for type macaddr8: "08002b010203xyza"`,
			},
			{
				Statement:   `SELECT '08:00-2b:01:02:03:04:05'::macaddr8; -- invalid`,
				ErrorString: `invalid input syntax for type macaddr8: "08:00-2b:01:02:03:04:05"`,
			},
			{
				Statement:   `SELECT '08:00-2b:01:02:03:04:05'::macaddr8; -- invalid`,
				ErrorString: `invalid input syntax for type macaddr8: "08:00-2b:01:02:03:04:05"`,
			},
			{
				Statement:   `SELECT '08:00:2b:01.02:03:04:05'::macaddr8; -- invalid`,
				ErrorString: `invalid input syntax for type macaddr8: "08:00:2b:01.02:03:04:05"`,
			},
			{
				Statement:   `SELECT '08:00:2b:01.02:03:04:05'::macaddr8; -- invalid`,
				ErrorString: `invalid input syntax for type macaddr8: "08:00:2b:01.02:03:04:05"`,
			},
			{
				Statement: `SELECT macaddr8_set7bit('00:08:2b:01:02:03'::macaddr8);`,
				Results:   []sql.Row{{`02:08:2b:ff:fe:01:02:03`}},
			},
			{
				Statement: `CREATE TABLE macaddr8_data (a int, b macaddr8);`,
			},
			{
				Statement: `INSERT INTO macaddr8_data VALUES (1, '08:00:2b:01:02:03');`,
			},
			{
				Statement: `INSERT INTO macaddr8_data VALUES (2, '08-00-2b-01-02-03');`,
			},
			{
				Statement: `INSERT INTO macaddr8_data VALUES (3, '08002b:010203');`,
			},
			{
				Statement: `INSERT INTO macaddr8_data VALUES (4, '08002b-010203');`,
			},
			{
				Statement: `INSERT INTO macaddr8_data VALUES (5, '0800.2b01.0203');`,
			},
			{
				Statement: `INSERT INTO macaddr8_data VALUES (6, '0800-2b01-0203');`,
			},
			{
				Statement: `INSERT INTO macaddr8_data VALUES (7, '08002b010203');`,
			},
			{
				Statement: `INSERT INTO macaddr8_data VALUES (8, '0800:2b01:0203');`,
			},
			{
				Statement:   `INSERT INTO macaddr8_data VALUES (9, 'not even close'); -- invalid`,
				ErrorString: `invalid input syntax for type macaddr8: "not even close"`,
			},
			{
				Statement: `INSERT INTO macaddr8_data VALUES (10, '08:00:2b:01:02:04');`,
			},
			{
				Statement: `INSERT INTO macaddr8_data VALUES (11, '08:00:2b:01:02:02');`,
			},
			{
				Statement: `INSERT INTO macaddr8_data VALUES (12, '08:00:2a:01:02:03');`,
			},
			{
				Statement: `INSERT INTO macaddr8_data VALUES (13, '08:00:2c:01:02:03');`,
			},
			{
				Statement: `INSERT INTO macaddr8_data VALUES (14, '08:00:2a:01:02:04');`,
			},
			{
				Statement: `INSERT INTO macaddr8_data VALUES (15, '08:00:2b:01:02:03:04:05');`,
			},
			{
				Statement: `INSERT INTO macaddr8_data VALUES (16, '08-00-2b-01-02-03-04-05');`,
			},
			{
				Statement: `INSERT INTO macaddr8_data VALUES (17, '08002b:0102030405');`,
			},
			{
				Statement: `INSERT INTO macaddr8_data VALUES (18, '08002b-0102030405');`,
			},
			{
				Statement: `INSERT INTO macaddr8_data VALUES (19, '0800.2b01.0203.0405');`,
			},
			{
				Statement: `INSERT INTO macaddr8_data VALUES (20, '08002b01:02030405');`,
			},
			{
				Statement: `INSERT INTO macaddr8_data VALUES (21, '08002b0102030405');`,
			},
			{
				Statement: `SELECT * FROM macaddr8_data ORDER BY 1;`,
				Results:   []sql.Row{{1, `08:00:2b:ff:fe:01:02:03`}, {2, `08:00:2b:ff:fe:01:02:03`}, {3, `08:00:2b:ff:fe:01:02:03`}, {4, `08:00:2b:ff:fe:01:02:03`}, {5, `08:00:2b:ff:fe:01:02:03`}, {6, `08:00:2b:ff:fe:01:02:03`}, {7, `08:00:2b:ff:fe:01:02:03`}, {8, `08:00:2b:ff:fe:01:02:03`}, {10, `08:00:2b:ff:fe:01:02:04`}, {11, `08:00:2b:ff:fe:01:02:02`}, {12, `08:00:2a:ff:fe:01:02:03`}, {13, `08:00:2c:ff:fe:01:02:03`}, {14, `08:00:2a:ff:fe:01:02:04`}, {15, `08:00:2b:01:02:03:04:05`}, {16, `08:00:2b:01:02:03:04:05`}, {17, `08:00:2b:01:02:03:04:05`}, {18, `08:00:2b:01:02:03:04:05`}, {19, `08:00:2b:01:02:03:04:05`}, {20, `08:00:2b:01:02:03:04:05`}, {21, `08:00:2b:01:02:03:04:05`}},
			},
			{
				Statement: `CREATE INDEX macaddr8_data_btree ON macaddr8_data USING btree (b);`,
			},
			{
				Statement: `CREATE INDEX macaddr8_data_hash ON macaddr8_data USING hash (b);`,
			},
			{
				Statement: `SELECT a, b, trunc(b) FROM macaddr8_data ORDER BY 2, 1;`,
				Results:   []sql.Row{{12, `08:00:2a:ff:fe:01:02:03`, `08:00:2a:00:00:00:00:00`}, {14, `08:00:2a:ff:fe:01:02:04`, `08:00:2a:00:00:00:00:00`}, {15, `08:00:2b:01:02:03:04:05`, `08:00:2b:00:00:00:00:00`}, {16, `08:00:2b:01:02:03:04:05`, `08:00:2b:00:00:00:00:00`}, {17, `08:00:2b:01:02:03:04:05`, `08:00:2b:00:00:00:00:00`}, {18, `08:00:2b:01:02:03:04:05`, `08:00:2b:00:00:00:00:00`}, {19, `08:00:2b:01:02:03:04:05`, `08:00:2b:00:00:00:00:00`}, {20, `08:00:2b:01:02:03:04:05`, `08:00:2b:00:00:00:00:00`}, {21, `08:00:2b:01:02:03:04:05`, `08:00:2b:00:00:00:00:00`}, {11, `08:00:2b:ff:fe:01:02:02`, `08:00:2b:00:00:00:00:00`}, {1, `08:00:2b:ff:fe:01:02:03`, `08:00:2b:00:00:00:00:00`}, {2, `08:00:2b:ff:fe:01:02:03`, `08:00:2b:00:00:00:00:00`}, {3, `08:00:2b:ff:fe:01:02:03`, `08:00:2b:00:00:00:00:00`}, {4, `08:00:2b:ff:fe:01:02:03`, `08:00:2b:00:00:00:00:00`}, {5, `08:00:2b:ff:fe:01:02:03`, `08:00:2b:00:00:00:00:00`}, {6, `08:00:2b:ff:fe:01:02:03`, `08:00:2b:00:00:00:00:00`}, {7, `08:00:2b:ff:fe:01:02:03`, `08:00:2b:00:00:00:00:00`}, {8, `08:00:2b:ff:fe:01:02:03`, `08:00:2b:00:00:00:00:00`}, {10, `08:00:2b:ff:fe:01:02:04`, `08:00:2b:00:00:00:00:00`}, {13, `08:00:2c:ff:fe:01:02:03`, `08:00:2c:00:00:00:00:00`}},
			},
			{
				Statement: `SELECT b <  '08:00:2b:01:02:04' FROM macaddr8_data WHERE a = 1; -- true`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT b >  '08:00:2b:ff:fe:01:02:04' FROM macaddr8_data WHERE a = 1; -- false`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT b >  '08:00:2b:ff:fe:01:02:03' FROM macaddr8_data WHERE a = 1; -- false`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT b::macaddr <= '08:00:2b:01:02:04' FROM macaddr8_data WHERE a = 1; -- true`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT b::macaddr >= '08:00:2b:01:02:04' FROM macaddr8_data WHERE a = 1; -- false`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT b =  '08:00:2b:ff:fe:01:02:03' FROM macaddr8_data WHERE a = 1; -- true`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT b::macaddr <> '08:00:2b:01:02:04'::macaddr FROM macaddr8_data WHERE a = 1; -- true`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT b::macaddr <> '08:00:2b:01:02:03'::macaddr FROM macaddr8_data WHERE a = 1; -- false`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT b <  '08:00:2b:01:02:03:04:06' FROM macaddr8_data WHERE a = 15; -- true`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT b >  '08:00:2b:01:02:03:04:06' FROM macaddr8_data WHERE a = 15; -- false`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT b >  '08:00:2b:01:02:03:04:05' FROM macaddr8_data WHERE a = 15; -- false`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT b <= '08:00:2b:01:02:03:04:06' FROM macaddr8_data WHERE a = 15; -- true`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT b >= '08:00:2b:01:02:03:04:06' FROM macaddr8_data WHERE a = 15; -- false`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT b =  '08:00:2b:01:02:03:04:05' FROM macaddr8_data WHERE a = 15; -- true`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT b <> '08:00:2b:01:02:03:04:06' FROM macaddr8_data WHERE a = 15; -- true`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT b <> '08:00:2b:01:02:03:04:05' FROM macaddr8_data WHERE a = 15; -- false`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT ~b                       FROM macaddr8_data;`,
				Results:   []sql.Row{{`f7:ff:d4:00:01:fe:fd:fc`}, {`f7:ff:d4:00:01:fe:fd:fc`}, {`f7:ff:d4:00:01:fe:fd:fc`}, {`f7:ff:d4:00:01:fe:fd:fc`}, {`f7:ff:d4:00:01:fe:fd:fc`}, {`f7:ff:d4:00:01:fe:fd:fc`}, {`f7:ff:d4:00:01:fe:fd:fc`}, {`f7:ff:d4:00:01:fe:fd:fc`}, {`f7:ff:d4:00:01:fe:fd:fb`}, {`f7:ff:d4:00:01:fe:fd:fd`}, {`f7:ff:d5:00:01:fe:fd:fc`}, {`f7:ff:d3:00:01:fe:fd:fc`}, {`f7:ff:d5:00:01:fe:fd:fb`}, {`f7:ff:d4:fe:fd:fc:fb:fa`}, {`f7:ff:d4:fe:fd:fc:fb:fa`}, {`f7:ff:d4:fe:fd:fc:fb:fa`}, {`f7:ff:d4:fe:fd:fc:fb:fa`}, {`f7:ff:d4:fe:fd:fc:fb:fa`}, {`f7:ff:d4:fe:fd:fc:fb:fa`}, {`f7:ff:d4:fe:fd:fc:fb:fa`}},
			},
			{
				Statement: `SELECT  b & '00:00:00:ff:ff:ff' FROM macaddr8_data;`,
				Results:   []sql.Row{{`00:00:00:ff:fe:01:02:03`}, {`00:00:00:ff:fe:01:02:03`}, {`00:00:00:ff:fe:01:02:03`}, {`00:00:00:ff:fe:01:02:03`}, {`00:00:00:ff:fe:01:02:03`}, {`00:00:00:ff:fe:01:02:03`}, {`00:00:00:ff:fe:01:02:03`}, {`00:00:00:ff:fe:01:02:03`}, {`00:00:00:ff:fe:01:02:04`}, {`00:00:00:ff:fe:01:02:02`}, {`00:00:00:ff:fe:01:02:03`}, {`00:00:00:ff:fe:01:02:03`}, {`00:00:00:ff:fe:01:02:04`}, {`00:00:00:01:02:03:04:05`}, {`00:00:00:01:02:03:04:05`}, {`00:00:00:01:02:03:04:05`}, {`00:00:00:01:02:03:04:05`}, {`00:00:00:01:02:03:04:05`}, {`00:00:00:01:02:03:04:05`}, {`00:00:00:01:02:03:04:05`}},
			},
			{
				Statement: `SELECT  b | '01:02:03:04:05:06' FROM macaddr8_data;`,
				Results:   []sql.Row{{`09:02:2b:ff:fe:05:07:07`}, {`09:02:2b:ff:fe:05:07:07`}, {`09:02:2b:ff:fe:05:07:07`}, {`09:02:2b:ff:fe:05:07:07`}, {`09:02:2b:ff:fe:05:07:07`}, {`09:02:2b:ff:fe:05:07:07`}, {`09:02:2b:ff:fe:05:07:07`}, {`09:02:2b:ff:fe:05:07:07`}, {`09:02:2b:ff:fe:05:07:06`}, {`09:02:2b:ff:fe:05:07:06`}, {`09:02:2b:ff:fe:05:07:07`}, {`09:02:2f:ff:fe:05:07:07`}, {`09:02:2b:ff:fe:05:07:06`}, {`09:02:2b:ff:fe:07:05:07`}, {`09:02:2b:ff:fe:07:05:07`}, {`09:02:2b:ff:fe:07:05:07`}, {`09:02:2b:ff:fe:07:05:07`}, {`09:02:2b:ff:fe:07:05:07`}, {`09:02:2b:ff:fe:07:05:07`}, {`09:02:2b:ff:fe:07:05:07`}},
			},
			{
				Statement: `DROP TABLE macaddr8_data;`,
			},
		},
	})
}
