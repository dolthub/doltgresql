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
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// TableFunctionDefinition is the sole argument of a table function whose SQL form is not a function call, such as
// XMLTABLE, carrying the table function that the expression converts to so that the planner resolves its expressions.
type TableFunctionDefinition struct {
	table sql.TableFunction
}

var _ sql.Expression = (*TableFunctionDefinition)(nil)
var _ vitess.Injectable = (*TableFunctionDefinition)(nil)

// NewTableFunctionDefinition returns a new TableFunctionDefinition for `table`.
func NewTableFunctionDefinition(table sql.TableFunction) *TableFunctionDefinition {
	return &TableFunctionDefinition{table: table}
}

// Children implements the interface sql.Expression.
func (d *TableFunctionDefinition) Children() []sql.Expression {
	return d.table.Expressions()
}

// Eval implements the interface sql.Expression.
func (d *TableFunctionDefinition) Eval(ctx *sql.Context, row sql.Row) (any, error) {
	return nil, errors.Errorf("%s may only appear in a FROM clause", strings.ToUpper(d.table.Name()))
}

// IsNullable implements the interface sql.Expression.
func (d *TableFunctionDefinition) IsNullable(ctx *sql.Context) bool {
	return false
}

// Resolved implements the interface sql.Expression.
func (d *TableFunctionDefinition) Resolved() bool {
	return d.table.Resolved()
}

// String implements the interface sql.Expression.
func (d *TableFunctionDefinition) String() string {
	return d.table.Name()
}

// Type implements the interface sql.Expression.
func (d *TableFunctionDefinition) Type(ctx *sql.Context) sql.Type {
	return pgtypes.Unknown
}

// WithChildren implements the interface sql.Expression.
func (d *TableFunctionDefinition) WithChildren(ctx *sql.Context, children ...sql.Expression) (sql.Expression, error) {
	table, err := d.table.WithExpressions(ctx, children...)
	if err != nil {
		return nil, err
	}
	return &TableFunctionDefinition{table: table.(sql.TableFunction)}, nil
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (d *TableFunctionDefinition) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	exprs := make([]sql.Expression, len(children))
	for i, child := range children {
		expr, ok := child.(sql.Expression)
		if !ok {
			return nil, errors.Errorf("expected vitess child to be an expression but has type `%T`", child)
		}
		exprs[i] = expr
	}
	return d.WithChildren(ctx.(*sql.Context), exprs...)
}
