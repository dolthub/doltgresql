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

package typecollection

import (
	"context"
	"io"
	"sort"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/store/hash"
	"github.com/dolthub/dolt/go/store/prolly"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/rootobject/objinterface"
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	parsertypes "github.com/dolthub/doltgresql/postgres/parser/types"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// anonymousCompositePrefix is the prefix for anonymous composite type names. These types are not stored on
// disk, but instead are created dynamically as needed.
const anonymousCompositePrefix = "table("

// anonymousCompositeSuffix is the suffix for anonymous composite type names.
const anonymousCompositeSuffix = ")"

// TypeCollection is a collection of all types (both built-in and user defined).
type TypeCollection struct {
	objinterface.RootObjectMap
	accessedMap map[id.Type]*pgtypes.DoltgresType
	initCache   map[id.Type]*pgtypes.DoltgresType // This is only used by the function `WithCachedType`
}

// TypeWrapper is a wrapper around a type that allows it to be used as a root object.
type TypeWrapper struct {
	Type *pgtypes.DoltgresType
}

var _ objinterface.Collection = (*TypeCollection)(nil)
var _ objinterface.RootObject = TypeWrapper{}
var _ doltdb.RootObject = TypeWrapper{}

// CreateType creates a new type.
func (pgs *TypeCollection) CreateType(ctx context.Context, typ *pgtypes.DoltgresType) error {
	// First we check the built-in types
	if _, ok := pgtypes.IDToBuiltInDoltgresType[typ.ID]; ok {
		return pgtypes.ErrTypeAlreadyExists.New(typ.Name())
	}

	// Ensure that the type does not already exist in the cache or underlying map
	if _, ok := pgs.accessedMap[typ.ID]; ok {
		return pgtypes.ErrTypeAlreadyExists.New(typ.Name())
	}
	if ok, err := pgs.Contents().Has(ctx, string(typ.ID)); err != nil {
		return err
	} else if ok {
		return pgtypes.ErrTypeAlreadyExists.New(typ.Name())
	}
	// Add it to our cache, which will be written when we do anything permanent
	pgs.accessedMap[typ.ID] = typ
	return nil
}

// DropType drops an existing type.
func (pgs *TypeCollection) DropType(ctx context.Context, names ...id.Type) (err error) {
	// First we'll check if we're trying to drop a built-in type
	for _, name := range names {
		if _, ok := pgtypes.IDToBuiltInDoltgresType[name]; ok {
			// TODO: investigate why we sometimes attempt to drop built-in types
			return nil
		}
	}

	// We need to clear the cache so that we only need to worry about the underlying map
	if err = pgs.writeCache(ctx); err != nil {
		return err
	}
	for _, name := range names {
		if ok, err := pgs.Contents().Has(ctx, string(name)); err != nil {
			return err
		} else if !ok {
			return pgtypes.ErrTypeDoesNotExist.New(name.TypeName())
		}
	}
	// Now we'll remove the types from the underlying map
	mapEditor := pgs.Contents().Editor()
	for _, name := range names {
		if err = mapEditor.Delete(ctx, string(name)); err != nil {
			return err
		}
	}
	flushed, err := mapEditor.Flush(ctx)
	if err != nil {
		return err
	}
	pgs.SetContents(flushed)
	return nil
}

// GetAllTypes returns a map containing all types in the collection, grouped by the schema they're contained in.
// Each type array is also sorted by the type name. It includes built-in types.
func (pgs *TypeCollection) GetAllTypes(ctx context.Context) (typeMap map[string][]*pgtypes.DoltgresType, schemaNames []string, totalCount int, err error) {
	schemaNamesMap := make(map[string]struct{})
	typeMap = make(map[string][]*pgtypes.DoltgresType)
	err = pgs.IterateTypes(ctx, func(t *pgtypes.DoltgresType) (stop bool, err error) {
		schemaNamesMap[t.ID.SchemaName()] = struct{}{}
		typeMap[t.ID.SchemaName()] = append(typeMap[t.ID.SchemaName()], t)
		totalCount++
		return false, nil
	})
	if err != nil {
		return nil, nil, 0, err
	}
	// Sort the types in the type map
	for _, seqs := range typeMap {
		sort.Slice(seqs, func(i, j int) bool {
			return seqs[i].ID < seqs[j].ID
		})
	}
	// Create and sort the schema names
	schemaNames = make([]string, 0, len(schemaNamesMap))
	for name := range schemaNamesMap {
		schemaNames = append(schemaNames, name)
	}
	sort.Slice(schemaNames, func(i, j int) bool {
		return schemaNames[i] < schemaNames[j]
	})
	return
}

// GetDomainType returns a domain type with the given schema and name.
// Returns nil if the type cannot be found. It checks for domain type.
func (pgs *TypeCollection) GetDomainType(ctx context.Context, name id.Type) (*pgtypes.DoltgresType, error) {
	t, err := pgs.GetType(ctx, name)
	if err != nil || t == nil {
		return nil, err
	}
	if t.TypType == pgtypes.TypeType_Domain {
		return t, nil
	}
	return nil, nil
}

// GetType returns the type with the given schema and name.
// Returns nil if the type cannot be found.
func (pgs *TypeCollection) GetType(ctx context.Context, name id.Type) (*pgtypes.DoltgresType, error) {
	return pgs.GetTypeWithTypmod(ctx, name, nil)
}

// GetTypeWithTypmod returns the type with the given schema and name, with the given type modifiers applied.
// Returns nil if the type cannot be found.
func (pgs *TypeCollection) GetTypeWithTypmod(ctx context.Context, name id.Type, typmods []any) (*pgtypes.DoltgresType, error) {
	// Check the built-in types first
	if t, ok := pgtypes.IDToBuiltInDoltgresType[name]; ok {
		return typeWithTypmod(ctx, t, typmods)
	}

	// Subsequent loads are cached
	if t, ok := pgs.accessedMap[name]; ok {
		return typeWithTypmod(ctx, t, typmods)
	}
	if t, ok := pgs.initCache[name]; ok {
		return typeWithTypmod(ctx, t, typmods)
	}
	sqlCtx, ok := ctx.(*sql.Context)
	if !ok {
		return nil, errors.New("type collection requires a SQL context")
	}
	// The initial load is from the internal map
	h, err := pgs.Contents().Get(ctx, string(name))
	if err != nil {
		return nil, err
	}
	if h.IsEmpty() {
		// If this is an anonymous composite type, create it dynamically
		if isAnonymousCompositeType(name) {
			t, err := pgs.createAnonymousCompositeType(sqlCtx, name)
			if err != nil {
				return nil, err
			}
			return typeWithTypmod(ctx, t, typmods)
		}

		// Table composite types are computed on the fly from the live table schema rather than
		// stored as root objects (storing them would create a naming collision with the actual
		// table in Dolt's diff layer, since both map to the same doltdb.TableName).
		typeName := name.TypeName()

		// A name starting with "_" may be the implicit array type for a table's composite row
		// type. Resolve the element type first (which handles the table lookup), then wrap it.
		if strings.HasPrefix(typeName, "_") {
			elemType, err := pgs.GetType(ctx, id.NewType(name.SchemaName(), typeName[1:]))
			if err != nil || elemType == nil {
				return nil, err
			}
			return typeWithTypmod(ctx, pgtypes.CreateArrayTypeFromBaseType(elemType), typmods)
		}

		tbl, schema, err := pgs.getTable(sqlCtx, name.SchemaName(), typeName)
		if err != nil || tbl == nil {
			return nil, err
		}
		t, err := pgs.tableToType(sqlCtx, tbl, schema)
		if err != nil {
			return nil, err
		}
		return typeWithTypmod(ctx, t, typmods)
	}
	data, err := pgs.NodeStore().ReadBytes(ctx, h)
	if err != nil {
		return nil, err
	}
	t, err := pgtypes.DeserializeTypeFromCollection(sqlCtx, pgs, data)
	if err != nil {
		return nil, err
	}
	pgt := t.(*pgtypes.DoltgresType)
	pgs.accessedMap[pgt.ID] = pgt

	return typeWithTypmod(ctx, pgt, typmods)
}

// ResolveType returns the type given if there's an exact match, or the closest matching type if the exact ID cannot be
// found. In general, this should only be used in cases where we are not sure of the schema name, as this is
// significantly slower than GetType. Returns an error if the type cannot be resolved, unlike GetType which returns a
// nil if the type is not found.
func (pgs *TypeCollection) ResolveType(ctx context.Context, name id.Type) (*pgtypes.DoltgresType, error) {
	return pgs.ResolveTypeWithTypmod(ctx, name, nil)
}

// ResolveTypeWithTypmod resolves the type in the same manner as ResolveType, with the given type modifiers applied.
func (pgs *TypeCollection) ResolveTypeWithTypmod(ctx context.Context, name id.Type, typmods []any) (*pgtypes.DoltgresType, error) {
	if t, err := pgs.GetTypeWithTypmod(ctx, name, typmods); err != nil {
		return nil, err
	} else if t != nil && t.IsResolvedType() {
		return t, nil
	}
	resolvedId, err := pgs.resolveName(ctx, name.SchemaName(), name.TypeName())
	if err != nil {
		return nil, err
	}
	t, err := pgs.GetTypeWithTypmod(ctx, resolvedId, typmods)
	if err != nil {
		return nil, err
	}
	if !t.IsResolvedType() {
		return nil, pgerror.WithCandidateCode(errors.Errorf("unable to resolve type `%s`", name.TypeName()), pgcode.UndefinedObject)
	}
	return t, nil
}

// typeWithTypmod returns the given type with the given type modifiers applied, or the type unchanged if there are none.
func typeWithTypmod(ctx context.Context, t *pgtypes.DoltgresType, typmods []any) (*pgtypes.DoltgresType, error) {
	if len(typmods) == 0 {
		return t, nil
	}
	sqlCtx, ok := ctx.(*sql.Context)
	if !ok {
		return nil, errors.New("type collection requires a SQL context")
	}
	if t.ModInFunc == 0 {
		return nil, errors.Errorf(`type modifier is not allowed for type "%s"`, t.Name())
	}
	typmod, err := t.TypModIn(sqlCtx, typmods)
	if err != nil {
		return nil, err
	}
	return t.WithAttTypMod(typmod), nil
}

// WithCachedType executes the given function while caching the given type, which allows for recursive type
// initialization to reference unfinished types.
func (pgs *TypeCollection) WithCachedType(typeToCache *pgtypes.DoltgresType, f func()) {
	pgs.initCache[typeToCache.ID] = typeToCache
	defer func() {
		delete(pgs.initCache, typeToCache.ID)
	}()
	f()
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
func (pgs *TypeCollection) createAnonymousCompositeType(ctx *sql.Context, returnType id.Type) (*pgtypes.DoltgresType, error) {
	typeName := returnType.TypeName()
	attributeTypes := typeName[len(anonymousCompositePrefix) : len(typeName)-len(anonymousCompositeSuffix)]
	attributeTypesSlice := strings.Split(attributeTypes, ",")

	attrs := make([]pgtypes.CompositeAttribute, len(attributeTypesSlice))
	for i, attributeNameAndType := range attributeTypesSlice {
		split := strings.Split(attributeNameAndType, ":")
		if len(split) != 2 {
			return nil, errors.Errorf("unexpected anonymous composite type attribute syntax: %s", attributeNameAndType)
		}
		// Attribute names may be standard SQL names such as "SMALLINT", so we have to normalize such names
		attributeName := split[1]
		var attributeType *pgtypes.DoltgresType
		var err error
		if parserT, ok, _ := parsertypes.TypeForNonKeywordTypeName(strings.ToLower(attributeName)); ok {
			typeId := id.Cache().ToInternal(uint32(parserT.Oid()))
			if typeId.IsValid() && typeId.Section() == id.Section_Type {
				attributeType, err = pgs.ResolveType(ctx, id.Type(typeId))
				if err != nil {
					return nil, err
				}
			}
		}
		if !attributeType.IsResolvedType() {
			// We check if a schema is present by the existence of a "."
			schemaName := ""
			if strings.Contains(attributeName, ".") {
				typeSplit := strings.SplitN(attributeName, ".", 2)
				schemaName = typeSplit[0]
				attributeName = typeSplit[1]
			}
			attributeType, err = pgs.ResolveType(ctx, id.NewType(schemaName, attributeName))
			if err != nil {
				return nil, err
			}
		}
		attrs[i] = pgtypes.NewCompositeAttribute(ctx, id.Null, split[0], attributeType, int16(i), "")
	}
	return pgtypes.NewCompositeType(ctx, id.Null, nil, returnType, attrs), nil
}

// HasType checks if a type exists with given schema and type name.
func (pgs *TypeCollection) HasType(ctx context.Context, name id.Type) bool {
	// We can check the built-in types first
	if _, ok := pgtypes.IDToBuiltInDoltgresType[name]; ok {
		return true
	}
	// Now we'll check our created types
	if _, ok := pgs.accessedMap[name]; ok {
		return true
	}
	ok, err := pgs.Contents().Has(ctx, string(name))
	if err == nil && ok {
		return true
	}
	// Table composite types are not stored; check the table as a fallback.
	sqlCtx, ok := ctx.(*sql.Context)
	if !ok {
		return false
	}
	tbl, _, err := pgs.getTable(sqlCtx, name.SchemaName(), name.TypeName())
	return err == nil && tbl != nil
}

// resolveName returns the fully resolved name of the given type. Returns an error if the name is ambiguous.
func (pgs *TypeCollection) resolveName(ctx context.Context, schemaName string, typeName string) (id.Type, error) {
	// TODO: this should probably check table names as well since tables create composite types matching their rows
	// First check for an exact match in the built-in types
	inputID := id.NewType(schemaName, typeName)
	if _, ok := pgtypes.IDToBuiltInDoltgresType[inputID]; ok {
		return inputID, nil
	}

	// Iterate over all the built-in names for a relative match
	var resolvedID id.Type
	for internalID, typ := range pgtypes.IDToBuiltInDoltgresType { // TODO: make a map by typeName?
		if typ.ID == pgtypes.Unknown.ID && internalID.TypeName() != "unknown" {
			continue
		}
		if strings.EqualFold(typeName, typ.ID.TypeName()) {
			if len(schemaName) > 0 && !strings.EqualFold(schemaName, typ.ID.SchemaName()) {
				continue
			}
			if resolvedID.IsValid() {
				return id.NullType, errors.Errorf("`%s.%s` is ambiguous, matches `%s.%s` and `%s.%s`",
					schemaName, typeName, typ.ID.SchemaName(), typ.ID.TypeName(), resolvedID.SchemaName(), resolvedID.TypeName())
			}
			resolvedID = typ.ID
		}
	}

	// Iterate over the initialization cache in case this is during a type initialization loop
	for _, typ := range pgs.initCache {
		if strings.EqualFold(typeName, typ.ID.TypeName()) {
			if len(schemaName) > 0 && !strings.EqualFold(schemaName, typ.ID.SchemaName()) {
				continue
			}
			if resolvedID.IsValid() {
				return id.NullType, errors.Errorf("`%s.%s` is ambiguous, matches `%s.%s` and `%s.%s`",
					schemaName, typeName, typ.ID.SchemaName(), typ.ID.TypeName(), resolvedID.SchemaName(), resolvedID.TypeName())
			}
			resolvedID = typ.ID
		}
	}

	// We write the cache so that we only need to worry about the underlying map
	if err := pgs.writeCache(ctx); err != nil {
		return id.NullType, err
	}

	// Check for an exact match in the underlying map
	ok, err := pgs.Contents().Has(ctx, string(inputID))
	if err != nil {
		return id.NullType, err
	} else if ok {
		// We don't bother looking if there's an existing match, since this is an exact match (so no ambiguity)
		return inputID, nil
	}

	// Iterate over all the names in the map
	err = pgs.Contents().IterAll(ctx, func(k string, _ hash.Hash) error {
		typeID := id.Type(k)
		if strings.EqualFold(typeName, typeID.TypeName()) {
			if len(schemaName) > 0 && !strings.EqualFold(schemaName, typeID.SchemaName()) {
				return nil
			}
			if resolvedID.IsValid() {
				return errors.Errorf("`%s.%s` is ambiguous, matches `%s.%s` and `%s.%s`",
					schemaName, typeName, typeID.SchemaName(), typeID.TypeName(), resolvedID.SchemaName(), resolvedID.TypeName())
			}
			resolvedID = typeID
		}
		return nil
	})
	if err != nil {
		return id.NullType, err
	}
	return resolvedID, nil
}

// IterateTypes iterates over all types in the collection.
func (pgs *TypeCollection) IterateTypes(ctx context.Context, f func(typ *pgtypes.DoltgresType) (stop bool, err error)) error {
	// TODO: this should probably iterate tables as well since tables create composite types matching their rows
	// We can iterate the built-in types first
	for _, t := range pgtypes.GetAllBuitInTypes() {
		stop, err := f(t)
		if err != nil || stop {
			return err
		}
	}

	sqlCtx, ok := ctx.(*sql.Context)
	if !ok {
		return errors.New("type collection requires a SQL context")
	}
	// We write the cache so that we only need to worry about the underlying map
	if err := pgs.writeCache(ctx); err != nil {
		return err
	}
	err := pgs.Contents().IterAll(ctx, func(_ string, v hash.Hash) error {
		data, err := pgs.NodeStore().ReadBytes(ctx, v)
		if err != nil {
			return err
		}
		t, err := pgtypes.DeserializeTypeFromCollection(sqlCtx, pgs, data)
		if err != nil {
			return err
		}
		stop, err := f(t.(*pgtypes.DoltgresType))
		if err != nil {
			return err
		} else if stop {
			return io.EOF
		} else {
			return nil
		}
	})
	return err
}

// Map writes any cached types to the collection's contents, and then returns those contents.
func (pgs *TypeCollection) Map(ctx context.Context) (prolly.AddressMap, error) {
	if err := pgs.writeCache(ctx); err != nil {
		return prolly.AddressMap{}, err
	}
	return pgs.Contents(), nil
}

// GetID implements the interface objinterface.RootObject.
func (t TypeWrapper) GetID() id.Id {
	if t.Type != nil {
		return t.Type.ID.AsId()
	}
	return id.Null
}

// GetRootObjectID implements the interface objinterface.RootObject.
func (t TypeWrapper) GetRootObjectID() objinterface.RootObjectID {
	return objinterface.RootObjectID_Types
}

// HashOf implements the interface objinterface.RootObject.
func (t TypeWrapper) HashOf(ctx context.Context) (hash.Hash, error) {
	if t.Type != nil {
		return hash.Of(t.Type.Serialize()), nil
	}
	return hash.Hash{}, nil
}

// Name implements the interface objinterface.RootObject.
func (t TypeWrapper) Name() doltdb.TableName {
	if t.Type != nil {
		return doltdb.TableName{
			Name:   t.Type.ID.TypeName(),
			Schema: t.Type.ID.SchemaName(),
		}
	}
	return doltdb.TableName{}
}

// Serialize implements the interface objinterface.RootObject.
func (t TypeWrapper) Serialize(ctx context.Context) ([]byte, error) {
	if t.Type != nil {
		return t.Type.Serialize(), nil
	}
	return nil, nil
}

// writeCache writes every type in the cache to the underlying map.
func (pgs *TypeCollection) writeCache(ctx context.Context) (err error) {
	if len(pgs.accessedMap) == 0 {
		return nil
	}
	mapEditor := pgs.Contents().Editor()
	for _, t := range pgs.accessedMap {
		data := t.Serialize()
		h, err := pgs.NodeStore().WriteBytes(ctx, data)
		if err != nil {
			return err
		}
		if err = mapEditor.Update(ctx, string(t.ID), h); err != nil {
			return err
		}
	}
	// Set the contents only after the error check. Flush returns a
	// zero AddressMap on failure, which would corrupt the TypeCollection.
	flushed, err := mapEditor.Flush(ctx)
	if err != nil {
		return err
	}
	pgs.SetContents(flushed)
	clear(pgs.accessedMap)
	return nil
}

// getTable returns the SQL table that matches the given schema and table name. Returns a nil table if one is not found.
// This is intended for use with tableToType.
func (*TypeCollection) getTable(ctx *sql.Context, schema string, tblName string) (tbl sql.Table, actualSchema string, err error) {
	// With no schema to search there is no matching table, which is a miss rather than a failure: callers can still
	// resolve an unqualified name by other means.
	// TODO: an unqualified name should be resolved against every schema on the search path, not just the first.
	actualSchema, ok, err := LookupSchemaName(ctx, nil, schema)
	if err != nil || !ok {
		return nil, "", err
	}
	tbl, err = GetSqlTableFromContext(ctx, "", doltdb.TableName{
		Name:   tblName,
		Schema: actualSchema,
	})
	if err != nil || tbl == nil {
		return nil, "", err
	}
	if schTbl, ok := tbl.(sql.DatabaseSchemaTable); ok {
		actualSchema = schTbl.DatabaseSchema().SchemaName()
	}
	return tbl, actualSchema, nil
}

// tableToType handles type creation related to a table's composite row type.
// https://www.postgresql.org/docs/15/sql-createtable.html
func (pgs *TypeCollection) tableToType(ctx *sql.Context, tbl sql.Table, schema string) (*pgtypes.DoltgresType, error) {
	tblName := tbl.Name()
	tblSch := tbl.Schema(ctx)
	typeID := id.NewType(schema, tblName)
	relID := id.NewTable(schema, tblName).AsId()
	arrayID := id.NewType(schema, "_"+tblName)
	attrs := make([]pgtypes.CompositeAttribute, len(tblSch))
	for i, col := range tblSch {
		collation := "" // TODO: what should we use for the collation?
		colType, ok := col.Type.(*pgtypes.DoltgresType)
		if !ok {
			// TODO: perhaps we should use a better error message stating that it uses a non-Doltgres type?
			return nil, pgtypes.ErrTypeDoesNotExist.New(tblName)
		}
		attrs[i] = pgtypes.NewCompositeAttribute(ctx, relID, col.Name, colType, int16(i+1), collation)
	}
	tableType := pgtypes.NewCompositeType(ctx, relID, pgtypes.NewUnresolvedDoltgresTypeFromID(arrayID), typeID, attrs)
	_ = pgtypes.CreateArrayTypeFromBaseType(tableType) // This sets the tableType's `Array` field as well
	return tableType, nil
}

// GetSqlTableFromContext is a forward declaration to get around import cycles
var GetSqlTableFromContext func(ctx *sql.Context, databaseName string, tableName doltdb.TableName) (sql.Table, error)

// LookupSchemaName is a forward declaration to get around import cycles
var LookupSchemaName func(ctx *sql.Context, db sql.Database, schemaName string) (string, bool, error)
