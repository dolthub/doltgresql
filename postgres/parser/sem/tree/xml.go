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

package tree

import (
	"context"

	"github.com/cockroachdb/errors"

	"github.com/dolthub/doltgresql/postgres/parser/types"
)

// XmlParse represents an XMLPARSE expression.
type XmlParse struct {
	Document bool
	Expr     Expr
}

var _ Expr = (*XmlParse)(nil)

// Format implements the NodeFormatter interface.
func (node *XmlParse) Format(ctx *FmtCtx) {
	if ctx.HasFlags(FmtOmitFunctionArgs) {
		ctx.WriteString("xmlparse")
		return
	}
	if node.Document {
		ctx.WriteString("XMLPARSE(DOCUMENT ")
	} else {
		ctx.WriteString("XMLPARSE(CONTENT ")
	}
	ctx.FormatNode(node.Expr)
	ctx.WriteByte(')')
}

// String implements the fmt.Stringer interface.
func (node *XmlParse) String() string { return AsString(node) }

// Walk implements the Expr interface.
func (node *XmlParse) Walk(v Visitor) Expr {
	if expr, changed := WalkExpr(v, node.Expr); changed {
		exprCopy := *node
		exprCopy.Expr = expr
		return &exprCopy
	}
	return node
}

// TypeCheck implements the Expr interface.
func (node *XmlParse) TypeCheck(_ context.Context, _ *SemaContext, _ *types.T) (TypedExpr, error) {
	return nil, errors.New("XMLPARSE cannot be type checked")
}

// XmlConcat represents an XMLCONCAT expression.
type XmlConcat struct {
	Exprs Exprs
}

var _ Expr = (*XmlConcat)(nil)

// Format implements the NodeFormatter interface.
func (node *XmlConcat) Format(ctx *FmtCtx) {
	if ctx.HasFlags(FmtOmitFunctionArgs) {
		ctx.WriteString("xmlconcat")
		return
	}
	ctx.WriteString("XMLCONCAT(")
	ctx.FormatNode(&node.Exprs)
	ctx.WriteByte(')')
}

// String implements the fmt.Stringer interface.
func (node *XmlConcat) String() string { return AsString(node) }

// Walk implements the Expr interface.
func (node *XmlConcat) Walk(v Visitor) Expr {
	if exprs, changed := walkExprSlice(v, node.Exprs); changed {
		exprCopy := *node
		exprCopy.Exprs = exprs
		return &exprCopy
	}
	return node
}

// TypeCheck implements the Expr interface.
func (node *XmlConcat) TypeCheck(_ context.Context, _ *SemaContext, _ *types.T) (TypedExpr, error) {
	return nil, errors.New("XMLCONCAT cannot be type checked")
}

// XmlAttribute represents one entry of an XMLATTRIBUTES list, whose Name is empty when the entry is unnamed.
type XmlAttribute struct {
	Expr Expr
	Name Name
}

// Format implements the NodeFormatter interface.
func (node *XmlAttribute) Format(ctx *FmtCtx) {
	ctx.FormatNode(node.Expr)
	if node.Name != "" {
		ctx.WriteString(" AS ")
		ctx.FormatNode(&node.Name)
	}
}

// XmlElement represents an XMLELEMENT expression.
type XmlElement struct {
	Name       Name
	Attributes []XmlAttribute
	Content    Exprs
}

var _ Expr = (*XmlElement)(nil)

// Format implements the NodeFormatter interface.
func (node *XmlElement) Format(ctx *FmtCtx) {
	if ctx.HasFlags(FmtOmitFunctionArgs) {
		ctx.WriteString("xmlelement")
		return
	}
	ctx.WriteString("XMLELEMENT(NAME ")
	ctx.FormatNode(&node.Name)
	if len(node.Attributes) > 0 {
		ctx.WriteString(", XMLATTRIBUTES(")
		for i := range node.Attributes {
			if i > 0 {
				ctx.WriteString(", ")
			}
			ctx.FormatNode(&node.Attributes[i])
		}
		ctx.WriteByte(')')
	}
	if len(node.Content) > 0 {
		ctx.WriteString(", ")
		ctx.FormatNode(&node.Content)
	}
	ctx.WriteByte(')')
}

