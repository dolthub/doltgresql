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

package framework

import (
	"fmt"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/lib/pq/oid"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// NewLiteral is the implementation for NewLiteral function
// that is being set from expression package to avoid circular dependencies.
var NewLiteral func(input any, t pgtypes.DoltgresType) sql.Expression

// IoInput converts input string value to given type value.
func IoInput(ctx *sql.Context, t pgtypes.DoltgresType, input string) (any, error) {
	receivedVal := NewLiteral(input, pgtypes.Cstring)
	return receiveInputFunction(ctx, t.InputFunc, t, receivedVal)
}

// IoOutput converts given type value to output string.
func IoOutput(ctx *sql.Context, t pgtypes.DoltgresType, val any) (string, error) {
	o, err := sendOutputFunction(ctx, t.OutputFunc, t, val)
	if err != nil {
		return "", err
	}
	output, ok := o.(string)
	if !ok {
		return "", fmt.Errorf(`expected string, got %T`, output)
	}
	return output, nil
}

// IoReceive converts external binary format (which is a byte array) to given type value.
// Receive functions match and used for given type's deserialize value function.
func IoReceive(ctx *sql.Context, t pgtypes.DoltgresType, val any) (any, error) {
	if !t.ReceiveFuncExists() {
		return nil, fmt.Errorf("receive function for type '%s' doesn't exist", t.Name)
	}

	receivedVal := NewLiteral(val, pgtypes.NewInternalTypeWithBaseType(t.OID))
	return receiveInputFunction(ctx, t.ReceiveFunc, t, receivedVal)
}

// IoSend converts given type value to a byte array.
// Send functions match and used for given type's serialize value function.
func IoSend(ctx *sql.Context, t pgtypes.DoltgresType, val any) ([]byte, error) {
	if !t.SendFuncExists() {
		return nil, fmt.Errorf("send function for type '%s' doesn't exist", t.Name)
	}

	o, err := sendOutputFunction(ctx, t.SendFunc, t, val)
	if err != nil {
		return nil, err
	}
	if o == nil {
		return nil, nil
	}
	output, ok := o.([]byte)
	if !ok {
		return nil, fmt.Errorf(`expected []byte, got %T`, output)
	}
	return output, nil
}

// receiveInputFunction handles given IoInput and IoReceive functions.
func receiveInputFunction(ctx *sql.Context, funcName string, t pgtypes.DoltgresType, val sql.Expression) (any, error) {
	var cf *CompiledFunction
	var ok bool
	var err error
	if t.IsArrayType() {
		baseType := t.ArrayBaseType()
		typmod := int32(0)
		if baseType.ModInFunc != "-" {
			typmod = t.AttTypMod
		}
		cf, ok, err = GetFunction(funcName, val, NewLiteral(baseType.OID, pgtypes.Oid), NewLiteral(typmod, pgtypes.Int32))
	} else if t.TypType == pgtypes.TypeType_Domain {
		baseType := t.DomainUnderlyingBaseType()
		cf, ok, err = GetFunction(funcName, val, NewLiteral(baseType.OID, pgtypes.Oid), NewLiteral(t.AttTypMod, pgtypes.Int32))
	} else if t.ModInFunc != "-" {
		cf, ok, err = GetFunction(funcName, val, NewLiteral(t.OID, pgtypes.Oid), NewLiteral(t.AttTypMod, pgtypes.Int32))
	} else {
		cf, ok, err = GetFunction(funcName, val)
	}
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrFunctionDoesNotExist.New(funcName)
	}
	return cf.Eval(ctx, nil)
}

// sendOutputFunction handles given IoOutput and IoSend functions.
func sendOutputFunction(ctx *sql.Context, funcName string, t pgtypes.DoltgresType, val any) (any, error) {
	outputVal, ok, err := GetFunction(funcName, NewLiteral(val, t))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrFunctionDoesNotExist.New(funcName)
	}
	return outputVal.Eval(ctx, nil)
}

