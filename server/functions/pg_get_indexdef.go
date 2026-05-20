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
	"strings"

	"github.com/cockroachdb/errors"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
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
	Parameters:         [1]*pgtypes.DoltgresType{pgtypes.Oid},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		oidVal := val.(id.Id)
		result := ""
		err := RunCallback(ctx, oidVal, Callbacks{
			Index: func(ctx *sql.Context, schema ItemSchema, table ItemTable, index ItemIndex) (cont bool, err error) {
				result = buildIndexDef(index.Item, schema.Item.SchemaName())
				return false, nil
			},
		})
		if err != nil {
			return "", err
		}
		return result, nil
	},
}

// buildIndexDef generates a CREATE INDEX DDL statement for the given index.
func buildIndexDef(index sql.Index, schemaName string) string {
	name := index.ID()
	using := strings.ToLower(index.IndexType())
	unique := ""
	if index.IsUnique() {
		unique = " UNIQUE"
	}

	cols := make([]string, len(index.Expressions()))
	for i, expr := range index.Expressions() {
		split := strings.Split(expr, ".")
		if len(split) > 1 {
			cols[i] = split[1]
		} else {
			cols[i] = expr
		}
	}
	colsStr := strings.Join(cols, ", ")

	def := fmt.Sprintf("CREATE%s INDEX %s ON %s.%s USING %s (%s)", unique, name, schemaName, index.Table(), using, colsStr)
	if pi, ok := index.(sql.PartialIndex); ok && pi.Predicate() != "" {
		def += " WHERE (" + pi.Predicate() + ")"
	}
	return def
}

// pg_get_indexdef_oid_integer_bool represents the PostgreSQL system catalog information function.
var pg_get_indexdef_oid_integer_bool = framework.Function3{
	Name:               "pg_get_indexdef",
	Return:             pgtypes.Text,
	Parameters:         [3]*pgtypes.DoltgresType{pgtypes.Oid, pgtypes.Int32, pgtypes.Bool},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		oidVal := val1.(id.Id)
		colNo := val2.(int32)
		pretty := val3.(bool)
		if pretty {
			return "", errors.Errorf("pretty printing is not yet supported")
		}
		err := RunCallback(ctx, oidVal, Callbacks{
			Index: func(ctx *sql.Context, schema ItemSchema, table ItemTable, index ItemIndex) (cont bool, err error) {
				exprs := index.Item.Expressions()
				if int(colNo) >= len(exprs) {
					return false, errors.Errorf("column not found")
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
