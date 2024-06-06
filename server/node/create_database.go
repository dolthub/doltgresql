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

	"github.com/dolthub/dolt/go/libraries/doltcore/sqle"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/rowexec"

	"github.com/dolthub/doltgresql/server/tables"
)

// CreateDatabase is a node that creates a database and handles its initial setup.
type CreateDatabase struct {
	createDB *plan.CreateDB
}

var _ sql.ExecSourceRel = (*CreateDatabase)(nil)

// NewCreateDatabase returns a new *CreateDatabase.
func NewCreateDatabase(createDB *plan.CreateDB) *CreateDatabase {
	return &CreateDatabase{
		createDB: createDB,
	}
}

// CheckPrivileges implements the interface sql.ExecSourceRel.
func (c *CreateDatabase) CheckPrivileges(ctx *sql.Context, opChecker sql.PrivilegedOperationChecker) bool {
	return c.createDB.CheckPrivileges(ctx, opChecker)
}

// Children implements the interface sql.ExecSourceRel.
func (c *CreateDatabase) Children() []sql.Node {
	return []sql.Node{c.createDB}
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (c *CreateDatabase) IsReadOnly() bool {
	return false
}

// Resolved implements the interface sql.ExecSourceRel.
func (c *CreateDatabase) Resolved() bool {
	return c.createDB != nil && c.createDB.Resolved()
}

// RowIter implements the interface sql.ExecSourceRel.
func (c *CreateDatabase) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	// First we'll create the database
	createDbIter, err := rowexec.DefaultBuilder.Build(ctx, c.createDB, r)
	if err != nil {
		return nil, err
	}
	// Next we'll create the "pg_catalog" schema, since it should exist in all databases by default
	db, err := c.createDB.Catalog.Database(ctx, c.createDB.DbName)
	if err != nil {
		return nil, err
	}
	sqlDb, ok := db.(dsess.SqlDatabase)
	if !ok {
		return nil, fmt.Errorf("`%T` is not of type `%T`", db, dsess.SqlDatabase(nil))
	}
	tx, ok := ctx.GetTransaction().(*dsess.DoltTransaction)
	if tx == nil || !ok {
		return nil, fmt.Errorf("nil transaction encountered while creating database: %s", c.createDB.DbName)
	}
	// Since the database was created after the start of the transaction, we have to add it so that the session can find
	// it, since it uses the transaction's initial root when a database's state has not yet been loaded by the session.
	if err = tx.AddDb(ctx, sqlDb); err != nil {
		return nil, err
	}
	schemaDb, ok := db.(sql.SchemaDatabase)
	if !ok {
		return nil, fmt.Errorf("`%T` is not of type `%T`", db, sql.SchemaDatabase(nil))
	}
	const pgCatalogName = "pg_catalog"
	if err = schemaDb.CreateSchema(ctx, pgCatalogName); err != nil {
		return nil, err
	}
	// Now we'll initialize all of the tables that should be in pg_catalog
	createdSchema, ok, err := schemaDb.GetSchema(ctx, pgCatalogName)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("cannot find the newly created `%s` schema", pgCatalogName)
	}
	if err = tables.InitializeTables(ctx, createdSchema.(sqle.Database), pgCatalogName); err != nil {
		return nil, err
	}
	return createDbIter, nil
}

// Schema implements the interface sql.ExecSourceRel.
func (c *CreateDatabase) Schema() sql.Schema {
	return c.createDB.Schema()
}

// String implements the interface sql.ExecSourceRel.
func (c *CreateDatabase) String() string {
	return c.createDB.String()
}

// WithChildren implements the interface sql.ExecSourceRel.
func (c *CreateDatabase) WithChildren(children ...sql.Node) (sql.Node, error) {
	if len(children) != 1 {
		return nil, sql.ErrInvalidChildrenNumber.New(c, len(children), 1)
	}
	createDb, ok := children[0].(*plan.CreateDB)
	if !ok {
		return nil, fmt.Errorf("%T: expected child to be `%T` but got `%T`", c, (*plan.CreateDB)(nil), children[0])
	}
	return &CreateDatabase{
		createDB: createDb,
	}, nil
}
