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

// XmlRoot represents an XMLROOT expression. Its children are the xml value and, when HasVersion is set, the version.
// Standalone is nil when the STANDALONE option is omitted, and otherwise holds "yes", "no", or "" for NO VALUE.
type XmlRoot struct {
	HasVersion bool
	Standalone *string
	children   []sql.Expression
}

var _ vitess.Injectable = (*XmlRoot)(nil)
var _ sql.Expression = (*XmlRoot)(nil)

// Children implements the sql.Expression interface.
func (x *XmlRoot) Children() []sql.Expression {
	return x.children
}

// Eval implements the sql.Expression interface.
func (x *XmlRoot) Eval(ctx *sql.Context, row sql.Row) (any, error) {
	val, err := xmlArgValue(ctx, x.children[0], row)
	if err != nil || val == nil {
		return nil, err
	}
	_, standalone, rest := xml.SplitDeclaration(*val)
	version := ""
	if x.HasVersion {
		newVersion, err := xmlArgToString(ctx, x.children[1], row)
		if err != nil {
			return nil, err
		} else if newVersion != nil {
			version = *newVersion
		}
	}
	if x.Standalone != nil {
		standalone = *x.Standalone
	}
	return xml.Declaration(version, standalone) + rest, nil
}

// IsNullable implements the sql.Expression interface.
func (x *XmlRoot) IsNullable(ctx *sql.Context) bool {
	return x.children[0].IsNullable(ctx)
}

// Resolved implements the sql.Expression interface.
func (x *XmlRoot) Resolved() bool {
	return argsResolved(x.children)
}

// String implements the sql.Expression interface.
func (x *XmlRoot) String() string {
	return "XMLROOT(" + argsString(x.children) + ")"
}

// Type implements the sql.Expression interface.
func (x *XmlRoot) Type(ctx *sql.Context) sql.Type {
	return pgtypes.Xml
}

// WithChildren implements the sql.Expression interface.
func (x *XmlRoot) WithChildren(ctx *sql.Context, children ...sql.Expression) (sql.Expression, error) {
	count := 1
	if x.HasVersion {
		count = 2
	}
	if len(children) != count {
		return nil, sql.ErrInvalidChildrenNumber.New(x, len(children), count)
	}
	if err := checkXmlArgType(ctx, children[0], "XMLROOT"); err != nil {
		return nil, err
	}
	return &XmlRoot{HasVersion: x.HasVersion, Standalone: x.Standalone, children: children}, nil
}

// WithResolvedChildren implements the vitess.Injectable interface.
func (x *XmlRoot) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	return x.WithChildren(ctx.(*sql.Context), injectedChildren(children)...)
}
