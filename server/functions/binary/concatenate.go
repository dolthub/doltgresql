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

package binary

import (
	"fmt"

	"github.com/dolthub/go-mysql-server/sql"
	gmstypes "github.com/dolthub/go-mysql-server/sql/types"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// These functions can be gathered using the following query from a Postgres 15 instance:
// SELECT * FROM pg_operator o WHERE o.oprname = '||' ORDER BY o.oprcode::varchar;

// initBinaryConcatenate registers the functions to the catalog.
func initBinaryConcatenate() {
	framework.RegisterBinaryFunction(framework.Operator_BinaryConcatenate, anytextcat)
	framework.RegisterBinaryFunction(framework.Operator_BinaryConcatenate, array_append)
	framework.RegisterBinaryFunction(framework.Operator_BinaryConcatenate, array_cat)
	framework.RegisterBinaryFunction(framework.Operator_BinaryConcatenate, array_prepend)
	framework.RegisterBinaryFunction(framework.Operator_BinaryConcatenate, byteacat)
	framework.RegisterBinaryFunction(framework.Operator_BinaryConcatenate, jsonb_concat)
	framework.RegisterBinaryFunction(framework.Operator_BinaryConcatenate, textanycat)
	framework.RegisterBinaryFunction(framework.Operator_BinaryConcatenate, textcat)
	// TODO: bitcat, tsquery_or, tsvector_concat
}

// anytextcat_callable is the callable logic for the anytextcat function.
func anytextcat_callable(ctx *sql.Context, paramsAndReturn [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	valType := paramsAndReturn[0]
	val1String, err := valType.IoOutput(ctx, val1)
	if err != nil {
		return nil, err
	}
	return val1String + val2.(string), nil
}

// anytextcat represents the PostgreSQL function of the same name, taking the same parameters.
var anytextcat = framework.Function2{
	Name:       "anytextcat",
	Return:     pgtypes.Text,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.AnyNonArray, pgtypes.Text},
	Strict:     true,
	Callable:   anytextcat_callable,
}

// array_append represents the PostgreSQL function of the same name, taking the same parameters.
var array_append = framework.Function2{
	Name:       "array_append",
	Return:     pgtypes.AnyArray,                                               // TODO: should be anycompatiblearray
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.AnyArray, pgtypes.AnyElement}, // TODO: should be anycompatiblearray, anycompatible
	Strict:     false,
	Callable: func(ctx *sql.Context, paramsAndReturn [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		if val1 == nil {
			return []any{val2}, nil
		}
		array := val1.([]any)
		returnArray := make([]any, len(array)+1)
		copy(returnArray, array)
		returnArray[len(returnArray)-1] = val2
		return returnArray, nil
	},
}

// array_cat represents the PostgreSQL function of the same name, taking the same parameters.
var array_cat = framework.Function2{
	Name:       "array_cat",
	Return:     pgtypes.AnyArray,                                             // TODO: should be anycompatiblearray
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.AnyArray, pgtypes.AnyArray}, // TODO: should be anycompatiblearray, anycompatiblearray
	Strict:     false,
	Callable: func(ctx *sql.Context, paramsAndReturn [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		if val1 == nil && val2 == nil {
			return nil, nil
		} else if val1 == nil {
			return val2, nil
		} else if val2 == nil {
			return val1, nil
		}

		array1 := val1.([]any)
		array2 := val2.([]any)

		// Concatenate the arrays
		result := make([]any, len(array1)+len(array2))
		copy(result, array1)
		copy(result[len(array1):], array2)

		return result, nil
	},
}

// array_prepend represents the PostgreSQL function of the same name, taking the same parameters.
var array_prepend = framework.Function2{
	Name:       "array_prepend",
	Return:     pgtypes.AnyArray,                                               // TODO: should be anycompatiblearray
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.AnyElement, pgtypes.AnyArray}, // TODO: should be anycompatible, anycompatiblearray
	Strict:     false,
	Callable: func(ctx *sql.Context, paramsAndReturn [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		if val2 == nil {
			return []any{val1}, nil
		}
		return append([]any{val1}, val2.([]any)...), nil
	},
}

// byteacat_callable is the callable logic for the byteacat function.
func byteacat_callable(ctx *sql.Context, paramsAndReturn [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	v1 := val1.([]byte)
	v2 := val2.([]byte)
	copied := make([]byte, len(v1)+len(v2))
	copy(copied, v1)
	copy(copied[len(v1):], v2)
	return copied, nil
}

// byteacat represents the PostgreSQL function of the same name, taking the same parameters.
var byteacat = framework.Function2{
	Name:       "byteacat",
	Return:     pgtypes.Bytea,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Bytea, pgtypes.Bytea},
	Strict:     true,
	Callable:   byteacat_callable,
}

// jsonb_concat_callable is the callable logic for the jsonb_concat function.
func jsonb_concat_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1Interface any, val2Interface any) (any, error) {
	wrapper1, ok1 := val1Interface.(sql.JSONWrapper)
	wrapper2, ok2 := val2Interface.(sql.JSONWrapper)
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("jsonb_concat: unexpected types %T, %T", val1Interface, val2Interface)
	}
	v1, err := wrapper1.ToInterface(ctx)
	if err != nil {
		return nil, err
	}
	v2, err := wrapper2.ToInterface(ctx)
	if err != nil {
		return nil, err
	}
	// Merge objects if both are objects
	obj1, isObj1 := v1.(map[string]interface{})
	obj2, isObj2 := v2.(map[string]interface{})
	if isObj1 && isObj2 {
		newObj := make(map[string]interface{}, len(obj1)+len(obj2))
		for k, v := range obj1 {
			newObj[k] = v
		}
		for k, v := range obj2 {
			newObj[k] = v
		}
		return gmstypes.JSONDocument{Val: newObj}, nil
	}
	// Not both objects: wrap non-arrays in single-element arrays and concatenate
	arr1, isArr1 := v1.([]interface{})
	arr2, isArr2 := v2.([]interface{})
	if !isArr1 {
		arr1 = []interface{}{v1}
	}
	if !isArr2 {
		arr2 = []interface{}{v2}
	}
	newArray := make([]interface{}, len(arr1)+len(arr2))
	copy(newArray, arr1)
	copy(newArray[len(arr1):], arr2)
	return gmstypes.JSONDocument{Val: newArray}, nil
}

// jsonb_concat represents the PostgreSQL function of the same name, taking the same parameters.
var jsonb_concat = framework.Function2{
	Name:       "jsonb_concat",
	Return:     pgtypes.JsonB,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.JsonB, pgtypes.JsonB},
	Strict:     true,
	Callable:   jsonb_concat_callable,
}

// textanycat_callable is the callable logic for the textanycat function.
func textanycat_callable(ctx *sql.Context, paramsAndReturn [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	valType := paramsAndReturn[1]
	val2String, err := valType.IoOutput(ctx, val2)
	if err != nil {
		return nil, err
	}
	return val1.(string) + val2String, nil
}

// textanycat represents the PostgreSQL function of the same name, taking the same parameters.
var textanycat = framework.Function2{
	Name:       "textanycat",
	Return:     pgtypes.Text,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.AnyNonArray},
	Strict:     true,
	Callable:   textanycat_callable,
}

// textcat_callable is the callable logic for the textcat function.
func textcat_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	return val1.(string) + val2.(string), nil
}

// textcat represents the PostgreSQL function of the same name, taking the same parameters.
var textcat = framework.Function2{
	Name:       "textcat",
	Return:     pgtypes.Text,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Text},
	Strict:     true,
	Callable:   textcat_callable,
}
