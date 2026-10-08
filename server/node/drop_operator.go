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

package node

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// OperatorToDrop represents an operator in DROP OPERATOR.
type OperatorToDrop struct {
	Symbol string
	Left   *pgtypes.DoltgresType // This is nil for prefix operators
	Right  *pgtypes.DoltgresType
}

// DropOperator implements DROP OPERATOR.
type DropOperator struct {
	Operators []*OperatorToDrop
	IfExists  bool
}

var _ sql.ExecSourceRel = (*DropOperator)(nil)
var _ vitess.Injectable = (*DropOperator)(nil)

// NewDropOperator returns a new *DropOperator.
func NewDropOperator(ifExists bool, operators []*OperatorToDrop) *DropOperator {
	return &DropOperator{
		IfExists:  ifExists,
		Operators: operators,
	}
}

// Children implements the interface sql.ExecSourceRel.
func (d *DropOperator) Children() []sql.Node {
	return nil
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (d *DropOperator) IsReadOnly() bool {
	return false
}

// Resolved implements the interface sql.ExecSourceRel.
func (d *DropOperator) Resolved() bool {
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (d *DropOperator) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	opCollection, err := core.GetOperatorsCollectionFromContext(ctx, "")
	if err != nil {
		return nil, err
	}
	searchPath, err := core.SearchPath(ctx)
	if err != nil {
		return nil, err
	}
	for _, op := range d.Operators {
		leftTypeID := id.NullType
		if op.Left != nil {
			leftTypeID = op.Left.ID
		}
		operator, found, err := opCollection.ResolveOperator(ctx, searchPath, op.Symbol, leftTypeID, op.Right.ID)
		if err != nil {
			return nil, err
		}
		if !found {
			if d.IfExists {
				continue
			}
			if op.Left == nil {
				return nil, errors.Errorf("operator does not exist: %s %s", op.Symbol, op.Right.String())
			}
			return nil, errors.Errorf("operator does not exist: %s %s %s", op.Left.String(), op.Symbol, op.Right.String())
		}
		if err = opCollection.DropOperator(ctx, operator.ID); err != nil {
			return nil, err
		}
	}
	return sql.RowsToRowIter(), nil
}

// Schema implements the interface sql.ExecSourceRel.
func (d *DropOperator) Schema(ctx *sql.Context) sql.Schema {
	return nil
}

// String implements the interface sql.ExecSourceRel.
func (d *DropOperator) String() string {
	return "DROP OPERATOR"
}

// WithChildren implements the interface sql.ExecSourceRel.
func (d *DropOperator) WithChildren(ctx *sql.Context, children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(d, children...)
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (d *DropOperator) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return d, nil
}
