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
	"github.com/cockroachdb/errors"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// nodeLimit handles *tree.Limit nodes.
func nodeLimit(ctx *Context, node *tree.Limit) (*vitess.Limit, error) {
	if node == nil || (node.Count == nil && node.Offset == nil) {
		return nil, nil
	}
	var count vitess.Expr
	if !node.LimitAll {
		var err error
		count, err = nodeExpr(ctx, node.Count)
		if err != nil {
			return nil, err
		}
	}
	offset, err := nodeExpr(ctx, node.Offset)
	if err != nil {
		return nil, err
	}

	return &vitess.Limit{
		Offset:   offset,
		Rowcount: count,
	}, nil
}

// int64ValueForLimit converts a literal value to an int64
func int64ValueForLimit(l any) (int64, error) {
	var limitValue int64
	switch l := l.(type) {
	case int:
		limitValue = int64(l)
	case int32:
		limitValue = int64(l)
	case int64:
		limitValue = l
	case float64:
		limitValue = int64(l)
	case float32:
		limitValue = int64(l)
	default:
		return 0, errors.Errorf("unsupported limit/offset value type %T", l)
	}
	return limitValue, nil
}
