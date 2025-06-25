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

package node

import (
	"fmt"

	"github.com/dolthub/go-mysql-server/sql"
)

var _ sql.Expressioner = (*SetReturningFunctionTable)(nil)
var _ sql.ExecSourceRel = (*SetReturningFunctionTable)(nil)

// SetReturningFunctionTable is a node for set returning function.
type SetReturningFunctionTable struct {
	Name     string
	sch      sql.Schema
	function sql.Expression
}

func NewSetReturningFunctionTable(ctx *sql.Context, name string, e sql.Expression) (*SetReturningFunctionTable, error) {
	t := e.Type()
	return &SetReturningFunctionTable{
		Name: name,
		sch: []*sql.Column{{
			Name: name,
			Type: t,
		}},
		function: e,
	}, nil
}

// Resolved implements the ExecSourceRel interface.
func (srf *SetReturningFunctionTable) Resolved() bool {
	return true
}

// String implements the ExecSourceRel interface.
func (srf *SetReturningFunctionTable) String() string {
	// TODO
	return fmt.Sprintf("set returning function %s", srf.Name)
}

// Schema implements the ExecSourceRel interface.
func (srf *SetReturningFunctionTable) Schema() sql.Schema {
	return srf.sch
}

// Children implements the ExecSourceRel interface.
func (srf *SetReturningFunctionTable) Children() []sql.Node {
	return nil
}

// WithChildren implements the ExecSourceRel interface.
func (srf *SetReturningFunctionTable) WithChildren(children ...sql.Node) (sql.Node, error) {
	if len(children) != 0 {
		return nil, sql.ErrInvalidChildrenNumber.New(srf, len(children), 0)
	}
	return srf, nil
}

// IsReadOnly implements the ExecSourceRel interface.
func (srf *SetReturningFunctionTable) IsReadOnly() bool {
	return true
}

// RowIter implements the ExecSourceRel interface.
func (srf *SetReturningFunctionTable) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	val, err := srf.function.Eval(ctx, r)
	if err != nil {
		return nil, err
	}
	if rv, ok := val.(sql.RowIter); ok {
		return rv, nil
	} else if val == nil {
		return sql.RowsToRowIter(), nil
	} else {
		return nil, fmt.Errorf("expected row iter, found %T", val)
	}
}

// Expressions implements the Expressioner interface.
func (srf *SetReturningFunctionTable) Expressions() []sql.Expression {
	return []sql.Expression{srf.function}
}

// WithExpressions implements the Expressioner interface.
func (srf *SetReturningFunctionTable) WithExpressions(exprs ...sql.Expression) (sql.Node, error) {
	if len(exprs) != 1 {
		return nil, sql.ErrInvalidChildrenNumber.New(srf, len(exprs), 1)
	}
	np := *srf
	np.function = exprs[0]
	return &np, nil
}
