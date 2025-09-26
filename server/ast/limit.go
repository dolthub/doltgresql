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
	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql/expression"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	pgexprs "github.com/dolthub/doltgresql/server/expression"
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
		if literal, ok := injectedExpr.Expression.(*expression.Literal); ok {
			l := literal.Value()
			limitValue, err := int64ValueForLimit(l)
			if err != nil {
				return nil, err
			}

			if limitValue < 0 {
				return nil, errors.Errorf("LIMIT must be greater than or equal to 0")
			}

			count = pgexprs.ToVitessLiteral(literal)
		}
	}
	if injectedExpr, ok := offset.(vitess.InjectedExpr); ok {
		if literal, ok := injectedExpr.Expression.(*expression.Literal); ok {
			o := literal.Value()
			offsetVal, err := int64ValueForLimit(o)
			if err != nil {
				return nil, err
			}

			if offsetVal < 0 {
				return nil, errors.Errorf("OFFSET must be greater than or equal to 0")
			}

			offset = pgexprs.ToVitessLiteral(literal)
		}
	}
	return &vitess.Limit{
		Offset:   offset,
		Rowcount: count,
	}, nil
}

// int64ValueForLimit converts a literal value to an int64
func int64ValueForLimit(l any) (int64, error) {
	var limitValue int64
	switch l := l.(type) {
	case int:
		limitValue = int64(l)
	case int32:
		limitValue = int64(l)
	case int64:
		limitValue = l
	case float64:
		limitValue = int64(l)
	case float32:
		limitValue = int64(l)
	default:
		return 0, errors.Errorf("limit/offset value type %T is not yet supported", l)
	}
	return limitValue, nil
}
