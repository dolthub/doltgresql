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
)

// DeclareCursor creates a cursor that reads the rows of its query as they are fetched. The connection handler plans the
// query and stores the cursor, so the engine never executes this node.
type DeclareCursor struct {
	Name         string
	Select       vitess.SelectStatement
	IsHoldable   bool
	IsScrollable bool
	// Bindings holds the values of the query's parameters, which the extended query protocol's Bind message provides.
	Bindings map[string]vitess.Expr
}

var _ sql.ExecSourceRel = (*DeclareCursor)(nil)
var _ vitess.Injectable = (*DeclareCursor)(nil)

// Children implements the interface sql.ExecSourceRel.
func (d *DeclareCursor) Children() []sql.Node {
	return nil
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (d *DeclareCursor) IsReadOnly() bool {
	return true
}

// Resolved implements the interface sql.ExecSourceRel.
func (d *DeclareCursor) Resolved() bool {
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (d *DeclareCursor) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	return nil, errors.Errorf("DECLARE CURSOR must be handled by the connection handler")
}

// Schema implements the interface sql.ExecSourceRel.
func (d *DeclareCursor) Schema(ctx *sql.Context) sql.Schema {
	return nil
}

// String implements the interface sql.ExecSourceRel.
func (d *DeclareCursor) String() string {
	return "DECLARE CURSOR"
}

// WithChildren implements the interface sql.ExecSourceRel.
func (d *DeclareCursor) WithChildren(ctx *sql.Context, children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(d, children...)
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (d *DeclareCursor) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return d, nil
}
