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
	"strconv"
	"strings"

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/expranalysis"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/overrides"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/transform"

	pgexprs "github.com/dolthub/doltgresql/server/expression"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// DisplayCheckExpression returns the text shown for the persisted expression of a check constraint on the named table,
// with its function calls shown as DisplayExpression shows them. The persisted text is returned when it cannot be
// resolved.
func DisplayCheckExpression(ctx *sql.Context, tableName doltdb.TableName, checkExpression string) string {
	expr, err := expranalysis.ResolveExpression(ctx, tableName, checkExpression)
	if err != nil {
		return checkExpression
	}

	// Check expressions are persisted without table names, and with their column names quoted
	formatter := overrides.SchemaFormatterFromContext(ctx)
	expr, _, err = transform.Expr(ctx, expr, func(ctx *sql.Context, e sql.Expression) (sql.Expression, transform.TreeIdentity, error) {
		if field, ok := e.(*expression.GetField); ok {
			return field.WithTable("").WithQuotedNames(formatter, true), transform.NewTree, nil
		}
		return e, transform.SameTree, nil
	})
	if err != nil {
		return checkExpression
	}

	displayed, err := DisplayExpression(ctx, expr)
	if err != nil {
		return checkExpression
	}
	return displayed.String()
}

// PostgresDefaultString returns the text that Postgres shows for a default built from constants and casts, where a string
// constant takes on the type of `columnType`. Any other expression returns its own text.
func PostgresDefaultString(ctx *sql.Context, expr sql.Expression, columnType *pgtypes.DoltgresType) (string, error) {
	switch expr := expr.(type) {
	case *expression.Literal:
		literalType, ok := expr.Type(ctx).(*pgtypes.DoltgresType)
		if !ok {
			break
		}
		if columnType != nil && literalType.ID == pgtypes.Unknown.ID {
			return PostgresDefaultString(ctx, pgexprs.NewExplicitCast(expr, columnType.WithAttTypMod(-1)), nil)
		}
		return postgresConstantString(ctx, expr.Value(), literalType, true)
	case *pgexprs.ExplicitCast:
		castType := expr.Type(ctx).(*pgtypes.DoltgresType)
		childType, ok := expr.Child().Type(ctx).(*pgtypes.DoltgresType)
		if !ok {
			break
		}
		if childType.Equals(castType) {
			return PostgresDefaultString(ctx, expr.Child(), nil)
		}
		literal, isLiteral := expr.Child().(*expression.Literal)
		var childText string
		var err error
		if isLiteral && childType.ID == pgtypes.Unknown.ID {
			var val any
			if val, err = expr.Eval(ctx, nil); err != nil {
				return "", err
			}
			if castType.GetAttTypMod() < 0 {
				return postgresConstantString(ctx, val, castType, true)
			}
			childText, err = postgresConstantString(ctx, val, castType, false)
		} else if isLiteral && childType.ID == castType.ID {
			childText, err = postgresConstantString(ctx, literal.Value(), childType, false)
		} else {
			childText, err = PostgresDefaultString(ctx, expr.Child(), nil)
			childText = "(" + childText + ")"
		}
		if err != nil {
			return "", err
		}
		typeName, err := FormatType(ctx, castType.ID.AsId(), castType.GetAttTypMod())
		if err != nil {
			return "", err
		}
		return childText + "::" + typeName, nil
	case *pgexprs.BinaryOperator:
		// Postgres encloses an operator in parentheses, while nested operators are already enclosed
		return "(" + expr.String() + ")", nil
	}
	return expr.String(), nil
}

// DisplayExpression returns the given expression with its function calls shown as Postgres shows them. Column defaults,
// generated columns, and check constraints are persisted with the schema of every function that they call, while
// Postgres shows a function's schema only when the function is not visible on the search path, and shows the type of the
// constants passed to it. The returned expression is only for display.
func DisplayExpression(ctx *sql.Context, expr sql.Expression) (sql.Expression, error) {
	newExpr, _, err := transform.Expr(ctx, expr, func(ctx *sql.Context, e sql.Expression) (sql.Expression, transform.TreeIdentity, error) {
		var text string
		var err error
		switch function := e.(type) {
		case *framework.CompiledFunction:
			if function.IsOperator {
				return e, transform.SameTree, nil
			}
			schema := ""
			if !function.IsVisible(ctx) {
				schema = function.SchemaName()
			}
			text, err = postgresFunctionString(ctx, schema, function.Name, function.Arguments, function.ResolvedParameterTypes())
		case framework.QuickFunction:
			// Quick functions are built-in functions, which only need their schema when the search path hides them
			resolvedTypes := function.ResolvedTypes()
			paramTypes := resolvedTypes[:len(resolvedTypes)-1]
			schema := ""
			if !framework.BuiltInIsVisible(ctx, function.FunctionName(), paramTypes) {
				schema = "pg_catalog"
			}
			text, err = postgresFunctionString(ctx, schema, function.FunctionName(), function.Children(), paramTypes)
		default:
			return e, transform.SameTree, nil
		}
		if err != nil {
			return nil, transform.SameTree, err
		}
		return displayedFunction{Expression: e, text: text}, transform.NewTree, nil
	})
	return newExpr, err
}

