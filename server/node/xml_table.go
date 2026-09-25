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

package node

import (
	"context"
	"io"
	"strings"

	"github.com/antchfx/xmlquery"
	"github.com/antchfx/xpath"
	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/server/xml"
)

// XmlTableName is the name of the table function that implements XMLTABLE.
const XmlTableName = "xmltable"

// XmlTableNamespace is one entry of the XMLNAMESPACES list of an XMLTABLE expression.
type XmlTableNamespace struct {
	Prefix string
	URI    sql.Expression
}

// XmlTableColumn is one output column of an XMLTABLE expression. Path and Default are nil for a FOR ORDINALITY column,
// and Default evaluates to NULL when the column has no DEFAULT.
type XmlTableColumn struct {
	Name          string
	Type          *pgtypes.DoltgresType
	Path          sql.Expression
	Default       sql.Expression
	NotNull       bool
	ForOrdinality bool
}

// XmlTable is the table function that implements XMLTABLE, producing one row per node matched by RowPath in Document.
type XmlTable struct {
	Namespaces []XmlTableNamespace
	RowPath    sql.Expression
	Document   sql.Expression
	Columns    []XmlTableColumn
	database   sql.Database
}

var _ sql.TableFunction = (*XmlTable)(nil)
var _ sql.ExecSourceRel = (*XmlTable)(nil)

// Children implements the interface sql.TableFunction.
func (x *XmlTable) Children() []sql.Node {
	return nil
}

// Database implements the interface sql.TableFunction.
func (x *XmlTable) Database() sql.Database {
	return x.database
}

// Expressions implements the interface sql.TableFunction.
func (x *XmlTable) Expressions() []sql.Expression {
	exprs := []sql.Expression{x.RowPath, x.Document}
	for _, namespace := range x.Namespaces {
		exprs = append(exprs, namespace.URI)
	}
	for _, column := range x.Columns {
		if !column.ForOrdinality {
			exprs = append(exprs, column.Path, column.Default)
		}
	}
	return exprs
}

// IsReadOnly implements the interface sql.TableFunction.
func (x *XmlTable) IsReadOnly() bool {
	return true
}

// Name implements the interface sql.TableFunction.
func (x *XmlTable) Name() string {
	return XmlTableName
}

// NewInstance implements the interface sql.TableFunction.
func (x *XmlTable) NewInstance(ctx *sql.Context, db sql.Database, args []sql.Expression) (sql.Node, error) {
	if len(args) != 1 {
		return nil, sql.ErrInvalidArgumentNumber.New(XmlTableName, 1, len(args))
	}
	definition, ok := args[0].(*XmlTableDefinition)
	if !ok {
		return nil, errors.Errorf("expected an XMLTABLE definition but found `%T`", args[0])
	}
	if documentType, ok := definition.table.Document.Type(ctx).(*pgtypes.DoltgresType); ok && documentType.ID != pgtypes.Xml.ID && documentType.ID != pgtypes.Unknown.ID {
		return nil, pgerror.Newf(pgcode.DatatypeMismatch, "argument of XMLTABLE must be type xml, not type %s", documentType.String())
	}
	table := *definition.table
	table.database = db
	table.Columns = make([]XmlTableColumn, len(definition.table.Columns))
	for i, column := range definition.table.Columns {
		if !column.Type.IsResolvedType() {
			typeColl, err := core.GetTypesCollectionFromContext(ctx, "")
			if err != nil {
				return nil, err
			}
			if column.Type, err = typeColl.ResolveTypeWithTypmod(ctx, column.Type.ID, column.Type.UnresolvedTypmods); err != nil {
				return nil, err
			}
		}
		table.Columns[i] = column
	}
	return &table, nil
}

