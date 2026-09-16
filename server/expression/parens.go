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

package expression

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"
)

// Parens represents a parenthesized expression whose parentheses must survive in its string form.
type Parens struct {
	child sql.Expression
}

var _ vitess.Injectable = (*Parens)(nil)
var _ sql.Expression = (*Parens)(nil)

// NewParens returns a new *Parens.
func NewParens() *Parens {
	return &Parens{}
}

// Children implements the sql.Expression interface.
func (p *Parens) Children() []sql.Expression {
	return []sql.Expression{p.child}
}

// Eval implements the sql.Expression interface.
func (p *Parens) Eval(ctx *sql.Context, row sql.Row) (any, error) {
	return p.child.Eval(ctx, row)
}

// IsNullable implements the sql.Expression interface.
func (p *Parens) IsNullable(ctx *sql.Context) bool {
	return p.child.IsNullable(ctx)
}

// Resolved implements the sql.Expression interface.
func (p *Parens) Resolved() bool {
	return p.child != nil && p.child.Resolved()
}

// String implements the sql.Expression interface.
func (p *Parens) String() string {
	if p.child == nil {
		return "(?)"
	}
	return "(" + p.child.String() + ")"
}

// Type implements the sql.Expression interface.
func (p *Parens) Type(ctx *sql.Context) sql.Type {
	return p.child.Type(ctx)
}

// WithChildren implements the sql.Expression interface.
func (p *Parens) WithChildren(ctx *sql.Context, children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 1 {
		return nil, sql.ErrInvalidChildrenNumber.New(p, len(children), 1)
	}
	return &Parens{
		child: children[0],
	}, nil
}

// WithResolvedChildren implements the vitess.InjectableExpression interface.
func (p *Parens) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	if len(children) != 1 {
		return nil, errors.Errorf("invalid vitess child count, expected `1` but got `%d`", len(children))
	}
	child, ok := children[0].(sql.Expression)
	if !ok {
		return nil, errors.Errorf("expected vitess child to be an expression but has type `%T`", children[0])
	}
	return &Parens{
		child: child,
	}, nil
}
