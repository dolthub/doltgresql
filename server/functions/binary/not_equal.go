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

package binary

import (
	"time"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/postgres/parser/uuid"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// These functions can be gathered using the following query from a Postgres 15 instance:
// SELECT * FROM pg_operator o WHERE o.oprname = '<>' ORDER BY o.oprcode::varchar;

// initBinaryNotEqual registers the functions to the catalog.
func initBinaryNotEqual() {
	framework.RegisterBinaryFunction(framework.Operator_BinaryNotEqual, boolne)
	framework.RegisterBinaryFunction(framework.Operator_BinaryNotEqual, bpcharne)
	framework.RegisterBinaryFunction(framework.Operator_BinaryNotEqual, byteane)
	framework.RegisterBinaryFunction(framework.Operator_BinaryNotEqual, date_ne)
	framework.RegisterBinaryFunction(framework.Operator_BinaryNotEqual, date_ne_timestamp)
	framework.RegisterBinaryFunction(framework.Operator_BinaryNotEqual, date_ne_timestamptz)
	framework.RegisterBinaryFunction(framework.Operator_BinaryNotEqual, float4ne)
	framework.RegisterBinaryFunction(framework.Operator_BinaryNotEqual, float48ne)
	framework.RegisterBinaryFunction(framework.Operator_BinaryNotEqual, float84ne)
	framework.RegisterBinaryFunction(framework.Operator_BinaryNotEqual, float8ne)
	framework.RegisterBinaryFunction(framework.Operator_BinaryNotEqual, int2ne)
	framework.RegisterBinaryFunction(framework.Operator_BinaryNotEqual, int24ne)
	framework.RegisterBinaryFunction(framework.Operator_BinaryNotEqual, int28ne)
	framework.RegisterBinaryFunction(framework.Operator_BinaryNotEqual, int42ne)
	framework.RegisterBinaryFunction(framework.Operator_BinaryNotEqual, int4ne)
	framework.RegisterBinaryFunction(framework.Operator_BinaryNotEqual, int48ne)
	framework.RegisterBinaryFunction(framework.Operator_BinaryNotEqual, int82ne)
	framework.RegisterBinaryFunction(framework.Operator_BinaryNotEqual, int84ne)
	framework.RegisterBinaryFunction(framework.Operator_BinaryNotEqual, int8ne)
	framework.RegisterBinaryFunction(framework.Operator_BinaryNotEqual, jsonb_ne)
	framework.RegisterBinaryFunction(framework.Operator_BinaryNotEqual, namene)
	framework.RegisterBinaryFunction(framework.Operator_BinaryNotEqual, namenetext)
	framework.RegisterBinaryFunction(framework.Operator_BinaryNotEqual, numeric_ne)
	framework.RegisterBinaryFunction(framework.Operator_BinaryNotEqual, oidne)
	framework.RegisterBinaryFunction(framework.Operator_BinaryNotEqual, textnename)
	framework.RegisterBinaryFunction(framework.Operator_BinaryNotEqual, text_ne)
	framework.RegisterBinaryFunction(framework.Operator_BinaryNotEqual, time_ne)
	framework.RegisterBinaryFunction(framework.Operator_BinaryNotEqual, timestamp_ne_date)
	framework.RegisterBinaryFunction(framework.Operator_BinaryNotEqual, timestamp_ne)
	framework.RegisterBinaryFunction(framework.Operator_BinaryNotEqual, timestamp_ne_timestamptz)
	framework.RegisterBinaryFunction(framework.Operator_BinaryNotEqual, timestamptz_ne_date)
	framework.RegisterBinaryFunction(framework.Operator_BinaryNotEqual, timestamptz_ne_timestamp)
	framework.RegisterBinaryFunction(framework.Operator_BinaryNotEqual, timestamptz_ne)
	framework.RegisterBinaryFunction(framework.Operator_BinaryNotEqual, timetz_ne)
	framework.RegisterBinaryFunction(framework.Operator_BinaryNotEqual, uuid_ne)
	framework.RegisterBinaryFunction(framework.Operator_BinaryNotEqual, xidneqint4)
	framework.RegisterBinaryFunction(framework.Operator_BinaryNotEqual, xidneq)
}

// boolne represents the PostgreSQL function of the same name, taking the same parameters.
var boolne = framework.Function2{
	Name:       "boolne",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Bool, pgtypes.Bool},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Bool.Compare(val1.(bool), val2.(bool))
		return res != 0, err
	},
	Strict: true,
}

