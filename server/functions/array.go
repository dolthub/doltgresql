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

package functions

import (
	"fmt"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initBinaryNotEqual registers the functions to the catalog.
func initArray() {
	framework.RegisterFunction(array_in)
	framework.RegisterFunction(array_out)
	framework.RegisterFunction(array_recv)
	framework.RegisterFunction(array_send)
	framework.RegisterFunction(btarraycmp)
}

// array_in represents the PostgreSQL function of array type IO input.
var array_in = framework.Function3{
	Name:       "array_in",
	Return:     pgtypes.AnyArray,
	Parameters: [3]pgtypes.DoltgresType{pgtypes.Text, pgtypes.Oid, pgtypes.Int32}, // cstring
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		input := val1.(string)
		oid := val2.(uint32) // TODO: is this oid of base type??
		// TODO: what is the third typmod
		baseType := pgtypes.OidToBuildInDoltgresType[oid]
		if len(input) < 2 || input[0] != '{' || input[len(input)-1] != '}' {
			// This error is regarded as a critical error, and thus we immediately return the error alongside a nil
			// value. Returning a nil value is a signal to not ignore the error.
			return nil, fmt.Errorf(`malformed array literal: "%s"`, input)
		}
		// We'll remove the surrounding braces since we've already verified that they're there
		input = input[1 : len(input)-1]
		var values []any
		var err error
		sb := strings.Builder{}
		quoteStartCount := 0
		quoteEndCount := 0
		escaped := false
		// Iterate over each rune in the input to collect and process the rune elements
		for _, r := range input {
			if escaped {
				sb.WriteRune(r)
				escaped = false
			} else if quoteStartCount > quoteEndCount {
				switch r {
				case '\\':
					escaped = true
				case '"':
					quoteEndCount++
				default:
					sb.WriteRune(r)
				}
			} else {
				switch r {
				case ' ', '\t', '\n', '\r':
					continue
				case '\\':
					escaped = true
				case '"':
					quoteStartCount++
				case ',':
					if quoteStartCount >= 2 {
						// This is a malformed string, thus we treat it as a critical error.
						return nil, fmt.Errorf(`malformed array literal: "%s"`, input)
					}
					str := sb.String()
					var innerValue any
					if quoteStartCount == 0 && strings.EqualFold(str, "null") {
						// An unquoted case-insensitive NULL is treated as an actual null value
						innerValue = nil
					} else {
						var nErr error
						innerValue, nErr = framework.IoInput(ctx, baseType, str)
						if nErr != nil && err == nil {
							// This is a non-critical error, therefore the error may be ignored at a higher layer (such as
							// an explicit cast) and the inner type will still return a valid result, so we must allow the
							// values to propagate.
							err = nErr
						}
					}
					values = append(values, innerValue)
					sb.Reset()
					quoteStartCount = 0
					quoteEndCount = 0
				default:
					sb.WriteRune(r)
				}
			}
		}
		// Use anything remaining in the buffer as the last element
		if sb.Len() > 0 {
			if escaped || quoteStartCount > quoteEndCount || quoteStartCount >= 2 {
				// These errors are regarded as critical errors, and thus we immediately return the error alongside a nil
				// value. Returning a nil value is a signal to not ignore the error.
				return nil, fmt.Errorf(`malformed array literal: "%s"`, input)
			} else {
				str := sb.String()
				var innerValue any
				if quoteStartCount == 0 && strings.EqualFold(str, "NULL") {
					// An unquoted case-insensitive NULL is treated as an actual null value
					innerValue = nil
				} else {
					var nErr error
					innerValue, nErr = framework.IoInput(ctx, baseType, str)
					if nErr != nil && err == nil {
						// This is a non-critical error, therefore the error may be ignored at a higher layer (such as
						// an explicit cast) and the inner type will still return a valid result, so we must allow the
						// values to propagate.
						err = nErr
					}
				}
				values = append(values, innerValue)
			}
		}

		return values, err
	},
}

