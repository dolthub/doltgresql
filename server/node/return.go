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
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"
)

// Return represents the statement RETURN statement.
type Return struct {
	Expr     sql.Expression
	exprStmt string
}

var _ sql.ExecSourceRel = (*Return)(nil)
var _ sql.Expressioner = (*Return)(nil)
var _ vitess.Injectable = (*Return)(nil)

// NewReturn creates a new *Return node.
func NewReturn(exprStmt string) *Return {
	return &Return{
		Expr:     nil,
		exprStmt: exprStmt,
	}
}

// Children implements the interface sql.ExecSourceRel.
func (r *Return) Children() []sql.Node {
	return nil
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (r *Return) IsReadOnly() bool {
	return true
}

// Resolved implements the interface sql.ExecSourceRel.
func (r *Return) Resolved() bool {
	if r.Expr == nil {
		return false
	}
	return !r.Expr.Resolved()
}

// RowIter implements the interface sql.ExecSourceRel.
func (r *Return) RowIter(ctx *sql.Context, row sql.Row) (sql.RowIter, error) {
	return nil, errors.Errorf(`cannot call RowIter on Return node`)
}

// String implements the interface sql.ExecSourceRel.
func (r *Return) String() string {
	if r.Expr == nil {
		return fmt.Sprintf("RETURN %s", r.exprStmt)
	}
	return fmt.Sprintf("RETURN %s", r.Expr.String())
}

// Schema implements the interface sql.ExecSourceRel.
func (r *Return) Schema(ctx *sql.Context) sql.Schema {
	return sql.Schema{
		{Name: r.Expr.String(), Type: r.Expr.Type(ctx), Source: ""},
	}
}

// WithChildren implements the interface sql.ExecSourceRel.
func (r *Return) WithChildren(ctx *sql.Context, children ...sql.Node) (sql.Node, error) {
	if len(children) != 0 {
		return nil, sql.ErrInvalidChildrenNumber.New(r, len(children), 0)
	}
	return r, nil
}

// WithResolvedChildren implements the interface sql.ExecSourceRel.
func (r *Return) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	if len(children) != 1 {
		return nil, sql.ErrInvalidChildrenNumber.New(r, len(children), 1)
	}

	nr := *r
	nr.Expr = children[0].(sql.Expression)
	return &nr, nil
}

// Expressions implements the interface sql.Expressioner.
func (r *Return) Expressions() []sql.Expression {
	return []sql.Expression{r.Expr}
}

// WithExpressions implements the interface sql.Expressioner.
func (r *Return) WithExpressions(ctx *sql.Context, exprs ...sql.Expression) (sql.Node, error) {
	if len(exprs) != 1 {
		return nil, sql.ErrInvalidChildrenNumber.New(r, len(exprs), 1)
	}

	nr := *r
	nr.Expr = exprs[0]
	return &nr, nil
}
