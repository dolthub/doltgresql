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

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/postgres/parser/duration"
	"github.com/dolthub/doltgresql/postgres/parser/uuid"
	"github.com/dolthub/doltgresql/server/compare"
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
	framework.RegisterBinaryFunction(framework.Operator_BinaryEqual, enum_eq)
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
	framework.RegisterBinaryFunction(framework.Operator_BinaryEqual, record_eq)
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

// booleq_callable is the callable logic for the booleq function.
func booleq_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	res, err := pgtypes.Bool.Compare(ctx, val1.(bool), val2.(bool))
	return res == 0, err
}

// booleq represents the PostgreSQL function of the same name, taking the same parameters.
var booleq = framework.Function2{
	Name:       "booleq",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Bool, pgtypes.Bool},
	Strict:     true,
	Callable:   booleq_callable,
}

// bpchareq_callable is the callable logic for the bpchareq function.
func bpchareq_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	res, err := pgtypes.BpChar.Compare(ctx, val1.(string), val2.(string))
	return res == 0, err
}

// bpchareq represents the PostgreSQL function of the same name, taking the same parameters.
var bpchareq = framework.Function2{
	Name:       "bpchareq",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.BpChar, pgtypes.BpChar},
	Strict:     true,
	Callable:   bpchareq_callable,
}

// byteaeq_callable is the callable logic for the byteaeq function.
func byteaeq_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	res, err := pgtypes.Bytea.Compare(ctx, val1.([]byte), val2.([]byte))
	return res == 0, err
}

// byteaeq represents the PostgreSQL function of the same name, taking the same parameters.
var byteaeq = framework.Function2{
	Name:       "byteaeq",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Bytea, pgtypes.Bytea},
	Strict:     true,
	Callable:   byteaeq_callable,
}

// chareq_callable is the callable logic for the chareq function.
func chareq_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	res, err := pgtypes.InternalChar.Compare(ctx, val1.(string), val2.(string))
	return res == 0, err
}

// chareq represents the PostgreSQL function of the same name, taking the same parameters.
var chareq = framework.Function2{
	Name:       "chareq",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.InternalChar, pgtypes.InternalChar},
	Strict:     true,
	Callable:   chareq_callable,
}

// date_eq_callable is the callable logic for the date_eq function.
func date_eq_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	res, err := pgtypes.Date.Compare(ctx, val1.(time.Time), val2.(time.Time))
	return res == 0, err
}

// date_eq represents the PostgreSQL function of the same name, taking the same parameters.
var date_eq = framework.Function2{
	Name:       "date_eq",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Date, pgtypes.Date},
	Strict:     true,
	Callable:   date_eq_callable,
}

// date_eq_timestamp_callable is the callable logic for the date_eq_timestamp function.
func date_eq_timestamp_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	res := val1.(time.Time).Compare(val2.(time.Time))
	return res == 0, nil
}

// date_eq_timestamp represents the PostgreSQL function of the same name, taking the same parameters.
var date_eq_timestamp = framework.Function2{
	Name:       "date_eq_timestamp",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Date, pgtypes.Timestamp},
	Strict:     true,
	Callable:   date_eq_timestamp_callable,
}

// date_eq_timestamptz_callable is the callable logic for the date_eq_timestamptz function.
func date_eq_timestamptz_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	res := val1.(time.Time).Compare(val2.(time.Time))
	return res == 0, nil
}

// date_eq_timestamptz represents the PostgreSQL function of the same name, taking the same parameters.
var date_eq_timestamptz = framework.Function2{
	Name:       "date_eq_timestamptz",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Date, pgtypes.TimestampTZ},
	Strict:     true,
	Callable:   date_eq_timestamptz_callable,
}

// enum_eq_callable is the callable logic for the enum_eq function.
func enum_eq_callable(ctx *sql.Context, t [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	// TODO: if given types are not the same enum type, it cannot compare
	res, err := t[0].Compare(ctx, val1, val2)
	return res == 0, err
}

// enum_eq represents the PostgreSQL function of the same name, taking the same parameters.
var enum_eq = framework.Function2{
	Name:       "enum_eq",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.AnyEnum, pgtypes.AnyEnum},
	Strict:     true,
	Callable:   enum_eq_callable,
}

