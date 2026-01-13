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
	"strings"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/extensions"
	"github.com/dolthub/doltgresql/core/id"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

const anonymousCompositePrefix = "table("
const anonymousCompositeSuffix = ")"

// FunctionProvider is the special sql.FunctionProvider for Doltgres that allows us to handle functions that
// are created by users.
type FunctionProvider struct{}

var _ sql.FunctionProvider = (*FunctionProvider)(nil)

// Function implements the interface sql.FunctionProvider.
func (fp *FunctionProvider) Function(ctx *sql.Context, name string) (sql.Function, bool) {
	// TODO: this should be configurable from within Dolt, rather than set on an external variable
	if !core.IsContextValid(ctx) {
		return nil, false
	}
	funcCollection, err := core.GetFunctionsCollectionFromContext(ctx)
	if err != nil {
		return nil, false
	}
	typesCollection, err := core.GetTypesCollectionFromContext(ctx)
	if err != nil {
		return nil, false
	}
	// TODO: this should search all schemas in the search path, but the search path doesn't handle pg_catalog yet
	funcName := id.NewFunction("pg_catalog", name)
	overloads, err := funcCollection.GetFunctionOverloads(ctx, funcName)
	if err != nil {
		return nil, false
	}
	if len(overloads) == 0 {
		currentSchema, err := core.GetCurrentSchema(ctx)
		if err != nil {
			return nil, false
		}
		funcName = id.NewFunction(currentSchema, name)
		overloads, err = funcCollection.GetFunctionOverloads(ctx, funcName)
		if err != nil {
			return nil, false
		}
		if len(overloads) == 0 {
			return nil, false
		}
	}

	overloadTree := NewOverloads()
	for _, overload := range overloads {
		var returnType *pgtypes.DoltgresType
		if isAnonymousCompositeType(overload.ReturnType) {
			// If this is an anonymous composite type, then we can't load it
			// from typesCollection, so we create it dynamically when needed.
			returnType = createAnonymousCompositeType(ctx, overload.ReturnType)
		} else {
			returnType, err = typesCollection.GetType(ctx, overload.ReturnType)
			if err != nil || returnType == nil {
				return nil, false
			}
		}

		paramTypes := make([]*pgtypes.DoltgresType, len(overload.ParameterTypes))
		for i, paramType := range overload.ParameterTypes {
			paramTypes[i], err = typesCollection.GetType(ctx, paramType)
			if err != nil || paramTypes[i] == nil {
				return nil, false
			}
		}
		if len(overload.ExtensionName) > 0 {
			if err = overloadTree.Add(CFunction{
				ID:                 overload.ID,
				ReturnType:         returnType,
				ParameterTypes:     paramTypes,
				Variadic:           overload.Variadic,
				IsNonDeterministic: overload.IsNonDeterministic,
				Strict:             overload.Strict,
				ExtensionName:      extensions.LibraryIdentifier(overload.ExtensionName),
				ExtensionSymbol:    overload.ExtensionSymbol,
			}); err != nil {
				return nil, false
			}
		} else if len(overload.SQLDefinition) > 0 {
			if err = overloadTree.Add(SQLFunction{
				ID:                 overload.ID,
				ReturnType:         returnType,
				ParameterNames:     overload.ParameterNames,
				ParameterTypes:     paramTypes,
				Variadic:           overload.Variadic,
				IsNonDeterministic: overload.IsNonDeterministic,
				Strict:             overload.Strict,
				SqlStatement:       overload.SQLDefinition,
				SetOf:              overload.SetOf,
			}); err != nil {
				return nil, false
			}
		} else {
			if err = overloadTree.Add(InterpretedFunction{
				ID:                 overload.ID,
				ReturnType:         returnType,
				ParameterNames:     overload.ParameterNames,
				ParameterTypes:     paramTypes,
				Variadic:           overload.Variadic,
				IsNonDeterministic: overload.IsNonDeterministic,
				Strict:             overload.Strict,
				Statements:         overload.Operations,
			}); err != nil {
				return nil, false
			}
		}
	}
	return sql.FunctionN{
		Name: name,
		Fn: func(params ...sql.Expression) (sql.Expression, error) {
			return NewCompiledFunction(name, params, overloadTree, false), nil
		},
	}, true
}

// isAnonymousCompositeType return true if |returnType| represents an anonymous composite return type
// for a function (i.e. the function was declared as "RETURNS TABLE(...)").
func isAnonymousCompositeType(returnType id.Type) bool {
	typeName := returnType.TypeName()
	return strings.HasPrefix(typeName, anonymousCompositePrefix) &&
		strings.HasSuffix(typeName, anonymousCompositeSuffix)
}

// createAnonymousCompositeType creates a new DoltgresType for the anonymous composite return type for a function,
// as represented by |returnType|.
func createAnonymousCompositeType(ctx *sql.Context, returnType id.Type) *pgtypes.DoltgresType {
	typeName := returnType.TypeName()
	attributeTypes := typeName[len(anonymousCompositePrefix) : len(typeName)-len(anonymousCompositeSuffix)]
	attributeTypesSlice := strings.Split(attributeTypes, ",")

	attrs := make([]pgtypes.CompositeAttribute, len(attributeTypesSlice))
	for i, attributeNameAndType := range attributeTypesSlice {
		split := strings.Split(attributeNameAndType, ":")
		if len(split) != 2 {
			// TODO: We could return an error here, but the only place this function is
			//       called (FunctionProvider.Function) would require updating the
			//       sql.FunctionProvider interface in GMS, too.
			panic("unexpected anonymous composite type attribute syntax: " + attributeNameAndType)
		}

		typeId := id.NewType("", split[1])
		attrs[i] = pgtypes.NewCompositeAttribute(nil, id.Null, split[0], typeId, int16(i), "")
	}
	return pgtypes.NewCompositeType(ctx, id.Null, id.NullType, returnType, attrs)
}
