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
	"io"

	pgtypes "github.com/dolthub/doltgresql/server/types"

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

func NewSetReturningFunctionTable(name string, e sql.Expression) *SetReturningFunctionTable {
	// todo set schema
	return &SetReturningFunctionTable{
		Name:     name,
		function: e,
	}
}

func (srf *SetReturningFunctionTable) Resolved() bool {
	return true
}

func (srf *SetReturningFunctionTable) String() string {
	// TODO
	return fmt.Sprintf("set returning function %s", srf.Name)
}

func (srf *SetReturningFunctionTable) Schema() sql.Schema {
	return srf.sch
}

func (srf *SetReturningFunctionTable) Children() []sql.Node {
	return nil
}

func (srf *SetReturningFunctionTable) WithChildren(children ...sql.Node) (sql.Node, error) {
	if len(children) != 0 {
		return nil, sql.ErrInvalidChildrenNumber.New(srf, len(children), 0)
	}
	return srf, nil
}

func (srf *SetReturningFunctionTable) IsReadOnly() bool {
	return true
}

func (srf *SetReturningFunctionTable) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	val, err := srf.function.Eval(ctx, r)
	if err != nil {
		return nil, err
	}
	if val == nil {
		// TODO
		return sql.RowsToRowIter(), nil
	} else if rv, ok := val.(*pgtypes.RowValues); ok {
		srf.sch = []*sql.Column{{
			Name: rv.Type().String(),
			Type: rv.Type(),
		}}
		return NewSetRowIter(rv), nil
	}
	return sql.RowsToRowIter(), nil // TODO
}

func (srf *SetReturningFunctionTable) Expressions() []sql.Expression {
	return []sql.Expression{srf.function}
}

func (srf *SetReturningFunctionTable) WithExpressions(exprs ...sql.Expression) (sql.Node, error) {
	if len(exprs) != 1 {
		return nil, sql.ErrInvalidChildrenNumber.New(srf, len(exprs), 1)
	}
	np := *srf
	np.function = exprs[0]
	return &np, nil
}

var _ sql.RowIter = (*SetRowIter)(nil)

type SetRowIter struct {
	values *pgtypes.RowValues
	idx    int32
}

func NewSetRowIter(values *pgtypes.RowValues) *SetRowIter {
	return &SetRowIter{
		values: values,
	}
}

func (s *SetRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	if s.idx >= s.values.Count() {
		return nil, io.EOF
	}
	s.idx++

	val, err := s.values.GetRow(ctx, s.idx-1)
	if err != nil {
		return nil, err
	}
	return sql.Row{val}, nil
}

func (s *SetRowIter) Close(_ *sql.Context) error {
	return nil
}
