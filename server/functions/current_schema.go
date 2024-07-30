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
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/resolve"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initCurrentSchema registers the functions to the catalog.
func initCurrentSchema() {
	framework.RegisterFunction(current_schema)
}

// current_schema represents the PostgreSQL system information function of the same name, taking no parameters.
var current_schema = framework.Function0{
	Name:               "current_schema",
	Return:             pgtypes.Name,
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, varargs ...any) (any, error) {
		schemas, err := resolve.SearchPath(ctx)
		if err != nil {
			return nil, err
		}
		if len(schemas) == 0 {
			return nil, nil
		}
		return schemas[0], nil
	},
}
