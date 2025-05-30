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
	"strings"

	"github.com/cockroachdb/errors"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	pgexprs "github.com/dolthub/doltgresql/server/expression"
)

// nodeFuncExpr handles *tree.FuncExpr nodes.
func nodeFuncExpr(ctx *Context, node *tree.FuncExpr) (vitess.Expr, error) {
	if node == nil {
		return nil, nil
	}
	if node.Filter != nil {
		return nil, errors.Errorf("function filters are not yet supported")
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

	switch strings.ToLower(name.String()) {
	// special case for string_agg, which maps to the mysql aggregate function group_concat
	case "string_agg":
		if len(node.Exprs) != 2 {
			return nil, errors.Errorf("string_agg requires two arguments")
		}
		sep, ok := node.Exprs[1].(*tree.StrVal)
		if !ok {
			return nil, errors.Errorf("string_agg requires a string separator")
		}
		sepString := strings.Trim(sep.String(), "'")

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
				Expression:         &pgexprs.ArrayAgg{},
				SelectExprChildren: exprs,
				Auth:               vitess.AuthInformation{},
			},
			OrderBy: orderBy,
		}, nil
	}

	if len(node.OrderBy) > 0 {
		return nil, errors.Errorf("function ORDER BY is not yet supported")
	}

	return &vitess.FuncExpr{
		Qualifier: qualifier,
		Name:      name,
		Distinct:  distinct,
		Exprs:     exprs,
		Over:      (*vitess.Over)(windowDef),
	}, nil
}
