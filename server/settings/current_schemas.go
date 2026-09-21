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

package settings

import (
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/resolve"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core"
)

// GetCurrentSchemas returns the session's effective search path: every schema named by the search_path setting that
// exists in the current database, in search_path order. Postgres only ever reports searchable schemas from
// current_schema() and current_schemas(), so entries naming no existing schema are dropped, as are repeats of a schema
// already on the path. That includes "$user", which expands to the session's user name and is typically dropped
// because no schema by that name exists. pg_catalog is not added here: it is only implicitly searched, and reporting
// it is current_schemas(true)'s job.
func GetCurrentSchemas(ctx *sql.Context) ([]string, error) {
	pathElems, err := resolve.SearchPath(ctx)
	if err != nil {
		return nil, err
	}

	db, err := core.GetSqlDatabaseFromContext(ctx, "")
	if err != nil || db == nil {
		return nil, err
	}
	schemaDb, ok := db.(sql.SchemaDatabase)
	if !ok {
		return nil, nil
	}

	var path []string
	seen := make(map[string]struct{}, len(pathElems))
	for _, schemaName := range pathElems {
		schema, exists, err := schemaDb.GetSchema(ctx, schemaName)
		if err != nil {
			return nil, err
		}
		if !exists {
			continue
		}
		// Schema lookup is case-insensitive, so report the name the schema was created with, like Postgres does
		name := schema.SchemaName()
		if _, ok = seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		path = append(path, name)
	}

	return path, nil
}

// GetCurrentSchemasAsMap returns the schemas from GetCurrentSchemas as a map for easy lookup.
func GetCurrentSchemasAsMap(ctx *sql.Context) (map[string]struct{}, error) {
	schemas, err := GetCurrentSchemas(ctx)
	if err != nil {
		return nil, err
	}
	schemaMap := make(map[string]struct{}, len(schemas))
	for _, schema := range schemas {
		schemaMap[schema] = struct{}{}
	}
	return schemaMap, nil
}
