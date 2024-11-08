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
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	pgexprs "github.com/dolthub/doltgresql/server/expression"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// nodeGroupBy handles tree.GroupBy nodes.
func nodeGroupBy(ctx *Context, node tree.GroupBy) (vitess.GroupBy, error) {
	if len(node) == 0 {
		return nil, nil
	}

	groupBys := make(vitess.GroupBy, len(node))
	var err error
	for i, expr := range node {
		groupBys[i], err = nodeExpr(ctx, expr)
		if err != nil {
			return nil, err
		}

		// GMS order by is hardcoded to expect vitess.SQLVal for expressions such as `ORDER BY 1`.
		// In addition, there is the requirement that columns in the order by also need to be referenced somewhere in
		// the query, which is not a requirement for Postgres. Whenever we add that functionality, we also need to
		// remove the dependency on vitess.SQLVal. For now, we'll just convert our literals to a vitess.SQLVal.
		if injectedExpr, ok := groupBys[i].(vitess.InjectedExpr); ok {
			if literal, ok := injectedExpr.Expression.(*pgexprs.Literal); ok {
				groupBys[i] = literal.ToVitessLiteral()
			}
		}
	}

	return groupBys, nil
}
