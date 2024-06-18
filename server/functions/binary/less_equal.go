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
// SELECT * FROM pg_operator o WHERE o.oprname = '<=' ORDER BY o.oprcode::varchar;

// initBinaryLessOrEqual registers the functions to the catalog.
func initBinaryLessOrEqual() {
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessOrEqual, boolle)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessOrEqual, bpcharle)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessOrEqual, byteale)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessOrEqual, date_le)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessOrEqual, date_le_timestamp)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessOrEqual, date_le_timestamptz)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessOrEqual, float4le)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessOrEqual, float48le)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessOrEqual, float84le)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessOrEqual, float8le)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessOrEqual, int2le)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessOrEqual, int24le)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessOrEqual, int28le)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessOrEqual, int42le)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessOrEqual, int4le)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessOrEqual, int48le)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessOrEqual, int82le)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessOrEqual, int84le)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessOrEqual, int8le)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessOrEqual, jsonb_le)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessOrEqual, namele)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessOrEqual, nameletext)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessOrEqual, numeric_le)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessOrEqual, oidle)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessOrEqual, textlename)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessOrEqual, text_le)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessOrEqual, time_le)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessOrEqual, timestamp_le_date)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessOrEqual, timestamp_le)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessOrEqual, timestamp_le_timestamptz)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessOrEqual, timestamptz_le_date)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessOrEqual, timestamptz_le_timestamp)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessOrEqual, timestamptz_le)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessOrEqual, timetz_le)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessOrEqual, uuid_le)
}

// boolle represents the PostgreSQL function of the same name, taking the same parameters.
var boolle = framework.Function2{
	Name:       "boolle",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Bool, pgtypes.Bool},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Bool.Compare(val1.(bool), val2.(bool))
		return res <= 0, err
	},
	Strict: true,
}

// bpcharle represents the PostgreSQL function of the same name, taking the same parameters.
var bpcharle = framework.Function2{
	Name:       "bpcharle",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.BpChar, pgtypes.BpChar},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.BpChar.Compare(val1.(string), val2.(string))
		return res <= 0, err
	},
	Strict: true,
}

// byteale represents the PostgreSQL function of the same name, taking the same parameters.
var byteale = framework.Function2{
	Name:       "byteale",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Bytea, pgtypes.Bytea},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Bytea.Compare(val1.([]byte), val2.([]byte))
		return res <= 0, err
	},
	Strict: true,
}

// date_le represents the PostgreSQL function of the same name, taking the same parameters.
var date_le = framework.Function2{
	Name:       "date_le",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Date, pgtypes.Date},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Date.Compare(val1.(time.Time), val2.(time.Time))
		return res <= 0, err
	},
	Strict: true,
}

// date_le_timestamp represents the PostgreSQL function of the same name, taking the same parameters.
var date_le_timestamp = framework.Function2{
	Name:       "date_le_timestamp",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Date, pgtypes.Timestamp},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res := val1.(time.Time).Compare(val2.(time.Time))
		return res <= 0, nil
	},
	Strict: true,
}

// date_le_timestamptz represents the PostgreSQL function of the same name, taking the same parameters.
var date_le_timestamptz = framework.Function2{
	Name:       "date_le_timestamptz",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Date, pgtypes.TimestampTZ},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res := val1.(time.Time).Compare(val2.(time.Time))
		return res <= 0, nil
	},
	Strict: true,
}

// float4le represents the PostgreSQL function of the same name, taking the same parameters.
var float4le = framework.Function2{
	Name:       "float4le",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Float32, pgtypes.Float32},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Float32.Compare(val1.(float32), val2.(float32))
		return res <= 0, err
	},
	Strict: true,
}

// float48le represents the PostgreSQL function of the same name, taking the same parameters.
var float48le = framework.Function2{
	Name:       "float48le",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Float32, pgtypes.Float64},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Float64.Compare(float64(val1.(float32)), val2.(float64))
		return res <= 0, err
	},
	Strict: true,
}

