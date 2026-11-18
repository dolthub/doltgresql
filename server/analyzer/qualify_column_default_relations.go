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
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"

	"github.com/dolthub/doltgresql/core/id"
	pgexprs "github.com/dolthub/doltgresql/server/expression"
	"github.com/dolthub/doltgresql/server/functions"
	pgnodes "github.com/dolthub/doltgresql/server/node"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// sequenceFunctionNames are the functions whose first argument names a sequence. Postgres declares that argument as a
// regclass.
var sequenceFunctionNames = map[string]struct{}{
	"currval": {},
	"nextval": {},
	"setval":  {},
}

// QualifyColumnDefaultRelations qualifies the relation names in the column defaults and generated columns of a table
// that is being created or altered. Postgres resolves a constant regclass, such as the sequence given to nextval, when
// the default is created, so the default keeps naming the same relation under any later search path. Column defaults
// are persisted as strings, so the resolved name is written with its schema.
func QualifyColumnDefaultRelations(ctx *sql.Context, _ *analyzer.Analyzer, n sql.Node, _ *plan.Scope, _ analyzer.RuleSelector, _ *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	return transform.Node(ctx, n, func(ctx *sql.Context, n sql.Node) (sql.Node, transform.TreeIdentity, error) {
		switch n.(type) {
		case *pgnodes.CreateTable, *plan.CreateTable, *plan.AddColumn, *plan.ModifyColumn, *plan.AlterDefaultSet:
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
			newExpr, same, err := transform.Expr(ctx, e, qualifyRelationArgument)
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

// qualifyRelationArgument qualifies the relation named by a string literal that is cast to regclass, or that is the
// first argument of a sequence function.
func qualifyRelationArgument(ctx *sql.Context, e sql.Expression) (sql.Expression, transform.TreeIdentity, error) {
	switch e := e.(type) {
	case *pgexprs.ExplicitCast:
		castType, ok := e.Type(ctx).(*pgtypes.DoltgresType)
		if !ok || castType.ID != pgtypes.Regclass.ID {
			return e, transform.SameTree, nil
		}
	case sql.FunctionExpression:
		if _, ok := sequenceFunctionNames[strings.ToLower(e.FunctionName())]; !ok {
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
	qualifiedName, err := qualifiedRelationName(ctx, name)
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
// its table.
func qualifiedRelationName(ctx *sql.Context, name string) (string, error) {
	schema, _, err := functions.ParseRelationName(ctx, name)
	if err != nil || schema != "" {
		return name, err
	}

	resolved, err := pgtypes.Regclass.IoInput(ctx, name)
	if err != nil {
		return "", err
	}
	relationID, ok := resolved.(id.Id)
	if !ok {
		return "", errors.Errorf(`unexpected regclass value for "%s"`, name)
	}

	var relation string
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

// plainIdentifier matches the identifiers that do not need to be quoted.
var plainIdentifier = regexp.MustCompile(`^[a-z_][a-z0-9_$]*$`)

// quoteIdentifierIfNeeded quotes the given identifier when it would not otherwise be read back unchanged.
func quoteIdentifierIfNeeded(identifier string) string {
	if plainIdentifier.MatchString(identifier) {
		return identifier
	}
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
}
