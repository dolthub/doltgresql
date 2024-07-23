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

package expression

import (
	"fmt"

	"github.com/dolthub/go-mysql-server/sql"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// AnyExpr represents the ANY expression.
type AnyExpr struct {
	left  sql.Expression
	right sql.Expression
}

// NewAnyExpr creates a new AnyExpr expression.
func NewAnyExpr() *AnyExpr {
	return &AnyExpr{left: nil, right: nil}
}

// Children implements the Expression interface.
func (a *AnyExpr) Children() []sql.Expression {
	return []sql.Expression{a.left, a.right}
}

// Resolved implements the Expression interface.
func (a *AnyExpr) Resolved() bool {
	return a.left.Resolved() && a.right.Resolved()
}

// IsNullable implements the Expression interface.
func (a *AnyExpr) IsNullable() bool {
	return a.left.IsNullable() || a.right.IsNullable()
}

// Type implements the Expression interface.
func (a *AnyExpr) Type() sql.Type {
	return pgtypes.Bool
}

// Eval implements the Expression interface.
func (a *AnyExpr) Eval(ctx *sql.Context, row sql.Row) (interface{}, error) {
	leftVal, err := a.left.Eval(ctx, row)
	if err != nil {
		return nil, err
	}

	rightVal, err := a.right.Eval(ctx, row)
	if err != nil {
		return nil, err
	}

	switch v := rightVal.(type) {
	case []interface{}:
		// If the arrayVal is an array of interface{}
		for _, val := range v {
			if leftVal == val {
				return true, nil
			}
		}
	default:
		if v == leftVal {
			return true, nil
		}
	}

	return false, nil
}

// WithChildren implements the Expression interface.
func (a *AnyExpr) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 2 {
		return nil, sql.ErrInvalidChildrenNumber.New(a, len(children), 2)
	}
	return &AnyExpr{left: children[0], right: children[1]}, nil
}

// WithResolvedChildren implements the Expression interface.
func (a *AnyExpr) WithResolvedChildren(children []any) (any, error) {
	if len(children) != 2 {
		return nil, fmt.Errorf("invalid vitess child count, expected `2` but got `%d`", len(children))
	}
	left, ok := children[0].(sql.Expression)
	if !ok {
		return nil, fmt.Errorf("expected vitess child to be an expression but has type `%T`", children[0])
	}
	right, ok := children[1].(sql.Expression)
	if !ok {
		return nil, fmt.Errorf("expected vitess child to be an expression but has type `%T`", children[1])
	}
	return a.WithChildren(left, right)
}

// String implements the fmt.Stringer interface.
func (a *AnyExpr) String() string {
	return fmt.Sprintf("%s ANY (%s)", a.left, a.right)
}

// DebugString implements the Expression interface.
func (a *AnyExpr) DebugString() string {
	return fmt.Sprintf("%s ANY (%s)", sql.DebugString(a.left), sql.DebugString(a.right))
}
