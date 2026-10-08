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

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/functions"
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
	entries, err := getStatioSequenceEntries(ctx, statSchemaAll)
	if err != nil {
		return nil, err
	}
	return &pgStatioSequencesRowIter{entries: entries}, nil
}

// PkSchema implements the interface tables.Handler.
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

// statioSequenceEntry identifies a sequence row in the pg_statio_*_sequences tables.
type statioSequenceEntry struct {
	oid          id.Id
	schemaName   string
	sequenceName string
	isSystem     bool
}

// getStatioSequenceEntries returns a statioSequenceEntry for each sequence in the current database
// whose schema matches the given filter. The unfiltered entry list is cached in the session's
// pg_catalog cache, since iterating all schema elements is expensive.
func getStatioSequenceEntries(ctx *sql.Context, filter statSchemaFilter) ([]statioSequenceEntry, error) {
	pgCatalogCache, err := getPgCatalogCache(ctx)
	if err != nil {
		return nil, err
	}

	if pgCatalogCache.statioSequenceEntries == nil {
		var entries []statioSequenceEntry
		err := functions.IterateCurrentDatabase(ctx, functions.Callbacks{
			Sequence: func(ctx *sql.Context, schema functions.ItemSchema, sequence functions.ItemSequence) (cont bool, err error) {
				entries = append(entries, statioSequenceEntry{
					oid:          sequence.OID.AsId(),
					schemaName:   schema.Item.SchemaName(),
					sequenceName: sequence.Item.Id.SequenceName(),
					isSystem:     schema.IsSystemSchema(),
				})
				return true, nil
			},
		})
		if err != nil {
			return nil, err
		}
		pgCatalogCache.statioSequenceEntries = entries
	}

	var filtered []statioSequenceEntry
	for _, entry := range pgCatalogCache.statioSequenceEntries {
		if filter(entry.isSystem) {
			filtered = append(filtered, entry)
		}
	}
	return filtered, nil
}

// pgStatioSequencesRowIter is the sql.RowIter for the pg_statio_all_sequences,
// pg_statio_sys_sequences, and pg_statio_user_sequences tables. All block I/O counters are zero,
// since Doltgres does not track block-level I/O statistics. This matches what a freshly-started
// Postgres server reports.
// TODO: fill in the I/O statistics columns when block I/O statistics are tracked
type pgStatioSequencesRowIter struct {
	entries []statioSequenceEntry
	idx     int
}

var _ sql.RowIter = (*pgStatioSequencesRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatioSequencesRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.entries) {
		return nil, io.EOF
	}
	iter.idx++
	entry := iter.entries[iter.idx-1]
	return sql.Row{
		entry.oid,          // relid
		entry.schemaName,   // schemaname
		entry.sequenceName, // relname
		int64(0),           // blks_read
		int64(0),           // blks_hit
	}, nil
}

// Close implements the interface sql.RowIter.
func (iter *pgStatioSequencesRowIter) Close(ctx *sql.Context) error {
	return nil
}
