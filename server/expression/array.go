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

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// Array represents an ARRAY[...] expression.
type Array struct {
	children    []sql.Expression
	coercedType pgtypes.DoltgresArrayType
}

var _ vitess.Injectable = (*Array)(nil)
var _ sql.Expression = (*Array)(nil)

// NewArray returns a new *Array.
func NewArray(coercedType sql.Type) (*Array, error) {
	var arrayCoercedType pgtypes.DoltgresArrayType
	if dat, ok := coercedType.(pgtypes.DoltgresArrayType); ok {
		arrayCoercedType = dat
	} else if coercedType != nil {
		return nil, fmt.Errorf("cannot cast array to %s", coercedType.String())
	}
	return &Array{
		children:    nil,
		coercedType: arrayCoercedType,
	}, nil
}

// Children implements the sql.Expression interface.
func (array *Array) Children() []sql.Expression {
	return array.children
}

// Eval implements the sql.Expression interface.
func (array *Array) Eval(ctx *sql.Context, row sql.Row) (any, error) {
	resultArrayType := array.getTargetType()
	if resultArrayType.Equals(pgtypes.AnyArray) {
		// TODO: error should look like "ARRAY types XXXX and YYYY cannot be matched", need to display conflicting types
		return nil, fmt.Errorf("ARRAY types cannot be matched")
	}
	values := make([]any, len(array.children))
	for i, expr := range array.children {
		val, err := expr.Eval(ctx, row)
		if err != nil {
			return nil, err
		}

		doltgresType, ok := expr.Type().(pgtypes.DoltgresType)
		if !ok {
			return nil, fmt.Errorf("expected DoltgresType, but got %s", expr.Type().String())
		}

		values[i], err = framework.ConvertValToCommonType(ctx, val, doltgresType, resultArrayType.BaseType())
		if err != nil {
			return nil, err
		}
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
	for _, child := range array.children {
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
	for i, child := range array.children {
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
	return array.getTargetType()
}

// WithChildren implements the sql.Expression interface.
func (array *Array) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	return &Array{
		children:    children,
		coercedType: array.coercedType,
	}, nil
}

// WithResolvedChildren implements the vitess.InjectableExpression interface.
func (array *Array) WithResolvedChildren(children []any) (any, error) {
	newExpressions := make([]sql.Expression, len(children))
	for i, resolvedChild := range children {
		resolvedExpression, ok := resolvedChild.(sql.Expression)
		if !ok {
			return nil, fmt.Errorf("expected vitess child to be an expression but has type `%T`", resolvedChild)
		}
		newExpressions[i] = resolvedExpression
	}
	return &Array{
		children:    newExpressions,
		coercedType: array.coercedType,
	}, nil
}

// getTargetType returns the evaluated type for this expression. Returns the "anyarray" type if the type combination is
// invalid.
func (array *Array) getTargetType() pgtypes.DoltgresArrayType {
	if array.coercedType != nil {
		return array.coercedType
	}
	var childrenTypes []pgtypes.DoltgresTypeBaseID
	for _, child := range array.children {
		if child != nil {
			childType, ok := child.Type().(pgtypes.DoltgresType)
			if !ok {
				// We use "anyarray" as the indeterminate/invalid type
				return pgtypes.AnyArray
			}
			childrenTypes = append(childrenTypes, childType.BaseID())
		}
	}
	targetType, ok := framework.FindCommonType(childrenTypes)
	if !ok {
		// We use "anyarray" as the indeterminate/invalid type
		return pgtypes.AnyArray
	}
	return targetType.GetRepresentativeType().ToArrayType()
}