// Resolved implements the interface sql.TableFunction.
func (x *XmlTable) Resolved() bool {
	for _, expr := range x.Expressions() {
		if expr == nil || !expr.Resolved() {
			return false
		}
	}
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (x *XmlTable) RowIter(ctx *sql.Context, row sql.Row) (sql.RowIter, error) {
	document, err := x.evalString(ctx, x.Document, row)
	if err != nil || document == nil {
		return sql.RowsToRowIter(), err
	}
	if _, err = pgtypes.Xml.IoInput(ctx, *document); err != nil {
		return nil, err
	}
	doc, err := xml.ParseDocument(*document)
	if err != nil {
		return nil, err
	}
	rowPath, err := x.evalString(ctx, x.RowPath, row)
	if err != nil {
		return nil, err
	} else if rowPath == nil {
		return nil, pgerror.New(pgcode.NullValueNotAllowed, "row filter expression must not be null")
	} else if *rowPath == "" {
		return nil, pgerror.New(pgcode.DataException, "row path filter must not be empty string")
	}
	namespaces := make(map[string]string, len(x.Namespaces))
	for _, namespace := range x.Namespaces {
		uri, err := x.evalString(ctx, namespace.URI, row)
		if err != nil {
			return nil, err
		} else if uri == nil {
			return nil, pgerror.New(pgcode.NullValueNotAllowed, "namespace URI must not be null")
		}
		namespaces[namespace.Prefix] = *uri
	}
	rowExpr, err := xml.Compile(*rowPath, namespaces)
	if err != nil {
		return nil, pgerror.WithCandidateCode(err, pgcode.Syntax)
	}
	columnExprs := make([]*xpath.Expr, len(x.Columns))
	for i, column := range x.Columns {
		if column.ForOrdinality {
			continue
		}
		path, err := x.evalString(ctx, column.Path, row)
		if err != nil {
			return nil, err
		} else if path == nil {
			return nil, pgerror.New(pgcode.NullValueNotAllowed, "column filter expression must not be null")
		} else if *path == "" {
			return nil, pgerror.New(pgcode.DataException, "column path filter must not be empty string")
		}
		if columnExprs[i], err = xml.Compile(*path, namespaces); err != nil {
			return nil, pgerror.WithCandidateCode(err, pgcode.DataException)
		}
	}
	result, err := xml.Evaluate(rowExpr, xmlquery.CreateXPathNavigator(doc))
	if err != nil {
		return nil, err
	}
	nodes, ok := result.(*xpath.NodeIterator)
	if !ok {
		return sql.RowsToRowIter(), nil
	}
	ordinality := int32(0)
	return pgtypes.NewSetReturningFunctionRowIter(func(ctx *sql.Context) (sql.Row, error) {
		if !nodes.MoveNext() {
			return nil, io.EOF
		}
		ordinality++
		rowNode := nodes.Current().Copy()
		outputRow := make(sql.Row, len(x.Columns))
		for i, column := range x.Columns {
			if column.ForOrdinality {
				outputRow[i] = ordinality
				continue
			}
			var err error
			if outputRow[i], err = x.columnValue(ctx, column, columnExprs[i], rowNode.Copy(), row); err != nil {
				return nil, err
			}
		}
		return outputRow, nil
	}), nil
}

// Schema implements the interface sql.TableFunction.
func (x *XmlTable) Schema(ctx *sql.Context) sql.Schema {
	schema := make(sql.Schema, len(x.Columns))
	for i, column := range x.Columns {
		schema[i] = &sql.Column{
			Name:     column.Name,
			Type:     column.Type,
			Nullable: !column.NotNull && !column.ForOrdinality,
			Source:   XmlTableName,
		}
	}
	return schema
}

// String implements the interface sql.TableFunction.
func (x *XmlTable) String() string {
	columns := make([]string, len(x.Columns))
	for i, column := range x.Columns {
		columns[i] = column.Name
	}
	return "XMLTABLE(" + x.RowPath.String() + " PASSING " + x.Document.String() + " COLUMNS " + strings.Join(columns, ", ") + ")"
}

// WithChildren implements the interface sql.TableFunction.
func (x *XmlTable) WithChildren(ctx *sql.Context, children ...sql.Node) (sql.Node, error) {
	if len(children) != 0 {
		return nil, sql.ErrInvalidChildrenNumber.New(x, len(children), 0)
	}
	return x, nil
}

// WithDatabase implements the interface sql.TableFunction.
func (x *XmlTable) WithDatabase(database sql.Database) (sql.Node, error) {
	table := *x
	table.database = database
	return &table, nil
}

// WithExpressions implements the interface sql.TableFunction.
func (x *XmlTable) WithExpressions(ctx *sql.Context, exprs ...sql.Expression) (sql.Node, error) {
	if len(exprs) != len(x.Expressions()) {
		return nil, sql.ErrInvalidChildrenNumber.New(x, len(exprs), len(x.Expressions()))
	}
	table := *x
	table.RowPath, table.Document, exprs = exprs[0], exprs[1], exprs[2:]
	table.Namespaces = make([]XmlTableNamespace, len(x.Namespaces))
	for i, namespace := range x.Namespaces {
		table.Namespaces[i] = XmlTableNamespace{Prefix: namespace.Prefix, URI: exprs[i]}
	}
	exprs = exprs[len(x.Namespaces):]
	table.Columns = make([]XmlTableColumn, len(x.Columns))
	for i, column := range x.Columns {
		table.Columns[i] = column
		if !column.ForOrdinality {
			table.Columns[i].Path, table.Columns[i].Default, exprs = exprs[0], exprs[1], exprs[2:]
		}
	}
	return &table, nil
}

// columnValue returns the value of `column` for the row node `rowNode`, falling back to the column's DEFAULT (evaluated
// against the outer `row`) when the column's XPath expression selects nothing.
func (x *XmlTable) columnValue(ctx *sql.Context, column XmlTableColumn, columnExpr *xpath.Expr, rowNode xpath.NodeNavigator, row sql.Row) (any, error) {
	result, err := xml.Evaluate(columnExpr, rowNode)
	if err != nil {
		return nil, err
	}
	var str *string
	if nodes, ok := result.(*xpath.NodeIterator); ok {
		str, err = x.nodesValue(column, nodes)
		if err != nil {
			return nil, err
		}
	} else {
		scalar := xml.ScalarToString(result)
		str = &scalar
	}
	var value any
	if str != nil {
		if value, err = column.Type.IoInput(ctx, *str); err != nil {
			return nil, err
		}
	} else if value, err = x.defaultValue(ctx, column, row); err != nil {
		return nil, err
	}
	if value == nil && column.NotNull {
		return nil, pgerror.Newf(pgcode.NullValueNotAllowed, `null is not allowed in column "%s"`, column.Name)
	}
	return value, nil
}

// nodesValue returns the text of the nodes selected for `column`, which is nil when no node was selected. A column of
// type xml receives the serialization of every node, while any other type accepts exactly one node's string value.
func (x *XmlTable) nodesValue(column XmlTableColumn, nodes *xpath.NodeIterator) (*string, error) {
	sb := strings.Builder{}
	count := 0
	for nodes.MoveNext() {
		count++
		nav := nodes.Current().(*xmlquery.NodeNavigator)
		if column.Type.ID == pgtypes.Xml.ID {
			sb.WriteString(xml.NodeToXml(nav))
		} else if count > 1 {
			return nil, pgerror.New(pgcode.CardinalityViolation, "more than one value returned by column XPath expression")
		} else if nav.NodeType() != xpath.AttributeNode && nav.Current().Type == xmlquery.CharDataNode {
			sb.WriteString(nav.Current().Data)
		} else {
			sb.WriteString(nav.Value())
		}
	}
	if count == 0 {
		return nil, nil
	}
	str := sb.String()
	return &str, nil
}

// defaultValue evaluates the DEFAULT of `column` against the outer `row` and casts it to the column's type.
func (x *XmlTable) defaultValue(ctx *sql.Context, column XmlTableColumn, row sql.Row) (any, error) {
	value, err := column.Default.Eval(ctx, row)
	if err != nil || value == nil {
		return nil, err
	}
	defaultType, ok := column.Default.Type(ctx).(*pgtypes.DoltgresType)
	if !ok {
		return nil, errors.Errorf("expected a Doltgres type but found `%T`", column.Default.Type(ctx))
	}
	if defaultType.Equals(column.Type) {
		return value, nil
	}
	castsColl, err := core.GetCastsCollectionFromContext(ctx, "")
	if err != nil {
		return nil, err
	}
	cast, err := castsColl.GetAssignmentCast(ctx, defaultType, column.Type)
	if err != nil {
		return nil, err
	}
	if !cast.ID.IsValid() {
		return nil, pgerror.Newf(pgcode.DatatypeMismatch, "argument of XMLTABLE must be type %s, not type %s", column.Type.String(), defaultType.String())
	}
	return cast.Eval(ctx, value, defaultType, column.Type)
}

// evalString evaluates `expr` against `row` and returns the text form of its value, or nil for NULL.
func (x *XmlTable) evalString(ctx *sql.Context, expr sql.Expression, row sql.Row) (*string, error) {
	value, err := expr.Eval(ctx, row)
	if err != nil || value == nil {
		return nil, err
	}
	typ, ok := expr.Type(ctx).(*pgtypes.DoltgresType)
	if !ok {
		return nil, errors.Errorf("expected a Doltgres type but found `%T`", expr.Type(ctx))
	}
	str, err := typ.IoOutput(ctx, value)
	if err != nil {
		return nil, err
	}
	return &str, nil
}

// XmlTableDefinition is the sole argument of the xmltable table function, carrying the XmlTable that an XMLTABLE
// expression converts to so that the planner resolves its expressions.
type XmlTableDefinition struct {
	table *XmlTable
}

var _ sql.Expression = (*XmlTableDefinition)(nil)
var _ vitess.Injectable = (*XmlTableDefinition)(nil)

// NewXmlTableDefinition returns a new XmlTableDefinition for `table`.
func NewXmlTableDefinition(table *XmlTable) *XmlTableDefinition {
	return &XmlTableDefinition{table: table}
}

// Children implements the interface sql.Expression.
func (d *XmlTableDefinition) Children() []sql.Expression {
	return d.table.Expressions()
}

// Eval implements the interface sql.Expression.
func (d *XmlTableDefinition) Eval(ctx *sql.Context, row sql.Row) (any, error) {
	return nil, errors.Errorf("XMLTABLE may only appear in a FROM clause")
}

// IsNullable implements the interface sql.Expression.
func (d *XmlTableDefinition) IsNullable(ctx *sql.Context) bool {
	return false
}

// Resolved implements the interface sql.Expression.
func (d *XmlTableDefinition) Resolved() bool {
	return d.table.Resolved()
}

// String implements the interface sql.Expression.
func (d *XmlTableDefinition) String() string {
	return XmlTableName
}

// Type implements the interface sql.Expression.
func (d *XmlTableDefinition) Type(ctx *sql.Context) sql.Type {
	return pgtypes.Unknown
}

// WithChildren implements the interface sql.Expression.
func (d *XmlTableDefinition) WithChildren(ctx *sql.Context, children ...sql.Expression) (sql.Expression, error) {
	table, err := d.table.WithExpressions(ctx, children...)
	if err != nil {
		return nil, err
	}
	return &XmlTableDefinition{table: table.(*XmlTable)}, nil
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (d *XmlTableDefinition) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	exprs := make([]sql.Expression, len(children))
	for i, child := range children {
		expr, ok := child.(sql.Expression)
		if !ok {
			return nil, errors.Errorf("expected vitess child to be an expression but has type `%T`", child)
		}
		exprs[i] = expr
	}
	return d.WithChildren(ctx.(*sql.Context), exprs...)
}
