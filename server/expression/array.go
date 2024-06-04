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
	"github.com/dolthub/vitess/go/vt/proto/query"
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
		values[i] = val
	}
	// We always cast the elements, as there may be parameter restrictions in place
	baseResultType := resultArrayType.BaseType()
	var err error
	for i := range values {
		if values[i] == nil {
			continue
		}
		values[i], err = array.handleEvaluationCast(ctx, baseResultType, array.children[i].Type(), &values[i])
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

// castable returns the largest type if both types are automatically castable to each other. Returns false if neither
// type allows for automatic casting.
func (array *Array) castable(t1, t2 pgtypes.DoltgresType) (pgtypes.DoltgresType, bool) {
	t1BaseID := t1.BaseID()
	t2BaseID := t2.BaseID()
	// We set these to different negative numbers so that the default matching behavior will fail, since it checks for
	// the equivalence of these two numbers. This is in an array so that we can simply loop over the logic.
	generalTyping := [2]int16{-1, -2}
	// Check for castable groups specific to the ARRAY expression. We'll assign integers to broadly represent each group.
	for i, baseID := range []pgtypes.DoltgresTypeBaseID{t1BaseID, t2BaseID} {
		switch baseID {
		// TODO: fill out the remaining convertable groups
		case pgtypes.DoltgresTypeBaseID_Float32, pgtypes.DoltgresTypeBaseID_Float64, pgtypes.DoltgresTypeBaseID_Int16,
			pgtypes.DoltgresTypeBaseID_Int32, pgtypes.DoltgresTypeBaseID_Int64, pgtypes.DoltgresTypeBaseID_Numeric:
			generalTyping[i] = 1
		}
	}
	// If the types are not in the same group, then we'll return false
	if generalTyping[0] != generalTyping[1] {
		return nil, false
	}
	// Check for each cast group
	if generalTyping[0] == 1 {
		if array.numberCastGroupPriority(t1BaseID) < array.numberCastGroupPriority(t2BaseID) {
			return t1, true
		} else {
			return t2, true
		}
	}
	return nil, false
}

// handleEvaluationCast handles the casts performed during evaluation. This is only called if casting is required.
func (array *Array) handleEvaluationCast(ctx *sql.Context, baseResultType pgtypes.DoltgresType, paramSqlType sql.Type, val *any) (any, error) {
	var paramType pgtypes.DoltgresType
	if doltgresType, ok := paramSqlType.(pgtypes.DoltgresType); ok {
		paramType = doltgresType
	} else {
		// TODO: we need to remove GMS types from all of our expressions so that we can remove this. For now, we have to
		// handle all possible GMS types and make any conversions for types that are not supported by Postgres
		switch paramType.Type() {
		case query.Type_INT8:
			*val = int16((*val).(int8))
			paramType = pgtypes.Int16
		case query.Type_INT16:
			paramType = pgtypes.Int16
		case query.Type_INT24, query.Type_INT32:
			paramType = pgtypes.Int32
		case query.Type_INT64:
			paramType = pgtypes.Int64
		case query.Type_UINT8:
			*val = int64((*val).(uint8))
			paramType = pgtypes.Int64
		case query.Type_UINT16:
			*val = int64((*val).(uint16))
			paramType = pgtypes.Int64
		case query.Type_UINT24, query.Type_UINT32:
			*val = int64((*val).(uint32))
			paramType = pgtypes.Int64
		case query.Type_UINT64:
			*val = int64((*val).(uint64))
			paramType = pgtypes.Int64
		case query.Type_YEAR:
			paramType = pgtypes.Int16
		case query.Type_FLOAT32:
			paramType = pgtypes.Float32
		case query.Type_FLOAT64:
			paramType = pgtypes.Float64
		case query.Type_DECIMAL:
			paramType = pgtypes.Numeric
		case query.Type_DATE, query.Type_DATETIME, query.Type_TIMESTAMP:
			return nil, fmt.Errorf("need to add DoltgresType equivalents to DATETIME")
		case query.Type_CHAR, query.Type_VARCHAR, query.Type_TEXT:
			paramType = pgtypes.Text
		case query.Type_ENUM:
			paramType = pgtypes.Int16
		case query.Type_SET:
			paramType = pgtypes.Int64
		case query.Type_NULL_TYPE:
			paramType = pgtypes.Null
		default:
			return nil, fmt.Errorf("encountered an unknown GMS type")
		}
	}
	castFunc := framework.GetImplicitCast(paramType.BaseID(), baseResultType.BaseID())
	if castFunc == nil {
		// This should never happen, but we'll check here just to be safe
		resultType := array.getTargetType()
		return nil, fmt.Errorf("cannot cast type %s to %s", resultType.BaseType().String(), paramType.String())
	}
	return castFunc(ctx, *val, baseResultType)
}

// isNullType returns whether the given type is a NULL type.
func (array *Array) isNullType(t sql.Type) bool {
	return t.Type() == query.Type_NULL_TYPE
}

// numberCastGroupPriority returns the priority for the given type belonging to the number group. The lower the value,
// the higher the priority.
func (array *Array) numberCastGroupPriority(t pgtypes.DoltgresTypeBaseID) int {
	switch t {
	case pgtypes.DoltgresTypeBaseID_Float64:
		return 1
	case pgtypes.DoltgresTypeBaseID_Float32:
		return 2
	case pgtypes.DoltgresTypeBaseID_Numeric:
		return 3
	case pgtypes.DoltgresTypeBaseID_Int64:
		return 4
	case pgtypes.DoltgresTypeBaseID_Int32:
		return 5
	case pgtypes.DoltgresTypeBaseID_Int16:
		return 6
	default:
		return 7
	}
}

// getTargetType returns the evaluated type for this expression. Returns the "anyarray" type if the type combination is
// invalid.
func (array *Array) getTargetType() pgtypes.DoltgresArrayType {
	// TODO: figure out the conditions that result in this being set
	if array.coercedType != nil {
		return array.coercedType
	}
	var lastChildType pgtypes.DoltgresType
	for _, child := range array.children {
		if child != nil {
			gmsChildType := child.Type()
			// We ignore NULL values here since they do not affect the array's type
			if array.isNullType(gmsChildType) {
				continue
			}
			// Ensure that the type is a DoltgresType
			childType, ok := gmsChildType.(pgtypes.DoltgresType)
			if !ok {
				// We use "anyarray" as the indeterminate/invalid type
				return pgtypes.AnyArray
			}
			// Ensure that all of the types align to a common type
			if lastChildType == nil {
				lastChildType = childType
			} else if !lastChildType.Equals(childType) {
				if castableType, ok := array.castable(lastChildType, childType); ok {
					lastChildType = castableType
				} else {
					lastChildType = nil
					break
				}
			}
		}
	}
	// If this is not nil, then all types either match this type, or are automatically castable to this type
	if lastChildType != nil {
		return lastChildType.ToArrayType()
	}
	// We use "anyarray" as the indeterminate/invalid type
	return pgtypes.AnyArray
}
