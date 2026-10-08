// Copyright 2026 Dolthub, Inc.
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

package ast

import (
	"maps"
	"strings"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	pgexprs "github.com/dolthub/doltgresql/server/expression"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// catalogRowsForTables preserves named catalog row types through table aliases
// and joins. Derived tables shadow outer aliases without acquiring a named type.
func catalogRowsForTables(parent map[string]*pgtypes.DoltgresType, tables tree.TableExprs) map[string]*pgtypes.DoltgresType {
	rows := maps.Clone(parent)
	if rows == nil {
		rows = make(map[string]*pgtypes.DoltgresType)
	}
	var visit func(tree.TableExpr)
	visit = func(expr tree.TableExpr) {
		switch expr := expr.(type) {
		case *tree.JoinTableExpr:
			visit(expr.Left)
			visit(expr.Right)
		case *tree.ParenTableExpr:
			visit(expr.Expr)
		case *tree.AliasedTableExpr:
			if table, ok := expr.Expr.(*tree.TableName); ok {
				alias := string(expr.As.Alias)
				if alias == "" {
					alias = table.Table()
				}
				rows[strings.ToLower(alias)] = catalogRowType(table)
			} else if expr.As.Alias != "" {
				rows[strings.ToLower(string(expr.As.Alias))] = nil
			}
		case *tree.TableName:
			rows[strings.ToLower(expr.Table())] = catalogRowType(expr)
		case *tree.UnresolvedObjectName:
			table := expr.ToTableName()
			rows[strings.ToLower(table.Table())] = catalogRowType(&table)
		}
	}
	for _, table := range tables {
		visit(table)
	}
	return rows
}

func catalogRowType(table *tree.TableName) *pgtypes.DoltgresType {
	schema := table.Schema()
	if schema == "" {
		schema = "pg_catalog"
	}
	typ := pgtypes.GetTypeByID(id.NewType(schema, table.Table()))
	if typ != nil && typ.IsCompositeType() {
		return typ
	}
	return nil
}

// catalogRowArgument expands catalog whole-row arguments using their shared
// schema, preserving the named type required by PostgreSQL function signatures.
func catalogRowArgument(ctx *Context, expr vitess.SelectExpr) vitess.SelectExpr {
	var alias string
	var bareReference vitess.Expr
	switch expr := expr.(type) {
	case *vitess.StarExpr:
		alias = expr.TableName.Name.String()
	case *vitess.AliasedExpr:
		if col, ok := expr.Expr.(*vitess.ColName); ok && col.Qualifier.IsEmpty() {
			alias = col.Name.String()
			bareReference = col
		}
	}
	typ := ctx.catalogRows[strings.ToLower(alias)]
	if typ == nil {
		return expr
	}
	children := make([]vitess.Expr, len(typ.CompositeAttrs))
	for i, attr := range typ.CompositeAttrs {
		children[i] = &vitess.ColName{
			Name:      vitess.NewColIdent(attr.Name),
			Qualifier: vitess.TableName{Name: vitess.NewTableIdent(alias)},
		}
	}
	if bareReference != nil {
		children = append([]vitess.Expr{bareReference}, children...)
	}
	return &vitess.AliasedExpr{Expr: vitess.InjectedExpr{
		Expression: pgexprs.NewCatalogRowExpr(typ, bareReference != nil),
		Children:   children,
	}}
}
