// Copyright 2025 Dolthub, Inc.
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
	"slices"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/aggregates"
	"github.com/dolthub/doltgresql/core/functions"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/typecollection"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// FunctionProvider is the special sql.FunctionProvider for Doltgres that allows us to handle functions that
// are created by users.
type FunctionProvider struct{}

var _ sql.FunctionProvider = (*FunctionProvider)(nil)
var _ sql.AggregateCallResolver = (*FunctionProvider)(nil)

// overloads returns the overloads of the named function. `overloadTree` holds every overload, both functions and
// aggregates, `hasAggregate` is whether any of them is an aggregate, and `ok` is false when the name has no
// user-defined overloads, which leaves built-in functions to be found in the function registry.
//
// A qualified name only finds overloads in its own schema. An unqualified name finds the overloads in every schema on
// the search path, where pg_catalog, and so the built-in functions, come first unless the path places pg_catalog later.
// An overload hides any with the same parameter types in a schema later on the path.
func (fp *FunctionProvider) overloads(ctx *sql.Context, schema, name string) (overloadTree *Overloads, hasAggregate bool, ok bool) {
	// TODO: this should be configurable from within Dolt, rather than set on an external variable
	if !core.IsContextValid(ctx) {
		return nil, false, false
	}
	funcCollection, err := core.GetFunctionsCollectionFromContext(ctx, "")
	if err != nil {
		return nil, false, false
	}
	typesCollection, err := core.GetTypesCollectionFromContext(ctx, "")
	if err != nil {
		return nil, false, false
	}
	aggCollection, err := core.GetAggregatesCollectionFromContext(ctx, "")
	if err != nil {
		return nil, false, false
	}

	schemasToSearch := []string{schema}
	if schema == "" {
		schemasToSearch, err = core.SearchPath(ctx)
		if err != nil {
			return nil, false, false
		}
	}

	overloadTree = NewOverloads()
	hasUserOverloads := false
	for _, searchSchema := range schemasToSearch {
		funcName := id.NewFunction(searchSchema, name)
		overloads, err := funcCollection.GetFunctionOverloads(ctx, funcName)
		if err != nil {
			return nil, false, false
		}
		aggOverloads, err := aggCollection.GetAggregateOverloads(ctx, funcName)
		if err != nil {
			return nil, false, false
		}
		if len(overloads) > 0 || len(aggOverloads) > 0 {
			hasUserOverloads = true
		}
		if len(aggOverloads) > 0 {
			hasAggregate = true
		}

		if err = addUserOverloads(ctx, overloadTree, typesCollection, overloads, aggOverloads); err != nil {
			return nil, false, false
		}
		if searchSchema == "pg_catalog" {
			if err = addBuiltInOverloads(overloadTree, name); err != nil {
				return nil, false, false
			}
			if len(AggregateCatalog[name]) > 0 {
				hasAggregate = true
			}
		}
	}

	if !hasUserOverloads {
		return nil, false, false
	}
	return overloadTree, hasAggregate, true
}

// addUserOverloads adds the given user-defined function and aggregate overloads to the tree, skipping any signature
// that an overload earlier on the search path has already taken.
func addUserOverloads(
	ctx *sql.Context,
	overloadTree *Overloads,
	typesCollection *typecollection.TypeCollection,
	overloads []functions.Function,
	aggOverloads []aggregates.Aggregate,
) error {
	for _, overload := range overloads {
		returnType, err := getType(ctx, typesCollection, overload.ReturnType)
		if err != nil {
			return err
		}

		paramTypes := make([]*pgtypes.DoltgresType, len(overload.AllParams))
		for i, param := range overload.AllParams {
			paramTypes[i], err = getType(ctx, typesCollection, param.Type)
			if err != nil {
				return err
			}
		}

		var function FunctionInterface
		if len(overload.ExtensionName) > 0 {
			function = ExtensionFunction{
				ID:                 overload.ID,
				ReturnType:         returnType,
				ParameterTypes:     paramTypes,
				Variadic:           overload.Variadic,
				IsNonDeterministic: overload.IsNonDeterministic,
				Strict:             overload.Strict,
				SetOf:              overload.SetOf,
				ExtensionName:      overload.ExtensionName,
				ExtensionSymbol:    overload.ExtensionSymbol,
			}
		} else if len(overload.SQLDefinition) > 0 {
			function = SQLFunction{
				ID:                 overload.ID,
				ReturnType:         returnType,
				AllParams:          overload.AllParams,
				AllTypes:           paramTypes,
				Variadic:           overload.Variadic,
				IsNonDeterministic: overload.IsNonDeterministic,
				Strict:             overload.Strict,
				SqlStatement:       overload.SQLDefinition,
				SetOf:              overload.SetOf,
			}
		} else {
			function = InterpretedFunction{
				ID:                 overload.ID,
				ReturnType:         returnType,
				AllParams:          overload.AllParams,
				AllTypes:           paramTypes,
				Variadic:           overload.Variadic,
				IsNonDeterministic: overload.IsNonDeterministic,
				Strict:             overload.Strict,
				SRF:                overload.SetOf,
				Statements:         overload.Operations,
			}
		}
		if err = addUnlessHidden(overloadTree, function); err != nil {
			return err
		}
	}

	for _, aggOverload := range aggOverloads {
		stateType, err := getType(ctx, typesCollection, aggOverload.SType)
		if err != nil {
			return err
		}
		returnType, err := getType(ctx, typesCollection, aggOverload.ReturnType)
		if err != nil {
			return err
		}
		paramTypes := make([]*pgtypes.DoltgresType, aggOverload.ID.ParameterCount())
		for i, param := range aggOverload.ID.Parameters() {
			paramTypes[i], err = getType(ctx, typesCollection, param)
			if err != nil {
				return err
			}
		}
		if err = addUnlessHidden(overloadTree, UserAggregate{
			ID:             aggOverload.ID,
			ReturnType:     returnType,
			ParameterTypes: paramTypes,
			StateType:      stateType,
			SFunc:          aggOverload.SFunc,
			FinalFunc:      aggOverload.FinalFunc,
			InitCond:       aggOverload.InitCond,
			HasInitCond:    aggOverload.HasInitCond,
		}); err != nil {
			return err
		}
	}
	return nil
}

