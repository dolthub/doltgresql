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

// PgStatioUserSequencesName is a constant to the pg_statio_user_sequences name.
const PgStatioUserSequencesName = "pg_statio_user_sequences"

// InitPgStatioUserSequences handles registration of the pg_statio_user_sequences handler.
func InitPgStatioUserSequences() {
	tables.AddHandler(PgCatalogName, PgStatioUserSequencesName, PgStatioUserSequencesHandler{})
}

// PgStatioUserSequencesHandler is the handler for the pg_statio_user_sequences table.
type PgStatioUserSequencesHandler struct{}

var _ tables.Handler = PgStatioUserSequencesHandler{}

// Name implements the interface tables.Handler.
func (p PgStatioUserSequencesHandler) Name() string {
	return PgStatioUserSequencesName
}

// RowIter implements the interface tables.Handler.
func (p PgStatioUserSequencesHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_statio_user_sequences row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgStatioUserSequencesHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatioUserSequencesSchema,
		PkOrdinals: nil,
	}
}

// pgStatioUserSequencesSchema is the schema for pg_statio_user_sequences.
var pgStatioUserSequencesSchema = sql.Schema{
	{Name: "relid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatioUserSequencesName},
	{Name: "schemaname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatioUserSequencesName},
	{Name: "relname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatioUserSequencesName},
	{Name: "blks_read", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatioUserSequencesName},
	{Name: "blks_hit", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatioUserSequencesName},
}

// pgStatioUserSequencesRowIter is the sql.RowIter for the pg_statio_user_sequences table.
type pgStatioUserSequencesRowIter struct {
}

var _ sql.RowIter = (*pgStatioUserSequencesRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatioUserSequencesRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgStatioUserSequencesRowIter) Close(ctx *sql.Context) error {
	return nil
}
