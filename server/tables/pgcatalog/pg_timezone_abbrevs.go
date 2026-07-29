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

package pgcatalog

import (
	"io"
	"time"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/duration"
	"github.com/dolthub/doltgresql/server/tables"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// PgTimezoneAbbrevsName is a constant to the pg_timezone_abbrevs name.
const PgTimezoneAbbrevsName = "pg_timezone_abbrevs"

// InitPgTimezoneAbbrevs handles registration of the pg_timezone_abbrevs handler.
func InitPgTimezoneAbbrevs() {
	tables.AddHandler(PgCatalogName, PgTimezoneAbbrevsName, PgTimezoneAbbrevsHandler{})
}

// PgTimezoneAbbrevsHandler is the handler for the pg_timezone_abbrevs table.
type PgTimezoneAbbrevsHandler struct{}

var _ tables.Handler = PgTimezoneAbbrevsHandler{}

// Name implements the interface tables.Handler.
func (p PgTimezoneAbbrevsHandler) Name() string {
	return PgTimezoneAbbrevsName
}

// RowIter implements the interface tables.Handler.
func (p PgTimezoneAbbrevsHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	return &pgTimezoneAbbrevsRowIter{
		abbrevs: defaultTimezoneAbbrevs,
		idx:     0,
	}, nil
}

// PkSchema implements the interface tables.Handler.
func (p PgTimezoneAbbrevsHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgTimezoneAbbrevsSchema,
		PkOrdinals: nil,
	}
}

// pgTimezoneAbbrevsSchema is the schema for pg_timezone_abbrevs.
var pgTimezoneAbbrevsSchema = sql.Schema{
	{Name: "abbrev", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgTimezoneAbbrevsName},
	{Name: "utc_offset", Type: pgtypes.Interval, Default: nil, Nullable: true, Source: PgTimezoneAbbrevsName},
	{Name: "is_dst", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgTimezoneAbbrevsName},
}

// pgTimezoneAbbrevsRowIter is the sql.RowIter for the pg_timezone_abbrevs table.
type pgTimezoneAbbrevsRowIter struct {
	abbrevs []timezoneAbbrev
	idx     int
}

var _ sql.RowIter = (*pgTimezoneAbbrevsRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgTimezoneAbbrevsRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.abbrevs) {
		return nil, io.EOF
	}
	iter.idx++
	abbrev := iter.abbrevs[iter.idx-1]

	return sql.Row{
		abbrev.abbrev, // abbrev
		duration.MakeDuration(abbrev.offsetSecs*int64(time.Second), 0, 0), // utc_offset
		abbrev.isDST, // is_dst
	}, nil
}

// Close implements the interface sql.RowIter.
func (iter *pgTimezoneAbbrevsRowIter) Close(ctx *sql.Context) error {
	return nil
}

// timezoneAbbrev is a single timezone abbreviation, with its UTC offset in seconds and whether it
// is a daylight-savings abbreviation.
type timezoneAbbrev struct {
	abbrev     string
	offsetSecs int64
	isDST      bool
}

