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

	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/postgres/parser/duration"
	"github.com/dolthub/doltgresql/postgres/parser/uuid"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// NewNumericLiteral returns a new *expression.Literal containing a NUMERIC value.
func NewNumericLiteral(numericValue string) (*expression.Literal, error) {
	d, err := decimal.NewFromString(numericValue)
	return expression.NewLiteral(d, pgtypes.Numeric), err
}

// NewIntegerLiteral returns a new *expression.Literal containing an integer (INT2/4/8 or NUMERIC) value.
func NewIntegerLiteral(integerValue string) (*expression.Literal, error) {
	i, err := strconv.ParseInt(integerValue, 10, 64)
	// If we don't get an error, then we know the value is either an INT32 or INT64
	if err == nil {
		if i >= -2147483648 && i <= 2147483647 {
			return expression.NewLiteral(int32(i), pgtypes.Int32), err
		} else {
			return expression.NewLiteral(i, pgtypes.Int64), err
		}
	} else {
		// If we errored the first time, then we'll assume it's a NUMERIC value
		d, err := decimal.NewFromString(integerValue)
		return expression.NewLiteral(d, pgtypes.Numeric), err
	}
}

// NewNullLiteral returns a new *expression.Literal containing a null value.
func NewNullLiteral() *expression.Literal {
return expression.NewLiteral(nil, pgtypes.Unknown)
}

// NewUnknownLiteral returns a new *expression.Literal containing a UNKNOWN type value.
func NewUnknownLiteral(stringValue string) *expression.Literal {
return expression.NewLiteral(stringValue, pgtypes.Unknown)
}

// NewTextLiteral returns a new *expression.Literal containing a TEXT type value.
// This should be used for internal uses when the type of the value is certain.
func NewTextLiteral(stringValue string) *expression.Literal {
return expression.NewLiteral(stringValue, pgtypes.Text)
}

// NewIntervalLiteral returns a new *expression.Literal containing a INTERVAL value.
func NewIntervalLiteral(duration duration.Duration) *expression.Literal {
return expression.NewLiteral(duration, pgtypes.Interval)
}

// NewJSONLiteral returns a new *expression.Literal containing a JSON value. This is different from JSONB.
func NewJSONLiteral(jsonValue string) *expression.Literal {
return expression.NewLiteral(jsonValue, pgtypes.Json)
}

// NewRawLiteralBool returns a new *expression.Literal containing a boolean value.
func NewRawLiteralBool(val bool) *expression.Literal {
return expression.NewLiteral(val, pgtypes.Bool)
}

// NewRawLiteralInt16 returns a new *expression.Literal containing an int16 value.
func NewRawLiteralInt16(val int16) *expression.Literal {
return expression.NewLiteral(val, pgtypes.Int16)
}

// NewRawLiteralInt32 returns a new *expression.Literal containing an int32 value.
func NewRawLiteralInt32(val int32) *expression.Literal {
return expression.NewLiteral(val, pgtypes.Int32)
}

// NewRawLiteralInt64 returns a new *expression.Literal containing an int64 value.
func NewRawLiteralInt64(val int64) *expression.Literal {
return expression.NewLiteral(val, pgtypes.Int64)
}

// NewRawLiteralFloat32 returns a new *expression.Literal containing a float32 value.
func NewRawLiteralFloat32(val float32) *expression.Literal {
return expression.NewLiteral(val, pgtypes.Float32)
}

// NewRawLiteralFloat64 returns a new *expression.Literal containing a float64 value.
func NewRawLiteralFloat64(val float64) *expression.Literal {
return expression.NewLiteral(val, pgtypes.Float64)
}

// NewRawLiteralNumeric returns a new *expression.Literal containing a decimal.Decimal value.
func NewRawLiteralNumeric(val decimal.Decimal) *expression.Literal {
return expression.NewLiteral(val, pgtypes.Numeric)
}

