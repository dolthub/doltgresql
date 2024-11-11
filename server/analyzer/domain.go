// Copyright 2024 Dolthub, Inc.
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
	"fmt"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/planbuilder"
	"github.com/dolthub/go-mysql-server/sql/transform"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/parser"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/server/ast"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// AddDomainConstraints adds domain type's default value and check constraints
// to the destination table schema and InsertNode/Update node's checks.
func AddDomainConstraints(ctx *sql.Context, a *analyzer.Analyzer, node sql.Node, scope *plan.Scope, selector analyzer.RuleSelector, qFlags *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	switch n := node.(type) {
	case *plan.InsertInto:
		return resolveDomainTypeAndLoadCheckConstraints(ctx, a, n, n.Schema())
	case *plan.Update:
		return resolveDomainTypeAndLoadCheckConstraints(ctx, a, n, n.Schema())
	default:
		return node, transform.SameTree, nil
	}
}

// resolveDomainTypeAndLoadCheckConstraints retrieves and assigns domain type's default value, nullable and check constraints
// to the destination table schema and InsertNode/Update node's checks.
func resolveDomainTypeAndLoadCheckConstraints(ctx *sql.Context, a *analyzer.Analyzer, c sql.CheckConstraintNode, schema sql.Schema) (sql.Node, transform.TreeIdentity, error) {
	// get current checks to append the domain checks to.
	checks := c.Checks()
	var same = transform.SameTree
	for _, col := range schema {
		if domainType, ok := col.Type.(pgtypes.DomainType); ok {
			// assign column nullable
			col.Nullable = !domainType.NotNull
			// get domain default value and assign to the column default value
			defVal, err := getDefault(ctx, a, domainType.DefaultExpr, col.Source, col.Type, col.Nullable)
			if err != nil {
				return nil, transform.SameTree, err
			}
			col.Default = defVal
			// get domain checks
			colChecks, err := getCheckConstraints(ctx, a, col.Name, col.Source, domainType.Checks)
			if err != nil {
				return nil, transform.SameTree, err
			}
			checks = append(checks, colChecks...)
			same = transform.NewTree
		}
	}
	if same == transform.SameTree {
		return c, same, nil
	}
	return c.WithChecks(checks), same, nil
}

// getDefault takes the default value definition, parses, builds and returns sql.ColumnDefaultValue.
func getDefault(ctx *sql.Context, a *analyzer.Analyzer, defExpr, tblName string, typ sql.Type, nullable bool) (*sql.ColumnDefaultValue, error) {
	if defExpr == "" {
		return nil, nil
	}
	parsed, err := sql.GlobalParser.ParseSimple(fmt.Sprintf("select %s from %s", defExpr, tblName))
	if err != nil {
		return nil, err
	}
	selectStmt, ok := parsed.(*vitess.Select)
	if !ok || len(selectStmt.SelectExprs) != 1 {
		return nil, sql.ErrInvalidColumnDefaultValue.New(defExpr)
	}
	expr := selectStmt.SelectExprs[0]
	ae, ok := expr.(*vitess.AliasedExpr)
	if !ok {
		return nil, sql.ErrInvalidColumnDefaultValue.New(defExpr)
	}
	builder := planbuilder.New(ctx, a.Catalog, nil, sql.GlobalParser)
	return builder.BuildColumnDefaultValueWithTable(ae.Expr, selectStmt.From[0], typ, nullable), nil
}

// getCheckConstraints takes the check constraint definitions, parses, builds and returns sql.CheckConstraints.
func getCheckConstraints(ctx *sql.Context, a *analyzer.Analyzer, colName string, tblName string, checkDefs []*sql.CheckDefinition) (sql.CheckConstraints, error) {
	checks := make(sql.CheckConstraints, len(checkDefs))
	for i, check := range checkDefs {
		parsed, err := parseAndReplaceDomainCheckConstraint(check.CheckExpression, colName, tblName)
		if err != nil {
			return nil, err
		}
		selectStmt, ok := parsed.(*vitess.Select)
		if !ok || len(selectStmt.SelectExprs) != 1 {
			return nil, sql.ErrInvalidCheckConstraint.New(check.CheckExpression)
		}
		expr := selectStmt.SelectExprs[0]
		ae, ok := expr.(*vitess.AliasedExpr)
		if !ok {
			return nil, sql.ErrInvalidCheckConstraint.New(check.CheckExpression)
		}

		builder := planbuilder.New(ctx, a.Catalog, nil, sql.GlobalParser)
		checks[i] = &sql.CheckConstraint{
			Name:     check.Name,
			Expr:     builder.BuildScalarWithTable(ae.Expr, selectStmt.From[0]),
			Enforced: true,
		}
	}

	return checks, nil
}

// parseAndReplaceDomainCheckConstraint parses check constraint and replaces the `VALUE` column
// reference with resolved column given table and column names.
func parseAndReplaceDomainCheckConstraint(checkExpr string, colName, tblName string) (vitess.Statement, error) {
	stmt, err := parser.ParseOne(fmt.Sprintf("select %s from %s", checkExpr, tblName))
	if err != nil {
		return nil, err
	}

	selectStmt, ok := stmt.AST.(*tree.Select)
	if !ok {
		return nil, sql.ErrInvalidCheckConstraint.New(checkExpr)
	}
	selectClause, ok := selectStmt.Select.(*tree.SelectClause)
	if !ok {
		return nil, sql.ErrInvalidCheckConstraint.New(checkExpr)
	}
	exprs := selectClause.Exprs
	if len(exprs) != 1 {
		return nil, sql.ErrInvalidCheckConstraint.New(checkExpr)
	}

	updatedCheckExpr, err := tree.SimpleVisit(exprs[0].Expr, func(visitingExpr tree.Expr) (recurse bool, newExpr tree.Expr, err error) {
		switch v := visitingExpr.(type) {
		case *tree.UnresolvedName:
			if strings.ToLower(v.String()) != "value" {
				return false, nil, fmt.Errorf(`column "%s" does not exist`, v.String())
			}
			return false, &tree.ColumnItem{
				ColumnName: tree.Name(colName),
				TableName:  &tree.UnresolvedObjectName{NumParts: 1, Parts: [3]string{tblName}},
			}, nil
		}
		return true, visitingExpr, nil
	})
	if err != nil {
		return nil, err
	}

	exprs[0].Expr = updatedCheckExpr
	return ast.Convert(stmt)
}
