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
	"time"

	"github.com/dolthub/go-mysql-server/sql"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/postgres/parser/duration"
	"github.com/dolthub/doltgresql/postgres/parser/uuid"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// Literal represents a raw literal (number, string, etc.).
type Literal struct {
	value any
	typ   pgtypes.DoltgresType
}

var _ vitess.Injectable = (*Literal)(nil)
var _ sql.Expression = (*Literal)(nil)
var _ framework.LiteralInterface = (*Literal)(nil)

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

// NewNullLiteral returns a new *Literal containing a null value.
func NewNullLiteral() *Literal {
	return &Literal{
		value: nil,
		typ:   pgtypes.Unknown,
	}
}

// NewUnknownLiteral returns a new *Literal containing a UNKNOWN type value.
func NewUnknownLiteral(stringValue string) *Literal {
	return &Literal{
		value: stringValue,
		typ:   pgtypes.Unknown,
	}
}

// NewTextLiteral returns a new *Literal containing a TEXT type value.
// This should be used for internal uses when the type of the value is certain.
func NewTextLiteral(stringValue string) *Literal {
	return &Literal{
		value: stringValue,
		typ:   pgtypes.Text,
	}
}

// NewIntervalLiteral returns a new *Literal containing a INTERVAL value.
func NewIntervalLiteral(duration duration.Duration) *Literal {
	return &Literal{
		value: duration,
		typ:   pgtypes.Interval,
	}
}

// NewJSONLiteral returns a new *Literal containing a JSON value. This is different from JSONB.
func NewJSONLiteral(jsonValue string) *Literal {
	return &Literal{
		value: jsonValue,
		typ:   pgtypes.Json,
	}
}

// NewRawLiteralBool returns a new *Literal containing a boolean value.
func NewRawLiteralBool(val bool) *Literal {
	return &Literal{
		value: val,
		typ:   pgtypes.Bool,
	}
}

// NewRawLiteralInt64 returns a new *Literal containing an int64 value.
func NewRawLiteralInt64(val int64) *Literal {
	return &Literal{
		value: val,
		typ:   pgtypes.Int64,
	}
}

// NewRawLiteralFloat64 returns a new *Literal containing a float64 value.
func NewRawLiteralFloat64(val float64) *Literal {
	return &Literal{
		value: val,
		typ:   pgtypes.Float64,
	}
}

// NewRawLiteralNumeric returns a new *Literal containing a decimal.Decimal value.
func NewRawLiteralNumeric(val decimal.Decimal) *Literal {
	return &Literal{
		value: val,
		typ:   pgtypes.Numeric,
	}
}

// NewRawLiteralDate returns a new *Literal containing a DATE value.
func NewRawLiteralDate(date time.Time) *Literal {
	return &Literal{
		value: date,
		typ:   pgtypes.Date,
	}
}

// NewRawLiteralTime returns a new *Literal containing a TIME value.
func NewRawLiteralTime(t time.Time) *Literal {
	return &Literal{
		value: t,
		typ:   pgtypes.Time,
	}
}

// NewRawLiteralTimeTZ returns a new *Literal containing a TIMETZ value.
func NewRawLiteralTimeTZ(ttz time.Time) *Literal {
	return &Literal{
		value: ttz,
		typ:   pgtypes.TimeTZ,
	}
}

// NewRawLiteralTimestamp returns a new *Literal containing a TIMESTAMP value. This is the variant without a time zone.
func NewRawLiteralTimestamp(val time.Time) *Literal {
	return &Literal{
		value: val,
		typ:   pgtypes.Timestamp,
	}
}

// NewRawLiteralTimestampTZ returns a new *Literal containing a TIMESTAMPTZ value. This is the variant with a time zone.
func NewRawLiteralTimestampTZ(val time.Time) *Literal {
	return &Literal{
		value: val,
		typ:   pgtypes.TimestampTZ,
	}
}

// NewRawLiteralJSON returns a new *Literal containing a JSON value.
func NewRawLiteralJSON(val string) *Literal {
	return &Literal{
		value: val,
		typ:   pgtypes.Json,
	}
}

// NewRawLiteralOid returns a new *Literal containing a OID value.
func NewRawLiteralOid(val uint32) *Literal {
	return &Literal{
		value: val,
		typ:   pgtypes.Oid,
	}
}

// NewRawLiteralUuid returns a new *Literal containing a UUID value.
func NewRawLiteralUuid(val uuid.UUID) *Literal {
	return &Literal{
		value: val,
		typ:   pgtypes.Uuid,
	}
}

// NewUnsafeLiteral returns a new *Literal containing the given value and type. This should almost never be used, as
// it does not perform any checking and circumvents type safety, which may lead to hard-to-debug errors. This is
// currently only used within the analyzer, and will likely be removed in the future.
func NewUnsafeLiteral(val any, t pgtypes.DoltgresType) *Literal {
	return &Literal{
		value: val,
		typ:   t,
	}
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
	if l.value == nil {
		return ""
	}
	str, err := l.typ.IoOutput(nil, l.value)
	if err != nil {
		panic("got error from IoOutput")
	}
	return pgtypes.QuoteString(l.typ.BaseID(), str)
}

// ToVitessLiteral returns the literal as a Vitess literal. This is strictly for situations where GMS is hardcoded to
// expect a Vitess literal. This should only be used as a temporary measure, as the GMS code needs to be updated, or the
// equivalent functionality should be built into Doltgres (recommend the second approach).
func (l *Literal) ToVitessLiteral() *vitess.SQLVal {
	switch l.typ.BaseID() {
	case pgtypes.DoltgresTypeBaseID_Bool:
		if l.value.(bool) {
			return vitess.NewIntVal([]byte("1"))
		} else {
			return vitess.NewIntVal([]byte("0"))
		}
	case pgtypes.DoltgresTypeBaseID_Int32:
		return vitess.NewIntVal([]byte(strconv.FormatInt(int64(l.value.(int32)), 10)))
	case pgtypes.DoltgresTypeBaseID_Int64:
		return vitess.NewIntVal([]byte(strconv.FormatInt(l.value.(int64), 10)))
	case pgtypes.DoltgresTypeBaseID_Numeric:
		return vitess.NewFloatVal([]byte(l.value.(decimal.Decimal).String()))
	case pgtypes.DoltgresTypeBaseID_Text:
		return vitess.NewStrVal([]byte(l.value.(string)))
	case pgtypes.DoltgresTypeBaseID_Unknown:
		if l.value == nil {
			return nil
		} else if str, ok := l.value.(string); ok {
			return vitess.NewStrVal([]byte(str))
		} else {
			panic("unhandled value of 'unknown' type in temporary literal conversion: " + l.typ.String())
		}
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
	if len(children) != 0 {
		return nil, sql.ErrInvalidChildrenNumber.New(l, len(children), 0)
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
