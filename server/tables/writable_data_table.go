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

package tables

import (
	"fmt"

	"github.com/dolthub/dolt/go/libraries/doltcore/sqle"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/go-mysql-server/sql"
)

// WritableDataTable represents a system table.
type WritableDataTable struct {
	DataTable
	writableDoltTable *sqle.WritableDoltTable
}

// WritableIndexedDataTable represents a system table that has had a sql.IndexLookup applied.
type WritableIndexedDataTable struct {
	WritableDataTable
	writableIndexedDoltTable *sqle.WritableIndexedDoltTable
}

var _ sql.Databaseable = (*WritableDataTable)(nil)
var _ sql.DeletableTable = (*WritableDataTable)(nil)
var _ sql.InsertableTable = (*WritableDataTable)(nil)
var _ sql.TruncateableTable = (*WritableDataTable)(nil)
var _ sql.UpdatableTable = (*WritableDataTable)(nil)
var _ sql.IndexedTable = (*WritableIndexedDataTable)(nil)

// NewWritableDataTable creates a new *WritableDataTable from the given *sqle.WritableDoltTable.
func NewWritableDataTable(doltTable *sqle.WritableDoltTable, schema string) *WritableDataTable {
	return &WritableDataTable{
		DataTable:         *NewDataTable(doltTable.DoltTable, schema),
		writableDoltTable: doltTable,
	}
}

// NewWritableIndexedDataTable creates a new *WritableIndexedDataTable from the given *sqle.WritableIndexedDoltTable.
func NewWritableIndexedDataTable(doltTable *sqle.WritableIndexedDoltTable, schema string) *WritableIndexedDataTable {
	return &WritableIndexedDataTable{
		WritableDataTable:        *NewWritableDataTable(doltTable.WritableDoltTable, schema),
		writableIndexedDoltTable: doltTable,
	}
}

// newWritableDataTable is used internally to construct a new *WritableDataTable.
func newWritableDataTable(doltTable *sqle.WritableDoltTable, schema string, handler DataTableHandler) WritableDataTable {
	return WritableDataTable{
		DataTable:         newDataTable(doltTable.DoltTable, schema, handler),
		writableDoltTable: doltTable,
	}
}

// newWritableIndexedDataTable is used internally to construct a new *WritableIndexedDataTable.
func newWritableIndexedDataTable(doltTable *sqle.WritableIndexedDoltTable, schema string, handler DataTableHandler) *WritableIndexedDataTable {
	return &WritableIndexedDataTable{
		WritableDataTable:        newWritableDataTable(doltTable.WritableDoltTable, schema, handler),
		writableIndexedDoltTable: doltTable,
	}
}

// AddForeignKey implements the interface sql.ForeignKeyTable.
func (tbl *WritableDataTable) AddForeignKey(ctx *sql.Context, fk sql.ForeignKeyConstraint) error {
	return fmt.Errorf("adding foreign key constraints on `%s` is not currently supported", tbl.writableDoltTable.Name())
}

// CreateIndexForForeignKey implements the interface sql.ForeignKeyTable.
func (tbl *WritableDataTable) CreateIndexForForeignKey(ctx *sql.Context, indexDef sql.IndexDef) error {
	return fmt.Errorf("creating indexes for foreign key constraints on `%s` is not currently supported", tbl.writableDoltTable.Name())
}

// Database implements the interface sql.Databaseable.
func (tbl *WritableDataTable) Database() string {
	return tbl.writableDoltTable.Database()
}

// Deleter implements the interface sql.DeletableTable.
func (tbl *WritableDataTable) Deleter(ctx *sql.Context) sql.RowDeleter {
	return newDataTableEditorInterface(tbl.writableDoltTable.Deleter(ctx).(dsess.TableWriter), tbl.handler)
}

// DropForeignKey implements the interface sql.ForeignKeyTable.
func (tbl *WritableDataTable) DropForeignKey(ctx *sql.Context, fkName string) error {
	return fmt.Errorf("dropping foreign key constraints on `%s` is not currently supported", tbl.writableDoltTable.Name())
}

// GetDeclaredForeignKeys implements the interface sql.ForeignKeyTable.
func (tbl *WritableDataTable) GetDeclaredForeignKeys(ctx *sql.Context) ([]sql.ForeignKeyConstraint, error) {
	return nil, nil
}

// GetForeignKeyEditor implements the interface sql.ForeignKeyTable.
func (tbl *WritableDataTable) GetForeignKeyEditor(ctx *sql.Context) sql.ForeignKeyEditor {
	return nil
}

// GetReferencedForeignKeys implements the interface sql.ForeignKeyTable.
func (tbl *WritableDataTable) GetReferencedForeignKeys(ctx *sql.Context) ([]sql.ForeignKeyConstraint, error) {
	return nil, nil
}

// IndexedAccess implements the interface sql.IndexAddressableTable.
func (tbl *WritableDataTable) IndexedAccess(lookup sql.IndexLookup) sql.IndexedTable {
	return newWritableIndexedDataTable(tbl.writableDoltTable.IndexedAccess(lookup).(*sqle.WritableIndexedDoltTable), tbl.schema, tbl.handler)
}

// Inserter implements the interface sql.InsertableTable.
func (tbl *WritableDataTable) Inserter(ctx *sql.Context) sql.RowInserter {
	return newDataTableEditorInterface(tbl.writableDoltTable.Inserter(ctx).(dsess.TableWriter), tbl.handler)
}

// Truncate implements the interface sql.TruncateableTable.
func (tbl *WritableDataTable) Truncate(ctx *sql.Context) (int, error) {
	return tbl.writableDoltTable.Truncate(ctx)
}

// UpdateForeignKey implements the interface sql.ForeignKeyTable.
func (tbl *WritableDataTable) UpdateForeignKey(ctx *sql.Context, fkName string, fk sql.ForeignKeyConstraint) error {
	return fmt.Errorf("updating foreign key constraints on `%s` is not currently supported", tbl.writableDoltTable.Name())
}

// Updater implements the interface sql.UpdatableTable.
func (tbl *WritableDataTable) Updater(ctx *sql.Context) sql.RowUpdater {
	return newDataTableEditorInterface(tbl.writableDoltTable.Updater(ctx).(dsess.TableWriter), tbl.handler)
}

// LookupPartitions implements the interface sql.IndexedTable.
func (tbl *WritableIndexedDataTable) LookupPartitions(ctx *sql.Context, lookup sql.IndexLookup) (sql.PartitionIter, error) {
	return tbl.writableIndexedDoltTable.LookupPartitions(ctx, lookup)
}

// Partitions implements the interface sql.Table.
func (tbl *WritableIndexedDataTable) Partitions(ctx *sql.Context) (sql.PartitionIter, error) {
	return tbl.writableIndexedDoltTable.Partitions(ctx)
}

// PartitionRows implements the interface sql.Table.
func (tbl *WritableIndexedDataTable) PartitionRows(ctx *sql.Context, part sql.Partition) (sql.RowIter, error) {
	rowIter, err := tbl.writableIndexedDoltTable.PartitionRows(ctx, part)
	if err != nil {
		return nil, err
	}
	return tbl.handler.RowIter(ctx, rowIter)
}
