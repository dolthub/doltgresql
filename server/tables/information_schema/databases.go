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

package information_schema

import (
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/information_schema"
	"github.com/dolthub/go-mysql-server/sql/mysql_db"
)

// allDatabasesWithNames returns the current database(s) and their catalog and schema names.
func allDatabasesWithNames(ctx *sql.Context, cat sql.Catalog, privCheck bool) ([]information_schema.DbWithNames, error) {
	var dbs []information_schema.DbWithNames

	currentDB := ctx.GetCurrentDatabase()
	currentRevDB, _ := dsess.SplitRevisionDbName(currentDB)

	allDbs := cat.AllDatabases(ctx)
	for _, db := range allDbs {
		if privCheck {
			if privDatabase, ok := db.(mysql_db.PrivilegedDatabase); ok {
				db = privDatabase.Unwrap()
			}
		}

		sdb, ok := db.(sql.SchemaDatabase)
		if ok {
			var dbsForSchema []information_schema.DbWithNames
			schemas, err := sdb.AllSchemas(ctx)
			if err != nil {
				return nil, err
			}

			for _, schema := range schemas {
				dbName := db.Name()
				revDb, _ := dsess.SplitRevisionDbName(dbName)
				// Add database it is the current database/revision database and if SchemaName exists
				if schema.SchemaName() != "" && (dbName == currentDB || revDb == currentRevDB) {
					dbsForSchema = append(dbsForSchema, information_schema.DbWithNames{
						Database:    schema,
						CatalogName: schema.Name(),
						SchemaName:  schema.SchemaName(),
					})
				}
			}
			dbs = append(dbs, dbsForSchema...)
		}
	}

	return dbs, nil
}
