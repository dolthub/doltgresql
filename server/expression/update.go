// Copyright 2026 Dolthub, Inc.
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
	gmsexpression "github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/plan"
)

// UpdateExpressionApplier evaluates explicit UPDATE assignments against the pre-update row.
type UpdateExpressionApplier struct{}

var _ sql.UpdateExpressionApplier = UpdateExpressionApplier{}

// ApplyRowUpdate implements sql.UpdateExpressionApplier. PostgreSQL does not support
// UPDATE IGNORE, so assignment errors are always returned to the caller.
func (UpdateExpressionApplier) ApplyRowUpdate(ctx *sql.Context, updateExprs *sql.UpdateExprs, tableSchema sql.Schema, oldRow sql.Row, _ bool) (sql.Row, error) {
	newRow := oldRow.Copy()
	for _, expr := range updateExprs.ExplicitUpdateExprs() {
		assignment, ok := expr.(*gmsexpression.SetField)
		if !ok {
			return nil, fmt.Errorf("UPDATE: expected SetField, found %T", expr)
		}
		field, ok := assignment.LeftChild.(*gmsexpression.GetField)
		if !ok {
			return nil, fmt.Errorf("UPDATE: expected GetField target, found %T", assignment.LeftChild)
		}
		// SetField performs assignment conversion and returns a copy of oldRow.
		// Merge only its target, so later assignments cannot undo earlier writes.
		value, err := assignment.Eval(ctx, oldRow)
		if err != nil {
			return nil, err
		}
		newRow[field.Index()] = value.(sql.Row)[field.Index()]
	}

	if updateExprs.HasDerivedUpdates() {
		// Outer-scope values precede the table fields and are not in tableSchema.
		offset := len(oldRow) - len(tableSchema)
		same, err := oldRow[offset:].Equals(ctx, newRow[offset:], tableSchema)
		if err != nil {
			return nil, err
		}
		if !same {
			for _, expr := range updateExprs.DerivedUpdateExprs() {
				value, err := expr.Eval(ctx, newRow)
				if err != nil {
					return nil, err
				}
				var ok bool
				newRow, ok = value.(sql.Row)
				if !ok {
					return nil, plan.ErrUpdateUnexpectedSetResult.New(value)
				}
			}
		}
	}
	return newRow, nil
}
