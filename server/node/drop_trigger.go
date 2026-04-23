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

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
)

// DropTrigger handles the DROP TRIGGER statement.
type DropTrigger struct {
	ifExists    bool
	trigger     string
	onTblSchema string
	onTable     string
	cascade     bool
}

var _ sql.ExecSourceRel = (*DropTrigger)(nil)
var _ vitess.Injectable = (*DropTrigger)(nil)

// NewDropTrigger returns a new *DropTrigger.
func NewDropTrigger(ifExists bool, trigger, schema, table string, cascade bool) *DropTrigger {
	return &DropTrigger{
		ifExists:    ifExists,
		trigger:     trigger,
		onTblSchema: schema,
		onTable:     table,
		cascade:     cascade,
	}
}

// Children implements the interface sql.ExecSourceRel.
func (c *DropTrigger) Children() []sql.Node {
	return nil
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (c *DropTrigger) IsReadOnly() bool {
	return false
}

// Resolved implements the interface sql.ExecSourceRel.
func (c *DropTrigger) Resolved() bool {
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (c *DropTrigger) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	schema, err := core.GetSchemaName(ctx, nil, c.onTblSchema)
	if err != nil {
		return nil, err
	}
	collection, err := core.GetTriggersCollectionFromContext(ctx, ctx.GetCurrentDatabase())
	if err != nil {
		return nil, err
	}
	triggerID := id.NewTrigger(schema, c.onTable, c.trigger)
	hasTrigger := collection.HasTrigger(ctx, triggerID)
	if !hasTrigger && c.ifExists {
		return sql.RowsToRowIter(), nil
	}
	if err = collection.DropTrigger(ctx, triggerID); err != nil {
		return nil, err
	}
	return sql.RowsToRowIter(), nil
}

// Schema implements the interface sql.ExecSourceRel.
func (c *DropTrigger) Schema(ctx *sql.Context) sql.Schema {
	return nil
}

// String implements the interface sql.ExecSourceRel.
func (c *DropTrigger) String() string {
	return "DROP TRIGGER"
}

// WithChildren implements the interface sql.ExecSourceRel.
func (c *DropTrigger) WithChildren(ctx *sql.Context, children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(c, children...)
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (c *DropTrigger) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return c, nil
}
