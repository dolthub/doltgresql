// Copyright 2025 Dolthub, Inc.
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

package _go

import (
	"fmt"
	"testing"

	"github.com/dolthub/go-mysql-server/enginetest/scriptgen/setup"

	"github.com/dolthub/go-mysql-server/sql"
)

// This is a correctness test based on the AdaptiveEncodingTest in Dolt.
// Because Doltgres serializes its results over a Postgres connection, we can't inspect the encoding
// of the result row. But we can still confirm that using adaptable encoding doesn't change any expected behavior.

func makeTestBytes(size int, firstbyte byte) []byte {
	bytes := make([]byte, size)
	bytes[0] = firstbyte
	return bytes
}

// A 4000 byte file starting with 0x01 and then consisting of all zeros.
// This is larger than default target tuple size for outlining adaptive types.
// We expect a tuple to always store this value out-of-band
// The same value is inserted via LOAD_FILE('testdata/fullSize')
var fullSizeString = string(makeTestBytes(4000, 1))

// A 2000 byte file starting with 0x02 and then consisting of all zeros.
// This is over half of the default target tuple size for outlining adaptive types.
// We expect a tuple to be able to store this value inline once, but not twice.
// The same value is inserted via LOAD_FILE('testdata/halfSize')
var halfSizeString = string(makeTestBytes(2000, 2))

// A 10 byte file starting with 0x03 and then consisting of 10 zero bytes.
// This is file is smaller than an address hash.
// We expect a tuple to never store this value out-of-band.
// The same value is inserted via LOAD_FILE('testdata/tinyFile')
var tinyString = string(makeTestBytes(10, 3))

// A 72K byte file starting with 0x04 and then consisting of all zeros.
// This is larger than the max tuple size. We should be able to write this without issues, and
// we expect a tuple to always store this value out-of-band.
// The same value is inserted via LOAD_FILE('testdata/tooBigFile')

// var tooBigString = string(makeTestBytes(72000, 4))

