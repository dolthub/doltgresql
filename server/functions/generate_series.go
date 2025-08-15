// Copyright 2025 Dolthub, Inc.
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
	"io"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/postgres/parser/duration"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initGenerateSeries registers the functions to the catalog.
func initGenerateSeries() {
	framework.RegisterFunction(generate_series_int32_int32)
	framework.RegisterFunction(generate_series_int32_int32_int32)
	framework.RegisterFunction(generate_series_int64_int64)
	framework.RegisterFunction(generate_series_int64_int64_int64)
	framework.RegisterFunction(generate_series_numeric_numeric)
	framework.RegisterFunction(generate_series_numeric_numeric_numeric)
	framework.RegisterFunction(generate_series_timestamp_timestamp_interval)
}

// errStepSizeZero is an error for a step size of zero in the generate_series functions.
var errStepSizeZero = errors.New("step size cannot equal zero")

// generate_series_int32_int32 represents the PostgreSQL function of the same name, taking the same parameters.
var generate_series_int32_int32 = framework.Function2{
	Name:       "generate_series",
	Return:     pgtypes.RowTypeWithReturnType(pgtypes.Int32),
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int32},
	Strict:     true,
	SRF:        true,
	Callable: func(ctx *sql.Context, t [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		start := val1.(int32)
		finish := val2.(int32)
		step := int32(1) // by default
		return int32GenerateSeries(start, finish, step)
	},
}

// generate_series_int32_int32_int32 represents the PostgreSQL function of the same name, taking the same parameters.
var generate_series_int32_int32_int32 = framework.Function3{
	Name:       "generate_series",
	Return:     pgtypes.RowTypeWithReturnType(pgtypes.Int32),
	Parameters: [3]*pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int32, pgtypes.Int32},
	Strict:     true,
	SRF:        true,
	Callable: func(ctx *sql.Context, t [4]*pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		start := val1.(int32)
		finish := val2.(int32)
		step := val3.(int32)
		return int32GenerateSeries(start, finish, step)
	},
}

func int32GenerateSeries(start, finish, step int32) (*pgtypes.SetReturningFunctionRowIter, error) {
	if step == 0 {
		return nil, errStepSizeZero
	}
	return pgtypes.NewSetReturningFunctionRowIter(func(ctx *sql.Context) (sql.Row, error) {
		defer func() {
			start += step
		}()
		if (step > 0 && start > finish) || (step < 0 && start < finish) {
			return nil, io.EOF
		}
		return sql.Row{start}, nil
	}), nil
}

// generate_series_int64_int64 represents the PostgreSQL function of the same name, taking the same parameters.
var generate_series_int64_int64 = framework.Function2{
	Name:       "generate_series",
	Return:     pgtypes.RowTypeWithReturnType(pgtypes.Int64),
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int64},
	Strict:     true,
	SRF:        true,
	Callable: func(ctx *sql.Context, t [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		start := val1.(int64)
		finish := val2.(int64)
		step := int64(1) // by default
		return int64GenerateSeries(start, finish, step)
	},
}

// generate_series_int64_int64_int64 represents the PostgreSQL function of the same name, taking the same parameters.
var generate_series_int64_int64_int64 = framework.Function3{
	Name:       "generate_series",
	Return:     pgtypes.RowTypeWithReturnType(pgtypes.Int64),
	Parameters: [3]*pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int64, pgtypes.Int64},
	Strict:     true,
	SRF:        true,
	Callable: func(ctx *sql.Context, t [4]*pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		start := val1.(int64)
		finish := val2.(int64)
		step := val3.(int64)
		return int64GenerateSeries(start, finish, step)
	},
}

func int64GenerateSeries(start, finish, step int64) (*pgtypes.SetReturningFunctionRowIter, error) {
	if step == 0 {
		return nil, errStepSizeZero
	}
	return pgtypes.NewSetReturningFunctionRowIter(func(ctx *sql.Context) (sql.Row, error) {
		defer func() {
			start += step
		}()
		if (step > 0 && start > finish) || (step < 0 && start < finish) {
			return nil, io.EOF
		}
		return sql.Row{start}, nil
	}), nil
}

// generate_series_numeric_numeric represents the PostgreSQL function of the same name, taking the same parameters.
var generate_series_numeric_numeric = framework.Function2{
	Name:       "generate_series",
	Return:     pgtypes.RowTypeWithReturnType(pgtypes.Numeric),
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Numeric, pgtypes.Numeric},
	Strict:     true,
	SRF:        true,
	Callable: func(ctx *sql.Context, t [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		start := val1.(decimal.Decimal)
		finish := val2.(decimal.Decimal)
		step := numericOne // by default
		return numericGenerateSeries(start, finish, step)
	},
}

// generate_series_numeric_numeric_numeric represents the PostgreSQL function of the same name, taking the same parameters.
var generate_series_numeric_numeric_numeric = framework.Function3{
	Name:       "generate_series",
	Return:     pgtypes.RowTypeWithReturnType(pgtypes.Numeric),
	Parameters: [3]*pgtypes.DoltgresType{pgtypes.Numeric, pgtypes.Numeric, pgtypes.Numeric},
	Strict:     true,
	SRF:        true,
	Callable: func(ctx *sql.Context, t [4]*pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		start := val1.(decimal.Decimal)
		finish := val2.(decimal.Decimal)
		step := val3.(decimal.Decimal)
		return numericGenerateSeries(start, finish, step)
	},
}

func numericGenerateSeries(start, finish, step decimal.Decimal) (*pgtypes.SetReturningFunctionRowIter, error) {
	if step.Equal(decimal.Zero) {
		return nil, errStepSizeZero
	}
	return pgtypes.NewSetReturningFunctionRowIter(func(ctx *sql.Context) (sql.Row, error) {
		defer func() {
			start = start.Add(step)
		}()
		if (step.GreaterThan(decimal.Zero) && start.GreaterThan(finish)) ||
			(step.LessThan(decimal.Zero) && start.LessThan(finish)) {
			return nil, io.EOF
		}
		return sql.Row{start}, nil
	}), nil
}

// generate_series_timestamp_timestamp_interval represents the PostgreSQL function of the same name, taking the same parameters.
var generate_series_timestamp_timestamp_interval = framework.Function3{
	Name:       "generate_series",
	Return:     pgtypes.RowTypeWithReturnType(pgtypes.Timestamp),
	Parameters: [3]*pgtypes.DoltgresType{pgtypes.Timestamp, pgtypes.Timestamp, pgtypes.Interval},
	Strict:     true,
	SRF:        true,
	Callable: func(ctx *sql.Context, t [4]*pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		start := val1.(time.Time)
		finish := val2.(time.Time)
		step := val3.(duration.Duration)
		stepInt, ok := step.AsInt64()
		if !ok {
			// TODO: overflown
			return nil, errors.Errorf("step argument of generate_series function is overflown")
		}

		return pgtypes.NewSetReturningFunctionRowIter(func(ctx *sql.Context) (sql.Row, error) {
			defer func() {
				start = start.Add(time.Duration(stepInt) * time.Second)
			}()
			if (stepInt > 0 && start.After(finish)) || (stepInt < 0 && start.Before(finish)) {
				return nil, io.EOF
			}
			return sql.Row{start}, nil
		}), nil
	},
}
