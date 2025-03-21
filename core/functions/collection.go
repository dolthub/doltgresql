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
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/store/hash"
	"github.com/dolthub/dolt/go/store/prolly"
	"github.com/dolthub/dolt/go/store/prolly/tree"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/rootobject"
	"github.com/dolthub/doltgresql/server/plpgsql"
)

type functionNamespace string

const (
	functionNamespace_FunctionMap functionNamespace = "1"
	functionNamespace_OverloadMap functionNamespace = "2"
)

// Collection contains a collection of functions.
type Collection struct {
	underlyingMap prolly.AddressMap
	ns            tree.NodeStore
	mutex         *sync.Mutex
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
	Operations         []plpgsql.InterpreterOperation
}

var _ rootobject.Collection = (*Collection)(nil)
var _ rootobject.RootObject = Function{}

// GetFunction returns the function with the given ID. Returns a function with an invalid ID if it cannot be found.
func (pgf *Collection) GetFunction(ctx context.Context, funcID id.Function) (Function, error) {
	pgf.mutex.Lock()
	defer pgf.mutex.Unlock()

	functionMap, ok, err := pgf.getFunctionMap(ctx)
	if err != nil || !ok {
		return Function{}, err
	}
	h, err := functionMap.Get(ctx, string(funcID))
	if err != nil || h.IsEmpty() {
		return Function{}, err
	}
	data, err := pgf.ns.ReadBytes(ctx, h)
	if err != nil {
		return Function{}, err
	}
	return DeserializeFunction(ctx, data)
}

// GetFunctionOverloads returns the overloads for the function matching the schema and the function name. The parameter
// types are ignored when searching for overloads.
func (pgf *Collection) GetFunctionOverloads(ctx context.Context, funcID id.Function) ([]Function, error) {
	pgf.mutex.Lock()
	defer pgf.mutex.Unlock()

	functionMap, ok, err := pgf.getFunctionMap(ctx)
	if err != nil || !ok {
		return nil, err
	}
	overloadIDs, err := pgf.getOverloadIDs(ctx, funcID)
	if err != nil {
		return nil, err
	}
	funcs := make([]Function, len(overloadIDs))
	for i := range overloadIDs {
		h, err := functionMap.Get(ctx, string(overloadIDs[i]))
		if err != nil {
			return nil, err
		}
		if h.IsEmpty() {
			return nil, fmt.Errorf("function %s is listed as an overload but cannot be found", overloadIDs[i].FunctionName())
		}
		data, err := pgf.ns.ReadBytes(ctx, h)
		if err != nil {
			return nil, err
		}
		funcs[i], err = DeserializeFunction(ctx, data)
		if err != nil {
			return nil, err
		}
	}
	return funcs, nil
}

// HasFunction returns whether the function is present.
func (pgf *Collection) HasFunction(ctx context.Context, funcID id.Function) bool {
	pgf.mutex.Lock()
	defer pgf.mutex.Unlock()

	functionMap, ok, err := pgf.getFunctionMap(ctx)
	if err != nil || !ok {
		return false
	}
	ok, err = functionMap.Has(ctx, string(funcID))
	return err == nil && ok
}

