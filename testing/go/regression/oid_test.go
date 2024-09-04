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

func TestOid(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_oid)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_oid,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TABLE OID_TBL(f1 oid);`,
			},
			{
				Statement: `INSERT INTO OID_TBL(f1) VALUES ('1234');`,
			},
			{
				Statement: `INSERT INTO OID_TBL(f1) VALUES ('1235');`,
			},
			{
				Statement: `INSERT INTO OID_TBL(f1) VALUES ('987');`,
			},
			{
				Statement: `INSERT INTO OID_TBL(f1) VALUES ('-1040');`,
			},
			{
				Statement: `INSERT INTO OID_TBL(f1) VALUES ('99999999');`,
			},
			{
				Statement: `INSERT INTO OID_TBL(f1) VALUES ('5     ');`,
			},
			{
				Statement: `INSERT INTO OID_TBL(f1) VALUES ('   10  ');`,
			},
			{
				Statement: `INSERT INTO OID_TBL(f1) VALUES ('	  15 	  ');`,
			},
			{
				Statement:   `INSERT INTO OID_TBL(f1) VALUES ('');`,
				ErrorString: `invalid input syntax for type oid: ""`,
			},
			{
				Statement:   `INSERT INTO OID_TBL(f1) VALUES ('    ');`,
				ErrorString: `invalid input syntax for type oid: "    "`,
			},
			{
				Statement:   `INSERT INTO OID_TBL(f1) VALUES ('asdfasd');`,
				ErrorString: `invalid input syntax for type oid: "asdfasd"`,
			},
			{
				Statement:   `INSERT INTO OID_TBL(f1) VALUES ('99asdfasd');`,
				ErrorString: `invalid input syntax for type oid: "99asdfasd"`,
			},
			{
				Statement:   `INSERT INTO OID_TBL(f1) VALUES ('5    d');`,
				ErrorString: `invalid input syntax for type oid: "5    d"`,
			},
			{
				Statement:   `INSERT INTO OID_TBL(f1) VALUES ('    5d');`,
				ErrorString: `invalid input syntax for type oid: "    5d"`,
			},
			{
				Statement:   `INSERT INTO OID_TBL(f1) VALUES ('5    5');`,
				ErrorString: `invalid input syntax for type oid: "5    5"`,
			},
			{
				Statement:   `INSERT INTO OID_TBL(f1) VALUES (' - 500');`,
				ErrorString: `invalid input syntax for type oid: " - 500"`,
			},
			{
				Statement:   `INSERT INTO OID_TBL(f1) VALUES ('32958209582039852935');`,
				ErrorString: `value "32958209582039852935" is out of range for type oid`,
			},
			{
				Statement:   `INSERT INTO OID_TBL(f1) VALUES ('-23582358720398502385');`,
				ErrorString: `value "-23582358720398502385" is out of range for type oid`,
			},
			{
				Statement: `SELECT * FROM OID_TBL;`,
				Results:   []sql.Row{{1234}, {1235}, {987}, {4294966256}, {99999999}, {5}, {10}, {15}},
			},
			{
				Statement: `SELECT o.* FROM OID_TBL o WHERE o.f1 = 1234;`,
				Results:   []sql.Row{{1234}},
			},
			{
				Statement: `SELECT o.* FROM OID_TBL o WHERE o.f1 <> '1234';`,
				Results:   []sql.Row{{1235}, {987}, {4294966256}, {99999999}, {5}, {10}, {15}},
			},
			{
				Statement: `SELECT o.* FROM OID_TBL o WHERE o.f1 <= '1234';`,
				Results:   []sql.Row{{1234}, {987}, {5}, {10}, {15}},
			},
			{
				Statement: `SELECT o.* FROM OID_TBL o WHERE o.f1 < '1234';`,
				Results:   []sql.Row{{987}, {5}, {10}, {15}},
			},
			{
				Statement: `SELECT o.* FROM OID_TBL o WHERE o.f1 >= '1234';`,
				Results:   []sql.Row{{1234}, {1235}, {4294966256}, {99999999}},
			},
			{
				Statement: `SELECT o.* FROM OID_TBL o WHERE o.f1 > '1234';`,
				Results:   []sql.Row{{1235}, {4294966256}, {99999999}},
			},
			{
				Statement: `DROP TABLE OID_TBL;`,
			},
		},
	})
}
