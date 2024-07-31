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
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/go-mysql-server/sql"
	"strings"
)

// initStringAgg registers the functions to the catalog.
func initStringAgg() {
	framework.RegisterFunction(string_agg)
}

// string_agg represents the PostgreSQL built-in aggregate function.
var string_agg = framework.Function2{
	Name:               "string_agg",
	Return:             pgtypes.Text,
	Parameters:         [2]pgtypes.DoltgresType{pgtypes.Text, pgtypes.Text},
	IsNonDeterministic: true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1, val2 any) (any, error) {
		if val1 == nil {
			return nil, nil
		}
		delimiter := ""
		if val2 != nil {
			delimiter = val2.(string)
		}
		// TODO: extract row values from val1
		expr := []string{val1.(string)}
		return strings.Join(expr, delimiter), nil
	},
	Strict: false,
}
