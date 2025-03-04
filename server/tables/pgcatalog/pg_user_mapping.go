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

// PgUserMappingName is a constant to the pg_user_mapping name.
const PgUserMappingName = "pg_user_mapping"

// InitPgUserMapping handles registration of the pg_user_mapping handler.
func InitPgUserMapping() {
	tables.AddHandler(PgCatalogName, PgUserMappingName, PgUserMappingHandler{})
}

// PgUserMappingHandler is the handler for the pg_user_mapping table.
type PgUserMappingHandler struct{}

var _ tables.Handler = PgUserMappingHandler{}

// Name implements the interface tables.Handler.
func (p PgUserMappingHandler) Name() string {
	return PgUserMappingName
}

// RowIter implements the interface tables.Handler.
func (p PgUserMappingHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_user_mapping row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgUserMappingHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgUserMappingSchema,
		PkOrdinals: nil,
	}
}

// pgUserMappingSchema is the schema for pg_user_mapping.
var pgUserMappingSchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgUserMappingName},
	{Name: "umuser", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgUserMappingName},
	{Name: "umserver", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgUserMappingName},
	{Name: "umoptions", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgUserMappingName}, // TODO: collation C
}

// pgUserMappingRowIter is the sql.RowIter for the pg_user_mapping table.
type pgUserMappingRowIter struct {
}

var _ sql.RowIter = (*pgUserMappingRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgUserMappingRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgUserMappingRowIter) Close(ctx *sql.Context) error {
	return nil
}
