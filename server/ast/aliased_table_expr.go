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
	"sync/atomic"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// uniqueAliasCounter is used to create unique aliases. Aliases are required by GMS when some expressions are contained
// within a vitess.AliasedTableExpr. The Postgres AST does not have this restriction, so we must do this for
// compatibility. Callers are free to change the alias if need be, but we'll always set an alias to be safe.
var uniqueAliasCounter atomic.Uint64

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
	default:
		tableExpr, err := nodeTableExpr(expr)
		if err != nil {
			return nil, err
		}
		subquery := &vitess.Subquery{
			Select: &vitess.Select{
				From: vitess.TableExprs{tableExpr},
			},
		}
		//TODO: make sure that this actually works
		if len(node.As.Cols) > 0 {
			columns := make([]vitess.ColIdent, len(node.As.Cols))
			for i := range node.As.Cols {
				columns[i] = vitess.NewColIdent(string(node.As.Cols[i]))
			}
			subquery.Columns = columns
		}
		aliasExpr = subquery
	}
	alias := string(node.As.Alias)
	if len(alias) == 0 {
		alias = generateUniqueAlias()
	}
	return &vitess.AliasedTableExpr{
		Expr:    aliasExpr,
		As:      vitess.NewTableIdent(alias),
		AsOf:    nil,
		Lateral: node.Lateral,
	}, nil
}

// generateUniqueAlias generates a unique alias. This is thread-safe.
func generateUniqueAlias() string {
	return fmt.Sprintf("doltgres!|alias|%d", uniqueAliasCounter.Add(1))
}
