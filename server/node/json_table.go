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
	"encoding/json"
	"io"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"
	"github.com/theory/sqljson/path"
	"github.com/theory/sqljson/path/exec"
	"github.com/theory/sqljson/path/types"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	pgexprs "github.com/dolthub/doltgresql/server/expression"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// JsonTableName is the name of the table function that implements JSON_TABLE.
const JsonTableName = "json_table"

// JsonTableBehavior is an ON EMPTY or ON ERROR behavior of a JSON_TABLE column, whose Default is only set for DEFAULT.
type JsonTableBehavior struct {
	Type    tree.JsonBehaviorType
	Default sql.Expression
}

// JsonTableColumn is one output column of a JSON_TABLE expression.
type JsonTableColumn struct {
	Name      string
	Type      *pgtypes.DoltgresType
	Kind      tree.JsonTableColumnKind
	Path      *path.Path
	Format    *tree.JsonFormat
	Wrapper   tree.JsonWrapper
	Quotes    tree.JsonQuotes
	OnEmpty   *JsonTableBehavior
	OnError   *JsonTableBehavior
	formatted bool
}

// JsonTablePath is a row path of a JSON_TABLE expression, whose Columns are indexes into the table's columns.
type JsonTablePath struct {
	Path    *path.Path
	Columns []int
	Nested  []*JsonTablePath
}

// JsonTablePassing is one entry of the PASSING clause of a JSON_TABLE expression.
type JsonTablePassing struct {
	Name   string
	Value  sql.Expression
	Format *tree.JsonFormat
}

// JsonTable is the table function that implements JSON_TABLE, producing one row per item matched by Root in Context.
type JsonTable struct {
	Context       sql.Expression
	ContextFormat *tree.JsonFormat
	Passing       []JsonTablePassing
	Root          *JsonTablePath
	Columns       []JsonTableColumn
	ErrorOnError  bool
	database      sql.Database
}

var _ sql.TableFunction = (*JsonTable)(nil)
var _ sql.ExecSourceRel = (*JsonTable)(nil)

// Children implements the interface sql.TableFunction.
func (j *JsonTable) Children() []sql.Node {
	return nil
}

// Database implements the interface sql.TableFunction.
func (j *JsonTable) Database() sql.Database {
	return j.database
}

// Expressions implements the interface sql.TableFunction.
func (j *JsonTable) Expressions() []sql.Expression {
	exprs := []sql.Expression{j.Context}
	for _, passing := range j.Passing {
		exprs = append(exprs, passing.Value)
	}
	for _, column := range j.Columns {
		for _, behavior := range []*JsonTableBehavior{column.OnEmpty, column.OnError} {
			if behavior != nil && behavior.Type == tree.JsonBehaviorDefault {
				exprs = append(exprs, behavior.Default)
			}
		}
	}
	return exprs
}

// IsReadOnly implements the interface sql.TableFunction.
func (j *JsonTable) IsReadOnly() bool {
	return true
}

// Name implements the interface sql.TableFunction.
func (j *JsonTable) Name() string {
	return JsonTableName
}

