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
	"github.com/dolthub/doltgresql/postgres/parser/duration"
	"time"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/postgres/parser/uuid"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// These functions can be gathered using the following query from a Postgres 15 instance:
// SELECT * FROM pg_operator o WHERE o.oprname = '=' ORDER BY o.oprcode::varchar;

// initBinaryEqual registers the functions to the catalog.
func initBinaryEqual() {
	framework.RegisterBinaryFunction(framework.Operator_BinaryEqual, booleq)
	framework.RegisterBinaryFunction(framework.Operator_BinaryEqual, bpchareq)
	framework.RegisterBinaryFunction(framework.Operator_BinaryEqual, byteaeq)
	framework.RegisterBinaryFunction(framework.Operator_BinaryEqual, chareq)
	framework.RegisterBinaryFunction(framework.Operator_BinaryEqual, date_eq)
	framework.RegisterBinaryFunction(framework.Operator_BinaryEqual, date_eq_timestamp)
	framework.RegisterBinaryFunction(framework.Operator_BinaryEqual, date_eq_timestamptz)
	framework.RegisterBinaryFunction(framework.Operator_BinaryEqual, float4eq)
	framework.RegisterBinaryFunction(framework.Operator_BinaryEqual, float48eq)
	framework.RegisterBinaryFunction(framework.Operator_BinaryEqual, float84eq)
	framework.RegisterBinaryFunction(framework.Operator_BinaryEqual, float8eq)
	framework.RegisterBinaryFunction(framework.Operator_BinaryEqual, int2eq)
	framework.RegisterBinaryFunction(framework.Operator_BinaryEqual, int24eq)
	framework.RegisterBinaryFunction(framework.Operator_BinaryEqual, int28eq)
	framework.RegisterBinaryFunction(framework.Operator_BinaryEqual, int42eq)
	framework.RegisterBinaryFunction(framework.Operator_BinaryEqual, int4eq)
	framework.RegisterBinaryFunction(framework.Operator_BinaryEqual, int48eq)
	framework.RegisterBinaryFunction(framework.Operator_BinaryEqual, int82eq)
	framework.RegisterBinaryFunction(framework.Operator_BinaryEqual, int84eq)
	framework.RegisterBinaryFunction(framework.Operator_BinaryEqual, int8eq)
	framework.RegisterBinaryFunction(framework.Operator_BinaryEqual, interval_eq)
	framework.RegisterBinaryFunction(framework.Operator_BinaryEqual, jsonb_eq)
	framework.RegisterBinaryFunction(framework.Operator_BinaryEqual, nameeq)
	framework.RegisterBinaryFunction(framework.Operator_BinaryEqual, nameeqtext)
	framework.RegisterBinaryFunction(framework.Operator_BinaryEqual, numeric_eq)
	framework.RegisterBinaryFunction(framework.Operator_BinaryEqual, oideq)
	framework.RegisterBinaryFunction(framework.Operator_BinaryEqual, texteqname)
	framework.RegisterBinaryFunction(framework.Operator_BinaryEqual, text_eq)
	framework.RegisterBinaryFunction(framework.Operator_BinaryEqual, time_eq)
	framework.RegisterBinaryFunction(framework.Operator_BinaryEqual, timestamp_eq_date)
	framework.RegisterBinaryFunction(framework.Operator_BinaryEqual, timestamp_eq)
	framework.RegisterBinaryFunction(framework.Operator_BinaryEqual, timestamp_eq_timestamptz)
	framework.RegisterBinaryFunction(framework.Operator_BinaryEqual, timestamptz_eq_date)
	framework.RegisterBinaryFunction(framework.Operator_BinaryEqual, timestamptz_eq_timestamp)
	framework.RegisterBinaryFunction(framework.Operator_BinaryEqual, timestamptz_eq)
	framework.RegisterBinaryFunction(framework.Operator_BinaryEqual, timetz_eq)
	framework.RegisterBinaryFunction(framework.Operator_BinaryEqual, uuid_eq)
	framework.RegisterBinaryFunction(framework.Operator_BinaryEqual, xideqint4)
	framework.RegisterBinaryFunction(framework.Operator_BinaryEqual, xideq)
}

// booleq represents the PostgreSQL function of the same name, taking the same parameters.
var booleq = framework.Function2{
	Name:       "booleq",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Bool, pgtypes.Bool},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Bool.Compare(val1.(bool), val2.(bool))
		return res == 0, err
	},
}

// bpchareq represents the PostgreSQL function of the same name, taking the same parameters.
var bpchareq = framework.Function2{
	Name:       "bpchareq",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.BpChar, pgtypes.BpChar},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.BpChar.Compare(val1.(string), val2.(string))
		return res == 0, err
	},
}

// byteaeq represents the PostgreSQL function of the same name, taking the same parameters.
var byteaeq = framework.Function2{
	Name:       "byteaeq",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Bytea, pgtypes.Bytea},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Bytea.Compare(val1.([]byte), val2.([]byte))
		return res == 0, err
	},
}

