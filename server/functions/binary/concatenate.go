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
	"sort"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// These functions can be gathered using the following query from a Postgres 15 instance:
// SELECT * FROM pg_operator o WHERE o.oprname = '||' ORDER BY o.oprcode::varchar;

// initBinaryConcatenate registers the functions to the catalog.
func initBinaryConcatenate() {
	framework.RegisterBinaryFunction(framework.Operator_BinaryConcatenate, anytextcat)
	framework.RegisterBinaryFunction(framework.Operator_BinaryConcatenate, byteacat)
	framework.RegisterBinaryFunction(framework.Operator_BinaryConcatenate, jsonb_concat)
	framework.RegisterBinaryFunction(framework.Operator_BinaryConcatenate, textanycat)
	framework.RegisterBinaryFunction(framework.Operator_BinaryConcatenate, textcat)
	// TODO: array_append, array_cat, array_prepend, bitcat, tsquery_or, tsvector_concat
}

// anytextcat represents the PostgreSQL function of the same name, taking the same parameters.
var anytextcat = framework.Function2{
	Name:       "anytextcat",
	Return:     pgtypes.Text,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.AnyNonArray, pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, paramsAndReturn [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		valType := paramsAndReturn[0]
		val1String, err := framework.IoOutput(ctx, valType, val1)
		if err != nil {
			return nil, err
		}
		return val1String + val2.(string), nil
	},
}

// byteacat represents the PostgreSQL function of the same name, taking the same parameters.
var byteacat = framework.Function2{
	Name:       "byteacat",
	Return:     pgtypes.Bytea,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Bytea, pgtypes.Bytea},
	Strict:     true,
	Callable: func(ctx *sql.Context, paramsAndReturn [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		v1 := val1.([]byte)
		v2 := val2.([]byte)
		copied := make([]byte, len(v1)+len(v2))
		copy(copied, v1)
		copy(copied[len(v1):], v2)
		return copied, nil
	},
}

// jsonb_concat represents the PostgreSQL function of the same name, taking the same parameters.
var jsonb_concat = framework.Function2{
	Name:       "jsonb_concat",
	Return:     pgtypes.JsonB,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.JsonB, pgtypes.JsonB},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1Interface any, val2Interface any) (any, error) {
		val1 := val1Interface.(pgtypes.JsonDocument).Value
		val2 := val2Interface.(pgtypes.JsonDocument).Value
		// First we'll merge objects if they're both objects
		val1Obj, isVal1Obj := val1.(pgtypes.JsonValueObject)
		val2Obj, isVal2Obj := val2.(pgtypes.JsonValueObject)
		if isVal1Obj && isVal2Obj {
			newObj := pgtypes.JsonValueCopy(val1Obj).(pgtypes.JsonValueObject)
			for _, item := range val2Obj.Items {
				if existingIdx, ok := newObj.Index[item.Key]; ok {
					newObj.Items[existingIdx].Value = pgtypes.JsonValueCopy(item.Value)
				} else {
					newObj.Items = append(newObj.Items, pgtypes.JsonValueObjectItem{
						Key:   item.Key,
						Value: pgtypes.JsonValueCopy(item.Value),
					})
				}
			}
			sort.Slice(newObj.Items, func(i, j int) bool {
				if len(newObj.Items[i].Key) < len(newObj.Items[j].Key) {
					return true
				} else if len(newObj.Items[i].Key) > len(newObj.Items[j].Key) {
					return false
				} else {
					return newObj.Items[i].Key < newObj.Items[j].Key
				}
			})
			for i, item := range newObj.Items {
				newObj.Index[item.Key] = i
			}
			return pgtypes.JsonDocument{Value: newObj}, nil
		}
		// They're not both objects, so we'll make them both arrays if they're not already arrays
		if _, ok := val1.(pgtypes.JsonValueArray); !ok {
			val1 = pgtypes.JsonValueArray{val1}
		}
		if _, ok := val2.(pgtypes.JsonValueArray); !ok {
			val2 = pgtypes.JsonValueArray{val2}
		}
		val1Array := pgtypes.JsonValueCopy(val1.(pgtypes.JsonValueArray)).(pgtypes.JsonValueArray)
		val2Array := pgtypes.JsonValueCopy(val2.(pgtypes.JsonValueArray)).(pgtypes.JsonValueArray)
		newArray := make(pgtypes.JsonValueArray, len(val1Array)+len(val2Array))
		copy(newArray, val1Array)
		copy(newArray[len(val1Array):], val2Array)
		return pgtypes.JsonDocument{Value: newArray}, nil
	},
}

// textanycat represents the PostgreSQL function of the same name, taking the same parameters.
var textanycat = framework.Function2{
	Name:       "textanycat",
	Return:     pgtypes.Text,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Text, pgtypes.AnyNonArray},
	Strict:     true,
	Callable: func(ctx *sql.Context, paramsAndReturn [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		valType := paramsAndReturn[1]
		val2String, err := framework.IoOutput(ctx, valType, val2)
		if err != nil {
			return nil, err
		}
		return val1.(string) + val2String, nil
	},
}

// textcat represents the PostgreSQL function of the same name, taking the same parameters.
var textcat = framework.Function2{
	Name:       "textcat",
	Return:     pgtypes.Text,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Text, pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		return val1.(string) + val2.(string), nil
	},
}
