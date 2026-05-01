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

// pgSequence represents a row in the pg_sequence table and pg_sequences view
type pgSequence struct {
	sequence *sequences.Sequence
	schema   string
	oid      id.Id
}

var _ tables.Handler = PgSequenceHandler{}

// Name implements the interface tables.Handler.
func (p PgSequenceHandler) Name() string {
	return PgSequenceName
}

// RowIter implements the interface tables.Handler.
func (p PgSequenceHandler) RowIter(ctx *sql.Context, _ sql.Partition) (sql.RowIter, error) {
	// Use cached data from this process if it exists
	pgCatalogCache, err := getPgCatalogCache(ctx)
	if err != nil {
		return nil, err
	}

	if pgCatalogCache.sequences == nil {
		err = cachePgSequences(ctx, pgCatalogCache)
		if err != nil {
			return nil, err
		}
	}

	return &pgSequenceRowIter{
		sequences: pgCatalogCache.sequences,
		idx:       0,
	}, nil
}

func cachePgSequences(ctx *sql.Context, pgCatalogCache *pgCatalogCache) error {
	var sequences []*pgSequence

	err := functions.IterateCurrentDatabase(ctx, functions.Callbacks{
		Sequence: func(ctx *sql.Context, schema functions.ItemSchema, sequence functions.ItemSequence) (cont bool, err error) {
			pgSeq := &pgSequence{
				sequence: sequence.Item,
				schema:   schema.Item.SchemaName(),
				oid:      sequence.OID.AsId(),
			}

			sequences = append(sequences, pgSeq)
			return true, nil
		},
	})
	if err != nil {
		return err
	}

	pgCatalogCache.sequences = sequences

	return nil
}

// Schema implements the interface tables.Handler.
func (p PgSequenceHandler) PkSchema() sql.PrimaryKeySchema {
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
	sequences []*pgSequence
	idx       int
}

var _ sql.RowIter = (*pgSequenceRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgSequenceRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.sequences) {
		return nil, io.EOF
	}
	iter.idx++
	sequence := iter.sequences[iter.idx-1].sequence
	oid := iter.sequences[iter.idx-1].oid
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
