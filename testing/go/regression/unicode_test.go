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

func TestUnicode(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_unicode)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_unicode,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `SELECT getdatabaseencoding() <> 'UTF8' AS skip_test \gset
\if :skip_test
\quit
\endif
SELECT U&'\0061\0308bc' <> U&'\00E4bc' COLLATE "C" AS sanity_check;`,
				Results: []sql.Row{{true}},
			},
			{
				Statement: `SELECT normalize('');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT normalize(U&'\0061\0308\24D1c') = U&'\00E4\24D1c' COLLATE "C" AS test_default;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT normalize(U&'\0061\0308\24D1c', NFC) = U&'\00E4\24D1c' COLLATE "C" AS test_nfc;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT normalize(U&'\00E4bc', NFC) = U&'\00E4bc' COLLATE "C" AS test_nfc_idem;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT normalize(U&'\00E4\24D1c', NFD) = U&'\0061\0308\24D1c' COLLATE "C" AS test_nfd;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT normalize(U&'\0061\0308\24D1c', NFKC) = U&'\00E4bc' COLLATE "C" AS test_nfkc;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT normalize(U&'\00E4\24D1c', NFKD) = U&'\0061\0308bc' COLLATE "C" AS test_nfkd;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement:   `SELECT "normalize"('abc', 'def');  -- run-time error`,
				ErrorString: `invalid normalization form: def`,
			},
			{
				Statement: `SELECT U&'\00E4\24D1c' IS NORMALIZED AS test_default;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT U&'\00E4\24D1c' IS NFC NORMALIZED AS test_nfc;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT num, val,
    val IS NFC NORMALIZED AS NFC,
    val IS NFD NORMALIZED AS NFD,
    val IS NFKC NORMALIZED AS NFKC,
    val IS NFKD NORMALIZED AS NFKD
FROM
  (VALUES (1, U&'\00E4bc'),
          (2, U&'\0061\0308bc'),
          (3, U&'\00E4\24D1c'),
          (4, U&'\0061\0308\24D1c'),
          (5, '')) vals (num, val)
ORDER BY num;`,
				Results: []sql.Row{{1, `äbc`, true, false, true, false}, {2, `äbc`, false, true, false, true}, {3, `äⓑc`, true, false, false, false}, {4, `äⓑc`, false, true, false, false}, {5, ``, true, true, true, true}},
			},
			{
				Statement:   `SELECT is_normalized('abc', 'def');  -- run-time error`,
				ErrorString: `invalid normalization form: def`,
			},
		},
	})
}