// TypModIn encodes given text array value to type modifier in int32 format.
func TypModIn(ctx *sql.Context, t pgtypes.DoltgresType, val []any) (int32, error) {
	// takes []string and return int32
	if t.ModInFunc == "-" {
		return 0, fmt.Errorf("typmodin function for type '%s' doesn't exist", t.Name)
	}
	v, ok, err := GetFunction(t.ModInFunc, NewLiteral(val, pgtypes.TextArray))
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, ErrFunctionDoesNotExist.New(t.ModInFunc)
	}
	o, err := v.Eval(ctx, nil)
	if err != nil {
		return 0, err
	}
	output, ok := o.(int32)
	if !ok {
		return 0, fmt.Errorf(`expected int32, got %T`, output)
	}
	return output, nil
}

// TypModOut decodes type modifier in int32 format to string representation of it.
func TypModOut(ctx *sql.Context, t pgtypes.DoltgresType, val int32) (string, error) {
	// takes int32 and returns string
	if t.ModOutFunc == "-" {
		return "", fmt.Errorf("typmodout function for type '%s' doesn't exist", t.Name)
	}
	v, ok, err := GetFunction(t.ModOutFunc, NewLiteral(val, pgtypes.Int32))
	if err != nil {
		return "", err
	}
	if !ok {
		return "", ErrFunctionDoesNotExist.New(t.ModOutFunc)
	}
	o, err := v.Eval(ctx, nil)
	if err != nil {
		return "", err
	}
	output, ok := o.(string)
	if !ok {
		return "", fmt.Errorf(`expected string, got %T`, output)
	}
	return output, nil
}

// IoCompare compares given two values using the given type.
// TODO: both values should have types. E.g.: to compare between float32 and float64
func IoCompare(ctx *sql.Context, t pgtypes.DoltgresType, v1, v2 any) (int32, error) {
	if v1 == nil && v2 == nil {
		return 0, nil
	} else if v1 != nil && v2 == nil {
		return 1, nil
	} else if v1 == nil && v2 != nil {
		return -1, nil
	}

	if t.CompareFunc == "-" {
		// TODO: use the type category's preferred type's compare function?
		return 0, fmt.Errorf("compare function does not exist for %s type", t.Name)
	}

	v, ok, err := GetFunction(t.CompareFunc, NewLiteral(v1, t), NewLiteral(v2, t))
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, ErrFunctionDoesNotExist.New(t.CompareFunc)
	}

	i, err := v.Eval(ctx, nil)
	if err != nil {
		return 0, err
	}
	output, ok := i.(int32)
	if !ok {
		return 0, fmt.Errorf(`expected int32, got %T`, output)
	}
	return output, nil
}

// SQL converts given type value to output string. This is the same as IoOutput function
// with an exception to BOOLEAN type. It returns "t" instead of "true".
func SQL(ctx *sql.Context, t pgtypes.DoltgresType, val any) (string, error) {
	if t.IsArrayType() {
		baseType := t.ArrayBaseType()
		if baseType.ModInFunc != "-" {
			baseType.AttTypMod = t.AttTypMod
		}
		return ArrToString(ctx, val.([]any), baseType, true)
	}
	// calling `out` function
	outputVal, ok, err := GetFunction(t.OutputFunc, NewLiteral(val, t))
	if err != nil {
		return "", err
	}
	if !ok {
		return "", ErrFunctionDoesNotExist.New(t.OutputFunc)
	}
	o, err := outputVal.Eval(ctx, nil)
	if err != nil {
		return "", err
	}
	output, ok := o.(string)
	if t.OID == uint32(oid.T_bool) {
		output = string(output[0])
	}
	if !ok {
		return "", fmt.Errorf(`expected string, got %T`, output)
	}
	return output, nil
}

// ArrToString is used for array_out function. |trimBool| parameter allows replacing
// boolean result of "true" to "t" if the function is `Type.SQL()`.
func ArrToString(ctx *sql.Context, arr []any, baseType pgtypes.DoltgresType, trimBool bool) (string, error) {
	sb := strings.Builder{}
	sb.WriteRune('{')
	for i, v := range arr {
		if i > 0 {
			sb.WriteString(",")
		}
		if v != nil {
			str, err := IoOutput(ctx, baseType, v)
			if err != nil {
				return "", err
			}
			if baseType.OID == uint32(oid.T_bool) && trimBool {
				str = string(str[0])
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
}
