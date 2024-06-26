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
	"github.com/dolthub/doltgresql/postgres/parser/sessiondata"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
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
	Parameters:         []pgtypes.DoltgresType{pgtypes.Bool},
	IsNonDeterministic: true,
	Callable: func(ctx *sql.Context, val any) (any, error) {
		schemas := make([]string, 0)
		if val.(bool) {
			schemas = append(schemas, sessiondata.PgCatalogName)
		}
		searchPaths, err := resolve.SearchPath(ctx)
		if err != nil {
			return nil, err
		}
		for _, searchPath := range searchPaths {
			schemas = append(schemas, searchPath)
		}
		return schemas, nil
	},
	Strict: true,
}
