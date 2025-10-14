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

package procedures

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

// ParameterMode represents the mode of the given parameter (whether it's IN, OUT, INOUT, or VARIADIC).
type ParameterMode uint8

const (
	ParameterMode_IN       ParameterMode = 0
	ParameterMode_OUT      ParameterMode = 1
	ParameterMode_INOUT    ParameterMode = 2
	ParameterMode_VARIADIC ParameterMode = 3
)

// Collection contains a collection of procedures.
type Collection struct {
	accessCache   map[id.Procedure]Procedure      // This cache is used for general access when you know the exact ID
	overloadCache map[id.Procedure][]id.Procedure // This cache is used to find overloads if you know the name
	idCache       []id.Procedure                  // This cache simply contains the name of every procedure
	mapHash       hash.Hash                       // This is cached so that we don't have to calculate the hash every time
	underlyingMap prolly.AddressMap
	ns            tree.NodeStore
}

// Procedure represents a created procedure.
type Procedure struct {
	ID              id.Procedure
	ParameterNames  []string
	ParameterTypes  []id.Type
	ParameterModes  []ParameterMode
	Definition      string
	ExtensionName   string                         // Only used when this is an extension procedure
	ExtensionSymbol string                         // Only used when this is an extension procedure
	Operations      []plpgsql.InterpreterOperation // Only used when this is a plpgsql language
	SQLDefinition   string                         // Only used when this is a sql language
}

var _ objinterface.Collection = (*Collection)(nil)
var _ objinterface.RootObject = Procedure{}

// NewCollection returns a new Collection.
func NewCollection(ctx context.Context, underlyingMap prolly.AddressMap, ns tree.NodeStore) (*Collection, error) {
	collection := &Collection{
		accessCache:   make(map[id.Procedure]Procedure),
		overloadCache: make(map[id.Procedure][]id.Procedure),
		idCache:       nil,
		mapHash:       hash.Hash{},
		underlyingMap: underlyingMap,
		ns:            ns,
	}
	return collection, collection.reloadCaches(ctx)
}

// GetProcedure returns the procedure with the given ID. Returns a procedure with an invalid ID if it cannot be found
// (Procedure.ID.IsValid() == false).
func (pgp *Collection) GetProcedure(_ context.Context, procID id.Procedure) (Procedure, error) {
	if f, ok := pgp.accessCache[procID]; ok {
		return f, nil
	}
	return Procedure{}, nil
}

// GetProcedureOverloads returns the overloads for the procedure matching the schema and the procedure name. The
// parameter types are ignored when searching for overloads.
func (pgp *Collection) GetProcedureOverloads(_ context.Context, procID id.Procedure) ([]Procedure, error) {
	overloads, ok := pgp.overloadCache[id.NewProcedure(procID.SchemaName(), procID.ProcedureName())]
	if !ok || len(overloads) == 0 {
		return nil, nil
	}
	procs := make([]Procedure, len(overloads))
	for i, overload := range overloads {
		procs[i] = pgp.accessCache[overload]
	}
	return procs, nil
}

// HasProcedure returns whether the procedure is present.
func (pgp *Collection) HasProcedure(_ context.Context, procID id.Procedure) bool {
	_, ok := pgp.accessCache[procID]
	return ok
}

// AddProcedure adds a new procedure.
func (pgp *Collection) AddProcedure(ctx context.Context, proc Procedure) error {
	// First we'll check to see if it exists
	if _, ok := pgp.accessCache[proc.ID]; ok {
		return errors.Errorf(`procedure "%s" already exists with same argument types`, proc.ID.ProcedureName())
	}

	// Now we'll add the procedure to our map
	data, err := proc.Serialize(ctx)
	if err != nil {
		return err
	}
	h, err := pgp.ns.WriteBytes(ctx, data)
	if err != nil {
		return err
	}
	mapEditor := pgp.underlyingMap.Editor()
	if err = mapEditor.Add(ctx, string(proc.ID), h); err != nil {
		return err
	}
	newMap, err := mapEditor.Flush(ctx)
	if err != nil {
		return err
	}
	pgp.underlyingMap = newMap
	pgp.mapHash = pgp.underlyingMap.HashOf()
	return pgp.reloadCaches(ctx)
}

// DropProcedure drops an existing procedure.
func (pgp *Collection) DropProcedure(ctx context.Context, procIDs ...id.Procedure) error {
	if len(procIDs) == 0 {
		return nil
	}
	// Check that each name exists before performing any deletions
	for _, procID := range procIDs {
		if _, ok := pgp.accessCache[procID]; !ok {
			return errors.Errorf(`procedure %s does not exist`, procID.ProcedureName())
		}
	}

	// Now we'll remove the procedure from the map
	mapEditor := pgp.underlyingMap.Editor()
	for _, procID := range procIDs {
		err := mapEditor.Delete(ctx, string(procID))
		if err != nil {
			return err
		}
	}
	newMap, err := mapEditor.Flush(ctx)
	if err != nil {
		return err
	}
	pgp.underlyingMap = newMap
	pgp.mapHash = pgp.underlyingMap.HashOf()
	return pgp.reloadCaches(ctx)
}

