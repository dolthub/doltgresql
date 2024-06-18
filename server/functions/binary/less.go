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
// SELECT * FROM pg_operator o WHERE o.oprname = '<' ORDER BY o.oprcode::varchar;

// initBinaryLessThan registers the functions to the catalog.
func initBinaryLessThan() {
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessThan, boollt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessThan, bpcharlt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessThan, bytealt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessThan, date_lt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessThan, date_lt_timestamp)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessThan, date_lt_timestamptz)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessThan, float4lt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessThan, float48lt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessThan, float84lt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessThan, float8lt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessThan, int2lt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessThan, int24lt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessThan, int28lt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessThan, int42lt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessThan, int4lt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessThan, int48lt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessThan, int82lt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessThan, int84lt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessThan, int8lt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessThan, jsonb_lt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessThan, namelt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessThan, namelttext)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessThan, numeric_lt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessThan, oidlt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessThan, textltname)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessThan, text_lt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessThan, time_lt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessThan, timestamp_lt_date)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessThan, timestamp_lt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessThan, timestamp_lt_timestamptz)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessThan, timestamptz_lt_date)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessThan, timestamptz_lt_timestamp)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessThan, timestamptz_lt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessThan, timetz_lt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessThan, uuid_lt)
}

// boollt represents the PostgreSQL function of the same name, taking the same parameters.
var boollt = framework.Function2{
	Name:       "boollt",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Bool, pgtypes.Bool},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Bool.Compare(val1.(bool), val2.(bool))
		return res == -1, err
	},
	Strict: true,
}

// bpcharlt represents the PostgreSQL function of the same name, taking the same parameters.
var bpcharlt = framework.Function2{
	Name:       "bpcharlt",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.BpChar, pgtypes.BpChar},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.BpChar.Compare(val1.(string), val2.(string))
		return res == -1, err
	},
	Strict: true,
}

// bytealt represents the PostgreSQL function of the same name, taking the same parameters.
var bytealt = framework.Function2{
	Name:       "bytealt",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Bytea, pgtypes.Bytea},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Bytea.Compare(val1.([]byte), val2.([]byte))
		return res == -1, err
	},
	Strict: true,
}

// date_lt represents the PostgreSQL function of the same name, taking the same parameters.
var date_lt = framework.Function2{
	Name:       "date_lt",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Date, pgtypes.Date},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Date.Compare(val1.(time.Time), val2.(time.Time))
		return res == -1, err
	},
	Strict: true,
}

// date_lt_timestamp represents the PostgreSQL function of the same name, taking the same parameters.
var date_lt_timestamp = framework.Function2{
	Name:       "date_lt_timestamp",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Date, pgtypes.Timestamp},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res := val1.(time.Time).Compare(val2.(time.Time))
		return res == -1, nil
	},
	Strict: true,
}

// date_lt_timestamptz represents the PostgreSQL function of the same name, taking the same parameters.
var date_lt_timestamptz = framework.Function2{
	Name:       "date_lt_timestamptz",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Date, pgtypes.TimestampTZ},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res := val1.(time.Time).Compare(val2.(time.Time))
		return res == -1, nil
	},
	Strict: true,
}

// float4lt represents the PostgreSQL function of the same name, taking the same parameters.
var float4lt = framework.Function2{
	Name:       "float4lt",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Float32, pgtypes.Float32},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Float32.Compare(val1.(float32), val2.(float32))
		return res == -1, err
	},
	Strict: true,
}

// float48lt represents the PostgreSQL function of the same name, taking the same parameters.
var float48lt = framework.Function2{
	Name:       "float48lt",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Float32, pgtypes.Float64},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Float64.Compare(float64(val1.(float32)), val2.(float64))
		return res == -1, err
	},
	Strict: true,
}

// float84lt represents the PostgreSQL function of the same name, taking the same parameters.
var float84lt = framework.Function2{
	Name:       "float84lt",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Float64, pgtypes.Float32},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Float64.Compare(val1.(float64), float64(val2.(float32)))
		return res == -1, err
	},
	Strict: true,
}

