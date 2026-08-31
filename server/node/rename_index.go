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
	"context"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"
)

// RenameIndex handles the ALTER INDEX ... RENAME TO ... statement. The table owning the index is resolved at
// analysis time by the resolveTableForDDL analyzer rule.
type RenameIndex struct {
	IfExists   bool
	DbName     string // optional catalog qualifier for the index name
	SchemaName string // optional schema qualifier for the index name
	TableName  string // optional table name (from the CockroachDB `table@index` syntax)
	IndexName  string
	NewName    string

	// table is the table owning the index, assigned during analysis. It is nil when IfExists is set and no
	// matching index was found, in which case execution is a no-op.
	table sql.TableNode
	// resolved is set once analysis has resolved the target table (or determined there is none to resolve)
	resolved bool
}

var _ sql.ExecSourceRel = (*RenameIndex)(nil)
var _ vitess.Injectable = (*RenameIndex)(nil)

// NewRenameIndex returns a new *RenameIndex.
func NewRenameIndex(ifExists bool, dbName, schemaName, tableName, indexName, newName string) *RenameIndex {
	return &RenameIndex{
		IfExists:   ifExists,
		DbName:     dbName,
		SchemaName: schemaName,
		TableName:  tableName,
		IndexName:  indexName,
		NewName:    newName,
	}
}

// WithResolvedTable returns a copy of this node with the table owning the index assigned. A nil table is only
// valid when IfExists is set, and makes execution a no-op.
func (r *RenameIndex) WithResolvedTable(table sql.TableNode) *RenameIndex {
	nr := *r
	nr.table = table
	nr.resolved = true
	return &nr
}

// Children implements the interface sql.ExecSourceRel.
func (r *RenameIndex) Children() []sql.Node {
	return nil
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (r *RenameIndex) IsReadOnly() bool {
	return false
}

// Resolved implements the interface sql.ExecSourceRel.
func (r *RenameIndex) Resolved() bool {
	return r.resolved
}

// RowIter implements the interface sql.ExecSourceRel.
func (r *RenameIndex) RowIter(ctx *sql.Context, _ sql.Row) (sql.RowIter, error) {
	if r.table == nil {
		if r.IfExists {
			// TODO: send notice "relation ... does not exist, skipping"
			return sql.RowsToRowIter(), nil
		}
		return nil, errors.Errorf(`relation "%s" does not exist`, r.IndexName)
	}
	alterable, ok := r.table.UnderlyingTable().(sql.IndexAlterableTable)
	if !ok {
		return nil, errors.Errorf(`table "%s" does not support renaming indexes`, r.table.Name())
	}
	if err := alterable.RenameIndex(ctx, r.IndexName, r.NewName); err != nil {
		return nil, err
	}
	return sql.RowsToRowIter(), nil
}

// Schema implements the interface sql.ExecSourceRel.
func (r *RenameIndex) Schema(ctx *sql.Context) sql.Schema {
	return nil
}

// String implements the interface sql.ExecSourceRel.
func (r *RenameIndex) String() string {
	return "ALTER INDEX RENAME"
}

// WithChildren implements the interface sql.ExecSourceRel.
func (r *RenameIndex) WithChildren(ctx *sql.Context, children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(r, children...)
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (r *RenameIndex) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return r, nil
}
