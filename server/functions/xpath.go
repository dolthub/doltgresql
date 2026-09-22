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

package functions

import (
	"github.com/antchfx/xmlquery"
	"github.com/antchfx/xpath"
	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/server/xml"
)

// initXpath registers the functions to the catalog.
func initXpath() {
	framework.RegisterFunction(xpath_text_xml)
	framework.RegisterFunction(xpath_text_xml_textarray)
}

// xpath_text_xml represents the PostgreSQL function of the same name, taking the same parameters.
var xpath_text_xml = framework.Function2{
	Name:       "xpath",
	Return:     pgtypes.XmlArray,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Xml},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		return xpath_text_xml_textarray.Callable(ctx, [4]*pgtypes.DoltgresType{}, val1, val2, []any{})
	},
}

// xpath_text_xml_textarray represents the PostgreSQL function of the same name, taking the same parameters.
var xpath_text_xml_textarray = framework.Function3{
	Name:       "xpath",
	Return:     pgtypes.XmlArray,
	Parameters: [3]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Xml, pgtypes.TextArray},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, val1 any, val2 any, val3 any) (any, error) {
		result, err := evaluateXpath(ctx, val1, val2, val3)
		if err != nil {
			return nil, err
		}
		if nodes, ok := result.(*xpath.NodeIterator); ok {
			values := []any{}
			for nodes.MoveNext() {
				values = append(values, xml.NodeToXml(nodes.Current().(*xmlquery.NodeNavigator)))
			}
			return values, nil
		}
		return []any{xml.EscapeText(xml.ScalarToString(result))}, nil
	},
}

// evaluateXpath evaluates the XPath expression `exprVal` against the document `xmlVal` using the namespace array
// `namespacesVal`, returning either a `*xpath.NodeIterator` or a scalar bool, float64, or string.
func evaluateXpath(ctx *sql.Context, exprVal any, xmlVal any, namespacesVal any) (any, error) {
	exprStr, err := framework.UnwrapString(ctx, exprVal)
	if err != nil {
		return nil, err
	}
	xmlStr, err := framework.UnwrapString(ctx, xmlVal)
	if err != nil {
		return nil, err
	}
	namespaces, err := xpathNamespaces(ctx, namespacesVal.([]any))
	if err != nil {
		return nil, err
	}
	expr, err := xml.Compile(exprStr, namespaces)
	if err != nil {
		return nil, err
	}
	doc, err := xml.ParseDocument(xmlStr)
	if err != nil {
		return nil, err
	}
	return xml.Evaluate(expr, xmlquery.CreateXPathNavigator(doc))
}

// xpathNamespaces converts the namespace array of `xpath` and `xpath_exists` into a prefix-to-URI map.
func xpathNamespaces(ctx *sql.Context, vals []any) (map[string]string, error) {
	if len(vals)%2 != 0 {
		return nil, errors.Errorf("invalid array for XML namespace mapping")
	}
	namespaces := make(map[string]string, len(vals)/2)
	for i := 0; i < len(vals); i += 2 {
		if vals[i] == nil || vals[i+1] == nil {
			return nil, errors.Errorf("neither namespace name nor URI may be null")
		}
		prefix, err := framework.UnwrapString(ctx, vals[i])
		if err != nil {
			return nil, err
		}
		uri, err := framework.UnwrapString(ctx, vals[i+1])
		if err != nil {
			return nil, err
		}
		namespaces[prefix] = uri
	}
	return namespaces, nil
}
