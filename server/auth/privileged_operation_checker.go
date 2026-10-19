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

package auth

import (
	"strings"

	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dtablefunctions"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core"
)

// privilegedOperationChecker answers the MySQL-style privilege checks that nodes make through
// sql.AuthorizationCheckerNode, such as Dolt's table functions (dolt_diff, dolt_patch, etc.), using Doltgres
// privileges. The global lock must be held for reading while it is used.
type privilegedOperationChecker struct {
	state AuthorizationQueryState
	// err is the first denial, which is reported in place of a generic error.
	err error
}

var _ sql.PrivilegedOperationChecker = (*privilegedOperationChecker)(nil)

// UserHasPrivileges implements the sql.PrivilegedOperationChecker interface.
func (c *privilegedOperationChecker) UserHasPrivileges(ctx *sql.Context, operations ...sql.PrivilegedOperation) bool {
	if c.state.role.IsSuperUser {
		return true
	}
	for _, operation := range operations {
		if err := c.checkOperation(ctx, operation); err != nil {
			if c.err == nil {
				c.err = err
			}
			return false
		}
	}
	return true
}

// RoutineAdminCheck implements the sql.PrivilegedOperationChecker interface. Admin-only routines are restricted to
// superusers.
func (c *privilegedOperationChecker) RoutineAdminCheck(ctx *sql.Context, operations ...sql.PrivilegedOperation) bool {
	return c.state.role.IsSuperUser
}

// checkOperation returns an error if the current role lacks the privileges that the operation requires.
func (c *privilegedOperationChecker) checkOperation(ctx *sql.Context, operation sql.PrivilegedOperation) error {
	if len(operation.DynamicPrivileges) > 0 || len(operation.Routine) > 0 {
		// Doltgres has no equivalent for these, so only superusers are allowed
		return errPermissionDenied()
	}
	if len(operation.Table) == 0 {
		// Doltgres does not have database-level SELECT, INSERT, etc. Database-level operations remain allowed, as they
		// were before table functions were checked.
		return nil
	}
	privileges := make([]Privilege, len(operation.StaticPrivileges))
	for i, privilegeType := range operation.StaticPrivileges {
		switch privilegeType {
		case sql.PrivilegeType_Select:
			privileges[i] = Privilege_SELECT
		case sql.PrivilegeType_Insert:
			privileges[i] = Privilege_INSERT
		case sql.PrivilegeType_Update:
			privileges[i] = Privilege_UPDATE
		case sql.PrivilegeType_Delete:
			privileges[i] = Privilege_DELETE
		default:
			return errPermissionDenied()
		}
	}
	schemaNames, tableName, err := operationTableSchemas(ctx, operation)
	if err != nil {
		return err
	}
	for _, schemaName := range schemaNames {
		if err = checkPrivilegeOnTable(c.state, "table", schemaName, tableName, privileges); err != nil {
			return err
		}
	}
	return nil
}

// operationTableSchemas returns the schemas whose table the operation may target, along with the table's name. A
// table qualified by its schema targets that schema. Dolt's table functions resolve unqualified names in different
// ways, and may read tables that exist only in a commit, so an unqualified name targets every schema with a table of
// that name, or the current schema if no schema has one.
func operationTableSchemas(ctx *sql.Context, operation sql.PrivilegedOperation) (schemaNames []string, tableName string, err error) {
	if schema, table, ok := strings.Cut(operation.Table, "."); ok {
		return []string{schema}, table, nil
	}
	tableName = operation.Table
	db, err := core.GetSqlDatabaseFromContext(ctx, operation.Database)
	if err != nil {
		return nil, "", err
	}
	if schemaDB, ok := db.(sql.SchemaDatabase); ok {
		schemas, err := schemaDB.AllSchemas(ctx)
		if err != nil {
			return nil, "", err
		}
		for _, schema := range schemas {
			if isSystemCatalogSchema(schema.SchemaName()) {
				continue
			}
			_, found, err := schema.GetTableInsensitive(ctx, tableName)
			if err != nil {
				return nil, "", err
			}
			if found {
				schemaNames = append(schemaNames, schema.SchemaName())
			}
		}
	}
	if len(schemaNames) == 0 {
		currentSchema, err := core.GetSchemaName(ctx, nil, "")
		if err != nil {
			return nil, "", err
		}
		schemaNames = []string{currentSchema}
	}
	return schemaNames, tableName, nil
}

// isSystemCatalogSchema returns whether the schema holds system catalogs, which are not versioned, so they never
// appear in Dolt's diffs or history.
func isSystemCatalogSchema(schemaName string) bool {
	return schemaName == "pg_catalog" || schemaName == sql.InformationSchemaDatabaseName
}

// checkAllTablesRead returns an error if the node reads every table in its database and the current role cannot read
// all of them. Dolt's table functions check every table by listing the database's tables, which only covers the
// current schema in Doltgres, so tables in the other schemas are checked here.
func (c *privilegedOperationChecker) checkAllTablesRead(ctx *sql.Context, node sql.AuthorizationCheckerNode) error {
	if c.state.role.IsSuperUser || !readsAllTables(node) {
		return nil
	}
	databaser, ok := node.(sql.Databaser)
	if !ok {
		return errPermissionDenied()
	}
	schemaDB, ok := databaser.Database().(sql.SchemaDatabase)
	if !ok {
		return nil
	}
	schemas, err := schemaDB.AllSchemas(ctx)
	if err != nil {
		return err
	}
	for _, schema := range schemas {
		schemaName := schema.SchemaName()
		if isSystemCatalogSchema(schemaName) {
			continue
		}
		tableNames, err := schema.GetTableNames(ctx)
		if err != nil {
			return err
		}
		for _, tableName := range tableNames {
			if err = checkPrivilegeOnTable(c.state, "table", schemaName, tableName, []Privilege{Privilege_SELECT}); err != nil {
				return err
			}
		}
	}
	return nil
}

// readsAllTables returns whether the node is one of Dolt's table functions that reads every table of its database,
// either always or because it was not given a table name.
func readsAllTables(node sql.AuthorizationCheckerNode) bool {
	switch node := node.(type) {
	case *dtablefunctions.LogTableFunction, *dtablefunctions.MergeBasesTableFunction,
		*dtablefunctions.PreviewMergeConflictsSummaryTableFunction:
		return true
	case *dtablefunctions.PatchTableFunction, *dtablefunctions.DiffStatTableFunction,
		*dtablefunctions.DiffSummaryTableFunction, *dtablefunctions.SchemaDiffTableFunction:
		// These take either a "from..to" argument or separate from and to arguments, followed by an optional table name
		exprs := node.(sql.Expressioner).Expressions()
		if len(exprs) == 0 {
			return true
		}
		if strings.Contains(exprs[0].String(), "..") {
			return len(exprs) < 2
		}
		return len(exprs) < 3
	default:
		return false
	}
}
