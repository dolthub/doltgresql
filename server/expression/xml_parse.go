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

	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/server/xml"
)

// XmlParse represents an XMLPARSE expression.
type XmlParse struct {
	Document bool
	Expr     sql.Expression
}

var _ vitess.Injectable = (*XmlParse)(nil)
var _ sql.Expression = (*XmlParse)(nil)

// Children implements the sql.Expression interface.
func (x *XmlParse) Children() []sql.Expression {
	return []sql.Expression{x.Expr}
}

// Eval implements the sql.Expression interface.
func (x *XmlParse) Eval(ctx *sql.Context, row sql.Row) (any, error) {
	str, err := xmlArgToString(ctx, x.Expr, row)
	if err != nil || str == nil {
		return nil, err
	}
	if err = xml.CheckWellFormed(*str, x.Document); err != nil {
		return nil, err
	}
	return *str, nil
}

// IsNullable implements the sql.Expression interface.
func (x *XmlParse) IsNullable(ctx *sql.Context) bool {
	return true
}

// Resolved implements the sql.Expression interface.
func (x *XmlParse) Resolved() bool {
	return x.Expr != nil && x.Expr.Resolved()
}

// String implements the sql.Expression interface.
func (x *XmlParse) String() string {
	if x.Document {
		return "XMLPARSE(DOCUMENT " + x.Expr.String() + ")"
	}
	return "XMLPARSE(CONTENT " + x.Expr.String() + ")"
}

// Type implements the sql.Expression interface.
func (x *XmlParse) Type(ctx *sql.Context) sql.Type {
	return pgtypes.Xml
}

// WithChildren implements the sql.Expression interface.
func (x *XmlParse) WithChildren(ctx *sql.Context, children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 1 {
		return nil, sql.ErrInvalidChildrenNumber.New(x, len(children), 1)
	}
	return &XmlParse{Document: x.Document, Expr: children[0]}, nil
}

// WithResolvedChildren implements the vitess.Injectable interface.
func (x *XmlParse) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	return x.WithChildren(ctx.(*sql.Context), injectedChildren(children)...)
}

// injectedChildren converts the children given to a vitess.Injectable into expressions.
func injectedChildren(children []any) []sql.Expression {
	exprs := make([]sql.Expression, len(children))
	for i, child := range children {
		exprs[i] = child.(sql.Expression)
	}
	return exprs
}

// xmlArgToString evaluates `expr` for `row` and returns the text form of its value, or nil for NULL.
func xmlArgToString(ctx *sql.Context, expr sql.Expression, row sql.Row) (*string, error) {
	val, err := expr.Eval(ctx, row)
	if err != nil || val == nil {
		return nil, err
	}
	typ, ok := expr.Type(ctx).(*pgtypes.DoltgresType)
	if !ok {
		return nil, errors.Errorf("expected a Doltgres type but found `%T`", expr.Type(ctx))
	}
	str, err := typ.IoOutput(ctx, val)
	if err != nil {
		return nil, err
	}
	return &str, nil
}