func TestAdaptiveEncoding(t *testing.T) {
	columnType := "text"
	fullSizeOutOfLineRepr := fullSizeString
	RunScripts(t, []ScriptTest{
		{
			Name: "Adaptive Encoding With One Column",
			SetUpScript: setup.SetupScript{
				fmt.Sprintf(`create table blobt (i char(1) primary key, b %s);`, columnType),
				`insert into blobt values
    ('F', LOAD_FILE('testdata/fullSize')),
    ('H', LOAD_FILE('testdata/halfSize')),
    ('T', LOAD_FILE('testdata/tinyFile'))`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "select b from blobt where i = 'F'",
					Expected: []sql.Row{{fullSizeString}},
				},
				{
					// Files that can fit within a tuple are stored inline.
					Query:    "select b from blobt where i = 'H'",
					Expected: []sql.Row{{halfSizeString}},
				},
				{
					// An inlined adaptive column can be used in a filter.
					Query:    "select i from blobt where b = LOAD_FILE('testdata/fullSize')",
					Expected: []sql.Row{{"F"}},
				},
				{
					// An out-of-line adaptive column can be used in a filter.
					Query:    "select i from blobt where b = LOAD_FILE('testdata/halfSize')",
					Expected: []sql.Row{{"H"}},
				},
			},
		},
		{
			Name: "Adaptive Encoding With Two Columns",
			SetUpScript: setup.SetupScript{
				fmt.Sprintf(`create table blobt2 (i char(2) primary key, b1 %s, b2 %s);`, columnType, columnType),
				`insert into blobt2 values
    ('FF', LOAD_FILE('testdata/fullSize'), LOAD_FILE('testdata/fullSize')),
    ('HF', LOAD_FILE('testdata/halfSize'), LOAD_FILE('testdata/fullSize')),
    ('TF', LOAD_FILE('testdata/tinyFile'), LOAD_FILE('testdata/fullSize')),
	('FH', LOAD_FILE('testdata/fullSize'), LOAD_FILE('testdata/halfSize')),
	('HH', LOAD_FILE('testdata/halfSize'), LOAD_FILE('testdata/halfSize')),
	('TH', LOAD_FILE('testdata/tinyFile'), LOAD_FILE('testdata/halfSize')),
    ('FT', LOAD_FILE('testdata/fullSize'), LOAD_FILE('testdata/tinyFile')),
    ('HT', LOAD_FILE('testdata/halfSize'), LOAD_FILE('testdata/tinyFile')),
    ('TT', LOAD_FILE('testdata/tinyFile'), LOAD_FILE('testdata/tinyFile'))`,
			},
			Assertions: []ScriptTestAssertion{
				{
					// When a tuple with multiple adaptive columns is too large, columns are moved out-of-band from left to right.
					// However, strings smaller than the address size (20 bytes) are never stored out-of-band.
					Query: "select i, b1, b2 from blobt2",
					Expected: []sql.Row{
						{"FF", fullSizeString, fullSizeString},
						{"HF", halfSizeString, fullSizeString},
						{"TF", tinyString, fullSizeString},
						{"FH", fullSizeString, halfSizeString},
						{"HH", halfSizeString, halfSizeString},
						{"TH", tinyString, halfSizeString},
						{"FT", fullSizeString, tinyString},
						{"HT", halfSizeString, tinyString},
						{"TT", tinyString, tinyString},
					},
				},
				{
					// An adaptive column can be used in a filter when it doesn't have the same encoding in all rows.
					Query:    "select i from blobt2 where b1 = LOAD_FILE('testdata/halfSize')",
					Expected: []sql.Row{{"HF"}, {"HH"}, {"HT"}},
				},
				{
					// An adaptive column can be used in a filter when it doesn't have the same encoding in all rows.
					Query:    "select i from blobt2 where b2 = LOAD_FILE('testdata/halfSize')",
					Expected: []sql.Row{{"FH"}, {"HH"}, {"TH"}},
				},
				{
					// Test creating an index on an adaptive encoding column, matching against out-of-band values
					Query: "CREATE INDEX bidx ON blobt2 (b1)",
				},
				{
					Query: "select i, b1 FROM blobt2 WHERE b1 LIKE '\x01%'",
					Expected: []sql.Row{
						{"FF", fullSizeOutOfLineRepr},
						{"FH", fullSizeOutOfLineRepr},
						{"FT", fullSizeOutOfLineRepr},
					},
				},
				{
					// Test creating an index on an adaptive encoding column, matching against inline values
					Query: "CREATE INDEX bidx2 ON blobt2 (b2)",
				},
				{
					Query: "select i, b2 FROM blobt2 WHERE b2 LIKE '\x02%'",
					Expected: []sql.Row{
						{"FH", halfSizeString},
						{"HH", halfSizeString},
						{"TH", halfSizeString},
					},
				},
				{
					// Tuples containing adaptive columns should be independent of how the tuple was created.
					// And adaptive values are always outlined starting from the left.
					// This means that in a table with two adaptive columns where both columns were previously stored out-of line,
					// Decreasing the size of the second column may allow both columns to be stored inline.
					Query: "UPDATE blobt2 SET b2 = LOAD_FILE('testdata/tinyFile') WHERE i = 'HH'",
				},
				{
					Query:    "select i, b1, b2 from blobt2 where i = 'HH'",
					Expected: []sql.Row{{"HH", halfSizeString, tinyString}},
				},
				{
					// Similar to the above, dropping a column can change whether the other column is inlined.
					Query: "ALTER TABLE blobt2 DROP COLUMN b2",
				},
				{
					Query: "select i, b1 from blobt2",
					Expected: []sql.Row{
						{"FF", fullSizeString},
						{"HF", halfSizeString},
						{"TF", tinyString},
						{"FH", fullSizeString},
						{"HH", halfSizeString},
						{"TH", tinyString},
						{"FT", fullSizeString},
						{"HT", halfSizeString},
						{"TT", tinyString},
					},
				},
			},
		},
		{
			Name: "Adaptive Encoding With values > 64K is not truncated",
			SetUpScript: setup.SetupScript{
				fmt.Sprintf(`create table blobt (i char(1) primary key, b %s);`, columnType),
				`insert into blobt values
    ('F', LOAD_FILE('testdata/tooBigFile')),
    ('H', LOAD_FILE('testdata/halfSize')),
    ('T', LOAD_FILE('testdata/tinyFile'))`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "select LENGTH(b) from blobt where i = 'F'",
					Expected: []sql.Row{{72001}},
				},
				{
					// An out-of-line adaptive column can be used in a filter.
					Query:    "select i from blobt where b = LOAD_FILE('testdata/tooBigFile')",
					Expected: []sql.Row{{"F"}},
				},
			},
		},
		{
			Name: "Adaptive Extended values can be read and re-inserted",
			SetUpScript: setup.SetupScript{
				fmt.Sprintf(`create table blobt2 (i1 char(1), i2 char(1), primary key (i1, i2), b1 %s, b2 %s);`, columnType, columnType),
				`insert into blobt2 values
    ('F', 'F', LOAD_FILE('testdata/fullSize'), LOAD_FILE('testdata/fullSize')),
    ('H', 'H', LOAD_FILE('testdata/halfSize'), LOAD_FILE('testdata/halfSize')),
    ('T', 'T', LOAD_FILE('testdata/tinyFile'), LOAD_FILE('testdata/tinyFile'))`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "insert into blobt2 select one.i1, two.i2, one.b1, two.b2 from blobt2 AS one join blobt2 AS two on one.i1 != two.i2;",
				},
				{
					// Check that the out-of-band value halfSize is written correctly both when it fits in the new row inline, and when it doesn't.
					Query: "select * from blobt2",
					Expected: []sql.Row{
						{"F", "F", fullSizeString, fullSizeString},
						{"H", "F", halfSizeString, fullSizeString},
						{"T", "F", tinyString, fullSizeString},
						{"F", "H", fullSizeString, halfSizeString},
						{"H", "H", halfSizeString, halfSizeString},
						{"T", "H", tinyString, halfSizeString},
						{"F", "T", fullSizeString, tinyString},
						{"H", "T", halfSizeString, tinyString},
						{"T", "T", tinyString, tinyString},
					},
				},
			},
		},
	})
}
