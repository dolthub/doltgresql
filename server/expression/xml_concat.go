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

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/server/xml"
)

// XmlConcat represents an XMLCONCAT expression.
type XmlConcat struct {
	Args []sql.Expression
}

var _ vitess.Injectable = (*XmlConcat)(nil)
var _ sql.Expression = (*XmlConcat)(nil)

// Children implements the sql.Expression interface.
func (x *XmlConcat) Children() []sql.Expression {
	return x.Args
}

// Eval implements the sql.Expression interface. Declarations are dropped from the values, and the result carries one
// only when every value agrees on a version other than 1.0 or any value declares standalone, as PostgreSQL does.
func (x *XmlConcat) Eval(ctx *sql.Context, row sql.Row) (any, error) {
	values := make([]string, 0, len(x.Args))
	for _, arg := range x.Args {
		val, err := xmlArgValue(ctx, arg, row)
		if err != nil {
			return nil, err
		} else if val != nil {
			values = append(values, *val)
		}
	}
	if len(values) == 0 {
		return nil, nil
	}
	return xml.Concat(values...), nil
}

// IsNullable implements the sql.Expression interface.
func (x *XmlConcat) IsNullable(ctx *sql.Context) bool {
	return true
}

// Resolved implements the sql.Expression interface.
func (x *XmlConcat) Resolved() bool {
	return argsResolved(x.Args)
}

// String implements the sql.Expression interface.
func (x *XmlConcat) String() string {
	return "XMLCONCAT(" + argsString(x.Args) + ")"
}

// Type implements the sql.Expression interface.
func (x *XmlConcat) Type(ctx *sql.Context) sql.Type {
	return pgtypes.Xml
}

// WithChildren implements the sql.Expression interface.
func (x *XmlConcat) WithChildren(ctx *sql.Context, children ...sql.Expression) (sql.Expression, error) {
	for _, child := range children {
		if err := checkXmlArgType(ctx, child, "XMLCONCAT"); err != nil {
			return nil, err
		}
	}
	return &XmlConcat{Args: children}, nil
}

// WithResolvedChildren implements the vitess.Injectable interface.
func (x *XmlConcat) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	return x.WithChildren(ctx.(*sql.Context), injectedChildren(children)...)
}

// checkXmlArgType returns PostgreSQL's error when `expr` is neither of type xml nor an untyped literal.
func checkXmlArgType(ctx *sql.Context, expr sql.Expression, syntax string) error {
	typ, ok := expr.Type(ctx).(*pgtypes.DoltgresType)
	if !ok {
		if nullType, isNull := expr.Type(ctx).(sql.NullType); isNull && nullType.IsNullType() {
			return nil
		}
		return errors.Errorf("expected a Doltgres type but found `%T`", expr.Type(ctx))
	}
	if typ.ID != pgtypes.Xml.ID && typ.ID != pgtypes.Unknown.ID {
		return pgerror.Newf(pgcode.DatatypeMismatch, "argument of %s must be type xml, not type %s", syntax, typ.String())
	}
	return nil
}

// xmlArgValue evaluates `expr`, which has passed checkXmlArgType, returning its xml value or nil for NULL.
func xmlArgValue(ctx *sql.Context, expr sql.Expression, row sql.Row) (*string, error) {
	str, err := xmlArgToString(ctx, expr, row)
	if err != nil || str == nil {
		return nil, err
	}
	if expr.Type(ctx).(*pgtypes.DoltgresType).ID == pgtypes.Unknown.ID {
		if _, err = pgtypes.Xml.IoInput(ctx, *str); err != nil {
			return nil, err
		}
	}
	return str, nil
}
