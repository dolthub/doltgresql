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
	"github.com/dolthub/vitess/go/vt/proto/query"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// TODO: This actually needs to defer to function calls, as addition, subtraction, etc. are just normal
//  Postgres functions accessed using an operator. Since we're not going through that just yet, this
//  will appear much more complicated as it's handling some of that logic.
//  Just to restate it once again, this functionality will not be necessary once all of the operator functions
//  are in, so this expression is temporary and only to make tests pass for now.

// Addition represents a VALUE + VALUE expression.
type Addition struct {
	left  sql.Expression
	right sql.Expression
}

var _ vitess.InjectableExpression = (*Addition)(nil)
var _ sql.Expression = (*Addition)(nil)

// NewAddition returns a new *Addition.
func NewAddition() *Addition {
	return &Addition{}
}

// Children implements the sql.Expression interface.
func (a *Addition) Children() []sql.Expression {
	return []sql.Expression{a.left, a.right}
}

// Eval implements the sql.Expression interface.
func (a *Addition) Eval(ctx *sql.Context, row sql.Row) (any, error) {
	// TODO: read the big TODO at the top of this file for more information regarding the logic here
	leftVal, err := a.left.Eval(ctx, row)
	if err != nil {
		return nil, err
	}
	rightVal, err := a.right.Eval(ctx, row)
	if err != nil {
		return nil, err
	}
	if leftVal == nil || rightVal == nil {
		return nil, nil
	}

	// Ensure that each value conforms to the expected type. This is only necessary due to the presence of GMS types
	leftType, rightType, err := a.parameterTypes()
	if err != nil {
		return nil, err
	}
	leftVal, _, err = leftType.Convert(leftVal)
	if err != nil {
		return nil, err
	}
	rightVal, _, err = rightType.Convert(rightVal)
	if err != nil {
		return nil, err
	}

	// Handle casts through the casting framework
	largestType, err := a.getLargestType()
	if err != nil {
		return nil, err
	}
	castFunc := framework.GetCast(leftType.BaseID(), largestType.BaseID())
	if castFunc == nil {
		return nil, fmt.Errorf("%T: cast from `%s` to `%s` does not exist", a, leftType.String(), largestType.String())
	}
	leftVal, err = castFunc(framework.Context{Context: ctx}, leftVal)
	if err != nil {
		return nil, err
	}
	castFunc = framework.GetCast(rightType.BaseID(), largestType.BaseID())
	if castFunc == nil {
		return nil, fmt.Errorf("%T: cast from `%s` to `%s` does not exist", a, rightType.String(), largestType.String())
	}
	rightVal, err = castFunc(framework.Context{Context: ctx}, rightVal)
	if err != nil {
		return nil, err
	}

	switch leftVal.(type) {
	case float32:
		return leftVal.(float32) + rightVal.(float32), nil
	case float64:
		return leftVal.(float64) + rightVal.(float64), nil
	case int16:
		return leftVal.(int16) + rightVal.(int16), nil
	case int32:
		return leftVal.(int32) + rightVal.(int32), nil
	case int64:
		return leftVal.(int64) + rightVal.(int64), nil
	case decimal.Decimal:
		return leftVal.(decimal.Decimal).Add(rightVal.(decimal.Decimal)), nil
	default:
		return nil, fmt.Errorf("%T: somehow encountered an unknown value type: `%T`", a, leftVal)
	}
}

// IsNullable implements the sql.Expression interface.
func (a *Addition) IsNullable() bool {
	// TODO: verify if this is actually nullable
	return false
}

// Resolved implements the sql.Expression interface.
func (a *Addition) Resolved() bool {
	if a.left == nil || !a.left.Resolved() {
		return false
	}
	if a.right == nil || !a.right.Resolved() {
		return false
	}
	return true
}

// String implements the sql.Expression interface.
func (a *Addition) String() string {
	return a.left.String() + " + " + a.right.String()
}

// Type implements the sql.Expression interface.
func (a *Addition) Type() sql.Type {
	t, err := a.getLargestType()
	if err != nil {
		// If we can't determine the type, then we'll just return numeric
		return pgtypes.Numeric
	}
	return t
}

// WithChildren implements the sql.Expression interface.
func (a *Addition) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 2 {
		return nil, sql.ErrInvalidChildrenNumber.New(a, len(children), 2)
	}
	return &Addition{
		left:  children[0],
		right: children[1],
	}, nil
}

// WithResolvedChildren implements the vitess.InjectableExpression interface.
func (a *Addition) WithResolvedChildren(children []any) (any, error) {
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
	return &Addition{
		left:  left,
		right: right,
	}, nil
}

