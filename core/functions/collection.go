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
	"maps"
	"slices"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/store/hash"
	"github.com/dolthub/dolt/go/store/prolly"
	"github.com/dolthub/dolt/go/store/prolly/tree"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/rootobject/objinterface"
	"github.com/dolthub/doltgresql/server/plpgsql"
)

// Collection contains a collection of functions.
type Collection struct {
	accessCache   map[id.Function]Function      // This cache is used for general access when you know the exact ID
	overloadCache map[id.Function][]id.Function // This cache is used to find overloads if you know the name
	idCache       []id.Function                 // This cache simply contains the name of every function
	mapHash       hash.Hash                     // This is cached so that we don't have to calculate the hash every time
	underlyingMap prolly.AddressMap
	ns            tree.NodeStore
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
	Definition         string
	ExtensionName      string                         // Only used when this is an extension function
	ExtensionSymbol    string                         // Only used when this is an extension function
	Operations         []plpgsql.InterpreterOperation // Only used when this is a plpgsql language
	SQLDefinition      string                         // Only used when this is a sql language
}

var _ objinterface.Collection = (*Collection)(nil)
var _ objinterface.RootObject = Function{}

// NewCollection returns a new Collection.
func NewCollection(ctx context.Context, underlyingMap prolly.AddressMap, ns tree.NodeStore) (*Collection, error) {
	collection := &Collection{
		accessCache:   make(map[id.Function]Function),
		overloadCache: make(map[id.Function][]id.Function),
		idCache:       nil,
		mapHash:       hash.Hash{},
		underlyingMap: underlyingMap,
		ns:            ns,
	}
	return collection, collection.reloadCaches(ctx)
}

// GetFunction returns the function with the given ID. Returns a function with an invalid ID if it cannot be found
// (Function.ID.IsValid() == false).
func (pgf *Collection) GetFunction(ctx context.Context, funcID id.Function) (Function, error) {
	if f, ok := pgf.accessCache[funcID]; ok {
		return f, nil
	}
	return Function{}, nil
}

// GetFunctionOverloads returns the overloads for the function matching the schema and the function name. The parameter
// types are ignored when searching for overloads.
func (pgf *Collection) GetFunctionOverloads(ctx context.Context, funcID id.Function) ([]Function, error) {
	overloads, ok := pgf.overloadCache[id.NewFunction(funcID.SchemaName(), funcID.FunctionName())]
	if !ok || len(overloads) == 0 {
		return nil, nil
	}
	funcs := make([]Function, len(overloads))
	for i, overload := range overloads {
		funcs[i] = pgf.accessCache[overload]
	}
	return funcs, nil
}

// HasFunction returns whether the function is present.
func (pgf *Collection) HasFunction(ctx context.Context, funcID id.Function) bool {
	_, ok := pgf.accessCache[funcID]
	return ok
}

// AddFunction adds a new function.
func (pgf *Collection) AddFunction(ctx context.Context, f Function) error {
	// First we'll check to see if it exists
	if _, ok := pgf.accessCache[f.ID]; ok {
		return errors.Errorf(`function "%s" already exists with same argument types`, f.ID.FunctionName())
	}

	// Now we'll add the function to our map
	data, err := f.Serialize(ctx)
	if err != nil {
		return err
	}
	h, err := pgf.ns.WriteBytes(ctx, data)
	if err != nil {
		return err
	}
	mapEditor := pgf.underlyingMap.Editor()
	if err = mapEditor.Add(ctx, string(f.ID), h); err != nil {
		return err
	}
	newMap, err := mapEditor.Flush(ctx)
	if err != nil {
		return err
	}
	pgf.underlyingMap = newMap
	pgf.mapHash = pgf.underlyingMap.HashOf()
	return pgf.reloadCaches(ctx)
}

// DropFunction drops an existing function.
func (pgf *Collection) DropFunction(ctx context.Context, funcIDs ...id.Function) error {
	if len(funcIDs) == 0 {
		return nil
	}
	// Check that each name exists before performing any deletions
	for _, funcID := range funcIDs {
		if _, ok := pgf.accessCache[funcID]; !ok {
			return errors.Errorf(`function %s does not exist`, funcID.FunctionName())
		}
	}

	// Now we'll remove the functions from the map
	mapEditor := pgf.underlyingMap.Editor()
	for _, funcID := range funcIDs {
		err := mapEditor.Delete(ctx, string(funcID))
		if err != nil {
			return err
		}
	}
	newMap, err := mapEditor.Flush(ctx)
	if err != nil {
		return err
	}
	pgf.underlyingMap = newMap
	pgf.mapHash = pgf.underlyingMap.HashOf()
	return pgf.reloadCaches(ctx)
}

