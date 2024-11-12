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
	"github.com/lib/pq/oid"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// Array represents an ARRAY[...] expression.
type Array struct {
	children    []sql.Expression
	coercedType pgtypes.DoltgresType
}

var _ vitess.Injectable = (*Array)(nil)
var _ sql.Expression = (*Array)(nil)

// NewArray returns a new *Array.
func NewArray(coercedType sql.Type) (*Array, error) {
	var arrayCoercedType pgtypes.DoltgresType
	if dt, ok := coercedType.(pgtypes.DoltgresType); ok {
		if dt.IsArrayType() {
			arrayCoercedType = dt
		} else if !dt.IsEmptyType() {
			return nil, fmt.Errorf("cannot cast array to %s", coercedType.String())
		}
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
	resultTyp := array.coercedType.ArrayBaseType()
	values := make([]any, len(array.children))
	for i, expr := range array.children {
		val, err := expr.Eval(ctx, row)
		if err != nil {
			return nil, err
		}

		if val == nil {
			values[i] = val
			continue
		}

		doltgresType, ok := expr.Type().(pgtypes.DoltgresType)
		if !ok {
			return nil, fmt.Errorf("expected DoltgresType, but got %s", expr.Type().String())
		}

		// We always cast the element, as there may be parameter restrictions in place
		castFunc := framework.GetImplicitCast(doltgresType, resultTyp)
		if castFunc == nil {
			if doltgresType.OID == uint32(oid.T_unknown) {
				castFunc = framework.UnknownLiteralCast
			} else {
				return nil, fmt.Errorf("cannot find cast function from %s to %s", doltgresType.String(), resultTyp.String())
			}
		}

		values[i], err = castFunc(ctx, val, resultTyp)
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
	return array.coercedType
}

// WithChildren implements the sql.Expression interface.
func (array *Array) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	resultType, err := array.getTargetType(children...)
	if err != nil {
		return nil, err
	}
	return &Array{
		children:    children,
		coercedType: resultType,
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
	return array.WithChildren(newExpressions...)
}

// getTargetType returns the evaluated type for this expression.
// Returns the "anyarray" type if the type combination is invalid.
func (array *Array) getTargetType(children ...sql.Expression) (pgtypes.DoltgresType, error) {
	var childrenTypes []pgtypes.DoltgresType
	for _, child := range children {
		if child != nil {
			childType, ok := child.Type().(pgtypes.DoltgresType)
			if !ok {
				// We use "anyarray" as the indeterminate/invalid type
				return pgtypes.AnyArray, nil
			}
			childrenTypes = append(childrenTypes, childType)
		}
	}
	targetType, err := framework.FindCommonType(childrenTypes)
	if err != nil {
		return pgtypes.DoltgresType{}, fmt.Errorf("ARRAY %s", err.Error())
	}
	return targetType.ToArrayType(), nil
}
