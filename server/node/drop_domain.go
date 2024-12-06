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
	"github.com/dolthub/doltgresql/server/auth"
	"github.com/dolthub/doltgresql/server/types"
)

// DropDomain handles the DROP DOMAIN statement.
type DropDomain struct {
	database string
	schema   string
	domain   string
	ifExists bool
	cascade  bool
}

var _ sql.ExecSourceRel = (*DropDomain)(nil)
var _ vitess.Injectable = (*DropDomain)(nil)

// NewDropDomain returns a new *DropDomain.
func NewDropDomain(ifExists bool, db string, schema string, domain string, cascade bool) *DropDomain {
	return &DropDomain{
		database: db,
		schema:   schema,
		domain:   domain,
		ifExists: ifExists,
		cascade:  cascade,
	}
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
	var userRole auth.Role
	auth.LockRead(func() {
		userRole = auth.GetRole(ctx.Client().User)
	})
	if !userRole.IsValid() {
		return nil, fmt.Errorf(`role "%s" does not exist`, ctx.Client().User)
	}

	currentDb := ctx.GetCurrentDatabase()
	if len(c.database) > 0 && c.database != currentDb {
		return nil, fmt.Errorf("DROP DOMAIN is currently only supported for the current database")
	}
	schema, err := core.GetSchemaName(ctx, nil, c.schema)
	if err != nil {
		return nil, err
	}
	collection, err := core.GetTypesCollectionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	domain, exists := collection.GetDomainType(schema, c.domain)
	if !exists {
		if c.ifExists {
			// TODO: issue a notice
			return sql.RowsToRowIter(), nil
		} else {
			return nil, types.ErrTypeDoesNotExist.New(c.domain)
		}
	}
	if c.cascade {
		// TODO: handle cascade
		return nil, fmt.Errorf(`cascading domain drops are not yet supported`)
	}

	// iterate on all table columns to check if this domain is currently used.
	db, err := core.GetSqlDatabaseFromContext(ctx, "")
	if err != nil {
		return nil, err
	}
	tableNames, err := db.GetTableNames(ctx)
	if err != nil {
		return nil, err
	}
	for _, tableName := range tableNames {
		t, ok, err := db.GetTableInsensitive(ctx, tableName)
		if err != nil {
			return nil, err
		}
		if ok {
			for _, col := range t.Schema() {
				if dt, isDoltgresType := col.Type.(*types.DoltgresType); isDoltgresType && dt.TypType == types.TypeType_Domain {
					if dt.Name == domain.Name {
						// TODO: issue a detail (list of all columns and tables that uses this domain)
						//  and a hint (when we support CASCADE)
						return nil, fmt.Errorf(`cannot drop type %s because other objects depend on it - column %s of table %s depends on type %s'`, c.domain, col.Name, t.Name(), c.domain)
					}
				}
			}
		}
	}

	if err = collection.DropType(schema, c.domain); err != nil {
		return nil, err
	}
	arrayType := fmt.Sprintf(`_%s`, c.domain)
	if err = collection.DropType(schema, arrayType); err != nil {
		return nil, err
	}
	auth.LockWrite(func() {
		auth.RemoveOwner(auth.OwnershipKey{
			PrivilegeObject: auth.PrivilegeObject_DOMAIN,
			Schema:          schema,
			Name:            c.domain,
		}, userRole.ID())
		err = auth.PersistChanges()
	})
	if err != nil {
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
