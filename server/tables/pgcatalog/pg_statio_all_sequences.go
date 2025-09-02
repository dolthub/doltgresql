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

// PgStatioAllSequencesName is a constant to the pg_statio_all_sequences name.
const PgStatioAllSequencesName = "pg_statio_all_sequences"

// InitPgStatioAllSequences handles registration of the pg_statio_all_sequences handler.
func InitPgStatioAllSequences() {
	tables.AddHandler(PgCatalogName, PgStatioAllSequencesName, PgStatioAllSequencesHandler{})
}

// PgStatioAllSequencesHandler is the handler for the pg_statio_all_sequences table.
type PgStatioAllSequencesHandler struct{}

var _ tables.Handler = PgStatioAllSequencesHandler{}

// Name implements the interface tables.Handler.
func (p PgStatioAllSequencesHandler) Name() string {
	return PgStatioAllSequencesName
}

// RowIter implements the interface tables.Handler.
func (p PgStatioAllSequencesHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_statio_all_sequences row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgStatioAllSequencesHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatioAllSequencesSchema,
		PkOrdinals: nil,
	}
}

// pgStatioAllSequencesSchema is the schema for pg_statio_all_sequences.
var pgStatioAllSequencesSchema = sql.Schema{
	{Name: "relid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatioAllSequencesName},
	{Name: "schemaname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatioAllSequencesName},
	{Name: "relname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatioAllSequencesName},
	{Name: "blks_read", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatioAllSequencesName},
	{Name: "blks_hit", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatioAllSequencesName},
}

// pgStatioAllSequencesRowIter is the sql.RowIter for the pg_statio_all_sequences table.
type pgStatioAllSequencesRowIter struct {
}

var _ sql.RowIter = (*pgStatioAllSequencesRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatioAllSequencesRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgStatioAllSequencesRowIter) Close(ctx *sql.Context) error {
	return nil
}
