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

// PgStatProgressBasebackupName is a constant to the pg_stat_progress_basebackup name.
const PgStatProgressBasebackupName = "pg_stat_progress_basebackup"

// InitPgStatProgressBasebackup handles registration of the pg_stat_progress_basebackup handler.
func InitPgStatProgressBasebackup() {
	tables.AddHandler(PgCatalogName, PgStatProgressBasebackupName, PgStatProgressBasebackupHandler{})
}

// PgStatProgressBasebackupHandler is the handler for the pg_stat_progress_basebackup table.
type PgStatProgressBasebackupHandler struct{}

var _ tables.Handler = PgStatProgressBasebackupHandler{}

// Name implements the interface tables.Handler.
func (p PgStatProgressBasebackupHandler) Name() string {
	return PgStatProgressBasebackupName
}

// RowIter implements the interface tables.Handler.
func (p PgStatProgressBasebackupHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_stat_progress_basebackup row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgStatProgressBasebackupHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatProgressBasebackupSchema,
		PkOrdinals: nil,
	}
}

// pgStatProgressBasebackupSchema is the schema for pg_stat_progress_basebackup.
var pgStatProgressBasebackupSchema = sql.Schema{
	{Name: "pid", Type: pgtypes.Int32, Default: nil, Nullable: true, Source: PgStatProgressBasebackupName},
	{Name: "phase", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatProgressBasebackupName},
	{Name: "backup_total", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatProgressBasebackupName},
	{Name: "backup_streamed", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatProgressBasebackupName},
	{Name: "tablespaces_total", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatProgressBasebackupName},
	{Name: "tablespaces_streamed", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatProgressBasebackupName},
}

// pgStatProgressBasebackupRowIter is the sql.RowIter for the pg_stat_progress_basebackup table.
type pgStatProgressBasebackupRowIter struct {
}

var _ sql.RowIter = (*pgStatProgressBasebackupRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatProgressBasebackupRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgStatProgressBasebackupRowIter) Close(ctx *sql.Context) error {
	return nil
}