// AddFunction adds a new function.
func (pgf *Collection) AddFunction(ctx context.Context, f Function) error {
	pgf.mutex.Lock()
	defer pgf.mutex.Unlock()

	// First we'll check the regular function map to see if it exists
	functionMap, ok, err := pgf.getFunctionMap(ctx)
	if err != nil {
		return err
	}
	if !ok {
		functionMap, err = prolly.NewEmptyAddressMap(pgf.ns)
		if err != nil {
			return err
		}
	}
	ok, err = functionMap.Has(ctx, string(f.ID))
	if err != nil {
		return err
	}
	if ok {
		return errors.Errorf(`function "%s" already exists with same argument types`, f.ID.FunctionName())
	}

	// Now we'll add the function to our first map
	functionMap, err = func(ctx context.Context, functionMap prolly.AddressMap, f Function) (_ prolly.AddressMap, err error) {
		data, err := f.Serialize(ctx)
		if err != nil {
			return prolly.AddressMap{}, err
		}
		h, err := pgf.ns.WriteBytes(ctx, data)
		if err != nil {
			return prolly.AddressMap{}, err
		}
		mapEditor := functionMap.Editor()
		if err = mapEditor.Add(ctx, string(f.ID), h); err != nil {
			return prolly.AddressMap{}, err
		}
		return mapEditor.Flush(ctx)
	}(ctx, functionMap, f)
	if err != nil {
		return err
	}

	// Then we'll add to the overload map (and set)
	overloadMap, ok, err := pgf.getOverloadMap(ctx)
	if err != nil {
		return err
	}
	if !ok {
		overloadMap, err = prolly.NewEmptyAddressMap(pgf.ns)
		if err != nil {
			return err
		}
	}
	partialID := id.NewFunction(f.ID.SchemaName(), f.ID.FunctionName())
	overloadSetHash, err := overloadMap.Get(ctx, string(partialID))
	if err != nil {
		return err
	}
	var overloadSet prolly.AddressMap
	if overloadSetHash.IsEmpty() {
		overloadSet, err = prolly.NewEmptyAddressMap(pgf.ns)
		if err != nil {
			return err
		}
	} else {
		node, err := pgf.ns.Read(ctx, overloadSetHash)
		if err != nil {
			return err
		}
		overloadSet, err = prolly.NewAddressMap(node, pgf.ns)
		if err != nil {
			return err
		}
	}
	// Although we use this like a set, the AddressMap enforces that the value is non-empty, so we write something to it
	byteHash, err := pgf.ns.WriteBytes(ctx, []byte{42})
	if err != nil {
		return err
	}
	setEditor := overloadSet.Editor()
	if err = setEditor.Add(ctx, string(f.ID), byteHash); err != nil {
		return err
	}
	overloadSet, err = setEditor.Flush(ctx)
	if err != nil {
		return err
	}
	overloadMap, err = func(ctx context.Context, overloadMap prolly.AddressMap, partialID id.Function, overloadSet prolly.AddressMap) (_ prolly.AddressMap, err error) {
		mapEditor := overloadMap.Editor()
		overloadSetHash, err := pgf.ns.Write(ctx, overloadSet.Node())
		if err != nil {
			return prolly.AddressMap{}, err
		}
		if err = mapEditor.Update(ctx, string(partialID), overloadSetHash); err != nil {
			return prolly.AddressMap{}, err
		}
		return mapEditor.Flush(ctx)
	}(ctx, overloadMap, partialID, overloadSet)
	if err != nil {
		return err
	}

	// Write the new maps to the underlying map
	underlyingEditor := pgf.underlyingMap.Editor()
	functionMapHash, err := pgf.ns.Write(ctx, functionMap.Node())
	if err != nil {
		return err
	}
	overloadMapHash, err := pgf.ns.Write(ctx, overloadMap.Node())
	if err != nil {
		return err
	}
	if err = underlyingEditor.Update(ctx, string(functionNamespace_FunctionMap), functionMapHash); err != nil {
		return err
	}
	if err = underlyingEditor.Update(ctx, string(functionNamespace_OverloadMap), overloadMapHash); err != nil {
		return err
	}
	pgf.underlyingMap, err = underlyingEditor.Flush(ctx)
	return err
}