// float4eq_callable is the callable logic for the float4eq function.
func float4eq_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	res, err := pgtypes.Float32.Compare(ctx, val1.(float32), val2.(float32))
	return res == 0, err
}

// float4eq represents the PostgreSQL function of the same name, taking the same parameters.
var float4eq = framework.Function2{
	Name:       "float4eq",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Float32, pgtypes.Float32},
	Strict:     true,
	Callable:   float4eq_callable,
}

// float48eq_callable is the callable logic for the float48eq function.
func float48eq_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	res, err := pgtypes.Float64.Compare(ctx, float64(val1.(float32)), val2.(float64))
	return res == 0, err
}

// float48eq represents the PostgreSQL function of the same name, taking the same parameters.
var float48eq = framework.Function2{
	Name:       "float48eq",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Float32, pgtypes.Float64},
	Strict:     true,
	Callable:   float48eq_callable,
}

// float84eq_callable is the callable logic for the float84eq function.
func float84eq_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	res, err := pgtypes.Float64.Compare(ctx, val1.(float64), float64(val2.(float32)))
	return res == 0, err
}

// float84eq represents the PostgreSQL function of the same name, taking the same parameters.
var float84eq = framework.Function2{
	Name:       "float84eq",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Float64, pgtypes.Float32},
	Strict:     true,
	Callable:   float84eq_callable,
}

// float8eq_callable is the callable logic for the float8eq function.
func float8eq_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	res, err := pgtypes.Float64.Compare(ctx, val1.(float64), val2.(float64))
	return res == 0, err
}

// float8eq represents the PostgreSQL function of the same name, taking the same parameters.
var float8eq = framework.Function2{
	Name:       "float8eq",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Float64, pgtypes.Float64},
	Strict:     true,
	Callable:   float8eq_callable,
}

// int2eq_callable is the callable logic for the int2eq function.
func int2eq_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	res, err := pgtypes.Int16.Compare(ctx, val1.(int16), val2.(int16))
	return res == 0, err
}

// int2eq represents the PostgreSQL function of the same name, taking the same parameters.
var int2eq = framework.Function2{
	Name:       "int2eq",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int16},
	Strict:     true,
	Callable:   int2eq_callable,
}

// int24eq_callable is the callable logic for the int24eq function.
func int24eq_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	res, err := pgtypes.Int32.Compare(ctx, int32(val1.(int16)), val2.(int32))
	return res == 0, err
}

// int24eq represents the PostgreSQL function of the same name, taking the same parameters.
var int24eq = framework.Function2{
	Name:       "int24eq",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int32},
	Strict:     true,
	Callable:   int24eq_callable,
}

// int28eq_callable is the callable logic for the int28eq function.
func int28eq_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	res, err := pgtypes.Int64.Compare(ctx, int64(val1.(int16)), val2.(int64))
	return res == 0, err
}

// int28eq represents the PostgreSQL function of the same name, taking the same parameters.
var int28eq = framework.Function2{
	Name:       "int28eq",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int64},
	Strict:     true,
	Callable:   int28eq_callable,
}

// int42eq_callable is the callable logic for the int42eq function.
func int42eq_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	res, err := pgtypes.Int32.Compare(ctx, val1.(int32), int32(val2.(int16)))
	return res == 0, err
}

// int42eq represents the PostgreSQL function of the same name, taking the same parameters.
var int42eq = framework.Function2{
	Name:       "int42eq",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int16},
	Strict:     true,
	Callable:   int42eq_callable,
}

// int4eq_callable is the callable logic for the int4eq function.
func int4eq_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	res, err := pgtypes.Int32.Compare(ctx, val1.(int32), val2.(int32))
	return res == 0, err
}

