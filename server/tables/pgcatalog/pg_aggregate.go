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
func (p PgAggregateHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_aggregate row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgAggregateHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgAggregateSchema,
		PkOrdinals: nil,
	}
}

// pgAggregateSchema is the schema for pg_aggregate.
var pgAggregateSchema = sql.Schema{
	{Name: "aggfnoid", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgAggregateName}, // TODO: regproc type
	{Name: "aggkind", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgAggregateName},
	{Name: "aggnumdirectargs", Type: pgtypes.Int16, Default: nil, Nullable: false, Source: PgAggregateName},
	{Name: "aggtransfn", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgAggregateName},     // TODO: regproc type
	{Name: "aggfinalfn", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgAggregateName},     // TODO: regproc type
	{Name: "aggcombinefn", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgAggregateName},   // TODO: regproc type
	{Name: "aggserialfn", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgAggregateName},    // TODO: regproc type
	{Name: "aggdeserialfn", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgAggregateName},  // TODO: regproc type
	{Name: "aggmtransfn", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgAggregateName},    // TODO: regproc type
	{Name: "aggminvtransfn", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgAggregateName}, // TODO: regproc type
	{Name: "aggmfinalfn", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgAggregateName},    // TODO: regproc type
	{Name: "aggfinalextra", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgAggregateName},
	{Name: "aggmfinalextra", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgAggregateName},
	{Name: "aggfinalmodify", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgAggregateName},
	{Name: "aggmfinalmodify", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgAggregateName},
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
}

var _ sql.RowIter = (*pgAggregateRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgAggregateRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgAggregateRowIter) Close(ctx *sql.Context) error {
	return nil
}
