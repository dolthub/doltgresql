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
	"crypto/md5"
	"fmt"
	"strings"
	"testing"

	"github.com/dolthub/go-mysql-server/enginetest/scriptgen/setup"
	"github.com/jackc/pgx/v5/pgtype"

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

func makeTestVarbit(sizeInBits int) pgtype.Bits {
	numBytes := sizeInBits / 8
	leftoverBits := sizeInBits % 8
	bytes := make([]byte, numBytes)
	for i := 0; i < numBytes; i++ {
		bytes[i] = 0xaa
	}

	// this is a little annoying, but if we have leftover bits, we need to pad out the remaining bits in the last byte with 0s
	endingByte := byte(0)
	for i := 0; i < leftoverBits/2; i++ {
		endingByte |= 0b10 << (6 - i*2)
	}
	if leftoverBits > 0 {
		bytes = append(bytes, endingByte)
	}

	return pgtype.Bits{
		Bytes: bytes,
		Len:   int32(sizeInBits),
		Valid: true,
	}
}

// A 4000 byte file starting with 0x01 and then consisting of all zeros.
// This is larger than default target tuple size for outlining adaptive types.
// We expect a tuple to always store this value out-of-band
var fullSizeString = string(makeTestBytes(4000, 1))

// A 2000 byte file starting with 0x02 and then consisting of all zeros.
// This is over half of the default target tuple size for outlining adaptive types.
// We expect a tuple to be able to store this value inline once, but not twice.
var halfSizeString = string(makeTestBytes(2000, 2))

// A 10 byte file starting with 0x03 and then consisting of 10 zero bytes.
// This is file is smaller than an address hash.
// We expect a tuple to never store this value out-of-band.
var tinyString = string(makeTestBytes(10, 3))

// A 4000 byte file starting with ascii byte 1 and then consisting of all zeros.
// This is larger than default target tuple size for outlining adaptive types.
// We expect a tuple to always store this value out-of-band
var fullSizeVarbit = makeTestVarbit(4000)

// A 2000 byte file starting with ascii byte 1 and then consisting of all zeros.
// This is over half of the default target tuple size for outlining adaptive types.
// We expect a tuple to be able to store this value inline once, but not twice.
var halfSizeVarbit = makeTestVarbit(2000)

// A 10 byte file starting with ascii byte 1 and then consisting of 10 zero bytes.
// This is file is smaller than an address hash.
// We expect a tuple to never store this value out-of-band.
var tinyVarbit = makeTestVarbit(10)

