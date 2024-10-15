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
	"fmt"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/server/types"
)

// DropDomain handles the DROP DOMAIN statement.
type DropDomain struct {
	schema   string
	domain   string
	ifExists bool
	cascade  bool
}

var _ sql.ExecSourceRel = (*DropDomain)(nil)
var _ vitess.Injectable = (*DropDomain)(nil)

// NewDropDomain returns a new *DropDomain.
func NewDropDomain(ifExists bool, schema string, domain string, cascade bool) *DropDomain {
	return &DropDomain{
		schema:   schema,
		domain:   domain,
		ifExists: ifExists,
		cascade:  cascade,
	}
}

// CheckPrivileges implements the interface sql.ExecSourceRel.
func (c *DropDomain) CheckPrivileges(ctx *sql.Context, opChecker sql.PrivilegedOperationChecker) bool {
	// TODO: implement privilege checking
	return true
}

// Children implements the interface sql.ExecSourceRel.
func (c *DropDomain) Children() []sql.Node {
	return nil
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (c *DropDomain) IsReadOnly() bool {
	return false
}

// Resolved implements the interface sql.ExecSourceRel.
func (c *DropDomain) Resolved() bool {
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (c *DropDomain) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	schema, err := core.GetSchemaName(ctx, nil, c.schema)
	if err != nil {
		return nil, err
	}
	collection, err := core.GetDomainsCollectionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	_, exists := collection.GetDomain(schema, c.domain)
	if !exists {
		if c.ifExists {
			// TODO: issue a notice
			return sql.RowsToRowIter(), nil
		} else {
			return nil, types.ErrTypeDoesNotExist.New(c.domain)
		}
	}

	// TODO: return nil, fmt.Errorf(`cannot drop type %s because other objects depend on it`, c.domain)

	if c.cascade {
		// TODO: handle cascade
		return nil, fmt.Errorf(`cascading domain drops are not yet supported`)
	}
	if err = collection.DropDomain(schema, c.domain); err != nil {
		return nil, err
	}
	return sql.RowsToRowIter(), nil
}

// Schema implements the interface sql.ExecSourceRel.
func (c *DropDomain) Schema() sql.Schema {
	return nil
}

// String implements the interface sql.ExecSourceRel.
func (c *DropDomain) String() string {
	return "DROP DOMAIN"
}

// WithChildren implements the interface sql.ExecSourceRel.
func (c *DropDomain) WithChildren(children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(c, children...)
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (c *DropDomain) WithResolvedChildren(children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return c, nil
}
