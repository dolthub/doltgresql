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

package pgtransform

import (
	"errors"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	gmstransform "github.com/dolthub/go-mysql-server/sql/transform"
)

// InspectNode functions similarly to GMS' InspectUp function, except it also walks through opaque and disjointed nodes.
func InspectNode(node sql.Node, nodeFunc func(sql.Node) bool) bool {
	// This implementation is based on the one in GMS, except that we use our functions instead to handle disjointed
	stop := errors.New("stop")
	_, _, err := NodeWithOpaque(node, func(node sql.Node) (sql.Node, gmstransform.TreeIdentity, error) {
		ok := nodeFunc(node)
		if ok {
			return nil, gmstransform.NewTree, stop
		}
		return node, gmstransform.SameTree, nil
	})
	return errors.Is(err, stop)
}

// InspectNodeExprs functions similarly to GMS' InspectUp function, except that it traverses expressions (there is no
// InspectUp derivative for expressions in GMS), and it also walks through opaque and disjointed nodes.
func InspectNodeExprs(node sql.Node, exprFunc func(expr sql.Expression) bool) bool {
	// This implementation is based on the one in GMS, except that we use our functions instead to handle disjointed
	stop := errors.New("stop")
	_, _, err := NodeExprsWithOpaque(node, func(expr sql.Expression) (sql.Expression, gmstransform.TreeIdentity, error) {
		ok := exprFunc(expr)
		if ok {
			return nil, gmstransform.NewTree, stop
		}
		return expr, gmstransform.SameTree, nil
	})
	return errors.Is(err, stop)
}

// NodeWithOpaque functions similarly to GMS' NodeWithOpaque function, except it also walks through disjointed nodes.
func NodeWithOpaque(node sql.Node, nodeFunc gmstransform.NodeFunc) (sql.Node, gmstransform.TreeIdentity, error) {
	return gmstransform.NodeWithOpaque(node, func(node sql.Node) (sql.Node, gmstransform.TreeIdentity, error) {
		treeIdentity := gmstransform.SameTree
		if disjointedNode, ok := node.(plan.DisjointedChildrenNode); ok {
			var err error
			node, treeIdentity, err = handleDisjointedNodes(disjointedNode, func(node sql.Node) (sql.Node, gmstransform.TreeIdentity, error) {
				return NodeWithOpaque(node, nodeFunc)
			})
			if err != nil {
				return nil, gmstransform.NewTree, err
			}
		}
		node, newTreeIdentity, err := nodeFunc(node)
		if err != nil {
			return nil, gmstransform.NewTree, err
		}
		return node, treeIdentity && newTreeIdentity, nil
	})
}

// NodeExprsWithOpaque functions similarly to GMS' NodeExprsWithOpaque function, except it also walks through disjointed
// nodes.
func NodeExprsWithOpaque(node sql.Node, exprFunc gmstransform.ExprFunc) (sql.Node, gmstransform.TreeIdentity, error) {
	node, disjointCheck, err := gmstransform.NodeWithOpaque(node, func(node sql.Node) (sql.Node, gmstransform.TreeIdentity, error) {
		if disjointedNode, ok := node.(plan.DisjointedChildrenNode); ok {
			return handleDisjointedNodes(disjointedNode, func(node sql.Node) (sql.Node, gmstransform.TreeIdentity, error) {
				return NodeExprsWithOpaque(node, exprFunc)
			})
		}
		return node, gmstransform.SameTree, nil
	})
	if err != nil {
		return nil, gmstransform.NewTree, err
	}
	node, exprCheck, err := gmstransform.NodeExprsWithOpaque(node, exprFunc)
	if err != nil {
		return nil, gmstransform.NewTree, err
	}
	return node, disjointCheck && exprCheck, nil
}

// NodeExprsWithNodeWithOpaque functions similarly to GMS' NodeExprsWithNodeWithOpaque function, except it also walks
// through disjointed nodes.
func NodeExprsWithNodeWithOpaque(node sql.Node, exprFunc gmstransform.ExprWithNodeFunc) (sql.Node, gmstransform.TreeIdentity, error) {
	node, disjointCheck, err := gmstransform.NodeWithOpaque(node, func(node sql.Node) (sql.Node, gmstransform.TreeIdentity, error) {
		if disjointedNode, ok := node.(plan.DisjointedChildrenNode); ok {
			return handleDisjointedNodes(disjointedNode, func(node sql.Node) (sql.Node, gmstransform.TreeIdentity, error) {
				return NodeExprsWithNodeWithOpaque(node, exprFunc)
			})
		}
		return node, gmstransform.SameTree, nil
	})
	if err != nil {
		return nil, gmstransform.NewTree, err
	}
	node, exprCheck, err := gmstransform.NodeExprsWithNodeWithOpaque(node, exprFunc)
	if err != nil {
		return nil, gmstransform.NewTree, err
	}
	return node, disjointCheck && exprCheck, nil
}

// handleDisjointedNodes handles disjointed nodes for the typical transform functions. This also includes the call on
// the given disjointed node, so the caller should avoid making the call themselves.
func handleDisjointedNodes(node plan.DisjointedChildrenNode, f func(sql.Node) (sql.Node, gmstransform.TreeIdentity, error)) (sql.Node, gmstransform.TreeIdentity, error) {
	disjointedChildren := node.DisjointedChildren()
	tree := gmstransform.SameTree
	newChildren := make([][]sql.Node, len(disjointedChildren))
	for firstIndex := range disjointedChildren {
		newSubChildren := make([]sql.Node, len(disjointedChildren[firstIndex]))
		for secondIndex := range disjointedChildren[firstIndex] {
			newSubChild, newTree, err := f(disjointedChildren[firstIndex][secondIndex])
			if err != nil {
				return nil, gmstransform.NewTree, err
			}
			tree = tree && newTree
			newSubChildren[secondIndex] = newSubChild
		}
		newChildren[firstIndex] = newSubChildren
	}
	if tree == gmstransform.SameTree {
		return node, gmstransform.SameTree, nil
	}
	newNode, err := node.WithDisjointedChildren(newChildren)
	return newNode, gmstransform.NewTree, err
}
