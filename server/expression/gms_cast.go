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
	"strconv"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/vt/proto/query"
	"github.com/shopspring/decimal"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// GMSCast handles the conversion from a GMS expression's type to its Doltgres type that is most similar.
type GMSCast struct {
	sqlChild sql.Expression
}

var _ sql.Expression = (*GMSCast)(nil)

// NewGMSCast returns a new *GMSCast.
func NewGMSCast(child sql.Expression) *GMSCast {
	return &GMSCast{
		sqlChild: child,
	}
}

// Children implements the sql.Expression interface.
func (c *GMSCast) Children() []sql.Expression {
	return []sql.Expression{c.sqlChild}
}

// Child returns the child that is being cast.
func (c *GMSCast) Child() sql.Expression {
	return c.sqlChild
}

// DoltgresType returns the DoltgresType that the cast evaluates to. This is the same value that is returned by Type().
func (c *GMSCast) DoltgresType() *pgtypes.DoltgresType {
	// GMSCast shouldn't receive a DoltgresType, but we shouldn't error if it happens
	if t, ok := c.sqlChild.Type().(*pgtypes.DoltgresType); ok {
		return t
	}

	return pgtypes.FromGmsType(c.sqlChild.Type())
}

// Eval implements the sql.Expression interface.
func (c *GMSCast) Eval(ctx *sql.Context, row sql.Row) (any, error) {
	val, err := c.sqlChild.Eval(ctx, row)
	if err != nil {
		return nil, err
	}
	if val == nil {
		return nil, nil
	}
	// GMSCast shouldn't receive a DoltgresType, but we shouldn't error if it happens
	if _, ok := c.sqlChild.Type().(*pgtypes.DoltgresType); ok {
		return val, nil
	}
	sqlTyp := c.sqlChild.Type()
	switch sqlTyp.Type() {
	// Boolean types are a special case because of how they are translated on the wire in Postgres. If we identify a
	// boolean result, we want to convert it from an int back to a boolean.
	case query.Type_INT8:
		if sqlTyp == types.Boolean {
			newVal, _, err := types.Int32.Convert(ctx, val)
			if err != nil {
				return nil, err
			}
			if _, ok := newVal.(int32); !ok {
				return nil, errors.Errorf("GMSCast expected type `int32`, got `%T`", val)
			}
			if newVal.(int32) == 0 {
				return false, nil
			} else {
				return true, nil
			}
		}
		fallthrough
	// Although Int16 would be a closer fit for some of these types, in Postgres, Int32 is generally the smallest value
	// used. To maximize overall compatibility, it's better to interpret these values as Int32 instead.
	case query.Type_INT16, query.Type_INT24, query.Type_INT32, query.Type_YEAR, query.Type_ENUM:
		newVal, _, err := types.Int32.Convert(ctx, val)
		if err != nil {
			return nil, err
		}
		if _, ok := newVal.(int32); !ok {
			return nil, errors.Errorf("GMSCast expected type `int32`, got `%T`", val)
		}
		return newVal, nil
	case query.Type_INT64, query.Type_SET, query.Type_BIT, query.Type_UINT8, query.Type_UINT16, query.Type_UINT24, query.Type_UINT32:
		newVal, _, err := types.Int64.Convert(ctx, val)
		if err != nil {
			return nil, err
		}
		if _, ok := newVal.(int64); !ok {
			return nil, errors.Errorf("GMSCast expected type `int64`, got `%T`", val)
		}
		return newVal, nil
	case query.Type_UINT64:
		if val, ok := val.(uint64); ok {
			return decimal.NewFromString(strconv.FormatUint(val, 10))
		}
		return nil, errors.Errorf("GMSCast expected type `uint64`, got `%T`", val)
	case query.Type_FLOAT32:
		if val, ok := val.(float32); ok {
			return val, nil
		}
		return nil, errors.Errorf("GMSCast expected type `float32`, got `%T`", val)
	case query.Type_FLOAT64:
		if val, ok := val.(float64); ok {
			return val, nil
		}
		return nil, errors.Errorf("GMSCast expected type `float64`, got `%T`", val)
	case query.Type_DECIMAL:
		if val, ok := val.(decimal.Decimal); ok {
			return val, nil
		}
		return nil, errors.Errorf("GMSCast expected type `Decimal`, got `%T`", val)
	case query.Type_DATE, query.Type_DATETIME, query.Type_TIMESTAMP:
		if val, ok := val.(time.Time); ok {
			return val, nil
		}
		return nil, errors.Errorf("GMSCast expected type `Time`, got `%T`", val)
	case query.Type_TIME:
		if val, ok := val.(types.Timespan); ok {
			return val.String(), nil
		}
		return nil, errors.Errorf("GMSCast expected type `Timespan`, got `%T`", val)
	case query.Type_CHAR, query.Type_VARCHAR, query.Type_TEXT, query.Type_BINARY, query.Type_VARBINARY, query.Type_BLOB:
		newVal, _, err := types.LongText.Convert(ctx, val)
		if err != nil {
			return nil, err
		}
		if _, ok := newVal.(string); !ok {
			return nil, errors.Errorf("GMSCast expected type `string`, got `%T`", val)
		}
		return newVal, nil
	case query.Type_JSON:
		if val, ok := val.(types.JSONDocument); ok {
			return val.JSONString()
		}
		return nil, errors.Errorf("GMSCast expected type `JSONDocument`, got `%T`", val)
	case query.Type_NULL_TYPE:
		return nil, nil
	case query.Type_GEOMETRY:
		return nil, errors.Errorf("GMS geometry types are not supported")
	default:
		return nil, errors.Errorf("GMS type `%s` is not supported", c.sqlChild.Type().String())
	}
}

// IsNullable implements the sql.Expression interface.
func (c *GMSCast) IsNullable() bool {
	return true
}

// Resolved implements the sql.Expression interface.
func (c *GMSCast) Resolved() bool {
	return c.sqlChild.Resolved()
}

// String implements the sql.Expression interface.
func (c *GMSCast) String() string {
	return c.sqlChild.String()
}

// Type implements the sql.Expression interface.
func (c *GMSCast) Type() sql.Type {
	return c.DoltgresType()
}

// WithChildren implements the sql.Expression interface.
func (c *GMSCast) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 1 {
		return nil, sql.ErrInvalidChildrenNumber.New(c, len(children), 1)
	}
	return &GMSCast{
		sqlChild: children[0],
	}, nil
}
