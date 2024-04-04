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
	"strconv"

	"github.com/dolthub/go-mysql-server/sql"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// Literal represents a raw literal (number, string, etc.).
type Literal struct {
	value any
	typ   pgtypes.DoltgresType
}

var _ vitess.InjectableExpression = (*Literal)(nil)
var _ sql.Expression = (*Literal)(nil)
var _ framework.LiteralInterface = (*Literal)(nil)

// NewBoolLiteral returns a new *Literal containing a boolean value.
func NewBoolLiteral(val bool) *Literal {
	return &Literal{
		value: val,
		typ:   pgtypes.Bool,
	}
}

// NewNumericLiteral returns a new *Literal containing a NUMERIC value.
func NewNumericLiteral(numericValue string) (*Literal, error) {
	d, err := decimal.NewFromString(numericValue)
	return &Literal{
		value: d,
		typ:   pgtypes.Numeric,
	}, err
}

// NewIntegerLiteral returns a new *Literal containing an integer (INT2/4/8 or NUMERIC) value.
func NewIntegerLiteral(integerValue string) (*Literal, error) {
	i, err := strconv.ParseInt(integerValue, 10, 64)
	// If we don't get an error, then we know the value is either an INT32 or INT64
	if err == nil {
		if i >= -2147483648 && i <= 2147483647 {
			return &Literal{
				value: int32(i),
				typ:   pgtypes.Int32,
			}, nil
		} else {
			return &Literal{
				value: i,
				typ:   pgtypes.Int64,
			}, nil
		}
	} else {
		// If we errored the first time, then we'll assume it's a NUMERIC value
		d, err := decimal.NewFromString(integerValue)
		return &Literal{
			value: d,
			typ:   pgtypes.Numeric,
		}, err
	}
}

// NewStringLiteral returns a new *Literal containing a VARCHAR value.
func NewStringLiteral(stringValue string) (*Literal, error) {
	return &Literal{
		value: stringValue,
		typ:   pgtypes.VarChar,
	}, nil
}

// Children implements the sql.Expression interface.
func (l *Literal) Children() []sql.Expression {
	return nil
}

// ConformsToLiteralInterface implements the framework.LiteralInterface interface.
func (l *Literal) ConformsToLiteralInterface() {}

// Eval implements the sql.Expression interface.
func (l *Literal) Eval(ctx *sql.Context, row sql.Row) (any, error) {
	return l.value, nil
}

// GetDoltgresType implements the framework.LiteralInterface interface.
func (l *Literal) GetDoltgresType() pgtypes.DoltgresType {
	return l.typ
}

// IsNullable implements the sql.Expression interface.
func (l *Literal) IsNullable() bool {
	return l.value == nil
}

// Resolved implements the sql.Expression interface.
func (l *Literal) Resolved() bool {
	return true
}

// String implements the sql.Expression interface.
func (l *Literal) String() string {
	return fmt.Sprintf("%v", l.value)
}

// ToVitessLiteral returns the literal as a Vitess literal. This is strictly for situations where GMS is hardcoded to
// expect a Vitess literal. This should only be used as a temporary measure, as the GMS code needs to be updated, or the
// equivalent functionality should be built into Doltgres (recommend the second approach).
func (l *Literal) ToVitessLiteral() *vitess.SQLVal {
	switch l.typ.BaseID() {
	case pgtypes.Bool.BaseID():
		if l.value.(bool) {
			return vitess.NewIntVal([]byte("1"))
		} else {
			return vitess.NewIntVal([]byte("0"))
		}
	case pgtypes.Int32.BaseID():
		return vitess.NewIntVal([]byte(strconv.FormatInt(int64(l.value.(int32)), 10)))
	case pgtypes.Int64.BaseID():
		return vitess.NewIntVal([]byte(strconv.FormatInt(l.value.(int64), 10)))
	case pgtypes.Numeric.BaseID():
		return vitess.NewFloatVal([]byte(l.value.(decimal.Decimal).String()))
	case pgtypes.VarChar.BaseID():
		return vitess.NewStrVal([]byte(l.value.(string)))
	default:
		panic("unhandled type in temporary literal conversion: " + l.typ.String())
	}
}

// Type implements the sql.Expression interface.
func (l *Literal) Type() sql.Type {
	return l.typ
}

// Value returns the literal value.
func (l *Literal) Value() any {
	return l.value
}

// WithChildren implements the sql.Expression interface.
func (l *Literal) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 1 {
		return nil, sql.ErrInvalidChildrenNumber.New(l, len(children), 1)
	}
	return l, nil
}

// WithResolvedChildren implements the vitess.InjectableExpression interface.
func (l *Literal) WithResolvedChildren(children []any) (any, error) {
	if len(children) != 0 {
		return nil, fmt.Errorf("invalid vitess child count, expected `0` but got `%d`", len(children))
	}
	return l, nil
}
