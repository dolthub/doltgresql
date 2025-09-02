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

// PgOpfamilyName is a constant to the pg_opfamily name.
const PgOpfamilyName = "pg_opfamily"

// InitPgOpfamily handles registration of the pg_opfamily handler.
func InitPgOpfamily() {
	tables.AddHandler(PgCatalogName, PgOpfamilyName, PgOpfamilyHandler{})
}

// PgOpfamilyHandler is the handler for the pg_opfamily table.
type PgOpfamilyHandler struct{}

var _ tables.Handler = PgOpfamilyHandler{}

// Name implements the interface tables.Handler.
func (p PgOpfamilyHandler) Name() string {
	return PgOpfamilyName
}

// RowIter implements the interface tables.Handler.
func (p PgOpfamilyHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_opfamily row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgOpfamilyHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgOpfamilySchema,
		PkOrdinals: nil,
	}
}

// pgOpfamilySchema is the schema for pg_opfamily.
var pgOpfamilySchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgOpfamilyName},
	{Name: "opfmethod", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgOpfamilyName},
	{Name: "opfname", Type: pgtypes.Name, Default: nil, Nullable: false, Source: PgOpfamilyName},
	{Name: "opfnamespace", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgOpfamilyName},
	{Name: "opfowner", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgOpfamilyName},
}

// pgOpfamilyRowIter is the sql.RowIter for the pg_opfamily table.
type pgOpfamilyRowIter struct {
}

var _ sql.RowIter = (*pgOpfamilyRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgOpfamilyRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgOpfamilyRowIter) Close(ctx *sql.Context) error {
	return nil
}
