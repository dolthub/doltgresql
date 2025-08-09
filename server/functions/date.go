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
	"time"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/pgdate"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initDate registers the functions to the catalog.
func initDate() {
	framework.RegisterFunction(date_in)
	framework.RegisterFunction(date_out)
	framework.RegisterFunction(date_recv)
	framework.RegisterFunction(date_send)
	framework.RegisterFunction(date_cmp)
}

// date_in represents the PostgreSQL function of date type IO input.
var date_in = framework.Function1{
	Name:       "date_in",
	Return:     pgtypes.Date,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Cstring},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		input := val.(string)
		if date, _, err := pgdate.ParseDate(time.Now(), pgdate.ParseModeYMD, input); err == nil {
			return date.ToTime()
		} else if date, _, err = pgdate.ParseDate(time.Now(), pgdate.ParseModeDMY, input); err == nil {
			return date.ToTime()
		} else if date, _, err = pgdate.ParseDate(time.Now(), pgdate.ParseModeMDY, input); err == nil {
			return date.ToTime()
		} else {
			return nil, err
		}
	},
}

// date_out represents the PostgreSQL function of date type IO output.
var date_out = framework.Function1{
	Name:       "date_out",
	Return:     pgtypes.Cstring,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Date},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		return formatDateWithBC(val.(time.Time)), nil
	},
}

// date_recv represents the PostgreSQL function of date type IO receive.
var date_recv = framework.Function1{
	Name:       "date_recv",
	Return:     pgtypes.Date,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Internal},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		data := val.([]byte)
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

// date_send represents the PostgreSQL function of date type IO send.
var date_send = framework.Function1{
	Name:       "date_send",
	Return:     pgtypes.Bytea,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Date},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		return val.(time.Time).MarshalBinary()
	},
}

// date_cmp represents the PostgreSQL function of date type compare.
var date_cmp = framework.Function2{
	Name:       "date_cmp",
	Return:     pgtypes.Int32,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Date, pgtypes.Date},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		ab := val1.(time.Time)
		bb := val2.(time.Time)
		return int32(ab.Compare(bb)), nil
	},
}

// formatDateWithBC formats a time.Time that may represent BC dates (negative years)
// PostgreSQL represents BC years as negative years in time.Time, but Go's Format() doesn't handle this correctly
func formatDateWithBC(t time.Time) string {
	year := t.Year()
	isBC := year <= 0

	layout := "2006-01-02"

	if isBC {
		// Convert from PostgreSQL's BC representation to positive year for formatting
		// PostgreSQL: year 0 = 1 BC, year -1 = 2 BC, etc.
		absYear := 1 - year

		// Create a new time with the positive year for formatting
		positiveTime := time.Date(absYear, t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)

		// Format with the positive year, then append " BC"
		formatted := positiveTime.Format(layout)
		return fmt.Sprintf("%s BC", formatted)
	} else {
		// For AD years (positive), use normal formatting
		return t.Format(layout)
	}
}
