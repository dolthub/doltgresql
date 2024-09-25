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

// tablesRowIter implements the sql.RowIter for the information_schema.TABLES table.
// func tablesRowIter2(ctx *sql.Context, cat sql.Catalog) (sql.RowIter, error) {
// 	var rows []sql.Row
// 	var (
// 		tableType      string
// 		tableRows      uint64
// 		avgRowLength   uint64
// 		dataLength     uint64
// 		engine         interface{}
// 		rowFormat      interface{}
// 		tableCollation interface{}
// 		autoInc        interface{}
// 	)

// 	databases, err := allDatabasesWithNames(ctx, cat, true)
// 	if err != nil {
// 		return nil, err
// 	}

// 	for _, db := range databases {
// 		if db.Database.Name() == sql.InformationSchemaDatabaseName {
// 			tableType = "SYSTEM VIEW"
// 		} else {
// 			tableType = "BASE TABLE"
// 			engine = "InnoDB"
// 			rowFormat = "Dynamic"
// 		}

// 		y2k, _, _ := types.Timestamp.Convert("2000-01-01 00:00:00")
// 		err := sql.DBTableIter(ctx, db.Database, func(t sql.Table) (cont bool, err error) {
// 			tableCollation = t.Collation().String()
// 			comment := ""
// 			if db.Database.Name() != sql.InformationSchemaDatabaseName {
// 				if st, ok := t.(sql.StatisticsTable); ok {
// 					tableRows, _, err = st.RowCount(ctx)
// 					if err != nil {
// 						return false, err
// 					}

// 					// TODO: correct values for avg_row_length, data_length, max_data_length are missing (current values varies on gms vs Dolt)
// 					//  index_length and data_free columns are not supported yet
// 					//  the data length values differ from MySQL
// 					// MySQL uses default page size (16384B) as data length, and it adds another page size, if table data fills the current page block.
// 					// https://stackoverflow.com/questions/34211377/average-row-length-higher-than-possible has good explanation.
// 					dataLength, err = st.DataLength(ctx)
// 					if err != nil {
// 						return false, err
// 					}

// 					if tableRows > uint64(0) {
// 						avgRowLength = dataLength / tableRows
// 					}
// 				}

// 				if ai, ok := t.(sql.AutoIncrementTable); ok {
// 					autoInc, err = ai.PeekNextAutoIncrementValue(ctx)
// 					if !errors.Is(err, sql.ErrNoAutoIncrementCol) && err != nil {
// 						return false, err
// 					}

// 					// table with no auto incremented column is qualified as AutoIncrementTable, and the nextAutoInc value is 0
// 					// table with auto incremented column and no rows, the nextAutoInc value is 1
// 					if autoInc == uint64(0) || autoInc == uint64(1) {
// 						autoInc = nil
// 					}
// 				}

// 				if commentedTable, ok := t.(sql.CommentedTable); ok {
// 					comment = commentedTable.Comment()
// 				}
// 			}

// 			rows = append(rows, sql.Row{
// 				db.CatalogName, // table_catalog
// 				db.SchemaName,  // table_schema
// 				t.Name(),       // table_name
// 				tableType,      // table_type
// 				engine,         // engine
// 				10,             // version (protocol, always 10)
// 				rowFormat,      // row_format
// 				tableRows,      // table_rows
// 				avgRowLength,   // avg_row_length
// 				dataLength,     // data_length
// 				0,              // max_data_length
// 				0,              // index_length
// 				0,              // data_free
// 				autoInc,        // auto_increment
// 				y2k,            // create_time
// 				y2k,            // update_time
// 				nil,            // check_time
// 				tableCollation, // table_collation
// 				nil,            // checksum
// 				"",             // create_options
// 				comment,        // table_comment
// 			})

// 			return true, nil
// 		})

// 		if err != nil {
// 			return nil, err
// 		}

// 		views, err := information_schema.ViewsInDatabase(ctx, db.Database)
// 		if err != nil {
// 			return nil, err
// 		}

// 		for _, view := range views {
// 			rows = append(rows, sql.Row{
// 				db.CatalogName, // table_catalog
// 				db.SchemaName,  // table_schema
// 				view.Name,      // table_name
// 				"VIEW",         // table_type
// 				nil,            // engine
// 				nil,            // version (protocol, always 10)
// 				nil,            // row_format
// 				nil,            // table_rows
// 				nil,            // avg_row_length
// 				nil,            // data_length
// 				nil,            // max_data_length
// 				nil,            // max_data_length
// 				nil,            // data_free
// 				nil,            // auto_increment
// 				y2k,            // create_time
// 				nil,            // update_time
// 				nil,            // check_time
// 				nil,            // table_collation
// 				nil,            // checksum
// 				nil,            // create_options
// 				"VIEW",         // table_comment
// 			})
// 		}
// 	}

// 	return sql.RowsToRowIter(rows...), nil
// }
