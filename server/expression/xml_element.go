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
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"unicode"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/server/xml"
)

// xmlContentEscaper escapes a value placed in the content of an element the way PostgreSQL does.
var xmlContentEscaper = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", "\r", "&#x0d;")

// xmlAttributeEscaper escapes a value placed in an attribute the way libxml2's writer does.
var xmlAttributeEscaper = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "\n", "&#10;", "\t", "&#9;", "\r", "&#13;")

// XmlElement represents an XMLELEMENT expression, whose children are the attribute values followed by the content.
type XmlElement struct {
	Name           string
	AttributeNames []string
	children       []sql.Expression
}

var _ vitess.Injectable = (*XmlElement)(nil)
var _ sql.Expression = (*XmlElement)(nil)

// NewXmlElement returns a new XmlElement for the SQL identifiers `name` and `attributeNames`, which are mapped to XML
// names as PostgreSQL does.
func NewXmlElement(name string, attributeNames []string) (*XmlElement, error) {
	seen := make(map[string]struct{}, len(attributeNames))
	mappedAttributes := make([]string, len(attributeNames))
	for i, attributeName := range attributeNames {
		if _, ok := seen[attributeName]; ok {
			return nil, pgerror.Newf(pgcode.Syntax, `XML attribute name "%s" appears more than once`, attributeName)
		}
		seen[attributeName] = struct{}{}
		mappedAttributes[i] = xmlName(attributeName)
	}
	return &XmlElement{Name: xmlName(name), AttributeNames: mappedAttributes}, nil
}

// Children implements the sql.Expression interface.
func (x *XmlElement) Children() []sql.Expression {
	return x.children
}

// Eval implements the sql.Expression interface.
func (x *XmlElement) Eval(ctx *sql.Context, row sql.Row) (any, error) {
	sb := strings.Builder{}
	sb.WriteString("<" + x.Name)
	for i, attributeName := range x.AttributeNames {
		str, err := xmlValueToString(ctx, x.children[i], row, false)
		if err != nil || str == nil {
			if err != nil {
				return nil, err
			}
			continue
		}
		sb.WriteString(` ` + attributeName + `="` + xmlAttributeEscaper.Replace(*str) + `"`)
	}
	content := strings.Builder{}
	for _, child := range x.children[len(x.AttributeNames):] {
		str, err := xmlValueToString(ctx, child, row, true)
		if err != nil || str == nil {
			if err != nil {
				return nil, err
			}
			continue
		}
		content.WriteString(*str)
	}
	if content.Len() == 0 {
		sb.WriteString("/>")
	} else {
		sb.WriteString(">" + content.String() + "</" + x.Name + ">")
	}
	return sb.String(), nil
}

// IsNullable implements the sql.Expression interface.
func (x *XmlElement) IsNullable(ctx *sql.Context) bool {
	return false
}

// Resolved implements the sql.Expression interface.
func (x *XmlElement) Resolved() bool {
	return argsResolved(x.children)
}

// String implements the sql.Expression interface.
func (x *XmlElement) String() string {
	return "XMLELEMENT(NAME " + x.Name + ", " + argsString(x.children) + ")"
}

// Type implements the sql.Expression interface.
func (x *XmlElement) Type(ctx *sql.Context) sql.Type {
	return pgtypes.Xml
}

// WithChildren implements the sql.Expression interface.
func (x *XmlElement) WithChildren(ctx *sql.Context, children ...sql.Expression) (sql.Expression, error) {
	if len(children) < len(x.AttributeNames) {
		return nil, sql.ErrInvalidChildrenNumber.New(x, len(children), len(x.AttributeNames))
	}
	return &XmlElement{Name: x.Name, AttributeNames: x.AttributeNames, children: children}, nil
}

// WithResolvedChildren implements the vitess.Injectable interface.
func (x *XmlElement) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	return x.WithChildren(ctx.(*sql.Context), injectedChildren(children)...)
}

// xmlName maps the SQL identifier `ident` to an XML name, replacing characters that are not valid in a name with their
// `_xHHHH_` escape the way PostgreSQL does.
func xmlName(ident string) string {
	sb := strings.Builder{}
	for i, r := range ident {
		switch {
		case r == ':' && i == 0:
			sb.WriteString("_x003A_")
		case r == '_' && strings.HasPrefix(ident[i+1:], "x"):
			sb.WriteString("_x005F_")
		case i == 0 && !(unicode.IsLetter(r) || r == '_'),
			i > 0 && !(unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Mc, r) || strings.ContainsRune(".-_:", r)):
			sb.WriteString(fmt.Sprintf("_x%04X_", r))
		default:
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

// xmlValueToString evaluates `expr` and maps its value to the text PostgreSQL places in an element, or nil for NULL.
// Values of type xml are returned verbatim, and other values are escaped when `escape` is set.
func xmlValueToString(ctx *sql.Context, expr sql.Expression, row sql.Row, escape bool) (*string, error) {
	val, err := expr.Eval(ctx, row)
	if err != nil || val == nil {
		return nil, err
	}
	typ, ok := expr.Type(ctx).(*pgtypes.DoltgresType)
	if !ok {
		return nil, errors.Errorf("expected a Doltgres type but found `%T`", expr.Type(ctx))
	}
	str, err := xmlTypedValueToString(ctx, typ, val, escape)
	if err != nil {
		return nil, err
	}
	return &str, nil
}

// xmlTypedValueToString maps `val` of type `typ` to the text PostgreSQL places in an element.
func xmlTypedValueToString(ctx *sql.Context, typ *pgtypes.DoltgresType, val any, escape bool) (string, error) {
	switch {
	case typ.ID == pgtypes.Xml.ID:
		str, err := framework.UnwrapString(ctx, val)
		return xml.Output(str), err
	case typ.IsArrayType():
		sb := strings.Builder{}
		for _, elem := range val.([]any) {
			if elem == nil {
				continue
			}
			str, err := xmlTypedValueToString(ctx, typ.ArrayBaseType(), elem, true)
			if err != nil {
				return "", err
			}
			sb.WriteString("<element>" + str + "</element>")
		}
		return sb.String(), nil
	}
	var str string
	switch typ.ID {
	case pgtypes.Bool.ID:
		str = fmt.Sprintf("%t", val.(bool))
	case pgtypes.Bytea.ID:
		data, err := framework.UnwrapBytes(ctx, val)
		if err != nil {
			return "", err
		}
		xmlBinary, err := ctx.GetSessionVariable(ctx, "xmlbinary")
		if err != nil {
			return "", err
		}
		if xmlBinary.(string) == "hex" {
			str = hex.EncodeToString(data)
		} else {
			str = base64.StdEncoding.EncodeToString(data)
		}
	case pgtypes.Timestamp.ID, pgtypes.TimestampTZ.ID:
		output, err := typ.IoOutput(ctx, val)
		if err != nil {
			return "", err
		}
		str = xmlDateTime(output)
	default:
		output, err := typ.IoOutput(ctx, val)
		if err != nil {
			return "", err
		}
		str = output
	}
	if escape {
		return xmlContentEscaper.Replace(str), nil
	}
	return str, nil
}

// xmlDateTime converts the output form of a timestamp to the XML Schema dateTime form, separating the date and time
// with `T` and adding minutes to an hour-only zone offset.
func xmlDateTime(output string) string {
	if len(output) > 10 && output[10] == ' ' {
		output = output[:10] + "T" + output[11:]
	}
	if len(output) >= 3 && (output[len(output)-3] == '+' || output[len(output)-3] == '-') {
		output += ":00"
	}
	return output
}
