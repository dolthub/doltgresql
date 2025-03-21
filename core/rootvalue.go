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

package core

import (
	"bytes"
	"context"
	"sort"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb/durable"
	"github.com/dolthub/dolt/go/libraries/doltcore/schema"
	"github.com/dolthub/dolt/go/store/hash"
	"github.com/dolthub/dolt/go/store/prolly/tree"
	"github.com/dolthub/dolt/go/store/types"

	"github.com/dolthub/doltgresql/core/functions"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/rootobject"
	"github.com/dolthub/doltgresql/core/sequences"
	"github.com/dolthub/doltgresql/core/storage"
	"github.com/dolthub/doltgresql/core/typecollection"
)

const (
	ddbRootStructName = "dolt_db_root"
	tablesKey         = "tables"
	foreignKeyKey     = "foreign_key"
	featureVersKey    = "feature_ver"
)

// DoltgresFeatureVersion is Doltgres' feature version. We use Dolt's feature version added to our own.
var DoltgresFeatureVersion = doltdb.DoltFeatureVersion + 0

// RootValue is Doltgres' implementation of doltdb.RootValue.
type RootValue struct {
	vrw  types.ValueReadWriter
	ns   tree.NodeStore
	st   storage.RootStorage
	fkc  *doltdb.ForeignKeyCollection // cache the first load
	hash hash.Hash                    // cache the first load
}

var _ doltdb.RootValue = (*RootValue)(nil)
var _ rootobject.RootValue = (*RootValue)(nil)

// CreateDatabaseSchema implements the interface doltdb.RootValue.
func (root *RootValue) CreateDatabaseSchema(ctx context.Context, dbSchema schema.DatabaseSchema) (doltdb.RootValue, error) {
	existingSchemas, err := root.st.GetSchemas(ctx)
	if err != nil {
		return nil, err
	}

	for _, s := range existingSchemas {
		if strings.EqualFold(s.Name, dbSchema.Name) {
			return nil, errors.Errorf("A schema with the name %s already exists", dbSchema.Name)
		}
	}

	existingSchemas = append(existingSchemas, dbSchema)
	sort.Slice(existingSchemas, func(i, j int) bool {
		return existingSchemas[i].Name < existingSchemas[j].Name
	})

	r, err := root.st.SetSchemas(ctx, existingSchemas)
	if err != nil {
		return nil, err
	}

	return root.withStorage(r), nil
}

func (root *RootValue) TableListHash() uint64 {
	return 0
}

// DebugString implements the interface doltdb.RootValue.
func (root *RootValue) DebugString(ctx context.Context, transitive bool) string {
	var buf bytes.Buffer
	buf.WriteString(root.st.DebugString(ctx))

	if transitive {
		buf.WriteString("\nTables:")
		root.IterTables(ctx, func(name doltdb.TableName, table *doltdb.Table, sch schema.Schema) (stop bool, err error) {
			buf.WriteString("\nTable ")
			buf.WriteString(name.Name)
			buf.WriteString(":\n")

			buf.WriteString(table.DebugString(ctx, root.ns))

			return false, nil
		})

		buf.WriteString("\nSchemas:")
		schemas, err := root.GetDatabaseSchemas(ctx)
		if err != nil {
			return ""
		}

		for _, schema := range schemas {
			buf.WriteString("\nSchema ")
			buf.WriteString(schema.Name)
		}

		fkc, err := root.GetForeignKeyCollection(ctx)
		if err == nil && fkc.Count() > 0 {
			buf.WriteString("\nForeign Keys:")
			fkc.Iter(func(fk doltdb.ForeignKey) (stop bool, err error) {
				buf.WriteString("\n")
				buf.WriteString(fk.Name)
				buf.WriteString(": ")
				buf.WriteString(fk.TableName.String())
				buf.WriteString("(")
				for i, tag := range fk.ReferencedTableColumns {
					if i > 0 {
						buf.WriteString(",")
					}
					buf.WriteString(strconv.Itoa(int(tag)))
				}
				buf.WriteString(") ON ")
				buf.WriteString(fk.ReferencedTableName.String())
				buf.WriteString("(")
				for i, tag := range fk.ReferencedTableColumns {
					if i > 0 {
						buf.WriteString(",")
					}
					buf.WriteString(strconv.Itoa(int(tag)))
				}
				buf.WriteString(")\n")
				return false, nil
			})
		}
	}

	return buf.String()
}

