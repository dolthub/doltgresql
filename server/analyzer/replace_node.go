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

	pgnodes "github.com/dolthub/doltgresql/server/node"
)

// ReplaceNode is used to replace generic top-level nodes with Doltgres versions that wrap them, without performing any
// additional analysis. This is used to handle relatively straightforward tasks, like delete cascading, etc.
func ReplaceNode(ctx *sql.Context, a *analyzer.Analyzer, node sql.Node, scope *plan.Scope, selector analyzer.RuleSelector, qFlags *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	// TODO: need to add the majority of other DDL operations here
	switch node := node.(type) {
	case *plan.Block:
		// Multi-action ALTER TABLE statements are represented as a block.
		if node.Pref != nil {
			return node, transform.SameTree, nil
		}
		children := append([]sql.Node(nil), node.Children()...)
		changed := false
		for i, child := range children {
			replacement, identity, err := ReplaceNode(ctx, a, child, scope, selector, qFlags)
			if err != nil {
				return nil, transform.SameTree, err
			}
			children[i] = replacement
			changed = changed || identity == transform.NewTree
		}
		if !changed {
			return node, transform.SameTree, nil
		}
		replacement, err := node.WithChildren(ctx, children...)
		return replacement, transform.NewTree, err
	case *plan.ModifyColumn:
		widening, err := pgnodes.NewForeignKeyIntegerWidening(ctx, node, a.Catalog.DbProvider)
		if err != nil {
			return nil, transform.SameTree, err
		}
		if widening != nil {
			return widening, transform.NewTree, nil
		}
		return node, transform.SameTree, nil
	case *plan.DropTable:
		return pgnodes.NewDropTable(node), transform.NewTree, nil
	default:
		return node, transform.SameTree, nil
	}
}
