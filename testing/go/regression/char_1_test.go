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

func TestChar1(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_char_1)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_char_1,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `SELECT char 'c' = char 'c' AS true;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `CREATE TEMP TABLE CHAR_TBL(f1 char);`,
			},
			{
				Statement: `INSERT INTO CHAR_TBL (f1) VALUES ('a');`,
			},
			{
				Statement: `INSERT INTO CHAR_TBL (f1) VALUES ('A');`,
			},
			{
				Statement: `INSERT INTO CHAR_TBL (f1) VALUES ('1');`,
			},
			{
				Statement: `INSERT INTO CHAR_TBL (f1) VALUES (2);`,
			},
			{
				Statement: `INSERT INTO CHAR_TBL (f1) VALUES ('3');`,
			},
			{
				Statement: `INSERT INTO CHAR_TBL (f1) VALUES ('');`,
			},
			{
				Statement:   `INSERT INTO CHAR_TBL (f1) VALUES ('cd');`,
				ErrorString: `value too long for type character(1)`,
			},
			{
				Statement: `INSERT INTO CHAR_TBL (f1) VALUES ('c     ');`,
			},
			{
				Statement: `SELECT * FROM CHAR_TBL;`,
				Results:   []sql.Row{{`a`}, {`A`}, {1}, {2}, {3}, {``}, {`c`}},
			},
			{
				Statement: `SELECT c.*
   FROM CHAR_TBL c
   WHERE c.f1 <> 'a';`,
				Results: []sql.Row{{`A`}, {1}, {2}, {3}, {``}, {`c`}},
			},
			{
				Statement: `SELECT c.*
   FROM CHAR_TBL c
   WHERE c.f1 = 'a';`,
				Results: []sql.Row{{`a`}},
			},
			{
				Statement: `SELECT c.*
   FROM CHAR_TBL c
   WHERE c.f1 < 'a';`,
				Results: []sql.Row{{1}, {2}, {3}, {``}},
			},
			{
				Statement: `SELECT c.*
   FROM CHAR_TBL c
   WHERE c.f1 <= 'a';`,
				Results: []sql.Row{{`a`}, {1}, {2}, {3}, {``}},
			},
			{
				Statement: `SELECT c.*
   FROM CHAR_TBL c
   WHERE c.f1 > 'a';`,
				Results: []sql.Row{{`A`}, {`c`}},
			},
			{
				Statement: `SELECT c.*
   FROM CHAR_TBL c
   WHERE c.f1 >= 'a';`,
				Results: []sql.Row{{`a`}, {`A`}, {`c`}},
			},
			{
				Statement: `DROP TABLE CHAR_TBL;`,
			},
			{
				Statement:   `INSERT INTO CHAR_TBL (f1) VALUES ('abcde');`,
				ErrorString: `value too long for type character(4)`,
			},
			{
				Statement: `SELECT * FROM CHAR_TBL;`,
				Results:   []sql.Row{{`a`}, {`ab`}, {`abcd`}, {`abcd`}},
			},
			{
				Statement: `SELECT 'a'::"char";`,
				Results:   []sql.Row{{`a`}},
			},
			{
				Statement: `SELECT '\101'::"char";`,
				Results:   []sql.Row{{`A`}},
			},
			{
				Statement: `SELECT '\377'::"char";`,
				Results:   []sql.Row{{`\377`}},
			},
			{
				Statement: `SELECT 'a'::"char"::text;`,
				Results:   []sql.Row{{`a`}},
			},
			{
				Statement: `SELECT '\377'::"char"::text;`,
				Results:   []sql.Row{{`\377`}},
			},
			{
				Statement: `SELECT '\000'::"char"::text;`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT 'a'::text::"char";`,
				Results:   []sql.Row{{`a`}},
			},
			{
				Statement: `SELECT '\377'::text::"char";`,
				Results:   []sql.Row{{`\377`}},
			},
			{
				Statement: `SELECT ''::text::"char";`,
				Results:   []sql.Row{{``}},
			},
		},
	})
}
