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
	"fmt"

	"github.com/cockroachdb/errors"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	pgnodes "github.com/dolthub/doltgresql/server/node"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// nodeDropOperator handles *tree.DropOperator nodes.
func nodeDropOperator(ctx *Context, node *tree.DropOperator) (vitess.Statement, error) {
	if node == nil {
		return nil, nil
	}
	if node.DropBehavior == tree.DropCascade {
		return nil, fmt.Errorf("DROP OPERATOR with CASCADE is not supported yet")
	}
	ops := make([]*pgnodes.OperatorToDrop, len(node.Operators))
	for i, op := range node.Operators {
		if op.Right == nil {
			return nil, errors.New("postfix operators are not supported")
		}
		var err error
		var leftType *pgtypes.DoltgresType
		if op.Left != nil {
			_, leftType, err = nodeResolvableTypeReference(ctx, op.Left, false)
			if err != nil {
				return nil, err
			}
		}
		_, rightType, err := nodeResolvableTypeReference(ctx, op.Right, false)
		if err != nil {
			return nil, err
		}
		ops[i] = &pgnodes.OperatorToDrop{
			Symbol: tree.OperatorSymbol(op.Op),
			Left:   leftType,
			Right:  rightType,
		}
	}
	return vitess.InjectedStatement{
		Statement: pgnodes.NewDropOperator(node.IfExists, ops),
		Children:  nil,
	}, nil
}
