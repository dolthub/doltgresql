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

// PgPublicationName is a constant to the pg_publication name.
const PgPublicationName = "pg_publication"

// InitPgPublication handles registration of the pg_publication handler.
func InitPgPublication() {
	tables.AddHandler(PgCatalogName, PgPublicationName, PgPublicationHandler{})
}

// PgPublicationHandler is the handler for the pg_publication table.
type PgPublicationHandler struct{}

var _ tables.Handler = PgPublicationHandler{}

// Name implements the interface tables.Handler.
func (p PgPublicationHandler) Name() string {
	return PgPublicationName
}

// RowIter implements the interface tables.Handler.
func (p PgPublicationHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_publication row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgPublicationHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgPublicationSchema,
		PkOrdinals: nil,
	}
}

// pgPublicationSchema is the schema for pg_publication.
var pgPublicationSchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgPublicationName},
	{Name: "pubname", Type: pgtypes.Name, Default: nil, Nullable: false, Source: PgPublicationName},
	{Name: "pubowner", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgPublicationName},
	{Name: "puballtables", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgPublicationName},
	{Name: "pubinsert", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgPublicationName},
	{Name: "pubupdate", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgPublicationName},
	{Name: "pubdelete", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgPublicationName},
	{Name: "pubtruncate", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgPublicationName},
	{Name: "pubviaroot", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgPublicationName},
}

// pgPublicationRowIter is the sql.RowIter for the pg_publication table.
type pgPublicationRowIter struct {
}

var _ sql.RowIter = (*pgPublicationRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgPublicationRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgPublicationRowIter) Close(ctx *sql.Context) error {
	return nil
}
