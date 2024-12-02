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
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/lib/pq/oid"
)

// functionNameSplitter is used to delineate the parameter OIDs in a function string.
const functionNameSplitter = ";"

// QuickFunction is an interface redefinition of the one defined in the `server/functions/framework` package to avoid cycles.
type QuickFunction interface {
	CallVariadic(ctx *sql.Context, args ...any) (interface{}, error)
	ResolvedTypes() []DoltgresType
	WithResolvedTypes(newTypes []DoltgresType) any
}

// LoadFunctionFromCatalog returns the function matching the given name and parameter types. This is intended solely for
// functions that are used for types, as the returned functions are not valid using the Eval function.
var LoadFunctionFromCatalog func(funcName string, parameterTypes []DoltgresType) any

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
	mapping    map[string]uint32
	revMapping map[uint32]string
	functions  [256]QuickFunction // Arbitrary number, big enough for now to fit every function in it
}

// globalFunctionRegistry is the global functionRegistry. Only one needs to exist since we do not yet allow deleting
// built-in functions.
var globalFunctionRegistry = functionRegistry{
	mutex:      &sync.Mutex{},
	counter:    1,
	mapping:    map[string]uint32{"-": 0},
	revMapping: map[uint32]string{0: "-"},
}

// StringToID returns an ID for the given function string.
func (r *functionRegistry) StringToID(functionString string) uint32 {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if id, ok := r.mapping[functionString]; ok {
		return id
	}
	if r.counter >= uint32(len(r.functions)) {
		panic("max function count reached in static array")
	}
	r.mapping[functionString] = r.counter
	r.revMapping[r.counter] = functionString
	r.counter++
	return r.counter - 1
}

// GetFunction returns the associated function for the given ID. This will always return a valid function.
func (r *functionRegistry) GetFunction(id uint32) QuickFunction {
	f := r.functions[id]
	if f != nil {
		return f
	}
	if id == 0 {
		return nil
	}
	r.mutex.Lock()
	defer r.mutex.Unlock()
	f = r.loadFunction(id)
	if f == nil {
		// If we hit this panic, then we're missing a test that uses this function (and we should add that test)
		panic(fmt.Errorf("cannot find function: `%s`", r.revMapping[id]))
	}
	return f
}

// GetFullString returns the function string associated with the given ID.
func (r *functionRegistry) GetFullString(id uint32) string {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.revMapping[id]
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
	functionString := r.revMapping[id]
	name, params, ok := r.nameWithParams(functionString)
	if !ok {
		return nil
	}
	potentialFunction := LoadFunctionFromCatalog(name, params)
	if potentialFunction == nil {
		return nil
	}
	f = potentialFunction.(QuickFunction)
	r.functions[id] = f
	return f
}

// nameWithoutParams returns the name only from the given function string.
func (*functionRegistry) nameWithoutParams(funcString string) string {
	return strings.Split(funcString, functionNameSplitter)[0]
}

// nameWithParams returns the name and parameter types from the given function string. Return false if there were issues
// with the OID parameters.
func (*functionRegistry) nameWithParams(funcString string) (string, []DoltgresType, bool) {
	parts := strings.Split(funcString, functionNameSplitter)
	if len(parts) == 1 {
		return parts[0], nil, true
	}
	dTypes := make([]DoltgresType, len(parts)-1)
	for i := 1; i < len(parts); i++ {
		oidVal, err := strconv.Atoi(parts[i])
		if err != nil {
			return parts[0], nil, false
		}
		typ, ok := OidToBuiltInDoltgresType[uint32(oidVal)]
		if !ok {
			return parts[0], nil, false
		}
		dTypes[i-1] = typ
	}
	return parts[0], dTypes, true
}

// toFuncID creates a valid function string for the given name and parameters, then registers the name with the
// global functionRegistry. The ID from the registry is returned.
func toFuncID(functionName string, params ...oid.Oid) uint32 {
	if functionName == "-" {
		return 0
	}
	if len(params) > 0 {
		paramStrs := make([]string, len(params))
		for i := range params {
			paramStrs[i] = strconv.Itoa(int(params[i]))
		}
		functionName = fmt.Sprintf("%s%s%s", functionName, functionNameSplitter, strings.Join(paramStrs, functionNameSplitter))
	}
	return globalFunctionRegistry.StringToID(functionName)
}
