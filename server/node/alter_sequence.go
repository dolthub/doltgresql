// Copyright 2025 Dolthub, Inc.
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
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/jackc/pgx/v5/pgproto3"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
)

// AlterSequence handles the ALTER SEQUENCE statement.
type AlterSequence struct {
	ifExists       bool
	targetSchema   string
	targetSequence string
	ownedBy        AlterSequenceOwnedBy
	warnings       []string
}

// AlterSequenceOwnedBy is an option in AlterSequence to represent OWNED BY.
type AlterSequenceOwnedBy struct {
	IsSet  bool
	Table  string
	Column string
}

var _ sql.ExecSourceRel = (*AlterSequence)(nil)
var _ vitess.Injectable = (*AlterSequence)(nil)

// NewAlterSequence returns a new *AlterSequence.
func NewAlterSequence(ifExists bool, targetSchema string, targetSequence string, ownedBy AlterSequenceOwnedBy, warnings ...string) *AlterSequence {
	return &AlterSequence{
		ifExists:       ifExists,
		targetSchema:   targetSchema,
		targetSequence: targetSequence,
		ownedBy:        ownedBy,
		warnings:       warnings,
	}
}

// Children implements the interface sql.ExecSourceRel.
func (c *AlterSequence) Children() []sql.Node {
	return nil
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (c *AlterSequence) IsReadOnly() bool {
	return false
}

// Resolved implements the interface sql.ExecSourceRel.
func (c *AlterSequence) Resolved() bool {
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (c *AlterSequence) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	targetSchema, err := core.GetSchemaName(ctx, nil, c.targetSchema)
	if err != nil {
		return nil, err
	}
	target := id.NewSequence(targetSchema, c.targetSequence)
	collection, err := core.GetSequencesCollectionFromContext(ctx, ctx.GetCurrentDatabase())
	if err != nil {
		return nil, err
	}
	if !collection.HasSequence(ctx, target) {
		if c.ifExists {
			return sql.RowsToRowIter(), nil
		}
		return nil, errors.Errorf(`relation "%s" does not exist`, c.targetSequence)
	}
	seq, err := collection.GetSequence(ctx, target)
	if err != nil {
		return nil, err
	}

	if c.ownedBy.IsSet {
		if len(c.ownedBy.Table) > 0 {
			relationType, err := core.GetRelationType(ctx, targetSchema, c.ownedBy.Table)
			if err != nil {
				return nil, err
			}
			if relationType == core.RelationType_DoesNotExist {
				return nil, errors.Errorf(`relation "%s" does not exist`, c.ownedBy.Table)
			} else if relationType != core.RelationType_Table {
				return nil, errors.Errorf(`sequence cannot be owned by relation "%s"`, c.ownedBy.Table)
			}

			table, err := core.GetSqlTableFromContext(ctx, "", doltdb.TableName{Name: c.ownedBy.Table, Schema: targetSchema})
			if err != nil {
				return nil, err
			}
			if table == nil {
				return nil, errors.Errorf(`table "%s" cannot be found but says it exists`, c.ownedBy.Table)
			}
			var tableColumn *sql.Column
			for _, col := range table.Schema() {
				if col.Name == c.ownedBy.Column {
					tableColumn = col.Copy()
					break
				}
			}
			if tableColumn == nil {
				return nil, errors.Errorf(`column "%s" of relation "%s" does not exist`,
					c.ownedBy.Column, c.ownedBy.Table)
			}
			// We've verified the existence of the table's column, so we can assign it now
			seq.OwnerTable = id.NewTable(targetSchema, c.ownedBy.Table)
			seq.OwnerColumn = c.ownedBy.Column
		} else {
			seq.OwnerTable = ""
			seq.OwnerColumn = ""
		}
	}
	// Display any warnings that were encountered during parsing
	for _, warning := range c.warnings {
		noticeResponse := &pgproto3.NoticeResponse{
			Severity: "WARNING",
			Message:  warning,
		}
		sess := dsess.DSessFromSess(ctx.Session)
		sess.Notice(noticeResponse)
	}
	// Any changes made to the sequence will be persisted at the end of the transaction, so we can just return now
	return sql.RowsToRowIter(), nil
}

// Schema implements the interface sql.ExecSourceRel.
func (c *AlterSequence) Schema() sql.Schema {
	return nil
}

// String implements the interface sql.ExecSourceRel.
func (c *AlterSequence) String() string {
	return "ALTER SEQUENCE"
}

// WithChildren implements the interface sql.ExecSourceRel.
func (c *AlterSequence) WithChildren(children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(c, children...)
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (c *AlterSequence) WithResolvedChildren(children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return c, nil
}