// DropFunction drops an existing function.
func (pgf *Collection) DropFunction(ctx context.Context, funcIDs ...id.Function) error {
	if len(funcIDs) == 0 {
		return nil
	}
	pgf.mutex.Lock()
	defer pgf.mutex.Unlock()

	// First we'll delete from the regular function map
	functionMap, ok, err := pgf.getFunctionMap(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return errors.Errorf(`function %s does not exist`, funcIDs[0].FunctionName())
	}
	// Check that each name exists before performing any deletions
	for _, funcID := range funcIDs {
		ok, err = functionMap.Has(ctx, string(funcID))
		if err != nil {
			return err
		}
		if !ok {
			return errors.Errorf(`function %s does not exist`, funcID.FunctionName())
		}
	}

	// Now we'll remove the functions from the map
	functionMap, err = func(ctx context.Context, functionMap prolly.AddressMap, funcIDs []id.Function) (_ prolly.AddressMap, err error) {
		mapEditor := functionMap.Editor()
		for _, funcID := range funcIDs {
			err := mapEditor.Delete(ctx, string(funcID))
			if err != nil {
				return prolly.AddressMap{}, err
			}
		}
		return mapEditor.Flush(ctx)
	}(ctx, functionMap, funcIDs)
	if err != nil {
		return err
	}

	// Then we'll delete from the overload map
	overloadMap, ok, err := pgf.getOverloadMap(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return errors.Errorf(`could not find the overload map while deleting function %s`, funcIDs[0].FunctionName())
	}
	for _, funcID := range funcIDs {
		overloadMap, err = func(ctx context.Context, overloadMap prolly.AddressMap, funcID id.Function) (_ prolly.AddressMap, err error) {
			partialID := id.NewFunction(funcID.SchemaName(), funcID.FunctionName())
			overloadSetHash, err := overloadMap.Get(ctx, string(partialID))
			if err != nil {
				return prolly.AddressMap{}, err
			}
			if overloadSetHash.IsEmpty() {
				return prolly.AddressMap{}, errors.Errorf(`could not find the overload set while deleting function %s`, funcID.FunctionName())
			}
			overloadSetNode, err := pgf.ns.Read(ctx, overloadSetHash)
			if err != nil {
				return prolly.AddressMap{}, err
			}
			overloadSet, err := prolly.NewAddressMap(overloadSetNode, pgf.ns)
			if err != nil {
				return prolly.AddressMap{}, err
			}
			ok, err = overloadSet.Has(ctx, string(funcID))
			if err != nil {
				return prolly.AddressMap{}, err
			}
			if !ok {
				return prolly.AddressMap{}, errors.Errorf(`could not find %s in the overload set while deleting the function`, funcID.FunctionName())
			}
			// If there's only the one entry, then we'll delete the set altogether, otherwise we just delete the entry
			count, err := overloadSet.Count()
			if err != nil {
				return prolly.AddressMap{}, err
			}
			if count == 1 {
				mapEditor := overloadMap.Editor()
				if err = mapEditor.Delete(ctx, string(partialID)); err != nil {
					return prolly.AddressMap{}, err
				}
				return mapEditor.Flush(ctx)
			} else {
				setEditor := overloadSet.Editor()
				if err = setEditor.Delete(ctx, string(funcID)); err != nil {
					return prolly.AddressMap{}, err
				}
				overloadSet, err = setEditor.Flush(ctx)
				if err != nil {
					return prolly.AddressMap{}, err
				}
				overloadSetHash, err := pgf.ns.Write(ctx, overloadSet.Node())
				if err != nil {
					return prolly.AddressMap{}, err
				}
				mapEditor := overloadMap.Editor()
				if err = mapEditor.Update(ctx, string(partialID), overloadSetHash); err != nil {
					return prolly.AddressMap{}, err
				}
				return mapEditor.Flush(ctx)
			}
		}(ctx, overloadMap, funcID)
		if err != nil {
			return err
		}
	}

	// Write the new maps to the underlying map
	underlyingEditor := pgf.underlyingMap.Editor()
	functionMapHash, err := pgf.ns.Write(ctx, functionMap.Node())
	if err != nil {
		return err
	}
	overloadMapHash, err := pgf.ns.Write(ctx, overloadMap.Node())
	if err != nil {
		return err
	}
	if err = underlyingEditor.Update(ctx, string(functionNamespace_FunctionMap), functionMapHash); err != nil {
		return err
	}
	if err = underlyingEditor.Update(ctx, string(functionNamespace_OverloadMap), overloadMapHash); err != nil {
		return err
	}
	pgf.underlyingMap, err = underlyingEditor.Flush(ctx)
	return err
}

