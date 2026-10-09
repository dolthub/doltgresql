// Copyright 2025 Dolthub, Inc.
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
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/planbuilder"
	"github.com/dolthub/go-mysql-server/sql/transform"

	pgnodes "github.com/dolthub/doltgresql/server/node"
)

// resolveDropConstraint resolves PostgreSQL constraint names to concrete DDL operations before GMS's resolver.
// Unique constraints use AlterIndex to retain the resolved database schema through execution.
func resolveDropConstraint(ctx *sql.Context, a *analyzer.Analyzer, n sql.Node, _ *plan.Scope, _ analyzer.RuleSelector, _ *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	return transform.Node(ctx, n, func(ctx *sql.Context, n sql.Node) (sql.Node, transform.TreeIdentity, error) {
		dropConstraint, ok := n.(*plan.DropConstraint)
		if !ok {
			return n, transform.SameTree, nil
		}

		rt, ok := dropConstraint.Child.(*plan.ResolvedTable)
		if !ok {
			return nil, transform.SameTree, analyzer.ErrInAnalysis.New(
				"Expected a TableNode for ALTER TABLE DROP CONSTRAINT statement")
		}

		// Expression indexes introduce virtual columns, whose table wrapper doesn't expose the underlying
		// constraint interfaces. DDL needs the underlying table, including for GMS's CHECK validation.
		resolvedTable := *rt
		resolvedTable.Table = sql.GetUnderlyingTable(rt.Table)
		rt = &resolvedTable
		table := rt.Table

		if fkt, ok := table.(sql.ForeignKeyTable); ok {
			foreignKeys, err := fkt.GetDeclaredForeignKeys(ctx)
			if err != nil {
				return nil, transform.SameTree, err
			}
			for _, fk := range foreignKeys {
				if strings.EqualFold(fk.Name, dropConstraint.Name) {
					return pgnodes.NewDropForeignKey(rt, fk.Name), transform.NewTree, nil
				}
			}
		}

		if ct, ok := table.(sql.CheckTable); ok {
			checks, err := ct.GetChecks(ctx)
			if err != nil {
				return nil, transform.SameTree, err
			}
			for _, check := range checks {
				if strings.EqualFold(check.Name, dropConstraint.Name) {
					return plan.NewAlterDropCheck(rt, check.Name), transform.NewTree, nil
				}
			}
		}

		if it, ok := table.(sql.IndexAddressable); ok {
			indexes, err := it.GetIndexes(ctx)
			if err != nil {
				return nil, transform.SameTree, err
			}
			for _, index := range indexes {
				if index.ID() == "PRIMARY" && dropConstraint.Name == rt.Name()+"_pkey" {
					// AlterPK exposes generated expressions to analysis. Resolve the persisted defaults
					// before introducing it, including hidden columns backing expression indexes.
					builder := planbuilder.NewBuilderForColumnDefaultResolution(ctx, a.Overrides)
					targetSchema := builder.ResolveSchemaDefaults(rt.Database().Name(), rt.Name(), rt.Schema(ctx))
					alterDropPk := plan.NewAlterDropPk(rt.Database(), rt)
					newNode, err := alterDropPk.WithTargetSchema(targetSchema)
					if err != nil {
						return n, transform.SameTree, err
					}
					return newNode, transform.NewTree, nil
				}
				if index.ID() != "PRIMARY" && index.IsUnique() && index.ID() == dropConstraint.Name {
					// DropIndex reacquires the database by name and loses the target schema. AlterIndex
					// retains the schema-qualified database and also handles expression-index cleanup.
					newNode, err := plan.NewAlterDropIndex(rt.Database(), rt, dropConstraint.IfExists, index.ID()).
						WithTargetSchema(rt.Schema(ctx))
					return newNode, transform.NewTree, err
				}
			}
		}

		// Apply IF EXISTS only after looking for every supported constraint type. The optional CHECK
		// drop provides the existing no-op execution behavior for a missing constraint.
		if dropConstraint.IfExists {
			newNode := plan.NewAlterDropCheck(rt, dropConstraint.Name)
			newNode.IfExists = true
			return newNode, transform.NewTree, nil
		}
		return nil, transform.SameTree, sql.ErrUnknownConstraint.New(dropConstraint.Name)
	})
}
