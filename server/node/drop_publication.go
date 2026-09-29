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
	"fmt"

	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/jackc/pgx/v5/pgproto3"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/auth"
)

// DropPublication drops definitions, checking every target before changing the collection.
type DropPublication struct {
	Names    []string
	IfExists bool
}

var _ sql.ExecSourceRel = (*DropPublication)(nil)
var _ vitess.Injectable = (*DropPublication)(nil)

// Children implements sql.ExecSourceRel.
func (d *DropPublication) Children() []sql.Node { return nil }

// IsReadOnly implements sql.ExecSourceRel.
func (d *DropPublication) IsReadOnly() bool { return false }

// Resolved implements sql.ExecSourceRel.
func (d *DropPublication) Resolved() bool { return true }

// Schema implements sql.ExecSourceRel.
func (d *DropPublication) Schema(ctx *sql.Context) sql.Schema { return nil }

// String implements sql.ExecSourceRel.
func (d *DropPublication) String() string { return "DROP PUBLICATION" }

// WithChildren implements sql.ExecSourceRel.
func (d *DropPublication) WithChildren(ctx *sql.Context, children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(d, children...)
}

// WithResolvedChildren implements vitess.Injectable.
func (d *DropPublication) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return d, nil
}

// RowIter implements sql.ExecSourceRel.
func (d *DropPublication) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	if tx := ctx.GetTransaction(); tx != nil && tx.IsReadOnly() {
		return nil, pgerror.New(pgcode.ReadOnlySQLTransaction, "cannot execute DROP PUBLICATION in a read-only transaction")
	}
	collection, err := core.GetPublicationsCollectionFromContext(ctx, "")
	if err != nil {
		return nil, err
	}
	var targets []id.Publication
	seen := make(map[id.Publication]bool)
	for _, name := range d.Names {
		publicationID := id.NewPublication(name)
		if seen[publicationID] {
			continue
		}
		publication, err := collection.GetPublication(ctx, publicationID)
		if err != nil {
			return nil, err
		}
		if !publication.ID.IsValid() {
			if d.IfExists {
				dsess.DSessFromSess(ctx.Session).Notice(&pgproto3.NoticeResponse{Severity: "NOTICE", Code: pgcode.SuccessfulCompletion.String(), Message: fmt.Sprintf(`publication "%s" does not exist, skipping`, name)})
				continue
			}
			return nil, pgerror.Newf(pgcode.UndefinedObject, `publication "%s" does not exist`, name)
		}
		seen[publicationID] = true
		var canDrop bool
		auth.LockRead(func() {
			role := auth.GetRole(ctx.Client().User)
			_, inherits, _ := auth.IsRoleAMember(role.ID(), auth.RoleID(publication.OwnerRoleID))
			canDrop = role.IsSuperUser || inherits
		})
		if !canDrop {
			return nil, pgerror.Newf(pgcode.InsufficientPrivilege, `must be owner of publication %s`, name)
		}
		targets = append(targets, publicationID)
	}
	if len(targets) > 0 {
		if err = collection.DropPublication(ctx, targets...); err != nil {
			return nil, err
		}
	}
	return sql.RowsToRowIter(), nil
}
