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

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initAsin registers the functions to the catalog.
func initCurrentSetting() {
	framework.RegisterFunction(current_setting)
}

// asin_float64 represents the PostgreSQL function of the same name, taking the same parameters.
var current_setting = framework.Function1{
	Name:       "current_setting",
	Return:     pgtypes.Text, // TODO: it would be nice to support non-text values as well, but this is all postgres supports
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val1 any) (any, error) {
		s := val1.(string)
		_, variable, err := ctx.GetUserVariable(ctx, s)
		if err != nil {
			return nil, err
		}

		if variable != nil {
			return variable, nil
		}

		variable, err = ctx.GetSessionVariable(ctx, s)
		if err != nil {
			return nil, fmt.Errorf("unrecognized configuration parameter %s", s)
		}

		if variable != nil {
			return variable, nil
		}

		return nil, fmt.Errorf("unrecognized configuration parameter %s", s)
	},
}
