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
	Return:     pgtypes.Text, // cstring
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Timestamp},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		return val.(time.Time).Format("2006-01-02 15:04:05.999999999"), nil
	},
}

// timestamp_recv represents the PostgreSQL function of timestamp type IO receive.
var timestamp_recv = framework.Function3{
	Name:       "timestamp_recv",
	Return:     pgtypes.Timestamp,
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

// timestamp_send represents the PostgreSQL function of timestamp type IO send.
var timestamp_send = framework.Function1{
	Name:       "timestamp_send",
	Return:     pgtypes.Bytea,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Timestamp},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		return val.(time.Time).MarshalBinary()
	},
}

// timestamptypmodin represents the PostgreSQL function of timestamp type IO typmod input.
var timestamptypmodin = framework.Function1{
	Name:       "timestamptypmodin",
	Return:     pgtypes.Int32,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.TextArray}, // cstring[]
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		// TODO: typmod=(precision<<16)∣scale
		return nil, nil
	},
}

// timestamptypmodout represents the PostgreSQL function of timestamp type IO typmod output.
var timestamptypmodout = framework.Function1{
	Name:       "timestamptypmodout",
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

// timestamp_cmp represents the PostgreSQL function of timestamp type compare.
var timestamp_cmp = framework.Function2{
	Name:       "timestamp_cmp",
	Return:     pgtypes.Int32,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Timestamp, pgtypes.Timestamp},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1, val2 any) (any, error) {
		ab := val1.(time.Time)
		bb := val2.(time.Time)
		return int32(ab.Compare(bb)), nil
	},
}
