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
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"

	pgnodes "github.com/dolthub/doltgresql/server/node"
)

// InsertContextRootFinalizer inserts a ContextRootFinalizer node right before the transaction commits, yet after all
// other nodes have finished. This ensures that the ContextRootFinalizer does not overwrite any changes from its
// children.
func InsertContextRootFinalizer(ctx *sql.Context, a *analyzer.Analyzer, node sql.Node, scope *plan.Scope, selector analyzer.RuleSelector) (sql.Node, transform.TreeIdentity, error) {
	if _, ok := node.(*pgnodes.ContextRootFinalizer); ok {
		return node, transform.SameTree, nil
	}
	// Analysis may occur separately on child nodes, so we have to ensure that only one finalizer exists in the tree
	newNode, _, err := transform.NodeWithOpaque(node, transformRemoveContextRootFinalizer)
	if err != nil {
		return nil, transform.NewTree, err
	}
	return pgnodes.NewContextRootFinalizer(newNode), transform.NewTree, nil
}

// transformRemoveContextRootFinalizer is the function used by the transform from within InsertContextRootFinalizer.
func transformRemoveContextRootFinalizer(node sql.Node) (sql.Node, transform.TreeIdentity, error) {
	if finalizer, ok := node.(*pgnodes.ContextRootFinalizer); ok {
		return finalizer.Child(), transform.NewTree, nil
	} else if disjointedNode, ok := node.(plan.DisjointedChildrenNode); ok {
		var err error
		same := transform.SameTree
		disjointedChildGroups := disjointedNode.DisjointedChildren()
		newDisjointedChildGroups := make([][]sql.Node, len(disjointedChildGroups))
		for groupIdx, disjointedChildGroup := range disjointedChildGroups {
			newDisjointedChildGroups[groupIdx] = make([]sql.Node, len(disjointedChildGroup))
			for childIdx, disjointedChild := range disjointedChildGroup {
				var childIdentity transform.TreeIdentity
				newDisjointedChildGroups[groupIdx][childIdx], childIdentity, err = transform.NodeWithOpaque(disjointedChild, transformRemoveContextRootFinalizer)
				if err != nil {
					return nil, transform.NewTree, err
				}
				same = same && childIdentity
			}
		}
		if same == transform.NewTree {
			if newChild, err := disjointedNode.WithDisjointedChildren(newDisjointedChildGroups); err != nil {
				return nil, transform.NewTree, err
			} else {
				return newChild, transform.NewTree, nil
			}
		}
	}
	return node, transform.SameTree, nil
}
