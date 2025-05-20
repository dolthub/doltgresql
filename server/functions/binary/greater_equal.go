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
	"cmp"
	"time"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/postgres/parser/duration"
	"github.com/dolthub/doltgresql/postgres/parser/uuid"
	"github.com/dolthub/doltgresql/server/compare"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// These functions can be gathered using the following query from a Postgres 15 instance:
// SELECT * FROM pg_operator o WHERE o.oprname = '>=' ORDER BY o.oprcode::varchar;

// initBinaryGreaterOrEqual registers the functions to the catalog.
func initBinaryGreaterOrEqual() {
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterOrEqual, boolge)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterOrEqual, bpcharge)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterOrEqual, byteage)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterOrEqual, charge)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterOrEqual, date_ge)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterOrEqual, date_ge_timestamp)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterOrEqual, date_ge_timestamptz)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterOrEqual, enum_ge)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterOrEqual, float4ge)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterOrEqual, float48ge)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterOrEqual, float84ge)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterOrEqual, float8ge)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterOrEqual, int2ge)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterOrEqual, int24ge)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterOrEqual, int28ge)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterOrEqual, int42ge)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterOrEqual, int4ge)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterOrEqual, int48ge)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterOrEqual, int82ge)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterOrEqual, int84ge)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterOrEqual, int8ge)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterOrEqual, interval_ge)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterOrEqual, jsonb_ge)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterOrEqual, namege)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterOrEqual, namegetext)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterOrEqual, numeric_ge)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterOrEqual, oidge)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterOrEqual, textgename)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterOrEqual, text_ge)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterOrEqual, time_ge)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterOrEqual, timestamp_ge_date)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterOrEqual, timestamp_ge)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterOrEqual, timestamp_ge_timestamptz)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterOrEqual, timestamptz_ge_date)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterOrEqual, timestamptz_ge_timestamp)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterOrEqual, timestamptz_ge)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterOrEqual, timetz_ge)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterOrEqual, record_ge)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterOrEqual, uuid_ge)
}

// boolge represents the PostgreSQL function of the same name, taking the same parameters.
var boolge = framework.Function2{
	Name:       "boolge",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Bool, pgtypes.Bool},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Bool.Compare(ctx, val1.(bool), val2.(bool))
		return res >= 0, err
	},
}

// bpcharge represents the PostgreSQL function of the same name, taking the same parameters.
var bpcharge = framework.Function2{
	Name:       "bpcharge",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.BpChar, pgtypes.BpChar},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.BpChar.Compare(ctx, val1.(string), val2.(string))
		return res >= 0, err
	},
}

// byteage represents the PostgreSQL function of the same name, taking the same parameters.
var byteage = framework.Function2{
	Name:       "byteage",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Bytea, pgtypes.Bytea},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Bytea.Compare(ctx, val1.([]byte), val2.([]byte))
		return res >= 0, err
	},
}

// charge represents the PostgreSQL function of the same name, taking the same parameters.
var charge = framework.Function2{
	Name:       "charge",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.InternalChar, pgtypes.InternalChar},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.InternalChar.Compare(ctx, val1.(string), val2.(string))
		return res >= 0, err
	},
}

// date_ge represents the PostgreSQL function of the same name, taking the same parameters.
var date_ge = framework.Function2{
	Name:       "date_ge",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Date, pgtypes.Date},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Date.Compare(ctx, val1.(time.Time), val2.(time.Time))
		return res >= 0, err
	},
}

// date_ge_timestamp represents the PostgreSQL function of the same name, taking the same parameters.
var date_ge_timestamp = framework.Function2{
	Name:       "date_ge_timestamp",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Date, pgtypes.Timestamp},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res := val1.(time.Time).Compare(val2.(time.Time))
		return res >= 0, nil
	},
}

// date_ge_timestamptz represents the PostgreSQL function of the same name, taking the same parameters.
var date_ge_timestamptz = framework.Function2{
	Name:       "date_ge_timestamptz",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Date, pgtypes.TimestampTZ},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res := val1.(time.Time).Compare(val2.(time.Time))
		return res >= 0, nil
	},
}

// enum_ge represents the PostgreSQL function of the same name, taking the same parameters.
var enum_ge = framework.Function2{
	Name:       "enum_ge",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.AnyEnum, pgtypes.AnyEnum},
	Strict:     true,
	Callable: func(ctx *sql.Context, t [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := t[0].Compare(ctx, val1, val2)
		return res >= 0, err
	},
}

