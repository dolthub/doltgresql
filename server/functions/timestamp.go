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

package functions

import (
	"fmt"
	"strings"
	"time"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initTimestamp registers the functions to the catalog.
func initTimestamp() {
	framework.RegisterFunction(timestamp_in)
	framework.RegisterFunction(timestamp_out)
	framework.RegisterFunction(timestamp_recv)
	framework.RegisterFunction(timestamp_send)
	framework.RegisterFunction(timestamptypmodin)
	framework.RegisterFunction(timestamptypmodout)
	framework.RegisterFunction(timestamp_cmp)
}

// timestamp_in represents the PostgreSQL function of timestamp type IO input.
var timestamp_in = framework.Function3{
	Name:       "timestamp_in",
	Return:     pgtypes.Timestamp,
	Parameters: [3]*pgtypes.DoltgresType{pgtypes.Cstring, pgtypes.Oid, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		input := val1.(string)
		//oid := val2.(id.Id)
		//typmod := val3.(int32)
		// TODO: decode typmod to precision
		p := 6
		t, _, err := tree.ParseDTimestamp(nil, input, tree.TimeFamilyPrecisionToRoundDuration(int32(p)))
		if err != nil {
			return nil, err
		}
		return t.Time, nil
	},
}

// timestamp_out represents the PostgreSQL function of timestamp type IO output.
var timestamp_out = framework.Function1{
	Name:       "timestamp_out",
	Return:     pgtypes.Cstring,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Timestamp},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		t := val.(time.Time)
		return FormatDateTimeWithBC(t, getLayoutStringFormat(ctx, false), false), nil
	},
}

// timestamp_recv represents the PostgreSQL function of timestamp type IO receive.
var timestamp_recv = framework.Function3{
	Name:       "timestamp_recv",
	Return:     pgtypes.Timestamp,
	Parameters: [3]*pgtypes.DoltgresType{pgtypes.Internal, pgtypes.Oid, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		data := val1.([]byte)
		//oid := val2.(id.Id)
		//typmod := val3.(int32)
		// TODO: decode typmod to precision
		if len(data) == 0 {
			return nil, nil
		}
		t := time.Time{}
		if err := t.UnmarshalBinary(data); err != nil {
			return nil, err
		}
		return t, nil
	},
}

// timestamp_send represents the PostgreSQL function of timestamp type IO send.
var timestamp_send = framework.Function1{
	Name:       "timestamp_send",
	Return:     pgtypes.Bytea,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Timestamp},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		return val.(time.Time).MarshalBinary()
	},
}

// timestamptypmodin represents the PostgreSQL function of timestamp type IO typmod input.
var timestamptypmodin = framework.Function1{
	Name:       "timestamptypmodin",
	Return:     pgtypes.Int32,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.CstringArray},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		// TODO: typmod=(precision<<16)∣scale
		return nil, nil
	},
}

// timestamptypmodout represents the PostgreSQL function of timestamp type IO typmod output.
var timestamptypmodout = framework.Function1{
	Name:       "timestamptypmodout",
	Return:     pgtypes.Cstring,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		// TODO
		// Precision = typmod & 0xFFFF
		// Scale = (typmod >> 16) & 0xFFFF
		return nil, nil
	},
}

// timestamp_cmp represents the PostgreSQL function of timestamp type compare.
var timestamp_cmp = framework.Function2{
	Name:       "timestamp_cmp",
	Return:     pgtypes.Int32,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Timestamp, pgtypes.Timestamp},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		ab := val1.(time.Time)
		bb := val2.(time.Time)
		return int32(ab.Compare(bb)), nil
	},
}

// https://www.postgresql.org/docs/15/datatype-datetime.html
const (
	DateStyleISO      = "ISO"
	DateStyleSQL      = "SQL"
	DateStylePostgres = "Postgres"
	DateStyleGerman   = "German"
)

