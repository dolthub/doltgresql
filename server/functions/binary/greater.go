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
// SELECT * FROM pg_operator o WHERE o.oprname = '>' ORDER BY o.oprcode::varchar;

// initBinaryGreaterThan registers the functions to the catalog.
func initBinaryGreaterThan() {
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterThan, boolgt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterThan, bpchargt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterThan, byteagt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterThan, chargt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterThan, date_gt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterThan, date_gt_timestamp)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterThan, date_gt_timestamptz)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterThan, float4gt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterThan, float48gt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterThan, float84gt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterThan, float8gt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterThan, int2gt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterThan, int24gt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterThan, int28gt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterThan, int42gt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterThan, int4gt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterThan, int48gt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterThan, int82gt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterThan, int84gt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterThan, int8gt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterThan, interval_gt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterThan, jsonb_gt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterThan, namegt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterThan, namegttext)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterThan, numeric_gt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterThan, oidgt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterThan, textgtname)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterThan, text_gt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterThan, time_gt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterThan, timestamp_gt_date)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterThan, timestamp_gt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterThan, timestamp_gt_timestamptz)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterThan, timestamptz_gt_date)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterThan, timestamptz_gt_timestamp)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterThan, timestamptz_gt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterThan, timetz_gt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterThan, uuid_gt)
}

// boolgt represents the PostgreSQL function of the same name, taking the same parameters.
var boolgt = framework.Function2{
	Name:       "boolgt",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Bool, pgtypes.Bool},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Bool.Compare(val1.(bool), val2.(bool))
		return res == 1, err
	},
}

// bpchargt represents the PostgreSQL function of the same name, taking the same parameters.
var bpchargt = framework.Function2{
	Name:       "bpchargt",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.BpChar, pgtypes.BpChar},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.BpChar.Compare(val1.(string), val2.(string))
		return res == 1, err
	},
}

// byteagt represents the PostgreSQL function of the same name, taking the same parameters.
var byteagt = framework.Function2{
	Name:       "byteagt",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Bytea, pgtypes.Bytea},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Bytea.Compare(val1.([]byte), val2.([]byte))
		return res == 1, err
	},
}

// chargt represents the PostgreSQL function of the same name, taking the same parameters.
var chargt = framework.Function2{
	Name:       "chargt",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.InternalChar, pgtypes.InternalChar},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.InternalChar.Compare(val1.(string), val2.(string))
		return res == 1, err
	},
}

// date_gt represents the PostgreSQL function of the same name, taking the same parameters.
var date_gt = framework.Function2{
	Name:       "date_gt",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Date, pgtypes.Date},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Date.Compare(val1.(time.Time), val2.(time.Time))
		return res == 1, err
	},
}

// date_gt_timestamp represents the PostgreSQL function of the same name, taking the same parameters.
var date_gt_timestamp = framework.Function2{
	Name:       "date_gt_timestamp",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Date, pgtypes.Timestamp},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res := val1.(time.Time).Compare(val2.(time.Time))
		return res == 1, nil
	},
}

// date_gt_timestamptz represents the PostgreSQL function of the same name, taking the same parameters.
var date_gt_timestamptz = framework.Function2{
	Name:       "date_gt_timestamptz",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Date, pgtypes.TimestampTZ},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res := val1.(time.Time).Compare(val2.(time.Time))
		return res == 1, nil
	},
}

// float4gt represents the PostgreSQL function of the same name, taking the same parameters.
var float4gt = framework.Function2{
	Name:       "float4gt",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Float32, pgtypes.Float32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Float32.Compare(val1.(float32), val2.(float32))
		return res == 1, err
	},
}

// float48gt represents the PostgreSQL function of the same name, taking the same parameters.
var float48gt = framework.Function2{
	Name:       "float48gt",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Float32, pgtypes.Float64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Float64.Compare(float64(val1.(float32)), val2.(float64))
		return res == 1, err
	},
}

