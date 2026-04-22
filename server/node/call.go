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
	"context"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core"
	pgexprs "github.com/dolthub/doltgresql/server/expression"
	"github.com/dolthub/doltgresql/server/functions/framework"
)

// Call is used to call stored procedures.
type Call struct {
	SchemaName    string
	ProcedureName string
	Exprs         []sql.Expression
	Runner        pgexprs.StatementRunner
	cachedSch     sql.Schema
	originalExprs vitess.Exprs
	CompiledFunc  *framework.CompiledFunction
}

var _ sql.ExecSourceRel = (*Call)(nil)
var _ sql.Expressioner = (*Call)(nil)
var _ vitess.Injectable = (*Call)(nil)

// NewCall returns a new *Call.
func NewCall(schema string, name string, originalExprs vitess.Exprs) *Call {
	return &Call{
		SchemaName:    schema,
		ProcedureName: name,
		Exprs:         nil,
		originalExprs: originalExprs,
	}
}

// Children implements the interface sql.ExecSourceRel.
func (c *Call) Children() []sql.Node {
	return nil
}

// Expressions implements the interface sql.Expressioner.
func (c *Call) Expressions() []sql.Expression {
	exprs := make([]sql.Expression, len(c.Exprs)+1)
	exprs[0] = c.Runner
	copy(exprs[1:], c.Exprs)
	return exprs
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (c *Call) IsReadOnly() bool {
	return false
}

// Resolved implements the interface sql.ExecSourceRel.
func (c *Call) Resolved() bool {
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (c *Call) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	if c.CompiledFunc == nil || !c.CompiledFunc.Resolved() {
		return nil, errors.New("cannot call unresolved procedure")
	}
	if !core.IsContextValid(ctx) {
		return nil, errors.New("invalid context while attempting to call a procedure")
	}

	cf := c.CompiledFunc.SetStatementRunner(ctx, c.Runner.Runner).(*framework.CompiledFunction)
	_, err := cf.Eval(ctx, nil)
	if err != nil {
		return nil, err
	}
	return sql.RowsToRowIter(), nil
}

// Schema implements the interface sql.ExecSourceRel.
func (c *Call) Schema(ctx *sql.Context) sql.Schema {
	// TODO: this should be the INOUT and OUT parameters of the target procedure assuming we're not using the cached schema
	return c.cachedSch
}

// String implements the interface sql.ExecSourceRel.
func (c *Call) String() string {
	return "CALL"
}

// WithChildren implements the interface sql.ExecSourceRel.
func (c *Call) WithChildren(ctx *sql.Context, children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(c, children...)
}

// WithExpressions implements the interface sql.Expressioner.
func (c *Call) WithExpressions(ctx *sql.Context, exprs ...sql.Expression) (sql.Node, error) {
	if len(c.Exprs)+1 != len(exprs) {
		return nil, errors.Errorf("expected `%d` child expressions but received `%d`", len(c.Exprs), len(exprs))
	}
	nc := *c
	nc.Runner = exprs[0].(pgexprs.StatementRunner)
	nc.Exprs = exprs[1:]
	return &nc, nil
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (c *Call) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	resolvedChildren := make([]sql.Expression, len(children))
	for i, child := range children {
		var ok bool
		resolvedChildren[i], ok = child.(sql.Expression)
		if !ok {
			return nil, errors.Errorf("expected vitess child to be an expression but has type `%T`", child)
		}
	}
	nc := *c
	nc.Exprs = resolvedChildren
	return &nc, nil
}
