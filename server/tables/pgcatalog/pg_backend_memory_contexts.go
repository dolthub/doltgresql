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

// PgBackendMemoryContextsName is a constant to the pg_backend_memory_contexts name.
const PgBackendMemoryContextsName = "pg_backend_memory_contexts"

// InitPgBackendMemoryContexts handles registration of the pg_backend_memory_contexts handler.
func InitPgBackendMemoryContexts() {
	tables.AddHandler(PgCatalogName, PgBackendMemoryContextsName, PgBackendMemoryContextsHandler{})
}

// PgBackendMemoryContextsHandler is the handler for the pg_backend_memory_contexts table.
type PgBackendMemoryContextsHandler struct{}

var _ tables.Handler = PgBackendMemoryContextsHandler{}

// Name implements the interface tables.Handler.
func (p PgBackendMemoryContextsHandler) Name() string {
	return PgBackendMemoryContextsName
}

// RowIter implements the interface tables.Handler.
func (p PgBackendMemoryContextsHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_backend_memory_contexts row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgBackendMemoryContextsHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgBackendMemoryContextsSchema,
		PkOrdinals: nil,
	}
}

// pgBackendMemoryContextsSchema is the schema for pg_backend_memory_contexts.
var pgBackendMemoryContextsSchema = sql.Schema{
	{Name: "name", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgBackendMemoryContextsName},
	{Name: "ident", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgBackendMemoryContextsName},
	{Name: "parent", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgBackendMemoryContextsName},
	{Name: "level", Type: pgtypes.Int32, Default: nil, Nullable: true, Source: PgBackendMemoryContextsName},
	{Name: "total_bytes", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgBackendMemoryContextsName},
	{Name: "total_nblocks", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgBackendMemoryContextsName},
	{Name: "free_bytes", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgBackendMemoryContextsName},
	{Name: "free_chunks", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgBackendMemoryContextsName},
	{Name: "used_bytes", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgBackendMemoryContextsName},
}

// pgBackendMemoryContextsRowIter is the sql.RowIter for the pg_backend_memory_contexts table.
type pgBackendMemoryContextsRowIter struct {
}

var _ sql.RowIter = (*pgBackendMemoryContextsRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgBackendMemoryContextsRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgBackendMemoryContextsRowIter) Close(ctx *sql.Context) error {
	return nil
}
