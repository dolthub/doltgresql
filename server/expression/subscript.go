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
	"fmt"

	"github.com/dolthub/go-mysql-server/sql"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/server/types"
)

type Subscript struct {
	Child sql.Expression
	Index sql.Expression
}

var _ vitess.Injectable = (*Subscript)(nil)
var _ sql.Expression = (*Subscript)(nil)

func NewSubscript(child, index sql.Expression) *Subscript {
	return &Subscript{
		Child: child,
		Index: index,
	}
}

func (s Subscript) Resolved() bool {
	return s.Child.Resolved() && s.Index.Resolved()
}

func (s Subscript) String() string {
	return fmt.Sprintf("%s[%s]", s.Child, s.Index)
}

func (s Subscript) Type() sql.Type {
	dt, ok := s.Child.Type().(*types.DoltgresType)
	if !ok {
		panic(fmt.Sprintf("unexpected type %T for subscript", s.Child.Type()))
	}
	return dt.ArrayBaseType()
}

func (s Subscript) IsNullable() bool {
	return true
}

// Eval implements the sql.Expression interface.
func (s Subscript) Eval(ctx *sql.Context, row sql.Row) (interface{}, error) {
	childVal, err := s.Child.Eval(ctx, row)
	if err != nil {
		return nil, err
	}
	if childVal == nil {
		return nil, nil
	}

	indexVal, err := s.Index.Eval(ctx, row)
	if err != nil {
		return nil, err
	}
	if indexVal == nil {
		return nil, nil
	}

	switch child := childVal.(type) {
	case []interface{}:
		index, ok := indexVal.(int32)
		if !ok {
			converted, _, err := types.Int32.Convert(ctx, indexVal)
			if err != nil {
				return nil, err
			}
			index = converted.(int32)
		}

		// subscripts are 1-based
		if index < 1 || int(index) > len(child) {
			return nil, nil
		}
		return child[index-1], nil
	default:
		return nil, fmt.Errorf("unsupported type %T for subscript", child)
	}
}

func (s Subscript) Children() []sql.Expression {
	return []sql.Expression{s.Child, s.Index}
}

func (s Subscript) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 2 {
		return nil, fmt.Errorf("expected 2 children, got %d", len(children))
	}
	return NewSubscript(children[0], children[1]), nil
}

func (s Subscript) WithResolvedChildren(children []any) (any, error) {
	if len(children) != 2 {
		return nil, fmt.Errorf("expected 2 children, got %d", len(children))
	}
	child, ok := children[0].(sql.Expression)
	if !ok {
		return nil, fmt.Errorf("expected child to be an expression but has type `%T`", children[0])
	}
	index, ok := children[1].(sql.Expression)
	if !ok {
		return nil, fmt.Errorf("expected index to be an expression but has type `%T`", children[1])
	}

	return NewSubscript(child, index), nil
}
