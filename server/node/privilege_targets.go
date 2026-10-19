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
	"strings"

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/functions/framework"
)

// PrivilegeTarget is a table, sequence, or routine named by a GRANT or REVOKE statement.
type PrivilegeTarget struct {
	Database string // The database qualifier as written, which is empty when the target is not database-qualified
	Schema   string // The schema qualifier as written, which is empty when the target is not schema-qualified
	Name     string // The object's name, which is empty for ALL ... IN SCHEMA
	ArgTypes string // A routine's argument types for its privilege key
	// TODO: there can be 2 routines with the same name but different argument types, so ArgTypes is always empty
	//  until we can get argument types from a CALL statement
	Args string // A routine's argument types as written, which are used only in error messages
}

// String returns the target's name as written.
func (t PrivilegeTarget) String() string {
	parts := make([]string, 0, 3)
	for _, part := range []string{t.Database, t.Schema, t.Name} {
		if len(part) > 0 {
			parts = append(parts, part)
		}
	}
	return strings.Join(parts, ".")
}

// checkTargetDatabase returns an error if the target is qualified by a database other than the current one, which
// PostgreSQL does not support. A qualifier naming the current database is accepted.
func checkTargetDatabase(ctx *sql.Context, target PrivilegeTarget, quoteName bool) error {
	if len(target.Database) == 0 || sameDatabase(target.Database, ctx.GetCurrentDatabase()) {
		return nil
	}
	name := target.String()
	if quoteName {
		name = `"` + name + `"`
	}
	return pgerror.Newf(pgcode.FeatureNotSupported, "cross-database references are not implemented: %s", name)
}

// sameDatabase returns whether the two database names refer to the same database, ignoring any revision and case.
func sameDatabase(a string, b string) bool {
	aBase, _ := doltdb.SplitRevisionDbName(a)
	bBase, _ := doltdb.SplitRevisionDbName(b)
	return strings.EqualFold(aBase, bBase)
}

// currentSchemaDatabase returns the current database as a sql.SchemaDatabase. Returns false if the current database
// does not have schemas, in which case no existence checks can be made.
func currentSchemaDatabase(ctx *sql.Context) (sql.SchemaDatabase, bool, error) {
	db, err := core.GetSqlDatabaseFromContext(ctx, "")
	if err != nil || db == nil {
		return nil, false, err
	}
	schemaDB, ok := db.(sql.SchemaDatabase)
	return schemaDB, ok, nil
}

// schemaExists returns whether the given schema exists in the current database.
func schemaExists(ctx *sql.Context, schemaName string) (bool, error) {
	schemaDB, ok, err := currentSchemaDatabase(ctx)
	if err != nil || !ok {
		return !ok, err
	}
	_, exists, err := schemaDB.GetSchema(ctx, schemaName)
	return exists, err
}

// resolveTargetSchema returns the schema that a schema-wide target (ALL ... IN SCHEMA) refers to, which must exist.
func resolveTargetSchema(ctx *sql.Context, schemaName string) (string, error) {
	schemaName, err := core.GetSchemaName(ctx, nil, schemaName)
	if err != nil {
		return "", err
	}
	if err = checkSchemaExists(ctx, schemaName); err != nil {
		return "", err
	}
	return schemaName, nil
}

// checkSchemaExists returns an error if the given schema does not exist in the current database.
func checkSchemaExists(ctx *sql.Context, schemaName string) error {
	exists, err := schemaExists(ctx, schemaName)
	if err != nil {
		return err
	}
	if !exists {
		return pgerror.Newf(pgcode.UndefinedSchema, `schema "%s" does not exist`, schemaName)
	}
	return nil
}

// candidateSchemas returns the schemas that may contain the target: its own schema if it is qualified, and otherwise
// the search path. A qualified target's schema must exist.
func candidateSchemas(ctx *sql.Context, target PrivilegeTarget) ([]string, error) {
	if len(target.Schema) > 0 {
		if err := checkSchemaExists(ctx, target.Schema); err != nil {
			return nil, err
		}
		return []string{target.Schema}, nil
	}
	return core.SearchPath(ctx)
}

