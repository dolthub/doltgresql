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

package pgcatalog

import (
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/tables"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// PgDescriptionName is a constant to the pg_description name.
const PgDescriptionName = "pg_description"

// InitPgDescription handles registration of the pg_description handler.
func InitPgDescription() {
	tables.AddHandler(PgCatalogName, PgDescriptionName, PgDescriptionHandler{})
	tables.AddInitializeTable(PgCatalogName, pgDescriptionInitializeTable)
}

// PgDescriptionHandler is the handler for the pg_description table.
type PgDescriptionHandler struct{}

var _ tables.DataTableHandler = PgDescriptionHandler{}

// Insert implements the interface tables.DataTableHandler.
func (p PgDescriptionHandler) Insert(ctx *sql.Context, editor *tables.DataTableEditor, row sql.Row) error {
	return editor.Insert(ctx, row)
}

// Update implements the interface tables.DataTableHandler.
func (p PgDescriptionHandler) Update(ctx *sql.Context, editor *tables.DataTableEditor, old sql.Row, new sql.Row) error {
	return editor.Update(ctx, old, new)
}

// Delete implements the interface tables.DataTableHandler.
func (p PgDescriptionHandler) Delete(ctx *sql.Context, editor *tables.DataTableEditor, row sql.Row) error {
	return editor.Delete(ctx, row)
}

// UsesIndexes implements the interface tables.DataTableHandler.
func (p PgDescriptionHandler) UsesIndexes() bool {
	return true
}

// RowIter implements the interface tables.DataTableHandler.
func (p PgDescriptionHandler) RowIter(ctx *sql.Context, rowIter sql.RowIter) (sql.RowIter, error) {
	return rowIter, nil
}

// pgDescriptionSchema is the schema for pg_description.
var pgDescriptionSchema = sql.Schema{
	{Name: "objoid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgDescriptionName},
	{Name: "classoid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgDescriptionName},
	{Name: "objsubid", Type: pgtypes.Int32, Default: nil, Nullable: false, Source: PgDescriptionName},
	{Name: "description", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgDescriptionName},
}

const (
	pgDescription_objoid      int = 0
	pgDescription_classoid    int = 1
	pgDescription_objsubid    int = 2
	pgDescription_description int = 3
)

// pgDescriptionInitializeTable is the tables.InitializeTable function for pg_description.
func pgDescriptionInitializeTable(ctx *sql.Context, db sqle.Database) error {
	return db.CreateIndexedTable(ctx, PgDescriptionName, sql.PrimaryKeySchema{
		Schema: pgDescriptionSchema,
	}, sql.IndexDef{
		Name: "pg_description_o_c_o_index",
		Columns: []sql.IndexColumn{
			{Name: pgDescriptionSchema[pgDescription_objoid].Name},
			{Name: pgDescriptionSchema[pgDescription_classoid].Name},
			{Name: pgDescriptionSchema[pgDescription_objsubid].Name},
		},
		Constraint: sql.IndexConstraint_Unique,
		Storage:    sql.IndexUsing_BTree,
		Comment:    "",
	}, sql.Collation_Default)
}
