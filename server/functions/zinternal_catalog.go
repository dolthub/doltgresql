// Copyright 2023 Dolthub, Inc.
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
	"reflect"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression/function"
)

// Function is a name, along with a collection of functions, that represent a single PostgreSQL function with all of its
// overloads.
type Function struct {
	Name      string
	Overloads []any
}

// Catalog contains all of the PostgreSQL functions. If a new function is added, make sure to add it to the catalog here.
var Catalog = []Function{
	abs,
	acos,
	acosd,
	acosh,
	ascii,
	asin,
	asind,
	asinh,
	atan,
	atan2,
	atan2d,
	atand,
	atanh,
	bit_length,
	btrim,
	cbrt,
	ceil,
	ceiling,
	char_length,
	character_length,
	chr,
	cos,
	cosd,
	cosh,
	cot,
	cotd,
	degrees,
	div,
	exp,
	factorial,
	floor,
	gcd,
	initcap,
	lcm,
	left,
	length,
	ln,
	log,
	log10,
	lower,
	lpad,
	ltrim,
	md5,
	min_scale,
	mod,
	octet_length,
	pi,
	power,
	radians,
	random,
	repeat,
	replace,
	reverse,
	right,
	round,
	rpad,
	rtrim,
	scale,
	sign,
	sin,
	sind,
	sinh,
	split_part,
	sqrt,
	strpos,
	substr,
	tan,
	tand,
	tanh,
	to_hex,
	trim_scale,
	trunc,
	upper,
	width_bucket,
}

// init handles the initialization of the catalog by overwriting the built-in GMS functions, since they do not apply to
// PostgreSQL (and functions of the same name often have different behavior).
func init() {
	catalogMap := make(map[string]struct{})
	for _, f := range Catalog {
		catalogMap[strings.ToLower(f.Name)] = struct{}{}
	}
	var newBuiltIns []sql.Function
	for _, f := range function.BuiltIns {
		if _, ok := catalogMap[strings.ToLower(f.FunctionName())]; !ok {
			newBuiltIns = append(newBuiltIns, f)
		}
	}
	function.BuiltIns = newBuiltIns

	allNames := make(map[string]struct{})
	for _, catalogItem := range Catalog {
		funcName := strings.ToLower(catalogItem.Name)
		if _, ok := allNames[funcName]; ok {
			panic("duplicate name: " + catalogItem.Name)
		}
		allNames[funcName] = struct{}{}

		baseOverload := &OverloadDeduction{}
		for _, functionOverload := range catalogItem.Overloads {
			// For each function overload, we first need to ensure that it has an acceptable signature
			funcVal := reflect.ValueOf(functionOverload)
			if !funcVal.IsValid() || funcVal.IsNil() {
				panic(fmt.Errorf("function `%s` has an invalid item", catalogItem.Name))
			}
			if funcVal.Kind() != reflect.Func {
				panic(fmt.Errorf("function `%s` has a non-function item", catalogItem.Name))
			}
			if funcVal.Type().NumOut() != 2 {
				panic(fmt.Errorf("function `%s` has an overload that does not return two values", catalogItem.Name))
			}
			if funcVal.Type().Out(1) != reflect.TypeOf((*error)(nil)).Elem() {
				panic(fmt.Errorf("function `%s` has an overload that does not return an error", catalogItem.Name))
			}
			returnValType, returnSqlType, ok := ParameterTypeFromReflection(funcVal.Type().Out(0))
			if !ok {
				panic(fmt.Errorf("function `%s` has an overload that returns as invalid type (`%s`)",
					catalogItem.Name, funcVal.Type().Out(0).String()))
			}

			// Loop through all of the parameters to ensure uniqueness, then store it
			currentOverload := baseOverload
			for i := 0; i < funcVal.Type().NumIn(); i++ {
				paramValType, _, ok := ParameterTypeFromReflection(funcVal.Type().In(i))
				if !ok {
					panic(fmt.Errorf("function `%s` has an overload with an invalid parameter type (`%s`)",
						catalogItem.Name, funcVal.Type().In(i).String()))
				}
				nextOverload := currentOverload.Parameter[paramValType]
				if nextOverload == nil {
					nextOverload = &OverloadDeduction{}
					currentOverload.Parameter[paramValType] = nextOverload
				}
				currentOverload = nextOverload
			}
			if currentOverload.Function.IsValid() && !currentOverload.Function.IsNil() {
				panic(fmt.Errorf("function `%s` has duplicate overloads", catalogItem.Name))
			}
			currentOverload.Function = funcVal
			currentOverload.ReturnValType = returnValType
			currentOverload.ReturnSqlType = returnSqlType
		}

		// Store the compiled function into the engine's built-in functions
		function.BuiltIns = append(function.BuiltIns, sql.FunctionN{
			Name: funcName,
			Fn: func(params ...sql.Expression) (sql.Expression, error) {
				return &CompiledFunction{
					Name:       catalogItem.Name,
					Parameters: params,
					Functions:  baseOverload,
				}, nil
			},
		})
	}
}
