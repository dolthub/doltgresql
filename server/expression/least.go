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
	"github.com/dolthub/doltgresql/core/casts"
	"github.com/dolthub/go-mysql-server/sql"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// Least represents a LEAST expression.
type Least struct {
	casts   []casts.Cast
	retType *pgtypes.DoltgresType
	Args    []sql.Expression
}

var _ vitess.Injectable = (*Least)(nil)
var _ sql.Expression = (*Least)(nil)

// Children implements the sql.Expression interface.
func (n *Least) Children() []sql.Expression {
	return n.Args
}

// Eval implements the sql.Expression interface.
func (n *Least) Eval(ctx *sql.Context, row sql.Row) (any, error) {
	return evalGreatestLeast(ctx, row, n.Args, n.retType, n.casts, -1)
}

// IsNullable implements the sql.Expression interface.
func (n *Least) IsNullable(ctx *sql.Context) bool {
	return true
}

// Resolved implements the sql.Expression interface.
func (n *Least) Resolved() bool {
	return argsResolved(n.Args) && n.retType != nil
}

// String implements the sql.Expression interface.
func (n *Least) String() string {
	return "LEAST(" + argsString(n.Args) + ")"
}

// Type implements the sql.Expression interface.
func (n *Least) Type(ctx *sql.Context) sql.Type {
	if n.retType == nil {
		return pgtypes.Unknown
	}
	return n.retType
}

// WithChildren implements the sql.Expression interface.
func (n *Least) WithChildren(ctx *sql.Context, children ...sql.Expression) (sql.Expression, error) {
	if len(children) == 0 {
		return nil, sql.ErrInvalidArgumentNumber.New("LEAST", "1 or more", 0)
	}

	retType, castList, err := getRetTypeAndCasts(ctx, children)
	if err != nil {
		return nil, err
	}

	return &Least{
		casts:   castList,
		retType: retType,
		Args:    children,
	}, nil
}

// WithResolvedChildren implements the vitess.InjectableExpression interface.
func (n *Least) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	args := make([]sql.Expression, len(children))
	for i, child := range children {
		expr, ok := child.(sql.Expression)
		if !ok {
			return nil, errors.Errorf("expected vitess child to be an expression but has type `%T`", child)
		}
		args[i] = expr
	}
	return n.WithChildren(ctx.(*sql.Context), args...)
}
