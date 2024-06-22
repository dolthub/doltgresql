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

// PgStatisticExtName is a constant to the pg_statistic_ext name.
const PgStatisticExtName = "pg_statistic_ext"

// InitPgStatisticExt handles registration of the pg_statistic_ext handler.
func InitPgStatisticExt() {
	tables.AddHandler(PgCatalogName, PgStatisticExtName, PgStatisticExtHandler{})
}

// PgStatisticExtHandler is the handler for the pg_statistic_ext table.
type PgStatisticExtHandler struct{}

var _ tables.Handler = PgStatisticExtHandler{}

// Name implements the interface tables.Handler.
func (p PgStatisticExtHandler) Name() string {
	return PgStatisticExtName
}

// RowIter implements the interface tables.Handler.
func (p PgStatisticExtHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_statistic_ext row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgStatisticExtHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatisticExtSchema,
		PkOrdinals: nil,
	}
}

// pgStatisticExtSchema is the schema for pg_statistic_ext.
var pgStatisticExtSchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgStatisticExtName},
	{Name: "stxrelid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgStatisticExtName},
	{Name: "stxname", Type: pgtypes.Name, Default: nil, Nullable: false, Source: PgStatisticExtName},
	{Name: "stxnamespace", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgStatisticExtName},
	{Name: "stxowner", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgStatisticExtName},
	{Name: "stxstattarget", Type: pgtypes.Int32, Default: nil, Nullable: false, Source: PgStatisticExtName},
	{Name: "stxkeys", Type: pgtypes.Int16Array, Default: nil, Nullable: false, Source: PgStatisticExtName}, // TODO: int2vector type
	{Name: "stxkind", Type: pgtypes.BpCharArray, Default: nil, Nullable: false, Source: PgStatisticExtName},
	{Name: "stxexprs", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatisticExtName}, // TODO: collation C, pg_node_tree type
}

// pgStatisticExtRowIter is the sql.RowIter for the pg_statistic_ext table.
type pgStatisticExtRowIter struct {
}

var _ sql.RowIter = (*pgStatisticExtRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatisticExtRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgStatisticExtRowIter) Close(ctx *sql.Context) error {
	return nil
}
