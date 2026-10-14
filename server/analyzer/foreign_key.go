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

package analyzer

import (
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/plan"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/casts"
	"github.com/dolthub/doltgresql/server/functions/framework"
	"github.com/dolthub/doltgresql/server/types"
)

// validateForeignKeyDefinition validates that the given foreign key definition is valid for creation
func validateForeignKeyDefinition(ctx *sql.Context, fkDef sql.ForeignKeyConstraint, cols map[string]*sql.Column, parentCols map[string]*sql.Column) error {
	var castsColl *casts.Collection
	if len(fkDef.Columns) > 0 {
		var err error
		// TODO: which database is this supposed to use?
		castsColl, err = core.GetCastsCollectionFromContext(ctx, "")
		if err != nil {
			return err
		}
	}
	for i := range fkDef.Columns {
		col := cols[strings.ToLower(fkDef.Columns[i])]
		parentCol := parentCols[strings.ToLower(fkDef.ParentColumns[i])]
		if !foreignKeyComparableTypes(ctx, castsColl, col.Type, parentCol.Type) {
			return errors.Errorf("Key columns %q and %q are of incompatible types: %s and %s", col.Name, parentCol.Name, col.Type.String(), parentCol.Type.String())
		}
	}
	return nil
}

// foreignKeyComparableTypes returns whether the two given types are able to be used as parent/child columns in a
// foreign key.
func foreignKeyComparableTypes(ctx *sql.Context, castColl *casts.Collection, from sql.Type, to sql.Type) bool {
	dtFrom, ok := from.(*types.DoltgresType)
	if !ok {
		return false // should never be possible
	}

	dtTo, ok := to.(*types.DoltgresType)
	if !ok {
		return false // should never be possible
	}

	if dtFrom.Equals(dtTo) {
		return true
	}

	fromLiteral := expression.NewLiteral(dtFrom.Zero(), from)
	toLiteral := expression.NewLiteral(dtTo.Zero(), to)

	// a foreign key between two different types is valid if there is an equality operator on the two types
	// TODO: there are some subtleties in postgres not captured by this logic, e.g. a foreign key from double -> int
	//  is valid, but the reverse is not. This works fine, but is more permissive than postgres is.
	eq := framework.GetBinaryFunction(framework.Operator_BinaryEqual).Compile(ctx, "=", fromLiteral, toLiteral)
	if eq == nil || eq.StashedError() != nil {
		return false
	}

	// Additionally, we need to be able to convert freely between the two types in both directions, since we do this
	// during the process of enforcing the constraints
	forwardConversion, fErr := castColl.GetAssignmentCast(ctx, dtFrom, dtTo)
	reverseConversion, rErr := castColl.GetAssignmentCast(ctx, dtTo, dtFrom)

	return fErr == nil && rErr == nil && forwardConversion.ID.IsValid() && reverseConversion.ID.IsValid()
}

// resolveImplicitForeignKey fills omitted referenced columns in primary-key
// order. A self-reference during CREATE TABLE uses the schema being created.
func resolveImplicitForeignKey(ctx *sql.Context, fk *sql.ForeignKeyConstraint, creating *plan.CreateTable) (bool, error) {
	if len(fk.ParentColumns) != 0 {
		return false, nil
	}
	var creatingSchema string
	if creating != nil && fk.ParentTable == creating.Name() && fk.ParentDatabase == fk.Database {
		var err error
		creatingSchema, err = core.GetSchemaName(ctx, creating.Db, "")
		if err != nil {
			return false, err
		}
	}
	schemas, err := schemasOnSearchPath(ctx, fk.ParentDatabase, fk.ParentSchema)
	if err != nil {
		return false, err
	}
	var pk sql.PrimaryKeySchema
	found := false
	for _, schema := range schemas {
		// Treat the new table as visible in its own schema, without
		// overriding an earlier search-path match in another schema.
		if creatingSchema != "" && schema.SchemaName() == creatingSchema {
			pk = creating.PkSchema()
		} else {
			table, exists, err := schema.GetTableInsensitive(ctx, fk.ParentTable)
			if err != nil {
				return false, err
			}
			if !exists {
				continue
			}
			if pkTable, ok := plan.NewResolvedTable(table, schema, nil).UnderlyingTable().(sql.PrimaryKeyTable); ok {
				pk = pkTable.PrimaryKeySchema(ctx)
			}
		}
		fk.ParentSchema = schema.SchemaName()
		found = true
		break
	}
	if !found {
		return false, pgerror.Newf(pgcode.UndefinedTable, "relation %q does not exist", fk.ParentTable)
	}
	if len(pk.PkOrdinals) == 0 {
		return false, pgerror.Newf(pgcode.UndefinedObject, "there is no primary key for referenced table %q", fk.ParentTable)
	}
	if len(fk.Columns) != len(pk.PkOrdinals) {
		return false, pgerror.New(pgcode.InvalidForeignKey, "number of referencing and referenced columns for foreign key disagree")
	}
	fk.ParentColumns = make([]string, len(pk.PkOrdinals))
	for i, ordinal := range pk.PkOrdinals {
		fk.ParentColumns[i] = pk.Schema[ordinal].Name
	}
	return true, nil
}