// float84gt represents the PostgreSQL function of the same name, taking the same parameters.
var float84gt = framework.Function2{
	Name:       "float84gt",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Float64, pgtypes.Float32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Float64.Compare(val1.(float64), float64(val2.(float32)))
		return res == 1, err
	},
}

// float8gt represents the PostgreSQL function of the same name, taking the same parameters.
var float8gt = framework.Function2{
	Name:       "float8gt",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Float64, pgtypes.Float64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Float64.Compare(val1.(float64), val2.(float64))
		return res == 1, err
	},
}

// int2gt represents the PostgreSQL function of the same name, taking the same parameters.
var int2gt = framework.Function2{
	Name:       "int2gt",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int16},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int16.Compare(val1.(int16), val2.(int16))
		return res == 1, err
	},
}

// int24gt represents the PostgreSQL function of the same name, taking the same parameters.
var int24gt = framework.Function2{
	Name:       "int24gt",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int32.Compare(int32(val1.(int16)), val2.(int32))
		return res == 1, err
	},
}

// int28gt represents the PostgreSQL function of the same name, taking the same parameters.
var int28gt = framework.Function2{
	Name:       "int28gt",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int64.Compare(int64(val1.(int16)), val2.(int64))
		return res == 1, err
	},
}

// int42gt represents the PostgreSQL function of the same name, taking the same parameters.
var int42gt = framework.Function2{
	Name:       "int42gt",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int16},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int32.Compare(val1.(int32), int32(val2.(int16)))
		return res == 1, err
	},
}

// int4gt represents the PostgreSQL function of the same name, taking the same parameters.
var int4gt = framework.Function2{
	Name:       "int4gt",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int32.Compare(val1.(int32), val2.(int32))
		return res == 1, err
	},
}

// int48gt represents the PostgreSQL function of the same name, taking the same parameters.
var int48gt = framework.Function2{
	Name:       "int48gt",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int64.Compare(int64(val1.(int32)), val2.(int64))
		return res == 1, err
	},
}

// int82gt represents the PostgreSQL function of the same name, taking the same parameters.
var int82gt = framework.Function2{
	Name:       "int82gt",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int16},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int64.Compare(val1.(int64), int64(val2.(int16)))
		return res == 1, err
	},
}

// int84gt represents the PostgreSQL function of the same name, taking the same parameters.
var int84gt = framework.Function2{
	Name:       "int84gt",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int64.Compare(val1.(int64), int64(val2.(int32)))
		return res == 1, err
	},
}

// int8gt represents the PostgreSQL function of the same name, taking the same parameters.
var int8gt = framework.Function2{
	Name:       "int8gt",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int64.Compare(val1.(int64), val2.(int64))
		return res == 1, err
	},
}

// interval_gt represents the PostgreSQL function of the same name, taking the same parameters.
var interval_gt = framework.Function2{
	Name:       "interval_gt",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Interval, pgtypes.Interval},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Interval.Compare(val1.(duration.Duration), val2.(duration.Duration))
		return res == 1, err
	},
}

// jsonb_gt represents the PostgreSQL function of the same name, taking the same parameters.
var jsonb_gt = framework.Function2{
	Name:       "jsonb_gt",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.JsonB, pgtypes.JsonB},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.JsonB.Compare(val1.(pgtypes.JsonDocument), val2.(pgtypes.JsonDocument))
		return res == 1, err
	},
}

// namegt represents the PostgreSQL function of the same name, taking the same parameters.
var namegt = framework.Function2{
	Name:       "namegt",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Name, pgtypes.Name},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Name.Compare(val1.(string), val2.(string))
		return res == 1, err
	},
}

// namegttext represents the PostgreSQL function of the same name, taking the same parameters.
var namegttext = framework.Function2{
	Name:       "namegttext",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Name, pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Text.Compare(val1.(string), val2.(string))
		return res == 1, err
	},
}

// numeric_gt represents the PostgreSQL function of the same name, taking the same parameters.
var numeric_gt = framework.Function2{
	Name:       "numeric_gt",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Numeric, pgtypes.Numeric},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Numeric.Compare(val1.(decimal.Decimal), val2.(decimal.Decimal))
		return res == 1, err
	},
}