// float4ge represents the PostgreSQL function of the same name, taking the same parameters.
var float4ge = framework.Function2{
	Name:       "float4ge",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Float32, pgtypes.Float32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Float32.Compare(ctx, val1.(float32), val2.(float32))
		return res >= 0, err
	},
}

// float48ge represents the PostgreSQL function of the same name, taking the same parameters.
var float48ge = framework.Function2{
	Name:       "float48ge",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Float32, pgtypes.Float64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Float64.Compare(ctx, float64(val1.(float32)), val2.(float64))
		return res >= 0, err
	},
}

// float84ge represents the PostgreSQL function of the same name, taking the same parameters.
var float84ge = framework.Function2{
	Name:       "float84ge",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Float64, pgtypes.Float32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Float64.Compare(ctx, val1.(float64), float64(val2.(float32)))
		return res >= 0, err
	},
}

// float8ge represents the PostgreSQL function of the same name, taking the same parameters.
var float8ge = framework.Function2{
	Name:       "float8ge",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Float64, pgtypes.Float64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Float64.Compare(ctx, val1.(float64), val2.(float64))
		return res >= 0, err
	},
}

// int2ge represents the PostgreSQL function of the same name, taking the same parameters.
var int2ge = framework.Function2{
	Name:       "int2ge",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int16},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int16.Compare(ctx, val1.(int16), val2.(int16))
		return res >= 0, err
	},
}

// int24ge represents the PostgreSQL function of the same name, taking the same parameters.
var int24ge = framework.Function2{
	Name:       "int24ge",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int32.Compare(ctx, int32(val1.(int16)), val2.(int32))
		return res >= 0, err
	},
}

// int28ge represents the PostgreSQL function of the same name, taking the same parameters.
var int28ge = framework.Function2{
	Name:       "int28ge",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int64.Compare(ctx, int64(val1.(int16)), val2.(int64))
		return res >= 0, err
	},
}

// int42ge represents the PostgreSQL function of the same name, taking the same parameters.
var int42ge = framework.Function2{
	Name:       "int42ge",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int16},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int32.Compare(ctx, val1.(int32), int32(val2.(int16)))
		return res >= 0, err
	},
}

// int4ge represents the PostgreSQL function of the same name, taking the same parameters.
var int4ge = framework.Function2{
	Name:       "int4ge",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int32.Compare(ctx, val1.(int32), val2.(int32))
		return res >= 0, err
	},
}

// int48ge represents the PostgreSQL function of the same name, taking the same parameters.
var int48ge = framework.Function2{
	Name:       "int48ge",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int64.Compare(ctx, int64(val1.(int32)), val2.(int64))
		return res >= 0, err
	},
}

// int82ge represents the PostgreSQL function of the same name, taking the same parameters.
var int82ge = framework.Function2{
	Name:       "int82ge",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int16},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int64.Compare(ctx, val1.(int64), int64(val2.(int16)))
		return res >= 0, err
	},
}

// int84ge represents the PostgreSQL function of the same name, taking the same parameters.
var int84ge = framework.Function2{
	Name:       "int84ge",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int64.Compare(ctx, val1.(int64), int64(val2.(int32)))
		return res >= 0, err
	},
}

// int8ge represents the PostgreSQL function of the same name, taking the same parameters.
var int8ge = framework.Function2{
	Name:       "int8ge",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int64.Compare(ctx, val1.(int64), val2.(int64))
		return res >= 0, err
	},
}

// interval_ge represents the PostgreSQL function of the same name, taking the same parameters.
var interval_ge = framework.Function2{
	Name:       "interval_ge",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Interval, pgtypes.Interval},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Interval.Compare(ctx, val1.(duration.Duration), val2.(duration.Duration))
		return res >= 0, err
	},
}

// jsonb_ge represents the PostgreSQL function of the same name, taking the same parameters.
var jsonb_ge = framework.Function2{
	Name:       "jsonb_ge",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.JsonB, pgtypes.JsonB},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.JsonB.Compare(ctx, val1.(pgtypes.JsonDocument), val2.(pgtypes.JsonDocument))
		return res >= 0, err
	},
}

// namege represents the PostgreSQL function of the same name, taking the same parameters.
var namege = framework.Function2{
	Name:       "namege",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Name, pgtypes.Name},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Name.Compare(ctx, val1.(string), val2.(string))
		return res >= 0, err
	},
}

// namegetext represents the PostgreSQL function of the same name, taking the same parameters.
var namegetext = framework.Function2{
	Name:       "namegetext",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Name, pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Text.Compare(ctx, val1.(string), val2.(string))
		return res >= 0, err
	},
}

