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
	"math"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/sequences"
	pgexprs "github.com/dolthub/doltgresql/server/expression"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// CreateSequence handles the CREATE SEQUENCE statement, along with SERIAL type definitions.
type CreateSequence struct {
	schema      string
	ifNotExists bool
	fromAlter   bool
	sequence    *sequences.Sequence
}

var _ sql.ExecSourceRel = (*CreateSequence)(nil)
var _ vitess.Injectable = (*CreateSequence)(nil)

// NewCreateSequence returns a new *CreateSequence.
func NewCreateSequence(ifNotExists bool, schema string, fromAlter bool, sequence *sequences.Sequence) *CreateSequence {
	return &CreateSequence{
		schema:      schema,
		ifNotExists: ifNotExists,
		fromAlter:   fromAlter,
		sequence:    sequence,
	}
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
	if strings.HasPrefix(strings.ToLower(c.sequence.Id.SequenceName()), "dolt") {
		return nil, errors.Errorf("sequences cannot be prefixed with 'dolt'")
	}
	schema, err := core.GetSchemaName(ctx, nil, c.schema)
	if err != nil {
		return nil, err
	}
	// The sequence won't have the schema filled in, so we have to do that now
	c.sequence.Id = id.NewSequence(schema, c.sequence.Id.SequenceName())

	// Check that the sequence name is free
	relationType, err := core.GetRelationType(ctx, schema, c.sequence.Id.SequenceName())
	if err != nil {
		return nil, err
	}
	if relationType != core.RelationType_DoesNotExist && c.ifNotExists {
		if c.ifNotExists {
			// TODO: issue a notice
			return sql.RowsToRowIter(), nil
		}
		return nil, errors.Errorf(`relation "%s" already exists`, c.sequence.Id)
	}
	// Check that the OWNED BY is valid, if it exists
	var table sql.Table
	var tableSch sql.Schema
	var tableColumn *sql.Column
	if c.sequence.OwnerTable.IsValid() {
		// The table will only have its name set, so we need to fill in the schema as well
		c.sequence.OwnerTable = id.NewTable(schema, c.sequence.OwnerTable.TableName())
		relationType, err = core.GetRelationType(ctx, schema, c.sequence.OwnerTable.TableName())
		if err != nil {
			return nil, err
		}
		if relationType == core.RelationType_DoesNotExist {
			return nil, errors.Errorf(`relation "%s" does not exist`, c.sequence.OwnerTable.TableName())
		} else if relationType != core.RelationType_Table {
			return nil, errors.Errorf(`sequence cannot be owned by relation "%s"`, c.sequence.OwnerTable.TableName())
		}

		table, err = core.GetSqlTableFromContext(ctx, "", doltdb.TableName{Name: c.sequence.OwnerTable.TableName(), Schema: schema})
		if err != nil {
			return nil, err
		}
		if table == nil {
			return nil, errors.Errorf(`table "%s" cannot be found but says it exists`, c.sequence.OwnerTable.TableName())
		}
		tableSch = table.Schema()
		for _, col := range tableSch {
			if col.Name == c.sequence.OwnerColumn {
				tableColumn = col.Copy()
				break
			}
		}
		if tableColumn == nil {
			return nil, errors.Errorf(`column "%s" of relation "%s" does not exist`,
				c.sequence.OwnerColumn, c.sequence.OwnerTable.TableName())
		}
		// If this is from an ALTER TABLE statement, then we have to adjust the type since we didn't have that information earlier
		if c.fromAlter {
			dgType, ok := tableColumn.Type.(*pgtypes.DoltgresType)
			if !ok {
				return nil, errors.Errorf(`column "%s" of relation "%s" has unexpected type: "%s"`,
					c.sequence.OwnerColumn, c.sequence.OwnerTable.TableName(), tableColumn.Type.String())
			}
			switch dgType.ID {
			case pgtypes.Int16.ID:
				c.sequence.DataTypeID = pgtypes.Int16.ID
				if c.sequence.Minimum < int64(math.MinInt16) {
					c.sequence.Minimum = int64(math.MinInt16)
				}
				if c.sequence.Maximum > int64(math.MaxInt16) {
					c.sequence.Maximum = int64(math.MaxInt16)
				}
			case pgtypes.Int32.ID:
				c.sequence.DataTypeID = pgtypes.Int32.ID
				if c.sequence.Minimum < int64(math.MinInt32) {
					c.sequence.Minimum = int64(math.MinInt32)
				}
				if c.sequence.Maximum > int64(math.MaxInt32) {
					c.sequence.Maximum = int64(math.MaxInt32)
				}
			case pgtypes.Int64.ID:
				c.sequence.DataTypeID = pgtypes.Int64.ID
				// Minimum and Maximum are already set to the Int64 values
			default:
				// Not sure what we should do if we encounter a non-integer type, so we'll error for now
				return nil, errors.Errorf(`column "%s" of relation "%s" has an unsupported type: "%s"`,
					c.sequence.OwnerColumn, c.sequence.OwnerTable.TableName(), tableColumn.Type.String())
			}
		}
	}
	// Create the sequence since we know it's completely valid
	// TODO: we always create the sequence in the current database, but there's no need to require this, and in fact we
	//  need to support it to create a sequence on another branch than the current one
	collection, err := core.GetSequencesCollectionFromContext(ctx, ctx.GetCurrentDatabase())
	if err != nil {
		return nil, err
	}
	if err = collection.CreateSequence(ctx, c.sequence); err != nil {
		return nil, err
	}
	if c.fromAlter {
		if tableColumn == nil {
			// This check is to satisfy the linter
			return nil, errors.New(`somehow unable to find the matching table column which should have failed a previous step`)
		}
		// TODO: What happens when we add a sequence to a column that already has a default value? We'll error until we find out
		if tableColumn.Default != nil || tableColumn.Generated != nil {
			return nil, errors.New(`cannot add a sequence to a column that already has a default value`)
		}
		alterableTable, ok := table.(*sqle.AlterableDoltTable)
		if !ok {
			return nil, errors.Errorf(`expected a Dolt table but received "%T"`, table)
		}
		// TODO: Do we need to convert to a TableName and then call String? Are we reliant on the specific way it's formatted?
		//  This is how it's done in the analyzer for SERIAL types, so assuming it's for a good reason.
		seqName := doltdb.TableName{Name: c.sequence.Id.SequenceName(), Schema: c.sequence.Id.SchemaName()}.String()
		nextVal, foundFunc, err := framework.GetFunction("nextval", pgexprs.NewTextLiteral(seqName))
		if err != nil {
			return nil, err
		}
		if !foundFunc {
			return nil, errors.Errorf(`function "nextval" could not be found`)
		}
		tableColumn.Default = &sql.ColumnDefaultValue{
			Expr:          nextVal,
			OutType:       tableColumn.Type.(*pgtypes.DoltgresType),
			Literal:       false,
			ReturnNil:     false,
			Parenthesized: false,
		}
		err = alterableTable.ModifyColumn(ctx, tableColumn.Name, tableColumn, nil)
		if err != nil {
			return nil, err
		}
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
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return c, nil
}