// NewInstance implements the interface sql.TableFunction.
func (j *JsonTable) NewInstance(ctx *sql.Context, db sql.Database, args []sql.Expression) (sql.Node, error) {
	if len(args) != 1 {
		return nil, sql.ErrInvalidArgumentNumber.New(JsonTableName, 1, len(args))
	}
	var jsonTable *JsonTable
	definition, ok := args[0].(*TableFunctionDefinition)
	if ok {
		jsonTable, ok = definition.table.(*JsonTable)
	}
	if !ok {
		return nil, errors.Errorf("expected a JSON_TABLE definition but found `%T`", args[0])
	}
	table := *jsonTable
	table.database = db
	contextType, err := jsonTableInputType(ctx, table.Context, table.ContextFormat)
	if err != nil {
		return nil, err
	}
	switch {
	case contextType.ID == pgtypes.Json.ID, contextType.ID == pgtypes.JsonB.ID, contextType.ID == pgtypes.Unknown.ID, contextType.IsStringType():
	case contextType.ID == pgtypes.Bytea.ID && table.ContextFormat != nil:
	default:
		return nil, pgerror.Newf(pgcode.CannotCoerce, "cannot cast type %s to jsonb", contextType.String())
	}
	table.Passing = make([]JsonTablePassing, len(jsonTable.Passing))
	for i, passing := range jsonTable.Passing {
		if table.Passing[i], err = newJsonTablePassing(ctx, passing); err != nil {
			return nil, err
		}
	}
	table.Columns = make([]JsonTableColumn, len(jsonTable.Columns))
	for i, column := range jsonTable.Columns {
		if table.Columns[i], err = newJsonTableColumn(ctx, column); err != nil {
			return nil, err
		}
	}
	return &table, nil
}