// bpcharne represents the PostgreSQL function of the same name, taking the same parameters.
var bpcharne = framework.Function2{
	Name:       "bpcharne",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.BpChar, pgtypes.BpChar},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.BpChar.Compare(val1.(string), val2.(string))
		return res != 0, err
	},
	Strict: true,
}

// byteane represents the PostgreSQL function of the same name, taking the same parameters.
var byteane = framework.Function2{
	Name:       "byteane",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Bytea, pgtypes.Bytea},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Bytea.Compare(val1.([]byte), val2.([]byte))
		return res != 0, err
	},
	Strict: true,
}

// date_ne represents the PostgreSQL function of the same name, taking the same parameters.
var date_ne = framework.Function2{
	Name:       "date_ne",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Date, pgtypes.Date},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Date.Compare(val1.(time.Time), val2.(time.Time))
		return res != 0, err
	},
	Strict: true,
}

// date_ne_timestamp represents the PostgreSQL function of the same name, taking the same parameters.
var date_ne_timestamp = framework.Function2{
	Name:       "date_ne_timestamp",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Date, pgtypes.Timestamp},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res := val1.(time.Time).Compare(val2.(time.Time))
		return res != 0, nil
	},
	Strict: true,
}

// date_ne_timestamptz represents the PostgreSQL function of the same name, taking the same parameters.
var date_ne_timestamptz = framework.Function2{
	Name:       "date_ne_timestamptz",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Date, pgtypes.TimestampTZ},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res := val1.(time.Time).Compare(val2.(time.Time))
		return res != 0, nil
	},
	Strict: true,
}

// float4ne represents the PostgreSQL function of the same name, taking the same parameters.
var float4ne = framework.Function2{
	Name:       "float4ne",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Float32, pgtypes.Float32},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Float32.Compare(val1.(float32), val2.(float32))
		return res != 0, err
	},
	Strict: true,
}

// float48ne represents the PostgreSQL function of the same name, taking the same parameters.
var float48ne = framework.Function2{
	Name:       "float48ne",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Float32, pgtypes.Float64},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Float64.Compare(float64(val1.(float32)), val2.(float64))
		return res != 0, err
	},
	Strict: true,
}

// float84ne represents the PostgreSQL function of the same name, taking the same parameters.
var float84ne = framework.Function2{
	Name:       "float84ne",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Float64, pgtypes.Float32},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Float64.Compare(val1.(float64), float64(val2.(float32)))
		return res != 0, err
	},
	Strict: true,
}

// float8ne represents the PostgreSQL function of the same name, taking the same parameters.
var float8ne = framework.Function2{
	Name:       "float8ne",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Float64, pgtypes.Float64},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Float64.Compare(val1.(float64), val2.(float64))
		return res != 0, err
	},
	Strict: true,
}

// int2ne represents the PostgreSQL function of the same name, taking the same parameters.
var int2ne = framework.Function2{
	Name:       "int2ne",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int16},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int16.Compare(val1.(int16), val2.(int16))
		return res != 0, err
	},
	Strict: true,
}

// int24ne represents the PostgreSQL function of the same name, taking the same parameters.
var int24ne = framework.Function2{
	Name:       "int24ne",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int32},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int32.Compare(int32(val1.(int16)), val2.(int32))
		return res != 0, err
	},
	Strict: true,
}

