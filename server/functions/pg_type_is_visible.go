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

	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/resolve"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initPgTableSize registers the functions to the catalog.
func initPgTypeIsVisible() {
	framework.RegisterFunction(pg_type_is_visible)
}

// pg_type_is_visible represents the PostgreSQL function of the same name, taking the same parameters.
var pg_type_is_visible = framework.Function1{
	Name:               "pg_type_is_visible",
	Return:             pgtypes.Bool,
	Parameters:         [1]*pgtypes.DoltgresType{pgtypes.Oid},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		typeId := id.Cache().ToInternal(val.(uint32))
		if !typeId.IsValid() {
			return false, nil
		}
		
		searchPath, err := resolve.SearchPath(ctx)
		if err != nil {
			return nil, err
		}

		typ := id.Type(typeId)
		for _, schema := range searchPath {
			if strings.EqualFold(schema, typ.SchemaName()) {
				return true, nil
			}
		}
		
		return false, nil
	},
}