// numeric_ge represents the PostgreSQL function of the same name, taking the same parameters.
var numeric_ge = framework.Function2{
	Name:       "numeric_ge",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Numeric, pgtypes.Numeric},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Numeric.Compare(ctx, val1.(decimal.Decimal), val2.(decimal.Decimal))
		return res >= 0, err
	},
}

// oidge represents the PostgreSQL function of the same name, taking the same parameters.
var oidge = framework.Function2{
	Name:       "oidge",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Oid, pgtypes.Oid},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res := cmp.Compare(id.Cache().ToOID(val1.(id.Id)), id.Cache().ToOID(val2.(id.Id)))
		return res >= 0, nil
	},
}

// textgename represents the PostgreSQL function of the same name, taking the same parameters.
var textgename = framework.Function2{
	Name:       "textgename",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Name},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Text.Compare(ctx, val1.(string), val2.(string))
		return res >= 0, err
	},
}

// text_ge represents the PostgreSQL function of the same name, taking the same parameters.
var text_ge = framework.Function2{
	Name:       "text_ge",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Text.Compare(ctx, val1.(string), val2.(string))
		return res >= 0, err
	},
}

// time_ge represents the PostgreSQL function of the same name, taking the same parameters.
var time_ge = framework.Function2{
	Name:       "time_ge",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Time, pgtypes.Time},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Time.Compare(ctx, val1.(time.Time), val2.(time.Time))
		return res >= 0, err
	},
}

// timestamp_ge_date represents the PostgreSQL function of the same name, taking the same parameters.
var timestamp_ge_date = framework.Function2{
	Name:       "timestamp_ge_date",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Timestamp, pgtypes.Date},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res := val1.(time.Time).Compare(val2.(time.Time))
		return res >= 0, nil
	},
}

// timestamp_ge represents the PostgreSQL function of the same name, taking the same parameters.
var timestamp_ge = framework.Function2{
	Name:       "timestamp_ge",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Timestamp, pgtypes.Timestamp},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Timestamp.Compare(ctx, val1.(time.Time), val2.(time.Time))
		return res >= 0, err
	},
}

// timestamp_ge_timestamptz represents the PostgreSQL function of the same name, taking the same parameters.
var timestamp_ge_timestamptz = framework.Function2{
	Name:       "timestamp_ge_timestamptz",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Timestamp, pgtypes.TimestampTZ},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.TimestampTZ.Compare(ctx, val1.(time.Time), val2.(time.Time))
		return res >= 0, err
	},
}

// timestamptz_ge_date represents the PostgreSQL function of the same name, taking the same parameters.
var timestamptz_ge_date = framework.Function2{
	Name:       "timestamptz_ge_date",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.TimestampTZ, pgtypes.Date},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res := val1.(time.Time).Compare(val2.(time.Time))
		return res >= 0, nil
	},
}

// timestamptz_ge_timestamp represents the PostgreSQL function of the same name, taking the same parameters.
var timestamptz_ge_timestamp = framework.Function2{
	Name:       "timestamptz_ge_timestamp",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.TimestampTZ, pgtypes.Timestamp},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.TimestampTZ.Compare(ctx, val1.(time.Time), val2.(time.Time))
		return res >= 0, err
	},
}

// timestamptz_ge represents the PostgreSQL function of the same name, taking the same parameters.
var timestamptz_ge = framework.Function2{
	Name:       "timestamptz_ge",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.TimestampTZ, pgtypes.TimestampTZ},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.TimestampTZ.Compare(ctx, val1.(time.Time), val2.(time.Time))
		return res >= 0, err
	},
}

// timetz_ge represents the PostgreSQL function of the same name, taking the same parameters.
var timetz_ge = framework.Function2{
	Name:       "timetz_ge",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.TimeTZ, pgtypes.TimeTZ},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.TimeTZ.Compare(ctx, val1.(time.Time), val2.(time.Time))
		return res >= 0, err
	},
}

// record_ge represents the PostgreSQL function of the same name, taking the same parameters.
var record_ge = framework.Function2{
	Name:       "record_ge",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Record, pgtypes.Record},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		return compare.CompareRecords(ctx, framework.Operator_BinaryGreaterOrEqual, val1, val2)
	},
}

// uuid_ge represents the PostgreSQL function of the same name, taking the same parameters.
var uuid_ge = framework.Function2{
	Name:       "uuid_ge",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Uuid, pgtypes.Uuid},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Uuid.Compare(ctx, val1.(uuid.UUID), val2.(uuid.UUID))
		return res >= 0, err
	},
}
