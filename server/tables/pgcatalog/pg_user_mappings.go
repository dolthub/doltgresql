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

// PgUserMappingsName is a constant to the pg_user_mappings name.
const PgUserMappingsName = "pg_user_mappings"

// InitPgUserMappings handles registration of the pg_user_mappings handler.
func InitPgUserMappings() {
	tables.AddHandler(PgCatalogName, PgUserMappingsName, PgUserMappingsHandler{})
}

// PgUserMappingsHandler is the handler for the pg_user_mappings table.
type PgUserMappingsHandler struct{}

var _ tables.Handler = PgUserMappingsHandler{}

// Name implements the interface tables.Handler.
func (p PgUserMappingsHandler) Name() string {
	return PgUserMappingsName
}

// RowIter implements the interface tables.Handler.
func (p PgUserMappingsHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_user_mappings row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgUserMappingsHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgUserMappingsSchema,
		PkOrdinals: nil,
	}
}

// pgUserMappingsSchema is the schema for pg_user_mappings.
var pgUserMappingsSchema = sql.Schema{
	{Name: "umid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgUserMappingsName},
	{Name: "srvid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgUserMappingsName},
	{Name: "srvname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgUserMappingsName},
	{Name: "umuser", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgUserMappingsName},
	{Name: "usename", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgUserMappingsName},
	{Name: "umoptions", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgUserMappingsName}, // TODO: collation C
}

// pgUserMappingsRowIter is the sql.RowIter for the pg_user_mappings table.
type pgUserMappingsRowIter struct {
}

var _ sql.RowIter = (*pgUserMappingsRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgUserMappingsRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgUserMappingsRowIter) Close(ctx *sql.Context) error {
	return nil
}
