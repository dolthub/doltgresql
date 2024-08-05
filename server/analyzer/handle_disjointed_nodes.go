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
)

// handleDisjointedNodes handles disjointed nodes.
func handleDisjointedNodes(ctx *sql.Context, a *analyzer.Analyzer, node plan.DisjointedChildrenNode, scope *plan.Scope, selector analyzer.RuleSelector, ruleFunc analyzer.RuleFunc, qFlags *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	// TODO: should move this to the transform package in GMS, rather than have it here
	disjointedChildren := node.DisjointedChildren()
	tree := transform.SameTree
	newChildren := make([][]sql.Node, len(disjointedChildren))
	for firstIndex := range disjointedChildren {
		newSubChildren := make([]sql.Node, len(disjointedChildren[firstIndex]))
		for secondIndex := range disjointedChildren[firstIndex] {
			newSubChild, newTree, err := ruleFunc(ctx, a, disjointedChildren[firstIndex][secondIndex], scope, selector, qFlags)
			if err != nil {
				return nil, transform.NewTree, err
			}
			tree = tree && newTree
			newSubChildren[secondIndex] = newSubChild
		}
		newChildren[firstIndex] = newSubChildren
	}
	if tree == transform.SameTree {
		return node, transform.SameTree, nil
	}
	newNode, err := node.WithDisjointedChildren(newChildren)
	return newNode, transform.NewTree, err
}