// String implements the fmt.Stringer interface.
func (node *XmlElement) String() string { return AsString(node) }

// Walk implements the Expr interface.
func (node *XmlElement) Walk(v Visitor) Expr {
	ret := node
	for i := range node.Attributes {
		if expr, changed := WalkExpr(v, node.Attributes[i].Expr); changed {
			if ret == node {
				exprCopy := *node
				exprCopy.Attributes = append([]XmlAttribute{}, node.Attributes...)
				ret = &exprCopy
			}
			ret.Attributes[i].Expr = expr
		}
	}
	if exprs, changed := walkExprSlice(v, node.Content); changed {
		if ret == node {
			exprCopy := *node
			ret = &exprCopy
		}
		ret.Content = exprs
	}
	return ret
}

// TypeCheck implements the Expr interface.
func (node *XmlElement) TypeCheck(_ context.Context, _ *SemaContext, _ *types.T) (TypedExpr, error) {
	return nil, errors.New("XMLELEMENT cannot be type checked")
}

// XmlNamespace represents one entry of an XMLNAMESPACES list, whose Prefix is empty for the DEFAULT namespace.
type XmlNamespace struct {
	URI    Expr
	Prefix Name
}

// Format implements the NodeFormatter interface.
func (node *XmlNamespace) Format(ctx *FmtCtx) {
	if node.Prefix == "" {
		ctx.WriteString("DEFAULT ")
		ctx.FormatNode(node.URI)
		return
	}
	ctx.FormatNode(node.URI)
	ctx.WriteString(" AS ")
	ctx.FormatNode(&node.Prefix)
}

// XmlTableColumn represents one column definition of an XMLTABLE expression.
type XmlTableColumn struct {
	Name          Name
	Type          ResolvableTypeReference
	Path          Expr
	Default       Expr
	NotNull       bool
	ForOrdinality bool
}

// Format implements the NodeFormatter interface.
func (node *XmlTableColumn) Format(ctx *FmtCtx) {
	ctx.FormatNode(&node.Name)
	if node.ForOrdinality {
		ctx.WriteString(" FOR ORDINALITY")
		return
	}
	ctx.WriteByte(' ')
	ctx.WriteString(node.Type.SQLString())
	if node.Path != nil {
		ctx.WriteString(" PATH ")
		ctx.FormatNode(node.Path)
	}
	if node.Default != nil {
		ctx.WriteString(" DEFAULT ")
		ctx.FormatNode(node.Default)
	}
	if node.NotNull {
		ctx.WriteString(" NOT NULL")
	}
}

// XmlTableExpr represents an XMLTABLE expression in a FROM clause.
type XmlTableExpr struct {
	Namespaces []XmlNamespace
	RowPath    Expr
	Document   Expr
	Columns    []XmlTableColumn
}

var _ TableExpr = (*XmlTableExpr)(nil)

// Format implements the NodeFormatter interface.
func (node *XmlTableExpr) Format(ctx *FmtCtx) {
	ctx.WriteString("XMLTABLE(")
	if len(node.Namespaces) > 0 {
		ctx.WriteString("XMLNAMESPACES(")
		for i := range node.Namespaces {
			if i > 0 {
				ctx.WriteString(", ")
			}
			ctx.FormatNode(&node.Namespaces[i])
		}
		ctx.WriteString("), ")
	}
	ctx.FormatNode(node.RowPath)
	ctx.WriteString(" PASSING ")
	ctx.FormatNode(node.Document)
	ctx.WriteString(" COLUMNS ")
	for i := range node.Columns {
		if i > 0 {
			ctx.WriteString(", ")
		}
		ctx.FormatNode(&node.Columns[i])
	}
	ctx.WriteByte(')')
}

