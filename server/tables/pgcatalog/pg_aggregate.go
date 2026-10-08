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

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/aggregates"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/tables"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// PgAggregateName is a constant to the pg_aggregate name.
const PgAggregateName = "pg_aggregate"

// InitPgAggregate handles registration of the pg_aggregate handler.
func InitPgAggregate() {
	tables.AddHandler(PgCatalogName, PgAggregateName, PgAggregateHandler{})
}

// PgAggregateHandler is the handler for the pg_aggregate table.
type PgAggregateHandler struct{}

var _ tables.Handler = PgAggregateHandler{}

// Name implements the interface tables.Handler.
func (p PgAggregateHandler) Name() string {
	return PgAggregateName
}

// RowIter implements the interface tables.Handler.
func (p PgAggregateHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: fill this in alongside built-in function entries in pg_proc
	aggregateCollection, err := core.GetAggregatesCollectionFromContext(ctx, "")
	if err != nil {
		return nil, err
	}
	var aggs []aggregates.Aggregate
	err = aggregateCollection.IterateAggregates(ctx, func(a aggregates.Aggregate) (stop bool, err error) {
		aggs = append(aggs, a)
		return false, nil
	})
	if err != nil {
		return nil, err
	}
	return &pgAggregateRowIter{aggs: aggs}, nil
}

// PkSchema implements the interface tables.Handler.
func (p PgAggregateHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgAggregateSchema,
		PkOrdinals: nil,
	}
}

// pgAggregateSchema is the schema for pg_aggregate.
var pgAggregateSchema = sql.Schema{
	{Name: "aggfnoid", Type: pgtypes.Regproc, Default: nil, Nullable: false, Source: PgAggregateName},
	{Name: "aggkind", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgAggregateName},
	{Name: "aggnumdirectargs", Type: pgtypes.Int16, Default: nil, Nullable: false, Source: PgAggregateName},
	{Name: "aggtransfn", Type: pgtypes.Regproc, Default: nil, Nullable: false, Source: PgAggregateName},
	{Name: "aggfinalfn", Type: pgtypes.Regproc, Default: nil, Nullable: false, Source: PgAggregateName},
	{Name: "aggcombinefn", Type: pgtypes.Regproc, Default: nil, Nullable: false, Source: PgAggregateName},
	{Name: "aggserialfn", Type: pgtypes.Regproc, Default: nil, Nullable: false, Source: PgAggregateName},
	{Name: "aggdeserialfn", Type: pgtypes.Regproc, Default: nil, Nullable: false, Source: PgAggregateName},
	{Name: "aggmtransfn", Type: pgtypes.Regproc, Default: nil, Nullable: false, Source: PgAggregateName},
	{Name: "aggminvtransfn", Type: pgtypes.Regproc, Default: nil, Nullable: false, Source: PgAggregateName},
	{Name: "aggmfinalfn", Type: pgtypes.Regproc, Default: nil, Nullable: false, Source: PgAggregateName},
	{Name: "aggfinalextra", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgAggregateName},
	{Name: "aggmfinalextra", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgAggregateName},
	{Name: "aggfinalmodify", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgAggregateName},
	{Name: "aggmfinalmodify", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgAggregateName},
	{Name: "aggsortop", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgAggregateName},
	{Name: "aggtranstype", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgAggregateName},
	{Name: "aggtransspace", Type: pgtypes.Int32, Default: nil, Nullable: false, Source: PgAggregateName},
	{Name: "aggmtranstype", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgAggregateName},
	{Name: "aggmtransspace", Type: pgtypes.Int32, Default: nil, Nullable: false, Source: PgAggregateName},
	{Name: "agginitval", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgAggregateName},  // TODO: collation C
	{Name: "aggminitval", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgAggregateName}, // TODO: collation C
}

// pgAggregateRowIter is the sql.RowIter for the pg_aggregate table.
type pgAggregateRowIter struct {
	aggs []aggregates.Aggregate
	idx  int
}

var _ sql.RowIter = (*pgAggregateRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgAggregateRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.aggs) {
		return nil, io.EOF
	}
	agg := iter.aggs[iter.idx]
	iter.idx++
	var initVal any
	if agg.HasInitCond {
		initVal = agg.InitCond
	}
	return sql.Row{
		agg.ID.AsId(),          // aggfnoid
		"n",                    // aggkind
		int16(0),               // aggnumdirectargs
		agg.SFunc.AsId(),       // aggtransfn
		agg.FinalFunc.AsId(),   // aggfinalfn
		agg.CombineFunc.AsId(), // aggcombinefn
		id.Null,                // aggserialfn
		id.Null,                // aggdeserialfn
		id.Null,                // aggmtransfn
		id.Null,                // aggminvtransfn
		id.Null,                // aggmfinalfn
		false,                  // aggfinalextra
		false,                  // aggmfinalextra
		"r",                    // aggfinalmodify
		"r",                    // aggmfinalmodify
		id.Null,                // aggsortop
		agg.SType.AsId(),       // aggtranstype
		int32(0),               // aggtransspace
		id.Null,                // aggmtranstype
		int32(0),               // aggmtransspace
		initVal,                // agginitval
		nil,                    // aggminitval
	}, nil
}

// Close implements the interface sql.RowIter.
func (iter *pgAggregateRowIter) Close(ctx *sql.Context) error {
	return nil
}
