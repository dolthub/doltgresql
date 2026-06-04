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

package tables

import (
	"fmt"
	"strings"

	"github.com/dolthub/dolt/go/libraries/doltcore/sqle"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/resolve"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
)

// PgDatabase wraps a sqle.Database to add PostgreSQL-specific behavior.
type PgDatabase struct {
	sqle.Database
}

var _ sql.DatabaseSchema = &PgDatabase{}
var _ sql.SchemaDatabase = &PgDatabase{}
var _ sql.SchemaObjectNameValidator = &PgDatabase{}
var _ sql.IndexNameGenerator = &PgDatabase{}

// PgReadOnlyDatabase is the read-only variant of PgDatabase, used for revision databases
// such as "postgres/main" returned by sqle.DoltDatabaseProvider for detached-HEAD sessions.
// It applies the same schema-wrapping logic as PgDatabase.
type PgReadOnlyDatabase struct {
	sqle.ReadOnlyDatabase
}

var _ sql.DatabaseSchema = &PgReadOnlyDatabase{}
var _ sql.SchemaDatabase = &PgReadOnlyDatabase{}

// WrapSqleDatabase creates a PgDatabase from a sqle.Database.
func WrapSqleDatabase(db sqle.Database) *PgDatabase {
	return &PgDatabase{db}
}

// WrapSqlDatabase wraps any Dolt database variant as a Pg-aware database.
// ReadOnlyDatabase embeds sqle.Database by value, so a plain sqle.Database type
// assertion does not match it; this function handles both cases.
func WrapSqlDatabase(db sql.Database) sql.Database {
	if rodb, ok := db.(sqle.ReadOnlyDatabase); ok {
		return &PgReadOnlyDatabase{rodb}
	}
	if sdb, ok := db.(sqle.Database); ok {
		return WrapSqleDatabase(sdb)
	}
	return db
}

// applySchemaWrap wraps a single schema returned by the underlying sqle.Database methods.
// System schemas (those with registered virtual-table handlers) get a Database wrapper
// that exposes only virtual tables; all others get a PgDatabase wrapper.
func applySchemaWrap(requestedName string, schema sql.DatabaseSchema) sql.DatabaseSchema {
	sdb, ok := schema.(sqle.Database)
	if !ok {
		// information_schema and any other non-sqle schema: leave as-is.
		return schema
	}
	if _, isSystem := handlers[requestedName]; isSystem {
		return Database{sdb}
	}
	return &PgDatabase{sdb}
}

// AllSchemas overrides sqle.Database.AllSchemas to apply Doltgres schema wrapping.
func (d *PgDatabase) AllSchemas(ctx *sql.Context) ([]sql.DatabaseSchema, error) {
	schemas, err := d.Database.AllSchemas(ctx)
	if err != nil {
		return nil, err
	}
	for i, s := range schemas {
		schemas[i] = applySchemaWrap(s.SchemaName(), s)
	}
	return schemas, nil
}

// GetSchema overrides sqle.Database.GetSchema to apply Doltgres schema wrapping.
func (d *PgDatabase) GetSchema(ctx *sql.Context, schemaName string) (sql.DatabaseSchema, bool, error) {
	schema, ok, err := d.Database.GetSchema(ctx, schemaName)
	if !ok || err != nil {
		return schema, ok, err
	}
	return applySchemaWrap(schemaName, schema), true, nil
}

// GetTableInsensitive overrides sqle.Database.GetTableInsensitive to check the pg_catalog
// virtual schema before falling back to user tables.
func (d *PgDatabase) GetTableInsensitive(ctx *sql.Context, tblName string) (sql.Table, bool, error) {
	if resolve.UseSearchPath && d.Database.Schema() == "" && strings.HasPrefix(strings.ToLower(tblName), "pg_") {
		sdb, found, err := d.GetSchema(ctx, "pg_catalog")
		if err != nil {
			return nil, false, err
		}
		if found {
			tbl, foundTbl, err := sdb.GetTableInsensitive(ctx, tblName)
			if err != nil {
				return nil, false, err
			}
			if foundTbl {
				return tbl, true, nil
			}
		}
	}
	return d.Database.GetTableInsensitive(ctx, tblName)
}

