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
	"github.com/dolthub/doltgresql/server/types/oid"
)

// initPgGetIndexDef registers the functions to the catalog.
func initPgGetIndexDef() {
	framework.RegisterFunction(pg_get_indexdef_oid)
	framework.RegisterFunction(pg_get_indexdef_oid_integer_bool)
}

// pg_get_indexdef_oid represents the PostgreSQL system catalog information function.
var pg_get_indexdef_oid = framework.Function1{
	Name:               "pg_get_indexdef",
	Return:             pgtypes.Text,
	Parameters:         [1]pgtypes.DoltgresType{pgtypes.Oid},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		oidVal := val.(uint32)
		err := oid.RunCallback(ctx, oidVal, oid.Callbacks{
			Index: func(ctx *sql.Context, schema oid.ItemSchema, table oid.ItemTable, index oid.ItemIndex) (cont bool, err error) {
				// TODO: make `create index` statement
				return false, nil
			},
		})
		if err != nil {
			return "", err
		}
		return "", nil
	},
}

// pg_get_indexdef_oid_integer_bool represents the PostgreSQL system catalog information function.
var pg_get_indexdef_oid_integer_bool = framework.Function3{
	Name:               "pg_get_indexdef",
	Return:             pgtypes.Text,
	Parameters:         [3]pgtypes.DoltgresType{pgtypes.Oid, pgtypes.Int32, pgtypes.Bool},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [4]pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		oidVal := val1.(uint32)
		colNo := val2.(int32)
		pretty := val3.(bool)
		if pretty {
			return "", fmt.Errorf("pretty printing is not yet supported")
		}
		err := oid.RunCallback(ctx, oidVal, oid.Callbacks{
			Index: func(ctx *sql.Context, schema oid.ItemSchema, table oid.ItemTable, index oid.ItemIndex) (cont bool, err error) {
				exprs := index.Item.Expressions()
				if int(colNo) >= len(exprs) {
					return false, fmt.Errorf("column not found")
				}
				// TODO: make `create index` statement
				return false, nil
			},
		})
		if err != nil {
			return "", err
		}
		return "", nil
	},
}
