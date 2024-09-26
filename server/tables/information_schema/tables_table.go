// Copyright 2022 Dolthub, Inc.
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
	"github.com/dolthub/doltgresql/server/types/oid"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/information_schema"
)

// newTablesTable returns a InformationSchemaTable for MySQL.
func newTablesTable() *information_schema.InformationSchemaTable {
	return &information_schema.InformationSchemaTable{
		TableName:   information_schema.TablesTableName,
		TableSchema: tablesSchema,
		Reader:      tablesRowIter,
	}
}

// tablesSchema is the schema for the information_schema.TABLES table.
var tablesSchema = sql.Schema{
	{Name: "table_catalog", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.TablesTableName},
	{Name: "table_schema", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.TablesTableName},
	{Name: "table_name", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.TablesTableName},
	{Name: "table_type", Type: character_data, Default: nil, Nullable: true, Source: information_schema.TablesTableName},
	{Name: "self_referencing_column_name", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.TablesTableName},
	{Name: "reference_generation", Type: character_data, Default: nil, Nullable: true, Source: information_schema.TablesTableName},
	{Name: "user_defined_type_catalog", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.TablesTableName},
	{Name: "user_defined_type_schema", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.TablesTableName},
	{Name: "user_defined_type_name", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.TablesTableName},
	{Name: "is_insertable_into", Type: yes_or_no, Default: nil, Nullable: true, Source: information_schema.TablesTableName},
	{Name: "is_typed", Type: yes_or_no, Default: nil, Nullable: true, Source: information_schema.TablesTableName},
	{Name: "commit_action", Type: yes_or_no, Default: nil, Nullable: true, Source: information_schema.TablesTableName},
}

// tablesRowIter implements the sql.RowIter for the information_schema.TABLES table.
func tablesRowIter(ctx *sql.Context, cat sql.Catalog) (sql.RowIter, error) {
	var rows []sql.Row

	err := oid.IterateCurrentDatabase(ctx, oid.Callbacks{
		Table: func(ctx *sql.Context, schema oid.ItemSchema, table oid.ItemTable) (cont bool, err error) {
			// TODO: Foreign and temporary tables.
			rows = append(rows, sql.Row{
				schema.Item.Name(),       // table_catalog
				schema.Item.SchemaName(), // table_schema
				table.Item.Name(),        // table_name
				"BASE TABLE",             // table_type
				nil,                      // self_referencing_column_name
				nil,                      // reference_generation
				nil,                      // user_defined_type_catalog
				nil,                      // user_defined_type_schema
				nil,                      // user_defined_type_name
				"YES",                    // is_insertable_into
				"NO",                     // is_typed
				nil,                      // commit_action
			})
			return true, nil
		},
		View: func(ctx *sql.Context, schema oid.ItemSchema, view oid.ItemView) (cont bool, err error) {
			// TODO: Fill out the rest of the columns.
			rows = append(rows, sql.Row{
				schema.Item.Name(),       // table_catalog
				schema.Item.SchemaName(), // table_schema
				view.Item.Name,           // table_name
				"VIEW",                   // table_type
				nil,                      // self_referencing_column_name
				nil,                      // reference_generation
				nil,                      // user_defined_type_catalog
				nil,                      // user_defined_type_schema
				nil,                      // user_defined_type_name
				nil,                      // is_insertable_into
				"NO",                     // is_typed
				nil,                      // commit_action
			})
			return true, nil
		},
	})
	if err != nil {
		return nil, err
	}

	return sql.RowsToRowIter(rows...), nil
}
