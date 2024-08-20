// Copyright 2023 Dolthub, Inc.
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
	"fmt"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// nodeAliasedTableExpr handles *tree.AliasedTableExpr nodes.
func nodeAliasedTableExpr(node *tree.AliasedTableExpr) (*vitess.AliasedTableExpr, error) {
	if node.Ordinality {
		return nil, fmt.Errorf("ordinality is not yet supported")
	}
	if node.IndexFlags != nil {
		return nil, fmt.Errorf("index flags are not yet supported")
	}
	var aliasExpr vitess.SimpleTableExpr
	switch expr := node.Expr.(type) {
	case *tree.TableName:
		var err error
		aliasExpr, err = nodeTableName(expr)
		if err != nil {
			return nil, err
		}
	case *tree.Subquery:
		tableExpr, err := nodeTableExpr(expr)
		if err != nil {
			return nil, err
		}

		ate, ok := tableExpr.(*vitess.AliasedTableExpr)
		if !ok {
			return nil, fmt.Errorf("expected *vitess.AliasedTableExpr, found %T", tableExpr)
		}

		var selectStmt vitess.SelectStatement
		switch ate.Expr.(type) {
		case *vitess.Subquery:
			selectStmt = ate.Expr.(*vitess.Subquery).Select
		default:
			return nil, fmt.Errorf("unhandled subquery table expression: `%T`", tableExpr)
		}

		subquery := &vitess.Subquery{
			Select: selectStmt,
		}

		// TODO: make sure that this actually works
		if len(node.As.Cols) > 0 {
			columns := make([]vitess.ColIdent, len(node.As.Cols))
			for i := range node.As.Cols {
				columns[i] = vitess.NewColIdent(string(node.As.Cols[i]))
			}
			subquery.Columns = columns
		}
		aliasExpr = subquery
	default:
		return nil, fmt.Errorf("unhandled table expression: `%T`", expr)
	}
	alias := string(node.As.Alias)
	return &vitess.AliasedTableExpr{
		Expr:    aliasExpr,
		As:      vitess.NewTableIdent(alias),
		AsOf:    nil,
		Lateral: node.Lateral,
	}, nil
}
