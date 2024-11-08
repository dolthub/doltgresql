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

	pgexprs "github.com/dolthub/doltgresql/server/expression"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// nodeLimit handles *tree.Limit nodes.
func nodeLimit(ctx *Context, node *tree.Limit) (*vitess.Limit, error) {
	if node == nil || (node.Count == nil && node.Offset == nil) {
		return nil, nil
	}
	var count vitess.Expr
	if !node.LimitAll {
		var err error
		count, err = nodeExpr(ctx, node.Count)
		if err != nil {
			return nil, err
		}
	}
	offset, err := nodeExpr(ctx, node.Offset)
	if err != nil {
		return nil, err
	}
	// GMS is hardcoded to expect vitess.SQLVal for expressions such as `LIMIT 1 OFFSET 1`.
	// We need to remove the hard dependency, but for now, we'll just convert our literals to a vitess.SQLVal.
	if injectedExpr, ok := count.(vitess.InjectedExpr); ok {
		if literal, ok := injectedExpr.Expression.(*pgexprs.Literal); ok {
			count = literal.ToVitessLiteral()
		}
	}
	if injectedExpr, ok := offset.(vitess.InjectedExpr); ok {
		if literal, ok := injectedExpr.Expression.(*pgexprs.Literal); ok {
			offset = literal.ToVitessLiteral()
		}
	}
	return &vitess.Limit{
		Offset:   offset,
		Rowcount: count,
	}, nil
}
