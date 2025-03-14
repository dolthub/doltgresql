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
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"
)

// convertDropPrimaryKeyConstraint converts a DropConstraint node dropping a primary key constraint into
// an AlterPK node that GMS can process to remove the primary key.
func convertDropPrimaryKeyConstraint(ctx *sql.Context, _ *analyzer.Analyzer, n sql.Node, _ *plan.Scope, _ analyzer.RuleSelector, _ *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	return transform.Node(n, func(n sql.Node) (sql.Node, transform.TreeIdentity, error) {
		dropConstraint, ok := n.(*plan.DropConstraint)
		if !ok {
			return n, transform.SameTree, nil
		}

		rt, ok := dropConstraint.Child.(*plan.ResolvedTable)
		if !ok {
			return nil, transform.SameTree, analyzer.ErrInAnalysis.New(
				"Expected a TableNode for ALTER TABLE DROP CONSTRAINT statement")
		}

		table := rt.Table
		if it, ok := table.(sql.IndexAddressableTable); ok {
			indexes, err := it.GetIndexes(ctx)
			if err != nil {
				return nil, transform.SameTree, err
			}
			for _, index := range indexes {
				if index.ID() == "PRIMARY" && dropConstraint.Name == rt.Name()+"_pkey" {
					alterDropPk := plan.NewAlterDropPk(rt.Database(), rt)
					newNode, err := alterDropPk.WithTargetSchema(rt.Schema())
					if err != nil {
						return n, transform.SameTree, err
					}
					return newNode, transform.NewTree, nil
				}
			}
		}

		return n, transform.SameTree, nil
	})
}
