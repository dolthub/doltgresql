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

// initFactorial registers the functions to the catalog.
func initFactorial() {
	framework.RegisterFunction(factorial_int64)
}

// factorial_int64 represents the PostgreSQL function of the same name, taking the same parameters.
var factorial_int64 = framework.Function1{
	Name:       "factorial",
	Return:     pgtypes.Numeric,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int64},
	Callable: func(ctx *sql.Context, val1Interface any) (any, error) {
		if val1Interface == nil {
			return nil, nil
		}
		val1 := val1Interface.(int64)
		if val1 < 0 {
			return nil, fmt.Errorf("factorial of a negative number is undefined")
		}
		total := int64(1)
		for i := int64(2); i <= val1; i++ {
			total *= i
		}
		return decimal.NewFromInt(total), nil
	},
}