// resolveTableTargets resolves each target of a GRANT or REVOKE ON TABLE to the schema it lives in, returning an
// error for targets that do not exist or that are in another database.
func resolveTableTargets(ctx *sql.Context, targets []PrivilegeTarget) ([]PrivilegeTarget, error) {
	resolved := make([]PrivilegeTarget, len(targets))
	for i, target := range targets {
		if err := checkTargetDatabase(ctx, target, true); err != nil {
			return nil, err
		}
		var err error
		switch {
		case len(target.Name) == 0:
			target.Schema, err = resolveTargetSchema(ctx, target.Schema)
		case doltdb.HasDoltPrefix(target.Name):
			// Dolt system tables retain their existing SQL privilege namespace
			target.Schema, err = core.GetSchemaName(ctx, nil, target.Schema)
		default:
			target.Schema, err = relationSchema(ctx, target)
		}
		if err != nil {
			return nil, err
		}
		target.Database = ""
		resolved[i] = target
	}
	return resolved, nil
}

// relationSchema returns the schema containing the table, view, or sequence named by the target, mirroring how an
// unqualified relation resolves on the search path.
func relationSchema(ctx *sql.Context, target PrivilegeTarget) (string, error) {
	schemaDB, ok, err := currentSchemaDatabase(ctx)
	if err != nil {
		return "", err
	}
	if !ok {
		return core.GetSchemaName(ctx, nil, target.Schema)
	}
	candidates, err := candidateSchemas(ctx, target)
	if err != nil {
		return "", err
	}
	sequences, err := core.GetSequencesCollectionFromContext(ctx, "")
	if err != nil {
		return "", err
	}
	for _, candidate := range candidates {
		schema, exists, err := schemaDB.GetSchema(ctx, candidate)
		if err != nil {
			return "", err
		}
		if !exists {
			continue
		}
		_, found, err := schema.GetTableInsensitive(ctx, target.Name)
		if err != nil {
			return "", err
		}
		if found {
			return schema.SchemaName(), nil
		}
		if viewDB, ok := schema.(sql.ViewDatabase); ok {
			_, found, err = viewDB.GetViewDefinition(ctx, target.Name)
			if err != nil {
				return "", err
			}
			if found {
				return schema.SchemaName(), nil
			}
		}
		if sequences.HasSequence(ctx, id.NewSequence(schema.SchemaName(), target.Name)) {
			return schema.SchemaName(), nil
		}
	}
	return "", pgerror.Newf(pgcode.UndefinedTable, `relation "%s" does not exist`, target.String())
}

// resolveSequenceTargets resolves each target of a GRANT or REVOKE ON SEQUENCE to the schema it lives in, returning an
// error for targets that do not exist or that are in another database.
func resolveSequenceTargets(ctx *sql.Context, targets []PrivilegeTarget) ([]PrivilegeTarget, error) {
	resolved := make([]PrivilegeTarget, len(targets))
	for i, target := range targets {
		if err := checkTargetDatabase(ctx, target, true); err != nil {
			return nil, err
		}
		var err error
		if len(target.Name) == 0 {
			target.Schema, err = resolveTargetSchema(ctx, target.Schema)
		} else {
			target.Schema, err = sequenceSchema(ctx, target)
		}
		if err != nil {
			return nil, err
		}
		target.Database = ""
		resolved[i] = target
	}
	return resolved, nil
}

// sequenceSchema returns the schema containing the sequence named by the target, mirroring how nextval resolves an
// unqualified sequence on the search path.
func sequenceSchema(ctx *sql.Context, target PrivilegeTarget) (string, error) {
	candidates, err := candidateSchemas(ctx, target)
	if err != nil {
		return "", err
	}
	sequences, err := core.GetSequencesCollectionFromContext(ctx, "")
	if err != nil {
		return "", err
	}
	for _, candidate := range candidates {
		if sequences.HasSequence(ctx, id.NewSequence(candidate, target.Name)) {
			return candidate, nil
		}
	}
	return "", pgerror.Newf(pgcode.UndefinedTable, `relation "%s" does not exist`, target.String())
}

