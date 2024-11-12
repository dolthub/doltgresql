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

// TODO: move this else where
// nodeWith handles *tree.CTE nodes.
func nodeCTE(ctx *Context, node *tree.CTE) (*vitess.CommonTableExpr, error) {
	if node == nil {
		return nil, nil
	}

	alias := vitess.NewTableIdent(string(node.Name.Alias))
	cols := make([]vitess.ColIdent, len(node.Name.Cols))
	for i, col := range node.Name.Cols {
		cols[i] = vitess.NewColIdent(string(col))
	}

	subSelect, ok := node.Stmt.(*tree.Select)
	if !ok {
		return nil, fmt.Errorf("unsupported CTE statement type: %T", node.Stmt)
	}

	selectStmt, err := nodeSelect(ctx, subSelect)
	if err != nil {
		return nil, err
	}

	subQuery := &vitess.Subquery{
		Select: selectStmt,
	}

	return &vitess.CommonTableExpr{
		AliasedTableExpr: &vitess.AliasedTableExpr{
			Expr: subQuery,
			As: alias,
			Auth: vitess.AuthInformation{AuthType: vitess.AuthType_IGNORE},
		},
		Columns: cols,
	}, nil
}

// nodeWith handles *tree.With nodes.
func nodeWith(ctx *Context, node *tree.With) (*vitess.With, error) {
	if node == nil {
		return nil, nil
	}

	ctes := make([]vitess.TableExpr, len(node.CTEList))
	for i, cte := range node.CTEList {
		var err error
		ctes[i], err = nodeCTE(ctx, cte)
		if err != nil {
			return nil, err
		}
	}

	return &vitess.With{
		Recursive: node.Recursive,
		Ctes:      ctes,
	}, nil
}
