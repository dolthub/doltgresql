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
	"github.com/dolthub/go-mysql-server/sql"
)

// VirtualTable represents a table that does not enforce any particular storage of its data.
type VirtualTable struct {
	handler Handler
	schema  sql.DatabaseSchema
}

var _ sql.DebugStringer = (*VirtualTable)(nil)
var _ sql.PrimaryKeyTable = (*VirtualTable)(nil)
var _ sql.Table = (*VirtualTable)(nil)
var _ sql.DatabaseSchemaTable = (*VirtualTable)(nil)

// NewVirtualTable creates a new *VirtualTable from the given Handler.
func NewVirtualTable(handler Handler, schema sql.DatabaseSchema) *VirtualTable {
	return &VirtualTable{
		handler: handler,
		schema:  schema,
	}
}

// Collation implements the interface sql.Table.
func (tbl *VirtualTable) Collation() sql.CollationID {
	return sql.Collation_Default
}

// DebugString implements the interface sql.DebugStringer.
func (tbl *VirtualTable) DebugString() string {
	return "virt_table_" + tbl.String()
}

// Name implements the interface sql.Table.
func (tbl *VirtualTable) Name() string {
	return tbl.handler.Name()
}

// PartitionRows implements the interface sql.Table.
func (tbl *VirtualTable) PartitionRows(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	return tbl.handler.RowIter(ctx)
}

// Partitions implements the interface sql.Table.
func (tbl *VirtualTable) Partitions(ctx *sql.Context) (sql.PartitionIter, error) {
	return &partitionIter{
		used: false,
	}, nil
}

// PrimaryKeySchema implements the interface sql.PrimaryKeyTable.
func (tbl *VirtualTable) PrimaryKeySchema() sql.PrimaryKeySchema {
	return tbl.handler.Schema()
}

// Schema implements the interface sql.Table.
func (tbl *VirtualTable) Schema() sql.Schema {
	return tbl.PrimaryKeySchema().Schema
}

// String implements the interface sql.Table.
func (tbl *VirtualTable) String() string {
	return tbl.Name()
}

func (tbl *VirtualTable) DatabaseSchema() sql.DatabaseSchema {
	return tbl.schema
}
