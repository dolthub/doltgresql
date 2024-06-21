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

// PgAmprocName is a constant to the pg_amproc name.
const PgAmprocName = "pg_amproc"

// InitPgAmproc handles registration of the pg_amproc handler.
func InitPgAmproc() {
	tables.AddHandler(PgCatalogName, PgAmprocName, PgAmprocHandler{})
}

// PgAmprocHandler is the handler for the pg_amproc table.
type PgAmprocHandler struct{}

var _ tables.Handler = PgAmprocHandler{}

// Name implements the interface tables.Handler.
func (p PgAmprocHandler) Name() string {
	return PgAmprocName
}

// RowIter implements the interface tables.Handler.
func (p PgAmprocHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_amproc row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgAmprocHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgAmprocSchema,
		PkOrdinals: nil,
	}
}

// pgAmprocSchema is the schema for pg_amproc.
var pgAmprocSchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgAmprocName},
	{Name: "amprocfamily", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgAmprocName},
	{Name: "amproclefttype", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgAmprocName},
	{Name: "amprocrighttype", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgAmprocName},
	{Name: "amprocnum", Type: pgtypes.Int16, Default: nil, Nullable: false, Source: PgAmprocName},
	{Name: "amproc", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgAmprocName}, // TODO: regproc type
}

// pgAmprocRowIter is the sql.RowIter for the pg_amproc table.
type pgAmprocRowIter struct {
}

var _ sql.RowIter = (*pgAmprocRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgAmprocRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgAmprocRowIter) Close(ctx *sql.Context) error {
	return nil
}
