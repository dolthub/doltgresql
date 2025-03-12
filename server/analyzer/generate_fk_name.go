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
)

// generateForeignKeyName populates a generated foreign key name, in the Postgres default foreign key name format,
// when a foreign key is created without an explicit name specified.
func generateForeignKeyName(_ *sql.Context, _ *analyzer.Analyzer, n sql.Node, _ *plan.Scope, _ analyzer.RuleSelector, _ *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
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
					generatedName, err := generateFkName(n.Name(), fk, nil)
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
				generatedName, err := generateFkName(copiedFk.Table, &copiedFk, nil)
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

// generateFkName creates a default foreign key name, according to Postgres naming rules (i.e. "<tablename>_<col1name>_<col2name>_fkey").
// |existingFks| is used to check that the generated name doesn't conflict with an existing foreign key name. If a
// conflicting name is generated, this function returns an error.
func generateFkName(tableName string, newFk *sql.ForeignKeyConstraint, existingFks []sql.ForeignKeyConstraint) (string, error) {
	columnNames := strings.Join(newFk.Columns, "_")
	generatedFkName := fmt.Sprintf("%s_%s_fkey", tableName, columnNames)

	for _, existingFk := range existingFks {
		if existingFk.Name == generatedFkName {
			// TODO: Instead of returning an error, we should follow Postgres' behavior for disambiguating the name.
			return "", fmt.Errorf("unable to create foreign key %s: "+
				"a foreign key constraint already exists with this name", generatedFkName)
		}
	}

	return generatedFkName, nil
}