// DropTable overrides sqle.Database.DropTable to prevent dropping virtual pg_catalog tables.
func (d *PgDatabase) DropTable(ctx *sql.Context, tableName string) error {
	if resolve.UseSearchPath && d.Database.Schema() == "" && strings.HasPrefix(strings.ToLower(tableName), "pg_") {
		sdb, found, err := d.GetSchema(ctx, "pg_catalog")
		if err != nil {
			return err
		}
		if found {
			_, foundTbl, err := sdb.GetTableInsensitive(ctx, tableName)
			if err != nil {
				return err
			}
			if foundTbl {
				return sql.ErrDropTableNotSupported.New("pg_catalog")
			}
		}
	}
	return d.Database.DropTable(ctx, tableName)
}

// AllSchemas overrides sqle.ReadOnlyDatabase.AllSchemas to apply Doltgres schema wrapping.
func (d *PgReadOnlyDatabase) AllSchemas(ctx *sql.Context) ([]sql.DatabaseSchema, error) {
	schemas, err := d.ReadOnlyDatabase.AllSchemas(ctx)
	if err != nil {
		return nil, err
	}
	for i, s := range schemas {
		schemas[i] = applySchemaWrap(s.SchemaName(), s)
	}
	return schemas, nil
}

// GetSchema overrides sqle.ReadOnlyDatabase.GetSchema to apply Doltgres schema wrapping.
func (d *PgReadOnlyDatabase) GetSchema(ctx *sql.Context, schemaName string) (sql.DatabaseSchema, bool, error) {
	schema, ok, err := d.ReadOnlyDatabase.GetSchema(ctx, schemaName)
	if !ok || err != nil {
		return schema, ok, err
	}
	return applySchemaWrap(schemaName, schema), true, nil
}

// GetTableInsensitive overrides sqle.Database.GetTableInsensitive to check the pg_catalog
// virtual schema before falling back to user tables.
func (d *PgReadOnlyDatabase) GetTableInsensitive(ctx *sql.Context, tblName string) (sql.Table, bool, error) {
	if resolve.UseSearchPath && d.ReadOnlyDatabase.Schema() == "" && strings.HasPrefix(strings.ToLower(tblName), "pg_") {
		sdb, found, err := d.GetSchema(ctx, "pg_catalog")
		if err != nil {
			return nil, false, err
		}
		if found {
			tbl, foundTbl, err := sdb.GetTableInsensitive(ctx, tblName)
			if err != nil {
				return nil, false, err
			}
			if foundTbl {
				return tbl, true, nil
			}
		}
	}
	return d.ReadOnlyDatabase.GetTableInsensitive(ctx, tblName)
}

// ValidateNewIndexName implements the sql.SchemaObjectNameValidator interface
func (d *PgDatabase) ValidateNewIndexName(ctx *sql.Context, newIndexName string, skipIfExists bool) (nameAlreadyUsed bool, err error) {
	nameAlreadyUsed, _, err = d.doesRelationExist(ctx, newIndexName)
	if err != nil {
		return false, err
	}

	if !nameAlreadyUsed {
		return false, nil
	}

	if skipIfExists {
		return true, nil
	}

	return nameAlreadyUsed, fmt.Errorf(`relation "%s" already exists`, newIndexName)
}

// ValidateNewSequenceName implements the sql.SchemaObjectNameValidator interface
func (d *PgDatabase) ValidateNewSequenceName(ctx *sql.Context, newSequenceName string, skipIfExists bool) (nameAlreadyUsed bool, err error) {
	nameAlreadyUsed, _, err = d.doesRelationExist(ctx, newSequenceName)
	if err != nil {
		return false, err
	}

	if !nameAlreadyUsed {
		return false, nil
	}

	if skipIfExists {
		return true, nil
	}

	return nameAlreadyUsed, fmt.Errorf(`relation "%s" already exists`, newSequenceName)
}

// ValidateNewViewName implements the sql.SchemaObjectNameValidator interface
func (d *PgDatabase) ValidateNewViewName(ctx *sql.Context, newViewName string, replaceAllowed bool) error {
	relationExists, relationType, err := d.doesRelationExist(ctx, newViewName)
	if err != nil {
		return err
	}

	if !relationExists {
		return nil
	}

	// When REPLACE is used, Postgres sends a different error message
	if replaceAllowed {
		if relationType == "view" {
			return nil
		} else {
			return fmt.Errorf(`"%s" is not a view`, newViewName)
		}
	}

	return fmt.Errorf(`relation "%s" already exists`, newViewName)
}

