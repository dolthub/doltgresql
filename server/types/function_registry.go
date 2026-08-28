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

package types

import (
	"sync"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"
)

// QuickFunction is an interface redefinition of the one defined in the `server/functions/framework` package to avoid cycles.
type QuickFunction interface {
	CallVariadic(ctx *sql.Context, args ...any) (interface{}, error)
	ResolvedTypes() []*DoltgresType
	WithResolvedTypes(newTypes []*DoltgresType) any
}

// LoadFunctionFromCatalog returns the function matching the given schema, name and parameter types. This is intended
// solely for functions that are used for types, as the returned functions are not valid using the Eval function.
var LoadFunctionFromCatalog func(ctx *sql.Context, schemaName string, funcName string, parameterTypes []*DoltgresType) any

// LoadExtensionFunction returns the extension-provided function matching the given ID. This is the fallback for
// LoadFunctionFromCatalog in contexts that have no session, such as index comparators.
var LoadExtensionFunction func(functionID id.Function) any

// functionRegistry is a local registry that holds a mapping from ID to QuickFunction. This is done as types are now
// passed by struct, meaning that we need to cache the loading of functions somewhere. Only the functions in pg_catalog
// are cached, since a user-defined function may be replaced or dropped, and it may differ between databases.
//
// In a way, one can view this as associated an OID to a function. With a proper OID system, this would not need to
// exist. It should be removed once OIDs are figured out.
type functionRegistry struct {
	mutex      *sync.Mutex
	counter    uint32
	mapping    map[id.Function]uint32
	revMapping map[uint32]id.Function
	functions  []QuickFunction
}

// globalFunctionRegistry is the global functionRegistry. Only one needs to exist since we do not yet allow deleting
// built-in functions.
var globalFunctionRegistry = functionRegistry{
	mutex:      &sync.Mutex{},
	counter:    1,
	mapping:    map[id.Function]uint32{id.NullFunction: 0},
	revMapping: map[uint32]id.Function{0: id.NullFunction},
	functions:  make([]QuickFunction, 1, 256),
}

// InternalToRegistryID returns an ID for the given Internal ID.
func (r *functionRegistry) InternalToRegistryID(functionID id.Function) uint32 {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if registryID, ok := r.mapping[functionID]; ok {
		return registryID
	}
	r.mapping[functionID] = r.counter
	r.revMapping[r.counter] = functionID
	r.functions = append(r.functions, nil)
	r.counter++
	return r.counter - 1
}

// GetFunction returns the associated function for the given ID. This will always return a valid function.
func (r *functionRegistry) GetFunction(ctx *sql.Context, id uint32) QuickFunction {
	if id == 0 {
		return nil
	}
	f := r.loadFunction(ctx, id)
	if f == nil {
		// If we hit this panic, then we're missing a test that uses this function (and we should add that test)
		panic(errors.Errorf("cannot find function: `%s`", r.GetInternalID(id)))
	}
	return f
}

// GetInternalID returns the function's Internal ID associated with the given registry ID.
func (r *functionRegistry) GetInternalID(registryID uint32) id.Function {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.revMapping[registryID]
}

// GetString returns the extracted function name from the function string associated with the given ID.
func (r *functionRegistry) GetString(id uint32) string {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.nameWithoutParams(r.revMapping[id])
}

// loadFunction loads the given function
func (r *functionRegistry) loadFunction(ctx *sql.Context, id uint32) QuickFunction {
	// The mutex is only held while accessing the cache, since loading through the catalog may re-enter the registry
	// (extension types resolve their I/O functions during deserialization).
	r.mutex.Lock()
	f := r.functions[id]
	functionID := r.revMapping[id]
	r.mutex.Unlock()
	if f != nil {
		return f
	}
	if !functionID.IsValid() {
		return nil
	}
	if LoadFunctionFromCatalog != nil {
		if funcName, types, ok := r.toFuncSignature(ctx, functionID); ok {
			if potentialFunction := LoadFunctionFromCatalog(ctx, functionID.SchemaName(), funcName, types); potentialFunction != nil {
				f = potentialFunction.(QuickFunction)
				if functionID.SchemaName() == "pg_catalog" {
					r.mutex.Lock()
					r.functions[id] = f
					r.mutex.Unlock()
				}
				return f
			}
		}
	}
	if LoadExtensionFunction != nil {
		if potentialFunction := LoadExtensionFunction(functionID); potentialFunction != nil {
			return potentialFunction.(QuickFunction)
		}
	}
	return nil
}

// nameWithoutParams returns the name only from the given function string.
func (*functionRegistry) nameWithoutParams(functionID id.Function) string {
	if !functionID.IsValid() {
		return "-"
	}
	return functionID.FunctionName()
}

// toFuncSignature returns a function signature for the given Internal ID. Returns false when a parameter names a type
// that cannot be resolved, which may happen when a user-defined type has been dropped.
func (*functionRegistry) toFuncSignature(ctx *sql.Context, functionID id.Function) (string, []*DoltgresType, bool) {
	internalParams := functionID.Parameters()
	params := make([]*DoltgresType, len(internalParams))
	var collection TypeCollection
	for i, internalParam := range internalParams {
		if builtIn, ok := IDToBuiltInDoltgresType[internalParam]; ok {
			params[i] = builtIn
			continue
		}
		if collection == nil {
			if GetTypesCollectionFromContext == nil {
				return "", nil, false
			}
			var err error
			if collection, err = GetTypesCollectionFromContext(ctx, ""); err != nil {
				return "", nil, false
			}
		}
		param, err := collection.GetType(ctx, internalParam)
		if err != nil || param == nil {
			return "", nil, false
		}
		params[i] = param
	}
	return functionID.FunctionName(), params, true
}

// toFuncID creates a valid function string for the given name and parameters, then registers the name with the
// global functionRegistry. The ID from the registry is returned.
func toFuncID(functionName string, params ...id.Type) uint32 {
	if functionName == "-" || len(functionName) == 0 {
		return 0
	}
	return ToFuncID(id.NewFunction("pg_catalog", functionName, params...))
}

// ToFuncID registers the given function with the global function registry, and returns the ID it was given. Primarily
// used by extensions.
func ToFuncID(functionID id.Function) uint32 {
	return globalFunctionRegistry.InternalToRegistryID(functionID)
}

// FromFuncID creates a valid function string for the given name and parameters, then registers the name with the
// global functionRegistry. The ID from the registry is returned.
func FromFuncID(u uint32) id.Function {
	return globalFunctionRegistry.GetInternalID(u)
}