// NewRawLiteralDate returns a new *expression.Literal containing a DATE value.
func NewRawLiteralDate(date time.Time) *expression.Literal {
return expression.NewLiteral(date, pgtypes.Date)
}

// NewRawLiteralTime returns a new *expression.Literal containing a TIME value.
func NewRawLiteralTime(t time.Time) *expression.Literal {
return expression.NewLiteral(t, pgtypes.Time)
}

// NewRawLiteralTimeTZ returns a new *expression.Literal containing a TIMETZ value.
func NewRawLiteralTimeTZ(ttz time.Time) *expression.Literal {
return expression.NewLiteral(ttz, pgtypes.TimeTZ)
}

// NewRawLiteralTimestamp returns a new *expression.Literal containing a TIMESTAMP value. This is the variant without a time zone.
func NewRawLiteralTimestamp(val time.Time) *expression.Literal {
return expression.NewLiteral(val, pgtypes.Timestamp)
}

// NewRawLiteralTimestampTZ returns a new *expression.Literal containing a TIMESTAMPTZ value. This is the variant with a time zone.
func NewRawLiteralTimestampTZ(val time.Time) *expression.Literal {
return expression.NewLiteral(val, pgtypes.TimestampTZ)
}

// NewRawLiteralJSON returns a new *expression.Literal containing a JSON value.
func NewRawLiteralJSON(val string) *expression.Literal {
return expression.NewLiteral(val, pgtypes.Json)
}

// NewRawLiteralOid returns a new *expression.Literal containing a OID value.
func NewRawLiteralOid(val id.Id) *expression.Literal {
return expression.NewLiteral(val, pgtypes.Oid)
}

// NewRawLiteralUuid returns a new *expression.Literal containing a UUID value.
func NewRawLiteralUuid(val uuid.UUID) *expression.Literal {
return expression.NewLiteral(val, pgtypes.Uuid)
}

// NewUnsafeLiteral returns a new *expression.Literal containing the given value and type. This should almost never be used, as
// it does not perform any checking and circumvents type safety, which may lead to hard-to-debug errors. This is
// currently only used within the analyzer, and will likely be removed in the future.
func NewUnsafeLiteral(val any, t *pgtypes.DoltgresType) *expression.Literal {
return expression.NewLiteral(val, t)
}

// // ToVitessLiteral returns the literal as a Vitess literal. This is strictly for situations where GMS is hardcoded to
// // expect a Vitess literal. This should only be used as a temporary measure, as the GMS code needs to be updated, or the
// // equivalent functionality should be built into Doltgres (recommend the second approach).
// func (l *expression.Literal) ToVitessLiteral() *vitess.SQLVal {
// 	switch l.typ.ID {
// 	case pgtypes.Bool.ID:
// 		if l.value.(bool) {
// 			return vitess.NewIntVal([]byte("1"))
// 		} else {
// 			return vitess.NewIntVal([]byte("0"))
// 		}
// 	case pgtypes.Int32.ID:
// 		return vitess.NewIntVal([]byte(strconv.FormatInt(int64(l.value.(int32)), 10)))
// 	case pgtypes.Int64.ID:
// 		return vitess.NewIntVal([]byte(strconv.FormatInt(l.value.(int64), 10)))
// 	case pgtypes.Numeric.ID:
// 		return vitess.NewFloatVal([]byte(l.value.(decimal.Decimal).String()))
// 	case pgtypes.Text.ID:
// 		return vitess.NewStrVal([]byte(l.value.(string)))
// 	case pgtypes.Unknown.ID:
// 		if l.value == nil {
// 			return nil
// 		} else if str, ok := l.value.(string); ok {
// 			return vitess.NewStrVal([]byte(str))
// 		} else {
// 			panic("unhandled value of 'unknown' type in temporary literal conversion: " + l.typ.String())
// 		}
// 	default:
// 		panic("unhandled type in temporary literal conversion: " + l.typ.String())
// 	}
// }