// FindRootObjectNames implements the interface doltdb.RootValue.
func (root *RootValue) FindRootObjectNames(ctx context.Context, names []doltdb.TableName) (map[doltdb.TableName]struct{}, []doltdb.TableName, error) {
	nameMap := make(map[doltdb.TableName]struct{})
	var returnNames []doltdb.TableName
	for _, name := range names {
		_, _, objID, err := rootobject.ResolveName(ctx, root, name)
		if err != nil {
			return nil, nil, err
		}
		if objID != rootobject.RootObjectID_None {
			nameMap[name] = struct{}{}
			returnNames = append(returnNames, name)
		}
	}
	return nameMap, returnNames, nil
}

// GetCollation implements the interface doltdb.RootValue.
func (root *RootValue) GetCollation(ctx context.Context) (schema.Collation, error) {
	return root.st.GetCollation(ctx)
}

// GetCollection returns the root object collection that matches the given ID.
func (root *RootValue) GetCollection(ctx context.Context, rootObjID rootobject.RootObjectID) (rootobject.Collection, error) {
	return rootobject.LoadCollection(ctx, root, rootObjID)
}

// GetRootObject implements the interface doltdb.RootValue.
func (root *RootValue) GetRootObject(ctx context.Context, tName doltdb.TableName) (doltdb.RootObject, bool, error) {
	return rootobject.GetRootObject(ctx, root, tName)
}

// GetDatabaseSchemas implements the interface doltdb.RootValue.
func (root *RootValue) GetDatabaseSchemas(ctx context.Context) ([]schema.DatabaseSchema, error) {
	existingSchemas, err := root.st.GetSchemas(ctx)
	if err != nil {
		return nil, err
	}

	return existingSchemas, nil
}

// GetFeatureVersion implements the interface doltdb.RootValue.
func (root *RootValue) GetFeatureVersion(ctx context.Context) (ver doltdb.FeatureVersion, ok bool, err error) {
	return root.st.GetFeatureVersion(), true, nil
}

// GetForeignKeyCollection implements the interface doltdb.RootValue.
func (root *RootValue) GetForeignKeyCollection(ctx context.Context) (*doltdb.ForeignKeyCollection, error) {
	if root.fkc == nil {
		fkMap, ok, err := root.st.GetForeignKeys(ctx, root.vrw)
		if err != nil {
			return nil, err
		}
		if !ok {
			return doltdb.NewForeignKeyCollection()
		}

		root.fkc, err = doltdb.DeserializeForeignKeys(ctx, root.vrw.Format(), fkMap)
		if err != nil {
			return nil, err
		}
	}
	return root.fkc.Copy(), nil
}

// GetFunctions returns all functions that are on the root.
func (root *RootValue) GetFunctions(ctx context.Context) (*functions.Collection, error) {
	coll, err := rootobject.LoadCollection(ctx, root, rootobject.RootObjectID_Functions)
	if err != nil {
		return nil, err
	}
	return coll.(*functions.Collection), nil
}

// GetSequences returns all sequences that are on the root.
func (root *RootValue) GetSequences(ctx context.Context) (*sequences.Collection, error) {
	coll, err := rootobject.LoadCollection(ctx, root, rootobject.RootObjectID_Sequences)
	if err != nil {
		return nil, err
	}
	return coll.(*sequences.Collection), nil
}

// GetStorage returns the underlying storage.
func (root *RootValue) GetStorage(ctx context.Context) storage.RootStorage {
	return root.st
}

// GetTable implements the interface doltdb.RootValue.
func (root *RootValue) GetTable(ctx context.Context, tName doltdb.TableName) (*doltdb.Table, bool, error) {
	tableMap, err := root.getTableMap(ctx, tName.Schema)
	if err != nil {
		return nil, false, err
	}

	addr, err := tableMap.Get(ctx, tName.Name)
	if err != nil {
		return nil, false, err
	}

	return doltdb.GetTable(ctx, root, addr)
}

