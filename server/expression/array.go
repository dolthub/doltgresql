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
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// Array represents an ARRAY[...] expression.
type Array struct {
	sqlChildren    []sql.Expression
	vitessChildren vitess.Exprs
	vitessIndexes  []int
	coercedType    sql.Type
}

var _ vitess.InjectableExpression = (*Array)(nil)
var _ sql.Expression = (*Array)(nil)

// NewArray returns a new *Array.
func NewArray(expressions []sql.Expression, unresolvedChildren vitess.Exprs, unresolvedIndexes []int, coercedType sql.Type) (*Array, error) {
	if len(unresolvedChildren) != len(unresolvedIndexes) {
		return nil, fmt.Errorf("ARRAY has an invalid number of unresolved children (%d) and indexes (%d)",
			len(unresolvedChildren), len(unresolvedIndexes))
	}
	for _, index := range unresolvedIndexes {
		if index >= len(expressions) || index < 0 {
			return nil, fmt.Errorf("ARRAY unresolved index (%d) is out of bounds (expression count: %d)",
				index, len(expressions))
		}
		if expressions[index] != nil {
			return nil, fmt.Errorf("ARRAY unresolved index (%d) points to a resolved expression", index)
		}
	}
	return &Array{
		sqlChildren:    expressions,
		vitessChildren: unresolvedChildren,
		vitessIndexes:  unresolvedIndexes,
		coercedType:    coercedType,
	}, nil
}

// Children implements the sql.Expression interface.
func (array *Array) Children() []sql.Expression {
	return array.sqlChildren
}

// Eval implements the sql.Expression interface.
func (array *Array) Eval(ctx *sql.Context, row sql.Row) (interface{}, error) {
	// TODO: make this an actual implementation instead of the mock boolean implementation
	values := make([]bool, len(array.sqlChildren))
	for i, expr := range array.sqlChildren {
		val, err := expr.Eval(ctx, row)
		if err != nil {
			return nil, err
		}
		boolVal, ok := val.(bool)
		if !ok {
			return nil, fmt.Errorf("only boolean values are supported within an ARRAY for now")
		}
		values[i] = boolVal
	}
	return values, nil
}

// IsNullable implements the sql.Expression interface.
func (array *Array) IsNullable() bool {
	// TODO: verify if this is actually nullable
	return false
}

// Resolved implements the sql.Expression interface.
func (array *Array) Resolved() bool {
	for _, child := range array.sqlChildren {
		if child == nil || !child.Resolved() {
			return false
		}
	}
	return true
}

// String implements the sql.Expression interface.
func (array *Array) String() string {
	sb := strings.Builder{}
	sb.WriteString("ARRAY[")
	for i, child := range array.sqlChildren {
		if i > 0 {
			sb.WriteString(", ")
		}
		if child == nil {
			sb.WriteString("...")
		} else {
			sb.WriteString(child.String())
		}
	}
	sb.WriteRune(']')
	return sb.String()
}

// Type implements the sql.Expression interface.
func (array *Array) Type() sql.Type {
	if array.coercedType != nil {
		return array.coercedType
	}
	// TODO: how do we handle multiple children with different types?
	for _, child := range array.sqlChildren {
		if child != nil {
			childType := child.Type()
			if _, ok := childType.(pgtypes.BoolType); ok {
				return pgtypes.BoolArray
			}
		}
	}
	// TODO: remove mock boolean array demonstration
	return pgtypes.BoolArray
}

// WithChildren implements the sql.Expression interface.
func (array *Array) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	return &Array{
		sqlChildren: children,
		coercedType: array.coercedType,
	}, nil
}

// WithResolvedChildren implements the vitess.InjectableExpression interface.
func (array *Array) WithResolvedChildren(children []any) (any, error) {
	if len(children) != len(array.vitessIndexes) {
		return nil, fmt.Errorf("invalid vitess child count, expected `%d` but got `%d`",
			len(array.vitessIndexes), len(children))
	}
	newExpressions := make([]sql.Expression, len(array.sqlChildren))
	copy(newExpressions, array.sqlChildren)
	for i, resolvedChild := range children {
		resolvedExpression, ok := resolvedChild.(sql.Expression)
		if !ok {
			return nil, fmt.Errorf("expected vitess child to be an expression but has type `%T`", resolvedChild)
		}
		newExpressions[array.vitessIndexes[i]] = resolvedExpression
	}
	return &Array{
		sqlChildren: newExpressions,
		coercedType: array.coercedType,
	}, nil
}
