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
	"context"
	"fmt"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/server/types"
)

// Subscript represents a subscript expression, e.g. `a[1]` or `a[1][2]`.
type Subscript struct {
	Child   sql.Expression
	Indexes []sql.Expression
}

var _ vitess.Injectable = (*Subscript)(nil)
var _ sql.Expression = (*Subscript)(nil)

// NewSubscript creates a new Subscript expression.
func NewSubscript(child sql.Expression, indexes ...sql.Expression) *Subscript {
	return &Subscript{
		Child:   child,
		Indexes: indexes,
	}
}

// Resolved implements the sql.Expression interface.
func (s Subscript) Resolved() bool {
	for _, index := range s.Indexes {
		if !index.Resolved() {
			return false
		}
	}
	return s.Child.Resolved()
}

// String implements the sql.Expression interface.
func (s Subscript) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprint(s.Child))
	for _, index := range s.Indexes {
		sb.WriteString(fmt.Sprintf("[%s]", index))
	}
	return sb.String()
}

// Type implements the sql.Expression interface.
func (s Subscript) Type(ctx *sql.Context) sql.Type {

	dt, ok := s.childType(ctx)
	if !ok {
		return types.Unknown
		//panic(fmt.Sprintf("unexpected type %T for subscript", s.Child.Type(ctx)))
	}
	// can be either array type or vector type, so use its base type if it exists
	return dt.BaseType()
}

// IsNullable implements the sql.Expression interface.
func (s Subscript) IsNullable(ctx *sql.Context) bool {
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

	scalarElements := false
	if dt, ok := s.childType(ctx); ok {
		scalarElements = dt.BaseType().IsArrayCategory()
	}
	for i, indexExpr := range s.Indexes {
		indexVal, err := indexExpr.Eval(ctx, row)
		if err != nil {
			return nil, err
		}
		if indexVal == nil {
			return nil, nil
		}
		array, ok := childVal.([]any)
		if !ok {
			if i == 0 {
				return nil, fmt.Errorf("unsupported type %T for subscript", childVal)
			}
			return nil, nil
		}
		if i > 0 && scalarElements {
			return nil, nil
		}
		index, ok := indexVal.(int32)
		if !ok {
			converted, _, err := types.Int32.Convert(ctx, indexVal)
			if err != nil {
				return nil, err
			}
			index = converted.(int32)
		}

		// subscripts are 1-based
		if index < 1 || int(index) > len(array) {
			return nil, nil
		}
		childVal = array[index-1]
	}
	if _, isSubArray := childVal.([]any); isSubArray && !scalarElements {
		return nil, nil
	}
	return childVal, nil
}

// childType returns the type of the subscripted expression, with a domain resolved to its underlying type.
func (s Subscript) childType(ctx *sql.Context) (*types.DoltgresType, bool) {
	dt, ok := s.Child.Type(ctx).(*types.DoltgresType)
	if ok && dt.TypType == types.TypeType_Domain {
		dt = dt.DomainUnderlyingBaseType()
	}
	return dt, ok
}

// Children implements the sql.Expression interface.
func (s Subscript) Children() []sql.Expression {
	return append([]sql.Expression{s.Child}, s.Indexes...)
}

// WithChildren implements the sql.Expression interface.
func (s Subscript) WithChildren(ctx *sql.Context, children ...sql.Expression) (sql.Expression, error) {
	if len(children) < 2 {
		return nil, fmt.Errorf("expected at least 2 children, got %d", len(children))
	}
	// The subscript index is always int4, regardless of the array's element type, so an untyped bind variable
	// (e.g. `arr[$1]`) can be resolved to int4 immediately.
	for _, index := range children[1:] {
		if bv, ok := index.(*expression.BindVar); ok {
			if _, ok := bv.Typ.(*types.DoltgresType); !ok {
				bv.Typ = types.Int32
			}
		}
	}
	return NewSubscript(children[0], children[1:]...), nil
}

// WithResolvedChildren implements the vitess.Injectable interface.
func (s Subscript) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	expressions := make([]sql.Expression, len(children))
	for i, child := range children {
		var ok bool
		if expressions[i], ok = child.(sql.Expression); !ok {
			return nil, fmt.Errorf("expected child to be an expression but has type `%T`", child)
		}
	}
	return s.WithChildren(ctx.(*sql.Context), expressions...)
}
