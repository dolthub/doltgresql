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

// PgTransformName is a constant to the pg_transform name.
const PgTransformName = "pg_transform"

// InitPgTransform handles registration of the pg_transform handler.
func InitPgTransform() {
	tables.AddHandler(PgCatalogName, PgTransformName, PgTransformHandler{})
}

// PgTransformHandler is the handler for the pg_transform table.
type PgTransformHandler struct{}

var _ tables.Handler = PgTransformHandler{}

// Name implements the interface tables.Handler.
func (p PgTransformHandler) Name() string {
	return PgTransformName
}

// RowIter implements the interface tables.Handler.
func (p PgTransformHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_transform row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgTransformHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgTransformSchema,
		PkOrdinals: nil,
	}
}

// pgTransformSchema is the schema for pg_transform.
var pgTransformSchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgTransformName},
	{Name: "trftype", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgTransformName},
	{Name: "trflang", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgTransformName},
	{Name: "trffromsql", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgTransformName}, // TODO: regproc type
	{Name: "trftosql", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgTransformName},   // TODO: regproc type
}

// pgTransformRowIter is the sql.RowIter for the pg_transform table.
type pgTransformRowIter struct {
}

var _ sql.RowIter = (*pgTransformRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgTransformRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgTransformRowIter) Close(ctx *sql.Context) error {
	return nil
}
