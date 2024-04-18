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
	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/server/functions/framework"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initTrimScale registers the functions to the catalog.
func initTrimScale() {
	framework.RegisterFunction(trim_scale_numeric)
}

// trim_scale_numeric represents the PostgreSQL function of the same name, taking the same parameters.
var trim_scale_numeric = framework.Function1{
	Name:       "trim_scale",
	Return:     pgtypes.Numeric,
	Parameters: []pgtypes.DoltgresType{pgtypes.Numeric},
	Callable: func(ctx framework.Context, val1 any) (any, error) {
		if val1 == nil {
			return nil, nil
		}
		// We don't store the scale in the value, so I'm not sure if this is functionally correct.
		// Seems like we'd need to modify the type of the return value (by trimming the scale), rather than the value itself.
		return val1.(decimal.Decimal), nil
	},
}
