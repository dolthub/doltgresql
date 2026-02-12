// Copyright 2024 Dolthub, Inc.
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
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/types"
)

// RowsFrom represents a ROWS FROM table function that executes multiple
// set-returning functions in parallel and zips their results together.
// This is the PostgreSQL-compatible syntax: ROWS FROM(func1(...), func2(...), ...)
type RowsFrom struct {
	// Functions contains the set-returning function expressions to execute
	Functions []sql.Expression
	// withOrdinality when true, adds an ordinality column to the result
	withOrdinality bool
	// alias is the table alias for this ROWS FROM expression
	alias string
	// columnAliases are optional column names for the result columns
	columnAliases []string
	// colset tracks the column IDs for this node
	colset sql.ColSet
	// id is the table ID for this node
	id sql.TableId
}

var _ sql.Node = (*RowsFrom)(nil)
var _ sql.Expressioner = (*RowsFrom)(nil)
var _ sql.CollationCoercible = (*RowsFrom)(nil)
var _ plan.TableIdNode = (*RowsFrom)(nil)
var _ sql.RenameableNode = (*RowsFrom)(nil)
var _ sql.ExecBuilderNode = (*RowsFrom)(nil)

// NewRowsFrom creates a new RowsFrom node with the given function expressions.
func NewRowsFrom(exprs []sql.Expression, alias string, withOrdinality bool, columnAliases []string) *RowsFrom {
	return &RowsFrom{
		Functions:      exprs,
		withOrdinality: withOrdinality,
		alias:          alias,
		columnAliases:  columnAliases,
	}
}

// BuildRowIter implements sql.ExecBuilderNode.
func (r *RowsFrom) BuildRowIter(ctx *sql.Context, b sql.NodeExecBuilder, row sql.Row) (sql.RowIter, error) {
	return NewRowsFromIter(r.Functions, r.withOrdinality, row), nil
}

// WithId implements plan.TableIdNode
func (r *RowsFrom) WithId(id sql.TableId) plan.TableIdNode {
	ret := *r
	ret.id = id
	return &ret
}

// Id implements plan.TableIdNode
func (r *RowsFrom) Id() sql.TableId {
	return r.id
}

// WithColumns implements plan.TableIdNode
func (r *RowsFrom) WithColumns(set sql.ColSet) plan.TableIdNode {
	ret := *r
	ret.colset = set
	return &ret
}

// Columns implements plan.TableIdNode
func (r *RowsFrom) Columns() sql.ColSet {
	return r.colset
}

// Name returns the alias name for this ROWS FROM expression
func (r *RowsFrom) Name() string {
	if r.alias != "" {
		return r.alias
	}
	return "rows_from"
}

// WithName implements sql.RenameableNode
func (r *RowsFrom) WithName(s string) sql.Node {
	ret := *r
	ret.alias = s
	return &ret
}

// Schema implements the sql.Node interface.
func (r *RowsFrom) Schema() sql.Schema {
	var schema sql.Schema

	for i, f := range r.Functions {
		colName := fmt.Sprintf("col%d", i)
		if i < len(r.columnAliases) && r.columnAliases[i] != "" {
			colName = r.columnAliases[i]
		} else if nameable, ok := f.(sql.Nameable); ok {
			colName = nameable.Name()
		}

		schema = append(schema, &sql.Column{
			Name:     colName,
			Type:     f.Type(),
			Nullable: true, // SRF results can be NULL when zipping unequal-length results
			Source:   r.Name(),
		})
	}

	if r.withOrdinality {
		schema = append(schema, &sql.Column{
			Name:     "ordinality",
			Type:     types.Int64,
			Nullable: false,
			Source:   r.Name(),
		})
	}

	return schema
}

// Children implements the sql.Node interface.
func (r *RowsFrom) Children() []sql.Node {
	return nil
}

// Resolved implements the sql.Resolvable interface.
func (r *RowsFrom) Resolved() bool {
	for _, f := range r.Functions {
		if !f.Resolved() {
			return false
		}
	}
	return true
}

// IsReadOnly implements the sql.Node interface.
func (r *RowsFrom) IsReadOnly() bool {
	return true
}

// String implements the sql.Node interface.
func (r *RowsFrom) String() string {
	var sb strings.Builder
	sb.WriteString("ROWS FROM(")
	for i, f := range r.Functions {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(f.String())
	}
	sb.WriteString(")")
	if r.withOrdinality {
		sb.WriteString(" WITH ORDINALITY")
	}
	if r.alias != "" {
		sb.WriteString(" AS ")
		sb.WriteString(r.alias)
	}
	return sb.String()
}

