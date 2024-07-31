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

	"github.com/dolthub/doltgresql/postgres/parser/parser"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/server/types/oid"
)

// initPgGetViewDef registers the functions to the catalog.
func initPgGetViewDef() {
	framework.RegisterFunction(pg_get_viewdef1)
	framework.RegisterFunction(pg_get_viewdef2bool)
	framework.RegisterFunction(pg_get_viewdef2int)
}

// pg_get_viewdef represents the PostgreSQL system catalog information function taking 1 parameter, {oid}.
var pg_get_viewdef1 = framework.Function1{
	Name:               "pg_get_viewdef",
	Return:             pgtypes.Text,
	Parameters:         [1]pgtypes.DoltgresType{pgtypes.Oid},
	IsNonDeterministic: true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		oidVal := val.(uint32)
		return getViewDef(ctx, oidVal)
	},
	Strict: true,
}

// pg_get_viewdef represents the PostgreSQL system catalog information function taking 2 parameters, {oid, bool}.
var pg_get_viewdef2bool = framework.Function2{
	Name:               "pg_get_viewdef",
	Return:             pgtypes.Text,
	Parameters:         [2]pgtypes.DoltgresType{pgtypes.Oid, pgtypes.Bool},
	IsNonDeterministic: true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1, val2 any) (any, error) {
		oidVal := val1.(uint32)
		pretty := val2.(bool)
		if pretty {
			return "", fmt.Errorf("pretty printing is not yet supported")
		}
		return getViewDef(ctx, oidVal)
	},
	Strict: true,
}

// pg_get_viewdef represents the PostgreSQL system catalog information function taking 2 parameters, {oid, int}.
var pg_get_viewdef2int = framework.Function2{
	Name:               "pg_get_viewdef",
	Return:             pgtypes.Text,
	Parameters:         [2]pgtypes.DoltgresType{pgtypes.Oid, pgtypes.Int64},
	IsNonDeterministic: true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1, val2 any) (any, error) {
		// TODO: prettyprint is implied
		return "", fmt.Errorf("pretty printing is not yet supported")
	},
	Strict: true,
}

func getViewDef(ctx *sql.Context, oidVal uint32) (string, error) {
	var result string
	err := oid.RunCallback(ctx, oidVal, oid.Callbacks{
		View: func(ctx *sql.Context, sch oid.ItemSchema, view oid.ItemView) (cont bool, err error) {
			result = view.Item.TextDefinition
			if result == "" {
				stmts, err := parser.Parse(view.Item.CreateViewStatement)
				if err != nil {
					return false, err
				}
				if len(stmts) == 0 {
					return false, fmt.Errorf("expected CREATE VIEW statement, got none")
				}
				cv, ok := stmts[0].AST.(*tree.CreateView)
				if !ok {
					return false, fmt.Errorf("expected CREATE VIEW statement, got %s", stmts[0].SQL)
				}
				result = cv.AsSource.String()
			}
			return false, nil
		},
	})
	if err != nil {
		return "", err
	}
	return result, nil
}