// GetTableHash implements the interface doltdb.RootValue.
func (root *RootValue) GetTableHash(ctx context.Context, tName doltdb.TableName) (hash.Hash, bool, error) {
	_, rawID, objID, err := rootobject.ResolveName(ctx, root, tName)
	if err != nil {
		return hash.Hash{}, false, err
	}
	if objID == rootobject.RootObjectID_None {
		tableMap, err := root.getTableMap(ctx, tName.Schema)
		if err != nil {
			return hash.Hash{}, false, err
		}

		tVal, err := tableMap.Get(ctx, tName.Name)
		if err != nil {
			return hash.Hash{}, false, err
		}

		return tVal, !tVal.IsEmpty(), nil
	} else {
		coll, err := rootobject.LoadCollection(ctx, root, objID)
		if err != nil {
			return hash.Hash{}, false, err
		}
		obj, ok, err := coll.GetRootObject(ctx, rawID)
		if err != nil || !ok {
			return hash.Hash{}, false, err
		}
		h, err := obj.HashOf(ctx)
		return h, err == nil && !h.IsEmpty(), err
	}
}

// GetTableNames implements the interface doltdb.RootValue.
func (root *RootValue) GetTableNames(ctx context.Context, schemaName string) ([]string, error) {
	tableMap, err := root.getTableMap(ctx, schemaName)
	if err != nil {
		return nil, err
	}

	var names []string
	err = tableMap.Iter(ctx, func(name string, _ hash.Hash) (bool, error) {
		names = append(names, name)
		return false, nil
	})
	if err != nil {
		return nil, err
	}
	// Iterate collections
	colls, err := rootobject.LoadAllCollections(ctx, root)
	if err != nil {
		return nil, err
	}
	for _, coll := range colls {
		err = coll.IterateIDs(ctx, func(identifier id.Id) (stop bool, err error) {
			tName := coll.IDToTableName(identifier)
			if tName.Schema == schemaName {
				names = append(names, tName.Name)
			}
			return false, nil
		})
		if err != nil {
			return nil, err
		}
	}
	return names, nil
}

// GetTableSchemaHash implements the interface doltdb.RootValue.
func (root *RootValue) GetTableSchemaHash(ctx context.Context, tName doltdb.TableName) (hash.Hash, error) {
	tab, ok, err := root.GetTable(ctx, tName)
	if err != nil {
		return hash.Hash{}, err
	}
	if !ok {
		return hash.Hash{}, nil
	}
	return tab.GetSchemaHash(ctx)
}

// GetTypes returns all types that are on the root.
func (root *RootValue) GetTypes(ctx context.Context) (*typecollection.TypeCollection, error) {
	coll, err := rootobject.LoadCollection(ctx, root, rootobject.RootObjectID_Types)
	if err != nil {
		return nil, err
	}
	return coll.(*typecollection.TypeCollection), nil
}

// HandlePostMerge implements the interface doltdb.RootValue.
func (root *RootValue) HandlePostMerge(ctx context.Context, ourRoot, theirRoot, ancRoot doltdb.RootValue) (doltdb.RootValue, error) {
	// Handle sequences
	_, err := root.handlePostSequencesMerge(ctx, ourRoot, theirRoot, ancRoot)
	if err != nil {
		return nil, err
	}
	// Handle types
	_, err = root.handlePostTypesMerge(ctx, ourRoot, theirRoot, ancRoot)
	if err != nil {
		return nil, err
	}
	// Handle functions
	return root.handlePostFunctionsMerge(ctx, ourRoot, theirRoot, ancRoot)
}

// handlePostFunctionsMerge merges functions.
func (root *RootValue) handlePostFunctionsMerge(ctx context.Context, ourRoot, theirRoot, ancRoot doltdb.RootValue) (doltdb.RootValue, error) {
	ourFunctions, err := ourRoot.(*RootValue).GetFunctions(ctx)
	if err != nil {
		return nil, err
	}
	theirFunctions, err := theirRoot.(*RootValue).GetFunctions(ctx)
	if err != nil {
		return nil, err
	}
	ancFunctions, err := ancRoot.(*RootValue).GetFunctions(ctx)
	if err != nil {
		return nil, err
	}
	mergedFunctions, err := functions.Merge(ctx, ourFunctions, theirFunctions, ancFunctions)
	if err != nil {
		return nil, err
	}
	return root.PutFunctions(ctx, mergedFunctions)
}

