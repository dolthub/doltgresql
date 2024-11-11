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
)

// nodeWhere handles *tree.Where nodes.
func nodeWhere(ctx *Context, node *tree.Where) (*vitess.Where, error) {
	if node == nil {
		return nil, nil
	}
	expr, err := nodeExpr(ctx, node.Expr)
	if err != nil {
		return nil, err
	}
	var whereType string
	switch node.Type {
	case tree.AstWhere:
		whereType = vitess.WhereStr
	case tree.AstHaving:
		whereType = vitess.HavingStr
	default:
		return nil, fmt.Errorf("WHERE-type statement not yet supported: `%s`", node.Type)
	}
	return &vitess.Where{
		Type: whereType,
		Expr: expr,
	}, nil
}
