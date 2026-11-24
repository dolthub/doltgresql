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

package functions

import (
	"math"
	"strings"
	"unicode/utf8"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// maxFormatLength is the longest result, in bytes, that PostgreSQL can return from format().
const maxFormatLength = 1073741819

var (
	// errUnterminatedFormatSpecifier is returned when a format() string ends before a `%` reaches its type letter.
	errUnterminatedFormatSpecifier = pgerror.New(pgcode.InvalidParameterValue, "unterminated format() type specifier")
	// errTooFewFormatArguments is returned when a format() string uses more arguments than were given.
	errTooFewFormatArguments = pgerror.New(pgcode.InvalidParameterValue, "too few arguments for format()")
	// errFormatNumberOutOfRange is returned when a number in a format() string, or a width argument, is too large.
	errFormatNumberOutOfRange = pgerror.New(pgcode.NumericValueOutOfRange, "number is out of range")
	// errFormatArgumentZero is returned when a format() string asks for argument 0, since arguments count from 1.
	errFormatArgumentZero = pgerror.New(pgcode.InvalidParameterValue, "format specifies argument 0, but arguments are numbered from 1")
)

// initFormat registers the functions to the catalog.
func initFormat() {
	framework.RegisterFunction(format_text)
}

// format_text represents the PostgreSQL functions format(text) and format(text, VARIADIC "any").
var format_text = framework.Function1N{
	Name:       "format",
	Return:     pgtypes.Text,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Text},
	Strict:     false,
	Callable: func(ctx *sql.Context, t []*pgtypes.DoltgresType, val1 any, vals []any) (any, error) {
		if val1 == nil {
			return nil, nil
		}
		formatStr, err := framework.UnwrapString(ctx, val1)
		if err != nil {
			return nil, err
		}
		sb := strings.Builder{}
		nextArg := 0
		for i := 0; i < len(formatStr); i++ {
			if formatStr[i] != '%' {
				sb.WriteByte(formatStr[i])
				continue
			}
			if i++; i >= len(formatStr) {
				return nil, errUnterminatedFormatSpecifier
			}
			if formatStr[i] == '%' {
				sb.WriteByte('%')
				continue
			}
			argPos, widthPos, leftAlign, width, next, err := parseFormatSpecifier(formatStr, i)
			if err != nil {
				return nil, err
			}
			i = next
			conversion := formatStr[i]
			if conversion != 's' && conversion != 'I' && conversion != 'L' {
				r, _ := utf8.DecodeRuneInString(formatStr[i:])
				return nil, pgerror.Newf(pgcode.InvalidParameterValue, `unrecognized format() type specifier "%c"`, r)
			}
			if widthPos >= 0 {
				if widthPos > 0 {
					nextArg = widthPos - 1
				}
				if nextArg >= len(vals) {
					return nil, errTooFewFormatArguments
				}
				if width, err = formatWidthArgument(ctx, t[nextArg+1], vals[nextArg]); err != nil {
					return nil, err
				}
				nextArg++
			}
			if argPos > 0 {
				nextArg = argPos - 1
			}
			if nextArg >= len(vals) {
				return nil, errTooFewFormatArguments
			}
			val := vals[nextArg]
			var str string
			if val == nil {
				switch conversion {
				case 'I':
					return nil, pgerror.New(pgcode.NullValueNotAllowed, "null values cannot be formatted as an SQL identifier")
				case 'L':
					str = "NULL"
				}
			} else {
				output, err := t[nextArg+1].SQL(ctx, nil, val)
				if err != nil {
					return nil, err
				}
				str = output.ToString()
				switch conversion {
				case 'I':
					str = quoteTypeIdentifier(str)
				case 'L':
					str = quoteLiteral(str)
				}
			}
			nextArg++
			if width < 0 {
				if width == math.MinInt32 {
					return nil, errFormatNumberOutOfRange
				}
				leftAlign = true
				width = -width
			}
			paddingLen := max(int(width)-utf8.RuneCountInString(str), 0)
			if sb.Len()+len(str)+paddingLen > maxFormatLength {
				return nil, pgerror.New(pgcode.ProgramLimitExceeded, "out of memory")
			}
			padding := strings.Repeat(" ", paddingLen)
			if leftAlign {
				sb.WriteString(str)
				sb.WriteString(padding)
			} else {
				sb.WriteString(padding)
				sb.WriteString(str)
			}
		}
		return sb.String(), nil
	},
}

