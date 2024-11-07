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

// initArrayToString registers the functions to the catalog.
func initArrayToString() {
	framework.RegisterFunction(array_to_string_anyarray_text)
	framework.RegisterFunction(array_to_string_anyarray_text_text)
}

// array_to_string_anyarray_text represents the PostgreSQL array function taking 2 parameters.
var array_to_string_anyarray_text = framework.Function2{
	Name:               "array_to_string",
	Return:             pgtypes.Text,
	Parameters:         [2]pgtypes.DoltgresType{pgtypes.AnyArray, pgtypes.Text},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, paramsAndReturn [3]pgtypes.DoltgresType, val1, val2 any) (any, error) {
		arr := val1.([]any)
		delimiter := val2.(string)
		return getStringArrFromAnyArray(ctx, paramsAndReturn[0], arr, delimiter, nil)
	},
}

// array_to_string_anyarray_text_text represents the PostgreSQL array function taking 3 parameter.
var array_to_string_anyarray_text_text = framework.Function3{
	Name:               "array_to_string",
	Return:             pgtypes.Text,
	Parameters:         [3]pgtypes.DoltgresType{pgtypes.AnyArray, pgtypes.Text, pgtypes.Text},
	IsNonDeterministic: true,
	Strict:             false,
	Callable: func(ctx *sql.Context, paramsAndReturn [4]pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		if val1 == nil {
			return nil, fmt.Errorf("could not determine polymorphic type because input has type unknown")
		} else if val2 == nil {
			return nil, nil
		}
		arr := val1.([]any)
		delimiter := val2.(string)
		return getStringArrFromAnyArray(ctx, paramsAndReturn[0], arr, delimiter, val3)
	},
}

// getStringArrFromAnyArray takes inputs of any array, delimiter and null entry replacement. It uses the IoOutput() of the
// base type of the AnyArray type to get string representation of array elements.
func getStringArrFromAnyArray(ctx *sql.Context, arrType pgtypes.DoltgresType, arr []any, delimiter string, nullEntry any) (string, error) {
	baseType, ok := arrType.ArrayBaseType()
	if !ok {
		return "", fmt.Errorf("cannot get base type from %s", arrType.Name)
	}
	strs := make([]string, 0)
	for _, el := range arr {
		if el != nil {
			v, err := framework.IoOutput(ctx, baseType, el)
			if err != nil {
				return "", err
			}
			strs = append(strs, v)
		} else if nullEntry != nil {
			strs = append(strs, nullEntry.(string))
		}
	}
	return strings.Join(strs, delimiter), nil
}
