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

package core

import (
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/resolve"
	"github.com/dolthub/go-mysql-server/sql"
)

// LookupCurrentSchema returns the first schema on the search path that exists, reporting whether there was one.
// Finding none is not an error here: only the caller knows whether it needs somewhere to create an object or is
// resolving a name that may simply not exist.
func LookupCurrentSchema(ctx *sql.Context) (string, bool, error) {
	_, root, err := GetRootFromContext(ctx)
	if err != nil {
		return "", false, err
	}
	// The current database may not be backed by a Doltgres *RootValue (e.g. Dolt's synthetic dolt_cluster system
	// database), in which case there's no search path to resolve against.
	if root == nil {
		return "public", true, nil
	}
	schemaName, err := resolve.FirstExistingSchemaOnSearchPath(ctx, root)
	if err != nil {
		// Dolt reports "no schema on the search path exists" with an error phrased for its CREATE TABLE caller.
		if sql.ErrDatabaseNoDatabaseSchemaSelectedCreate.Is(err) {
			return "", false, nil
		}
		return "", false, err
	}
	return schemaName, true, nil
}

// LookupSchemaName is GetSchemaName for callers that treat a missing schema as "not found" rather than an error.
func LookupSchemaName(ctx *sql.Context, db sql.Database, schemaName string) (string, bool, error) {
	if schemaName == "" {
		if schema, isSch := db.(sql.DatabaseSchema); isSch {
			schemaName = schema.SchemaName()
		}
		if schemaName == "" {
			return LookupCurrentSchema(ctx)
		}
	}
	return schemaName, true, nil
}

// GetCurrentSchema returns the current schema used by the context. Defaults to "public" if the context does not specify
// a schema. Returns sql.ErrDatabaseNoDatabaseSchemaSelectedCreate when no schema on the search path exists, which suits
// callers creating an object; use LookupCurrentSchema to treat that as "not found" instead.
func GetCurrentSchema(ctx *sql.Context) (string, error) {
	schemaName, ok, err := LookupCurrentSchema(ctx)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", sql.ErrDatabaseNoDatabaseSchemaSelectedCreate.New()
	}
	return schemaName, nil
}

// GetSchemaName returns the schema name if there is any exist.
// If the given schema is not empty, it's returned.
// If it is empty, uses given database to get schema name if it's DatabaseSchema.
// If it's not of DatabaseSchema type or the schema name of it is empty,
// it tries retrieving the current schema used by the context.
// Defaults to "public" if the context does not specify a schema.
func GetSchemaName(ctx *sql.Context, db sql.Database, schemaName string) (string, error) {
	name, ok, err := LookupSchemaName(ctx, db, schemaName)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", sql.ErrDatabaseNoDatabaseSchemaSelectedCreate.New()
	}
	return name, nil
}
