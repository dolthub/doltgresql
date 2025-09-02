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

// PgSequencesName is a constant to the pg_sequences name.
const PgSequencesName = "pg_sequences"

// InitPgSequences handles registration of the pg_sequences handler.
func InitPgSequences() {
	tables.AddHandler(PgCatalogName, PgSequencesName, PgSequencesHandler{})
}

// PgSequencesHandler is the handler for the pg_sequences table.
type PgSequencesHandler struct{}

var _ tables.Handler = PgSequencesHandler{}

// Name implements the interface tables.Handler.
func (p PgSequencesHandler) Name() string {
	return PgSequencesName
}

// RowIter implements the interface tables.Handler.
func (p PgSequencesHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_sequences row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgSequencesHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgSequencesSchema,
		PkOrdinals: nil,
	}
}

// pgSequencesSchema is the schema for pg_sequences.
var pgSequencesSchema = sql.Schema{
	{Name: "schemaname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgSequencesName},
	{Name: "sequencename", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgSequencesName},
	{Name: "sequenceowner", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgSequencesName},
	{Name: "data_type", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgSequencesName}, // TODO: regtype type
	{Name: "start_value", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgSequencesName},
	{Name: "min_value", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgSequencesName},
	{Name: "max_value", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgSequencesName},
	{Name: "increment_by", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgSequencesName},
	{Name: "cycle", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgSequencesName},
	{Name: "cache_size", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgSequencesName},
	{Name: "last_value", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgSequencesName},
}

// pgSequencesRowIter is the sql.RowIter for the pg_sequences table.
type pgSequencesRowIter struct {
}

var _ sql.RowIter = (*pgSequencesRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgSequencesRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgSequencesRowIter) Close(ctx *sql.Context) error {
	return nil
}
