// Copyright 2026 Dolthub, Inc.
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

package framework

import (
	"encoding/json"
	"time"

	"github.com/cockroachdb/apd/v3"
	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/store/prolly/tree"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/vt/proto/query"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// ConvertGMSValueToDoltgresValue converts a value produced by an expression with the given GMS type into the Go
// representation expected by the Doltgres type most similar to it (the type that pgtypes.FromGmsTypeToDoltgresType
// returns). Expressions with GMS types appear wherever tables or functions are implemented in GMS rather than
// Doltgres, most notably the dolt_* system tables, whose values (e.g. the uint64 counts of dolt_statistics) don't
// necessarily match the Go representation that Doltgres consumers of the corresponding Doltgres type expect.
func ConvertGMSValueToDoltgresValue(ctx *sql.Context, sqlTyp sql.Type, val any) (any, error) {
	if val == nil {
		return nil, nil
	}
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
		// In Postgres, Int32 is generally the smallest value returned. But we convert int8 and int16 to this type during
		// schema conversion, which means we must do so here as well to avoid runtime panics.
	case query.Type_INT16, query.Type_INT24, query.Type_INT32, query.Type_YEAR:
		newVal, _, err := types.Int32.Convert(ctx, val)
		if err != nil {
			return nil, err
		}
		if _, ok := newVal.(int32); !ok {
			return nil, errors.Errorf("GMSCast expected type `int32`, got `%T`", val)
		}
		return newVal, nil
	case query.Type_INT64, query.Type_BIT, query.Type_UINT8, query.Type_UINT16, query.Type_UINT24, query.Type_UINT32:
		newVal, _, err := types.Int64.Convert(ctx, val)
		if err != nil {
			return nil, err
		}
		if _, ok := newVal.(int64); !ok {
			return nil, errors.Errorf("GMSCast expected type `int64`, got `%T`", val)
		}
		return newVal, nil
	case query.Type_UINT64:
		// Postgres doesn't have a "public" Uint64 type, so we return a Numeric value
		newVal, _, err := types.InternalDecimalType.Convert(ctx, val)
		if err != nil {
			return nil, err
		}
		dec, ok := newVal.(*apd.Decimal)
		if !ok {
			return nil, errors.Errorf("GMSCast expected type `*apd.Decimal`, got `%T`", val)
		}
		return dec, nil
	case query.Type_FLOAT32:
		newVal, _, err := types.Float32.Convert(ctx, val)
		if err != nil {
			return nil, err
		}
		if _, ok := newVal.(float32); !ok {
			return nil, errors.Errorf("GMSCast expected type `float32`, got `%T`", val)
		}
		return newVal, nil
	case query.Type_FLOAT64:
		newVal, _, err := types.Float64.Convert(ctx, val)
		if err != nil {
			return nil, err
		}
		if _, ok := newVal.(float64); !ok {
			return nil, errors.Errorf("GMSCast expected type `float64`, got `%T`", val)
		}
		return newVal, nil
	case query.Type_DECIMAL:
		newVal, _, err := types.InternalDecimalType.Convert(ctx, val)
		if err != nil {
			return nil, err
		}
		dec, ok := newVal.(*apd.Decimal)
		if !ok {
			return nil, errors.Errorf("GMSCast expected type `*apd.Decimal`, got `%T`", val)
		}
		return dec, nil
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
	case query.Type_CHAR, query.Type_VARCHAR, query.Type_TEXT, query.Type_BINARY, query.Type_VARBINARY, query.Type_BLOB, query.Type_SET, query.Type_ENUM:
		newVal, _, err := types.LongText.Convert(ctx, val)
		if err != nil {
			return nil, err
		}
		switch newVal := newVal.(type) {
		case string:
			return newVal, nil
		case sql.StringWrapper:
			return newVal.Unwrap(ctx)
		default:
			return nil, errors.Errorf("GMSCast expected type `string`, got `%T`", val)
		}
	case query.Type_JSON:
		switch val := val.(type) {
		case types.JSONDocument:
			return val.JSONString()
		case tree.IndexedJsonDocument:
			return val.String(), nil
		default:
			// TODO: there are particular dolt tables (dolt_constraint_violations) that return json-marshallable structs
			//  that we need to handle here like this
			bytes, err := json.Marshal(val)
			return string(bytes), err
		}
	case query.Type_NULL_TYPE:
		return nil, nil
	case query.Type_GEOMETRY:
		return nil, errors.Errorf("GMS geometry types are not supported")
	default:
		return nil, errors.Errorf("GMS type `%s` is not supported", sqlTyp.String())
	}
}

// gmsCastExpr wraps an expression that has a GMS type, converting the values it produces into the Go representation
// expected by the corresponding Doltgres type. It serves the same purpose as the GMSCast expression in the
// server/expression package (which cannot be used here, as that package depends on this one): aggregate and window
// function arguments are evaluated directly by their buffer implementations rather than passing through
// CompiledFunction.Eval, so any GMS-typed arguments must convert their values themselves.
type gmsCastExpr struct {
	child sql.Expression
}

var _ sql.Expression = (*gmsCastExpr)(nil)

// castGMSArguments wraps each argument that has a GMS type in a gmsCastExpr, so that its evaluated values are
// converted to Doltgres values. Arguments that already have Doltgres types, or no type at all, are left alone.
func castGMSArguments(ctx *sql.Context, args []sql.Expression) []sql.Expression {
	for i, arg := range args {
		if t := arg.Type(ctx); t != nil {
			if _, ok := t.(*pgtypes.DoltgresType); !ok {
				args[i] = &gmsCastExpr{child: arg}
			}
		}
	}
	return args
}

// Children implements the sql.Expression interface.
func (c *gmsCastExpr) Children() []sql.Expression {
	return []sql.Expression{c.child}
}

// Eval implements the sql.Expression interface.
func (c *gmsCastExpr) Eval(ctx *sql.Context, row sql.Row) (any, error) {
	val, err := c.child.Eval(ctx, row)
	if err != nil {
		return nil, err
	}
	return ConvertGMSValueToDoltgresValue(ctx, c.child.Type(ctx), val)
}

// IsNullable implements the sql.Expression interface.
func (c *gmsCastExpr) IsNullable(ctx *sql.Context) bool {
	return c.child.IsNullable(ctx)
}

// Resolved implements the sql.Expression interface.
func (c *gmsCastExpr) Resolved() bool {
	return c.child.Resolved()
}

// String implements the sql.Expression interface.
func (c *gmsCastExpr) String() string {
	return c.child.String()
}

// Type implements the sql.Expression interface.
func (c *gmsCastExpr) Type(ctx *sql.Context) sql.Type {
	return pgtypes.FromGmsType(c.child.Type(ctx))
}

// WithChildren implements the sql.Expression interface.
func (c *gmsCastExpr) WithChildren(ctx *sql.Context, children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 1 {
		return nil, sql.ErrInvalidChildrenNumber.New(c, len(children), 1)
	}
	return &gmsCastExpr{child: children[0]}, nil
}
