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

// JsonFormat represents a FORMAT JSON clause, whose Encoding is empty when no ENCODING is given.
type JsonFormat struct {
	Encoding Name
}

var _ NodeFormatter = (*JsonFormat)(nil)

// Format implements the NodeFormatter interface.
func (node *JsonFormat) Format(ctx *FmtCtx) {
	ctx.WriteString("FORMAT JSON")
	if node.Encoding != "" {
		ctx.WriteString(" ENCODING ")
		ctx.FormatNode(&node.Encoding)
	}
}

// JsonValueExpr represents an expression followed by an optional FORMAT JSON clause, whose JsonFormat is nil when
// omitted.
type JsonValueExpr struct {
	Expr       Expr
	JsonFormat *JsonFormat
}

var _ NodeFormatter = (*JsonValueExpr)(nil)

// Format implements the NodeFormatter interface.
func (node *JsonValueExpr) Format(ctx *FmtCtx) {
	ctx.FormatNode(node.Expr)
	if node.JsonFormat != nil {
		ctx.WriteByte(' ')
		ctx.FormatNode(node.JsonFormat)
	}
}

// JsonArgument represents one entry of a PASSING clause.
type JsonArgument struct {
	Value JsonValueExpr
	Name  Name
}

var _ NodeFormatter = (*JsonArgument)(nil)

// Format implements the NodeFormatter interface.
func (node *JsonArgument) Format(ctx *FmtCtx) {
	ctx.FormatNode(&node.Value)
	ctx.WriteString(" AS ")
	ctx.FormatNode(&node.Name)
}

// JsonBehaviorType is the kind of an ON EMPTY or ON ERROR behavior.
type JsonBehaviorType uint8

const (
	JsonBehaviorError JsonBehaviorType = iota
	JsonBehaviorNull
	JsonBehaviorTrue
	JsonBehaviorFalse
	JsonBehaviorUnknown
	JsonBehaviorEmptyArray
	JsonBehaviorEmptyObject
	JsonBehaviorDefault
)

// String returns the SQL spelling of the behavior, which is used in error messages.
func (b JsonBehaviorType) String() string {
	switch b {
	case JsonBehaviorError:
		return "ERROR"
	case JsonBehaviorNull:
		return "NULL"
	case JsonBehaviorTrue:
		return "TRUE"
	case JsonBehaviorFalse:
		return "FALSE"
	case JsonBehaviorUnknown:
		return "UNKNOWN"
	case JsonBehaviorEmptyArray:
		return "EMPTY ARRAY"
	case JsonBehaviorEmptyObject:
		return "EMPTY OBJECT"
	default:
		return "DEFAULT"
	}
}

// JsonBehavior represents an ON EMPTY or ON ERROR behavior, whose Default is only set for DEFAULT.
type JsonBehavior struct {
	Type    JsonBehaviorType
	Default Expr
}

var _ NodeFormatter = (*JsonBehavior)(nil)

// Format implements the NodeFormatter interface.
func (node *JsonBehavior) Format(ctx *FmtCtx) {
	ctx.WriteString(node.Type.String())
	if node.Type == JsonBehaviorDefault {
		ctx.WriteByte(' ')
		ctx.FormatNode(node.Default)
	}
}

// JsonWrapper is the WRAPPER clause of a JSON_TABLE column.
type JsonWrapper uint8

const (
	JsonWrapperUnspecified JsonWrapper = iota
	JsonWrapperNone
	JsonWrapperConditional
	JsonWrapperUnconditional
)

// JsonQuotes is the QUOTES clause of a JSON_TABLE column.
type JsonQuotes uint8

const (
	JsonQuotesUnspecified JsonQuotes = iota
	JsonQuotesKeep
	JsonQuotesOmit
)

// JsonTableColumnKind is the kind of a JSON_TABLE column.
type JsonTableColumnKind uint8

const (
	JsonTableColumnRegular JsonTableColumnKind = iota
	JsonTableColumnExists
	JsonTableColumnForOrdinality
	JsonTableColumnNested
)

// JsonTableColumn represents one entry of a JSON_TABLE COLUMNS clause. A NESTED column uses Name for its optional path
// name and holds its own Columns, while Path is nil when a regular or EXISTS column omits its PATH.
type JsonTableColumn struct {
	Kind       JsonTableColumnKind
	Name       Name
	Type       ResolvableTypeReference
	JsonFormat *JsonFormat
	Path       Expr
	Wrapper    JsonWrapper
	Quotes     JsonQuotes
	OnEmpty    *JsonBehavior
	OnError    *JsonBehavior
	Columns    []JsonTableColumn
}

