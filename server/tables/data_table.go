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

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dtables"
	"github.com/dolthub/go-mysql-server/sql"
)

// DataTable represents a system table in its read-only state.
type DataTable struct {
	doltTable *sqle.DoltTable
	schema    string
	handler   DataTableHandler
}

// IndexedDataTable represents a system table, in its read-only state, that has had a sql.IndexLookup applied.
type IndexedDataTable struct {
	DataTable
	indexedDoltTable *sqle.IndexedDoltTable
}

var _ dtables.VersionableTable = (*DataTable)(nil)
var _ sql.CheckTable = (*DataTable)(nil)
var _ sql.CommentedTable = (*DataTable)(nil)
var _ sql.DebugStringer = (*DataTable)(nil)
var _ sql.ForeignKeyTable = (*DataTable)(nil)
var _ sql.IndexAddressableTable = (*DataTable)(nil)
var _ sql.PrimaryKeyTable = (*DataTable)(nil)
var _ sql.StatisticsTable = (*DataTable)(nil)
var _ sql.Table = (*DataTable)(nil)
var _ sql.IndexedTable = (*IndexedDataTable)(nil)

// NewDataTable creates a new *DataTable from the given *sqle.DoltTable.
func NewDataTable(doltTable *sqle.DoltTable, schema string) *DataTable {
	return &DataTable{
		doltTable: doltTable,
		schema:    schema,
		handler:   getHandler(schema, doltTable.Name()),
	}
}

// NewIndexedDataTable creates a new *IndexedDataTable from the given *sqle.IndexedDoltTable.
func NewIndexedDataTable(indexedDoltTable *sqle.IndexedDoltTable, schema string) *IndexedDataTable {
	return &IndexedDataTable{
		DataTable:        *NewDataTable(indexedDoltTable.DoltTable, schema),
		indexedDoltTable: indexedDoltTable,
	}
}

// newDataTable is used internally to construct a new *DataTable.
func newDataTable(doltTable *sqle.DoltTable, schema string, handler DataTableHandler) DataTable {
	return DataTable{
		doltTable: doltTable,
		schema:    schema,
		handler:   handler,
	}
}

// newIndexedDataTable is used internally to construct a new *IndexedDataTable.
func newIndexedDataTable(indexedDoltTable *sqle.IndexedDoltTable, schema string, handler DataTableHandler) *IndexedDataTable {
	return &IndexedDataTable{
		DataTable:        newDataTable(indexedDoltTable.DoltTable, schema, handler),
		indexedDoltTable: indexedDoltTable,
	}
}

// AddForeignKey implements the interface sql.ForeignKeyTable.
func (tbl *DataTable) AddForeignKey(ctx *sql.Context, fk sql.ForeignKeyConstraint) error {
	return fmt.Errorf("adding foreign key constraints on `%s` is not supported: read-only table", tbl.doltTable.Name())
}

// Collation implements the interface sql.Table.
func (tbl *DataTable) Collation() sql.CollationID {
	return tbl.doltTable.Collation()
}

// Comment implements the interface sql.CommentedTable.
func (tbl *DataTable) Comment() string {
	return tbl.doltTable.Comment()
}

// CreateIndexForForeignKey implements the interface sql.ForeignKeyTable.
func (tbl *DataTable) CreateIndexForForeignKey(ctx *sql.Context, indexDef sql.IndexDef) error {
	return fmt.Errorf("creating indexes for foreign key constraints on `%s` is not supported: read-only table", tbl.doltTable.Name())
}

// DataLength implements the interface sql.StatisticsTable.
func (tbl *DataTable) DataLength(ctx *sql.Context) (uint64, error) {
	return tbl.doltTable.DataLength(ctx)
}

// DebugString implements the interface sql.DebugStringer.
func (tbl *DataTable) DebugString() string {
	return tbl.doltTable.DebugString()
}

// DropForeignKey implements the interface sql.ForeignKeyTable.
func (tbl *DataTable) DropForeignKey(ctx *sql.Context, fkName string) error {
	return fmt.Errorf("dropping foreign key constraints on `%s` is not supported: read-only table", tbl.doltTable.Name())
}

// GetChecks implements the interface sql.CheckTable.
func (tbl *DataTable) GetChecks(ctx *sql.Context) ([]sql.CheckDefinition, error) {
	return tbl.doltTable.GetChecks(ctx)
}

