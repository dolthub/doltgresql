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
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/information_schema"

	"github.com/dolthub/doltgresql/server/functions"

	"github.com/dolthub/doltgresql/postgres/parser/parser"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// newViewsTable creates a new information_schema.VIEWS table.
func newViewsTable() *information_schema.InformationSchemaTable {
	return &information_schema.InformationSchemaTable{
		TableName:   information_schema.ViewsTableName,
		TableSchema: viewsSchema,
		Reader:      viewsRowIter,
	}
}

// viewsSchema is the schema for the information_schema.VIEWS table.
var viewsSchema = sql.Schema{
	{Name: "table_catalog", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.ViewsTableName},
	{Name: "table_schema", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.ViewsTableName},
	{Name: "table_name", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.ViewsTableName},
	{Name: "view_definition", Type: character_data, Default: nil, Nullable: true, Source: information_schema.ViewsTableName},
	{Name: "check_option", Type: character_data, Default: nil, Nullable: true, Source: information_schema.ViewsTableName},
	{Name: "is_updatable", Type: yes_or_no, Default: nil, Nullable: true, Source: information_schema.ViewsTableName},
	{Name: "is_insertable_into", Type: yes_or_no, Default: nil, Nullable: true, Source: information_schema.ViewsTableName},
	{Name: "is_trigger_updatable", Type: yes_or_no, Default: nil, Nullable: true, Source: information_schema.ViewsTableName},
	{Name: "is_trigger_deletable", Type: yes_or_no, Default: nil, Nullable: false, Source: information_schema.ViewsTableName},
	{Name: "is_trigger_insertable_into", Type: yes_or_no, Default: nil, Nullable: false, Source: information_schema.ViewsTableName},
}

// viewsRowIter implements the sql.RowIter for the information_schema.VIEWS table.
func viewsRowIter(ctx *sql.Context, catalog sql.Catalog) (sql.RowIter, error) {
	var rows []sql.Row

	err := functions.IterateCurrentDatabase(ctx, functions.Callbacks{
		View: func(ctx *sql.Context, schema functions.ItemSchema, view functions.ItemView) (cont bool, err error) {
			stmts, err := parser.Parse(view.Item.CreateViewStatement)
			if err != nil {
				return false, err
			}
			if len(stmts) == 0 {
				return false, sql.ErrViewCreateStatementInvalid.New(view.Item.CreateViewStatement)
			}
			cv, ok := stmts[0].AST.(*tree.CreateView)
			if !ok {
				return false, sql.ErrViewCreateStatementInvalid.New(view.Item.CreateViewStatement)
			}

			viewDef := cv.AsSource.String()

			checkOpt := "NONE"
			if cv.CheckOption == tree.ViewCheckOptionCascaded {
				checkOpt = "CASCADED"
			}
			if cv.CheckOption == tree.ViewCheckOptionLocal {
				checkOpt = "LOCAL"
			}

			// TODO: Fill out the rest of the columns.
			rows = append(rows, sql.Row{
				schema.Item.Name(),       // table_catalog
				schema.Item.SchemaName(), // table_schema
				view.Item.Name,           // table_name
				viewDef,                  // view_definition
				checkOpt,                 // check_option
				nil,                      // is_updatable
				nil,                      // is_insertable_into
				nil,                      // is_trigger_updatable
				nil,                      // is_trigger_deletable
				nil,                      // is_trigger_insertable_into
			})
			return true, nil
		},
	})
	if err != nil {
		return nil, err
	}

	return sql.RowsToRowIter(rows...), nil
}
