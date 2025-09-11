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
	"io"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/tables"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// PgForeignDataWrapperName is a constant to the pg_foreign_data_wrapper name.
const PgForeignDataWrapperName = "pg_foreign_data_wrapper"

// InitPgForeignDataWrapper handles registration of the pg_foreign_data_wrapper handler.
func InitPgForeignDataWrapper() {
	tables.AddHandler(PgCatalogName, PgForeignDataWrapperName, PgForeignDataWrapperHandler{})
}

// PgForeignDataWrapperHandler is the handler for the pg_foreign_data_wrapper table.
type PgForeignDataWrapperHandler struct{}

var _ tables.Handler = PgForeignDataWrapperHandler{}

// Name implements the interface tables.Handler.
func (p PgForeignDataWrapperHandler) Name() string {
	return PgForeignDataWrapperName
}

// RowIter implements the interface tables.Handler.
func (p PgForeignDataWrapperHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_foreign_data_wrapper row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgForeignDataWrapperHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgForeignDataWrapperSchema,
		PkOrdinals: nil,
	}
}

// pgForeignDataWrapperSchema is the schema for pg_foreign_data_wrapper.
var pgForeignDataWrapperSchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgForeignDataWrapperName},
	{Name: "fdwname", Type: pgtypes.Name, Default: nil, Nullable: false, Source: PgForeignDataWrapperName},
	{Name: "fdwowner", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgForeignDataWrapperName},
	{Name: "fdwhandler", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgForeignDataWrapperName},
	{Name: "fdwvalidator", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgForeignDataWrapperName},
	{Name: "fdwacl", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgForeignDataWrapperName},     // TODO: aclitem[] type
	{Name: "fdwoptions", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgForeignDataWrapperName}, // TODO: collation C
}

// pgForeignDataWrapperRowIter is the sql.RowIter for the pg_foreign_data_wrapper table.
type pgForeignDataWrapperRowIter struct {
}

var _ sql.RowIter = (*pgForeignDataWrapperRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgForeignDataWrapperRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgForeignDataWrapperRowIter) Close(ctx *sql.Context) error {
	return nil
}
