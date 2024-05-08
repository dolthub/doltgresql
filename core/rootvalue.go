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
	"sort"
	"strings"

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb/durable"
	"github.com/dolthub/dolt/go/libraries/doltcore/schema"
	"github.com/dolthub/dolt/go/store/hash"
	"github.com/dolthub/dolt/go/store/prolly/tree"
	"github.com/dolthub/dolt/go/store/types"
)

const (
	ddbRootStructName = "dolt_db_root"
	tablesKey         = "tables"
	foreignKeyKey     = "foreign_key"
	featureVersKey    = "feature_ver"
)

// DoltgresFeatureVersion is Doltgres' feature version. We use Dolt's feature version added to our own.
var DoltgresFeatureVersion = doltdb.DoltFeatureVersion + 0

// rootValue is Dolt's implementation of RootValue.
type rootValue struct {
	vrw  types.ValueReadWriter
	ns   tree.NodeStore
	st   rootStorage
	fkc  *doltdb.ForeignKeyCollection // cache the first load
	hash hash.Hash                    // cache first load
}

type tableEdit struct {
	name doltdb.TableName
	ref  *types.Ref

	// Used for rename.
	old_name string
}

// CreateDatabaseSchema implements the interface doltdb.RootValue.
func (root *rootValue) CreateDatabaseSchema(ctx context.Context, dbSchema schema.DatabaseSchema) (doltdb.RootValue, error) {
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

// DebugString implements the interface doltdb.RootValue.
func (root *rootValue) DebugString(ctx context.Context, transitive bool) string {
	var buf bytes.Buffer
	buf.WriteString(root.st.DebugString(ctx))

	if transitive {
		buf.WriteString("\nTables:")
		root.IterTables(ctx, func(name string, table *doltdb.Table, sch schema.Schema) (stop bool, err error) {
			buf.WriteString("\nTable ")
			buf.WriteString(name)
			buf.WriteString(":\n")

			buf.WriteString(table.DebugString(ctx, root.ns))

			return false, nil
		})
	}

	return buf.String()
}

// GetCollation implements the interface doltdb.RootValue.
func (root *rootValue) GetCollation(ctx context.Context) (schema.Collation, error) {
	return root.st.GetCollation(ctx)
}

// GetDatabaseSchemas implements the interface doltdb.RootValue.
func (root *rootValue) GetDatabaseSchemas(ctx context.Context) ([]schema.DatabaseSchema, error) {
	existingSchemas, err := root.st.GetSchemas(ctx)
	if err != nil {
		return nil, err
	}

	return existingSchemas, nil
}

// GetFeatureVersion implements the interface doltdb.RootValue.
func (root *rootValue) GetFeatureVersion(ctx context.Context) (ver doltdb.FeatureVersion, ok bool, err error) {
	return root.st.GetFeatureVersion(), true, nil
}

// GetForeignKeyCollection implements the interface doltdb.RootValue.
func (root *rootValue) GetForeignKeyCollection(ctx context.Context) (*doltdb.ForeignKeyCollection, error) {
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

// GetTable implements the interface doltdb.RootValue.
func (root *rootValue) GetTable(ctx context.Context, tName doltdb.TableName) (*doltdb.Table, bool, error) {
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
func (root *rootValue) GetTableHash(ctx context.Context, tName string) (hash.Hash, bool, error) {
	tableMap, err := root.getTableMap(ctx, doltdb.DefaultSchemaName)
	if err != nil {
		return hash.Hash{}, false, err
	}

	tVal, err := tableMap.Get(ctx, tName)
	if err != nil {
		return hash.Hash{}, false, err
	}

	return tVal, !tVal.IsEmpty(), nil
}

// GetTableNames implements the interface doltdb.RootValue.
func (root *rootValue) GetTableNames(ctx context.Context, schemaName string) ([]string, error) {
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

// HashOf implements the interface doltdb.RootValue.
func (root *rootValue) HashOf() (hash.Hash, error) {
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
func (root *rootValue) HasTable(ctx context.Context, tName string) (bool, error) {
	tableMap, err := root.st.GetTablesMap(ctx, root.vrw, root.ns, doltdb.DefaultSchemaName)
	if err != nil {
		return false, err
	}
	a, err := tableMap.Get(ctx, tName)
	if err != nil {
		return false, err
	}
	return !a.IsEmpty(), nil
}

// IterTables implements the interface doltdb.RootValue.
func (root *rootValue) IterTables(ctx context.Context, cb func(name string, table *doltdb.Table, sch schema.Schema) (stop bool, err error)) error {
	// TODO: schema name
	tm, err := root.getTableMap(ctx, doltdb.DefaultSchemaName)
	if err != nil {
		return err
	}

	return tm.Iter(ctx, func(name string, addr hash.Hash) (bool, error) {
		nt, err := durable.TableFromAddr(ctx, root.VRW(), root.ns, addr)
		if err != nil {
			return true, err
		}
		tbl := doltdb.NewTableFromDurable(nt)

		sch, err := tbl.GetSchema(ctx)
		if err != nil {
			return true, err
		}

		return cb(name, tbl, sch)
	})
}

// NodeStore implements the interface doltdb.RootValue.
func (root *rootValue) NodeStore() tree.NodeStore {
	return root.ns
}

// NomsValue implements the interface doltdb.RootValue.
func (root *rootValue) NomsValue() types.Value {
	return root.st.nomsValue()
}

// PutForeignKeyCollection implements the interface doltdb.RootValue.
func (root *rootValue) PutForeignKeyCollection(ctx context.Context, fkc *doltdb.ForeignKeyCollection) (doltdb.RootValue, error) {
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

// PutTable implements the interface doltdb.RootValue.
func (root *rootValue) PutTable(ctx context.Context, tName doltdb.TableName, table *doltdb.Table) (doltdb.RootValue, error) {
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
func (root *rootValue) RemoveTables(ctx context.Context, skipFKHandling bool, allowDroppingFKReferenced bool, tables ...string) (doltdb.RootValue, error) {
	// TODO: schema name
	tableMap, err := root.getTableMap(ctx, doltdb.DefaultSchemaName)
	if err != nil {
		return nil, err
	}

	edits := make([]tableEdit, len(tables))
	for i, name := range tables {
		a, err := tableMap.Get(ctx, name)
		if err != nil {
			return nil, err
		}
		if a.IsEmpty() {
			return nil, fmt.Errorf("%w: '%s'", doltdb.ErrTableNotFound, name)
		}
		edits[i].name = doltdb.TableName{
			Name: name,
		}
	}

	newStorage, err := root.st.EditTablesMap(ctx, root.vrw, root.ns, edits)
	if err != nil {
		return nil, err
	}

	newRoot := root.withStorage(newStorage)
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
func (root *rootValue) RenameTable(ctx context.Context, oldName, newName string) (doltdb.RootValue, error) {
	newStorage, err := root.st.EditTablesMap(ctx, root.vrw, root.ns, []tableEdit{{old_name: oldName, name: doltdb.TableName{Name: newName}}})
	if err != nil {
		return nil, err
	}
	return root.withStorage(newStorage), nil
}

// ResolveRootValue implements the interface doltdb.RootValue.
func (root *rootValue) ResolveRootValue(ctx context.Context) (doltdb.RootValue, error) {
	return root, nil
}

// ResolveTableName implements the interface doltdb.RootValue.
func (root *rootValue) ResolveTableName(ctx context.Context, tName string) (string, bool, error) {
	tableMap, err := root.getTableMap(ctx, doltdb.DefaultSchemaName)
	if err != nil {
		return "", false, err
	}

	a, err := tableMap.Get(ctx, tName)
	if err != nil {
		return "", false, err
	}
	if !a.IsEmpty() {
		return tName, true, nil
	}

	found := false
	err = tableMap.Iter(ctx, func(name string, addr hash.Hash) (bool, error) {
		if !found && strings.EqualFold(tName, name) {
			tName = name
			found = true
		}
		return false, nil
	})
	if err != nil {
		return "", false, nil
	}
	return tName, found, nil
}

// SetCollation implements the interface doltdb.RootValue.
func (root *rootValue) SetCollation(ctx context.Context, collation schema.Collation) (doltdb.RootValue, error) {
	newStorage, err := root.st.SetCollation(ctx, collation)
	if err != nil {
		return nil, err
	}
	return root.withStorage(newStorage), nil
}

// SetFeatureVersion implements the interface doltdb.RootValue.
func (root *rootValue) SetFeatureVersion(v doltdb.FeatureVersion) (doltdb.RootValue, error) {
	newStorage, err := root.st.SetFeatureVersion(v)
	if err != nil {
		return nil, err
	}
	return root.withStorage(newStorage), nil
}

// SetTableHash implements the interface doltdb.RootValue.
func (root *rootValue) SetTableHash(ctx context.Context, tName string, h hash.Hash) (doltdb.RootValue, error) {
	val, err := root.vrw.ReadValue(ctx, h)

	if err != nil {
		return nil, err
	}

	ref, err := types.NewRef(val, root.vrw.Format())

	if err != nil {
		return nil, err
	}

	// TODO: schema
	return root.putTable(ctx, doltdb.TableName{Name: tName}, ref)
}

// VRW implements the interface doltdb.RootValue.
func (root *rootValue) VRW() types.ValueReadWriter {
	return root.vrw
}

// getTableMap returns the tableMap for this root.
func (root *rootValue) getTableMap(ctx context.Context, schemaName string) (rootTableMap, error) {
	if schemaName == "" {
		schemaName = doltdb.DefaultSchemaName
	}
	return root.st.GetTablesMap(ctx, root.vrw, root.ns, schemaName)
}

// putTable provides an inner implementation that is called from multiple other functions.
func (root *rootValue) putTable(ctx context.Context, tName doltdb.TableName, ref types.Ref) (doltdb.RootValue, error) {
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
func (root *rootValue) withStorage(st rootStorage) *rootValue {
	return &rootValue{root.vrw, root.ns, st, nil, hash.Hash{}}
}