// int4eq represents the PostgreSQL function of the same name, taking the same parameters.
var int4eq = framework.Function2{
	Name:       "int4eq",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int32},
	Strict:     true,
	Callable:   int4eq_callable,
}

// int48eq_callable is the callable logic for the int48eq function.
func int48eq_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	res, err := pgtypes.Int64.Compare(ctx, int64(val1.(int32)), val2.(int64))
	return res == 0, err
}

// int48eq represents the PostgreSQL function of the same name, taking the same parameters.
var int48eq = framework.Function2{
	Name:       "int48eq",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int64},
	Strict:     true,
	Callable:   int48eq_callable,
}

// int82eq_callable is the callable logic for the int82eq function.
func int82eq_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	res, err := pgtypes.Int64.Compare(ctx, val1.(int64), int64(val2.(int16)))
	return res == 0, err
}

// int82eq represents the PostgreSQL function of the same name, taking the same parameters.
var int82eq = framework.Function2{
	Name:       "int82eq",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int16},
	Strict:     true,
	Callable:   int82eq_callable,
}

// int84eq_callable is the callable logic for the int84eq function.
func int84eq_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	res, err := pgtypes.Int64.Compare(ctx, val1.(int64), int64(val2.(int32)))
	return res == 0, err
}

// int84eq represents the PostgreSQL function of the same name, taking the same parameters.
var int84eq = framework.Function2{
	Name:       "int84eq",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int32},
	Strict:     true,
	Callable:   int84eq_callable,
}

// int8eq_callable is the callable logic for the int8eq function.
func int8eq_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	res, err := pgtypes.Int64.Compare(ctx, val1.(int64), val2.(int64))
	return res == 0, err
}

// int8eq represents the PostgreSQL function of the same name, taking the same parameters.
var int8eq = framework.Function2{
	Name:       "int8eq",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int64},
	Strict:     true,
	Callable:   int8eq_callable,
}

// interval_eq_callable is the callable logic for the interval_eq function.
func interval_eq_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	res, err := pgtypes.Interval.Compare(ctx, val1.(duration.Duration), val2.(duration.Duration))
	return res == 0, err
}

// interval_eq represents the PostgreSQL function of the same name, taking the same parameters.
var interval_eq = framework.Function2{
	Name:       "interval_eq",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Interval, pgtypes.Interval},
	Strict:     true,
	Callable:   interval_eq_callable,
}

// jsonb_eq_callable is the callable logic for the jsonb_eq function.
func jsonb_eq_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	res, err := pgtypes.JsonB.Compare(ctx, val1.(pgtypes.JsonDocument), val2.(pgtypes.JsonDocument))
	return res == 0, err
}

// jsonb_eq represents the PostgreSQL function of the same name, taking the same parameters.
var jsonb_eq = framework.Function2{
	Name:       "jsonb_eq",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.JsonB, pgtypes.JsonB},
	Strict:     true,
	Callable:   jsonb_eq_callable,
}

// nameeq_callable is the callable logic for the nameeq function.
func nameeq_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	res, err := pgtypes.Name.Compare(ctx, val1.(string), val2.(string))
	return res == 0, err
}

// nameeq represents the PostgreSQL function of the same name, taking the same parameters.
var nameeq = framework.Function2{
	Name:       "nameeq",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Name, pgtypes.Name},
	Strict:     true,
	Callable:   nameeq_callable,
}

// nameeqtext_callable is the callable logic for the nameeqtext function.
func nameeqtext_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	res, err := pgtypes.Text.Compare(ctx, val1.(string), val2.(string))
	return res == 0, err
}

// nameeqtext represents the PostgreSQL function of the same name, taking the same parameters.
var nameeqtext = framework.Function2{
	Name:       "nameeqtext",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Name, pgtypes.Text},
	Strict:     true,
	Callable:   nameeqtext_callable,
}

// numeric_eq_callable is the callable logic for the numeric_eq function.
func numeric_eq_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	res, err := pgtypes.Numeric.Compare(ctx, val1.(decimal.Decimal), val2.(decimal.Decimal))
	return res == 0, err
}

