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
	"strings"

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/jackc/pgx/v5/pgproto3"

	"github.com/dolthub/doltgresql/server/auth"
)

// DropDatabase handles the DROP DATABASE statement.
type DropDatabase struct {
	Database string
	IfExists bool
	catalog  sql.Catalog
}

var _ sql.ExecSourceRel = (*DropDatabase)(nil)
var _ sql.MultiDatabaser = (*DropDatabase)(nil)
var _ vitess.Injectable = (*DropDatabase)(nil)

// NewDropDatabase returns a new *DropDatabase.
func NewDropDatabase(database string, ifExists bool) *DropDatabase {
	return &DropDatabase{
		Database: database,
		IfExists: ifExists,
	}
}

// Children implements the interface sql.ExecSourceRel.
func (c *DropDatabase) Children() []sql.Node {
	return nil
}

// DatabaseProvider implements the interface sql.MultiDatabaser.
func (c *DropDatabase) DatabaseProvider() sql.DatabaseProvider {
	return c.catalog
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (c *DropDatabase) IsReadOnly() bool {
	return false
}

// Resolved implements the interface sql.ExecSourceRel.
func (c *DropDatabase) Resolved() bool {
	return c.catalog != nil
}

// RowIter implements the interface sql.ExecSourceRel.
func (c *DropDatabase) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	if !c.catalog.HasDatabase(ctx, c.Database) {
		if c.IfExists {
			dsess.DSessFromSess(ctx.Session).Notice(&pgproto3.NoticeResponse{
				Severity: "NOTICE",
				Message:  fmt.Sprintf(`database "%s" does not exist, skipping`, c.Database),
			})
			return sql.RowsToRowIter(), nil
		}
		return nil, sql.ErrDatabaseNotFound.New(c.Database)
	}
	if err := c.catalog.RemoveDatabase(ctx, c.Database); err != nil {
		// Another session may have dropped the database between our existence check and the removal
		if c.IfExists && sql.ErrDatabaseNotFound.Is(err) {
			return sql.RowsToRowIter(), nil
		}
		return nil, err
	}
	if strings.EqualFold(ctx.GetCurrentDatabase(), c.Database) {
		ctx.SetCurrentDatabase("")
	}
	// Settings from ALTER DATABASE and ALTER ROLE ... IN DATABASE are keyed by the database's name, so they must be
	// removed to prevent them from applying to a new database with the same name
	var err error
	var rsc doltdb.ReplicationStatusController
	auth.LockWrite(func() {
		auth.RemoveDatabaseRoleSettings(c.Database)
		err = auth.PersistChanges(ctx, &rsc)
	})
	if err != nil {
		return nil, err
	}
	auth.WaitForReplication(ctx, rsc)
	return sql.RowsToRowIter(), nil
}

// Schema implements the interface sql.ExecSourceRel.
func (c *DropDatabase) Schema(ctx *sql.Context) sql.Schema {
	return nil
}

// String implements the interface sql.ExecSourceRel.
func (c *DropDatabase) String() string {
	return "DROP DATABASE"
}

// WithChildren implements the interface sql.ExecSourceRel.
func (c *DropDatabase) WithChildren(ctx *sql.Context, children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(c, children...)
}

// WithDatabaseProvider implements the interface sql.MultiDatabaser.
func (c *DropDatabase) WithDatabaseProvider(provider sql.DatabaseProvider) (sql.Node, error) {
	catalog, ok := provider.(sql.Catalog)
	if !ok {
		return nil, fmt.Errorf("DROP DATABASE expected a catalog but received `%T`", provider)
	}
	nc := *c
	nc.catalog = catalog
	return &nc, nil
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (c *DropDatabase) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return c, nil
}
