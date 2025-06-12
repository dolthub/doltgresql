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

	"github.com/dolthub/doltgresql/utils"
)

// Database is a wrapper around Dolt's database object, allowing for functionality specific to Doltgres (such as system
// tables).
type Database struct {
	db sqle.Database
}

var _ sql.DatabaseSchema = Database{}

// GetTableInsensitive implements the interface sql.DatabaseSchema.
func (d Database) GetTableInsensitive(ctx *sql.Context, tblName string) (sql.Table, bool, error) {
	// Even though this is named "GetTableInsensitive", due to differences in Postgres and MySQL, this should perform an
	// exact search.
	if tableMap, ok := handlers[d.db.Schema()]; ok {
		if handler, ok := tableMap[tblName]; ok {
			return NewVirtualTable(handler, d.db), true, nil
		}
	}
	return nil, false, nil
}

// GetTableNames implements the interface sql.DatabaseSchema.
func (d Database) GetTableNames(ctx *sql.Context) ([]string, error) {
	tableMap := handlers[d.db.Schema()]
	return utils.GetMapKeysSorted(tableMap), nil
}

// Name implements the interface sql.DatabaseSchema.
func (d Database) Name() string {
	return d.db.Name()
}

func (d Database) SchemaName() string {
	return d.db.SchemaName()
}
