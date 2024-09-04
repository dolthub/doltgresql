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

func TestVarchar2(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_varchar_2)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_varchar_2,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TEMP TABLE VARCHAR_TBL(f1 varchar(1));`,
			},
			{
				Statement: `INSERT INTO VARCHAR_TBL (f1) VALUES ('a');`,
			},
			{
				Statement: `INSERT INTO VARCHAR_TBL (f1) VALUES ('A');`,
			},
			{
				Statement: `INSERT INTO VARCHAR_TBL (f1) VALUES ('1');`,
			},
			{
				Statement: `INSERT INTO VARCHAR_TBL (f1) VALUES (2);`,
			},
			{
				Statement: `INSERT INTO VARCHAR_TBL (f1) VALUES ('3');`,
			},
			{
				Statement: `INSERT INTO VARCHAR_TBL (f1) VALUES ('');`,
			},
			{
				Statement:   `INSERT INTO VARCHAR_TBL (f1) VALUES ('cd');`,
				ErrorString: `value too long for type character varying(1)`,
			},
			{
				Statement: `INSERT INTO VARCHAR_TBL (f1) VALUES ('c     ');`,
			},
			{
				Statement: `SELECT * FROM VARCHAR_TBL;`,
				Results:   []sql.Row{{`a`}, {`A`}, {1}, {2}, {3}, {``}, {`c`}},
			},
			{
				Statement: `SELECT c.*
   FROM VARCHAR_TBL c
   WHERE c.f1 <> 'a';`,
				Results: []sql.Row{{`A`}, {1}, {2}, {3}, {``}, {`c`}},
			},
			{
				Statement: `SELECT c.*
   FROM VARCHAR_TBL c
   WHERE c.f1 = 'a';`,
				Results: []sql.Row{{`a`}},
			},
			{
				Statement: `SELECT c.*
   FROM VARCHAR_TBL c
   WHERE c.f1 < 'a';`,
				Results: []sql.Row{{``}},
			},
			{
				Statement: `SELECT c.*
   FROM VARCHAR_TBL c
   WHERE c.f1 <= 'a';`,
				Results: []sql.Row{{`a`}, {``}},
			},
			{
				Statement: `SELECT c.*
   FROM VARCHAR_TBL c
   WHERE c.f1 > 'a';`,
				Results: []sql.Row{{`A`}, {1}, {2}, {3}, {`c`}},
			},
			{
				Statement: `SELECT c.*
   FROM VARCHAR_TBL c
   WHERE c.f1 >= 'a';`,
				Results: []sql.Row{{`a`}, {`A`}, {1}, {2}, {3}, {`c`}},
			},
			{
				Statement: `DROP TABLE VARCHAR_TBL;`,
			},
			{
				Statement:   `INSERT INTO VARCHAR_TBL (f1) VALUES ('abcde');`,
				ErrorString: `value too long for type character varying(4)`,
			},
			{
				Statement: `SELECT * FROM VARCHAR_TBL;`,
				Results:   []sql.Row{{`a`}, {`ab`}, {`abcd`}, {`abcd`}},
			},
		},
	})
}
