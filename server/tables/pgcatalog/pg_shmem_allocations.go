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

// PgShmemAllocationsName is a constant to the pg_shmem_allocations name.
const PgShmemAllocationsName = "pg_shmem_allocations"

// InitPgShmemAllocations handles registration of the pg_shmem_allocations handler.
func InitPgShmemAllocations() {
	tables.AddHandler(PgCatalogName, PgShmemAllocationsName, PgShmemAllocationsHandler{})
}

// PgShmemAllocationsHandler is the handler for the pg_shmem_allocations table.
type PgShmemAllocationsHandler struct{}

var _ tables.Handler = PgShmemAllocationsHandler{}

// Name implements the interface tables.Handler.
func (p PgShmemAllocationsHandler) Name() string {
	return PgShmemAllocationsName
}

// RowIter implements the interface tables.Handler.
func (p PgShmemAllocationsHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_shmem_allocations row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgShmemAllocationsHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgShmemAllocationsSchema,
		PkOrdinals: nil,
	}
}

// pgShmemAllocationsSchema is the schema for pg_shmem_allocations.
var pgShmemAllocationsSchema = sql.Schema{
	{Name: "name", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgShmemAllocationsName},
	{Name: "off", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgShmemAllocationsName},
	{Name: "size", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgShmemAllocationsName},
	{Name: "allocated_size", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgShmemAllocationsName},
}

// pgShmemAllocationsRowIter is the sql.RowIter for the pg_shmem_allocations table.
type pgShmemAllocationsRowIter struct {
}

var _ sql.RowIter = (*pgShmemAllocationsRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgShmemAllocationsRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgShmemAllocationsRowIter) Close(ctx *sql.Context) error {
	return nil
}
