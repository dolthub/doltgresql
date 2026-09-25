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

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/server/xml"
)

// XmlSerialize represents an XMLSERIALIZE expression, which casts the text of its xml child to TargetType.
type XmlSerialize struct {
	Document   bool
	TargetType *pgtypes.DoltgresType
	Expr       sql.Expression
}

var _ vitess.Injectable = (*XmlSerialize)(nil)
var _ sql.Expression = (*XmlSerialize)(nil)

// Children implements the sql.Expression interface.
func (x *XmlSerialize) Children() []sql.Expression {
	return []sql.Expression{x.Expr}
}

// Eval implements the sql.Expression interface.
func (x *XmlSerialize) Eval(ctx *sql.Context, row sql.Row) (any, error) {
	val, err := x.Expr.Eval(ctx, row)
	if err != nil || val == nil {
		return nil, err
	}
	str, err := framework.UnwrapString(ctx, val)
	if err != nil {
		return nil, err
	}
	if x.Expr.Type(ctx).(*pgtypes.DoltgresType).ID == pgtypes.Unknown.ID {
		if _, err = pgtypes.Xml.IoInput(ctx, str); err != nil {
			return nil, err
		}
	}
	if x.Document && xml.CheckWellFormed(str, true) != nil {
		return nil, pgerror.New(pgcode.NotAnXMLDocument, "not an XML document")
	}
	if x.TargetType.Equals(pgtypes.Text) {
		return str, nil
	}
	castsColl, err := core.GetCastsCollectionFromContext(ctx, "")
	if err != nil {
		return nil, err
	}
	cast, err := castsColl.GetImplicitCast(ctx, pgtypes.Text, x.TargetType)
	if err != nil {
		return nil, err
	}
	if !cast.ID.IsValid() {
		return nil, pgerror.Newf(pgcode.CannotCoerce, "cannot cast XMLSERIALIZE result to %s", x.TargetType.String())
	}
	return cast.Eval(ctx, str, pgtypes.Text, x.TargetType)
}

// IsNullable implements the sql.Expression interface.
func (x *XmlSerialize) IsNullable(ctx *sql.Context) bool {
	return x.Expr.IsNullable(ctx)
}

// Resolved implements the sql.Expression interface.
func (x *XmlSerialize) Resolved() bool {
	return x.Expr.Resolved() && x.TargetType.IsResolvedType()
}

// String implements the sql.Expression interface.
func (x *XmlSerialize) String() string {
	return "XMLSERIALIZE(" + x.Expr.String() + " AS " + x.TargetType.String() + ")"
}

// Type implements the sql.Expression interface.
func (x *XmlSerialize) Type(ctx *sql.Context) sql.Type {
	return x.TargetType
}

// WithChildren implements the sql.Expression interface.
func (x *XmlSerialize) WithChildren(ctx *sql.Context, children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 1 {
		return nil, sql.ErrInvalidChildrenNumber.New(x, len(children), 1)
	}
	if err := checkXmlArgType(ctx, children[0], "XMLSERIALIZE"); err != nil {
		return nil, err
	}
	return &XmlSerialize{Document: x.Document, TargetType: x.TargetType, Expr: children[0]}, nil
}

// WithResolvedChildren implements the vitess.Injectable interface.
func (x *XmlSerialize) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	return x.WithChildren(ctx.(*sql.Context), injectedChildren(children)...)
}

// WithType returns a copy of the expression with the given target type.
func (x *XmlSerialize) WithType(typ *pgtypes.DoltgresType) *XmlSerialize {
	return &XmlSerialize{Document: x.Document, TargetType: typ, Expr: x.Expr}
}
