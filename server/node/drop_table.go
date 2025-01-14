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

package node

import (
	"fmt"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/rowexec"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
)

// DropTable is a node that implements functionality specifically relevant to Doltgres' table dropping needs.
type DropTable struct {
	gmsDropTable *plan.DropTable
}

var _ sql.ExecSourceRel = (*DropTable)(nil)

// NewDropTable returns a new *DropTable.
func NewDropTable(dropTable *plan.DropTable) *DropTable {
	return &DropTable{
		gmsDropTable: dropTable,
	}
}

// Children implements the interface sql.ExecSourceRel.
func (c *DropTable) Children() []sql.Node {
	return c.gmsDropTable.Children()
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (c *DropTable) IsReadOnly() bool {
	return false
}

// Resolved implements the interface sql.ExecSourceRel.
func (c *DropTable) Resolved() bool {
	return c.gmsDropTable != nil && c.gmsDropTable.Resolved()
}

// RowIter implements the interface sql.ExecSourceRel.
func (c *DropTable) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	dropTableIter, err := rowexec.DefaultBuilder.Build(ctx, c.gmsDropTable, r)
	if err != nil {
		return nil, err
	}

	for _, table := range c.gmsDropTable.Tables {
		var schemaName string
		var tableName string
		switch table := table.(type) {
		case *plan.ResolvedTable:
			schemaName, err = core.GetSchemaName(ctx, table.Database(), "")
			if err != nil {
				return nil, err
			}
			tableName = table.Name()
		default:
			return nil, fmt.Errorf("encountered unexpected table type `%T` during DROP TABLE", table)
		}

		tableID := id.NewTable(schemaName, tableName).AsId()
		if err = id.ValidateOperation(ctx, id.Section_Table, id.Operation_Delete, tableID, id.Null); err != nil {
			return nil, err
		}
		if err = id.PerformOperation(ctx, id.Section_Table, id.Operation_Delete, tableID, id.Null); err != nil {
			return nil, err
		}
	}
	return dropTableIter, err
}

// Schema implements the interface sql.ExecSourceRel.
func (c *DropTable) Schema() sql.Schema {
	return c.gmsDropTable.Schema()
}

// String implements the interface sql.ExecSourceRel.
func (c *DropTable) String() string {
	return c.gmsDropTable.String()
}

// WithChildren implements the interface sql.ExecSourceRel.
func (c *DropTable) WithChildren(children ...sql.Node) (sql.Node, error) {
	gmsDropTable, err := c.gmsDropTable.WithChildren(children...)
	if err != nil {
		return nil, err
	}
	return &DropTable{
		gmsDropTable: gmsDropTable.(*plan.DropTable),
	}, nil
}