// resolveName returns the fully resolved name of the given function. Returns an error if the name is ambiguous.
//
// The following formats are examples of a formatted name:
// name()
// name(type1, schema.type2)
// name(,,)
func (pgf *Collection) resolveName(ctx context.Context, schemaName string, formattedName string) (id.Function, error) {
	pgf.mutex.Lock()
	defer pgf.mutex.Unlock()

	// Basic checks to skip most of the logic for empty maps
	count, err := pgf.underlyingMap.Count()
	if err != nil || count == 0 || len(formattedName) == 0 {
		return id.NullFunction, err
	}
	overloadMap, ok, err := pgf.getOverloadMap(ctx)
	if err != nil || !ok {
		return id.NullFunction, err
	}
	count, err = overloadMap.Count()
	if err != nil || count == 0 {
		return id.NullFunction, err
	}

	// Extract the actual name from the format
	leftParenIndex := strings.IndexByte(formattedName, '(')
	if leftParenIndex == -1 {
		return id.NullFunction, nil
	}
	if formattedName[len(formattedName)-1] != ')' {
		return id.NullFunction, nil
	}
	functionName := strings.TrimSpace(formattedName[:leftParenIndex])
	var typeIDs []id.Type
	typePortion := strings.TrimSpace(formattedName[leftParenIndex+1 : len(formattedName)-1])
	if len(typePortion) > 0 {
		// If the type portion is just an empty string, then we don't want any type IDs
		typeStrings := strings.Split(strings.TrimSpace(formattedName[leftParenIndex+1:len(formattedName)-1]), ",")
		typeIDs = make([]id.Type, len(typeStrings))
		for i, typeString := range typeStrings {
			typeParts := strings.Split(typeString, ".")
			switch len(typeParts) {
			case 1:
				typeIDs[i] = id.NewType("", strings.TrimSpace(typeParts[0]))
			case 2:
				typeIDs[i] = id.NewType(strings.TrimSpace(typeParts[0]), strings.TrimSpace(typeParts[1]))
			default:
				return id.NullFunction, nil
			}
		}
	}

	// If there's an exact match, then we return exactly that
	{
		prefixID := id.NewFunction(schemaName, functionName)
		fullID := id.NewFunction(schemaName, functionName, typeIDs...)
		h, err := overloadMap.Get(ctx, string(prefixID))
		if err != nil {
			return id.NullFunction, err
		}
		if !h.IsEmpty() {
			node, err := pgf.ns.Read(ctx, h)
			if err != nil {
				return id.NullFunction, err
			}
			overloadSet, err := prolly.NewAddressMap(node, pgf.ns)
			if err != nil {
				return id.NullFunction, err
			}
			if ok, err = overloadSet.Has(ctx, string(fullID)); err != nil {
				return id.NullFunction, err
			} else if ok {
				return fullID, nil
			}
		}
	}

	// Now we'll iterate over all the names
	var resolvedID id.Function
	err = overloadMap.IterAll(ctx, func(k string, v hash.Hash) error {
		partialID := id.Function(k)
		if strings.EqualFold(functionName, partialID.FunctionName()) {
			if len(schemaName) > 0 && !strings.EqualFold(schemaName, partialID.SchemaName()) {
				return nil
			}
			node, err := pgf.ns.Read(ctx, v)
			if err != nil {
				return err
			}
			overloadSet, err := prolly.NewAddressMap(node, pgf.ns)
			if err != nil {
				return err
			}
			return overloadSet.IterAll(ctx, func(k string, _ hash.Hash) error {
				funcID := id.Function(k)
				if len(typeIDs) > 0 {
					if funcID.ParameterCount() != len(typeIDs) {
						return nil
					}
					for i, param := range funcID.Parameters() {
						if len(typeIDs[i].TypeName()) > 0 && !strings.EqualFold(typeIDs[i].TypeName(), param.TypeName()) {
							return nil
						}
						if len(typeIDs[i].SchemaName()) > 0 && !strings.EqualFold(typeIDs[i].SchemaName(), param.SchemaName()) {
							return nil
						}
					}
				}
				// Everything must have matched to have made it here
				if resolvedID.IsValid() {
					funcTableName := FunctionIDToTableName(funcID)
					resolvedTableName := FunctionIDToTableName(resolvedID)
					return fmt.Errorf("`%s.%s` is ambiguous, matches `%s` and `%s`",
						schemaName, formattedName, funcTableName.String(), resolvedTableName.String())
				}
				resolvedID = funcID
				return nil
			})
		}
		return nil
	})
	if err != nil {
		return id.NullFunction, err
	}
	return resolvedID, nil
}

