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

package tables

import (
	"fmt"
	"strings"

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/schema"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle"
	"github.com/dolthub/go-mysql-server/sql"
)

// Init handles initialization of all Postgres-specific and Doltgres-specific tables.
func Init() {
	originalTableFunc := sqle.NewDoltSqlTable
	sqle.NewDoltSqlTable = func(db sqle.Database, tableName string, sch schema.Schema, tbl *doltdb.Table) (sql.Table, error) {
		sqlTable, err := originalTableFunc(db, tableName, sch, tbl)
		if err != nil {
			return nil, err
		}
		const pgCatalogName = "pg_catalog"
		if strings.EqualFold(db.Schema(), pgCatalogName) {
			switch t := sqlTable.(type) {
			case *sqle.DoltTable:
				sqlTable = NewDataTable(t, pgCatalogName)
			case *sqle.WritableDoltTable:
				sqlTable = NewWritableDataTable(t, pgCatalogName)
			case *sqle.AlterableDoltTable:
				sqlTable = NewWritableDataTable(&t.WritableDoltTable, pgCatalogName)
			default:
				return nil, fmt.Errorf("unexpected Dolt table type in pg_catalog: %T", sqlTable)
			}
		}
		return sqlTable, nil
	}
}
