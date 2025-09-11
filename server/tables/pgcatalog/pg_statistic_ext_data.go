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

// PgStatisticExtDataName is a constant to the pg_statistic_ext_data name.
const PgStatisticExtDataName = "pg_statistic_ext_data"

// InitPgStatisticExtData handles registration of the pg_statistic_ext_data handler.
func InitPgStatisticExtData() {
	tables.AddHandler(PgCatalogName, PgStatisticExtDataName, PgStatisticExtDataHandler{})
}

// PgStatisticExtDataHandler is the handler for the pg_statistic_ext_data table.
type PgStatisticExtDataHandler struct{}

var _ tables.Handler = PgStatisticExtDataHandler{}

// Name implements the interface tables.Handler.
func (p PgStatisticExtDataHandler) Name() string {
	return PgStatisticExtDataName
}

// RowIter implements the interface tables.Handler.
func (p PgStatisticExtDataHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_statistic_ext_data row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgStatisticExtDataHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatisticExtDataSchema,
		PkOrdinals: nil,
	}
}

// pgStatisticExtDataSchema is the schema for pg_statistic_ext_data.
var pgStatisticExtDataSchema = sql.Schema{
	{Name: "stxoid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgStatisticExtDataName},
	{Name: "stxdinherit", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgStatisticExtDataName},
	{Name: "stxdndistinct", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatisticExtDataName},    // TODO: pg_ndistinct type, collation C
	{Name: "stxddependencies", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatisticExtDataName}, // TODO: pg_dependencies type, collation C
	{Name: "stxdmcv", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatisticExtDataName},          // TODO: pg_mcv_list type, collation C
	{Name: "stxdexpr", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgStatisticExtDataName},    // TODO: pg_statistic[] type
}

// pgStatisticExtDataRowIter is the sql.RowIter for the pg_statistic_ext_data table.
type pgStatisticExtDataRowIter struct {
}

var _ sql.RowIter = (*pgStatisticExtDataRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatisticExtDataRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgStatisticExtDataRowIter) Close(ctx *sql.Context) error {
	return nil
}
