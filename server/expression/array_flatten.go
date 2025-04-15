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
	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	gmstypes "github.com/dolthub/go-mysql-server/sql/types"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/server/types"
)

// ArrayFlatten is an expression that represents the results of a subquery expression as an array.
// Currently only subqueries that return a single column are supported.
type ArrayFlatten struct {
	Subquery sql.Expression
}

var _ vitess.Injectable = (*ArrayFlatten)(nil)
var _ sql.Expression = (*ArrayFlatten)(nil)

// Resolved implements sql.Expression.
func (a ArrayFlatten) Resolved() bool {
	return a.Subquery.Resolved()
}

// String implements sql.Expression.
func (a ArrayFlatten) String() string {
	return "ARRAY(" + a.Subquery.String() + ")"
}

// Type implements sql.Expression.
func (a ArrayFlatten) Type() sql.Type {
	sqType := a.Subquery.Type()
	dt, ok := sqType.(*types.DoltgresType)
	if !ok {
		// If we don't have a doltgres type, we'll error out at execution time. A special case is the tuple type,
		// where we need to choose a single one to avoid erroring out too early.
		if tt, ok := sqType.(gmstypes.TupleType); ok {
			return tt[0]
		}
		return sqType
	}
	return dt.ToArrayType()
}

// IsNullable implements sql.Expression.
func (a ArrayFlatten) IsNullable() bool {
	return false
}

// Eval implements sql.Expression.
func (a ArrayFlatten) Eval(ctx *sql.Context, row sql.Row) (interface{}, error) {
	subquery, ok := a.Subquery.(*plan.Subquery)
	if !ok {
		return nil, errors.Errorf("expected subquery, got %T", a.Subquery)
	}

	sqType := subquery.Type()
	_, ok = sqType.(*types.DoltgresType)
	if !ok {
		if tt, ok := sqType.(gmstypes.TupleType); ok {
			return nil, errors.Errorf("only a single column subquery is supported in ARRAY(), got %d columns", len(tt))
		}
		return nil, errors.Errorf("expected doltgres type, got %T", sqType)
	}

	return subquery.EvalMultiple(ctx, row)
}

// Children implements sql.Expression.
func (a ArrayFlatten) Children() []sql.Expression {
	return []sql.Expression{a.Subquery}
}

// WithChildren implements sql.Expression.
func (a ArrayFlatten) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 1 {
		return nil, sql.ErrInvalidChildrenNumber.New(a, len(children), 1)
	}

	return ArrayFlatten{Subquery: children[0]}, nil
}

// WithResolvedChildren implements vitess.Injectable.
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
