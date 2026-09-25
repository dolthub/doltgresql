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
	"fmt"
	"strings"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/theory/sqljson/path"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	pgnodes "github.com/dolthub/doltgresql/server/node"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// nodeJsonTable handles *tree.AliasedTableExpr nodes that wrap a *tree.JsonTableExpr, converting them to a call of the
// json_table table function.
func nodeJsonTable(ctx *Context, node *tree.AliasedTableExpr, jsonTable *tree.JsonTableExpr) (vitess.TableExpr, error) {
	if jsonTable.OnError != nil && jsonTable.OnError.Type != tree.JsonBehaviorError && jsonTable.OnError.Type != tree.JsonBehaviorEmptyArray {
		return nil, pgerror.New(pgcode.Syntax, "invalid ON ERROR behavior")
	}
	table := &pgnodes.JsonTable{ErrorOnError: jsonTable.OnError != nil && jsonTable.OnError.Type == tree.JsonBehaviorError}
	var err error
	if table.ContextFormat, err = nodeJsonFormat(jsonTable.Context.JsonFormat); err != nil {
		return nil, err
	}
	children, err := nodeExprs(ctx, tree.Exprs{jsonTable.Context.Expr})
	if err != nil {
		return nil, err
	}
	for _, argument := range jsonTable.Passing {
		format, err := nodeJsonFormat(argument.Value.JsonFormat)
		if err != nil {
			return nil, err
		}
		value, err := nodeExpr(ctx, argument.Value.Expr)
		if err != nil {
			return nil, err
		}
		table.Passing = append(table.Passing, pgnodes.JsonTablePassing{Name: string(argument.Name), Format: format})
		children = append(children, value)
	}
	names := make(map[string]struct{})
	if table.Root, err = nodeJsonTablePath(ctx, table, names, &children, jsonTable.Path, jsonTable.PathName, jsonTable.Columns); err != nil {
		return nil, err
	}
	columns := make(vitess.Columns, len(node.As.Cols))
	for i, col := range node.As.Cols {
		columns[i] = vitess.NewColIdent(string(col))
	}
	tableFuncExpr := &vitess.TableFuncExpr{
		Name: pgnodes.JsonTableName,
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

// nodeJsonTablePath converts a row path of a JSON_TABLE expression, appending its columns to `table` and their DEFAULT
// expressions to `children`.
func nodeJsonTablePath(ctx *Context, table *pgnodes.JsonTable, names map[string]struct{}, children *vitess.Exprs, pathExpr tree.Expr, pathName tree.Name, columns []tree.JsonTableColumn) (*pgnodes.JsonTablePath, error) {
	rowPath, err := nodeJsonPath(pathExpr)
	if err != nil {
		return nil, err
	}
	if err = addJsonTableName(names, pathName); err != nil {
		return nil, err
	}
	tablePath := &pgnodes.JsonTablePath{Path: rowPath}
	hasOrdinality := false
	for _, column := range columns {
		if column.Kind == tree.JsonTableColumnNested {
			nested, err := nodeJsonTablePath(ctx, table, names, children, column.Path, column.Name, column.Columns)
			if err != nil {
				return nil, err
			}
			tablePath.Nested = append(tablePath.Nested, nested)
			continue
		}
		if err = addJsonTableName(names, column.Name); err != nil {
			return nil, err
		}
		tableColumn := pgnodes.JsonTableColumn{
			Name:    string(column.Name),
			Kind:    column.Kind,
			Wrapper: column.Wrapper,
			Quotes:  column.Quotes,
		}
		switch column.Kind {
		case tree.JsonTableColumnForOrdinality:
			if hasOrdinality {
				return nil, pgerror.New(pgcode.Syntax, "only one FOR ORDINALITY column is allowed")
			}
			hasOrdinality = true
			tableColumn.Type = pgtypes.Int32
		case tree.JsonTableColumnExists:
			if column.OnError != nil {
				switch column.OnError.Type {
				case tree.JsonBehaviorError, tree.JsonBehaviorTrue, tree.JsonBehaviorFalse, tree.JsonBehaviorUnknown:
				default:
					return nil, pgerror.Newf(pgcode.Syntax, `invalid ON ERROR behavior for column "%s"`, column.Name)
				}
			}
			fallthrough
		default:
			if column.Quotes != tree.JsonQuotesUnspecified && (column.Wrapper == tree.JsonWrapperConditional || column.Wrapper == tree.JsonWrapperUnconditional) {
				return nil, pgerror.New(pgcode.Syntax, "SQL/JSON QUOTES behavior must not be specified when WITH WRAPPER is used")
			}
			if tableColumn.Format, err = nodeJsonFormat(column.JsonFormat); err != nil {
				return nil, err
			}
			if _, tableColumn.Type, err = nodeResolvableTypeReference(ctx, column.Type, false); err != nil {
				return nil, err
			}
			columnPath := column.Path
			if columnPath == nil {
				columnPath = tree.NewStrVal(`$."` + strings.ReplaceAll(strings.ReplaceAll(string(column.Name), `\`, `\\`), `"`, `\"`) + `"`)
			}
			if tableColumn.Path, err = nodeJsonPath(columnPath); err != nil {
				return nil, err
			}
			if tableColumn.OnEmpty, err = nodeJsonTableBehavior(ctx, children, column.OnEmpty); err != nil {
				return nil, err
			}
			if tableColumn.OnError, err = nodeJsonTableBehavior(ctx, children, column.OnError); err != nil {
				return nil, err
			}
		}
		tablePath.Columns = append(tablePath.Columns, len(table.Columns))
		table.Columns = append(table.Columns, tableColumn)
	}
	return tablePath, nil
}

// nodeJsonTableBehavior converts an ON EMPTY or ON ERROR behavior, appending its DEFAULT expression to `children`.
func nodeJsonTableBehavior(ctx *Context, children *vitess.Exprs, behavior *tree.JsonBehavior) (*pgnodes.JsonTableBehavior, error) {
	if behavior == nil {
		return nil, nil
	}
	if behavior.Type == tree.JsonBehaviorDefault {
		defaultExpr, err := nodeExpr(ctx, behavior.Default)
		if err != nil {
			return nil, err
		}
		*children = append(*children, defaultExpr)
	}
	return &pgnodes.JsonTableBehavior{Type: behavior.Type}, nil
}

// nodeJsonFormat validates the ENCODING of a FORMAT JSON clause.
func nodeJsonFormat(format *tree.JsonFormat) (*tree.JsonFormat, error) {
	if format == nil || format.Encoding == "" {
		return format, nil
	}
	switch strings.ToLower(string(format.Encoding)) {
	case "utf8":
		return format, nil
	case "utf16", "utf32":
		return nil, pgerror.New(pgcode.FeatureNotSupported, "unsupported JSON encoding")
	default:
		return nil, pgerror.Newf(pgcode.InvalidParameterValue, "unrecognized JSON encoding: %s", format.Encoding)
	}
}

// addJsonTableName adds `name` to `names`, returning an error when it is already present.
func addJsonTableName(names map[string]struct{}, name tree.Name) error {
	if name == "" {
		return nil
	}
	if _, ok := names[string(name)]; ok {
		return pgerror.Newf(pgcode.DuplicateAlias, "duplicate JSON_TABLE column or path name: %s", name)
	}
	names[string(name)] = struct{}{}
	return nil
}

// nodeJsonPath parses the SQL/JSON path of a JSON_TABLE row path or column, which must be a string constant.
func nodeJsonPath(expr tree.Expr) (*path.Path, error) {
	str, ok := expr.(*tree.StrVal)
	if !ok {
		return nil, pgerror.New(pgcode.FeatureNotSupported, "only string constants are supported in JSON_TABLE path specification")
	}
	jsonPath := str.RawString()
	parsed, err := path.Parse(jsonPath)
	if err == nil {
		return parsed, nil
	}
	if strings.TrimSpace(jsonPath) == "" {
		return nil, pgerror.Newf(pgcode.InvalidTextRepresentation, `invalid input syntax for type jsonpath: "%s"`, jsonPath)
	}
	msg := strings.TrimPrefix(err.Error(), "path: parser: ")
	var line, column int
	if idx := strings.LastIndex(msg, " at "); idx == -1 || !strings.HasPrefix(msg, "syntax error") {
		return nil, pgerror.New(pgcode.Syntax, msg)
	} else if _, scanErr := fmt.Sscanf(msg[idx+4:], "%d:%d", &line, &column); scanErr != nil {
		return nil, pgerror.New(pgcode.Syntax, msg)
	}
	lines := strings.Split(jsonPath, "\n")
	if line >= 1 && line <= len(lines) && column >= 2 && column <= len([]rune(lines[line-1])) {
		return nil, pgerror.Newf(pgcode.Syntax, `syntax error at or near "%c" of jsonpath input`, []rune(lines[line-1])[column-2])
	}
	return nil, pgerror.New(pgcode.Syntax, "syntax error at end of jsonpath input")
}
