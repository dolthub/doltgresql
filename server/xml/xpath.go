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

package xml

import (
	"math"
	"strconv"
	"strings"

	"github.com/antchfx/xmlquery"
	"github.com/antchfx/xpath"
	"github.com/cockroachdb/errors"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
)

// textEscaper escapes the characters that may not appear literally in XML text.
var textEscaper = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")

// attributeEscaper escapes the characters that may not appear literally in a double-quoted XML attribute value.
var attributeEscaper = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;")

// EscapeText escapes the characters of `str` that may not appear literally in XML text.
func EscapeText(str string) string {
	return textEscaper.Replace(str)
}

// Compile compiles the XPath expression `expr` using the prefix-to-URI map `namespaces`.
func Compile(expr string, namespaces map[string]string) (*xpath.Expr, error) {
	if len(expr) == 0 {
		return nil, pgerror.New(pgcode.DataException, "empty XPath expression")
	}
	compiled, err := xpath.CompileWithNS(expr, namespaces)
	if err != nil {
		return nil, errors.Errorf("invalid XPath expression")
	}
	return compiled, nil
}

// ParseDocument parses `str` as an XML document.
func ParseDocument(str string) (*xmlquery.Node, error) {
	if err := CheckWellFormed(str, true); err != nil {
		return nil, pgerror.New(pgcode.InvalidXMLDocument, "could not parse XML document")
	}
	doc, err := xmlquery.Parse(strings.NewReader(str))
	if err != nil {
		return nil, pgerror.New(pgcode.InvalidXMLDocument, "could not parse XML document")
	}
	return doc, nil
}

// Evaluate evaluates `expr` with `nav` as the context node, returning a `*xpath.NodeIterator`, bool, float64, or string.
func Evaluate(expr *xpath.Expr, nav xpath.NodeNavigator) (result any, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.Errorf("could not evaluate XPath expression: %v", r)
		}
	}()
	return expr.Evaluate(nav), nil
}

// ScalarToString formats a non-node XPath result the way the XPath `string` function does.
func ScalarToString(result any) string {
	switch result := result.(type) {
	case bool:
		return strconv.FormatBool(result)
	case float64:
		if math.IsInf(result, 1) {
			return "Infinity"
		} else if math.IsInf(result, -1) {
			return "-Infinity"
		}
		return strconv.FormatFloat(result, 'f', -1, 64)
	default:
		return result.(string)
	}
}

// NodeToXml converts a node selected by an XPath expression into the text of an xml value.
func NodeToXml(nav *xmlquery.NodeNavigator) string {
	if nav.NodeType() == xpath.AttributeNode {
		return textEscaper.Replace(nav.Value())
	}
	node := nav.Current()
	if node.Type == xmlquery.TextNode {
		return textEscaper.Replace(node.Data)
	}
	sb := &strings.Builder{}
	if node.Type == xmlquery.DocumentNode {
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if child.Type != xmlquery.DeclarationNode {
				writeNode(sb, child, true)
			}
		}
		sb.WriteString("\n")
	} else {
		writeNode(sb, node, true)
	}
	return sb.String()
}

// writeNode serializes `node` and its descendants, declaring the namespace of a `root` element that inherits it from
// an ancestor.
func writeNode(sb *strings.Builder, node *xmlquery.Node, root bool) {
	switch node.Type {
	case xmlquery.TextNode:
		sb.WriteString(textEscaper.Replace(node.Data))
	case xmlquery.CharDataNode:
		sb.WriteString("<![CDATA[" + node.Data + "]]>")
	case xmlquery.CommentNode:
		sb.WriteString("<!--" + node.Data + "-->")
	case xmlquery.NotationNode:
		sb.WriteString("<!" + node.Data + ">")
	case xmlquery.DeclarationNode:
		sb.WriteString("<?" + node.Data)
		writeAttributes(sb, node.Attr)
		sb.WriteString("?>")
	case xmlquery.ElementNode:
		name := node.Data
		if node.Prefix != "" {
			name = node.Prefix + ":" + node.Data
		}
		sb.WriteString("<" + name)
		if root && node.NamespaceURI != "" && !declaresNamespace(node) {
			if node.Prefix == "" {
				sb.WriteString(` xmlns="` + attributeEscaper.Replace(node.NamespaceURI) + `"`)
			} else {
				sb.WriteString(" xmlns:" + node.Prefix + `="` + attributeEscaper.Replace(node.NamespaceURI) + `"`)
			}
		}
		writeAttributes(sb, node.Attr)
		if node.FirstChild == nil {
			sb.WriteString("/>")
			return
		}
		sb.WriteString(">")
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			writeNode(sb, child, false)
		}
		sb.WriteString("</" + name + ">")
	}
}

// writeAttributes serializes `attrs` in order, each preceded by a space.
func writeAttributes(sb *strings.Builder, attrs []xmlquery.Attr) {
	for _, attr := range attrs {
		sb.WriteString(" ")
		if attr.Name.Space != "" {
			sb.WriteString(attr.Name.Space + ":")
		}
		sb.WriteString(attr.Name.Local + `="` + attributeEscaper.Replace(attr.Value) + `"`)
	}
}

// declaresNamespace returns whether the attributes of `node` declare the namespace that `node` belongs to.
func declaresNamespace(node *xmlquery.Node) bool {
	for _, attr := range node.Attr {
		if (node.Prefix == "" && attr.Name.Space == "" && attr.Name.Local == "xmlns") ||
			(node.Prefix != "" && attr.Name.Space == "xmlns" && attr.Name.Local == node.Prefix) {
			return true
		}
	}
	return false
}
