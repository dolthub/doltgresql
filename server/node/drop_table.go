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
	"github.com/dolthub/doltgresql/server/auth"
)

// DropTable is a node that implements functionality specifically relevant to Doltgres' table removal needs.
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
func (d *DropTable) Children() []sql.Node {
	return d.gmsDropTable.Children()
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (d *DropTable) IsReadOnly() bool {
	return false
}

// Resolved implements the interface sql.ExecSourceRel.
func (d *DropTable) Resolved() bool {
	return d.gmsDropTable != nil && d.gmsDropTable.Resolved()
}

// RowIter implements the interface sql.ExecSourceRel.
func (d *DropTable) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	var userRole auth.Role
	auth.LockRead(func() {
		userRole = auth.GetRole(ctx.Client().User)
	})
	if !userRole.IsValid() {
		return nil, fmt.Errorf(`role "%s" does not exist`, ctx.Client().User)
	}

	dropTableIter, err := rowexec.DefaultBuilder.Build(ctx, d.gmsDropTable, r)
	if err != nil {
		return nil, err
	}

	for _, node := range d.gmsDropTable.Tables {
		// Since we handle orphaned owners anyway, we won't treat this as a failure condition
		if table, ok := node.(sql.Table); ok {
			if databaser, ok := table.(sql.Databaser); ok {
				schemaName, err := core.GetSchemaName(ctx, databaser.Database(), "")
				if err != nil {
					return nil, err
				}
				auth.LockWrite(func() {
					auth.RemoveOwner(auth.OwnershipKey{
						PrivilegeObject: auth.PrivilegeObject_TABLE,
						Schema:          schemaName,
						Name:            table.Name(),
					}, userRole.ID())
					err = auth.PersistChanges()
				})
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return dropTableIter, err
}

// Schema implements the interface sql.ExecSourceRel.
func (d *DropTable) Schema() sql.Schema {
	return d.gmsDropTable.Schema()
}

// String implements the interface sql.ExecSourceRel.
func (d *DropTable) String() string {
	return d.gmsDropTable.String()
}

// WithChildren implements the interface sql.ExecSourceRel.
func (d *DropTable) WithChildren(children ...sql.Node) (sql.Node, error) {
	gmsDropTable, err := d.gmsDropTable.WithChildren(children...)
	if err != nil {
		return nil, err
	}
	return &DropTable{
		gmsDropTable: gmsDropTable.(*plan.DropTable),
	}, nil
}