// resolveRoutineTargets resolves each target of a GRANT or REVOKE ON FUNCTION/PROCEDURE/ROUTINE to the schema it lives
// in, returning an error for targets that do not exist or that are in another database.
func resolveRoutineTargets(ctx *sql.Context, targets []PrivilegeTarget) ([]PrivilegeTarget, error) {
	resolved := make([]PrivilegeTarget, len(targets))
	for i, target := range targets {
		if err := checkTargetDatabase(ctx, target, false); err != nil {
			return nil, err
		}
		var err error
		if len(target.Name) == 0 {
			target.Schema, err = resolveTargetSchema(ctx, target.Schema)
		} else {
			target.Schema, err = routineSchema(ctx, target)
		}
		if err != nil {
			return nil, err
		}
		target.Database = ""
		resolved[i] = target
	}
	return resolved, nil
}

// routineSchema returns the schema containing the function, aggregate, or procedure named by the target, mirroring
// how an unqualified routine resolves on the search path.
func routineSchema(ctx *sql.Context, target PrivilegeTarget) (string, error) {
	candidates, err := candidateSchemas(ctx, target)
	if err != nil {
		return "", err
	}
	funcCollection, err := core.GetFunctionsCollectionFromContext(ctx, "")
	if err != nil {
		return "", err
	}
	aggCollection, err := core.GetAggregatesCollectionFromContext(ctx, "")
	if err != nil {
		return "", err
	}
	procCollection, err := core.GetProceduresCollectionFromContext(ctx, "")
	if err != nil {
		return "", err
	}
	for _, candidate := range candidates {
		if candidate == "pg_catalog" && isBuiltInRoutine(target.Name) {
			return candidate, nil
		}
		funcID := id.NewFunction(candidate, target.Name)
		overloads, err := funcCollection.GetFunctionOverloads(ctx, funcID)
		if err != nil {
			return "", err
		}
		if len(overloads) > 0 {
			return candidate, nil
		}
		aggOverloads, err := aggCollection.GetAggregateOverloads(ctx, funcID)
		if err != nil {
			return "", err
		}
		if len(aggOverloads) > 0 {
			return candidate, nil
		}
		procOverloads, err := procCollection.GetProcedureOverloads(ctx, id.NewProcedure(candidate, target.Name))
		if err != nil {
			return "", err
		}
		if len(procOverloads) > 0 {
			return candidate, nil
		}
	}
	return "", pgerror.Newf(pgcode.UndefinedFunction, `function %s(%s) does not exist`, target.String(), target.Args)
}

// isBuiltInRoutine returns whether a built-in function or aggregate has the given name. Built-in routines live in
// pg_catalog.
func isBuiltInRoutine(name string) bool {
	if _, ok := framework.Catalog[name]; ok {
		return true
	}
	_, ok := framework.AggregateCatalog[name]
	return ok
}

// checkSchemaTargets returns an error if any of the schemas named by a GRANT or REVOKE ON SCHEMA do not exist.
func checkSchemaTargets(ctx *sql.Context, schemas []string) error {
	for _, schema := range schemas {
		if err := checkSchemaExists(ctx, schema); err != nil {
			return err
		}
	}
	return nil
}

// checkDatabaseTargets returns an error if any of the databases named by a GRANT or REVOKE ON DATABASE do not exist.
func checkDatabaseTargets(ctx *sql.Context, databases []string) error {
	for _, database := range databases {
		db, err := core.GetSqlDatabaseFromContext(ctx, database)
		if err != nil {
			return err
		}
		if db == nil {
			return pgerror.Newf(pgcode.UndefinedDatabase, `database "%s" does not exist`, database)
		}
	}
	return nil
}
