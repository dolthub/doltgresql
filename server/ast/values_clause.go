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
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// nodeValuesClause handles tree.ValuesClause nodes.
func nodeValuesClause(ctx *Context, node *tree.ValuesClause) (*vitess.Select, error) {
	if node == nil {
		return nil, nil
	}
	valTuples := make([]vitess.ValTuple, len(node.Rows))
	for i := range node.Rows {
		exprs, err := nodeExprs(ctx, node.Rows[i])
		if err != nil {
			return nil, err
		}
		valTuples[i] = vitess.ValTuple(exprs)
	}
	//TODO: ValuesStatement might need to be aliased
	//TODO: is the SelectExprs necessary?
	return &vitess.Select{
		SelectExprs: vitess.SelectExprs{
			&vitess.StarExpr{
				TableName: vitess.TableName{
					Name: vitess.NewTableIdent("*"),
				},
			},
		},
		From: vitess.TableExprs{
			&vitess.ValuesStatement{
				Rows: valTuples,
			},
		},
	}, nil
}