var _ NodeFormatter = (*JsonTableColumn)(nil)

// Format implements the NodeFormatter interface.
func (node *JsonTableColumn) Format(ctx *FmtCtx) {
	switch node.Kind {
	case JsonTableColumnForOrdinality:
		ctx.FormatNode(&node.Name)
		ctx.WriteString(" FOR ORDINALITY")
		return
	case JsonTableColumnNested:
		ctx.WriteString("NESTED PATH ")
		ctx.FormatNode(node.Path)
		if node.Name != "" {
			ctx.WriteString(" AS ")
			ctx.FormatNode(&node.Name)
		}
		ctx.WriteString(" COLUMNS ")
		formatJsonTableColumns(ctx, node.Columns)
		return
	}
	ctx.FormatNode(&node.Name)
	ctx.WriteByte(' ')
	ctx.WriteString(node.Type.SQLString())
	if node.Kind == JsonTableColumnExists {
		ctx.WriteString(" EXISTS")
	}
	if node.JsonFormat != nil {
		ctx.WriteByte(' ')
		ctx.FormatNode(node.JsonFormat)
	}
	if node.Path != nil {
		ctx.WriteString(" PATH ")
		ctx.FormatNode(node.Path)
	}
	switch node.Wrapper {
	case JsonWrapperNone:
		ctx.WriteString(" WITHOUT WRAPPER")
	case JsonWrapperConditional:
		ctx.WriteString(" WITH CONDITIONAL WRAPPER")
	case JsonWrapperUnconditional:
		ctx.WriteString(" WITH UNCONDITIONAL WRAPPER")
	}
	switch node.Quotes {
	case JsonQuotesKeep:
		ctx.WriteString(" KEEP QUOTES")
	case JsonQuotesOmit:
		ctx.WriteString(" OMIT QUOTES")
	}
	if node.OnEmpty != nil {
		ctx.WriteByte(' ')
		ctx.FormatNode(node.OnEmpty)
		ctx.WriteString(" ON EMPTY")
	}
	if node.OnError != nil {
		ctx.WriteByte(' ')
		ctx.FormatNode(node.OnError)
		ctx.WriteString(" ON ERROR")
	}
}

// JsonTableExpr represents a JSON_TABLE expression in a FROM clause.
type JsonTableExpr struct {
	Context  JsonValueExpr
	Path     Expr
	PathName Name
	Passing  []JsonArgument
	Columns  []JsonTableColumn
	OnError  *JsonBehavior
}

var _ TableExpr = (*JsonTableExpr)(nil)

// Format implements the NodeFormatter interface.
func (node *JsonTableExpr) Format(ctx *FmtCtx) {
	ctx.WriteString("JSON_TABLE(")
	ctx.FormatNode(&node.Context)
	ctx.WriteString(", ")
	ctx.FormatNode(node.Path)
	if node.PathName != "" {
		ctx.WriteString(" AS ")
		ctx.FormatNode(&node.PathName)
	}
	if len(node.Passing) > 0 {
		ctx.WriteString(" PASSING ")
		for i := range node.Passing {
			if i > 0 {
				ctx.WriteString(", ")
			}
			ctx.FormatNode(&node.Passing[i])
		}
	}
	ctx.WriteString(" COLUMNS ")
	formatJsonTableColumns(ctx, node.Columns)
	if node.OnError != nil {
		ctx.WriteByte(' ')
		ctx.FormatNode(node.OnError)
		ctx.WriteString(" ON ERROR")
	}
	ctx.WriteByte(')')
}

// String implements the fmt.Stringer interface.
func (node *JsonTableExpr) String() string { return AsString(node) }

// tableExpr implements the TableExpr interface.
func (*JsonTableExpr) tableExpr() {}

// formatJsonTableColumns writes the parenthesized COLUMNS list of a JSON_TABLE expression or NESTED column.
func formatJsonTableColumns(ctx *FmtCtx, columns []JsonTableColumn) {
	ctx.WriteByte('(')
	for i := range columns {
		if i > 0 {
			ctx.WriteString(", ")
		}
		ctx.FormatNode(&columns[i])
	}
	ctx.WriteByte(')')
}
