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
	"strings"

	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// init registers the functions to the catalog.
func init() {
	framework.RegisterFunction(min_scale_numeric)
}

// min_scale_numeric represents the PostgreSQL function of the same name, taking the same parameters.
var min_scale_numeric = framework.Function1{
	Name:       "min_scale",
	Return:     pgtypes.Numeric,
	Parameters: []pgtypes.DoltgresType{pgtypes.Numeric},
	Callable: func(ctx framework.Context, val1 any) (any, error) {
		if val1 == nil {
			return nil, nil
		}
		str := val1.(decimal.Decimal).String()
		if idx := strings.Index(str, "."); idx != -1 {
			str = str[idx+1:]
			i := len(str) - 1
			for ; i >= 0; i-- {
				if str[i] != '0' {
					break
				}
			}
			return decimal.NewFromInt(int64(i + 1)), nil
		}
		return decimal.Zero, nil
	},
}
