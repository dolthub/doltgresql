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

// GetCurrentSchema returns the current schema used by the context. Defaults to "public" if the context does not specify
// a schema.
func GetCurrentSchema(ctx *sql.Context) (string, error) {
	_, root, err := GetRootFromContext(ctx)
	if err != nil {
		return "", nil
	}

	return resolve.FirstExistingSchemaOnSearchPath(ctx, root)
}

// GetSchemaName returns the schema name if there is any exist.
// If the given schema is not empty, it's returned.
// If it is empty, uses given database to get schema name if it's DatabaseSchema.
// If it's not of DatabaseSchema type or the schema name of it is empty,
// it tries retrieving the current schema used by the context.
// Defaults to "public" if the context does not specify a schema.
func GetSchemaName(ctx *sql.Context, db sql.Database, schemaName string) (string, error) {
	if schemaName == "" {
		if schema, isSch := db.(sql.DatabaseSchema); isSch {
			schemaName = schema.SchemaName()
		}
		if schemaName == "" {
			return GetCurrentSchema(ctx)
		}
	}
	return schemaName, nil
}
