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
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/server/types/oid"
	"github.com/dolthub/go-mysql-server/sql"
)

// initPgTableIsVisible registers the functions to the catalog.
func initPgTableIsVisible() {
	framework.RegisterFunction(pg_table_is_visible)
}

// pg_table_is_visible represents the PostgreSQL system schema visibility inquiry function.
var pg_table_is_visible = framework.Function1{
	Name:               "pg_table_is_visible",
	Return:             pgtypes.Bool,
	Parameters:         [1]pgtypes.DoltgresType{pgtypes.Oid},
	IsNonDeterministic: true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		oidVal := val.(uint32)
		paths, err := resolve.SearchPath(ctx)
		lookUpPaths := make(map[string]bool)
		for _, path := range paths {
			lookUpPaths[path] = true
		}

		var isVisible bool
		err = oid.RunCallback(ctx, oidVal, oid.Callbacks{
			Table: func(ctx *sql.Context, sch oid.ItemSchema, table oid.ItemTable) (cont bool, err error) {
				_, isVisible = lookUpPaths[sch.Item.SchemaName()]
				return false, nil
			},
			View: func(ctx *sql.Context, sch oid.ItemSchema, view oid.ItemView) (cont bool, err error) {
				_, isVisible = lookUpPaths[sch.Item.SchemaName()]
				return false, nil
			},
			Index: func(ctx *sql.Context, sch oid.ItemSchema, table oid.ItemTable, index oid.ItemIndex) (cont bool, err error) {
				_, isVisible = lookUpPaths[sch.Item.SchemaName()]
				return false, nil
			},
			Sequence: func(ctx *sql.Context, sch oid.ItemSchema, sequence oid.ItemSequence) (cont bool, err error) {
				_, isVisible = lookUpPaths[sch.Item.SchemaName()]
				return false, nil
			},
			// TODO: This works for all types of relations, including views, materialized views, indexes, sequences and foreign tables.
		})
		if err != nil {
			return false, err
		}
		return isVisible, nil
	},
	Strict: true,
}
