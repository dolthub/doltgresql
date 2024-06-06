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
	"fmt"
	"io"

	"github.com/dolthub/dolt/go/libraries/doltcore/sqle"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/sequences"
	"github.com/dolthub/doltgresql/server/tables"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// PgSequenceName is a constant to the pg_sequence name.
const PgSequenceName = "pg_sequence"

// InitPgSequence handles registration of the pg_sequence handler.
func InitPgSequence() {
	tables.AddHandler(PgCatalogName, PgSequenceName, PgSequenceHandler{})
	tables.AddInitializeTable(PgCatalogName, pgSequenceInitializeTable)
}

// PgSequenceHandler is the handler for the pg_sequence table.
type PgSequenceHandler struct{}

var _ tables.DataTableHandler = PgSequenceHandler{}

// Insert implements the interface tables.DataTableHandler.
func (p PgSequenceHandler) Insert(ctx *sql.Context, editor *tables.DataTableEditor, row sql.Row) error {
	// Sequences are spread over multiple system tables, so we'll block insertion for now
	return fmt.Errorf("inserting into pg_sequence is not yet supported")
}

// Update implements the interface tables.DataTableHandler.
func (p PgSequenceHandler) Update(ctx *sql.Context, editor *tables.DataTableEditor, old sql.Row, new sql.Row) error {
	if len(old) != len(new) || len(old) != 8 {
		return fmt.Errorf("invalid row count given to %s: %d, %d", PgSequenceName, len(old), len(new))
	}
	allSequences, err := p.getAllSequencesOrdered(ctx)
	if err != nil {
		return err
	}
	idx := old[pgSequence_seqrelid].(uint32) - 1
	if idx >= uint32(len(allSequences)) {
		return fmt.Errorf("invalid %s given to %s: %d", pgSequenceSchema[pgSequence_seqrelid].Name, PgSequenceName, idx+1)
	}
	if old[pgSequence_seqrelid].(uint32) != new[pgSequence_seqrelid].(uint32) {
		return fmt.Errorf("%s does not support changing %s", PgSequenceName, pgSequenceSchema[pgSequence_seqrelid].Name)
	}
	if old[pgSequence_seqtypid].(uint32) != new[pgSequence_seqtypid].(uint32) {
		return fmt.Errorf("%s does not support changing %s", PgSequenceName, pgSequenceSchema[pgSequence_seqtypid].Name)
	}
	seq := allSequences[idx]
	seq.Start = new[pgSequence_seqstart].(int64)
	seq.Increment = new[pgSequence_seqincrement].(int64)
	seq.Maximum = new[pgSequence_seqmax].(int64)
	seq.Minimum = new[pgSequence_seqmin].(int64)
	seq.Cache = new[pgSequence_seqcache].(int64)
	seq.Cycle = new[pgSequence_seqcycle].(bool)
	if seq.Current < seq.Minimum || seq.Current > seq.Maximum {
		seq.IsAtEnd = true
	}
	return nil
}

// Delete implements the interface tables.DataTableHandler.
func (p PgSequenceHandler) Delete(ctx *sql.Context, editor *tables.DataTableEditor, row sql.Row) error {
	// Sequences are spread over multiple system tables, so we'll block deletion for now
	return fmt.Errorf("deleting from pg_sequence is not yet supported")
}

// UsesIndexes implements the interface tables.DataTableHandler.
func (p PgSequenceHandler) UsesIndexes() bool {
	return false
}

// RowIter implements the interface tables.DataTableHandler.
func (p PgSequenceHandler) RowIter(ctx *sql.Context, rowIter sql.RowIter) (sql.RowIter, error) {
	allSequences, err := p.getAllSequencesOrdered(ctx)
	if err != nil {
		return nil, err
	}
	return &pgSequenceRowIter{
		sequences: allSequences,
		idx:       0,
	}, nil
}

// getAllSequencesOrdered returns all sequences on the root, ordered by their schema and name.
func (p PgSequenceHandler) getAllSequencesOrdered(ctx *sql.Context) ([]*sequences.Sequence, error) {
	collection, err := core.GetCollectionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	allSequencesMap, schemas, count := collection.GetAllSequences()
	allSequences := make([]*sequences.Sequence, 0, count)
	for _, schemaName := range schemas {
		allSequences = append(allSequences, allSequencesMap[schemaName]...)
	}
	return allSequences, nil
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

const (
	pgSequence_seqrelid     int = 0
	pgSequence_seqtypid     int = 1
	pgSequence_seqstart     int = 2
	pgSequence_seqincrement int = 3
	pgSequence_seqmax       int = 4
	pgSequence_seqmin       int = 5
	pgSequence_seqcache     int = 6
	pgSequence_seqcycle     int = 7
)

// pgSequenceInitializeTable is the tables.InitializeTable function for pg_sequence.
func pgSequenceInitializeTable(ctx *sql.Context, db sqle.Database) error {
	return db.CreateIndexedTable(ctx, PgSequenceName, sql.PrimaryKeySchema{
		Schema: pgSequenceSchema,
	}, sql.IndexDef{
		Name:       "pg_sequence_seqrelid_index",
		Columns:    []sql.IndexColumn{{Name: pgSequenceSchema[pgSequence_seqrelid].Name}},
		Constraint: sql.IndexConstraint_Unique,
		Storage:    sql.IndexUsing_BTree,
		Comment:    "",
	}, sql.Collation_Default)
}

// pgSequenceRowIter is the sql.RowIter for the pg_sequence table.
type pgSequenceRowIter struct {
	sequences []*sequences.Sequence
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
	return sql.Row{
		uint32(iter.idx),             // seqrelid
		uint32(sequence.DataTypeOID), // seqtypid
		int64(sequence.Start),        // seqstart
		int64(sequence.Increment),    // seqincrement
		int64(sequence.Maximum),      // seqmax
		int64(sequence.Minimum),      // seqmin
		int64(sequence.Cache),        // seqcache
		bool(sequence.Cycle),         // seqcycle
	}, nil
}

// Close implements the interface sql.RowIter.
func (iter *pgSequenceRowIter) Close(ctx *sql.Context) error {
	return nil
}
