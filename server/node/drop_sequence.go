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

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
)

// DropSequence handles the DROP SEQUENCE statement.
type DropSequence struct {
	schema   string
	sequence string
	ifExists bool
	cascade  bool
}

var _ sql.ExecSourceRel = (*DropSequence)(nil)
var _ vitess.Injectable = (*DropSequence)(nil)

// NewDropSequence returns a new *DropSequence.
func NewDropSequence(ifExists bool, schema string, sequence string, cascade bool) *DropSequence {
	return &DropSequence{
		schema:   schema,
		sequence: sequence,
		ifExists: ifExists,
		cascade:  cascade,
	}
}

// Children implements the interface sql.ExecSourceRel.
func (c *DropSequence) Children() []sql.Node {
	return nil
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (c *DropSequence) IsReadOnly() bool {
	return false
}

// Resolved implements the interface sql.ExecSourceRel.
func (c *DropSequence) Resolved() bool {
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (c *DropSequence) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	schema, err := core.GetSchemaName(ctx, nil, c.schema)
	if err != nil {
		return nil, err
	}
	relationType, err := core.GetRelationType(ctx, schema, c.sequence)
	if err != nil {
		return nil, err
	}
	if relationType == core.RelationType_DoesNotExist {
		if c.ifExists {
			// TODO: issue a notice
			return sql.RowsToRowIter(), nil
		}
		return nil, errors.Errorf(`sequence "%s" does not exist`, c.sequence)
	}
	collection, err := core.GetSequencesCollectionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	sequenceID := id.NewSequence(schema, c.sequence)
	sequence, err := collection.GetSequence(ctx, sequenceID)
	if err != nil {
		return nil, err
	}
	if sequence.OwnerTable.IsValid() {
		if c.cascade {
			// TODO: if the sequence is referenced by the column's default value, then we also need to delete the default
			return nil, errors.Errorf(`cascading sequence drops are not yet supported`)
		} else {
			// TODO: this error is only true if the sequence is referenced by the column's default value
			return nil, errors.Errorf(`cannot drop sequence %s because other objects depend on it`, c.sequence)
		}
	}
	if err = collection.DropSequence(ctx, sequenceID); err != nil {
		return nil, err
	}
	return sql.RowsToRowIter(), nil
}

// Schema implements the interface sql.ExecSourceRel.
func (c *DropSequence) Schema() sql.Schema {
	return nil
}

// String implements the interface sql.ExecSourceRel.
func (c *DropSequence) String() string {
	return "DROP SEQUENCE"
}

// WithChildren implements the interface sql.ExecSourceRel.
func (c *DropSequence) WithChildren(children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(c, children...)
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (c *DropSequence) WithResolvedChildren(children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return c, nil
}