// String implements the fmt.Stringer interface.
func (node *XmlTableExpr) String() string { return AsString(node) }

// tableExpr implements the TableExpr interface.
func (*XmlTableExpr) tableExpr() {}

// XmlForest represents an XMLFOREST expression, whose Elements reuse the XMLATTRIBUTES entry form.
type XmlForest struct {
	Elements []XmlAttribute
}

var _ Expr = (*XmlForest)(nil)

// Format implements the NodeFormatter interface.
func (node *XmlForest) Format(ctx *FmtCtx) {
	if ctx.HasFlags(FmtOmitFunctionArgs) {
		ctx.WriteString("xmlforest")
		return
	}
	ctx.WriteString("XMLFOREST(")
	for i := range node.Elements {
		if i > 0 {
			ctx.WriteString(", ")
		}
		ctx.FormatNode(&node.Elements[i])
	}
	ctx.WriteByte(')')
}

// String implements the fmt.Stringer interface.
func (node *XmlForest) String() string { return AsString(node) }

// Walk implements the Expr interface.
func (node *XmlForest) Walk(v Visitor) Expr {
	ret := node
	for i := range node.Elements {
		if expr, changed := WalkExpr(v, node.Elements[i].Expr); changed {
			if ret == node {
				exprCopy := *node
				exprCopy.Elements = append([]XmlAttribute{}, node.Elements...)
				ret = &exprCopy
			}
			ret.Elements[i].Expr = expr
		}
	}
	return ret
}

// TypeCheck implements the Expr interface.
func (node *XmlForest) TypeCheck(_ context.Context, _ *SemaContext, _ *types.T) (TypedExpr, error) {
	return nil, errors.New("XMLFOREST cannot be type checked")
}

// XmlPi represents an XMLPI expression, whose Content is nil when omitted.
type XmlPi struct {
	Name    Name
	Content Expr
}

var _ Expr = (*XmlPi)(nil)

// Format implements the NodeFormatter interface.
func (node *XmlPi) Format(ctx *FmtCtx) {
	if ctx.HasFlags(FmtOmitFunctionArgs) {
		ctx.WriteString("xmlpi")
		return
	}
	ctx.WriteString("XMLPI(NAME ")
	ctx.FormatNode(&node.Name)
	if node.Content != nil {
		ctx.WriteString(", ")
		ctx.FormatNode(node.Content)
	}
	ctx.WriteByte(')')
}

// String implements the fmt.Stringer interface.
func (node *XmlPi) String() string { return AsString(node) }

// Walk implements the Expr interface.
func (node *XmlPi) Walk(v Visitor) Expr {
	if node.Content == nil {
		return node
	}
	if expr, changed := WalkExpr(v, node.Content); changed {
		exprCopy := *node
		exprCopy.Content = expr
		return &exprCopy
	}
	return node
}

// TypeCheck implements the Expr interface.
func (node *XmlPi) TypeCheck(_ context.Context, _ *SemaContext, _ *types.T) (TypedExpr, error) {
	return nil, errors.New("XMLPI cannot be type checked")
}

// XmlRootStandalone is the STANDALONE option of an XMLROOT expression.
type XmlRootStandalone int

const (
	XmlRootStandaloneOmitted XmlRootStandalone = iota
	XmlRootStandaloneYes
	XmlRootStandaloneNo
	XmlRootStandaloneNoValue
)

// XmlRoot represents an XMLROOT expression, whose Version is nil for VERSION NO VALUE.
type XmlRoot struct {
	Xml        Expr
	Version    Expr
	Standalone XmlRootStandalone
}

var _ Expr = (*XmlRoot)(nil)

