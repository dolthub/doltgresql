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
	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/postgres/parser/parser"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	ast2 "github.com/dolthub/doltgresql/server/ast"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/planbuilder"
	"github.com/dolthub/go-mysql-server/sql/transform"
	ast "github.com/dolthub/vitess/go/vt/sqlparser"
	"strings"
)

// ReplaceDomainType replaces a CreateTable node containing a domain type with a node
func ReplaceDomainType(ctx *sql.Context, a *analyzer.Analyzer, node sql.Node, scope *plan.Scope, selector analyzer.RuleSelector, qFlags *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	createTable, ok := node.(*plan.CreateTable)
	if !ok {
		return node, transform.SameTree, nil
	}

	for _, col := range createTable.PkSchema().Schema {
		if domainType, ok := col.Type.(pgtypes.DomainType); ok {
			schemaName, err := core.GetSchemaName(ctx, createTable.Db, domainType.SchemaName)
			if err != nil {
				return nil, false, err
			}

			domains, err := core.GetDomainsCollectionFromContext(ctx)
			if err != nil {
				return nil, false, err
			}
			domain := domains.GetDomain(schemaName, domainType.Name)
			// TODO : builder!
			domainType.DataType = domain.DataType
			col.Type = domainType
		}
	}
	return createTable, transform.NewTree, nil
}

// InsertOnDomainType replaces a InsertInto node containing a domain type with a node
func InsertOnDomainType(ctx *sql.Context, a *analyzer.Analyzer, node sql.Node, scope *plan.Scope, selector analyzer.RuleSelector, qFlags *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	insertInto, ok := node.(*plan.InsertInto)
	if !ok {
		return node, transform.SameTree, nil
	}

	// TODO:
	//checks := insertInto.Checks()

	destSch := insertInto.Destination.Schema() // table schema
	//srcSch := insertInto.Source.Schema()       // row values

	for _, col := range destSch {
		if domainType, ok := col.Type.(pgtypes.DomainType); ok {
			schemaName, err := core.GetSchemaName(ctx, insertInto.Database(), domainType.SchemaName)
			if err != nil {
				return nil, false, err
			}

			domains, err := core.GetDomainsCollectionFromContext(ctx)
			if err != nil {
				return nil, false, err
			}
			domain := domains.GetDomain(schemaName, domainType.Name)

			defVal, err := getDefault(ctx, a, domain.DefaultExpr)
			if err != nil {
				return nil, transform.SameTree, err
			}
			checks, err := getCheckConstraints(ctx, a, col.Name, col.Source, domain.Checks)
			if err != nil {
				return nil, transform.SameTree, err
			}
			domainType.ResolveType(defVal, domain.NotNull, checks)
		}
	}

	return insertInto, transform.SameTree, nil
}

func getCheckConstraints(ctx *sql.Context, a *analyzer.Analyzer, colName string, tblName string, checkDefs []*sql.CheckDefinition) (sql.CheckConstraints, error) {
	builder := planbuilder.New(ctx, a.Catalog, sql.GlobalParser)
	checks := make(sql.CheckConstraints, len(checkDefs))
	for i, check := range checkDefs {
		parsed, err := parseAndReplace(fmt.Sprintf("select %s", check.CheckExpression), colName, tblName)
		if err != nil {
			return nil, err
		}
		selectStmt, ok := parsed.(*ast.Select)
		if !ok || len(selectStmt.SelectExprs) != 1 {
			err := sql.ErrInvalidCheckConstraint.New(check.CheckExpression)
			if err != nil {
				return nil, err
			}
		}
		expr := selectStmt.SelectExprs[0]
		ae, ok := expr.(*ast.AliasedExpr)
		if !ok {
			err := sql.ErrInvalidCheckConstraint.New(check.CheckExpression)
			if err != nil {
				return nil, err
			}
		}

		checks[i] = &sql.CheckConstraint{
			Name:     check.Name,
			Expr:     builder.BuildScalar(ae.Expr),
			Enforced: true,
		}

	}

	return checks, nil
}

func getDefault(ctx *sql.Context, a *analyzer.Analyzer, defExpr string) (sql.Expression, error) {
	if defExpr == "" {
		return nil, nil
	}
	builder := planbuilder.New(ctx, a.Catalog, sql.GlobalParser)
	parsed, err := sql.GlobalParser.ParseSimple(fmt.Sprintf("select %s", defExpr))
	if err != nil {
		return nil, err
	}
	selectStmt, ok := parsed.(*ast.Select)
	if !ok || len(selectStmt.SelectExprs) != 1 {
		err := sql.ErrInvalidColumnDefaultValue.New(defExpr)
		if err != nil {
			return nil, err
		}
	}
	expr := selectStmt.SelectExprs[0]
	ae, ok := expr.(*ast.AliasedExpr)
	if !ok {
		err := sql.ErrInvalidColumnDefaultValue.New(defExpr)
		if err != nil {
			return nil, err
		}
	}

	return builder.BuildScalar(ae.Expr), nil
}

func parseAndReplace(q string, colName, tblName string) (ast.Statement, error) {
	stmt, err := parser.ParseOne(q)
	if err != nil {
		return nil, err
	}

	exprs := stmt.AST.(*tree.Select).Select.(*tree.SelectClause).Exprs
	if len(exprs) != 1 {

	}
	e, err := tree.SimpleVisit(exprs[0].Expr, func(visitingExpr tree.Expr) (recurse bool, newExpr tree.Expr, err error) {
		switch v := visitingExpr.(type) {
		case *tree.UnresolvedName:
			if strings.ToLower(v.String()) != "value" {
				return false, nil, fmt.Errorf(`column "%s" does not exist`, v.String())
			}
			return false, &tree.ColumnItem{ColumnName: tree.Name(colName), TableName: &tree.UnresolvedObjectName{NumParts: 1, Parts: [3]string{tblName}}}, nil
		case *tree.ColumnItem:
			if strings.ToLower(v.Column()) != "value" {
				return false, nil, fmt.Errorf(`column "%s" does not exist`, v.String())
			}
			return false, &tree.ColumnItem{ColumnName: tree.Name(colName), TableName: &tree.UnresolvedObjectName{NumParts: 1, Parts: [3]string{tblName}}}, nil
		}
		return true, visitingExpr, nil
	})
	stmt.AST.(*tree.Select).Select.(*tree.SelectClause).Exprs[0].Expr = e
	return ast2.Convert(stmt)
}