// float8lt represents the PostgreSQL function of the same name, taking the same parameters.
var float8lt = framework.Function2{
	Name:       "float8lt",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Float64, pgtypes.Float64},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Float64.Compare(val1.(float64), val2.(float64))
		return res == -1, err
	},
	Strict: true,
}

// int2lt represents the PostgreSQL function of the same name, taking the same parameters.
var int2lt = framework.Function2{
	Name:       "int2lt",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int16},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int16.Compare(val1.(int16), val2.(int16))
		return res == -1, err
	},
	Strict: true,
}

// int24lt represents the PostgreSQL function of the same name, taking the same parameters.
var int24lt = framework.Function2{
	Name:       "int24lt",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int32},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int32.Compare(int32(val1.(int16)), val2.(int32))
		return res == -1, err
	},
	Strict: true,
}

// int28lt represents the PostgreSQL function of the same name, taking the same parameters.
var int28lt = framework.Function2{
	Name:       "int28lt",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int64},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int64.Compare(int64(val1.(int16)), val2.(int64))
		return res == -1, err
	},
	Strict: true,
}

// int42lt represents the PostgreSQL function of the same name, taking the same parameters.
var int42lt = framework.Function2{
	Name:       "int42lt",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int16},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int32.Compare(val1.(int32), int32(val2.(int16)))
		return res == -1, err
	},
	Strict: true,
}

// int4lt represents the PostgreSQL function of the same name, taking the same parameters.
var int4lt = framework.Function2{
	Name:       "int4lt",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int32},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int32.Compare(val1.(int32), val2.(int32))
		return res == -1, err
	},
	Strict: true,
}

// int48lt represents the PostgreSQL function of the same name, taking the same parameters.
var int48lt = framework.Function2{
	Name:       "int48lt",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int64},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int64.Compare(int64(val1.(int32)), val2.(int64))
		return res == -1, err
	},
	Strict: true,
}

// int82lt represents the PostgreSQL function of the same name, taking the same parameters.
var int82lt = framework.Function2{
	Name:       "int82lt",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int16},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int64.Compare(val1.(int64), int64(val2.(int16)))
		return res == -1, err
	},
	Strict: true,
}

// int84lt represents the PostgreSQL function of the same name, taking the same parameters.
var int84lt = framework.Function2{
	Name:       "int84lt",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int32},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int64.Compare(val1.(int64), int64(val2.(int32)))
		return res == -1, err
	},
	Strict: true,
}

// int8lt represents the PostgreSQL function of the same name, taking the same parameters.
var int8lt = framework.Function2{
	Name:       "int8lt",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int64},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Int64.Compare(val1.(int64), val2.(int64))
		return res == -1, err
	},
	Strict: true,
}

// jsonb_lt represents the PostgreSQL function of the same name, taking the same parameters.
var jsonb_lt = framework.Function2{
	Name:       "jsonb_lt",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.JsonB, pgtypes.JsonB},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.JsonB.Compare(val1.(pgtypes.JsonDocument), val2.(pgtypes.JsonDocument))
		return res == -1, err
	},
	Strict: true,
}

// namelt represents the PostgreSQL function of the same name, taking the same parameters.
var namelt = framework.Function2{
	Name:       "namelt",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Name, pgtypes.Name},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Name.Compare(val1.(string), val2.(string))
		return res == -1, err
	},
	Strict: true,
}

// namelttext represents the PostgreSQL function of the same name, taking the same parameters.
var namelttext = framework.Function2{
	Name:       "namelttext",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Name, pgtypes.Text},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Text.Compare(val1.(string), val2.(string))
		return res == -1, err
	},
	Strict: true,
}

// numeric_lt represents the PostgreSQL function of the same name, taking the same parameters.
var numeric_lt = framework.Function2{
	Name:       "numeric_lt",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Numeric, pgtypes.Numeric},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Numeric.Compare(val1.(decimal.Decimal), val2.(decimal.Decimal))
		return res == -1, err
	},
	Strict: true,
}

