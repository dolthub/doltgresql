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
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Text}, // cstring
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
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
	Return:     pgtypes.Text, // cstring
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Date},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		return val.(time.Time).Format("2006-01-02"), nil
	},
}

// date_recv represents the PostgreSQL function of date type IO receive.
var date_recv = framework.Function1{
	Name:       "date_recv",
	Return:     pgtypes.Date,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Internal},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		switch v := val.(type) {
		case time.Time:
			return v, nil
		default:
			return nil, pgtypes.ErrUnhandledType.New("date", v)
		}
	},
}

// date_send represents the PostgreSQL function of date type IO send.
var date_send = framework.Function1{
	Name:       "date_send",
	Return:     pgtypes.Bytea,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Date},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		return []byte(val.(time.Time).Format("2006-01-02")), nil
	},
}

// date_cmp represents the PostgreSQL function of date type compare.
var date_cmp = framework.Function2{
	Name:       "date_cmp",
	Return:     pgtypes.Int32,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Date, pgtypes.Date},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1, val2 any) (any, error) {
		ab := val1.(time.Time)
		bb := val2.(time.Time)
		return int32(ab.Compare(bb)), nil
	},
}
