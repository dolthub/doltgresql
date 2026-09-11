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

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	pgexprs "github.com/dolthub/doltgresql/server/expression"
	"github.com/dolthub/doltgresql/server/functions/framework"
	"github.com/dolthub/doltgresql/server/plpgsql"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// Do executes an anonymous PL/pgSQL code block.
type Do struct {
	Statements []plpgsql.InterpreterOperation
	Runner     pgexprs.StatementRunner
}

var _ sql.ExecSourceRel = (*Do)(nil)
var _ sql.Expressioner = (*Do)(nil)
var _ vitess.Injectable = (*Do)(nil)

// NewDo returns a new anonymous code block.
func NewDo(statements []plpgsql.InterpreterOperation) *Do {
	return &Do{Statements: statements}
}

// Children implements sql.ExecSourceRel.
func (*Do) Children() []sql.Node { return nil }

// Expressions implements sql.Expressioner.
func (d *Do) Expressions() []sql.Expression { return []sql.Expression{d.Runner} }

// IsReadOnly implements sql.ExecSourceRel.
func (*Do) IsReadOnly() bool { return false }

// Resolved implements sql.ExecSourceRel.
func (*Do) Resolved() bool { return true }

// RowIter implements sql.ExecSourceRel.
func (d *Do) RowIter(ctx *sql.Context, _ sql.Row) (sql.RowIter, error) {
	interpreted := framework.InterpretedFunction{
		ReturnType: pgtypes.Void,
		Statements: d.Statements,
	}
	if _, err := plpgsql.Call(ctx, interpreted, d.Runner.Runner, nil, nil); err != nil {
		return nil, err
	}
	return sql.RowsToRowIter(), nil
}

// Schema implements sql.ExecSourceRel.
func (*Do) Schema(*sql.Context) sql.Schema { return nil }

// String implements sql.ExecSourceRel.
func (*Do) String() string { return "DO" }

// WithChildren implements sql.ExecSourceRel.
func (d *Do) WithChildren(ctx *sql.Context, children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(d, children...)
}

// WithExpressions implements sql.Expressioner.
func (d *Do) WithExpressions(_ *sql.Context, exprs ...sql.Expression) (sql.Node, error) {
	if len(exprs) != 1 {
		return nil, errors.Errorf("expected 1 child expression but received %d", len(exprs))
	}
	runner, ok := exprs[0].(pgexprs.StatementRunner)
	if !ok {
		return nil, errors.Errorf("expected statement runner but received %T", exprs[0])
	}
	nd := *d
	nd.Runner = runner
	return &nd, nil
}

// WithResolvedChildren implements vitess.Injectable.
func (d *Do) WithResolvedChildren(context.Context, []any) (any, error) { return d, nil }