// int28ne represents the PostgreSQL function of the same name, taking the same parameters.
var int28ne = framework.Function2{
	Name:       "int28ne",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int64},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int64.Compare(int64(val1.(int16)), val2.(int64))
		return res != 0, err
	},
	Strict: true,
}

// int42ne represents the PostgreSQL function of the same name, taking the same parameters.
var int42ne = framework.Function2{
	Name:       "int42ne",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int16},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int32.Compare(val1.(int32), int32(val2.(int16)))
		return res != 0, err
	},
	Strict: true,
}

// int4ne represents the PostgreSQL function of the same name, taking the same parameters.
var int4ne = framework.Function2{
	Name:       "int4ne",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int32},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int32.Compare(val1.(int32), val2.(int32))
		return res != 0, err
	},
	Strict: true,
}

// int48ne represents the PostgreSQL function of the same name, taking the same parameters.
var int48ne = framework.Function2{
	Name:       "int48ne",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int64},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int64.Compare(int64(val1.(int32)), val2.(int64))
		return res != 0, err
	},
	Strict: true,
}

// int82ne represents the PostgreSQL function of the same name, taking the same parameters.
var int82ne = framework.Function2{
	Name:       "int82ne",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int16},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int64.Compare(val1.(int64), int64(val2.(int16)))
		return res != 0, err
	},
	Strict: true,
}

// int84ne represents the PostgreSQL function of the same name, taking the same parameters.
var int84ne = framework.Function2{
	Name:       "int84ne",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int32},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int64.Compare(val1.(int64), int64(val2.(int32)))
		return res != 0, err
	},
	Strict: true,
}

// int8ne represents the PostgreSQL function of the same name, taking the same parameters.
var int8ne = framework.Function2{
	Name:       "int8ne",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int64},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int64.Compare(val1.(int64), val2.(int64))
		return res != 0, err
	},
	Strict: true,
}

// jsonb_ne represents the PostgreSQL function of the same name, taking the same parameters.
var jsonb_ne = framework.Function2{
	Name:       "jsonb_ne",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.JsonB, pgtypes.JsonB},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.JsonB.Compare(val1.(pgtypes.JsonDocument), val2.(pgtypes.JsonDocument))
		return res != 0, err
	},
	Strict: true,
}

// namene represents the PostgreSQL function of the same name, taking the same parameters.
var namene = framework.Function2{
	Name:       "namene",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Name, pgtypes.Name},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Name.Compare(val1.(string), val2.(string))
		return res != 0, err
	},
	Strict: true,
}

// namenetext represents the PostgreSQL function of the same name, taking the same parameters.
var namenetext = framework.Function2{
	Name:       "namenetext",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Name, pgtypes.Text},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Text.Compare(val1.(string), val2.(string))
		return res != 0, err
	},
	Strict: true,
}

// numeric_ne represents the PostgreSQL function of the same name, taking the same parameters.
var numeric_ne = framework.Function2{
	Name:       "numeric_ne",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Numeric, pgtypes.Numeric},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Numeric.Compare(val1.(decimal.Decimal), val2.(decimal.Decimal))
		return res != 0, err
	},
	Strict: true,
}

// oidne represents the PostgreSQL function of the same name, taking the same parameters.
var oidne = framework.Function2{
	Name:       "oidne",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Oid, pgtypes.Oid},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Oid.Compare(val1.(uint32), val2.(uint32))
		return res != 0, err
	},
	Strict: true,
}

// textnename represents the PostgreSQL function of the same name, taking the same parameters.
var textnename = framework.Function2{
	Name:       "textnename",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Text, pgtypes.Name},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Text.Compare(val1.(string), val2.(string))
		return res != 0, err
	},
	Strict: true,
}

// text_ne represents the PostgreSQL function of the same name, taking the same parameters.
var text_ne = framework.Function2{
	Name:       "text_ne",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Text, pgtypes.Text},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Text.Compare(val1.(string), val2.(string))
		return res != 0, err
	},
	Strict: true,
}