// oidlt represents the PostgreSQL function of the same name, taking the same parameters.
var oidlt = framework.Function2{
	Name:       "oidlt",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Oid, pgtypes.Oid},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Oid.Compare(val1.(uint32), val2.(uint32))
		return res == -1, err
	},
	Strict: true,
}

// textltname represents the PostgreSQL function of the same name, taking the same parameters.
var textltname = framework.Function2{
	Name:       "textltname",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Text, pgtypes.Name},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Text.Compare(val1.(string), val2.(string))
		return res == -1, err
	},
	Strict: true,
}

// text_lt represents the PostgreSQL function of the same name, taking the same parameters.
var text_lt = framework.Function2{
	Name:       "text_lt",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Text, pgtypes.Text},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Text.Compare(val1.(string), val2.(string))
		return res == -1, err
	},
	Strict: true,
}

// time_lt represents the PostgreSQL function of the same name, taking the same parameters.
var time_lt = framework.Function2{
	Name:       "time_lt",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Time, pgtypes.Time},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Time.Compare(val1.(time.Time), val2.(time.Time))
		return res == -1, err
	},
	Strict: true,
}

// timestamp_lt_date represents the PostgreSQL function of the same name, taking the same parameters.
var timestamp_lt_date = framework.Function2{
	Name:       "timestamp_lt_date",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Timestamp, pgtypes.Date},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res := val1.(time.Time).Compare(val2.(time.Time))
		return res == -1, nil
	},
	Strict: true,
}

// timestamp_lt represents the PostgreSQL function of the same name, taking the same parameters.
var timestamp_lt = framework.Function2{
	Name:       "timestamp_lt",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Timestamp, pgtypes.Timestamp},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Timestamp.Compare(val1.(time.Time), val2.(time.Time))
		return res == -1, err
	},
	Strict: true,
}

// timestamp_lt_timestamptz represents the PostgreSQL function of the same name, taking the same parameters.
var timestamp_lt_timestamptz = framework.Function2{
	Name:       "timestamp_lt_timestamptz",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Timestamp, pgtypes.TimestampTZ},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.TimestampTZ.Compare(val1.(time.Time), val2.(time.Time))
		return res == -1, err
	},
	Strict: true,
}

// timestamptz_lt_date represents the PostgreSQL function of the same name, taking the same parameters.
var timestamptz_lt_date = framework.Function2{
	Name:       "timestamptz_lt_date",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.TimestampTZ, pgtypes.Date},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res := val1.(time.Time).Compare(val2.(time.Time))
		return res == -1, nil
	},
	Strict: true,
}

// timestamptz_lt_timestamp represents the PostgreSQL function of the same name, taking the same parameters.
var timestamptz_lt_timestamp = framework.Function2{
	Name:       "timestamptz_lt_timestamp",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.TimestampTZ, pgtypes.Timestamp},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.TimestampTZ.Compare(val1.(time.Time), val2.(time.Time))
		return res == -1, err
	},
	Strict: true,
}

// timestamptz_lt represents the PostgreSQL function of the same name, taking the same parameters.
var timestamptz_lt = framework.Function2{
	Name:       "timestamptz_lt",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.TimestampTZ, pgtypes.TimestampTZ},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.TimestampTZ.Compare(val1.(time.Time), val2.(time.Time))
		return res == -1, err
	},
	Strict: true,
}

// timetz_lt represents the PostgreSQL function of the same name, taking the same parameters.
var timetz_lt = framework.Function2{
	Name:       "timetz_lt",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.TimeTZ, pgtypes.TimeTZ},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.TimeTZ.Compare(val1.(time.Time), val2.(time.Time))
		return res == -1, err
	},
	Strict: true,
}

// uuid_lt represents the PostgreSQL function of the same name, taking the same parameters.
var uuid_lt = framework.Function2{
	Name:       "uuid_lt",
	Return:     pgtypes.Bool,
	Parameters: []pgtypes.DoltgresType{pgtypes.Uuid, pgtypes.Uuid},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		res, err := pgtypes.Uuid.Compare(val1.(uuid.UUID), val2.(uuid.UUID))
		return res == -1, err
	},
	Strict: true,
}
