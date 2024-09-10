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
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"
)

// DiscardStatement is just a marker type, since all functionality is handled by the connection handler,
// rather than the engine. It has to conform to the sql.ExecSourceRel interface to be used in the handler, but this
// functionality is all unused.
type DiscardStatement struct{}

var _ vitess.Injectable = DiscardStatement{}
var _ sql.ExecSourceRel = DiscardStatement{}

// CheckPrivileges implements the interface sql.ExecSourceRel.
func (d DiscardStatement) CheckPrivileges(ctx *sql.Context, opChecker sql.PrivilegedOperationChecker) bool {
	return true
}

// Children implements the interface sql.ExecSourceRel.
func (d DiscardStatement) Children() []sql.Node {
	return nil
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (d DiscardStatement) IsReadOnly() bool {
	return true
}

// Resolved implements the interface sql.ExecSourceRel.
func (d DiscardStatement) Resolved() bool {
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (d DiscardStatement) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	panic("DISCARD ALL should be handled by the connection handler")
}

// Schema implements the interface sql.ExecSourceRel.
func (d DiscardStatement) Schema() sql.Schema {
	return nil
}

// String implements the interface sql.ExecSourceRel.
func (d DiscardStatement) String() string {
	return "DISCARD ALL"
}

// WithChildren implements the interface sql.ExecSourceRel.
func (d DiscardStatement) WithChildren(children ...sql.Node) (sql.Node, error) {
	return d, nil
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (d DiscardStatement) WithResolvedChildren(children []any) (any, error) {
	return d, nil
}