// DebugString implements the sql.DebugStringer interface.
func (r *RowsFrom) DebugString() string {
	var sb strings.Builder
	sb.WriteString("RowsFrom(")
	for i, f := range r.Functions {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(sql.DebugString(f))
	}
	sb.WriteString(")")
	if r.withOrdinality {
		sb.WriteString(" WITH ORDINALITY")
	}
	if r.alias != "" {
		sb.WriteString(" AS ")
		sb.WriteString(r.alias)
	}
	return sb.String()
}

// Expressions implements the sql.Expressioner interface.
func (r *RowsFrom) Expressions() []sql.Expression {
	return r.Functions
}

// WithExpressions implements the sql.Expressioner interface.
func (r *RowsFrom) WithExpressions(exprs ...sql.Expression) (sql.Node, error) {
	if len(exprs) != len(r.Functions) {
		return nil, sql.ErrInvalidChildrenNumber.New(r, len(exprs), len(r.Functions))
	}
	ret := *r
	ret.Functions = exprs
	return &ret, nil
}

// WithChildren implements the sql.Node interface.
func (r *RowsFrom) WithChildren(children ...sql.Node) (sql.Node, error) {
	if len(children) != 0 {
		return nil, sql.ErrInvalidChildrenNumber.New(r, len(children), 0)
	}
	return r, nil
}

// CollationCoercibility implements the interface sql.CollationCoercible.
func (*RowsFrom) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	return sql.Collation_binary, 7
}

// RowsFromIter is an iterator for the RowsFrom node.
// It executes multiple set-returning functions in parallel and zips their results together.
// When one function is exhausted before another, NULL is used for its values.
type RowsFromIter struct {
	functions      []sql.Expression
	iters          []sql.RowIter
	finished       []bool
	withOrdinality bool
	ordinality     int64
	initialized    bool
	sourceRow      sql.Row
}

var _ sql.RowIter = (*RowsFromIter)(nil)

// NewRowsFromIter creates a new RowsFromIter.
func NewRowsFromIter(functions []sql.Expression, withOrdinality bool, row sql.Row) *RowsFromIter {
	return &RowsFromIter{
		functions:      functions,
		withOrdinality: withOrdinality,
		sourceRow:      row,
		finished:       make([]bool, len(functions)),
	}
}

// Next implements the sql.RowIter interface.
func (r *RowsFromIter) Next(ctx *sql.Context) (sql.Row, error) {
	if !r.initialized {
		if err := r.initIterators(ctx); err != nil {
			return nil, err
		}
		r.initialized = true
	}

	allFinished := true
	for _, f := range r.finished {
		if !f {
			allFinished = false
			break
		}
	}
	if allFinished {
		return nil, io.EOF
	}

	row := make(sql.Row, len(r.functions))
	for i, iter := range r.iters {
		if r.finished[i] {
			row[i] = nil
			continue
		}

		nextRow, err := iter.Next(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				r.finished[i] = true
				row[i] = nil
				continue
			}
			return nil, err
		}

		if len(nextRow) > 0 {
			row[i] = nextRow[0]
		} else {
			row[i] = nil
		}
	}

	allFinished = true
	for _, f := range r.finished {
		if !f {
			allFinished = false
			break
		}
	}

	allNulls := true
	for _, v := range row {
		if v != nil {
			allNulls = false
			break
		}
	}

	if allFinished && allNulls {
		return nil, io.EOF
	}

	r.ordinality++
	if r.withOrdinality {
		row = append(row, r.ordinality)
	}

	return row, nil
}

func (r *RowsFromIter) initIterators(ctx *sql.Context) error {
	r.iters = make([]sql.RowIter, len(r.functions))

	for i, f := range r.functions {
		if rie, ok := f.(sql.RowIterExpression); ok && rie.ReturnsRowIter() {
			iter, err := rie.EvalRowIter(ctx, r.sourceRow)
			if err != nil {
				for j := 0; j < i; j++ {
					if r.iters[j] != nil {
						r.iters[j].Close(ctx)
					}
				}
				return err
			}
			r.iters[i] = iter
		} else {
			val, err := f.Eval(ctx, r.sourceRow)
			if err != nil {
				for j := 0; j < i; j++ {
					if r.iters[j] != nil {
						r.iters[j].Close(ctx)
					}
				}
				return err
			}
			r.iters[i] = &singleValueIter{value: val}
		}
	}

	return nil
}

// Close implements the sql.RowIter interface.
func (r *RowsFromIter) Close(ctx *sql.Context) error {
	var firstErr error
	for _, iter := range r.iters {
		if iter != nil {
			if err := iter.Close(ctx); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

type singleValueIter struct {
	value    interface{}
	consumed bool
}

func (s *singleValueIter) Next(ctx *sql.Context) (sql.Row, error) {
	if s.consumed {
		return nil, io.EOF
	}
	s.consumed = true
	return sql.Row{s.value}, nil
}

func (s *singleValueIter) Close(ctx *sql.Context) error {
	return nil
}
