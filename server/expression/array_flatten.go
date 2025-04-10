// Copyright 2025 Dolthub, Inc.
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

package expression

import (
	"github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/go-mysql-server/sql"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"
)

// ArrayFlatten is an expression that represents the results of a subquery expression as an array 
type ArrayFlatten struct {
		Subquery sql.Expression
}

var _ vitess.Injectable = (*ArrayFlatten)(nil)
var _ sql.Expression = (*ArrayFlatten)(nil)

func (a ArrayFlatten) Resolved() bool {
	return a.Subquery.Resolved()
}

func (a ArrayFlatten) String() string {
	return "ARRAY(" + a.Subquery.String() + ")"
}

func (a ArrayFlatten) Type() sql.Type {
	return types.AnyArray
}

func (a ArrayFlatten) IsNullable() bool {
	return false
}

func (a ArrayFlatten) Eval(ctx *sql.Context, row sql.Row) (interface{}, error) {
	result, err := a.Subquery.Eval(ctx, row)
	if err != nil {
		return nil, err
	}
	
	return result, nil
}

func (a ArrayFlatten) Children() []sql.Expression {
	return []sql.Expression{a.Subquery}
}

func (a ArrayFlatten) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 1 {
		return nil, sql.ErrInvalidChildrenNumber.New(a, len(children), 1)
	}
	
	return ArrayFlatten{Subquery: children[0]}, nil
}

func (a ArrayFlatten) WithResolvedChildren(children []any) (any, error) {
	if len(children) != 1 {
		return nil, sql.ErrInvalidChildrenNumber.New(a, len(children), 1)
	}

	subquery, ok := children[0].(sql.Expression)
	if !ok {
		return nil, sql.ErrInvalidChildType.New(a, 0, children[0])
	}

	return ArrayFlatten{Subquery: subquery}, nil
}