// numeric_eq represents the PostgreSQL function of the same name, taking the same parameters.
var numeric_eq = framework.Function2{
	Name:       "numeric_eq",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Numeric, pgtypes.Numeric},
	Strict:     true,
	Callable:   numeric_eq_callable,
}

// oideq_callable is the callable logic for the oideq function.
// This method doesn't use DotlgresType.Compare because it's on the critical path for many tooling queries that
// examine the pg_catalog tables.
func oideq_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	if val1 == nil || val2 == nil {
		return false, nil
	}

	val1id, val2id := val1.(id.Id), val2.(id.Id)
	return val1id == val2id, nil
}

// oideq represents the PostgreSQL function of the same name, taking the same parameters.
var oideq = framework.Function2{
	Name:       "oideq",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Oid, pgtypes.Oid},
	Strict:     true,
	Callable:   oideq_callable,
}

// texteqname_callable is the callable logic for the texteqname function.
func texteqname_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	res, err := pgtypes.Text.Compare(ctx, val1.(string), val2.(string))
	return res == 0, err
}

// texteqname represents the PostgreSQL function of the same name, taking the same parameters.
var texteqname = framework.Function2{
	Name:       "texteqname",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Name},
	Strict:     true,
	Callable:   texteqname_callable,
}

// text_eq_callable is the callable logic for the text_eq function.
func text_eq_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	res, err := pgtypes.Text.Compare(ctx, val1, val2)
	return res == 0, err
}

// text_eq represents the PostgreSQL function of the same name, taking the same parameters.
var text_eq = framework.Function2{
	Name:       "text_eq",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Text},
	Strict:     true,
	Callable:   text_eq_callable,
}

// record_eq_callable is the callable logic for the record_eq function.
func record_eq_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	return compare.CompareRecords(ctx, framework.Operator_BinaryEqual, val1, val2)
}

// record_eq represents the PostgreSQL function of the same name, taking the same parameters.
var record_eq = framework.Function2{
	Name:       "record_eq",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Record, pgtypes.Record},
	Strict:     true,
	Callable:   record_eq_callable,
}

// time_eq_callable is the callable logic for the time_eq function.
func time_eq_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	res, err := pgtypes.Time.Compare(ctx, val1.(time.Time), val2.(time.Time))
	return res == 0, err
}

// time_eq represents the PostgreSQL function of the same name, taking the same parameters.
var time_eq = framework.Function2{
	Name:       "time_eq",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Time, pgtypes.Time},
	Strict:     true,
	Callable:   time_eq_callable,
}

// timestamp_eq_date_callable is the callable logic for the timestamp_eq_date function.
func timestamp_eq_date_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	res := val1.(time.Time).Compare(val2.(time.Time))
	return res == 0, nil
}

// timestamp_eq_date represents the PostgreSQL function of the same name, taking the same parameters.
var timestamp_eq_date = framework.Function2{
	Name:       "timestamp_eq_date",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Timestamp, pgtypes.Date},
	Strict:     true,
	Callable:   timestamp_eq_date_callable,
}

// timestamp_eq_callable is the callable logic for the timestamp_eq function.
func timestamp_eq_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	res, err := pgtypes.Timestamp.Compare(ctx, val1.(time.Time), val2.(time.Time))
	return res == 0, err
}

// timestamp_eq represents the PostgreSQL function of the same name, taking the same parameters.
var timestamp_eq = framework.Function2{
	Name:       "timestamp_eq",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Timestamp, pgtypes.Timestamp},
	Strict:     true,
	Callable:   timestamp_eq_callable,
}

// timestamp_eq_timestamptz_callable is the callable logic for the timestamp_eq_timestamptz function.
func timestamp_eq_timestamptz_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	res, err := pgtypes.TimestampTZ.Compare(ctx, val1.(time.Time), val2.(time.Time))
	return res == 0, err
}

