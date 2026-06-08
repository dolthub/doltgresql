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

// DropCast implements DROP CAST.
type DropCast struct {
	Source   *pgtypes.DoltgresType
	Target   *pgtypes.DoltgresType
	IfExists bool
}

var _ sql.ExecSourceRel = (*DropCast)(nil)
var _ vitess.Injectable = (*DropCast)(nil)

// NewDropCast returns a new *DropCast.
func NewDropCast(source *pgtypes.DoltgresType, target *pgtypes.DoltgresType, ifExists bool) *DropCast {
	return &DropCast{
		Source:   source,
		Target:   target,
		IfExists: ifExists,
	}
}

// Children implements the interface sql.ExecSourceRel.
func (c *DropCast) Children() []sql.Node {
	return nil
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (c *DropCast) IsReadOnly() bool {
	return false
}

// Resolved implements the interface sql.ExecSourceRel.
func (c *DropCast) Resolved() bool {
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (c *DropCast) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	castCollection, err := core.GetCastsCollectionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	castID := id.NewCast(c.Source.ID, c.Target.ID)
	if !castCollection.HasCast(ctx, castID) {
		if c.IfExists {
			// TODO: send notice "cast from type SOURCE_NAME to type TARGET_NAME does not exist, skipping"
			return sql.RowsToRowIter(), nil
		}
		return nil, errors.Errorf("cast from type %s to type %s does not exist", c.Source.Name(), c.Target.Name())
	}
	if err = castCollection.DropCast(ctx, castID); err != nil {
		return nil, err
	}
	return sql.RowsToRowIter(), nil
}

// Schema implements the interface sql.ExecSourceRel.
func (c *DropCast) Schema(ctx *sql.Context) sql.Schema {
	return nil
}

// String implements the interface sql.ExecSourceRel.
func (c *DropCast) String() string {
	return "DROP CAST"
}

// WithChildren implements the interface sql.ExecSourceRel.
func (c *DropCast) WithChildren(ctx *sql.Context, children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(c, children...)
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (c *DropCast) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return c, nil
}
