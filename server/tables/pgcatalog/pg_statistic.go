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

// PgStatisticName is a constant to the pg_statistic name.
const PgStatisticName = "pg_statistic"

// InitPgStatistic handles registration of the pg_statistic handler.
func InitPgStatistic() {
	tables.AddHandler(PgCatalogName, PgStatisticName, PgStatisticHandler{})
}

// PgStatisticHandler is the handler for the pg_statistic table.
type PgStatisticHandler struct{}

var _ tables.Handler = PgStatisticHandler{}

// Name implements the interface tables.Handler.
func (p PgStatisticHandler) Name() string {
	return PgStatisticName
}

// RowIter implements the interface tables.Handler.
func (p PgStatisticHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_statistic row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgStatisticHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatisticSchema,
		PkOrdinals: nil,
	}
}

// pgStatisticSchema is the schema for pg_statistic.
var pgStatisticSchema = sql.Schema{
	{Name: "starelid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgStatisticName},
	{Name: "staattnum", Type: pgtypes.Int16, Default: nil, Nullable: false, Source: PgStatisticName},
	{Name: "stainherit", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgStatisticName},
	{Name: "stanullfrac", Type: pgtypes.Float32, Default: nil, Nullable: false, Source: PgStatisticName},
	{Name: "stawidth", Type: pgtypes.Int32, Default: nil, Nullable: false, Source: PgStatisticName},
	{Name: "stadistinct", Type: pgtypes.Float32, Default: nil, Nullable: false, Source: PgStatisticName},
	{Name: "stakind1", Type: pgtypes.Int16, Default: nil, Nullable: false, Source: PgStatisticName},
	{Name: "stakind2", Type: pgtypes.Int16, Default: nil, Nullable: false, Source: PgStatisticName},
	{Name: "stakind3", Type: pgtypes.Int16, Default: nil, Nullable: false, Source: PgStatisticName},
	{Name: "stakind4", Type: pgtypes.Int16, Default: nil, Nullable: false, Source: PgStatisticName},
	{Name: "stakind5", Type: pgtypes.Int16, Default: nil, Nullable: false, Source: PgStatisticName},
	{Name: "staop1", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgStatisticName},
	{Name: "staop2", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgStatisticName},
	{Name: "staop3", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgStatisticName},
	{Name: "staop4", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgStatisticName},
	{Name: "staop5", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgStatisticName},
	{Name: "stacoll1", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgStatisticName},
	{Name: "stacoll2", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgStatisticName},
	{Name: "stacoll3", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgStatisticName},
	{Name: "stacoll4", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgStatisticName},
	{Name: "stacoll5", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgStatisticName},
	{Name: "stanumbers1", Type: pgtypes.Float32Array, Default: nil, Nullable: true, Source: PgStatisticName},
	{Name: "stanumbers2", Type: pgtypes.Float32Array, Default: nil, Nullable: true, Source: PgStatisticName},
	{Name: "stanumbers3", Type: pgtypes.Float32Array, Default: nil, Nullable: true, Source: PgStatisticName},
	{Name: "stanumbers4", Type: pgtypes.Float32Array, Default: nil, Nullable: true, Source: PgStatisticName},
	{Name: "stanumbers5", Type: pgtypes.Float32Array, Default: nil, Nullable: true, Source: PgStatisticName},
	{Name: "stavalues1", Type: pgtypes.AnyArray, Default: nil, Nullable: true, Source: PgStatisticName},
	{Name: "stavalues2", Type: pgtypes.AnyArray, Default: nil, Nullable: true, Source: PgStatisticName},
	{Name: "stavalues3", Type: pgtypes.AnyArray, Default: nil, Nullable: true, Source: PgStatisticName},
	{Name: "stavalues4", Type: pgtypes.AnyArray, Default: nil, Nullable: true, Source: PgStatisticName},
	{Name: "stavalues5", Type: pgtypes.AnyArray, Default: nil, Nullable: true, Source: PgStatisticName},
}

// pgStatisticRowIter is the sql.RowIter for the pg_statistic table.
type pgStatisticRowIter struct {
}

var _ sql.RowIter = (*pgStatisticRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatisticRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgStatisticRowIter) Close(ctx *sql.Context) error {
	return nil
}