// getType returns the type with the given ID, which must exist.
func getType(ctx *sql.Context, typesCollection *typecollection.TypeCollection, typeID id.Type) (*pgtypes.DoltgresType, error) {
	typ, err := typesCollection.GetType(ctx, typeID)
	if err != nil {
		return nil, err
	}
	if typ == nil {
		return nil, pgtypes.ErrTypeDoesNotExist.New(typeID.TypeName())
	}
	return typ, nil
}

// addUnlessHidden adds the function to the tree, unless an overload with the same parameter types, which comes earlier
// on the search path, hides it.
func addUnlessHidden(overloadTree *Overloads, function FunctionInterface) error {
	if _, ok := overloadTree.ByParamType[keyForParamTypes(function.GetInputParameterTypes())]; ok {
		return nil
	}
	return overloadTree.Add(function)
}

// CallsAggregate implements the interface sql.AggregateCallResolver. A call that resolves to one of the name's function
// overloads calls that function instead of the aggregate, and the arguments are only built when the name has both
// kinds of overload.
func (fp *FunctionProvider) CallsAggregate(ctx *sql.Context, schema, name string, args func() []sql.Expression) (bool, error) {
	overloadTree, _, ok := fp.overloads(ctx, schema, name)
	if !ok || !slices.ContainsFunc(overloadTree.AllOverloads, func(overload FunctionInterface) bool {
		_, isAggregate := overload.(AggregateFunctionInterface)
		return !isAggregate
	}) {
		return true, nil
	}
	aggregate := NewCompiledAggregateFunction(ctx, name, args(), overloadTree)
	if !aggregate.overload.Valid() {
		return true, nil
	}
	_, isAggregate := aggregate.overload.Function().(AggregateFunctionInterface)
	return isAggregate, nil
}

// Function implements the interface sql.FunctionProvider.
func (fp *FunctionProvider) Function(ctx *sql.Context, schema, name string) (sql.Function, bool) {
	overloadTree, hasAggregate, ok := fp.overloads(ctx, schema, name)
	if !ok {
		return nil, false
	}
	if hasAggregate {
		return sql.FunctionN{
			Name: name,
			Fn: func(ctx *sql.Context, params ...sql.Expression) (sql.Expression, error) {
				aggregate := NewCompiledAggregateFunction(ctx, name, params, overloadTree)
				if aggregate.overload.Valid() {
					if _, ok := aggregate.overload.Function().(AggregateFunctionInterface); !ok {
						return NewCompiledFunction(ctx, name, params, overloadTree, false), nil
					}
				}
				return aggregate, nil
			},
		}, true
	}
	return sql.FunctionN{
		Name: name,
		Fn: func(ctx *sql.Context, params ...sql.Expression) (sql.Expression, error) {
			return NewCompiledFunction(ctx, name, params, overloadTree, false), nil
		},
	}, true
}

// addBuiltInOverloads adds the built-in overloads of the given name to the tree, skipping any signature that an
// overload earlier on the search path has already taken.
func addBuiltInOverloads(overloadTree *Overloads, name string) error {
	for _, builtIn := range Catalog[name] {
		if err := addUnlessHidden(overloadTree, builtIn); err != nil {
			return err
		}
	}
	for _, builtIn := range AggregateCatalog[name] {
		if err := addUnlessHidden(overloadTree, builtIn); err != nil {
			return err
		}
	}
	return nil
}
