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
	"fmt"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/server/auth"
)

// nodeUpdate handles *tree.Update nodes.
func nodeUpdate(ctx *Context, node *tree.Update) (*vitess.Update, error) {
	if node == nil {
		return nil, nil
	}
	ctx.Auth().PushAuthType(auth.AuthType_UPDATE)
	defer ctx.Auth().PopAuthType()

	if _, ok := node.Returning.(*tree.NoReturningClause); !ok {
		return nil, fmt.Errorf("RETURNING is not yet supported")
	}
	if len(node.From) > 0 {
		return nil, fmt.Errorf("FROM is not yet supported")
	}
	with, err := nodeWith(ctx, node.With)
	if err != nil {
		return nil, err
	}
	table, err := nodeTableExpr(ctx, node.Table)
	if err != nil {
		return nil, err
	}
	exprs, err := nodeUpdateExprs(ctx, node.Exprs)
	if err != nil {
		return nil, err
	}
	where, err := nodeWhere(ctx, node.Where)
	if err != nil {
		return nil, err
	}
	orderBy, err := nodeOrderBy(ctx, node.OrderBy)
	if err != nil {
		return nil, err
	}
	limit, err := nodeLimit(ctx, node.Limit)
	if err != nil {
		return nil, err
	}
	return &vitess.Update{
		TableExprs: vitess.TableExprs{table},
		With:       with,
		Exprs:      exprs,
		Where:      where,
		OrderBy:    orderBy,
		Limit:      limit,
	}, nil
}
