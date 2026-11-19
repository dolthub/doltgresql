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

package analyzer

import (
	"regexp"
	"slices"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/ast"
	pgexprs "github.com/dolthub/doltgresql/server/expression"
	"github.com/dolthub/doltgresql/server/functions"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgnodes "github.com/dolthub/doltgresql/server/node"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// QualifyColumnDefaultRelations qualifies the relation names and built-in function calls in the column defaults,
// generated columns, and check constraints of a table that is being created or altered. Postgres resolves a constant
// regclass, such as the sequence given to nextval, and the function that each call names, when the expression is
// created, so the expression keeps naming the same objects under any later search path. These expressions are persisted
// as strings, so the resolved names are written with their schemas.
func QualifyColumnDefaultRelations(ctx *sql.Context, _ *analyzer.Analyzer, n sql.Node, _ *plan.Scope, _ analyzer.RuleSelector, _ *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	return transform.Node(ctx, n, func(ctx *sql.Context, n sql.Node) (sql.Node, transform.TreeIdentity, error) {
		// Postgres creates a table before it resolves the expressions in its columns, so those expressions may name the
		// table itself
		var newTable *doltdb.TableName
		switch n := n.(type) {
		case *pgnodes.CreateTable:
			var err error
			newTable, err = createdTableName(ctx, n.GmsCreateTable())
			if err != nil {
				return nil, transform.SameTree, err
			}
		case *plan.CreateTable:
			var err error
			newTable, err = createdTableName(ctx, n)
			if err != nil {
				return nil, transform.SameTree, err
			}
		case *plan.AddColumn, *plan.ModifyColumn, *plan.AlterDefaultSet, *plan.CreateCheck:
		default:
			return n, transform.SameTree, nil
		}

		expressioner, ok := n.(sql.Expressioner)
		if !ok {
			return n, transform.SameTree, nil
		}
		exprs := expressioner.Expressions()
		var newExprs []sql.Expression
		for i, e := range exprs {
			newExpr, same, err := transform.Expr(ctx, e, func(ctx *sql.Context, e sql.Expression) (sql.Expression, transform.TreeIdentity, error) {
				e, sameRelation, err := qualifyRelationArgument(ctx, e, newTable)
				if err != nil {
					return nil, transform.SameTree, err
				}
				e, sameFunction := qualifyBuiltInFunction(e)
				return e, sameRelation && sameFunction, nil
			})
			if err != nil {
				return nil, transform.SameTree, err
			}
			if same {
				continue
			}
			if newExprs == nil {
				newExprs = make([]sql.Expression, len(exprs))
				copy(newExprs, exprs)
			}
			newExprs[i] = newExpr
		}
		if newExprs == nil {
			return n, transform.SameTree, nil
		}

		newNode, err := expressioner.WithExpressions(ctx, newExprs...)
		if err != nil {
			return nil, transform.SameTree, err
		}
		return newNode, transform.NewTree, nil
	})
}

// qualifyBuiltInFunction qualifies a call to a built-in function by `pg_catalog`. Postgres stores the function that a
// call resolved to, so a later search path that puts a function of the same name ahead of `pg_catalog` does not change
// which function the expression calls. Other functions are already qualified by the schema they resolved to.
func qualifyBuiltInFunction(e sql.Expression) (sql.Expression, transform.TreeIdentity) {
	function, ok := e.(*framework.CompiledFunction)
	if !ok || function.IsOperator || function.SchemaName() != "pg_catalog" {
		return e, transform.SameTree
	}
	return function.WithQualifiedName(), transform.NewTree
}

// createdTableName returns the schema and name of the table that the given node creates.
func createdTableName(ctx *sql.Context, createTable *plan.CreateTable) (*doltdb.TableName, error) {
	schema, err := core.GetSchemaName(ctx, createTable.Db, "")
	if err != nil {
		return nil, err
	}
	return &doltdb.TableName{Schema: schema, Name: createTable.Name()}, nil
}

// qualifyRelationArgument qualifies the relation named by a string literal that is cast to regclass, or that is the
// first argument of a sequence function. |newTable| is the table being created, if any, which does not exist yet.
func qualifyRelationArgument(ctx *sql.Context, e sql.Expression, newTable *doltdb.TableName) (sql.Expression, transform.TreeIdentity, error) {
	switch e := e.(type) {
	case *pgexprs.ExplicitCast:
		castType, ok := e.Type(ctx).(*pgtypes.DoltgresType)
		if !ok || castType.ID != pgtypes.Regclass.ID {
			return e, transform.SameTree, nil
		}
	case sql.FunctionExpression:
		if !functions.IsSequenceFunction(e.FunctionName()) {
			return e, transform.SameTree, nil
		}
	default:
		return e, transform.SameTree, nil
	}

	// The relation is the only child of a cast, and the first argument of a sequence function
	children := e.Children()
	if len(children) == 0 {
		return e, transform.SameTree, nil
	}
	literal, ok := children[0].(*expression.Literal)
	if !ok {
		return e, transform.SameTree, nil
	}
	name, ok := literal.Value().(string)
	if !ok {
		return e, transform.SameTree, nil
	}
	qualifiedName, err := qualifiedRelationName(ctx, name, newTable)
	if err != nil || qualifiedName == name {
		return e, transform.SameTree, err
	}

	newChildren := make([]sql.Expression, len(children))
	copy(newChildren, children)
	newChildren[0] = expression.NewLiteral(qualifiedName, literal.Type(ctx))
	newExpr, err := e.WithChildren(ctx, newChildren...)
	if err != nil {
		return nil, transform.SameTree, err
	}
	return newExpr, transform.NewTree, nil
}

// qualifiedRelationName resolves the given relation name as regclass input does, and returns it qualified by the
// relation's schema. A name that is already qualified, or that is a number, which regclass input reads as an OID, is
// returned unchanged. A qualified name may not exist yet, such as the sequence that a SERIAL column creates along with
// its table. |newTable| is the table being created, if any, which is found ahead of any relation of the same name in a
// later schema on the search path.
func qualifiedRelationName(ctx *sql.Context, name string, newTable *doltdb.TableName) (string, error) {
	// The placeholder for an identity column's sequence is replaced once the sequence has a name
	if name == ast.DoltCreateTablePlaceholderSequenceName {
		return name, nil
	}

	schema, relation, err := functions.ParseRelationName(ctx, name)
	if err != nil || schema != "" {
		return name, err
	}

	resolved, err := pgtypes.Regclass.IoInput(ctx, name)
	if newTable != nil && relation == newTable.Name {
		newTableFirst, pathErr := newTableComesFirst(ctx, *newTable, resolved, err)
		if pathErr != nil {
			return "", pathErr
		}
		if newTableFirst {
			return quoteIdentifierIfNeeded(newTable.Schema) + "." + quoteIdentifierIfNeeded(newTable.Name), nil
		}
	}
	if err != nil {
		return "", err
	}
	relationID, ok := resolved.(id.Id)
	if !ok {
		return "", errors.Errorf(`unexpected regclass value for "%s"`, name)
	}

	switch relationID.Section() {
	case id.Section_Sequence:
		schema, relation = id.Sequence(relationID).SchemaName(), id.Sequence(relationID).SequenceName()
	case id.Section_Table:
		schema, relation = id.Table(relationID).SchemaName(), id.Table(relationID).TableName()
	case id.Section_View:
		schema, relation = id.View(relationID).SchemaName(), id.View(relationID).ViewName()
	case id.Section_Index:
		schema, relation = id.Index(relationID).SchemaName(), id.Index(relationID).IndexName()
	default:
		return name, nil
	}
	return quoteIdentifierIfNeeded(schema) + "." + quoteIdentifierIfNeeded(relation), nil
}

// newTableComesFirst returns whether the search path finds the table being created ahead of the relation that regclass
// input resolved, given as |resolved| and |resolveErr|. The table being created can only be found when its schema is on
// the search path.
func newTableComesFirst(ctx *sql.Context, newTable doltdb.TableName, resolved any, resolveErr error) (bool, error) {
	searchPath, err := core.SearchPath(ctx)
	if err != nil {
		return false, err
	}

	newTableIndex := slices.Index(searchPath, newTable.Schema)
	if newTableIndex < 0 {
		return false, nil
	}
	if resolveErr != nil {
		return true, nil
	}

	relationID, ok := resolved.(id.Id)
	if !ok {
		return false, nil
	}
	resolvedIndex := slices.Index(searchPath, relationID.Segment(0))
	return resolvedIndex < 0 || newTableIndex < resolvedIndex, nil
}

// plainIdentifier matches the identifiers that do not need to be quoted.
var plainIdentifier = regexp.MustCompile(`^[a-z_][a-z0-9_$]*$`)

// quoteIdentifierIfNeeded quotes the given identifier when it would not otherwise be read back unchanged.
func quoteIdentifierIfNeeded(identifier string) string {
	if plainIdentifier.MatchString(identifier) {
		return identifier
	}
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
}
