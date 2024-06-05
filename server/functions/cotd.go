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

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initCotd registers the functions to the catalog.
func initCotd() {
	framework.RegisterFunction(cotd_float64)
}

// cot_float64 represents the PostgreSQL function of the same name, taking the same parameters.
var cotd_float64 = framework.Function1{
	Name:       "cotd",
	Return:     pgtypes.Float64,
	Parameters: []pgtypes.DoltgresType{pgtypes.Float64},
	Callable: func(ctx *sql.Context, val1Interface any) (any, error) {
		if val1Interface == nil {
			return nil, nil
		}
		val1 := toRadians(val1Interface.(float64))
		if val1 == 0 {
			return math.Inf(1), nil
		}
		return math.Cos(val1) / math.Sin(val1), nil
	},
}