// chareq represents the PostgreSQL function of the same name, taking the same parameters.
var chareq = framework.Function2{
	Name:       "chareq",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.InternalChar, pgtypes.InternalChar},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.InternalChar.Compare(val1.(string), val2.(string))
		return res == 0, err
	},
}

// date_eq represents the PostgreSQL function of the same name, taking the same parameters.
var date_eq = framework.Function2{
	Name:       "date_eq",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Date, pgtypes.Date},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Date.Compare(val1.(time.Time), val2.(time.Time))
		return res == 0, err
	},
}

// date_eq_timestamp represents the PostgreSQL function of the same name, taking the same parameters.
var date_eq_timestamp = framework.Function2{
	Name:       "date_eq_timestamp",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Date, pgtypes.Timestamp},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res := val1.(time.Time).Compare(val2.(time.Time))
		return res == 0, nil
	},
}

// date_eq_timestamptz represents the PostgreSQL function of the same name, taking the same parameters.
var date_eq_timestamptz = framework.Function2{
	Name:       "date_eq_timestamptz",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Date, pgtypes.TimestampTZ},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res := val1.(time.Time).Compare(val2.(time.Time))
		return res == 0, nil
	},
}

// float4eq represents the PostgreSQL function of the same name, taking the same parameters.
var float4eq = framework.Function2{
	Name:       "float4eq",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Float32, pgtypes.Float32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Float32.Compare(val1.(float32), val2.(float32))
		return res == 0, err
	},
}

// float48eq represents the PostgreSQL function of the same name, taking the same parameters.
var float48eq = framework.Function2{
	Name:       "float48eq",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Float32, pgtypes.Float64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Float64.Compare(float64(val1.(float32)), val2.(float64))
		return res == 0, err
	},
}

// float84eq represents the PostgreSQL function of the same name, taking the same parameters.
var float84eq = framework.Function2{
	Name:       "float84eq",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Float64, pgtypes.Float32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Float64.Compare(val1.(float64), float64(val2.(float32)))
		return res == 0, err
	},
}

// float8eq represents the PostgreSQL function of the same name, taking the same parameters.
var float8eq = framework.Function2{
	Name:       "float8eq",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Float64, pgtypes.Float64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Float64.Compare(val1.(float64), val2.(float64))
		return res == 0, err
	},
}

// int2eq represents the PostgreSQL function of the same name, taking the same parameters.
var int2eq = framework.Function2{
	Name:       "int2eq",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int16},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int16.Compare(val1.(int16), val2.(int16))
		return res == 0, err
	},
}

// int24eq represents the PostgreSQL function of the same name, taking the same parameters.
var int24eq = framework.Function2{
	Name:       "int24eq",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int32.Compare(int32(val1.(int16)), val2.(int32))
		return res == 0, err
	},
}

// int28eq represents the PostgreSQL function of the same name, taking the same parameters.
var int28eq = framework.Function2{
	Name:       "int28eq",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int64.Compare(int64(val1.(int16)), val2.(int64))
		return res == 0, err
	},
}

// int42eq represents the PostgreSQL function of the same name, taking the same parameters.
var int42eq = framework.Function2{
	Name:       "int42eq",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int16},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int32.Compare(val1.(int32), int32(val2.(int16)))
		return res == 0, err
	},
}

// int4eq represents the PostgreSQL function of the same name, taking the same parameters.
var int4eq = framework.Function2{
	Name:       "int4eq",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int32.Compare(val1.(int32), val2.(int32))
		return res == 0, err
	},
}

// int48eq represents the PostgreSQL function of the same name, taking the same parameters.
var int48eq = framework.Function2{
	Name:       "int48eq",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int64.Compare(int64(val1.(int32)), val2.(int64))
		return res == 0, err
	},
}

// int82eq represents the PostgreSQL function of the same name, taking the same parameters.
var int82eq = framework.Function2{
	Name:       "int82eq",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int16},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int64.Compare(val1.(int64), int64(val2.(int16)))
		return res == 0, err
	},
}

// int84eq represents the PostgreSQL function of the same name, taking the same parameters.
var int84eq = framework.Function2{
	Name:       "int84eq",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int64.Compare(val1.(int64), int64(val2.(int32)))
		return res == 0, err
	},
}

// int8eq represents the PostgreSQL function of the same name, taking the same parameters.
var int8eq = framework.Function2{
	Name:       "int8eq",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int64.Compare(val1.(int64), val2.(int64))
		return res == 0, err
	},
}

// interval_eq represents the PostgreSQL function of the same name, taking the same parameters.
var interval_eq = framework.Function2{
	Name:       "interval_eq",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Interval, pgtypes.Interval},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Interval.Compare(val1.(duration.Duration), val2.(duration.Duration))
		return res == 0, err
	},
}

// jsonb_eq represents the PostgreSQL function of the same name, taking the same parameters.
var jsonb_eq = framework.Function2{
	Name:       "jsonb_eq",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.JsonB, pgtypes.JsonB},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.JsonB.Compare(val1.(pgtypes.JsonDocument), val2.(pgtypes.JsonDocument))
		return res == 0, err
	},
}