// getDateStyleOutputFormat returns the format layout for date/time values defined in the 'DateStyle' configuration parameter.
func getDateStyleOutputFormat(ctx *sql.Context) string {
	format := DateStyleISO // default
	if ctx == nil {
		return format
	}

	val, err := ctx.GetSessionVariable(ctx, "datestyle")
	if err != nil {
		return format
	}

	values := strings.Split(strings.ReplaceAll(val.(string), " ", ""), ",")
	for _, value := range values {
		switch value {
		case DateStyleISO, DateStyleSQL, DateStylePostgres, DateStyleGerman:
			return value
		}
	}
	return format
}

const (
	// ISO is ISO 8601, SQL standard
	dateStyleFormat_ISO         = "2006-01-02 15:04:05.999999"
	dateStyleFormatDateOnly_ISO = "2006-01-02"
	// SQL is traditional style
	dateStyleFormat_SQL         = "01/02/2006 15:04:05.999999"
	dateStyleFormatDateOnly_SQL = "01/02/2006"
	// Postgres is original style
	dateStyleFormat_Postgres         = "Mon Jan 02 15:04:05.999999 2006"
	dateStyleFormatDateOnly_Postgres = "Mon Jan 02 2006"
	// regional style
	dateStyleFormat_German         = "02.01.2006 15:04:05.999999"
	dateStyleFormatDateOnly_German = "02.01.2006"
)

func getLayoutStringFormat(ctx *sql.Context, dateOnly bool) string {
	layout := getDateStyleOutputFormat(ctx)
	switch layout {
	case DateStyleISO:
		if dateOnly {
			return dateStyleFormatDateOnly_ISO
		}
		return dateStyleFormat_ISO
	case DateStyleSQL:
		if dateOnly {
			return dateStyleFormatDateOnly_SQL
		}
		return dateStyleFormat_SQL
	case DateStylePostgres:
		if dateOnly {
			return dateStyleFormatDateOnly_Postgres
		}
		return dateStyleFormat_Postgres
	case DateStyleGerman:
		if dateOnly {
			return dateStyleFormatDateOnly_German
		}
		return dateStyleFormat_German
	}
	// shouldn't happen but return default
	if dateOnly {
		return dateStyleFormatDateOnly_ISO
	}
	return dateStyleFormat_ISO
}

// FormatDateTimeWithBC formats a time.Time that may represent BC dates (negative years)
// PostgreSQL represents BC years as negative years in time.Time, but Go's Format() doesn't handle this correctly
// tz is optional timezone value to be appended to formatted value
func FormatDateTimeWithBC(t time.Time, layout string, hasTZ bool) string {
	year := t.Year()
	isBC := year <= 0

	var formattedTime string
	if isBC {
		// Convert from PostgreSQL's BC representation to positive year for formatting
		// PostgreSQL: year 0 = 1 BC, year -1 = 2 BC, etc.
		absYear := 1 - year

		// Create a new time with the positive year for formatting
		positiveTime := time.Date(absYear, t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())

		// Convert the negative year to positive year to get formatted result, but it creates issue on
		// the day of week value, so we need to get day of the week value from original negative time value.
		if strings.HasPrefix(layout, "Mon") {
			formattedTime = fmt.Sprintf("%s%s", t.Format("Mon"), positiveTime.Format(strings.TrimPrefix(layout, "Mon")))
		} else {
			// Format with the positive year, then append " BC"
			formattedTime = positiveTime.Format(layout)
		}
	} else {
		// For AD years (positive), use normal formatting
		formattedTime = t.Format(layout)
	}

	if hasTZ {
		name, offset := t.Zone()
		if strings.HasPrefix(layout, "Mon") {
			// Postgres doesn't show timezone for ones that don't have timezone abbreviation.
			if name != "" {
				name = t.Format("MST")
			}
			formattedTime += fmt.Sprintf(" %s", name)
		} else {
			if offset%3600 != 0 {
				formattedTime += t.Format("-07:00")
			} else {
				formattedTime += t.Format("-07")
			}
		}
	}

	if isBC {
		return formattedTime + " BC"
	}
	return formattedTime
}
