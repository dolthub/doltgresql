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

	"github.com/dolthub/go-mysql-server/sql"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// XmlForest represents an XMLFOREST expression, whose children are the values of the named elements.
type XmlForest struct {
	Names    []string
	children []sql.Expression
}

var _ vitess.Injectable = (*XmlForest)(nil)
var _ sql.Expression = (*XmlForest)(nil)

// NewXmlForest returns a new XmlForest for the given SQL identifiers, which are mapped to XML names.
func NewXmlForest(names []string) *XmlForest {
	mappedNames := make([]string, len(names))
	for i, name := range names {
		mappedNames[i] = xmlName(name)
	}
	return &XmlForest{Names: mappedNames}
}

// Children implements the sql.Expression interface.
func (x *XmlForest) Children() []sql.Expression {
	return x.children
}

// Eval implements the sql.Expression interface.
func (x *XmlForest) Eval(ctx *sql.Context, row sql.Row) (any, error) {
	sb := strings.Builder{}
	for i, name := range x.Names {
		str, err := xmlValueToString(ctx, x.children[i], row, true)
		if err != nil {
			return nil, err
		} else if str != nil {
			sb.WriteString("<" + name + ">" + *str + "</" + name + ">")
		}
	}
	return sb.String(), nil
}

// IsNullable implements the sql.Expression interface.
func (x *XmlForest) IsNullable(ctx *sql.Context) bool {
	return false
}

// Resolved implements the sql.Expression interface.
func (x *XmlForest) Resolved() bool {
	return argsResolved(x.children)
}

// String implements the sql.Expression interface.
func (x *XmlForest) String() string {
	return "XMLFOREST(" + argsString(x.children) + ")"
}

// Type implements the sql.Expression interface.
func (x *XmlForest) Type(ctx *sql.Context) sql.Type {
	return pgtypes.Xml
}

// WithChildren implements the sql.Expression interface.
func (x *XmlForest) WithChildren(ctx *sql.Context, children ...sql.Expression) (sql.Expression, error) {
	if len(children) != len(x.Names) {
		return nil, sql.ErrInvalidChildrenNumber.New(x, len(children), len(x.Names))
	}
	return &XmlForest{Names: x.Names, children: children}, nil
}

// WithResolvedChildren implements the vitess.Injectable interface.
func (x *XmlForest) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	return x.WithChildren(ctx.(*sql.Context), injectedChildren(children)...)
}
