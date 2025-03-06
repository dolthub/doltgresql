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
	"maps"
	"sync"

	"github.com/cockroachdb/errors"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/interpreter"
)

// Collection contains a collection of functions.
type Collection struct {
	funcMap     map[id.Function]*Function
	overloadMap map[id.Function][]*Function
	mutex       *sync.Mutex
}

// Function represents a created function.
type Function struct {
	ID                 id.Function
	ReturnType         id.Type
	ParameterNames     []string
	ParameterTypes     []id.Type
	Variadic           bool
	IsNonDeterministic bool
	Strict             bool
	Operations         []interpreter.InterpreterOperation
}

// GetFunction returns the function with the given ID. Returns nil if the function cannot be found.
func (pgf *Collection) GetFunction(funcID id.Function) *Function {
	pgf.mutex.Lock()
	defer pgf.mutex.Unlock()

	if f, ok := pgf.funcMap[funcID]; ok {
		return f
	}
	return nil
}

// GetFunctionOverloads returns the overloads for the function matching the schema and the function name. The parameter
// types are ignored when searching for overloads.
func (pgf *Collection) GetFunctionOverloads(funcID id.Function) []*Function {
	pgf.mutex.Lock()
	defer pgf.mutex.Unlock()

	funcNameOnly := id.NewFunction(funcID.SchemaName(), funcID.FunctionName())
	return pgf.overloadMap[funcNameOnly]
}

// HasFunction returns whether the function is present.
func (pgf *Collection) HasFunction(funcID id.Function) bool {
	return pgf.GetFunction(funcID) != nil
}

// AddFunction adds a new function.
func (pgf *Collection) AddFunction(f *Function) error {
	pgf.mutex.Lock()
	defer pgf.mutex.Unlock()

	if _, ok := pgf.funcMap[f.ID]; ok {
		return errors.Errorf(`function "%s" already exists with same argument types`, f.ID.FunctionName())
	}
	pgf.funcMap[f.ID] = f
	funcNameOnly := id.NewFunction(f.ID.SchemaName(), f.ID.FunctionName())
	pgf.overloadMap[funcNameOnly] = append(pgf.overloadMap[funcNameOnly], f)
	return nil
}

// DropFunction drops an existing function.
func (pgf *Collection) DropFunction(funcID id.Function) error {
	pgf.mutex.Lock()
	defer pgf.mutex.Unlock()

	if _, ok := pgf.funcMap[funcID]; ok {
		delete(pgf.funcMap, funcID)
		funcNameOnly := id.NewFunction(funcID.SchemaName(), funcID.FunctionName())
		for i, f := range pgf.overloadMap[funcNameOnly] {
			if f.ID == funcID {
				pgf.overloadMap[funcNameOnly] = append(pgf.overloadMap[funcNameOnly][:i], pgf.overloadMap[funcNameOnly][i+1:]...)
				break
			}
		}
		return nil
	}
	return errors.Errorf(`function %s does not exist`, funcID.FunctionName())
}

// IterateFunctions iterates over all functions in the collection.
func (pgf *Collection) IterateFunctions(callback func(f *Function) error) error {
	pgf.mutex.Lock()
	defer pgf.mutex.Unlock()

	for _, f := range pgf.funcMap {
		if err := callback(f); err != nil {
			return err
		}
	}
	return nil
}

// Clone returns a new *Collection with the same contents as the original.
func (pgf *Collection) Clone() *Collection {
	pgf.mutex.Lock()
	defer pgf.mutex.Unlock()

	return &Collection{
		funcMap:     maps.Clone(pgf.funcMap),
		overloadMap: maps.Clone(pgf.overloadMap),
		mutex:       &sync.Mutex{},
	}
}
