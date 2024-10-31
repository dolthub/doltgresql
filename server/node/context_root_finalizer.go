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
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/rowexec"

	"github.com/dolthub/doltgresql/core"
)

// ContextRootFinalizer is a node that finalizes any changes persisted within the context.
type ContextRootFinalizer struct {
	child sql.Node
}

var _ sql.ExecSourceRel = (*ContextRootFinalizer)(nil)

// NewContextRootFinalizer returns a new *ContextRootFinalizer.
func NewContextRootFinalizer(child sql.Node) *ContextRootFinalizer {
	return &ContextRootFinalizer{
		child: child,
	}
}

// CheckPrivileges implements the interface sql.ExecSourceRel.
func (rf *ContextRootFinalizer) CheckPrivileges(ctx *sql.Context, opChecker sql.PrivilegedOperationChecker) bool {
	return rf.child.CheckPrivileges(ctx, opChecker)
}

// Child returns the single child of this node
func (rf *ContextRootFinalizer) Child() sql.Node {
	return rf.child
}

// Children implements the interface sql.ExecSourceRel.
func (rf *ContextRootFinalizer) Children() []sql.Node {
	return []sql.Node{rf.child}
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (rf *ContextRootFinalizer) IsReadOnly() bool {
	return false
}

// Resolved implements the interface sql.ExecSourceRel.
func (rf *ContextRootFinalizer) Resolved() bool {
	return rf.child.Resolved()
}

// RowIter implements the interface sql.ExecSourceRel.
func (rf *ContextRootFinalizer) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	childIter, err := rowexec.DefaultBuilder.Build(ctx, rf.child, r)
	if err != nil {
		return nil, err
	}
	if childIter == nil {
		childIter = sql.RowsToRowIter()
	}
	childIter, _, err = rowexec.AddAccumulatorIter(ctx, childIter)
	if err != nil {
		return nil, err
	}
	return &rootFinalizerIter{childIter: childIter}, nil
}

// Schema implements the interface sql.ExecSourceRel.
func (rf *ContextRootFinalizer) Schema() sql.Schema {
	return rf.child.Schema()
}

// String implements the interface sql.ExecSourceRel.
func (rf *ContextRootFinalizer) String() string {
	return rf.child.String()
}

func (rf *ContextRootFinalizer) DebugString() string {
	return sql.DebugString(rf.child)
}

// WithChildren implements the interface sql.ExecSourceRel.
func (rf *ContextRootFinalizer) WithChildren(children ...sql.Node) (sql.Node, error) {
	if len(children) != 1 {
		return nil, sql.ErrInvalidChildrenNumber.New(rf, len(children), 1)
	}
	return NewContextRootFinalizer(children[0]), nil
}

// rootFinalizerIter is the iterator for *ContextRootFinalizer that finalizes the context.
type rootFinalizerIter struct {
	childIter sql.RowIter
}

var _ sql.RowIter = (*rootFinalizerIter)(nil)

// Next implements the interface sql.RowIter.
func (r *rootFinalizerIter) Next(ctx *sql.Context) (sql.Row, error) {
	return r.childIter.Next(ctx)
}

// Close implements the interface sql.RowIter.
func (r *rootFinalizerIter) Close(ctx *sql.Context) error {
	err := r.childIter.Close(ctx)
	if err != nil {
		_ = core.CloseContextRootFinalizer(ctx)
		return err
	}
	return core.CloseContextRootFinalizer(ctx)
}
