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

// initDegrees registers the functions to the catalog.
func initDegrees() {
	framework.RegisterFunction(degrees_float64)
}

// degrees_float64 represents the PostgreSQL function of the same name, taking the same parameters.
var degrees_float64 = framework.Function1{
	Name:       "degrees",
	Return:     pgtypes.Float64,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Float64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val1 any) (any, error) {
		return toDegrees(val1.(float64)), nil
	},
}

// toDegrees converts the given radians to degrees.
func toDegrees(radians float64) float64 {
	return (radians * 180.0) / math.Pi
}