// float84le represents the PostgreSQL function of the same name, taking the same parameters.
var float84le = framework.Function2{
	Name:       "float84le",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Float64, pgtypes.Float32},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Float64.Compare(val1.(float64), float64(val2.(float32)))
		return res <= 0, err
	},
	Strict: true,
}

// float8le represents the PostgreSQL function of the same name, taking the same parameters.
var float8le = framework.Function2{
	Name:       "float8le",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Float64, pgtypes.Float64},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Float64.Compare(val1.(float64), val2.(float64))
		return res <= 0, err
	},
	Strict: true,
}

// int2le represents the PostgreSQL function of the same name, taking the same parameters.
var int2le = framework.Function2{
	Name:       "int2le",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int16},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int16.Compare(val1.(int16), val2.(int16))
		return res <= 0, err
	},
	Strict: true,
}

// int24le represents the PostgreSQL function of the same name, taking the same parameters.
var int24le = framework.Function2{
	Name:       "int24le",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int32},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int32.Compare(int32(val1.(int16)), val2.(int32))
		return res <= 0, err
	},
	Strict: true,
}

// int28le represents the PostgreSQL function of the same name, taking the same parameters.
var int28le = framework.Function2{
	Name:       "int28le",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int64},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int64.Compare(int64(val1.(int16)), val2.(int64))
		return res <= 0, err
	},
	Strict: true,
}

// int42le represents the PostgreSQL function of the same name, taking the same parameters.
var int42le = framework.Function2{
	Name:       "int42le",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int16},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int32.Compare(val1.(int32), int32(val2.(int16)))
		return res <= 0, err
	},
	Strict: true,
}

// int4le represents the PostgreSQL function of the same name, taking the same parameters.
var int4le = framework.Function2{
	Name:       "int4le",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int32},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int32.Compare(val1.(int32), val2.(int32))
		return res <= 0, err
	},
	Strict: true,
}

// int48le represents the PostgreSQL function of the same name, taking the same parameters.
var int48le = framework.Function2{
	Name:       "int48le",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int64},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int64.Compare(int64(val1.(int32)), val2.(int64))
		return res <= 0, err
	},
	Strict: true,
}

// int82le represents the PostgreSQL function of the same name, taking the same parameters.
var int82le = framework.Function2{
	Name:       "int82le",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int16},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int64.Compare(val1.(int64), int64(val2.(int16)))
		return res <= 0, err
	},
	Strict: true,
}

// int84le represents the PostgreSQL function of the same name, taking the same parameters.
var int84le = framework.Function2{
	Name:       "int84le",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int32},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int64.Compare(val1.(int64), int64(val2.(int32)))
		return res <= 0, err
	},
	Strict: true,
}

// int8le represents the PostgreSQL function of the same name, taking the same parameters.
var int8le = framework.Function2{
	Name:       "int8le",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int64},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int64.Compare(val1.(int64), val2.(int64))
		return res <= 0, err
	},
	Strict: true,
}

// jsonb_le represents the PostgreSQL function of the same name, taking the same parameters.
var jsonb_le = framework.Function2{
	Name:       "jsonb_le",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.JsonB, pgtypes.JsonB},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.JsonB.Compare(val1.(pgtypes.JsonDocument), val2.(pgtypes.JsonDocument))
		return res <= 0, err
	},
	Strict: true,
}

// namele represents the PostgreSQL function of the same name, taking the same parameters.
var namele = framework.Function2{
	Name:       "namele",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Name, pgtypes.Name},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Name.Compare(val1.(string), val2.(string))
		return res <= 0, err
	},
	Strict: true,
}

// nameletext represents the PostgreSQL function of the same name, taking the same parameters.
var nameletext = framework.Function2{
	Name:       "nameletext",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Name, pgtypes.Text},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Text.Compare(val1.(string), val2.(string))
		return res <= 0, err
	},
	Strict: true,
}

// numeric_le represents the PostgreSQL function of the same name, taking the same parameters.
var numeric_le = framework.Function2{
	Name:       "numeric_le",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Numeric, pgtypes.Numeric},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Numeric.Compare(val1.(decimal.Decimal), val2.(decimal.Decimal))
		return res <= 0, err
	},
	Strict: true,
}