// time_ne represents the PostgreSQL function of the same name, taking the same parameters.
var time_ne = framework.Function2{
	Name:       "time_ne",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Time, pgtypes.Time},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Time.Compare(val1.(time.Time), val2.(time.Time))
		return res != 0, err
	},
	Strict: true,
}

// timestamp_ne_date represents the PostgreSQL function of the same name, taking the same parameters.
var timestamp_ne_date = framework.Function2{
	Name:       "timestamp_ne_date",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Timestamp, pgtypes.Date},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res := val1.(time.Time).Compare(val2.(time.Time))
		return res != 0, nil
	},
	Strict: true,
}

// timestamp_ne represents the PostgreSQL function of the same name, taking the same parameters.
var timestamp_ne = framework.Function2{
	Name:       "timestamp_ne",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Timestamp, pgtypes.Timestamp},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Timestamp.Compare(val1.(time.Time), val2.(time.Time))
		return res != 0, err
	},
	Strict: true,
}

// timestamp_ne_timestamptz represents the PostgreSQL function of the same name, taking the same parameters.
var timestamp_ne_timestamptz = framework.Function2{
	Name:       "timestamp_ne_timestamptz",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Timestamp, pgtypes.TimestampTZ},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.TimestampTZ.Compare(val1.(time.Time), val2.(time.Time))
		return res != 0, err
	},
	Strict: true,
}

// timestamptz_ne_date represents the PostgreSQL function of the same name, taking the same parameters.
var timestamptz_ne_date = framework.Function2{
	Name:       "timestamptz_ne_date",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.TimestampTZ, pgtypes.Date},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res := val1.(time.Time).Compare(val2.(time.Time))
		return res != 0, nil
	},
	Strict: true,
}

// timestamptz_ne_timestamp represents the PostgreSQL function of the same name, taking the same parameters.
var timestamptz_ne_timestamp = framework.Function2{
	Name:       "timestamptz_ne_timestamp",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.TimestampTZ, pgtypes.Timestamp},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.TimestampTZ.Compare(val1.(time.Time), val2.(time.Time))
		return res != 0, err
	},
	Strict: true,
}

// timestamptz_ne represents the PostgreSQL function of the same name, taking the same parameters.
var timestamptz_ne = framework.Function2{
	Name:       "timestamptz_ne",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.TimestampTZ, pgtypes.TimestampTZ},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.TimestampTZ.Compare(val1.(time.Time), val2.(time.Time))
		return res != 0, err
	},
	Strict: true,
}

// timetz_ne represents the PostgreSQL function of the same name, taking the same parameters.
var timetz_ne = framework.Function2{
	Name:       "timetz_ne",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.TimeTZ, pgtypes.TimeTZ},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.TimeTZ.Compare(val1.(time.Time), val2.(time.Time))
		return res != 0, err
	},
	Strict: true,
}

// uuid_ne represents the PostgreSQL function of the same name, taking the same parameters.
var uuid_ne = framework.Function2{
	Name:       "uuid_ne",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Uuid, pgtypes.Uuid},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Uuid.Compare(val1.(uuid.UUID), val2.(uuid.UUID))
		return res != 0, err
	},
	Strict: true,
}

// xidneqint4 represents the PostgreSQL function of the same name, taking the same parameters.
var xidneqint4 = framework.Function2{
	Name:       "xidneqint4",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Xid, pgtypes.Int32},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		// TODO: investigate the edge cases
		res, err := pgtypes.Int64.Compare(int64(val1.(uint32)), int64(val2.(int32)))
		return res != 0, err
	},
	Strict: true,
}

// xidneq represents the PostgreSQL function of the same name, taking the same parameters.
var xidneq = framework.Function2{
	Name:       "xidneq",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Xid, pgtypes.Xid},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Xid.Compare(val1.(uint32), val2.(uint32))
		return res != 0, err
	},
	Strict: true,
}