// Resolved implements the interface sql.TableFunction.
func (j *JsonTable) Resolved() bool {
	for _, expr := range j.Expressions() {
		if !expr.Resolved() {
			return false
		}
	}
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (j *JsonTable) RowIter(ctx *sql.Context, row sql.Row) (sql.RowIter, error) {
	contextFormat := j.ContextFormat
	if contextFormat == nil {
		contextFormat = &tree.JsonFormat{}
	}
	document, err := evalJsonTableInput(ctx, j.Context, contextFormat, row)
	if err != nil {
		return nil, err
	} else if document == nil {
		return sql.RowsToRowIter(), nil
	}
	vars := make(exec.Vars, len(j.Passing))
	for _, passing := range j.Passing {
		if _, ok := vars[passing.Name]; ok {
			continue
		}
		value, err := evalJsonTableInput(ctx, passing.Value, passing.Format, row)
		if err != nil {
			return nil, err
		}
		vars[passing.Name] = nil
		if value != nil {
			vars[passing.Name] = *value
		}
	}
	next, err := j.pathRowIter(ctx, j.Root, *document, vars, row)
	if err != nil {
		return nil, err
	}
	return pgtypes.NewSetReturningFunctionRowIter(next), nil
}

// Schema implements the interface sql.TableFunction.
func (j *JsonTable) Schema(ctx *sql.Context) sql.Schema {
	schema := make(sql.Schema, len(j.Columns))
	for i, column := range j.Columns {
		schema[i] = &sql.Column{
			Name:     column.Name,
			Type:     column.Type,
			Nullable: true,
			Source:   JsonTableName,
		}
	}
	return schema
}

// String implements the interface sql.TableFunction.
func (j *JsonTable) String() string {
	columns := make([]string, len(j.Columns))
	for i, column := range j.Columns {
		columns[i] = column.Name
	}
	return "JSON_TABLE(" + j.Context.String() + ", " + j.Root.Path.String() + " COLUMNS " + strings.Join(columns, ", ") + ")"
}

// WithChildren implements the interface sql.TableFunction.
func (j *JsonTable) WithChildren(ctx *sql.Context, children ...sql.Node) (sql.Node, error) {
	if len(children) != 0 {
		return nil, sql.ErrInvalidChildrenNumber.New(j, len(children), 0)
	}
	return j, nil
}

// WithDatabase implements the interface sql.TableFunction.
func (j *JsonTable) WithDatabase(database sql.Database) (sql.Node, error) {
	table := *j
	table.database = database
	return &table, nil
}

// WithExpressions implements the interface sql.TableFunction.
func (j *JsonTable) WithExpressions(ctx *sql.Context, exprs ...sql.Expression) (sql.Node, error) {
	if len(exprs) != len(j.Expressions()) {
		return nil, sql.ErrInvalidChildrenNumber.New(j, len(exprs), len(j.Expressions()))
	}
	table := *j
	table.Context, exprs = exprs[0], exprs[1:]
	table.Passing = make([]JsonTablePassing, len(j.Passing))
	for i, passing := range j.Passing {
		table.Passing[i] = passing
		table.Passing[i].Value, exprs = exprs[0], exprs[1:]
	}
	table.Columns = make([]JsonTableColumn, len(j.Columns))
	for i, column := range j.Columns {
		for _, behavior := range []**JsonTableBehavior{&column.OnEmpty, &column.OnError} {
			if *behavior != nil && (*behavior).Type == tree.JsonBehaviorDefault {
				*behavior = &JsonTableBehavior{Type: tree.JsonBehaviorDefault, Default: exprs[0]}
				exprs = exprs[1:]
			}
		}
		table.Columns[i] = column
	}
	return &table, nil
}

// pathRowIter returns a function that produces the rows of `tablePath` for `item` one at a time, leaving the columns of
// other paths NULL, and that returns io.EOF once the rows run out.
func (j *JsonTable) pathRowIter(ctx *sql.Context, tablePath *JsonTablePath, item any, vars exec.Vars, row sql.Row) (func(ctx *sql.Context) (sql.Row, error), error) {
	items, err := tablePath.Path.Query(ctx, item, exec.WithVars(vars))
	if err != nil && (j.ErrorOnError || !errors.Is(err, exec.ErrVerbose)) {
		return nil, jsonPathError(err)
	}
	itemIdx := 0
	nestedIdx := 0
	hasNestedRows := false
	var outputRow sql.Row
	var nestedNext func(ctx *sql.Context) (sql.Row, error)
	return func(ctx *sql.Context) (sql.Row, error) {
		for {
			if nestedNext != nil {
				nestedRow, err := nestedNext(ctx)
				if err == io.EOF {
					nestedNext = nil
					continue
				} else if err != nil {
					return nil, err
				}
				hasNestedRows = true
				for _, columnIdx := range tablePath.Columns {
					nestedRow[columnIdx] = outputRow[columnIdx]
				}
				return nestedRow, nil
			}
			if outputRow != nil {
				if nestedIdx < len(tablePath.Nested) {
					var err error
					if nestedNext, err = j.pathRowIter(ctx, tablePath.Nested[nestedIdx], items[itemIdx-1], vars, row); err != nil {
						return nil, err
					}
					nestedIdx++
					continue
				}
				parentRow := outputRow
				outputRow = nil
				if !hasNestedRows {
					return parentRow, nil
				}
				continue
			}
			if itemIdx >= len(items) {
				return nil, io.EOF
			}
			outputRow = make(sql.Row, len(j.Columns))
			for _, columnIdx := range tablePath.Columns {
				column := j.Columns[columnIdx]
				if column.Kind == tree.JsonTableColumnForOrdinality {
					outputRow[columnIdx] = int32(itemIdx + 1)
				} else {
					var err error
					if outputRow[columnIdx], err = j.columnValue(ctx, column, items[itemIdx], vars, row); err != nil {
						return nil, err
					}
				}
			}
			itemIdx++
			nestedIdx = 0
			hasNestedRows = false
		}
	}, nil
}

// columnValue returns the value of `column` for the row item `item`, applying its ON EMPTY and ON ERROR behaviors.
func (j *JsonTable) columnValue(ctx *sql.Context, column JsonTableColumn, item any, vars exec.Vars, row sql.Row) (any, error) {
	if column.Kind == tree.JsonTableColumnExists {
		exists, err := column.Path.Exists(ctx, item, exec.WithVars(vars))
		if err != nil {
			if !errors.Is(err, exec.ErrVerbose) && !errors.Is(err, exec.NULL) {
				return nil, jsonPathError(err)
			}
			return j.behaviorValue(ctx, column, column.OnError, jsonPathError(err), row)
		}
		return jsonTableExistsValue(ctx, column.Type, exists)
	}
	items, err := column.Path.Query(ctx, item, exec.WithVars(vars))
	if err != nil {
		if !errors.Is(err, exec.ErrVerbose) {
			return nil, jsonPathError(err)
		}
		return j.behaviorValue(ctx, column, column.OnError, jsonPathError(err), row)
	}
	if len(items) == 0 {
		return j.behaviorValue(ctx, column, column.OnEmpty, pgerror.Newf(pgcode.NoSQLJSONItem, `no SQL/JSON item found for specified path of column "%s"`, column.Name), row)
	}
	value, err := jsonTableItemsValue(ctx, column, items)
	if err != nil {
		return j.behaviorValue(ctx, column, column.OnError, err, row)
	}
	return value, nil
}

// behaviorValue returns the value that `behavior` produces for `column`, returning `err` for the ERROR behavior.
func (j *JsonTable) behaviorValue(ctx *sql.Context, column JsonTableColumn, behavior *JsonTableBehavior, err error, row sql.Row) (any, error) {
	switch behavior.Type {
	case tree.JsonBehaviorError:
		return nil, err
	case tree.JsonBehaviorTrue, tree.JsonBehaviorFalse:
		return jsonTableExistsValue(ctx, column.Type, behavior.Type == tree.JsonBehaviorTrue)
	case tree.JsonBehaviorEmptyArray:
		return jsonToColumnValue(ctx, column, []any{})
	case tree.JsonBehaviorEmptyObject:
		return jsonToColumnValue(ctx, column, map[string]any{})
	case tree.JsonBehaviorDefault:
		return behavior.Default.Eval(ctx, row)
	default:
		return nil, nil
	}
}

// newJsonTableColumn returns `column` with its type resolved and its behaviors validated and defaulted.
func newJsonTableColumn(ctx *sql.Context, column JsonTableColumn) (JsonTableColumn, error) {
	if !column.Type.IsResolvedType() {
		typeColl, err := core.GetTypesCollectionFromContext(ctx, "")
		if err != nil {
			return JsonTableColumn{}, err
		}
		if column.Type, err = typeColl.ResolveTypeWithTypmod(ctx, column.Type.ID, column.Type.UnresolvedTypmods); err != nil {
			return JsonTableColumn{}, err
		}
	}
	baseType := column.Type
	if baseType.TypType == pgtypes.TypeType_Domain {
		baseType = baseType.DomainUnderlyingBaseType()
	}
	switch column.Kind {
	case tree.JsonTableColumnForOrdinality:
		return column, nil
	case tree.JsonTableColumnExists:
		if column.OnError == nil {
			column.OnError = &JsonTableBehavior{Type: tree.JsonBehaviorFalse}
		}
		if column.OnError.Type == tree.JsonBehaviorTrue || column.OnError.Type == tree.JsonBehaviorFalse {
			if _, err := jsonTableExistsValue(ctx, column.Type, column.OnError.Type == tree.JsonBehaviorTrue); err != nil {
				return JsonTableColumn{}, pgerror.Newf(pgcode.DatatypeMismatch, "could not coerce ON ERROR expression (%s) to the RETURNING type", column.OnError.Type)
			}
		}
		return column, nil
	}
	isJson := baseType.ID == pgtypes.Json.ID || baseType.ID == pgtypes.JsonB.ID
	if column.Format != nil {
		if !isJson && !baseType.IsStringType() && baseType.ID != pgtypes.Bytea.ID {
			return JsonTableColumn{}, pgerror.New(pgcode.FeatureNotSupported, "cannot use JSON format with non-string output types")
		} else if column.Format.Encoding != "" && baseType.ID != pgtypes.Bytea.ID {
			return JsonTableColumn{}, pgerror.New(pgcode.FeatureNotSupported, "cannot set JSON encoding for non-bytea output types")
		}
	}
	column.formatted = column.Format != nil || column.Wrapper != tree.JsonWrapperUnspecified ||
		column.Quotes != tree.JsonQuotesUnspecified || isJson || baseType.IsArrayType() || baseType.IsCompositeType()
	for _, behavior := range []struct {
		behavior **JsonTableBehavior
		clause   string
	}{{&column.OnEmpty, "EMPTY"}, {&column.OnError, "ERROR"}} {
		if *behavior.behavior == nil {
			*behavior.behavior = &JsonTableBehavior{Type: tree.JsonBehaviorNull}
			continue
		}
		switch (*behavior.behavior).Type {
		case tree.JsonBehaviorError, tree.JsonBehaviorNull:
		case tree.JsonBehaviorEmptyArray, tree.JsonBehaviorEmptyObject:
			if !column.formatted {
				return JsonTableColumn{}, pgerror.Newf(pgcode.Syntax, `invalid ON %s behavior for column "%s"`, behavior.clause, column.Name)
			}
		case tree.JsonBehaviorDefault:
			defaultExpr := (*behavior.behavior).Default
			if transform.InspectExpr(ctx, defaultExpr, func(ctx *sql.Context, expr sql.Expression) bool {
				switch expr.(type) {
				case *expression.GetField, *plan.Subquery:
					return true
				}
				return false
			}) {
				return JsonTableColumn{}, pgerror.New(pgcode.DatatypeMismatch, "can only specify a constant, non-aggregate function, or operator expression for DEFAULT")
			}
			defaultType, ok := defaultExpr.Type(ctx).(*pgtypes.DoltgresType)
			if !ok {
				defaultType = pgtypes.FromGmsType(defaultExpr.Type(ctx))
			}
			*behavior.behavior = &JsonTableBehavior{Type: tree.JsonBehaviorDefault, Default: pgexprs.NewAssignmentCast(defaultExpr, defaultType, column.Type)}
		default:
			return JsonTableColumn{}, pgerror.Newf(pgcode.Syntax, `invalid ON %s behavior for column "%s"`, behavior.clause, column.Name)
		}
	}
	return column, nil
}

// newJsonTablePassing returns `passing` with its value converted to jsonb when SQL/JSON has no equivalent type.
func newJsonTablePassing(ctx *sql.Context, passing JsonTablePassing) (JsonTablePassing, error) {
	typ, err := jsonTableInputType(ctx, passing.Value, passing.Format)
	if err != nil || passing.Format != nil {
		return passing, err
	}
	switch typ.ID {
	case pgtypes.Bool.ID, pgtypes.Int16.ID, pgtypes.Int32.ID, pgtypes.Int64.ID, pgtypes.Float32.ID, pgtypes.Float64.ID,
		pgtypes.Numeric.ID, pgtypes.Text.ID, pgtypes.VarChar.ID, pgtypes.Unknown.ID, pgtypes.Date.ID, pgtypes.Time.ID,
		pgtypes.TimeTZ.ID, pgtypes.Timestamp.ID, pgtypes.TimestampTZ.ID, pgtypes.Json.ID, pgtypes.JsonB.ID:
		return passing, nil
	}
	if typ.IsStringType() {
		return JsonTablePassing{}, pgerror.Newf(pgcode.InvalidParameterValue, "could not convert value of type %s to jsonpath", typ.String())
	}
	toJsonb, ok, err := framework.GetFunction(ctx, "to_jsonb", passing.Value)
	if err != nil {
		return JsonTablePassing{}, err
	} else if !ok {
		return JsonTablePassing{}, errors.Errorf("function to_jsonb does not exist")
	}
	passing.Value = toJsonb
	return passing, nil
}

// jsonTableInputType returns the type of the context item or PASSING value `expr`, validating its ENCODING.
func jsonTableInputType(ctx *sql.Context, expr sql.Expression, format *tree.JsonFormat) (*pgtypes.DoltgresType, error) {
	typ, ok := expr.Type(ctx).(*pgtypes.DoltgresType)
	if !ok {
		typ = pgtypes.FromGmsType(expr.Type(ctx))
	}
	if format != nil && format.Encoding != "" && typ.ID != pgtypes.Bytea.ID {
		return nil, pgerror.New(pgcode.DatatypeMismatch, "JSON ENCODING clause is only allowed for bytea input type")
	}
	return typ, nil
}

// evalJsonTableInput evaluates the context item or PASSING value `expr` as a SQL/JSON item, returning nil for NULL.
func evalJsonTableInput(ctx *sql.Context, expr sql.Expression, format *tree.JsonFormat, row sql.Row) (*any, error) {
	value, err := expr.Eval(ctx, row)
	if err != nil || value == nil {
		return nil, err
	}
	typ := expr.Type(ctx).(*pgtypes.DoltgresType)
	var text string
	if typ.ID == pgtypes.Bytea.ID && format != nil {
		text = string(value.([]byte))
	} else if text, err = typ.IoOutput(ctx, value); err != nil {
		return nil, err
	}
	var item any
	switch {
	case format != nil, typ.ID == pgtypes.Json.ID, typ.ID == pgtypes.JsonB.ID:
		if item, err = parseJsonDocument(ctx, text); err != nil {
			return nil, err
		}
	case typ.ID == pgtypes.Bool.ID:
		item = value.(bool)
	case typ.ID == pgtypes.Int16.ID, typ.ID == pgtypes.Int32.ID, typ.ID == pgtypes.Int64.ID, typ.ID == pgtypes.Float32.ID,
		typ.ID == pgtypes.Float64.ID, typ.ID == pgtypes.Numeric.ID:
		item = json.Number(text)
	case typ.ID == pgtypes.Date.ID, typ.ID == pgtypes.Time.ID, typ.ID == pgtypes.TimeTZ.ID, typ.ID == pgtypes.Timestamp.ID,
		typ.ID == pgtypes.TimestampTZ.ID:
		dateTime, ok := types.ParseTime(ctx, text, -1)
		if !ok {
			return nil, pgerror.Newf(pgcode.InvalidParameterValue, "could not convert value of type %s to jsonpath", typ.String())
		}
		item = dateTime
	default:
		item = text
	}
	return &item, nil
}

// parseJsonDocument parses `text` as a jsonb document and returns it as a SQL/JSON item.
func parseJsonDocument(ctx *sql.Context, text string) (any, error) {
	document, err := pgtypes.JsonB.IoInput(ctx, text)
	if err != nil {
		return nil, err
	}
	if text, err = pgtypes.JsonB.IoOutput(ctx, document); err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(strings.NewReader(text))
	decoder.UseNumber()
	var item any
	err = decoder.Decode(&item)
	return item, err
}

// jsonTableItemsValue returns the value of `column` for the SQL/JSON items that its path matched.
func jsonTableItemsValue(ctx *sql.Context, column JsonTableColumn, items []any) (any, error) {
	if !column.formatted {
		if len(items) > 1 {
			return nil, pgerror.Newf(pgcode.MoreThanOneSQLJSONItem, `JSON path expression for column "%s" must return single scalar item`, column.Name)
		}
		switch item := items[0].(type) {
		case nil:
			return nil, nil
		case map[string]any, []any:
			return nil, pgerror.Newf(pgcode.SQLJSONScalarRequired, `JSON path expression for column "%s" must return single scalar item`, column.Name)
		case string:
			return column.Type.IoInput(ctx, item)
		case bool:
			if item {
				return column.Type.IoInput(ctx, "t")
			}
			return column.Type.IoInput(ctx, "f")
		case types.DateTime:
			dateTimeType := jsonDateTimeType(item)
			value, err := dateTimeType.IoInput(ctx, item.String())
			if err != nil {
				return nil, err
			}
			text, err := dateTimeType.IoOutput(ctx, value)
			if err != nil {
				return nil, err
			}
			return column.Type.IoInput(ctx, text)
		}
		text, err := jsonText(ctx, items[0])
		if err != nil {
			return nil, err
		}
		return column.Type.IoInput(ctx, text)
	}
	result := items[0]
	switch column.Wrapper {
	case tree.JsonWrapperUnconditional:
		result = items
	case tree.JsonWrapperConditional:
		if len(items) > 1 {
			result = items
		}
	default:
		if len(items) > 1 {
			return nil, pgerror.Newf(pgcode.MoreThanOneSQLJSONItem, `JSON path expression for column "%s" must return single item when no wrapper is requested`, column.Name)
		}
	}
	return jsonToColumnValue(ctx, column, result)
}

// jsonToColumnValue converts the SQL/JSON item `item` to the type of the formatted `column`.
func jsonToColumnValue(ctx *sql.Context, column JsonTableColumn, item any) (any, error) {
	if str, ok := item.(string); ok && column.Quotes == tree.JsonQuotesOmit {
		return column.Type.IoInput(ctx, str)
	}
	baseType := column.Type
	if baseType.TypType == pgtypes.TypeType_Domain {
		baseType = baseType.DomainUnderlyingBaseType()
	}
	if baseType.IsArrayType() || baseType.IsCompositeType() {
		if item == nil {
			return nil, nil
		}
		literal, err := jsonToLiteral(ctx, baseType, item)
		if err != nil {
			return nil, err
		}
		return column.Type.IoInput(ctx, literal)
	}
	text, err := jsonText(ctx, item)
	if err != nil {
		return nil, err
	}
	return column.Type.IoInput(ctx, text)
}

// jsonToLiteral converts the SQL/JSON item `item` into an input literal of the array or composite type `typ`.
func jsonToLiteral(ctx *sql.Context, typ *pgtypes.DoltgresType, item any) (string, error) {
	sb := strings.Builder{}
	if typ.IsArrayType() {
		array, ok := item.([]any)
		if !ok {
			return "", pgerror.New(pgcode.InvalidTextRepresentation, "expected JSON array")
		}
		isMultidimensional := false
		if len(array) > 0 {
			_, isMultidimensional = array[0].([]any)
		}
		sb.WriteByte('{')
		for i, element := range array {
			if i > 0 {
				sb.WriteByte(',')
			}
			if isMultidimensional {
				if _, ok := element.([]any); !ok {
					return "", pgerror.New(pgcode.InvalidTextRepresentation, "expected JSON array")
				}
				literal, err := jsonToLiteral(ctx, typ, element)
				if err != nil {
					return "", err
				}
				sb.WriteString(literal)
			} else if element == nil {
				sb.WriteString("NULL")
			} else if err := writeQuotedJsonField(ctx, &sb, typ.ArrayBaseType(), element); err != nil {
				return "", err
			}
		}
		sb.WriteByte('}')
		return sb.String(), nil
	}
	object, ok := item.(map[string]any)
	if _, isArray := item.([]any); isArray {
		return "", pgerror.New(pgcode.InvalidParameterValue, "cannot call populate_composite on an array")
	} else if !ok {
		return "", pgerror.New(pgcode.InvalidParameterValue, "cannot call populate_composite on a scalar")
	}
	sb.WriteByte('(')
	for i, attr := range typ.CompositeAttrs {
		if i > 0 {
			sb.WriteByte(',')
		}
		if field, ok := object[attr.Name]; ok && field != nil {
			if err := writeQuotedJsonField(ctx, &sb, attr.Type, field); err != nil {
				return "", err
			}
		}
	}
	sb.WriteByte(')')
	return sb.String(), nil
}

// writeQuotedJsonField writes the quoted input literal of `item` as an array element or composite field of type `typ`.
func writeQuotedJsonField(ctx *sql.Context, sb *strings.Builder, typ *pgtypes.DoltgresType, item any) error {
	var text string
	var err error
	if str, ok := item.(string); ok {
		text = str
	} else if typ.IsArrayType() || typ.IsCompositeType() {
		text, err = jsonToLiteral(ctx, typ, item)
	} else {
		text, err = jsonText(ctx, item)
	}
	if err != nil {
		return err
	}
	sb.WriteByte('"')
	sb.WriteString(strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(text))
	sb.WriteByte('"')
	return nil
}

// jsonText returns the jsonb text of the SQL/JSON item `item`.
func jsonText(ctx *sql.Context, item any) (string, error) {
	text, err := json.Marshal(item)
	if err != nil {
		return "", err
	}
	document, err := pgtypes.JsonB.IoInput(ctx, string(text))
	if err != nil {
		return "", err
	}
	return pgtypes.JsonB.IoOutput(ctx, document)
}

// jsonDateTimeType returns the type that matches the SQL/JSON datetime item `item`.
func jsonDateTimeType(item types.DateTime) *pgtypes.DoltgresType {
	switch item.(type) {
	case *types.Date:
		return pgtypes.Date
	case *types.Time:
		return pgtypes.Time
	case *types.TimeTZ:
		return pgtypes.TimeTZ
	case *types.Timestamp:
		return pgtypes.Timestamp
	default:
		return pgtypes.TimestampTZ
	}
}

// jsonTableExistsValue converts the result of an EXISTS column to `typ`.
func jsonTableExistsValue(ctx *sql.Context, typ *pgtypes.DoltgresType, exists bool) (any, error) {
	baseType := typ
	if baseType.TypType == pgtypes.TypeType_Domain {
		baseType = baseType.DomainUnderlyingBaseType()
	}
	switch baseType.ID {
	case pgtypes.Bool.ID:
		return exists, nil
	case pgtypes.Int32.ID:
		if exists {
			return int32(1), nil
		}
		return int32(0), nil
	}
	return typ.IoInput(ctx, strconv.FormatBool(exists))
}

// jsonPathErrorCodes maps each message of a SQL/JSON path query error to the code Postgres returns for it, checked in
// order.
var jsonPathErrorCodes = []struct {
	message string
	code    pgcode.Code
}{
	{"JSON object does not contain key", pgcode.SQLJSONMemberNotFound},
	{"jsonpath member accessor", pgcode.SQLJSONMemberNotFound},
	{"can only be applied to an object", pgcode.SQLJSONObjectNotFound},
	{"can only be applied to an array", pgcode.SQLJSONArrayNotFound},
	{"can only be applied to a string or numeric value", pgcode.NonNumericSQLJSONItem},
	{"can only be applied to a string", pgcode.InvalidArgumentForSQLJSONDatetimeFunction},
	{"format is not recognized", pgcode.InvalidArgumentForSQLJSONDatetimeFunction},
	{"jsonpath item method", pgcode.NonNumericSQLJSONItem},
	{"jsonpath array subscript", pgcode.InvalidSQLJSONSubscript},
	{"division by zero", pgcode.DivisionByZero},
	{"operand of unary jsonpath operator", pgcode.SQLJSONNumberNotFound},
	{"operand of jsonpath operator", pgcode.SingletonSQLJSONItemRequired},
	{"single boolean result is expected", pgcode.SingletonSQLJSONItemRequired},
	{"could not find jsonpath variable", pgcode.UndefinedObject},
	{"without time zone usage", pgcode.FeatureNotSupported},
}

// jsonPathError removes the package prefix from an error returned by a SQL/JSON path query and attaches its code.
func jsonPathError(err error) error {
	message := strings.TrimPrefix(err.Error(), "exec: ")
	for _, errorCode := range jsonPathErrorCodes {
		if strings.Contains(message, errorCode.message) {
			return pgerror.New(errorCode.code, message)
		}
	}
	return errors.New(message)
}
