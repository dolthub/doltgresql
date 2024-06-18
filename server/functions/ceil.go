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
	"math"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initCeil registers the functions to the catalog.
func initCeil() {
	framework.RegisterFunction(ceil_float64)
	framework.RegisterFunction(ceil_numeric)
	// Register aliases
	ceiling_float64 := ceil_float64
	ceiling_numeric := ceil_numeric
	ceiling_float64.Name = "ceiling"
	ceiling_numeric.Name = "ceiling"
	framework.RegisterFunction(ceiling_float64)
	framework.RegisterFunction(ceiling_numeric)
}

// ceil_float64 represents the PostgreSQL function of the same name, taking the same parameters.
var ceil_float64 = framework.Function1{
	Name:       "ceil",
	Return:     pgtypes.Float64,
	Parameters: []pgtypes.DoltgresType{pgtypes.Float64},
	Callable: func(ctx *sql.Context, val1 any) (any, error) {
		return math.Ceil(val1.(float64)), nil
	},
	Strict: true,
}

// ceil_numeric represents the PostgreSQL function of the same name, taking the same parameters.
var ceil_numeric = framework.Function1{
	Name:       "ceil",
	Return:     pgtypes.Numeric,
	Parameters: []pgtypes.DoltgresType{pgtypes.Numeric},
	Callable: func(ctx *sql.Context, val1 any) (any, error) {
		return val1.(decimal.Decimal).Ceil(), nil
	},
	Strict: true,
}
