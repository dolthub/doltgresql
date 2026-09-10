// Copyright 2026 Dolthub, Inc.
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
	"unicode/utf8"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	gms_expression "github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"

	"github.com/dolthub/doltgresql/server/expression"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// AddLikePrefixRanges bounds each LIKE in a filter by the fixed prefix of its pattern, so that an index on the column
// can serve it under the C collation.
func AddLikePrefixRanges(ctx *sql.Context, a *analyzer.Analyzer, node sql.Node, scope *plan.Scope, selector analyzer.RuleSelector, qFlags *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	return transform.Node(ctx, node, func(ctx *sql.Context, node sql.Node) (sql.Node, transform.TreeIdentity, error) {
		filter, ok := node.(*plan.Filter)
		if !ok {
			return node, transform.SameTree, nil
		}
		newExpr, same, err := addLikePrefixBounds(ctx, filter.Expression)
		if err != nil || same {
			return node, transform.SameTree, err
		}
		newNode, err := filter.WithExpressions(ctx, newExpr)
		return newNode, transform.NewTree, err
	})
}

// addLikePrefixBounds returns `e` with prefix bounds added to every LIKE reachable through AND.
func addLikePrefixBounds(ctx *sql.Context, e sql.Expression) (sql.Expression, transform.TreeIdentity, error) {
	switch e := e.(type) {
	case *gms_expression.And:
		children := e.Children()
		newChildren := make([]sql.Expression, len(children))
		same := transform.SameTree
		for i, child := range children {
			newChild, childSame, err := addLikePrefixBounds(ctx, child)
			if err != nil {
				return nil, transform.SameTree, err
			}
			newChildren[i] = newChild
			same = same && childSame
		}
		if same {
			return e, transform.SameTree, nil
		}
		newExpr, err := e.WithChildren(ctx, newChildren...)
		return newExpr, transform.NewTree, err
	case *gms_expression.Like:
		bounds, err := likePrefixBounds(ctx, e)
		if err != nil || bounds == nil {
			return e, transform.SameTree, err
		}
		return gms_expression.JoinAnd(append(bounds, e)...), transform.NewTree, nil
	}
	return e, transform.SameTree, nil
}

// likePrefixBounds returns the comparisons bounding a text or varchar column to the prefix, or nil when the pattern has
// no such prefix.
func likePrefixBounds(ctx *sql.Context, like *gms_expression.Like) ([]sql.Expression, error) {
	column, ok := like.LeftChild.(*gms_expression.GetField)
	if !ok || like.Escape != nil {
		return nil, nil
	}
	columnType, ok := column.Type(ctx).(*pgtypes.DoltgresType)
	if !ok || (columnType.ID != pgtypes.Text.ID && columnType.ID != pgtypes.VarChar.ID) {
		return nil, nil
	}
	literal, ok := like.RightChild.(*gms_expression.Literal)
	if !ok {
		return nil, nil
	}
	pattern, ok := literal.Value().(string)
	if !ok {
		return nil, nil
	}
	prefix := likePrefix(pattern)
	if prefix == "" {
		return nil, nil
	}
	lower, err := expression.NewBinaryOperator(framework.Operator_BinaryGreaterOrEqual).
		WithResolvedChildren(ctx, []any{column, expression.NewTextLiteral(prefix)})
	if err != nil {
		return nil, err
	}
	bounds := []sql.Expression{lower.(sql.Expression)}
	next, ok := incrementLastRune(prefix)
	if !ok {
		return bounds, nil
	}
	upper, err := expression.NewBinaryOperator(framework.Operator_BinaryLessThan).
		WithResolvedChildren(ctx, []any{column, expression.NewTextLiteral(next)})
	if err != nil {
		return nil, err
	}
	return append(bounds, upper.(sql.Expression)), nil
}

// likePrefix returns the characters of `pattern` before its first wildcard, or an empty string when an escape precedes them.
func likePrefix(pattern string) string {
	for i, r := range pattern {
		switch r {
		case '%', '_':
			return pattern[:i]
		case '\\':
			return ""
		}
	}
	return pattern
}

// incrementLastRune returns the smallest string greater than every string with the prefix `prefix`, and false when
// there is no such string. For example, the increment of `abc` is `abd`.
func incrementLastRune(prefix string) (string, bool) {
	last, size := utf8.DecodeLastRuneInString(prefix)
	if last == utf8.MaxRune {
		return "", false
	}
	next := last + 1
	if next == 0xD800 {
		next = 0xE000
	}
	return prefix[:len(prefix)-size] + string(next), true
}
