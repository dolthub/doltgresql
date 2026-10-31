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
	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
)

// DropTable is a node that implements functionality specifically relevant to Doltgres' table dropping needs.
type DropTable struct {
	gmsDropTable *plan.DropTable
}

var _ sql.ExecBuilderNode = (*DropTable)(nil)

// NewDropTable returns a new *DropTable.
func NewDropTable(dropTable *plan.DropTable) *DropTable {
	return &DropTable{
		gmsDropTable: dropTable,
	}
}

// Children implements the interface sql.ExecBuilderNode.
func (c *DropTable) Children() []sql.Node {
	return c.gmsDropTable.Children()
}

// IsReadOnly implements the interface sql.ExecBuilderNode.
func (c *DropTable) IsReadOnly() bool {
	return false
}

// Resolved implements the interface sql.ExecBuilderNode.
func (c *DropTable) Resolved() bool {
	return c.gmsDropTable != nil && c.gmsDropTable.Resolved()
}

// BuildRowIter implements the interface sql.ExecBuilderNode.
func (c *DropTable) BuildRowIter(ctx *sql.Context, b sql.NodeExecBuilder, r sql.Row) (sql.RowIter, error) {
	dbNames := make([]string, len(c.gmsDropTable.Tables))
	tableIDs := make([]id.Table, len(c.gmsDropTable.Tables))
	var sequenceDbNames []string
	var sequenceIDs []id.Sequence
	for i, table := range c.gmsDropTable.Tables {
		resolvedTable, ok := table.(*plan.ResolvedTable)
		if !ok {
			return nil, errors.Errorf("encountered unexpected table type `%T` during DROP TABLE", table)
		}
		schemaName, err := core.GetSchemaName(ctx, resolvedTable.Database(), "")
		if err != nil {
			return nil, err
		}
		dbNames[i] = resolvedTable.Database().Name()
		tableIDs[i] = id.NewTable(schemaName, resolvedTable.Name())
		collection, err := core.GetSequencesCollectionFromContext(ctx, dbNames[i])
		if err != nil {
			return nil, err
		}
		ownedSequences, err := collection.GetSequencesWithTable(ctx, doltdb.TableName{Name: resolvedTable.Name(), Schema: schemaName})
		if err != nil {
			return nil, err
		}
		for _, sequence := range ownedSequences {
			sequenceDbNames = append(sequenceDbNames, dbNames[i])
			sequenceIDs = append(sequenceIDs, sequence.Id)
		}
	}

	dropTableIter, err := b.Build(ctx, c.gmsDropTable, r)
	if err != nil {
		return nil, err
	}
	for i, tableID := range tableIDs {
		if err = id.ValidateOperation(ctx, id.Section_Table, id.Operation_Delete, dbNames[i], tableID.AsId(), id.Null); err != nil {
			return nil, err
		}
		if err = id.PerformOperation(ctx, id.Section_Table, id.Operation_Delete, dbNames[i], tableID.AsId(), id.Null); err != nil {
			return nil, err
		}
	}
	for i, sequenceID := range sequenceIDs {
		if err = core.DropSequenceFromTracker(ctx, sequenceDbNames[i], sequenceID); err != nil {
			return nil, err
		}
	}
	return dropTableIter, nil
}

// Schema implements the interface sql.ExecBuilderNode.
func (c *DropTable) Schema(ctx *sql.Context) sql.Schema {
	return c.gmsDropTable.Schema(ctx)
}

// String implements the interface sql.ExecBuilderNode.
func (c *DropTable) String() string {
	return c.gmsDropTable.String()
}

// WithChildren implements the interface sql.ExecBuilderNode.
func (c *DropTable) WithChildren(ctx *sql.Context, children ...sql.Node) (sql.Node, error) {
	gmsDropTable, err := c.gmsDropTable.WithChildren(ctx, children...)
	if err != nil {
		return nil, err
	}
	return &DropTable{
		gmsDropTable: gmsDropTable.(*plan.DropTable),
	}, nil
}
