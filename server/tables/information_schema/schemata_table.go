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
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/information_schema"
)

// newSchemataTable creates a new information_schema.SCHEMATA table.
func newSchemataTable() *information_schema.InformationSchemaTable {
	return &information_schema.InformationSchemaTable{
		TableName:   information_schema.SchemataTableName,
		TableSchema: schemataSchema,
		Reader:      schemataRowIter,
	}
}

// schemataSchema is the schema for the information_schema.SCHEMATA table.
var schemataSchema = sql.Schema{
	{Name: "catalog_name", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.SchemataTableName},
	{Name: "schema_name", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.SchemataTableName},
	{Name: "schema_owner", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.SchemataTableName},
	{Name: "default_character_set_catalog", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.SchemataTableName},
	{Name: "default_character_set_schema", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.SchemataTableName},
	{Name: "default_character_set_name", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.SchemataTableName},
	{Name: "sql_path", Type: character_data, Default: nil, Nullable: true, Source: information_schema.SchemataTableName},
}

// schemataRowIter implements the sql.RowIter for the information_schema.SCHEMATA table.
func schemataRowIter(ctx *sql.Context, c sql.Catalog) (sql.RowIter, error) {
	dbs, err := information_schema.AllDatabases(ctx, c, false)
	if err != nil {
		return nil, err
	}

	var rows []sql.Row

	for _, db := range dbs {
		rows = append(rows, sql.Row{
			db.CatalogName, // catalog_name
			db.SchemaName,  // schema_name
			"",             // schema_owner
			nil,            // default_character_set_catalog
			nil,            // default_character_set_schema
			nil,            // default_character_set_name
			nil,            // sql_path
		})
	}

	return sql.RowsToRowIter(rows...), nil
}
