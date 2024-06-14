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
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle"
	"github.com/dolthub/go-mysql-server/sql"
)

// Init handles initialization of all Postgres-specific and Doltgres-specific tables.
func Init() {
	sqle.HandleSchema = func(ctx *sql.Context, schemaName string, db sqle.Database) (sql.DatabaseSchema, error) {
		if _, ok := handlers[schemaName]; ok {
			return Database{db}, nil
		}
		return db, nil
	}
}
