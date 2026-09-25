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

package ast

import (
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	pgexprs "github.com/dolthub/doltgresql/server/expression"
	pgnodes "github.com/dolthub/doltgresql/server/node"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// nodeXmlTable handles *tree.AliasedTableExpr nodes that wrap a *tree.XmlTableExpr, converting them to a call of the
// xmltable table function.
func nodeXmlTable(ctx *Context, node *tree.AliasedTableExpr, xmlTable *tree.XmlTableExpr) (vitess.TableExpr, error) {
	table := &pgnodes.XmlTable{}
	children, err := nodeExprs(ctx, tree.Exprs{xmlTable.RowPath, xmlTable.Document})
	if err != nil {
		return nil, err
	}
	for _, namespace := range xmlTable.Namespaces {
		if namespace.Prefix == "" {
			return nil, pgerror.New(pgcode.FeatureNotSupported, "DEFAULT namespace is not supported")
		}
		uri, err := nodeExpr(ctx, namespace.URI)
		if err != nil {
			return nil, err
		}
		table.Namespaces = append(table.Namespaces, pgnodes.XmlTableNamespace{Prefix: string(namespace.Prefix)})
		children = append(children, uri)
	}
	names := make(map[string]struct{}, len(xmlTable.Columns))
	hasOrdinality := false
	for _, column := range xmlTable.Columns {
		name := string(column.Name)
		if _, ok := names[name]; ok {
			return nil, pgerror.Newf(pgcode.Syntax, `column name "%s" is not unique`, name)
		}
		names[name] = struct{}{}
		if column.ForOrdinality {
			if hasOrdinality {
				return nil, pgerror.New(pgcode.Syntax, "only one FOR ORDINALITY column is allowed")
			}
			hasOrdinality = true
			table.Columns = append(table.Columns, pgnodes.XmlTableColumn{Name: name, Type: pgtypes.Int32, ForOrdinality: true})
			continue
		}
		_, typ, err := nodeResolvableTypeReference(ctx, column.Type, false)
		if err != nil {
			return nil, err
		}
		path := vitess.Expr(vitess.InjectedExpr{Expression: pgexprs.NewUnknownLiteral(name)})
		if column.Path != nil {
			if path, err = nodeExpr(ctx, column.Path); err != nil {
				return nil, err
			}
		}
		defaultValue := vitess.Expr(vitess.InjectedExpr{Expression: pgexprs.NewNullLiteral()})
		if column.Default != nil {
			if defaultValue, err = nodeExpr(ctx, column.Default); err != nil {
				return nil, err
			}
		}
		table.Columns = append(table.Columns, pgnodes.XmlTableColumn{Name: name, Type: typ, NotNull: column.NotNull})
		children = append(children, path, defaultValue)
	}
	columns := make(vitess.Columns, len(node.As.Cols))
	for i, col := range node.As.Cols {
		columns[i] = vitess.NewColIdent(string(col))
	}
	tableFuncExpr := &vitess.TableFuncExpr{
		Name: pgnodes.XmlTableName,
		Exprs: vitess.SelectExprs{&vitess.AliasedExpr{Expr: vitess.InjectedExpr{
			Expression: pgnodes.NewTableFunctionDefinition(table),
			Children:   children,
		}}},
		Alias:   vitess.NewTableIdent(string(node.As.Alias)),
		Columns: columns,
	}
	if node.Lateral {
		return wrapLateralTableFunc(tableFuncExpr), nil
	}
	return tableFuncExpr, nil
}

// nodeXmlElement handles *tree.XmlElement nodes.
func nodeXmlElement(ctx *Context, node *tree.XmlElement) (vitess.Expr, error) {
	attributeNames, children, err := nodeXmlAttributes(ctx, node.Attributes, "attribute")
	if err != nil {
		return nil, err
	}
	content, err := nodeExprs(ctx, node.Content)
	if err != nil {
		return nil, err
	}
	element, err := pgexprs.NewXmlElement(string(node.Name), attributeNames)
	if err != nil {
		return nil, err
	}
	return vitess.InjectedExpr{Expression: element, Children: append(children, content...)}, nil
}

// nodeXmlForest handles *tree.XmlForest nodes.
func nodeXmlForest(ctx *Context, node *tree.XmlForest) (vitess.Expr, error) {
	names, children, err := nodeXmlAttributes(ctx, node.Elements, "element")
	if err != nil {
		return nil, err
	}
	return vitess.InjectedExpr{Expression: pgexprs.NewXmlForest(names), Children: children}, nil
}

// nodeXmlAttributes converts the entries of an XMLATTRIBUTES or XMLFOREST list, deriving the name of an unnamed entry
// from the column it references. `kind` names the entry in errors.
func nodeXmlAttributes(ctx *Context, attributes []tree.XmlAttribute, kind string) ([]string, vitess.Exprs, error) {
	names := make([]string, len(attributes))
	children := make(vitess.Exprs, len(attributes))
	for i, attribute := range attributes {
		names[i] = string(attribute.Name)
		if names[i] == "" {
			columnName, ok := attribute.Expr.(*tree.UnresolvedName)
			if !ok || columnName.Star {
				return nil, nil, pgerror.Newf(pgcode.Syntax, "unnamed XML %s value must be a column reference", kind)
			}
			names[i] = columnName.Parts[0]
		}
		var err error
		if children[i], err = nodeExpr(ctx, attribute.Expr); err != nil {
			return nil, nil, err
		}
	}
	return names, children, nil
}

// nodeXmlPi handles *tree.XmlPi nodes.
func nodeXmlPi(ctx *Context, node *tree.XmlPi) (vitess.Expr, error) {
	pi, err := pgexprs.NewXmlPi(string(node.Name))
	if err != nil {
		return nil, err
	}
	if node.Content == nil {
		return vitess.InjectedExpr{Expression: pi}, nil
	}
	content, err := nodeExpr(ctx, node.Content)
	if err != nil {
		return nil, err
	}
	return vitess.InjectedExpr{Expression: pi, Children: vitess.Exprs{content}}, nil
}

// nodeXmlRoot handles *tree.XmlRoot nodes.
func nodeXmlRoot(ctx *Context, node *tree.XmlRoot) (vitess.Expr, error) {
	root := &pgexprs.XmlRoot{HasVersion: node.Version != nil}
	children, err := nodeExprs(ctx, tree.Exprs{node.Xml})
	if err != nil {
		return nil, err
	}
	if node.Version != nil {
		version, err := nodeExpr(ctx, node.Version)
		if err != nil {
			return nil, err
		}
		children = append(children, version)
	}
	if node.Standalone != tree.XmlRootStandaloneOmitted {
		standalone := ""
		switch node.Standalone {
		case tree.XmlRootStandaloneYes:
			standalone = "yes"
		case tree.XmlRootStandaloneNo:
			standalone = "no"
		}
		root.Standalone = &standalone
	}
	return vitess.InjectedExpr{Expression: root, Children: children}, nil
}

// nodeXmlSerialize handles *tree.XmlSerialize nodes.
func nodeXmlSerialize(ctx *Context, node *tree.XmlSerialize) (vitess.Expr, error) {
	child, err := nodeExpr(ctx, node.Expr)
	if err != nil {
		return nil, err
	}
	_, typ, err := nodeResolvableTypeReference(ctx, node.Type, false)
	if err != nil {
		return nil, err
	}
	return vitess.InjectedExpr{Expression: &pgexprs.XmlSerialize{Document: node.Document, TargetType: typ}, Children: vitess.Exprs{child}}, nil
}
