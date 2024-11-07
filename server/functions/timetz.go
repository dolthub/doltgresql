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

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/postgres/parser/timetz"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initTimeTZ registers the functions to the catalog.
func initTimeTZ() {
	framework.RegisterFunction(timetz_in)
	framework.RegisterFunction(timetz_out)
	framework.RegisterFunction(timetz_recv)
	framework.RegisterFunction(timetz_send)
	framework.RegisterFunction(timetztypmodin)
	framework.RegisterFunction(timetztypmodout)
	framework.RegisterFunction(timetz_cmp)
}

// timetz_in represents the PostgreSQL function of timetz type IO input.
var timetz_in = framework.Function3{
	Name:       "timetz_in",
	Return:     pgtypes.TimeTZ,
	Parameters: [3]pgtypes.DoltgresType{pgtypes.Text, pgtypes.Oid, pgtypes.Int32}, // cstring
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		input := val1.(string)
		//oid := val2.(uint32)
		//typmod := val3.(int32)
		// TODO: decode typmod to precision
		p := 6
		//if b.Precision == -1 {
		//	p = b.Precision
		//}
		loc, err := GetServerLocation(ctx)
		if err != nil {
			return nil, err
		}
		t, _, err := timetz.ParseTimeTZ(time.Now().In(loc), input, tree.TimeFamilyPrecisionToRoundDuration(int32(p)))
		if err != nil {
			return nil, err
		}
		return t.ToTime(), nil
	},
}

// timetz_out represents the PostgreSQL function of timetz type IO output.
var timetz_out = framework.Function1{
	Name:       "timetz_out",
	Return:     pgtypes.Text, // cstring
	Parameters: [1]pgtypes.DoltgresType{pgtypes.TimeTZ},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		// TODO: this always displays the time with an offset relevant to the server location
		return timetz.MakeTimeTZFromTime(val.(time.Time)).String(), nil
	},
}

// timetz_recv represents the PostgreSQL function of timetz type IO receive.
var timetz_recv = framework.Function3{
	Name:       "timetz_recv",
	Return:     pgtypes.TimeTZ,
	Parameters: [3]pgtypes.DoltgresType{pgtypes.Internal, pgtypes.Oid, pgtypes.Int32}, // cstring
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		data := val1.([]byte)
		//oid := val2.(uint32)
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

// timetz_send represents the PostgreSQL function of timetz type IO send.
var timetz_send = framework.Function1{
	Name:       "timetz_send",
	Return:     pgtypes.Bytea,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.TimeTZ},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		return val.(time.Time).MarshalBinary()
	},
}

// timetztypmodin represents the PostgreSQL function of timetz type IO typmod input.
var timetztypmodin = framework.Function1{
	Name:       "timetztypmodin",
	Return:     pgtypes.Int32,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.TextArray}, // cstring[]
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		// TODO: typmod=(precision<<16)∣scale
		return nil, nil
	},
}

// timetztypmodout represents the PostgreSQL function of timetz type IO typmod output.
var timetztypmodout = framework.Function1{
	Name:       "timetztypmodout",
	Return:     pgtypes.Text, // cstring
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		// TODO
		// Precision = typmod & 0xFFFF
		// Scale = (typmod >> 16) & 0xFFFF
		return nil, nil
	},
}

// timetz_cmp represents the PostgreSQL function of timetz type compare.
var timetz_cmp = framework.Function2{
	Name:       "timetz_cmp",
	Return:     pgtypes.Int32,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.TimeTZ, pgtypes.TimeTZ},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1, val2 any) (any, error) {
		ab := val1.(time.Time)
		bb := val2.(time.Time)
		return int32(ab.Compare(bb)), nil
	},
}

// GetServerLocation returns timezone value set for the server.
func GetServerLocation(ctx *sql.Context) (*time.Location, error) {
	if ctx == nil {
		return time.Local, nil
	}
	val, err := ctx.GetSessionVariable(ctx, "timezone")
	if err != nil {
		return nil, err
	}

	tz := val.(string)
	loc, err := time.LoadLocation(tz)
	if err == nil {
		return loc, nil
	}

	var t time.Time
	if t, err = time.Parse("Z07", tz); err == nil {
	} else if t, err = time.Parse("Z07:00", tz); err == nil {
	} else if t, err = time.Parse("Z07:00:00", tz); err != nil {
		return nil, err
	}

	_, offsetSecsUnconverted := t.Zone()
	return time.FixedZone(fmt.Sprintf("fixed offset:%d", offsetSecsUnconverted), -offsetSecsUnconverted), nil
}
