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

	"github.com/dolthub/doltgresql/postgres/parser/sessiondata"
	"github.com/dolthub/doltgresql/server/functions/framework"
	"github.com/dolthub/doltgresql/server/settings"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initCurrentSchemas registers the functions to the catalog.
func initCurrentSchemas() {
	framework.RegisterFunction(current_schemas)
}

// current_schemas represents the PostgreSQL system information function taking a bool parameter.
var current_schemas = framework.Function1{
	Name:               "current_schemas",
	Return:             pgtypes.NameArray,
	Parameters:         [1]*pgtypes.DoltgresType{pgtypes.Bool},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		schemas := make([]any, 0)
		if val.(bool) {
			schemas = append(schemas, sessiondata.PgCatalogName)
		}
		searchPaths, err := settings.GetCurrentSchemas(ctx)
		if err != nil {
			return nil, err
		}
		for _, schema := range searchPaths {
			schemas = append(schemas, schema)
		}
		return schemas, nil
	},
}
