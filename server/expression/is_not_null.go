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
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// IsNotNull is an implementation of sql.Expression for the IS NOT NULL operator
// and includes Postgres-specific logic for handling records and composites.
type IsNotNull struct {
	expression.UnaryExpression
}

var _ sql.Expression = (*IsNotNull)(nil)
var _ sql.CollationCoercible = (*IsNotNull)(nil)
var _ sql.IsNotNullExpression = (*IsNotNull)(nil)

// NewIsNotNull creates a new IsNotNull expression.
func NewIsNotNull(child sql.Expression) *IsNotNull {
	return &IsNotNull{expression.UnaryExpression{Child: child}}
}

// IsNotNullExpression implements the sql.IsNotNullExpression interface. This function exists primarily
// to ensure the IsNotNullExpression interface has a unique signature.
func (e *IsNotNull) IsNotNullExpression() bool {
	return true
}

// Type implements the Expression interface.
func (e *IsNotNull) Type() sql.Type {
	return pgtypes.Bool
}

// CollationCoercibility implements the interface sql.CollationCoercible.
func (*IsNotNull) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	return sql.Collation_binary, 5
}

// IsNullable implements the Expression interface.
func (e *IsNotNull) IsNullable() bool {
	return false
}

// Eval implements the Expression interface.
func (e *IsNotNull) Eval(ctx *sql.Context, row sql.Row) (interface{}, error) {
	v, err := e.Child.Eval(ctx, row)
	if err != nil {
		return nil, err
	}

	// Slices of typed values (e.g. Record and Composite types in Postgres) evaluate
	// true for IS NOT NULL only if ALL of their entries are not NULL.
	if tupleValue, ok := v.([]pgtypes.RecordValue); ok {
		for _, typedValue := range tupleValue {
			if typedValue.Value == nil {
				return false, nil
			}
		}
		return true, nil
	}

	return v != nil, nil
}

func (e *IsNotNull) String() string {
	return e.Child.String() + " IS NOT NULL"
}

func (e *IsNotNull) DebugString() string {
	return sql.DebugString(e.Child) + " IS NOT NULL"
}

// WithChildren implements the Expression interface.
func (e *IsNotNull) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 1 {
		return nil, sql.ErrInvalidChildrenNumber.New(e, len(children), 1)
	}
	return NewIsNotNull(children[0]), nil
}
