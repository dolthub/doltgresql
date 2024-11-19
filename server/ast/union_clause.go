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

// nodeUnionClause handles tree.UnionClause nodes.
func nodeUnionClause(ctx *Context, node *tree.UnionClause) (*vitess.SetOp, error) {
	if node == nil {
		return nil, nil
	}
	left, err := nodeSelect(ctx, node.Left)
	if err != nil {
		return nil, err
	}
	right, err := nodeSelect(ctx, node.Right)
	if err != nil {
		return nil, err
	}
	var unionType string
	switch node.Type {
	case tree.UnionOp:
		if node.All {
			unionType = vitess.UnionAllStr
		} else {
			unionType = vitess.UnionStr
		}
	case tree.IntersectOp:
		if node.All {
			unionType = vitess.IntersectAllStr
		} else {
			unionType = vitess.IntersectStr
		}
	case tree.ExceptOp:
		if node.All {
			unionType = vitess.ExceptAllStr
		} else {
			unionType = vitess.ExceptStr
		}
	default:
		return nil, fmt.Errorf("unknown type of UNION operator: `%s`", node.Type.String())
	}
	return &vitess.SetOp{
		Type:  unionType,
		Left:  left,
		Right: right,
	}, nil
}
