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

	"github.com/dolthub/dolt/go/libraries/doltcore/schema"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/extensions"
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
				result = buildIndexDef(ctx, index.Item, table.Item, schema.Item.SchemaName())
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
func buildIndexDef(ctx *sql.Context, index sql.Index, table sql.Table, schemaName string) string {
	name := index.ID()
	if name == "PRIMARY" {
		// Primary key indexes are displayed with their postgres-convention name, matching pg_class
		name = fmt.Sprintf("%s_pkey", index.Table())
	}
	using := strings.ToLower(index.IndexType())
	unique := ""
	if index.IsUnique() {
		unique = " UNIQUE"
	}

	cols := indexColumnExprs(ctx, index, table)
	if len(cols) == 1 {
		col := plan.GetColumnFromIndexExpr(ctx, index.Expressions()[0], table)
		if method, opclass, ok := VectorIndexRendering(index, col); ok {
			using = method
			cols[0] += " " + opclass
		}
	}
	colsStr := strings.Join(cols, ", ")

	def := fmt.Sprintf("CREATE%s INDEX %s ON %s.%s USING %s (%s)", unique, name, schemaName, index.Table(), using, colsStr)
	if pi, ok := index.(sql.PartialIndex); ok && pi.Predicate() != "" {
		def += " WHERE (" + pi.Predicate() + ")"
	}
	return def
}

// indexColumnExprs returns the rendered text of each column of the given index, in index column
// order. Plain columns render as the bare column name, functional expressions as their original
// SQL text.
func indexColumnExprs(ctx *sql.Context, index sql.Index, table sql.Table) []string {
	cols := make([]string, len(index.Expressions()))
	for i, expr := range index.Expressions() {
		if exprText, ok := RenderHiddenIndexColumnExpr(plan.GetColumnFromIndexExpr(ctx, expr, table)); ok {
			cols[i] = exprText
			continue
		}

		split := strings.Split(expr, ".")
		if len(split) > 1 {
			cols[i] = split[1]
		} else {
			cols[i] = expr
		}
	}
	return cols
}

// VectorIndexRendering returns the access method and operator class rendered for the given vector index over the given
// column.
func VectorIndexRendering(index sql.Index, col *sql.Column) (method string, opclass string, ok bool) {
	if col == nil || !index.IsVector() {
		return "", "", false
	}
	vectorIndex, ok := index.(interface {
		VectorProperties() schema.VectorProperties
	})
	if !ok {
		return "", "", false
	}
	colType, ok := col.Type.(*pgtypes.DoltgresType)
	if !ok {
		return "", "", false
	}
	declared, ok := extensions.GetOperatorClassForIndex(colType.Name(), vectorIndex.VectorProperties().DistanceType)
	if !ok {
		return "", "", false
	}
	return "hnsw", declared.Name, true
}

// RenderHiddenIndexColumnExpr returns the original SQL text of the functional expression backing
// |col|, a hidden system column created for an indexed functional expression (e.g. `upper(name)`
// rather than that column's internal identifier, `!hidden!idx1!0!0`). ok is false if col is nil or
// isn't such a column (e.g. it's a plain, user-visible column), in which case expr is "".
func RenderHiddenIndexColumnExpr(col *sql.Column) (expr string, ok bool) {
	if col == nil || !col.HiddenSystem || col.Generated == nil {
		return "", false
	}
	if unresolved, isUnresolved := col.Generated.Expr.(*sql.UnresolvedColumnDefault); isUnresolved {
		return unresolved.String(), true
	}
	return col.Generated.String(), true
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
		// The pretty flag only affects the formatting of expressions, which we don't reproduce, so
		// we return the same text either way.
		result := ""
		err := RunCallback(ctx, oidVal, Callbacks{
			Index: func(ctx *sql.Context, schema ItemSchema, table ItemTable, index ItemIndex) (cont bool, err error) {
				if colNo == 0 {
					result = buildIndexDef(ctx, index.Item, table.Item, schema.Item.SchemaName())
					return false, nil
				}
				// A non-zero column number selects just that column's definition, or an empty
				// string if the index has no such column.
				cols := indexColumnExprs(ctx, index.Item, table.Item)
				if colNo >= 1 && int(colNo) <= len(cols) {
					result = cols[colNo-1]
				}
				return false, nil
			},
		})
		if err != nil {
			return "", err
		}
		return result, nil
	},
}
