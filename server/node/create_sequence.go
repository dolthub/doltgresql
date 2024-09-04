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
	"strings"

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/sequences"
)

// CreateSequence handles the CREATE SEQUENCE statement, along with SERIAL type definitions.
type CreateSequence struct {
	schema      string
	ifNotExists bool
	sequence    *sequences.Sequence
}

var _ sql.ExecSourceRel = (*CreateSequence)(nil)
var _ vitess.Injectable = (*CreateSequence)(nil)

// NewCreateSequence returns a new *CreateSequence.
func NewCreateSequence(ifNotExists bool, schema string, sequence *sequences.Sequence) *CreateSequence {
	return &CreateSequence{
		schema:      schema,
		ifNotExists: ifNotExists,
		sequence:    sequence,
	}
}

// CheckPrivileges implements the interface sql.ExecSourceRel.
func (c *CreateSequence) CheckPrivileges(ctx *sql.Context, opChecker sql.PrivilegedOperationChecker) bool {
	// TODO: implement privilege checking
	return true
}

// Children implements the interface sql.ExecSourceRel.
func (c *CreateSequence) Children() []sql.Node {
	return nil
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (c *CreateSequence) IsReadOnly() bool {
	return false
}

// Resolved implements the interface sql.ExecSourceRel.
func (c *CreateSequence) Resolved() bool {
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (c *CreateSequence) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	if strings.HasPrefix(strings.ToLower(c.sequence.Name), "dolt") {
		return nil, fmt.Errorf("sequences cannot be prefixed with 'dolt'")
	}
	schema, err := core.GetSchemaName(ctx, nil, c.schema)
	if err != nil {
		return nil, err
	}

	// Check that the sequence name is free
	relationType, err := core.GetRelationType(ctx, schema, c.sequence.Name)
	if err != nil {
		return nil, err
	}
	if relationType != core.RelationType_DoesNotExist && c.ifNotExists {
		if c.ifNotExists {
			// TODO: issue a notice
			return sql.RowsToRowIter(), nil
		}
		return nil, fmt.Errorf(`relation "%s" already exists`, c.sequence.Name)
	}
	// Check that the OWNED BY is valid, if it exists
	if len(c.sequence.OwnerTable) > 0 {
		relationType, err = core.GetRelationType(ctx, schema, c.sequence.OwnerTable)
		if err != nil {
			return nil, err
		}
		if relationType == core.RelationType_DoesNotExist {
			return nil, fmt.Errorf(`relation "%s" does not exist`, c.sequence.OwnerTable)
		} else if relationType != core.RelationType_Table {
			return nil, fmt.Errorf(`sequence cannot be owned by relation "%s"`, c.sequence.OwnerTable)
		}

		table, err := core.GetDoltTableFromContext(ctx, doltdb.TableName{Name: c.sequence.OwnerTable, Schema: schema})
		if err != nil {
			return nil, err
		}
		if table == nil {
			return nil, fmt.Errorf(`table "%s" cannot be found but says it exists`, c.sequence.OwnerTable)
		}
		tableSch, err := table.GetSchema(ctx)
		if err != nil {
			return nil, err
		}
		colFound := false
		for _, col := range tableSch.GetAllCols().GetColumns() {
			if col.Name == c.sequence.OwnerColumn {
				colFound = true
				break
			}
		}
		if !colFound {
			return nil, fmt.Errorf(`column "%s" of relation "%s" does not exist`, c.sequence.OwnerColumn, c.sequence.OwnerTable)
		}
	}
	// Create the sequence since we know it's completely valid
	collection, err := core.GetCollectionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if err = collection.CreateSequence(schema, c.sequence); err != nil {
		return nil, err
	}
	return sql.RowsToRowIter(), nil
}

// Schema implements the interface sql.ExecSourceRel.
func (c *CreateSequence) Schema() sql.Schema {
	return nil
}

// String implements the interface sql.ExecSourceRel.
func (c *CreateSequence) String() string {
	return "CREATE SEQUENCE"
}

// WithChildren implements the interface sql.ExecSourceRel.
func (c *CreateSequence) WithChildren(children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(c, children...)
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (c *CreateSequence) WithResolvedChildren(children []any) (any, error) {
	if len(children) != 0 {
		return nil, fmt.Errorf("invalid vitess child count, expected `0` but got `%d`", len(children))
	}
	return c, nil
}
