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

	"github.com/dolthub/go-mysql-server/sql"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/server/xml"
)

// XmlIsDocument represents an IS DOCUMENT expression.
type XmlIsDocument struct {
	Expr sql.Expression
}

var _ vitess.Injectable = (*XmlIsDocument)(nil)
var _ sql.Expression = (*XmlIsDocument)(nil)

// Children implements the sql.Expression interface.
func (x *XmlIsDocument) Children() []sql.Expression {
	return []sql.Expression{x.Expr}
}

// Eval implements the sql.Expression interface.
func (x *XmlIsDocument) Eval(ctx *sql.Context, row sql.Row) (any, error) {
	val, err := xmlArgValue(ctx, x.Expr, row)
	if err != nil || val == nil {
		return nil, err
	}
	return xml.CheckWellFormed(*val, true) == nil, nil
}

// IsNullable implements the sql.Expression interface.
func (x *XmlIsDocument) IsNullable(ctx *sql.Context) bool {
	return x.Expr.IsNullable(ctx)
}

// Resolved implements the sql.Expression interface.
func (x *XmlIsDocument) Resolved() bool {
	return x.Expr.Resolved()
}

// String implements the sql.Expression interface.
func (x *XmlIsDocument) String() string {
	return x.Expr.String() + " IS DOCUMENT"
}

// Type implements the sql.Expression interface.
func (x *XmlIsDocument) Type(ctx *sql.Context) sql.Type {
	return pgtypes.Bool
}

// WithChildren implements the sql.Expression interface.
func (x *XmlIsDocument) WithChildren(ctx *sql.Context, children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 1 {
		return nil, sql.ErrInvalidChildrenNumber.New(x, len(children), 1)
	}
	if err := checkXmlArgType(ctx, children[0], "IS DOCUMENT"); err != nil {
		return nil, err
	}
	return &XmlIsDocument{Expr: children[0]}, nil
}

// WithResolvedChildren implements the vitess.Injectable interface.
func (x *XmlIsDocument) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	return x.WithChildren(ctx.(*sql.Context), injectedChildren(children)...)
}
