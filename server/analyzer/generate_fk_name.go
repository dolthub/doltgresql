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
	"fmt"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"

	"github.com/dolthub/doltgresql/server/functions"
)

// generateForeignKeyName populates a generated foreign key name, in the Postgres default foreign key name format,
// when a foreign key is created without an explicit name specified.
func generateForeignKeyName(ctx *sql.Context, _ *analyzer.Analyzer, n sql.Node, _ *plan.Scope, _ analyzer.RuleSelector, _ *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	return transform.Node(n, func(n sql.Node) (sql.Node, transform.TreeIdentity, error) {
		switch n := n.(type) {
		case *plan.CreateTable:
			copiedForeignKeys := make([]*sql.ForeignKeyConstraint, len(n.ForeignKeys()))
			for i := range n.ForeignKeys() {
				fk := *n.ForeignKeys()[i]
				copiedForeignKeys[i] = &fk
			}

			changedForeignKey := false
			for _, fk := range copiedForeignKeys {
				if fk.Name == "" {
					generatedName, err := generateFkName(ctx, n.Name(), fk)
					if err != nil {
						return nil, transform.SameTree, err
					}
					changedForeignKey = true
					fk.Name = generatedName
				}
			}
			if changedForeignKey {
				newCreateTable := plan.NewCreateTable(n.Db, n.Name(), n.IfNotExists(), n.Temporary(), &plan.TableSpec{
					Schema:    n.PkSchema(),
					FkDefs:    copiedForeignKeys,
					ChDefs:    n.Checks(),
					IdxDefs:   n.Indexes(),
					Collation: n.Collation,
					TableOpts: n.TableOpts,
				})
				return newCreateTable, transform.NewTree, nil
			} else {
				return n, transform.SameTree, nil
			}

		case *plan.CreateForeignKey:
			if n.FkDef.Name == "" {
				copiedFk := *n.FkDef
				generatedName, err := generateFkName(ctx, copiedFk.Table, &copiedFk)
				if err != nil {
					return nil, transform.SameTree, err
				}
				copiedFk.Name = generatedName
				return &plan.CreateForeignKey{
					DbProvider: n.DbProvider,
					FkDef:      &copiedFk,
				}, transform.NewTree, nil
			} else {
				return n, transform.SameTree, nil
			}

		default:
			return n, transform.SameTree, nil
		}
	})
}

// generateFkName creates a default foreign key name, according to Postgres naming rules
// (i.e. "<tablename>_<col1name>_<col2name>_fkey"). If an existing foreign key is found with the default, generated
// name, the generated name will be suffixed with a number to ensure uniqueness.
func generateFkName(ctx *sql.Context, tableName string, newFk *sql.ForeignKeyConstraint) (string, error) {
	columnNames := strings.Join(newFk.Columns, "_")
	generatedBaseName := fmt.Sprintf("%s_%s_fkey", tableName, columnNames)

	for counter := 0; counter < 100; counter += 1 {
		generatedFkName := generatedBaseName
		if counter > 0 {
			generatedFkName = fmt.Sprintf("%s%d", generatedBaseName, counter)
		}

		duplicate := false
		err := functions.IterateCurrentDatabase(ctx, functions.Callbacks{
			ForeignKey: func(ctx *sql.Context, schema functions.ItemSchema, table functions.ItemTable, foreignKey functions.ItemForeignKey) (cont bool, err error) {
				if foreignKey.Item.Name == generatedFkName {
					duplicate = true
					return false, nil
				}
				return true, nil
			},
		})
		if err != nil {
			return "", err
		}

		if !duplicate {
			return generatedFkName, nil
		}
	}

	return "", fmt.Errorf("unable to create unique foreign key %s: "+
		"a foreign key constraint already exists with this name", generatedBaseName)
}
