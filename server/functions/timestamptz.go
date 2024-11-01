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

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initTimestampTZ registers the functions to the catalog.
func initTimestampTZ() {
	framework.RegisterFunction(timestamptz_in)
	framework.RegisterFunction(timestamptz_out)
	framework.RegisterFunction(timestamptz_recv)
	framework.RegisterFunction(timestamptz_send)
	framework.RegisterFunction(timestamptztypmodin)
	framework.RegisterFunction(timestamptztypmodout)
	framework.RegisterFunction(timestamptz_cmp)
}

// timestamptz_in represents the PostgreSQL function of timestamptz type IO input.
var timestamptz_in = framework.Function3{
	Name:       "timestamptz_in",
	Return:     pgtypes.TimestampTZ,
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
		t, _, err := tree.ParseDTimestampTZ(nil, input, tree.TimeFamilyPrecisionToRoundDuration(int32(p)), loc)
		if err != nil {
			return nil, err
		}
		return t.Time, nil
	},
}

// timestamptz_out represents the PostgreSQL function of timestamptz type IO output.
var timestamptz_out = framework.Function1{
	Name:       "timestamptz_out",
	Return:     pgtypes.Text, // cstring
	Parameters: [1]pgtypes.DoltgresType{pgtypes.TimestampTZ},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		serverLoc, err := GetServerLocation(ctx)
		if err != nil {
			return "", err
		}
		t := val.(time.Time).In(serverLoc)
		_, offset := t.Zone()
		if offset%3600 != 0 {
			return t.Format("2006-01-02 15:04:05.999999999-07:00"), nil
		} else {
			return t.Format("2006-01-02 15:04:05.999999999-07"), nil
		}
	},
}

// timestamptz_recv represents the PostgreSQL function of timestamptz type IO receive.
var timestamptz_recv = framework.Function3{
	Name:       "timestamptz_recv",
	Return:     pgtypes.TimestampTZ,
	Parameters: [3]pgtypes.DoltgresType{pgtypes.Internal, pgtypes.Oid, pgtypes.Int32}, // cstring
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		// TODO
		switch val := val1.(type) {
		case time.Time:
			return val, nil
		default:
			return nil, pgtypes.ErrUnhandledType.New("timestamptz", val)
		}
	},
}

// timestamptz_send represents the PostgreSQL function of timestamptz type IO send.
var timestamptz_send = framework.Function1{
	Name:       "timestamptz_send",
	Return:     pgtypes.Bytea,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.TimestampTZ},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		serverLoc, err := GetServerLocation(ctx)
		if err != nil {
			return "", err
		}
		t := val.(time.Time).In(serverLoc)
		_, offset := t.Zone()
		if offset%3600 != 0 {
			return []byte(t.Format("2006-01-02 15:04:05.999999999-07:00")), nil
		} else {
			return []byte(t.Format("2006-01-02 15:04:05.999999999-07")), nil
		}
	},
}

// timestamptztypmodin represents the PostgreSQL function of timestamptz type IO typmod input.
var timestamptztypmodin = framework.Function1{
	Name:       "timestamptztypmodin",
	Return:     pgtypes.Int32,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.TextArray}, // cstring[]
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		// TODO: typmod=(precision<<16)∣scale
		return nil, nil
	},
}

// timestamptztypmodout represents the PostgreSQL function of timestamptz type IO typmod output.
var timestamptztypmodout = framework.Function1{
	Name:       "timestamptztypmodout",
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

// timestamptz_cmp represents the PostgreSQL function of timestamptz type compare.
var timestamptz_cmp = framework.Function2{
	Name:       "bttimestamptz_cmp",
	Return:     pgtypes.Int32,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Text, pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1, val2 any) (any, error) {
		ab := val1.(time.Time)
		bb := val2.(time.Time)
		return int32(ab.Compare(bb)), nil
	},
}
