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
	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"
)

// DropExtension implements DROP EXTENSION.
type DropExtension struct {
	Names    []string
	IfExists bool
	Cascade  bool
}

var _ sql.ExecSourceRel = (*DropExtension)(nil)
var _ vitess.Injectable = (*DropExtension)(nil)

// NewDropExtension returns a new *DropExtension.
func NewDropExtension(names []string, ifExists bool, cascade bool) *DropExtension {
	return &DropExtension{
		Names:    names,
		IfExists: ifExists,
		Cascade:  cascade,
	}
}

// Children implements the interface sql.ExecSourceRel.
func (c *DropExtension) Children() []sql.Node {
	return nil
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (c *DropExtension) IsReadOnly() bool {
	return false
}

// Resolved implements the interface sql.ExecSourceRel.
func (c *DropExtension) Resolved() bool {
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (c *DropExtension) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	// TODO: implement this
	return nil, errors.Errorf("DROP EXTENSION is not yet implemented")
}

// Schema implements the interface sql.ExecSourceRel.
func (c *DropExtension) Schema() sql.Schema {
	return nil
}

// String implements the interface sql.ExecSourceRel.
func (c *DropExtension) String() string {
	return "DROP EXTENSION"
}

// WithChildren implements the interface sql.ExecSourceRel.
func (c *DropExtension) WithChildren(children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(c, children...)
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (c *DropExtension) WithResolvedChildren(children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return c, nil
}