// parameterTypes returns the Postgres types for each parameter.
func (a *Addition) parameterTypes() (left pgtypes.DoltgresType, right pgtypes.DoltgresType, err error) {
	var types [2]pgtypes.DoltgresType
	for i, parameterType := range []sql.Type{a.left.Type(), a.right.Type()} {
		var ok bool
		types[i], ok = parameterType.(pgtypes.DoltgresType)
		if !ok {
			// TODO: this expression is only temporary, so we need this block to deal with GMS types for now
			switch parameterType.Type() {
			case query.Type_INT8, query.Type_INT16, query.Type_YEAR:
				types[i] = pgtypes.Int16
			case query.Type_INT24, query.Type_INT32:
				types[i] = pgtypes.Int32
			case query.Type_INT64:
				types[i] = pgtypes.Int64
			case query.Type_UINT8, query.Type_UINT16, query.Type_UINT24, query.Type_UINT32:
				types[i] = pgtypes.Int64
			case query.Type_UINT64:
				types[i] = pgtypes.Numeric
			case query.Type_FLOAT32:
				types[i] = pgtypes.Float32
			case query.Type_FLOAT64:
				types[i] = pgtypes.Float64
			case query.Type_DECIMAL:
				types[i] = pgtypes.Numeric
			case query.Type_DATE, query.Type_DATETIME, query.Type_TIMESTAMP:
				return nil, nil, fmt.Errorf("need to add DoltgresType equivalents to DATETIME")
			case query.Type_CHAR, query.Type_VARCHAR, query.Type_TEXT:
				types[i] = pgtypes.VarCharMax
			case query.Type_ENUM:
				types[i] = pgtypes.Int16
			case query.Type_SET:
				types[i] = pgtypes.Int64
			default:
				return nil, nil, fmt.Errorf("encountered a GMS type that cannot be handled")
			}
		}
	}
	return types[0], types[1], nil
}

// getLargestType returns the largest type between the two parameters.
func (a *Addition) getLargestType() (pgtypes.DoltgresType, error) {
	leftType, rightType, err := a.parameterTypes()
	if err != nil {
		return nil, err
	}
	switch leftType.(type) {
	case pgtypes.Float32Type:
		switch rightType.(type) {
		case pgtypes.Float32Type, pgtypes.Float64Type:
			return rightType, nil
		case pgtypes.Int16Type, pgtypes.Int32Type, pgtypes.Int64Type:
			return leftType, nil
		case pgtypes.NumericType:
			return rightType, nil
		default:
			return nil, fmt.Errorf("type `%s` is invalid for addition", rightType.String())
		}
	case pgtypes.Float64Type:
		switch rightType.(type) {
		case pgtypes.Float32Type, pgtypes.Float64Type, pgtypes.Int16Type, pgtypes.Int32Type, pgtypes.Int64Type:
			return leftType, nil
		case pgtypes.NumericType:
			return rightType, nil
		default:
			return nil, fmt.Errorf("type `%s` is invalid for addition", rightType.String())
		}
	case pgtypes.Int16Type:
		switch rightType.(type) {
		case pgtypes.Float32Type, pgtypes.Float64Type, pgtypes.Int16Type, pgtypes.Int32Type, pgtypes.Int64Type, pgtypes.NumericType:
			return rightType, nil
		default:
			return nil, fmt.Errorf("type `%s` is invalid for addition", rightType.String())
		}
	case pgtypes.Int32Type:
		switch rightType.(type) {
		case pgtypes.Float32Type, pgtypes.Float64Type:
			return rightType, nil
		case pgtypes.Int16Type:
			return leftType, nil
		case pgtypes.Int32Type, pgtypes.Int64Type, pgtypes.NumericType:
			return rightType, nil
		default:
			return nil, fmt.Errorf("type `%s` is invalid for addition", rightType.String())
		}
	case pgtypes.Int64Type:
		switch rightType.(type) {
		case pgtypes.Float32Type, pgtypes.Float64Type:
			return rightType, nil
		case pgtypes.Int16Type, pgtypes.Int32Type, pgtypes.Int64Type:
			return leftType, nil
		case pgtypes.NumericType:
			return rightType, nil
		default:
			return nil, fmt.Errorf("type `%s` is invalid for addition", rightType.String())
		}
	case pgtypes.NumericType:
		switch rightType.(type) {
		case pgtypes.Float32Type, pgtypes.Float64Type, pgtypes.Int16Type, pgtypes.Int32Type, pgtypes.Int64Type, pgtypes.NumericType:
			return leftType, nil
		default:
			return nil, fmt.Errorf("type `%s` is invalid for addition", rightType.String())
		}
	default:
		return nil, fmt.Errorf("type `%s` is invalid for addition", leftType.String())
	}
}
