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
	"strings"
	"unicode"

	"github.com/dolthub/go-mysql-server/sql"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// XmlPi represents an XMLPI expression, whose optional child is the content of the processing instruction.
type XmlPi struct {
	Name     string
	children []sql.Expression
}

var _ vitess.Injectable = (*XmlPi)(nil)
var _ sql.Expression = (*XmlPi)(nil)

// NewXmlPi returns a new XmlPi targeting the SQL identifier `name`, which is mapped to an XML name.
func NewXmlPi(name string) (*XmlPi, error) {
	if strings.EqualFold(name, "xml") {
		return nil, pgerror.Newf(pgcode.Syntax, `invalid XML processing instruction: XML processing instruction target name cannot be "%s"`, name)
	}
	return &XmlPi{Name: xmlName(name)}, nil
}

// Children implements the sql.Expression interface.
func (x *XmlPi) Children() []sql.Expression {
	return x.children
}

// Eval implements the sql.Expression interface.
func (x *XmlPi) Eval(ctx *sql.Context, row sql.Row) (any, error) {
	if len(x.children) == 0 {
		return "<?" + x.Name + "?>", nil
	}
	str, err := xmlArgToString(ctx, x.children[0], row)
	if err != nil || str == nil {
		return nil, err
	}
	if strings.Contains(*str, "?>") {
		return nil, pgerror.New(pgcode.InvalidXMLProcessingInstruction, `invalid XML processing instruction: XML processing instruction cannot contain "?>"`)
	}
	return "<?" + x.Name + " " + strings.TrimLeftFunc(*str, unicode.IsSpace) + "?>", nil
}

// IsNullable implements the sql.Expression interface.
func (x *XmlPi) IsNullable(ctx *sql.Context) bool {
	return len(x.children) > 0
}

// Resolved implements the sql.Expression interface.
func (x *XmlPi) Resolved() bool {
	return argsResolved(x.children)
}

// String implements the sql.Expression interface.
func (x *XmlPi) String() string {
	if len(x.children) == 0 {
		return "XMLPI(NAME " + x.Name + ")"
	}
	return "XMLPI(NAME " + x.Name + ", " + x.children[0].String() + ")"
}

// Type implements the sql.Expression interface.
func (x *XmlPi) Type(ctx *sql.Context) sql.Type {
	return pgtypes.Xml
}

// WithChildren implements the sql.Expression interface.
func (x *XmlPi) WithChildren(ctx *sql.Context, children ...sql.Expression) (sql.Expression, error) {
	if len(children) > 1 {
		return nil, sql.ErrInvalidChildrenNumber.New(x, len(children), 1)
	}
	return &XmlPi{Name: x.Name, children: children}, nil
}

// WithResolvedChildren implements the vitess.Injectable interface.
func (x *XmlPi) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	return x.WithChildren(ctx.(*sql.Context), injectedChildren(children)...)
}
