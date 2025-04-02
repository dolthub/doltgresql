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

// LoadFunctionFromCatalog returns the function matching the given name and parameter types. This is intended solely for
// functions that are used for types, as the returned functions are not valid using the Eval function.
var LoadFunctionFromCatalog func(funcName string, parameterTypes []*DoltgresType) any

// functionRegistry is a local registry that holds a mapping from ID to QuickFunction. This is done as types are now
// passed by struct, meaning that we need to cache the loading of functions somewhere. In addition, we don't yet support
// deleting built-in functions, so we can make a global cache. This makes a hard assumption that all functions being
// referenced actually exist, which should be true until built-in function deletion is implemented.
//
// In a way, one can view this as associated an OID to a function. With a proper OID system, this would not need to
// exist. It should be removed once OIDs are figured out.
type functionRegistry struct {
	mutex      *sync.Mutex
	counter    uint32
	mapping    map[id.Function]uint32
	revMapping map[uint32]id.Function
	functions  [256]QuickFunction // Arbitrary number, big enough for now to fit every function in it
}

// globalFunctionRegistry is the global functionRegistry. Only one needs to exist since we do not yet allow deleting
// built-in functions.
var globalFunctionRegistry = functionRegistry{
	mutex:      &sync.Mutex{},
	counter:    1,
	mapping:    map[id.Function]uint32{id.NullFunction: 0},
	revMapping: map[uint32]id.Function{0: id.NullFunction},
}

// InternalToRegistryID returns an ID for the given Internal ID.
func (r *functionRegistry) InternalToRegistryID(functionID id.Function) uint32 {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if registryID, ok := r.mapping[functionID]; ok {
		return registryID
	}
	if r.counter >= uint32(len(r.functions)) {
		panic("max function count reached in static array")
	}
	r.mapping[functionID] = r.counter
	r.revMapping[r.counter] = functionID
	r.counter++
	return r.counter - 1
}

// GetFunction returns the associated function for the given ID. This will always return a valid function.
func (r *functionRegistry) GetFunction(id uint32) QuickFunction {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	f := r.functions[id]
	if f != nil {
		return f
	}
	if id == 0 {
		return nil
	}
	f = r.loadFunction(id)
	if f == nil {
		// If we hit this panic, then we're missing a test that uses this function (and we should add that test)
		panic(errors.Errorf("cannot find function: `%s`", r.revMapping[id]))
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
func (r *functionRegistry) loadFunction(id uint32) QuickFunction {
	// We make this check a second time (first in GetFunction) since the function may have been added while another
	// function acquired the lock.
	f := r.functions[id]
	if f != nil {
		return f
	}
	if LoadFunctionFromCatalog == nil {
		return nil
	}
	functionID := r.revMapping[id]
	if !functionID.IsValid() {
		return nil
	}
	funcName, types := r.toFuncSignature(functionID)
	potentialFunction := LoadFunctionFromCatalog(funcName, types)
	if potentialFunction == nil {
		return nil
	}
	f = potentialFunction.(QuickFunction)
	r.functions[id] = f
	return f
}

// nameWithoutParams returns the name only from the given function string.
func (*functionRegistry) nameWithoutParams(functionID id.Function) string {
	if !functionID.IsValid() {
		return "-"
	}
	return functionID.FunctionName()
}

// toFuncSignature returns a function signature for the given Internal ID.
func (*functionRegistry) toFuncSignature(functionID id.Function) (string, []*DoltgresType) {
	internalParams := functionID.Parameters()
	params := make([]*DoltgresType, len(internalParams))
	for i, internalParam := range internalParams {
		params[i] = IDToBuiltInDoltgresType[internalParam]
	}
	return functionID.FunctionName(), params
}

// toFuncID creates a valid function string for the given name and parameters, then registers the name with the
// global functionRegistry. The ID from the registry is returned.
func toFuncID(functionName string, params ...id.Type) uint32 {
	if functionName == "-" || len(functionName) == 0 {
		return 0
	}
	functionID := id.NewFunction("pg_catalog", functionName, params...)
	return globalFunctionRegistry.InternalToRegistryID(functionID)
}
