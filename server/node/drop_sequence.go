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

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core"
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

// CheckPrivileges implements the interface sql.ExecSourceRel.
func (c *DropSequence) CheckPrivileges(ctx *sql.Context, opChecker sql.PrivilegedOperationChecker) bool {
	// TODO: implement privilege checking
	return true
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
		return nil, fmt.Errorf(`sequence "%s" does not exist`, c.sequence)
	}
	collection, err := core.GetCollectionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if sequence := collection.GetSequence(doltdb.TableName{Name: c.sequence, Schema: schema}); len(sequence.OwnerTable) > 0 {
		if c.cascade {
			// TODO: handle cascade
			return nil, fmt.Errorf(`cascading sequence drops are not yet supported`)
		} else {
			return nil, fmt.Errorf(`cannot drop sequence %s because other objects depend on it`, c.sequence)
		}
	}
	if err = collection.DropSequence(doltdb.TableName{Name: c.sequence, Schema: schema}); err != nil {
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