// handlePostSequencesMerge merges sequences.
func (root *RootValue) handlePostSequencesMerge(ctx context.Context, ourRoot, theirRoot, ancRoot doltdb.RootValue) (doltdb.RootValue, error) {
	ourSequence, err := ourRoot.(*RootValue).GetSequences(ctx)
	if err != nil {
		return nil, err
	}
	theirSequence, err := theirRoot.(*RootValue).GetSequences(ctx)
	if err != nil {
		return nil, err
	}
	ancSequence, err := ancRoot.(*RootValue).GetSequences(ctx)
	if err != nil {
		return nil, err
	}
	mergedSequence, err := sequences.Merge(ctx, ourSequence, theirSequence, ancSequence)
	if err != nil {
		return nil, err
	}
	return root.PutSequences(ctx, mergedSequence)
}

// handlePostTypesMerge merges types.
func (root *RootValue) handlePostTypesMerge(ctx context.Context, ourRoot, theirRoot, ancRoot doltdb.RootValue) (doltdb.RootValue, error) {
	ourTypes, err := ourRoot.(*RootValue).GetTypes(ctx)
	if err != nil {
		return nil, err
	}
	theirTypes, err := theirRoot.(*RootValue).GetTypes(ctx)
	if err != nil {
		return nil, err
	}
	ancTypes, err := ancRoot.(*RootValue).GetTypes(ctx)
	if err != nil {
		return nil, err
	}
	mergedTypes, err := typecollection.Merge(ctx, ourTypes, theirTypes, ancTypes)
	if err != nil {
		return nil, err
	}
	return root.PutTypes(ctx, mergedTypes)
}

// HashOf implements the interface doltdb.RootValue.
func (root *RootValue) HashOf() (hash.Hash, error) {
	if root.hash.IsEmpty() {
		var err error
		root.hash, err = root.st.NomsValue().Hash(root.vrw.Format())
		if err != nil {
			return hash.Hash{}, nil
		}
	}
	return root.hash, nil
}

// HasTable implements the interface doltdb.RootValue.
func (root *RootValue) HasTable(ctx context.Context, tName doltdb.TableName) (bool, error) {
	_, _, objID, err := rootobject.ResolveName(ctx, root, tName)
	if err != nil {
		return false, err
	}
	if objID == rootobject.RootObjectID_None {
		tableMap, err := root.st.GetTablesMap(ctx, root.vrw, root.ns, tName.Schema)
		if err != nil {
			return false, err
		}
		a, err := tableMap.Get(ctx, tName.Name)
		if err != nil {
			return false, err
		}
		return !a.IsEmpty(), nil
	} else {
		return true, nil
	}
}