// resolveName returns the fully resolved name of the given procedure. Returns an error if the name is ambiguous.
//
// The following formats are examples of a formatted name:
// name()
// name(type1, schema.type2)
// name(,,)
func (pgp *Collection) resolveName(_ context.Context, schemaName string, formattedName string) (id.Procedure, error) {
	if len(pgp.accessCache) == 0 || len(formattedName) == 0 {
		return id.NullProcedure, nil
	}

	// Extract the actual name from the format
	leftParenIndex := strings.IndexByte(formattedName, '(')
	if leftParenIndex == -1 {
		return id.NullProcedure, nil
	}
	if formattedName[len(formattedName)-1] != ')' {
		return id.NullProcedure, nil
	}
	procedureName := strings.TrimSpace(formattedName[:leftParenIndex])
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
				return id.NullProcedure, nil
			}
		}
	}

	// If there's an exact match, then we return exactly that
	fullID := id.NewProcedure(schemaName, procedureName, typeIDs...)
	if _, ok := pgp.accessCache[fullID]; ok {
		return fullID, nil
	}

	// Otherwise we'll iterate over all the names
	var resolvedID id.Procedure
OuterLoop:
	for _, procID := range pgp.idCache {
		if !strings.EqualFold(procedureName, procID.ProcedureName()) {
			continue
		}
		if len(schemaName) > 0 && !strings.EqualFold(schemaName, procID.SchemaName()) {
			continue
		}
		if len(typeIDs) > 0 {
			if procID.ParameterCount() != len(typeIDs) {
				continue
			}
			for i, param := range procID.Parameters() {
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
			procTableName := ProcedureIDToTableName(procID)
			resolvedTableName := ProcedureIDToTableName(resolvedID)
			return id.NullProcedure, fmt.Errorf("`%s.%s` is ambiguous, matches `%s` and `%s`",
				schemaName, formattedName, procTableName.String(), resolvedTableName.String())
		}
		resolvedID = procID
	}
	return resolvedID, nil
}

// iterateIDs iterates over all procedure IDs in the collection.
func (pgp *Collection) iterateIDs(_ context.Context, callback func(procID id.Procedure) (stop bool, err error)) error {
	for _, procID := range pgp.idCache {
		stop, err := callback(procID)
		if err != nil {
			return err
		} else if stop {
			return nil
		}
	}
	return nil
}

// IterateProcedures iterates over all procedures in the collection.
func (pgp *Collection) IterateProcedures(_ context.Context, callback func(f Procedure) (stop bool, err error)) error {
	for _, procID := range pgp.idCache {
		stop, err := callback(pgp.accessCache[procID])
		if err != nil {
			return err
		} else if stop {
			return nil
		}
	}
	return nil
}

// Clone returns a new *Collection with the same contents as the original.
func (pgp *Collection) Clone(_ context.Context) *Collection {
	return &Collection{
		accessCache:   maps.Clone(pgp.accessCache),
		overloadCache: maps.Clone(pgp.overloadCache),
		idCache:       slices.Clone(pgp.idCache),
		mapHash:       pgp.mapHash,
		underlyingMap: pgp.underlyingMap,
		ns:            pgp.ns,
	}
}

// Map returns the underlying map.
func (pgp *Collection) Map(_ context.Context) (prolly.AddressMap, error) {
	return pgp.underlyingMap, nil
}

// DiffersFrom returns true when the hash that is associated with the underlying map for this collection is different
// from the hash in the given root.
func (pgp *Collection) DiffersFrom(ctx context.Context, root objinterface.RootValue) bool {
	hashOnGivenRoot, err := pgp.LoadCollectionHash(ctx, root)
	if err != nil {
		return true
	}
	if pgp.mapHash.Equal(hashOnGivenRoot) {
		return false
	}
	// An empty map should match an uninitialized collection on the root
	count, err := pgp.underlyingMap.Count()
	if err == nil && count == 0 && hashOnGivenRoot.IsEmpty() {
		return false
	}
	return true
}

// reloadCaches writes the underlying map's contents to the caches.
func (pgp *Collection) reloadCaches(ctx context.Context) error {
	count, err := pgp.underlyingMap.Count()
	if err != nil {
		return err
	}

	clear(pgp.accessCache)
	clear(pgp.overloadCache)
	pgp.mapHash = pgp.underlyingMap.HashOf()
	pgp.idCache = make([]id.Procedure, 0, count)

	return pgp.underlyingMap.IterAll(ctx, func(_ string, h hash.Hash) error {
		if h.IsEmpty() {
			return nil
		}
		data, err := pgp.ns.ReadBytes(ctx, h)
		if err != nil {
			return err
		}
		f, err := DeserializeProcedure(ctx, data)
		if err != nil {
			return err
		}
		pgp.accessCache[f.ID] = f
		partialID := id.NewProcedure(f.ID.SchemaName(), f.ID.ProcedureName())
		pgp.overloadCache[partialID] = append(pgp.overloadCache[partialID], f.ID)
		pgp.idCache = append(pgp.idCache, f.ID)
		return nil
	})
}