// oidgt represents the PostgreSQL function of the same name, taking the same parameters.
var oidgt = framework.Function2{
	Name:       "oidgt",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Oid, pgtypes.Oid},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Oid.Compare(val1.(uint32), val2.(uint32))
		return res == 1, err
	},
}

// textgtname represents the PostgreSQL function of the same name, taking the same parameters.
var textgtname = framework.Function2{
	Name:       "textgtname",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Text, pgtypes.Name},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Text.Compare(val1.(string), val2.(string))
		return res == 1, err
	},
}

// text_gt represents the PostgreSQL function of the same name, taking the same parameters.
var text_gt = framework.Function2{
	Name:       "text_gt",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Text, pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Text.Compare(val1.(string), val2.(string))
		return res == 1, err
	},
}

// time_gt represents the PostgreSQL function of the same name, taking the same parameters.
var time_gt = framework.Function2{
	Name:       "time_gt",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Time, pgtypes.Time},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Time.Compare(val1.(time.Time), val2.(time.Time))
		return res == 1, err
	},
}

// timestamp_gt_date represents the PostgreSQL function of the same name, taking the same parameters.
var timestamp_gt_date = framework.Function2{
	Name:       "timestamp_gt_date",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Timestamp, pgtypes.Date},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res := val1.(time.Time).Compare(val2.(time.Time))
		return res == 1, nil
	},
}

// timestamp_gt represents the PostgreSQL function of the same name, taking the same parameters.
var timestamp_gt = framework.Function2{
	Name:       "timestamp_gt",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Timestamp, pgtypes.Timestamp},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Timestamp.Compare(val1.(time.Time), val2.(time.Time))
		return res == 1, err
	},
}

// timestamp_gt_timestamptz represents the PostgreSQL function of the same name, taking the same parameters.
var timestamp_gt_timestamptz = framework.Function2{
	Name:       "timestamp_gt_timestamptz",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Timestamp, pgtypes.TimestampTZ},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.TimestampTZ.Compare(val1.(time.Time), val2.(time.Time))
		return res == 1, err
	},
}

// timestamptz_gt_date represents the PostgreSQL function of the same name, taking the same parameters.
var timestamptz_gt_date = framework.Function2{
	Name:       "timestamptz_gt_date",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.TimestampTZ, pgtypes.Date},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res := val1.(time.Time).Compare(val2.(time.Time))
		return res == 1, nil
	},
}

// timestamptz_gt_timestamp represents the PostgreSQL function of the same name, taking the same parameters.
var timestamptz_gt_timestamp = framework.Function2{
	Name:       "timestamptz_gt_timestamp",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.TimestampTZ, pgtypes.Timestamp},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.TimestampTZ.Compare(val1.(time.Time), val2.(time.Time))
		return res == 1, err
	},
}

// timestamptz_gt represents the PostgreSQL function of the same name, taking the same parameters.
var timestamptz_gt = framework.Function2{
	Name:       "timestamptz_gt",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.TimestampTZ, pgtypes.TimestampTZ},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.TimestampTZ.Compare(val1.(time.Time), val2.(time.Time))
		return res == 1, err
	},
}

// timetz_gt represents the PostgreSQL function of the same name, taking the same parameters.
var timetz_gt = framework.Function2{
	Name:       "timetz_gt",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.TimeTZ, pgtypes.TimeTZ},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.TimeTZ.Compare(val1.(time.Time), val2.(time.Time))
		return res == 1, err
	},
}

// uuid_gt represents the PostgreSQL function of the same name, taking the same parameters.
var uuid_gt = framework.Function2{
	Name:       "uuid_gt",
	Return:     pgtypes.Bool,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Uuid, pgtypes.Uuid},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Uuid.Compare(val1.(uuid.UUID), val2.(uuid.UUID))
		return res == 1, err
	},
}