// defaultTimezoneAbbrevs is the list of timezone abbreviations reported by pg_timezone_abbrevs,
// sorted by abbreviation. This is a curated subset of Postgres's "Default" tznames set.
// TODO: this is a subset of the ~200 abbreviations in Postgres's default tznames set; fill in the
// rest (and support the timezone_abbreviations setting) if the full set is ever needed.
var defaultTimezoneAbbrevs = []timezoneAbbrev{
	{abbrev: "ACDT", offsetSecs: 37800, isDST: true},   // Australian Central Daylight Time
	{abbrev: "ACST", offsetSecs: 34200, isDST: false},  // Australian Central Standard Time
	{abbrev: "ADT", offsetSecs: -10800, isDST: true},   // Atlantic Daylight Time
	{abbrev: "AEDT", offsetSecs: 39600, isDST: true},   // Australian Eastern Daylight Time
	{abbrev: "AEST", offsetSecs: 36000, isDST: false},  // Australian Eastern Standard Time
	{abbrev: "AKDT", offsetSecs: -28800, isDST: true},  // Alaska Daylight Time
	{abbrev: "AKST", offsetSecs: -32400, isDST: false}, // Alaska Standard Time
	{abbrev: "AST", offsetSecs: -14400, isDST: false},  // Atlantic Standard Time
	{abbrev: "AWST", offsetSecs: 28800, isDST: false},  // Australian Western Standard Time
	{abbrev: "BST", offsetSecs: 3600, isDST: true},     // British Summer Time
	{abbrev: "CDT", offsetSecs: -18000, isDST: true},   // Central Daylight Time
	{abbrev: "CEST", offsetSecs: 7200, isDST: true},    // Central European Summer Time
	{abbrev: "CET", offsetSecs: 3600, isDST: false},    // Central European Time
	{abbrev: "CST", offsetSecs: -21600, isDST: false},  // Central Standard Time
	{abbrev: "EAT", offsetSecs: 10800, isDST: false},   // East Africa Time
	{abbrev: "EDT", offsetSecs: -14400, isDST: true},   // Eastern Daylight Time
	{abbrev: "EEST", offsetSecs: 10800, isDST: true},   // Eastern European Summer Time
	{abbrev: "EET", offsetSecs: 7200, isDST: false},    // Eastern European Time
	{abbrev: "EST", offsetSecs: -18000, isDST: false},  // Eastern Standard Time
	{abbrev: "GMT", offsetSecs: 0, isDST: false},       // Greenwich Mean Time
	{abbrev: "HKT", offsetSecs: 28800, isDST: false},   // Hong Kong Time
	{abbrev: "HST", offsetSecs: -36000, isDST: false},  // Hawaii Standard Time
	{abbrev: "IST", offsetSecs: 7200, isDST: false},    // Israel Standard Time (per Postgres default tznames)
	{abbrev: "JST", offsetSecs: 32400, isDST: false},   // Japan Standard Time
	{abbrev: "KST", offsetSecs: 32400, isDST: false},   // Korea Standard Time
	{abbrev: "MDT", offsetSecs: -21600, isDST: true},   // Mountain Daylight Time
	{abbrev: "MEST", offsetSecs: 7200, isDST: true},    // Middle Europe Summer Time
	{abbrev: "MET", offsetSecs: 3600, isDST: false},    // Middle Europe Time
	{abbrev: "MSK", offsetSecs: 10800, isDST: false},   // Moscow Time
	{abbrev: "MST", offsetSecs: -25200, isDST: false},  // Mountain Standard Time
	{abbrev: "NDT", offsetSecs: -9000, isDST: true},    // Newfoundland Daylight Time
	{abbrev: "NST", offsetSecs: -12600, isDST: false},  // Newfoundland Standard Time
	{abbrev: "NZDT", offsetSecs: 46800, isDST: true},   // New Zealand Daylight Time
	{abbrev: "NZST", offsetSecs: 43200, isDST: false},  // New Zealand Standard Time
	{abbrev: "PDT", offsetSecs: -25200, isDST: true},   // Pacific Daylight Time
	{abbrev: "PKT", offsetSecs: 18000, isDST: false},   // Pakistan Time
	{abbrev: "PST", offsetSecs: -28800, isDST: false},  // Pacific Standard Time
	{abbrev: "SAST", offsetSecs: 7200, isDST: false},   // South Africa Standard Time
	{abbrev: "SGT", offsetSecs: 28800, isDST: false},   // Singapore Time
	{abbrev: "UCT", offsetSecs: 0, isDST: false},       // Universal Coordinated Time
	{abbrev: "UT", offsetSecs: 0, isDST: false},        // Universal Time
	{abbrev: "UTC", offsetSecs: 0, isDST: false},       // Coordinated Universal Time
	{abbrev: "WAT", offsetSecs: 3600, isDST: false},    // West Africa Time
	{abbrev: "WEST", offsetSecs: 3600, isDST: true},    // Western European Summer Time
	{abbrev: "WET", offsetSecs: 0, isDST: false},       // Western European Time
	{abbrev: "Z", offsetSecs: 0, isDST: false},         // Zulu
	{abbrev: "ZULU", offsetSecs: 0, isDST: false},      // Zulu
}
