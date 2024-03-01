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

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// Cast represents a VALUE::TYPE expression.
type Cast struct {
	sqlChild   sql.Expression
	castToType pgtypes.DoltgresType
}

var _ vitess.InjectableExpression = (*Cast)(nil)
var _ sql.Expression = (*Cast)(nil)

// NewCast returns a new *Cast.
func NewCast(castToType sql.Type) (*Cast, error) {
	pgtype, ok := castToType.(pgtypes.DoltgresType)
	if !ok {
		return nil, fmt.Errorf("cast expects a Doltgres type as the target type")
	}
	return &Cast{
		sqlChild:   nil,
		castToType: pgtype,
	}, nil
}

// Children implements the sql.Expression interface.
func (c *Cast) Children() []sql.Expression {
	return []sql.Expression{c.sqlChild}
}

// Child returns the child that is being cast.
func (c *Cast) Child() sql.Expression {
	return c.sqlChild
}

// Eval implements the sql.Expression interface.
func (c *Cast) Eval(ctx *sql.Context, row sql.Row) (any, error) {
	val, err := c.sqlChild.Eval(ctx, row)
	if err != nil {
		return nil, err
	}
	fromType, ok := c.sqlChild.Type().(pgtypes.DoltgresType)
	if !ok {
		// TODO: we need to remove GMS types from all expressions, then we can remove this block
		switch c.sqlChild.Type().Type() {
		case query.Type_INT8, query.Type_INT16:
			fromType = pgtypes.Int16
			val, _, err = pgtypes.Int16.Convert(val)
			if err != nil {
				return nil, err
			}
		case query.Type_INT24, query.Type_INT32:
			fromType = pgtypes.Int32
			val, _, err = pgtypes.Int32.Convert(val)
			if err != nil {
				return nil, err
			}
		case query.Type_INT64:
			fromType = pgtypes.Int64
			val, _, err = pgtypes.Int64.Convert(val)
			if err != nil {
				return nil, err
			}
		case query.Type_UINT8, query.Type_UINT16, query.Type_UINT24, query.Type_UINT32:
			fromType = pgtypes.Int64
			val, _, err = pgtypes.Int64.Convert(val)
			if err != nil {
				return nil, err
			}
		case query.Type_UINT64:
			fromType = pgtypes.Numeric
			val, _, err = pgtypes.Numeric.Convert(val)
			if err != nil {
				return nil, err
			}
		case query.Type_YEAR:
			fromType = pgtypes.Int16
			val, _, err = pgtypes.Int16.Convert(val)
			if err != nil {
				return nil, err
			}
		case query.Type_FLOAT32:
			fromType = pgtypes.Float32
			val, _, err = pgtypes.Float32.Convert(val)
			if err != nil {
				return nil, err
			}
		case query.Type_FLOAT64:
			fromType = pgtypes.Float64
			val, _, err = pgtypes.Float64.Convert(val)
			if err != nil {
				return nil, err
			}
		case query.Type_DECIMAL:
			fromType = pgtypes.Numeric
			val, _, err = pgtypes.Numeric.Convert(val)
			if err != nil {
				return nil, err
			}
		case query.Type_DATE, query.Type_DATETIME, query.Type_TIMESTAMP:
			return nil, fmt.Errorf("need to add DoltgresType equivalents to DATETIME")
		case query.Type_CHAR, query.Type_VARCHAR:
			fromType = pgtypes.VarCharMax
			val, _, err = pgtypes.VarCharMax.Convert(val)
			if err != nil {
				return nil, err
			}
		case query.Type_TEXT:
			fromType = pgtypes.VarCharMax
			val, _, err = pgtypes.VarCharMax.Convert(val)
			if err != nil {
				return nil, err
			}
		case query.Type_ENUM:
			fromType = pgtypes.Int16
			val, _, err = pgtypes.Int16.Convert(val)
			if err != nil {
				return nil, err
			}
		case query.Type_SET:
			fromType = pgtypes.Int64
			val, _, err = pgtypes.Int64.Convert(val)
			if err != nil {
				return nil, err
			}
		case query.Type_NULL_TYPE:
			fromType = pgtypes.Null
		default:
			return nil, fmt.Errorf("encountered a GMS type that cannot be handled")
		}
	}
	castFunction := framework.GetCast(fromType.BaseID(), c.castToType.BaseID())
	if castFunction == nil {
		return nil, fmt.Errorf("cast from `%s` to `%s` does not exist", fromType.String(), c.castToType.String())
	}
	return castFunction(framework.Context{Context: ctx}, val)
}

// IsNullable implements the sql.Expression interface.
func (c *Cast) IsNullable() bool {
	// TODO: verify if this is actually nullable
	return false
}

// Resolved implements the sql.Expression interface.
func (c *Cast) Resolved() bool {
	if c.sqlChild != nil && c.sqlChild.Resolved() {
		return true
	}
	return false
}

// String implements the sql.Expression interface.
func (c *Cast) String() string {
	return c.sqlChild.String() + "::" + c.castToType.String()
}

// Type implements the sql.Expression interface.
func (c *Cast) Type() sql.Type {
	return c.castToType
}

// WithChildren implements the sql.Expression interface.
func (c *Cast) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 1 {
		return nil, sql.ErrInvalidChildrenNumber.New(c, len(children), 1)
	}
	return &Cast{
		sqlChild:   children[0],
		castToType: c.castToType,
	}, nil
}

// WithResolvedChildren implements the vitess.InjectableExpression interface.
func (c *Cast) WithResolvedChildren(children []any) (any, error) {
	if len(children) != 1 {
		return nil, fmt.Errorf("invalid vitess child count, expected `1` but got `%d`", len(children))
	}
	resolvedExpression, ok := children[0].(sql.Expression)
	if !ok {
		return nil, fmt.Errorf("expected vitess child to be an expression but has type `%T`", children[0])
	}
	return &Cast{
		sqlChild:   resolvedExpression,
		castToType: c.castToType,
	}, nil
}
