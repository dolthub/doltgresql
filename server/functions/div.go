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

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initDiv registers the functions to the catalog.
func initDiv() {
	framework.RegisterFunction(div_numeric)
}

// div_numeric represents the PostgreSQL function of the same name, taking the same parameters.
var div_numeric = framework.Function2{
	Name:       "div",
	Return:     pgtypes.Numeric,
	Parameters: []pgtypes.DoltgresType{pgtypes.Numeric, pgtypes.Numeric},
	Callable: func(ctx *sql.Context, val1Interface any, val2Interface any) (any, error) {
		val1 := val1Interface.(decimal.Decimal)
		val2 := val2Interface.(decimal.Decimal)
		if val2.Cmp(decimal.Zero) == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		val := val1.Div(val2)
		return val.Truncate(0), nil
	},
	Strict: true,
}