// Format implements the NodeFormatter interface.
func (node *XmlRoot) Format(ctx *FmtCtx) {
	if ctx.HasFlags(FmtOmitFunctionArgs) {
		ctx.WriteString("xmlroot")
		return
	}
	ctx.WriteString("XMLROOT(")
	ctx.FormatNode(node.Xml)
	ctx.WriteString(", VERSION ")
	if node.Version == nil {
		ctx.WriteString("NO VALUE")
	} else {
		ctx.FormatNode(node.Version)
	}
	switch node.Standalone {
	case XmlRootStandaloneYes:
		ctx.WriteString(", STANDALONE YES")
	case XmlRootStandaloneNo:
		ctx.WriteString(", STANDALONE NO")
	case XmlRootStandaloneNoValue:
		ctx.WriteString(", STANDALONE NO VALUE")
	}
	ctx.WriteByte(')')
}

// String implements the fmt.Stringer interface.
func (node *XmlRoot) String() string { return AsString(node) }

// Walk implements the Expr interface.
func (node *XmlRoot) Walk(v Visitor) Expr {
	ret := node
	if expr, changed := WalkExpr(v, node.Xml); changed {
		exprCopy := *node
		exprCopy.Xml = expr
		ret = &exprCopy
	}
	if node.Version != nil {
		if expr, changed := WalkExpr(v, node.Version); changed {
			if ret == node {
				exprCopy := *node
				ret = &exprCopy
			}
			ret.Version = expr
		}
	}
	return ret
}

// TypeCheck implements the Expr interface.
func (node *XmlRoot) TypeCheck(_ context.Context, _ *SemaContext, _ *types.T) (TypedExpr, error) {
	return nil, errors.New("XMLROOT cannot be type checked")
}

// XmlSerialize represents an XMLSERIALIZE expression.
type XmlSerialize struct {
	Document bool
	Expr     Expr
	Type     ResolvableTypeReference
}

var _ Expr = (*XmlSerialize)(nil)

// Format implements the NodeFormatter interface.
func (node *XmlSerialize) Format(ctx *FmtCtx) {
	if ctx.HasFlags(FmtOmitFunctionArgs) {
		ctx.WriteString("xmlserialize")
		return
	}
	if node.Document {
		ctx.WriteString("XMLSERIALIZE(DOCUMENT ")
	} else {
		ctx.WriteString("XMLSERIALIZE(CONTENT ")
	}
	ctx.FormatNode(node.Expr)
	ctx.WriteString(" AS ")
	ctx.WriteString(node.Type.SQLString())
	ctx.WriteByte(')')
}

// String implements the fmt.Stringer interface.
func (node *XmlSerialize) String() string { return AsString(node) }

// Walk implements the Expr interface.
func (node *XmlSerialize) Walk(v Visitor) Expr {
	if expr, changed := WalkExpr(v, node.Expr); changed {
		exprCopy := *node
		exprCopy.Expr = expr
		return &exprCopy
	}
	return node
}

// TypeCheck implements the Expr interface.
func (node *XmlSerialize) TypeCheck(_ context.Context, _ *SemaContext, _ *types.T) (TypedExpr, error) {
	return nil, errors.New("XMLSERIALIZE cannot be type checked")
}

// XmlIsDocument represents an IS DOCUMENT expression.
type XmlIsDocument struct {
	Expr Expr
}

var _ Expr = (*XmlIsDocument)(nil)

// Format implements the NodeFormatter interface.
func (node *XmlIsDocument) Format(ctx *FmtCtx) {
	ctx.FormatNode(node.Expr)
	ctx.WriteString(" IS DOCUMENT")
}

// String implements the fmt.Stringer interface.
func (node *XmlIsDocument) String() string { return AsString(node) }

// Walk implements the Expr interface.
func (node *XmlIsDocument) Walk(v Visitor) Expr {
	if expr, changed := WalkExpr(v, node.Expr); changed {
		exprCopy := *node
		exprCopy.Expr = expr
		return &exprCopy
	}
	return node
}

// TypeCheck implements the Expr interface.
func (node *XmlIsDocument) TypeCheck(_ context.Context, _ *SemaContext, _ *types.T) (TypedExpr, error) {
	return nil, errors.New("IS DOCUMENT cannot be type checked")
}
