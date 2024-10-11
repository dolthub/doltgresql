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

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/postgres/parser/parser"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	ast "github.com/dolthub/doltgresql/server/ast"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// ReplaceDomainType replaces a CreateTable node containing a domain type with its
// underlying type defined as retrieved from storage.
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
			domainType.DataType = domain.DataType
			col.Type = domainType
		}
	}
	return createTable, transform.NewTree, nil
}

// InsertOnDomainType retrieves and assigns domain type's default value, nullable and check constraints
// to the destination table schema and InsertInto node's checks.
func InsertOnDomainType(ctx *sql.Context, a *analyzer.Analyzer, node sql.Node, scope *plan.Scope, selector analyzer.RuleSelector, qFlags *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	insertInto, ok := node.(*plan.InsertInto)
	if !ok {
		return node, transform.SameTree, nil
	}

	domains, err := core.GetDomainsCollectionFromContext(ctx)
	if err != nil {
		return nil, false, err
	}

	builder := planbuilder.New(ctx, a.Catalog, sql.GlobalParser)

	// get current checks to append the domain checks to.
	checks := insertInto.Checks()
	destSch := insertInto.Destination.Schema()
	for _, col := range destSch {
		if domainType, ok := col.Type.(pgtypes.DomainType); ok {
			schemaName, err := core.GetSchemaName(ctx, insertInto.Database(), domainType.SchemaName)
			if err != nil {
				return nil, false, err
			}
			domain := domains.GetDomain(schemaName, domainType.Name)
			// assign column nullable
			col.Nullable = !domain.NotNull
			// get domain default value and assign to the column default value
			defVal, err := getDefault(builder, domain.DefaultExpr, col.Source, col.Type, col.Nullable)
			if err != nil {
				return nil, transform.SameTree, err
			}
			col.Default = defVal
			// get domain checks
			colChecks, err := getCheckConstraints(builder, col.Name, col.Source, domain.Checks)
			if err != nil {
				return nil, transform.SameTree, err
			}
			checks = append(checks, colChecks...)
		}
	}

	return insertInto.WithChecks(checks), transform.NewTree, nil
}

// getCheckConstraints takes the check constraint definitions, parses, builds and returns sql.CheckConstraints.
func getCheckConstraints(builder *planbuilder.Builder, colName string, tblName string, checkDefs []*sql.CheckDefinition) (sql.CheckConstraints, error) {
	checks := make(sql.CheckConstraints, len(checkDefs))
	for i, check := range checkDefs {
		parsed, err := parseAndReplaceDomainCheckConstraint(fmt.Sprintf("select %s from %s", check.CheckExpression, tblName), colName, tblName)
		if err != nil {
			return nil, err
		}
		selectStmt, ok := parsed.(*vitess.Select)
		if !ok || len(selectStmt.SelectExprs) != 1 {
			err := sql.ErrInvalidCheckConstraint.New(check.CheckExpression)
			if err != nil {
				return nil, err
			}
		}
		expr := selectStmt.SelectExprs[0]
		ae, ok := expr.(*vitess.AliasedExpr)
		if !ok {
			err := sql.ErrInvalidCheckConstraint.New(check.CheckExpression)
			if err != nil {
				return nil, err
			}
		}

		checks[i] = &sql.CheckConstraint{
			Name:     check.Name,
			Expr:     builder.BuildScalarWithTable(ae.Expr, selectStmt.From[0]),
			Enforced: true,
		}
	}

	return checks, nil
}

// getDefault takes the default value definition, parses, builds and returns sql.CheckConstraints.
func getDefault(builder *planbuilder.Builder, defExpr, tblName string, typ sql.Type, nullable bool) (*sql.ColumnDefaultValue, error) {
	if defExpr == "" {
		return nil, nil
	}
	parsed, err := sql.GlobalParser.ParseSimple(fmt.Sprintf("select %s from %s", defExpr, tblName))
	if err != nil {
		return nil, err
	}
	selectStmt, ok := parsed.(*vitess.Select)
	if !ok || len(selectStmt.SelectExprs) != 1 {
		err := sql.ErrInvalidColumnDefaultValue.New(defExpr)
		if err != nil {
			return nil, err
		}
	}
	expr := selectStmt.SelectExprs[0]
	ae, ok := expr.(*vitess.AliasedExpr)
	if !ok {
		err := sql.ErrInvalidColumnDefaultValue.New(defExpr)
		if err != nil {
			return nil, err
		}
	}
	return builder.BuildColumnDefaultValueWithTable(ae.Expr, selectStmt.From[0], typ, nullable), nil
}

// parseAndReplaceDomainCheckConstraint parses check constraint and replaces the `VALUE` column
// reference with resolved column given table and column names.
func parseAndReplaceDomainCheckConstraint(q string, colName, tblName string) (vitess.Statement, error) {
	stmt, err := parser.ParseOne(q)
	if err != nil {
		return nil, err
	}

	exprs := stmt.AST.(*tree.Select).Select.(*tree.SelectClause).Exprs
	if len(exprs) != 1 {
		return nil, fmt.Errorf("expected single select exprtession from domain check constraint parsing")
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
	if err != nil {
		return nil, err
	}
	stmt.AST.(*tree.Select).Select.(*tree.SelectClause).Exprs[0].Expr = e
	return ast.Convert(stmt)
}
