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
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/auth"
	"github.com/dolthub/doltgresql/server/types"
)

// DropType handles the DROP TYPE statement.
type DropType struct {
	database string
	schName  string
	typName  string
	ifExists bool
	cascade  bool
}

var _ sql.ExecSourceRel = (*DropType)(nil)
var _ vitess.Injectable = (*DropType)(nil)

// NewDropType returns a new *DropType.
func NewDropType(ifExists bool, db, sch, typ string, cascade bool) *DropType {
	return &DropType{
		database: db,
		schName:  sch,
		typName:  typ,
		ifExists: ifExists,
		cascade:  cascade,
	}
}

// Children implements the interface sql.ExecSourceRel.
func (c *DropType) Children() []sql.Node {
	return nil
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (c *DropType) IsReadOnly() bool {
	return false
}

// Resolved implements the interface sql.ExecSourceRel.
func (c *DropType) Resolved() bool {
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (c *DropType) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	var userRole auth.Role
	auth.LockRead(func() {
		userRole = auth.GetRole(ctx.Client().User)
	})
	if !userRole.IsValid() {
		return nil, errors.Errorf(`role "%s" does not exist`, ctx.Client().User)
	}

	currentDb := ctx.GetCurrentDatabase()
	if len(c.database) > 0 && c.database != currentDb {
		return nil, errors.Errorf("DROP TYPE is currently only supported for the current database")
	}
	schema, err := core.GetSchemaName(ctx, nil, c.schName)
	if err != nil {
		return nil, err
	}
	collection, err := core.GetTypesCollectionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	typeID := id.NewType(schema, c.typName)
	typ, err := collection.GetType(ctx, typeID)
	if err != nil {
		return nil, err
	}
	if typ == nil {
		if c.ifExists {
			// TODO: issue a notice
			return sql.RowsToRowIter(), nil
		} else {
			return nil, types.ErrTypeDoesNotExist.New(c.typName)
		}
	}
	if c.cascade {
		// TODO: handle cascade
		return nil, errors.Errorf(`cascading type drops are not yet supported`)
	}

	if _, ok := types.IDToBuiltInDoltgresType[typ.ID]; ok {
		return nil, types.ErrCannotDropSystemType.New(typ.String())
	}

	// TODO: use .IsArrayType() when we support OIDs, so Elem OID isn't 0
	if typ.TypCategory == types.TypeCategory_ArrayTypes {
		// TODO: get the base type name
		//  add HINT:  You can drop type ___ instead. (base type)
		arrTypeName := typ.String()
		return nil, types.ErrCannotDropArrayType.New(arrTypeName, strings.TrimSuffix(arrTypeName, "[]"))
	}

	// iterate on all table columns to check if this type is currently used.
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
				if dt, isDoltgresType := col.Type.(*types.DoltgresType); isDoltgresType {
					if dt.Name() == typ.Name() {
						// TODO: issue a detail (list of all columns and tables that uses this type)
						//  and a hint (when we support CASCADE)
						return nil, errors.Errorf(`cannot drop type %s because other objects depend on it - column %s of table %s depends on type %s'`, c.typName, col.Name, t.Name(), c.typName)
					}
				}
			}
		}
	}

	if err = collection.DropType(ctx, typeID); err != nil {
		return nil, err
	}

	// undefined/shell type doesn't create array type.
	if typ.IsDefined {
		arrayTypeName := fmt.Sprintf(`_%s`, c.typName)
		arrayID := id.NewType(schema, arrayTypeName)
		if err = collection.DropType(ctx, arrayID); err != nil {
			return nil, err
		}
	}

	return sql.RowsToRowIter(), nil
}

// Schema implements the interface sql.ExecSourceRel.
func (c *DropType) Schema() sql.Schema {
	return nil
}

// String implements the interface sql.ExecSourceRel.
func (c *DropType) String() string {
	return "DROP TYPE"
}

// WithChildren implements the interface sql.ExecSourceRel.
func (c *DropType) WithChildren(children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(c, children...)
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (c *DropType) WithResolvedChildren(children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return c, nil
}