// postgresFunctionString returns the text that Postgres shows for a call of the named function, qualified by the given
// schema unless it is empty, whose arguments are passed to parameters of the given types.
func postgresFunctionString(ctx *sql.Context, schema string, name string, args []sql.Expression, paramTypes []*pgtypes.DoltgresType) (string, error) {
	sb := strings.Builder{}
	if schema != "" {
		sb.WriteString(schema + ".")
	}
	sb.WriteString(name + "(")
	for i, arg := range args {
		if i > 0 {
			sb.WriteString(", ")
		}
		var paramType *pgtypes.DoltgresType
		if i < len(paramTypes) {
			paramType = paramTypes[i]
		}
		argText, err := postgresArgumentString(ctx, arg, paramType, i == 0 && IsSequenceFunction(name))
		if err != nil {
			return "", err
		}
		sb.WriteString(argText)
	}
	sb.WriteString(")")
	return sb.String(), nil
}

// postgresArgumentString returns the text that Postgres shows for a function argument, which is passed to a parameter
// of the given type. A constant shows the type that it is passed as, and a constant that names a relation is shown as a
// regclass when `isRelation` is true, as Postgres declares the parameters of sequence functions.
func postgresArgumentString(ctx *sql.Context, arg sql.Expression, paramType *pgtypes.DoltgresType, isRelation bool) (string, error) {
	if alias, ok := arg.(*expression.Alias); ok {
		arg = alias.Child
	}
	switch arg := arg.(type) {
	case *expression.Literal:
		if name, ok := arg.Value().(string); ok && isRelation {
			if relation, err := pgtypes.Regclass.IoInput(ctx, name); err == nil {
				return postgresConstantString(ctx, relation, pgtypes.Regclass, true)
			}
		}
		literalType, ok := arg.Type(ctx).(*pgtypes.DoltgresType)
		if ok && literalType.ID == pgtypes.Unknown.ID && paramType != nil {
			return PostgresDefaultString(ctx, pgexprs.NewExplicitCast(arg, paramType.WithAttTypMod(-1)), nil)
		}
		return PostgresDefaultString(ctx, arg, nil)
	case *pgexprs.ExplicitCast:
		return PostgresDefaultString(ctx, arg, nil)
	default:
		return arg.String(), nil
	}
}

// displayedFunction is a function call that is shown as the given text.
type displayedFunction struct {
	sql.Expression
	text string
}

var _ sql.Expression = displayedFunction{}

// String implements the interface sql.Expression.
func (d displayedFunction) String() string {
	return d.text
}

// postgresConstantString returns the text that Postgres shows for a constant of the given type. When `labeled` is true, the
// type name follows any constant that would otherwise read back as a different type.
func postgresConstantString(ctx *sql.Context, val any, typ *pgtypes.DoltgresType, labeled bool) (string, error) {
	var text string
	if val == nil {
		text = "NULL"
	} else if typ.ID == pgtypes.Bool.ID {
		return strconv.FormatBool(val.(bool)), nil
	} else {
		output, err := typ.IoOutput(ctx, val)
		if err != nil {
			return "", err
		}
		if (typ.ID == pgtypes.Int32.ID && !strings.HasPrefix(output, "-")) ||
			(typ.ID == pgtypes.Numeric.ID && output[0] >= '0' && output[0] <= '9' && strings.ContainsAny(output, ".eE")) {
			return output, nil
		}
		text = "'" + strings.ReplaceAll(output, "'", "''") + "'"
	}
	if !labeled || typ.ID == pgtypes.Unknown.ID {
		return text, nil
	}
	typeName, err := FormatType(ctx, typ.ID.AsId(), typ.GetAttTypMod())
	if err != nil {
		return "", err
	}
	return text + "::" + typeName, nil
}