// iterateIDs iterates over all function IDs in the collection.
func (pgf *Collection) iterateIDs(ctx context.Context, callback func(funcID id.Function) (stop bool, err error)) error {
	pgf.mutex.Lock()
	defer pgf.mutex.Unlock()

	functionMap, ok, err := pgf.getFunctionMap(ctx)
	if err != nil {
		return err
	}
	if ok {
		return functionMap.IterAll(ctx, func(k string, _ hash.Hash) error {
			stop, err := callback(id.Function(k))
			if err != nil {
				return err
			} else if stop {
				return io.EOF
			} else {
				return nil
			}
		})
	}
	return nil
}

// IterateFunctions iterates over all functions in the collection.
func (pgf *Collection) IterateFunctions(ctx context.Context, callback func(f Function) (stop bool, err error)) error {
	pgf.mutex.Lock()
	defer pgf.mutex.Unlock()

	functionMap, ok, err := pgf.getFunctionMap(ctx)
	if err != nil {
		return err
	}
	if ok {
		return functionMap.IterAll(ctx, func(_ string, v hash.Hash) error {
			data, err := pgf.ns.ReadBytes(ctx, v)
			if err != nil {
				return err
			}
			f, err := DeserializeFunction(ctx, data)
			if err != nil {
				return err
			}
			stop, err := callback(f)
			if err != nil {
				return err
			} else if stop {
				return io.EOF
			} else {
				return nil
			}
		})
	}
	return nil
}

// Clone returns a new *Collection with the same contents as the original.
func (pgf *Collection) Clone(ctx context.Context) *Collection {
	// The lock here is so that we don't request a Map while it's being written and end up with some intermediate value
	// due to a race
	pgf.mutex.Lock()
	defer pgf.mutex.Unlock()

	return &Collection{
		underlyingMap: pgf.underlyingMap,
		ns:            pgf.ns,
		mutex:         &sync.Mutex{},
	}
}

// Map writes any cached sequences to the underlying map, and then returns the underlying map.
func (pgf *Collection) Map(ctx context.Context) (prolly.AddressMap, error) {
	// The lock here is so that we don't request a Map while it's being written and end up with some intermediate value
	// due to a race
	pgf.mutex.Lock()
	defer pgf.mutex.Unlock()
	return pgf.underlyingMap, nil
}

// getFunctionMap returns the map that maps a full function ID to the definition. This does not lock the collection, as
// it is assumed that the calling function already holds the lock.
func (pgf *Collection) getFunctionMap(ctx context.Context) (prolly.AddressMap, bool, error) {
	h, err := pgf.underlyingMap.Get(ctx, string(functionNamespace_FunctionMap))
	if err != nil || h.IsEmpty() {
		return prolly.AddressMap{}, false, err
	}
	node, err := pgf.ns.Read(ctx, h)
	if err != nil {
		return prolly.AddressMap{}, false, err
	}
	m, err := prolly.NewAddressMap(node, pgf.ns)
	return m, true, err
}

// getOverloadMap returns the map that maps base names to their full names (overloads). This does not lock the
// collection, as it is assumed that the calling function already holds the lock.
func (pgf *Collection) getOverloadMap(ctx context.Context) (prolly.AddressMap, bool, error) {
	h, err := pgf.underlyingMap.Get(ctx, string(functionNamespace_OverloadMap))
	if err != nil || h.IsEmpty() {
		return prolly.AddressMap{}, false, err
	}
	node, err := pgf.ns.Read(ctx, h)
	if err != nil {
		return prolly.AddressMap{}, false, err
	}
	m, err := prolly.NewAddressMap(node, pgf.ns)
	return m, true, err
}

