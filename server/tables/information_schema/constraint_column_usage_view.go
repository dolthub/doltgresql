// Copyright 2026 Dolthub, Inc.
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
)

// ConstraintColumnUsageViewName is the name of the CONSTRAINT_COLUMN_USAGE view.
const ConstraintColumnUsageViewName = "constraint_column_usage"

// newConstraintColumnUsageView creates a new information_schema.CONSTRAINT_COLUMN_USAGE view.
func newConstraintColumnUsageView() *information_schema.InformationSchemaTable {
	return &information_schema.InformationSchemaTable{
		TableName:   ConstraintColumnUsageViewName,
		TableSchema: constraintColumnUsageSchema,
		Reader:      constraintColumnUsageRowIter,
	}
}

// constraintColumnUsage is the schema for the information_schema.CONSTRAINT_COLUMN_USAGE view.
var constraintColumnUsageSchema = sql.Schema{
	{Name: "table_catalog", Type: sql_identifier, Default: nil, Nullable: true, Source: ConstraintColumnUsageViewName},
	{Name: "table_schema", Type: sql_identifier, Default: nil, Nullable: true, Source: ConstraintColumnUsageViewName},
	{Name: "table_name", Type: sql_identifier, Default: nil, Nullable: true, Source: ConstraintColumnUsageViewName},
	{Name: "column_name", Type: character_data, Default: nil, Nullable: true, Source: ConstraintColumnUsageViewName},
	{Name: "constraint_catalog", Type: character_data, Default: nil, Nullable: true, Source: ConstraintColumnUsageViewName},
	{Name: "constraint_schema", Type: yes_or_no, Default: nil, Nullable: true, Source: ConstraintColumnUsageViewName},
	{Name: "constraint_name", Type: yes_or_no, Default: nil, Nullable: true, Source: ConstraintColumnUsageViewName},
}

// constraintColumnUsageRowIter implements the sql.RowIter for the information_schema.CONSTRAINT_COLUMN_USAGE view.
func constraintColumnUsageRowIter(ctx *sql.Context, catalog sql.Catalog) (sql.RowIter, error) {
	var rows []sql.Row

	err := functions.IterateCurrentDatabase(ctx, functions.Callbacks{
		Check: func(ctx *sql.Context, schema functions.ItemSchema, table functions.ItemTable, check functions.ItemCheck) (cont bool, err error) {

			// TODO: Fill out the rest of the columns.
			rows = append(rows, sql.Row{
				schema.Item.Name(),       // table_catalog
				schema.Item.SchemaName(), // table_schema
				table.Item.Name(),        // table_name
				nil,                      // column_name
				nil,                      // constraint_catalog
				nil,                      // constraint_schema
				check.Item.Name,          // constraint_name
			})
			return true, nil
		},
	})
	if err != nil {
		return nil, err
	}

	return sql.RowsToRowIter(rows...), nil
}
