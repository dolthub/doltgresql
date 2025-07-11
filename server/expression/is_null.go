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

// IsNull is an implementation of sql.Expression for the IS NULL operator and
// includes Postgres-specific logic for handling records and composites.
type IsNull struct {
	expression.UnaryExpression
}

var _ sql.Expression = (*IsNull)(nil)
var _ sql.CollationCoercible = (*IsNull)(nil)
var _ sql.IsNullExpression = (*IsNull)(nil)

// NewIsNull creates a new IsNull expression.
func NewIsNull(child sql.Expression) *IsNull {
	return &IsNull{expression.UnaryExpression{Child: child}}
}

// IsNullExpression implements the sql.IsNullExpression interface. This function exists primarily
// to ensure the IsNullExpression interface has a unique signature.
func (e *IsNull) IsNullExpression() bool {
	return true
}

// Type implements the Expression interface.
func (e *IsNull) Type() sql.Type {
	return pgtypes.Bool
}

// CollationCoercibility implements the interface sql.CollationCoercible.
func (*IsNull) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	return sql.Collation_binary, 5
}

// IsNullable implements the Expression interface.
func (e *IsNull) IsNullable() bool {
	return false
}

// Eval implements the Expression interface.
func (e *IsNull) Eval(ctx *sql.Context, row sql.Row) (interface{}, error) {
	v, err := e.Child.Eval(ctx, row)
	if err != nil {
		return nil, err
	}

	// Slices of typed values (e.g. Record and Composite types in Postgres) evaluate
	// to NULL if all of their entries are NULL.
	if tupleValue, ok := v.([]pgtypes.RecordValue); ok {
		for _, typedValue := range tupleValue {
			if typedValue.Value != nil {
				return false, nil
			}
		}
		return true, nil
	}

	return v == nil, nil
}

func (e *IsNull) String() string {
	return e.Child.String() + " IS NULL"
}

func (e *IsNull) DebugString() string {
	return sql.DebugString(e.Child) + " IS NULL"
}

// WithChildren implements the Expression interface.
func (e *IsNull) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 1 {
		return nil, sql.ErrInvalidChildrenNumber.New(e, len(children), 1)
	}
	return NewIsNull(children[0]), nil
}