// getOverloadIDs returns the overloads for the function matching the schema and the function name. The parameter
// types are ignored when searching for overload IDs. This does not lock the collection, as it is assumed that the
// calling function already holds the lock.
func (pgf *Collection) getOverloadIDs(ctx context.Context, funcID id.Function) ([]id.Function, error) {
	overloadMap, ok, err := pgf.getOverloadMap(ctx)
	if err != nil || !ok {
		return nil, err
	}
	partialID := id.NewFunction(funcID.SchemaName(), funcID.FunctionName())
	overloadSetHash, err := overloadMap.Get(ctx, string(partialID))
	if err != nil || overloadSetHash.IsEmpty() {
		return nil, err
	}
	overloadSetNode, err := pgf.ns.Read(ctx, overloadSetHash)
	if err != nil {
		return nil, err
	}
	overloadSet, err := prolly.NewAddressMap(overloadSetNode, pgf.ns)
	if err != nil {
		return nil, err
	}
	count, err := overloadSet.Count()
	if err != nil {
		return nil, err
	}
	funcIDs := make([]id.Function, 0, count)
	err = overloadSet.IterAll(ctx, func(k string, _ hash.Hash) error {
		funcIDs = append(funcIDs, id.Function(k))
		return nil
	})
	return funcIDs, err
}

// tableNameToID returns the ID that was created from the within the table name.
func (pgf *Collection) tableNameToID(schemaName string, formattedName string) id.Function {
	leftParenIndex := strings.IndexByte(formattedName, '(')
	if leftParenIndex == -1 {
		return id.NullFunction
	}
	if formattedName[len(formattedName)-1] != ')' {
		return id.NullFunction
	}
	functionName := strings.TrimSpace(formattedName[:leftParenIndex])
	var typeIDs []id.Type
	typePortion := strings.TrimSpace(formattedName[leftParenIndex+1 : len(formattedName)-1])
	if len(typePortion) > 0 {
		// If the type portion is just an empty string, then we don't want any type IDs
		typeStrings := strings.Split(strings.TrimSpace(formattedName[leftParenIndex+1:len(formattedName)-1]), ",")
		typeIDs = make([]id.Type, len(typeStrings))
		for i, typeString := range typeStrings {
			typeParts := strings.Split(typeString, ".")
			switch len(typeParts) {
			case 1:
				typeIDs[i] = id.NewType("", strings.TrimSpace(typeParts[0]))
			case 2:
				typeIDs[i] = id.NewType(strings.TrimSpace(typeParts[0]), strings.TrimSpace(typeParts[1]))
			default:
				return id.NullFunction
			}
		}
	}
	return id.NewFunction(schemaName, functionName, typeIDs...)
}

// GetID implements the interface rootobject.RootObject.
func (function Function) GetID() rootobject.RootObjectID {
	return rootobject.RootObjectID_Functions
}

// HashOf implements the interface rootobject.RootObject.
func (function Function) HashOf(ctx context.Context) (hash.Hash, error) {
	data, err := function.Serialize(ctx)
	if err != nil {
		return hash.Hash{}, err
	}
	return hash.Of(data), nil
}

// Name implements the interface rootobject.RootObject.
func (function Function) Name() doltdb.TableName {
	return FunctionIDToTableName(function.ID)
}

// FunctionIDToTableName returns the ID in a format that's better for user consumption.
func FunctionIDToTableName(funcID id.Function) doltdb.TableName {
	paramTypes := funcID.Parameters()
	strTypes := make([]string, len(paramTypes))
	for i, paramType := range paramTypes {
		if paramType.SchemaName() == "pg_catalog" || paramType.SchemaName() == funcID.SchemaName() {
			strTypes[i] = paramType.TypeName()
		} else {
			strTypes[i] = fmt.Sprintf("%s.%s", paramType.SchemaName(), paramType.TypeName())
		}
	}
	return doltdb.TableName{
		Name:   fmt.Sprintf("%s(%s)", funcID.FunctionName(), strings.Join(strTypes, ",")),
		Schema: funcID.SchemaName(),
	}
}