func TestAdaptiveEncodingText(t *testing.T) {
	fullSizeOutOfLineRepr := fullSizeString
	columnTypes := []string{"varchar", "text"}
	for _, columnType := range columnTypes {
		t.Run(columnType, func(t *testing.T) {
			RunScripts(t, []ScriptTest{
				{
					Name: "Adaptive Encoding With One Column",
					SetUpScript: setup.SetupScript{
						fmt.Sprintf(`create table blobt (i char(1) primary key, b %s);`, columnType),
						fmt.Sprintf(`create table blobt2 (i char(2) primary key, b1 %s, b2 %s);`, columnType, columnType),
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
			})
		})
	}
}

// TestStringFunctionsOnOutOfBandValues checks that functions, operators, and casts taking string arguments work when
// the argument is a large value stored out-of-band (represented in memory as a wrapper such as *val.TextStorage,
// rather than a string).
func TestStringFunctionsOnOutOfBandValues(t *testing.T) {
	bigString := strings.Repeat("x", 20000)
	bigStringMd5 := fmt.Sprintf("%x", md5.Sum([]byte(bigString)))
	for _, columnType := range []string{"varchar", "text"} {
		t.Run(columnType, func(t *testing.T) {
			RunScripts(t, []ScriptTest{
				{
					Name: "string functions on out-of-band values",
					SetUpScript: setup.SetupScript{
						fmt.Sprintf(`create table t_big (id int primary key, body %s);`, columnType),
						`insert into t_big values (1, repeat('x', 20000));`,
					},
					Assertions: []ScriptTestAssertion{
						{
							Query:    "select length(body) from t_big;",
							Expected: []sql.Row{{int32(20000)}},
						},
						{
							Query:    "select length(replace(body, 'x', 'y')) from t_big;",
							Expected: []sql.Row{{int32(20000)}},
						},
						{
							Query:    "select length(upper(body)), length(lower(body)) from t_big;",
							Expected: []sql.Row{{int32(20000), int32(20000)}},
						},
						{
							Query:    "select left(body, 3), right(body, 3) from t_big;",
							Expected: []sql.Row{{"xxx", "xxx"}},
						},
						{
							Query:    "select ascii(body) from t_big;",
							Expected: []sql.Row{{int32(120)}},
						},
						{
							Query:    "select strpos(body, 'x') from t_big;",
							Expected: []sql.Row{{int32(1)}},
						},
						{
							Query:    "select length(reverse(body)) from t_big;",
							Expected: []sql.Row{{int32(20000)}},
						},
						{
							Query:    "select length(translate(body, 'x', 'z')) from t_big;",
							Expected: []sql.Row{{int32(20000)}},
						},
						{
							Query:    "select length(ltrim(body, 'x')), length(rtrim(body, 'x')), length(btrim(body, 'x')) from t_big;",
							Expected: []sql.Row{{int32(0), int32(0), int32(0)}},
						},
						{
							Query:    "select length(split_part(body, ',', 1)) from t_big;",
							Expected: []sql.Row{{int32(20000)}},
						},
						{
							Query:    "select length(repeat(body, 1)) from t_big;",
							Expected: []sql.Row{{int32(20000)}},
						},
						{
							Query:    "select md5(body) from t_big;",
							Expected: []sql.Row{{bigStringMd5}},
						},
						{
							Query:    "select octet_length(body) from t_big;",
							Expected: []sql.Row{{int32(20000)}},
						},
						{
							Query:    "select initcap(left(body, 3)) from t_big;",
							Expected: []sql.Row{{"Xxx"}},
						},
						{
							Query:    "select body::varchar(5) from t_big;",
							Expected: []sql.Row{{"xxxxx"}},
						},
						{
							Query:    "select length(body || 'y') from t_big;",
							Expected: []sql.Row{{int32(20001)}},
						},
						{
							Query:    "select body = repeat('x', 20000) from t_big;",
							Expected: []sql.Row{{"t"}},
						},
						{
							Query:    "select body < repeat('y', 5) from t_big;",
							Expected: []sql.Row{{"t"}},
						},
					},
				},
			})
		})
	}
}

// TestByteaFunctionsOnOutOfBandValues checks that functions, operators, and casts taking bytea arguments work when
// the argument is a large value stored out-of-band (represented in memory as a wrapper such as *val.ByteArray,
// rather than a []byte).
func TestByteaFunctionsOnOutOfBandValues(t *testing.T) {
	bigBytes := []byte(strings.Repeat("x", 20000))
	bigBytesHexLiteral := "'\\x" + strings.Repeat("78", 20000) + "'"
	RunScripts(t, []ScriptTest{
		{
			Name: "bytea functions on out-of-band values",
			SetUpScript: setup.SetupScript{
				`create table t_bytes (id int primary key, body bytea);`,
				"insert into t_bytes values (1, " + bigBytesHexLiteral + ");",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "select body from t_bytes;",
					Expected: []sql.Row{{bigBytes}},
				},
				{
					Query:    "select length(encode(body, 'hex')) from t_bytes;",
					Expected: []sql.Row{{int32(40000)}},
				},
				{
					Query:    "select left(encode(body, 'hex'), 4) from t_bytes;",
					Expected: []sql.Row{{"7878"}},
				},
				{
					Query:    "select length(encode(body || '\\x79', 'hex')) from t_bytes;",
					Expected: []sql.Row{{int32(40002)}},
				},
				{
					Query:    "select body = " + bigBytesHexLiteral + " from t_bytes;",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "select body <= '\\x79', body <> '\\x78', body >= '\\x78' from t_bytes;",
					Expected: []sql.Row{{"t", "t", "t"}},
				},
				{
					Query:    "select byteacmp(body, '\\x79') from t_bytes;",
					Expected: []sql.Row{{int32(-1)}},
				},
			},
		},
	})
}

func TestAdaptiveEncodingVarbit(t *testing.T) {
	columnType := "varbit"
	RunScripts(t, []ScriptTest{
		{
			Name: "Adaptive Encoding With One Column",
			SetUpScript: setup.SetupScript{
				fmt.Sprintf(`create table blobt (i char(1) primary key, b %s);`, columnType),
				fmt.Sprintf(`create table blobt2 (i char(2) primary key, b1 %s, b2 %s);`, columnType, columnType),
				fmt.Sprintf(`insert into blobt values
    ('F', LOAD_FILE('testdata/fullSizeVarbit')::%s),
    ('H', LOAD_FILE('testdata/halfSizeVarbit')::%s),
    ('T', LOAD_FILE('testdata/tinyFileVarbit')::%s)`, columnType, columnType, columnType),
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "select b from blobt where i = 'F'",
					Expected: []sql.Row{{fullSizeVarbit}},
				},
				{
					// Files that can fit within a tuple are stored inline.
					Query:    "select b from blobt where i = 'H'",
					Expected: []sql.Row{{halfSizeVarbit}},
				},
				{
					// An inlined adaptive column can be used in a filter.
					Query:    "select i from blobt where b = LOAD_FILE('testdata/fullSizeVarbit')::varbit",
					Expected: []sql.Row{{"F"}},
				},
				{
					// An out-of-line adaptive column can be used in a filter.
					Query:    "select i from blobt where b = LOAD_FILE('testdata/halfSizeVarbit')::varbit",
					Expected: []sql.Row{{"H"}},
				},
			},
		},
		{
			Name: "Adaptive Encoding With Two Columns",
			SetUpScript: setup.SetupScript{
				fmt.Sprintf(`create table blobt2 (i char(2) primary key, b1 %s, b2 %s);`, columnType, columnType),
				fmt.Sprintf(`insert into blobt2 values
    ('FF', LOAD_FILE('testdata/fullSizeVarbit')::%s, LOAD_FILE('testdata/fullSizeVarbit')::%s),
    ('HF', LOAD_FILE('testdata/halfSizeVarbit')::%s, LOAD_FILE('testdata/fullSizeVarbit')::%s),
    ('TF', LOAD_FILE('testdata/tinyFileVarbit')::%s, LOAD_FILE('testdata/fullSizeVarbit')::%s),
	('FH', LOAD_FILE('testdata/fullSizeVarbit')::%s, LOAD_FILE('testdata/halfSizeVarbit')::%s),
	('HH', LOAD_FILE('testdata/halfSizeVarbit')::%s, LOAD_FILE('testdata/halfSizeVarbit')::%s),
	('TH', LOAD_FILE('testdata/tinyFileVarbit')::%s, LOAD_FILE('testdata/halfSizeVarbit')::%s),
    ('FT', LOAD_FILE('testdata/fullSizeVarbit')::%s, LOAD_FILE('testdata/tinyFileVarbit')::%s),
    ('HT', LOAD_FILE('testdata/halfSizeVarbit')::%s, LOAD_FILE('testdata/tinyFileVarbit')::%s),
    ('TT', LOAD_FILE('testdata/tinyFileVarbit')::%s, LOAD_FILE('testdata/tinyFileVarbit')::%s)`, columnType, columnType,
					columnType, columnType, columnType, columnType, columnType, columnType, columnType, columnType,
					columnType, columnType, columnType, columnType, columnType, columnType, columnType, columnType),
			},
			Assertions: []ScriptTestAssertion{
				{
					// When a tuple with multiple adaptive columns is too large, columns are moved out-of-band from left to right.
					// However, strings smaller than the address size (20 bytes) are never stored out-of-band.
					Query: "select i, b1, b2 from blobt2 order by 1",
					Expected: []sql.Row{
						{"FF", fullSizeVarbit, fullSizeVarbit},
						{"FH", fullSizeVarbit, halfSizeVarbit},
						{"FT", fullSizeVarbit, tinyVarbit},
						{"HF", halfSizeVarbit, fullSizeVarbit},
						{"HH", halfSizeVarbit, halfSizeVarbit},
						{"HT", halfSizeVarbit, tinyVarbit},
						{"TF", tinyVarbit, fullSizeVarbit},
						{"TH", tinyVarbit, halfSizeVarbit},
						{"TT", tinyVarbit, tinyVarbit},
					},
				},
				{
					// An adaptive column can be used in a filter when it doesn't have the same encoding in all rows.
					Query:    "select i from blobt2 where b1 = LOAD_FILE('testdata/halfSizeVarbit')::varbit",
					Expected: []sql.Row{{"HF"}, {"HH"}, {"HT"}},
				},
				{
					// An adaptive column can be used in a filter when it doesn't have the same encoding in all rows.
					Query:    "select i from blobt2 where b2 = LOAD_FILE('testdata/halfSizeVarbit')::varbit",
					Expected: []sql.Row{{"FH"}, {"HH"}, {"TH"}},
				},
			},
		},
	})
}
