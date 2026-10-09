// Copyright 2026 Dolthub, Inc.
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

package node

import (
	"fmt"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/types"

	"github.com/dolthub/doltgresql/core"
)

// DropForeignKey retains the resolved database schema when removing a foreign key.
// GMS's DropForeignKey reacquires an unqualified database and does not pass a schema to DropForeignKey.
type DropForeignKey struct {
	database  sql.Database
	tableName string
	name      string
}

var _ sql.ExecSourceRel = (*DropForeignKey)(nil)

// NewDropForeignKey returns a foreign-key drop targeting the given resolved table.
func NewDropForeignKey(table *plan.ResolvedTable, name string) *DropForeignKey {
	return &DropForeignKey{database: table.Database(), tableName: table.Name(), name: name}
}

// Children implements the interface sql.ExecSourceRel.
func (d *DropForeignKey) Children() []sql.Node {
	return nil
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (d *DropForeignKey) IsReadOnly() bool {
	return false
}

// Resolved implements the interface sql.ExecSourceRel.
func (d *DropForeignKey) Resolved() bool {
	return d.database != nil
}

// RowIter implements the interface sql.ExecSourceRel.
func (d *DropForeignKey) RowIter(ctx *sql.Context, _ sql.Row) (sql.RowIter, error) {
	// Reacquire the table at execution time so preceding operations in a multi-action ALTER are visible,
	// while keeping the schema resolved during analysis rather than searching the current search_path.
	table, ok, err := d.database.GetTableInsensitive(ctx, d.tableName)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, sql.ErrTableNotFound.New(d.tableName)
	}
	fkTable, ok := sql.GetUnderlyingTable(table).(sql.ForeignKeyTable)
	if !ok {
		return nil, sql.ErrNoForeignKeySupport.New(d.tableName)
	}
	schemaName, err := core.GetSchemaName(ctx, d.database, "")
	if err != nil {
		return nil, err
	}
	if err = fkTable.DropForeignKey(ctx, d.name, fkTable.Name(), schemaName); err != nil {
		return nil, err
	}
	return sql.RowsToRowIter(sql.NewRow(types.NewOkResult(0))), nil
}

// Schema implements the interface sql.ExecSourceRel.
func (d *DropForeignKey) Schema(ctx *sql.Context) sql.Schema {
	return types.OkResultSchema
}

// String implements the interface sql.ExecSourceRel.
func (d *DropForeignKey) String() string {
	return fmt.Sprintf("DropForeignKey(%s)", d.name)
}

// WithChildren implements the interface sql.ExecSourceRel.
func (d *DropForeignKey) WithChildren(ctx *sql.Context, children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(d, children...)
}