// ValidateNewTableName implements the sql.SchemaObjectNameValidator interface
func (d *PgDatabase) ValidateNewTableName(ctx *sql.Context, newTableName string, skipIfExists bool) (nameAlreadyUsed bool, err error) {
	relationExists, _, err := d.doesRelationExist(ctx, newTableName)
	if err != nil {
		return false, err
	}

	if !relationExists {
		return false, nil
	}

	if skipIfExists {
		return true, nil
	}

	return true, fmt.Errorf(`relation "%s" already exists`, newTableName)
}

// GenerateIndexName implements the sql.IndexNameGenerator interface with PostgreSQL-compatible naming conventions:
//   - UNIQUE indexes:   <table>_<col1>[_col2...]_key
//   - All other indexes: <table>_<col1>[_col2...]_idx
//
// Collisions are resolved by appending a numeric suffix (1, 2, …) to the base name.
// The collision check uses doesRelationExist so that any schema-level relation — table,
// view, sequence, or another index — blocks a candidate name, matching PostgreSQL's
// behavior where all relations in a schema share one namespace.
func (d *PgDatabase) GenerateIndexName(ctx *sql.Context, tableName string, idxDef sql.IndexDef, _ sql.Table) (string, error) {
	colPart := strings.Join(idxDef.ColumnNames(), "_")
	suffix := "_idx"
	if idxDef.IsUnique() {
		suffix = "_key"
	}
	base := tableName + "_" + colPart + suffix

	exists, _, err := d.doesRelationExist(ctx, base)
	if err != nil {
		return "", err
	}
	if !exists {
		return base, nil
	}
	for i := 1; ; i++ {
		candidate := fmt.Sprintf("%s%d", base, i)
		exists, _, err = d.doesRelationExist(ctx, candidate)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}
}

// doesRelationExist tests if a relation with the specified |name| exists in this database. If any relation with that
// name exists, this function returns true for |exists|, the relation type (e.g. index, view, table, sequence) for
// |relationType|. If any problems are encountered looking up a relation, an error is returned in |err|.
func (d *PgDatabase) doesRelationExist(ctx *sql.Context, name string) (exists bool, relationType string, err error) {
	lowerName := strings.ToLower(name)

	// Resolve the effective schema: when the database was obtained without a schema qualifier
	// (e.g. from GMS's plan builder for CREATE INDEX), schemaName is "" and we must fall back to
	// the session's current schema so that sequence/view checks use the right namespace.
	schema := d.Database.Schema()
	if schema == "" {
		var err error
		schema, err = core.GetCurrentSchema(ctx)
		if err != nil || schema == "" {
			schema = "public"
		}
	}

	// Tables: use GetTableNames which reads the session's working root directly.
	tableNames, err := d.Database.GetTableNames(ctx)
	if err != nil {
		return false, "", err
	}
	for _, tableName := range tableNames {
		if strings.ToLower(tableName) == lowerName {
			return true, "table", nil
		}
	}

	// Sequences: use the session-cached collection so uncommitted sequences are visible.
	seqCollection, err := core.GetSequencesCollectionFromContext(ctx, d.Database.Name())
	if err != nil {
		return false, "", err
	}
	if seqCollection.HasSequence(ctx, id.NewSequence(schema, name)) {
		return true, "sequence", nil
	}

	// Views: sqle.Database implements sql.ViewDatabase, so call AllViews directly.
	views, err := d.Database.AllViews(ctx)
	if err != nil {
		return false, "", err
	}
	for _, view := range views {
		if strings.ToLower(view.Name) == lowerName {
			return true, "view", nil
		}
	}

	// Indexes are per-table; reuse the tableNames slice from the table check above.
	for _, tableName := range tableNames {
		tbl, ok, err := d.Database.GetTableInsensitive(ctx, tableName)
		if err != nil {
			return false, "", err
		}
		if !ok {
			continue
		}
		idxTbl, ok := tbl.(sql.IndexAddressableTable)
		if !ok {
			continue
		}
		indexes, err := idxTbl.GetIndexes(ctx)
		if err != nil {
			return false, "", err
		}
		for _, idx := range indexes {
			if strings.ToLower(idx.ID()) == lowerName {
				return true, "index", nil
			}
		}
	}

	return false, "", nil
}