// timestamp_eq_timestamptz represents the PostgreSQL function of the same name, taking the same parameters.
var timestamp_eq_timestamptz = framework.Function2{
	Name:       "timestamp_eq_timestamptz",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Timestamp, pgtypes.TimestampTZ},
	Strict:     true,
	Callable:   timestamp_eq_timestamptz_callable,
}

// timestamptz_eq_date_callable is the callable logic for the timestamptz_eq_date function.
func timestamptz_eq_date_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	res := val1.(time.Time).Compare(val2.(time.Time))
	return res == 0, nil
}

// timestamptz_eq_date represents the PostgreSQL function of the same name, taking the same parameters.
var timestamptz_eq_date = framework.Function2{
	Name:       "timestamptz_eq_date",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.TimestampTZ, pgtypes.Date},
	Strict:     true,
	Callable:   timestamptz_eq_date_callable,
}

// timestamptz_eq_timestamp_callable is the callable logic for the timestamptz_eq_timestamp function.
func timestamptz_eq_timestamp_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	res, err := pgtypes.TimestampTZ.Compare(ctx, val1.(time.Time), val2.(time.Time))
	return res == 0, err
}

// timestamptz_eq_timestamp represents the PostgreSQL function of the same name, taking the same parameters.
var timestamptz_eq_timestamp = framework.Function2{
	Name:       "timestamptz_eq_timestamp",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.TimestampTZ, pgtypes.Timestamp},
	Strict:     true,
	Callable:   timestamptz_eq_timestamp_callable,
}

// timestamptz_eq_callable is the callable logic for the timestamptz_eq function.
func timestamptz_eq_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	res, err := pgtypes.TimestampTZ.Compare(ctx, val1.(time.Time), val2.(time.Time))
	return res == 0, err
}

// timestamptz_eq represents the PostgreSQL function of the same name, taking the same parameters.
var timestamptz_eq = framework.Function2{
	Name:       "timestamptz_eq",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.TimestampTZ, pgtypes.TimestampTZ},
	Strict:     true,
	Callable:   timestamptz_eq_callable,
}

// timetz_eq_callable is the callable logic for the timetz_eq function.
func timetz_eq_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	res, err := pgtypes.TimeTZ.Compare(ctx, val1.(time.Time), val2.(time.Time))
	return res == 0, err
}

// timetz_eq represents the PostgreSQL function of the same name, taking the same parameters.
var timetz_eq = framework.Function2{
	Name:       "timetz_eq",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.TimeTZ, pgtypes.TimeTZ},
	Strict:     true,
	Callable:   timetz_eq_callable,
}

// uuid_eq_callable is the callable logic for the uuid_eq function.
func uuid_eq_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	res, err := pgtypes.Uuid.Compare(ctx, val1.(uuid.UUID), val2.(uuid.UUID))
	return res == 0, err
}

// uuid_eq represents the PostgreSQL function of the same name, taking the same parameters.
var uuid_eq = framework.Function2{
	Name:       "uuid_eq",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Uuid, pgtypes.Uuid},
	Strict:     true,
	Callable:   uuid_eq_callable,
}

// xideqint4_callable is the callable logic for the xideqint4 function.
func xideqint4_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	// TODO: investigate the edge cases
	res, err := pgtypes.Int64.Compare(ctx, int64(val1.(uint32)), int64(val2.(int32)))
	return res == 0, err
}

// xideqint4 represents the PostgreSQL function of the same name, taking the same parameters.
var xideqint4 = framework.Function2{
	Name:       "xideqint4",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Xid, pgtypes.Int32},
	Strict:     true,
	Callable:   xideqint4_callable,
}

// xideq_callable is the callable logic for the xideq function.
func xideq_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	res, err := pgtypes.Xid.Compare(ctx, val1.(uint32), val2.(uint32))
	return res == 0, err
}

// xideq represents the PostgreSQL function of the same name, taking the same parameters.
var xideq = framework.Function2{
	Name:       "xideq",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Xid, pgtypes.Xid},
	Strict:     true,
	Callable:   xideq_callable,
}
