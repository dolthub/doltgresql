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
	"github.com/dolthub/doltgresql/core/sequences"
	"github.com/dolthub/doltgresql/server/functions"
	"github.com/dolthub/doltgresql/server/tables"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// PgSequenceName is a constant to the pg_sequence name.
const PgSequenceName = "pg_sequence"

// InitPgSequence handles registration of the pg_sequence handler.
func InitPgSequence() {
	tables.AddHandler(PgCatalogName, PgSequenceName, PgSequenceHandler{})
}

// PgSequenceHandler is the handler for the pg_sequence table.
type PgSequenceHandler struct{}

var _ tables.Handler = PgSequenceHandler{}

// Name implements the interface tables.Handler.
func (p PgSequenceHandler) Name() string {
	return PgSequenceName
}

// RowIter implements the interface tables.Handler.
func (p PgSequenceHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// Use cached data from this process if it exists
	pgCatalogCache, err := getPgCatalogCache(ctx)
	if err != nil {
		return nil, err
	}

	if pgCatalogCache.sequences == nil {
		var sequences []*sequences.Sequence
		var sequenceOids []id.Id
		err := functions.IterateCurrentDatabase(ctx, functions.Callbacks{
			Sequence: func(ctx *sql.Context, _ functions.ItemSchema, sequence functions.ItemSequence) (cont bool, err error) {
				sequences = append(sequences, sequence.Item)
				sequenceOids = append(sequenceOids, sequence.OID.AsId())
				return true, nil
			},
		})
		if err != nil {
			return nil, err
		}
		pgCatalogCache.sequences = sequences
		pgCatalogCache.sequenceOids = sequenceOids
	}

	return &pgSequenceRowIter{
		sequences: pgCatalogCache.sequences,
		oids:      pgCatalogCache.sequenceOids,
		idx:       0,
	}, nil
}

// Schema implements the interface tables.Handler.
func (p PgSequenceHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgSequenceSchema,
		PkOrdinals: nil,
	}
}

// pgSequenceSchema is the schema for pg_sequence.
var pgSequenceSchema = sql.Schema{
	{Name: "seqrelid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgSequenceName},
	{Name: "seqtypid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgSequenceName},
	{Name: "seqstart", Type: pgtypes.Int64, Default: nil, Nullable: false, Source: PgSequenceName},
	{Name: "seqincrement", Type: pgtypes.Int64, Default: nil, Nullable: false, Source: PgSequenceName},
	{Name: "seqmax", Type: pgtypes.Int64, Default: nil, Nullable: false, Source: PgSequenceName},
	{Name: "seqmin", Type: pgtypes.Int64, Default: nil, Nullable: false, Source: PgSequenceName},
	{Name: "seqcache", Type: pgtypes.Int64, Default: nil, Nullable: false, Source: PgSequenceName},
	{Name: "seqcycle", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgSequenceName},
}

// pgSequenceRowIter is the sql.RowIter for the pg_sequence table.
type pgSequenceRowIter struct {
	sequences []*sequences.Sequence
	oids      []id.Id
	idx       int
}

var _ sql.RowIter = (*pgSequenceRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgSequenceRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.sequences) {
		return nil, io.EOF
	}
	iter.idx++
	sequence := iter.sequences[iter.idx-1]
	oid := iter.oids[iter.idx-1]
	return sql.Row{
		oid,                        // seqrelid
		sequence.DataTypeID.AsId(), // seqtypid
		int64(sequence.Start),      // seqstart
		int64(sequence.Increment),  // seqincrement
		int64(sequence.Maximum),    // seqmax
		int64(sequence.Minimum),    // seqmin
		int64(sequence.Cache),      // seqcache
		bool(sequence.Cycle),       // seqcycle
	}, nil
}

// Close implements the interface sql.RowIter.
func (iter *pgSequenceRowIter) Close(ctx *sql.Context) error {
	return nil
}
