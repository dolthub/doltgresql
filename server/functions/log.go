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
	"fmt"
	"math"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initLog registers the functions to the catalog.
func initLog() {
	framework.RegisterFunction(log_float64)
	framework.RegisterFunction(log_numeric)
	framework.RegisterFunction(log_numeric_numeric)
}

// log_float64 represents the PostgreSQL function of the same name, taking the same parameters.
var log_float64 = framework.Function1{
	Name:       "log",
	Return:     pgtypes.Float64,
	Parameters: []pgtypes.DoltgresType{pgtypes.Float64},
	Callable: func(ctx *sql.Context, val1Interface any) (any, error) {
		val1 := val1Interface.(float64)
		if val1 == 0 {
			return nil, fmt.Errorf("cannot take logarithm of zero")
		} else if val1 < 0 {
			return nil, fmt.Errorf("cannot take logarithm of a negative number")
		}
		return math.Log10(val1), nil
	},
	Strict: true,
}

// log_numeric represents the PostgreSQL function of the same name, taking the same parameters.
var log_numeric = framework.Function1{
	Name:       "log",
	Return:     pgtypes.Numeric,
	Parameters: []pgtypes.DoltgresType{pgtypes.Numeric},
	Callable: func(ctx *sql.Context, val1Interface any) (any, error) {
		if val1Interface == nil {
			return nil, nil
		}
		val1 := val1Interface.(decimal.Decimal)
		if val1.Equal(decimal.Zero) {
			return nil, fmt.Errorf("cannot take logarithm of zero")
		} else if val1.LessThan(decimal.Zero) {
			return nil, fmt.Errorf("cannot take logarithm of a negative number")
		}
		// TODO: implement log for numeric instead of relying on float64
		f, _ := val1.Float64()
		return decimal.NewFromFloat(math.Log10(f)), nil
	},
	Strict: true,
}

// log_numeric_numeric represents the PostgreSQL function of the same name, taking the same parameters.
var log_numeric_numeric = framework.Function2{
	Name:       "log",
	Return:     pgtypes.Numeric,
	Parameters: []pgtypes.DoltgresType{pgtypes.Numeric, pgtypes.Numeric},
	Callable: func(ctx *sql.Context, val1Interface any, val2Interface any) (any, error) {
		if val1Interface == nil || val2Interface == nil {
			return nil, nil
		}
		val1 := val1Interface.(decimal.Decimal)
		val2 := val2Interface.(decimal.Decimal)
		if val1.Equal(decimal.Zero) || val2.Equal(decimal.Zero) {
			return nil, fmt.Errorf("cannot take logarithm of zero")
		} else if val1.LessThan(decimal.Zero) || val2.LessThan(decimal.Zero) {
			return nil, fmt.Errorf("cannot take logarithm of a negative number")
		}
		// TODO: implement log for numeric instead of relying on float64
		base, _ := val1.Float64()
		num, _ := val2.Float64()
		logNum := math.Log(num)
		logBase := math.Log(base)
		if logBase == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return decimal.NewFromFloat(logNum / logBase), nil
	},
	Strict: true,
}