// parseFormatSpecifier reads what sits between a `%` and its type letter in a format() string, such as the `2$-10`
// in `%2$-10s`, and returns the index of the type letter. Argument positions count from 1 and are -1 when omitted.
// A width written as `*` is read from an argument, whose position is 0 when the `*` does not give one.
func parseFormatSpecifier(formatStr string, i int) (argPos int, widthPos int, leftAlign bool, width int32, next int, err error) {
	argPos, widthPos = -1, -1
	n, found, i, err := parseFormatDigits(formatStr, i)
	if err != nil {
		return 0, 0, false, 0, 0, err
	}
	if found {
		if formatStr[i] != '$' {
			return argPos, widthPos, false, n, i, nil
		}
		if n == 0 {
			return 0, 0, false, 0, 0, errFormatArgumentZero
		}
		argPos = int(n)
		if i++; i >= len(formatStr) {
			return 0, 0, false, 0, 0, errUnterminatedFormatSpecifier
		}
	}
	for formatStr[i] == '-' {
		leftAlign = true
		if i++; i >= len(formatStr) {
			return 0, 0, false, 0, 0, errUnterminatedFormatSpecifier
		}
	}
	if formatStr[i] != '*' {
		width, _, i, err = parseFormatDigits(formatStr, i)
		return argPos, widthPos, leftAlign, width, i, err
	}
	if i++; i >= len(formatStr) {
		return 0, 0, false, 0, 0, errUnterminatedFormatSpecifier
	}
	n, found, i, err = parseFormatDigits(formatStr, i)
	if err != nil {
		return 0, 0, false, 0, 0, err
	}
	if !found {
		return argPos, 0, leftAlign, 0, i, nil
	}
	if formatStr[i] != '$' {
		return 0, 0, false, 0, 0, pgerror.New(pgcode.InvalidParameterValue, `width argument position must be ended by "$"`)
	}
	if n == 0 {
		return 0, 0, false, 0, 0, errFormatArgumentZero
	}
	if i++; i >= len(formatStr) {
		return 0, 0, false, 0, 0, errUnterminatedFormatSpecifier
	}
	return argPos, int(n), leftAlign, 0, i, nil
}

// parseFormatDigits reads the number starting at index `i` of a format() string, and returns it along with whether
// a number was there and the index just after it.
func parseFormatDigits(formatStr string, i int) (n int32, found bool, next int, err error) {
	for formatStr[i] >= '0' && formatStr[i] <= '9' {
		digit := int64(formatStr[i] - '0')
		if int64(n)*10+digit > math.MaxInt32 {
			return 0, false, 0, errFormatNumberOutOfRange
		}
		n = n*10 + int32(digit)
		found = true
		if i++; i >= len(formatStr) {
			return 0, false, 0, errUnterminatedFormatSpecifier
		}
	}
	return n, found, i, nil
}

// formatWidthArgument returns the width that a `*` in a format() string reads from an argument. A NULL counts as
// zero, and any type other than an integer is converted through its text form.
func formatWidthArgument(ctx *sql.Context, typ *pgtypes.DoltgresType, val any) (int32, error) {
	switch val := val.(type) {
	case nil:
		return 0, nil
	case int32:
		return val, nil
	case int16:
		return int32(val), nil
	}
	str, err := typ.IoOutput(ctx, val)
	if err != nil {
		return 0, err
	}
	width, err := pgtypes.Int32.IoInput(ctx, str)
	if err != nil {
		return 0, err
	}
	return width.(int32), nil
}

// quoteLiteral returns the given string as a quoted SQL literal, matching PostgreSQL's quote_literal.
func quoteLiteral(str string) string {
	quoted := "'" + strings.ReplaceAll(strings.ReplaceAll(str, `\`, `\\`), "'", "''") + "'"
	if strings.Contains(str, `\`) {
		return "E" + quoted
	}
	return quoted
}
