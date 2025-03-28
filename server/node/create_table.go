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

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/rowexec"

	"github.com/dolthub/doltgresql/core"
)

// CreateTable is a node that implements functionality specifically relevant to Doltgres' table creation needs.
type CreateTable struct {
	gmsCreateTable *plan.CreateTable
	sequences      []*CreateSequence
}

var _ sql.ExecSourceRel = (*CreateTable)(nil)
var _ sql.SchemaTarget = (*CreateTable)(nil)
var _ sql.Expressioner = (*CreateTable)(nil)

// NewCreateTable returns a new *CreateTable.
func NewCreateTable(createTable *plan.CreateTable, sequences []*CreateSequence) *CreateTable {
	return &CreateTable{
		gmsCreateTable: createTable,
		sequences:      sequences,
	}
}

// Children implements the interface sql.ExecSourceRel.
func (c *CreateTable) Children() []sql.Node {
	return c.gmsCreateTable.Children()
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (c *CreateTable) IsReadOnly() bool {
	return false
}

// Resolved implements the interface sql.ExecSourceRel.
func (c *CreateTable) Resolved() bool {
	return c.gmsCreateTable != nil && c.gmsCreateTable.Resolved()
}

// RowIter implements the interface sql.ExecSourceRel.
func (c *CreateTable) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	// Prevent tables from having names like `guid()`, which resembles a function
	leftParen := strings.IndexByte(c.gmsCreateTable.Name(), '(')
	rightParen := strings.IndexByte(c.gmsCreateTable.Name(), ')')
	if leftParen != -1 && rightParen != -1 && rightParen > leftParen {
		return nil, fmt.Errorf("table name `%s` cannot contain a parenthesized portion", c.gmsCreateTable.Name())
	}

	createTableIter, err := rowexec.DefaultBuilder.Build(ctx, c.gmsCreateTable, r)
	if err != nil {
		return nil, err
	}

	schemaName, err := core.GetSchemaName(ctx, c.gmsCreateTable.Db, "")
	if err != nil {
		return nil, err
	}
	for _, sequence := range c.sequences {
		sequence.schema = schemaName
		_, err = sequence.RowIter(ctx, r)
		if err != nil {
			_ = createTableIter.Close(ctx)
			return nil, err
		}
	}
	return createTableIter, err
}

// Schema implements the interface sql.ExecSourceRel.
func (c *CreateTable) Schema() sql.Schema {
	return c.gmsCreateTable.Schema()
}

// String implements the interface sql.ExecSourceRel.
func (c *CreateTable) String() string {
	return c.gmsCreateTable.String()
}

// DebugString implements the sql.DebugStringer interface
func (c *CreateTable) DebugString() string {
	return sql.DebugString(c.gmsCreateTable)
}

// WithChildren implements the interface sql.ExecSourceRel.
func (c *CreateTable) WithChildren(children ...sql.Node) (sql.Node, error) {
	gmsCreateTable, err := c.gmsCreateTable.WithChildren(children...)
	if err != nil {
		return nil, err
	}
	return &CreateTable{
		gmsCreateTable: gmsCreateTable.(*plan.CreateTable),
		sequences:      c.sequences,
	}, nil
}

func (c *CreateTable) TargetSchema() sql.Schema {
	return c.gmsCreateTable.TargetSchema()
}

func (c CreateTable) WithTargetSchema(schema sql.Schema) (sql.Node, error) {
	n, err := c.gmsCreateTable.WithTargetSchema(schema)
	if err != nil {
		return nil, err
	}

	c.gmsCreateTable = n.(*plan.CreateTable)

	return &c, nil
}

// Expressions implements the sql.Expressioner interface.
func (c *CreateTable) Expressions() []sql.Expression {
	return c.gmsCreateTable.Expressions()
}

// WithExpressions implements the sql.Expressioner interface.
func (c *CreateTable) WithExpressions(expression ...sql.Expression) (sql.Node, error) {
	nc := *c
	n, err := nc.gmsCreateTable.WithExpressions(expression...)
	if err != nil {
		return nil, err
	}

	nc.gmsCreateTable = n.(*plan.CreateTable)
	return &nc, nil
}
