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
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb/durable"
	"github.com/dolthub/dolt/go/libraries/doltcore/schema"
	"github.com/dolthub/dolt/go/store/hash"
	"github.com/dolthub/dolt/go/store/prolly/tree"
	"github.com/dolthub/dolt/go/store/types"

	"github.com/dolthub/doltgresql/core/sequences"
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
	st   rootStorage
	fkc  *doltdb.ForeignKeyCollection // cache the first load
	hash hash.Hash                    // cache the first load
}

var _ doltdb.RootValue = (*RootValue)(nil)

type tableEdit struct {
	name doltdb.TableName
	ref  *types.Ref

	// Used for rename.
	old_name doltdb.TableName
}

// CreateDatabaseSchema implements the interface doltdb.RootValue.
func (root *RootValue) CreateDatabaseSchema(ctx context.Context, dbSchema schema.DatabaseSchema) (doltdb.RootValue, error) {
	existingSchemas, err := root.st.GetSchemas(ctx)
	if err != nil {
		return nil, err
	}

	for _, s := range existingSchemas {
		if strings.EqualFold(s.Name, dbSchema.Name) {
			return nil, fmt.Errorf("A schema with the name %s already exists", dbSchema.Name)
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

// GetCollation implements the interface doltdb.RootValue.
func (root *RootValue) GetCollation(ctx context.Context) (schema.Collation, error) {
	return root.st.GetCollation(ctx)
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

// GetSequences returns all sequences that are on the root.
func (root *RootValue) GetSequences(ctx context.Context) (*sequences.Collection, error) {
	h := root.st.GetSequences()
	if h.IsEmpty() {
		return sequences.Deserialize(ctx, nil)
	}
	dataValue, err := root.vrw.ReadValue(ctx, h)
	if err != nil {
		return nil, err
	}
	dataBlob := dataValue.(types.Blob)
	dataBlobLength := dataBlob.Len()
	data := make([]byte, dataBlobLength)
	n, err := dataBlob.ReadAt(context.Background(), data, 0)
	if err != nil && err != io.EOF {
		return nil, err
	}
	if uint64(n) != dataBlobLength {
		return nil, fmt.Errorf("wanted %d bytes from blob for sequences, got %d", dataBlobLength, n)
	}
	return sequences.Deserialize(ctx, data)
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
	tableMap, err := root.getTableMap(ctx, tName.Schema)
	if err != nil {
		return hash.Hash{}, false, err
	}

	tVal, err := tableMap.Get(ctx, tName.Name)
	if err != nil {
		return hash.Hash{}, false, err
	}

	return tVal, !tVal.IsEmpty(), nil
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

	return names, nil
}

// HandlePostMerge implements the interface doltdb.RootValue.
func (root *RootValue) HandlePostMerge(ctx context.Context, ourRoot, theirRoot, ancRoot doltdb.RootValue) (doltdb.RootValue, error) {
	// Handle sequences
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

// HashOf implements the interface doltdb.RootValue.
func (root *RootValue) HashOf() (hash.Hash, error) {
	if root.hash.IsEmpty() {
		var err error
		root.hash, err = root.st.nomsValue().Hash(root.vrw.Format())
		if err != nil {
			return hash.Hash{}, nil
		}
	}
	return root.hash, nil
}

// HasTable implements the interface doltdb.RootValue.
func (root *RootValue) HasTable(ctx context.Context, tName doltdb.TableName) (bool, error) {
	tableMap, err := root.st.GetTablesMap(ctx, root.vrw, root.ns, tName.Schema)
	if err != nil {
		return false, err
	}
	a, err := tableMap.Get(ctx, tName.Name)
	if err != nil {
		return false, err
	}
	return !a.IsEmpty(), nil
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

	schemaNames := make([]string, len(dbSchemas)+1)
	for i, dbSchema := range dbSchemas {
		schemaNames[i] = dbSchema.Name
	}
	return schemaNames, nil
}

// NodeStore implements the interface doltdb.RootValue.
func (root *RootValue) NodeStore() tree.NodeStore {
	return root.ns
}

// NomsValue implements the interface doltdb.RootValue.
func (root *RootValue) NomsValue() types.Value {
	return root.st.nomsValue()
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

// PutSequences writes the given sequences to the returned root value.
func (root *RootValue) PutSequences(ctx context.Context, seq *sequences.Collection) (*RootValue, error) {
	data, err := seq.Serialize(ctx)
	if err != nil {
		return nil, err
	}
	dataBlob, err := types.NewBlob(ctx, root.vrw, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	ref, err := root.vrw.WriteValue(ctx, dataBlob)
	if err != nil {
		return nil, err
	}
	newStorage, err := root.st.SetSequences(ctx, ref.TargetHash())
	if err != nil {
		return nil, err
	}
	return root.withStorage(newStorage), nil
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

// RemoveTables implements the interface doltdb.RootValue.
func (root *RootValue) RemoveTables(
	ctx context.Context,
	skipFKHandling bool,
	allowDroppingFKReferenced bool,
	tables ...doltdb.TableName,
) (doltdb.RootValue, error) {
	if len(tables) == 0 {
		return root, nil
	}

	// TODO: support multiple schemas in the same set
	tableMap, err := root.getTableMap(ctx, tables[0].Schema)
	if err != nil {
		return nil, err
	}

	edits := make([]tableEdit, len(tables))
	for i, name := range tables {
		a, err := tableMap.Get(ctx, name.Name)
		if err != nil {
			return nil, err
		}
		if a.IsEmpty() {
			return nil, fmt.Errorf("%w: '%s'", doltdb.ErrTableNotFound, name)
		}
		edits[i].name = name
	}

	newStorage, err := root.st.EditTablesMap(ctx, root.vrw, root.ns, edits)
	if err != nil {
		return nil, err
	}
	newRoot := root.withStorage(newStorage)

	collection, err := newRoot.GetSequences(ctx)
	if err != nil {
		return nil, err
	}
	for _, tableName := range tables {
		for _, seq := range collection.GetSequencesWithTable(tableName) {
			if err = collection.DropSequence(doltdb.TableName{Name: seq.Name, Schema: tableName.Schema}); err != nil {
				return nil, err
			}
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
		err = fkc.RemoveAndUnresolveTables(ctx, root, tables...)
	} else {
		err = fkc.RemoveTables(ctx, tables...)
	}
	if err != nil {
		return nil, err
	}

	return newRoot.PutForeignKeyCollection(ctx, fkc)
}

// RenameTable implements the interface doltdb.RootValue.
func (root *RootValue) RenameTable(ctx context.Context, oldName, newName doltdb.TableName) (doltdb.RootValue, error) {
	newStorage, err := root.st.EditTablesMap(ctx, root.vrw, root.ns, []tableEdit{{old_name: oldName, name: newName}})
	if err != nil {
		return nil, err
	}
	newRoot := root.withStorage(newStorage)

	collection, err := newRoot.GetSequences(ctx)
	if err != nil {
		return nil, err
	}
	for _, seq := range collection.GetSequencesWithTable(oldName) {
		seq.OwnerTable = newName.Name
	}
	newRoot, err = newRoot.PutSequences(ctx, collection)
	if err != nil {
		return nil, err
	}

	return newRoot, nil
}

// ResolveRootValue implements the interface doltdb.RootValue.
func (root *RootValue) ResolveRootValue(ctx context.Context) (doltdb.RootValue, error) {
	return root, nil
}

// ResolveTableName implements the interface doltdb.RootValue.
func (root *RootValue) ResolveTableName(ctx context.Context, tName doltdb.TableName) (string, bool, error) {
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

// getTableMap returns the tableMap for this root.
func (root *RootValue) getTableMap(ctx context.Context, schemaName string) (rootTableMap, error) {
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

	newStorage, err := root.st.EditTablesMap(ctx, root.VRW(), root.NodeStore(), []tableEdit{{name: tName, ref: &ref}})
	if err != nil {
		return nil, err
	}

	return root.withStorage(newStorage), nil
}

// withStorage returns a new root value with the given storage.
func (root *RootValue) withStorage(st rootStorage) *RootValue {
	return &RootValue{root.vrw, root.ns, st, nil, hash.Hash{}}
}
