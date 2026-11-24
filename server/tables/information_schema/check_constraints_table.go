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
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/information_schema"

	"github.com/dolthub/doltgresql/server/functions"
)

// newCheckConstraintsTable returns an InformationSchemaTable for the information_schema.check_constraints table.
func newCheckConstraintsTable() *information_schema.InformationSchemaTable {
	return &information_schema.InformationSchemaTable{
		TableName:   information_schema.CheckConstraintsTableName,
		TableSchema: checkConstraintsSchema,
		Reader:      checkConstraintsRowIter,
	}
}

// checkConstraintsSchema is the schema for the information_schema.check_constraints table.
var checkConstraintsSchema = sql.Schema{
	{Name: "constraint_catalog", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.CheckConstraintsTableName},
	{Name: "constraint_schema", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.CheckConstraintsTableName},
	{Name: "constraint_name", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.CheckConstraintsTableName},
	{Name: "check_clause", Type: character_data, Default: nil, Nullable: true, Source: information_schema.CheckConstraintsTableName},
}

// checkConstraintsRowIter implements the sql.RowIter for the information_schema.check_constraints table. Check
// expressions are persisted with the schema of every function that they call, and are shown with their function calls
// as Postgres shows them.
func checkConstraintsRowIter(ctx *sql.Context, _ sql.Catalog) (sql.RowIter, error) {
	var rows []sql.Row
	catalogName, _ := doltdb.SplitRevisionDbName(ctx.GetCurrentDatabase())
	err := functions.IterateCurrentDatabase(ctx, functions.Callbacks{
		Check: func(ctx *sql.Context, schema functions.ItemSchema, table functions.ItemTable, check functions.ItemCheck) (cont bool, err error) {
			tableName := doltdb.TableName{Schema: schema.Item.SchemaName(), Name: table.Item.Name()}
			rows = append(rows, sql.Row{
				catalogName,              // constraint_catalog
				schema.Item.SchemaName(), // constraint_schema
				check.Item.Name,          // constraint_name
				functions.DisplayCheckExpression(ctx, tableName, check.Item.CheckExpression), // check_clause
			})
			return true, nil
		},
	})
	if err != nil {
		return nil, err
	}
	return sql.RowsToRowIter(rows...), nil
}