// resolveName returns the fully resolved name of the given function. Returns an error if the name is ambiguous.
//
// The following formats are examples of a formatted name:
// name()
// name(type1, schema.type2)
// name(,,)
func (pgf *Collection) resolveName(ctx context.Context, schemaName string, formattedName string) (id.Function, error) {
	if len(pgf.accessCache) == 0 || len(formattedName) == 0 {
		return id.NullFunction, nil
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
	fullID := id.NewFunction(schemaName, functionName, typeIDs...)
	if _, ok := pgf.accessCache[fullID]; ok {
		return fullID, nil
	}

	// Otherwise we'll iterate over all the names
	var resolvedID id.Function
OuterLoop:
	for _, funcID := range pgf.idCache {
		if !strings.EqualFold(functionName, funcID.FunctionName()) {
			continue
		}
		if len(schemaName) > 0 && !strings.EqualFold(schemaName, funcID.SchemaName()) {
			continue
		}
		if len(typeIDs) > 0 {
			if funcID.ParameterCount() != len(typeIDs) {
				continue
			}
			for i, param := range funcID.Parameters() {
				if len(typeIDs[i].TypeName()) > 0 && !strings.EqualFold(typeIDs[i].TypeName(), param.TypeName()) {
					continue OuterLoop
				}
				if len(typeIDs[i].SchemaName()) > 0 && !strings.EqualFold(typeIDs[i].SchemaName(), param.SchemaName()) {
					continue OuterLoop
				}
			}
		}
		// Everything must have matched to have made it here
		if resolvedID.IsValid() {
			funcTableName := FunctionIDToTableName(funcID)
			resolvedTableName := FunctionIDToTableName(resolvedID)
			return id.NullFunction, fmt.Errorf("`%s.%s` is ambiguous, matches `%s` and `%s`",
				schemaName, formattedName, funcTableName.String(), resolvedTableName.String())
		}
		resolvedID = funcID
	}
	return resolvedID, nil
}

// iterateIDs iterates over all function IDs in the collection.
func (pgf *Collection) iterateIDs(ctx context.Context, callback func(funcID id.Function) (stop bool, err error)) error {
	for _, funcID := range pgf.idCache {
		stop, err := callback(funcID)
		if err != nil {
			return err
		} else if stop {
			return nil
		}
	}
	return nil
}

// IterateFunctions iterates over all functions in the collection.
func (pgf *Collection) IterateFunctions(ctx context.Context, callback func(f Function) (stop bool, err error)) error {
	for _, funcID := range pgf.idCache {
		stop, err := callback(pgf.accessCache[funcID])
		if err != nil {
			return err
		} else if stop {
			return nil
		}
	}
	return nil
}

// Clone returns a new *Collection with the same contents as the original.
func (pgf *Collection) Clone(ctx context.Context) *Collection {
	return &Collection{
		accessCache:   maps.Clone(pgf.accessCache),
		overloadCache: maps.Clone(pgf.overloadCache),
		idCache:       slices.Clone(pgf.idCache),
		underlyingMap: pgf.underlyingMap,
		mapHash:       pgf.mapHash,
		ns:            pgf.ns,
	}
}

// Map writes any cached sequences to the underlying map, and then returns the underlying map.
func (pgf *Collection) Map(ctx context.Context) (prolly.AddressMap, error) {
	return pgf.underlyingMap, nil
}

// DiffersFrom returns true when the hash that is associated with the underlying map for this collection is different
// from the hash in the given root.
func (pgf *Collection) DiffersFrom(ctx context.Context, root objinterface.RootValue) bool {
	hashOnGivenRoot, err := pgf.LoadCollectionHash(ctx, root)
	if err != nil {
		return true
	}
	if pgf.mapHash.Equal(hashOnGivenRoot) {
		return false
	}
	// An empty map should match an uninitialized collection on the root
	count, err := pgf.underlyingMap.Count()
	if err == nil && count == 0 && hashOnGivenRoot.IsEmpty() {
		return false
	}
	return true
}

// reloadCaches writes the underlying map's contents to the caches.
func (pgf *Collection) reloadCaches(ctx context.Context) error {
	count, err := pgf.underlyingMap.Count()
	if err != nil {
		return err
	}

	clear(pgf.accessCache)
	clear(pgf.overloadCache)
	pgf.mapHash = pgf.underlyingMap.HashOf()
	pgf.idCache = make([]id.Function, 0, count)

	return pgf.underlyingMap.IterAll(ctx, func(_ string, h hash.Hash) error {
		if h.IsEmpty() {
			return nil
		}
		data, err := pgf.ns.ReadBytes(ctx, h)
		if err != nil {
			return err
		}
		f, err := DeserializeFunction(ctx, data)
		if err != nil {
			return err
		}
		pgf.accessCache[f.ID] = f
		partialID := id.NewFunction(f.ID.SchemaName(), f.ID.FunctionName())
		pgf.overloadCache[partialID] = append(pgf.overloadCache[partialID], f.ID)
		pgf.idCache = append(pgf.idCache, f.ID)
		return nil
	})
}

// tableNameToID returns the ID that was encoded via the Name() call, as the returned TableName contains additional
// information (which this is able to process).
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

// GetID implements the interface objinterface.RootObject.
func (function Function) GetID() id.Id {
	return function.ID.AsId()
}

// GetInnerDefinition returns the inner definition inside the CREATE FUNCTION statement.
func (function Function) GetInnerDefinition() string {
	// TODO: right now we're hardcode searching for $$, which will fail for some definition strings
	start := strings.Index(function.Definition, "$$")
	end := strings.LastIndex(function.Definition, "$$")
	if start == -1 || end == -1 {
		// Return the whole definition for now
		return function.Definition
	}
	return strings.TrimSpace(function.Definition[start+2 : end])
}

// ReplaceDefinition returns a new definition with the inner portion replaced with the given string.
func (function Function) ReplaceDefinition(newInner string) string {
	return strings.Replace(function.Definition, function.GetInnerDefinition(), newInner, 1)
}

// GetRootObjectID implements the interface objinterface.RootObject.
func (function Function) GetRootObjectID() objinterface.RootObjectID {
	return objinterface.RootObjectID_Functions
}

// HashOf implements the interface objinterface.RootObject.
func (function Function) HashOf(ctx context.Context) (hash.Hash, error) {
	data, err := function.Serialize(ctx)
	if err != nil {
		return hash.Hash{}, err
	}
	return hash.Of(data), nil
}

// Name implements the interface objinterface.RootObject.
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
