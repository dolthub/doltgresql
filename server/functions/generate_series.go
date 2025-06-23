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
	"gopkg.in/src-d/go-errors.v1"
	"math"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

var ErrStepSizeCannotEqualZero = errors.NewKind("step size cannot equal zero")

// initGenerateSeries registers the functions to the catalog.
func initGenerateSeries() {
	framework.RegisterFunction(generate_series_int32_int32)
	framework.RegisterFunction(generate_series_int32_int32_int32)
	framework.RegisterFunction(generate_series_int64_int64)
	framework.RegisterFunction(generate_series_int64_int64_int64)
	framework.RegisterFunction(generate_series_numeric_numeric)
	framework.RegisterFunction(generate_series_numeric_numeric_numeric)
}

// generate_series_int32_int32 represents the PostgreSQL function of the same name, taking the same parameters.
var generate_series_int32_int32 = framework.Function2{
	Name:       "generate_series",
	Return:     pgtypes.Row,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int32},
	Strict:     true,
	SRF:        true,
	Callable: func(ctx *sql.Context, t [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		start := val1.(int32)
		finish := val2.(int32)
		step := int32(1) // by default
		count := int64(math.Floor(float64(finish-start+step) / float64(step)))

		rows := make([]any, count)
		if start > finish {
			return nil, nil
		}

		for i := 0; start <= finish; i++ {
			rows[i] = start
			start += step
		}
		return pgtypes.NewRowValues(rows, pgtypes.Int32, count), nil
	},
}

// generate_series_int32_int32_int32 represents the PostgreSQL function of the same name, taking the same parameters.
var generate_series_int32_int32_int32 = framework.Function3{
	Name:       "generate_series",
	Return:     pgtypes.Row,
	Parameters: [3]*pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int32, pgtypes.Int32},
	Strict:     true,
	SRF:        true,
	Callable: func(ctx *sql.Context, t [4]*pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		start := val1.(int32)
		finish := val2.(int32)
		step := val3.(int32)
		if step == 0 {
			return nil, ErrStepSizeCannotEqualZero.New()
		}
		count := int64(math.Floor(float64(finish-start+step) / float64(step)))

		if step > 0 && start > finish || step < 0 && start < finish {
			return nil, nil
		}

		rows := make([]any, count)
		if step > 0 {
			for i := 0; start <= finish; i++ {
				rows[i] = start
				start += step
			}
		} else {
			for i := 0; start >= finish; i++ {
				rows[i] = start
				start += step
			}
		}

		return pgtypes.NewRowValues(rows, pgtypes.Int32, count), nil
	},
}

// generate_series_int64_int64 represents the PostgreSQL function of the same name, taking the same parameters.
var generate_series_int64_int64 = framework.Function2{
	Name:       "generate_series",
	Return:     pgtypes.Row,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int64},
	Strict:     true,
	SRF:        true,
	Callable: func(ctx *sql.Context, t [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		start := val1.(int64)
		finish := val2.(int64)
		step := int64(1) // by default
		count := int64(math.Floor(float64(finish-start+step) / float64(step)))

		rows := make([]any, count)
		if start > finish {
			return nil, nil
		}

		for i := 0; start <= finish; i++ {
			rows[i] = start
			start += step
		}
		return pgtypes.NewRowValues(rows, pgtypes.Int64, count), nil
	},
}

// generate_series_int64_int64_int64 represents the PostgreSQL function of the same name, taking the same parameters.
var generate_series_int64_int64_int64 = framework.Function3{
	Name:       "generate_series",
	Return:     pgtypes.Row,
	Parameters: [3]*pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int64, pgtypes.Int64},
	Strict:     true,
	SRF:        true,
	Callable: func(ctx *sql.Context, t [4]*pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		start := val1.(int64)
		finish := val2.(int64)
		step := val3.(int64)
		if step == 0 {
			return nil, ErrStepSizeCannotEqualZero.New()
		}
		count := int64(math.Floor(float64(finish-start+step) / float64(step)))

		if step > 0 && start > finish || step < 0 && start < finish {
			return nil, nil
		}

		rows := make([]any, count)
		if step > 0 {
			for i := 0; start <= finish; i++ {
				rows[i] = start
				start += step
			}
		} else {
			for i := 0; start >= finish; i++ {
				rows[i] = start
				start += step
			}
		}

		return pgtypes.NewRowValues(rows, pgtypes.Int64, count), nil
	},
}

// generate_series_numeric_numeric represents the PostgreSQL function of the same name, taking the same parameters.
var generate_series_numeric_numeric = framework.Function2{
	Name:       "generate_series",
	Return:     pgtypes.Row,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Numeric, pgtypes.Numeric},
	Strict:     true,
	SRF:        true,
	Callable: func(ctx *sql.Context, t [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		start := val1.(decimal.Decimal)
		finish := val2.(decimal.Decimal)
		step := decimal.NewFromInt(1) // by default
		count := (finish.Sub(start).Add(step)).Div(step).Floor().IntPart()

		rows := make([]any, count)
		if start.GreaterThan(finish) {
			return nil, nil
		}

		for i := 0; start.GreaterThanOrEqual(finish); i++ {
			rows[i] = start
			start = start.Add(step)
		}
		return pgtypes.NewRowValues(rows, pgtypes.Numeric, count), nil
	},
}

// generate_series_numeric_numeric_numeric represents the PostgreSQL function of the same name, taking the same parameters.
var generate_series_numeric_numeric_numeric = framework.Function3{
	Name:       "generate_series",
	Return:     pgtypes.Row,
	Parameters: [3]*pgtypes.DoltgresType{pgtypes.Numeric, pgtypes.Numeric, pgtypes.Numeric},
	Strict:     true,
	SRF:        true,
	Callable: func(ctx *sql.Context, t [4]*pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		start := val1.(decimal.Decimal)
		finish := val2.(decimal.Decimal)
		step := val3.(decimal.Decimal)
		if step == decimal.Zero {
			return nil, ErrStepSizeCannotEqualZero.New()
		}

		count := (finish.Sub(start).Add(step)).Div(step).Floor().IntPart()

		if step.GreaterThan(decimal.Zero) && start.GreaterThan(finish) || step.LessThan(decimal.Zero) && start.LessThan(finish) {
			return nil, nil
		}

		rows := make([]any, count)
		if step.GreaterThan(decimal.Zero) {
			for i := 0; start.GreaterThanOrEqual(finish); i++ {
				rows[i] = start
				start = start.Add(step)
			}
		} else {
			for i := 0; start.LessThanOrEqual(finish); i++ {
				rows[i] = start
				start = start.Add(step)
			}
		}

		return pgtypes.NewRowValues(rows, pgtypes.Numeric, count), nil
	},
}
