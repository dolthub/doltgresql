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
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initCurrentDatabase registers the functions to the catalog.
func initCurrentDatabase() {
	framework.RegisterFunction(current_database)
}

// current_database represents the PostgreSQL system information function of the same name, taking no parameters.
var current_database = framework.Function0{
	Name:               "current_database",
	Return:             pgtypes.Name,
	Parameters:         []pgtypes.DoltgresType{},
	IsNonDeterministic: true,
	Callable: func(ctx *sql.Context) (any, error) {
		if ctx.GetCurrentDatabase() == "" {
			return nil, nil
		}
		return ctx.GetCurrentDatabase(), nil
	},
	Strict: true,
}
