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

package pgcatalog

import "github.com/dolthub/go-mysql-server/sql"

// emptyRowIter implements the sql.RowIter for empty table.
func emptyRowIter() (sql.RowIter, error) {
	return sql.RowsToRowIter(), nil
}

// currentDatabaseSchemaIter iterates over all schemas in the current database, calling cb
// for each schema. Once all schemas have been processed or the callback returns
// false or an error, the iteration stops.
func currentDatabaseSchemaIter(ctx *sql.Context, c sql.Catalog, cb func(schema sql.DatabaseSchema) (bool, error)) (sql.Database, error) {
	currentDB := ctx.GetCurrentDatabase()
	db, err := c.Database(ctx, currentDB)
	if err != nil {
		return nil, err
	}

	if schDB, ok := db.(sql.SchemaDatabase); ok {
		schemas, err := schDB.AllSchemas(ctx)
		if err != nil {
			return nil, err
		}

		for _, schema := range schemas {
			cont, err := cb(schema)
			if err != nil {
				return nil, err
			}
			if !cont {
				break
			}
		}
	}

	return db, nil
}
