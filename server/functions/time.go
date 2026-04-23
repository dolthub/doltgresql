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
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgtype"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/postgres/parser/timeofday"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/utils"
)

// initTime registers the functions to the catalog.
func initTime() {
	framework.RegisterFunction(time_in)
	framework.RegisterFunction(time_out)
	framework.RegisterFunction(time_recv)
	framework.RegisterFunction(time_send)
	framework.RegisterFunction(timetypmodin)
	framework.RegisterFunction(timetypmodout)
	framework.RegisterFunction(time_cmp)
}

// time_in represents the PostgreSQL function of time type IO input.
var time_in = framework.Function3{
	Name:       "time_in",
	Return:     pgtypes.Time,
	Parameters: [3]*pgtypes.DoltgresType{pgtypes.Cstring, pgtypes.Oid, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		input := strings.TrimSpace(val1.(string))
		typmod := val3.(int32)
		if typmod == -1 {
			typmod = 6
		}
		precision := tree.TimeFamilyPrecisionToRoundDuration(typmod)
		if strings.EqualFold(input, "now") {
			t := ctx.QueryTime()
			t = t.Round(precision)
			return timeofday.New(t.Hour(), t.Minute(), t.Second(), t.Nanosecond()/1000), nil
		}
		t, _, err := tree.ParseDTime(nil, input, precision)
		if err != nil {
			return nil, err
		}
		return timeofday.TimeOfDay(*t), nil
	},
}

// time_out represents the PostgreSQL function of time type IO output.
var time_out = framework.Function1{
	Name:       "time_out",
	Return:     pgtypes.Cstring,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Time},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		return val.(timeofday.TimeOfDay).String(), nil
	},
}

// time_recv represents the PostgreSQL function of time type IO receive.
var time_recv = framework.Function3{
	Name:       "time_recv",
	Return:     pgtypes.Time,
	Parameters: [3]*pgtypes.DoltgresType{pgtypes.Internal, pgtypes.Oid, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		data := val1.([]byte)
		if data == nil {
			return nil, nil
		}
		var out pgtype.Time
		err := out.DecodeBinary(nil, data)
		if err != nil {
			return nil, err
		}
		return timeofday.New(0, 0, 0, int(out.Microseconds)), nil
	},
}

// time_send represents the PostgreSQL function of time type IO send.
var time_send = framework.Function1{
	Name:       "time_send",
	Return:     pgtypes.Bytea,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Time},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		writer := utils.NewWireWriter()
		writer.WriteInt64(int64(val.(timeofday.TimeOfDay)))
		return writer.BufferData(), nil
	},
}

// timetypmodin represents the PostgreSQL function of time type IO typmod input.
var timetypmodin = framework.Function1{
	Name:       "timetypmodin",
	Return:     pgtypes.Int32,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.CstringArray},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		// TODO: typmod=(precision<<16)∣scale
		return nil, nil
	},
}

// timetypmodout represents the PostgreSQL function of time type IO typmod output.
var timetypmodout = framework.Function1{
	Name:       "timetypmodout",
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

// time_cmp represents the PostgreSQL function of time type compare.
var time_cmp = framework.Function2{
	Name:       "time_cmp",
	Return:     pgtypes.Int32,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Time, pgtypes.Time},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		ab := val1.(timeofday.TimeOfDay)
		bb := val2.(timeofday.TimeOfDay)
		return int32(ab.Compare(bb)), nil
	},
}