// GetDeclaredForeignKeys implements the interface sql.ForeignKeyTable.
func (tbl *DataTable) GetDeclaredForeignKeys(ctx *sql.Context) ([]sql.ForeignKeyConstraint, error) {
	return nil, nil
}

// GetForeignKeyEditor implements the interface sql.ForeignKeyTable.
func (tbl *DataTable) GetForeignKeyEditor(ctx *sql.Context) sql.ForeignKeyEditor {
	return nil
}

// GetIndexes implements the interface sql.IndexAddressableTable.
func (tbl *DataTable) GetIndexes(ctx *sql.Context) ([]sql.Index, error) {
	if tbl.handler.UsesIndexes() {
		return tbl.doltTable.GetIndexes(ctx)
	}
	return nil, nil
}

// GetReferencedForeignKeys implements the interface sql.ForeignKeyTable.
func (tbl *DataTable) GetReferencedForeignKeys(ctx *sql.Context) ([]sql.ForeignKeyConstraint, error) {
	return nil, nil
}

// IndexedAccess implements the interface sql.IndexAddressableTable.
func (tbl *DataTable) IndexedAccess(lookup sql.IndexLookup) sql.IndexedTable {
	return newIndexedDataTable(tbl.doltTable.IndexedAccess(lookup).(*sqle.IndexedDoltTable), tbl.schema, tbl.handler)
}

// LockedToRoot implements the interface sql.VersionableTable.
func (tbl *DataTable) LockedToRoot(ctx *sql.Context, root doltdb.RootValue) (sql.IndexAddressableTable, error) {
	lockedTable, err := tbl.doltTable.LockedToRoot(ctx, root)
	if err != nil {
		return nil, err
	}
	newTbl := newDataTable(lockedTable.(*sqle.DoltTable), tbl.schema, tbl.handler)
	return &newTbl, nil
}

// Name implements the interface sql.Table.
func (tbl *DataTable) Name() string {
	return tbl.doltTable.Name()
}

// PartitionRows implements the interface sql.Table.
func (tbl *DataTable) PartitionRows(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	rowIter, err := tbl.doltTable.PartitionRows(ctx, partition)
	if err != nil {
		return nil, err
	}
	return tbl.handler.RowIter(ctx, rowIter)
}

// Partitions implements the interface sql.Table.
func (tbl *DataTable) Partitions(ctx *sql.Context) (sql.PartitionIter, error) {
	return tbl.doltTable.Partitions(ctx)
}

// PreciseMatch implements the interface sql.IndexAddressableTable.
func (tbl *DataTable) PreciseMatch() bool {
	return tbl.doltTable.PreciseMatch()
}

// PrimaryKeySchema implements the interface sql.PrimaryKeyTable.
func (tbl *DataTable) PrimaryKeySchema() sql.PrimaryKeySchema {
	return tbl.doltTable.PrimaryKeySchema()
}

// RowCount implements the interface sql.StatisticsTable.
func (tbl *DataTable) RowCount(ctx *sql.Context) (uint64, bool, error) {
	return tbl.doltTable.RowCount(ctx)
}

// Schema implements the interface sql.Table.
func (tbl *DataTable) Schema() sql.Schema {
	return tbl.doltTable.Schema()
}

// String implements the interface sql.Table.
func (tbl *DataTable) String() string {
	return tbl.doltTable.String()
}

// UpdateForeignKey implements the interface sql.ForeignKeyTable.
func (tbl *DataTable) UpdateForeignKey(ctx *sql.Context, fkName string, fk sql.ForeignKeyConstraint) error {
	return fmt.Errorf("updating foreign key constraints on `%s` is not supported: read-only table", tbl.doltTable.Name())
}

// LookupPartitions implements the interface sql.IndexedTable.
func (tbl *IndexedDataTable) LookupPartitions(ctx *sql.Context, lookup sql.IndexLookup) (sql.PartitionIter, error) {
	return tbl.indexedDoltTable.LookupPartitions(ctx, lookup)
}

// Partitions implements the interface sql.Table.
func (tbl *IndexedDataTable) Partitions(ctx *sql.Context) (sql.PartitionIter, error) {
	return tbl.indexedDoltTable.Partitions(ctx)
}

// PartitionRows implements the interface sql.Table.
func (tbl *IndexedDataTable) PartitionRows(ctx *sql.Context, part sql.Partition) (sql.RowIter, error) {
	rowIter, err := tbl.indexedDoltTable.PartitionRows(ctx, part)
	if err != nil {
		return nil, err
	}
	return tbl.handler.RowIter(ctx, rowIter)
}