// nameeq represents the PostgreSQL function of the same name, taking the same parameters.
var nameeq = framework.Function2{
	Name:       "nameeq",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Name, pgtypes.Name},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Name.Compare(val1.(string), val2.(string))
		return res == 0, err
	},
}

// nameeqtext represents the PostgreSQL function of the same name, taking the same parameters.
var nameeqtext = framework.Function2{
	Name:       "nameeqtext",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Name, pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Text.Compare(val1.(string), val2.(string))
		return res == 0, err
	},
}

// numeric_eq represents the PostgreSQL function of the same name, taking the same parameters.
var numeric_eq = framework.Function2{
	Name:       "numeric_eq",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Numeric, pgtypes.Numeric},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Numeric.Compare(val1.(decimal.Decimal), val2.(decimal.Decimal))
		return res == 0, err
	},
}

// oideq represents the PostgreSQL function of the same name, taking the same parameters.
var oideq = framework.Function2{
	Name:       "oideq",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Oid, pgtypes.Oid},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Oid.Compare(val1.(uint32), val2.(uint32))
		return res == 0, err
	},
}

// texteqname represents the PostgreSQL function of the same name, taking the same parameters.
var texteqname = framework.Function2{
	Name:       "texteqname",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Text, pgtypes.Name},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Text.Compare(val1.(string), val2.(string))
		return res == 0, err
	},
}

// text_eq represents the PostgreSQL function of the same name, taking the same parameters.
var text_eq = framework.Function2{
	Name:       "text_eq",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Text, pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Text.Compare(val1.(string), val2.(string))
		return res == 0, err
	},
}

// time_eq represents the PostgreSQL function of the same name, taking the same parameters.
var time_eq = framework.Function2{
	Name:       "time_eq",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Time, pgtypes.Time},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Time.Compare(val1.(time.Time), val2.(time.Time))
		return res == 0, err
	},
}

// timestamp_eq_date represents the PostgreSQL function of the same name, taking the same parameters.
var timestamp_eq_date = framework.Function2{
	Name:       "timestamp_eq_date",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Timestamp, pgtypes.Date},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res := val1.(time.Time).Compare(val2.(time.Time))
		return res == 0, nil
	},
}

// timestamp_eq represents the PostgreSQL function of the same name, taking the same parameters.
var timestamp_eq = framework.Function2{
	Name:       "timestamp_eq",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Timestamp, pgtypes.Timestamp},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Timestamp.Compare(val1.(time.Time), val2.(time.Time))
		return res == 0, err
	},
}

// timestamp_eq_timestamptz represents the PostgreSQL function of the same name, taking the same parameters.
var timestamp_eq_timestamptz = framework.Function2{
	Name:       "timestamp_eq_timestamptz",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Timestamp, pgtypes.TimestampTZ},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.TimestampTZ.Compare(val1.(time.Time), val2.(time.Time))
		return res == 0, err
	},
}

// timestamptz_eq_date represents the PostgreSQL function of the same name, taking the same parameters.
var timestamptz_eq_date = framework.Function2{
	Name:       "timestamptz_eq_date",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.TimestampTZ, pgtypes.Date},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res := val1.(time.Time).Compare(val2.(time.Time))
		return res == 0, nil
	},
}

// timestamptz_eq_timestamp represents the PostgreSQL function of the same name, taking the same parameters.
var timestamptz_eq_timestamp = framework.Function2{
	Name:       "timestamptz_eq_timestamp",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.TimestampTZ, pgtypes.Timestamp},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.TimestampTZ.Compare(val1.(time.Time), val2.(time.Time))
		return res == 0, err
	},
}

// timestamptz_eq represents the PostgreSQL function of the same name, taking the same parameters.
var timestamptz_eq = framework.Function2{
	Name:       "timestamptz_eq",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.TimestampTZ, pgtypes.TimestampTZ},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.TimestampTZ.Compare(val1.(time.Time), val2.(time.Time))
		return res == 0, err
	},
}

// timetz_eq represents the PostgreSQL function of the same name, taking the same parameters.
var timetz_eq = framework.Function2{
	Name:       "timetz_eq",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.TimeTZ, pgtypes.TimeTZ},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.TimeTZ.Compare(val1.(time.Time), val2.(time.Time))
		return res == 0, err
	},
}

// uuid_eq represents the PostgreSQL function of the same name, taking the same parameters.
var uuid_eq = framework.Function2{
	Name:       "uuid_eq",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Uuid, pgtypes.Uuid},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Uuid.Compare(val1.(uuid.UUID), val2.(uuid.UUID))
		return res == 0, err
	},
}

// xideqint4 represents the PostgreSQL function of the same name, taking the same parameters.
var xideqint4 = framework.Function2{
	Name:       "xideqint4",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Xid, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		// TODO: investigate the edge cases
		res, err := pgtypes.Int64.Compare(int64(val1.(uint32)), int64(val2.(int32)))
		return res == 0, err
	},
}

// xideq represents the PostgreSQL function of the same name, taking the same parameters.
var xideq = framework.Function2{
	Name:       "xideq",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Xid, pgtypes.Xid},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Xid.Compare(val1.(uint32), val2.(uint32))
		return res == 0, err
	},
}