// tableNameToID returns the ID that was encoded via the Name() call, as the returned TableName contains additional
// information (which this is able to process).
func (pgp *Collection) tableNameToID(schemaName string, formattedName string) id.Procedure {
	leftParenIndex := strings.IndexByte(formattedName, '(')
	if leftParenIndex == -1 {
		return id.NullProcedure
	}
	if formattedName[len(formattedName)-1] != ')' {
		return id.NullProcedure
	}
	procedureName := strings.TrimSpace(formattedName[:leftParenIndex])
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
				return id.NullProcedure
			}
		}
	}
	return id.NewProcedure(schemaName, procedureName, typeIDs...)
}

// GetID implements the interface objinterface.RootObject.
func (procedure Procedure) GetID() id.Id {
	return procedure.ID.AsId()
}

// GetInnerDefinition returns the inner definition inside the CREATE PROCEDURE statement.
func (procedure Procedure) GetInnerDefinition() string {
	// TODO: right now we're hardcode searching for $$, which will fail for some definition strings
	start := strings.Index(procedure.Definition, "$$")
	end := strings.LastIndex(procedure.Definition, "$$")
	if start == -1 || end == -1 {
		// Return the whole definition for now
		return procedure.Definition
	}
	return strings.TrimSpace(procedure.Definition[start+2 : end])
}

// ReplaceDefinition returns a new definition with the inner portion replaced with the given string.
func (procedure Procedure) ReplaceDefinition(newInner string) string {
	return strings.Replace(procedure.Definition, procedure.GetInnerDefinition(), newInner, 1)
}

// GetRootObjectID implements the interface objinterface.RootObject.
func (procedure Procedure) GetRootObjectID() objinterface.RootObjectID {
	return objinterface.RootObjectID_Procedures
}

// HashOf implements the interface objinterface.RootObject.
func (procedure Procedure) HashOf(ctx context.Context) (hash.Hash, error) {
	data, err := procedure.Serialize(ctx)
	if err != nil {
		return hash.Hash{}, err
	}
	return hash.Of(data), nil
}

// Name implements the interface objinterface.RootObject.
func (procedure Procedure) Name() doltdb.TableName {
	return ProcedureIDToTableName(procedure.ID)
}

// ParameterModesAsString returns a string that represents the parameter modes. The string may be converted back to a
// slice using ParameterModesFromString.
func (procedure Procedure) ParameterModesAsString() string {
	sb := strings.Builder{}
	for i, mode := range procedure.ParameterModes {
		if i > 0 {
			sb.WriteRune(',')
		}
		switch mode {
		case ParameterMode_IN:
			sb.WriteString("in")
		case ParameterMode_OUT:
			sb.WriteString("out")
		case ParameterMode_INOUT:
			sb.WriteString("inout")
		case ParameterMode_VARIADIC:
			sb.WriteString("variadic")
		default:
			panic("unhandled procedure parameter mode")
		}
	}
	return sb.String()
}

// ProcedureIDToTableName returns the ID in a format that's better for user consumption.
func ProcedureIDToTableName(procID id.Procedure) doltdb.TableName {
	paramTypes := procID.Parameters()
	strTypes := make([]string, len(paramTypes))
	for i, paramType := range paramTypes {
		if paramType.SchemaName() == "pg_catalog" || paramType.SchemaName() == procID.SchemaName() {
			strTypes[i] = paramType.TypeName()
		} else {
			strTypes[i] = fmt.Sprintf("%s.%s", paramType.SchemaName(), paramType.TypeName())
		}
	}
	return doltdb.TableName{
		Name:   fmt.Sprintf("%s(%s)", procID.ProcedureName(), strings.Join(strTypes, ",")),
		Schema: procID.SchemaName(),
	}
}

// ParameterModesFromString returns a ParameterMode slice from the given string. It is assumed that this string was
// originally created using Procedure.ParameterModesAsString.
func ParameterModesFromString(str string) ([]ParameterMode, error) {
	if len(str) == 0 {
		return nil, nil
	}
	modeStrings := strings.Split(str, ",")
	modes := make([]ParameterMode, len(modeStrings))
	for i, modeString := range modeStrings {
		switch modeString {
		case "in":
			modes[i] = ParameterMode_IN
		case "out":
			modes[i] = ParameterMode_OUT
		case "inout":
			modes[i] = ParameterMode_INOUT
		case "variadic":
			modes[i] = ParameterMode_VARIADIC
		default:
			return nil, errors.Errorf("`%s` is not a valid parameter argmode, it may be one of the following: in, out, inout, variadic", modeString)
		}
	}
	return modes, nil
}