// IterRootObjects implements the interface doltdb.RootValue.
func (root *RootValue) IterRootObjects(ctx context.Context, cb func(name doltdb.TableName, table doltdb.RootObject) (stop bool, err error)) error {
	colls, err := rootobject.LoadAllCollections(ctx, root)
	if err != nil {
		return err
	}
	for _, coll := range colls {
		err = coll.IterateRootObjects(ctx, func(rootObj rootobject.RootObject) (stop bool, err error) {
			return cb(rootObj.Name(), rootObj)
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// IterTables implements the interface doltdb.RootValue.
func (root *RootValue) IterTables(ctx context.Context, cb func(name doltdb.TableName, table *doltdb.Table, sch schema.Schema) (stop bool, err error)) error {
	schemaNames, err := schemaNames(ctx, root)
	if err != nil {
		return err
	}

	for _, schemaName := range schemaNames {
		tm, err := root.getTableMap(ctx, schemaName)
		if err != nil {
			return err
		}

		err = tm.Iter(ctx, func(name string, addr hash.Hash) (bool, error) {
			nt, err := durable.TableFromAddr(ctx, root.VRW(), root.ns, addr)
			if err != nil {
				return true, err
			}
			tbl := doltdb.NewTableFromDurable(nt)

			sch, err := tbl.GetSchema(ctx)
			if err != nil {
				return true, err
			}

			return cb(doltdb.TableName{
				Name:   name,
				Schema: schemaName,
			}, tbl, sch)
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// schemaNames returns all names of all schemas which may have tables
func schemaNames(ctx context.Context, root doltdb.RootValue) ([]string, error) {
	dbSchemas, err := root.GetDatabaseSchemas(ctx)
	if err != nil {
		return nil, err
	}

	schNames := make([]string, len(dbSchemas)+1)
	for i, dbSchema := range dbSchemas {
		schNames[i] = dbSchema.Name
	}
	return schNames, nil
}

// NodeStore implements the interface doltdb.RootValue.
func (root *RootValue) NodeStore() tree.NodeStore {
	return root.ns
}

// NomsValue implements the interface doltdb.RootValue.
func (root *RootValue) NomsValue() types.Value {
	return root.st.NomsValue()
}

// PutCollection updates the root value with the given collection.
func (root *RootValue) PutCollection(ctx context.Context, coll rootobject.Collection) (doltdb.RootValue, error) {
	if coll == nil {
		return root, nil
	}
	return rootobject.PutCollection(ctx, root, coll)
}

// PutForeignKeyCollection implements the interface doltdb.RootValue.
func (root *RootValue) PutForeignKeyCollection(ctx context.Context, fkc *doltdb.ForeignKeyCollection) (doltdb.RootValue, error) {
	value, err := doltdb.SerializeForeignKeys(ctx, root.vrw, fkc)
	if err != nil {
		return nil, err
	}
	newStorage, err := root.st.SetForeignKeyMap(ctx, root.vrw, value)
	if err != nil {
		return nil, err
	}
	return root.withStorage(newStorage), nil
}

// PutFunctions writes the given functions to the returned root value.
func (root *RootValue) PutFunctions(ctx context.Context, funcs *functions.Collection) (*RootValue, error) {
	if funcs == nil {
		return root, nil
	}
	newRoot, err := rootobject.PutCollection(ctx, root, funcs)
	if err != nil {
		return nil, err
	}
	return newRoot.(*RootValue), nil
}

// PutRootObject implements the interface doltdb.RootValue.
func (root *RootValue) PutRootObject(ctx context.Context, tName doltdb.TableName, rootObj doltdb.RootObject) (doltdb.RootValue, error) {
	if rootObj == nil {
		return root, nil
	}
	return rootobject.PutRootObject(ctx, root, tName, rootObj.(rootobject.RootObject))
}

// PutSequences writes the given sequences to the returned root value.
func (root *RootValue) PutSequences(ctx context.Context, seq *sequences.Collection) (*RootValue, error) {
	if seq == nil {
		return root, nil
	}
	newRoot, err := rootobject.PutCollection(ctx, root, seq)
	if err != nil {
		return nil, err
	}
	return newRoot.(*RootValue), nil
}

// PutTable implements the interface doltdb.RootValue.
func (root *RootValue) PutTable(ctx context.Context, tName doltdb.TableName, table *doltdb.Table) (doltdb.RootValue, error) {
	// TODO: modify owned sequences based on schema changes
	err := doltdb.ValidateTagUniqueness(ctx, root, tName.Name, table)
	if err != nil {
		return nil, err
	}

	tableRef, err := doltdb.RefFromNomsTable(ctx, table)
	if err != nil {
		return nil, err
	}

	return root.putTable(ctx, tName, tableRef)
}

// PutTypes writes the given types to the returned root value.
func (root *RootValue) PutTypes(ctx context.Context, typ *typecollection.TypeCollection) (*RootValue, error) {
	if typ == nil {
		return root, nil
	}
	newRoot, err := rootobject.PutCollection(ctx, root, typ)
	if err != nil {
		return nil, err
	}
	return newRoot.(*RootValue), nil
}

// RemoveTables implements the interface doltdb.RootValue.
func (root *RootValue) RemoveTables(
	ctx context.Context,
	skipFKHandling bool,
	allowDroppingFKReferenced bool,
	originalTables ...doltdb.TableName,
) (doltdb.RootValue, error) {
	if len(originalTables) == 0 {
		return root, nil
	}

	var tables []doltdb.TableName
	var rootObjNames []struct {
		rawID id.Id
		objID rootobject.RootObjectID
	}
	for _, name := range originalTables {
		_, rawID, objID, err := rootobject.ResolveName(ctx, root, name)
		if err != nil {
			return nil, err
		}
		if objID == rootobject.RootObjectID_None {
			tables = append(tables, name)
		} else {
			rootObjNames = append(rootObjNames, struct {
				rawID id.Id
				objID rootobject.RootObjectID
			}{rawID: rawID, objID: objID})
		}
	}
	newRoot := root

	// First we'll handle regular table names
	if len(tables) > 0 {
		// TODO: support multiple schemas in the same set
		tableMap, err := newRoot.getTableMap(ctx, tables[0].Schema)
		if err != nil {
			return nil, err
		}

		edits := make([]storage.TableEdit, len(tables))
		for i, name := range tables {
			a, err := tableMap.Get(ctx, name.Name)
			if err != nil {
				return nil, err
			}
			if a.IsEmpty() {
				return nil, errors.Errorf("%w: '%s'", doltdb.ErrTableNotFound, name)
			}
			edits[i].Name = name
		}

		newStorage, err := newRoot.st.EditTablesMap(ctx, newRoot.vrw, newRoot.ns, edits)
		if err != nil {
			return nil, err
		}
		newRoot = newRoot.withStorage(newStorage)

		collection, err := newRoot.GetSequences(ctx)
		if err != nil {
			return nil, err
		}
		for _, tableName := range tables {
			seqs, err := collection.GetSequencesWithTable(ctx, tableName)
			if err != nil {
				return nil, err
			}
			if len(seqs) > 0 {
				for _, seq := range seqs {
					if err = collection.DropSequence(ctx, seq.Id); err != nil {
						return nil, err
					}
				}
			} else {

			}
		}
		newRoot, err = newRoot.PutSequences(ctx, collection)
		if err != nil {
			return nil, err
		}

		if skipFKHandling {
			return newRoot, nil
		}
		fkc, err := newRoot.GetForeignKeyCollection(ctx)
		if err != nil {
			return nil, err
		}
		if allowDroppingFKReferenced {
			err = fkc.RemoveAndUnresolveTables(ctx, newRoot, tables...)
		} else {
			err = fkc.RemoveTables(ctx, tables...)
		}
		if err != nil {
			return nil, err
		}
		newRootInterface, err := newRoot.PutForeignKeyCollection(ctx, fkc)
		if err != nil {
			return nil, err
		}
		newRoot = newRootInterface.(*RootValue)
	}

	// Then we'll handle root objects
	for _, rootObjName := range rootObjNames {
		newRootInt, err := rootobject.RemoveRootObject(ctx, newRoot, rootObjName.rawID, rootObjName.objID)
		if err != nil {
			return nil, err
		}
		newRoot = newRootInt.(*RootValue)
	}

	return newRoot, nil
}

// RenameTable implements the interface doltdb.RootValue.
func (root *RootValue) RenameTable(ctx context.Context, oldName, newName doltdb.TableName) (doltdb.RootValue, error) {
	_, rawOldID, objID, err := rootobject.ResolveName(ctx, root, oldName)
	if err != nil {
		return nil, err
	}
	if objID == rootobject.RootObjectID_None {
		newStorage, err := root.st.EditTablesMap(ctx, root.vrw, root.ns, []storage.TableEdit{{OldName: oldName, Name: newName}})
		if err != nil {
			return nil, err
		}
		newRoot := root.withStorage(newStorage)

		collection, err := newRoot.GetSequences(ctx)
		if err != nil {
			return nil, err
		}
		seqs, err := collection.GetSequencesWithTable(ctx, oldName)
		if err != nil {
			return nil, err
		}
		for _, seq := range seqs {
			seq.OwnerTable = id.NewTable(seq.OwnerTable.SchemaName(), newName.Name)
		}
		newRoot, err = newRoot.PutSequences(ctx, collection)
		if err != nil {
			return nil, err
		}

		return newRoot, nil
	} else {
		coll, err := rootobject.LoadCollection(ctx, root, objID)
		if err != nil {
			return nil, err
		}
		rawNewID := coll.TableNameToID(newName)
		if err = coll.RenameRootObject(ctx, rawOldID, rawNewID); err != nil {
			return nil, err
		}
		return rootobject.PutCollection(ctx, root, coll)
	}
}

// ResolveRootValue implements the interface doltdb.RootValue.
func (root *RootValue) ResolveRootValue(ctx context.Context) (doltdb.RootValue, error) {
	return root, nil
}

// ResolveTableName implements the interface doltdb.RootValue.
func (root *RootValue) ResolveTableName(ctx context.Context, tName doltdb.TableName) (string, bool, error) {
	resolvedTableName, _, objID, err := rootobject.ResolveName(ctx, root, tName)
	if err != nil {
		return "", false, err
	}
	if objID == rootobject.RootObjectID_None {
		tableMap, err := root.getTableMap(ctx, tName.Schema)
		if err != nil {
			return "", false, err
		}

		a, err := tableMap.Get(ctx, tName.Name)
		if err != nil {
			return "", false, err
		}
		if !a.IsEmpty() {
			return tName.Name, true, nil
		}

		found := false
		resolvedName := tName.Name
		err = tableMap.Iter(ctx, func(name string, addr hash.Hash) (bool, error) {
			if !found && strings.EqualFold(tName.Name, name) {
				resolvedName = name
				found = true
			}
			return false, nil
		})
		if err != nil {
			return "", false, nil
		}
		return resolvedName, found, nil
	} else {
		return resolvedTableName.Name, true, nil
	}
}

// SetCollation implements the interface doltdb.RootValue.
func (root *RootValue) SetCollation(ctx context.Context, collation schema.Collation) (doltdb.RootValue, error) {
	newStorage, err := root.st.SetCollation(ctx, collation)
	if err != nil {
		return nil, err
	}
	return root.withStorage(newStorage), nil
}

// SetFeatureVersion implements the interface doltdb.RootValue.
func (root *RootValue) SetFeatureVersion(v doltdb.FeatureVersion) (doltdb.RootValue, error) {
	newStorage, err := root.st.SetFeatureVersion(v)
	if err != nil {
		return nil, err
	}
	return root.withStorage(newStorage), nil
}

// SetTableHash implements the interface doltdb.RootValue.
func (root *RootValue) SetTableHash(ctx context.Context, tName doltdb.TableName, h hash.Hash) (doltdb.RootValue, error) {
	// TODO: error for root object tables?
	val, err := root.vrw.ReadValue(ctx, h)

	if err != nil {
		return nil, err
	}

	ref, err := types.NewRef(val, root.vrw.Format())

	if err != nil {
		return nil, err
	}

	return root.putTable(ctx, tName, ref)
}

// VRW implements the interface doltdb.RootValue.
func (root *RootValue) VRW() types.ValueReadWriter {
	return root.vrw
}

// WithStorage returns a new root value with the given storage.
func (root *RootValue) WithStorage(ctx context.Context, st storage.RootStorage) rootobject.RootValue {
	return root.withStorage(st)
}

// getTableMap returns the tableMap for this root.
func (root *RootValue) getTableMap(ctx context.Context, schemaName string) (storage.RootTableMap, error) {
	if schemaName == "" {
		schemaName = doltdb.DefaultSchemaName
	}
	return root.st.GetTablesMap(ctx, root.vrw, root.ns, schemaName)
}

// putTable provides an inner implementation that is called from multiple other functions.
func (root *RootValue) putTable(ctx context.Context, tName doltdb.TableName, ref types.Ref) (doltdb.RootValue, error) {
	if !doltdb.IsValidTableName(tName.Name) {
		panic("Don't attempt to put a table with a name that fails the IsValidTableName check")
	}

	newStorage, err := root.st.EditTablesMap(ctx, root.VRW(), root.NodeStore(), []storage.TableEdit{{Name: tName, Ref: &ref}})
	if err != nil {
		return nil, err
	}

	return root.withStorage(newStorage), nil
}

// withStorage returns a new root value with the given storage.
func (root *RootValue) withStorage(st storage.RootStorage) *RootValue {
	return &RootValue{root.vrw, root.ns, st, nil, hash.Hash{}}
}