// oidle represents the PostgreSQL function of the same name, taking the same parameters.
var oidle = framework.Function2{
	Name:       "oidle",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Oid, pgtypes.Oid},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Oid.Compare(val1.(uint32), val2.(uint32))
		return res <= 0, err
	},
	Strict: true,
}

// textlename represents the PostgreSQL function of the same name, taking the same parameters.
var textlename = framework.Function2{
	Name:       "textlename",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Text, pgtypes.Name},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Text.Compare(val1.(string), val2.(string))
		return res <= 0, err
	},
	Strict: true,
}

// text_le represents the PostgreSQL function of the same name, taking the same parameters.
var text_le = framework.Function2{
	Name:       "text_le",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Text, pgtypes.Text},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Text.Compare(val1.(string), val2.(string))
		return res <= 0, err
	},
	Strict: true,
}

// time_le represents the PostgreSQL function of the same name, taking the same parameters.
var time_le = framework.Function2{
	Name:       "time_le",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Time, pgtypes.Time},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Time.Compare(val1.(time.Time), val2.(time.Time))
		return res <= 0, err
	},
	Strict: true,
}

// timestamp_le_date represents the PostgreSQL function of the same name, taking the same parameters.
var timestamp_le_date = framework.Function2{
	Name:       "timestamp_le_date",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Timestamp, pgtypes.Date},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res := val1.(time.Time).Compare(val2.(time.Time))
		return res <= 0, nil
	},
	Strict: true,
}

// timestamp_le represents the PostgreSQL function of the same name, taking the same parameters.
var timestamp_le = framework.Function2{
	Name:       "timestamp_le",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Timestamp, pgtypes.Timestamp},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Timestamp.Compare(val1.(time.Time), val2.(time.Time))
		return res <= 0, err
	},
	Strict: true,
}

// timestamp_le_timestamptz represents the PostgreSQL function of the same name, taking the same parameters.
var timestamp_le_timestamptz = framework.Function2{
	Name:       "timestamp_le_timestamptz",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Timestamp, pgtypes.TimestampTZ},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.TimestampTZ.Compare(val1.(time.Time), val2.(time.Time))
		return res <= 0, err
	},
	Strict: true,
}

// timestamptz_le_date represents the PostgreSQL function of the same name, taking the same parameters.
var timestamptz_le_date = framework.Function2{
	Name:       "timestamptz_le_date",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.TimestampTZ, pgtypes.Date},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res := val1.(time.Time).Compare(val2.(time.Time))
		return res <= 0, nil
	},
	Strict: true,
}

// timestamptz_le_timestamp represents the PostgreSQL function of the same name, taking the same parameters.
var timestamptz_le_timestamp = framework.Function2{
	Name:       "timestamptz_le_timestamp",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.TimestampTZ, pgtypes.Timestamp},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.TimestampTZ.Compare(val1.(time.Time), val2.(time.Time))
		return res <= 0, err
	},
	Strict: true,
}

// timestamptz_le represents the PostgreSQL function of the same name, taking the same parameters.
var timestamptz_le = framework.Function2{
	Name:       "timestamptz_le",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.TimestampTZ, pgtypes.TimestampTZ},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.TimestampTZ.Compare(val1.(time.Time), val2.(time.Time))
		return res <= 0, err
	},
	Strict: true,
}

// timetz_le represents the PostgreSQL function of the same name, taking the same parameters.
var timetz_le = framework.Function2{
	Name:       "timetz_le",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.TimeTZ, pgtypes.TimeTZ},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.TimeTZ.Compare(val1.(time.Time), val2.(time.Time))
		return res <= 0, err
	},
	Strict: true,
}

// uuid_le represents the PostgreSQL function of the same name, taking the same parameters.
var uuid_le = framework.Function2{
	Name:       "uuid_le",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Uuid, pgtypes.Uuid},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Uuid.Compare(val1.(uuid.UUID), val2.(uuid.UUID))
		return res <= 0, err
	},
	Strict: true,
}