// array_out represents the PostgreSQL function of array type IO output.
var array_out = framework.Function1{
	Name:       "array_out",
	Return:     pgtypes.Text, // cstring
	Parameters: [1]pgtypes.DoltgresType{pgtypes.AnyArray},
	Strict:     true,
	Callable: func(ctx *sql.Context, t [2]pgtypes.DoltgresType, val any) (any, error) {
		// TODO: should the input be converted or should be converted here?
		//converted, _, err := ac.Convert(output)
		//if err != nil {
		//	return "", err
		//}

		arrType := t[0]
		if !arrType.IsArrayType() {
			// TODO: shouldn't happen but check??
			return nil, fmt.Errorf(`not array type`)
		}
		baseType, ok := arrType.ArrayBaseType()
		if !ok {
			// TODO: shouldn't happen but check??
			return nil, fmt.Errorf(`cannot find base type for array type`)
		}

		sb := strings.Builder{}
		sb.WriteRune('{')
		for i, v := range val.([]any) {
			if i > 0 {
				sb.WriteString(",")
			}
			if v != nil {
				str, err := framework.IoOutput(ctx, baseType, v)
				if err != nil {
					return "", err
				}
				shouldQuote := false
				for _, r := range str {
					switch r {
					case ' ', ',', '{', '}', '\\', '"':
						shouldQuote = true
					}
				}
				if shouldQuote || strings.EqualFold(str, "NULL") {
					sb.WriteRune('"')
					sb.WriteString(strings.ReplaceAll(str, `"`, `\"`))
					sb.WriteRune('"')
				} else {
					sb.WriteString(str)
				}
			} else {
				sb.WriteString("NULL")
			}
		}
		sb.WriteRune('}')
		return sb.String(), nil
	},
}

// array_recv represents the PostgreSQL function of array type IO receive.
var array_recv = framework.Function3{
	Name:       "array_recv",
	Return:     pgtypes.AnyArray,
	Parameters: [3]pgtypes.DoltgresType{pgtypes.Internal, pgtypes.Oid, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		input := val1.(string)
		oid := val2.(uint32) // TODO: is this oid of base type??
		// TODO: what is the third argument for??
		baseType := pgtypes.OidToBuildInDoltgresType[oid]
		return framework.IoReceive(ctx, baseType, input)
	},
}

// array_send represents the PostgreSQL function of array type IO send.
var array_send = framework.Function1{
	Name:       "array_send",
	Return:     pgtypes.Bytea,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.AnyArray},
	Strict:     true,
	Callable: func(ctx *sql.Context, t [2]pgtypes.DoltgresType, val any) (any, error) {
		arrType := t[0]
		if !arrType.IsArrayType() {
			// TODO: shouldn't happen but check??
			return nil, fmt.Errorf(`not array type`)
		}
		baseType, ok := arrType.ArrayBaseType()
		if !ok {
			// TODO: shouldn't happen but check??
			return nil, fmt.Errorf(`cannot find base type for array type`)
		}

		sb := strings.Builder{}
		sb.WriteRune('{')
		for i, v := range val.([]any) {
			if i > 0 {
				sb.WriteString(",")
			}
			if v != nil {
				str, err := framework.IoSend(ctx, baseType, v)
				if err != nil {
					return "", err
				}
				shouldQuote := false
				for _, r := range str {
					switch r {
					case ' ', ',', '{', '}', '\\', '"':
						shouldQuote = true
					}
				}
				if shouldQuote || strings.EqualFold(string(str), "NULL") {
					sb.WriteRune('"')
					sb.WriteString(strings.ReplaceAll(string(str), `"`, `\"`))
					sb.WriteRune('"')
				} else {
					sb.WriteString(string(str))
				}
			} else {
				sb.WriteString("NULL")
			}
		}
		sb.WriteRune('}')
		return []byte(sb.String()), nil
	},
}

// btarraycmp represents the PostgreSQL function of array type byte compare.
var btarraycmp = framework.Function2{
	Name:       "btarraycmp",
	Return:     pgtypes.Int32,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.AnyArray, pgtypes.AnyArray},
	Strict:     true,
	Callable: func(ctx *sql.Context, t [3]pgtypes.DoltgresType, val1, val2 any) (any, error) {
		// TODO
		return int32(1), nil
	},
}
