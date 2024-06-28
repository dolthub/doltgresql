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

// PgStatioSysSequencesName is a constant to the pg_statio_sys_sequences name.
const PgStatioSysSequencesName = "pg_statio_sys_sequences"

// InitPgStatioSysSequences handles registration of the pg_statio_sys_sequences handler.
func InitPgStatioSysSequences() {
	tables.AddHandler(PgCatalogName, PgStatioSysSequencesName, PgStatioSysSequencesHandler{})
}

// PgStatioSysSequencesHandler is the handler for the pg_statio_sys_sequences table.
type PgStatioSysSequencesHandler struct{}

var _ tables.Handler = PgStatioSysSequencesHandler{}

// Name implements the interface tables.Handler.
func (p PgStatioSysSequencesHandler) Name() string {
	return PgStatioSysSequencesName
}

// RowIter implements the interface tables.Handler.
func (p PgStatioSysSequencesHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_statio_sys_sequences row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgStatioSysSequencesHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatioSysSequencesSchema,
		PkOrdinals: nil,
	}
}

// pgStatioSysSequencesSchema is the schema for pg_statio_sys_sequences.
var pgStatioSysSequencesSchema = sql.Schema{
	{Name: "relid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatioSysSequencesName},
	{Name: "schemaname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatioSysSequencesName},
	{Name: "relname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatioSysSequencesName},
	{Name: "blks_read", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatioSysSequencesName},
	{Name: "blks_hit", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatioSysSequencesName},
}

// pgStatioSysSequencesRowIter is the sql.RowIter for the pg_statio_sys_sequences table.
type pgStatioSysSequencesRowIter struct {
}

var _ sql.RowIter = (*pgStatioSysSequencesRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatioSysSequencesRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgStatioSysSequencesRowIter) Close(ctx *sql.Context) error {
	return nil
}
