// Copyright 2023 Dolthub, Inc.
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

package ast

import (
	"go/constant"
	"strings"

	"github.com/cockroachdb/errors"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/server/auth"
	pgexprs "github.com/dolthub/doltgresql/server/expression"
)

// nodeFuncExpr handles *tree.FuncExpr nodes.
func nodeFuncExpr(ctx *Context, node *tree.FuncExpr) (vitess.Expr, error) {
	if node == nil {
		return nil, nil
	}
	if node.Filter != nil {
		var err error
		if node, err = rewriteAggregateFilter(node); err != nil {
			return nil, err
		}
	}
	if node.AggType == tree.OrderedSetAgg {
		return nil, errors.Errorf("WITHIN GROUP is not yet supported")
	}

	var qualifier vitess.TableIdent
	var name vitess.ColIdent
	switch funcRef := node.Func.FunctionReference.(type) {
	case *tree.FunctionDefinition:
		name = vitess.NewColIdent(funcRef.Name)
	case *tree.UnresolvedName:
		colName, err := unresolvedNameToColName(funcRef)
		if err != nil {
			return nil, err
		}
		qualifier = colName.Qualifier.Name
		name = colName.Name
	default:
		return nil, errors.Errorf("unknown function reference")
	}
	var distinct bool
	switch node.Type {
	case 0, tree.AllFuncType:
		distinct = false
	case tree.DistinctFuncType:
		distinct = true
	default:
		return nil, errors.Errorf("unknown function spec type %d", node.Type)
	}
	windowDef, err := nodeWindowDef(ctx, node.WindowDef)
	if err != nil {
		return nil, err
	}
	exprs, err := nodeExprsToSelectExprs(ctx, node.Exprs)
	if err != nil {
		return nil, err
	}
	for i, expr := range exprs {
		exprs[i] = catalogRowArgument(ctx, expr)
	}

	switch strings.ToLower(name.String()) {
	// special case for string_agg, which maps to the mysql aggregate function group_concat
	case "string_agg":
		if len(node.Exprs) != 2 {
			return nil, errors.Errorf("string_agg requires two arguments")
		}

		sepString := ""
		if sep, ok := node.Exprs[1].(*tree.StrVal); ok {
			sepString = strings.Trim(sep.String(), "'")
		} else {
			// TODO: need to support this function in doltgres
			c, is := node.Exprs[1].(*tree.CastExpr)
			if !is && c.Type.SQLString() != "TEXT" {
				return nil, errors.Errorf("string_agg requires a string separator")
			}
			sepString = strings.Trim(c.Expr.String(), "'")
		}

		var orderBy vitess.OrderBy
		if len(node.OrderBy) > 0 {
			orderBy, err = nodeOrderBy(ctx, node.OrderBy)
			if err != nil {
				return nil, err
			}
		}

		return &vitess.GroupConcatExpr{
			Exprs: exprs[:1],
			Separator: vitess.Separator{
				SeparatorString: sepString,
			},
			OrderBy: orderBy,
		}, nil
	case "array_agg":
		var orderBy vitess.OrderBy
		if len(node.OrderBy) > 0 {
			orderBy, err = nodeOrderBy(ctx, node.OrderBy)
			if err != nil {
				return nil, err
			}
		}

		return &vitess.OrderedInjectedExpr{
			InjectedExpr: vitess.InjectedExpr{
				Expression:         &pgexprs.ArrayAgg{Distinct: distinct},
				SelectExprChildren: exprs,
				Auth:               vitess.AuthInformation{},
			},
			OrderBy: orderBy,
		}, nil
	case "greatest":
		return vitess.InjectedExpr{
			Expression:         &pgexprs.Greatest{},
			SelectExprChildren: exprs,
			Auth:               vitess.AuthInformation{},
		}, nil
	case "least":
		return vitess.InjectedExpr{
			Expression:         &pgexprs.Least{},
			SelectExprChildren: exprs,
			Auth:               vitess.AuthInformation{},
		}, nil
	}

	if len(node.OrderBy) > 0 {
		return nil, errors.Errorf("function ORDER BY is not yet supported")
	}
	if !strings.EqualFold(name.String(), "count") {
		exprs = dropStarArgument(exprs)
	}

	return &vitess.FuncExpr{
		Qualifier: qualifier,
		Name:      name,
		Distinct:  distinct,
		Exprs:     exprs,
		Over:      (*vitess.Over)(windowDef),
		Auth: vitess.AuthInformation{
			AuthType:    auth.AuthType_EXECUTE,
			TargetType:  auth.AuthTargetType_FunctionIdentifiers,
			TargetNames: []string{qualifier.String(), name.String()},
		},
	}, nil
}

// dropStarArgument removes a lone unqualified `*` argument, which PostgreSQL parses as a call with no arguments. Only
// count has a form that accepts it, so every other function resolves as though it were given none. A qualified star
// such as `t.*` names the whole row and is left alone.
func dropStarArgument(exprs vitess.SelectExprs) vitess.SelectExprs {
	if len(exprs) != 1 {
		return exprs
	}
	if starExpr, ok := exprs[0].(*vitess.StarExpr); !ok || !starExpr.TableName.IsEmpty() {
		return exprs
	}
	return nil
}

// filterableAggregates are the aggregates that ignore NULL inputs. For these, FILTER (WHERE c)
// is the same as passing NULL for every row where c is not true.
var filterableAggregates = map[string]struct{}{
	"avg": {}, "bit_and": {}, "bit_or": {}, "bool_and": {}, "bool_or": {}, "count": {},
	"every": {}, "max": {}, "min": {}, "stddev": {}, "stddev_pop": {}, "stddev_samp": {},
	"string_agg": {}, "sum": {}, "var_pop": {}, "var_samp": {}, "variance": {},
}

// rewriteAggregateFilter turns agg(x) FILTER (WHERE c) into agg(CASE WHEN c THEN x END), and
// count(*) FILTER (WHERE c) into count(CASE WHEN c THEN 1 END). Only the aggregated value is
// filtered; later arguments, such as the separator of string_agg, are left as they are.
func rewriteAggregateFilter(node *tree.FuncExpr) (*tree.FuncExpr, error) {
	var name string
	switch ref := node.Func.FunctionReference.(type) {
	case *tree.FunctionDefinition:
		name = ref.Name
	case *tree.UnresolvedName:
		name = ref.Parts[0]
	}
	name = strings.ToLower(name)
	if _, ok := filterableAggregates[name]; !ok || len(node.Exprs) == 0 {
		return nil, errors.Errorf("FILTER is not yet supported for %s", name)
	}
	filtered := *node
	filtered.Filter = nil
	filtered.Exprs = append(tree.Exprs(nil), node.Exprs...)
	value := filtered.Exprs[0]
	switch value.(type) {
	case tree.UnqualifiedStar, *tree.UnqualifiedStar:
		value = tree.NewNumVal(constant.MakeInt64(1), "1", false)
	}
	filtered.Exprs[0] = &tree.CaseExpr{Whens: []*tree.When{{Cond: node.Filter, Val: value}}}
	return &filtered, nil
}
